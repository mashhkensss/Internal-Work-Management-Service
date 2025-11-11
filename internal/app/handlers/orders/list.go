package orders

import (
	"net/http"

	"internal-work-management-service/internal/app/handlers"
)

func (h Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orders, err := h.Service.List(ctx)
	if err != nil {
		handlers.WriteInternalError(w, err)
		return
	}

	res := make([]response, 0, len(orders))
	for _, o := range orders {
		res = append(res, newOrderResponse(o))
	}
	handlers.WriteJSON(w, http.StatusOK, res)
}
