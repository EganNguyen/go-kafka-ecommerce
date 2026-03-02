package entity

import (
	"encoding/json"
	"fmt"
)

// InventoryAggregate manages the stock of a product by replaying events.
type InventoryAggregate struct {
	AggregateBase
	HardStock     int // Total physical items
	ReservedStock int // Items locked for pending orders
}

// AvailableStock returns the stock available for new reservations.
func (a *InventoryAggregate) AvailableStock() int {
	return a.HardStock - a.ReservedStock
}

// NewInventoryAggregate creates a new InventoryAggregate.
func NewInventoryAggregate(productID string) *InventoryAggregate {
	return &InventoryAggregate{
		AggregateBase: AggregateBase{ID: productID, Version: 0},
	}
}

// ApplyEvent mutates the aggregate state based on the event.
func (a *InventoryAggregate) ApplyEvent(e Event) error {
	switch e := e.(type) {
	case ProductStockUpdated:
		// Used for initial loading or manual adjustments
		a.HardStock = e.NewStock
	case InventoryReserved:
		a.ReservedStock += e.Quantity
	case ReservationReleased:
		a.ReservedStock -= e.Quantity
	case ReservationConfirmed:
		a.ReservedStock -= e.Quantity
		a.HardStock -= e.Quantity
	default:
		return fmt.Errorf("unknown event type for InventoryAggregate: %s", e.EventType())
	}
	a.Version++
	return nil
}

// Rehydrate rebuilds the aggregate from a list of records.
func (a *InventoryAggregate) Rehydrate(records []EventStoreRecord) error {
	for _, rec := range records {
		var err error
		switch rec.EventType {
		case "ProductStockUpdated":
			var e ProductStockUpdated
			if err = json.Unmarshal(rec.Payload, &e); err == nil {
				err = a.ApplyEvent(e)
			}
		case "InventoryReserved":
			var e InventoryReserved
			if err = json.Unmarshal(rec.Payload, &e); err == nil {
				err = a.ApplyEvent(e)
			}
		case "ReservationReleased":
			var e ReservationReleased
			if err = json.Unmarshal(rec.Payload, &e); err == nil {
				err = a.ApplyEvent(e)
			}
		case "ReservationConfirmed":
			var e ReservationConfirmed
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
	return nil
}
