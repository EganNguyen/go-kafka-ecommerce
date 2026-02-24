package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/egannguyen/go-kafka-ecommerce/backend/internal/entity"
	"github.com/egannguyen/go-kafka-ecommerce/backend/internal/repository"
)

type productRepository struct {
	db *sql.DB
}

// NewProductRepository creates a new ProductRepository backed by Postgres.
func NewProductRepository(db *sql.DB) repository.ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) FindAll(ctx context.Context) ([]entity.Product, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name, description, price, image_url, category, stock FROM products ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var products []entity.Product
	for rows.Next() {
		var p entity.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.ImageURL, &p.Category, &p.Stock); err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, p)
	}
	return products, nil
}

func (r *productRepository) Seed(ctx context.Context, products []entity.Product) error {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM products").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // already seeded
	}

	for _, p := range products {
		_, err := r.db.ExecContext(ctx,
			"INSERT INTO products (id, name, description, price, image_url, category, stock) VALUES ($1, $2, $3, $4, $5, $6, $7)",
			p.ID, p.Name, p.Description, p.Price, p.ImageURL, p.Category, p.Stock,
		)
		if err != nil {
			return fmt.Errorf("failed to seed product %s: %w", p.ID, err)
		}
	}
	return nil
}
