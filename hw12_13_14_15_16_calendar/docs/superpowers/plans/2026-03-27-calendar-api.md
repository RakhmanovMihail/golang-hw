# HTTP API для Календаря с OpenAPI спецификацией

> **Для агентов:** REQUIRED SUB-SKILL: Используйте `subagent-driven-development` для реализации этого плана по шагам.

**Goal:** Реализовать HTTP API для сервиса календаря на основе OpenAPI спецификации с использованием oapi-codegen для генерации кода.

**Architecture:** Спецификация OpenAPI V3 описывает все эндпоинты API. oapi-codegen генерирует типы и интерфейс сервера. Реализация хендлеров использует слой App (методы App, не напрямую storage).

**Tech Stack:** Go 1.25, oapi-codegen, chi router (для строгого следования спецификации), OpenAPI V3.

**Важные решения:**
- ID события генерируется автоматически в storage (не передаётся клиентом)
- oapi-codegen middleware автоматически валидирует required поля
- Хендлеры вызывают методы App, App вызывает методы Storage
- Пакет сервера: `internal/server/http` с именем пакета `internalhttp`
- API включает методы для получения событий по дате/неделе/месяцу (согласно ТЗ)

---

## Файлы для создания/изменения

**Создать:**
- `api/openapi.yaml` — спецификация OpenAPI V3 (с методами по дате/неделе/месяцу)
- `internal/server/http/handlers.go` — реализация хендлеров (тип `APIServer`)
- `internal/server/http/handlers_test.go` — юнит-тесты для API
- `internal/server/http/router.go` — регистрация API роутов

**Модифицировать:**
- `internal/app/app.go` — добавить методы: GetEvents, GetEvent, UpdateEvent, DeleteEvent, GetEventsByDay, GetEventsByWeek, GetEventsByMonth
- `internal/storage/storage.go` — добавить метод GetByID, GetEventsByDay, GetEventsByWeek, GetEventsByMonth
- `internal/storage/event.go` — добавить поля Description, NotifyBefore
- `internal/storage/memory/storage.go` — реализовать новые методы
- `internal/server/http/server.go` — переименовать Server → HTTPServer, интегрировать chi router
- `cmd/calendar/main.go` — обновление инициализации сервера
- `go.mod` — добавить зависимости (oapi-codegen, chi)
- `Makefile` — добавить target для генерации

---

## Task 1: Создание спецификации OpenAPI V3

**Files:**
- Create: `api/openapi.yaml`

- [ ] **Step 1: Создать спецификацию OpenAPI V3**

```yaml
openapi: 3.0.3
info:
  title: Calendar API
  description: HTTP API для сервиса календаря
  version: 1.0.0
servers:
  - url: http://localhost:8888
    description: Local development server

components:
  schemas:
    Event:
      type: object
      properties:
        id:
          type: integer
          format: int64
        title:
          type: string
        start_time:
          type: string
          format: date-time
        end_time:
          type: string
          format: date-time
        user_id:
          type: integer
          format: int32
        description:
          type: string
          nullable: true
        notify_before:
          type: integer
          format: int32
          nullable: true
          description: За сколько минут до события отправлять уведомление
    
    CreateEventRequest:
      type: object
      required:
        - title
        - start_time
        - end_time
        - user_id
      properties:
        title:
          type: string
        start_time:
          type: string
          format: date-time
        end_time:
          type: string
          format: date-time
        user_id:
          type: integer
          format: int32
        description:
          type: string
          nullable: true
        notify_before:
          type: integer
          format: int32
          nullable: true
    
    UpdateEventRequest:
      type: object
      properties:
        title:
          type: string
        start_time:
          type: string
          format: date-time
        end_time:
          type: string
          format: date-time
        user_id:
          type: integer
          format: int32
        description:
          type: string
          nullable: true
        notify_before:
          type: integer
          format: int32
          nullable: true
    
    ErrorResponse:
      type: object
      properties:
        error:
          type: string

paths:
  /api/v1/events:
    get:
      summary: Получить все события
      operationId: getEvents
      tags:
        - events
      responses:
        '200':
          description: Успешный ответ
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Event'
        '500':
          description: Внутренняя ошибка сервера
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
    
    post:
      summary: Создать новое событие
      operationId: createEvent
      tags:
        - events
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateEventRequest'
      responses:
        '201':
          description: Событие создано
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Event'
        '400':
          description: Некорректный запрос
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '409':
          description: Дата занята
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '500':
          description: Внутренняя ошибка сервера
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
  
  /api/v1/events/{id}:
    get:
      summary: Получить событие по ID
      operationId: getEvent
      tags:
        - events
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
            format: int64
      responses:
        '200':
          description: Успешный ответ
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Event'
        '404':
          description: Событие не найдено
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '500':
          description: Внутренняя ошибка сервера
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
    
    put:
      summary: Обновить событие
      operationId: updateEvent
      tags:
        - events
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
            format: int64
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UpdateEventRequest'
      responses:
        '200':
          description: Событие обновлено
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Event'
        '400':
          description: Некорректный запрос
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '404':
          description: Событие не найдено
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '409':
          description: Дата занята
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '500':
          description: Внутренняя ошибка сервера
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
    
    delete:
      summary: Удалить событие
      operationId: deleteEvent
      tags:
        - events
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
            format: int64
      responses:
        '200':
          description: Событие удалено
        '404':
          description: Событие не найдено
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '500':
          description: Внутренняя ошибка сервера
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'

  /api/v1/events/day/{date}:
    get:
      summary: Получить события на день
      operationId: getEventsByDay
      tags:
        - events
      parameters:
        - name: date
          in: path
          required: true
          schema:
            type: string
            format: date
          description: Дата в формате YYYY-MM-DD
      responses:
        '200':
          description: Успешный ответ
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Event'
        '400':
          description: Некорректный формат даты
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '500':
          description: Внутренняя ошибка сервера
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'

  /api/v1/events/week/{date}:
    get:
      summary: Получить события на неделю
      operationId: getEventsByWeek
      tags:
        - events
      parameters:
        - name: date
          in: path
          required: true
          schema:
            type: string
            format: date
          description: Дата начала недели в формате YYYY-MM-DD
      responses:
        '200':
          description: Успешный ответ
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Event'
        '400':
          description: Некорректный формат даты
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '500':
          description: Внутренняя ошибка сервера
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'

  /api/v1/events/month/{date}:
    get:
      summary: Получить события на месяц
      operationId: getEventsByMonth
      tags:
        - events
      parameters:
        - name: date
          in: path
          required: true
          schema:
            type: string
            format: date
          description: Дата начала месяца в формате YYYY-MM-DD
      responses:
        '200':
          description: Успешный ответ
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Event'
        '400':
          description: Некорректный формат даты
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '500':
          description: Внутренняя ошибка сервера
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
```

**Примечание:** 
- Поля в `CreateEventRequest` указаны как `required` — oapi-codegen middleware автоматически проверит их наличие
- ID события не передаётся клиентом, генерируется сервером
- Все поля дат в запросах — опциональные указатели (для частичного обновления в UpdateEvent)
- Добавлены эндпоинты для получения событий по дате/неделе/месяцу (согласно ТЗ)
- Событие включает дополнительные поля: `description`, `notify_before`

- [ ] **Step 2: Коммит**

```bash
git add api/openapi.yaml
git commit -m "feat(api): добавить OpenAPI V3 спецификацию для календаря"
```

---

## Task 2: Добавление методов в App и Storage

**Files:**
- Modify: `internal/app/app.go`
- Modify: `internal/storage/storage.go`

- [ ] **Step 1: Добавить метод GetByID в интерфейс Storage**

```go
// internal/storage/storage.go

// Storage is the interface for event storage.
type Storage interface {
	Create(ctx context.Context, e *Event) (*Event, error)
	Read(ctx context.Context) ([]Event, error)
	Update(ctx context.Context, id uint64, e *Event) (*Event, error)
	Delete(ctx context.Context, id uint64) error
	GetByID(ctx context.Context, id uint64) (*Event, error) // Новый метод
}
```

- [ ] **Step 2: Добавить методы в App**

```go
// internal/app/app.go

package app

import (
	"context"
	"time"

	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/storage"
)

// App represents the calendar application.
type App struct {
	Storage storage.Storage
	Logger  logger.Logger
}

// New creates a new App instance.
func New(logger logger.Logger, storage storage.Storage) *App {
	return &App{
		Logger:  logger,
		Storage: storage,
	}
}

// CreateEvent creates a new event.
func (a *App) CreateEvent(ctx context.Context, id uint64, title string, startTime, endTime time.Time, userID int) error {
	_, err := a.Storage.Create(ctx, &storage.Event{
		ID:        id,
		Title:     title,
		StartTime: startTime,
		EndTime:   endTime,
		UserID:    userID,
	})
	return err
}

// GetEvents returns all events.
func (a *App) GetEvents(ctx context.Context) ([]storage.Event, error) {
	return a.Storage.Read(ctx)
}

// GetEvent returns an event by ID.
func (a *App) GetEvent(ctx context.Context, id uint64) (*storage.Event, error) {
	return a.Storage.GetByID(ctx, id)
}

// UpdateEvent updates an existing event.
func (a *App) UpdateEvent(ctx context.Context, id uint64, title *string, startTime, endTime *time.Time, userID *int) (*storage.Event, error) {
	existing, err := a.Storage.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	newEvent := *existing
	if title != nil {
		newEvent.Title = *title
	}
	if startTime != nil {
		newEvent.StartTime = *startTime
	}
	if endTime != nil {
		newEvent.EndTime = *endTime
	}
	if userID != nil {
		newEvent.UserID = *userID
	}

	return a.Storage.Update(ctx, id, &newEvent)
}

// DeleteEvent deletes an event by ID.
func (a *App) DeleteEvent(ctx context.Context, id uint64) error {
	return a.Storage.Delete(ctx, id)
}
```

- [ ] **Step 3: Обновить memory storage (если нужно)**

Проверить, что `internal/storage/memory/storage.go` уже имеет метод `GetByID` — он уже реализован.

- [ ] **Step 4: Коммит**

```bash
git add internal/app/app.go internal/storage/storage.go
git commit -m "feat(app): добавить методы для работы с событиями"
```

---

## Task 3: Установка oapi-codegen и генерация кода

**Files:**
- Modify: `go.mod`
- Create: `internal/server/http/generated.go`

- [ ] **Step 1: Установить oapi-codegen и chi**

```bash
go get github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest
go get github.com/go-chi/chi/v5@v5.2.0
```

- [ ] **Step 2: Создать Makefile target для генерации**

Добавить в `Makefile`:
```makefile
generate:
	oapi-codegen -generate types,server,spec -package internalhttp -o internal/server/http/generated.go api/openapi.yaml

.PHONY: generate
```

- [ ] **Step 3: Запустить генерацию**

```bash
make generate
```

- [ ] **Step 4: Проверить сгенерированный файл**

Убедиться, что `internal/server/http/generated.go` содержит:
- Типы: `Event`, `CreateEventRequest`, `UpdateEventRequest`, `ErrorResponse`
- Интерфейс `ServerInterface` с методами: `GetEvents`, `CreateEvent`, `GetEvent`, `UpdateEvent`, `DeleteEvent`
- Типы параметров: `*time.Time` для полей даты

- [ ] **Step 5: Коммит**

```bash
git add go.mod go.sum internal/server/http/generated.go Makefile
git commit -m "chore: добавить oapi-codegen, chi и сгенерировать типы"
```

---

## Task 4: Реализация хендлеров API

**Files:**
- Create: `internal/server/http/handlers.go`
- Modify: `internal/server/http/server.go`

- [ ] **Step 1: Создать файл хендлеров**

```go
package internalhttp

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/app"
	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/storage"
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
// Note: oapi-codegen middleware validates required fields automatically.
func (s *APIServer) CreateEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var req CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	
	// ID генерируется в storage, не передаётся клиентом
	event := storage.Event{
		Title:     req.Title,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		UserID:    int(req.UserID),
	}
	
	// Вызываем метод Storage напрямую через App
	// (App.Storage — это поле, а не метод)
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
	
	// Получаем существующее событие
	existing, err := s.app.GetEvent(ctx, uint64(id))
	if err != nil {
		if errors.Is(err, storage.ErrEventNotFound) {
			s.writeError(w, http.StatusNotFound, "event not found")
			return
		}
		s.writeError(w, http.StatusInternalServerError, "failed to get event")
		return
	}
	
	// Обновляем поля если они предоставлены
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
	if req.UserID != nil {
		newEvent.UserID = int(*req.UserID)
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
	return Event{
		Id:        int64(e.ID),
		Title:     e.Title,
		StartTime: e.StartTime,
		EndTime:   e.EndTime,
		UserId:    int32(e.UserID),
	}
}
```

- [ ] **Step 2: Обновить server.go — переименовать и интегрировать chi**

```go
package internalhttp

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Application is the interface for the application logic.
type Application interface{}

// HTTPServer represents an HTTP server.
type HTTPServer struct {
	server *http.Server
	logger logger.ILogger
	addr   string
}

// NewServer creates a new HTTP server instance.
func NewServer(logger logger.ILogger, app Application, addr string) *HTTPServer {
	router := chi.NewRouter()
	
	router.Use(middleware.Logger)
	router.Use(loggingMiddleware(logger))
	
	// Hello endpoint
	router.Get("/", helloHandler(logger))
	
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
	}
}

// Start starts the HTTP server.
func (s *HTTPServer) Start(ctx context.Context) error {
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
func (s *HTTPServer) Stop(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
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
```

- [ ] **Step 3: Создать файл для регистрации API роутов**

```go
// internal/server/http/router.go

package internalhttp

import (
	"github.com/go-chi/chi/v5"
	oapiMiddleware "github.com/deepmap/oapi-codegen/pkg/chi-middleware"
)

// RegisterAPIRoutes registers API routes from generated spec.
func RegisterAPIRoutes(r chi.Router, apiServer *APIServer) {
	r.Group(func(r chi.Router) {
		r.Use(oapiMiddleware.OapiRequestValidatorWithOptions(nil))
		r.Mount("/api/v1/", HandlerFromMux(apiServer, r))
	})
}
```

- [ ] **Step 4: Коммит**

```bash
git add internal/server/http/handlers.go internal/server/http/server.go internal/server/http/router.go
git commit -m "feat(server): реализовать хендлеры API и интегрировать chi router"
```

---

## Task 5: Интеграция в main.go

**Files:**
- Modify: `cmd/calendar/main.go`

- [ ] **Step 1: Обновить main.go**

```go
// В функцию run() после создания calendar:
calendar := app.New(*logg, store)
apiServer := internalhttp.NewAPIServer(calendar)

// Создаём HTTP сервер с chi router
server := internalhttp.NewServer(logg, calendar, fmt.Sprintf("%s:%s", cfg.API.Host, cfg.API.Port))

// Получаем chi.Mux из сервера и регистрируем API роуты
// NewServer уже создаёт chi.NewRouter() внутри
// Нужно добавить метод для получения router или зарегистрировать роуты внутри NewServer
```

**Альтернативный подход (более чистый):**

Обновить `NewServer` чтобы он принимал `chi.Router` и API сервер:

```go
// internal/server/http/server.go
func NewServer(logger logger.ILogger, app Application, addr string, apiRouter chi.Router) *HTTPServer {
	router := chi.NewRouter()
	
	router.Use(middleware.Logger)
	router.Use(loggingMiddleware(logger))
	
	// Hello endpoint
	router.Get("/", helloHandler(logger))
	
	// API routes
	if apiRouter != nil {
		router.Mount("/api/v1", apiRouter)
	}
	
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
	}
}
```

Тогда в `main.go`:

```go
calendar := app.New(*logg, store)
apiServer := internalhttp.NewAPIServer(calendar)

// Создаём API router через oapi-codegen
apiRouter := chi.NewRouter()
apiRouter.Mount("/", internalhttp.HandlerFromMux(apiServer, apiRouter))

server := internalhttp.NewServer(logg, calendar, fmt.Sprintf("%s:%s", cfg.API.Host, cfg.API.Port), apiRouter)
```

- [ ] **Step 2: Коммит**

```bash
git add cmd/calendar/main.go
git commit -m "feat(main): интегрировать API сервер"
```

---

## Task 6: Написание юнит-тестов для API

**Files:**
- Create: `internal/server/http/handlers_test.go`

- [ ] **Step 1: Создать тесты**

```go
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

func setupTestServer(t *testing.T) (*chi.Mux, *app.App) {
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
	router, _ := setupTestServer(t)
	
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "[]\n", w.Body.String())
}

func TestCreateEvent_Success(t *testing.T) {
	router, _ := setupTestServer(t)
	
	now := time.Now().UTC()
	event := internalhttp.CreateEventRequest{
		Title:     ptr("Test Event"),
		StartTime: ptr(now.Add(time.Hour)),
		EndTime:   ptr(now.Add(2 * time.Hour)),
		UserId:    ptr(int32(1)),
	}
	
	body, _ := json.Marshal(event)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/events", bytes.NewReader(body))
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var created internalhttp.Event
	err := json.Unmarshal(w.Body.Bytes(), &created)
	require.NoError(t, err)
	assert.Equal(t, "Test Event", created.Title)
}

func TestGetEvent_NotFound(t *testing.T) {
	router, _ := setupTestServer(t)
	
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events/999", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateEvent_Success(t *testing.T) {
	router, app := setupTestServer(t)
	
	// Сначала создадим событие
	ctx := context.Background()
	created, err := app.Storage.Create(ctx, &storage.Event{
		Title:     "Original",
		StartTime: time.Now().Add(time.Hour),
		EndTime:   time.Now().Add(2 * time.Hour),
		UserID:    1,
	})
	require.NoError(t, err)
	
	// Теперь обновим
	update := internalhttp.UpdateEventRequest{
		Title: ptr("Updated"),
	}
	
	body, _ := json.Marshal(update)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/events/"+strconv.Itoa(int(created.ID)), bytes.NewReader(body))
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteEvent_Success(t *testing.T) {
	router, app := setupTestServer(t)
	
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
	router, app := setupTestServer(t)
	
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
		Title:     ptr("Conflict Event"),
		StartTime: ptr(now.Add(time.Hour)),
		EndTime:   ptr(now.Add(2 * time.Hour)),
		UserId:    ptr(int32(2)),
	}
	
	body, _ := json.Marshal(event)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/events", bytes.NewReader(body))
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusConflict, w.Code)
}

func ptr[T any](v T) *T {
	return &v
}
```

- [ ] **Step 2: Запустить тесты**

```bash
go test -v ./internal/server/http/...
```

- [ ] **Step 3: Коммит**

```bash
git add internal/server/http/handlers_test.go
git commit -m "test(api): добавить юнит-тесты для API"
```

---

## Task 7: Финальная проверка и документация

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Обновить README.md**

Добавить секцию про API:

```markdown
## API

API описано в спецификации OpenAPI V3: [`api/openapi.yaml`](api/openapi.yaml)

### Запуск сервера

```bash
make run
# или
go run cmd/calendar/main.go -config configs/config.toml
```

### Генерация кода

```bash
make generate
```

### Endpoints

- `GET /api/v1/events` - Получить все события
- `POST /api/v1/events` - Создать событие
- `GET /api/v1/events/{id}` - Получить событие по ID
- `PUT /api/v1/events/{id}` - Обновить событие
- `DELETE /api/v1/events/{id}` - Удалить событие
```

- [ ] **Step 2: Запустить сборку**

```bash
go build -o bin/calendar ./cmd/calendar
```

- [ ] **Step 3: Запустить тесты**

```bash
go test -v ./...
```

- [ ] **Step 4: Отформатировать код**

```bash
go fmt ./...
go vet ./...
```

- [ ] **Step 5: Коммит**

```bash
git add README.md
git commit -m "docs: обновить README с информацией об API"
```

---

## Критерии готовности

- [ ] Спецификация OpenAPI V3 создана и валидна
- [ ] oapi-codegen установлен и генерирует код
- [ ] Все хендлеры реализованы
- [ ] Юнит-тесты написаны и проходят
- [ ] Сборка работает без ошибок
- [ ] Код отформатирован (`go fmt ./...`)
- [ ] Линтер проходит (`go vet ./...`)
