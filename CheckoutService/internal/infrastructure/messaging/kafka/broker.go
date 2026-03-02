package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/egannguyen/go-kafka-ecommerce/checkout-service/internal/domain"
	kafkaGo "github.com/segmentio/kafka-go"
)

type kafkaBroker struct {
	brokers []string
}

func NewKafkaBroker(brokers []string) (domain.Publisher, domain.Subscriber) {
	kb := &kafkaBroker{brokers: brokers}
	return kb, kb
}

func (k *kafkaBroker) PublishEvent(ctx context.Context, topic string, key string, event interface{}) error {
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

func (k *kafkaBroker) Consume(ctx context.Context, topic string, groupID string, handler func(ctx context.Context, payload []byte) error) error {
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
				return nil
			}
			slog.Error("Error reading message", "topic", topic, "err", err)
			continue
		}

		if err := handler(ctx, msg.Value); err != nil {
			slog.Error("Error handling message", "topic", topic, "err", err)
		}
	}
}
