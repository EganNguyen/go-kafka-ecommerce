package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/egannguyen/go-kafka-ecommerce/backend/internal/entity"
	"github.com/egannguyen/go-kafka-ecommerce/backend/internal/messaging"
	"github.com/egannguyen/go-kafka-ecommerce/backend/internal/repository"
)

// OrderService orchestrates order-related business logic.
type OrderService struct {
	orderRepo   repository.OrderRepository
	productRepo repository.ProductRepository
	publisher   messaging.Publisher
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	productRepo repository.ProductRepository,
	publisher messaging.Publisher,
) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		productRepo: productRepo,
		publisher:   publisher,
	}
}

// GetProducts returns all available products.
func (s *OrderService) GetProducts(ctx context.Context) ([]entity.Product, error) {
	return s.productRepo.FindAll(ctx)
}

// GetRecentOrders returns the latest orders.
func (s *OrderService) GetRecentOrders(ctx context.Context, limit int) ([]entity.Order, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.orderRepo.FindRecent(ctx, limit)
}

// PlaceOrder initiates the order placement process.
func (s *OrderService) PlaceOrder(ctx context.Context, cmd *entity.PlaceOrder) error {
	slog.Info("Service: Placing order", "order_id", cmd.OrderID, "items", len(cmd.Items))

	// Validate command
	if len(cmd.Items) == 0 {
		return fmt.Errorf("order must have at least one item")
	}

	// 1. Save order to DB and deduct stock within a transaction
	event, err := s.orderRepo.PlaceOrder(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to save order to repository: %w", err)
	}

	// If event is nil, it means it was an idempotent skip (already processed)
	if event == nil {
		slog.Info("Order already processed (idempotent)", "order_id", cmd.OrderID)
		return nil
	}

	// 2. Publish OrderPlaced event
	if err := s.publisher.PublishEvent(ctx, "orders.placed", event.OrderID, event); err != nil {
		return fmt.Errorf("failed to publish OrderPlaced event: %w", err)
	}

	return nil
}

// HandleOrderPlaced is triggered by the message broker when an order is placed.
func (s *OrderService) HandleOrderPlaced(ctx context.Context, event *entity.OrderPlaced) error {
	slog.Info("Service: Handling OrderPlaced event", "order_id", event.OrderID)

	if err := s.orderRepo.ConfirmOrder(ctx, event.OrderID); err != nil {
		return fmt.Errorf("failed to confirm order in repository: %w", err)
	}

	slog.Info("✅ Order confirmed", "order_id", event.OrderID)
	return nil
}
