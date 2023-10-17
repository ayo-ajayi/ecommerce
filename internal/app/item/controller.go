package item

import (
	"context"
	"mime/multipart"
	"net/http"

	"strconv"

	"github.com/ayo-ajayi/ecommerce/internal/errors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ItemController struct {
	itemServices ItemServices
}

type ItemServices interface {
	CreateItem(item *Item) *errors.AppError
	DeleteItem(itemId string, userId primitive.ObjectID) *errors.AppError
	GetItems() ([]*Item, *errors.AppError)
	UpdateItem(item *Item) *errors.AppError
	GetItemByID(itemId string) (*Item, *errors.AppError)
	GetItemBySlug(slug string) (*Item, *errors.AppError)
	GetVendorItems(vendorId primitive.ObjectID) ([]*Item, *errors.AppError)
	UploadImage(ctx context.Context, files []*multipart.FileHeader, collection string) ([]string, *errors.AppError)
}

func NewItemController(itemServices ItemServices) *ItemController {
	return &ItemController{
		itemServices: itemServices,
	}
}
func (ic *ItemController) CreateItem(c *gin.Context) {
	userid := c.MustGet("userId").(primitive.ObjectID)
	if userid.IsZero() || userid.Hex() == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid user"}})
		return
	}
	item, err := ic.createItem(c)
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	item.VendorID = userid
	if err := ic.itemServices.CreateItem(item); err != nil {
		c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(200, gin.H{"message": "item created successfully", "data": gin.H{"name": item.Name}})
}

func (ic *ItemController) createItem(c *gin.Context) (*Item, *errors.AppError) {

	err := c.Request.ParseMultipartForm(32 << 20)
	if err != nil {
		return nil, errors.NewError("file too large"+err.Error(), 400)
	}
	req := struct {
		Name        string               `json:"name" binding:"required"`
		Description string               `json:"description" binding:"required"`
		CategoryID  []primitive.ObjectID `json:"category_id"`
		Price       float64              `json:"price" binding:"required"`
		Quantity    int                  `json:"quantity" binding:"required"`
		Discount    float64              `json:"discount"`
		Images      []string             `json:"images"`
	}{}

	req.Name = c.PostForm("name")
	req.Description = c.PostForm("description")
	if req.Name == "" || req.Description == "" {
		return nil, errors.NewError("invalid name or description", 400)
	}
	priceStr := c.PostForm("price")
	if priceStr == "" {
		return nil, errors.NewError("invalid price", 400)
	}
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return nil, errors.NewError("invalid price", 400)
	}
	req.Price = price

	quantityStr := c.PostForm("quantity")
	if quantityStr == "" {
		return nil, errors.NewError("invalid quantity", 400)
	}
	quantity, err := strconv.Atoi(quantityStr)
	if err != nil {
		return nil, errors.NewError("invalid quantity", 400)
	}
	req.Quantity = quantity

	discountStr := c.PostForm("discount")
	if discountStr != "" {
		discount, err := strconv.ParseFloat(discountStr, 64)
		if err != nil {
			return nil, errors.NewError("invalid discount", 400)
		}
		req.Discount = discount
	}

	categoryIDs := c.PostFormArray("category_id")
	for _, v := range categoryIDs {
		categoryID, err := primitive.ObjectIDFromHex(v)
		if err != nil {
			return nil, errors.NewError("invalid category id", 400)
		}
		req.CategoryID = append(req.CategoryID, categoryID)
	}

	files := c.Request.MultipartForm.File["images"]

	if imgURLs, err := ic.itemServices.UploadImage(c.Request.Context(), files, "items"); err != nil {
		return nil, errors.NewError("failed to upload images: "+err.Error(), 400)
	} else {
		req.Images = imgURLs
	}
	return &Item{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Quantity:    req.Quantity,
		Discount:    req.Discount,
		CategoryID:  req.CategoryID,
		Images:      req.Images,
	}, nil
}

func (ic *ItemController) DeleteItem(c *gin.Context) {
	id := c.Param("id")
	userid := c.MustGet("userId").(primitive.ObjectID)
	if userid.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid user"}})
		return
	}
	if err := ic.itemServices.DeleteItem(id, userid); err != nil {
		c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(200, gin.H{"message": "item deleted successfully"})
}

func (ic *ItemController) GetItems(c *gin.Context) {
	items, err := ic.itemServices.GetItems()
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(200, gin.H{"items": items})
}

func (ic *ItemController) UpdateItem(c *gin.Context) {
	id := c.Param("id")
	itemID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid item id"}})
		return
	}
	vendorId := c.MustGet("userId").(primitive.ObjectID)
	if vendorId.IsZero() || vendorId.Hex() == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid user"}})
		return
	}
	var cErr *errors.AppError
	item, cErr := ic.createItem(c)
	if cErr != nil {
		c.JSON(cErr.StatusCode, gin.H{"error": gin.H{"message": cErr.Error()}})
		return
	}
	item.ID = itemID

	if err := ic.itemServices.UpdateItem(item); err != nil {
		c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}

	c.JSON(200, gin.H{"message": "item updated successfully", "data": gin.H{"name": item.Name}})
}

func (ic *ItemController) GetItemByID(c *gin.Context) {
	id := c.Param("id")
	item, err := ic.itemServices.GetItemByID(id)
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(200, gin.H{"data": gin.H{"item": item}})
}

func (ic *ItemController) GetItemBySlug(c *gin.Context) {
	slug := c.Param("slug")
	item, err := ic.itemServices.GetItemBySlug(slug)
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(200, gin.H{"data": gin.H{"item": item}})
}

func (ic *ItemController) GetVendorItems(c *gin.Context) {
	userid := c.MustGet("userId").(primitive.ObjectID)
	if userid.IsZero() || userid.Hex() == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid user"}})
		return
	}
	items, err := ic.itemServices.GetVendorItems(userid)
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(200, gin.H{"data": gin.H{"items": items}})
}
