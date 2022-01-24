package handlers

import (
	"log"
	"net/http"
	"regexp"
)

var (
	createRe = regexp.MustCompile(`^/api/products/create[/]*$`)
	updateRe = regexp.MustCompile(`^/api/products/(\d+)$`)
	changeRe = regexp.MustCompile(`^/api/products/(\d+)$`)
	deleteRe = regexp.MustCompile(`^/api/products/(\d+)$`)
	getRe    = regexp.MustCompile(`^/api/products/(\d+)$`)
	listRe   = regexp.MustCompile(`^/api/products[/]*$`)
)

type Product struct {
	logger *log.Logger
}

func NewProduct(l *log.Logger) *Product {
	return &Product{l}
}

func (h *Product) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	reqPath := r.URL.Path

	switch {
	case r.Method == http.MethodPost && createRe.MatchString(reqPath):
		h.Create(rw, r)
		return

	case r.Method == http.MethodPut && updateRe.MatchString(reqPath):
		h.Update(rw, r)
		return

	case r.Method == http.MethodPatch && changeRe.MatchString(reqPath):
		h.Change(rw, r)
		return

	case r.Method == http.MethodDelete && deleteRe.MatchString(reqPath):
		h.Delete(rw, r)
		return

	case r.Method == http.MethodGet && getRe.MatchString(reqPath):
		h.Get(rw, r)
		return

	case r.Method == http.MethodGet && listRe.MatchString(reqPath):
		h.List(rw, r)
		return

	default:
		msg := "method not implemented"
		h.logger.Println("[ERROR]", msg)

		http.Error(rw, msg, http.StatusNotImplemented)
		return
	}
}

// Create create and add a new product on the data-store
func (h *Product) Create(rw http.ResponseWriter, r *http.Request) {
	h.logger.Println("received a POST request")
}

// Update update all attributes of a particular product.
func (h *Product) Update(rw http.ResponseWriter, r *http.Request) {
	h.logger.Println("received a PUT request")
}

// Change update specific attributes of a particular product.
func (h *Product) Change(rw http.ResponseWriter, r *http.Request) {
	h.logger.Println("received a PATCH request")
}

// Delete get and remove a product form the data-store.
func (h *Product) Delete(rw http.ResponseWriter, r *http.Request) {
	h.logger.Println("received a DELETE request")
}

// Get search and retrieve a product given by ID.
func (h *Product) Get(rw http.ResponseWriter, r *http.Request) {
	h.logger.Println("received a GET request")
}

// List get all products stored by a user.
func (h *Product) List(rw http.ResponseWriter, r *http.Request) {
	h.logger.Println("received a GET-list request")
}
