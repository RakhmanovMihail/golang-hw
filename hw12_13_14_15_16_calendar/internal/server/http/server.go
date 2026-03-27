package internalhttp

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/logger"
)

// Application is the interface for the application logic.
type Application interface{}

// Server represents an HTTP server.
type Server struct {
	server *http.Server
	logger logger.ILogger
	addr   string
}

type responseWriter struct { // ← В ЭТОМ ЖЕ ФАЙЛЕ
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.size += len(b)
	return rw.ResponseWriter.Write(b)
}

// NewServer creates a new HTTP server instance.
func NewServer(logger logger.ILogger, app Application, addr string) *Server {
	router := http.NewServeMux()
	router.HandleFunc("/", helloHandler(logger))

	srv := &http.Server{
		Addr:         addr,
		Handler:      loggingMiddleware(logger)(router),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return &Server{
		server: srv,
		logger: logger,
		addr:   addr,
	}
}

// Start starts the HTTP server.
func (s *Server) Start(ctx context.Context) error {
	s.logger.Info(fmt.Sprintf("starting HTTP server at %s", s.addr))

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("server failed: " + err.Error())
		}
	}()

	<-ctx.Done()
	return nil
}

// Stop stops the HTTP server gracefully.
func (s *Server) Stop(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("server shutdown error: " + err.Error())
		return err
	}

	s.logger.Info("server stopped")
	return nil
}

func queryString(r *http.Request) string {
	if r.URL.RawQuery != "" {
		return "?" + r.URL.RawQuery
	}
	return ""
}

func helloHandler(logger logger.ILogger) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		logger.Debug("hello endpoint called")
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("hello-world")); err != nil {
			logger.Error("write response: " + err.Error())
		}
	}
}
