package handlers

import (
	"log"
	"net/http"
	"regexp"

	"github.com/imariom/products-api/data"
)

var (
	listCartsRe = regexp.MustCompile(`/cart[/]?`)
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

	switch {
	case r.Method == http.MethodGet && listCartsRe.MatchString(r.URL.Path):
		h.List(rw, r)
		return

	default:
		http.Error(rw, "HTTP ver not implemented", http.StatusNotImplemented)
	}
}

func (h *Cart) List(rw http.ResponseWriter, r *http.Request) {
	h.logger.Println("received a GET-list request")

	carts := data.GetAllCarts()
	if err := carts.ToJSON(rw); err != nil {
		msg := "internal server error, while converting carts to JSON"
		h.logger.Println("[ERROR]", msg)
		http.Error(rw, msg, http.StatusInternalServerError)
	}
}

func (h *Cart) Get(rw http.ResponseWriter, r *http.Request) {
	h.logger.Println("received a GET request")
}
