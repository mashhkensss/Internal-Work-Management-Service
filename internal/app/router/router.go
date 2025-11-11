package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"internal-work-management-service/internal/app/handlers/notes"
	"internal-work-management-service/internal/app/handlers/orders"
	"internal-work-management-service/internal/app/handlers/tasks"
	"internal-work-management-service/internal/app/middleware"
)

// Deps aggregates router dependencies
type Deps struct {
	Notes  notes.Handler
	Tasks  tasks.Handler
	Orders orders.Handler

	Middlewares Middlewares
}

// Middlewares lists middleware
type Middlewares struct {
    RequestID func(http.Handler) http.Handler
    Logging   func(http.Handler) http.Handler
    Recovery  func(http.Handler) http.Handler
    Auth      func(http.Handler) http.Handler
}

// Router builds the HTTP handler tree
func Router(deps Deps) http.Handler {
	r := chi.NewRouter()

	if deps.Middlewares.Recovery != nil {
		r.Use(deps.Middlewares.Recovery)
	} else {
		r.Use(middleware.Recovery)
	}
	if deps.Middlewares.RequestID != nil {
		r.Use(deps.Middlewares.RequestID)
	} else {
		r.Use(middleware.RequestID)
	}
    if deps.Middlewares.Logging != nil {
        r.Use(deps.Middlewares.Logging)
    } else {
        r.Use(middleware.Logging)
    }

    if deps.Middlewares.Auth != nil {
        r.Use(deps.Middlewares.Auth)
    }

	r.Route("/notes", func(r chi.Router) {
		r.Post("/", deps.Notes.Create)
		r.Get("/", deps.Notes.List)
		r.Delete("/", deps.Notes.Delete)
		r.Get("/{id}", deps.Notes.GetByID)
		r.Delete("/{id}", deps.Notes.Delete)
	})

	r.Route("/tasks", func(r chi.Router) {
		r.Post("/", deps.Tasks.Create)
		r.Get("/", deps.Tasks.List)
		r.Post("/{id}/done", deps.Tasks.MarkDone)
	})

	r.Route("/orders", func(r chi.Router) {
		r.Post("/", deps.Orders.Create)
		r.Get("/", deps.Orders.List)
		r.Get("/{id}", deps.Orders.GetByID)
		r.Delete("/{id}", deps.Orders.Delete)
	})

	return r
}
