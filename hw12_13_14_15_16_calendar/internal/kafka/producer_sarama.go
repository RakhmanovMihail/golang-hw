package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/sarama"
)

// SaramaProducer implements Producer interface using IBM/sarama.
type SaramaProducer struct {
	producer sarama.SyncProducer
}

// NewSaramaProducer creates a new Kafka producer with retry logic.
func NewSaramaProducer(brokers []string) (*SaramaProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Retry.Backoff = 100 * time.Millisecond

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return &SaramaProducer{producer: producer}, nil
}

// Send sends a message to the specified topic.
func (p *SaramaProducer) Send(_ context.Context, topic, key string, value []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(value),
	}

	_, _, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message to topic %s: %w", topic, err)
	}

	return nil
}

// Close gracefully closes the producer.
func (p *SaramaProducer) Close() error {
	return p.producer.Close()
}
