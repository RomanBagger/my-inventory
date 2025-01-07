package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

var a App

func TestMain(m *testing.M) {

	err := a.Initialize(DbUser, DbPassword, PublicIP, Port, DbName)
	if err != nil {
		log.Fatalf("Error initializing app: %v", err)
	}
	createTable()
	m.Run()
}

func createTable() {
	createTableQuery := `CREATE TABLE IF NOT EXISTS products(
	id int NOT NULL AUTO_INCREMENT,
	name VARCHAR(255) NOT NULL,
	quantity int,
	price float(10,7),
	PRIMARY KEY (id)
	);`

	_, err := a.DB.Exec(createTableQuery)
	if err != nil {
		log.Fatalf("Error creating table: %v", err)
	}
}

func clearTable() {
	a.DB.Exec("DELETE FROM products")
	a.DB.Exec("ALTER TABLE products AUTO_INCREMENT = 1")
	log.Println("Table cleared")
}

func addProduct(name string, quantity int, price float64) int {
	query := fmt.Sprintf("INSERT INTO products(name, quantity, price) VALUES('%v', %v, %v)", name, quantity, price)
	result, err := a.DB.Exec(query)
	if err != nil {
		log.Fatalf("Error adding product: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		log.Fatalf("Error getting last insert ID: %v", err)
	}
	return int(id)
}

func TestGetProducts(t *testing.T) {
	clearTable()
	id := addProduct("keyboard", 100, 500)
	request, _ := http.NewRequest("GET", fmt.Sprintf("/product/%d", id), nil)
	response := sendRequest(request)
	checkStatusCode(http.StatusOK, response.Code, t)
}

func checkStatusCode(expectedStatusCode int, actualStatusCode int, t *testing.T) {
	if expectedStatusCode != actualStatusCode {
		t.Errorf("Expected status code %v but got %v", expectedStatusCode, actualStatusCode)
	}
}

func sendRequest(request *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	a.Router.ServeHTTP(recorder, request)
	return recorder
}

// func TestCreateProduct(t *testing.T) {
// 	clearTable()
// 	var product = []byte(`{"name":"chair", "quantity":1, "price":100}`)
// 	req, _ := http.NewRequest("POST", "/product", bytes.NewBuffer(product))
// 	req.Header.Set("Content-Type", "application/json")

// 	response := sendRequest(req)
// 	checkStatusCode(http.StatusCreated, response.Code, t)

// 	var m map[string]interface{}
// 	json.Unmarshal(response.Body.Bytes(), &m)

// 	if m["name"] != "chair" {
// 		t.Errorf("Expected product name to be 'chair'. Got '%v'", m["name"])
// 	}
// 	log.Printf("%T", m["quantity"])
// 	if m["quantity"] != 1.0 {
// 		t.Errorf("Expected product quantity to be 1. Got %v", m["quantity"])
// 	}
// 	if m["price"] != 100.0 {
// 		t.Errorf("Expected product price to be 100. Got %v", m["price"])
// 	}
// }

func TestDeleteProduct(t *testing.T) {
	clearTable()
	addProduct("connector", 10, 10)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := sendRequest(req)
	checkStatusCode(http.StatusOK, response.Code, t)

	req, _ = http.NewRequest("DELETE", "/product/1", nil)
	response = sendRequest(req)
	checkStatusCode(http.StatusOK, response.Code, t)
}

func TestUpdateProduct(t *testing.T) {
	clearTable()
	addProduct("connector", 10, 10)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := sendRequest(req)
	checkStatusCode(http.StatusOK, response.Code, t)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	m["name"] = "connector updated"
	m["quantity"] = 20
	m["price"] = 20

	product, _ := json.Marshal(m)
	req, _ = http.NewRequest("PUT", "/product/1", bytes.NewBuffer(product))
	req.Header.Set("Content-Type", "application/json")
	response = sendRequest(req)
	checkStatusCode(http.StatusOK, response.Code, t)

	req, _ = http.NewRequest("GET", "/product/1", nil)
	response = sendRequest(req)
	checkStatusCode(http.StatusOK, response.Code, t)

	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] != "connector updated" {
		t.Errorf("Expected product name to be 'connector updated'. Got '%v'", m["name"])
	}
	if m["quantity"] != 20.0 {
		t.Errorf("Expected product quantity to be 20. Got %v", m["quantity"])
	}
	if m["price"] != 20.0 {
		t.Errorf("Expected product price to be 20. Got %v", m["price"])
	}
}