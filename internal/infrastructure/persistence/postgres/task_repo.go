package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"

	"internal-work-management-service/internal/domain"
	domainTask "internal-work-management-service/internal/domain/task"
	tasksvc "internal-work-management-service/internal/service/task"
)

// TaskRepo stores tasks in PostgreSQL
type TaskRepo struct {
	Pool ConnPool
}

func NewTaskRepo(pool ConnPool) *TaskRepo {
	return &TaskRepo{Pool: pool}
}

func (r TaskRepo) Insert(ctx context.Context, t domain.Task) (domain.Task, error) {
	query := NewStmtBuilder().
		Insert("tasks").
		Columns("title", "description", "done", "created_at", "completed_at").
		Values(t.Title, t.Description, t.Done, t.CreatedAt, t.CompletedAt).
		Suffix("RETURNING id, title, description, done, created_at, completed_at")

	sql, args, err := query.ToSql()
	if err != nil {
		return domain.Task{}, fmt.Errorf("task insert err sql: %w", err)
	}

	row := r.Pool.QueryRow(ctx, sql, args...)
	var stored domain.Task
	if err := row.
		Scan(&stored.ID, &stored.Title, &stored.Description, &stored.Done, &stored.CreatedAt, &stored.CompletedAt); err != nil {
		return domain.Task{}, fmt.Errorf("task insert err scan: %w", err)
	}

	return stored, nil
}

func (r TaskRepo) List(ctx context.Context, filter tasksvc.Filter) ([]domain.Task, error) {
	builder := NewStmtBuilder().
		Select("id", "title", "description", "done", "created_at", "completed_at").
		From("tasks").
		OrderBy("id")

	if filter.Done != nil {
		builder = builder.Where(squirrel.Eq{"done": *filter.Done})
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("task list err sql: %w", err)
	}

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("task list err query: %w", err)
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		var t domain.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Done, &t.CreatedAt, &t.CompletedAt); err != nil {
			return nil, fmt.Errorf("task list err scan: %w", err)
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("task list err rows: %w", err)
	}

	return tasks, nil
}

func (r TaskRepo) MarkDone(ctx context.Context, id int64) (domain.Task, error) {
	query := NewStmtBuilder().
		Update("tasks").
		Set("done", true).
		Set("completed_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": id, "done": false}).
		Suffix("RETURNING id, title, description, done, created_at, completed_at")

	sql, args, err := query.ToSql()
	if err != nil {
		return domain.Task{}, fmt.Errorf("task mark done err sql: %w", err)
	}

	row := r.Pool.QueryRow(ctx, sql, args...)
	var updated domain.Task
	if err := row.
		Scan(&updated.ID, &updated.Title, &updated.Description, &updated.Done, &updated.CreatedAt, &updated.CompletedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return r.handleMarkDone(ctx, id)
		}
		return domain.Task{}, fmt.Errorf("task mark done err scan: %w", err)
	}

	return updated, nil
}

func (r TaskRepo) handleMarkDone(ctx context.Context, id int64) (domain.Task, error) {
	builder := NewStmtBuilder().
		Select("id", "title", "description", "done", "created_at", "completed_at").
		From("tasks").
		Where(squirrel.Eq{"id": id})

	sql, args, err := builder.ToSql()
	if err != nil {
		return domain.Task{}, fmt.Errorf("task mark done err select sql: %w", err)
	}

	row := r.Pool.QueryRow(ctx, sql, args...)
	var existing domain.Task
	if err := row.
		Scan(&existing.ID, &existing.Title, &existing.Description, &existing.Done, &existing.CreatedAt, &existing.CompletedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Task{}, domain.ErrNotFound
		}
		return domain.Task{}, fmt.Errorf("task mark done select scan: %w", err)
	}

	if existing.Done {
		return existing, domainTask.ErrAlreadyDone
	}

	return existing, fmt.Errorf("task mark done: no rows updated for id %d", id)
}
