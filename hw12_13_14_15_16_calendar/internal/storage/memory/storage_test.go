package memory

import (
	"context"
	"testing"
	"time"

	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	store := New()
	assert.NotNil(t, store)
}

func TestStore_CRUD(t *testing.T) {
	store := New().(*Store)
	ctx := context.Background()

	now := time.Now()
	event1 := &storage.Event{
		Title:     "Event 1",
		StartTime: now.Add(2 * time.Hour),
		EndTime:   now.Add(3 * time.Hour),
		UserID:    123,
	}

	// CREATE
	created1, err := store.Create(ctx, event1)
	require.NoError(t, err)
	assert.NotZero(t, created1.ID)

	// READ
	events, err := store.Read(ctx)
	require.NoError(t, err)
	assert.Len(t, events, 1)

	// GET BY ID
	got, err := store.GetByID(ctx, created1.ID)
	require.NoError(t, err)
	assert.Equal(t, event1.Title, got.Title)

	// UPDATE - НЕ пересекается
	event2 := &storage.Event{
		Title:     "Updated Event",
		StartTime: now.Add(4 * time.Hour), // ✅ После первого события
		EndTime:   now.Add(5 * time.Hour),
		UserID:    123,
	}
	updated, err := store.Update(ctx, created1.ID, event2)
	require.NoError(t, err)
	assert.Equal(t, "Updated Event", updated.Title)

	// DELETE
	err = store.Delete(ctx, created1.ID)
	require.NoError(t, err)

	_, err = store.GetByID(ctx, created1.ID)
	assert.ErrorIs(t, err, storage.ErrEventNotFound)
}

func TestStore_Overlap(t *testing.T) {
	store := New().(*Store)
	ctx := context.Background()

	now := time.Now()

	// Первое событие
	event1 := &storage.Event{
		Title:     "Event 1",
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		UserID:    123,
	}
	_, err := store.Create(ctx, event1)
	require.NoError(t, err)

	// ✅ Пересекающееся - ErrDateBusy
	overlap1 := &storage.Event{
		Title:     "Overlap 1",
		StartTime: now.Add(1*time.Hour + 30*time.Minute),
		EndTime:   now.Add(1*time.Hour + 45*time.Minute),
		UserID:    456, // ✅ Любой UserID - overlap для ВСЕХ
	}
	_, err = store.Create(ctx, overlap1)
	assert.ErrorIs(t, err, storage.ErrDateBusy)

	// ✅ НЕ пересекающееся
	nonOverlap := &storage.Event{
		Title:     "Non Overlap",
		StartTime: now.Add(3 * time.Hour),
		EndTime:   now.Add(4 * time.Hour),
		UserID:    789,
	}
	_, err = store.Create(ctx, nonOverlap)
	assert.NoError(t, err)
}

func TestStore_SortRead(t *testing.T) {
	store := New().(*Store)
	ctx := context.Background()

	now := time.Now()
	eventsData := []struct {
		title string
		start time.Time
	}{
		{"Event 3", now.Add(3 * time.Hour)},
		{"Event 1", now.Add(time.Hour)},
		{"Event 2", now.Add(2 * time.Hour)},
	}

	var createdIDs []uint64
	for _, data := range eventsData {
		e := &storage.Event{
			Title:     data.title,
			StartTime: data.start,
			EndTime:   data.start.Add(time.Hour),
			UserID:    123,
		}
		created, err := store.Create(ctx, e)
		require.NoError(t, err)
		createdIDs = append(createdIDs, created.ID)
	}

	// ✅ Сортировка по StartTime
	readEvents, err := store.Read(ctx)
	require.NoError(t, err)
	assert.Len(t, readEvents, 3)

	assert.Equal(t, "Event 1", readEvents[0].Title)
	assert.Equal(t, "Event 2", readEvents[1].Title)
	assert.Equal(t, "Event 3", readEvents[2].Title)
}

func TestStore_Errors(t *testing.T) {
	store := New().(*Store)
	ctx := context.Background()

	// Update несуществующего
	_, err := store.Update(ctx, 999, &storage.Event{})
	assert.ErrorIs(t, err, storage.ErrEventNotFound)

	// Delete несуществующего
	_ = store.Delete(ctx, 999)
	assert.ErrorIs(t, err, storage.ErrEventNotFound)

	// GetByID несуществующего
	_, err = store.GetByID(ctx, 999)
	assert.ErrorIs(t, err, storage.ErrEventNotFound)
}

func TestStore_ContextIgnored(t *testing.T) {
	store := New().(*Store)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// ✅ Memory store игнорирует контекст (нет select)
	_, err := store.Create(ctx, &storage.Event{})
	assert.NoError(t, err) // ✅ Memory НЕ отменяется
}
