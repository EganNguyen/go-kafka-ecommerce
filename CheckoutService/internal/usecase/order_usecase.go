package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/egannguyen/go-kafka-ecommerce/checkout-service/internal/domain"
)

type CheckoutUseCase interface {
	GetProducts(ctx context.Context) ([]domain.Product, error)
	GetRecentOrders(ctx context.Context, limit int) ([]domain.Order, error)
	PlaceOrder(ctx context.Context, cmd *domain.PlaceOrder) error
	HandleOrderPlaced(ctx context.Context, event *domain.OrderPlaced) error
	HandleOrderConfirmed(ctx context.Context, event *domain.OrderConfirmed) error
}

type checkoutUseCase struct {
	orderRepo   domain.OrderRepository
	productRepo domain.ProductRepository
	eventStore  domain.EventStore
	publisher   domain.Publisher
}

func NewCheckoutUseCase(orderRepo domain.OrderRepository, productRepo domain.ProductRepository, eventStore domain.EventStore, publisher domain.Publisher) CheckoutUseCase {
	return &checkoutUseCase{
		orderRepo:   orderRepo,
		productRepo: productRepo,
		eventStore:  eventStore,
		publisher:   publisher,
	}
}

func (u *checkoutUseCase) GetProducts(ctx context.Context) ([]domain.Product, error) {
	return u.productRepo.FindAll(ctx)
}

func (u *checkoutUseCase) GetRecentOrders(ctx context.Context, limit int) ([]domain.Order, error) {
	if limit <= 0 {
		limit = 50
	}
	return u.orderRepo.FindRecent(ctx, limit)
}

func (u *checkoutUseCase) PlaceOrder(ctx context.Context, cmd *domain.PlaceOrder) error {
	slog.Info("UseCase: Placing order", "order_id", cmd.OrderID, "items", len(cmd.Items))

	if len(cmd.Items) == 0 {
		return fmt.Errorf("order must have at least one item")
	}

	records, err := u.eventStore.LoadEvents(ctx, cmd.OrderID)
	if err != nil {
		return fmt.Errorf("failed to load order history: %w", err)
	}

	if len(records) > 0 {
		slog.Info("Order already exists (idempotency)", "order_id", cmd.OrderID)
		return nil
	}

	for _, item := range cmd.Items {
		invRecords, err := u.eventStore.LoadEvents(ctx, item.ProductID)
		if err != nil {
			return fmt.Errorf("failed to load inventory history for %s: %w", item.ProductID, err)
		}

		invAgg := domain.NewInventoryAggregate(item.ProductID)
		// Simulating legacy seed for stock check if no events exist
		if len(invRecords) == 0 {
			prods, _ := u.productRepo.FindAll(ctx)
			for _, p := range prods {
				if p.ID == item.ProductID {
					invAgg.HardStock = p.Stock
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

		resEvent := domain.InventoryReserved{
			OrderID:   cmd.OrderID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
		err = u.eventStore.SaveEvents(ctx, item.ProductID, "inventory", -1, []domain.Event{resEvent})
		if err != nil {
			slog.Error("Failed to save InventoryReserved event", "err", err)
		}
	}

	placedEvent := domain.OrderPlaced{
		OrderID:    cmd.OrderID,
		Items:      cmd.Items,
		TotalPrice: totalPrice,
		PlacedAt:   time.Now(),
	}

	err = u.eventStore.SaveEvents(ctx, cmd.OrderID, "order", 0, []domain.Event{placedEvent})
	if err != nil {
		return fmt.Errorf("failed to save OrderPlaced event: %w", err)
	}

	if err := u.publisher.PublishEvent(ctx, "orders.placed", cmd.OrderID, placedEvent); err != nil {
		return fmt.Errorf("failed to publish OrderPlaced event: %w", err)
	}

	return nil
}

func (u *checkoutUseCase) HandleOrderPlaced(ctx context.Context, event *domain.OrderPlaced) error {
	slog.Info("UseCase: Confirming order", "order_id", event.OrderID)

	if err := u.orderRepo.UpdateOrderProjection(ctx, *event); err != nil {
		slog.Error("Failed to update projection for OrderPlaced", "err", err)
	}

	records, err := u.eventStore.LoadEvents(ctx, event.OrderID)
	if err != nil {
		return fmt.Errorf("failed to load order events: %w", err)
	}

	aggregate := domain.NewOrderAggregate(event.OrderID)
	if err := aggregate.Rehydrate(records); err != nil {
		return fmt.Errorf("failed to rehydrate order aggregate: %w", err)
	}

	if aggregate.Status == "confirmed" {
		return nil
	}

	confirmedEvent := domain.OrderConfirmed{
		OrderID:     event.OrderID,
		ConfirmedAt: time.Now(),
	}

	err = u.eventStore.SaveEvents(ctx, event.OrderID, "order", aggregate.GetVersion(), []domain.Event{confirmedEvent})
	if err != nil {
		return fmt.Errorf("failed to save OrderConfirmed event: %w", err)
	}

	return u.publisher.PublishEvent(ctx, "orders.confirmed", event.OrderID, confirmedEvent)
}

func (u *checkoutUseCase) HandleOrderConfirmed(ctx context.Context, event *domain.OrderConfirmed) error {
	return u.orderRepo.UpdateOrderProjection(ctx, *event)
}
