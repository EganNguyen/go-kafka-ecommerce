package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/egannguyen/go-kafka-ecommerce/backend/internal/entity"
	"github.com/egannguyen/go-kafka-ecommerce/backend/internal/repository"
)

type orderRepository struct {
	db *sql.DB
}

// NewOrderRepository creates a new OrderRepository backed by Postgres.
func NewOrderRepository(db *sql.DB) repository.OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) PlaceOrder(ctx context.Context, cmd *entity.PlaceOrder) (*entity.OrderPlaced, error) {
	var totalPrice float64
	for _, item := range cmd.Items {
		totalPrice += item.Price * float64(item.Quantity)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Idempotency check
	var alreadyExists bool
	err = tx.QueryRowContext(ctx,
		"INSERT INTO orders (id, total_price, status, created_at) VALUES ($1, $2, $3, $4) ON CONFLICT (id) DO NOTHING RETURNING true",
		cmd.OrderID, totalPrice, "placed", time.Now(),
	).Scan(&alreadyExists)

	if err == sql.ErrNoRows {
		// Already exists, just return nil indicating it was handled
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

		// Decrement stock
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

	event := &entity.OrderPlaced{
		OrderID:    cmd.OrderID,
		Items:      cmd.Items,
		TotalPrice: totalPrice,
		PlacedAt:   time.Now(),
	}
	return event, nil
}

func (r *orderRepository) ConfirmOrder(ctx context.Context, orderID string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE orders SET status = 'confirmed' WHERE id = $1",
		orderID,
	)
	if err != nil {
		return fmt.Errorf("failed to confirm order: %w", err)
	}
	return nil
}

func (r *orderRepository) FindRecent(ctx context.Context, limit int) ([]entity.Order, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, total_price, status, created_at FROM orders ORDER BY created_at DESC LIMIT $1", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	var orders []entity.Order
	for rows.Next() {
		var o entity.Order
		if err := rows.Scan(&o.ID, &o.TotalPrice, &o.Status, &o.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, o)
	}

	// Fetch items for each order
	for i := range orders {
		itemRows, err := r.db.QueryContext(ctx,
			"SELECT product_id, name, price, quantity FROM order_items WHERE order_id = $1",
			orders[i].ID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to query order items: %w", err)
		}

		for itemRows.Next() {
			var item entity.OrderItem
			if err := itemRows.Scan(&item.ProductID, &item.Name, &item.Price, &item.Quantity); err != nil {
				itemRows.Close()
				return nil, fmt.Errorf("failed to scan order item: %w", err)
			}
			orders[i].Items = append(orders[i].Items, item)
		}
		itemRows.Close()
	}

	return orders, nil
}
