package main

import (
	"database/sql"
	"fmt"
)

type product struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

func getProducts(DB *sql.DB) ([]product, error) {
	query := "SELECT id, name, quantity, price from products"
	rows, err := DB.Query(query)

	if err != nil {
		return nil, err
	}

	products := []product{}
	for rows.Next() {
		var p product
		if err := rows.Scan(&p.ID, &p.Name, &p.Quantity, &p.Price); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

func (p *product) getProduct(DB *sql.DB) error {
	query := fmt.Sprintf("SELECT name, quantity, price FROM products WHERE id=%v", p.ID)
	row := DB.QueryRow(query)
	err := row.Scan(&p.Name, &p.Quantity, &p.Price)
	if err != nil {
		return err
	}
	return nil
}

func (p *product) createProduct(DB *sql.DB) error {
	query := fmt.Sprintf("INSERT INTO products(name, quantity, price) VALUES('%v', %v, %v)", p.Name, p.Quantity, p.Price)
	result, err := DB.Exec(query)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	p.ID = int(id)
	return nil
}

func (p *product) updateProduct(DB *sql.DB) error {
	query := fmt.Sprintf("UPDATE products SET name='%v', quantity=%v, price=%v WHERE id=%v", p.Name, p.Quantity, p.Price, p.ID)
	result, err := DB.Exec(query)
	if err != nil {
		return err
	}
	rowsAffceted, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffceted == 0 {
		return fmt.Errorf("no such row exists")
	}
	return nil
}

func (p *product) deleteProduct(DB *sql.DB) error {
	query := fmt.Sprintf("DELETE FROM products WHERE id=%v", p.ID)
	_, err := DB.Exec(query)
	return err
}