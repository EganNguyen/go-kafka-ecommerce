package mongodb

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/egannguyen/go-kafka-ecommerce/product-catalog-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitDB(uri string) (*mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongodb: %w", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping mongodb: %w", err)
	}

	db := client.Database("ecommerce")

	if err := seedProducts(ctx, db); err != nil {
		slog.Warn("Failed to seed products", "err", err)
	}

	slog.Info("MongoDB connected and seeded")
	return db, nil
}

func seedProducts(ctx context.Context, db *mongo.Database) error {
	coll := db.Collection("products")

	count, err := coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	products := []interface{}{
		domain.Product{
			ID:          "prod-001",
			Name:        "Ergonomic Keyboard",
			Description: "Comfortable typing for long sessions",
			Price:       129.99,
			ImageURL:    "https://placehold.co/400x300?text=Keyboard",
			Category:    "Peripherals",
			Stock:       50,
		},
		domain.Product{
			ID:          "prod-002",
			Name:        "Wireless Mouse",
			Description: "High precision laser sensor",
			Price:       59.00,
			ImageURL:    "https://placehold.co/400x300?text=Mouse",
			Category:    "Peripherals",
			Stock:       100,
		},
		domain.Product{
			ID:          "prod-003",
			Name:        "4K Monitor",
			Description: "Stunning visuals and color accuracy",
			Price:       399.50,
			ImageURL:    "https://placehold.co/400x300?text=Monitor",
			Category:    "Displays",
			Stock:       20,
		},
		domain.Product{
			ID:          "prod-004",
			Name:        "Webcam HD",
			Description: "Clear video for your meetings",
			Price:       85.25,
			ImageURL:    "https://placehold.co/400x300?text=Webcam",
			Category:    "Video",
			Stock:       45,
		},
	}

	// Drop collection to re-seed with correct tags
	_ = coll.Drop(ctx)

	_, err = coll.InsertMany(ctx, products)
	if err != nil {
		return fmt.Errorf("failed to seed products: %w", err)
	}

	slog.Info("Database seeded with products")
	return nil
}
