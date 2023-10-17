package utils

import (
	"errors"

	"github.com/ayo-ajayi/ecommerce/internal/database"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"time"
)

type OTP struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Email     string             `json:"email" bson:"email"`
	URL       string             `json:"url" bson:"url"`
	Secret    string             `json:"secret" bson:"secret"`
	ExpiresAt time.Time          `json:"expires_at" bson:"expires_at"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}

type OTPManager struct {
	collection                      *mongo.Collection
	Issuer                          string
	SignUpOtpValidityInSecs         uint
	ForgotPasswordOtpValidityInSecs uint
}

func NewOTPManager(collection *mongo.Collection, issuer string, signUpOtpValidityInSecs, forgotPasswordOtpvalidityInSecs uint) *OTPManager {
	ou := &OTPManager{
		collection:                      collection,
		Issuer:                          issuer,
		SignUpOtpValidityInSecs:         signUpOtpValidityInSecs,
		ForgotPasswordOtpValidityInSecs: forgotPasswordOtpvalidityInSecs,
	}

	return ou
}

func InitOtpExpiryIndex(collection *mongo.Collection) error {
	indexModel := mongo.IndexModel{
		Keys: bson.M{"expires_at": 1}, Options: options.Index().SetExpireAfterSeconds(0),
	}
	ctx, cancel := database.DBReqContext(5)
	defer cancel()
	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return errors.New("Error creating TTL index for otp collection:" + err.Error())
	}
	return nil
}

func (ou *OTPManager) GenerateSignUpOTP(email string) (string, error) {
	return ou.generateOTP(email, ou.SignUpOtpValidityInSecs)
}
func (ou *OTPManager) GenerateForgotPasswordOTP(email string) (string, error) {
	return ou.generateOTP(email, ou.ForgotPasswordOtpValidityInSecs)
}

func (ou *OTPManager) VerifySignUpOTP(email, otpStr string) (bool, error) {
	return ou.verifyOTP(email, otpStr, ou.SignUpOtpValidityInSecs)
}

func (ou *OTPManager) VerifyForgotPasswordOTP(email, otpStr string) (bool, error) {
	return ou.verifyOTP(email, otpStr, ou.ForgotPasswordOtpValidityInSecs)
}

func (ou *OTPManager) DeleteOTP(email string) error {
	ctx, cancel := database.DBReqContext(5)
	defer cancel()
	_, err := ou.collection.DeleteMany(ctx, bson.M{"email": email})
	return err
}

func (ou *OTPManager) generateOTP(email string, otpValidityInSecs uint) (string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      ou.Issuer,
		AccountName: email,
		Period:      otpValidityInSecs,
	})
	if err != nil {
		return "", err
	}
	err = ou.saveOTP(&OTP{
		Email:     email,
		URL:       key.URL(),
		Secret:    key.Secret(),
		ExpiresAt: time.Now().Add(time.Duration(otpValidityInSecs) * time.Second),
		CreatedAt: time.Now(),
	})
	if err != nil {
		return "", errors.New("unable to save otp: " + err.Error())
	}
	otp, err := getOTPFromSecret(key.Secret(), otpValidityInSecs, time.Now())
	if err != nil {
		return "", errors.New("unable to generate otp properly: " + err.Error())
	}
	if otp == "" {
		return "", errors.New("unable to generate otp properly: " + err.Error())
	}
	return otp, nil
}

func getOTPFromSecret(secret string, otpValidityInSecs uint, t time.Time) (string, error) {
	return totp.GenerateCodeCustom(secret, t, totp.ValidateOpts{
		Period:    otpValidityInSecs,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
}

func (ou *OTPManager) saveOTP(otp *OTP) error {
	ctx, cancel := database.DBReqContext(5)
	defer cancel()

	_, err := ou.collection.InsertOne(ctx, otp)
	return err
}

func (ou *OTPManager) getLatestOTP(email string) (*OTP, error) {
	ctx, cancel := database.DBReqContext(5)
	defer cancel()
	var otpObj OTP
	findOptions := options.FindOneOptions{
		Sort: bson.D{{Key: "created_at", Value: -1}},
	}
	err := ou.collection.FindOne(ctx, bson.M{"email": email}, &findOptions).Decode(&otpObj) //get the latest otp
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("invalid otp from: " + email)
		}
		return nil, err
	}
	return &otpObj, nil
}

func (ou *OTPManager) verifyOTP(email, otpStr string, otpValidityInSecs uint) (bool, error) {
	otpObj, err := ou.getLatestOTP(email)
	if err != nil {
		return false, err
	}
	valid, err := totp.ValidateCustom(otpStr, otpObj.Secret, time.Now().UTC(), totp.ValidateOpts{
		Period:    otpValidityInSecs,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	return valid, err
}
