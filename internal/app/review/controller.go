package review

import (
	"net/http"

	"github.com/ayo-ajayi/ecommerce/internal/errors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReviewController struct {
	reviewServices ReviewServices
}

type ReviewServices interface {
	PostReview(review *Review) *errors.AppError
	GetReview(reviewId string) (*Review, *errors.AppError)
	GetReviews(itemId string) ([]*Review, *errors.AppError)
}

func NewReviewController(reviewServices ReviewServices) *ReviewController {
	return &ReviewController{
		reviewServices: reviewServices,
	}
}

func (rc *ReviewController) PostReview(c *gin.Context) {
	userid := c.MustGet("userId").(primitive.ObjectID)
	if userid.IsZero() || userid.Hex() == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid user"}})
		return
	}
	req := struct {
		ItemID    primitive.ObjectID `json:"item_id" binding:"required"`
		Star      int                `json:"star" binding:"required"`
		Content   string             `json:"content" binding:"required"`
		Anonymous bool               `json:"anonymous"`
	}{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	review := Review{
		ItemID:    req.ItemID,
		AuthorID:  userid,
		Star:      req.Star,
		Content:   req.Content,
		Anonymous: req.Anonymous,
	}
	if err := rc.reviewServices.PostReview(&review); err != nil {
		c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

func (rc *ReviewController) GetReview(c *gin.Context) {
	id := c.Param("id")
	review, err := rc.reviewServices.GetReview(id)
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"review": review}})
}

func (rc *ReviewController) GetReviews(c *gin.Context) {
	req := struct {
		ItemID string `json:"item_id" binding:"required"`
	}{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	reviews, err := rc.reviewServices.GetReviews(req.ItemID)
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"reviews": reviews}})
}
