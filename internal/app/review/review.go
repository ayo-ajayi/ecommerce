package review

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Review struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ItemID     primitive.ObjectID `json:"item_id" bson:"item_id"`
	AuthorID   primitive.ObjectID `json:"author_id" bson:"author_id"`
	AuthorName string             `json:"author_name" bson:"author_name"`
	Star       int                `json:"star" bson:"star"`
	Content    string             `json:"content" bson:"content"`
	UpdatedAt  time.Time          `json:"updated_at" bson:"updated_at"`
	Anonymous  bool               `json:"anonymous" bson:"anonymous"`
}
