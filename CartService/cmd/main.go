package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	deliveryHttp "github.com/egannguyen/go-kafka-ecommerce/cart-service/internal/delivery/http"
	"github.com/egannguyen/go-kafka-ecommerce/cart-service/internal/infrastructure/persistence/postgres"
	"github.com/egannguyen/go-kafka-ecommerce/cart-service/internal/usecase"
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

	eventStore := postgres.NewEventStore(db)

	// --- 2. Application Layer (Use Cases) ---
	cartUseCase := usecase.NewCartUseCase(eventStore)

	// --- 3. Interface Layer (HTTP Delivery) ---
	httpHandler := deliveryHttp.NewHandler(cartUseCase)

	mux := http.NewServeMux()
	httpHandler.RegisterRoutes(mux)

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: deliveryHttp.EnableCORS(mux),
	}

	// --- 4. Start Application ---
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go func() {
		slog.Info("🚀 Cart Service starting on :8080")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "err", err)
			cancel()
		}
	}()

	slog.Info("🔄 Cart Service started")

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
