package storage

import (
	"context"
	"errors"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrVirtualModelNotFound = errors.New("virtual model not found")
)

func (s *Store) CreateVirtualModel(ctx context.Context, tenantID string, req platform.CreateVirtualModelRequest, actor string) (*platform.VirtualModel, error) {
	id := uuid.NewString()
	alias := strings.TrimSpace(req.Alias)
	strategy := strings.TrimSpace(strings.ToLower(req.RoutingStrategy))
	if strategy == "" {
		strategy = "priority"
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		INSERT INTO virtual_models (virtual_model_id, tenant_id, alias, routing_strategy, enabled, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, id, tenantID, alias, strategy, enabled, emptyToNil(actor), emptyToNil(actor)); err != nil {
		return nil, err
	}

	targets, err := insertVirtualModelTargets(ctx, tx, id, req.Targets)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	out := &platform.VirtualModel{
		ID:              id,
		TenantID:        tenantID,
		Alias:           alias,
		RoutingStrategy: strategy,
		Enabled:         enabled,
		Version:         1,
		Targets:         targets,
		CreatedBy:       actor,
		UpdatedBy:       actor,
	}
	return out, nil
}

func (s *Store) ListVirtualModels(ctx context.Context, tenantID string) ([]platform.VirtualModel, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT virtual_model_id::text, tenant_id::text, alias, routing_strategy, enabled, version,
		       COALESCE(created_by, ''), COALESCE(updated_by, ''), created_at, updated_at
		FROM virtual_models
		WHERE tenant_id = $1
		ORDER BY created_at ASC
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]platform.VirtualModel, 0)
	for rows.Next() {
		var vm platform.VirtualModel
		if err := rows.Scan(&vm.ID, &vm.TenantID, &vm.Alias, &vm.RoutingStrategy, &vm.Enabled, &vm.Version, &vm.CreatedBy, &vm.UpdatedBy, &vm.CreatedAt, &vm.UpdatedAt); err != nil {
			return nil, err
		}
		targets, err := s.listVirtualModelTargets(ctx, vm.ID)
		if err != nil {
			return nil, err
		}
		vm.Targets = targets
		out = append(out, vm)
	}
	return out, rows.Err()
}

func (s *Store) GetVirtualModel(ctx context.Context, tenantID string, id string) (*platform.VirtualModel, error) {
	var vm platform.VirtualModel
	err := s.pool.QueryRow(ctx, `
		SELECT virtual_model_id::text, tenant_id::text, alias, routing_strategy, enabled, version,
		       COALESCE(created_by, ''), COALESCE(updated_by, ''), created_at, updated_at
		FROM virtual_models WHERE tenant_id = $1 AND virtual_model_id = $2
	`, tenantID, id).Scan(&vm.ID, &vm.TenantID, &vm.Alias, &vm.RoutingStrategy, &vm.Enabled, &vm.Version, &vm.CreatedBy, &vm.UpdatedBy, &vm.CreatedAt, &vm.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrVirtualModelNotFound
	}
	if err != nil {
		return nil, err
	}
	targets, err := s.listVirtualModelTargets(ctx, vm.ID)
	if err != nil {
		return nil, err
	}
	vm.Targets = targets
	return &vm, nil
}

func (s *Store) GetVirtualModelByAlias(ctx context.Context, tenantID string, alias string) (*platform.VirtualModel, error) {
	var vm platform.VirtualModel
	err := s.pool.QueryRow(ctx, `
		SELECT virtual_model_id::text, tenant_id::text, alias, routing_strategy, enabled, version,
		       COALESCE(created_by, ''), COALESCE(updated_by, ''), created_at, updated_at
		FROM virtual_models WHERE tenant_id = $1 AND alias = $2 AND enabled = TRUE
	`, tenantID, strings.TrimSpace(alias)).Scan(&vm.ID, &vm.TenantID, &vm.Alias, &vm.RoutingStrategy, &vm.Enabled, &vm.Version, &vm.CreatedBy, &vm.UpdatedBy, &vm.CreatedAt, &vm.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrVirtualModelNotFound
	}
	if err != nil {
		return nil, err
	}
	targets, err := s.listVirtualModelTargets(ctx, vm.ID)
	if err != nil {
		return nil, err
	}
	vm.Targets = targets
	return &vm, nil
}

func (s *Store) UpdateVirtualModel(ctx context.Context, tenantID string, id string, req platform.UpdateVirtualModelRequest, actor string) (*platform.VirtualModel, error) {
	curr, err := s.GetVirtualModel(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	alias := curr.Alias
	if req.Alias != nil {
		alias = strings.TrimSpace(*req.Alias)
	}
	strategy := curr.RoutingStrategy
	if req.RoutingStrategy != nil {
		strategy = strings.TrimSpace(strings.ToLower(*req.RoutingStrategy))
	}
	enabled := curr.Enabled
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		UPDATE virtual_models
		SET alias = $3, routing_strategy = $4, enabled = $5, version = version + 1, updated_by = $6, updated_at = NOW()
		WHERE tenant_id = $1 AND virtual_model_id = $2
	`, tenantID, id, alias, strategy, enabled, emptyToNil(actor)); err != nil {
		return nil, err
	}

	if len(req.Targets) > 0 {
		if _, err := tx.Exec(ctx, `DELETE FROM virtual_model_targets WHERE virtual_model_id = $1`, id); err != nil {
			return nil, err
		}
		if _, err := insertVirtualModelTargets(ctx, tx, id, req.Targets); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.GetVirtualModel(ctx, tenantID, id)
}

func (s *Store) DeleteVirtualModel(ctx context.Context, tenantID string, id string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM virtual_model_targets WHERE virtual_model_id = $1`, id); err != nil {
		return err
	}

	res, err := tx.Exec(ctx, `DELETE FROM virtual_models WHERE tenant_id = $1 AND virtual_model_id = $2`, tenantID, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrVirtualModelNotFound
	}

	return tx.Commit(ctx)
}

func insertVirtualModelTargets(ctx context.Context, tx pgx.Tx, vmID string, targets []platform.VirtualModelTarget) ([]platform.VirtualModelTarget, error) {
	out := make([]platform.VirtualModelTarget, 0, len(targets))
	for i, target := range targets {
		tid := target.ID
		if tid == "" {
			tid = uuid.NewString()
		}
		order := target.Order
		if order <= 0 {
			order = i + 1
		}
		retryAttempts := target.Policy.Retry.Attempts
		if retryAttempts <= 0 {
			retryAttempts = 2
		}
		retryDelay := target.Policy.Retry.DelayMS
		if retryDelay < 0 {
			retryDelay = 0
		}
		retryCodes := target.Policy.Retry.StatusCodes
		if len(retryCodes) == 0 {
			retryCodes = []int{429, 500, 502, 503}
		}
		fallbackCodes := target.Policy.Fallback.StatusCodes
		if len(fallbackCodes) == 0 {
			fallbackCodes = []int{429, 500, 502, 503, 504}
		}
		fallbackCandidate := target.Policy.Fallback.Candidate
		if _, err := tx.Exec(ctx, `
			INSERT INTO virtual_model_targets (
				target_id, virtual_model_id, target_order, provider, model, timeout_sec,
				retry_attempts, retry_delay_ms, retry_status_codes, fallback_status_codes, fallback_candidate
			)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		`, tid, vmID, order, strings.ToLower(strings.TrimSpace(target.Provider)), strings.TrimSpace(target.Model), target.Policy.TimeoutSec, retryAttempts, retryDelay, retryCodes, fallbackCodes, fallbackCandidate); err != nil {
			return nil, err
		}
		target.ID = tid
		target.Order = order
		target.Policy.Retry.Attempts = retryAttempts
		target.Policy.Retry.DelayMS = retryDelay
		target.Policy.Retry.StatusCodes = retryCodes
		target.Policy.Fallback.StatusCodes = fallbackCodes
		target.Policy.Fallback.Candidate = fallbackCandidate
		out = append(out, target)
	}
	return out, nil
}

func (s *Store) listVirtualModelTargets(ctx context.Context, vmID string) ([]platform.VirtualModelTarget, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT target_id::text, target_order, provider, model, COALESCE(timeout_sec, 0),
		       retry_attempts, retry_delay_ms, retry_status_codes, fallback_status_codes, fallback_candidate
		FROM virtual_model_targets
		WHERE virtual_model_id = $1
		ORDER BY target_order ASC
	`, vmID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]platform.VirtualModelTarget, 0)
	for rows.Next() {
		var t platform.VirtualModelTarget
		if err := rows.Scan(
			&t.ID,
			&t.Order,
			&t.Provider,
			&t.Model,
			&t.Policy.TimeoutSec,
			&t.Policy.Retry.Attempts,
			&t.Policy.Retry.DelayMS,
			&t.Policy.Retry.StatusCodes,
			&t.Policy.Fallback.StatusCodes,
			&t.Policy.Fallback.Candidate,
		); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
