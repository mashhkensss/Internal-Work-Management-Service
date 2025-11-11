package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"internal-work-management-service/internal/app/handlers/tasks"
	"internal-work-management-service/internal/domain"
	domainTask "internal-work-management-service/internal/domain/task"
	taskService "internal-work-management-service/internal/service/task"
)

type taskServiceStub struct {
	createInput taskService.CreateTask
	createOut   domain.Task
	createErr   error

	listInput taskService.Filter
	listOut   []domain.Task
	listErr   error

	markDoneID  int64
	markDoneOut domain.Task
	markDoneErr error
}

func (s *taskServiceStub) Create(_ context.Context, in taskService.CreateTask) (domain.Task, error) {
	s.createInput = in
	return s.createOut, s.createErr
}

func (s *taskServiceStub) List(_ context.Context, filter taskService.Filter) ([]domain.Task, error) {
	s.listInput = filter
	return s.listOut, s.listErr
}

func (s *taskServiceStub) MarkDone(_ context.Context, id int64) (domain.Task, error) {
	s.markDoneID = id
	return s.markDoneOut, s.markDoneErr
}

func TestTasksCreate(t *testing.T) {
	now := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	stub := &taskServiceStub{
		createOut: domain.Task{
			ID:        11,
			Title:     "Task",
			Done:      false,
			CreatedAt: now,
		},
	}
	handler := tasks.Handler{Service: stub}

	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(`{"title":" Do ","description":" Work "}`))
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}
	if stub.createInput.Title != " Do " || stub.createInput.Description != " Work " {
		t.Fatalf("expected raw payload forwarded to service, got %+v", stub.createInput)
	}
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["id"] != float64(11) {
		t.Fatalf("expected id 11, got %#v", resp["id"])
	}
}

func TestTasksCreateInvalidJSON(t *testing.T) {
	handler := tasks.Handler{Service: &taskServiceStub{}}
	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(`{"title":`))
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestTasksCreateValidationError(t *testing.T) {
	stub := &taskServiceStub{createErr: domainTask.ErrEmptyTitle}
	handler := tasks.Handler{Service: stub}
	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(`{"title":"","description":""}`))
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestTasksCreateInternalError(t *testing.T) {
	stub := &taskServiceStub{createErr: errors.New("boom")}
	handler := tasks.Handler{Service: stub}
	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(`{"title":"x"}`))
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}
}

func TestTasksList(t *testing.T) {
	now := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	stub := &taskServiceStub{
		listOut: []domain.Task{
			{ID: 1, Title: "A", CreatedAt: now},
			{ID: 2, Title: "B", CreatedAt: now},
		},
	}
	handler := tasks.Handler{Service: stub}

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var resp []map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(resp) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(resp))
	}
}

func TestTasksListWithFilter(t *testing.T) {
	stub := &taskServiceStub{}
	handler := tasks.Handler{Service: stub}

	req := httptest.NewRequest(http.MethodGet, "/tasks?done=true", nil)
	rr := httptest.NewRecorder()
	handler.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if stub.listInput.Done == nil || !*stub.listInput.Done {
		t.Fatalf("expected filter.Done=true, got %#v", stub.listInput.Done)
	}
}

func TestTasksListInvalidFilter(t *testing.T) {
	handler := tasks.Handler{Service: &taskServiceStub{}}
	req := httptest.NewRequest(http.MethodGet, "/tasks?done=maybe", nil)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestTasksListInternalError(t *testing.T) {
	stub := &taskServiceStub{listErr: errors.New("boom")}
	handler := tasks.Handler{Service: stub}
	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}
}

func TestTasksMarkDone(t *testing.T) {
	completed := time.Date(2025, 1, 2, 4, 5, 6, 0, time.UTC)
	stub := &taskServiceStub{
		markDoneOut: domain.Task{
			ID:          9,
			Title:       "T",
			Done:        true,
			CreatedAt:   completed.Add(-time.Hour),
			CompletedAt: &completed,
		},
	}
	handler := tasks.Handler{Service: stub}

	req := httptest.NewRequest(http.MethodPost, "/tasks/9/done", nil)
	req = withRouteParam(req, "id", "9")
	rr := httptest.NewRecorder()

	handler.MarkDone(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if stub.markDoneID != 9 {
		t.Fatalf("expected markDone called with id 9, got %d", stub.markDoneID)
	}
}

func TestTasksMarkDoneInvalidID(t *testing.T) {
	handler := tasks.Handler{Service: &taskServiceStub{}}
	req := httptest.NewRequest(http.MethodPost, "/tasks/x/done", nil)
	req = withRouteParam(req, "id", "x")
	rr := httptest.NewRecorder()

	handler.MarkDone(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestTasksMarkDoneNotFound(t *testing.T) {
	stub := &taskServiceStub{markDoneErr: domain.ErrNotFound}
	handler := tasks.Handler{Service: stub}
	req := httptest.NewRequest(http.MethodPost, "/tasks/5/done", nil)
	req = withRouteParam(req, "id", "5")
	rr := httptest.NewRecorder()

	handler.MarkDone(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}
}

func TestTasksMarkDoneInternalError(t *testing.T) {
	stub := &taskServiceStub{markDoneErr: errors.New("boom")}
	handler := tasks.Handler{Service: stub}
	req := httptest.NewRequest(http.MethodPost, "/tasks/5/done", nil)
	req = withRouteParam(req, "id", "5")
	rr := httptest.NewRecorder()

	handler.MarkDone(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}
}
