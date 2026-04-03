package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/storage"
)

// Store is an in-memory implementation of storage.Storage.
type Store struct {
	mu     sync.RWMutex
	events map[uint64]storage.Event
	nextID uint64
}

func New() storage.Storage {
	return &Store{
		events: make(map[uint64]storage.Event),
		nextID: 1,
	}
}

// Create creates a new event in the store.
func (s *Store) Create(ctx context.Context, e *storage.Event) (*storage.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.checkOverlap(*e); err != nil {
		return nil, err
	}

	e.ID = s.nextID
	s.nextID++
	s.events[e.ID] = *e

	return e, nil
}

// Read returns all events from the store.
func (s *Store) Read(_ context.Context) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events := make([]storage.Event, 0, len(s.events))
	for _, e := range s.events {
		events = append(events, e)
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].StartTime.Before(events[j].StartTime)
	})

	return events, nil
}

// Update updates an existing event in the store.
func (s *Store) Update(ctx context.Context, id uint64, e *storage.Event) (*storage.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[id]; !exists {
		return nil, storage.ErrEventNotFound
	}

	if err := s.checkOverlap(*e); err != nil {
		return nil, err
	}

	e.ID = id
	s.events[id] = *e

	return e, nil
}

// Delete deletes an event from the store.
func (s *Store) Delete(ctx context.Context, id uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[id]; !exists {
		return storage.ErrEventNotFound
	}

	delete(s.events, id)
	return nil
}

// GetByID returns an event by ID.
func (s *Store) GetByID(ctx context.Context, id uint64) (*storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if event, exists := s.events[id]; exists {
		return &event, nil
	}

	return nil, storage.ErrEventNotFound
}

func (s *Store) checkOverlap(e storage.Event) error {
	for _, event := range s.events {
		if s.eventsOverlap(event, e) {
			return storage.ErrDateBusy
		}
	}
	return nil
}

func (s *Store) eventsOverlap(a, b storage.Event) bool {
	return a.StartTime.Before(b.EndTime) && b.StartTime.Before(a.EndTime)
}
