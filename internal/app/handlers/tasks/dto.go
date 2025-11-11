package tasks

import (
	"time"

	"internal-work-management-service/internal/domain"
)

type createRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type response struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description,omitempty"`
	Done        bool    `json:"done"`
	CreatedAt   string  `json:"created_at"`
	CompletedAt *string `json:"completed_at,omitempty"`
}

func newTaskResponse(t domain.Task) response {
	var completedAt *string
	if t.CompletedAt != nil {
		formatted := t.CompletedAt.UTC().Format(time.RFC3339)
		completedAt = &formatted
	}

	return response{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Done:        t.Done,
		CreatedAt:   t.CreatedAt.UTC().Format(time.RFC3339),
		CompletedAt: completedAt,
	}
}
