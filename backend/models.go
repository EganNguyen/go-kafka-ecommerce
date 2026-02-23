package main

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

// OrderPlaced is emitted when an order is successfully placed.
type OrderPlaced struct {
	OrderID    string      `json:"order_id"`
	Items      []OrderItem `json:"items"`
	TotalPrice float64     `json:"total_price"`
	PlacedAt   time.Time   `json:"placed_at"`
}

// OrderConfirmed is emitted when an order is confirmed (e.g., payment verified).
type OrderConfirmed struct {
	OrderID     string    `json:"order_id"`
	ConfirmedAt time.Time `json:"confirmed_at"`
}

// ProductStockUpdated is emitted when product stock changes due to an order.
type ProductStockUpdated struct {
	ProductID string `json:"product_id"`
	NewStock  int    `json:"new_stock"`
}
