package data

import (
	"encoding/json"
	"io"
	"sync"
)

// to protect read and write operations on the productList data store
var rwMtx = &sync.RWMutex{}

// in-memory product list data store
var productList = Products{
	&Product{
		ID:          0,
		Name:        "The Go Programming Language",
		Description: "Modern, fast, reliable and productive programming language",
		Category:    "books",
		Image:       "",
		Price:       49.99,
	},
}

type Product struct {
	ID          uint64  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Image       string  `json:"image"`
	Price       float64 `json:"price"`
}

type Products []*Product

func GetAllProducts() Products {
	rwMtx.RLock()
	defer rwMtx.RUnlock()

	return productList
}

func AddNewProduct(p *Product) {
	p.ID = getNextID()

	rwMtx.Lock()
	productList = append(productList, p)
	rwMtx.Unlock()
}

func getNextID() uint64 {
	rwMtx.RLock()
	// assumes the product with ID 0 will never be deleted
	lastProduct := productList[len(productList)-1]
	rwMtx.RUnlock()

	return lastProduct.ID + 1
}

func (ps *Products) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(ps)
}

func (p *Product) FromJSON(r io.Reader) error {
	return json.NewDecoder(r).Decode(p)
}

func (p *Product) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(p)
}
