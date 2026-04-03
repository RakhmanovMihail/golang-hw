package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/sarama"
)

// SaramaConsumer implements Consumer interface using IBM/sarama Consumer Group.
type SaramaConsumer struct {
	group   sarama.ConsumerGroup
	topic   string
	groupID string
}

// NewSaramaConsumer creates a new Kafka consumer using Consumer Group API.
func NewSaramaConsumer(brokers []string, topic, groupID string) (*SaramaConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	group, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer group: %w", err)
	}

	return &SaramaConsumer{
		group:   group,
		topic:   topic,
		groupID: groupID,
	}, nil
}

// handler adapts kafka.Handler to use our Handler function type.
type handler struct {
	fn      Handler
	ctx     context.Context
	cancel  context.CancelFunc
}

func (h *handler) Setup(_ sarama.ConsumerGroupSession) error  { return nil }
func (h *handler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			kafkaMsg := Message{
				Topic: msg.Topic,
				Key:   string(msg.Key),
				Value: msg.Value,
			}
			if err := h.fn(h.ctx, kafkaMsg); err != nil {
				fmt.Printf("error handling message: %v\n", err)
			}
			session.MarkMessage(msg, "")
		case <-h.ctx.Done():
			return nil
		}
	}
}

// Start begins consuming messages and calls the handler for each.
func (c *SaramaConsumer) Start(ctx context.Context, handlerFn Handler) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	h := &handler{
		fn:     handlerFn,
		ctx:    ctx,
		cancel: cancel,
	}

	// Consume runs in a loop to handle rebalances
	for {
		if err := c.group.Consume(ctx, []string{c.topic}, h); err != nil {
			return fmt.Errorf("consumer group consume error: %w", err)
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// Close gracefully closes the consumer.
func (c *SaramaConsumer) Close() error {
	return c.group.Close()
}
