package event

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ncobase/ncore/logging/logger"
)

type SQLiteStore struct {
	db     *sql.DB
	logger *logger.Logger
}

func NewSQLiteStore(db *sql.DB, logger *logger.Logger) (*SQLiteStore, error) {
	store := &SQLiteStore{db: db, logger: logger}
	if err := store.initSchema(context.Background()); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *SQLiteStore) initSchema(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS events (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			aggregate_id TEXT,
			aggregate_name TEXT,
			payload TEXT NOT NULL,
			metadata TEXT,
			timestamp TEXT NOT NULL,
			version INTEGER NOT NULL
		);
	`)
	return err
}

func (s *SQLiteStore) Save(ctx context.Context, event *Event) error {
	payloadJSON, err := json.Marshal(event.Payload)
	if err != nil {
		return err
	}

	metadataJSON, err := json.Marshal(event.Metadata)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO events (id, type, aggregate_id, aggregate_name, payload, metadata, timestamp, version)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		event.ID,
		string(event.Type),
		event.AggregateID,
		event.AggregateName,
		string(payloadJSON),
		string(metadataJSON),
		event.Timestamp.UTC().Format(time.RFC3339Nano),
		event.Version,
	)

	if err != nil {
		s.logger.Error(ctx, "Failed to insert event", "error", err)
	}
	return err
}

func (s *SQLiteStore) Load(ctx context.Context, eventID string) (*Event, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, type, aggregate_id, aggregate_name, payload, metadata, timestamp, version
		FROM events WHERE id = ?
	`, eventID)

	return scanEvent(row)
}

func (s *SQLiteStore) LoadByAggregate(ctx context.Context, aggregateID string) ([]*Event, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, type, aggregate_id, aggregate_name, payload, metadata, timestamp, version
		FROM events WHERE aggregate_id = ?
		ORDER BY timestamp ASC
	`, aggregateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanEvents(rows)
}

func (s *SQLiteStore) LoadByType(ctx context.Context, eventType EventType) ([]*Event, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, type, aggregate_id, aggregate_name, payload, metadata, timestamp, version
		FROM events WHERE type = ?
		ORDER BY timestamp ASC
	`, string(eventType))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanEvents(rows)
}

func (s *SQLiteStore) LoadSince(ctx context.Context, since time.Time) ([]*Event, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, type, aggregate_id, aggregate_name, payload, metadata, timestamp, version
		FROM events WHERE timestamp >= ?
		ORDER BY timestamp ASC
	`, since.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanEvents(rows)
}

func scanEvent(row *sql.Row) (*Event, error) {
	var payloadJSON string
	var metadataJSON string
	var eventType string
	var timestamp string

	event := &Event{}
	if err := row.Scan(
		&event.ID,
		&eventType,
		&event.AggregateID,
		&event.AggregateName,
		&payloadJSON,
		&metadataJSON,
		&timestamp,
		&event.Version,
	); err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(payloadJSON), &event.Payload); err != nil {
		return nil, err
	}
	if metadataJSON != "" {
		if err := json.Unmarshal([]byte(metadataJSON), &event.Metadata); err != nil {
			return nil, err
		}
	}

	parsedTime, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp: %w", err)
	}

	event.Type = EventType(eventType)
	event.Timestamp = parsedTime

	return event, nil
}

func scanEvents(rows *sql.Rows) ([]*Event, error) {
	var events []*Event
	for rows.Next() {
		var payloadJSON string
		var metadataJSON string
		var eventType string
		var timestamp string
		event := &Event{}
		if err := rows.Scan(
			&event.ID,
			&eventType,
			&event.AggregateID,
			&event.AggregateName,
			&payloadJSON,
			&metadataJSON,
			&timestamp,
			&event.Version,
		); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(payloadJSON), &event.Payload); err != nil {
			return nil, err
		}
		if metadataJSON != "" {
			if err := json.Unmarshal([]byte(metadataJSON), &event.Metadata); err != nil {
				return nil, err
			}
		}

		parsedTime, err := time.Parse(time.RFC3339Nano, timestamp)
		if err != nil {
			return nil, fmt.Errorf("invalid timestamp: %w", err)
		}

		event.Type = EventType(eventType)
		event.Timestamp = parsedTime
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}
