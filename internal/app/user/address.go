package user

import (
	"time"

	"github.com/ayo-ajayi/ecommerce/internal/errors"
	"github.com/ayo-ajayi/ecommerce/internal/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func (us *UserService) AddAddress(userid primitive.ObjectID, address types.Address) *errors.AppError {
	address.ID = primitive.NewObjectID()
	address.CreatedAt = time.Now()
	address.UpdatedAt = time.Now()

	user, err := us.userRepository.GetUser(bson.M{"_id": userid})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err := errors.ErrNotFound
			return errors.NewError("user not found: "+err.Error(), err.StatusCode)
		}
		return errors.ErrInternalServer
	}
	if len(user.Addresses) == 0 {
		err := us.userRepository.UpdateUser(bson.M{"_id": userid}, bson.M{"$set": bson.M{"addresses": []types.Address{address}}})
		if err != nil {
			return errors.ErrInternalServer
		}
		return nil
	}
	err = us.userRepository.UpdateUser(bson.M{"_id": userid}, bson.M{"$push": bson.M{"addresses": address}})
	if err != nil {
		return errors.ErrInternalServer
	}
	return nil
}

func (us *UserService) RemoveAddress(userid, addressid primitive.ObjectID) *errors.AppError {
	user, err := us.userRepository.GetUser(bson.M{"_id": userid})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err := errors.ErrNotFound
			return errors.NewError("user not found: "+err.Error(), err.StatusCode)
		}
		return errors.ErrInternalServer
	}
	foundAddress := false
	for _, v := range user.Addresses {
		if v.ID == addressid {
			foundAddress = true
			break
		}
	}
	if !foundAddress {
		return errors.NewError("address not found", 404)
	}
	if len(user.Addresses) == 1 {
		err := us.userRepository.UpdateUser(bson.M{"_id": userid}, bson.M{"$set": bson.M{"addresses": []types.Address{}}})
		if err != nil {
			return errors.ErrInternalServer
		}
		return nil
	}
	err = us.userRepository.UpdateUser(bson.M{"_id": userid}, bson.M{"$pull": bson.M{"addresses": bson.M{"_id": addressid}}})
	if err != nil {
		return errors.ErrInternalServer
	}
	return nil
}

func (us *UserService) GetAddresses(userid primitive.ObjectID) ([]types.Address, *errors.AppError) {
	user, err := us.userRepository.GetUser(bson.M{"_id": userid})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err := errors.ErrNotFound
			return nil, errors.NewError("user not found: "+err.Error(), err.StatusCode)
		}
		return nil, errors.ErrInternalServer
	}
	return user.Addresses, nil
}

func (us *UserService) GetAddress(userid, addressid primitive.ObjectID) (*types.Address, *errors.AppError) {
	user, err := us.userRepository.GetUser(bson.M{"_id": userid})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err := errors.ErrNotFound
			return nil, errors.NewError("user not found: "+err.Error(), err.StatusCode)
		}
		return nil, errors.ErrInternalServer
	}
	for _, v := range user.Addresses {
		if v.ID == addressid {
			return &v, nil
		}
	}
	return nil, errors.NewError("address not found", 404)
}

func (us *UserService) UpdateAddress(userid, addressid primitive.ObjectID, address types.Address) *errors.AppError {
	user, err := us.userRepository.GetUser(bson.M{"_id": userid})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err := errors.ErrNotFound
			return errors.NewError("user not found: "+err.Error(), err.StatusCode)
		}
		return errors.ErrInternalServer
	}
	for i, v := range user.Addresses {
		if v.ID == addressid {
			address.CreatedAt = v.CreatedAt
			address.ID = v.ID
			address.UpdatedAt = time.Now()
			user.Addresses[i] = address
			err := us.userRepository.UpdateUser(bson.M{"_id": userid, "addresses._id": addressid}, bson.M{"$set": bson.M{"addresses": user.Addresses}})
			if err != nil {
				return errors.ErrInternalServer
			}
			return nil
		}
	}
	return errors.NewError("address not found", 404)
}
