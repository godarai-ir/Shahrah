package integration_test

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
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
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestIntegration_EndToEndChatCompletion(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/chat/completions", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(contractsopenai.ChatCompletionResponse{
			ID:      "chatcmpl-test",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "gpt-3.5-turbo",
			Choices: []contractsopenai.Choice{{
				Index: 0,
				Message: &contractsopenai.ChatMessage{
					Role:    "assistant",
					Content: "hello from provider",
				},
				FinishReason: "stop",
			}},
			Usage: &contractsopenai.Usage{
				PromptTokens:     10,
				CompletionTokens: 5,
				TotalTokens:      15,
			},
		})
	}))
	defer provider.Close()

	silkRoad := httptest.NewServer(silkserver.NewHandler(nil))
	defer silkRoad.Close()
	tunnelServer := httptest.NewServer(tserverserver.NewHandler(silkRoad.URL, nil))
	defer tunnelServer.Close()
	tunnelClient := httptest.NewServer(tclientserver.NewHandler(tunnelServer.URL, nil))
	defer tunnelClient.Close()

	tenantStore := integrationTenantStore{
		tenant: &platform.Tenant{ID: "tenant-1", Slug: "default", Domain: "default", Status: "active", CreatedAt: time.Now()},
	}
	authStore := auth.NewInMemoryKeyStore()
	authStore.AddKey(auth.HashKey("sk-integration-key"), &platform.AuthContext{
		TenantID:      "tenant-1",
		KeyID:         "key-1",
		KeyAlias:      "integration-key",
		Role:          "virtual_key",
		AllowedModels: []string{"gpt-3.5-turbo"},
		TokenHash:     auth.HashKey("sk-integration-key"),
	})

	modelStore := &integrationModelStore{
		deployments: []platform.Deployment{{
			ID:        "dep-1",
			TenantID:  "tenant-1",
			ModelName: "gpt-3.5-turbo",
			Provider:  "openai",
			Model:     "gpt-3.5-turbo",
			APIBase:   provider.URL + "/v1",
			APIKey:    "sk-provider-test",
			RPM:       100,
		}},
	}
	walletStore := &integrationWalletStore{
		wallet: &platform.Wallet{ID: "wallet-1", TenantID: "tenant-1", BalanceIRR: 10_000_000},
		rate:   &platform.ExchangeRate{ID: "rate-1", USDtoIRR: 850000, Source: "test", FetchedAt: time.Now()},
	}

	billingSvc := billing.NewService(memCache, walletStore, 850000, "test")
	llmRouter, err := router.NewRouter([]config.ModelDeployment{{
		ModelName: "gpt-3.5-turbo",
		ProviderEndpoint: config.ProviderEndpointConfig{
			Model:   "openai/gpt-3.5-turbo",
			APIBase: provider.URL + "/v1",
			APIKey:  "sk-provider-test",
			RPM:     100,
		},
		ModelInfo: map[string]interface{}{"id": "dep-1"},
	}}, config.RouterSettings{
		RoutingStrategy: "simple-shuffle",
		CooldownTime:    60,
	}, memCache, modelStore, logger)
	assert.NoError(t, err)

	app := controlserver.NewFiberApp(&config.Config{}, controlserver.Dependencies{
		AuthMiddleware: auth.NewAuthMiddleware(authStore, "sk-master", logger),
		TenantResolver: tenantmw.NewMiddleware(tenantStore, "default", true),
		AdminHandlers:  admin.New(noopKeyStore{}, modelStore, walletStore),
		ProxyHandler: proxy.NewChatCompletionHandler(
			llmRouter,
			tunneltransport.NewHTTPTransport(tunnelClient.URL, 30*time.Second),
			billingSvc,
			walletStore,
			walletStore,
			logger,
		),
		CompletionsHandler: proxy.NewCompletionsHandler(
			llmRouter,
			tunneltransport.NewHTTPTransport(tunnelClient.URL, 30*time.Second),
			billingSvc,
			walletStore,
			walletStore,
			logger,
		),
		ModelsHandler: proxy.NewModelsHandler(modelStore),
		Logger:        logger,
	})

	reqBody := `{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"hi"}]}`
	req := httptest.NewRequest(http.MethodPost, "http://localhost/v1/chat/completions", strings.NewReader(reqBody))
	req.Header.Set("Authorization", "Bearer sk-integration-key")
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, int((30 * time.Second).Milliseconds()))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	var completion contractsopenai.ChatCompletionResponse
	assert.NoError(t, json.Unmarshal(body, &completion))
	assert.Equal(t, "hello from provider", completion.Choices[0].Message.Content)
	assert.Len(t, walletStore.spendLogs, 1)
	assert.Len(t, walletStore.requestLogs, 1)
	assert.Len(t, walletStore.debits, 1)
	assert.Greater(t, walletStore.wallet.BalanceIRR, 0.0)
}

func TestIntegration_TenantOnboardingFlowViaManagementAPIs(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/chat/completions", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(contractsopenai.ChatCompletionResponse{
			ID:      "chatcmpl-managed",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "gpt-4o-mini",
			Choices: []contractsopenai.Choice{{
				Index: 0,
				Message: &contractsopenai.ChatMessage{
					Role:    "assistant",
					Content: "provisioned flow succeeded",
				},
				FinishReason: "stop",
			}},
			Usage: &contractsopenai.Usage{
				PromptTokens:     12,
				CompletionTokens: 6,
				TotalTokens:      18,
			},
		})
	}))
	defer provider.Close()

	app, keyStore, modelStore, walletStore := newIntegrationApp(t, logger, memCache, provider.URL)

	createdKey := createVirtualKey(t, app, `{"key_alias":"tenant-key","models":["gpt-4o-mini"]}`)
	assert.NotEmpty(t, createdKey.Key)
	assert.Equal(t, "tenant-key", createdKey.KeyAlias)

	createModel(t, app, fmt.Sprintf(`{"model_name":"gpt-4o-mini","provider":"openai","model":"gpt-4o-mini","api_base":"%s/v1","api_key":"sk-provider","rpm":250}`, provider.URL))
	topUpWallet(t, app, `{"amount_irr":1500000}`)

	keysResp := performJSONRequest(t, app, http.MethodGet, "http://localhost/api/v1/keys", "Bearer sk-master", "")
	assert.Equal(t, http.StatusOK, keysResp.StatusCode)
	var listedKeys struct {
		Data []platform.CreatedKey `json:"data"`
	}
	decodeJSONBody(t, keysResp, &listedKeys)
	require.Len(t, listedKeys.Data, 1)
	assert.Empty(t, listedKeys.Data[0].Key)
	assert.NotEmpty(t, listedKeys.Data[0].MaskedKey)

	modelsResp := performJSONRequest(t, app, http.MethodGet, "http://localhost/v1/models", "Bearer "+createdKey.Key, "")
	assert.Equal(t, http.StatusOK, modelsResp.StatusCode)
	var modelList contractsopenai.ModelListResponse
	decodeJSONBody(t, modelsResp, &modelList)
	require.Len(t, modelList.Data, 1)
	assert.Equal(t, "gpt-4o-mini", modelList.Data[0].ID)
	assert.Contains(t, modelList.Data[0].Aliases, "gpt-4o-mini")

	completionResp := performJSONRequest(t, app, http.MethodPost, "http://localhost/v1/chat/completions", "Bearer "+createdKey.Key, `{"model":"gpt-4o-mini","messages":[{"role":"user","content":"hello"}]}`)
	assert.Equal(t, http.StatusOK, completionResp.StatusCode)
	var completion contractsopenai.ChatCompletionResponse
	decodeJSONBody(t, completionResp, &completion)
	assert.Equal(t, "provisioned flow succeeded", completion.Choices[0].Message.Content)

	require.Len(t, walletStore.debits, 1)
	require.Len(t, walletStore.requestLogs, 1)
	require.Len(t, walletStore.spendLogs, 1)
	assert.Greater(t, walletStore.wallet.BalanceIRR, 0.0)

	authInfo, err := keyStore.GetKeyByHash(context.Background(), hashToken(createdKey.Key))
	require.NoError(t, err)
	assert.Greater(t, authInfo.Spend, 0.0)
	require.Len(t, modelStore.deployments, 1)
}

func TestIntegration_ChatCompletionRejectsInsufficientWalletBalance(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(contractsopenai.ChatCompletionResponse{
			ID:      "chatcmpl-no-balance",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "gpt-4",
			Choices: []contractsopenai.Choice{{
				Index: 0,
				Message: &contractsopenai.ChatMessage{
					Role:    "assistant",
					Content: "should not succeed",
				},
				FinishReason: "stop",
			}},
			Usage: &contractsopenai.Usage{
				PromptTokens:     50,
				CompletionTokens: 50,
				TotalTokens:      100,
			},
		})
	}))
	defer provider.Close()

	app, _, _, walletStore := newIntegrationApp(t, logger, memCache, provider.URL)

	createdKey := createVirtualKey(t, app, `{"key_alias":"low-balance","models":["gpt-4"]}`)
	createModel(t, app, fmt.Sprintf(`{"model_name":"gpt-4","provider":"openai","model":"gpt-4","api_base":"%s/v1","api_key":"sk-provider","rpm":50}`, provider.URL))
	topUpWallet(t, app, `{"amount_irr":1}`)

	resp := performJSONRequest(t, app, http.MethodPost, "http://localhost/v1/chat/completions", "Bearer "+createdKey.Key, `{"model":"gpt-4","messages":[{"role":"user","content":"expensive request"}]}`)
	assert.Equal(t, http.StatusPaymentRequired, resp.StatusCode)

	var apiErr contractsopenai.APIError
	decodeJSONBody(t, resp, &apiErr)
	assert.Contains(t, apiErr.Error.Type, "budget")
	assert.Empty(t, walletStore.debits)
	assert.Empty(t, walletStore.requestLogs)
	assert.Empty(t, walletStore.spendLogs)
	assert.Equal(t, 1.0, walletStore.wallet.BalanceIRR)
}

func TestIntegration_CreateServiceAccountKeyRequiresTeamID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer provider.Close()

	app, _, _, _ := newIntegrationApp(t, logger, memCache, provider.URL)

	resp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/keys", "Bearer sk-master", `{"key_alias":"svc-no-team","models":["gpt-4o-mini"],"role":"service_account"}`)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestIntegration_CreateServiceAccountKey(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer provider.Close()

	app, _, _, _ := newIntegrationApp(t, logger, memCache, provider.URL)

	createdKey := createVirtualKey(t, app, `{"key_alias":"svc-team-key","models":["gpt-4o-mini"],"role":"service_account","team_id":"team-ops","organization_id":"org-platform"}`)
	assert.Equal(t, "service_account", createdKey.Role)
	assert.Equal(t, "team-ops", createdKey.TeamID)
	assert.Equal(t, "org-platform", createdKey.OrgID)

	keysResp := performJSONRequest(t, app, http.MethodGet, "http://localhost/api/v1/keys", "Bearer sk-master", "")
	assert.Equal(t, http.StatusOK, keysResp.StatusCode)
	var listedKeys struct {
		Data []platform.CreatedKey `json:"data"`
	}
	decodeJSONBody(t, keysResp, &listedKeys)
	require.Len(t, listedKeys.Data, 1)
	assert.Equal(t, "service_account", listedKeys.Data[0].Role)
	assert.Equal(t, "team-ops", listedKeys.Data[0].TeamID)
}

func TestIntegration_ListTenantUsers(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer provider.Close()

	app, _, _, _ := newIntegrationApp(t, logger, memCache, provider.URL)
	resp := performJSONRequest(t, app, http.MethodGet, "http://localhost/api/v1/users?limit=20&offset=0", "Bearer sk-master", "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var out platform.TenantAdminUserListResponse
	decodeJSONBody(t, resp, &out)
	require.GreaterOrEqual(t, out.Total, 1)
	require.NotEmpty(t, out.Data)
	assert.Equal(t, "default", out.Data[0].TenantSlug)
}

func TestIntegration_InviteTenantUser(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer provider.Close()

	app, _, _, _ := newIntegrationApp(t, logger, memCache, provider.URL)
	resp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/users/invite", "Bearer sk-master", `{"email":"new-user@tenant.test","role":"tenant_member"}`)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var out platform.InviteTenantAdminUserResponse
	decodeJSONBody(t, resp, &out)
	assert.Equal(t, "new-user@tenant.test", out.Invitation.Email)
	assert.Equal(t, "tenant_member", out.Invitation.Role)
	assert.NotEmpty(t, out.Link)
}

func TestIntegration_AcceptTenantUserInvitation(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer provider.Close()

	app, _, _, _ := newIntegrationApp(t, logger, memCache, provider.URL)

	inviteResp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/users/invite", "Bearer sk-master", `{"email":"accept-me@tenant.test","role":"tenant_member"}`)
	assert.Equal(t, http.StatusCreated, inviteResp.StatusCode)

	acceptResp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/users/invitations/accept", "Bearer sk-master", `{"token":"test-token"}`)
	assert.Equal(t, http.StatusOK, acceptResp.StatusCode)

	var out platform.AcceptTenantAdminInvitationResponse
	decodeJSONBody(t, acceptResp, &out)
	assert.Equal(t, "accepted", out.Invitation.Status)
	assert.Equal(t, "accept-me@tenant.test", out.User.Email)
}

func TestIntegration_ActivateDeactivateTenantUser(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer provider.Close()

	app, _, _, _ := newIntegrationApp(t, logger, memCache, provider.URL)

	deactivateResp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/users/tenant-user-1/deactivate", "Bearer sk-master", "")
	assert.Equal(t, http.StatusOK, deactivateResp.StatusCode)
	var deactivated platform.TenantAdminUser
	decodeJSONBody(t, deactivateResp, &deactivated)
	assert.False(t, deactivated.Active)
	assert.Equal(t, "deactivated", deactivated.Role)

	activateResp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/users/tenant-user-1/activate", "Bearer sk-master", "")
	assert.Equal(t, http.StatusOK, activateResp.StatusCode)
	var activated platform.TenantAdminUser
	decodeJSONBody(t, activateResp, &activated)
	assert.True(t, activated.Active)
	assert.NotEqual(t, "deactivated", activated.Role)
}

func TestIntegration_LegacyCompletionsEnforcesAllowedModels(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(contractsopenai.ChatCompletionResponse{
			ID:      "chatcmpl-forbidden",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "gpt-4",
			Choices: []contractsopenai.Choice{{Index: 0, Message: &contractsopenai.ChatMessage{Role: "assistant", Content: "blocked"}}},
		})
	}))
	defer provider.Close()

	app, _, _, _ := newIntegrationApp(t, logger, memCache, provider.URL)
	createdKey := createVirtualKey(t, app, `{"key_alias":"restricted","models":["gpt-4o-mini"]}`)
	createModel(t, app, fmt.Sprintf(`{"model_name":"gpt-4","provider":"openai","model":"gpt-4","api_base":"%s/v1","api_key":"sk-provider","rpm":50}`, provider.URL))

	resp := performJSONRequest(t, app, http.MethodPost, "http://localhost/v1/completions", "Bearer "+createdKey.Key, `{"model":"gpt-4","prompt":"hello"}`)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)

	var apiErr contractsopenai.APIError
	decodeJSONBody(t, resp, &apiErr)
	assert.Equal(t, "authorization_error", apiErr.Error.Type)
}

func TestIntegration_LegacyCompletionsSetsFinalCostHeaders(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-Id", "provider-req-1")
		_ = json.NewEncoder(w).Encode(contractsopenai.ChatCompletionResponse{
			ID:      "chatcmpl-legacy",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "gpt-4o-mini",
			Choices: []contractsopenai.Choice{{Index: 0, Message: &contractsopenai.ChatMessage{Role: "assistant", Content: "ok"}}},
			Usage: &contractsopenai.Usage{
				PromptTokens:     40,
				CompletionTokens: 20,
				TotalTokens:      60,
			},
		})
	}))
	defer provider.Close()

	app, _, _, _ := newIntegrationApp(t, logger, memCache, provider.URL)
	createdKey := createVirtualKey(t, app, `{"key_alias":"legacy","models":["gpt-4o-mini"]}`)
	createModel(t, app, fmt.Sprintf(`{"model_name":"gpt-4o-mini","provider":"openai","model":"gpt-4o-mini","api_base":"%s/v1","api_key":"sk-provider","rpm":100}`, provider.URL))
	topUpWallet(t, app, `{"amount_irr":2000000}`)

	resp := performJSONRequest(t, app, http.MethodPost, "http://localhost/v1/completions", "Bearer "+createdKey.Key, `{"model":"gpt-4o-mini","prompt":"hello"}`)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, resp.Header.Get("X-Shahrah-Cost-Usd"))
	assert.NotEmpty(t, resp.Header.Get("X-Shahrah-Cost-Irr"))
	assert.Equal(t, "provider-req-1", resp.Header.Get("X-Request-Id"))
}

type integrationTenantStore struct {
	tenant *platform.Tenant
}

func (s integrationTenantStore) GetTenantByHost(ctx context.Context, host string) (*platform.Tenant, error) {
	return s.tenant, nil
}

func (s integrationTenantStore) EnsureTenant(ctx context.Context, slug string) (*platform.Tenant, error) {
	return s.tenant, nil
}

type integrationModelStore struct {
	deployments []platform.Deployment
	users       []platform.TenantAdminUser
	invitations []platform.TenantAdminInvitation
	inviteToken string
}

func (s *integrationModelStore) CreateDeployment(ctx context.Context, dep platform.Deployment) (*platform.Deployment, error) {
	s.deployments = append(s.deployments, dep)
	return &dep, nil
}

func (s *integrationModelStore) ListDeployments(ctx context.Context, tenantID string) ([]platform.Deployment, error) {
	return s.deployments, nil
}

func (s *integrationModelStore) ListTenantAdminUsers(ctx context.Context, tenantSlug string, query string, limit int, offset int) ([]platform.TenantAdminUser, int, error) {
	filtered := make([]platform.TenantAdminUser, 0, len(s.users))
	for _, item := range s.users {
		if item.TenantSlug != tenantSlug {
			continue
		}
		filtered = append(filtered, item)
	}

	total := len(filtered)
	if offset >= total {
		return []platform.TenantAdminUser{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return filtered[offset:end], total, nil
}

func (s *integrationModelStore) InviteTenantAdminUser(ctx context.Context, tenantID string, tenantSlug string, req platform.InviteTenantAdminUserRequest) (*platform.TenantAdminInvitation, string, error) {
	now := time.Now().UTC()
	invitation := platform.TenantAdminInvitation{
		InvitationID: "invite-1",
		TenantID:     tenantID,
		TenantSlug:   tenantSlug,
		Email:        req.Email,
		Mobile:       req.Mobile,
		Role:         req.Role,
		Status:       "pending",
		ExpiresAt:    now.Add(24 * time.Hour),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	s.invitations = append(s.invitations, invitation)
	s.inviteToken = "test-token"
	return &invitation, "https://tenant.app.sharah.ir/invite/accept?token=test-token", nil
}

func (s *integrationModelStore) AcceptTenantAdminInvitation(ctx context.Context, tenantID string, tenantSlug string, token string) (*platform.TenantAdminInvitation, *platform.TenantAdminUser, error) {
	for i := range s.invitations {
		if s.invitations[i].TenantID != tenantID || s.invitations[i].TenantSlug != tenantSlug {
			continue
		}
		if token != s.inviteToken {
			continue
		}
		now := time.Now().UTC()
		s.invitations[i].Status = "accepted"
		s.invitations[i].AcceptedAt = &now
		s.invitations[i].UpdatedAt = now
		user := platform.TenantAdminUser{
			ID:         "accepted-user-1",
			TenantSlug: tenantSlug,
			Email:      s.invitations[i].Email,
			Mobile:     s.invitations[i].Mobile,
			Role:       s.invitations[i].Role,
			Active:     true,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		s.users = append(s.users, user)
		return &s.invitations[i], &user, nil
	}
	return nil, nil, pgx.ErrNoRows
}

func (s *integrationModelStore) SetTenantAdminUserActive(ctx context.Context, tenantSlug string, userID string, active bool) (*platform.TenantAdminUser, error) {
	for i := range s.users {
		if s.users[i].TenantSlug != tenantSlug || s.users[i].ID != userID {
			continue
		}
		if active {
			if s.users[i].Role == "deactivated" {
				s.users[i].Role = "tenant_member"
			}
		} else {
			s.users[i].Role = "deactivated"
		}
		s.users[i].Active = active
		s.users[i].UpdatedAt = time.Now().UTC()
		return &s.users[i], nil
	}
	return nil, pgx.ErrNoRows
}

type integrationWalletStore struct {
	wallet      *platform.Wallet
	rate        *platform.ExchangeRate
	debits      []platform.UsageCostBreakdown
	requestLogs []platform.RequestLogEntry
	spendLogs   []platform.SpendLog
	onSpend     func(tokenHash string, amount float64)
	settings    platform.WalletTopUpSettings
}

func (s *integrationWalletStore) GetWallet(ctx context.Context, tenantID string) (*platform.Wallet, error) {
	return s.wallet, nil
}

func (s *integrationWalletStore) TopUpWallet(ctx context.Context, tenantID string, amountIRR float64, metadata map[string]interface{}) (*platform.Transaction, error) {
	s.wallet.BalanceIRR += amountIRR
	return &platform.Transaction{TenantID: tenantID, AmountIRR: amountIRR, Type: "DEPOSIT", Status: "COMPLETED"}, nil
}

func (s *integrationWalletStore) DebitWallet(ctx context.Context, tenantID string, breakdown platform.UsageCostBreakdown, requestID string) (*platform.Transaction, error) {
	s.debits = append(s.debits, breakdown)
	s.wallet.BalanceIRR -= breakdown.TotalCostIRR
	usd := breakdown.TotalCostUSD
	rate := breakdown.ExchangeRate
	return &platform.Transaction{
		TenantID:     tenantID,
		Type:         "CONSUMPTION",
		AmountIRR:    breakdown.TotalCostIRR,
		AmountUSD:    &usd,
		ExchangeRate: &rate,
		Status:       "COMPLETED",
	}, nil
}

func (s *integrationWalletStore) GetLatestExchangeRate(ctx context.Context) (*platform.ExchangeRate, error) {
	return s.rate, nil
}

func (s *integrationWalletStore) UpsertExchangeRate(ctx context.Context, rate platform.ExchangeRate) (*platform.ExchangeRate, error) {
	s.rate = &rate
	return &rate, nil
}

func (s *integrationWalletStore) GetWalletTopUpSettings(ctx context.Context) (*platform.WalletTopUpSettings, error) {
	return &s.settings, nil
}

func (s *integrationWalletStore) UpdateWalletTopUpSettings(ctx context.Context, minUSD, maxUSD, taxPercent, feePercent float64) (*platform.WalletTopUpSettings, error) {
	s.settings.MinAmountUSD = minUSD
	s.settings.MaxAmountUSD = maxUSD
	s.settings.TaxPercent = taxPercent
	s.settings.FeePercent = feePercent
	s.settings.UpdatedAt = time.Now()
	return &s.settings, nil
}

func (s *integrationWalletStore) CreateWalletTopUpIntent(ctx context.Context, p storage.CreateTopUpIntentParams) (*platform.WalletTopUpIntent, error) {
	return &platform.WalletTopUpIntent{
		ID:           "intent-1",
		PublicID:     "intent-public-1",
		TenantID:     p.TenantID,
		Provider:     p.Provider,
		Status:       platform.WalletTopUpStatusPending,
		AmountUSD:    p.AmountUSD,
		AmountIRR:    p.AmountIRR,
		ExchangeRate: p.ExchangeRate,
		TaxPercent:   p.TaxPercent,
		FeePercent:   p.FeePercent,
		TaxAmountIRR: p.TaxAmountIRR,
		FeeAmountIRR: p.FeeAmountIRR,
		PayableIRR:   p.PayableIRR,
		CallbackURL:  p.CallbackURL,
		ReturnURL:    p.ReturnURL,
		Metadata:     p.Metadata,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

func (s *integrationWalletStore) AttachWalletTopUpGatewayData(ctx context.Context, intentPublicID, authority, paymentURL string, raw map[string]interface{}) error {
	return nil
}

func (s *integrationWalletStore) GetWalletTopUpIntentByAuthority(ctx context.Context, authority string) (*platform.WalletTopUpIntent, error) {
	return &platform.WalletTopUpIntent{
		PublicID:   "intent-public-1",
		ReturnURL:  "https://default.app.sharah.ir/billing/topup/result",
		Authority:  authority,
		Status:     platform.WalletTopUpStatusPending,
		AmountUSD:  10,
		PayableIRR: 8500000,
	}, nil
}

func (s *integrationWalletStore) CompleteWalletTopUpByAuthority(ctx context.Context, authority string, verifyRaw map[string]interface{}, refID, gatewayTxnID, cardMask, payerPhone string) (*platform.WalletTopUpIntent, bool, error) {
	s.wallet.BalanceIRR += 1000
	return &platform.WalletTopUpIntent{
		PublicID:     "intent-public-1",
		ReturnURL:    "https://default.app.sharah.ir/billing/topup/result",
		Authority:    authority,
		Status:       platform.WalletTopUpStatusCompleted,
		RefID:        refID,
		GatewayTxnID: gatewayTxnID,
		CardMask:     cardMask,
		PayerPhone:   payerPhone,
	}, false, nil
}

func (s *integrationWalletStore) MarkWalletTopUpFailedByAuthority(ctx context.Context, authority string, reason string, raw map[string]interface{}) error {
	return nil
}

func (s *integrationWalletStore) ListWalletTransactions(ctx context.Context, tenantID string, limit int) ([]platform.Transaction, error) {
	return []platform.Transaction{}, nil
}

func (s *integrationWalletStore) LogRequest(ctx context.Context, entry platform.RequestLogEntry) error {
	s.requestLogs = append(s.requestLogs, entry)
	return nil
}

func (s *integrationWalletStore) LogSpend(ctx context.Context, log platform.SpendLog) error {
	s.spendLogs = append(s.spendLogs, log)
	if s.onSpend != nil {
		s.onSpend(log.TokenHash, log.CostUSD)
	}
	return nil
}

type noopKeyStore struct{}

func (noopKeyStore) CreateKey(ctx context.Context, params platform.CreateKeyParams) (*platform.CreatedKey, error) {
	return &platform.CreatedKey{}, nil
}

func (noopKeyStore) ListKeys(ctx context.Context, tenantID string) ([]platform.CreatedKey, error) {
	return nil, nil
}

type integrationKeyStore struct {
	mu      sync.RWMutex
	nextID  int
	byHash  map[string]*platform.AuthContext
	created []platform.CreatedKey
}

func newIntegrationKeyStore() *integrationKeyStore {
	return &integrationKeyStore{
		byHash: make(map[string]*platform.AuthContext),
	}
}

func (s *integrationKeyStore) GetKeyByHash(ctx context.Context, hash string) (*platform.AuthContext, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info, ok := s.byHash[hash]
	if !ok {
		return nil, auth.ErrInvalidAPIKey
	}
	copy := *info
	return &copy, nil
}

func (s *integrationKeyStore) CreateKey(ctx context.Context, params platform.CreateKeyParams) (*platform.CreatedKey, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	rawKey := fmt.Sprintf("sk-managed-%03d", s.nextID)
	keyHash := hashToken(rawKey)
	created := platform.CreatedKey{
		ID:        fmt.Sprintf("key-%03d", s.nextID),
		Key:       rawKey,
		KeyAlias:  params.KeyAlias,
		MaskedKey: maskToken(rawKey),
		TenantID:  params.TenantID,
		Role:      params.Role,
		Models:    append([]string(nil), params.Models...),
		MaxBudget: params.MaxBudget,
		RPMLimit:  params.RPMLimit,
		TPMLimit:  params.TPMLimit,
		TeamID:    params.TeamID,
		OrgID:     params.OrgID,
		TokenHash: keyHash,
	}
	s.created = append(s.created, created)
	s.byHash[keyHash] = &platform.AuthContext{
		TenantID:       params.TenantID,
		OrganizationID: params.OrgID,
		TeamID:         params.TeamID,
		KeyID:          created.ID,
		KeyAlias:       params.KeyAlias,
		Role:           params.Role,
		AllowedModels:  append([]string(nil), params.Models...),
		MaxBudget:      params.MaxBudget,
		RPMLimit:       params.RPMLimit,
		TPMLimit:       params.TPMLimit,
		TokenHash:      keyHash,
	}
	return &created, nil
}

func (s *integrationKeyStore) ListKeys(ctx context.Context, tenantID string) ([]platform.CreatedKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]platform.CreatedKey, 0, len(s.created))
	for _, item := range s.created {
		if item.TenantID != tenantID {
			continue
		}
		listItem := item
		listItem.Key = ""
		out = append(out, listItem)
	}
	return out, nil
}

func (s *integrationKeyStore) addSpend(tokenHash string, amount float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if info, ok := s.byHash[tokenHash]; ok {
		info.Spend += amount
	}
	for i := range s.created {
		if s.created[i].TokenHash == tokenHash {
			s.created[i].Spend += amount
		}
	}
}

func newIntegrationApp(t *testing.T, logger *zap.Logger, memCache cache.Cache, providerURL string) (*fiber.App, *integrationKeyStore, *integrationModelStore, *integrationWalletStore) {
	t.Helper()

	silkRoad := httptest.NewServer(silkserver.NewHandler(nil))
	t.Cleanup(silkRoad.Close)
	tunnelServer := httptest.NewServer(tserverserver.NewHandler(silkRoad.URL, nil))
	t.Cleanup(tunnelServer.Close)
	tunnelClient := httptest.NewServer(tclientserver.NewHandler(tunnelServer.URL, nil))
	t.Cleanup(tunnelClient.Close)

	tenantStore := integrationTenantStore{
		tenant: &platform.Tenant{ID: "tenant-1", Slug: "default", Domain: "default", Status: "active", CreatedAt: time.Now()},
	}
	keyStore := newIntegrationKeyStore()
	modelStore := &integrationModelStore{
		users: []platform.TenantAdminUser{
			{
				ID:         "tenant-user-1",
				TenantSlug: "default",
				Email:      "owner@tenant.test",
				Role:       "tenant_admin",
				Active:     true,
				CreatedAt:  time.Now().UTC(),
				UpdatedAt:  time.Now().UTC(),
			},
		},
	}
	walletStore := &integrationWalletStore{
		wallet: &platform.Wallet{ID: "wallet-1", TenantID: "tenant-1", BalanceIRR: 0},
		rate:   &platform.ExchangeRate{ID: "rate-1", USDtoIRR: 850000, Source: "test", FetchedAt: time.Now()},
		settings: platform.WalletTopUpSettings{
			MinAmountUSD: 1,
			MaxAmountUSD: 1000,
			TaxPercent:   0,
			FeePercent:   2,
			UpdatedAt:    time.Now(),
		},
	}

	billingSvc := billing.NewService(memCache, walletStore, 850000, "test")
	llmRouter, err := router.NewRouter(nil, config.RouterSettings{
		RoutingStrategy: "simple-shuffle",
		CooldownTime:    60,
	}, memCache, modelStore, logger)
	require.NoError(t, err)

	app := controlserver.NewFiberApp(&config.Config{}, controlserver.Dependencies{
		AuthMiddleware: auth.NewAuthMiddleware(keyStore, "sk-master", logger),
		TenantResolver: tenantmw.NewMiddleware(tenantStore, "default", true),
		AdminHandlers:  admin.New(keyStore, modelStore, walletStore),
		ProxyHandler: proxy.NewChatCompletionHandler(
			llmRouter,
			tunneltransport.NewHTTPTransport(tunnelClient.URL, 30*time.Second),
			billingSvc,
			walletStore,
			walletStore,
			logger,
		),
		CompletionsHandler: proxy.NewCompletionsHandler(
			llmRouter,
			tunneltransport.NewHTTPTransport(tunnelClient.URL, 30*time.Second),
			billingSvc,
			walletStore,
			walletStore,
			logger,
		),
		ModelsHandler: proxy.NewModelsHandler(modelStore),
		Logger:        logger,
	})

	walletStore.onSpend = keyStore.addSpend
	return app, keyStore, modelStore, walletStore
}

func createVirtualKey(t *testing.T, app *fiber.App, body string) platform.CreatedKey {
	t.Helper()
	resp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/keys", "Bearer sk-master", body)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var out platform.CreatedKey
	decodeJSONBody(t, resp, &out)
	return out
}

func createModel(t *testing.T, app *fiber.App, body string) platform.Deployment {
	t.Helper()
	resp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/models", "Bearer sk-master", body)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var out platform.Deployment
	decodeJSONBody(t, resp, &out)
	return out
}

func topUpWallet(t *testing.T, app *fiber.App, body string) platform.Transaction {
	t.Helper()
	resp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/wallet/topup", "Bearer sk-master", body)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var out platform.Transaction
	decodeJSONBody(t, resp, &out)
	return out
}

func performJSONRequest(t *testing.T, app *fiber.App, method string, url string, authz string, body string) *http.Response {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, url, reader)
	req.Header.Set("Authorization", authz)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := app.Test(req, int((30 * time.Second).Milliseconds()))
	require.NoError(t, err)
	return resp
}

func decodeJSONBody(t *testing.T, resp *http.Response, target interface{}) {
	t.Helper()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(body, target), string(body))
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", sum[:])
}

func maskToken(token string) string {
	if len(token) <= 8 {
		return token
	}
	return token[:5] + "..." + token[len(token)-4:]
}
