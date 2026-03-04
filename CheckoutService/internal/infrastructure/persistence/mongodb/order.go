package mongodb

import (
	"context"
	"fmt"

	"github.com/egannguyen/go-kafka-ecommerce/checkout-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type orderRepository struct {
	db *mongo.Database
}

func NewOrderRepository(db *mongo.Database) domain.OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) UpdateOrderProjection(ctx context.Context, event interface{}) error {
	coll := r.db.Collection("orders")

	switch e := event.(type) {
	case domain.OrderPlaced:
		order := domain.Order{
			ID:         e.OrderID,
			TotalPrice: e.TotalPrice,
			Status:     "placed",
			CreatedAt:  e.PlacedAt,
			Items:      e.Items,
		}
		opts := options.Update().SetUpsert(true)
		_, err := coll.UpdateOne(ctx, bson.M{"id": e.OrderID}, bson.M{"$set": order}, opts)
		if err != nil {
			return fmt.Errorf("failed to upsert order projection: %w", err)
		}

	case domain.OrderConfirmed:
		_, err := coll.UpdateOne(ctx, bson.M{"id": e.OrderID}, bson.M{"$set": bson.M{"status": "confirmed"}})
		if err != nil {
			return fmt.Errorf("failed to confirm order projection: %w", err)
		}
	}

	return nil
}

func (r *orderRepository) FindRecent(ctx context.Context, limit int) ([]domain.Order, error) {
	coll := r.db.Collection("orders")
	opts := options.Find().SetSort(bson.M{"created_at": -1}).SetLimit(int64(limit))
	cursor, err := coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer cursor.Close(ctx)

	var orders []domain.Order
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, fmt.Errorf("failed to decode orders: %w", err)
	}

	return orders, nil
}
