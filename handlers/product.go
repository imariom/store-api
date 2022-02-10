package handlers

import (
	"log"
	"net/http"
	"regexp"

	"github.com/imariom/products-api/data"
)

var (
	listProductsRe = regexp.MustCompile(`^/products[/]?$`)
)

type Product struct {
	logger *log.Logger
}

func NewProduct(l *log.Logger) *Product {
	return &Product{l}
}

func (h *Product) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// set API to be json based (send and receive JSON data)
	rw.Header().Set("Content-Type", "application/json")

	switch {
	case r.Method == http.MethodGet && listProductsRe.MatchString(r.URL.Path):
		h.List(rw, r)
		return

	default:
		msg := "HTTP verb not implemented"
		h.logger.Println(msg)
		http.Error(rw, msg, http.StatusNotImplemented)
		return
	}
}

func (h *Product) List(rw http.ResponseWriter, r *http.Request) {
	h.logger.Println("received a GET-list request")

	products := data.GetAllProducts()
	if err := products.FromJSON(rw); err != nil {
		msg := "internal server error, while converting products to JSON"
		h.logger.Println("[ERROR]", msg)
		http.Error(rw, msg, http.StatusInternalServerError)
		return
	}
}
