package item

import (
	"context"
	"mime/multipart"
	"sync"
	"time"

	"github.com/ayo-ajayi/ecommerce/internal/errors"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ItemService struct {
	itemRepository     ItemRepository
	categoryRepository CategoryRepository
	uploader           Uploader
}
type CategoryRepository interface {
	IsExists(filter interface{}, opts ...*options.FindOneOptions) (bool, error)
}
type ItemRepository interface {
	IsExists(filter interface{}, opts ...*options.FindOneOptions) (bool, error)
	CreateItem(item *Item) error
	UpdateItem(filter interface{}, update interface{}, opts ...*options.UpdateOptions) error
	GetItem(filter interface{}, opts ...*options.FindOneOptions) (*Item, error)
	GetItems(filter interface{}, opts ...*options.FindOptions) ([]*Item, error)
	DeleteItem(filter interface{}, opts ...*options.DeleteOptions) error
}
type Uploader interface {
	UploadImage(ctx context.Context, files []*multipart.FileHeader, collection string) ([]string, *errors.AppError)
	DeleteImageBySecureURL(ctx context.Context, secureUrl string) *errors.AppError
}

func NewItemService(itemRepository ItemRepository, categoryRepository CategoryRepository, uploader Uploader) *ItemService {
	return &ItemService{itemRepository, categoryRepository, uploader}
}

func (is *ItemService) CreateItem(item *Item) *errors.AppError {
	item.Slug = slug.Make(item.Name)
	itemExists, err := is.itemRepository.IsExists(bson.M{
		"slug": item.Slug,
	})
	if err != nil {
		return errors.ErrInternalServer
	}
	if itemExists {
		item.Slug = item.Slug + "-" + uuid.New().String()
	}
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()
	if err := is.checkCategoryID(item.CategoryID); err != nil {
		return err
	}
	err = is.itemRepository.CreateItem(item)
	if err != nil {
		return errors.ErrInternalServer
	}
	return nil
}

func (is *ItemService) DeleteItem(itemId string, userId primitive.ObjectID) *errors.AppError {
	item_id, err := primitive.ObjectIDFromHex(itemId)
	if err != nil {
		return errors.ErrInvalidObjectID
	}
	vendorId := userId
	item, err := is.itemRepository.GetItem(bson.M{"_id": item_id, "vendor_id": vendorId})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.NewError("item not found: "+err.Error(), 404)
		}
		return errors.ErrInternalServer
	}
	if err := is.itemRepository.DeleteItem(bson.M{"_id": item_id, "vendor_id": vendorId}); err != nil {
		return errors.NewError("internal error: "+err.Error(), 500)
	}
	oldImages := item.Images
	if len(oldImages) > 0 {
		for _, v := range oldImages {
			is.uploader.DeleteImageBySecureURL(context.Background(), v)
		}
	}
	return nil
}

func (is *ItemService) UploadImage(ctx context.Context, files []*multipart.FileHeader, collection string) ([]string, *errors.AppError) {
	return is.uploader.UploadImage(ctx, files, collection)
}

func (is *ItemService) GetItems() ([]*Item, *errors.AppError) {
	items, err := is.itemRepository.GetItems(bson.M{})
	if err != nil {
		return nil, errors.ErrInternalServer
	}
	return items, nil
}

func (is *ItemService) GetItemBySlug(slug string) (*Item, *errors.AppError) {
	item, err := is.itemRepository.GetItem(bson.M{"slug": slug})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err := errors.ErrNotFound
			return nil, errors.NewError("item not found: "+err.Error(), err.StatusCode)
		}
		return nil, errors.ErrInternalServer
	}
	return item, nil
}

func (is *ItemService) GetVendorItems(vendorid primitive.ObjectID) ([]*Item, *errors.AppError) {

	items, err := is.itemRepository.GetItems(bson.M{"vendor_id": vendorid})
	if err != nil {
		return nil, errors.ErrInternalServer
	}
	return items, nil
}

func (is *ItemService) GetItemByID(itemId string) (*Item, *errors.AppError) {
	item_id, err := primitive.ObjectIDFromHex(itemId)
	if err != nil {
		return nil, errors.ErrInvalidObjectID
	}
	item, err := is.itemRepository.GetItem(bson.M{"_id": item_id})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err := errors.ErrNotFound
			return nil, errors.NewError("item not found: "+err.Error(), err.StatusCode)
		}
		return nil, errors.ErrInternalServer
	}
	return item, nil
}

func (is *ItemService) UpdateItem(item *Item) *errors.AppError {

	oldItem, err := is.itemRepository.GetItem(bson.M{"_id": item.ID})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err := errors.ErrNotFound
			return errors.NewError("item not found: "+err.Error(), err.StatusCode)
		}
		return errors.ErrInternalServer
	}
	item.UpdatedAt = time.Now()
	item.CreatedAt = oldItem.CreatedAt
	item.Slug = slug.Make(item.Name)

	if err := is.checkCategoryID(item.CategoryID); err != nil {
		return err
	}
	if err := is.itemRepository.UpdateItem(bson.M{"_id": item.ID}, bson.M{"$set": item}); err != nil {
		return errors.ErrInternalServer
	}
	oldImages := oldItem.Images
	if len(oldImages) > 0 {
		for _, v := range oldImages {
			is.uploader.DeleteImageBySecureURL(context.Background(), v)
		}
	}
	return nil
}

func (is *ItemService) checkCategoryID(categoryID []primitive.ObjectID) *errors.AppError {
	var wg sync.WaitGroup
	itemExistChan := make(chan *errors.AppError, len(categoryID))
	var allErrors []*errors.AppError

	for _, v := range categoryID {
		wg.Add(1)
		go func(categoryID primitive.ObjectID) {
			exists, err := is.categoryRepository.IsExists(bson.M{"_id": categoryID})
			if err != nil {
				itemExistChan <- errors.NewError("internal error: "+err.Error(), 500)
			}
			if !exists {
				err := errors.ErrNotFound
				itemExistChan <- errors.NewError("category not found: "+err.Error(), err.StatusCode)
			}
			wg.Done()
		}(v)
	}
	go func() {
		wg.Wait()
		close(itemExistChan)
	}()
	for err := range itemExistChan {
		if err != nil {
			allErrors = append(allErrors, err)
		}
	}
	if len(allErrors) > 0 {
		return allErrors[0]
	}
	return nil
}
