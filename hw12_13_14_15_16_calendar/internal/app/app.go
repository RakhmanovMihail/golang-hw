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
func (a *App) CreateEvent(
	ctx context.Context,
	title string,
	startTime, endTime time.Time,
	userID int64,
	description *string,
	notifyBefore *int32,
) (*storage.Event, error) {
	event, err := a.Storage.Create(ctx, &storage.Event{
		Title:        title,
		StartTime:    startTime,
		EndTime:      endTime,
		UserID:       userID,
		Description:  description,
		NotifyBefore: notifyBefore,
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
func (a *App) UpdateEvent(
	ctx context.Context,
	id uint64,
	title *string,
	startTime, endTime *time.Time,
	userID *int64,
	description *string,
	notifyBefore *int32,
) (*storage.Event, error) {
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
	if description != nil {
		newEvent.Description = description
	}
	if notifyBefore != nil {
		newEvent.NotifyBefore = notifyBefore
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

// GetEventsByDay returns events for a specific day.
func (a *App) GetEventsByDay(ctx context.Context, date time.Time) ([]storage.Event, error) {
	all, err := a.Storage.Read(ctx)
	if err != nil {
		return nil, err
	}

	var result []storage.Event
	for _, e := range all {
		if e.StartTime.YearDay() == date.YearDay() && e.StartTime.Year() == date.Year() {
			result = append(result, e)
		}
	}
	return result, nil
}

// GetEventsByWeek returns events for a specific week.
func (a *App) GetEventsByWeek(ctx context.Context, date time.Time) ([]storage.Event, error) {
	all, err := a.Storage.Read(ctx)
	if err != nil {
		return nil, err
	}

	// Calculate start of week (Monday)
	startOfWeek := date.AddDate(0, 0, -int(date.Weekday())+1)
	if startOfWeek.Weekday() == time.Sunday {
		startOfWeek = startOfWeek.AddDate(0, 0, -6)
	}
	startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, date.Location())
	endOfWeek := startOfWeek.AddDate(0, 0, 7)

	var result []storage.Event
	for _, e := range all {
		if (e.StartTime.After(startOfWeek) || e.StartTime.Equal(startOfWeek)) && e.StartTime.Before(endOfWeek) {
			result = append(result, e)
		}
	}
	return result, nil
}

// GetEventsByMonth returns events for a specific month.
func (a *App) GetEventsByMonth(ctx context.Context, date time.Time) ([]storage.Event, error) {
	all, err := a.Storage.Read(ctx)
	if err != nil {
		return nil, err
	}

	var result []storage.Event
	for _, e := range all {
		if e.StartTime.Year() == date.Year() && e.StartTime.Month() == date.Month() {
			result = append(result, e)
		}
	}
	return result, nil
}
