package order

import (
	"context"
	"errors"
	"testing"

	"internal-work-management-service/internal/domain"
	"internal-work-management-service/internal/domain/discount"
	domainOrder "internal-work-management-service/internal/domain/order"
	"internal-work-management-service/tests/testutils"
)

type orderRepoMock struct {
	insertedOrder domain.Order
	insertIdem    *IdempotencyKey
	insertErr     error
	insertResult  *domain.Order

	listResult []domain.Order
	listErr    error

	getResult domain.Order
	getErr    error

	deleteErr error
}

func (m *orderRepoMock) Insert(_ context.Context, o domain.Order, idem *IdempotencyKey) (domain.Order, error) {
	if m.insertErr != nil {
		return domain.Order{}, m.insertErr
	}
	m.insertedOrder = o
	m.insertIdem = idem
	if m.insertResult != nil {
		return *m.insertResult, nil
	}
	created, err := testutils.OrderWithID(o, 1)
	if err != nil {
		return domain.Order{}, err
	}
	return created, nil
}

func (m *orderRepoMock) List(context.Context) ([]domain.Order, error) {
	return m.listResult, m.listErr
}

func (m *orderRepoMock) Get(context.Context, int64) (domain.Order, error) {
	return m.getResult, m.getErr
}

func (m *orderRepoMock) Delete(context.Context, int64) error {
	return m.deleteErr
}

func sampleOrderItems() []domainOrder.Item {
	return []domainOrder.Item{
		{Name: " first ", Price: 10},
		{Name: " second ", Price: 5},
	}
}

func TestOrderServiceCreate(t *testing.T) {
	mockRepo := &orderRepoMock{}
	svc := New(mockRepo, nil)

	created, err := svc.Create(
		context.Background(),
		CreateOrder{
			Customer: "  ACME  ",
			Items:    sampleOrderItems(),
		},
		discount.Fixed{Amount: 3},
		&IdempotencyKey{Key: "idem-123"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("expected created order to have ID")
	}
	if mockRepo.insertedOrder.Customer != "ACME" {
		t.Fatalf("expected trimmed customer, got %q", mockRepo.insertedOrder.Customer)
	}
	if len(mockRepo.insertedOrder.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(mockRepo.insertedOrder.Items))
	}
	if mockRepo.insertedOrder.Items[0].Name != "first" {
		t.Fatalf("item name should be trimmed, got %q", mockRepo.insertedOrder.Items[0].Name)
	}
	if mockRepo.insertedOrder.Discount != 3 {
		t.Fatalf("expected discount 3, got %f", mockRepo.insertedOrder.Discount)
	}
	if mockRepo.insertIdem == nil || mockRepo.insertIdem.Key != "idem-123" {
		t.Fatalf("expected idempotency key to be forwarded, got %#v", mockRepo.insertIdem)
	}
}

func TestOrderServiceCreateEmptyItems(t *testing.T) {
	svc := New(&orderRepoMock{}, nil)
	_, err := svc.Create(
		context.Background(),
		CreateOrder{Customer: "cust", Items: nil},
		discount.None{},
		nil,
	)
	if !errors.Is(err, domainOrder.ErrNoItems) {
		t.Fatalf("expected ErrNoItems, got %v", err)
	}
}

func TestOrderServiceCreateDomain(t *testing.T) {
	svc := New(&orderRepoMock{}, nil)
	_, err := svc.Create(
		context.Background(),
		CreateOrder{
			Customer: "cust",
			Items: []domainOrder.Item{
				{Name: "item", Price: -1},
			},
		},
		discount.None{},
		nil,
	)
	if !errors.Is(err, domainOrder.ErrNegativeItemCost) {
		t.Fatalf("expected ErrNegativeItemCost, got %v", err)
	}
}

func TestOrderServiceCreateRepo(t *testing.T) {
	wantErr := errors.New("insert failed")
	svc := New(&orderRepoMock{insertErr: wantErr}, nil)
	_, err := svc.Create(
		context.Background(),
		CreateOrder{
			Customer: "cust",
			Items: []domainOrder.Item{
				{Name: "item", Price: 1},
			},
		},
		discount.None{},
		nil,
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}

func TestOrderServiceList(t *testing.T) {
	expected := []domain.Order{{ID: 1}, {ID: 2}}
	svc := New(&orderRepoMock{listResult: expected}, nil)
	got, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(expected) {
		t.Fatalf("expected %d orders, got %d", len(expected), len(got))
	}
}

func TestOrderServiceGetNotFound(t *testing.T) {
	svc := New(&orderRepoMock{getErr: domain.ErrNotFound}, nil)
	_, err := svc.Get(context.Background(), 1)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected domain.ErrNotFound, got %v", err)
	}
}

func TestOrderServiceGetRepo(t *testing.T) {
	wantErr := errors.New("unexpected")
	svc := New(&orderRepoMock{getErr: wantErr}, nil)
	_, err := svc.Get(context.Background(), 2)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}

func TestOrderServiceGet(t *testing.T) {
	expected := domain.Order{ID: 5}
	svc := New(&orderRepoMock{getResult: expected}, nil)
	got, err := svc.Get(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != expected.ID {
		t.Fatalf("expected order ID %d, got %d", expected.ID, got.ID)
	}
}

func TestOrderServiceDelete(t *testing.T) {
	svc := New(&orderRepoMock{deleteErr: domain.ErrNotFound}, nil)
	if err := svc.Delete(context.Background(), 1); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected domain.ErrNotFound, got %v", err)
	}
}

func TestOrderServiceDeleteRepo(t *testing.T) {
	wantErr := errors.New("delete failed")
	svc := New(&orderRepoMock{deleteErr: wantErr}, nil)
	if err := svc.Delete(context.Background(), 1); !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}

func TestOrderServiceDeleteSuccess(t *testing.T) {
	svc := New(&orderRepoMock{}, nil)
	if err := svc.Delete(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
