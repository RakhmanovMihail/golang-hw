package app

import (
	"context"
	"time"

	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/storage"
)

// App represents the calendar application.
type App struct {
	Storage storage.Storage
	Logger  logger.Logger
}

// New creates a new App instance.
func New(logger logger.Logger, storage storage.Storage) *App {
	return &App{
		Logger:  logger,
		Storage: storage,
	}
}

// CreateEvent creates a new event.
func (a *App) CreateEvent(ctx context.Context, id uint64, title string, startTime, endTime time.Time, userID int) error {
	_, err := a.Storage.Create(ctx, &storage.Event{
		ID:        id,
		Title:     title,
		StartTime: startTime,
		EndTime:   endTime,
		UserID:    userID,
	})
	return err
}

// GetEvents returns all events.
func (a *App) GetEvents(ctx context.Context) ([]storage.Event, error) {
	return a.Storage.Read(ctx)
}

// GetEvent returns an event by ID.
func (a *App) GetEvent(ctx context.Context, id uint64) (*storage.Event, error) {
	return a.Storage.GetByID(ctx, id)
}

// UpdateEvent updates an existing event.
func (a *App) UpdateEvent(ctx context.Context, id uint64, title *string, startTime, endTime *time.Time, userID *int) (*storage.Event, error) {
	existing, err := a.Storage.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	newEvent := *existing
	if title != nil {
		newEvent.Title = *title
	}
	if startTime != nil {
		newEvent.StartTime = *startTime
	}
	if endTime != nil {
		newEvent.EndTime = *endTime
	}
	if userID != nil {
		newEvent.UserID = *userID
	}

	return a.Storage.Update(ctx, id, &newEvent)
}

// DeleteEvent deletes an event by ID.
func (a *App) DeleteEvent(ctx context.Context, id uint64) error {
	return a.Storage.Delete(ctx, id)
}
