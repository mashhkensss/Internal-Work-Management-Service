package notes

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"internal-work-management-service/internal/app/handlers"
)

func (h Handler) Delete(w http.ResponseWriter, r *http.Request) {
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

	if err := h.Service.Delete(ctx, id); err != nil {
		handlers.HandleError(w, err, mapError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
