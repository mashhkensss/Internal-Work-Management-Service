package notes

import (
	"context"

	"internal-work-management-service/internal/domain"
	noteService "internal-work-management-service/internal/service/note"
)

// Service exposes note operations required by the HTTP layer
type Service interface {
	Create(ctx context.Context, in noteService.CreateNote) (domain.Note, error)
	List(ctx context.Context) ([]domain.Note, error)
	Get(ctx context.Context, id int64) (domain.Note, error)
	Delete(ctx context.Context, id int64) error
}

// Handler wires note-related HTTP endpoints
type Handler struct {
	Service Service
}

func New(service Service) Handler {
	return Handler{Service: service}
}
