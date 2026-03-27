package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/app"
	loggerpkg "github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testStorage struct {
	events map[uint64]*storage.Event
	err    error
}

var _ storage.Storage = (*testStorage)(nil)

func (s *testStorage) Create(ctx context.Context, event *storage.Event) (*storage.Event, error) {
	// ✅ ПРЯМАЯ проверка ДО time.After!
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// ✅ Большие задержки для надёжности
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(10 * time.Millisecond):
	}

	if s.err != nil {
		return nil, s.err
	}

	if event.ID == 0 {
		return nil, errors.New("event ID cannot be zero")
	}

	if s.events == nil {
		s.events = make(map[uint64]*storage.Event)
	}

	for _, existing := range s.events {
		if event.UserID == existing.UserID &&
			event.StartTime.Before(existing.EndTime) &&
			event.EndTime.After(existing.StartTime) {
			return nil, storage.ErrDateBusy
		}
	}

	s.events[event.ID] = event
	return event, nil
}

func (s *testStorage) Read(ctx context.Context) ([]storage.Event, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(10 * time.Millisecond):
	}
	if s.err != nil {
		return nil, s.err
	}
	events := make([]storage.Event, 0, len(s.events))
	for _, event := range s.events {
		events = append(events, *event)
	}
	return events, nil
}

func (s *testStorage) Update(ctx context.Context, id uint64, event *storage.Event) (*storage.Event, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(10 * time.Millisecond):
	}
	if s.err != nil {
		return nil, s.err
	}
	if _, exists := s.events[id]; !exists {
		return nil, storage.ErrEventNotFound
	}
	s.events[id] = event
	return event, nil
}

func (s *testStorage) Delete(ctx context.Context, id uint64) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(10 * time.Millisecond):
	}
	if s.err != nil {
		return s.err
	}
	if _, exists := s.events[id]; !exists {
		return storage.ErrEventNotFound
	}
	delete(s.events, id)
	return nil
}

func TestApp_New(t *testing.T) {
	loggerInst := loggerpkg.New(loggerpkg.LevelInfo)
	storage := &testStorage{}
	app := app.New(*loggerInst, storage)
	require.NotNil(t, app)
	assert.Equal(t, loggerInst.Level, app.Logger.Level)
}

func TestApp_CreateEvent(t *testing.T) {
	tests := []struct {
		name            string
		setupStorage    func(*testStorage)
		ctxSetup        func() context.Context
		id              uint64
		title           string
		wantErrContains string
		wantEventsLen   int
	}{
		{
			name:            "успешное создание события",
			ctxSetup:        func() context.Context { return context.Background() },
			id:              1,
			title:           "Встреча с клиентом",
			wantErrContains: "",
			wantEventsLen:   1,
		},
		{
			name: "контекст с таймаутом",
			ctxSetup: func() context.Context {
				ctx, _ := context.WithTimeout(context.Background(), 5*time.Millisecond)
				return ctx
			},
			id:              1,
			title:           "test",
			wantErrContains: "context deadline exceeded",
			wantEventsLen:   0,
		},
		{
			name: "отменённый контекст",
			ctxSetup: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			id:              1,
			title:           "test",
			wantErrContains: "context canceled",
			wantEventsLen:   0,
		},
		{
			name:            "ошибка storage",
			setupStorage:    func(ts *testStorage) { ts.err = storage.ErrDateBusy },
			ctxSetup:        func() context.Context { return context.Background() },
			id:              1,
			title:           "test",
			wantErrContains: "date already busy",
			wantEventsLen:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &testStorage{}
			if tt.setupStorage != nil {
				tt.setupStorage(storage)
			}

			loggerInst := loggerpkg.New(loggerpkg.LevelInfo)
			appInst := app.New(*loggerInst, storage)

			ctx := tt.ctxSetup()
			err := appInst.CreateEvent(ctx, tt.id, tt.title)

			if tt.wantErrContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrContains)
			} else {
				require.NoError(t, err)
			}

			assert.Len(t, storage.events, tt.wantEventsLen)
		})
	}
}
