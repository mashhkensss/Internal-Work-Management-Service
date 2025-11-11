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

	"internal-work-management-service/internal/app/handlers/notes"
	"internal-work-management-service/internal/domain"
	domainNote "internal-work-management-service/internal/domain/note"
	noteService "internal-work-management-service/internal/service/note"
)

type noteServiceStub struct {
	createInput noteService.CreateNote
	createOut   domain.Note
	createErr   error

	listOut []domain.Note
	listErr error

	getID  int64
	getOut domain.Note
	getErr error

	deleteID  int64
	deleteErr error
}

func (s *noteServiceStub) Create(_ context.Context, in noteService.CreateNote) (domain.Note, error) {
	s.createInput = in
	return s.createOut, s.createErr
}

func (s *noteServiceStub) List(context.Context) ([]domain.Note, error) {
	return s.listOut, s.listErr
}

func (s *noteServiceStub) Get(_ context.Context, id int64) (domain.Note, error) {
	s.getID = id
	return s.getOut, s.getErr
}

func (s *noteServiceStub) Delete(_ context.Context, id int64) error {
	s.deleteID = id
	return s.deleteErr
}

func TestNotesCreate(t *testing.T) {
	createdAt := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	stub := &noteServiceStub{
		createOut: domain.Note{
			ID:        10,
			Title:     "Note",
			Content:   "Body",
			CreatedAt: createdAt,
		},
	}
	handler := notes.Handler{Service: stub}

	req := httptest.NewRequest(http.MethodPost, "/notes", strings.NewReader(`{"title":" Note ","content":" Body "}`))
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}
	if stub.createInput.Title != " Note " || stub.createInput.Content != " Body " {
		t.Fatalf("expected raw payload forwarded to service, got %+v", stub.createInput)
	}
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["id"] != float64(10) {
		t.Fatalf("expected id 10, got %#v", resp["id"])
	}
	if resp["created_at"] != createdAt.Format(time.RFC3339) {
		t.Fatalf("unexpected created_at: %v", resp["created_at"])
	}
}

func TestNotesCreateInvalidJSON(t *testing.T) {
	handler := notes.Handler{Service: &noteServiceStub{}}
	req := httptest.NewRequest(http.MethodPost, "/notes", strings.NewReader(`{"title":`))
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestNotesCreateValidationError(t *testing.T) {
	stub := &noteServiceStub{createErr: domainNote.ErrEmptyTitle}
	handler := notes.Handler{Service: stub}
	req := httptest.NewRequest(http.MethodPost, "/notes", strings.NewReader(`{"title":"","content":"x"}`))
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestNotesCreateInternalError(t *testing.T) {
	stub := &noteServiceStub{createErr: errors.New("boom")}
	handler := notes.Handler{Service: stub}
	req := httptest.NewRequest(http.MethodPost, "/notes", strings.NewReader(`{"title":"x","content":"y"}`))
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}
}

func TestNotesListAll(t *testing.T) {
	createdAt := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	stub := &noteServiceStub{
		listOut: []domain.Note{
			{ID: 1, Title: "A", Content: "B", CreatedAt: createdAt},
			{ID: 2, Title: "C", Content: "D", CreatedAt: createdAt},
		},
	}
	handler := notes.Handler{Service: stub}

	req := httptest.NewRequest(http.MethodGet, "/notes", nil)
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
		t.Fatalf("expected 2 notes, got %d", len(resp))
	}
}

func TestNotesListByID(t *testing.T) {
	stub := &noteServiceStub{
		getOut: domain.Note{ID: 5, Title: "A", Content: "B", CreatedAt: time.Now()},
	}
	handler := notes.Handler{Service: stub}

	req := httptest.NewRequest(http.MethodGet, "/notes?id=5", nil)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if stub.getID != 5 {
		t.Fatalf("expected Get to be called with id 5, got %d", stub.getID)
	}
}

func TestNotesListInvalidID(t *testing.T) {
	handler := notes.Handler{Service: &noteServiceStub{}}
	req := httptest.NewRequest(http.MethodGet, "/notes?id=abc", nil)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestNotesListNotFound(t *testing.T) {
	stub := &noteServiceStub{getErr: domain.ErrNotFound}
	handler := notes.Handler{Service: stub}
	req := httptest.NewRequest(http.MethodGet, "/notes?id=5", nil)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}
}

func TestNotesGetByID(t *testing.T) {
	stub := &noteServiceStub{
		getOut: domain.Note{ID: 3, Title: "A", Content: "B", CreatedAt: time.Now()},
	}
	handler := notes.Handler{Service: stub}

	req := httptest.NewRequest(http.MethodGet, "/notes/3", nil)
	req = withRouteParam(req, "id", "3")
	rr := httptest.NewRecorder()

	handler.GetByID(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if stub.getID != 3 {
		t.Fatalf("expected get called with id 3, got %d", stub.getID)
	}
}

func TestNotesGetInvalidID(t *testing.T) {
	handler := notes.Handler{Service: &noteServiceStub{}}
	req := httptest.NewRequest(http.MethodGet, "/notes/x", nil)
	req = withRouteParam(req, "id", "x")
	rr := httptest.NewRecorder()

	handler.GetByID(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestNotesGetNotFound(t *testing.T) {
	stub := &noteServiceStub{getErr: domain.ErrNotFound}
	handler := notes.Handler{Service: stub}
	req := httptest.NewRequest(http.MethodGet, "/notes/5", nil)
	req = withRouteParam(req, "id", "5")
	rr := httptest.NewRecorder()

	handler.GetByID(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}
}

func TestNotesDelete(t *testing.T) {
	stub := &noteServiceStub{}
	handler := notes.Handler{Service: stub}
	req := httptest.NewRequest(http.MethodDelete, "/notes?id=4", nil)
	rr := httptest.NewRecorder()

	handler.Delete(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rr.Code)
	}
	if stub.deleteID != 4 {
		t.Fatalf("expected delete called with id 4, got %d", stub.deleteID)
	}
}

func TestNotesDeleteInvalidID(t *testing.T) {
	handler := notes.Handler{Service: &noteServiceStub{}}
	req := httptest.NewRequest(http.MethodDelete, "/notes?id=x", nil)
	rr := httptest.NewRecorder()

	handler.Delete(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestNotesDeleteNotFound(t *testing.T) {
	stub := &noteServiceStub{deleteErr: domain.ErrNotFound}
	handler := notes.Handler{Service: stub}
	req := httptest.NewRequest(http.MethodDelete, "/notes/4", nil)
	req = withRouteParam(req, "id", "4")
	rr := httptest.NewRecorder()

	handler.Delete(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}
}
