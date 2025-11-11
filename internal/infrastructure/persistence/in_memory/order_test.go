package in_memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"internal-work-management-service/internal/domain"
	ordersvc "internal-work-management-service/internal/service/order"
)

func TestOrderRepoInsertAndList(t *testing.T) {
	repo := NewOrderRepo()
	ctx := context.Background()

	order := domain.Order{
		Customer:  "ACME",
		Items:     []domain.OrderItem{{Name: "Notebook", Price: 1000}},
		Total:     1000,
		CreatedAt: time.Now(),
	}

	created, err := repo.Insert(ctx, order, nil)
	if err != nil {
		t.Fatalf("insert failed: %v", err)
	}
	if created.ID != 1 {
		t.Fatalf("expected id 1, got %d", created.ID)
	}
	if len(created.Items) != 1 || created.Items[0].Name != "Notebook" {
		t.Fatalf("unexpected items %+v", created.Items)
	}

	list, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(list) != 1 || list[0].ID != created.ID {
		t.Fatalf("expected single order in list, got %+v", list)
	}
	if &list[0].Items[0] == &created.Items[0] {
		t.Fatalf("expected list items to be cloned")
	}
}

func TestOrderRepoIdempotency(t *testing.T) {
	repo := NewOrderRepo()
	ctx := context.Background()
	key := &ordersvc.IdempotencyKey{Key: "key-1"}

	base := domain.Order{
		Customer:  "ACME",
		Items:     []domain.OrderItem{{Name: "Notebook", Price: 1000}},
		Total:     1000,
		CreatedAt: time.Now(),
	}

	first, err := repo.Insert(ctx, base, key)
	if err != nil {
		t.Fatalf("first insert failed: %v", err)
	}

	// simulate retry with altered payload that should be ignored
	retryPayload := base
	retryPayload.Customer = "SHOULD-NOT-CHANGE"
	retry, err := repo.Insert(ctx, retryPayload, key)
	if err != nil {
		t.Fatalf("retry insert failed: %v", err)
	}
	if retry.ID != first.ID {
		t.Fatalf("expected same id on retry, got %d vs %d", retry.ID, first.ID)
	}
	if retry.Customer != first.Customer {
		t.Fatalf("expected customer %q, got %q", first.Customer, retry.Customer)
	}

	list, _ := repo.List(ctx)
	if len(list) != 1 {
		t.Fatalf("expected single stored order, got %d", len(list))
	}
}

func TestOrderRepoGetAndDelete(t *testing.T) {
	repo := NewOrderRepo()
	ctx := context.Background()

	order := domain.Order{
		Customer:  "ACME",
		Items:     []domain.OrderItem{{Name: "Notebook", Price: 1000}},
		Total:     1000,
		CreatedAt: time.Now(),
	}

	created, err := repo.Insert(ctx, order, nil)
	if err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	found, err := repo.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if found.ID != created.ID {
		t.Fatalf("expected id %d, got %d", created.ID, found.ID)
	}
	found.Items[0].Name = "Mutated"

	foundAgain, err := repo.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if foundAgain.Items[0].Name == "Mutated" {
		t.Fatalf("expected items to be cloned on get")
	}

	if err := repo.Delete(ctx, created.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	if _, err := repo.Get(ctx, created.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
	if err := repo.Delete(ctx, created.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound on repeat delete, got %v", err)
	}
}
