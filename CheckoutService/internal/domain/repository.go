package domain

import "context"

type OrderRepository interface {
	FindRecent(ctx context.Context, limit int) ([]Order, error)
	UpdateOrderProjection(ctx context.Context, event interface{}) error
}


type EventStore interface {
	SaveEvents(ctx context.Context, aggregateID string, aggregateType string, expectedVersion int, events []Event) error
	LoadEvents(ctx context.Context, aggregateID string) ([]EventRecord, error)
}

type Publisher interface {
	PublishEvent(ctx context.Context, topic string, key string, event interface{}) error
}

type Subscriber interface {
	Consume(ctx context.Context, topic string, groupID string, handler func(context.Context, []byte) error) error
}
