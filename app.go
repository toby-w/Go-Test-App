package main

import (
	"database/sql" // SQL library
	"fmt"
	"log"
	"net/http"
	"strconv"
	"encoding/json"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq" // Drivers for SQL
)

type App struct {
	Router *mux.Router
	DB *sql.DB
}

func (a *App) Initialize(user, password, dbname, ssl string) {
	// Establish connection with DB
	connectionString := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s", user, password, dbname, ssl)

	var err error
	// Q: How does sql.Open establish this connection?
	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	// Create router
	a.Router = mux.NewRouter()
	a.initializeRoutes()

}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func (a *App) initializeRoutes() {
	// Q: What does .Methods() do?
	// A: Adds a matcher (the path) for the specified HTTP method (ie. "GET") and a handler
	a.Router.HandleFunc("/products", a.getProducts).Methods("GET")
	a.Router.HandleFunc("/product", a.createProduct).Methods("POST")
	a.Router.HandleFunc("/product/{id:[0-9]+}", a.getProduct).Methods("GET")
	a.Router.HandleFunc("/product/{id:[0-9]+}", a.updateProduct).Methods("PUT")
	a.Router.HandleFunc("/product/{id:[0-9]+}", a.deleteProduct).Methods("DELETE")
}

// Response handlers

func (a *App) getProduct(w http.ResponseWriter, r *http.Request) {
	// Q: What does this return again?
	// A: Creates a map of route variables (whatever was allowed to vary in the route)
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		// Should be an error if we can't parse the string into an int
		respondWithError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	p := product{ID: id}
	if err := p.getProduct(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Product not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, p)
}

func (a *App) getProducts(w http.ResponseWriter, r *http.Request) {
	// Q: What is http.Request FormValue?
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	products, err := getProducts(a.DB, start, count)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, products)
}

func (a *App) createProduct(w http.ResponseWriter, r *http.Request) {
	var p product
	// Q: Why need a decoder here? Can't just specify values in the product?
	// A: Decodes json object body (which holds values of the product)
	decoder := json.NewDecoder(r.Body)
	// Q: what does decoder.Decode(&p) do?
	// A: Stores the fields of the JSON into p
	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Q: What type is err that lets us call err.Error()?
	// A: Whatever is returned by db.QueryRow
	if err := p.createProduct(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, p)
}

func (a *App) updateProduct(w http.ResponseWriter, r *http.Request) {
	// Q: What is the structure of r?
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	var p product
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	p.ID = id

	if err := p.updateProduct(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, p)
}

func (a *App) deleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithJSON(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	p := product{ID:id}
	if err := p.deleteProduct(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	// Q: What're valid http response/request headers?
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}