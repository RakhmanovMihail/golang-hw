package scheduler_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/kafka"
	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/scheduler"
	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProducer captures sent messages for testing.
type mockProducer struct {
	messages []kafka.Message
}

func (m *mockProducer) Send(_ context.Context, topic, key string, value []byte) error {
	m.messages = append(m.messages, kafka.Message{
		Topic: topic,
		Key:   key,
		Value: value,
	})
	return nil
}

func (m *mockProducer) Close() error { return nil }

func TestScheduler_SendNotifications(t *testing.T) {
	store := memory.New()
	producer := &mockProducer{}
	logg := logger.New(logger.LevelInfo)

	// Create an event with notify_before
	notifyBefore := int32(15)
	desc := "Test event"
	_, err := store.Create(context.Background(), &storage.Event{
		Title:        "Test Event",
		StartTime:    time.Now().Add(10 * time.Minute),
		EndTime:      time.Now().Add(20 * time.Minute),
		UserID:       1,
		Description:  &desc,
		NotifyBefore: &notifyBefore,
	})
	require.NoError(t, err)

	sched := scheduler.New(store, producer, *logg, "test-topic", 1*time.Second, 8760*time.Hour)

	// Run one tick
	ctx := context.Background()
	err = sched.Run(ctx)
	// Run will block, so we test indirectly via tick logic
	// For unit testing, we verify the producer received messages after a short run
	_ = sched // scheduler is designed for long-running use

	// Directly verify by creating a new scheduler and running a single tick
	// Since Run() blocks, we verify via the mock producer after context cancellation
	ctx, cancel := context.WithCancel(ctx)
	cancel() // Cancel immediately after first tick
	_ = sched.Run(ctx)

	// Verify notifications were sent
	require.NotEmpty(t, producer.messages, "Expected notifications to be sent")

	var msg kafka.NotificationMessage
	err = json.Unmarshal(producer.messages[0].Value, &msg)
	require.NoError(t, err)

	assert.Equal(t, "Test Event", msg.Title)
	assert.Equal(t, int64(1), msg.UserID)
}

func TestScheduler_CleanupOldEvents(t *testing.T) {
	store := memory.New()
	producer := &mockProducer{}
	logg := logger.New(logger.LevelInfo)

	// Create an old event
	oldTime := time.Now().Add(-2 * 365 * 24 * time.Hour) // 2 years ago
	_, err := store.Create(context.Background(), &storage.Event{
		Title:     "Old Event",
		StartTime: oldTime,
		EndTime:   oldTime.Add(time.Hour),
		UserID:    1,
	})
	require.NoError(t, err)

	// Create a recent event
	recentTime := time.Now().Add(-1 * time.Hour)
	_, err = store.Create(context.Background(), &storage.Event{
		Title:     "Recent Event",
		StartTime: recentTime,
		EndTime:   recentTime.Add(time.Hour),
		UserID:    2,
	})
	require.NoError(t, err)

	sched := scheduler.New(store, producer, *logg, "test-topic", 1*time.Second, 8760*time.Hour) // 1 year

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_ = sched.Run(ctx)

	// Verify old event was deleted, recent event remains
	events, err := store.Read(context.Background())
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "Recent Event", events[0].Title)
}
