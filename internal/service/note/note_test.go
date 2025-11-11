package note

import (
	"context"
	"errors"
	"testing"
	"time"

	"internal-work-management-service/internal/domain"
	domainNote "internal-work-management-service/internal/domain/note"
)

type noteRepoMock struct {
	inserted  domain.Note
	insertErr error

	listResult []domain.Note
	listErr    error

	getResult domain.Note
	getErr    error

	deleteErr error
}

func (m *noteRepoMock) Insert(_ context.Context, n domain.Note) (domain.Note, error) {
	if m.insertErr != nil {
		return domain.Note{}, m.insertErr
	}
	m.inserted = n
	if n.ID == 0 {
		n.ID = 1
	}
	return n, nil
}

func (m *noteRepoMock) List(context.Context) ([]domain.Note, error) {
	return m.listResult, m.listErr
}

func (m *noteRepoMock) Get(context.Context, int64) (domain.Note, error) {
	return m.getResult, m.getErr
}

func (m *noteRepoMock) Delete(context.Context, int64) error {
	return m.deleteErr
}

func TestNoteServiceCreate(t *testing.T) {
	mockRepo := &noteRepoMock{}
	svc := New(mockRepo, nil)

	n, err := svc.Create(context.Background(), CreateNote{Title: "  Title ", Content: " body "})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mockRepo.inserted.Title != "Title" || mockRepo.inserted.Content != "body" {
		t.Fatalf("note should be trimmed before insert: %+v", mockRepo.inserted)
	}
	if n.ID == 0 {
		t.Fatalf("expected id to be set")
	}
}

func TestNoteServiceCreateValidation(t *testing.T) {
	svc := New(&noteRepoMock{}, nil)
	_, err := svc.Create(context.Background(), CreateNote{Title: "", Content: "body"})
	if !errors.Is(err, domainNote.ErrEmptyTitle) {
		t.Fatalf("expected ErrEmptyTitle, got %v", err)
	}
}

func TestNoteServiceCreateRepo(t *testing.T) {
	wantErr := errors.New("insert failure")
	svc := New(&noteRepoMock{insertErr: wantErr}, nil)
	_, err := svc.Create(context.Background(), CreateNote{Title: "title", Content: "body"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}

func TestNoteServiceList(t *testing.T) {
	expected := []domain.Note{{ID: 1}, {ID: 2}}
	svc := New(&noteRepoMock{listResult: expected}, nil)
	got, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(expected) {
		t.Fatalf("expected %d notes, got %d", len(expected), len(got))
	}
}

func TestNoteServiceGet(t *testing.T) {
	svc := New(&noteRepoMock{getErr: domain.ErrNotFound}, nil)
	_, err := svc.Get(context.Background(), 1)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected domain.ErrNotFound, got %v", err)
	}
}

func TestNoteServiceDelete(t *testing.T) {
	svc := New(&noteRepoMock{deleteErr: domain.ErrNotFound}, nil)
	if err := svc.Delete(context.Background(), 1); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected domain.ErrNotFound, got %v", err)
	}
}

func TestNoteServiceDeleteSuccess(t *testing.T) {
	svc := New(&noteRepoMock{}, nil)
	if err := svc.Delete(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNoteServiceInjectsClock(t *testing.T) {
	fixed := time.Date(2025, 2, 3, 4, 5, 6, 0, time.UTC)
	repo := &noteRepoMock{}
	svc := New(repo, func() time.Time { return fixed })

	created, err := svc.Create(context.Background(), CreateNote{Title: "title", Content: "body"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !created.CreatedAt.Equal(fixed) {
		t.Fatalf("expected CreatedAt to equal injected clock, got %v", created.CreatedAt)
	}
}
