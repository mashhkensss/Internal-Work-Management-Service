package in_memory

import (
	"context"
	"sort"
	"sync"

	"internal-work-management-service/internal/domain"
	domainorder "internal-work-management-service/internal/domain/order"
	ordersvc "internal-work-management-service/internal/service/order"
)

// OrderRepo is an in-memory order storage used in tests.
type OrderRepo struct {
	mu          sync.RWMutex
	idx         int64
	items       map[int64]domain.Order
	idemMapping map[string]int64
}

func NewOrderRepo() *OrderRepo {
	return &OrderRepo{
		items:       make(map[int64]domain.Order),
		idemMapping: make(map[string]int64),
	}
}

func (r *OrderRepo) Insert(_ context.Context, o domain.Order, idem *ordersvc.IdempotencyKey) (domain.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if idem != nil {
		if existingID, ok := r.idemMapping[idem.Key]; ok {
			if order, exists := r.items[existingID]; exists {
				return cloneOrder(order), nil
			}
		}
	}

	r.idx++
	o.ID = r.idx
	o.Items = cloneItems(o.Items)
	r.items[o.ID] = o
	if idem != nil {
		r.idemMapping[idem.Key] = o.ID
	}
	return cloneOrder(o), nil
}

func (r *OrderRepo) List(_ context.Context) ([]domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	orders := make([]domain.Order, 0, len(r.items))
	for _, o := range r.items {
		orders = append(orders, cloneOrder(o))
	}
	sort.Slice(orders, func(i, j int) bool { return orders[i].ID < orders[j].ID })
	return orders, nil
}

func (r *OrderRepo) Get(_ context.Context, id int64) (domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	o, ok := r.items[id]
	if !ok {
		return domain.Order{}, domain.ErrNotFound
	}
	return cloneOrder(o), nil
}

func (r *OrderRepo) Delete(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.items[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.items, id)
	for key, storedID := range r.idemMapping {
		if storedID == id {
			delete(r.idemMapping, key)
		}
	}
	return nil
}

func cloneOrder(o domain.Order) domain.Order {
	clone := o
	clone.Items = cloneItems(o.Items)
	return clone
}

func cloneItems(items []domainorder.Item) []domainorder.Item {
	if len(items) == 0 {
		return nil
	}
	out := make([]domainorder.Item, len(items))
	copy(out, items)
	return out
}
