// main_test.go

package main_test

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	main "github.com/vlaguduva/relyf"
)

var a main.App

func TestMain(m *testing.M) {
	a.Initialize(
		os.Getenv("APP_DB_HOST"),
		os.Getenv("APP_DB_USERNAME"),
		os.Getenv("APP_DB_PASSWORD"),
		os.Getenv("APP_DB_NAME"))

	ensureTableExists()

	code := m.Run()

	clearTable()

	os.Exit(code)
}

func TestEmptyTable(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/products", nil)
	res := executeRequest(req)

	checkReponseCode(t, http.StatusOK, res.Code)
	if body := res.Body.String(); body != "[]" {
		t.Errorf("Expected empty array. Got %s\n", body)
	}
}

func TestNonExistentProduct(t *testing.T) {
	clearTable()
	req, _ := http.NewRequest("GET", "/product/999", nil)
	res := executeRequest(req)

	checkReponseCode(t, http.StatusNotFound, res.Code)

	var m map[string]string
	json.Unmarshal(res.Body.Bytes(), &m)
	if m["error"] != "Product not found." {
		t.Errorf("Expected the 'error' to be 'Product not found'. Got %s\n", m["error"])
	}
}

func TestGetProduct(t *testing.T) {

	TestNewProduct(t)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	res := executeRequest(req)

	checkReponseCode(t, http.StatusOK, res.Code)

	var m map[string]interface{}
	json.Unmarshal(res.Body.Bytes(), &m)
	if m["name"] != "test_product" {
		t.Errorf("Expected the 'name' to be 'test_product'. Got %v\n", m["name"])
	}
	if m["price"] != 11.21 {
		t.Errorf("Expected the 'price' to be '11.21'. Got %v\n", m["price"])
	}
	if m["id"] != 1.0 {
		t.Errorf("Expected the 'id' to be '1.0'. Got %v\n", m["price"])
	}
}

func TestNewProduct(t *testing.T) {
	clearTable()
	var reqBody = []byte(`{"name": "test_product", "price": 11.21}`)
	req, _ := http.NewRequest("POST", "/product", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	res := executeRequest(req)
	checkReponseCode(t, http.StatusOK, res.Code)

	var m map[string]interface{}
	json.Unmarshal(res.Body.Bytes(), &m)
	if m["name"] != "test_product" {
		t.Errorf("Expected the 'name' to be 'test_product'. Got %v\n", m["name"])
	}
	if m["price"] != 11.21 {
		t.Errorf("Expected the 'price' to be '11.21'. Got %v\n", m["price"])
	}
	if m["id"] != 1.0 {
		t.Errorf("Expected the 'id' to be '1.0'. Got %v\n", m["price"])
	}
}

func TestUpdateProduct(t *testing.T) {

	TestNewProduct(t)

	var reqBody = []byte(`{"name": "test_product_updated", "price": 11.22}`)
	req, _ := http.NewRequest("PUT", "/product/1", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application//1json")

	res := executeRequest(req)

	checkReponseCode(t, http.StatusOK, res.Code)

	var m map[string]interface{}
	json.Unmarshal(res.Body.Bytes(), &m)
	if m["name"] != "test_product_updated" {
		t.Errorf("Expected the 'name' to be 'test_product_updated'. Got %v\n", m["name"])
	}
	if m["price"] != 11.22 {
		t.Errorf("Expected the 'price' to be '11.22'. Got %v\n", m["price"])
	}
	if m["id"] != 1.0 {
		t.Errorf("Expected the 'id' to be '1.0'. Got %v\n", m["price"])
	}
}

func TestDeleteProduct(t *testing.T) {

	TestNewProduct(t)

	TestGetProduct(t)

	req, _ := http.NewRequest("DELETE", "/product/1", nil)
	res := executeRequest(req)

	checkReponseCode(t, http.StatusOK, res.Code)

	var m map[string]string

	json.Unmarshal(res.Body.Bytes(), &m)
	if m["error"] != "" {
		t.Errorf("Expected the 'error' to be 'nil'. Got %s\n", m["error"])
	}

	TestGetProduct(t)
}

func ensureTableExists() {
	if _, error := a.DB.Exec(tableCreationQuery); error != nil {
		log.Fatal(error)
	}
}

func clearTable() {
	a.DB.Exec("DELETE from products")
	a.DB.Exec("ALTER SEQUENCE products_id_seq RESTART WITH 1")
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	responseRecorder := httptest.NewRecorder()

	a.Router.ServeHTTP(responseRecorder, req)

	return responseRecorder
}

func checkReponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

const tableCreationQuery = `CREATE TABLE IF NOT EXISTS products (
	id SERIAL,
	name TEXT NOT NULL,
	price NUMERIC (10,2) NOT NULL DEFAULT 0.00,
	CONSTRAINT products_pkey PRIMARY KEY (id)
)`
