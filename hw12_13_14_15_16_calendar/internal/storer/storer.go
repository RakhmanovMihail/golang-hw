// Package storer implements the storer process that consumes notifications
// from Kafka and saves them to the database.
package storer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/kafka"
	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/storage"
)

// Storer consumes notifications from Kafka and saves them to the database.
type Storer struct {
	consumer kafka.Consumer
	storage  storage.Storage
	logger   logger.Logger
}

// New creates a new Storer.
func New(consumer kafka.Consumer, store storage.Storage, logger logger.Logger) *Storer {
	return &Storer{
		consumer: consumer,
		storage:  store,
		logger:   logger,
	}
}

// Run starts consuming notifications.
func (s *Storer) Run(ctx context.Context) error {
	s.logger.Info("storer started")

	handler := func(ctx context.Context, msg kafka.Message) error {
		return s.handleMessage(ctx, msg)
	}

	if err := s.consumer.Start(ctx, handler); err != nil {
		return fmt.Errorf("consumer start: %w", err)
	}

	s.logger.Info("storer stopped")
	return nil
}

func (s *Storer) handleMessage(_ context.Context, msg kafka.Message) error {
	var notification kafka.NotificationMessage
	if err := json.Unmarshal(msg.Value, &notification); err != nil {
		return fmt.Errorf("unmarshal notification: %w", err)
	}

	n := &storage.Notification{
		EventID:   notification.EventID,
		Title:     notification.Title,
		EventDate: notification.EventDate,
		UserID:    notification.UserID,
	}

	if err := s.storage.SaveNotification(context.Background(), n); err != nil {
		return fmt.Errorf("save notification: %w", err)
	}

	s.logger.Info(fmt.Sprintf("saved notification for event %d, user %d", n.EventID, n.UserID))
	return nil
}
