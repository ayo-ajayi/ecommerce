package category

import (
	"github.com/ayo-ajayi/ecommerce/internal/database"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CategoryRepo struct {
	Collection *mongo.Collection
}

func NewCategoryRepo(collection *mongo.Collection) *CategoryRepo {
	return &CategoryRepo{
		Collection: collection,
	}
}

func (cr *CategoryRepo) IsExists(filter interface{}, opts ...*options.FindOneOptions) (bool, error) {
	ctx, cancel := database.DBReqContext(5)
	defer cancel()
	var category Category
	err := cr.Collection.FindOne(ctx, filter, opts...).Decode(&category)
	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (cr *CategoryRepo) CreateCategory(category *Category) error {
	ctx, cancel := database.DBReqContext(5)
	defer cancel()
	_, err := cr.Collection.InsertOne(ctx, category)
	return err
}

func (cr *CategoryRepo) UpdateCategory(filter interface{}, update interface{}, opts ...*options.UpdateOptions) error {
	ctx, cancel := database.DBReqContext(5)
	defer cancel()
	_, err := cr.Collection.UpdateOne(ctx, filter, update, opts...)
	return err
}

func (cr *CategoryRepo) GetCategory(filter interface{}, opts ...*options.FindOneOptions) (*Category, error) {
	ctx, cancel := database.DBReqContext(5)
	defer cancel()
	var category Category
	err := cr.Collection.FindOne(ctx, filter, opts...).Decode(&category)
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (cr *CategoryRepo) GetCategories(filter interface{}, opts ...*options.FindOptions) ([]*Category, error) {
	ctx, cancel := database.DBReqContext(5)
	defer cancel()
	var categories []*Category
	cursor, err := cr.Collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	err = cursor.All(ctx, &categories)
	if err != nil {
		return nil, err
	}
	return categories, nil
}

func (cr *CategoryRepo) DeleteCategory(filter interface{}, opts ...*options.DeleteOptions) error {
	ctx, cancel := database.DBReqContext(5)
	defer cancel()
	_, err := cr.Collection.DeleteOne(ctx, filter, opts...)
	return err
}
