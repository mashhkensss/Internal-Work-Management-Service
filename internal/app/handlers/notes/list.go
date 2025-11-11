package notes

import (
	"errors"
	"net/http"
	"strconv"

	"internal-work-management-service/internal/app/handlers"
)

func (h Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if idStr := r.URL.Query().Get("id"); idStr != "" {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid id"))
			return
		}

		n, err := h.Service.Get(ctx, id)
		if err != nil {
			handlers.HandleError(w, err, mapError)
			return
		}
		handlers.WriteJSON(w, http.StatusOK, newNoteResponse(n))
		return
	}

	notes, err := h.Service.List(ctx)
	if err != nil {
		handlers.WriteInternalError(w, err)
		return
	}

	res := make([]noteResponse, 0, len(notes))
	for _, n := range notes {
		res = append(res, newNoteResponse(n))
	}
	handlers.WriteJSON(w, http.StatusOK, res)
}
