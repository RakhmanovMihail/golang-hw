package storage

import (
	"context"
	"time"
)

// Storage is the interface for event storage.
type Storage interface {
	Create(ctx context.Context, e *Event) (*Event, error)
	Read(ctx context.Context) ([]Event, error)
	Update(ctx context.Context, id uint64, e *Event) (*Event, error)
	Delete(ctx context.Context, id uint64) error
	GetByID(ctx context.Context, id uint64) (*Event, error)

	// GetEventsForNotify returns events that need notification sent.
	// notifyTime is the current time; events with start_time - notify_before <= notifyTime are returned.
	GetEventsForNotify(ctx context.Context, notifyTime time.Time) ([]Event, error)
	// DeleteOldEvents deletes events older than the given cutoff time.
	DeleteOldEvents(ctx context.Context, cutoffTime time.Time) (int64, error)
	// SaveNotification saves a notification to the database.
	SaveNotification(ctx context.Context, n *Notification) error
	// Close closes the storage connection.
	Close(ctx context.Context) error
}
