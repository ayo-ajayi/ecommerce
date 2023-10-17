package user

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ayo-ajayi/ecommerce/internal/errors"
	"github.com/ayo-ajayi/ecommerce/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserService struct {
	userRepository  UserRepository
	otpRepository   OTPRepository
	emailRepository EmailRepository
	tokenRepository TokenRepository
}

func NewUserService(userRepository UserRepository,
	otpRepository OTPRepository, emailRepository EmailRepository, tokenRepository TokenRepository) *UserService {
	return &UserService{
		userRepository,
		otpRepository,
		emailRepository,
		tokenRepository,
	}
}

type TokenRepository interface {
	GenerateToken(userId primitive.ObjectID) (*utils.TokenDetails, error)
	SaveToken(userId primitive.ObjectID, td *utils.TokenDetails) error
	DeleteToken(uuid string) error
	IdentifyUser(refreshToken string) (*utils.RefreshDetails, error)
}

type EmailRepository interface {
	SendAccountVerificationEmail(email, firstname, otp string) error
	SendResetPasswordEmail(email, firstname, otp string) error
}

type OTPRepository interface {
	GenerateSignUpOTP(email string) (string, error)
	GenerateForgotPasswordOTP(email string) (string, error)
	VerifySignUpOTP(email, otpStr string) (bool, error)
	VerifyForgotPasswordOTP(email, otpStr string) (bool, error)
	DeleteOTP(email string) error
}

func (us *UserService) SignUpAndSendVerificationEmail(user *User) *errors.AppError {

	var wg sync.WaitGroup

	var userExists bool
	var passwordHash string
	var otp string

	errCh := make(chan *errors.AppError, 3)

	wg.Add(3)
	go func() {

		defer wg.Done()
		exists, err := us.userRepository.IsExists(user.Email)
		if err != nil {

			errCh <- errors.ErrInternalServer
			return
		}
		userExists = exists
	}()
	go func() {
		defer wg.Done()
		hash, err := utils.HashPassword(user.Password)
		if err != nil || hash == "" {

			errCh <- errors.ErrInternalServer
			return
		}
		passwordHash = hash
	}()
	go func() {
		defer wg.Done()
		generatedOtp, err := us.otpRepository.GenerateSignUpOTP(user.Email)
		if err != nil {

			errCh <- errors.NewError("failed to generate otp"+err.Error(), http.StatusInternalServerError)
			return
		}
		if generatedOtp == "" {
			errCh <- errors.ErrInternalServer
			return
		}
		otp = generatedOtp
	}()
	go func() {
		wg.Wait()
		close(errCh)
	}()

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	if userExists {
		return errors.ErrUserAlreadyExists
	}

	user.Password = passwordHash
	user.Role = Customer
	user.CreatedAt = time.Now()
	user.UpdatedAt = user.CreatedAt

	if err := us.userRepository.CreateUser(user); err != nil {
		return errors.ErrInternalServer
	}

	if otp == "" {
		return errors.ErrInternalServer
	}

	if err := us.emailRepository.SendAccountVerificationEmail(user.Email, user.FirstName, otp); err != nil {
		return errors.NewError("failed to send user otp"+err.Error(), http.StatusInternalServerError)
	}
	return nil
}

func (us *UserService) VerifyUser(email, otp string) *errors.AppError {
	valid, err := us.otpRepository.VerifySignUpOTP(email, otp)
	if err != nil {
		us.otpRepository.DeleteOTP(email)
		return errors.NewError("failed to verify otp: invalid otp: "+err.Error(), errors.ErrInvalidOTP.StatusCode)
	}
	if !valid {
		us.otpRepository.DeleteOTP(email)
		return errors.ErrInvalidOTP
	}
	err = us.otpRepository.DeleteOTP(email)
	if err != nil {
		return errors.ErrInternalServer
	}
	u, e := us.userRepository.GetUser(bson.M{"email": email})
	if e != nil {
		if e == mongo.ErrNoDocuments {
			err := errors.ErrNotFound
			return errors.NewError("user not found: "+err.Error(), err.StatusCode)
		}
		return errors.ErrInternalServer
	}
	if u.IsVerified {
		return errors.ErrUserAlreadyVerified
	}
	err = us.userRepository.UpdateUser(bson.M{"email": email}, bson.M{"$set": bson.M{"is_verified": true}})
	if err != nil {
		return errors.ErrInternalServer
	}
	return nil
}

func (us *UserService) ResendEmailVerificationOTP(email string) *errors.AppError {
	user, err := us.userRepository.GetUser(bson.M{"email": email})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err := errors.ErrNotFound
			return errors.NewError("user not found: "+err.Error(), err.StatusCode)
		}
		return errors.ErrInternalServer
	}
	if user.IsVerified {
		return errors.ErrUserAlreadyVerified
	}

	otp, err := us.otpRepository.GenerateSignUpOTP(email)
	if err != nil {
		return errors.NewError("failed to generate otp"+err.Error(), http.StatusInternalServerError)
	}
	if otp == "" {
		return errors.ErrInternalServer
	}
	err = us.emailRepository.SendAccountVerificationEmail(email, user.FirstName, otp)
	if err != nil {
		return errors.NewError("failed to send user otp"+err.Error(), http.StatusInternalServerError)
	}
	return nil
}

func (us *UserService) Login(email, password string) (*User, *utils.TokenDetails, *errors.AppError) {
	user, err := us.userRepository.GetUser(bson.M{"email": email})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err := errors.ErrNotFound
			return nil, nil, errors.NewError("user not found: "+err.Error(), err.StatusCode)
		}
		return nil, nil, errors.ErrInternalServer
	}
	if !user.IsVerified {
		return nil, nil, errors.NewError("user is not verified", http.StatusForbidden)
	}

	var passwordMatches bool
	var token *utils.TokenDetails

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		passwordMatches = utils.CheckPasswordHash(password, user.Password)
		wg.Done()
	}()

	go func() {
		token, err = us.tokenRepository.GenerateToken(user.ID)
		if err != nil {
			token = nil
		}
		wg.Done()
	}()

	wg.Wait()
	if !passwordMatches {
		return nil, nil, errors.ErrInvalidEmailOrPassword
	}
	if token == nil {
		return nil, nil, errors.NewError("failed to generate token: ", http.StatusInternalServerError)
	}
	err = us.tokenRepository.SaveToken(user.ID, token)
	if err != nil {
		log.Println(err)
		return nil, nil, errors.NewError("failed to save token: "+err.Error(), http.StatusInternalServerError)
	}
	return &User{
			ID:          user.ID,
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			Email:       user.Email,
			IsVerified:  user.IsVerified,
			Role:        user.Role,
			PhoneNumber: user.PhoneNumber,
			Addresses:   user.Addresses,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		}, &utils.TokenDetails{
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			AtExpires:    token.AtExpires,
			RtExpires:    token.RtExpires,
		}, nil
}

func (us *UserService) Profile(userId primitive.ObjectID) (*User, *errors.AppError) {
	user, err := us.userRepository.GetUser(bson.M{"_id": userId})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err := errors.ErrNotFound
			return nil, errors.NewError("user not found: "+err.Error(), err.StatusCode)
		}
		return nil, errors.ErrInternalServer
	}
	return user, nil
}

func (us *UserService) Logout(accessUuid string) *errors.AppError {
	err := us.tokenRepository.DeleteToken(accessUuid)
	if err != nil {
		return errors.NewError("logout unsuccessful: invalid access token: "+err.Error(), http.StatusNotFound)
	}
	return nil
}

func (us *UserService) SendForgotPasswordOTP(email string) *errors.AppError {
	user, err := us.userRepository.GetUser(bson.M{"email": email})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err := errors.ErrNotFound
			return errors.NewError("user not found: "+err.Error(), err.StatusCode)
		}
		return errors.ErrInternalServer
	}
	if !user.IsVerified {
		return errors.NewError("user is not verified", http.StatusForbidden)
	}
	otp, err := us.otpRepository.GenerateForgotPasswordOTP(email)
	if err != nil {
		return errors.NewError("failed to generate otp"+err.Error(), http.StatusInternalServerError)
	}
	if otp == "" {
		return errors.ErrInternalServer
	}
	err = us.emailRepository.SendResetPasswordEmail(email, user.FirstName, otp)
	if err != nil {
		return errors.NewError("failed to send user otp"+err.Error(), http.StatusInternalServerError)
	}
	return nil
}

func (us *UserService) ResetPassword(email, otp, password string) *errors.AppError {
	valid, err := us.otpRepository.VerifyForgotPasswordOTP(email, otp)
	if err != nil {
		us.otpRepository.DeleteOTP(email)
		return errors.NewError("failed to verify otp: invalid otp: "+err.Error(), errors.ErrInvalidOTP.StatusCode)
	}
	if !valid {
		us.otpRepository.DeleteOTP(email)
		return errors.ErrInvalidOTP
	}
	err = us.otpRepository.DeleteOTP(email)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err := errors.ErrNotFound
			return errors.NewError("invalid otp: "+err.Error(), err.StatusCode)
		}
		return errors.ErrInternalServer
	}
	hash, err := utils.HashPassword(password)
	if err != nil || hash == "" {
		return errors.ErrInternalServer
	}
	err = us.userRepository.UpdateUser(bson.M{"email": email}, bson.M{"$set": bson.M{"password": hash}})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err := errors.ErrNotFound
			return errors.NewError("user not found: "+err.Error(), err.StatusCode)
		}
		return errors.ErrInternalServer
	}
	return nil
}

func (us *UserService) RefreshToken(refreshToken string) (*utils.TokenDetails, *errors.AppError) {
	refreshTokenDetails, err := us.tokenRepository.IdentifyUser(refreshToken)
	if err != nil {
		return nil, errors.ErrInternalServer
	}
	userId := refreshTokenDetails.UserId
	refreshUuid := refreshTokenDetails.RefreshUuid
	err = us.tokenRepository.DeleteToken(refreshUuid)
	if err != nil {
		return nil, errors.NewError("failed to delete old refresh token: "+err.Error(), http.StatusNotFound)
	}
	token, err := us.tokenRepository.GenerateToken(userId)
	if err != nil {
		return nil, errors.ErrInternalServer
	}
	err = us.tokenRepository.SaveToken(userId, token)
	if err != nil {
		return nil, errors.ErrInternalServer
	}
	return &utils.TokenDetails{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		AtExpires:    token.AtExpires,
		RtExpires:    token.RtExpires,
	}, nil
}

func (us *UserService) GetUsers() ([]*User, *errors.AppError) {
	users, err := us.userRepository.GetUsers(bson.M{})
	if err != nil {
		return nil, errors.ErrInternalServer
	}
	return users, nil
}
