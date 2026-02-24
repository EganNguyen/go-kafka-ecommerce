package entity

import (
	"encoding/json"
	"fmt"
	"time"
)

// OrderAggregate manages the state of an Order by replaying events.
type OrderAggregate struct {
	AggregateBase
	Items      []OrderItem
	TotalPrice float64
	Status     string
	CreatedAt  time.Time
}

// NewOrderAggregate creates a new OrderAggregate from history.
func NewOrderAggregate(id string) *OrderAggregate {
	return &OrderAggregate{
		AggregateBase: AggregateBase{ID: id, Version: 0},
		Status:        "pending",
	}
}

// ApplyEvent mutates the aggregate state based on the event.
func (a *OrderAggregate) ApplyEvent(e Event) error {
	switch e := e.(type) {
	case OrderPlaced:
		a.Items = e.Items
		a.TotalPrice = e.TotalPrice
		a.Status = "placed"
		if a.CreatedAt.IsZero() {
			a.CreatedAt = e.PlacedAt
		}
	case OrderConfirmed:
		a.Status = "confirmed"
	default:
		return fmt.Errorf("unknown event type for OrderAggregate: %s", e.EventType())
	}
	// Increment version after applying
	a.Version++
	return nil
}

// Rehydrate rebuilds the aggregate from a list of records.
func (a *OrderAggregate) Rehydrate(records []EventStoreRecord) error {
	for _, rec := range records {
		var err error
		switch rec.EventType {
		case "OrderPlaced":
			var e OrderPlaced
			if err = json.Unmarshal(rec.Payload, &e); err == nil {
				err = a.ApplyEvent(e)
			}
		case "OrderConfirmed":
			var e OrderConfirmed
			if err = json.Unmarshal(rec.Payload, &e); err == nil {
				err = a.ApplyEvent(e)
			}
		default:
			return fmt.Errorf("unknown event type in stream: %s", rec.EventType)
		}
		if err != nil {
			return fmt.Errorf("failed to apply event from stream: %w", err)
		}
	}
	// Version is updated in ApplyEvent
	return nil
}
