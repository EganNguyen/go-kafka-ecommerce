package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/egannguyen/go-kafka-ecommerce/checkout-service/internal/domain"
	"github.com/google/uuid"
)

type eventStore struct {
	db *sql.DB
}

func NewEventStore(db *sql.DB) domain.EventStore {
	return &eventStore{db: db}
}

func (s *eventStore) SaveEvents(ctx context.Context, streamID string, streamType string, expectedVersion int, events []domain.Event) error {
	if len(events) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var currentVersion int
	err = tx.QueryRowContext(ctx, "SELECT COALESCE(MAX(version), 0) FROM events WHERE stream_id = $1", streamID).Scan(&currentVersion)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to get current stream version: %w", err)
	}

	if expectedVersion != -1 && currentVersion != expectedVersion {
		return fmt.Errorf("concurrency exception: expected version %d, got %d", expectedVersion, currentVersion)
	}

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO events (id, stream_id, stream_type, version, event_type, payload, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)")
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	version := currentVersion
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

	return tx.Commit()
}

func (s *eventStore) LoadEvents(ctx context.Context, streamID string) ([]domain.EventRecord, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, stream_id, stream_type, version, event_type, payload, created_at FROM events WHERE stream_id = $1 ORDER BY version ASC", streamID)
	if err != nil {
		return nil, fmt.Errorf("failed to load events for stream %s: %w", streamID, err)
	}
	defer rows.Close()

	var events []domain.EventRecord
	for rows.Next() {
		var record domain.EventRecord
		if err := rows.Scan(&record.ID, &record.StreamID, &record.StreamType, &record.Version, &record.EventType, &record.Payload, &record.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan event record: %w", err)
		}
		events = append(events, record)
	}

	return events, rows.Err()
}
