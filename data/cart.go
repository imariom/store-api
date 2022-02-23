package data

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
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

func AddCart(c *Cart) error {
	// TODO: validate if provided user_id and each product_id
	// are valid (talk to users and products models to verify)
	c.ID = getNextCartID()

	cartsRWMtx.Lock()
	cartList = append(cartList, c)
	cartsRWMtx.Unlock()

	return nil
}

func RemoveCart(id uint64) (*Cart, error) {
	index, cart, err := cartExists(id)
	if err != nil {
		return nil, err
	}

	deletedCart := &Cart{}

	cartsRWMtx.Lock()
	*deletedCart = *cart
	tmpList := make(Carts, 0, len(cartList)-1)

	for i, c := range cartList {
		if i == index {
			continue
		}

		tmpList = append(tmpList, c)
	}
	cartList = tmpList
	cartsRWMtx.Unlock()

	return deletedCart, nil
}

func UpdateCart(cart *Cart) error {
	cartsRWMtx.Lock()
	defer cartsRWMtx.Unlock()

	for i, c := range cartList {
		if c.ID == cart.ID {
			cartList[i] = cart
			return nil
		}
	}

	return fmt.Errorf("requested cart does not exist")
}

func SetCart(cart *Cart) error {
	cartsRWMtx.Lock()
	defer cartsRWMtx.Unlock()

	// TODO: optimize this loop by using the userExists() function
	// and using a better data structure.
	for i, c := range cartList {
		if c.ID == cart.ID {
			if cart.UserID != 0 {
				cartList[i].UserID = cart.UserID
			}

			if cart.Products != nil {
				cartList[i].Products = cart.Products
			}

			// set temporary cart equal to original product
			*cart = *cartList[i]
			return nil
		}
	}

	return fmt.Errorf("requested cart does not exist")
}

func GetAllCarts(l int, s string) Carts {
	// sort cart list
	if s == "asc" {
		sort.Sort(cartList)
	} else if s == "desc" {
		sort.Sort(sort.Reverse(cartList))
	}

	// limit the result
	if l <= 0 || l >= cartList.Len() {
		l = cartList.Len()
	}

	cartsRWMtx.RLock()
	defer cartsRWMtx.RUnlock()

	// it is necessary to get a copy of each product from the
	// memory to avoid returning a cartList that while is being
	// used by the caller it is being modified by another goroutine.
	temp := make(Carts, 0, l)
	for i := 0; i != l; i++ {
		temp = append(temp, cartList[i])
	}

	return temp
}

func GetAllUserCarts(userID uint64) Carts {
	cartsRWMtx.RLock()
	defer cartsRWMtx.RUnlock()

	tmpCarts := make(Carts, 0)
	for _, c := range cartList {
		if c.UserID == userID {
			tmpCarts = append(tmpCarts, c)
		}
	}

	return tmpCarts
}

func GetCart(id uint64) (*Cart, error) {
	_, cart, err := cartExists(id)
	if err != nil {
		return nil, err
	}

	return cart, nil
}

func (cs *Carts) ToJSON(w io.Writer) error {
	cartsRWMtx.RLock()
	defer cartsRWMtx.RUnlock()

	return json.NewEncoder(w).Encode(cs)
}

func (c *Cart) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(c)
}

func (c *Cart) FromJSON(r io.Reader) error {
	return json.NewDecoder(r).Decode(c)
}

// sort.Interface implementation for Cart struct.
// This is to allow sorting on a list of carts.

// Len is the number of elements in the collection.
func (c Carts) Len() int {
	cartsRWMtx.RLock()
	defer cartsRWMtx.RUnlock()

	return len(cartList)
}

// Less reports whether the element with index i
// must sort before the element with index j.
//
// If both Less(i, j) and Less(j, i) are false,
// then the elements at index i and j are considered equal.
// Sort may place equal elements in any order in the final result,
// while Stable preserves the original input order of equal elements.
//
// Less must describe a transitive ordering:
//  - if both Less(i, j) and Less(j, k) are true, then Less(i, k) must be true as well.
//  - if both Less(i, j) and Less(j, k) are false, then Less(i, k) must be false as well.
func (c Carts) Less(i, j int) bool {
	cartsRWMtx.RLock()
	defer cartsRWMtx.RUnlock()

	return cartList[i].Date.Before(cartList[j].Date)
}

// Swap swaps the elements with indexes i and j.
func (c Carts) Swap(i, j int) {
	cartsRWMtx.Lock()

	cartList[i], cartList[j] = cartList[j], cartList[i]

	defer cartsRWMtx.Unlock()
}
