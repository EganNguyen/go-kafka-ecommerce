package domain

import "context"

// Product represents a product in the store.
type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	ImageURL    string  `json:"image_url"`
	Category    string  `json:"category"`
	Stock       int     `json:"stock"`
}

type ProductRepository interface {
	FindAll(ctx context.Context) ([]Product, error)
	FindByID(ctx context.Context, id string) (*Product, error)
	Update(ctx context.Context, product *Product) error
}
