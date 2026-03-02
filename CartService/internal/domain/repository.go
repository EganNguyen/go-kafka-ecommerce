package domain

import "context"

// EventStore defines the interface for persisting and loading events.
type EventStore interface {
	SaveEvents(ctx context.Context, aggregateID string, aggregateType string, expectedVersion int, events []Event) error
	LoadEvents(ctx context.Context, aggregateID string) ([]EventRecord, error)
}
