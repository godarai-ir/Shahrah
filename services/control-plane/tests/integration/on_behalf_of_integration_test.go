package integration_test

import (
	"context"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apperrors"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

type onBehalfOfFixture struct {
	tenantID       string
	orgUnitID      string
	teamID         string
	singleTeamUser string
	multiTeamUser  string
	virtualUserID  string
	vatID          string
	vatToken       string
	patID          string
}

func seedOnBehalfOfFixture(t *testing.T, pool *pgxpool.Pool) onBehalfOfFixture {
	t.Helper()

	ctx := context.Background()
	tenant := createIntegrationTenant(t, pool, "obo-"+uuid.NewString()[:8])
	f := onBehalfOfFixture{
		tenantID:       tenant.ID,
		orgUnitID:      uuid.NewString(),
		teamID:         uuid.NewString(),
		singleTeamUser: uuid.NewString(),
		multiTeamUser:  uuid.NewString(),
		virtualUserID:  uuid.NewString(),
		vatID:          uuid.NewString(),
		vatToken:       "sk-obo-" + uuid.NewString(),
		patID:          uuid.NewString(),
	}
	secondTeam := uuid.NewString()

	exec := func(query string, args ...interface{}) {
		t.Helper()

		_, err := pool.Exec(ctx, query, args...)
		require.NoError(t, err)
	}

	exec(`INSERT INTO organizations (organization_id, tenant_id, organization_alias, created_at, updated_at)
		VALUES ($1::uuid, $2::uuid, 'platform', NOW(), NOW())`, f.orgUnitID, f.tenantID)

	for _, team := range []string{f.teamID, secondTeam} {
		exec(`INSERT INTO teams (team_id, tenant_id, team_alias, organization_id, org_unit_id, workspace_type, created_at, updated_at)
			VALUES ($1::uuid, $2::uuid, $3, $4::uuid, $4::uuid, 'general', NOW(), NOW())`, team, f.tenantID, "team-"+team[:8], f.orgUnitID)
	}

	for _, user := range []string{f.singleTeamUser, f.multiTeamUser} {
		exec(`INSERT INTO users (user_id, tenant_id, user_alias, role, active, created_at, updated_at)
			VALUES ($1::uuid, $2::uuid, 'member', 'member', true, NOW(), NOW())`, user, f.tenantID)
	}

	membership := `INSERT INTO team_memberships (tenant_id, team_id, user_id, role, created_at, updated_at)
		VALUES ($1::uuid, $2::uuid, $3::uuid, 'team_member', NOW(), NOW())`
	exec(membership, f.tenantID, f.teamID, f.singleTeamUser)
	exec(membership, f.tenantID, f.teamID, f.multiTeamUser)
	exec(membership, f.tenantID, secondTeam, f.multiTeamUser)

	exec(`INSERT INTO virtual_users (virtual_user_id, tenant_id, team_id, alias, active)
		VALUES ($1::uuid, $2::uuid, $3::uuid, 'librechat', true)`, f.virtualUserID, f.tenantID, f.teamID)
	exec(`INSERT INTO verification_tokens (id, tenant_id, token_hash, key_alias, models, team_id, virtual_user_id, key_type, created_at, updated_at)
		VALUES ($1::uuid, $2::uuid, $3, 'librechat-service', '{}', $4::uuid, $5::uuid, 'vat', NOW(), NOW())`,
		f.vatID, f.tenantID, auth.HashKey(f.vatToken), f.teamID, f.virtualUserID)
	exec(`INSERT INTO verification_tokens (id, tenant_id, token_hash, key_alias, models, user_id, key_type, created_at, updated_at)
		VALUES ($1::uuid, $2::uuid, $3, 'personal', '{}', $4::uuid, 'pat', NOW(), NOW())`,
		f.patID, f.tenantID, auth.HashKey("sk-obo-pat-"+f.patID), f.singleTeamUser)

	return f
}

func TestIntegration_VATOnBehalfOfCapability(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()
	f := seedOnBehalfOfFixture(t, pool)

	key, err := store.GetKeyByHash(ctx, auth.HashKey(f.vatToken))
	require.NoError(t, err)
	require.False(t, key.OnBehalfOfEnabled, "capability is off by default")

	state, err := store.SetVATOnBehalfOf(ctx, f.tenantID, f.virtualUserID, f.vatID,
		platform.UpdateVATOnBehalfOfRequest{Enabled: true, Product: platform.CreditProductLibreChat}, f.singleTeamUser)
	require.NoError(t, err)
	require.True(t, state.Enabled)
	require.Equal(t, platform.CreditProductLibreChat, state.Product)
	require.Equal(t, f.singleTeamUser, state.UpdatedBy)
	require.NotNil(t, state.UpdatedAt)

	key, err = store.GetKeyByHash(ctx, auth.HashKey(f.vatToken))
	require.NoError(t, err)
	require.True(t, key.OnBehalfOfEnabled)
	require.Equal(t, platform.CreditProductLibreChat, key.CreditProduct)

	listed, _, err := store.ListVATTokens(ctx, f.tenantID, f.virtualUserID, "", 10, 0)
	require.NoError(t, err)
	require.Len(t, listed, 1)
	require.NotNil(t, listed[0].OnBehalfOf)
	require.True(t, listed[0].OnBehalfOf.Enabled)

	// PATs cannot carry the capability, through the repository or directly in SQL.
	_, err = store.SetVATOnBehalfOf(ctx, f.tenantID, f.virtualUserID, f.patID, platform.UpdateVATOnBehalfOfRequest{Enabled: true}, "")
	require.ErrorIs(t, err, apperrors.ErrNotFound)
	_, err = pool.Exec(ctx, `UPDATE verification_tokens SET on_behalf_of_enabled = true WHERE id = $1::uuid`, f.patID)
	require.Error(t, err, "check constraint must reject on-behalf-of on a PAT")

	state, err = store.SetVATOnBehalfOf(ctx, f.tenantID, f.virtualUserID, f.vatID, platform.UpdateVATOnBehalfOfRequest{Enabled: false, Product: "librechat"}, "")
	require.NoError(t, err)
	require.False(t, state.Enabled)
	require.Empty(t, state.Product)

	key, err = store.GetKeyByHash(ctx, auth.HashKey(f.vatToken))
	require.NoError(t, err)
	require.False(t, key.OnBehalfOfEnabled)
	require.Empty(t, key.CreditProduct)
}

func TestIntegration_ResolveOnBehalfOfUserAuth(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()
	f := seedOnBehalfOfFixture(t, pool)

	single, err := store.ResolveOnBehalfOfUserAuth(ctx, f.tenantID, f.singleTeamUser)
	require.NoError(t, err)
	require.Equal(t, f.singleTeamUser, single.UserID)
	require.Equal(t, f.teamID, single.TeamID)
	require.Equal(t, f.orgUnitID, single.OrgUnitID)

	multi, err := store.ResolveOnBehalfOfUserAuth(ctx, f.tenantID, f.multiTeamUser)
	require.NoError(t, err)
	require.Equal(t, f.multiTeamUser, multi.UserID)
	require.Empty(t, multi.TeamID, "members of several teams get no team attribution")
	require.Empty(t, multi.OrgUnitID)

	_, err = store.ResolveOnBehalfOfUserAuth(ctx, f.tenantID, uuid.NewString())
	require.Error(t, err)

	other := createIntegrationTenant(t, pool, "obo-other-"+uuid.NewString()[:8])
	_, err = store.ResolveOnBehalfOfUserAuth(ctx, other.ID, f.singleTeamUser)
	require.Error(t, err, "a member of another tenant must not resolve")
}
