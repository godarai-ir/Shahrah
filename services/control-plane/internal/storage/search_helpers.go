package storage

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"
)

type pagedSearcher interface {
	QueryRow(context.Context, string, ...any) pgx.Row
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func normalizePagedQuery(query string, limit, offset int) (string, string, int, int) {
	query = strings.TrimSpace(strings.ToLower(query))

	if limit <= 0 {
		limit = 50
	}

	if limit > 200 {
		limit = 200
	}

	if offset < 0 {
		offset = 0
	}

	return query, "%" + query + "%", limit, offset
}

func listSearchResults[T any](
	ctx context.Context,
	db pagedSearcher,
	countSQL string,
	countArgs []any,
	listSQL string,
	listArgs []any,
	limit int,
	scan func(pgx.Rows) (*T, error),
) ([]T, int, error) {
	var total int
	if err := db.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := db.Query(ctx, listSQL, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]T, 0, limit)

	for rows.Next() {
		item, scanErr := scan(rows)
		if scanErr != nil {
			return nil, 0, scanErr
		}

		items = append(items, *item)
	}

	return items, total, rows.Err()
}
