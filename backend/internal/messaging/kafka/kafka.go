package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/egannguyen/go-kafka-ecommerce/backend/internal/messaging"
	kafkaGo "github.com/segmentio/kafka-go"
)

type kafkaBroker struct {
	brokers []string
}

// NewKafkaBroker creates a new Kafka publisher and subscriber.
func NewKafkaBroker(brokers []string) (messaging.Publisher, messaging.Subscriber) {
	kb := &kafkaBroker{brokers: brokers}
	return kb, kb
}

func (k *kafkaBroker) PublishEvent(ctx context.Context, topic string, key string, event any) error {
	w := &kafkaGo.Writer{
		Addr:     kafkaGo.TCP(k.brokers...),
		Topic:    topic,
		Balancer: &kafkaGo.LeastBytes{},
	}
	defer w.Close()

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	return w.WriteMessages(ctx, kafkaGo.Message{
		Key:   []byte(key),
		Value: payload,
	})
}

func (k *kafkaBroker) Consume(ctx context.Context, topic string, groupID string, handler func(ctx context.Context, payload []byte) error) {
	reader := kafkaGo.NewReader(kafkaGo.ReaderConfig{
		Brokers: k.brokers,
		Topic:   topic,
		GroupID: groupID,
	})
	defer reader.Close()

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				slog.Info("Consumer shutting down", "topic", topic)
				return
			}
			slog.Error("Error reading message", "topic", topic, "err", err)
			continue
		}

		if err := handler(ctx, msg.Value); err != nil {
			slog.Error("Error handling message", "topic", topic, "err", err)
		}
	}
}
