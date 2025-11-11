package in_memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"internal-work-management-service/internal/domain"
)

func TestNoteRepoLifecycle(t *testing.T) {
	repo := NewNoteRepo()
	ctx := context.Background()

	n1, err := repo.Insert(ctx, domain.Note{Title: "first", Content: "body", CreatedAt: time.Now()})
	if err != nil {
		t.Fatalf("insert 1 failed: %v", err)
	}
	if n1.ID != 1 {
		t.Fatalf("expected id 1, got %d", n1.ID)
	}

	n2, err := repo.Insert(ctx, domain.Note{Title: "second", Content: "body", CreatedAt: time.Now()})
	if err != nil {
		t.Fatalf("insert 2 failed: %v", err)
	}
	if n2.ID != 2 {
		t.Fatalf("expected id 2, got %d", n2.ID)
	}

	all, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 notes, got %d", len(all))
	}
	if all[0].ID != 1 || all[1].ID != 2 {
		t.Fatalf("expected notes ordered by id, got %+v", all)
	}

	got, err := repo.Get(ctx, n1.ID)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if got.Title != n1.Title {
		t.Fatalf("expected title %q, got %q", n1.Title, got.Title)
	}

	if err := repo.Delete(ctx, n1.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	if _, err := repo.Get(ctx, n1.ID); err == nil || !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}

	if err := repo.Delete(ctx, 999); err == nil || !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound on missing delete, got %v", err)
	}
}
