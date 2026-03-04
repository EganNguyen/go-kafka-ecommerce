package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

// Event represents a domain event.
type Event interface {
	EventType() string
}

// ItemAddedToCart is emitted when a user drops an item into their cart.
type ItemAddedToCart struct {
	CartID    string  `json:"cart_id" bson:"cart_id"`
	ProductID string  `json:"product_id" bson:"product_id"`
	Quantity  int     `json:"quantity" bson:"quantity"`
	Price     float64 `json:"price" bson:"price"`
}

func (e ItemAddedToCart) EventType() string { return "ItemAddedToCart" }

// ItemRemovedFromCart is emitted when a user removes an item or it's bought.
type ItemRemovedFromCart struct {
	CartID    string `json:"cart_id" bson:"cart_id"`
	ProductID string `json:"product_id" bson:"product_id"`
	Quantity  int    `json:"quantity" bson:"quantity"`
}

func (e ItemRemovedFromCart) EventType() string { return "ItemRemovedFromCart" }

// CartItem represents an item in the cart.
type CartItem struct {
	ProductID string  `json:"product_id" bson:"product_id"`
	Quantity  int     `json:"quantity" bson:"quantity"`
	Price     float64 `json:"price" bson:"price"`
}

// EventRecord represents an event stored in the event store.
type EventRecord struct {
	ID            string    `json:"id" bson:"id"`
	StreamID      string    `json:"stream_id" bson:"stream_id"`
	StreamType    string    `json:"stream_type" bson:"stream_type"`
	Version       int       `json:"version" bson:"version"`
	EventType     string    `json:"event_type" bson:"event_type"`
	Payload       []byte    `json:"payload" bson:"payload"`
	CreatedAt     time.Time `json:"created_at" bson:"created_at"`
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

// CartAggregate manages the state of a shopping cart by replaying events.
type CartAggregate struct {
	AggregateBase
	Items map[string]*CartItem
}

// NewCartAggregate creates a new CartAggregate.
func NewCartAggregate(cartID string) *CartAggregate {
	return &CartAggregate{
		AggregateBase: AggregateBase{ID: cartID, Version: 0},
		Items:         make(map[string]*CartItem),
	}
}

// ApplyEvent mutates the aggregate state based on the event.
func (a *CartAggregate) ApplyEvent(e Event) error {
	switch e := e.(type) {
	case ItemAddedToCart:
		if item, exists := a.Items[e.ProductID]; exists {
			item.Quantity += e.Quantity
		} else {
			a.Items[e.ProductID] = &CartItem{
				ProductID: e.ProductID,
				Quantity:  e.Quantity,
				Price:     e.Price,
			}
		}
	case ItemRemovedFromCart:
		if item, exists := a.Items[e.ProductID]; exists {
			item.Quantity -= e.Quantity
			if item.Quantity <= 0 {
				delete(a.Items, e.ProductID)
			}
		}
	default:
		return fmt.Errorf("unknown event type for CartAggregate: %s", e.EventType())
	}
	a.Version++
	return nil
}

// Rehydrate rebuilds the aggregate from a list of records.
func (a *CartAggregate) Rehydrate(records []EventRecord) error {
	for _, rec := range records {
		var err error
		switch rec.EventType {
		case "ItemAddedToCart":
			var e ItemAddedToCart
			if err = json.Unmarshal(rec.Payload, &e); err == nil {
				err = a.ApplyEvent(e)
			}
		case "ItemRemovedFromCart":
			var e ItemRemovedFromCart
			if err = json.Unmarshal(rec.Payload, &e); err == nil {
				err = a.ApplyEvent(e)
			}
		default:
			return fmt.Errorf("unknown event type in cart stream: %s", rec.EventType)
		}
		if err != nil {
			return fmt.Errorf("failed to apply cart event from stream: %w", err)
		}
	}
	return nil
}

