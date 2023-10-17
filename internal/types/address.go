package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Address struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	AddressLine1 string             `json:"address_line_1" bson:"address_line_1" binding:"required"`
	City         string             `json:"city" bson:"city" binding:"required"`
	AddressLine2 string             `json:"address_line_2" bson:"address_line_2"`
	PostalCode   string             `json:"postal_code" bson:"postal_code" binding:"required"`
	Country      string             `json:"country" bson:"country" binding:"required"`
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at" bson:"updated_at"`
}
