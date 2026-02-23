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

	"github.com/IBM/sarama"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	logger := watermill.NewSlogLoggerWithLevelMapping(nil, map[slog.Level]slog.Level{
		slog.LevelInfo: slog.LevelDebug,
	})

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

	kafkaMarshaler := kafka.DefaultMarshaler{}

	publisher, err := kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:   brokers,
			Marshaler: kafkaMarshaler,
		},
		logger,
	)
	if err != nil {
		slog.Error("Failed to create Kafka publisher", "err", err)
		os.Exit(1)
	}

	saramaConfig := kafka.DefaultSaramaSubscriberConfig()
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	subscriber, err := kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers:               brokers,
			Unmarshaler:           kafkaMarshaler,
			OverwriteSaramaConfig: saramaConfig,
			ConsumerGroup:         "ecommerce-app",
		},
		logger,
	)
	if err != nil {
		slog.Error("Failed to create Kafka subscriber", "err", err)
		os.Exit(1)
	}

	// --- Watermill Router ---
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		slog.Error("Failed to create router", "err", err)
		os.Exit(1)
	}

	router.AddMiddleware(
		middleware.Recoverer,
		middleware.CorrelationID,
	)

	// Subscribe to the "orders.placed" topic and handle events.
	orderPlacedHandler := &OrderPlacedHandler{db: db}
	router.AddHandler(
		orderPlacedHandler.HandlerName(),
		"orders.placed", // subscribe topic
		subscriber,
		"orders.confirmed", // publish topic (for downstream)
		publisher,
		func(msg *message.Message) ([]*message.Message, error) {
			var event OrderPlaced
			if err := json.Unmarshal(msg.Payload, &event); err != nil {
				return nil, fmt.Errorf("failed to unmarshal OrderPlaced: %w", err)
			}

			if err := orderPlacedHandler.Handle(context.Background(), &event); err != nil {
				return nil, err
			}

			// Publish a confirmation event downstream.
			confirmedPayload, _ := json.Marshal(OrderConfirmed{
				OrderID:     event.OrderID,
				ConfirmedAt: event.PlacedAt,
			})
			confirmMsg := message.NewMessage(watermill.NewUUID(), confirmedPayload)

			return []*message.Message{confirmMsg}, nil
		},
	)

	// --- HTTP API ---
	api := &API{
		db: db,
		placeOrder: func(cmd *PlaceOrder) error {
			// Publish the PlaceOrder as a message to Kafka.
			payload, err := json.Marshal(cmd)
			if err != nil {
				return fmt.Errorf("failed to marshal PlaceOrder: %w", err)
			}
			msg := message.NewMessage(watermill.NewUUID(), payload)
			return publisher.Publish("orders.commands", msg)
		},
	}

	// We also add a simple handler that listens for commands and executes them.
	commandHandler := &PlaceOrderHandler{db: db, eventBus: nil}
	router.AddHandler(
		"PlaceOrderCommandHandler",
		"orders.commands",
		subscriber,
		"orders.placed",
		publisher,
		func(msg *message.Message) ([]*message.Message, error) {
			var cmd PlaceOrder
			if err := json.Unmarshal(msg.Payload, &cmd); err != nil {
				return nil, fmt.Errorf("failed to unmarshal PlaceOrder command: %w", err)
			}

			if err := commandHandler.Handle(context.Background(), &cmd); err != nil {
				return nil, err
			}

			// After handling the command, publish the OrderPlaced event.
			var totalPrice float64
			for _, item := range cmd.Items {
				totalPrice += item.Price * float64(item.Quantity)
			}

			event := OrderPlaced{
				OrderID:    cmd.OrderID,
				Items:      cmd.Items,
				TotalPrice: totalPrice,
			}
			eventPayload, _ := json.Marshal(event)
			eventMsg := message.NewMessage(watermill.NewUUID(), eventPayload)
			return []*message.Message{eventMsg}, nil
		},
	)

	mux := http.NewServeMux()
	api.RegisterRoutes(mux)

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: enableCORS(mux),
	}

	// --- Start everything ---
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go func() {
		slog.Info("🚀 HTTP server starting on :8080")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "err", err)
			cancel()
		}
	}()

	go func() {
		slog.Info("🔄 Watermill router starting...")
		if err := router.Run(ctx); err != nil {
			slog.Error("Router error", "err", err)
			cancel()
		}
	}()

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
