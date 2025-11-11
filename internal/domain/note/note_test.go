package note_test

import (
	"testing"
	"time"

	note "internal-work-management-service/internal/domain/note"
	"internal-work-management-service/tests/testutils"
)

func TestNoteNewSuccess(t *testing.T) {
	createdAt := time.Now()
	n, err := note.New("  Title  ", " Body ", createdAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.Title != "Title" {
		t.Fatalf("title not trimmed: %q", n.Title)
	}
	if n.Content != "Body" {
		t.Fatalf("content not trimmed: %q", n.Content)
	}
	if !n.CreatedAt.Equal(createdAt) {
		t.Fatalf("createdAt mismatch")
	}
}

func TestNoteNewValidationErrors(t *testing.T) {
	createdAt := time.Now()
	_, err := note.New("", "body", createdAt)
	if err != note.ErrEmptyTitle {
		t.Fatalf("expected ErrEmptyTitle, got %v", err)
	}
	_, err = note.New("title", "", createdAt)
	if err != note.ErrEmptyContent {
		t.Fatalf("expected ErrEmptyContent, got %v", err)
	}
	_, err = note.New("title", "body", time.Time{})
	if err != note.ErrCreatedAtZero {
		t.Fatalf("expected ErrCreatedAtZero, got %v", err)
	}
}

func TestNoteWithID(t *testing.T) {
	n := note.Note{Title: "a", Content: "b", CreatedAt: time.Now()}
	withID, err := testutils.NoteWithID(n, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if withID.ID != 10 {
		t.Fatalf("expected id 10, got %d", withID.ID)
	}
	_, err = testutils.NoteWithID(n, 0)
	if err != note.ErrInvalidID {
		t.Fatalf("expected ErrInvalidID, got %v", err)
	}
}

func TestNoteUpdate(t *testing.T) {
	n := note.Note{Title: "old", Content: "old", CreatedAt: time.Now()}
	if err := n.Update(" new ", "  body"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.Title != "new" || n.Content != "body" {
		t.Fatalf("update should trim inputs")
	}
	n2 := note.Note{Title: "old", Content: "old", CreatedAt: time.Now()}
	if err := n2.Update("", "body"); err != note.ErrEmptyTitle {
		t.Fatalf("expected ErrEmptyTitle, got %v", err)
	}
}
