package internalhttp

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	app "github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/app"
	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/stretchr/testify/assert"
)

type mockLogger struct {
	messages []string
}

func (m *mockLogger) Debug(msg string) { m.messages = append(m.messages, msg) }
func (m *mockLogger) Info(msg string)  { m.messages = append(m.messages, msg) }
func (m *mockLogger) Warn(msg string)  { m.messages = append(m.messages, msg) }
func (m *mockLogger) Error(msg string) { m.messages = append(m.messages, msg) }

type mockStorage struct{}

func (m *mockStorage) Create(ctx context.Context, event *storage.Event) (*storage.Event, error) {
	return event, nil
}

func (m *mockStorage) Read(ctx context.Context) ([]storage.Event, error) {
	return []storage.Event{}, nil
}

func (m *mockStorage) GetByID(ctx context.Context, id uint64) (*storage.Event, error) {
	return nil, storage.ErrEventNotFound
}

func (m *mockStorage) Update(ctx context.Context, id uint64, event *storage.Event) (*storage.Event, error) {
	return event, nil
}

func (m *mockStorage) Delete(ctx context.Context, id uint64) error {
	return nil
}

func (m *mockStorage) GetEventsForNotify(_ context.Context, _ time.Time) ([]storage.Event, error) {
	return nil, nil
}

func (m *mockStorage) DeleteOldEvents(_ context.Context, _ time.Time) (int64, error) {
	return 0, nil
}

func (m *mockStorage) SaveNotification(_ context.Context, _ *storage.Notification) error {
	return nil
}

func (m *mockStorage) Close(_ context.Context) error {
	return nil
}

func TestNewServer(t *testing.T) {
	mockLog := &logger.Logger{Level: logger.LevelInfo}
	mockStore := &mockStorage{}
	serverApp := app.New(*mockLog, mockStore)
	server := NewHTTPServer(mockLog, serverApp, ":8080")

	assert.NotNil(t, server)
	assert.Equal(t, ":8080", server.addr)
	assert.NotNil(t, server.server)
}

func TestHelloHandler(t *testing.T) {
	mockLog := &mockLogger{}
	handler := helloHandler(mockLog)

	rr := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), "GET", "/", nil)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "text/plain", rr.Header().Get("Content-Type"))
	body, _ := io.ReadAll(rr.Body)
	assert.Equal(t, "hello-world", string(body))
	assert.Contains(t, mockLog.messages, "hello endpoint called")
}
