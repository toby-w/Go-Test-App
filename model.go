package main

import (
	"database/sql" // TODO: Read this package
)

type product struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Price float64 `json:"price"`
}

func (p *product) getProduct(db *sql.DB) error {
	// Q: What is difference between QueryRow and Query and Exec?
	// A: Exec is used when no rows are returned; Query for multiple rows, QueryRow for single row
	// Q: What is Scan?
	// A: Scan looks like C's scanf, you read the the args and place value in the specified vars
	return db.QueryRow("SELECT name, price FROM product WHERE id=$1",
		p.ID).Scan(&p.Name, &p.Price)
}

func (p *product) updateProduct(db *sql.DB) error {
	_, err := db.Exec("UPDATE product SET name=$1, price=$2 WHERE id=$3",
		p.Name, p.Price, p.ID)

	return err
}

func (p *product) deleteProduct(db *sql.DB) error {
	// TODO: Review PostgreSQL syntax and operations
	_, err := db.Exec("DELETE FROM product WHERE id=$1", p.ID)

	return err
}

func (p *product) createProduct(db *sql.DB) error {
	err := db.QueryRow(
		"INSERT INTO product(name, price) VALUES($1, $2) RETURNING id",
		p.Name, p.Price).Scan(&p.ID)

	// TODO: Just return err below and note that nil means error free
	if err != nil {
		return err
	}

	return nil
}

// Get a list of products from db. start is probably the id. count is the number of items to fetch.
func getProducts(db *sql.DB, start, count int) ([]product, error) {
	// Q: Limit and Offset operations?
	// A: Limit limits the number of records returned, Offset determines how many records are skipped at beginning
	rows, err := db.Query(
		"SELECT id, name, price FROM product LIMIT $1 OFFSET $2",
		count, start)

	if err != nil {
		return nil, err
	}

	// Q: What is rows? why do we need to close it?
	defer rows.Close()

	// Q: Slice of product structs? Why need to instantiate? Can we write []product instead?
	// A: Can't write []product, need the {}, it's to instantiate an empty slice of product
	products := []product{}

	for rows.Next() { // can loop over the rows using Next()
		var p product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}