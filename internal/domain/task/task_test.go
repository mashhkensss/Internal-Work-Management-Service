package task

import (
	"testing"
	"time"
)

func TestTaskNewAndValidate(t *testing.T) {
	createdAt := time.Now()
	tk, err := New(" title ", " desc ", createdAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tk.Title != "title" || tk.Description != "desc" {
		t.Fatalf("strings should be trimmed")
	}
	if tk.Done {
		t.Fatalf("new task must be not done")
	}
}

func TestTaskNewValidationErrors(t *testing.T) {
	_, err := New("", "desc", time.Now())
	if err != ErrEmptyTitle {
		t.Fatalf("expected ErrEmptyTitle, got %v", err)
	}
	_, err = New("title", "desc", time.Time{})
	if err != ErrCreatedAtZero {
		t.Fatalf("expected ErrCreatedAtZero, got %v", err)
	}
}

func TestTaskMarkDoneAndUndone(t *testing.T) {
	now := time.Now()
	tk, _ := New("title", "", now)
	if err := tk.MarkDone(now.Add(time.Second)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !tk.Done || tk.CompletedAt == nil {
		t.Fatalf("task must be marked done with completion time")
	}
	if err := tk.MarkDone(time.Now()); err != ErrAlreadyDone {
		t.Fatalf("expected ErrAlreadyDone, got %v", err)
	}
	tk.MarkUndone()
	if tk.Done || tk.CompletedAt != nil {
		t.Fatalf("MarkUndone must reset state")
	}
}

func TestTaskMarkDoneValidation(t *testing.T) {
	now := time.Now()
	tk, _ := New("title", "", now)
	if err := tk.MarkDone(time.Time{}); err != ErrCompletedAtZero {
		t.Fatalf("expected ErrCompletedAtZero, got %v", err)
	}
}
