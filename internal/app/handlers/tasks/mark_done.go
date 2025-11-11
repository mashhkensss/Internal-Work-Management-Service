package tasks

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"internal-work-management-service/internal/app/handlers"
)

func (h Handler) MarkDone(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("err invalid id"))
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("err invalid id"))
		return
	}

	task, err := h.Service.MarkDone(ctx, id)
	if err != nil {
		handlers.HandleError(w, err, mapError)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, newTaskResponse(task))
}
