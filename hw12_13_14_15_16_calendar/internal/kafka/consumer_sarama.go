package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/sarama"
)

// SaramaConsumer implements Consumer interface using IBM/sarama.
type SaramaConsumer struct {
	consumer sarama.Consumer
	topic    string
	groupID  string
}

// NewSaramaConsumer creates a new Kafka consumer.
func NewSaramaConsumer(brokers []string, topic, groupID string) (*SaramaConsumer, error) {
	consumer, err := sarama.NewConsumer(brokers, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	return &SaramaConsumer{
		consumer: consumer,
		topic:    topic,
		groupID:  groupID,
	}, nil
}

// Start begins consuming messages in a loop and calls the handler for each.
func (c *SaramaConsumer) Start(ctx context.Context, handler Handler) error {
	partitions, err := c.consumer.Partitions(c.topic)
	if err != nil {
		return fmt.Errorf("failed to get partitions: %w", err)
	}

	for _, partition := range partitions {
		partitionConsumer, err := c.consumer.ConsumePartition(c.topic, partition, sarama.OffsetNewest)
		if err != nil {
			return fmt.Errorf("failed to consume partition %d: %w", partition, err)
		}

		go func(pc sarama.PartitionConsumer) {
			defer pc.Close()
			for {
				select {
				case msg := <-pc.Messages():
					kafkaMsg := Message{
						Topic: msg.Topic,
						Key:   string(msg.Key),
						Value: msg.Value,
					}
					if err := handler(ctx, kafkaMsg); err != nil {
						// Log error but continue consuming
						fmt.Printf("error handling message: %v\n", err)
					}
					time.Sleep(10 * time.Millisecond) // Small delay
				case <-ctx.Done():
					return
				}
			}
		}(partitionConsumer)
	}

	// Block until context is done
	<-ctx.Done()
	return nil
}

// Close gracefully closes the consumer.
func (c *SaramaConsumer) Close() error {
	return c.consumer.Close()
}
