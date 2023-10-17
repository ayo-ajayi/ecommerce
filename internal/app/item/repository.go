package item

import (
	"github.com/ayo-ajayi/ecommerce/internal/database"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ItemRepo struct {
	Collection *mongo.Collection
}

func NewItemRepo(collection *mongo.Collection) *ItemRepo {
	return &ItemRepo{
		Collection: collection,
	}
}

func (ir *ItemRepo) IsExists(filter interface{}, opts ...*options.FindOneOptions) (bool, error) {
	ctx, cancel := database.DBReqContext(5)
	defer cancel()
	var item Item
	err := ir.Collection.FindOne(ctx, filter, opts...).Decode(&item)
	if err == mongo.ErrNoDocuments {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (ir *ItemRepo) CreateItem(item *Item) error {
	ctx, cancel := database.DBReqContext(5)
	defer cancel()
	_, err := ir.Collection.InsertOne(ctx, item)
	return err
}

func (ir *ItemRepo) UpdateItem(filter interface{}, update interface{}, opts ...*options.UpdateOptions) error {
	ctx, cancel := database.DBReqContext(5)
	defer cancel()
	_, err := ir.Collection.UpdateOne(ctx, filter, update, opts...)
	return err
}

func (ir *ItemRepo) GetItem(filter interface{}, opts ...*options.FindOneOptions) (*Item, error) {
	ctx, cancel := database.DBReqContext(5)
	defer cancel()
	var item Item
	err := ir.Collection.FindOne(ctx, filter, opts...).Decode(&item)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (ir *ItemRepo) GetItems(filter interface{}, opts ...*options.FindOptions) ([]*Item, error) {
	ctx, cancel := database.DBReqContext(5)
	defer cancel()
	var items []*Item
	cur, err := ir.Collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	for cur.Next(ctx) {
		var item Item
		err := cur.Decode(&item)
		if err != nil {
			return nil, err
		}
		items = append(items, &item)
	}
	return items, nil
}

func (ir *ItemRepo) DeleteItem(filter interface{}, opts ...*options.DeleteOptions) error {
	ctx, cancel := database.DBReqContext(5)
	defer cancel()
	_, err := ir.Collection.DeleteOne(ctx, filter, opts...)
	return err
}
