package handlers

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/imariom/products-api/data"
)

type Cart struct {
	logger *log.Logger
}

func NewCart(l *log.Logger) *Cart {
	return &Cart{l}
}

// parseUser try to parse user data from incoming request.
func parseCart(regex *regexp.Regexp, r *http.Request) (*data.Cart, error) {
	// try to parse user id
	id, err := getItemID(regex, r.URL.Path)
	if err != nil {
		return nil, fmt.Errorf("invalid cart ID")
	}

	// try to decode user from request body
	cart := &data.Cart{}
	if err := cart.FromJSON(r.Body); err != nil {
		return nil, fmt.Errorf("invalid cart payload")
	}

	// this line update the current date and time, every
	// update request must be updated to reflect change.
	cart.Date = time.Now()

	cart.ID = id
	return cart, nil
}

func (h *Cart) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// set API to be json based (send and receive JSON data)
	rw.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		h.get(rw, r)
		return

	case http.MethodPost:
		h.create(rw, r)
		return

	case http.MethodDelete:
		h.delete(rw, r)
		return

	case http.MethodPut:
		fallthrough
	case http.MethodPatch:
		h.update(rw, r)
		return

	default:
		http.Error(rw, "HTTP verb not implemented", http.StatusNotImplemented)
	}
}

func (h *Cart) get(rw http.ResponseWriter, r *http.Request) {
	h.logger.Println("received a GET request")

	// list all carts
	listCartsRe := regexp.MustCompile(`^/cart[/]?$`)
	limitRes, sortCriteria := getQueryParams(r.URL.RawQuery)

	if listCartsRe.MatchString(r.URL.Path) {
		carts := data.GetAllCarts(limitRes, sortCriteria)

		if err := carts.ToJSON(rw); err != nil {
			msg := "internal server error, while converting carts to JSON"
			h.logger.Println("[ERROR]", msg)
			http.Error(rw, msg, http.StatusInternalServerError)
		}
		return
	}

	// list single cart
	getCartRe := regexp.MustCompile(`^/cart/(\d+)$`)

	if getCartRe.MatchString(r.URL.Path) {
		cartID, err := getItemID(getCartRe, r.URL.Path)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		}

		cart, err := data.GetCart(uint64(cartID))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		}

		if err := cart.ToJSON(rw); err != nil {
			http.Error(rw, "failed to convert cart", http.StatusInternalServerError)
		}
	}
}

func (h *Cart) create(rw http.ResponseWriter, r *http.Request) {
	h.logger.Println("received a PUT request")

	// parse cart from request object
	cart := &data.Cart{}
	if err := cart.FromJSON(r.Body); err != nil {
		http.Error(rw, "invalid cart payload", http.StatusBadRequest)
		return
	}
	cart.Date = time.Now()

	// add cart to data store
	data.AddCart(cart)

	// try to return created cart
	if err := cart.ToJSON(rw); err != nil {
		http.Error(rw,
			fmt.Sprintf("user created with ID: '%d', but failed to retrieve it",
				cart.ID),
			http.StatusInternalServerError)
	}
}

func (h *Cart) delete(rw http.ResponseWriter, r *http.Request) {
	h.logger.Println("received a DELETE cart request")

	deleteCartRe := regexp.MustCompile(`^/cart/(\d+)$`)

	// get cart id
	cartID, err := getItemID(deleteCartRe, r.URL.Path)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}

	// delete cart from datastore
	cart, err := data.RemoveCart(cartID)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}

	// return deleted cart to client
	if err := cart.ToJSON(rw); err != nil {
		http.Error(rw,
			fmt.Sprintf("cart with ID: '%d' was deleted, but failed to retrieve it",
				cart.ID),
			http.StatusInternalServerError)
	}
}

func (h *Cart) update(rw http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPut {
		h.logger.Println("received a PUT cart request")
	} else if r.Method == http.MethodPatch {
		h.logger.Println("received a PATCH cart request")
	}

	// try to parse cart from request object
	updateCartRe := regexp.MustCompile(`^/cart/(\d+)$`)

	cart, err := parseCart(updateCartRe, r)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}

	// match request method (PUT or PATCH)
	if r.Method == http.MethodPut {
		// update whole cart information
		if err := data.UpdateCart(cart); err != nil {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		}
	} else if r.Method == http.MethodPatch {
		// update cart attributes
		if err := data.SetCart(cart); err != nil {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		}
	}

	// return updated cart
	if err := cart.ToJSON(rw); err != nil {
		http.Error(rw,
			fmt.Sprintf("cart with ID: '%d' was updated sucessfully, but failed to retrieve it",
				cart.ID),
			http.StatusInternalServerError)
	}
}
