package domain

import "context"

// ProductService defines the interface for fetching product information from the catalog.
type ProductService interface {
	GetProduct(ctx context.Context, id string) (*Product, error)
	ListProducts(ctx context.Context) ([]Product, error)
}
