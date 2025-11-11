package task

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidID       = errors.New("task: id must be positive")
	ErrEmptyTitle      = errors.New("task: title is required")
	ErrCreatedAtZero   = errors.New("task: created at must be set")
	ErrCompletedAtZero = errors.New("task: completed at must be set when task is done")
	ErrAlreadyDone     = errors.New("task: already marked as done")
)

// Task represents a to-do task
type Task struct {
	ID          int64
	Title       string
	Description string
	Done        bool
	CreatedAt   time.Time
	CompletedAt *time.Time
}

// New constructs a task
func New(title, description string, createdAt time.Time) (Task, error) {
	t := Task{
		Title:       strings.TrimSpace(title),
		Description: strings.TrimSpace(description),
		CreatedAt:   createdAt,
	}
	if err := t.Validate(); err != nil {
		return Task{}, err
	}
	return t, nil
}

// Update rewrites title and description
func (t *Task) Update(title, description string) error {
	t.Title = strings.TrimSpace(title)
	t.Description = strings.TrimSpace(description)
	if err := t.Validate(); err != nil {
		return err
	}
	return nil
}

// MarkDone marks the task as completed at some provided time
func (t *Task) MarkDone(at time.Time) error {
	if t.Done {
		return ErrAlreadyDone
	}
	if at.IsZero() {
		return ErrCompletedAtZero
	}
	completion := at
	t.Done = true
	t.CompletedAt = &completion
	return t.Validate()
}

// MarkUndone clears completion state
func (t *Task) MarkUndone() {
	t.Done = false
	t.CompletedAt = nil
}

// Validate ensures task invariants
func (t Task) Validate() error {
	if t.Title == "" {
		return ErrEmptyTitle
	}
	if t.CreatedAt.IsZero() {
		return ErrCreatedAtZero
	}
	if t.Done {
		if t.CompletedAt == nil || t.CompletedAt.IsZero() {
			return ErrCompletedAtZero
		}
	}
	return nil
}
