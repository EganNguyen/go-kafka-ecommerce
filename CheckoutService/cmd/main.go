package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	deliveryHttp "github.com/egannguyen/go-kafka-ecommerce/checkout-service/internal/delivery/http"
	"github.com/egannguyen/go-kafka-ecommerce/checkout-service/internal/domain"
	"github.com/egannguyen/go-kafka-ecommerce/checkout-service/internal/infrastructure/messaging/kafka"
	"github.com/egannguyen/go-kafka-ecommerce/checkout-service/internal/infrastructure/persistence/postgres"
	"github.com/egannguyen/go-kafka-ecommerce/checkout-service/internal/usecase"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	// --- 1. Infrastructure Layer ---
	dsn := getEnv("DATABASE_URL", "postgres://ecommerce:ecommerce@localhost:5432/ecommerce?sslmode=disable")
	db, err := postgres.InitDB(dsn)
	if err != nil {
		slog.Error("Failed to init database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	productRepo := postgres.NewProductRepository(db)
	orderRepo := postgres.NewOrderRepository(db)
	eventStore := postgres.NewEventStore(db)

	// Seed products
	err = productRepo.Seed(context.Background(), []domain.Product{
		{ID: "prod-001", Name: "Wireless Noise-Cancelling Headphones", Description: "Premium over-ear headphones with active noise cancellation and 30-hour battery life.", Price: 349.99, ImageURL: "https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=400", Category: "Electronics", Stock: 50},
		{ID: "prod-002", Name: "Mechanical Keyboard RGB", Description: "Cherry MX switches with per-key RGB lighting and aluminum frame.", Price: 179.99, ImageURL: "https://images.unsplash.com/photo-1618384887929-16ec33fab9ef?w=400", Category: "Electronics", Stock: 120},
		{ID: "prod-003", Name: "Ultrawide Curved Monitor 34\"", Description: "UWQHD 3440x1440 144Hz IPS panel with USB-C connectivity.", Price: 699.99, ImageURL: "https://images.unsplash.com/photo-1527443224154-c4a3942d3acf?w=400", Category: "Electronics", Stock: 30},
		{ID: "prod-004", Name: "Ergonomic Office Chair", Description: "Adjustable lumbar support, breathable mesh, and 4D armrests.", Price: 549.99, ImageURL: "https://images.unsplash.com/photo-1592078615290-033ee584e267?w=400", Category: "Furniture", Stock: 25},
		{ID: "prod-005", Name: "Smart LED Desk Lamp", Description: "Adjustable color temperature, brightness levels, and USB charging port.", Price: 89.99, ImageURL: "https://images.unsplash.com/photo-1507473885765-e6ed057ab6fe?w=400", Category: "Home", Stock: 200},
		{ID: "prod-006", Name: "Premium Laptop Backpack", Description: "Water-resistant 17\" laptop compartment with anti-theft design.", Price: 129.99, ImageURL: "https://images.unsplash.com/photo-1553062407-98eeb64c6a62?w=400", Category: "Accessories", Stock: 80},
	})
	if err != nil {
		slog.Error("Failed to seed products", "err", err)
	}

	brokers := []string{getEnv("KAFKA_BROKERS", "localhost:9092")}
	publisher, subscriber := kafka.NewKafkaBroker(brokers)

	// --- 2. Application Layer (Use Cases) ---
	checkoutUseCase := usecase.NewCheckoutUseCase(orderRepo, productRepo, eventStore, publisher)

	// --- 3. Interface Layer (HTTP Delivery) ---
	httpHandler := deliveryHttp.NewHandler(checkoutUseCase)

	mux := http.NewServeMux()
	httpHandler.RegisterRoutes(mux)

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: deliveryHttp.EnableCORS(mux),
	}

	// --- 4. Start Application ---
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Kafka Consumer: orders.commands -> PlaceOrder Command
	go subscriber.Consume(ctx, "orders.commands", "checkout-commands", func(ctx context.Context, payload []byte) error {
		var cmd domain.PlaceOrder
		if err := json.Unmarshal(payload, &cmd); err != nil {
			return err
		}
		return checkoutUseCase.PlaceOrder(ctx, &cmd)
	})

	// Kafka Consumer: orders.placed -> HandleOrderPlaced Event
	go subscriber.Consume(ctx, "orders.placed", "checkout-placed", func(ctx context.Context, payload []byte) error {
		var event domain.OrderPlaced
		if err := json.Unmarshal(payload, &event); err != nil {
			return err
		}
		return checkoutUseCase.HandleOrderPlaced(ctx, &event)
	})

	// Kafka Consumer: orders.confirmed -> HandleOrderConfirmed Event
	go subscriber.Consume(ctx, "orders.confirmed", "checkout-confirmed-projection", func(ctx context.Context, payload []byte) error {
		var event domain.OrderConfirmed
		if err := json.Unmarshal(payload, &event); err != nil {
			return err
		}
		return checkoutUseCase.HandleOrderConfirmed(ctx, &event)
	})

	go func() {
		slog.Info("🚀 Checkout Service starting on :8080")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "err", err)
			cancel()
		}
	}()

	slog.Info("🔄 Checkout Service started")

	<-ctx.Done()
	slog.Info("Shutting down...")
	httpServer.Shutdown(context.Background())
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
