package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/egannguyen/go-kafka-ecommerce/backend/internal/entity"
	"github.com/egannguyen/go-kafka-ecommerce/backend/internal/repository"
	"github.com/google/uuid"
)

type eventStore struct {
	db *sql.DB
}

// NewEventStore creates a new EventStore backed by Postgres.
func NewEventStore(db *sql.DB) repository.EventStore {
	return &eventStore{db: db}
}

func (s *eventStore) SaveEvents(ctx context.Context, streamID string, streamType string, expectedVersion int, events []entity.Event) error {
	if len(events) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check concurrency
	var currentVersion int
	err = tx.QueryRowContext(ctx, "SELECT COALESCE(MAX(version), 0) FROM events WHERE stream_id = $1", streamID).Scan(&currentVersion)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to get current stream version: %w", err)
	}

	if currentVersion != expectedVersion {
		return fmt.Errorf("concurrency exception: expected version %d, got %d", expectedVersion, currentVersion)
	}

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO events (id, stream_id, stream_type, version, event_type, payload, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)")
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	version := expectedVersion
	now := time.Now()

	for _, event := range events {
		version++

		payload, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event %s: %w", event.EventType(), err)
		}

		_, err = stmt.ExecContext(ctx, uuid.NewString(), streamID, streamType, version, event.EventType(), payload, now)
		if err != nil {
			return fmt.Errorf("failed to insert event %s: %w", event.EventType(), err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *eventStore) LoadEvents(ctx context.Context, streamID string) ([]entity.EventStoreRecord, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, stream_id, stream_type, version, event_type, payload, created_at FROM events WHERE stream_id = $1 ORDER BY version ASC", streamID)
	if err != nil {
		return nil, fmt.Errorf("failed to load events for stream %s: %w", streamID, err)
	}
	defer rows.Close()

	var events []entity.EventStoreRecord
	for rows.Next() {
		var record entity.EventStoreRecord
		if err := rows.Scan(&record.ID, &record.StreamID, &record.StreamType, &record.Version, &record.EventType, &record.Payload, &record.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan event record: %w", err)
		}
		events = append(events, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating event rows: %w", err)
	}

	return events, nil
}
