package search

import (
	"sync"

	"github.com/ayo-ajayi/ecommerce/internal/app/category"
	"github.com/ayo-ajayi/ecommerce/internal/app/item"
	"github.com/ayo-ajayi/ecommerce/internal/database"
	"github.com/ayo-ajayi/ecommerce/internal/errors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SearchController struct {
	itemRepo     ItemRepository
	categoryRepo CategoryRepository
}

func NewSearchController(itemRepo ItemRepository, categoryRepo CategoryRepository) *SearchController {
	return &SearchController{
		itemRepo:     itemRepo,
		categoryRepo: categoryRepo,
	}
}

type ItemRepository interface {
	GetItems(filter interface{}, opts ...*options.FindOptions) ([]*item.Item, error)
}
type CategoryRepository interface {
	GetCategories(filter interface{}, opts ...*options.FindOptions) ([]*category.Category, error)
}

func InitSearchIndex(coll ...*mongo.Collection) *errors.AppError {
	ctx, cancel := database.DBReqContext(5)
	defer cancel()
	var errSlice []error
	for _, collection := range coll {
		_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys: bson.D{
				{Key: "name", Value: "text"}, {Key: "description", Value: "text"},
			},
			Options: options.Index().SetName("text_index"),
		})
		errSlice = append(errSlice, err)
	}
	for _, err := range errSlice {
		if err != nil {
			return errors.NewError("failed to create index: "+err.Error(), 500)
		}
	}
	return nil
}

func (sc *SearchController) Search(ctx *gin.Context) {
	query := ctx.Query("q")
	if query == "" {
		ctx.JSON(400, gin.H{"error": "query not found"})
		return
	}

	var itemErr, categoryErr error
	var items []*item.Item
	filter := bson.M{
		"$or": []bson.M{
			{"name": primitive.Regex{
				Pattern: query,
				Options: "i",
			}},
			{"description": primitive.Regex{
				Pattern: query,
				Options: "i",
			},
			},
		}}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		items, itemErr = sc.itemRepo.GetItems(filter)

	}()

	var categories []*category.Category
	wg.Add(1)
	go func() {
		defer wg.Done()
		categories, categoryErr = sc.categoryRepo.GetCategories(filter)
	}()
	wg.Wait()

	if itemErr != nil {
		ctx.JSON(500, gin.H{"error": itemErr.Error()})
		return
	}
	if categoryErr != nil {
		ctx.JSON(500, gin.H{"error": categoryErr.Error()})
		return
	}

	ctx.JSON(200, gin.H{"message": "searching for " + query, "data": gin.H{"items": items, "categories": categories}})
}
