// Package storage provides storage types and errors.
package storage

import (
	"errors"
	"time"
)

var (
	// ErrEventNotFound is returned when an event is not found.
	ErrEventNotFound = errors.New("event not found")
	// ErrDateBusy is returned when a date is already busy.
	ErrDateBusy = errors.New("date already busy")
)

// Event represents a calendar event.
type Event struct {
	ID        uint64    `db:"id"`
	Title     string    `db:"title"`
	StartTime time.Time `db:"start_time"`
	EndTime   time.Time `db:"end_time"`
	UserID    int       `db:"user_id"`
}
