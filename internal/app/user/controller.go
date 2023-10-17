package user

import (
	"net/http"

	"github.com/ayo-ajayi/ecommerce/internal/errors"
	"github.com/ayo-ajayi/ecommerce/internal/types"
	"github.com/ayo-ajayi/ecommerce/internal/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserController struct {
	userServices UserServices
}

func NewUserController(userServices UserServices) *UserController {
	return &UserController{
		userServices: userServices,
	}
}

type UserServices interface {
	SignUpAndSendVerificationEmail(user *User) *errors.AppError
	VerifyUser(email, otp string) *errors.AppError
	Login(email, password string) (*User, *utils.TokenDetails, *errors.AppError)
	Profile(userId primitive.ObjectID) (*User, *errors.AppError)
	Logout(accessUuid string) *errors.AppError
	SendForgotPasswordOTP(email string) *errors.AppError
	ResetPassword(email, otp, password string) *errors.AppError
	RefreshToken(refreshToken string) (*utils.TokenDetails, *errors.AppError)
	ResendEmailVerificationOTP(email string) *errors.AppError
	AddAddress(userid primitive.ObjectID, address types.Address) *errors.AppError
	RemoveAddress(userid, addressid primitive.ObjectID) *errors.AppError
	GetAddresses(userid primitive.ObjectID) ([]types.Address, *errors.AppError)
	GetAddress(userid, addressid primitive.ObjectID) (*types.Address, *errors.AppError)
	UpdateAddress(userid, addressid primitive.ObjectID, address types.Address) *errors.AppError
	GetUsers() ([]*User, *errors.AppError)
}

func (uc *UserController) SignUp(c *gin.Context) {
	req := struct {
		Email     string `json:"email" binding:"required"`
		Password  string `json:"password" binding:"required"`
		FirstName string `json:"first_name" binding:"required"`
		LastName  string `json:"last_name" binding:"required"`
	}{}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	err := uc.userServices.SignUpAndSendVerificationEmail(&User{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user created successfully"})
}

func (uc *UserController) ResendEmailVerificationOTP(c *gin.Context) {
	req := struct {
		Email string `json:"email" binding:"required"`
	}{}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	err := uc.userServices.ResendEmailVerificationOTP(req.Email)
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "otp sent successfully"})
}

func (uc *UserController) VerifyUser(c *gin.Context) {
	req := struct {
		Email string `json:"email" binding:"required"`
		OTP   string `json:"otp" binding:"required"`
	}{}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	err := uc.userServices.VerifyUser(req.Email, req.OTP)
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user verified successfully"})
}

func (uc *UserController) Login(c *gin.Context) {
	req := struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}{}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	user, tokenDetails, err := uc.userServices.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user, "token_details": tokenDetails})
}

func (uc *UserController) Profile(c *gin.Context) {
	userid := c.MustGet("userId").(primitive.ObjectID)
	if userid.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid user id"}})
		return
	}
	user, err := uc.userServices.Profile(userid)
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (uc *UserController) Logout(c *gin.Context) {
	accessUuid := c.GetString("accessUuid")
	err := uc.userServices.Logout(accessUuid)
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user logged out successfully"})
}

func (uc *UserController) SendForgotPasswordOTP(c *gin.Context) {

	req := struct {
		Email string `json:"email" binding:"required"`
	}{}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	err := uc.userServices.SendForgotPasswordOTP(req.Email)
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "otp sent successfully"})
}

// TODO:: verify the otp before redirecting user to where they can reset their password
func (uc *UserController) ResetPassword(c *gin.Context) {
	req := struct {
		Email    string `json:"email" binding:"required"`
		OTP      string `json:"otp" binding:"required"`
		Password string `json:"password" binding:"required"`
	}{}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	err := uc.userServices.ResetPassword(req.Email, req.OTP, req.Password)
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "password reset successfully"})
}

func (uc *UserController) RefreshToken(c *gin.Context) {
	req := struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}{}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	tokenDetails, err := uc.userServices.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "token refreshed successfully", "data": gin.H{"token_details": tokenDetails}})
}

func (uc *UserController) AddAddress(c *gin.Context) {
	userid := c.MustGet("userId").(primitive.ObjectID)
	if userid.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid user id"}})
		return
	}
	req := struct {
		types.Address
	}{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	err := uc.userServices.AddAddress(userid, req.Address)
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "address added successfully"})
}

func (uc *UserController) RemoveAddress(c *gin.Context) {
	id := c.Param("id")
	addressid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	userid := c.MustGet("userId").(primitive.ObjectID)
	if userid.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid user id"}})
		return
	}

	if err := uc.userServices.RemoveAddress(userid, addressid); err != nil {
		c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "address removed successfully"})
}

func (uc *UserController) GetAddresses(c *gin.Context) {
	userid := c.MustGet("userId").(primitive.ObjectID)
	if userid.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid user id"}})
		return
	}
	addresses, err := uc.userServices.GetAddresses(userid)
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, addresses)
}

func (uc *UserController) GetAddress(c *gin.Context) {
	id := c.Param("id")
	addressid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	userid := c.MustGet("userId").(primitive.ObjectID)
	if userid.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid user id"}})
		return
	}

	address, getErr := uc.userServices.GetAddress(userid, addressid)
	if getErr != nil {
		c.JSON(getErr.StatusCode, gin.H{"error": gin.H{"message": getErr.Error()}})
		return
	}
	c.JSON(http.StatusOK, address)
}

func (uc *UserController) UpdateAddress(c *gin.Context) {
	id := c.Param("id")
	addressid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	userid := c.MustGet("userId").(primitive.ObjectID)
	if userid.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid user id"}})
		return
	}
	req := types.Address{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	if err := uc.userServices.UpdateAddress(userid, addressid, req); err != nil {
		c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "address updated successfully"})
}

func (uc *UserController) GetUsers(c *gin.Context) {
	users, err := uc.userServices.GetUsers()
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, users)
}
