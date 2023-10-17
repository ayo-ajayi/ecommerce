package cart

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Cart struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID     primitive.ObjectID `json:"user_id" bson:"user_id,omitempty"`
	CartItems  []CartItem         `json:"cart_items" bson:"cart_items"`
	TotalPrice float64            `json:"total_price" bson:"total_price"`
	UpdatedAt  time.Time          `json:"updated_at" bson:"updated_at"`
}

type CartItem struct {
	ItemID     primitive.ObjectID `json:"item_id" bson:"item_id,omitempty"`
	Quantity   int                `json:"quantity" bson:"quantity"`
	Price      float64            `json:"price" bson:"price"`
	TotalPrice float64            `json:"total_price" bson:"total_price"`
}
