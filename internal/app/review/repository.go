package review

import (
	"github.com/ayo-ajayi/ecommerce/internal/database"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ReviewRepo struct {
	Collection *mongo.Collection
}

func NewReviewRepo(collection *mongo.Collection) *ReviewRepo {
	return &ReviewRepo{
		Collection: collection,
	}
}

func (rr *ReviewRepo) IsExists(filter interface{}, opts ...*options.FindOneOptions) (bool, error) {
	ctx, cancel := database.DBReqContext(5)
	defer cancel()
	err := rr.Collection.FindOne(ctx, filter, opts...).Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (rr *ReviewRepo) CreateReview(review *Review) error {
	ctx, cancel := database.DBReqContext(5)
	defer cancel()
	_, err := rr.Collection.InsertOne(ctx, review)
	return err
}

func (rr *ReviewRepo) UpdateReview(filter interface{}, update interface{}, opts ...*options.UpdateOptions) error {
	ctx, cancel := database.DBReqContext(5)
	defer cancel()
	_, err := rr.Collection.UpdateOne(ctx, filter, update, opts...)
	return err
}

func (rr *ReviewRepo) GetReview(filter interface{}, opts ...*options.FindOneOptions) (*Review, error) {
	ctx, cancel := database.DBReqContext(5)
	defer cancel()
	review := &Review{}
	err := rr.Collection.FindOne(ctx, filter, opts...).Decode(review)
	return review, err
}
func (rr *ReviewRepo) GetReviews(filter interface{}, opts ...*options.FindOptions) ([]*Review, error) {
	ctx, cancel := database.DBReqContext(5)
	defer cancel()
	var reviews []*Review
	cursor, err := rr.Collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &reviews); err != nil {
		return nil, err
	}
	return reviews, nil
}
