package entity

import (
	"time"
)

// Product represents a product in the store.
type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	ImageURL    string  `json:"image_url"`
	Category    string  `json:"category"`
	Stock       int     `json:"stock"`
}

// OrderItem is a line item within an order.
type OrderItem struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
}

// Order represents a customer order.
type Order struct {
	ID         string      `json:"id"`
	Items      []OrderItem `json:"items"`
	TotalPrice float64     `json:"total_price"`
	Status     string      `json:"status"` // "placed", "confirmed", "shipped"
	CreatedAt  time.Time   `json:"created_at"`
}

// --- Commands ---

// PlaceOrder is a command to create a new order.
type PlaceOrder struct {
	OrderID string      `json:"order_id"`
	Items   []OrderItem `json:"items"`
}

// --- Events ---

// OrderPlaced is emitted when an order is successfully placed (Command received).
type OrderPlaced struct {
	OrderID    string      `json:"order_id"`
	Items      []OrderItem `json:"items"`
	TotalPrice float64     `json:"total_price"`
	PlacedAt   time.Time   `json:"placed_at"`
}

func (e OrderPlaced) EventType() string { return "OrderPlaced" }

// OrderConfirmed is emitted when an order is confirmed (e.g., payment verified).
type OrderConfirmed struct {
	OrderID     string    `json:"order_id"`
	ConfirmedAt time.Time `json:"confirmed_at"`
}

func (e OrderConfirmed) EventType() string { return "OrderConfirmed" }

// ProductStockUpdated is emitted when product stock changes due to an order.
type ProductStockUpdated struct {
	ProductID string `json:"product_id"`
	NewStock  int    `json:"new_stock"`
}

func (e ProductStockUpdated) EventType() string { return "ProductStockUpdated" }

// InventoryReserved is emitted when stock is soft-locked for a pending order.
type InventoryReserved struct {
	OrderID   string `json:"order_id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

func (e InventoryReserved) EventType() string { return "InventoryReserved" }

// ReservationReleased is emitted if an order is cancelled or fails, unlocking the stock.
type ReservationReleased struct {
	OrderID   string `json:"order_id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

func (e ReservationReleased) EventType() string { return "ReservationReleased" }

// ReservationConfirmed is emitted when an order is finalized, turning soft-lock into hard-deduction.
type ReservationConfirmed struct {
	OrderID   string `json:"order_id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

func (e ReservationConfirmed) EventType() string { return "ReservationConfirmed" }

// ItemAddedToCart is emitted when a user drops an item into their cart.
type ItemAddedToCart struct {
	CartID    string  `json:"cart_id"`
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

func (e ItemAddedToCart) EventType() string { return "ItemAddedToCart" }

// ItemRemovedFromCart is emitted when a user removes an item or it's bought.
type ItemRemovedFromCart struct {
	CartID    string `json:"cart_id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

func (e ItemRemovedFromCart) EventType() string { return "ItemRemovedFromCart" }
