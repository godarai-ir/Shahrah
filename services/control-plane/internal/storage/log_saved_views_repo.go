package storage

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) ListSavedLogViewsV2(ctx context.Context, tenantID string, userID string) ([]platform.SavedLogViewV2, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT saved_view_id, tenant_id, name, coalesce(description, ''), filters_json, visible_columns_json,
		       shared_scope, coalesce(created_by_user_id::text, ''), created_at, updated_at
		FROM tenant_log_saved_views
		WHERE tenant_id = $1 AND (shared_scope = 'tenant' OR created_by_user_id::text = $2)
		ORDER BY updated_at DESC, name ASC
	`, tenantID, strings.TrimSpace(userID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]platform.SavedLogViewV2, 0, 8)

	for rows.Next() {
		item, err := scanSavedLogView(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, *item)
	}

	return items, rows.Err()
}

func (s *Store) CreateSavedLogViewV2(ctx context.Context, tenantID string, userID string, req platform.CreateSavedLogViewV2Request) (*platform.SavedLogViewV2, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, pgx.ErrNoRows
	}

	scope := normalizeSavedLogViewScope(req.SharedScope)
	filtersJSON := marshalJSONBestEffort(req.Filters)
	columnsJSON := marshalJSONBestEffort(req.VisibleColumns)

	var createdBy interface{}
	if strings.TrimSpace(userID) != "" {
		createdBy = strings.TrimSpace(userID)
	}

	row := s.pool.QueryRow(ctx, `
		INSERT INTO tenant_log_saved_views (
			saved_view_id, tenant_id, name, description, filters_json, visible_columns_json,
			shared_scope, created_by_user_id
		)
		VALUES ($1, $2, $3, nullif($4, ''), $5, $6, $7, $8)
		RETURNING saved_view_id, tenant_id, name, coalesce(description, ''), filters_json, visible_columns_json,
		          shared_scope, coalesce(created_by_user_id::text, ''), created_at, updated_at
	`, uuid.NewString(), tenantID, name, strings.TrimSpace(req.Description), filtersJSON, columnsJSON, scope, createdBy)

	return scanSavedLogView(row)
}

func (s *Store) UpdateSavedLogViewV2(ctx context.Context, tenantID string, savedViewID string, req platform.UpdateSavedLogViewV2Request) (*platform.SavedLogViewV2, error) {
	current, err := s.getSavedLogViewV2(ctx, tenantID, savedViewID)
	if err != nil {
		return nil, err
	}

	name := current.Name
	if req.Name != nil {
		name = strings.TrimSpace(*req.Name)
	}

	description := current.Description
	if req.Description != nil {
		description = strings.TrimSpace(*req.Description)
	}

	filters := current.Filters
	if req.Filters != nil {
		filters = req.Filters
	}

	columns := current.VisibleColumns
	if req.VisibleColumns != nil {
		columns = req.VisibleColumns
	}

	scope := current.SharedScope
	if req.SharedScope != nil {
		scope = normalizeSavedLogViewScope(*req.SharedScope)
	}

	row := s.pool.QueryRow(ctx, `
		UPDATE tenant_log_saved_views
		SET name = $3,
		    description = nullif($4, ''),
		    filters_json = $5,
		    visible_columns_json = $6,
		    shared_scope = $7,
		    updated_at = now()
		WHERE tenant_id = $1 AND saved_view_id = $2
		RETURNING saved_view_id, tenant_id, name, coalesce(description, ''), filters_json, visible_columns_json,
		          shared_scope, coalesce(created_by_user_id::text, ''), created_at, updated_at
	`, tenantID, strings.TrimSpace(savedViewID), name, description, marshalJSONBestEffort(filters), marshalJSONBestEffort(columns), scope)

	return scanSavedLogView(row)
}

func (s *Store) DeleteSavedLogViewV2(ctx context.Context, tenantID string, savedViewID string) error {
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM tenant_log_saved_views
		WHERE tenant_id = $1 AND saved_view_id = $2
	`, tenantID, strings.TrimSpace(savedViewID))
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (s *Store) ListLogExportsV2(ctx context.Context, tenantID string, limit int) ([]platform.LogExportJobV2, error) {
	if limit <= 0 || limit > 100 {
		limit = 25
	}

	rows, err := s.pool.Query(ctx, `
		SELECT export_id, tenant_id, filters_json, format, include_content, status, row_count,
		       coalesce(created_by_user_id::text, ''), created_at, completed_at
		FROM tenant_log_export_jobs
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]platform.LogExportJobV2, 0, limit)

	for rows.Next() {
		item, err := scanLogExportJob(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, *item)
	}

	return items, rows.Err()
}

func (s *Store) CreateLogExportV2(ctx context.Context, tenantID string, userID string, req platform.CreateLogExportV2Request) (*platform.LogExportJobV2, error) {
	format := strings.TrimSpace(strings.ToLower(req.Format))
	if format != "jsonl" && format != "csv" {
		format = "csv"
	}

	var createdBy interface{}
	if strings.TrimSpace(userID) != "" {
		createdBy = strings.TrimSpace(userID)
	}

	now := time.Now().UTC()

	row := s.pool.QueryRow(ctx, `
		INSERT INTO tenant_log_export_jobs (
			export_id, tenant_id, filters_json, format, include_content, status,
			row_count, created_by_user_id, completed_at
		)
		VALUES ($1, $2, $3, $4, $5, 'completed', 0, $6, $7)
		RETURNING export_id, tenant_id, filters_json, format, include_content, status, row_count,
		          coalesce(created_by_user_id::text, ''), created_at, completed_at
	`, uuid.NewString(), tenantID, marshalJSONBestEffort(req.Filters), format, req.IncludeContent, createdBy, now)

	return scanLogExportJob(row)
}

func (s *Store) GetLogExportV2(ctx context.Context, tenantID string, exportID string) (*platform.LogExportJobV2, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT export_id, tenant_id, filters_json, format, include_content, status, row_count,
		       coalesce(created_by_user_id::text, ''), created_at, completed_at
		FROM tenant_log_export_jobs
		WHERE tenant_id = $1 AND export_id = $2
	`, tenantID, strings.TrimSpace(exportID))

	return scanLogExportJob(row)
}

func (s *Store) CancelLogExportV2(ctx context.Context, tenantID string, exportID string) (*platform.LogExportJobV2, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE tenant_log_export_jobs
		SET status = CASE WHEN status = 'completed' THEN status ELSE 'canceled' END
		WHERE tenant_id = $1 AND export_id = $2
		RETURNING export_id, tenant_id, filters_json, format, include_content, status, row_count,
		          coalesce(created_by_user_id::text, ''), created_at, completed_at
	`, tenantID, strings.TrimSpace(exportID))

	return scanLogExportJob(row)
}

func (s *Store) getSavedLogViewV2(ctx context.Context, tenantID string, savedViewID string) (*platform.SavedLogViewV2, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT saved_view_id, tenant_id, name, coalesce(description, ''), filters_json, visible_columns_json,
		       shared_scope, coalesce(created_by_user_id::text, ''), created_at, updated_at
		FROM tenant_log_saved_views
		WHERE tenant_id = $1 AND saved_view_id = $2
	`, tenantID, strings.TrimSpace(savedViewID))

	return scanSavedLogView(row)
}

type logRowScanner interface {
	Scan(dest ...interface{}) error
}

func scanSavedLogView(row logRowScanner) (*platform.SavedLogViewV2, error) {
	var (
		item        platform.SavedLogViewV2
		filtersJSON []byte
		columnsJSON []byte
	)
	if err := row.Scan(
		&item.SavedViewID, &item.TenantID, &item.Name, &item.Description, &filtersJSON, &columnsJSON,
		&item.SharedScope, &item.CreatedByUserID, &item.CreatedAt, &item.UpdatedAt,
	); err != nil {
		return nil, err
	}

	if err := decodeJSONBytes(filtersJSON, &item.Filters); err != nil {
		return nil, err
	}

	if err := decodeJSONBytes(columnsJSON, &item.VisibleColumns); err != nil {
		return nil, err
	}

	if item.Filters == nil {
		item.Filters = map[string]interface{}{}
	}

	return &item, nil
}

func scanLogExportJob(row logRowScanner) (*platform.LogExportJobV2, error) {
	var (
		item        platform.LogExportJobV2
		filtersJSON []byte
	)
	if err := row.Scan(
		&item.ExportID, &item.TenantID, &filtersJSON, &item.Format, &item.IncludeContent, &item.Status,
		&item.RowCount, &item.CreatedByUserID, &item.CreatedAt, &item.CompletedAt,
	); err != nil {
		return nil, err
	}

	if err := decodeJSONBytes(filtersJSON, &item.Filters); err != nil {
		return nil, err
	}

	if item.Filters == nil {
		item.Filters = map[string]interface{}{}
	}

	return &item, nil
}

func decodeJSONBytes[T any](raw []byte, target *T) error {
	if len(raw) == 0 {
		return nil
	}

	return json.Unmarshal(raw, target)
}

func normalizeSavedLogViewScope(scope string) string {
	if strings.EqualFold(strings.TrimSpace(scope), "private") {
		return "private"
	}

	return "tenant"
}
