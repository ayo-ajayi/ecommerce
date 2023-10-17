package user

import (
	"github.com/ayo-ajayi/ecommerce/internal/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepo struct {
	Collection *mongo.Collection
}

type UserRepository interface {
	IsExists(email string) (bool, error)
	CreateUser(user *User) error
	UpdateUser(filter interface{}, update interface{}, opts ...*options.UpdateOptions) error
	GetUser(filter interface{}) (*User, error)
	GetUsers(filter interface{}) ([]*User, error)
}

func NewUserRepo(collection *mongo.Collection) *UserRepo {
	return &UserRepo{
		Collection: collection,
	}
}

func (ur *UserRepo) IsExists(email string) (bool, error) {
	ctx, cancel := database.DBReqContext(5)
	defer cancel()
	var user User
	err := ur.Collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (ur *UserRepo) CreateUser(user *User) error {
	ctx, cancel := database.DBReqContext(5)
	defer cancel()
	_, err := ur.Collection.InsertOne(ctx, user)
	return err
}

func (ur *UserRepo) UpdateUser(filter interface{}, update interface{}, opts ...*options.UpdateOptions) error {
	ctx, cancel := database.DBReqContext(5)
	defer cancel()
	_, err := ur.Collection.UpdateOne(ctx, filter, update, opts...)
	return err
}

func (ur *UserRepo) GetUser(filter interface{}) (*User, error) {
	ctx, cancel := database.DBReqContext(5)
	defer cancel()
	var user User
	err := ur.Collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (ur *UserRepo) GetUsers(filter interface{}) ([]*User, error) {
	ctx, cancel := database.DBReqContext(5)
	defer cancel()
	var users []*User
	cursor, err := ur.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}
