package notes

import (
	"net/http"

	"internal-work-management-service/internal/app/handlers"
	"internal-work-management-service/internal/service/note"
)

func (h Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req createRequest
	if err := handlers.DecodeJSON(w, r, &req); err != nil {
		handlers.WriteError(w, http.StatusBadRequest, err)
		return
	}

	n, err := h.Service.Create(ctx, note.CreateNote{
		Title:   req.Title,
		Content: req.Content,
	})
	if err != nil {
		handlers.HandleError(w, err, mapError)
		return
	}

	handlers.WriteJSON(w, http.StatusCreated, newNoteResponse(n))
}
