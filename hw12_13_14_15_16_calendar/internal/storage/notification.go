package storage

import "time"

// Notification represents a reminder notification about an event.
type Notification struct {
	ID        uint64    `db:"id"`
	EventID   uint64    `db:"event_id"`
	Title     string    `db:"title"`
	EventDate time.Time `db:"event_date"`
	UserID    int64     `db:"user_id"`
}
