package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/egannguyen/go-kafka-ecommerce/product-catalog-service/internal/domain"
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

func (r *productRepository) FindByID(ctx context.Context, id string) (*domain.Product, error) {
	var p domain.Product
	err := r.db.QueryRowContext(ctx, "SELECT id, name, description, price, image_url, category, stock FROM products WHERE id = $1", id).
		Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.ImageURL, &p.Category, &p.Stock)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get product by id: %w", err)
	}
	return &p, nil
}

func (r *productRepository) Update(ctx context.Context, p *domain.Product) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE products SET name = $1, description = $2, price = $3, image_url = $4, category = $5, stock = $6 WHERE id = $7",
		p.Name, p.Description, p.Price, p.ImageURL, p.Category, p.Stock, p.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}
	return nil
}

func (r *productRepository) scanProduct(rows *sql.Rows, p *domain.Product) error {
	return rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.ImageURL, &p.Category, &p.Stock)
}
