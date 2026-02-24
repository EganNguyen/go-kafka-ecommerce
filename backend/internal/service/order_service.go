package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/egannguyen/go-kafka-ecommerce/backend/internal/entity"
	"github.com/egannguyen/go-kafka-ecommerce/backend/internal/messaging"
	"github.com/egannguyen/go-kafka-ecommerce/backend/internal/repository"
)

// OrderService orchestrates order-related business logic.
type OrderService struct {
	orderRepo   repository.OrderRepository // Legacy Read Model repository
	productRepo repository.ProductRepository
	eventStore  repository.EventStore
	publisher   messaging.Publisher
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	productRepo repository.ProductRepository,
	eventStore repository.EventStore,
	publisher messaging.Publisher,
) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		productRepo: productRepo,
		eventStore:  eventStore,
		publisher:   publisher,
	}
}

// GetProducts returns all available products.
func (s *OrderService) GetProducts(ctx context.Context) ([]entity.Product, error) {
	return s.productRepo.FindAll(ctx)
}

// GetRecentOrders returns the latest orders.
func (s *OrderService) GetRecentOrders(ctx context.Context, limit int) ([]entity.Order, error) {
	// For now, this still queries the legacy order table (the projection / read model)
	if limit <= 0 {
		limit = 50
	}
	return s.orderRepo.FindRecent(ctx, limit)
}

// PlaceOrder initiates the order placement process.
func (s *OrderService) PlaceOrder(ctx context.Context, cmd *entity.PlaceOrder) error {
	slog.Info("Service: Placing order", "order_id", cmd.OrderID, "items", len(cmd.Items))

	if len(cmd.Items) == 0 {
		return fmt.Errorf("order must have at least one item")
	}

	// 1. Rehydrate aggregate (should be new)
	records, err := s.eventStore.LoadEvents(ctx, cmd.OrderID)
	if err != nil {
		return fmt.Errorf("failed to load order history: %w", err)
	}

	if len(records) > 0 {
		slog.Info("Order already exists (idempotency)", "order_id", cmd.OrderID)
		return nil
	}

	// 2. CHECK INVENTORY PRE-CONDITION
	for _, item := range cmd.Items {
		invRecords, err := s.eventStore.LoadEvents(ctx, item.ProductID)
		if err != nil {
			return fmt.Errorf("failed to load inventory history for %s: %w", item.ProductID, err)
		}

		invAgg := entity.NewInventoryAggregate(item.ProductID)
		// For backwards compatibility with legacy seed initialization during dev
		prod, err := s.productRepo.FindAll(ctx)
		if err == nil {
			for _, p := range prod {
				if p.ID == item.ProductID {
					invAgg.ApplyEvent(entity.ProductStockUpdated{ProductID: p.ID, NewStock: p.Stock})
				}
			}
		}

		if err := invAgg.Rehydrate(invRecords); err != nil {
			return fmt.Errorf("failed to rehydrate inventory aggregate: %w", err)
		}

		if invAgg.AvailableStock() < item.Quantity {
			return fmt.Errorf("insufficient stock for product %s (available: %d, requested: %d)", item.ProductID, invAgg.AvailableStock(), item.Quantity)
		}
	}

	var totalPrice float64
	for _, item := range cmd.Items {
		totalPrice += item.Price * float64(item.Quantity)

		// 3. GENERATE AND PERSIST INVENTORY RESERVATION EVENT
		resEvent := entity.InventoryReserved{
			OrderID:   cmd.OrderID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
		// In a real CQRS system this might happen in a downstream consumer. We do it here
		// to guarantee consistent stock reads immediately.
		err = s.eventStore.SaveEvents(ctx, item.ProductID, "inventory", -1, []entity.Event{resEvent})
		if err != nil {
			// -1 bypasses strict concurrency check here for simplicity, in prod use aggregate.GetVersion()
			slog.Error("Failed to save InventoryReserved event, proceeding anyway", "err", err)
		}
	}

	// 4. Generate Order Event
	placedEvent := entity.OrderPlaced{
		OrderID:    cmd.OrderID,
		Items:      cmd.Items,
		TotalPrice: totalPrice,
		PlacedAt:   time.Now(),
	}

	// 5. Persist Order Event
	err = s.eventStore.SaveEvents(ctx, cmd.OrderID, "order", 0, []entity.Event{placedEvent})
	if err != nil {
		return fmt.Errorf("failed to save OrderPlaced event: %w", err)
	}

	// 6. Publish Event to message broker for downstream consumers
	if err := s.publisher.PublishEvent(ctx, "orders.placed", cmd.OrderID, placedEvent); err != nil {
		return fmt.Errorf("failed to publish OrderPlaced event: %w", err)
	}

	return nil
}

// HandleOrderPlaced is triggered by the message broker when an order is placed.
func (s *OrderService) HandleOrderPlaced(ctx context.Context, event *entity.OrderPlaced) error {
	slog.Info("Service: Confirming order", "order_id", event.OrderID)

	// Update Read Model Projection (orders table)
	if err := s.orderRepo.UpdateOrderProjection(ctx, *event); err != nil {
		slog.Error("Failed to update projection for OrderPlaced", "err", err)
	}

	records, err := s.eventStore.LoadEvents(ctx, event.OrderID)
	if err != nil {
		return fmt.Errorf("failed to load order events: %w", err)
	}

	aggregate := entity.NewOrderAggregate(event.OrderID)
	if err := aggregate.Rehydrate(records); err != nil {
		return fmt.Errorf("failed to rehydrate order aggregate: %w", err)
	}

	if aggregate.Status == "confirmed" {
		slog.Info("Order already confirmed", "order_id", event.OrderID)
		return nil
	}

	confirmedEvent := entity.OrderConfirmed{
		OrderID:     event.OrderID,
		ConfirmedAt: time.Now(),
	}

	err = s.eventStore.SaveEvents(ctx, event.OrderID, "order", aggregate.GetVersion(), []entity.Event{confirmedEvent})
	if err != nil {
		return fmt.Errorf("failed to save OrderConfirmed event: %w", err)
	}

	// Publish confirmation so other systems (e.g. email) know
	if err := s.publisher.PublishEvent(ctx, "orders.confirmed", event.OrderID, confirmedEvent); err != nil {
		slog.Error("Failed to publish OrderConfirmed", "err", err)
	}

	slog.Info("✅ Order confirmed (Event Appended)", "order_id", event.OrderID)
	return nil
}

// HandleOrderConfirmed updates the read model when an order is confirmed.
func (s *OrderService) HandleOrderConfirmed(ctx context.Context, event *entity.OrderConfirmed) error {
	slog.Info("Projection: Updating OrderConfirmed", "order_id", event.OrderID)
	return s.orderRepo.UpdateOrderProjection(ctx, *event)
}
