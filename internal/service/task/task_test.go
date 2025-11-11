package task

import (
	"context"
	"errors"
	"testing"
	"time"

	"internal-work-management-service/internal/domain"
	domainTask "internal-work-management-service/internal/domain/task"
)

type taskRepoMock struct {
	inserted  domain.Task
	insertErr error

	listResult []domain.Task
	listErr    error

	markDoneResult domain.Task
	markDoneErr    error
}

func (m *taskRepoMock) Insert(_ context.Context, t domain.Task) (domain.Task, error) {
	if m.insertErr != nil {
		return domain.Task{}, m.insertErr
	}
	m.inserted = t
	if t.ID == 0 {
		t.ID = 1
	}
	return t, nil
}

func (m *taskRepoMock) List(_ context.Context, _ Filter) ([]domain.Task, error) {
	return m.listResult, m.listErr
}

func (m *taskRepoMock) MarkDone(_ context.Context, _ int64) (domain.Task, error) {
	return m.markDoneResult, m.markDoneErr
}

func TestTaskServiceCreate(t *testing.T) {
	mockRepo := &taskRepoMock{}
	svc := New(mockRepo, nil)
	task, err := svc.Create(context.Background(), CreateTask{Title: "  title ", Description: " desc "})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mockRepo.inserted.Title != "title" || mockRepo.inserted.Description != "desc" {
		t.Fatalf("task should be trimmed before insert: %+v", mockRepo.inserted)
	}
	if task.ID == 0 {
		t.Fatalf("expected id to be set")
	}
}

func TestTaskServiceCreateValidation(t *testing.T) {
	svc := New(&taskRepoMock{}, nil)
	_, err := svc.Create(context.Background(), CreateTask{Title: " "})
	if !errors.Is(err, domainTask.ErrEmptyTitle) {
		t.Fatalf("expected domainTask.ErrEmptyTitle, got %v", err)
	}
}

func TestTaskServiceCreateRepo(t *testing.T) {
	wantErr := errors.New("insert failed")
	svc := New(&taskRepoMock{insertErr: wantErr}, nil)
	_, err := svc.Create(context.Background(), CreateTask{Title: "title"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}

func TestTaskServiceList(t *testing.T) {
	expected := []domain.Task{{ID: 1}}
	svc := New(&taskRepoMock{listResult: expected}, nil)
	got, err := svc.List(context.Background(), Filter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(expected) {
		t.Fatalf("expected %d tasks, got %d", len(expected), len(got))
	}
}

func TestTaskServiceMarkDone(t *testing.T) {
	svc := New(&taskRepoMock{markDoneErr: domain.ErrNotFound}, nil)
	_, err := svc.MarkDone(context.Background(), 1)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected domain.ErrNotFound, got %v", err)
	}
}

func TestTaskServiceMarkDoneAlreadyDone(t *testing.T) {
	now := time.Now()
	doneTask := domain.Task{ID: 1, Title: "t", CreatedAt: now, Done: true, CompletedAt: &now}
	svc := New(&taskRepoMock{
		markDoneResult: doneTask,
		markDoneErr:    domainTask.ErrAlreadyDone,
	}, nil)
	got, err := svc.MarkDone(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !got.Done {
		t.Fatalf("expected done task")
	}
}

func TestTaskServiceMarkDoneOtherError(t *testing.T) {
	wantErr := errors.New("unexpected")
	svc := New(&taskRepoMock{markDoneErr: wantErr}, nil)
	_, err := svc.MarkDone(context.Background(), 1)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}
