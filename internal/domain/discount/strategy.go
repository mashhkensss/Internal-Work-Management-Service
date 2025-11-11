package discount

import "errors"

var (
	ErrNegativeResult = errors.New("discount: negative result")
)

// Strategy encapsulates discount calculation
type Strategy interface {
	Apply(total float64) float64
}

// Fixed discounts the order
type Fixed struct {
	Amount float64
}

// None represents an absence of discount
type None struct{}

// Apply returns the discount amount for the given order total
func (f Fixed) Apply(total float64) float64 {
	if total <= 0 || f.Amount <= 0 {
		return 0
	}
	if f.Amount > total {
		return total
	}
	return f.Amount
}

// Apply always returns zero discount for the None strategy
func (None) Apply(float64) float64 { return 0 }
