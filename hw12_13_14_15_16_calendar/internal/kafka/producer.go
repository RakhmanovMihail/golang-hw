// Package kafka provides abstractions for working with Apache Kafka.
package kafka

import (
	"context"
)

// Producer sends messages to Kafka topics.
type Producer interface {
	// Send sends a message to the specified topic.
	Send(ctx context.Context, topic, key string, value []byte) error
	// Close gracefully closes the producer.
	Close() error
}
