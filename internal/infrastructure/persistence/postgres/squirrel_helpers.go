package postgres

import "github.com/Masterminds/squirrel"

// NewStmtBuilder returns a statement builder configured with PostgreSQL placeholders
func NewStmtBuilder() squirrel.StatementBuilderType {
	return squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
}
