package orders

import (
	"context"

	"internal-work-management-service/internal/domain"
	"internal-work-management-service/internal/domain/discount"
	orderService "internal-work-management-service/internal/service/order"
)

// Service exposes order operations used by HTTP handlers.
type Service interface {
	Create(ctx context.Context, in orderService.CreateOrder, strat discount.Strategy, idem *orderService.IdempotencyKey) (domain.Order, error)
	List(ctx context.Context) ([]domain.Order, error)
	Get(ctx context.Context, id int64) (domain.Order, error)
	Delete(ctx context.Context, id int64) error
}

type Handler struct {
	Service Service
}

func New(service Service) Handler {
	return Handler{Service: service}
}
