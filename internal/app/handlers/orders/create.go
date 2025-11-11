package orders

import (
	"net/http"
	"strings"

	"internal-work-management-service/internal/app/handlers"
	"internal-work-management-service/internal/domain/discount"
	orderService "internal-work-management-service/internal/service/order"
)

func (h Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	key := r.Header.Get("Idempotency-Key")

	var req createRequest
	if err := handlers.DecodeJSON(w, r, &req); err != nil {
		handlers.WriteError(w, http.StatusBadRequest, err)
		return
	}

	items := mapItems(req.Items)

	var idemKey *orderService.IdempotencyKey
	if trimmed := strings.TrimSpace(key); trimmed != "" {
		idemKey = &orderService.IdempotencyKey{Key: trimmed}
	}

	created, err := h.Service.Create(ctx, orderService.CreateOrder{
		Customer: req.Customer,
		Items:    items,
	}, discount.None{}, idemKey)
	if err != nil {
		handlers.HandleError(w, err, mapError)
		return
	}

	handlers.WriteJSON(w, http.StatusCreated, newOrderResponse(created))
}
