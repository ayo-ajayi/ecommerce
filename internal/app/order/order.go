package order

import (
	"github.com/ayo-ajayi/ecommerce/internal/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Order struct {
	ID              primitive.ObjectID `json:"id" bson:"_id"`
	UserID          string             `json:"user_id" bson:"user_id"`
	OrderItems      []OrderItem        `json:"order_items" bson:"order_items"`
	TotalPrice      float64            `json:"total_price" bson:"total_price"`
	OrderDate       string             `json:"order_date" bson:"order_date"`
	ShippingAddress types.Address      `json:"shipping_address" bson:"shipping_address"`
	PaymentMethod   PaymentMethod      `json:"payment_method" bson:"payment_method"`
	OrderStatus     OrderStatus        `json:"order_status" bson:"order_status"`
	CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at" bson:"updated_at"`
}

type OrderItem struct {
	ItemID   string  `json:"item_id" bson:"item_id"`
	Quantity int     `json:"quantity" bson:"quantity"`
	Price    float64 `json:"price" bson:"price"`
}

type PaymentMethod struct {
	ID             string        `json:"id" bson:"id"`
	CardNumber     string        `json:"card_number" bson:"card_number"`
	CardHolder     string        `json:"card_holder" bson:"card_holder"`
	ExpirationDate string        `json:"expiration_date" bson:"expiration_date"`
	CVV            string        `json:"cvv" bson:"cvv"`
	BillingAddress types.Address `json:"billing_address" bson:"billing_address"`
}
type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "pending"
	PaymentStatusPaid     PaymentStatus = "paid"
	PaymentStatusFailed   PaymentStatus = "failed"
	PaymentStatusRefunded PaymentStatus = "refunded"
)

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCancelled  OrderStatus = "cancelled"
)
