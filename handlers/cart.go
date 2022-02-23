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

	default:
		http.Error(rw, "HTTP ver not implemented", http.StatusNotImplemented)
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
