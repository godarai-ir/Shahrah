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

	query := fmt.Sprintf("SELECT %s FROM %s LIMIT 0", strings.Join(columns, ", "), table)
	if _, err := s.clickhouse.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("validate clickhouse table %s: %w", table, err)
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
