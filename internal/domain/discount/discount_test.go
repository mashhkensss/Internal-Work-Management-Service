package discount

import (
	"testing"
)

func TestFixedDiscount(t *testing.T) {
	strat := Fixed{Amount: 10}
	if got := strat.Apply(0); got != 0 {
		t.Fatalf("expected 0 for zero total, got %f", got)
	}
	if got := strat.Apply(5); got != 5 {
		t.Fatalf("expected capped discount, got %f", got)
	}
	if got := strat.Apply(20); got != 10 {
		t.Fatalf("expected discount amount 10, got %f", got)
	}
}

func TestNoneDiscount(t *testing.T) {
	if (None{}).Apply(100) != 0 {
		t.Fatalf("none strategy must return 0")
	}
}
