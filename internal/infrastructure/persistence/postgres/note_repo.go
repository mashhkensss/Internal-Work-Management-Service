package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"

	"internal-work-management-service/internal/domain"
)

// NoteRepo stores notes in PostgreSQL
type NoteRepo struct {
	Pool ConnPool
}

func NewNoteRepo(pool ConnPool) *NoteRepo {
	return &NoteRepo{Pool: pool}
}

// Insert inserts note
func (r NoteRepo) Insert(ctx context.Context, n domain.Note) (domain.Note, error) {
	query := NewStmtBuilder().
		Insert("notes").
		Columns("title", "content", "created_at").
		Values(n.Title, n.Content, n.CreatedAt).
		Suffix("RETURNING id, title, content, created_at")

	sql, args, err := query.ToSql()
	if err != nil {
		return domain.Note{}, fmt.Errorf("note insert sql: %w", err)
	}

	row := r.Pool.QueryRow(ctx, sql, args...)
	var stored domain.Note
	if err := row.Scan(&stored.ID, &stored.Title, &stored.Content, &stored.CreatedAt); err != nil {
		return domain.Note{}, fmt.Errorf("note insert err: %w", err)
	}

	return stored, nil
}

// List returns all notes ordered by id
func (r NoteRepo) List(ctx context.Context) ([]domain.Note, error) {
	query := NewStmtBuilder().
		Select("id", "title", "content", "created_at").
		From("notes").
		OrderBy("id")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("note list sql: %w", err)
	}

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("note list query: %w", err)
	}
	defer rows.Close()

	notes := make([]domain.Note, 0)
	for rows.Next() {
		var n domain.Note
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("note list err: %w", err)
		}
		notes = append(notes, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("note list rows: %w", err)
	}

	return notes, nil
}

// Get retrieves a single note by id
func (r NoteRepo) Get(ctx context.Context, id int64) (domain.Note, error) {
	query := NewStmtBuilder().
		Select("id", "title", "content", "created_at").
		From("notes").
		Where(squirrel.Eq{"id": id})

	sql, args, err := query.ToSql()
	if err != nil {
		return domain.Note{}, fmt.Errorf("note get sql: %w", err)
	}

	row := r.Pool.QueryRow(ctx, sql, args...)
	var n domain.Note
	if err := row.Scan(&n.ID, &n.Title, &n.Content, &n.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Note{}, domain.ErrNotFound
		}
		return domain.Note{}, fmt.Errorf("note get err: %w", err)
	}

	return n, nil
}

// Delete removes a note if it presents
func (r NoteRepo) Delete(ctx context.Context, id int64) error {
	query := NewStmtBuilder().
		Delete("notes").
		Where(squirrel.Eq{"id": id})

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("note delete err sql: %w", err)
	}

	cmd, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("note delete err exec: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
