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
func (a *App) CreateEvent(ctx context.Context, id uint64, title string, startTime, endTime time.Time, userID int) (*storage.Event, error) {
	event, err := a.Storage.Create(ctx, &storage.Event{
		ID:        id,
		Title:     title,
		StartTime: startTime,
		EndTime:   endTime,
		UserID:    userID,
	})
	return event, err
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
		if *title == "" {
			return nil, storage.ErrInvalidEvent
		}
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

	// Validate: EndTime must be after StartTime
	if newEvent.EndTime.Before(newEvent.StartTime) || newEvent.EndTime.Equal(newEvent.StartTime) {
		return nil, storage.ErrInvalidEvent
	}

	return a.Storage.Update(ctx, id, &newEvent)
}

// DeleteEvent deletes an event by ID.
func (a *App) DeleteEvent(ctx context.Context, id uint64) error {
	return a.Storage.Delete(ctx, id)
}
