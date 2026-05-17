package integration_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/admin"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	"github.com/Delpak/shahrah/services/control-plane/internal/proxy"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	controlserver "github.com/Delpak/shahrah/services/control-plane/internal/server"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	tenantmw "github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	tunneltransport "github.com/Delpak/shahrah/services/control-plane/internal/tunnel"
	silkserver "github.com/Delpak/shahrah/services/silk-road/pkg/server"
	tclientserver "github.com/Delpak/shahrah/services/tunnel-client/pkg/server"
	tserverserver "github.com/Delpak/shahrah/services/tunnel-server/pkg/server"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestIntegration_PostgresBackedOnboardingFlow(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(databaseURL) == "" {
		t.Skip("DATABASE_URL not set")
	}

	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/chat/completions", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(contractsopenai.ChatCompletionResponse{
			ID:      "chatcmpl-db",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "gpt-4o-mini",
			Choices: []contractsopenai.Choice{{
				Index: 0,
				Message: &contractsopenai.ChatMessage{
					Role:    "assistant",
					Content: "db-backed flow succeeded",
				},
				FinishReason: "stop",
			}},
			Usage: &contractsopenai.Usage{
				PromptTokens:     14,
				CompletionTokens: 7,
				TotalTokens:      21,
			},
		})
	}))
	defer provider.Close()

	app, pool := newPersistentIntegrationApp(t, logger, memCache, databaseURL)
	resetPersistentState(t, pool)

	createdKey := createVirtualKey(t, app, `{"key_alias":"db-tenant-key","models":["gpt-4o-mini"]}`)
	createModel(t, app, fmt.Sprintf(`{"model_name":"gpt-4o-mini","provider":"openai","model":"gpt-4o-mini","api_base":"%s/v1","api_key":"sk-provider","rpm":200}`, provider.URL))
	topUpWallet(t, app, `{"amount_irr":1500000}`)

	modelsResp := performJSONRequest(t, app, http.MethodGet, "http://localhost/v1/models", "Bearer "+createdKey.Key, "")
	assert.Equal(t, http.StatusOK, modelsResp.StatusCode)

	completionResp := performJSONRequest(t, app, http.MethodPost, "http://localhost/v1/chat/completions", "Bearer "+createdKey.Key, `{"model":"gpt-4o-mini","messages":[{"role":"user","content":"hello from postgres"}]}`)
	assert.Equal(t, http.StatusOK, completionResp.StatusCode)

	var completion contractsopenai.ChatCompletionResponse
	decodeJSONBody(t, completionResp, &completion)
	assert.Equal(t, "db-backed flow succeeded", completion.Choices[0].Message.Content)

	ctx := context.Background()
	var keySpend float64
	require.NoError(t, pool.QueryRow(ctx, `SELECT spend FROM verification_tokens WHERE token_hash = $1`, hashToken(createdKey.Key)).Scan(&keySpend))
	assert.Greater(t, keySpend, 0.0)

	assertCount(t, pool, `SELECT COUNT(*) FROM model_deployments`, 1)
	assertCount(t, pool, `SELECT COUNT(*) FROM wallet_transactions WHERE tx_type = 'DEPOSIT'`, 1)
	assertCount(t, pool, `SELECT COUNT(*) FROM wallet_transactions WHERE tx_type = 'CONSUMPTION'`, 1)
	assertCount(t, pool, `SELECT COUNT(*) FROM request_logs`, 1)
	assertCount(t, pool, `SELECT COUNT(*) FROM spend_logs`, 1)

	var walletBalance float64
	require.NoError(t, pool.QueryRow(ctx, `SELECT balance_irr FROM wallets LIMIT 1`).Scan(&walletBalance))
	assert.Greater(t, walletBalance, 0.0)
	assert.Less(t, walletBalance, 1500000.0)
}

func TestIntegration_StorageSchemaContract(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(databaseURL) == "" {
		t.Skip("DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	requiredColumns := map[string][]string{
		"tenants": {
			"tenant_id", "slug", "domain", "status", "created_at",
		},
		"verification_tokens": {
			"id", "tenant_id", "organization_id", "team_id", "key_alias", "key_name", "token_hash",
			"models", "spend", "expires", "blocked", "created_at", "updated_at",
		},
		"model_deployments": {
			"id", "tenant_id", "model_name", "provider", "model", "provider_config", "model_info", "rpm", "tpm", "timeout", "created_at",
		},
		"wallets": {
			"wallet_id", "tenant_id", "balance_irr", "pending_irr", "updated_at",
		},
		"wallet_transactions": {
			"transaction_id", "wallet_id", "tenant_id", "tx_type", "amount_irr", "amount_usd", "exchange_rate", "status", "description", "metadata", "created_at",
		},
		"exchange_rates": {
			"rate_id", "usd_to_irr", "source", "fetched_at",
		},
		"request_logs": {
			"request_id", "token_hash", "tenant_id", "model", "request_body", "response_body", "status", "cost", "response_time_ms",
		},
		"spend_logs": {
			"request_id", "token_hash", "tenant_id", "organization_id", "team_id", "model",
			"input_tokens", "output_tokens", "total_tokens", "cost_usd", "cost_irr",
			"response_time_ms", "status", "request_metadata", "completion_start_time", "completion_end_time",
		},
	}

	for table, cols := range requiredColumns {
		assertColumnsExist(t, pool, table, cols)
	}

	assertIndexForColumn(t, pool, "verification_tokens", "token_hash")
	assertIndexForColumn(t, pool, "wallets", "tenant_id")
	assertIndexForColumn(t, pool, "tenants", "slug")
	assertIndexForColumn(t, pool, "model_deployments", "tenant_id")
}

func TestIntegration_OnboardingPendingRequestReusedAndExpiresAfterFiveMinutes(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(databaseURL) == "" {
		t.Skip("DATABASE_URL not set")
	}

	logger, _ := zap.NewDevelopment()
	ctx := context.Background()
	store, err := storage.NewStore(ctx, databaseURL, logger)
	require.NoError(t, err)
	t.Cleanup(store.Close)

	pool, err := pgxpool.New(ctx, databaseURL)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	resetPersistentState(t, pool)

	first, err := store.CreateOnboardingOTP(ctx, storage.CreateOnboardingOTPParams{
		TenantSlug: "ttl-tenant",
		Email:      "owner@example.com",
	})
	require.NoError(t, err)
	require.NotEmpty(t, first.RequestID)
	assert.Equal(t, 300, first.ExpiresInSeconds)
	assert.False(t, first.AlreadyPending)

	reused, err := store.CreateOnboardingOTP(ctx, storage.CreateOnboardingOTPParams{
		TenantSlug: "ttl-tenant",
		Email:      "owner@example.com",
	})
	require.NoError(t, err)
	assert.Equal(t, first.RequestID, reused.RequestID)
	assert.True(t, reused.AlreadyPending)
	assert.LessOrEqual(t, reused.ExpiresInSeconds, 300)
	assert.Greater(t, reused.ExpiresInSeconds, 0)

	_, err = store.CreateOnboardingOTP(ctx, storage.CreateOnboardingOTPParams{
		TenantSlug: "ttl-tenant",
		Email:      "other@example.com",
	})
	assert.ErrorIs(t, err, storage.ErrTenantSlugTaken)

	_, err = pool.Exec(ctx, `
		UPDATE tenant_onboarding_requests
		SET expires_at = NOW() - INTERVAL '1 second'
		WHERE request_id = $1::uuid
	`, first.RequestID)
	require.NoError(t, err)

	availability, err := store.CheckTenantSlugAvailability(ctx, "ttl-tenant")
	require.NoError(t, err)
	assert.True(t, availability.Available)

	second, err := store.CreateOnboardingOTP(ctx, storage.CreateOnboardingOTPParams{
		TenantSlug: "ttl-tenant",
		Email:      "other@example.com",
	})
	require.NoError(t, err)
	assert.NotEqual(t, first.RequestID, second.RequestID)
	assert.False(t, second.AlreadyPending)

	var expiredCount int
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM tenant_onboarding_requests
		WHERE request_id = $1::uuid AND status = 'expired'
	`, first.RequestID).Scan(&expiredCount))
	assert.Equal(t, 1, expiredCount)
}

func TestIntegration_OnboardingExpiredOTPIsRejected(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(databaseURL) == "" {
		t.Skip("DATABASE_URL not set")
	}

	logger, _ := zap.NewDevelopment()
	ctx := context.Background()
	store, err := storage.NewStore(ctx, databaseURL, logger)
	require.NoError(t, err)
	t.Cleanup(store.Close)

	pool, err := pgxpool.New(ctx, databaseURL)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	resetPersistentState(t, pool)

	otp, err := store.CreateOnboardingOTP(ctx, storage.CreateOnboardingOTPParams{
		TenantSlug: "expired-otp",
		Email:      "owner@example.com",
	})
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		UPDATE tenant_onboarding_requests
		SET expires_at = NOW() - INTERVAL '1 second'
		WHERE request_id = $1::uuid
	`, otp.RequestID)
	require.NoError(t, err)

	_, err = store.VerifyOnboardingOTPAndProvision(ctx, storage.VerifyOnboardingOTPParams{
		RequestID: otp.RequestID,
		OTPCode:   otp.DebugOTP,
	})
	assert.ErrorIs(t, err, storage.ErrOnboardingExpired)
}

func TestIntegration_TenantAdminInvitations_ValidationAndDedupe(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(databaseURL) == "" {
		t.Skip("DATABASE_URL not set")
	}

	logger, _ := zap.NewDevelopment()
	ctx := context.Background()
	store, err := storage.NewStore(ctx, databaseURL, logger)
	require.NoError(t, err)
	t.Cleanup(store.Close)

	pool, err := pgxpool.New(ctx, databaseURL)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	resetPersistentState(t, pool)

	defaultTenant := &platform.Tenant{}
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT tenant_id::text, slug, domain, status, created_at
		FROM tenants
		WHERE slug = 'default'
	`).Scan(&defaultTenant.ID, &defaultTenant.Slug, &defaultTenant.Domain, &defaultTenant.Status, &defaultTenant.CreatedAt))

	otherTenant := createIntegrationTenant(t, pool, "tenant-two")

	// 1) New email can be invited.
	firstInvite, firstLink, err := store.InviteTenantAdminUser(ctx, defaultTenant.ID, defaultTenant.Slug, platform.InviteTenantAdminUserRequest{
		Email: "new-user@tenant.test",
		Role:  "member",
	})
	require.NoError(t, err)
	require.NotNil(t, firstInvite)
	assert.NotEmpty(t, firstLink)

	// 2) Pending invitation is not duplicated.
	secondInvite, secondLink, err := store.InviteTenantAdminUser(ctx, defaultTenant.ID, defaultTenant.Slug, platform.InviteTenantAdminUserRequest{
		Email: "new-user@tenant.test",
		Role:  "member",
	})
	assert.Nil(t, secondInvite)
	assert.Empty(t, secondLink)
	assert.ErrorIs(t, err, storage.ErrInvitationAlreadyPending)

	// 3) Same email in another tenant does not block invitation in current tenant.
	otherInvite, otherLink, err := store.InviteTenantAdminUser(ctx, otherTenant.ID, otherTenant.Slug, platform.InviteTenantAdminUserRequest{
		Email: "new-user@tenant.test",
		Role:  "member",
	})
	require.NoError(t, err)
	require.NotNil(t, otherInvite)
	assert.NotEmpty(t, otherLink)

	// 4) Existing active member in same tenant cannot be invited (case-insensitive check).
	_, err = pool.Exec(ctx, `
		INSERT INTO users (user_id, tenant_id, user_alias, email, role, active, created_at, updated_at)
		VALUES (gen_random_uuid(), $1::uuid, 'existing-member', 'User1@Tenant2.com', 'member', true, NOW(), NOW())
	`, defaultTenant.ID)
	require.NoError(t, err)

	activeMemberInvite, _, err := store.InviteTenantAdminUser(ctx, defaultTenant.ID, defaultTenant.Slug, platform.InviteTenantAdminUserRequest{
		Email: "user1@tenant2.com",
		Role:  "member",
	})
	assert.Nil(t, activeMemberInvite)
	assert.ErrorIs(t, err, storage.ErrTenantMemberAlreadyExists)

	// 5) Same email in another tenant still allowed.
	_, err = pool.Exec(ctx, `
		INSERT INTO users (user_id, tenant_id, user_alias, email, role, active, created_at, updated_at)
		VALUES (gen_random_uuid(), $1::uuid, 'other-tenant-member', 'cross-tenant@example.com', 'member', true, NOW(), NOW())
	`, otherTenant.ID)
	require.NoError(t, err)

	crossTenantInvite, crossTenantLink, err := store.InviteTenantAdminUser(ctx, defaultTenant.ID, defaultTenant.Slug, platform.InviteTenantAdminUserRequest{
		Email: "cross-tenant@example.com",
		Role:  "member",
	})
	require.NoError(t, err)
	require.NotNil(t, crossTenantInvite)
	assert.NotEmpty(t, crossTenantLink)

	// 6) Deactivated members do not receive invitation links.
	_, err = pool.Exec(ctx, `
		INSERT INTO users (user_id, tenant_id, user_alias, email, role, active, created_at, updated_at)
		VALUES (gen_random_uuid(), $1::uuid, 'deactivated-member', 'deactivated@tenant.test', 'deactivated', false, NOW(), NOW())
	`, defaultTenant.ID)
	require.NoError(t, err)

	deactivatedInvite, _, err := store.InviteTenantAdminUser(ctx, defaultTenant.ID, defaultTenant.Slug, platform.InviteTenantAdminUserRequest{
		Email: "deactivated@tenant.test",
		Role:  "member",
	})
	assert.Nil(t, deactivatedInvite)
	assert.ErrorIs(t, err, storage.ErrTenantMemberDeactivated)

	// 7) Expired invitation is cleaned up and recreated.
	expiringEmail := "expired@tenant.test"
	originalExpired, originalLink, err := store.InviteTenantAdminUser(ctx, defaultTenant.ID, defaultTenant.Slug, platform.InviteTenantAdminUserRequest{
		Email: expiringEmail,
		Role:  "member",
	})
	require.NoError(t, err)
	require.NotNil(t, originalExpired)
	require.NotEmpty(t, originalLink)

	_, err = pool.Exec(ctx, `
		UPDATE tenant_admin_invitations
		SET expires_at = NOW() - INTERVAL '1 second'
		WHERE invitation_id = $1::uuid
	`, originalExpired.InvitationID)
	require.NoError(t, err)

	reissued, reissuedLink, err := store.InviteTenantAdminUser(ctx, defaultTenant.ID, defaultTenant.Slug, platform.InviteTenantAdminUserRequest{
		Email: expiringEmail,
		Role:  "member",
	})
	require.NoError(t, err)
	require.NotNil(t, reissued)
	assert.NotEqual(t, originalExpired.InvitationID, reissued.InvitationID)
	assert.NotEqual(t, originalLink, reissuedLink)
}

func TestIntegration_InviteTenantUser_InvalidEmailRejectedWithoutSideEffects(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(databaseURL) == "" {
		t.Skip("DATABASE_URL not set")
	}

	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)
	app, pool := newPersistentIntegrationApp(t, logger, memCache, databaseURL)
	resetPersistentState(t, pool)

	invalidEmails := []string{
		"admin@tenant",
		"admin@",
		"@tenant.com",
		"admin tenant.com",
		"admin@tenant.",
		"admin@.com",
		"",
		"   ",
	}

	for _, email := range invalidEmails {
		var before int
		require.NoError(t, pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM tenant_admin_invitations`).Scan(&before))

		body := fmt.Sprintf(`{"email":%q,"role":"member"}`, email)
		resp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/users/invite", "Bearer sk-master", body)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "email=%q", email)

		var payload map[string]interface{}
		decodeJSONBody(t, resp, &payload)
		assert.Contains(t, strings.ToLower(fmt.Sprint(payload["error"])), "valid email", "email=%q", email)

		var after int
		require.NoError(t, pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM tenant_admin_invitations`).Scan(&after))
		assert.Equal(t, before, after, "email=%q", email)
	}
}

func TestIntegration_RoleCatalogAndUserRoleAssignment(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(databaseURL) == "" {
		t.Skip("DATABASE_URL not set")
	}

	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)
	app, pool := newPersistentIntegrationApp(t, logger, memCache, databaseURL)
	resetPersistentState(t, pool)

	ctx := context.Background()
	userID := uuid.NewString()
	userEmail := "roles-user@tenant.test"
	_, err := pool.Exec(ctx, `
		INSERT INTO users (user_id, tenant_id, user_alias, email, role, active, created_at, updated_at)
		SELECT $1::uuid, tenant_id, 'roles-user', $2, 'member', true, NOW(), NOW()
		FROM tenants WHERE slug = 'default'
	`, userID, userEmail)
	require.NoError(t, err)

	listSystem := performJSONRequest(t, app, http.MethodGet, "http://localhost/api/v1/roles?limit=50&offset=0", "Bearer sk-master", "")
	assert.Equal(t, http.StatusOK, listSystem.StatusCode)
	var systemRoles platform.RoleListResponse
	decodeJSONBody(t, listSystem, &systemRoles)
	assert.GreaterOrEqual(t, len(systemRoles.Data), 6)

	createRoleResp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/roles", "Bearer sk-master", `{
		"name":"billing_auditor",
		"display_name":"Billing Auditor",
		"description":"Read financial and trace data",
		"permissions":["wallet:read","trace:list","trace:read","user:list"]
	}`)
	assert.Equal(t, http.StatusCreated, createRoleResp.StatusCode)
	var customRole platform.RoleDefinition
	decodeJSONBody(t, createRoleResp, &customRole)
	assert.Equal(t, "billing_auditor", customRole.Name)
	assert.False(t, customRole.IsSystem)

	updateRoleResp := performJSONRequest(t, app, http.MethodPatch, "http://localhost/api/v1/roles/"+customRole.ID, "Bearer sk-master", `{
		"display_name":"Billing & Trace Auditor",
		"permissions":["wallet:read","trace:list","trace:read","user:list","team:read"]
	}`)
	assert.Equal(t, http.StatusOK, updateRoleResp.StatusCode)
	var updatedRole platform.RoleDefinition
	decodeJSONBody(t, updateRoleResp, &updatedRole)
	assert.Equal(t, "Billing & Trace Auditor", updatedRole.DisplayName)
	assert.Contains(t, updatedRole.Permissions, "team:read")

	assignResp := performJSONRequest(t, app, http.MethodPatch, "http://localhost/api/v1/users/roles", "Bearer sk-master", fmt.Sprintf(`{
		"email":"%s",
		"roles":["billing_auditor","team_admin"],
		"resourceType":"team"
	}`, userEmail))
	assert.Equal(t, http.StatusOK, assignResp.StatusCode)
	var assignOut platform.UpdateUserRolesResponse
	decodeJSONBody(t, assignResp, &assignOut)
	assert.Equal(t, userEmail, assignOut.Email)
	assert.ElementsMatch(t, []string{"billing_auditor", "team_admin"}, assignOut.Roles)
	assert.Equal(t, "team", assignOut.ResourceType)

	usersResp := performJSONRequest(t, app, http.MethodGet, "http://localhost/api/v1/users?query=roles-user@tenant.test", "Bearer sk-master", "")
	assert.Equal(t, http.StatusOK, usersResp.StatusCode)
	var users platform.TenantAdminUserListResponse
	decodeJSONBody(t, usersResp, &users)
	require.Len(t, users.Data, 1)
	assert.ElementsMatch(t, []string{"billing_auditor", "team_admin"}, users.Data[0].Roles)
	assert.Equal(t, "team_admin", users.Data[0].Role)

	// Protected system role must not be mutable.
	var tenantAdminRoleID string
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT role_id::text
		FROM roles
		WHERE tenant_id = (SELECT tenant_id FROM tenants WHERE slug='default')
		  AND name = 'tenant_admin'
		LIMIT 1
	`).Scan(&tenantAdminRoleID))
	protectedUpdate := performJSONRequest(t, app, http.MethodPatch, "http://localhost/api/v1/roles/"+tenantAdminRoleID, "Bearer sk-master", `{"display_name":"Nope"}`)
	assert.Equal(t, http.StatusForbidden, protectedUpdate.StatusCode)
}

func TestIntegration_UpdateUserRoles_ValidationAndAuth(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(databaseURL) == "" {
		t.Skip("DATABASE_URL not set")
	}

	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)
	app, pool := newPersistentIntegrationApp(t, logger, memCache, databaseURL)
	resetPersistentState(t, pool)

	// unauthorized (invalid key)
	unauthorized := performJSONRequest(t, app, http.MethodPatch, "http://localhost/api/v1/users/roles", "Bearer sk-invalid", `{
		"email":"user@tenant.test",
		"roles":["member"]
	}`)
	assert.Equal(t, http.StatusUnauthorized, unauthorized.StatusCode)

	// validation error
	validation := performJSONRequest(t, app, http.MethodPatch, "http://localhost/api/v1/users/roles", "Bearer sk-master", `{
		"email":"",
		"roles":[]
	}`)
	assert.Equal(t, http.StatusUnprocessableEntity, validation.StatusCode)

	// unknown user
	notFound := performJSONRequest(t, app, http.MethodPatch, "http://localhost/api/v1/users/roles", "Bearer sk-master", `{
		"email":"missing@tenant.test",
		"roles":["member"]
	}`)
	assert.Equal(t, http.StatusNotFound, notFound.StatusCode)
}

func TestIntegration_ConcurrentWalletDebitUsesRowLock(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(databaseURL) == "" {
		t.Skip("DATABASE_URL not set")
	}

	logger, _ := zap.NewDevelopment()
	ctx := context.Background()

	store, err := storage.NewStore(ctx, databaseURL, logger)
	require.NoError(t, err)
	t.Cleanup(store.Close)

	pool, err := pgxpool.New(ctx, databaseURL)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	resetPersistentState(t, pool)

	tenant := createIntegrationTenant(t, pool, "race-lock")

	_, err = store.TopUpWallet(ctx, tenant.ID, 1_000_000, map[string]interface{}{"source": "concurrency-test"})
	require.NoError(t, err)

	const workers = 20
	debit := platform.UsageCostBreakdown{
		Model:        "gpt-4o-mini",
		TotalTokens:  1000,
		TotalCostUSD: 0.1,
		ExchangeRate: 850000,
		TotalCostIRR: 100_000,
	}

	var (
		wg           sync.WaitGroup
		mu           sync.Mutex
		successes    int
		insufficient int
		otherErrors  []error
	)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := store.DebitWallet(ctx, tenant.ID, debit, fmt.Sprintf("req-race-%d", i))

			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				successes++
			case errors.Is(err, storage.ErrInsufficientWalletBalance):
				insufficient++
			default:
				otherErrors = append(otherErrors, err)
			}
		}(i)
	}
	wg.Wait()

	require.Empty(t, otherErrors)
	assert.Equal(t, 10, successes)
	assert.Equal(t, workers-successes, insufficient)

	wallet, err := store.GetWallet(ctx, tenant.ID)
	require.NoError(t, err)
	assert.InDelta(t, 0.0, wallet.BalanceIRR, 0.0001)

	var debits int
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM wallet_transactions
		WHERE tenant_id = $1 AND tx_type = 'CONSUMPTION'
	`, tenant.ID).Scan(&debits))
	assert.Equal(t, successes, debits)
}

func TestIntegration_ProviderAccountAccessAndModelSelectionHardening(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(databaseURL) == "" {
		t.Skip("DATABASE_URL not set")
	}

	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)
	app, pool := newPersistentIntegrationApp(t, logger, memCache, databaseURL)
	resetPersistentState(t, pool)

	ctx := context.Background()
	userID := uuid.NewString()
	teamID := uuid.NewString()
	orgID := uuid.NewString()
	_, err := pool.Exec(ctx, `
		INSERT INTO organizations (organization_id, tenant_id, organization_alias, created_at, updated_at)
		SELECT $1::uuid, tenant_id, 'platform', NOW(), NOW()
		FROM tenants WHERE slug = 'default'
	`, orgID)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		INSERT INTO users (user_id, tenant_id, user_alias, role, active, created_at, updated_at)
		SELECT $1::uuid, tenant_id, 'owner', 'tenant_admin', true, NOW(), NOW()
		FROM tenants WHERE slug = 'default'
	`, userID)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		INSERT INTO teams (team_id, tenant_id, team_alias, organization_id, workspace_type, created_at, updated_at)
		SELECT $1::uuid, tenant_id, 'ml-team', $2::uuid, 'general', NOW(), NOW()
		FROM tenants WHERE slug = 'default'
	`, teamID, orgID)
	require.NoError(t, err)

	createAccountResp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/tenant-provider-accounts", "Bearer sk-master", `{
		"provider":"openai",
		"name":"openai-main",
		"auth_mode":"api_key",
		"credentials":{"api_key":"sk-provider-demo"}
	}`)
	assert.Equal(t, http.StatusCreated, createAccountResp.StatusCode)
	var account platform.TenantProviderAccount
	decodeJSONBody(t, createAccountResp, &account)
	require.NotEmpty(t, account.ID)

	modelsResp := performJSONRequest(t, app, http.MethodPut, "http://localhost/api/v1/tenant-provider-accounts/"+account.ID+"/models", "Bearer sk-master", `{
		"models":[
			{"model":"gpt-4.1-mini","enabled":true},
			{"model":"gpt-4.1-mini","enabled":true}
		]
	}`)
	assert.Equal(t, http.StatusOK, modelsResp.StatusCode)
	var modelsOut struct {
		Data []platform.TenantProviderAccountModel `json:"data"`
	}
	decodeJSONBody(t, modelsResp, &modelsOut)
	require.Len(t, modelsOut.Data, 1)
	assert.Equal(t, "gpt-4.1-mini", modelsOut.Data[0].Model)

	invalidModelResp := performJSONRequest(t, app, http.MethodPut, "http://localhost/api/v1/tenant-provider-accounts/"+account.ID+"/models", "Bearer sk-master", `{
		"models":[{"model":"not-a-real-model","enabled":true}]
	}`)
	assert.Equal(t, http.StatusBadRequest, invalidModelResp.StatusCode)
	var invalidModelErr map[string]interface{}
	decodeJSONBody(t, invalidModelResp, &invalidModelErr)
	errObj, ok := invalidModelErr["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "validation_error", errObj["type"])

	invalidPrincipalResp := performJSONRequest(t, app, http.MethodPut, "http://localhost/api/v1/tenant-provider-accounts/"+account.ID+"/access", "Bearer sk-master", fmt.Sprintf(`{
		"access":{
			"managers":[{"type":"user","principal_id":"%s"}],
			"users":[{"type":"team","principal_id":"%s"}]
		}
	}`, uuid.NewString(), teamID))
	assert.Equal(t, http.StatusBadRequest, invalidPrincipalResp.StatusCode)

	validAccessResp := performJSONRequest(t, app, http.MethodPut, "http://localhost/api/v1/tenant-provider-accounts/"+account.ID+"/access", "Bearer sk-master", fmt.Sprintf(`{
		"access":{
			"managers":[{"type":"user","principal_id":"%s"},{"type":"user","principal_id":"%s"}],
			"users":[{"type":"team","principal_id":"%s"},{"type":"team","principal_id":"%s"}]
		}
	}`, userID, userID, teamID, teamID))
	assert.Equal(t, http.StatusOK, validAccessResp.StatusCode)
	var accessOut struct {
		Access platform.TenantProviderAccountAccessControl `json:"access"`
	}
	decodeJSONBody(t, validAccessResp, &accessOut)
	require.Len(t, accessOut.Access.Managers, 1)
	require.Len(t, accessOut.Access.Users, 1)

	// Effective-models preview must not fail with 500 when user/team selectors send non-UUID strings.
	effectiveModelsResp := performJSONRequest(
		t,
		app,
		http.MethodGet,
		"http://localhost/api/v1/tenant-provider-accounts/"+account.ID+"/effective-models?key_type=pat&user_id=not-a-uuid",
		"Bearer sk-master",
		"",
	)
	assert.Equal(t, http.StatusOK, effectiveModelsResp.StatusCode)

	createModel(t, app, `{"model_name":"gpt-4.1-mini","provider":"openai","model":"gpt-4.1-mini","api_base":"http://127.0.0.1:65535/v1","api_key":"sk-provider","rpm":10}`)

	patResp := performJSONRequestWithHeaders(
		t,
		app,
		http.MethodPost,
		"http://localhost/api/v1/personal-access-tokens",
		"Bearer sk-master",
		map[string]string{"x-shahrah-user-id": userID},
		`{"name":"owner-pat"}`,
	)
	assert.Equal(t, http.StatusCreated, patResp.StatusCode)
	var pat platform.CreatedKey
	decodeJSONBody(t, patResp, &pat)
	require.NotEmpty(t, pat.Key)

	listModelsResp := performJSONRequest(t, app, http.MethodGet, "http://localhost/v1/models", "Bearer "+pat.Key, "")
	assert.Equal(t, http.StatusOK, listModelsResp.StatusCode)
	var listOut contractsopenai.ModelListResponse
	decodeJSONBody(t, listModelsResp, &listOut)
	require.Len(t, listOut.Data, 1)
	assert.Equal(t, "openai:openai-main:gpt-4.1-mini", listOut.Data[0].ID)
}

func TestIntegration_EffectiveModelsAccessMatrix_PAT_VAT_UserTeam(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(databaseURL) == "" {
		t.Skip("DATABASE_URL not set")
	}

	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)
	app, pool := newPersistentIntegrationApp(t, logger, memCache, databaseURL)
	resetPersistentState(t, pool)

	ctx := context.Background()
	userManagerID := uuid.NewString()
	userAID := uuid.NewString()
	userBID := uuid.NewString()
	teamAID := uuid.NewString()
	teamBID := uuid.NewString()
	orgID := uuid.NewString()

	_, err := pool.Exec(ctx, `
		INSERT INTO organizations (organization_id, tenant_id, organization_alias, created_at, updated_at)
		SELECT $1::uuid, tenant_id, 'platform', NOW(), NOW()
		FROM tenants WHERE slug = 'default'
	`, orgID)
	require.NoError(t, err)

	for _, uid := range []string{userManagerID, userAID, userBID} {
		_, err = pool.Exec(ctx, `
			INSERT INTO users (user_id, tenant_id, user_alias, role, active, created_at, updated_at)
			SELECT $1::uuid, tenant_id, 'member', 'member', true, NOW(), NOW()
			FROM tenants WHERE slug = 'default'
		`, uid)
		require.NoError(t, err)
	}

	for _, tid := range []string{teamAID, teamBID} {
		_, err = pool.Exec(ctx, `
			INSERT INTO teams (team_id, tenant_id, team_alias, organization_id, workspace_type, created_at, updated_at)
			SELECT $1::uuid, tenant_id, 'ml-team', $2::uuid, 'general', NOW(), NOW()
			FROM tenants WHERE slug = 'default'
		`, tid, orgID)
		require.NoError(t, err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO team_memberships (tenant_id, team_id, user_id, role, created_at, updated_at)
		SELECT t.tenant_id, t.team_id, $2::uuid, 'team_member', NOW(), NOW()
		FROM teams t
		WHERE t.team_id = $1::uuid
	`, teamAID, userAID)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		INSERT INTO team_memberships (tenant_id, team_id, user_id, role, created_at, updated_at)
		SELECT t.tenant_id, t.team_id, $2::uuid, 'team_member', NOW(), NOW()
		FROM teams t
		WHERE t.team_id = $1::uuid
	`, teamBID, userBID)
	require.NoError(t, err)

	createAccountResp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/tenant-provider-accounts", "Bearer sk-master", `{
		"provider":"openai",
		"name":"openai-access-matrix",
		"auth_mode":"api_key",
		"credentials":{"api_key":"sk-provider-demo"}
	}`)
	assert.Equal(t, http.StatusCreated, createAccountResp.StatusCode)
	var account platform.TenantProviderAccount
	decodeJSONBody(t, createAccountResp, &account)
	require.NotEmpty(t, account.ID)

	modelsResp := performJSONRequest(t, app, http.MethodPut, "http://localhost/api/v1/tenant-provider-accounts/"+account.ID+"/models", "Bearer sk-master", `{
		"models":[
			{"model":"gpt-4.1-mini","enabled":true},
			{"model":"gpt-4o","enabled":true}
		]
	}`)
	assert.Equal(t, http.StatusOK, modelsResp.StatusCode)

	setAccess := func(usersJSON string) {
		resp := performJSONRequest(
			t,
			app,
			http.MethodPut,
			"http://localhost/api/v1/tenant-provider-accounts/"+account.ID+"/access",
			"Bearer sk-master",
			fmt.Sprintf(`{
				"access":{
					"managers":[{"type":"user","principal_id":"%s"}],
					"users":%s
				}
			}`, userManagerID, usersJSON),
		)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	assertAllowed := func(keyType string, userID string, teamID string, model string, expected bool) {
		url := fmt.Sprintf(
			"http://localhost/api/v1/tenant-provider-accounts/%s/effective-models?key_type=%s&user_id=%s&team_id=%s",
			account.ID,
			keyType,
			userID,
			teamID,
		)
		resp := performJSONRequest(t, app, http.MethodGet, url, "Bearer sk-master", "")
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var out platform.EffectiveModelsResponse
		decodeJSONBody(t, resp, &out)
		found := false
		for _, item := range out.Data {
			if item.Model == model {
				found = true
				assert.Equal(t, expected, item.Allowed, "key_type=%s user_id=%s team_id=%s model=%s", keyType, userID, teamID, model)
			}
		}
		require.True(t, found, "model %s not found in effective-models response", model)
	}

	createPAT := func(userID string, name string) {
		resp := performJSONRequestWithHeaders(
			t,
			app,
			http.MethodPost,
			"http://localhost/api/v1/personal-access-tokens",
			"Bearer sk-master",
			map[string]string{"x-shahrah-user-id": userID},
			fmt.Sprintf(`{"name":"%s"}`, name),
		)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}
	createVAT := func(teamID string, name string) {
		resp := performJSONRequest(
			t,
			app,
			http.MethodPost,
			"http://localhost/api/v1/keys/vat",
			"Bearer sk-master",
			fmt.Sprintf(`{"key_alias":"%s","allowed_models":["*"],"team_id":"%s"}`, name, teamID),
		)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}
	createPAT(userManagerID, "pat-manager")
	createPAT(userAID, "pat-user-a")
	createPAT(userBID, "pat-user-b")
	createVAT(teamAID, "vat-team-a")
	createVAT(teamBID, "vat-team-b")

	// Scenario 1: everyone can access (PAT/VAT should both allow).
	setAccess(`[{"type":"everyone"}]`)
	assertAllowed("pat", userManagerID, "", "gpt-4.1-mini", true)
	assertAllowed("vat", "", teamAID, "gpt-4.1-mini", true)

	// Scenario 2: user-specific access.
	setAccess(fmt.Sprintf(`[{"type":"user","principal_id":"%s"}]`, userAID))
	assertAllowed("pat", userAID, "", "gpt-4.1-mini", true)
	assertAllowed("pat", userBID, "", "gpt-4.1-mini", false)
	assertAllowed("vat", "", teamAID, "gpt-4.1-mini", false)

	// Scenario 3: team-specific access.
	setAccess(fmt.Sprintf(`[{"type":"team","principal_id":"%s"}]`, teamAID))
	assertAllowed("vat", "", teamAID, "gpt-4.1-mini", true)
	assertAllowed("vat", "", teamBID, "gpt-4.1-mini", false)
	assertAllowed("pat", userAID, "", "gpt-4.1-mini", true)

	// Scenario 4: mixed user/team access (OR semantics).
	setAccess(fmt.Sprintf(`[
		{"type":"user","principal_id":"%s"},
		{"type":"team","principal_id":"%s"}
	]`, userAID, teamAID))
	assertAllowed("pat", userAID, "", "gpt-4o", true)
	assertAllowed("vat", "", teamAID, "gpt-4o", true)
	assertAllowed("pat", userBID, "", "gpt-4o", false)
	assertAllowed("vat", "", teamBID, "gpt-4o", false)
}

func TestIntegration_EffectiveModels_PATKeyAllowlistIsApplied(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(databaseURL) == "" {
		t.Skip("DATABASE_URL not set")
	}

	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)
	app, pool := newPersistentIntegrationApp(t, logger, memCache, databaseURL)
	resetPersistentState(t, pool)

	ctx := context.Background()
	userID := uuid.NewString()

	_, err := pool.Exec(ctx, `
		INSERT INTO users (user_id, tenant_id, user_alias, role, active, created_at, updated_at)
		SELECT $1::uuid, tenant_id, 'member', 'member', true, NOW(), NOW()
		FROM tenants WHERE slug = 'default'
	`, userID)
	require.NoError(t, err)

	createAccountResp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/tenant-provider-accounts", "Bearer sk-master", `{
		"provider":"openai",
		"name":"openai-key-allowlist-preview",
		"auth_mode":"api_key",
		"credentials":{"api_key":"sk-provider-demo"}
	}`)
	assert.Equal(t, http.StatusCreated, createAccountResp.StatusCode)
	var account platform.TenantProviderAccount
	decodeJSONBody(t, createAccountResp, &account)
	require.NotEmpty(t, account.ID)

	modelsResp := performJSONRequest(t, app, http.MethodPut, "http://localhost/api/v1/tenant-provider-accounts/"+account.ID+"/models", "Bearer sk-master", `{
		"models":[
			{"model":"gpt-4.1-mini","enabled":true},
			{"model":"gpt-4o","enabled":true}
		]
	}`)
	assert.Equal(t, http.StatusOK, modelsResp.StatusCode)

	accessResp := performJSONRequest(t, app, http.MethodPut, "http://localhost/api/v1/tenant-provider-accounts/"+account.ID+"/access", "Bearer sk-master", fmt.Sprintf(`{
		"access":{
			"managers":[{"type":"user","principal_id":"%s"}],
			"users":[{"type":"user","principal_id":"%s"}]
		}
	}`, userID, userID))
	assert.Equal(t, http.StatusOK, accessResp.StatusCode)

	patResp := performJSONRequestWithHeaders(
		t,
		app,
		http.MethodPost,
		"http://localhost/api/v1/personal-access-tokens",
		"Bearer sk-master",
		map[string]string{"x-shahrah-user-id": userID},
		`{"name":"pat-preview-allowlist"}`,
	)
	assert.Equal(t, http.StatusCreated, patResp.StatusCode)

	effectiveResp := performJSONRequest(
		t,
		app,
		http.MethodGet,
		fmt.Sprintf("http://localhost/api/v1/tenant-provider-accounts/%s/effective-models?key_type=pat&user_id=%s", account.ID, userID),
		"Bearer sk-master",
		"",
	)
	assert.Equal(t, http.StatusOK, effectiveResp.StatusCode)
	var out platform.EffectiveModelsResponse
	decodeJSONBody(t, effectiveResp, &out)
	require.NotEmpty(t, out.Data)

	byModel := make(map[string]platform.EffectiveModelView, len(out.Data))
	for _, item := range out.Data {
		byModel[item.Model] = item
	}
	require.Contains(t, byModel, "gpt-4.1-mini")
	require.Contains(t, byModel, "gpt-4o")
	assert.True(t, byModel["gpt-4.1-mini"].Allowed)
	assert.True(t, byModel["gpt-4o"].Allowed)
	assert.Empty(t, byModel["gpt-4o"].DeniedReason)
}

func TestIntegration_EffectiveModels_PATKeyAllowlistAcceptsFQN(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(databaseURL) == "" {
		t.Skip("DATABASE_URL not set")
	}

	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)
	app, pool := newPersistentIntegrationApp(t, logger, memCache, databaseURL)
	resetPersistentState(t, pool)

	ctx := context.Background()
	userID := uuid.NewString()

	_, err := pool.Exec(ctx, `
		INSERT INTO users (user_id, tenant_id, user_alias, role, active, created_at, updated_at)
		SELECT $1::uuid, tenant_id, 'member', 'member', true, NOW(), NOW()
		FROM tenants WHERE slug = 'default'
	`, userID)
	require.NoError(t, err)

	accountAlias := "openai-key-allowlist-preview-fqn"
	createAccountResp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/tenant-provider-accounts", "Bearer sk-master", fmt.Sprintf(`{
		"provider":"openai",
		"name":"%s",
		"auth_mode":"api_key",
		"credentials":{"api_key":"sk-provider-demo"}
	}`, accountAlias))
	assert.Equal(t, http.StatusCreated, createAccountResp.StatusCode)
	var account platform.TenantProviderAccount
	decodeJSONBody(t, createAccountResp, &account)
	require.NotEmpty(t, account.ID)

	modelsResp := performJSONRequest(t, app, http.MethodPut, "http://localhost/api/v1/tenant-provider-accounts/"+account.ID+"/models", "Bearer sk-master", `{
		"models":[
			{"model":"gpt-4.1-mini","enabled":true},
			{"model":"gpt-4o","enabled":true}
		]
	}`)
	assert.Equal(t, http.StatusOK, modelsResp.StatusCode)

	accessResp := performJSONRequest(t, app, http.MethodPut, "http://localhost/api/v1/tenant-provider-accounts/"+account.ID+"/access", "Bearer sk-master", fmt.Sprintf(`{
		"access":{
			"managers":[{"type":"user","principal_id":"%s"}],
			"users":[{"type":"user","principal_id":"%s"}]
		}
	}`, userID, userID))
	assert.Equal(t, http.StatusOK, accessResp.StatusCode)

	patResp := performJSONRequestWithHeaders(
		t,
		app,
		http.MethodPost,
		"http://localhost/api/v1/personal-access-tokens",
		"Bearer sk-master",
		map[string]string{"x-shahrah-user-id": userID},
		`{"name":"pat-preview-allowlist-fqn"}`,
	)
	assert.Equal(t, http.StatusCreated, patResp.StatusCode)

	effectiveResp := performJSONRequest(
		t,
		app,
		http.MethodGet,
		fmt.Sprintf("http://localhost/api/v1/tenant-provider-accounts/%s/effective-models?key_type=pat&user_id=%s", account.ID, userID),
		"Bearer sk-master",
		"",
	)
	assert.Equal(t, http.StatusOK, effectiveResp.StatusCode)
	var out platform.EffectiveModelsResponse
	decodeJSONBody(t, effectiveResp, &out)
	require.NotEmpty(t, out.Data)

	byModel := make(map[string]platform.EffectiveModelView, len(out.Data))
	for _, item := range out.Data {
		byModel[item.Model] = item
	}
	require.Contains(t, byModel, "gpt-4o")
	assert.True(t, byModel["gpt-4o"].Allowed)
	assert.Empty(t, byModel["gpt-4o"].DeniedReason)
}

func TestIntegration_DeleteTenantProviderAccount_Success(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(databaseURL) == "" {
		t.Skip("DATABASE_URL not set")
	}

	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)
	app, pool := newPersistentIntegrationApp(t, logger, memCache, databaseURL)
	resetPersistentState(t, pool)

	ctx := context.Background()
	userID := uuid.NewString()
	orgID := uuid.NewString()
	teamID := uuid.NewString()
	_, err := pool.Exec(ctx, `
		INSERT INTO organizations (organization_id, tenant_id, organization_alias, created_at, updated_at)
		SELECT $1::uuid, tenant_id, 'platform', NOW(), NOW()
		FROM tenants WHERE slug = 'default'
	`, orgID)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		INSERT INTO users (user_id, tenant_id, user_alias, role, active, created_at, updated_at)
		SELECT $1::uuid, tenant_id, 'owner', 'tenant_admin', true, NOW(), NOW()
		FROM tenants WHERE slug = 'default'
	`, userID)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		INSERT INTO teams (team_id, tenant_id, team_alias, organization_id, workspace_type, created_at, updated_at)
		SELECT $1::uuid, tenant_id, 'ml-team', $2::uuid, 'general', NOW(), NOW()
		FROM tenants WHERE slug = 'default'
	`, teamID, orgID)
	require.NoError(t, err)

	createResp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/tenant-provider-accounts", "Bearer sk-master", `{
		"provider":"openai",
		"name":"openai-delete-ok",
		"auth_mode":"api_key",
		"credentials":{"api_key":"sk-provider-demo"}
	}`)
	assert.Equal(t, http.StatusCreated, createResp.StatusCode)
	var account platform.TenantProviderAccount
	decodeJSONBody(t, createResp, &account)
	require.NotEmpty(t, account.ID)

	modelsResp := performJSONRequest(t, app, http.MethodPut, "http://localhost/api/v1/tenant-provider-accounts/"+account.ID+"/models", "Bearer sk-master", `{
		"models":[{"model":"gpt-4.1-mini","enabled":true}]
	}`)
	assert.Equal(t, http.StatusOK, modelsResp.StatusCode)

	accessResp := performJSONRequest(t, app, http.MethodPut, "http://localhost/api/v1/tenant-provider-accounts/"+account.ID+"/access", "Bearer sk-master", fmt.Sprintf(`{
		"access":{
			"managers":[{"type":"user","principal_id":"%s"}],
			"users":[{"type":"team","principal_id":"%s"}]
		}
	}`, userID, teamID))
	assert.Equal(t, http.StatusOK, accessResp.StatusCode)

	deleteResp := performJSONRequest(t, app, http.MethodDelete, "http://localhost/api/v1/tenant-provider-accounts/"+account.ID, "Bearer sk-master", "")
	assert.Equal(t, http.StatusNoContent, deleteResp.StatusCode)

	assertCount(t, pool, `SELECT COUNT(*) FROM tenant_provider_accounts WHERE account_id = '`+account.ID+`'`, 0)
	assertCount(t, pool, `SELECT COUNT(*) FROM tenant_provider_account_models WHERE account_id = '`+account.ID+`'`, 0)
	assertCount(t, pool, `SELECT COUNT(*) FROM tenant_provider_account_access WHERE account_id = '`+account.ID+`'`, 0)
}

func TestIntegration_DeleteTenantProviderAccount_ForbiddenForNonManager(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(databaseURL) == "" {
		t.Skip("DATABASE_URL not set")
	}

	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)
	app, pool := newPersistentIntegrationApp(t, logger, memCache, databaseURL)
	resetPersistentState(t, pool)

	ctx := context.Background()
	orgID := uuid.NewString()
	managerUserID := uuid.NewString()
	nonManagerUserID := uuid.NewString()
	_, err := pool.Exec(ctx, `
		INSERT INTO organizations (organization_id, tenant_id, organization_alias, created_at, updated_at)
		SELECT $1::uuid, tenant_id, 'platform', NOW(), NOW()
		FROM tenants WHERE slug = 'default'
	`, orgID)
	require.NoError(t, err)
	for _, uid := range []string{managerUserID, nonManagerUserID} {
		_, err = pool.Exec(ctx, `
			INSERT INTO users (user_id, tenant_id, user_alias, role, active, created_at, updated_at)
			SELECT $1::uuid, tenant_id, 'member', 'member', true, NOW(), NOW()
			FROM tenants WHERE slug = 'default'
		`, uid)
		require.NoError(t, err)
	}

	createResp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/tenant-provider-accounts", "Bearer sk-master", `{
		"provider":"openai",
		"name":"openai-delete-authz",
		"auth_mode":"api_key",
		"credentials":{"api_key":"sk-provider-demo"}
	}`)
	assert.Equal(t, http.StatusCreated, createResp.StatusCode)
	var account platform.TenantProviderAccount
	decodeJSONBody(t, createResp, &account)

	accessResp := performJSONRequest(t, app, http.MethodPut, "http://localhost/api/v1/tenant-provider-accounts/"+account.ID+"/access", "Bearer sk-master", fmt.Sprintf(`{
		"access":{
			"managers":[{"type":"user","principal_id":"%s"}],
			"users":[{"type":"user","principal_id":"%s"}]
		}
	}`, managerUserID, nonManagerUserID))
	assert.Equal(t, http.StatusOK, accessResp.StatusCode)

	patResp := performJSONRequestWithHeaders(
		t,
		app,
		http.MethodPost,
		"http://localhost/api/v1/personal-access-tokens",
		"Bearer sk-master",
		map[string]string{"x-shahrah-user-id": nonManagerUserID},
		`{"name":"non-manager-delete"}`,
	)
	assert.Equal(t, http.StatusCreated, patResp.StatusCode)
	var pat platform.CreatedKey
	decodeJSONBody(t, patResp, &pat)
	require.NotEmpty(t, pat.Key)

	deleteResp := performJSONRequest(t, app, http.MethodDelete, "http://localhost/api/v1/tenant-provider-accounts/"+account.ID, "Bearer "+pat.Key, "")
	assert.Equal(t, http.StatusForbidden, deleteResp.StatusCode)
}

func TestIntegration_DeleteTenantProviderAccount_NotFound(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(databaseURL) == "" {
		t.Skip("DATABASE_URL not set")
	}

	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)
	app, pool := newPersistentIntegrationApp(t, logger, memCache, databaseURL)
	resetPersistentState(t, pool)

	deleteResp := performJSONRequest(t, app, http.MethodDelete, "http://localhost/api/v1/tenant-provider-accounts/"+uuid.NewString(), "Bearer sk-master", "")
	assert.Equal(t, http.StatusNotFound, deleteResp.StatusCode)
}

func TestIntegration_DeleteTenantProviderAccount_ConflictWhenReferencedByActiveKey(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(databaseURL) == "" {
		t.Skip("DATABASE_URL not set")
	}

	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)
	app, pool := newPersistentIntegrationApp(t, logger, memCache, databaseURL)
	resetPersistentState(t, pool)

	ctx := context.Background()
	userID := uuid.NewString()
	_, err := pool.Exec(ctx, `
		INSERT INTO users (user_id, tenant_id, user_alias, role, active, created_at, updated_at)
		SELECT $1::uuid, tenant_id, 'owner', 'tenant_admin', true, NOW(), NOW()
		FROM tenants WHERE slug = 'default'
	`, userID)
	require.NoError(t, err)

	createResp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/tenant-provider-accounts", "Bearer sk-master", `{
		"provider":"openai",
		"name":"openai-delete-blocked",
		"auth_mode":"api_key",
		"credentials":{"api_key":"sk-provider-demo"}
	}`)
	assert.Equal(t, http.StatusCreated, createResp.StatusCode)
	var account platform.TenantProviderAccount
	decodeJSONBody(t, createResp, &account)

	modelsResp := performJSONRequest(t, app, http.MethodPut, "http://localhost/api/v1/tenant-provider-accounts/"+account.ID+"/models", "Bearer sk-master", `{
		"models":[{"model":"gpt-4.1-mini","enabled":true}]
	}`)
	assert.Equal(t, http.StatusOK, modelsResp.StatusCode)

	accessResp := performJSONRequest(t, app, http.MethodPut, "http://localhost/api/v1/tenant-provider-accounts/"+account.ID+"/access", "Bearer sk-master", fmt.Sprintf(`{
		"access":{
			"managers":[{"type":"user","principal_id":"%s"}],
			"users":[{"type":"user","principal_id":"%s"}]
		}
	}`, userID, userID))
	assert.Equal(t, http.StatusOK, accessResp.StatusCode)

	patResp := performJSONRequestWithHeaders(
		t,
		app,
		http.MethodPost,
		"http://localhost/api/v1/personal-access-tokens",
		"Bearer sk-master",
		map[string]string{"x-shahrah-user-id": userID},
		`{"name":"key-ref-delete"}`,
	)
	assert.Equal(t, http.StatusCreated, patResp.StatusCode)

	deleteResp := performJSONRequest(t, app, http.MethodDelete, "http://localhost/api/v1/tenant-provider-accounts/"+account.ID, "Bearer sk-master", "")
	assert.Equal(t, http.StatusConflict, deleteResp.StatusCode)
}

func createIntegrationTenant(t *testing.T, pool *pgxpool.Pool, slug string) *platform.Tenant {
	t.Helper()
	var tenant platform.Tenant
	err := pool.QueryRow(context.Background(), `
		INSERT INTO tenants (slug, domain, status)
		VALUES ($1, $2, 'active')
		RETURNING tenant_id::text, slug, domain, status, created_at
	`, slug, slug+".localhost").Scan(&tenant.ID, &tenant.Slug, &tenant.Domain, &tenant.Status, &tenant.CreatedAt)
	require.NoError(t, err)
	_, err = pool.Exec(context.Background(), `
		INSERT INTO wallets (tenant_id, balance_irr, pending_irr)
		VALUES ($1::uuid, 0, 0)
	`, tenant.ID)
	require.NoError(t, err)
	return &tenant
}

func newPersistentIntegrationApp(t *testing.T, logger *zap.Logger, memCache cache.Cache, databaseURL string) (*fiber.App, *pgxpool.Pool) {
	t.Helper()

	ctx := context.Background()
	store, err := storage.NewStore(ctx, databaseURL, logger)
	require.NoError(t, err)
	t.Cleanup(store.Close)

	pool, err := pgxpool.New(ctx, databaseURL)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	silkRoad := httptest.NewServer(silkserver.NewHandler(nil))
	t.Cleanup(silkRoad.Close)
	tunnelServer := httptest.NewServer(tserverserver.NewHandler(silkRoad.URL, nil))
	t.Cleanup(tunnelServer.Close)
	tunnelClient := httptest.NewServer(tclientserver.NewHandler(tunnelServer.URL, nil))
	t.Cleanup(tunnelClient.Close)

	llmRouter, err := router.NewRouter(nil, config.RouterSettings{
		RoutingStrategy: "simple-shuffle",
		CooldownTime:    60,
	}, memCache, store, logger)
	require.NoError(t, err)

	billingSvc := billing.NewService(memCache, store, 850000, "test")
	app := controlserver.NewFiberApp(&config.Config{}, controlserver.Dependencies{
		AuthMiddleware: auth.NewAuthMiddleware(store, "sk-master", logger),
		TenantResolver: tenantmw.NewMiddleware(store, "default", true),
		AdminHandlers:  admin.New(store, store, store),
		ProxyHandler: proxy.NewChatCompletionHandler(
			llmRouter,
			tunneltransport.NewHTTPTransport(tunnelClient.URL, 30*time.Second),
			billingSvc,
			store,
			store,
			logger,
		),
		CompletionsHandler: proxy.NewCompletionsHandler(
			llmRouter,
			tunneltransport.NewHTTPTransport(tunnelClient.URL, 30*time.Second),
			billingSvc,
			store,
			store,
			logger,
		),
		ModelsHandler: proxy.NewModelsHandler(store),
		Logger:        logger,
	})

	return app, pool
}

func resetPersistentState(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	ctx := context.Background()
	_, err := pool.Exec(ctx, `
		TRUNCATE TABLE
			tenant_onboarding_requests,
			wallet_transactions,
			wallets,
			exchange_rates,
			request_logs,
			spend_logs,
			verification_tokens,
			model_deployments,
			tenants
		RESTART IDENTITY CASCADE
	`)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO tenants (slug, domain, status)
		VALUES ('default', 'default', 'active')
		ON CONFLICT (slug) DO NOTHING
	`)
	require.NoError(t, err)
}

func assertCount(t *testing.T, pool *pgxpool.Pool, query string, expected int) {
	t.Helper()
	var count int
	require.NoError(t, pool.QueryRow(context.Background(), query).Scan(&count))
	assert.Equal(t, expected, count, query)
}

func assertColumnsExist(t *testing.T, pool *pgxpool.Pool, table string, required []string) {
	t.Helper()

	rows, err := pool.Query(context.Background(), `
		SELECT column_name
		FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = $1
	`, table)
	require.NoError(t, err)
	defer rows.Close()

	existing := make(map[string]struct{})
	for rows.Next() {
		var name string
		require.NoError(t, rows.Scan(&name))
		existing[name] = struct{}{}
	}
	require.NoError(t, rows.Err())

	for _, col := range required {
		_, ok := existing[col]
		assert.Truef(t, ok, "missing column %s.%s", table, col)
	}
}

func assertIndexForColumn(t *testing.T, pool *pgxpool.Pool, table, column string) {
	t.Helper()

	var count int
	err := pool.QueryRow(context.Background(), `
		SELECT COUNT(*)
		FROM pg_indexes
		WHERE schemaname = 'public'
		  AND tablename = $1
		  AND lower(indexdef) LIKE '%' || lower('(' || $2 || ')') || '%'
	`, table, column).Scan(&count)
	require.NoError(t, err)
	assert.Greaterf(t, count, 0, "missing index for %s(%s)", table, column)
}
