package in_memory

import (
	"context"
	"sort"
	"sync"

	"internal-work-management-service/internal/domain"
)

// NoteRepo is an in-memory note storage used in tests.
type NoteRepo struct {
	mu    sync.RWMutex
	idx   int64
	items map[int64]domain.Note
}

func NewNoteRepo() *NoteRepo {
	return &NoteRepo{
		items: make(map[int64]domain.Note),
	}
}

func (r *NoteRepo) Insert(_ context.Context, n domain.Note) (domain.Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.idx++
	n.ID = r.idx
	r.items[n.ID] = n
	return n, nil
}

func (r *NoteRepo) List(_ context.Context) ([]domain.Note, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.items) == 0 {
		return nil, nil
	}

	notes := make([]domain.Note, 0, len(r.items))
	for _, n := range r.items {
		notes = append(notes, n)
	}
	sort.Slice(notes, func(i, j int) bool { return notes[i].ID < notes[j].ID })
	return notes, nil
}

func (r *NoteRepo) Get(_ context.Context, id int64) (domain.Note, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	n, ok := r.items[id]
	if !ok {
		return domain.Note{}, domain.ErrNotFound
	}
	return n, nil
}

func (r *NoteRepo) Delete(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.items[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.items, id)
	return nil
}
