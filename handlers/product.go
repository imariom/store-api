package handlers

import (
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/imariom/products-api/data"
)

var (
	createProductRe         = regexp.MustCompile(`^/products[/]*$`)
	updateProductRe         = regexp.MustCompile(`^/products/(\d+)$`)
	productCategoriesRe     = regexp.MustCompile(`^/products/categories[/]*$`)
	getProductsByCategoryRe = regexp.MustCompile(`^/products/categories/(\w+)$`)
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

func getProduct(regex *regexp.Regexp, r *http.Request) (*data.Product, error) {
	// try get the id of the product
	id, err := getItemID(regex, r.URL.Path)
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

func (h *Product) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// set API to be json based (send and receive JSON data)
	rw.Header().Set("Content-Type", "application/json")

	switch {
	case r.Method == http.MethodGet:
		h.get(rw, r)
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

// get get all or a single product, it also allows to get
// all products in a specific category or all categories that
// exist on the data store.
func (h *Product) get(rw http.ResponseWriter, r *http.Request) {
	h.logger.Println("[INFO] received a GET product request")

	urlPath := r.URL.Path
	limitRes, sortCriteria := getQueryParams(r.URL.RawQuery)

	// get all products
	listProductsRe := regexp.MustCompile(`^/products[/]?$`)

	if listProductsRe.MatchString(urlPath) {
		products := data.GetAllProducts(limitRes, sortCriteria)

		if err := products.ToJSON(rw); err != nil {
			http.Error(rw, "failed to retrieve products", http.StatusInternalServerError)
		}
		return
	}

	// get a single product
	getProductRe := regexp.MustCompile(`^/products/(\d+)$`)

	if getProductRe.MatchString(urlPath) {
		// get product id
		productId, err := getItemID(getProductRe, urlPath)
		if err != nil {
			http.Error(rw, "invalid product ID", http.StatusBadRequest)
			return
		}

		// try to get product
		product, err := data.GetProduct(productId)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		}

		// try to return the product
		if err := product.ToJSON(rw); err != nil {
			http.Error(rw, "failed to retrieve product", http.StatusInternalServerError)
		}
		return
	}

	// get all product categories
	categoriesRe := regexp.MustCompile(`^/products/categories[/]?$`)

	if categoriesRe.MatchString(urlPath) {
		products := data.GetAllCategories()

		if err := products.ToJSON(rw); err != nil {
			http.Error(rw, "failed to retrieve categories", http.StatusInternalServerError)
		}
		return
	}

	// get all products by category
	productsByCategoryRe := regexp.MustCompile(`^/products/categories/(\w+)$`)

	if productsByCategoryRe.MatchString(urlPath) {
		// get category
		matches := productsByCategoryRe.FindStringSubmatch(urlPath)
		if len(matches) < 2 {
			http.Error(rw, "category not found", http.StatusNotFound)
			return
		}

		// try to get all products
		products := data.GetProductsByCategory(matches[1])
		if len(products) == 0 {
			http.Error(rw, "category not found", http.StatusNotFound)
			return
		}

		if err := products.ToJSON(rw); err != nil {
			http.Error(rw, "failed to retrieve products", http.StatusInternalServerError)
		}
		return
	}

	// if none of the above url paths is satisfied consider this a bad request
	http.Error(rw, "bad request", http.StatusBadRequest)
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
	product, err := getProduct(updateProductRe, r)
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
	product, err := getProduct(updateProductRe, r)
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

func (h *Product) Delete(rw http.ResponseWriter, r *http.Request) {
	h.logger.Println("received a DELETE request")

	// get product id
	productID, err := getItemID(deleteProductRe, r.URL.Path)
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
