package data

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"
)

var cartsRWMtx = &sync.RWMutex{}

type Item struct {
	ProductID uint64 `json:"product_id"`
	Quantity  uint64 `json:"quantity"`
}

type Cart struct {
	ID       uint64    `json:"id"`
	UserID   uint64    `json:"userId"`
	Date     time.Time `json:"date"` // YYYY-MM-DD
	Products []Item    `json:"products"`
}

type Carts []*Cart

var cartList = Carts{
	&Cart{
		ID:     0,
		UserID: 0,
		Date:   time.Now(),
		Products: []Item{
			{
				ProductID: 0,
				Quantity:  2,
			},
		},
	},
}

func cartExists(id uint64) (int, *Cart, error) {
	cartsRWMtx.RLock()
	defer cartsRWMtx.RUnlock()

	for i, c := range cartList {
		if id == c.ID {
			tmp := &Cart{}
			*tmp = *c
			return i, tmp, nil
		}
	}

	return -1, nil, fmt.Errorf("requested cart does not exist")
}

func getNextCartID() uint64 {
	cartsRWMtx.RLock()
	defer cartsRWMtx.RUnlock()

	if len(cartList) == 0 {
		return 0
	}

	lastCart := cartList[len(cartList)-1]

	return lastCart.ID + 1
}

func AddCart(c *Cart) {
	c.ID = getNextCartID()

	cartsRWMtx.Lock()
	cartList = append(cartList, c)
	cartsRWMtx.Unlock()
}

func GetAllCarts() Carts {
	// it is necessary to get a copy of each product from the
	// memory to avoid returning a cartList that while is being
	// used by the caller it is being modified by another goroutine.
	cartsRWMtx.RLock()
	temp := make(Carts, 0, len(cartList))

	temp = append(temp, cartList...)
	cartsRWMtx.RUnlock()

	return temp
}

func GetCart(id uint64) (*Cart, error) {
	_, cart, err := cartExists(id)
	if err != nil {
		return nil, err
	}

	return cart, nil
}

func (cs *Carts) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(cs)
}

func (c *Cart) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(c)
}

func (c *Cart) FromJSON(r io.Reader) error {
	return json.NewDecoder(r).Decode(c)
}
