package data

import (
	"encoding/json"
	"io"
	"sync"
	"time"
)

var cartsRWMtx = &sync.RWMutex{}

type Item struct {
	ProductID uint64 `json:"product_id"`
	Quantity  uint16 `json:"quantity"`
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
			Item{
				ProductID: 0,
				Quantity:  2,
			},
		},
	},
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

func (c *Carts) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(c)
}
