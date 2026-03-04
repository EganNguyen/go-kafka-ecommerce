package domain

import "context"

// Product represents a product in the store.
type Product struct {
	ID          string  `json:"id" bson:"id"`
	Name        string  `json:"name" bson:"name"`
	Description string  `json:"description" bson:"description"`
	Price       float64 `json:"price" bson:"price"`
	ImageURL    string  `json:"image_url" bson:"image_url"`
	Category    string  `json:"category" bson:"category"`
	Stock       int     `json:"stock" bson:"stock"`
}

type ProductRepository interface {
	FindAll(ctx context.Context) ([]Product, error)
	FindByID(ctx context.Context, id string) (*Product, error)
	Update(ctx context.Context, product *Product) error
}
