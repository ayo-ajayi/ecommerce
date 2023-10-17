package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/ayo-ajayi/ecommerce/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

type Middleware struct {
	accessTokenSecretKey string
	middlewareTokenRepo  MiddlewareTokenRepo
	middlewareUserRepo   MiddlewareUserRepo
}
type MiddlewareTokenRepo interface {
	FindToken(accessuuid string) (string, error)
	DeleteToken(accessUuid string) error
}

func NewMiddleware(accessTokenSecretKey string, middlewareTokenRepo MiddlewareTokenRepo, middlewareUserRepo MiddlewareUserRepo) *Middleware {
	return &Middleware{
		accessTokenSecretKey,
		middlewareTokenRepo,
		middlewareUserRepo,
	}
}

func (m *Middleware) Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c.Request)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"message": "access token required: unauthorized"},
			})
			return
		}
		jwtToken, err := utils.ValidateToken(token, m.accessTokenSecretKey)
		if err != nil {
			if errors.Is(err, jwt.ErrSignatureInvalid) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": gin.H{"message": "unauthorized: " + err.Error()},
				})
				return
			}
			if errors.Is(err, jwt.ErrTokenExpired) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": gin.H{"message": "unauthorized: " + err.Error()},
				})
				return
			}

			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{"message": "internal server error: " + err.Error()},
			})
			return
		}
		td, err := utils.ExtractTokenDetails(jwtToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": gin.H{"message": err.Error() + ": unauthorized"},
			})
			return
		}
		userId, err := m.middlewareTokenRepo.FindToken(td.AccessUuid)
		if err != nil {
			if err == redis.Nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": gin.H{"message": "access token is not valid: unauthorized"},
				})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{"message": "internal server error: " + err.Error()},
			})
			return
		}
		if userId != td.UserId.Hex() {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"message": "access token is not found: unauthorized"},
			})
			return
		}
		c.Set("userId", td.UserId)
		c.Set("accessUuid", td.AccessUuid)
		c.Next()
	}
}

func extractToken(r *http.Request) string {
	token := r.Header.Get("Authorization")
	ttoken := strings.Split(token, " ")
	if len(ttoken) != 2 {
		return ""
	}
	return ttoken[1]
}

func (m *Middleware) JsonMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.Next()
	}
}
