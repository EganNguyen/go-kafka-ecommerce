package entity

import "time"

// EventStoreRecord represents an event stored in the database.
type EventStoreRecord struct {
	ID         string    `json:"id"`
	StreamID   string    `json:"stream_id"`
	StreamType string    `json:"stream_type"`
	Version    int       `json:"version"`
	EventType  string    `json:"event_type"`
	Payload    []byte    `json:"payload"`
	CreatedAt  time.Time `json:"created_at"`
}

// Event represents a domain event.
type Event interface {
	EventType() string
}

// Aggregate represents a domain aggregate root.
type Aggregate interface {
	GetAggregateID() string
	GetVersion() int
	ApplyEvent(event Event) error
}

// AggregateBase provides a basic implementation for an aggregate.
type AggregateBase struct {
	ID      string
	Version int
}

func (a *AggregateBase) GetAggregateID() string {
	return a.ID
}

func (a *AggregateBase) GetVersion() int {
	return a.Version
}
