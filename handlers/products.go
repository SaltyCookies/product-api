//Package classification of Product API
//
//Documentation for Product API
//
// 	Schemes: http
// 	BasePath: /
// 	Version: 1.0.0
//
// 	Consumes:
// 	- application/json
//
// 	Produces:
// 	- application/json
//
//swagger:meta

package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"product-api/data"
	"strconv"

	"github.com/gorilla/mux"
)

// A list of products in the response
// swagger:response productsResponse
type productsResponseWrapper struct {
	// All the products in the system
	// in: Body
	Body []data.Product `json:"body"`
}

// A list of products in the request
// swagger:parameters postProducts
type productsRequestWrapper struct {
	// Product to send in system
	// in: Body
	Body []data.Product `json:"body"`
}

// A list of products in the request
// swagger:parameters putProduct
type productsUpdateWrapper struct {
	// Product(s) to update in system
	// in: Body
	Body []data.Product `json:"body"`
}

// No content
// swagger:response noContent
type productsNoContent struct {
}

// swagger:parameters deleteProduct
type productIDParameterWrapper struct {
	// The id of the product to delete from database
	// in: Path
	// required: true
	ID int `json:"id"`
}

// swagger:parameters putProduct
type productIDParameterUpdateWrapper struct {
	// The id of the product to update
	// in: Path
	// required: true
	ID int `json:"id"`
}

type Products struct {
	l *log.Logger
}

func NewProducts(l *log.Logger) *Products {
	return &Products{l}
}

// swagger:route GET / products getProducts
// Returns a list of products
// responses:
//  200: productsResponse
func (p *Products) GetProducts(w http.ResponseWriter, r *http.Request) {
	p.l.Println("GET PRODUCT")
	lp := data.GetProducts()
	err := lp.ToJSON(w)
	if err != nil {
		http.Error(w, "UnableToJson", http.StatusInternalServerError)
	}
}

// swagger:route POST / products postProducts
// Send product or more to DB
// responses:
//  201: noContent
func (p *Products) AddProduct(rw http.ResponseWriter, r *http.Request) {
	p.l.Println("POST PRODUCTS")
	prod := r.Context().Value(KeyProduct{}).(data.Product)
	p.l.Printf("Prod %#v", prod)
	data.AddProduct(&prod)
}

// swagger:route PUT /{id} products putProduct
// Update product by id
// responses:
//  201: noContent
func (p *Products) UpdateProduct(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(rw, "id convert failure", http.StatusBadRequest)
		return
	}
	p.l.Println("PUT PRODUCTS", id)
	prod := r.Context().Value(KeyProduct{}).(data.Product)
	errUpd := data.UpdateProduct(id, &prod)
	if errUpd == data.ErrProductNotFound {
		http.Error(rw, "Product not found", http.StatusNotFound)
		return
	}

	if errUpd != nil {
		http.Error(rw, "Product not found", http.StatusInternalServerError)
		return
	}
}

// swagger:route DELETE /{id} products deleteProduct
// Delete product by id
// responses:
//  201: noContent
func (p *Products) DeleteProduct(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(rw, "id convert failure", http.StatusBadRequest)
		return
	}
	p.l.Println("DELETE PRODUCTS", id)
	err = data.DeleteProduct(id)
	if err == data.ErrProductNotFound {
		http.Error(rw, "Product not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(rw, "Product not found", http.StatusInternalServerError)
		return
	}
}

type KeyProduct struct{}

func (p Products) MiddlewareProductValidation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		prod := data.Product{}
		err := prod.FromJSON(r.Body)
		if err != nil {
			p.l.Println("error serializing product", err)
			http.Error(rw, "unable to json convert", http.StatusBadRequest)
			return
		}

		err = prod.Validate()
		if err != nil {
			p.l.Println("error validating product", err)
			http.Error(rw, fmt.Sprintf("unable to validate: %s", err), http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), KeyProduct{}, prod)
		r = r.WithContext(ctx)
		next.ServeHTTP(rw, r)
	})
}
