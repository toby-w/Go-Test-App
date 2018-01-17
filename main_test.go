package main_test

import (
	"os"
	"testing" // TODO: Read through this package
	"log"
	"net/http" // TODO: Read through this package
	"net/http/httptest" // TODO: Read through this package
	"encoding/json"
	"bytes"
	"strconv" // TODO: Read through this package and relevant string operations

	"test_app" // to refer to current directory package, use main to get the app
)

var a main.App

// As described in testing docs, go test begins in this function to set-up and clean-up tests.
func TestMain(m *testing.M) {
	a = main.App{}
	/*
	a.Initialize(
		// Define these myself outside
		os.Getenv("TEST_DB_USERNAME"),
		os.Getenv("TEST_DB_PASSWORD"),
		os.Getenv("TEST_DB_NAME"))
	*/
	a.Initialize("postgres", "postgres", "postgres", "disable")

	// Pre-app work
	// TODO: Need a way to drop product table before run test if it exists or else error thrown
	ensureTableExists()

	// Run tests
	// Q: How do all the tests below get called?
	code := m.Run()

	// Clean-up
	clearTable()

	// Finish tests
	os.Exit(code)
}

// Note: A common error is miss typing the Table name in model.go, in all queries, search that they have been corrected
// 	 including product_id_seq

func TestEmptyTable(t *testing.T) {
	clearTable()

	// Test get route with empty table
	req, _ := http.NewRequest("GET", "/products", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		// Q: What is testing.T for?
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func TestGetNonExistentProduct(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/product/11", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	// Q: How does json.Unmarshal work?
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Product not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Product not found'. Got '%s'", m["error"])
	}
}

func TestCreateProduct(t *testing.T) {
	clearTable()

	// Becareful with JSON object string typo
	payload := []byte(`{"name":"test product","price":11.22}`)

	req, _ := http.NewRequest("POST", "/product", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] != "test product" {
		t.Errorf("Expected product name to be 'test product'. Got '%v'", m["name"])
	}

	if m["price"] != 11.22 {
		t.Errorf("Expected product price to be '11.22'. Got '%v'", m["price"])
	}

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	// floats, when the target is a map[string]interface{}
	// Q: How are IDs assigned in PostgreSQL?
	if m["id"] != 1.0 {
		t.Errorf("Expected product ID to be '1'. Got '%v'", m["id"])
	}
}

func TestGetProduct(t *testing.T) {
	clearTable()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}


func TestUpdateProduct(t *testing.T) {
	addProducts(1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)
	var originalProduct map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalProduct)

	payload := []byte(`{"name":"test product - updated name","price":11.23}`)

	req, _ = http.NewRequest("PUT", "/product/1", bytes.NewBuffer(payload))
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["id"] != originalProduct["id"] {
		t.Errorf("Expected the id to remain the same (%v). Got %v", originalProduct["id"], m["id"])
	}

	if m["name"] == originalProduct["name"] {
		// The message is flawed, expected is equivalent to actual here.
		t.Errorf("Expected the name to change from '%v' to '%v'. Got '%v'", originalProduct["name"], m["name"], m["name"])
	}

	if m["price"] == originalProduct["price"] {
		// The message is flawed, expected is equivalent to actual here.
		t.Errorf("Expected the price to change from '%v' to '%v'. Got '%v'", originalProduct["price"], m["price"], m["price"])
	}
}

func TestDeleteProduct(t *testing.T) {
	clearTable()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/product/1", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/product/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func ensureTableExists() {
	if _, err := a.DB.Exec(tableCreateQuery); err != nil {
		log.Fatal(err) // equivalent to Print and then exit(1)
	}
}

func clearTable() {
	a.DB.Exec("DELETE FROM product")
	// Q: What does this query do?
	a.DB.Exec("ALTER SEQUENCE product_id_seq RESTART WITH 1")
}

func addProducts(count int) {
	if count < 1 {
		count = 1
	}

	for i := 0; i < count; i++ {
		a.DB.Exec("INSERT INTO product(name, price) VALUES($1, $2)", "Product "+strconv.Itoa(i), (i+1.0)*10)
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		// Q: What is t.Errorf equivalent to?
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

// Q: What're valid SQL types?
// Q: If default value not specified for non-null, then what is the default value?
// Q: CONSTRAINT? PRIMARY KEY?
const tableCreateQuery = `
DROP TABLE product CASCADE;
CREATE TABLE product 
(
id	SERIAL,
name TEXT NOT NULL,
price NUMERIC(10,2) NOT NULL DEFAULT 0.00,
CONSTRAINT product_pkey PRIMARY KEY (id)
)`



