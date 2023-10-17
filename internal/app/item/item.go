package item

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Item struct {
	ID          primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	Name        string               `json:"name" bson:"name"`
	Slug        string               `json:"slug" bson:"slug"`
	Description string               `json:"description" bson:"description"`
	CategoryID  []primitive.ObjectID `json:"category_id" bson:"category_id,omitempty"`
	Price       float64              `json:"price" bson:"price"`
	Discount    float64              `json:"discount" bson:"discount"`
	Quantity    int                  `json:"quantity" bson:"quantity"`
	Images      []string             `json:"images" bson:"images"`
	VendorID    primitive.ObjectID   `json:"vendor_id" bson:"vendor_id,omitempty"`
	CreatedAt   time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at" bson:"updated_at"`
}
