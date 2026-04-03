package kafka

import "time"

// NotificationMessage represents a notification sent via Kafka.
type NotificationMessage struct {
	//nolint:tagliatelle // Kafka messages use snake_case by convention
	EventID   uint64    `json:"event_id"`
	Title     string    `json:"title"`
	EventDate time.Time `json:"event_date"`
	//nolint:tagliatelle // Kafka messages use snake_case by convention
	UserID int64 `json:"user_id"`
}
