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
	// Прямая проверка ДО time.After!
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Большие задержки для надёжности
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(10 * time.Millisecond):
	}

	if s.err != nil {
		return nil, s.err
	}

	// Генерируем ID на стороне storage (как и должно быть)
	nextID := uint64(1)
	for id := range s.events {
		if id >= nextID {
			nextID = id + 1
		}
	}
	event.ID = nextID

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

func (s *testStorage) GetByID(ctx context.Context, id uint64) (*storage.Event, error) {
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
	if event, exists := s.events[id]; exists {
		return event, nil
	}
	return nil, storage.ErrEventNotFound
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
		title           string
		startTime       time.Time
		endTime         time.Time
		userID          int64
		description     *string
		notifyBefore    *int32
		wantErrContains string
		wantEventsLen   int
	}{
		{
			name:            "успешное создание события",
			ctxSetup:        context.Background,
			title:           "Встреча с клиентом",
			startTime:       time.Date(2026, 3, 27, 10, 0, 0, 0, time.UTC),
			endTime:         time.Date(2026, 3, 27, 11, 0, 0, 0, time.UTC),
			userID:          1,
			wantErrContains: "",
			wantEventsLen:   1,
		},
		{
			name: "контекст с таймаутом",
			ctxSetup: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
				t.Cleanup(cancel)
				return ctx
			},
			title:           "test",
			startTime:       time.Date(2026, 3, 27, 10, 0, 0, 0, time.UTC),
			endTime:         time.Date(2026, 3, 27, 11, 0, 0, 0, time.UTC),
			userID:          1,
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
			title:           "test",
			startTime:       time.Date(2026, 3, 27, 10, 0, 0, 0, time.UTC),
			endTime:         time.Date(2026, 3, 27, 11, 0, 0, 0, time.UTC),
			userID:          1,
			wantErrContains: "context canceled",
			wantEventsLen:   0,
		},
		{
			name:            "ошибка storage",
			setupStorage:    func(ts *testStorage) { ts.err = storage.ErrDateBusy },
			ctxSetup:        context.Background,
			title:           "test",
			startTime:       time.Date(2026, 3, 27, 10, 0, 0, 0, time.UTC),
			endTime:         time.Date(2026, 3, 27, 11, 0, 0, 0, time.UTC),
			userID:          1,
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
			event, err := appInst.CreateEvent(
				ctx, tt.title, tt.startTime, tt.endTime,
				tt.userID, tt.description, tt.notifyBefore,
			)

			if tt.wantErrContains != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.wantErrContains)
				assert.Nil(t, event)
			} else {
				require.NoError(t, err)
				require.NotNil(t, event)
				assert.Equal(t, tt.title, event.Title)
				assert.NotZero(t, event.ID, "event ID should be generated by storage")
			}

			assert.Len(t, storage.events, tt.wantEventsLen)
		})
	}
}

func TestApp_GetEvents(t *testing.T) {
	tests := []struct {
		name            string
		setupStorage    func(*testStorage)
		ctxSetup        func() context.Context
		wantErrContains string
		wantLen         int
	}{
		{
			name: "успешное получение всех событий",
			setupStorage: func(ts *testStorage) {
				ts.events = map[uint64]*storage.Event{
					1: {ID: 1, Title: "Событие 1"},
					2: {ID: 2, Title: "Событие 2"},
				}
			},
			ctxSetup:        context.Background,
			wantErrContains: "",
			wantLen:         2,
		},
		{
			name:            "пустой список событий",
			setupStorage:    func(ts *testStorage) { ts.events = make(map[uint64]*storage.Event) },
			ctxSetup:        context.Background,
			wantErrContains: "",
			wantLen:         0,
		},
		{
			name: "контекст с таймаутом",
			setupStorage: func(ts *testStorage) {
				ts.events = map[uint64]*storage.Event{
					1: {ID: 1, Title: "Событие 1"},
				}
			},
			ctxSetup: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
				t.Cleanup(cancel)
				return ctx
			},
			wantErrContains: "context deadline exceeded",
			wantLen:         0,
		},
		{
			name:            "ошибка storage",
			setupStorage:    func(ts *testStorage) { ts.err = errors.New("storage error") },
			ctxSetup:        context.Background,
			wantErrContains: "storage error",
			wantLen:         0,
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
			events, err := appInst.GetEvents(ctx)

			if tt.wantErrContains != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.wantErrContains)
				assert.Nil(t, events)
			} else {
				require.NoError(t, err)
				assert.Len(t, events, tt.wantLen)
			}
		})
	}
}

func TestApp_GetEvent(t *testing.T) {
	tests := []struct {
		name            string
		setupStorage    func(*testStorage)
		ctxSetup        func() context.Context
		id              uint64
		wantErrContains string
		wantEventID     uint64
	}{
		{
			name: "успешное получение события",
			setupStorage: func(ts *testStorage) {
				ts.events = map[uint64]*storage.Event{
					1: {ID: 1, Title: "Событие 1"},
				}
			},
			ctxSetup:        context.Background,
			id:              1,
			wantErrContains: "",
			wantEventID:     1,
		},
		{
			name:            "событие не найдено",
			setupStorage:    func(ts *testStorage) { ts.events = make(map[uint64]*storage.Event) },
			ctxSetup:        context.Background,
			id:              999,
			wantErrContains: "event not found",
			wantEventID:     0,
		},
		{
			name: "контекст с таймаутом",
			setupStorage: func(ts *testStorage) {
				ts.events = map[uint64]*storage.Event{
					1: {ID: 1, Title: "Событие 1"},
				}
			},
			ctxSetup: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
				t.Cleanup(cancel)
				return ctx
			},
			id:              1,
			wantErrContains: "context deadline exceeded",
			wantEventID:     0,
		},
		{
			name:            "ошибка storage",
			setupStorage:    func(ts *testStorage) { ts.err = errors.New("storage error") },
			ctxSetup:        context.Background,
			id:              1,
			wantErrContains: "storage error",
			wantEventID:     0,
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
			event, err := appInst.GetEvent(ctx, tt.id)

			if tt.wantErrContains != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.wantErrContains)
				assert.Nil(t, event)
			} else {
				require.NoError(t, err)
				require.NotNil(t, event)
				assert.Equal(t, tt.wantEventID, event.ID)
			}
		})
	}
}

func TestApp_UpdateEvent(t *testing.T) {
	tests := []struct {
		name            string
		setupStorage    func(*testStorage)
		ctxSetup        func() context.Context
		id              uint64
		title           *string
		startTime       *time.Time
		endTime         *time.Time
		userID          *int64
		description     *string
		notifyBefore    *int32
		wantErrContains string
	}{
		{
			name: "успешное обновление события",
			setupStorage: func(ts *testStorage) {
				ts.events = map[uint64]*storage.Event{
					1: {
						ID: 1, Title: "Старое название",
						StartTime: time.Date(2026, 3, 27, 10, 0, 0, 0, time.UTC),
						EndTime:   time.Date(2026, 3, 27, 11, 0, 0, 0, time.UTC),
					},
				}
			},
			ctxSetup:        context.Background,
			id:              1,
			title:           ptrToString("Новое название"),
			startTime:       nil,
			endTime:         nil,
			userID:          nil,
			wantErrContains: "",
		},
		{
			name: "обновление с валидацией времени",
			setupStorage: func(ts *testStorage) {
				ts.events = map[uint64]*storage.Event{
					1: {
						ID: 1, Title: "Событие",
						StartTime: time.Date(2026, 3, 27, 10, 0, 0, 0, time.UTC),
						EndTime:   time.Date(2026, 3, 27, 11, 0, 0, 0, time.UTC),
					},
				}
			},
			ctxSetup:        context.Background,
			id:              1,
			title:           nil,
			startTime:       ptrToTime(time.Date(2026, 3, 27, 12, 0, 0, 0, time.UTC)),
			endTime:         ptrToTime(time.Date(2026, 3, 27, 11, 0, 0, 0, time.UTC)),
			userID:          nil,
			wantErrContains: "invalid event",
		},
		{
			name: "обновление с пустым заголовком",
			setupStorage: func(ts *testStorage) {
				ts.events = map[uint64]*storage.Event{
					1: {
						ID: 1, Title: "Событие",
						StartTime: time.Date(2026, 3, 27, 10, 0, 0, 0, time.UTC),
						EndTime:   time.Date(2026, 3, 27, 11, 0, 0, 0, time.UTC),
					},
				}
			},
			ctxSetup:        context.Background,
			id:              1,
			title:           ptrToString(""),
			startTime:       nil,
			endTime:         nil,
			userID:          nil,
			wantErrContains: "invalid event",
		},
		{
			name:            "событие не найдено",
			setupStorage:    func(ts *testStorage) { ts.events = make(map[uint64]*storage.Event) },
			ctxSetup:        context.Background,
			id:              999,
			title:           ptrToString("Новое название"),
			startTime:       nil,
			endTime:         nil,
			userID:          nil,
			wantErrContains: "event not found",
		},
		{
			name: "контекст с таймаутом",
			setupStorage: func(ts *testStorage) {
				ts.events = map[uint64]*storage.Event{
					1: {ID: 1, Title: "Событие"},
				}
			},
			ctxSetup: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
				t.Cleanup(cancel)
				return ctx
			},
			id:              1,
			title:           ptrToString("Новое название"),
			startTime:       nil,
			endTime:         nil,
			userID:          nil,
			wantErrContains: "context deadline exceeded",
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
			event, err := appInst.UpdateEvent(
				ctx, tt.id, tt.title, tt.startTime, tt.endTime,
				tt.userID, tt.description, tt.notifyBefore,
			)

			if tt.wantErrContains != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.wantErrContains)
				assert.Nil(t, event)
			} else {
				require.NoError(t, err)
				require.NotNil(t, event)
				if tt.title != nil {
					assert.Equal(t, *tt.title, event.Title)
				}
			}
		})
	}
}

func TestApp_DeleteEvent(t *testing.T) {
	tests := []struct {
		name            string
		setupStorage    func(*testStorage)
		ctxSetup        func() context.Context
		id              uint64
		wantErrContains string
	}{
		{
			name: "успешное удаление события",
			setupStorage: func(ts *testStorage) {
				ts.events = map[uint64]*storage.Event{
					1: {ID: 1, Title: "Событие"},
				}
			},
			ctxSetup:        context.Background,
			id:              1,
			wantErrContains: "",
		},
		{
			name:            "событие не найдено",
			setupStorage:    func(ts *testStorage) { ts.events = make(map[uint64]*storage.Event) },
			ctxSetup:        context.Background,
			id:              999,
			wantErrContains: "event not found",
		},
		{
			name: "контекст с таймаутом",
			setupStorage: func(ts *testStorage) {
				ts.events = map[uint64]*storage.Event{
					1: {ID: 1, Title: "Событие"},
				}
			},
			ctxSetup: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
				t.Cleanup(cancel)
				return ctx
			},
			id:              1,
			wantErrContains: "context deadline exceeded",
		},
		{
			name:            "ошибка storage",
			setupStorage:    func(ts *testStorage) { ts.err = errors.New("storage error") },
			ctxSetup:        context.Background,
			id:              1,
			wantErrContains: "storage error",
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
			err := appInst.DeleteEvent(ctx, tt.id)

			if tt.wantErrContains != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.wantErrContains)
			} else {
				require.NoError(t, err)
				assert.Len(t, storage.events, 0)
			}
		})
	}
}

// Helper functions.
func ptrToString(s string) *string {
	return &s
}

func ptrToTime(t time.Time) *time.Time {
	return &t
}
