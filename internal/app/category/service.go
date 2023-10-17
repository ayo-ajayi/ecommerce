package category

import (
	"sync"
	"time"

	"github.com/ayo-ajayi/ecommerce/internal/errors"
	"github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CategoryService struct {
	categoryRepo CategoryRepository
}

func NewCategoryService(categoryRepo CategoryRepository) *CategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,
	}
}

type CategoryRepository interface {
	IsExists(filter interface{}, opts ...*options.FindOneOptions) (bool, error)
	CreateCategory(category *Category) error
	UpdateCategory(filter interface{}, update interface{}, opts ...*options.UpdateOptions) error
	GetCategory(filter interface{}, opts ...*options.FindOneOptions) (*Category, error)
	GetCategories(filter interface{}, opts ...*options.FindOptions) ([]*Category, error)
	DeleteCategory(filter interface{}, opts ...*options.DeleteOptions) error
}

func (cs *CategoryService) CreateCategory(category *Category) *errors.AppError {
	category.Slug = slug.Make(category.Name)
	categoryExists, err := cs.categoryRepo.IsExists(bson.M{
		"slug": category.Slug,
		"name": category.Name,
	})
	if err != nil {
		return errors.NewError("internal error: "+err.Error(), 500)
	}
	if categoryExists {
		return errors.ErrCategoryAlreadyExists
	}
	category.CreatedAt = time.Now()
	category.UpdatedAt = time.Now()

	if err := cs.checkParentID(category.ParentID); err != nil {
		return err
	}
	if err := cs.categoryRepo.CreateCategory(category); err != nil {
		return errors.NewError("internal error: "+err.Error(), 500)
	}
	return nil
}

func (cs *CategoryService) GetCategoryByID(id string) (*Category, *errors.AppError) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.ErrInvalidObjectID
	}
	category, err := cs.categoryRepo.GetCategory(bson.M{
		"_id": objectID,
	})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err := errors.ErrNotFound
			return nil, errors.NewError("category not found: "+err.Error(), err.StatusCode)
		}
		return nil, errors.ErrInternalServer
	}
	return category, nil
}

func (cs *CategoryService) GetCategoryBySlug(slug string) (*Category, *errors.AppError) {

	category, err := cs.categoryRepo.GetCategory(bson.M{
		"slug": slug,
	})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err := errors.ErrNotFound
			return nil, errors.NewError("category not found: "+err.Error(), err.StatusCode)
		}
		return nil, errors.ErrInternalServer
	}
	return category, nil
}

func (cs *CategoryService) GetCategories() ([]*Category, *errors.AppError) {
	categories, err := cs.categoryRepo.GetCategories(bson.M{})
	if err != nil {
		return nil, errors.ErrInternalServer
	}
	return categories, nil
}

func (cs *CategoryService) UpdateCategory(category *Category) *errors.AppError {
	oldcategory, err := cs.categoryRepo.GetCategory(bson.M{
		"_id": category.ID,
	})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err := errors.ErrNotFound
			return errors.NewError("category not found: "+err.Error(), err.StatusCode)
		}
		return errors.ErrInternalServer
	}
	if category.ParentID != nil {
		if err := cs.checkParentID(category.ParentID); err != nil {
			return err
		}
	}
	category.UpdatedAt = time.Now()
	category.CreatedBy = oldcategory.CreatedBy
	category.CreatedAt = oldcategory.CreatedAt
	category.Slug = slug.Make(category.Name)
	getCategory, err := cs.categoryRepo.GetCategory(bson.M{
		"slug": category.Slug,
		"name": category.Name,
	})
	if err != nil && err != mongo.ErrNoDocuments {
		return errors.ErrInternalServer
	}
	if getCategory != nil && getCategory.ID != category.ID {
		return errors.ErrCategoryAlreadyExists
	}
	if err := cs.categoryRepo.UpdateCategory(bson.M{
		"_id": category.ID,
	}, bson.M{
		"$set": category,
	}); err != nil {
		return errors.NewError("internal error: "+err.Error(), 500)
	}
	return nil
}

func (cs *CategoryService) DeleteCategory(id string, userid primitive.ObjectID) *errors.AppError {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.ErrInvalidObjectID
	}
	category, err := cs.categoryRepo.IsExists(bson.M{
		"_id": objectID,
	})
	if err != nil {
		return errors.NewError("internal error: "+err.Error(), 500)
	}
	if !category {
		err := errors.ErrNotFound
		return errors.NewError("category not found: "+err.Error(), err.StatusCode)
	}
	if err := cs.categoryRepo.DeleteCategory(bson.M{"_id": objectID}); err != nil {
		return errors.NewError("internal error: "+err.Error(), 500)
	}

	//create_deleted category collection
	//use {deleted_by: userid} to add item to deleted category list
	//remove the category from all the items and all categories that have this category as parent
	return nil
}

func (cs *CategoryService) checkParentID(parentID []primitive.ObjectID) *errors.AppError {
	var wg sync.WaitGroup
	idExistChanErr := make(chan *errors.AppError, len(parentID))
	var allErrors []*errors.AppError
	for _, v := range parentID {
		wg.Add(1)
		go func(parentId primitive.ObjectID) {
			exists, err := cs.categoryRepo.IsExists(bson.M{
				"_id": parentId,
			})
			if err != nil {
				idExistChanErr <- errors.NewError("internal error: "+err.Error(), 500)
			}
			if !exists {
				err := errors.ErrNotFound
				idExistChanErr <- errors.NewError("category not found: "+err.Error(), err.StatusCode)
			}
			wg.Done()
		}(v)
		go func() {
			wg.Wait()
			close(idExistChanErr)
		}()
		for err := range idExistChanErr {
			if err != nil {
				allErrors = append(allErrors, err)
			}
		}
	}
	if len(allErrors) > 0 {
		return allErrors[0]
	}
	return nil
}
