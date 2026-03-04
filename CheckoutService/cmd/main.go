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
	"github.com/egannguyen/go-kafka-ecommerce/checkout-service/internal/infrastructure/grpc"
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

	orderRepo := postgres.NewOrderRepository(db)
	eventStore := postgres.NewEventStore(db)

	productCatalogAddr := getEnv("PRODUCT_CATALOG_ADDR", "localhost:50051")
	productService, err := grpc.NewProductServiceClient(productCatalogAddr)
	if err != nil {
		slog.Error("Failed to init product service client", "err", err)
		os.Exit(1)
	}

	currencyAddr := getEnv("CURRENCY_SERVICE_ADDR", "localhost:50051")
	currencyService, err := grpc.NewCurrencyServiceClient(currencyAddr)
	if err != nil {
		slog.Error("Failed to init currency service client", "err", err)
		os.Exit(1)
	}

	paymentAddr := getEnv("PAYMENT_SERVICE_ADDR", "localhost:50051")
	paymentService, err := grpc.NewPaymentServiceClient(paymentAddr)
	if err != nil {
		slog.Error("Failed to init payment service client", "err", err)
		os.Exit(1)
	}

	brokers := []string{getEnv("KAFKA_BROKERS", "localhost:9092")}
	publisher, subscriber := kafka.NewKafkaBroker(brokers)

	// --- 2. Application Layer (Use Cases) ---
	checkoutUseCase := usecase.NewCheckoutUseCase(orderRepo, productService, currencyService, paymentService, eventStore, publisher)

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
