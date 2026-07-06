package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/modelid"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrVirtualModelNotFound     = errors.New("virtual model not found")
	ErrInvalidVirtualModelModes = errors.New("invalid virtual model modes")
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

	exposedModes, err := s.resolveVirtualModelExposedModes(ctx, tenantID, req.Targets, req.ExposedModes)
	if err != nil {
		return nil, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	_, insertErr := tx.Exec(ctx, `
		INSERT INTO virtual_models (virtual_model_id, tenant_id, alias, routing_strategy, enabled, exposed_modes, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, id, tenantID, alias, strategy, enabled, exposedModes, emptyToNil(actor), emptyToNil(actor))
	if insertErr != nil {
		return nil, insertErr
	}

	targets, err := insertVirtualModelTargets(ctx, tx, id, req.Targets)
	if err != nil {
		return nil, err
	}

	access, err := upsertVirtualModelAccess(ctx, tx, tenantID, id, req.Access)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	s.invalidateVirtualModelsCache(ctx, tenantID)
	s.invalidateTenantRuntimeProfile(ctx, tenantID, "virtual_model.created")
	s.warmVirtualModelsCacheAsync(tenantID)

	out := &platform.VirtualModel{
		ID:              id,
		TenantID:        tenantID,
		Alias:           alias,
		RoutingStrategy: strategy,
		Enabled:         enabled,
		ExposedModes:    exposedModes,
		Version:         1,
		Targets:         targets,
		Access:          access,
		CreatedBy:       actor,
		UpdatedBy:       actor,
	}

	return out, nil
}

func (s *Store) ListVirtualModels(ctx context.Context, tenantID string) ([]platform.VirtualModel, error) {
	cacheKey := virtualModelsCacheKey(tenantID)
	if models, state := s.getCachedVirtualModels(cacheKey); state == hotPathCacheFresh {
		return models, nil
	} else if state == hotPathCacheStale {
		s.refreshVirtualModelsCacheAsync(tenantID, cacheKey)
		return models, nil
	}

	value, err, _ := s.hotPathLoadGroup.Do(cacheKey, func() (interface{}, error) {
		if models, state := s.getCachedVirtualModels(cacheKey); state == hotPathCacheFresh {
			return models, nil
		} else if state == hotPathCacheStale {
			s.refreshVirtualModelsCacheAsync(tenantID, cacheKey)
			return models, nil
		}

		models, err := s.listVirtualModelsFromDB(ctx, tenantID)
		if err != nil {
			return nil, err
		}
		s.cacheVirtualModels(cacheKey, models)
		return cloneVirtualModels(models), nil
	})
	if err != nil {
		return nil, err
	}
	models, ok := value.([]platform.VirtualModel)
	if !ok {
		return nil, fmt.Errorf("unexpected virtual model cache value %T", value)
	}

	return cloneVirtualModels(models), nil
}

func (s *Store) listVirtualModelsFromDB(ctx context.Context, tenantID string) ([]platform.VirtualModel, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT virtual_model_id::text, tenant_id::text, alias, routing_strategy, enabled, COALESCE(exposed_modes, '{}'), version,
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
		if err := rows.Scan(&vm.ID, &vm.TenantID, &vm.Alias, &vm.RoutingStrategy, &vm.Enabled, &vm.ExposedModes, &vm.Version, &vm.CreatedBy, &vm.UpdatedBy, &vm.CreatedAt, &vm.UpdatedAt); err != nil {
			return nil, err
		}

		targets, err := s.listVirtualModelTargets(ctx, vm.ID)
		if err != nil {
			return nil, err
		}

		access, err := s.listVirtualModelAccess(ctx, vm.ID)
		if err != nil {
			return nil, err
		}

		vm.Targets = targets
		vm.Access = access
		out = append(out, vm)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (s *Store) GetVirtualModel(ctx context.Context, tenantID string, id string) (*platform.VirtualModel, error) {
	var vm platform.VirtualModel

	err := s.pool.QueryRow(ctx, `
		SELECT virtual_model_id::text, tenant_id::text, alias, routing_strategy, enabled, COALESCE(exposed_modes, '{}'), version,
		       COALESCE(created_by, ''), COALESCE(updated_by, ''), created_at, updated_at
		FROM virtual_models WHERE tenant_id = $1 AND virtual_model_id = $2
	`, tenantID, id).Scan(&vm.ID, &vm.TenantID, &vm.Alias, &vm.RoutingStrategy, &vm.Enabled, &vm.ExposedModes, &vm.Version, &vm.CreatedBy, &vm.UpdatedBy, &vm.CreatedAt, &vm.UpdatedAt)
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

	access, err := s.listVirtualModelAccess(ctx, vm.ID)
	if err != nil {
		return nil, err
	}

	vm.Targets = targets
	vm.Access = access

	return &vm, nil
}

func (s *Store) GetVirtualModelByAlias(ctx context.Context, tenantID string, alias string) (*platform.VirtualModel, error) {
	var vm platform.VirtualModel

	err := s.pool.QueryRow(ctx, `
		SELECT virtual_model_id::text, tenant_id::text, alias, routing_strategy, enabled, COALESCE(exposed_modes, '{}'), version,
		       COALESCE(created_by, ''), COALESCE(updated_by, ''), created_at, updated_at
		FROM virtual_models WHERE tenant_id = $1 AND alias = $2 AND enabled = TRUE
	`, tenantID, strings.TrimSpace(alias)).Scan(&vm.ID, &vm.TenantID, &vm.Alias, &vm.RoutingStrategy, &vm.Enabled, &vm.ExposedModes, &vm.Version, &vm.CreatedBy, &vm.UpdatedBy, &vm.CreatedAt, &vm.UpdatedAt)
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

	access, err := s.listVirtualModelAccess(ctx, vm.ID)
	if err != nil {
		return nil, err
	}

	vm.Targets = targets
	vm.Access = access

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

	targetsForModeValidation := curr.Targets
	if len(req.Targets) > 0 {
		targetsForModeValidation = req.Targets
	}

	exposedModesInput := []string(nil)
	if req.ExposedModes != nil {
		exposedModesInput = *req.ExposedModes
	} else if len(req.Targets) == 0 && len(curr.ExposedModes) > 0 {
		exposedModesInput = curr.ExposedModes
	}

	exposedModes, err := s.resolveVirtualModelExposedModes(ctx, tenantID, targetsForModeValidation, exposedModesInput)
	if err != nil {
		return nil, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	_, updateErr := tx.Exec(ctx, `
		UPDATE virtual_models
		SET alias = $3, routing_strategy = $4, enabled = $5, exposed_modes = $6, version = version + 1, updated_by = $7, updated_at = NOW()
		WHERE tenant_id = $1 AND virtual_model_id = $2
	`, tenantID, id, alias, strategy, enabled, exposedModes, emptyToNil(actor))
	if updateErr != nil {
		return nil, updateErr
	}

	if len(req.Targets) > 0 {
		_, deleteTargetsErr := tx.Exec(ctx, `DELETE FROM virtual_model_targets WHERE virtual_model_id = $1`, id)
		if deleteTargetsErr != nil {
			return nil, deleteTargetsErr
		}

		if _, err := insertVirtualModelTargets(ctx, tx, id, req.Targets); err != nil {
			return nil, err
		}
	}

	if req.Access != nil {
		_, accessErr := upsertVirtualModelAccess(ctx, tx, tenantID, id, *req.Access)
		if accessErr != nil {
			return nil, accessErr
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	s.invalidateVirtualModelsCache(ctx, tenantID)
	s.invalidateTenantRuntimeProfile(ctx, tenantID, "virtual_model.updated")
	s.warmVirtualModelsCacheAsync(tenantID)

	return s.GetVirtualModel(ctx, tenantID, id)
}

func (s *Store) DeleteVirtualModel(ctx context.Context, tenantID string, id string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer rollbackTxBestEffort(ctx, tx)

	_, deleteTargetsErr := tx.Exec(ctx, `DELETE FROM virtual_model_targets WHERE virtual_model_id = $1`, id)
	if deleteTargetsErr != nil {
		return deleteTargetsErr
	}

	_, deleteAccessErr := tx.Exec(ctx, `DELETE FROM virtual_model_access WHERE virtual_model_id = $1`, id)
	if deleteAccessErr != nil {
		return deleteAccessErr
	}

	res, err := tx.Exec(ctx, `DELETE FROM virtual_models WHERE tenant_id = $1 AND virtual_model_id = $2`, tenantID, id)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return ErrVirtualModelNotFound
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	s.invalidateVirtualModelsCache(ctx, tenantID)
	s.invalidateTenantRuntimeProfile(ctx, tenantID, "virtual_model.deleted")
	s.warmVirtualModelsCacheAsync(tenantID)
	return nil
}

func (s *Store) ListVirtualModelsForPrincipal(ctx context.Context, tenantID string, userID string, teamID string) ([]platform.VirtualModel, error) {
	normalizedUserID := strings.TrimSpace(userID)
	normalizedTeamID := strings.TrimSpace(teamID)

	rows, err := s.pool.Query(ctx, `
		SELECT vm.virtual_model_id::text
		FROM virtual_models vm
		WHERE vm.tenant_id = $1::uuid
		  AND vm.enabled = true
		  AND EXISTS (
			SELECT 1
			FROM virtual_model_access vma
			WHERE vma.virtual_model_id = vm.virtual_model_id
			  AND vma.role = 'user'
			  AND (
				vma.principal_type = 'everyone'
				OR (vma.principal_type = 'user' AND vma.principal_id::text = NULLIF($2, ''))
				OR (vma.principal_type = 'team' AND (
					vma.principal_id::text = NULLIF($3, '')
					OR (
						NULLIF($2, '') IS NOT NULL
						AND EXISTS (
							SELECT 1
							FROM team_memberships tm
							WHERE tm.tenant_id = $1::uuid
							  AND tm.team_id::text = vma.principal_id::text
							  AND tm.user_id::text = NULLIF($2, '')
						)
					)
				))
				OR (vma.principal_type = 'org_unit' AND (
					(
						NULLIF($2, '') IS NOT NULL
						AND EXISTS (
							SELECT 1
							FROM organization_memberships om
							WHERE om.tenant_id = $1::uuid
							  AND om.organization_id::text = vma.principal_id::text
							  AND om.user_id::text = NULLIF($2, '')
						)
					)
					OR (
						NULLIF($2, '') IS NOT NULL
						AND EXISTS (
							SELECT 1
							FROM team_memberships tm
							JOIN teams t ON t.team_id = tm.team_id
							WHERE tm.tenant_id = $1::uuid
							  AND tm.user_id::text = NULLIF($2, '')
							  AND t.org_unit_id::text = vma.principal_id::text
						)
					)
					OR (
						NULLIF($3, '') IS NOT NULL
						AND EXISTS (
							SELECT 1
							FROM teams t
							WHERE t.tenant_id = $1::uuid
							  AND t.team_id::text = NULLIF($3, '')
							  AND t.org_unit_id::text = vma.principal_id::text
						)
					)
				))
			  )
		  )
		ORDER BY vm.created_at ASC
	`, tenantID, normalizedUserID, normalizedTeamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]platform.VirtualModel, 0)

	for rows.Next() {
		var vmID string
		if err := rows.Scan(&vmID); err != nil {
			return nil, err
		}

		vm, err := s.GetVirtualModel(ctx, tenantID, vmID)
		if err != nil {
			return nil, err
		}

		out = append(out, *vm)
	}

	return out, rows.Err()
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
		targetProvider := strings.ToLower(strings.TrimSpace(target.Provider))

		targetModel := normalizeModelFQN(target.Model)
		weight := target.Weight
		if weight < 0 {
			weight = 0
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO virtual_model_targets (
				target_id, virtual_model_id, target_order, weight, provider, model, timeout_sec,
				retry_attempts, retry_delay_ms, retry_status_codes, fallback_status_codes, fallback_candidate
			)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		`, tid, vmID, order, weight, targetProvider, targetModel, target.Policy.TimeoutSec, retryAttempts, retryDelay, retryCodes, fallbackCodes, fallbackCandidate); err != nil {
			return nil, err
		}

		target.ID = tid
		target.Order = order
		target.Weight = weight
		target.Provider = targetProvider
		target.Model = targetModel
		target.Policy.Retry.Attempts = retryAttempts
		target.Policy.Retry.DelayMS = retryDelay
		target.Policy.Retry.StatusCodes = retryCodes
		target.Policy.Fallback.StatusCodes = fallbackCodes
		target.Policy.Fallback.Candidate = fallbackCandidate
		out = append(out, target)
	}

	return out, nil
}

func upsertVirtualModelAccess(ctx context.Context, tx pgx.Tx, tenantID string, vmID string, access platform.TenantProviderAccountAccessControl) (platform.TenantProviderAccountAccessControl, error) {
	if _, err := tx.Exec(ctx, `DELETE FROM virtual_model_access WHERE virtual_model_id = $1`, vmID); err != nil {
		return platform.TenantProviderAccountAccessControl{}, err
	}

	normalized := normalizeVirtualModelAccess(access)

	seen := make(map[string]struct{})
	if err := insertVirtualModelAccessRows(ctx, tx, tenantID, vmID, platform.AccessRoleManager, normalized.Managers, seen); err != nil {
		return platform.TenantProviderAccountAccessControl{}, err
	}

	if err := insertVirtualModelAccessRows(ctx, tx, tenantID, vmID, platform.AccessRoleUser, normalized.Users, seen); err != nil {
		return platform.TenantProviderAccountAccessControl{}, err
	}

	return normalized, nil
}

func normalizeVirtualModelAccess(access platform.TenantProviderAccountAccessControl) platform.TenantProviderAccountAccessControl {
	normalized := platform.TenantProviderAccountAccessControl{
		Managers: append([]platform.AccessPrincipal(nil), access.Managers...),
		Users:    append([]platform.AccessPrincipal(nil), access.Users...),
	}
	if len(normalized.Users) == 0 && len(normalized.Managers) == 0 {
		normalized.Users = []platform.AccessPrincipal{{Type: platform.PrincipalTypeEveryone}}
	}

	return normalized
}

func insertVirtualModelAccessRows(ctx context.Context, tx pgx.Tx, tenantID string, vmID string, role platform.AccessRole, principals []platform.AccessPrincipal, seen map[string]struct{}) error {
	for _, principal := range principals {
		normalized, err := normalizeAccessPrincipal(ctx, tx, tenantID, role, principal)
		if err != nil {
			return err
		}

		key := string(role) + "|" + string(normalized.Type) + "|" + normalized.PrincipalID
		if _, exists := seen[key]; exists {
			continue
		}

		seen[key] = struct{}{}

		if _, err := tx.Exec(ctx, `
			INSERT INTO virtual_model_access (
				virtual_model_id, role, principal_type, principal_id, display_name
			)
			VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''))
		`, vmID, string(role), string(normalized.Type), normalized.PrincipalID, normalized.DisplayName); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) listVirtualModelAccess(ctx context.Context, vmID string) (platform.TenantProviderAccountAccessControl, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT role, principal_type, COALESCE(principal_id, ''), COALESCE(display_name, '')
		FROM virtual_model_access
		WHERE virtual_model_id = $1::uuid
		ORDER BY role, principal_type, principal_id
	`, vmID)
	if err != nil {
		return platform.TenantProviderAccountAccessControl{}, err
	}
	defer rows.Close()

	access := platform.TenantProviderAccountAccessControl{
		Managers: make([]platform.AccessPrincipal, 0),
		Users:    make([]platform.AccessPrincipal, 0),
	}

	for rows.Next() {
		var (
			role      string
			principal platform.AccessPrincipal
		)
		if err := rows.Scan(&role, &principal.Type, &principal.PrincipalID, &principal.DisplayName); err != nil {
			return platform.TenantProviderAccountAccessControl{}, err
		}

		if principal.Type == platform.PrincipalTypeEveryone {
			principal.PrincipalID = ""
		}

		switch platform.AccessRole(strings.TrimSpace(role)) {
		case platform.AccessRoleManager:
			access.Managers = append(access.Managers, principal)
		case platform.AccessRoleUser:
			access.Users = append(access.Users, principal)
		}
	}

	if err := rows.Err(); err != nil {
		return platform.TenantProviderAccountAccessControl{}, err
	}

	if len(access.Managers) == 0 && len(access.Users) == 0 {
		access.Users = []platform.AccessPrincipal{{Type: platform.PrincipalTypeEveryone}}
	}

	return access, nil
}

func (s *Store) listVirtualModelTargets(ctx context.Context, vmID string) ([]platform.VirtualModelTarget, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT target_id::text, target_order, COALESCE(weight, 0), provider, model, COALESCE(timeout_sec, 0),
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
			&t.Weight,
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

func (s *Store) resolveVirtualModelExposedModes(ctx context.Context, tenantID string, targets []platform.VirtualModelTarget, requested []string) ([]string, error) {
	intersection, err := s.virtualModelTargetModeIntersection(ctx, tenantID, targets)
	if err != nil {
		return nil, err
	}

	if len(intersection) == 0 {
		return nil, fmt.Errorf("%w: virtual model targets do not share any supported modes", ErrInvalidVirtualModelModes)
	}

	requested = normalizeStringList(requested)
	if len(requested) == 0 {
		return intersection, nil
	}

	allowed := make(map[string]struct{}, len(intersection))
	for _, mode := range intersection {
		allowed[mode] = struct{}{}
	}

	for _, mode := range requested {
		if _, ok := allowed[mode]; !ok {
			return nil, fmt.Errorf("%w: exposed_modes must be a subset of target mode intersection", ErrInvalidVirtualModelModes)
		}
	}

	return requested, nil
}

func (s *Store) virtualModelTargetModeIntersection(ctx context.Context, tenantID string, targets []platform.VirtualModelTarget) ([]string, error) {
	if len(targets) == 0 {
		return nil, fmt.Errorf("%w: at least one target is required", ErrInvalidVirtualModelModes)
	}

	var intersection []string

	for i, target := range targets {
		modes, err := s.supportedModesForVirtualModelTarget(ctx, tenantID, target)
		if err != nil {
			return nil, err
		}

		if len(modes) == 0 {
			return nil, fmt.Errorf("%w: target %q has no supported_modes", ErrInvalidVirtualModelModes, strings.TrimSpace(target.Model))
		}

		if i == 0 {
			intersection = modes
			continue
		}

		intersection = intersectStringLists(intersection, modes)
		if len(intersection) == 0 {
			return nil, nil
		}
	}

	return intersection, nil
}

func (s *Store) supportedModesForVirtualModelTarget(ctx context.Context, tenantID string, target platform.VirtualModelTarget) ([]string, error) {
	provider, accountAlias, officialModel, ok := modelid.SplitModelFQN(strings.TrimSpace(target.Model))
	if !ok {
		return nil, fmt.Errorf("%w: target.model must use provider:account_alias:model_id format", ErrInvalidVirtualModelModes)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT m.model, COALESCE(m.types, '{}')
		FROM tenant_provider_account_models m
		JOIN tenant_provider_accounts a ON a.account_id = m.account_id
		WHERE a.tenant_id = $1::uuid
		  AND lower(a.provider) = lower($2)
		  AND a.name = $3
		  AND m.enabled = true
	`, tenantID, provider, accountAlias)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			storedModel string
			modes       []string
		)
		if err := rows.Scan(&storedModel, &modes); err != nil {
			return nil, err
		}

		if modelid.MatchOfficialModel(provider, storedModel, officialModel) {
			return normalizeStringList(modes), nil
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("%w: target %q is not an enabled provider account model", ErrInvalidVirtualModelModes, target.Model)
}

func intersectStringLists(a []string, b []string) []string {
	bSet := make(map[string]struct{}, len(b))
	for _, item := range b {
		bSet[strings.TrimSpace(item)] = struct{}{}
	}

	out := make([]string, 0, len(a))

	seen := make(map[string]struct{}, len(a))
	for _, item := range a {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		if _, ok := bSet[item]; !ok {
			continue
		}

		if _, ok := seen[item]; ok {
			continue
		}

		seen[item] = struct{}{}
		out = append(out, item)
	}

	return out
}
