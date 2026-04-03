// Package scheduler implements the scheduler process that scans events,
// sends notifications via Kafka, and cleans up old events.
package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/kafka"
	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/storage"
)

// Scheduler scans events and sends notifications via Kafka.
type Scheduler struct {
	storage       storage.Storage
	producer      kafka.Producer
	logger        logger.Logger
	topic         string
	checkInterval time.Duration
	cleanupOlder  time.Duration
}

// New creates a new Scheduler.
func New(
	store storage.Storage,
	producer kafka.Producer,
	logger logger.Logger,
	topic string,
	checkInterval time.Duration,
	cleanupOlder time.Duration,
) *Scheduler {
	return &Scheduler{
		storage:       store,
		producer:      producer,
		logger:        logger,
		topic:         topic,
		checkInterval: checkInterval,
		cleanupOlder:  cleanupOlder,
	}
}

// Run starts the scheduler loop.
func (s *Scheduler) Run(ctx context.Context) error {
	s.logger.Info("scheduler started")

	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()

	// Run immediately on startup
	if err := s.tick(ctx); err != nil {
		s.logger.Error("scheduler tick error: " + err.Error())
	}

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("scheduler stopped")
			return ctx.Err()
		case <-ticker.C:
			if err := s.tick(ctx); err != nil {
				s.logger.Error("scheduler tick error: " + err.Error())
			}
		}
	}
}

func (s *Scheduler) tick(ctx context.Context) error {
	now := time.Now()

	// Send notifications
	if err := s.sendNotifications(ctx, now); err != nil {
		return fmt.Errorf("send notifications: %w", err)
	}

	// Clean up old events
	if err := s.cleanupOldEvents(ctx, now); err != nil {
		return fmt.Errorf("cleanup old events: %w", err)
	}

	return nil
}

func (s *Scheduler) sendNotifications(ctx context.Context, now time.Time) error {
	events, err := s.storage.GetEventsForNotify(ctx, now)
	if err != nil {
		return fmt.Errorf("get events for notify: %w", err)
	}

	sent := 0
	for _, e := range events {
		msg := kafka.NotificationMessage{
			EventID:   e.ID,
			Title:     e.Title,
			EventDate: e.StartTime,
			UserID:    e.UserID,
		}

		data, err := json.Marshal(msg)
		if err != nil {
			s.logger.Error("marshal notification: " + err.Error())
			continue
		}

		key := fmt.Sprintf("%d", e.ID)
		if err := s.producer.Send(ctx, s.topic, key, data); err != nil {
			s.logger.Error("send to kafka: " + err.Error())
			continue
		}

		sent++
	}

	if sent > 0 {
		s.logger.Info(fmt.Sprintf("sent %d notifications", sent))
	}

	return nil
}

func (s *Scheduler) cleanupOldEvents(ctx context.Context, now time.Time) error {
	cutoffTime := now.Add(-s.cleanupOlder)
	count, err := s.storage.DeleteOldEvents(ctx, cutoffTime)
	if err != nil {
		return err
	}

	if count > 0 {
		s.logger.Info(fmt.Sprintf("deleted %d old events", count))
	}

	return nil
}
