package messaging

import "context"

// Publisher defines an interface for publishing events to a message broker.
type Publisher interface {
	PublishEvent(ctx context.Context, topic string, key string, event any) error
}

// Subscriber defines an interface for subscribing to a message topic.
type Subscriber interface {
	Consume(ctx context.Context, topic string, groupID string, handler func(ctx context.Context, payload []byte) error)
}
