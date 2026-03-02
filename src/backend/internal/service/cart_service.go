package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/egannguyen/go-kafka-ecommerce/backend/internal/entity"
	"github.com/egannguyen/go-kafka-ecommerce/backend/internal/repository"
)

// CartService orchestrates shopping cart logic using Event Sourcing.
type CartService struct {
	eventStore repository.EventStore
}

func NewCartService(eventStore repository.EventStore) *CartService {
	return &CartService{
		eventStore: eventStore,
	}
}

// AddItemToCart appends an ItemAddedToCart event to the user's cart stream.
func (s *CartService) AddItemToCart(ctx context.Context, cartID, productID string, quantity int, price float64) error {
	slog.Info("Service: Adding item to cart", "cart_id", cartID, "product_id", productID)

	records, err := s.eventStore.LoadEvents(ctx, cartID)
	if err != nil {
		return fmt.Errorf("failed to load cart history: %w", err)
	}

	agg := entity.NewCartAggregate(cartID)
	if err := agg.Rehydrate(records); err != nil {
		return fmt.Errorf("failed to rehydrate cart aggregate: %w", err)
	}

	event := entity.ItemAddedToCart{
		CartID:    cartID,
		ProductID: productID,
		Quantity:  quantity,
		Price:     price,
	}

	err = s.eventStore.SaveEvents(ctx, cartID, "cart", agg.GetVersion(), []entity.Event{event})
	if err != nil {
		return fmt.Errorf("failed to save ItemAddedToCart event: %w", err)
	}

	return nil
}

// GetCart loading the current state of a cart by replaying its events.
func (s *CartService) GetCart(ctx context.Context, cartID string) (*entity.CartAggregate, error) {
	records, err := s.eventStore.LoadEvents(ctx, cartID)
	if err != nil {
		return nil, fmt.Errorf("failed to load cart history: %w", err)
	}

	agg := entity.NewCartAggregate(cartID)
	if err := agg.Rehydrate(records); err != nil {
		return nil, fmt.Errorf("failed to rehydrate cart aggregate: %w", err)
	}

	return agg, nil
}
