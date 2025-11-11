package tasks

import (
	"context"

	"internal-work-management-service/internal/domain"
	taskService "internal-work-management-service/internal/service/task"
)

// Service exposes task operations consumed by HTTP handlers.
type Service interface {
	Create(ctx context.Context, in taskService.CreateTask) (domain.Task, error)
	List(ctx context.Context, filter taskService.Filter) ([]domain.Task, error)
	MarkDone(ctx context.Context, id int64) (domain.Task, error)
}

type Handler struct {
	Service Service
}

func New(service Service) Handler {
	return Handler{Service: service}
}
