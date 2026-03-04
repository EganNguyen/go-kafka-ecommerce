package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/egannguyen/go-kafka-ecommerce/cart-service/internal/domain"
	"github.com/redis/go-redis/v9"
)

type cartRepository struct {
	client *redis.Client
}

func NewCartRepository(client *redis.Client) domain.CartRepository {
	return &cartRepository{client: client}
}

func (r *cartRepository) Save(ctx context.Context, cart *domain.CartAggregate) error {
	data, err := json.Marshal(cart)
	if err != nil {
		return fmt.Errorf("failed to marshal cart: %w", err)
	}

	key := fmt.Sprintf("cart:%s", cart.ID)
	err = r.client.Set(ctx, key, data, 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to save cart to redis: %w", err)
	}
	return nil
}

func (r *cartRepository) Get(ctx context.Context, cartID string) (*domain.CartAggregate, error) {
	key := fmt.Sprintf("cart:%s", cartID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get cart from redis: %w", err)
	}

	var cart domain.CartAggregate
	if err := json.Unmarshal(data, &cart); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cart: %w", err)
	}
	return &cart, nil
}

func (r *cartRepository) Delete(ctx context.Context, cartID string) error {
	key := fmt.Sprintf("cart:%s", cartID)
	return r.client.Del(ctx, key).Err()
}
