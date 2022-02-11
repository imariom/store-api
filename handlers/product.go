package handlers

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/imariom/products-api/data"
)

var (
	listProductsRe  = regexp.MustCompile(`^/products[/]?$`)
	createProductRe = regexp.MustCompile(`^/products[/]?$`)
	updateProductRe = regexp.MustCompile(`^/products/(\d+)$`)
)

var (
	ProductNotFound     = fmt.Errorf("product not found")
	InvalidPayload      = fmt.Errorf("invalid payload")
	InternalServerError = fmt.Errorf("internal server error")
	BadRequestError     = fmt.Errorf("bad request")
)

type Product struct {
	logger *log.Logger
}

func NewProduct(l *log.Logger) *Product {
	return &Product{l}
}


func getProduct(regex regexp.Regexp, r *http.Request) (*data.Product, error) {
	// try get the id of the product
	matches := regex.FindStringSubmatch(r.URL.Path)
	if len(matches) < 2 {
		return nil, ProductNotFound
	}

	// convert product id to integer
	id, _ := strconv.Atoi(matches[1])

	// decode the product from the request body
	product := &data.Product{}
	if err := product.FromJSON(r.Body); err != nil {
		return nil, ProductNotFound
	}
	product.ID = uint64(id)

	return product, nil
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

	case r.Method == http.MethodPut && updateProductRe.MatchString(r.URL.Path):
		h.Update(rw, r)
		return

	case r.Method == http.MethodPatch && updateProductRe.MatchString(r.URL.Path):
		h.Set(rw, r)
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

func (h *Product) Update(rw http.ResponseWriter, r *http.Request) {
	h.logger.Println("received a PUT request")

	// try to get the product payload and id to be updated
	product, err := getProduct(*updateProductRe, r)
	if err == ProductNotFound {
		http.Error(rw, ProductNotFound.Error(), http.StatusNotFound)
		return
	}

	// update whole product information
	if err := data.UpdateProduct(product); err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}

	// return updated product
	if err := product.ToJSON(rw); err != nil {
		http.Error(rw,
			fmt.Sprintf("product with ID: '%d' was updated, but failed to retrieve it",
				product.ID),
			http.StatusInternalServerError)
	}
}

func (h *Product) Set(rw http.ResponseWriter, r *http.Request) {
	h.logger.Println("received a PATCH request")

	// try to get the product payload and id to be updated (PATCH)
	product, err := getProduct(*updateProductRe, r)
	if err == ProductNotFound {
		http.Error(rw, ProductNotFound.Error(), http.StatusNotFound)
		return
	}

	// update product attributes
	if err := data.SetProduct(product); err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}

	// return updated product
	if err := product.ToJSON(rw); err != nil {
		http.Error(rw,
			fmt.Sprintf("product with ID: '%d' was updated, but failed to retrieve it",
				product.ID),
			http.StatusInternalServerError)
	}
}

// Get all categories
// Get products in a specific category
// Limit results
// Sort Results
// Get a Single Product
// Delete product
