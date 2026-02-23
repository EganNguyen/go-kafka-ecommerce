package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
)

// PlaceOrderHandler handles the PlaceOrder command.
type PlaceOrderHandler struct {
	db       *sql.DB
	eventBus *cqrs.EventBus
}

func (h PlaceOrderHandler) HandlerName() string {
	return "PlaceOrderHandler"
}

func (h PlaceOrderHandler) NewCommand() interface{} {
	return &PlaceOrder{}
}

func (h PlaceOrderHandler) Handle(ctx context.Context, cmd interface{}) error {
	placeOrder := cmd.(*PlaceOrder)

	slog.Info("Handling PlaceOrder command", "order_id", placeOrder.OrderID, "items", len(placeOrder.Items))

	var totalPrice float64
	for _, item := range placeOrder.Items {
		totalPrice += item.Price * float64(item.Quantity)
	}

	// Insert order into the database.
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Use ON CONFLICT for idempotency — if Kafka redelivers the message,
	// we skip the insert instead of crashing with a duplicate key error.
	var alreadyExists bool
	err = tx.QueryRowContext(ctx,
		"INSERT INTO orders (id, total_price, status, created_at) VALUES ($1, $2, $3, $4) ON CONFLICT (id) DO NOTHING RETURNING true",
		placeOrder.OrderID, totalPrice, "placed", time.Now(),
	).Scan(&alreadyExists)
	if err == sql.ErrNoRows {
		// ON CONFLICT DO NOTHING — order already exists, skip.
		slog.Info("Order already exists, skipping (idempotent)", "order_id", placeOrder.OrderID)
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	for _, item := range placeOrder.Items {
		_, err = tx.ExecContext(ctx,
			"INSERT INTO order_items (order_id, product_id, name, price, quantity) VALUES ($1, $2, $3, $4, $5)",
			placeOrder.OrderID, item.ProductID, item.Name, item.Price, item.Quantity,
		)
		if err != nil {
			return fmt.Errorf("failed to insert order item: %w", err)
		}

		// Decrement stock.
		_, err = tx.ExecContext(ctx,
			"UPDATE products SET stock = stock - $1 WHERE id = $2 AND stock >= $1",
			item.Quantity, item.ProductID,
		)
		if err != nil {
			return fmt.Errorf("failed to update product stock: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	slog.Info("✅ Order saved to database", "order_id", placeOrder.OrderID, "total", totalPrice)

	// Publish OrderPlaced event if eventBus is configured.
	if h.eventBus != nil {
		event := &OrderPlaced{
			OrderID:    placeOrder.OrderID,
			Items:      placeOrder.Items,
			TotalPrice: totalPrice,
			PlacedAt:   time.Now(),
		}
		return h.eventBus.Publish(ctx, event)
	}

	return nil
}

// OrderPlacedHandler handles the OrderPlaced event (e.g., sends notifications, confirms order).
type OrderPlacedHandler struct {
	db *sql.DB
}

func (h OrderPlacedHandler) HandlerName() string {
	return "OrderPlacedHandler"
}

func (h OrderPlacedHandler) NewEvent() interface{} {
	return &OrderPlaced{}
}

func (h OrderPlacedHandler) Handle(ctx context.Context, event interface{}) error {
	orderPlaced := event.(*OrderPlaced)

	slog.Info("📦 Order placed event received!",
		"order_id", orderPlaced.OrderID,
		"total_price", orderPlaced.TotalPrice,
		"items_count", len(orderPlaced.Items),
	)

	// Simulate confirming the order (e.g. after payment verification).
	_, err := h.db.ExecContext(ctx,
		"UPDATE orders SET status = 'confirmed' WHERE id = $1",
		orderPlaced.OrderID,
	)
	if err != nil {
		return fmt.Errorf("failed to confirm order: %w", err)
	}

	slog.Info("✅ Order confirmed", "order_id", orderPlaced.OrderID)
	return nil
}

// marshalCommand marshals a command into a Watermill message using JSON.
func marshalCommand(cmd interface{}) (*message.Message, error) {
	payload, err := json.Marshal(cmd)
	if err != nil {
		return nil, err
	}

	msg := message.NewMessage(uuid.New().String(), payload)
	return msg, nil
}
