package postgres

import (
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/lib/pq"
)

func InitDB(dsn string) (*sql.DB, error) {
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
