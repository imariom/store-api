package data

import (
	"encoding/json"
	"fmt"
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

func UpdateProduct(prod *Product) error {
	rwMtx.Lock()
	defer rwMtx.Unlock()

	// TODO: optimize the productList data structure for
	// reading and writing (e.g, map data structure)
	for i, p := range productList {
		if p.ID == prod.ID {
			productList[i] = prod
			return nil
		}
	}

	return fmt.Errorf("product does not exist")
}

func SetProduct(prod *Product) error {
	rwMtx.Lock()
	defer rwMtx.Unlock()

	for i, p := range productList {
		if p.ID == prod.ID {
			if prod.Name != "" {
				productList[i].Name = prod.Name
			}

			if prod.Description != "" {
				productList[i].Description = prod.Description
			}

			if prod.Category != "" {
				productList[i].Category = prod.Category
			}

			if prod.Price != 0.0 {
				productList[i].Price = prod.Price
			}

			if prod.Image != "" {
				productList[i].Image = prod.Image
			}

			// set temporary product equal to original product
			*prod = *productList[i]

			return nil
		}
	}

	return fmt.Errorf("product does not exist")
}

func getNextID() uint64 {
	rwMtx.RLock()
	defer rwMtx.RUnlock()

	if len(productList) == 0 {
		return 0
	}

	// assumes the product with ID 0 will never be deleted
	lastProduct := productList[len(productList)-1]

	return lastProduct.ID + 1
}

func productExists(id uint64) (int, bool) {
	rwMtx.RLock()
	for i, p := range productList {
		if p.ID == id {
			return i, true
		}
	}
	rwMtx.RUnlock()

	return -1, false
}

func (ps *Products) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(ps)
}

func (ps *Products) CategoriesToJSON(w io.Writer) error {
	// get all categories <category, COUNT>
	tmpCategories := make(map[string]uint64, 0)

	rwMtx.RLock()
	for _, p := range productList {
		tmpCategories[p.Category] = tmpCategories[p.Category] + 1
	}
	rwMtx.RUnlock()

	// get only category name
	categories := make([]string, 0)
	for category, _ := range tmpCategories {
		categories = append(categories, category)
	}

	return json.NewEncoder(w).Encode(categories)
}

func GetProductsByCategory(category string) Products {
	products := Products{}

	rwMtx.RLock()
	for _, p := range productList {
		if p.Category == category {
			products = append(products, p)
		}
	}
	rwMtx.RUnlock()

	return products
}

func GetProduct(id int) (*Product, error) {
	product := &Product{}

	// TODO: optimize productList to another data structure
	// for fast reading and writing
	rwMtx.RLock()
	defer rwMtx.RUnlock()
	for _, p := range productList {
		if p.ID == uint64(id) {
			*product = *p // to avoid reading concurrently accessed product
			return product, nil
		}
	}

	return nil, fmt.Errorf("product not found")
}

func RemoveProduct(id int) (*Product, error) {
	// checks wheter product exists
	index, exists := productExists(uint64(id))
	if !exists {
		return nil, fmt.Errorf("product does not exist")
	}

	deletedProduct := &Product{}

	// remove product from datastore
	rwMtx.RLock()
	*deletedProduct = *productList[index]
	tmpList := make(Products, 0, len(productList)-1)

	for i, p := range productList {
		if i == index {
			continue
		}

		tmpList = append(tmpList, p)
	}
	productList = tmpList

	rwMtx.RUnlock()

	return deletedProduct, nil
}

func (p *Product) FromJSON(r io.Reader) error {
	return json.NewDecoder(r).Decode(p)
}

func (p *Product) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(p)
}
