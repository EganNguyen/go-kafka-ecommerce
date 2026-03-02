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
}

func NewCartUseCase(eventStore domain.EventStore) CartUseCase {
	return &cartUseCase{
		eventStore: eventStore,
	}
}

func (u *cartUseCase) AddItemToCart(ctx context.Context, cartID, productID string, quantity int, price float64) error {
	slog.Info("UseCase: Adding item to cart", "cart_id", cartID, "product_id", productID)

	records, err := u.eventStore.LoadEvents(ctx, cartID)
	if err != nil {
		return fmt.Errorf("failed to load cart history: %w", err)
	}

	agg := domain.NewCartAggregate(cartID)
	if err := agg.Rehydrate(records); err != nil {
		return fmt.Errorf("failed to rehydrate cart aggregate: %w", err)
	}

	event := domain.ItemAddedToCart{
		CartID:    cartID,
		ProductID: productID,
		Quantity:  quantity,
		Price:     price,
	}

	err = u.eventStore.SaveEvents(ctx, cartID, "cart", agg.GetVersion(), []domain.Event{event})
	if err != nil {
		return fmt.Errorf("failed to save ItemAddedToCart event: %w", err)
	}

	return nil
}

func (u *cartUseCase) GetCart(ctx context.Context, cartID string) (*domain.CartAggregate, error) {
	records, err := u.eventStore.LoadEvents(ctx, cartID)
	if err != nil {
		return nil, fmt.Errorf("failed to load cart history: %w", err)
	}

	agg := domain.NewCartAggregate(cartID)
	if err := agg.Rehydrate(records); err != nil {
		return nil, fmt.Errorf("failed to rehydrate cart aggregate: %w", err)
	}

	return agg, nil
}
