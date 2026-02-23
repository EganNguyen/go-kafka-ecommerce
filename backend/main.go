package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	kafkaGo "github.com/segmentio/kafka-go"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	// --- Database ---
	dsn := getEnv("DATABASE_URL", "postgres://ecommerce:ecommerce@localhost:5432/ecommerce?sslmode=disable")
	db, err := initDB(dsn)
	if err != nil {
		slog.Error("Failed to init database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := seedProducts(db); err != nil {
		slog.Error("Failed to seed products", "err", err)
		os.Exit(1)
	}

	// --- Kafka ---
	brokers := []string{getEnv("KAFKA_BROKERS", "localhost:9092")}

	commandWriter := newWriter(brokers, "orders.commands")
	defer commandWriter.Close()

	placedWriter := newWriter(brokers, "orders.placed")
	defer placedWriter.Close()

	commandReader := newReader(brokers, "orders.commands", "ecommerce-commands")
	defer commandReader.Close()

	placedReader := newReader(brokers, "orders.placed", "ecommerce-placed")
	defer placedReader.Close()

	// --- Handlers ---
	placeOrderHandler := &PlaceOrderHandler{db: db, eventWriter: placedWriter}
	orderPlacedHandler := &OrderPlacedHandler{db: db}

	// --- HTTP API ---
	api := &API{
		db: db,
		placeOrder: func(cmd *PlaceOrder) error {
			payload, err := json.Marshal(cmd)
			if err != nil {
				return fmt.Errorf("failed to marshal PlaceOrder: %w", err)
			}
			return publishJSON(context.Background(), commandWriter, []byte(cmd.OrderID), payload)
		},
	}

	mux := http.NewServeMux()
	api.RegisterRoutes(mux)

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: enableCORS(mux),
	}

	// --- Start everything ---
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Consumer: orders.commands → PlaceOrderHandler → publish OrderPlaced
	go consume(ctx, commandReader, func(ctx context.Context, msg kafkaGo.Message) error {
		var cmd PlaceOrder
		if err := json.Unmarshal(msg.Value, &cmd); err != nil {
			return fmt.Errorf("failed to unmarshal PlaceOrder command: %w", err)
		}

		event, err := placeOrderHandler.Handle(ctx, &cmd)
		if err != nil {
			return err
		}

		// If the order was a duplicate (idempotent skip), don't publish event.
		if event == nil {
			return nil
		}

		return publishEvent(ctx, placedWriter, event.OrderID, event)
	})

	// Consumer: orders.placed → OrderPlacedHandler (confirms order)
	go consume(ctx, placedReader, func(ctx context.Context, msg kafkaGo.Message) error {
		var event OrderPlaced
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			return fmt.Errorf("failed to unmarshal OrderPlaced event: %w", err)
		}
		return orderPlacedHandler.Handle(ctx, &event)
	})

	go func() {
		slog.Info("🚀 HTTP server starting on :8080")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "err", err)
			cancel()
		}
	}()

	slog.Info("🔄 Kafka consumers started")

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
