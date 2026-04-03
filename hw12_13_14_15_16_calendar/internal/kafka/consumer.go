package kafka

import (
	"context"
)

// Message represents a Kafka message consumed from a topic.
type Message struct {
	Topic string
	Key   string
	Value []byte
}

// Handler processes a consumed message.
type Handler func(ctx context.Context, msg Message) error

// Consumer reads messages from Kafka topics.
type Consumer interface {
	// Start begins consuming messages and calling the handler.
	Start(ctx context.Context, handler Handler) error
	// Close gracefully closes the consumer.
	Close() error
}
