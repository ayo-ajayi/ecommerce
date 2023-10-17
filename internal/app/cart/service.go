package cart

import (
	"sync"
	"time"

	"github.com/ayo-ajayi/ecommerce/internal/app/item"

	"github.com/ayo-ajayi/ecommerce/internal/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CartService struct {
	cartRepo   CartRepository
	itemRepo   ItemRepository
	updateChan chan itemUpdateRequest
}
type itemUpdateRequest struct {
	itemID         primitive.ObjectID
	quantityChange int
	responseChan   chan error
}

type CartRepository interface {
	CreateCart(cart *Cart) error
	UpdateCart(filter interface{}, update interface{}, opts ...*options.UpdateOptions) error
	GetCart(filter interface{}, opts ...*options.FindOneOptions) (*Cart, error)
	GetCarts(filter interface{}, opts ...*options.FindOptions) ([]*Cart, error)
}

type ItemRepository interface {
	GetItem(filter interface{}, opts ...*options.FindOneOptions) (*item.Item, error)
	UpdateItem(filter interface{}, update interface{}, opts ...*options.UpdateOptions) error
}

func NewCartService(cartRepository CartRepository, itemRepository ItemRepository) *CartService {
	cs := &CartService{
		cartRepo:   cartRepository,
		itemRepo:   itemRepository,
		updateChan: make(chan itemUpdateRequest),
	}
	go cs.processItemUpdates()
	return cs
}

func (cs *CartService) processItemUpdates() {
	for req := range cs.updateChan {
		err := cs.itemRepo.UpdateItem(bson.M{"_id": req.itemID}, bson.M{"$inc": bson.M{"quantity": req.quantityChange}})
		req.responseChan <- err
	}
}

func (cs *CartService) AddToCart(userid primitive.ObjectID, cartItem CartItem) *errors.AppError {
	item, err := cs.itemRepo.GetItem(bson.M{"_id": cartItem.ItemID})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.NewError("item not found with ID: "+cartItem.ItemID.Hex(), 404)
		}
		return errors.ErrInternalServer
	}
	if item.Quantity < cartItem.Quantity {
		return errors.NewError("not enough quantity available in inventory", 400)
	}
	if item.Discount > 0 {
		cartItem.Price = item.Price - item.Price*item.Discount/100
	} else {
		cartItem.Price = item.Price
	}
	cartItem.TotalPrice = cartItem.Price * float64(cartItem.Quantity)
	ct, err := cs.cartRepo.GetCart(bson.M{"user_id": userid})
	if err != nil {
		if err != mongo.ErrNoDocuments {
			return errors.ErrInternalServer
		}
		cart := &Cart{
			UserID:     userid,
			CartItems:  []CartItem{cartItem},
			TotalPrice: cartItem.TotalPrice,
			UpdatedAt:  time.Now(),
		}
		err := cs.cartRepo.CreateCart(cart)
		if err != nil {
			return errors.ErrInternalServer
		}
	} else {
		itemExists := false
		for i, v := range ct.CartItems {
			if v.ItemID == cartItem.ItemID {
				ct.CartItems[i].Quantity += cartItem.Quantity
				ct.CartItems[i].TotalPrice += cartItem.TotalPrice
				itemExists = true
				break
			}
		}
		if !itemExists {
			ct.CartItems = append(ct.CartItems, cartItem)
		}

		ct.TotalPrice = 0.0
		for _, item := range ct.CartItems {
			ct.TotalPrice += item.TotalPrice
		}
	}
	return cs.updateConcurrently(ct, cartItem, add)
}

func (cs *CartService) updateItemQuantity(itemID primitive.ObjectID, quantityChange int) *errors.AppError {
	responseChan := make(chan error)
	cs.updateChan <- itemUpdateRequest{
		itemID:         itemID,
		quantityChange: quantityChange,
		responseChan:   responseChan,
	}
	err := <-responseChan
	close(responseChan)
	if err != nil {
		return errors.ErrInternalServer
	}
	return nil
}

func (cs *CartService) RemoveFromCart(userid primitive.ObjectID, cartItem CartItem) *errors.AppError {
	ct, err := cs.cartRepo.GetCart(bson.M{"user_id": userid})
	if err != nil {
		if err != mongo.ErrNoDocuments {
			return errors.ErrInternalServer
		}
		return errors.NewError("cart not found or empty", 404)
	}
	var foundCartItem *CartItem
	for i, v := range ct.CartItems {
		if v.ItemID == cartItem.ItemID {
			foundCartItem = &ct.CartItems[i]
			break
		}
	}
	if foundCartItem == nil {
		return errors.NewError("item not found in cart", 404)
	}
	if foundCartItem.Quantity < cartItem.Quantity {
		return errors.NewError("cannot remove more items than what is in the cart cart", 400)
	}

	itemFound := false
	for i, v := range ct.CartItems {
		if v.ItemID == cartItem.ItemID {
			if v.Quantity > cartItem.Quantity {
				ct.CartItems[i].Quantity -= cartItem.Quantity
				ct.CartItems[i].TotalPrice = ct.CartItems[i].Price * float64(ct.CartItems[i].Quantity)
			} else {
				ct.CartItems = append(ct.CartItems[:i], ct.CartItems[i+1:]...)
			}
			itemFound = true
			break
		}
	}

	if !itemFound {
		return errors.NewError("item not found in cart", 404)
	}

	ct.TotalPrice = 0.0
	for _, item := range ct.CartItems {
		ct.TotalPrice += item.TotalPrice
	}
	return cs.updateConcurrently(ct, cartItem, remove)
}

func (cs *CartService) GetCart(userid primitive.ObjectID) (*Cart, *errors.AppError) {
	ct, err := cs.cartRepo.GetCart(bson.M{"user_id": userid})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, errors.NewError("cart not found", 404)
	}
	return ct, nil
}

type action string

const add action = "add"
const remove action = "remove"

func (cs *CartService) updateConcurrently(ct *Cart, cartItem CartItem, act action) *errors.AppError {
	errChan := make(chan *errors.AppError, 2)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ct.UpdatedAt = time.Now()
		err := cs.cartRepo.UpdateCart(bson.M{"_id": ct.ID}, bson.M{"$set": bson.M{"cart_items": ct.CartItems, "total_price": ct.TotalPrice, "updated_at": ct.UpdatedAt}})
		if err != nil {
			errChan <- errors.ErrInternalServer
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		var err *errors.AppError
		if act == remove {
			err = cs.updateItemQuantity(cartItem.ItemID, cartItem.Quantity)
		} else {
			err = cs.updateItemQuantity(cartItem.ItemID, -1*cartItem.Quantity)
		}
		if err != nil {
			errChan <- errors.ErrInternalServer
		}
	}()
	wg.Wait()
	close(errChan)
	for err := range errChan {
		if err != nil {
			return err
		}
	}
	return nil
}
