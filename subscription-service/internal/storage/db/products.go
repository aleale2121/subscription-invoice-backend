package db

import (
	"context"
	"database/sql"
	"log"
)

// Product is the structure which holds one product from the database.
type Product struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type ProductPersistence struct {
	db *sql.DB
}

// NewProductsPersistence is the function used to create an instance of the ProductPersistence.
func NewProductsPersistence(dbPool *sql.DB) ProductPersistence {
	return ProductPersistence{db: dbPool}
}

// GetAllProducts returns a slice of all products
func (p *ProductPersistence) GetAllProducts() ([]*Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `SELECT code, name FROM products`

	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*Product

	for rows.Next() {
		var product Product
		err := rows.Scan(&product.Code, &product.Name)
		if err != nil {
			log.Println("Error scanning", err)
			return nil, err
		}

		products = append(products, &product)
	}

	return products, nil
}

// GetProductByCode returns a product by code
func (p *ProductPersistence) GetProductByCode(code string) (*Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `SELECT code, name FROM products WHERE code = $1`

	var product Product
	row := p.db.QueryRowContext(ctx, query, code)

	err := row.Scan(&product.Code, &product.Name)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

// UpdateProduct updates a product in the database
func (p *ProductPersistence) UpdateProduct(product Product) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `UPDATE products SET name = $1 WHERE code = $2`

	_, err := p.db.ExecContext(ctx, stmt, product.Name, product.Code)
	if err != nil {
		return err
	}

	return nil
}

// DeleteProduct deletes a product from the database
func (p *ProductPersistence) DeleteProduct(productCode string) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `DELETE FROM products WHERE code = $1`

	_, err := p.db.ExecContext(ctx, stmt, productCode)
	if err != nil {
		return err
	}

	return nil
}

// InsertProduct inserts a new product into the database
func (p *ProductPersistence) InsertProduct(product Product) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `INSERT INTO products (code, name) VALUES ($1, $2)`

	_, err := p.db.ExecContext(ctx, stmt, product.Code, product.Name)
	if err != nil {
		return err
	}

	return nil
}
