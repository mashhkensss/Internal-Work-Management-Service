package note

import (
	"context"
	"errors"
	"fmt"
	"time"

	"internal-work-management-service/internal/domain"
	domainNote "internal-work-management-service/internal/domain/note"
)

// Repository describes storage operations required by the note service
type Repository interface {
	Insert(ctx context.Context, n domain.Note) (domain.Note, error)
	List(ctx context.Context) ([]domain.Note, error)
	Get(ctx context.Context, id int64) (domain.Note, error)
	Delete(ctx context.Context, id int64) error
}

type CreateNote struct {
	Title   string
	Content string
}

// Service provides note application logic.
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

func (s *Service) Create(ctx context.Context, in CreateNote) (domain.Note, error) {
	noteEntity, err := domainNote.New(in.Title, in.Content, s.now())
	if err != nil {
		return domain.Note{}, fmt.Errorf("note service: new entity: %w", err)
	}

	created, err := s.repo.Insert(ctx, noteEntity)
	if err != nil {
		return domain.Note{}, fmt.Errorf("note service: insert note: %w", err)
	}
	return created, nil
}

func (s *Service) List(ctx context.Context) ([]domain.Note, error) {
	notes, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("note service: list notes: %w", err)
	}
	return notes, nil
}

func (s *Service) Get(ctx context.Context, id int64) (domain.Note, error) {
	n, err := s.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.Note{}, domain.ErrNotFound
		}
		return domain.Note{}, fmt.Errorf("note service: get note %d: %w", id, err)
	}
	return n, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("note service: delete note %d: %w", id, err)
	}
	return nil
}
