package order

import (
	"testing"
	"time"

	"internal-work-management-service/internal/domain/discount"
)

func sampleItems() []Item {
	return []Item{
		{Name: "A", Price: 10},
		{Name: "B", Price: 5},
	}
}

type discountStrategyFunc func(total float64) float64

func (f discountStrategyFunc) Apply(total float64) float64 {
	return f(total)
}

func TestOrderNew(t *testing.T) {
	createdAt := time.Now()
	o, err := New(" customer ", sampleItems(), createdAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if o.Customer != "customer" {
		t.Fatalf("customer should be trimmed")
	}
	if o.Total != 15 {
		t.Fatalf("expected total 15, got %f", o.Total)
	}
}

func TestOrderNewValidationErrors(t *testing.T) {
	_, err := New("", sampleItems(), time.Now())
	if err != ErrEmptyCustomer {
		t.Fatalf("expected ErrEmptyCustomer, got %v", err)
	}
	_, err = New("cust", nil, time.Now())
	if err != ErrNoItems {
		t.Fatalf("expected ErrNoItems, got %v", err)
	}
	_, err = New("cust", []Item{{Name: "", Price: 1}}, time.Now())
	if err != ErrEmptyItemName {
		t.Fatalf("expected ErrEmptyItemName, got %v", err)
	}
	_, err = New("cust", []Item{{Name: "x", Price: -1}}, time.Now())
	if err != ErrNegativeItemCost {
		t.Fatalf("expected ErrNegativeItemCost, got %v", err)
	}
	_, err = New("cust", sampleItems(), time.Time{})
	if err != ErrCreatedAtZero {
		t.Fatalf("expected ErrCreatedAtZero, got %v", err)
	}
}

func TestOrderApplyDiscount(t *testing.T) {
	o, _ := New("cust", sampleItems(), time.Now())
	if err := o.ApplyDiscount(discount.Fixed{Amount: 3}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if o.Discount != 3 {
		t.Fatalf("expected discount 3, got %f", o.Discount)
	}
	if err := o.ApplyDiscount(discountStrategyFunc(func(total float64) float64 {
		return total + 1
	})); err != ErrDiscountOverflow {
		t.Fatalf("expected ErrDiscountOverflow, got %v", err)
	}
}

func TestOrderUpdateItems(t *testing.T) {
	o, _ := New("cust", sampleItems(), time.Now())
	o.Discount = 5
	if err := o.UpdateItems([]Item{{Name: "X", Price: 2}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if o.Discount != 0 {
		t.Fatalf("discount must reset when items updated")
	}
}

func TestOrderFinalAmount(t *testing.T) {
	o, _ := New("cust", sampleItems(), time.Now())
	o.Discount = 4
	if o.FinalAmount() != 11 {
		t.Fatalf("expected final amount 11, got %f", o.FinalAmount())
	}
}
