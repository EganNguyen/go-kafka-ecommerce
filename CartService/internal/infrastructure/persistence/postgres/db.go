package postgres

import (
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/lib/pq"
)

func InitDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := migrateDB(db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	slog.Info("Database connected and migrated")
	return db, nil
}

func migrateDB(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS events (
			id TEXT PRIMARY KEY,
			stream_id TEXT NOT NULL,
			stream_type TEXT NOT NULL,
			version INT NOT NULL,
			event_type TEXT NOT NULL,
			payload JSONB NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			UNIQUE(stream_id, version)
		);
	`)
	return err
}
