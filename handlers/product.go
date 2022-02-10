package handlers

import (
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/imariom/products-api/data"
)

var (
	listProductsRe = regexp.MustCompile(`^/products[/]?$`)
	createProductRe = regexp.MustCompile(`^/products[/]?$`)
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

	case r.Method == http.MethodPost && createProductRe.MatchString(r.URL.Path):
		h.Create(rw, r)
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
	if err := products.ToJSON(rw); err != nil {
		msg := "internal server error, while converting products to JSON"
		h.logger.Println("[ERROR]", msg)
		http.Error(rw, msg, http.StatusInternalServerError)
		return
	}
}

func (h *Product) Create(rw http.ResponseWriter, r *http.Request) {
	h.logger.Println("received a POST request")

	// create and store new product on the data store
	newProduct := &data.Product{}
	if err := newProduct.FromJSON(r.Body); err != nil {
		http.Error(rw, "payload not valid", http.StatusBadRequest)
		return
	}
	data.AddNewProduct(newProduct)

	// try to return created product
	if err := newProduct.ToJSON(rw); err != nil {
		http.Error(rw,
			fmt.Sprintf("product created with ID: '%d', but failed to retrieve it",
				newProduct.ID),
			http.StatusInternalServerError)
	}
}

// Update product (PUT and PATCH)
// Get all categories
// Get products in a specific category
// Limit results
// Sort Results
// Get a Single Product
// Delete product
