package notes

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"internal-work-management-service/internal/app/handlers"
)

func (h Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		idStr = chi.URLParam(r, "id")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("err invalid id"))
		return
	}

	n, err := h.Service.Get(ctx, id)
	if err != nil {
		handlers.HandleError(w, err, mapError)
		return
	}
	handlers.WriteJSON(w, http.StatusOK, newNoteResponse(n))
}
