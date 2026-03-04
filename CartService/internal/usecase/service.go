package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/egannguyen/go-kafka-ecommerce/cart-service/internal/domain"
)

// CartUseCase orchestrates shopping cart logic.
type CartUseCase interface {
	AddItemToCart(ctx context.Context, cartID, productID string, quantity int, price float64) error
	GetCart(ctx context.Context, cartID string) (*domain.CartAggregate, error)
}

type cartUseCase struct {
	eventStore domain.EventStore
	repo       domain.CartRepository
}

func NewCartUseCase(eventStore domain.EventStore, repo domain.CartRepository) CartUseCase {
	return &cartUseCase{
		eventStore: eventStore,
		repo:       repo,
	}
}

func (u *cartUseCase) AddItemToCart(ctx context.Context, cartID, productID string, quantity int, price float64) error {
	slog.Info("UseCase: Adding item to cart", "cart_id", cartID, "product_id", productID)

	// Try to get from cache first
	agg, err := u.repo.Get(ctx, cartID)
	if err != nil {
		slog.Warn("Failed to get cart from cache, rehydrating from event store", "err", err)
	}

	if agg == nil {
		// Rehydrate if not in cache
		records, err := u.eventStore.LoadEvents(ctx, cartID)
		if err != nil {
			return fmt.Errorf("failed to load cart history: %w", err)
		}

		agg = domain.NewCartAggregate(cartID)
		if err := agg.Rehydrate(records); err != nil {
			return fmt.Errorf("failed to rehydrate cart aggregate: %w", err)
		}
	}

	event := domain.ItemAddedToCart{
		CartID:    cartID,
		ProductID: productID,
		Quantity:  quantity,
		Price:     price,
	}

	// Persist event
	err = u.eventStore.SaveEvents(ctx, cartID, "cart", agg.GetVersion(), []domain.Event{event})
	if err != nil {
		return fmt.Errorf("failed to save ItemAddedToCart event: %w", err)
	}

	// Apply event to aggregate and update cache
	if err := agg.ApplyEvent(event); err != nil {
		return fmt.Errorf("failed to apply event to aggregate: %w", err)
	}

	if err := u.repo.Save(ctx, agg); err != nil {
		slog.Error("Failed to save cart to cache", "err", err)
	}

	return nil
}

func (u *cartUseCase) GetCart(ctx context.Context, cartID string) (*domain.CartAggregate, error) {
	// Try cache first
	agg, err := u.repo.Get(ctx, cartID)
	if err == nil && agg != nil {
		return agg, nil
	}

	// Rehydrate if not in cache or error
	records, err := u.eventStore.LoadEvents(ctx, cartID)
	if err != nil {
		return nil, fmt.Errorf("failed to load cart history: %w", err)
	}
	if len(records) == 0 {
		return domain.NewCartAggregate(cartID), nil
	}

	agg = domain.NewCartAggregate(cartID)
	if err := agg.Rehydrate(records); err != nil {
		return nil, fmt.Errorf("failed to rehydrate cart aggregate: %w", err)
	}

	// Save back to cache
	_ = u.repo.Save(ctx, agg)

	return agg, nil
}
