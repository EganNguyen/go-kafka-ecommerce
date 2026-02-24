package repository

import (
	"context"

	"github.com/egannguyen/go-kafka-ecommerce/backend/internal/entity"
)

// ProductRepository handles persistence for Products.
type ProductRepository interface {
	FindAll(ctx context.Context) ([]entity.Product, error)
	// Seed inserts initial products if none exist.
	Seed(ctx context.Context, products []entity.Product) error
}

// OrderRepository handles persistence for Orders.
type OrderRepository interface {
	PlaceOrder(ctx context.Context, cmd *entity.PlaceOrder) (*entity.OrderPlaced, error) // Deprecated
	ConfirmOrder(ctx context.Context, orderID string) error                              // Deprecated
	UpdateOrderProjection(ctx context.Context, event entity.Event) error
	FindRecent(ctx context.Context, limit int) ([]entity.Order, error)
}

// EventStore handles appending and loading events for an aggregate stream.
type EventStore interface {
	SaveEvents(ctx context.Context, streamID string, streamType string, expectedVersion int, events []entity.Event) error
	LoadEvents(ctx context.Context, streamID string) ([]entity.EventStoreRecord, error)
}
