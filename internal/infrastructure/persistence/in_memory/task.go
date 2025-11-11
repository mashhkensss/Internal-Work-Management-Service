package in_memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"internal-work-management-service/internal/domain"
	domainTask "internal-work-management-service/internal/domain/task"
	tasksvc "internal-work-management-service/internal/service/task"
)

// TaskRepo is an in-memory task storage used in tests.
type TaskRepo struct {
	mu    sync.RWMutex
	idx   int64
	items map[int64]domain.Task
}

func NewTaskRepo() *TaskRepo {
	return &TaskRepo{
		items: make(map[int64]domain.Task),
	}
}

func (r *TaskRepo) Insert(_ context.Context, t domain.Task) (domain.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.idx++
	t.ID = r.idx
	r.items[t.ID] = t
	return t, nil
}

func (r *TaskRepo) List(_ context.Context, filter tasksvc.Filter) ([]domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tasks := make([]domain.Task, 0, len(r.items))
	for _, task := range r.items {
		if filter.Done != nil && task.Done != *filter.Done {
			continue
		}
		tasks = append(tasks, task)
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].ID < tasks[j].ID })
	return tasks, nil
}

func (r *TaskRepo) MarkDone(_ context.Context, id int64) (domain.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, ok := r.items[id]
	if !ok {
		return domain.Task{}, domain.ErrNotFound
	}

	if task.Done {
		return task, domainTask.ErrAlreadyDone
	}

	now := time.Now().UTC()
	task.Done = true
	task.CompletedAt = &now
	r.items[id] = task
	return task, nil
}
