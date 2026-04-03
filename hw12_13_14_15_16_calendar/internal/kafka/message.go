package kafka

import "time"

// NotificationMessage represents a notification sent via Kafka.
type NotificationMessage struct {
	EventID   uint64    `json:"event_id"`
	Title     string    `json:"title"`
	EventDate time.Time `json:"event_date"`
	UserID    int64     `json:"user_id"`
}
