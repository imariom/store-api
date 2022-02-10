package data

import (
	"encoding/json"
	"io"
)

type Product struct {
	ID          uint64  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	SKU         string  `json:"sku"`
	Image       string  `json:"image"`
	Price       float64 `json:"price"`
}

type Products []*Product

var productList = Products{
	&Product{
		ID:          0,
		Name:        "Test product",
		Description: "Product description",
		Category:    "",
		SKU:         "",
		Image:       "",
		Price:       0.0,
	},
}

func GetAllProducts() Products {
	return productList
}

func getNextID() uint64 {
	lastProduct := productList[len(productList)-1]
	return lastProduct.ID + 1
}

func (ps *Products) FromJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(ps)
}
