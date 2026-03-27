package internalhttp

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/app"
	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const shutdownTimeout = 5 * time.Second

// HTTPServer represents an HTTP server.
type HTTPServer struct {
	server *http.Server
	logger logger.ILogger
	addr   string
	router chi.Router
}

// Router returns the chi.Router instance.
func (s *HTTPServer) Router() chi.Router {
	return s.router
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

// NewHTTPServer creates a new HTTP server instance.
func NewHTTPServer(logger logger.ILogger, app *app.App, addr string) *HTTPServer {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(loggingMiddleware(logger))

	// Hello endpoint
	router.Get("/", helloHandler(logger))

	// API routes
	apiServer := NewAPIServer(app)
	RegisterAPIRoutes(router, apiServer)

	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return &HTTPServer{
		server: srv,
		logger: logger,
		addr:   addr,
		router: router,
	}
}

// Start starts the HTTP server.
func (s *HTTPServer) Start(ctx context.Context) error {
	s.logger.Info(fmt.Sprintf("starting HTTP server at %s", s.addr))

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if err != http.ErrServerClosed {
			s.logger.Error("server failed: " + err.Error())
			return err
		}
	case <-ctx.Done():
	}

	return nil
}

// Stop stops the HTTP server gracefully.
func (s *HTTPServer) Stop(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("server shutdown error: " + err.Error())
		return err
	}

	s.logger.Info("server stopped")
	return nil
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
