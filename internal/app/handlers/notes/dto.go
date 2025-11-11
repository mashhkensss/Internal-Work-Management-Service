package notes

import (
	"time"

	"internal-work-management-service/internal/domain"
)

type createRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type noteResponse struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

func newNoteResponse(n domain.Note) noteResponse {
	return noteResponse{
		ID:        n.ID,
		Title:     n.Title,
		Content:   n.Content,
		CreatedAt: n.CreatedAt.UTC().Format(time.RFC3339),
	}
}
