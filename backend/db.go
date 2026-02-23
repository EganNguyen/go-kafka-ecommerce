package main

import (
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/lib/pq"
)

func initDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := migrateDB(db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	slog.Info("Database connected and migrated")
	return db, nil
}

func migrateDB(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS products (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			price DOUBLE PRECISION NOT NULL DEFAULT 0,
			image_url TEXT NOT NULL DEFAULT '',
			category TEXT NOT NULL DEFAULT '',
			stock INT NOT NULL DEFAULT 0
		);

		CREATE TABLE IF NOT EXISTS orders (
			id TEXT PRIMARY KEY,
			total_price DOUBLE PRECISION NOT NULL DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'placed',
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS order_items (
			id SERIAL PRIMARY KEY,
			order_id TEXT NOT NULL REFERENCES orders(id),
			product_id TEXT NOT NULL,
			name TEXT NOT NULL,
			price DOUBLE PRECISION NOT NULL DEFAULT 0,
			quantity INT NOT NULL DEFAULT 1
		);
	`)
	return err
}

func seedProducts(db *sql.DB) error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM products").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	products := []Product{
		{ID: "prod-001", Name: "Wireless Noise-Cancelling Headphones", Description: "Premium over-ear headphones with active noise cancellation and 30-hour battery life.", Price: 349.99, ImageURL: "https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=400", Category: "Electronics", Stock: 50},
		{ID: "prod-002", Name: "Mechanical Keyboard RGB", Description: "Cherry MX switches with per-key RGB lighting and aluminum frame.", Price: 179.99, ImageURL: "https://images.unsplash.com/photo-1618384887929-16ec33fab9ef?w=400", Category: "Electronics", Stock: 120},
		{ID: "prod-003", Name: "Ultrawide Curved Monitor 34\"", Description: "UWQHD 3440x1440 144Hz IPS panel with USB-C connectivity.", Price: 699.99, ImageURL: "https://images.unsplash.com/photo-1527443224154-c4a3942d3acf?w=400", Category: "Electronics", Stock: 30},
		{ID: "prod-004", Name: "Ergonomic Office Chair", Description: "Adjustable lumbar support, breathable mesh, and 4D armrests.", Price: 549.99, ImageURL: "https://images.unsplash.com/photo-1592078615290-033ee584e267?w=400", Category: "Furniture", Stock: 25},
		{ID: "prod-005", Name: "Smart LED Desk Lamp", Description: "Adjustable color temperature, brightness levels, and USB charging port.", Price: 89.99, ImageURL: "https://images.unsplash.com/photo-1507473885765-e6ed057ab6fe?w=400", Category: "Home", Stock: 200},
		{ID: "prod-006", Name: "Premium Laptop Backpack", Description: "Water-resistant 17\" laptop compartment with anti-theft design.", Price: 129.99, ImageURL: "https://images.unsplash.com/photo-1553062407-98eeb64c6a62?w=400", Category: "Accessories", Stock: 80},
	}

	for _, p := range products {
		_, err := db.Exec(
			"INSERT INTO products (id, name, description, price, image_url, category, stock) VALUES ($1, $2, $3, $4, $5, $6, $7)",
			p.ID, p.Name, p.Description, p.Price, p.ImageURL, p.Category, p.Stock,
		)
		if err != nil {
			return fmt.Errorf("failed to seed product %s: %w", p.ID, err)
		}
	}

	slog.Info("Seeded products", "count", len(products))
	return nil
}
