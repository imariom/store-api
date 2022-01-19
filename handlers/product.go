package handlers

import (
	"fmt"
	"log"
	"net/http"
)

type Product struct {
	logger *log.Logger
}

func NewProduct(l *log.Logger) *Product {
	return &Product{l}
}

func (h *Product) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(rw, "Hello from Product Handler\n")
}
