package handler

import (
	"fmt"
	"net/http"
)

type Order struct{}

func (order *Order) Create(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Create an order")
}

func (order *Order) List(w http.ResponseWriter, r *http.Request) {
	fmt.Println("List all order")
}

func (order *Order) GetByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get an order")
}

func (order *Order) Update(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Update an order")
}

func (order *Order) Delete(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Delete an order")
}
