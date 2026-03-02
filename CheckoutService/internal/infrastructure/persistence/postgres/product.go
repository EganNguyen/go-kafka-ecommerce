package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/egannguyen/go-kafka-ecommerce/checkout-service/internal/domain"
)

type productRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) domain.ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) FindAll(ctx context.Context) ([]domain.Product, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name, description, price, image_url, category, stock FROM products ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var p domain.Product
		if err := r.scanProduct(rows, &p); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func (r *productRepository) scanProduct(rows *sql.Rows, p *domain.Product) error {
	return rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.ImageURL, &p.Category, &p.Stock)
}

func (r *productRepository) Seed(ctx context.Context, products []domain.Product) error {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM products").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
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
