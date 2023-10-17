package middleware

import (
	"github.com/ayo-ajayi/ecommerce/internal/app/user"
	"github.com/ayo-ajayi/ecommerce/internal/errors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MiddlewareUserRepo interface {
	GetUser(filter interface{}) (*user.User, error)
}

func (m *Middleware) Authorization(roles []user.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId, ok := c.MustGet("userId").(primitive.ObjectID)
		if !ok {
			err := errors.ErrInvalidObjectID
			c.AbortWithStatusJSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error() + ": you are not authorized to acess this resource"}})
			return
		}
		user, err := m.middlewareUserRepo.GetUser(bson.M{
			"_id": userId,
		})
		if err != nil {
			err := errors.ErrForbidden
			c.AbortWithStatusJSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error() + ": you are not authorized to acess this resource"}})
			return
		}
		allowed := false
		for _, role := range roles {
			if role == user.Role {
				allowed = true
				break
			}
		}
		if !allowed {
			err := errors.ErrForbidden
			c.AbortWithStatusJSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error() + ": you are not authorized to acess this resource"}})
			return
		}
		c.Next()
	}
}
