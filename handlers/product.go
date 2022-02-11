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
	listProductsRe          = regexp.MustCompile(`^/products[/]*$`)
	createProductRe         = regexp.MustCompile(`^/products[/]*$`)
	updateProductRe         = regexp.MustCompile(`^/products/(\d+)$`)
	productCategoriesRe     = regexp.MustCompile(`^/products/categories[/]*$`)
	getProductsByCategoryRe = regexp.MustCompile(`^/products/categories/(\w+)$`)
	getProductRe            = regexp.MustCompile(`^/products/(\d+)$`)
	deleteProductRe         = regexp.MustCompile(`^/products/(\d+)$`)
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
	id, err := getProductId(regex, r.URL.Path)
	if err != nil {
		return nil, ProductNotFound
	}

	// decode the product from the request body
	product := &data.Product{}
	if err := product.FromJSON(r.Body); err != nil {
		return nil, ProductNotFound
	}
	product.ID = uint64(id)

	return product, nil
}

func getProductId(regex regexp.Regexp, exp string) (int, error) {
	// parse id from expression
	matches := regex.FindStringSubmatch(exp)
	if len(matches) < 2 {
		return -1, fmt.Errorf("id not found")
	}

	// convert id to integer
	id, _ := strconv.Atoi(matches[1])

	return id, nil
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

	case r.Method == http.MethodGet && productCategoriesRe.MatchString(r.URL.Path):
		h.GetCategories(rw, r)
		return

	case r.Method == http.MethodGet && getProductsByCategoryRe.MatchString(r.URL.Path):
		h.GetInCategory(rw, r)
		return

	case r.Method == http.MethodGet && getProductRe.MatchString(r.URL.Path):
		h.Get(rw, r)
		return

	case r.Method == http.MethodDelete && deleteProductRe.MatchString(r.URL.Path):
		h.Delete(rw, r)
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

func (h *Product) GetCategories(rw http.ResponseWriter, r *http.Request) {
	h.logger.Println("recieved a GET-categories request")

	products := data.GetAllProducts()
	if err := products.CategoriesToJSON(rw); err != nil {
		http.Error(rw, InternalServerError.Error(), http.StatusInternalServerError)
	}
}

func (h *Product) GetInCategory(rw http.ResponseWriter, r *http.Request) {
	h.logger.Println("received a GET products in category request")

	// get product category
	matches := getProductsByCategoryRe.FindStringSubmatch(r.URL.Path)
	if len(matches) < 2 {
		http.Error(rw, "category not found", http.StatusNotFound)
		return
	}
	category := matches[1] // extract the category name

	// try to get all products in the category
	products := data.GetProductsByCategory(category)
	if len(products) == 0 {
		http.Error(rw, "category not found", http.StatusNotFound)
		return
	}

	if err := products.ToJSON(rw); err != nil {
		http.Error(rw, InternalServerError.Error(), http.StatusInternalServerError)
		return
	}
}

// Limit results
// Sort Results
func (h *Product) Get(rw http.ResponseWriter, r *http.Request) {
	h.logger.Println("received a GET request")

	// get product id
	productID, err := getProductId(*getProductRe, r.URL.Path)
	if err != nil {
		http.Error(rw, ProductNotFound.Error(), http.StatusNotFound)
		return
	}

	// get the product
	product, err := data.GetProduct(productID)
	if err != nil {
		http.Error(rw, ProductNotFound.Error(), http.StatusNotFound)
		return
	}

	// try to return the product
	if err := product.ToJSON(rw); err != nil {
		http.Error(rw, InternalServerError.Error(), http.StatusInternalServerError)
	}
}

func (h *Product) Delete(rw http.ResponseWriter, r *http.Request) {
	h.logger.Println("received a DELETE request")

	// get product id
	productID, err := getProductId(*deleteProductRe, r.URL.Path)
	if err != nil {
		http.Error(rw, ProductNotFound.Error(), http.StatusNotFound)
		return
	}

	// delete product from datastore
	product, err := data.RemoveProduct(productID)
	if err != nil {
		http.Error(rw, ProductNotFound.Error(), http.StatusNotFound)
		return
	}

	// return deleted product
	if err := product.ToJSON(rw); err != nil {
		http.Error(rw,
			fmt.Sprintf("product with ID: '%d' was deleted, but failed to retrieve it",
				product.ID),
			http.StatusInternalServerError)
	}
}
