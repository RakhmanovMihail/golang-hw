package internalhttp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/app"
	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/logger"
	internalhttp "github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/server/http"
	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/storage/memory"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter(t *testing.T) (*chi.Mux, *app.App) {
	t.Helper()

	logg := logger.New(logger.LevelDebug)
	store := memory.New()
	app := app.New(*logg, store)
	apiServer := internalhttp.NewAPIServer(app)

	router := chi.NewRouter()
	internalhttp.RegisterAPIRoutes(router, apiServer)

	return router, app
}

func TestGetEvents_Empty(t *testing.T) {
	router, _ := setupTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/events", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "[]\n", w.Body.String())
}

func TestCreateEvent_Success(t *testing.T) {
	router, _ := setupTestRouter(t)

	now := time.Now().UTC()
	event := internalhttp.CreateEventRequest{
		Title:     "Test Event",
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		UserId:    1,
	}

	body, _ := json.Marshal(event)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/events", bytes.NewReader(body))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var created internalhttp.Event
	err := json.Unmarshal(w.Body.Bytes(), &created)
	require.NoError(t, err)
	assert.Equal(t, "Test Event", *created.Title)
}

func TestGetEvent_NotFound(t *testing.T) {
	router, _ := setupTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/events/999", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateEvent_Success(t *testing.T) {
	router, app := setupTestRouter(t)

	// Сначала создадим событие
	ctx := context.Background()
	now := time.Now().UTC()
	created, err := app.Storage.Create(ctx, &storage.Event{
		Title:     "Original",
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		UserID:    1,
	})
	require.NoError(t, err)

	// Теперь обновим (меняем заголовок и сдвигаем время, чтобы не было конфликта)
	newStart := now.Add(3 * time.Hour)
	newEnd := now.Add(4 * time.Hour)
	update := internalhttp.UpdateEventRequest{
		Title:     ptr("Updated"),
		StartTime: &newStart,
		EndTime:   &newEnd,
	}

	body, _ := json.Marshal(update)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/events/"+strconv.Itoa(int(created.ID)), bytes.NewReader(body))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Проверим, что событие обновилось
	var updated internalhttp.Event
	err = json.Unmarshal(w.Body.Bytes(), &updated)
	require.NoError(t, err)
	assert.Equal(t, "Updated", *updated.Title)
}

func TestDeleteEvent_Success(t *testing.T) {
	router, app := setupTestRouter(t)

	// Сначала создадим событие
	ctx := context.Background()
	created, err := app.Storage.Create(ctx, &storage.Event{
		Title:     "To Delete",
		StartTime: time.Now().Add(time.Hour),
		EndTime:   time.Now().Add(2 * time.Hour),
		UserID:    1,
	})
	require.NoError(t, err)

	// Теперь удалим
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/events/"+strconv.Itoa(int(created.ID)), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCreateEvent_DateBusy(t *testing.T) {
	router, app := setupTestRouter(t)

	ctx := context.Background()
	now := time.Now().UTC()
	_, err := app.Storage.Create(ctx, &storage.Event{
		Title:     "Busy Event",
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		UserID:    1,
	})
	require.NoError(t, err)

	// Попытка создать событие с пересекающимся временем
	event := internalhttp.CreateEventRequest{
		Title:     "Conflict Event",
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		UserId:    2,
	}

	body, _ := json.Marshal(event)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/events", bytes.NewReader(body))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestGetEvent_Success(t *testing.T) {
	router, app := setupTestRouter(t)

	ctx := context.Background()
	now := time.Now().UTC()
	created, err := app.Storage.Create(ctx, &storage.Event{
		Title:     "Test Event",
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		UserID:    1,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/events/"+strconv.Itoa(int(created.ID)), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var event internalhttp.Event
	err = json.Unmarshal(w.Body.Bytes(), &event)
	require.NoError(t, err)
	assert.Equal(t, "Test Event", *event.Title)
	assert.Equal(t, int64(created.ID), *event.Id)
}

func TestUpdateEvent_NotFound(t *testing.T) {
	router, _ := setupTestRouter(t)

	now := time.Now().UTC()
	update := internalhttp.UpdateEventRequest{
		Title:     ptr("Updated"),
		StartTime: ptr(now.Add(time.Hour)),
		EndTime:   ptr(now.Add(2 * time.Hour)),
	}

	body, _ := json.Marshal(update)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/events/999", bytes.NewReader(body))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteEvent_NotFound(t *testing.T) {
	router, _ := setupTestRouter(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/events/999", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func ptr[T any](v T) *T {
	return &v
}
