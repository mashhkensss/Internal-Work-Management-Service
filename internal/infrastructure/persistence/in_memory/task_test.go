package in_memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"internal-work-management-service/internal/domain"
	domainTask "internal-work-management-service/internal/domain/task"
	tasksvc "internal-work-management-service/internal/service/task"
)

func TestTaskRepoInsertAndList(t *testing.T) {
	repo := NewTaskRepo()
	ctx := context.Background()

	task1, err := repo.Insert(ctx, domain.Task{Title: "first", CreatedAt: time.Now()})
	if err != nil {
		t.Fatalf("insert 1 failed: %v", err)
	}
	task2, err := repo.Insert(ctx, domain.Task{Title: "second", CreatedAt: time.Now(), Done: true})
	if err != nil {
		t.Fatalf("insert 2 failed: %v", err)
	}
	if task1.ID != 1 || task2.ID != 2 {
		t.Fatalf("expected ids 1 and 2, got %d and %d", task1.ID, task2.ID)
	}

	all, err := repo.List(ctx, tasksvc.Filter{})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(all))
	}
	if all[0].ID != 1 || all[1].ID != 2 {
		t.Fatalf("expected tasks ordered by id, got %+v", all)
	}

	onlyDone := true
	doneTasks, err := repo.List(ctx, tasksvc.Filter{Done: &onlyDone})
	if err != nil {
		t.Fatalf("list with filter failed: %v", err)
	}
	if len(doneTasks) != 1 || doneTasks[0].ID != task2.ID {
		t.Fatalf("expected only done task, got %+v", doneTasks)
	}
}

func TestTaskRepoMarkDone(t *testing.T) {
	repo := NewTaskRepo()
	ctx := context.Background()

	inserted, err := repo.Insert(ctx, domain.Task{Title: "todo", CreatedAt: time.Now()})
	if err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	doneTask, err := repo.MarkDone(ctx, inserted.ID)
	if err != nil {
		t.Fatalf("mark done failed: %v", err)
	}
	if !doneTask.Done {
		t.Fatalf("expected task to be done")
	}
	if doneTask.CompletedAt == nil {
		t.Fatalf("expected completedAt to be set")
	}

	doneTask2, err := repo.MarkDone(ctx, inserted.ID)
	if !errors.Is(err, domainTask.ErrAlreadyDone) {
		t.Fatalf("expected ErrAlreadyDone, got %v", err)
	}
	if !doneTask2.Done || doneTask2.CompletedAt == nil {
		t.Fatalf("expected returned task to stay done")
	}

	if _, err := repo.MarkDone(ctx, 999); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for missing task, got %v", err)
	}
}
