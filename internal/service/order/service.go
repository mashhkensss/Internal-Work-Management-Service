package order

import (
	"context"
	"errors"
	"fmt"
	"time"

	"internal-work-management-service/internal/domain"
	"internal-work-management-service/internal/domain/discount"
	domainOrder "internal-work-management-service/internal/domain/order"
)

// Repository describes storage operations required by the order service
type Repository interface {
	Insert(ctx context.Context, o domain.Order, idem *IdempotencyKey) (domain.Order, error)
	List(ctx context.Context) ([]domain.Order, error)
	Get(ctx context.Context, id int64) (domain.Order, error)
	Delete(ctx context.Context, id int64) error
}

type CreateOrder struct {
	Customer string
	Items    []domainOrder.Item
}

// IdempotencyKey references an idempotency token
type IdempotencyKey struct {
	Key string
}

// Service provides order application logic.
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

func (s *Service) Create(ctx context.Context, in CreateOrder, strat discount.Strategy, idem *IdempotencyKey) (domain.Order, error) {
	orderEntity, err := domainOrder.New(in.Customer, in.Items, s.now())
	if err != nil {
		return domain.Order{}, fmt.Errorf("order service: new entity: %w", err)
	}

	if err := orderEntity.ApplyDiscount(strat); err != nil {
		return domain.Order{}, fmt.Errorf("order service: apply discount: %w", err)
	}

	created, err := s.repo.Insert(ctx, orderEntity, idem)
	if err != nil {
		return domain.Order{}, fmt.Errorf("order service: insert order: %w", err)
	}
	return created, nil
}

func (s *Service) List(ctx context.Context) ([]domain.Order, error) {
	orders, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("order service: list orders: %w", err)
	}
	return orders, nil
}

func (s *Service) Get(ctx context.Context, id int64) (domain.Order, error) {
	res, err := s.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.Order{}, domain.ErrNotFound
		}
		return domain.Order{}, fmt.Errorf("order service: get order %d: %w", id, err)
	}
	return res, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("order service: delete order %d: %w", id, err)
	}
	return nil
}
