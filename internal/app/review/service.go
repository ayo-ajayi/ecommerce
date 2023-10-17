package review

import (
	"log"
	"time"

	"github.com/ayo-ajayi/ecommerce/internal/app/user"
	"github.com/ayo-ajayi/ecommerce/internal/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ReviewSevice struct {
	reviewRepository ReviewRepository
	userRepository   UserRepository
}
type UserRepository interface {
	GetUser(filter interface{}) (*user.User, error)
}

type ReviewRepository interface {
	CreateReview(review *Review) error
	UpdateReview(filter interface{}, update interface{}, opts ...*options.UpdateOptions) error
	GetReview(filter interface{}, opts ...*options.FindOneOptions) (*Review, error)
	GetReviews(filter interface{}, opts ...*options.FindOptions) ([]*Review, error)
	IsExists(filter interface{}, opts ...*options.FindOneOptions) (bool, error)
}

func NewReviewService(reviewRepository ReviewRepository, userRepository UserRepository) *ReviewSevice {
	return &ReviewSevice{
		reviewRepository: reviewRepository,
		userRepository:   userRepository,
	}
}

func (rs *ReviewSevice) PostReview(review *Review) *errors.AppError {
	//check if item to create review exists in user's order

	review.UpdatedAt = time.Now()
	if review.Star < 1 || review.Star > 5 {
		return errors.NewError("invalid star", 400)
	}

	if !review.Anonymous {
		user, err := rs.userRepository.GetUser(bson.M{"_id": review.AuthorID})
		if err != nil {
			return errors.ErrInternalServer
		}
		if user != nil {
			review.AuthorName = user.FirstName + " " + user.LastName
		}
	}

	revExists, err := rs.reviewRepository.IsExists(bson.M{"author_id": review.AuthorID, "item_id": review.ItemID})
	if err != nil {
		return errors.ErrInternalServer
	}
	if revExists {
		log.Println("review exists")
		rev, err := rs.reviewRepository.GetReview(bson.M{"author_id": review.AuthorID, "item_id": review.ItemID})
		if err != nil {
			return errors.ErrInternalServer
		}
		if err := rs.reviewRepository.UpdateReview(bson.M{"_id": rev.ID}, bson.M{"$set": review}); err != nil {
			return errors.ErrInternalServer
		}
		return nil
	}
	log.Println("review does not exist")
	if err := rs.reviewRepository.CreateReview(review); err != nil {
		return errors.ErrInternalServer
	}
	return nil
}

func (rs *ReviewSevice) GetReview(reviewId string) (*Review, *errors.AppError) {
	review_id, err := primitive.ObjectIDFromHex(reviewId)
	if err != nil {
		return nil, errors.ErrInvalidObjectID
	}
	review, err := rs.reviewRepository.GetReview(bson.M{"_id": review_id})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.ErrNotFound
		}
		return nil, errors.ErrInternalServer
	}
	return review, nil
}

func (rs *ReviewSevice) GetReviews(itemId string) ([]*Review, *errors.AppError) {
	itemid, err := primitive.ObjectIDFromHex(itemId)
	if err != nil {
		return nil, errors.NewError(errors.ErrInvalidObjectID.Error()+err.Error(), 400)
	}
	reviews, err := rs.reviewRepository.GetReviews(bson.M{"item_id": itemid})
	if err != nil {
		return nil, errors.ErrInternalServer
	}
	return reviews, nil
}
