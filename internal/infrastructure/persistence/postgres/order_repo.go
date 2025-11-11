package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"

	"internal-work-management-service/internal/domain"
	domainorder "internal-work-management-service/internal/domain/order"
	ordersvc "internal-work-management-service/internal/service/order"
)

// OrderRepo stores orders in PostgreSQL
type OrderRepo struct {
	Pool ConnPool
}

func NewOrderRepo(pool ConnPool) *OrderRepo {
	return &OrderRepo{Pool: pool}
}

func (r OrderRepo) Insert(ctx context.Context, o domain.Order, idem *ordersvc.IdempotencyKey) (stored domain.Order, err error) {
	tx, err := r.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.Order{}, fmt.Errorf("order insert begin: %w", err)
	}
	commit := true
	defer func() {
		if !commit {
			tx.Rollback(ctx)
			return
		}
		if err != nil {
			tx.Rollback(ctx)
			return
		}
		if commitErr := tx.
			Commit(ctx); commitErr != nil {
			err = fmt.Errorf("order insert commit err: %w", commitErr)
		}
	}()

	if idem != nil {
		if existing, err := r.handleIdempotency(ctx, tx, idem); err != nil {
			return domain.Order{}, err
		} else if existing != nil {
			commit = false
			return *existing, nil
		}
	}

	stored, err = r.insertOrder(ctx, tx, o)
	if err != nil {
		return domain.Order{}, err
	}

	if idem != nil {
		if err := r.updateIdempotency(ctx, tx, idem.Key, stored.ID); err != nil {
			return domain.Order{}, err
		}
	}

	return stored, nil
}

func (r OrderRepo) List(ctx context.Context) ([]domain.Order, error) {
	orders, err := r.fetchOrders(ctx, nil)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (r OrderRepo) Get(ctx context.Context, id int64) (domain.Order, error) {
	orders, err := r.fetchOrders(ctx, []int64{id})
	if err != nil {
		return domain.Order{}, err
	}
	if len(orders) == 0 {
		return domain.Order{}, domain.ErrNotFound
	}
	return orders[0], nil
}

func (r OrderRepo) Delete(ctx context.Context, id int64) error {
	query := NewStmtBuilder().
		Delete("orders").
		Where(squirrel.Eq{"id": id})

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("order delete err sql: %w", err)
	}

	cmd, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("order delete err exec: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r OrderRepo) handleIdempotency(ctx context.Context, tx pgx.Tx, idem *ordersvc.IdempotencyKey) (*domain.Order, error) {
	selectSQL := "SELECT order_id FROM idempotency WHERE key = $1 FOR UPDATE"
	var existingOrderID sql.NullInt64
	err := tx.QueryRow(ctx, selectSQL, idem.Key).Scan(&existingOrderID)

	switch {
	case err == nil:
		if existingOrderID.Valid {
			order, err := r.fetchOrderTx(ctx, tx, existingOrderID.Int64)
			if err != nil {
				return nil, err
			}
			return &order, nil
		}
	case errors.Is(err, pgx.ErrNoRows):
		if _, err := tx.
			Exec(ctx, "INSERT INTO idempotency (key, created_at) VALUES ($1, NOW())", idem.Key); err != nil {
			return nil, fmt.Errorf("order idempotency insert err: %w", err)
		}
	default:
		return nil, fmt.Errorf("order idempotency select err: %w", err)
	}
	return nil, nil
}

func (r OrderRepo) updateIdempotency(ctx context.Context, tx pgx.Tx, key string, orderID int64) error {
	if _, err := tx.
		Exec(ctx, "UPDATE idempotency SET order_id = $1, updated_at = NOW() WHERE key = $2", orderID, key); err != nil {
		return fmt.Errorf("order idempotency update err: %w", err)
	}

	return nil
}

func (r OrderRepo) insertOrder(ctx context.Context, tx pgx.Tx, o domain.Order) (domain.Order, error) {
	insertOrder := NewStmtBuilder().
		Insert("orders").
		Columns("customer", "total", "discount", "created_at").
		Values(o.Customer, o.Total, o.Discount, o.CreatedAt).
		Suffix("RETURNING id, customer, total, discount, created_at")

	sql, args, err := insertOrder.ToSql()
	if err != nil {
		return domain.Order{}, fmt.Errorf("order insert sql: %w", err)
	}

	var stored domain.Order
	if err := tx.QueryRow(ctx, sql, args...).
		Scan(&stored.ID, &stored.Customer, &stored.Total, &stored.Discount, &stored.CreatedAt); err != nil {
		return domain.Order{}, fmt.Errorf("order insert scan: %w", err)
	}

	if len(o.Items) > 0 {
		itemsBuilder := NewStmtBuilder().
			Insert("order_items").
			Columns("order_id", "name", "price")
		for _, item := range o.Items {
			itemsBuilder = itemsBuilder.Values(stored.ID, item.Name, item.Price)
		}
		sql, args, err := itemsBuilder.ToSql()
		if err != nil {
			return domain.Order{}, fmt.Errorf("order items err sql: %w", err)
		}
		if _, err := tx.Exec(ctx, sql, args...); err != nil {
			return domain.Order{}, fmt.Errorf("order items err exec: %w", err)
		}
	}

	if len(o.Items) > 0 {
		stored.Items = make([]domainorder.Item, len(o.Items))
		copy(stored.Items, o.Items)
	}
	return stored, nil
}

func (r OrderRepo) fetchOrders(ctx context.Context, ids []int64) ([]domain.Order, error) {
	return r.fetchOrdersWithQuerier(ctx, r.Pool, ids)
}

func (r OrderRepo) fetchOrderTx(ctx context.Context, tx pgx.Tx, id int64) (domain.Order, error) {
	orders, err := r.fetchOrdersWithQuerier(ctx, tx, []int64{id})

	if err != nil {
		return domain.Order{}, err
	}

	if len(orders) == 0 {
		return domain.Order{}, domain.ErrNotFound
	}

	return orders[0], nil
}

type querier interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func (r OrderRepo) fetchOrdersWithQuerier(ctx context.Context, q querier, ids []int64) ([]domain.Order, error) {
	builder := NewStmtBuilder().
		Select("id", "customer", "total", "discount", "created_at").
		From("orders")
	if len(ids) > 0 {
		builder = builder.Where(squirrel.Eq{"id": ids})
	}

	builder = builder.OrderBy("id")

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("order fetch err sql: %w", err)
	}

	rows, err := q.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("order fetch err query: %w", err)
	}
	defer rows.Close()

	orders := make([]domain.Order, 0)
	orderIndex := make(map[int64]int)

	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.Customer, &o.Total, &o.Discount, &o.CreatedAt); err != nil {
			return nil, fmt.Errorf("order fetch scan err: %w", err)
		}
		orders = append(orders, o)
		orderIndex[o.ID] = len(orders) - 1
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("order fetch err rows: %w", err)
	}

	if len(orders) == 0 {
		return orders, nil
	}

	if err := r.attachOrderItems(ctx, q, orderIndex, orders); err != nil {
		return nil, err
	}

	return orders, nil
}

func (r OrderRepo) attachOrderItems(ctx context.Context, q querier, index map[int64]int, orders []domain.Order) error {
	ids := make([]int64, 0, len(index))
	for id := range index {
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil
	}

	builder := NewStmtBuilder().
		Select("order_id", "name", "price").
		From("order_items").
		Where(squirrel.Eq{"order_id": ids})

	sql, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("order items err sql: %w", err)
	}

	rows, err := q.Query(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("order items err query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var orderID int64
		var item domainorder.Item
		if err := rows.Scan(&orderID, &item.Name, &item.Price); err != nil {
			return fmt.Errorf("order items err scan: %w", err)
		}
		if idx, ok := index[orderID]; ok {
			orders[idx].Items = append(orders[idx].Items, item)
		}
	}
	return rows.Err()
}
