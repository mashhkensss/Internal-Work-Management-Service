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

	"internal-work-management-service/internal/app/handlers/orders"
	"internal-work-management-service/internal/domain"
	"internal-work-management-service/internal/domain/discount"
	domainOrder "internal-work-management-service/internal/domain/order"
	orderService "internal-work-management-service/internal/service/order"
)

type orderServiceStub struct {
	createInput    orderService.CreateOrder
	createStrategy discount.Strategy
	createIdem     *orderService.IdempotencyKey
	createOut      domain.Order
	createErr      error

	listOut []domain.Order
	listErr error

	getID  int64
	getOut domain.Order
	getErr error

	deleteID  int64
	deleteErr error
}

func (s *orderServiceStub) Create(_ context.Context, in orderService.CreateOrder, strat discount.Strategy, idem *orderService.IdempotencyKey) (domain.Order, error) {
	s.createInput = in
	s.createStrategy = strat
	s.createIdem = idem
	return s.createOut, s.createErr
}

func (s *orderServiceStub) List(context.Context) ([]domain.Order, error) {
	return s.listOut, s.listErr
}

func (s *orderServiceStub) Get(_ context.Context, id int64) (domain.Order, error) {
	s.getID = id
	return s.getOut, s.getErr
}

func (s *orderServiceStub) Delete(_ context.Context, id int64) error {
	s.deleteID = id
	return s.deleteErr
}

func TestOrdersCreate(t *testing.T) {
	now := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	stub := &orderServiceStub{
		createOut: domain.Order{
			ID:        77,
			Customer:  "ACME",
			Items:     []domain.OrderItem{{Name: "Notebook", Price: 1000}},
			Total:     1000,
			Discount:  0,
			CreatedAt: now,
		},
	}
	handler := orders.Handler{Service: stub}

	body := `{"customer":" ACME ","items":[{"name":"Notebook","price":1000}]}`
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(body))
	req.Header.Set("Idempotency-Key", " idem-1 ")
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}
	if stub.createInput.Customer != " ACME " {
		t.Fatalf("expected raw customer forwarded, got %q", stub.createInput.Customer)
	}
	if len(stub.createInput.Items) != 1 || stub.createInput.Items[0].Name != "Notebook" {
		t.Fatalf("expected items forwarded, got %#v", stub.createInput.Items)
	}
	if _, ok := stub.createStrategy.(discount.None); !ok {
		t.Fatalf("expected discount.None strategy, got %T", stub.createStrategy)
	}
	if stub.createIdem == nil || stub.createIdem.Key != "idem-1" {
		t.Fatalf("expected idempotency key trimmed, got %#v", stub.createIdem)
	}
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["id"] != float64(77) {
		t.Fatalf("expected id 77, got %#v", resp["id"])
	}
}

func TestOrdersCreateInvalidJSON(t *testing.T) {
	handler := orders.Handler{Service: &orderServiceStub{}}
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"customer":`))
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestOrdersCreateValidationError(t *testing.T) {
	stub := &orderServiceStub{createErr: domainOrder.ErrNoItems}
	handler := orders.Handler{Service: stub}
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"customer":"A","items":[]}`))
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestOrdersCreateInternalError(t *testing.T) {
	stub := &orderServiceStub{createErr: errors.New("boom")}
	handler := orders.Handler{Service: stub}
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"customer":"A","items":[{"name":"X","price":1}]}`))
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}
}

func TestOrdersList(t *testing.T) {
	now := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	stub := &orderServiceStub{
		listOut: []domain.Order{
			{ID: 1, Customer: "A", CreatedAt: now},
			{ID: 2, Customer: "B", CreatedAt: now},
		},
	}
	handler := orders.Handler{Service: stub}

	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
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
		t.Fatalf("expected 2 orders, got %d", len(resp))
	}
}

func TestOrdersListInternalError(t *testing.T) {
	stub := &orderServiceStub{listErr: errors.New("boom")}
	handler := orders.Handler{Service: stub}
	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}
}

func TestOrdersGetByID(t *testing.T) {
	stub := &orderServiceStub{
		getOut: domain.Order{ID: 3, Customer: "A", CreatedAt: time.Now()},
	}
	handler := orders.Handler{Service: stub}

	req := httptest.NewRequest(http.MethodGet, "/orders/3", nil)
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

func TestOrdersGetInvalidID(t *testing.T) {
	handler := orders.Handler{Service: &orderServiceStub{}}
	req := httptest.NewRequest(http.MethodGet, "/orders/x", nil)
	req = withRouteParam(req, "id", "x")
	rr := httptest.NewRecorder()

	handler.GetByID(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestOrdersGetNotFound(t *testing.T) {
	stub := &orderServiceStub{getErr: domain.ErrNotFound}
	handler := orders.Handler{Service: stub}
	req := httptest.NewRequest(http.MethodGet, "/orders/5", nil)
	req = withRouteParam(req, "id", "5")
	rr := httptest.NewRecorder()

	handler.GetByID(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}
}

func TestOrdersDelete(t *testing.T) {
	stub := &orderServiceStub{}
	handler := orders.Handler{Service: stub}
	req := httptest.NewRequest(http.MethodDelete, "/orders/6", nil)
	req = withRouteParam(req, "id", "6")
	rr := httptest.NewRecorder()

	handler.Delete(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rr.Code)
	}
	if stub.deleteID != 6 {
		t.Fatalf("expected delete called with id 6, got %d", stub.deleteID)
	}
}

func TestOrdersDeleteInvalidID(t *testing.T) {
	handler := orders.Handler{Service: &orderServiceStub{}}
	req := httptest.NewRequest(http.MethodDelete, "/orders/x", nil)
	req = withRouteParam(req, "id", "x")
	rr := httptest.NewRecorder()

	handler.Delete(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestOrdersDeleteNotFound(t *testing.T) {
	stub := &orderServiceStub{deleteErr: domain.ErrNotFound}
	handler := orders.Handler{Service: stub}
	req := httptest.NewRequest(http.MethodDelete, "/orders/6", nil)
	req = withRouteParam(req, "id", "6")
	rr := httptest.NewRecorder()

	handler.Delete(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}
}
