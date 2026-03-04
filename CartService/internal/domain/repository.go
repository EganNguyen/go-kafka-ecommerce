package domain

import "context"

type CartRepository interface {
	Save(ctx context.Context, cart *CartAggregate) error
	Get(ctx context.Context, cartID string) (*CartAggregate, error)
	Delete(ctx context.Context, cartID string) error
}

// EventStore defines the interface for persisting and loading events.
type EventStore interface {
	SaveEvents(ctx context.Context, aggregateID string, aggregateType string, expectedVersion int, events []Event) error
	LoadEvents(ctx context.Context, aggregateID string) ([]EventRecord, error)
}
