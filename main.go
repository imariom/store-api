package main

import (
	"log"
	"net/http"
	"os"

	"github.com/imariom/products-api/handlers"
	"github.com/imariom/products-api/server"
)

func main() {
	// Logger for the API
	logger := log.New(os.Stdout, "[PRODUCT API]", log.LstdFlags)

	// api handlers
	productHandler := handlers.NewProduct(logger)

	// multiplexer
	mux := http.NewServeMux()
	mux.Handle("/products/", productHandler)
	mux.Handle("/products", productHandler)

	// create and run server
	server.Run(&server.Options{
		Addr:    "127.0.0.1:8080",
		Handler: mux,
		Logger:  logger,
	})
}
