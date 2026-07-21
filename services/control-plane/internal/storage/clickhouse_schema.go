package storage

import (
	"context"
	"fmt"
	"strings"
)

func (s *Store) validateClickHouseColumns(ctx context.Context, table string, columns []string) error {
	if strings.TrimSpace(table) == "" || len(columns) == 0 {
		return fmt.Errorf("invalid clickhouse schema validation target")
	}

	rows, err := s.clickhouse.QueryContext(ctx, `
		SELECT name
		FROM system.columns
		WHERE database = currentDatabase() AND table = ?
	`, table)
	if err != nil {
		return fmt.Errorf("validate clickhouse table %s: %w", table, err)
	}
	defer closeBestEffort(rows)

	found := make(map[string]struct{}, len(columns))

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return fmt.Errorf("validate clickhouse table %s: %w", table, err)
		}

		found[name] = struct{}{}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("validate clickhouse table %s: %w", table, err)
	}

	missing := make([]string, 0)

	for _, column := range columns {
		if _, ok := found[column]; !ok {
			missing = append(missing, column)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("validate clickhouse table %s: missing columns: %s", table, strings.Join(missing, ", "))
	}

	return nil
}

func (s *Store) validateClickHouseObject(ctx context.Context, name string) error {
	var count uint64
	if err := s.clickhouse.QueryRowContext(ctx, `
		SELECT count()
		FROM system.tables
		WHERE database = currentDatabase() AND name = ?
	`, name).Scan(&count); err != nil {
		return fmt.Errorf("validate clickhouse object %s: %w", name, err)
	}

	if count == 0 {
		return fmt.Errorf("validate clickhouse object %s: missing", name)
	}

	return nil
}
