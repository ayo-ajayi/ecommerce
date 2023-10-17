package cart

import (
	"github.com/ayo-ajayi/ecommerce/internal/database"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CartRepo struct {
	Collection *mongo.Collection
}

func NewCartRepo(collection *mongo.Collection) *CartRepo {
	return &CartRepo{
		Collection: collection,
	}
}

func (cr *CartRepo) CreateCart(cart *Cart) error {
	ctx, cancel := database.DBReqContext(5)
	defer cancel()
	_, err := cr.Collection.InsertOne(ctx, cart)
	return err
}

func (cr *CartRepo) UpdateCart(filter interface{}, update interface{}, opts ...*options.UpdateOptions) error {
	ctx, cancel := database.DBReqContext(5)
	defer cancel()
	_, err := cr.Collection.UpdateOne(ctx, filter, update, opts...)
	return err
}

func (cr *CartRepo) GetCart(filter interface{}, opts ...*options.FindOneOptions) (*Cart, error) {
	ctx, cancel := database.DBReqContext(5)
	defer cancel()
	cart := &Cart{}
	err := cr.Collection.FindOne(ctx, filter, opts...).Decode(cart)
	return cart, err
}
func (cr *CartRepo) GetCarts(filter interface{}, opts ...*options.FindOptions) ([]*Cart, error) {
	ctx, cancel := database.DBReqContext(5)
	defer cancel()
	var carts []*Cart
	cursor, err := cr.Collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &carts); err != nil {
		return nil, err
	}
	return carts, nil
}

func (cr *CartRepo) DeleteCart(filter interface{}, opts ...*options.DeleteOptions) error {
	ctx, cancel := database.DBReqContext(5)
	defer cancel()
	_, err := cr.Collection.DeleteOne(ctx, filter, opts...)
	return err
}
