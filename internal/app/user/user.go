package user

import (
	"time"

	"github.com/ayo-ajayi/ecommerce/internal/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	FirstName   string             `json:"first_name" bson:"first_name"`
	LastName    string             `json:"last_name" bson:"last_name"`
	Email       string             `json:"email" bson:"email"`
	Password    string             `json:"-" bson:"password"`
	PhoneNumber string             `json:"phone_number" bson:"phone_number"`
	IsVerified  bool               `json:"is_verified" bson:"is_verified"`
	Addresses   []types.Address    `json:"addresses" bson:"addresses"`
	Role        Role               `json:"role" bson:"role"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

type Role string

const (
	Admin    Role = "admin"
	Vendor   Role = "vendor"
	Customer Role = "customer"
)

type ItemVendor struct {
	ID     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID primitive.ObjectID `json:"user_id" bson:"user_id,omitempty"`
	Name   string             `json:"name" bson:"name"`
	Email  string             `json:"email" bson:"email"`
}
