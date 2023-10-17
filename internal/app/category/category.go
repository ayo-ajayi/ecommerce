package category

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Category struct {
	ID          primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	Name        string               `json:"name" bson:"name"`
	Slug        string               `json:"slug" bson:"slug"`
	Description string               `json:"description" bson:"description"`
	Images      []string             `json:"images" bson:"images"`
	ParentID    []primitive.ObjectID `json:"parent_id" bson:"parent_id,omitempty"`
	CreatedBy   primitive.ObjectID   `json:"created_by" bson:"created_by,omitempty"`
	UpdatedBy   primitive.ObjectID   `json:"updated_by" bson:"updated_by,omitempty"`
	CreatedAt   time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at" bson:"updated_at"`
}
