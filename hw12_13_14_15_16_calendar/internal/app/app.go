package app

import (
	"context"

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
func (a *App) CreateEvent(ctx context.Context, id uint64, title string) error {
	_, err := a.Storage.Create(ctx,
		&storage.Event{
			ID:    id,
			Title: title,
		},
	)
	return err
}
