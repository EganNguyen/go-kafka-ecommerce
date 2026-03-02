package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/egannguyen/go-kafka-ecommerce/checkout-service/internal/domain"
)

type orderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) domain.OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) UpdateOrderProjection(ctx context.Context, event interface{}) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	switch e := event.(type) {
	case domain.OrderPlaced:
		_, err = tx.ExecContext(ctx,
			"INSERT INTO orders (id, total_price, status, created_at) VALUES ($1, $2, $3, $4) ON CONFLICT (id) DO NOTHING",
			e.OrderID, e.TotalPrice, "placed", e.PlacedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert order projection: %w", err)
		}

		for _, item := range e.Items {
			_, err = tx.ExecContext(ctx,
				"INSERT INTO order_items (order_id, product_id, name, price, quantity) VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING",
				e.OrderID, item.ProductID, item.Name, item.Price, item.Quantity,
			)
			if err != nil {
				return fmt.Errorf("failed to insert order item projection: %w", err)
			}

			_, err = tx.ExecContext(ctx,
				"UPDATE products SET stock = stock - $1 WHERE id = $2 AND stock >= $1",
				item.Quantity, item.ProductID,
			)
			if err != nil {
				return fmt.Errorf("failed to update product stock: %w", err)
			}
		}

	case domain.OrderConfirmed:
		_, err = tx.ExecContext(ctx,
			"UPDATE orders SET status = 'confirmed' WHERE id = $1",
			e.OrderID,
		)
		if err != nil {
			return fmt.Errorf("failed to confirm order projection: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *orderRepository) FindRecent(ctx context.Context, limit int) ([]domain.Order, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, total_price, status, created_at FROM orders ORDER BY created_at DESC LIMIT $1", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.TotalPrice, &o.Status, &o.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, o)
	}

	for i := range orders {
		itemRows, err := r.db.QueryContext(ctx,
			"SELECT product_id, name, price, quantity FROM order_items WHERE order_id = $1",
			orders[i].ID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to query order items: %w", err)
		}

		for itemRows.Next() {
			var item domain.OrderItem
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
