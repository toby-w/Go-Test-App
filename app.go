package main

import (
	"database/sql" // SQL library

	"github.com/gorilla/mux"
	_ "github.com/lib/pq" // Drivers for SQL
	"fmt"
	"log"
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
}

func (a *App) Run(addr string) { }