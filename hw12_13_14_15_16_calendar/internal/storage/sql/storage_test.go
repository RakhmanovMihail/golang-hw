package sql

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	dsn := "postgres://user:pass@localhost/db?sslmode=disable" //nolint:gosec // Test DSN
	_, err := New(dsn)
	assert.Error(t, err)
}

func TestStore_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	store := &Store{db: sqlxDB}

	ctx := context.Background()
	event := &storage.Event{
		Title:     "Test Event",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Hour),
		UserID:    1,
	}

	rows := sqlmock.NewRows([]string{"id", "title", "start_time", "end_time", "user_id"}).
		AddRow(uint64(1), event.Title, event.StartTime, event.EndTime, event.UserID)

	mock.ExpectQuery("INSERT INTO events").
		WithArgs(event.Title, event.StartTime, event.EndTime, event.UserID).
		WillReturnRows(rows)

	created, err := store.Create(ctx, event)

	require.NoError(t, err)
	assert.Equal(t, uint64(1), created.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestStore_Read(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	store := &Store{db: sqlxDB}

	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "title", "start_time", "end_time", "user_id"}).
		AddRow(uint64(1), "Event 1", time.Now(), time.Now().Add(time.Hour), 1)

	mock.ExpectQuery("SELECT id, title, start_time, end_time, user_id FROM events").
		WillReturnRows(rows)

	events, err := store.Read(ctx)

	require.NoError(t, err)
	assert.Len(t, events, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestStore_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	store := &Store{db: sqlxDB}

	ctx := context.Background()
	id := uint64(1)
	event := &storage.Event{
		Title:     "Updated Event",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Hour),
		UserID:    1,
	}

	rows := sqlmock.NewRows([]string{"id", "title", "start_time", "end_time", "user_id"}).
		AddRow(id, event.Title, event.StartTime, event.EndTime, event.UserID)

	mock.ExpectQuery("UPDATE events").
		WithArgs(event.Title, event.StartTime, event.EndTime, event.UserID, id).
		WillReturnRows(rows)

	updated, err := store.Update(ctx, id, event)

	require.NoError(t, err)
	assert.Equal(t, id, updated.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestStore_Update_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	store := &Store{db: sqlxDB}

	ctx := context.Background()
	id := uint64(999)
	event := &storage.Event{
		Title:     "Test",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Hour),
		UserID:    1,
	}

	mock.ExpectQuery("UPDATE events").
		WithArgs(event.Title, event.StartTime, event.EndTime, event.UserID, id).
		WillReturnError(sql.ErrNoRows)

	_, err = store.Update(ctx, id, event)

	require.ErrorIs(t, err, storage.ErrEventNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestStore_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	store := &Store{db: sqlxDB}

	ctx := context.Background()
	id := uint64(1)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(id)

	mock.ExpectQuery("DELETE FROM events").
		WithArgs(id).
		WillReturnRows(rows)

	err = store.Delete(ctx, id)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestStore_Delete_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	store := &Store{db: sqlxDB}

	ctx := context.Background()
	id := uint64(999)

	mock.ExpectQuery("DELETE FROM events").
		WithArgs(id).
		WillReturnError(sql.ErrNoRows)

	err = store.Delete(ctx, id)

	require.ErrorIs(t, err, storage.ErrEventNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestStore_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	store := &Store{db: sqlxDB}

	ctx := context.Background()
	id := uint64(1)

	rows := sqlmock.NewRows([]string{"id", "title", "start_time", "end_time", "user_id"}).
		AddRow(id, "Test Event", time.Now(), time.Now().Add(time.Hour), 1)

	mock.ExpectQuery("SELECT id, title, start_time, end_time, user_id FROM events").
		WithArgs(id).
		WillReturnRows(rows)

	event, err := store.GetByID(ctx, id)

	require.NoError(t, err)
	assert.Equal(t, id, event.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestStore_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	store := &Store{db: sqlxDB}

	ctx := context.Background()
	id := uint64(999)

	mock.ExpectQuery("SELECT id, title, start_time, end_time, user_id FROM events").
		WithArgs(id).
		WillReturnError(sql.ErrNoRows)

	_, err = store.GetByID(ctx, id)

	require.ErrorIs(t, err, storage.ErrEventNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}
