package handlers

import (
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/imariom/products-api/data"
)

var (
	createProductRe = regexp.MustCompile(`^/products[/]*$`)

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

	// route each incoming request to specific handler
	switch r.Method {
	case http.MethodPost:
		h.create(rw, r)
		return

	case http.MethodGet:
		h.get(rw, r)
		return

	case http.MethodPut:
		fallthrough
	case http.MethodPatch:
		h.update(rw, r)
		return

	case http.MethodDelete:
		h.delete(rw, r)
		return

	default:
		http.Error(rw, "HTTP verb not implemented", http.StatusNotImplemented)
		return
	}
}

// create parse and create new product from request body and
// store this product on internal data store.
func (h *Product) create(rw http.ResponseWriter, r *http.Request) {
	h.logger.Println("[INFO] received a POST product request")

	// create and store new product on the data store
	newProduct := &data.Product{}
	if err := newProduct.FromJSON(r.Body); err != nil {
		http.Error(rw, "invalid product payload", http.StatusBadRequest)
		return
	}
	data.AddNewProduct(newProduct)

	// try to return created product
	if err := newProduct.ToJSON(rw); err != nil {
		http.Error(rw,
			fmt.Sprintf("product with ID '%d' was created, but failed to retrieve it",
				newProduct.ID),
			http.StatusInternalServerError)
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

// update handle PUT requests (when the whole product attributes
// need to be updated), it handle PATCH requests when specific
// attributes of a product need to be updated.
func (h *Product) update(rw http.ResponseWriter, r *http.Request) {
	// update all attrributes of a product
	updateProductRe := regexp.MustCompile(`^/products/(\d+)$`)

	if r.Method == http.MethodPut {
		h.logger.Println("[INFO] received a PUT product request")

		product, err := getProduct(updateProductRe, r)
		if err != nil {
			http.Error(rw, "product not found", http.StatusNotFound)
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

		return
	}

	// update specific attributes of a product
	if r.Method == http.MethodPatch {
		h.logger.Println("[INFO] received a PATCH product request")

		// try to get the product payload and id to be updated (PATCH)
		product, err := getProduct(updateProductRe, r)
		if err != nil {
			http.Error(rw, "product not found", http.StatusNotFound)
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
}

// delete removes (delete) and retrieve single product from the data store.
func (h *Product) delete(rw http.ResponseWriter, r *http.Request) {
	h.logger.Println("[INFO] received a DELETE product request")

	// get product id
	productID, err := getItemID(deleteProductRe, r.URL.Path)
	if err != nil {
		http.Error(rw, "invalid product ID", http.StatusNotFound)
		return
	}

	// delete product from datastore
	product, err := data.RemoveProduct(productID)
	if err != nil {
		http.Error(rw, "product not found", http.StatusNotFound)
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
