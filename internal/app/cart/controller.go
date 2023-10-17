package cart

import (
	"net/http"

	"github.com/ayo-ajayi/ecommerce/internal/errors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CartController struct {
	cartServices CartServices
}

type CartServices interface {
	AddToCart(userid primitive.ObjectID, cartItem CartItem) *errors.AppError
	RemoveFromCart(userid primitive.ObjectID, cartItem CartItem) *errors.AppError
	GetCart(userid primitive.ObjectID) (*Cart, *errors.AppError)
}

func NewCartController(cartServices CartServices) *CartController {
	return &CartController{
		cartServices: cartServices,
	}
}

func (cc *CartController) UpdateCart(c *gin.Context) {
	userid := c.MustGet("userId").(primitive.ObjectID)
	if userid.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid user"}})
		return
	}
	req := struct {
		ItemID   string `json:"item_id" binding:"required"`
		Quantity int    `json:"quantity" binding:"required"`
	}{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	if req.Quantity == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid quantity"}})
		return
	}
	objectID, err := primitive.ObjectIDFromHex(req.ItemID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	cartItem := CartItem{
		ItemID:   objectID,
		Quantity: req.Quantity,
	}
	if req.Quantity < 0 {
		cartItem.Quantity = -cartItem.Quantity
		if err := cc.cartServices.RemoveFromCart(userid, cartItem); err != nil {
			c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "success removed item from cart"})
		return
	}
	if err := cc.cartServices.AddToCart(userid, cartItem); err != nil {
		c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success added new item to cart"})
}

func (cc *CartController) GetCart(c *gin.Context) {
	userid := c.MustGet("userId").(primitive.ObjectID)
	if userid.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid user"}})
		return
	}
	ct, err := cc.cartServices.GetCart(userid)
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	if ct == nil {
		c.JSON(http.StatusOK, gin.H{"message": "cart is empty"})
		return
	}
	c.JSON(http.StatusOK, ct)
}
