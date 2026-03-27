package internalhttp

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/logger"
)

func loggingMiddleware(logger logger.ILogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Создаем response writer для перехвата статуса и размера
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				size:           0,
			}

			next.ServeHTTP(rw, r)

			latency := time.Since(start)
			logger.Info(fmt.Sprintf(
				"%s [%s] %s %s%s %s %d %dms \"%s\"",
				getClientIP(r),
				start.Format("02/Jan/2006:15:04:05 -0700"),
				r.Method,
				r.URL.Path,
				r.URL.RawQuery,
				r.Proto,
				rw.statusCode,
				latency.Milliseconds(),
				r.UserAgent(),
			))
		})
	}
}

func getClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return ip
	}
	return r.RemoteAddr
}
