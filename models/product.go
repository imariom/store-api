package models

type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

type Products []*Product
type dataStore map[string]Products

// Products data-store
var productList = dataStore{
	"admin@api.com": Products{
		&Product{
			ID:          0,
			Name:        "Test product",
			Description: "Product description",
			Price:       0.0,
		},
	},
}
