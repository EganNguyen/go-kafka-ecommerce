package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	kafkaGo "github.com/segmentio/kafka-go"
)

// PlaceOrderHandler handles the PlaceOrder command.
type PlaceOrderHandler struct {
	db          *sql.DB
	eventWriter *kafkaGo.Writer // publishes OrderPlaced events
}

func (h *PlaceOrderHandler) Handle(ctx context.Context, cmd *PlaceOrder) (*OrderPlaced, error) {
	slog.Info("Handling PlaceOrder command", "order_id", cmd.OrderID, "items", len(cmd.Items))

	var totalPrice float64
	for _, item := range cmd.Items {
		totalPrice += item.Price * float64(item.Quantity)
	}

	// Insert order into the database.
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Use ON CONFLICT for idempotency — if Kafka redelivers the message,
	// we skip the insert instead of crashing with a duplicate key error.
	var alreadyExists bool
	err = tx.QueryRowContext(ctx,
		"INSERT INTO orders (id, total_price, status, created_at) VALUES ($1, $2, $3, $4) ON CONFLICT (id) DO NOTHING RETURNING true",
		cmd.OrderID, totalPrice, "placed", time.Now(),
	).Scan(&alreadyExists)
	if err == sql.ErrNoRows {
		slog.Info("Order already exists, skipping (idempotent)", "order_id", cmd.OrderID)
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to insert order: %w", err)
	}

	for _, item := range cmd.Items {
		_, err = tx.ExecContext(ctx,
			"INSERT INTO order_items (order_id, product_id, name, price, quantity) VALUES ($1, $2, $3, $4, $5)",
			cmd.OrderID, item.ProductID, item.Name, item.Price, item.Quantity,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert order item: %w", err)
		}

		// Decrement stock.
		_, err = tx.ExecContext(ctx,
			"UPDATE products SET stock = stock - $1 WHERE id = $2 AND stock >= $1",
			item.Quantity, item.ProductID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update product stock: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	slog.Info("✅ Order saved to database", "order_id", cmd.OrderID, "total", totalPrice)

	event := &OrderPlaced{
		OrderID:    cmd.OrderID,
		Items:      cmd.Items,
		TotalPrice: totalPrice,
		PlacedAt:   time.Now(),
	}
	return event, nil
}

// OrderPlacedHandler handles the OrderPlaced event (confirms the order).
type OrderPlacedHandler struct {
	db *sql.DB
}

func (h *OrderPlacedHandler) Handle(ctx context.Context, event *OrderPlaced) error {
	slog.Info("📦 Order placed event received!",
		"order_id", event.OrderID,
		"total_price", event.TotalPrice,
		"items_count", len(event.Items),
	)

	_, err := h.db.ExecContext(ctx,
		"UPDATE orders SET status = 'confirmed' WHERE id = $1",
		event.OrderID,
	)
	if err != nil {
		return fmt.Errorf("failed to confirm order: %w", err)
	}

	slog.Info("✅ Order confirmed", "order_id", event.OrderID)
	return nil
}

// publishEvent is a helper to JSON-encode and publish an event to Kafka.
func publishEvent(ctx context.Context, w *kafkaGo.Writer, key string, event any) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	return publishJSON(ctx, w, []byte(key), payload)
}
