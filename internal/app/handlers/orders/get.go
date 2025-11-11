package orders

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"internal-work-management-service/internal/app/handlers"
)

func (h Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		idStr = r.URL.Query().Get("id")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("err invalid id"))
		return
	}

	order, err := h.Service.Get(ctx, id)
	if err != nil {
		handlers.HandleError(w, err, mapError)
		return
	}
	handlers.WriteJSON(w, http.StatusOK, newOrderResponse(order))
}
