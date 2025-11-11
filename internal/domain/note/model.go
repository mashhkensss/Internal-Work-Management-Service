package note

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidID     = errors.New("note: id must be positive")
	ErrEmptyTitle    = errors.New("note: title is required")
	ErrEmptyContent  = errors.New("note: content is required")
	ErrCreatedAtZero = errors.New("note: created at must be set")
)

// Note represents a user note
type Note struct {
	ID        int64
	Title     string
	Content   string
	CreatedAt time.Time
}

// New creates a note without id
func New(title, content string, createdAt time.Time) (Note, error) {
	n := Note{
		Title:     strings.TrimSpace(title),
		Content:   strings.TrimSpace(content),
		CreatedAt: createdAt,
	}
	if err := n.Validate(); err != nil {
		return Note{}, err
	}
	return n, nil
}

// Update updates title and content
func (n *Note) Update(title, content string) error {
	n.Title = strings.TrimSpace(title)
	n.Content = strings.TrimSpace(content)
	if err := n.Validate(); err != nil {
		return err
	}
	return nil
}

// Validate checks invariants required for a note to be considered valid
func (n Note) Validate() error {
	if n.Title == "" {
		return ErrEmptyTitle
	}
	if n.Content == "" {
		return ErrEmptyContent
	}
	if n.CreatedAt.IsZero() {
		return ErrCreatedAtZero
	}
	return nil
}
