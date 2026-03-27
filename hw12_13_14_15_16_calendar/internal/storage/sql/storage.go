package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/jmoiron/sqlx"
)

// Store is a SQL implementation of storage.Storage.
type Store struct {
	db *sqlx.DB
}

// New creates a new Store instance.
func New(dsn string) (storage.Storage, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return &Store{db: db}, nil
}

// Connect checks the database connection.
func (s *Store) Connect(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// Close closes the database connection.
func (s *Store) Close(_ context.Context) error {
	return s.db.Close()
}

// Create creates a new event in the database.
func (s *Store) Create(ctx context.Context, e *storage.Event) (*storage.Event, error) {
	query := `
       INSERT INTO events (title, start_time, end_time, user_id)
       VALUES ($1, $2, $3, $4)
       RETURNING id, title, start_time, end_time, user_id`

	created := &storage.Event{}
	err := s.db.QueryRowContext(ctx, query,
		e.Title, e.StartTime, e.EndTime, e.UserID).
		Scan(&created.ID, &created.Title, &created.StartTime,
			&created.EndTime, &created.UserID)
	if err != nil {
		return nil, err
	}

	return created, nil
}

// Read returns all events from the database.
func (s *Store) Read(ctx context.Context) ([]storage.Event, error) {
	query := `SELECT id, title, start_time, end_time, user_id FROM events ORDER BY start_time`

	var events []storage.Event
	err := s.db.SelectContext(ctx, &events, query)
	if err != nil {
		return nil, err
	}

	return events, nil
}

// Update updates an existing event in the database.
func (s *Store) Update(ctx context.Context, id uint64, e *storage.Event) (*storage.Event, error) {
	query := `
        UPDATE events
        SET title = $1, start_time = $2, end_time = $3, user_id = $4
        WHERE id = $5
        RETURNING id, title, start_time, end_time, user_id`

	updated := &storage.Event{}
	err := s.db.QueryRowContext(ctx, query,
		e.Title, e.StartTime, e.EndTime, e.UserID, id).
		Scan(&updated.ID, &updated.Title, &updated.StartTime,
			&updated.EndTime, &updated.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrEventNotFound
		}
		return nil, err
	}

	return updated, nil
}

// Delete deletes an event from the database.
func (s *Store) Delete(ctx context.Context, id uint64) error {
	query := `DELETE FROM events WHERE id = $1 RETURNING id`

	var returnedID uint64
	err := s.db.QueryRowContext(ctx, query, id).Scan(&returnedID)
	if err != nil {
		if err == sql.ErrNoRows {
			return storage.ErrEventNotFound
		}
		return err
	}

	return nil
}

// GetByID returns an event by ID from the database.
func (s *Store) GetByID(ctx context.Context, id uint64) (*storage.Event, error) {
	query := `SELECT id, title, start_time, end_time, user_id FROM events WHERE id = $1`

	event := &storage.Event{}
	err := s.db.GetContext(ctx, event, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrEventNotFound
		}
		return nil, err
	}

	return event, nil
}
