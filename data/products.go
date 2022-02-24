package data

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"sync"
)

// to protect read and write operations on the productList data store
var productsRWMtx = &sync.RWMutex{}

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

// Products represent the type of the In-Memory Data Store
// 'productList'.
//
// It also was created to allow methods on the slice of
// products used as data-store.
type Products []*Product

// Categories represent a key-value pair data structure to track
// the number of times a category appears on the data store.
// On the client side it is an object containing a list of
// key-value pairs (the key represent the name of the category, and
// the value the number of times that category appear on the data
// store).
//
// It also was created as an alias to allow objects of this type
// provide convenient methods such as ToJSON for readability
// and easy enconding and retrieve of the data for the client.
type Categories map[string]uint16

func getNextProductId() uint64 {
	productsRWMtx.RLock()
	defer productsRWMtx.RUnlock()

	if len(productList) == 0 {
		return 0
	}

	lastProduct := productList[len(productList)-1]

	return lastProduct.ID + 1
}

func productExists(id uint64) (int, bool) {
	productsRWMtx.RLock()
	defer productsRWMtx.RUnlock()

	for i, p := range productList {
		if p.ID == id {
			return i, true
		}
	}

	return -1, false
}

// GetAllProducts retrieve a slice of all products that
// exist on the data store.
func GetAllProducts(limitRes int, sortCriteria string) Products {
	// sort products
	if sortCriteria == "asc" {
		// sort products in ascending order of price
		sort.Sort(&productList)
	} else if sortCriteria == "desc" {
		// sort products in descending order of price
		sort.Sort(sort.Reverse(&productList))
	}

	// limit number of products to return
	if limitRes <= 0 || limitRes >= productList.Len() {
		limitRes = productList.Len()
	}

	productsRWMtx.RLock()
	defer productsRWMtx.RUnlock()

	tmpProducts := make(Products, 0, limitRes)
	for i := 0; i != limitRes; i++ {
		tmpProducts = append(tmpProducts, productList[i])
	}

	return tmpProducts
}

// GetProduct get and retrieve a product from the data store.
func GetProduct(prodId uint64) (*Product, error) {
	product := &Product{}

	productsRWMtx.RLock()
	defer productsRWMtx.RUnlock()

	// TODO: optimize productList to another data structure
	// for fast reading and writing.
	for _, prod := range productList {
		if prod.ID == prodId {
			*product = *prod // to avoid reading concurrently accessed product
			return product, nil
		}
	}

	return nil, fmt.Errorf("product not found")
}

// GetAllCategories return all the categories that exist on the data
// store in the form of an object.
// This object contains the name of the category and its count (the
// number of times that category appear on the data store).
// The object will contain a key-value pair in the form
// { "category0": count, "category1": count, ... }
func GetAllCategories() Categories {
	categories := make(Categories, 0)

	// prevent concurrent access
	productsRWMtx.RLock()
	defer productsRWMtx.RUnlock()

	for _, p := range productList {
		categories[p.Category] = categories[p.Category] + 1
	}

	return categories
}

// GetProductsByCategory retrieve all products on a specific
// category in the data store.
func GetProductsByCategory(category string) Products {
	products := Products{}

	productsRWMtx.RLock()
	defer productsRWMtx.RUnlock()

	for _, p := range productList {
		if p.Category == category {
			products = append(products, p)
		}
	}

	return products
}

func AddNewProduct(p *Product) {
	p.ID = getNextProductId()

	productsRWMtx.Lock()
	productList = append(productList, p)
	productsRWMtx.Unlock()
}

func UpdateProduct(prod *Product) error {
	productsRWMtx.Lock()
	defer productsRWMtx.Unlock()

	// TODO: optimize the productList data structure for
	// reading and writing (e.g, map data structure)
	for i, p := range productList {
		if p.ID == prod.ID {
			productList[i] = prod
			return nil
		}
	}

	return fmt.Errorf("product not found")
}

func SetProduct(prod *Product) error {
	productsRWMtx.Lock()
	defer productsRWMtx.Unlock()

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

	return fmt.Errorf("product not found")
}

func RemoveProduct(id uint64) (*Product, error) {
	// checks wheter product exists
	index, exists := productExists(uint64(id))
	if !exists {
		return nil, fmt.Errorf("product not found")
	}

	deletedProduct := &Product{}

	// remove product from datastore
	productsRWMtx.RLock()

	*deletedProduct = *productList[index]
	tmpList := make(Products, 0, len(productList)-1)

	for i, p := range productList {
		if i == index {
			continue
		}

		tmpList = append(tmpList, p)
	}
	productList = tmpList

	productsRWMtx.RUnlock()

	return deletedProduct, nil
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

func (c *Categories) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(c)
}

// Len is the number of elements in the collection productList.
func (p *Products) Len() int {
	productsRWMtx.RLock()
	length := len(productList)
	productsRWMtx.RUnlock()

	return length
}

// Less reports whether the product with index i
// must sort before the product with index j.
func (p *Products) Less(i, j int) bool {
	productsRWMtx.RLock()
	defer productsRWMtx.RUnlock()

	return productList[i].Price < productList[j].Price
}

// Swap swaps the products with indexes i and j.
func (p *Products) Swap(i, j int) {
	productsRWMtx.Lock()
	defer productsRWMtx.Unlock()

	productList[i], productList[j] = productList[j], productList[i]
}
