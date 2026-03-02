package entity

import (
	"encoding/json"
	"fmt"
)

// CartItem represents an item currently in a user's cart.
type CartItem struct {
	ProductID string
	Quantity  int
	Price     float64
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
func (a *CartAggregate) Rehydrate(records []EventStoreRecord) error {
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
