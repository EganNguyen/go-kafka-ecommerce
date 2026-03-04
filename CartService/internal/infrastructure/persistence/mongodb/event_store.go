package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/egannguyen/go-kafka-ecommerce/cart-service/internal/domain"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type eventStore struct {
	db *mongo.Database
}

func NewEventStore(db *mongo.Database) domain.EventStore {
	return &eventStore{db: db}
}

func (s *eventStore) SaveEvents(ctx context.Context, streamID string, streamType string, expectedVersion int, events []domain.Event) error {
	if len(events) == 0 {
		return nil
	}

	coll := s.db.Collection("events")

	// Start session for transaction
	session, err := s.db.Client().StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Check concurrency
		var result struct {
			Version int `bson:"version"`
		}
		opts := options.FindOne().SetSort(bson.M{"version": -1})
		err := coll.FindOne(sessCtx, bson.M{"stream_id": streamID}, opts).Decode(&result)
		
		currentVersion := 0
		if err == nil {
			currentVersion = result.Version
		} else if err != mongo.ErrNoDocuments {
			return nil, fmt.Errorf("failed to get current stream version: %w", err)
		}

		if currentVersion != expectedVersion {
			return nil, fmt.Errorf("concurrency exception: expected version %d, got %d", expectedVersion, currentVersion)
		}

		var docs []interface{}
		version := expectedVersion
		now := time.Now()

		for _, event := range events {
			version++
			docs = append(docs, bson.M{
				"id":          uuid.NewString(),
				"stream_id":   streamID,
				"stream_type": streamType,
				"version":     version,
				"event_type":  event.EventType(),
				"payload":     event, // MongoDB driver handles marshaling
				"created_at":  now,
			})
		}

		_, err = coll.InsertMany(sessCtx, docs)
		if err != nil {
			return nil, fmt.Errorf("failed to insert events: %w", err)
		}

		return nil, nil
	})

	return err
}

func (s *eventStore) LoadEvents(ctx context.Context, streamID string) ([]domain.EventRecord, error) {
	coll := s.db.Collection("events")
	opts := options.Find().SetSort(bson.M{"version": 1})
	cursor, err := coll.Find(ctx, bson.M{"stream_id": streamID}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to load events: %w", err)
	}
	defer cursor.Close(ctx)

	var events []domain.EventRecord
	if err := cursor.All(ctx, &events); err != nil {
		return nil, fmt.Errorf("failed to decode events: %w", err)
	}

	return events, nil
}
