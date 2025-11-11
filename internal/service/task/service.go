package task

import (
	"context"
	"errors"
	"fmt"
	"time"

	"internal-work-management-service/internal/domain"
	domainTask "internal-work-management-service/internal/domain/task"
)

// Repository describes storage operations used by the task service
type Repository interface {
	Insert(ctx context.Context, t domain.Task) (domain.Task, error)
	List(ctx context.Context, filter Filter) ([]domain.Task, error)
	MarkDone(ctx context.Context, id int64) (domain.Task, error)
}

type CreateTask struct {
	Title       string
	Description string
}

type Filter struct {
	Done *bool
}

// Service provides task application logic.
type Service struct {
	repo Repository
	now  func() time.Time
}

func New(repo Repository, now func() time.Time) *Service {
	if now == nil {
		now = defaultNow
	}

	return &Service{
		repo: repo,
		now:  now,
	}
}

func defaultNow() time.Time {
	return time.Now().UTC()
}

func (s *Service) Create(ctx context.Context, in CreateTask) (domain.Task, error) {
	taskEntity, err := domainTask.New(in.Title, in.Description, s.now())
	if err != nil {
		return domain.Task{}, fmt.Errorf("task service: new entity: %w", err)
	}

	created, err := s.repo.Insert(ctx, taskEntity)
	if err != nil {
		return domain.Task{}, fmt.Errorf("task service: insert task: %w", err)
	}
	return created, nil
}

func (s *Service) List(ctx context.Context, filter Filter) ([]domain.Task, error) {
	tasks, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("task service: list tasks: %w", err)
	}
	return tasks, nil
}

func (s *Service) MarkDone(ctx context.Context, id int64) (domain.Task, error) {
	taskResult, err := s.repo.MarkDone(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			return domain.Task{}, domain.ErrNotFound
		case errors.Is(err, domainTask.ErrAlreadyDone) && taskResult.Done:
			return taskResult, nil
		default:
			return domain.Task{}, fmt.Errorf("task service: mark done %d: %w", id, err)
		}
	}
	return taskResult, nil
}
