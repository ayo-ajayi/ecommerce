package category

import (
	"context"
	"mime/multipart"
	"net/http"

	"github.com/ayo-ajayi/ecommerce/internal/errors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CategoryController struct {
	categoryServices CategoryServices
	uploader         Uploader
}
type Uploader interface {
	UploadImage(ctx context.Context, files []*multipart.FileHeader, collection string) ([]string, *errors.AppError)
}

type CategoryServices interface {
	CreateCategory(category *Category) *errors.AppError
	GetCategoryByID(id string) (*Category, *errors.AppError)
	GetCategories() ([]*Category, *errors.AppError)
	UpdateCategory(category *Category) *errors.AppError
	DeleteCategory(id string, userid primitive.ObjectID) *errors.AppError
	GetCategoryBySlug(slug string) (*Category, *errors.AppError)
}

func NewCategoryController(categoryServices CategoryServices, uploader Uploader) *CategoryController {
	return &CategoryController{
		categoryServices: categoryServices,
		uploader:         uploader,
	}
}

func (cc *CategoryController) CreateCategory(c *gin.Context) {
	userid := c.MustGet("userId").(primitive.ObjectID)
	if userid.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid user"}})
		return
	}

	category, err := cc.createCategory(c)
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	category.CreatedBy = userid
	category.UpdatedBy = userid
	if err := cc.categoryServices.CreateCategory(category); err != nil {
		c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(201, gin.H{"message": "category created successfully", "data": gin.H{"name": category.Name}})
}

func (cc *CategoryController) createCategory(c *gin.Context) (*Category, *errors.AppError) {
	err := c.Request.ParseMultipartForm(10 << 20)
	if err != nil {
		return nil, errors.NewError("file too large"+err.Error(), 400)
	}
	req := struct {
		Name        string               `json:"name" binding:"required"`
		Description string               `json:"description" binding:"required"`
		ParentID    []primitive.ObjectID `json:"parent_id"`
		Images      []string             `json:"images"`
	}{}
	req.Name = c.PostForm("name")
	req.Description = c.PostForm("description")
	if req.Name == "" || req.Description == "" {
		return nil, errors.NewError("invalid name or description", 400)
	}
	parentIDs := c.PostFormArray("parent_id")
	for _, parentID := range parentIDs {
		objectID, err := primitive.ObjectIDFromHex(parentID)
		if err != nil {
			return nil, errors.NewError("invalid parent id", 400)
		}
		req.ParentID = append(req.ParentID, objectID)
	}
	files := c.Request.MultipartForm.File["images"]
	if imgURLs, err := cc.uploader.UploadImage(c.Request.Context(), files, "categories"); err != nil {
		return nil, errors.NewError("failed to upload images: "+err.Error(), 400)
	} else {
		req.Images = imgURLs
	}
	return &Category{
		Name:        req.Name,
		Description: req.Description,
		ParentID:    req.ParentID,
		Images:      req.Images,
	}, nil
}

func (cc *CategoryController) GetCategoryByID(c *gin.Context) {
	id := c.Param("id")
	category, err := cc.categoryServices.GetCategoryByID(id)
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(200, gin.H{"data": gin.H{"category": category}})
}

func (cc *CategoryController) GetCategoryBySlug(c *gin.Context) {
	slug := c.Param("slug")
	category, err := cc.categoryServices.GetCategoryBySlug(slug)
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(200, gin.H{"data": gin.H{"category": category}})
}

func (cc *CategoryController) UpdateCategory(c *gin.Context) {
	categoryIdStr := c.Param("id")
	categoryId, err := primitive.ObjectIDFromHex(categoryIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	userid := c.MustGet("userId").(primitive.ObjectID)
	if userid.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid user"}})
		return
	}
	var cErr *errors.AppError
	category, cErr := cc.createCategory(c)
	if cErr != nil {
		c.JSON(cErr.StatusCode, gin.H{"error": gin.H{"message": cErr.Error()}})
		return
	}
	category.ID = categoryId
	category.UpdatedBy = userid
	if err := cc.categoryServices.UpdateCategory(category); err != nil {
		c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(200, gin.H{"message": "category updated successfully", "data": gin.H{"category": category}})
}

func (cc *CategoryController) GetCategories(c *gin.Context) {
	categories, err := cc.categoryServices.GetCategories()
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(200, gin.H{"data": gin.H{"categories": categories}})

}

func (cc *CategoryController) DeleteCategory(c *gin.Context) {
	categoryId := c.Param("id")
	userid := c.MustGet("userId").(primitive.ObjectID)
	if userid.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid user"}})
		return
	}
	err := cc.categoryServices.DeleteCategory(categoryId, userid)
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(200, gin.H{"message": "category deleted successfully"})
}
