package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

type Money struct {
	CurrencyCode string  `json:"currency_code" bson:"currency_code"`
	Units        int64   `json:"units" bson:"units"`
	Nanos        int32   `json:"nanos" bson:"nanos"`
}

// Product represents a product in the store.
type Product struct {
	ID          string  `json:"id" bson:"id"`
	Name        string  `json:"name" bson:"name"`
	Description string  `json:"description" bson:"description"`
	Price       float64 `json:"price" bson:"price"`
	ImageURL    string  `json:"image_url" bson:"image_url"`
	Category    string  `json:"category" bson:"category"`
	Stock       int     `json:"stock" bson:"stock"`
}

// OrderItem is a line item within an order.
type OrderItem struct {
	ProductID string  `json:"product_id" bson:"product_id"`
	Name      string  `json:"name" bson:"name"`
	Price     float64 `json:"price" bson:"price"`
	Quantity  int     `json:"quantity" bson:"quantity"`
}

// Order represents a customer order.
type Order struct {
	ID         string      `json:"id" bson:"id"`
	Items      []OrderItem `json:"items" bson:"items"`
	TotalPrice float64     `json:"total_price" bson:"total_price"`
	Status     string      `json:"status" bson:"status"` // "placed", "confirmed", "shipped"
	CreatedAt  time.Time   `json:"created_at" bson:"created_at"`
}

// --- Commands ---

// PlaceOrder is a command to create a new order.
type PlaceOrder struct {
	OrderID string      `json:"order_id"`
	Items   []OrderItem `json:"items"`
}

// --- Events ---

type Event interface {
	EventType() string
}

type OrderPlaced struct {
	OrderID    string      `json:"order_id"`
	Items      []OrderItem `json:"items"`
	TotalPrice float64     `json:"total_price"`
	PlacedAt   time.Time   `json:"placed_at"`
}

func (e OrderPlaced) EventType() string { return "OrderPlaced" }

type OrderConfirmed struct {
	OrderID     string    `json:"order_id"`
	ConfirmedAt time.Time `json:"confirmed_at"`
}

func (e OrderConfirmed) EventType() string { return "OrderConfirmed" }

type ProductStockUpdated struct {
	ProductID string `json:"product_id"`
	NewStock  int    `json:"new_stock"`
}

func (e ProductStockUpdated) EventType() string { return "ProductStockUpdated" }

type InventoryReserved struct {
	OrderID   string `json:"order_id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

func (e InventoryReserved) EventType() string { return "InventoryReserved" }

type ReservationReleased struct {
	OrderID   string `json:"order_id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

func (e ReservationReleased) EventType() string { return "ReservationReleased" }

type ReservationConfirmed struct {
	OrderID   string `json:"order_id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

func (e ReservationConfirmed) EventType() string { return "ReservationConfirmed" }

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
	a.Version++
	return nil
}

// Rehydrate rebuilds the aggregate from a list of records.
func (a *OrderAggregate) Rehydrate(records []EventRecord) error {
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
	return nil
}

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
func (a *InventoryAggregate) Rehydrate(records []EventRecord) error {
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
