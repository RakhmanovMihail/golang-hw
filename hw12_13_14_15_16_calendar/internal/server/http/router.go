package internalhttp

import (
	"github.com/go-chi/chi/v5"
)

// RegisterAPIRoutes registers API routes from generated spec.
func RegisterAPIRoutes(r chi.Router, apiServer *APIServer) {
	r.Mount("/api/v1/", HandlerFromMux(apiServer, r))
}
