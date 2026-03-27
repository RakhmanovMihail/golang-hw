package storage

import "context"

type Storage interface {
	Create(ctx context.Context, e *Event) (*Event, error)
	Read(ctx context.Context) ([]Event, error)
	Update(ctx context.Context, id uint64, e *Event) (*Event, error)
	Delete(ctx context.Context, id uint64) error
}
