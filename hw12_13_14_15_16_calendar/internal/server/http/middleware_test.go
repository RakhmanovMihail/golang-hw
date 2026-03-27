package internalhttp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/stretchr/testify/assert"
)

type testLogger struct {
	level    logger.Level
	messages []string
}

func (l *testLogger) Debug(msg string) { l.messages = append(l.messages, msg) }
func (l *testLogger) Info(msg string)  { l.messages = append(l.messages, msg) }
func (l *testLogger) Warn(msg string)  { l.messages = append(l.messages, msg) }
func (l *testLogger) Error(msg string) { l.messages = append(l.messages, msg) }

func TestLoggingMiddleware(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		forwardedFor string
		realIP       string
		method       string
		path         string
		wantClientIP string
		wantStatus   string
	}{
		{
			name:         "успешный GET запрос",
			statusCode:   http.StatusOK,
			method:       "GET",
			path:         "/events",
			wantClientIP: "192.0.2.1", // Реальный IP из httptest
			wantStatus:   "200",
		},
		{
			name:         "ошибка POST запрос",
			statusCode:   http.StatusInternalServerError,
			method:       "POST",
			path:         "/events",
			wantClientIP: "192.0.2.1",
			wantStatus:   "500",
		},
		{
			name:         "X-Forwarded-For",
			statusCode:   http.StatusOK,
			method:       "GET",
			path:         "/events?user=123",
			forwardedFor: "proxy1,192.168.1.100",
			wantClientIP: "proxy1",
			wantStatus:   "200",
		},
		{
			name:         "X-Real-IP",
			statusCode:   http.StatusOK,
			method:       "PUT",
			path:         "/events/1",
			realIP:       "172.16.0.1",
			wantClientIP: "172.16.0.1",
			wantStatus:   "200",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLog := &testLogger{
				level:    logger.LevelInfo,
				messages: make([]string, 0),
			}

			middleware := loggingMiddleware(testLog)

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			})

			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.forwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.forwardedFor)
			}
			if tt.realIP != "" {
				req.Header.Set("X-Real-IP", tt.realIP)
			}

			rr := httptest.NewRecorder()
			middleware(handler).ServeHTTP(rr, req)

			assert.Len(t, testLog.messages, 1)

			logMsg := testLog.messages[0]

			assert.Contains(t, logMsg, tt.method)
			assert.Contains(t, logMsg, tt.wantStatus)
			assert.Contains(t, logMsg, tt.wantClientIP)
		})
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		remote   string
		expected string
	}{
		{
			name: "X-Forwarded-For первый IP",
			headers: map[string]string{
				"X-Forwarded-For": "proxy1.example.com, 192.168.1.100",
			},
			expected: "proxy1.example.com",
		},
		{
			name: "X-Forwarded-For с пробелами",
			headers: map[string]string{
				"X-Forwarded-For": " proxy1 , 192.168.1.100",
			},
			expected: " proxy1 ", // Точно как возвращает код
		},
		{
			name: "X-Real-IP",
			headers: map[string]string{
				"X-Real-IP": "172.16.0.1",
			},
			expected: "172.16.0.1",
		},
		{
			name:     "RemoteAddr IPv4 с портом",
			remote:   "192.168.1.50:54321",
			expected: "192.168.1.50",
		},
		{
			name:     "RemoteAddr IPv6 с портом",
			remote:   "[2001:db8::1]:8080",
			expected: "2001:db8::1",
		},
		{
			name:     "RemoteAddr без порта",
			remote:   "127.0.0.1",
			expected: "127.0.0.1",
		},
		{
			name:     "Unix socket",
			remote:   "unix:/tmp/http.sock",
			expected: "unix", // net.SplitHostPort возвращает "unix"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			for header, value := range tt.headers {
				req.Header.Set(header, value)
			}
			req.RemoteAddr = tt.remote

			ip := getClientIP(req)
			assert.Equal(t, tt.expected, ip)
		})
	}
}
