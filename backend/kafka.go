package main

import (
	"context"
	"log/slog"

	kafkaGo "github.com/segmentio/kafka-go"
)

// newWriter creates a Kafka writer for a specific topic.
func newWriter(brokers []string, topic string) *kafkaGo.Writer {
	return &kafkaGo.Writer{
		Addr:     kafkaGo.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafkaGo.LeastBytes{},
	}
}

// newReader creates a Kafka reader (consumer) for a specific topic and consumer group.
func newReader(brokers []string, topic, groupID string) *kafkaGo.Reader {
	return kafkaGo.NewReader(kafkaGo.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})
}

// publishJSON writes a JSON-encoded message to Kafka.
func publishJSON(ctx context.Context, w *kafkaGo.Writer, key, value []byte) error {
	return w.WriteMessages(ctx, kafkaGo.Message{
		Key:   key,
		Value: value,
	})
}

// consume reads messages from a Kafka reader in a loop and calls the handler for each message.
// It blocks until the context is cancelled.
func consume(ctx context.Context, reader *kafkaGo.Reader, handler func(ctx context.Context, msg kafkaGo.Message) error) {
	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				slog.Info("Consumer shutting down", "topic", reader.Config().Topic)
				return
			}
			slog.Error("Error reading message", "topic", reader.Config().Topic, "err", err)
			continue
		}

		if err := handler(ctx, msg); err != nil {
			slog.Error("Error handling message", "topic", reader.Config().Topic, "err", err)
		}
	}
}
