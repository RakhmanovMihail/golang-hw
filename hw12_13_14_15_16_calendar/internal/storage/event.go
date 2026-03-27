package storage

import (
	"errors"
	"time"
)

var (
	ErrEventNotFound = errors.New("event not found")
	ErrDateBusy      = errors.New("date already busy")
)

type Event struct {
	ID        uint64    `db:"id"`
	Title     string    `db:"title"`
	StartTime time.Time `db:"start_time"`
	EndTime   time.Time `db:"end_time"`
	UserID    int       `db:"user_id"`
}
