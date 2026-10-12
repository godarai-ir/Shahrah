package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apperrors"
	"github.com/jackc/pgx/v5"
)

// applyOnBehalfOfCapability marks a VAT that may act for tenant members and
// records its product label. Only VATs can carry the capability.
func applyOnBehalfOfCapability(auth *platform.AuthContext, enabled bool, product string) {
	if !enabled || auth.KeyType != platform.KeyTypeVAT {
		return
	}

	auth.OnBehalfOfEnabled = true
	auth.CreditProduct = strings.TrimSpace(product)

	if auth.CreditProduct == "" {
		auth.CreditProduct = platform.CreditProductLibreChat
	}
}

// SetVATOnBehalfOf enables or disables the on-behalf-of capability of a VAT.
func (s *AuthRepository) SetVATOnBehalfOf(
	ctx context.Context,
	tenantID string,
	virtualUserID string,
	keyID string,
	req platform.UpdateVATOnBehalfOfRequest,
	actor string,
) (*platform.VATOnBehalfOf, error) {
	product := strings.TrimSpace(req.Product)
	if !req.Enabled {
		product = ""
	}

	var (
		out       platform.VATOnBehalfOf
		updatedAt time.Time
	)

	err := s.pool.QueryRow(ctx, `
		UPDATE verification_tokens
		SET on_behalf_of_enabled = $4,
		    on_behalf_of_product = NULLIF($5, ''),
		    on_behalf_of_updated_by = NULLIF($6, ''),
		    on_behalf_of_updated_at = NOW(),
		    updated_at = NOW()
		WHERE id = $1::uuid
		  AND tenant_id = $2::uuid
		  AND virtual_user_id = $3::uuid
		  AND key_type = 'vat'
		  AND revoked_at IS NULL
		RETURNING on_behalf_of_enabled, COALESCE(on_behalf_of_product, ''),
		          COALESCE(on_behalf_of_updated_by, ''), on_behalf_of_updated_at
	`, keyID, tenantID, virtualUserID, req.Enabled, product, strings.TrimSpace(actor)).Scan(
		&out.Enabled, &out.Product, &out.UpdatedBy, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}

		return nil, fmt.Errorf("update vat on-behalf-of: %w", err)
	}

	out.UpdatedAt = &updatedAt

	s.invalidateTenantRuntimeProfile(ctx, tenantID, "vat.on_behalf_of.updated")

	return &out, nil
}

// attachVATOnBehalfOf loads the on-behalf-of state for listed VATs.
func (s *AuthRepository) attachVATOnBehalfOf(ctx context.Context, tenantID string, items []platform.CreatedKey) error {
	ids := make([]string, 0, len(items))
	for i := range items {
		ids = append(ids, items[i].ID)
	}

	if len(ids) == 0 {
		return nil
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id::text, on_behalf_of_enabled, COALESCE(on_behalf_of_product, ''),
		       COALESCE(on_behalf_of_updated_by, ''), on_behalf_of_updated_at
		FROM verification_tokens
		WHERE tenant_id = $1::uuid AND id::text = ANY($2)
	`, tenantID, ids)
	if err != nil {
		return fmt.Errorf("load vat on-behalf-of: %w", err)
	}
	defer rows.Close()

	states := make(map[string]*platform.VATOnBehalfOf, len(ids))

	for rows.Next() {
		var (
			id    string
			state platform.VATOnBehalfOf
		)

		if err := rows.Scan(&id, &state.Enabled, &state.Product, &state.UpdatedBy, &state.UpdatedAt); err != nil {
			return fmt.Errorf("scan vat on-behalf-of: %w", err)
		}

		states[id] = &state
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate vat on-behalf-of: %w", err)
	}

	for i := range items {
		if state, ok := states[items[i].ID]; ok {
			items[i].OnBehalfOf = state
		}
	}

	return nil
}

// ResolveOnBehalfOfUserAuth builds the member side of an on-behalf-of request:
// the same user-scoped model policy as the member's own PAT, plus team and org
// unit attribution when the member belongs to exactly one team.
func (s *AuthRepository) ResolveOnBehalfOfUserAuth(ctx context.Context, tenantID, userID string) (*platform.AuthContext, error) {
	member, err := s.ResolveTenantUserAuth(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}

	auth := platform.AuthContext{
		TenantID:    member.TenantID,
		UserID:      member.UserID,
		Role:        member.Role,
		Permissions: append([]string(nil), member.Permissions...),
	}

	if policyErr := s.applyRuntimeModelPolicy(ctx, &auth, nil); policyErr != nil {
		return nil, policyErr
	}

	teamID, orgUnitID, err := s.soleTeamMembership(ctx, auth.TenantID, auth.UserID)
	if err != nil {
		return nil, err
	}

	auth.TeamID = teamID
	auth.OrgUnitID = orgUnitID
	auth.OrganizationID = orgUnitID

	return &auth, nil
}

// soleTeamMembership returns the member's team and its org unit when the member
// belongs to exactly one team, and empty values otherwise.
func (s *AuthRepository) soleTeamMembership(ctx context.Context, tenantID, userID string) (string, string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT t.team_id::text, COALESCE(t.org_unit_id::text, '')
		FROM team_memberships tm
		JOIN teams t ON t.team_id = tm.team_id AND t.tenant_id = tm.tenant_id
		WHERE tm.tenant_id = $1::uuid AND tm.user_id = $2::uuid
		LIMIT 2
	`, tenantID, userID)
	if err != nil {
		return "", "", fmt.Errorf("load team memberships: %w", err)
	}
	defer rows.Close()

	var teams [][2]string

	for rows.Next() {
		var teamID, orgUnitID string
		if err := rows.Scan(&teamID, &orgUnitID); err != nil {
			return "", "", fmt.Errorf("scan team membership: %w", err)
		}

		teams = append(teams, [2]string{teamID, orgUnitID})
	}

	if err := rows.Err(); err != nil {
		return "", "", fmt.Errorf("iterate team memberships: %w", err)
	}

	if len(teams) != 1 {
		return "", "", nil
	}

	return teams[0][0], teams[0][1], nil
}
