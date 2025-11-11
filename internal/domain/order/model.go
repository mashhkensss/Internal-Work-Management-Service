package order

import (
	"errors"
	"strings"
	"time"

	"internal-work-management-service/internal/domain/discount"
)

var (
	ErrInvalidID        = errors.New("order: id must be positive")
	ErrEmptyCustomer    = errors.New("order: customer name is required")
	ErrCreatedAtZero    = errors.New("order: created at must be set")
	ErrNoItems          = errors.New("order: at least one item is required")
	ErrEmptyItemName    = errors.New("order: item name is required")
	ErrNegativeItemCost = errors.New("order: item price cannot be negative")
	ErrDiscountOverflow = errors.New("order: discount cannot exceed total")
	ErrNegativeTotal    = errors.New("order: total cannot be negative")
	ErrNegativeDiscount = errors.New("order: discount cannot be negative")
)

// Item describes a single position in the order cart
type Item struct {
	Name  string
	Price float64
}

// Order holds cart items
type Order struct {
	ID        int64
	Customer  string
	Items     []Item
	Total     float64
	Discount  float64
	CreatedAt time.Time
}

// New builds a new order
func New(customer string, items []Item, createdAt time.Time) (Order, error) {
	o := Order{
		Customer:  strings.TrimSpace(customer),
		Items:     normalizeItems(items),
		CreatedAt: createdAt,
	}
	o.Total = calculateTotal(o.Items)
	if err := o.Validate(); err != nil {
		return Order{}, err
	}
	return o, nil
}

// UpdateItems replaces the order items and resets discount state
func (o *Order) UpdateItems(items []Item) error {
	o.Items = normalizeItems(items)
	o.Total = calculateTotal(o.Items)
	o.Discount = 0
	return o.Validate()
}

// UpdateCustomer updates info about customer
func (o *Order) UpdateCustomer(customer string) error {
	o.Customer = strings.TrimSpace(customer)
	return o.Validate()
}

// ApplyDiscount evaluates the provided strategy and stores the discount result
func (o *Order) ApplyDiscount(strategy discount.Strategy) error {
	if strategy == nil {
		strategy = discount.None{}
	}
	discountValue := strategy.Apply(o.Total)
	if discountValue < 0 {
		return discount.ErrNegativeResult
	}
	if discountValue > o.Total {
		return ErrDiscountOverflow
	}
	o.Discount = discountValue
	return nil
}

// Validate checks the order invariants
func (o Order) Validate() error {
	if o.Customer == "" {
		return ErrEmptyCustomer
	}
	if o.CreatedAt.IsZero() {
		return ErrCreatedAtZero
	}
	if len(o.Items) == 0 {
		return ErrNoItems
	}
	for _, item := range o.Items {
		if item.Name == "" {
			return ErrEmptyItemName
		}
		if item.Price < 0 {
			return ErrNegativeItemCost
		}
	}
	if o.Total < 0 {
		return ErrNegativeTotal
	}
	if o.Discount < 0 {
		return ErrNegativeDiscount
	}
	if o.Discount > o.Total {
		return ErrDiscountOverflow
	}
	return nil
}

func calculateTotal(items []Item) float64 {
	var total float64
	for _, item := range items {
		total += item.Price
	}
	return total
}

func normalizeItems(items []Item) []Item {
	if len(items) == 0 {
		return nil
	}
	out := make([]Item, len(items))
	for i, item := range items {
		out[i] = Item{
			Name:  strings.TrimSpace(item.Name),
			Price: item.Price,
		}
	}
	return out
}

// FinalAmount returns the amount to be charged after applying the discount
func (o Order) FinalAmount() float64 {
	return o.Total - o.Discount
}
