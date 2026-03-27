package internalhttp

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/app"
	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/storage"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// APIServer implements the generated ServerInterface.
type APIServer struct {
	app *app.App
}

// NewAPIServer creates a new APIServer instance.
func NewAPIServer(app *app.App) *APIServer {
	return &APIServer{app: app}
}

// GetEvents implements ServerInterface.
func (s *APIServer) GetEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	events, err := s.app.GetEvents(ctx)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "failed to read events")
		return
	}

	response := make([]Event, 0, len(events))
	for _, e := range events {
		response = append(response, eventToAPI(e))
	}

	s.writeJSON(w, http.StatusOK, response)
}

// CreateEvent implements ServerInterface.
func (s *APIServer) CreateEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	event := storage.Event{
		Title:     req.Title,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		UserID:    int(req.UserId),
	}

	created, err := s.app.Storage.Create(ctx, &event)
	if err != nil {
		if errors.Is(err, storage.ErrDateBusy) {
			s.writeError(w, http.StatusConflict, "date is busy")
			return
		}
		s.writeError(w, http.StatusInternalServerError, "failed to create event")
		return
	}

	s.writeJSON(w, http.StatusCreated, eventToAPI(*created))
}

// GetEvent implements ServerInterface.
func (s *APIServer) GetEvent(w http.ResponseWriter, r *http.Request, id int64) {
	ctx := r.Context()

	event, err := s.app.GetEvent(ctx, uint64(id))
	if err != nil {
		if errors.Is(err, storage.ErrEventNotFound) {
			s.writeError(w, http.StatusNotFound, "event not found")
			return
		}
		s.writeError(w, http.StatusInternalServerError, "failed to get event")
		return
	}

	s.writeJSON(w, http.StatusOK, eventToAPI(*event))
}

// UpdateEvent implements ServerInterface.
func (s *APIServer) UpdateEvent(w http.ResponseWriter, r *http.Request, id int64) {
	ctx := r.Context()

	var req UpdateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	existing, err := s.app.GetEvent(ctx, uint64(id))
	if err != nil {
		if errors.Is(err, storage.ErrEventNotFound) {
			s.writeError(w, http.StatusNotFound, "event not found")
			return
		}
		s.writeError(w, http.StatusInternalServerError, "failed to get event")
		return
	}

	newEvent := *existing
	if req.Title != nil {
		newEvent.Title = *req.Title
	}
	if req.StartTime != nil {
		newEvent.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		newEvent.EndTime = *req.EndTime
	}
	if req.UserId != nil {
		newEvent.UserID = int(*req.UserId)
	}

	updated, err := s.app.Storage.Update(ctx, uint64(id), &newEvent)
	if err != nil {
		if errors.Is(err, storage.ErrEventNotFound) {
			s.writeError(w, http.StatusNotFound, "event not found")
			return
		}
		if errors.Is(err, storage.ErrDateBusy) {
			s.writeError(w, http.StatusConflict, "date is busy")
			return
		}
		s.writeError(w, http.StatusInternalServerError, "failed to update event")
		return
	}

	s.writeJSON(w, http.StatusOK, eventToAPI(*updated))
}

// DeleteEvent implements ServerInterface.
func (s *APIServer) DeleteEvent(w http.ResponseWriter, r *http.Request, id int64) {
	ctx := r.Context()

	err := s.app.DeleteEvent(ctx, uint64(id))
	if err != nil {
		if errors.Is(err, storage.ErrEventNotFound) {
			s.writeError(w, http.StatusNotFound, "event not found")
			return
		}
		s.writeError(w, http.StatusInternalServerError, "failed to delete event")
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetEventsByDay implements ServerInterface.
func (s *APIServer) GetEventsByDay(w http.ResponseWriter, r *http.Request, date openapi_types.Date) {
	ctx := r.Context()

	events, err := s.app.GetEvents(ctx)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "failed to read events")
		return
	}

	filtered := make([]Event, 0)
	for _, e := range events {
		eventDate := e.StartTime.Format("2006-01-02")
		if eventDate == date.String() {
			filtered = append(filtered, eventToAPI(e))
		}
	}

	s.writeJSON(w, http.StatusOK, filtered)
}

// GetEventsByWeek implements ServerInterface.
func (s *APIServer) GetEventsByWeek(w http.ResponseWriter, r *http.Request, date openapi_types.Date) {
	ctx := r.Context()

	events, err := s.app.GetEvents(ctx)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "failed to read events")
		return
	}

	targetTime := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	startOfWeek := targetTime.AddDate(0, 0, -int(targetTime.Weekday()))
	endOfWeek := startOfWeek.AddDate(0, 0, 7)

	filtered := make([]Event, 0)
	for _, e := range events {
		if (e.StartTime.After(startOfWeek) || e.StartTime.Equal(startOfWeek)) && e.StartTime.Before(endOfWeek) {
			filtered = append(filtered, eventToAPI(e))
		}
	}

	s.writeJSON(w, http.StatusOK, filtered)
}

// GetEventsByMonth implements ServerInterface.
func (s *APIServer) GetEventsByMonth(w http.ResponseWriter, r *http.Request, date openapi_types.Date) {
	ctx := r.Context()

	events, err := s.app.GetEvents(ctx)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "failed to read events")
		return
	}

	filtered := make([]Event, 0)
	for _, e := range events {
		if e.StartTime.Year() == date.Year() && e.StartTime.Month() == date.Month() {
			filtered = append(filtered, eventToAPI(e))
		}
	}

	s.writeJSON(w, http.StatusOK, filtered)
}

// Helper functions.

func (s *APIServer) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.app.Logger.Error("write response: " + err.Error())
	}
}

func (s *APIServer) writeError(w http.ResponseWriter, status int, message string) {
	s.writeJSON(w, status, ErrorResponse{Error: &message})
}

func eventToAPI(e storage.Event) Event {
	id := int64(e.ID)
	userID := int64(e.UserID)
	return Event{
		Id:        &id,
		Title:     &e.Title,
		StartTime: &e.StartTime,
		EndTime:   &e.EndTime,
		UserId:    &userID,
	}
}
