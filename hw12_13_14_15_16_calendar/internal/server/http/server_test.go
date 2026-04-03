package internalhttp

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockLogger struct {
	messages []string
}

func (m *mockLogger) Debug(msg string) { m.messages = append(m.messages, msg) }
func (m *mockLogger) Info(msg string)  { m.messages = append(m.messages, msg) }
func (m *mockLogger) Warn(msg string)  { m.messages = append(m.messages, msg) }
func (m *mockLogger) Error(msg string) { m.messages = append(m.messages, msg) }

type mockApp struct{}

func TestNewServer(t *testing.T) {
	mockLog := &mockLogger{}
	server := NewServer(mockLog, &mockApp{}, ":8080")

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

func TestQueryString(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/events?user=123", "?user=123"},
		{"/events", ""},
		{"/events?", ""},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), "GET", tt.path, nil)
			result := queryString(req)
			assert.Equal(t, tt.expected, result)
		})
	}
}
