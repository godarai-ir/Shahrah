package integration_test

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/admin"
	"github.com/Delpak/shahrah/services/control-plane/internal/apiheaders"
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

	baseProxyDeps := proxy.BaseHandlerDependencies{
		Router:        llmRouter,
		Tunnel:        tunneltransport.NewHTTPTransport(tunnelClient.URL, 30*time.Second),
		Billing:       billingSvc,
		Wallets:       walletStore,
		RequestLogger: walletStore,
		SpendLogger:   walletStore,
		Logger:        logger,
	}
	app := controlserver.NewFiberApp(&config.Config{}, controlserver.Dependencies{
		AuthMiddleware: auth.NewAuthMiddleware(authStore, "sk-master", logger),
		TenantResolver: tenantmw.NewMiddlewareWithCache(tenantStore, nil, tenantmw.MiddlewareConfig{
			AllowHeaderHint: true,
		}),
		AdminHandlers: admin.New(admin.Dependencies{
			Keys:    noopKeyStore{},
			Models:  modelStore,
			Wallets: walletStore,
		}),
		ProxyHandler: proxy.NewChatCompletionHandlerWithDependencies(proxy.ChatCompletionDependencies{
			Base: baseProxyDeps,
		}),
		CompletionsHandler: proxy.NewCompletionsHandlerWithDependencies(baseProxyDeps),
		ResponsesHandler:   proxy.NewResponsesHandlerWithDependencies(baseProxyDeps),
		ModelsHandler:      proxy.NewModelsHandler(modelStore),
		Logger:             logger,
	})

	reqBody := `{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"hi"}]}`
	req := httptest.NewRequest(http.MethodPost, "http://localhost/v1/chat/completions", strings.NewReader(reqBody))
	req.Header.Set("Authorization", "Bearer sk-integration-key")
	req.Header.Set("Content-Type", "application/json")

	resp, requestErr := app.Test(req, int((30 * time.Second).Milliseconds()))
	defer closeResponseBody(t, resp)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, requestErr)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(body))

	var completion contractsopenai.ChatCompletionResponse
	require.NoError(t, json.Unmarshal(body, &completion))
	require.NotEmpty(t, completion.Choices)
	assert.Equal(t, "hello from provider", completion.Choices[0].Message.Content)
	assert.Len(t, walletStore.spendLogs, 1)
	assert.Len(t, walletStore.requestLogs, 1)
	assert.Empty(t, walletStore.debits)
	assert.Equal(t, int64(10_000_000), walletStore.wallet.BalanceIRR)
}

func TestIntegration_EndToEndResponsesStreaming(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/responses", r.URL.Path)
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		require.True(t, ok)
		w.WriteHeader(http.StatusOK)

		for _, frame := range []string{
			"event: response.created\n",
			"data: {\"type\":\"response.created\"}\n\n",
			"data: {\"type\":\"response.output_text.delta\",\"delta\":\"Salam\"}\n\n",
			"data: [DONE]\n\n",
		} {
			_, _ = w.Write([]byte(frame))

			flusher.Flush()
		}
	}))
	defer provider.Close()

	app, _, _, _ := newIntegrationApp(t, logger, memCache, provider.URL)

	createdKey := createVirtualKey(t, app, `{"key_alias":"stream-key","models":["openai:default:gpt-4o-mini"]}`)
	createModel(t, app, fmt.Sprintf(`{"model_name":"openai:default:gpt-4o-mini","provider":"openai","model":"gpt-4o-mini","api_base":"%s/v1","api_key":"sk-provider","rpm":250}`, provider.URL))
	topUpWallet(t, app, `{"amount_irr":1500000}`)

	req := httptest.NewRequest(http.MethodPost, "http://localhost/v1/responses", strings.NewReader(`{"model":"openai:default:gpt-4o-mini","input":"hello","stream":true}`))
	req.Header.Set("Authorization", "Bearer "+createdKey.Key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	start := time.Now()

	resp, err := app.Test(req, int((30 * time.Second).Milliseconds()))
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Less(t, time.Since(start), 5*time.Second)
	assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	bodyText := string(body)
	assert.Contains(t, bodyText, "event: response.created")
	assert.Contains(t, bodyText, `response.output_text.delta`)
	assert.Contains(t, bodyText, "data: [DONE]")
	assert.NotContains(t, strings.ToLower(bodyText), "empty reply")
	assert.NotContains(t, strings.ToLower(bodyText), "socket hang up")
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

	createdKey := createVirtualKey(t, app, `{"key_alias":"tenant-key","models":["openai:default:gpt-4o-mini"]}`)
	assert.NotEmpty(t, createdKey.Key)
	assert.Equal(t, "tenant-key", createdKey.KeyAlias)

	createModel(t, app, fmt.Sprintf(`{"model_name":"openai:default:gpt-4o-mini","provider":"openai","model":"gpt-4o-mini","api_base":"%s/v1","api_key":"sk-provider","rpm":250}`, provider.URL))
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
	assert.Equal(t, "openai:default:gpt-4o-mini", modelList.Data[0].ID)
	assert.Contains(t, modelList.Data[0].Aliases, "gpt-4o-mini")

	completionResp := performJSONRequest(t, app, http.MethodPost, "http://localhost/v1/chat/completions", "Bearer "+createdKey.Key, `{"model":"openai:default:gpt-4o-mini","messages":[{"role":"user","content":"hello"}]}`)
	assert.Equal(t, http.StatusOK, completionResp.StatusCode)

	var completion contractsopenai.ChatCompletionResponse
	decodeJSONBody(t, completionResp, &completion)
	assert.Equal(t, "provisioned flow succeeded", completion.Choices[0].Message.Content)

	require.Empty(t, walletStore.debits)
	require.Len(t, walletStore.requestLogs, 1)
	require.Len(t, walletStore.spendLogs, 1)
	assert.Greater(t, walletStore.wallet.BalanceIRR, int64(0))

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

	createdKey := createVirtualKey(t, app, `{"key_alias":"low-balance","models":["openai:default:gpt-4"]}`)
	createModel(t, app, fmt.Sprintf(`{"model_name":"openai:default:gpt-4","provider":"openai","model":"gpt-4","api_base":"%s/v1","api_key":"sk-provider","rpm":50}`, provider.URL))
	topUpWallet(t, app, `{"amount_irr":1}`)

	resp := performJSONRequest(t, app, http.MethodPost, "http://localhost/v1/chat/completions", "Bearer "+createdKey.Key, `{"model":"openai:default:gpt-4","messages":[{"role":"user","content":"expensive request"}]}`)
	assert.Equal(t, http.StatusPaymentRequired, resp.StatusCode)

	var apiErr contractsopenai.APIError
	decodeJSONBody(t, resp, &apiErr)
	assert.Contains(t, apiErr.Error.Type, "budget")
	assert.Empty(t, walletStore.debits)
	assert.Empty(t, walletStore.requestLogs)
	assert.Empty(t, walletStore.spendLogs)
	assert.Equal(t, int64(1), walletStore.wallet.BalanceIRR)
}

func TestIntegration_CreateServiceAccountKeyRequiresTeamID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer provider.Close()

	app, _, _, _ := newIntegrationApp(t, logger, memCache, provider.URL)

	resp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/keys", "Bearer sk-master", `{"key_alias":"svc-no-team","models":["openai:default:gpt-4o-mini"],"role":"service_account"}`)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestIntegration_CreateServiceAccountKey(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer provider.Close()

	app, _, _, _ := newIntegrationApp(t, logger, memCache, provider.URL)

	createdKey := createVirtualKey(t, app, `{"key_alias":"svc-team-key","models":["openai:default:gpt-4o-mini"],"role":"service_account","team_id":"team-ops","organization_id":"org-platform"}`)
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

func TestIntegration_CreateVATKeyViaCanonicalEndpoint(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer provider.Close()

	app, _, _, _ := newIntegrationApp(t, logger, memCache, provider.URL)

	resp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/keys/vat", "Bearer sk-master", `{"key_alias":"svc-team-key","allowed_models":["openai:default:gpt-4o-mini"],"team_id":"team-ops","org_unit_id":"org-platform"}`)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var out platform.CreatedKey
	decodeJSONBody(t, resp, &out)
	assert.Equal(t, platform.KeyTypeVAT, out.KeyType)
	assert.Equal(t, platform.LegacyKeyRoleServiceAccount, out.Role)
	assert.Equal(t, "team-ops", out.TeamID)
	assert.Equal(t, "org-platform", out.OrgUnitID)
	assert.Equal(t, "org-platform", out.OrgID)
	assert.Equal(t, []string{"openai:default:gpt-4o-mini"}, out.AllowedModels)
}

func TestIntegration_VirtualModelCRUDSupportsWeightedStrategy(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer provider.Close()

	app, _, _, _ := newIntegrationApp(t, logger, memCache, provider.URL)

	created := createVirtualModel(t, app, `{
		"alias":"smart-router",
		"routing_strategy":"weighted",
		"targets":[
			{"order":1,"weight":60,"provider":"openai","model":"gpt-4o-primary","policy":{"retry":{"attempts":2,"delay_ms":100,"status_codes":[429,500,502,503]},"fallback":{"status_codes":[429,500,502,503,504],"candidate":true}}},
			{"order":2,"weight":40,"provider":"openai","model":"gpt-4o-backup","policy":{"retry":{"attempts":2,"delay_ms":100,"status_codes":[429,500,502,503]},"fallback":{"status_codes":[429,500,502,503,504],"candidate":true}}}
		]
	}`)
	assert.Equal(t, "weighted", created.RoutingStrategy)
	require.Len(t, created.Targets, 2)
	assert.Equal(t, 60.0, created.Targets[0].Weight)
	assert.Equal(t, 40.0, created.Targets[1].Weight)

	getResp := performJSONRequest(t, app, http.MethodGet, "http://localhost/api/v1/virtual-models/"+created.ID, "Bearer sk-master", "")
	assert.Equal(t, http.StatusOK, getResp.StatusCode)

	var fetched platform.VirtualModel
	decodeJSONBody(t, getResp, &fetched)
	assert.Equal(t, "weighted", fetched.RoutingStrategy)
	assert.Equal(t, 60.0, fetched.Targets[0].Weight)

	updateResp := performJSONRequest(t, app, http.MethodPatch, "http://localhost/api/v1/virtual-models/"+created.ID, "Bearer sk-master", `{
		"routing_strategy":"weighted",
		"targets":[
			{"order":1,"weight":55,"provider":"openai","model":"gpt-4o-primary","policy":{"retry":{"attempts":2,"delay_ms":100,"status_codes":[429,500,502,503]},"fallback":{"status_codes":[429,500,502,503,504],"candidate":true}}},
			{"order":2,"weight":45,"provider":"openai","model":"gpt-4o-backup","policy":{"retry":{"attempts":2,"delay_ms":100,"status_codes":[429,500,502,503]},"fallback":{"status_codes":[429,500,502,503,504],"candidate":true}}}
		]
	}`)
	assert.Equal(t, http.StatusOK, updateResp.StatusCode)

	var updated platform.VirtualModel
	decodeJSONBody(t, updateResp, &updated)
	assert.Equal(t, 55.0, updated.Targets[0].Weight)
	assert.Equal(t, 45.0, updated.Targets[1].Weight)
}

func TestIntegration_WeightedVirtualModelFallsBackInOrderAfterWeightedPrimary(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req contractsopenai.ChatCompletionRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))

		w.Header().Set("Content-Type", "application/json")

		if req.Model == "gpt-4o-primary" {
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "rate limited"})

			return
		}

		_ = json.NewEncoder(w).Encode(contractsopenai.ChatCompletionResponse{
			ID:      "chatcmpl-weighted",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   req.Model,
			Choices: []contractsopenai.Choice{{
				Index: 0,
				Message: &contractsopenai.ChatMessage{
					Role:    "assistant",
					Content: "backup selected",
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

	app, _, modelStore, _ := newIntegrationApp(t, logger, memCache, provider.URL)

	createModel(t, app, fmt.Sprintf(`{"model_name":"gpt-4o-primary","provider":"openai","model":"gpt-4o-primary","api_base":"%s/v1","api_key":"sk-provider","rpm":250}`, provider.URL))
	createModel(t, app, fmt.Sprintf(`{"model_name":"gpt-4o-backup","provider":"openai","model":"gpt-4o-backup","api_base":"%s/v1","api_key":"sk-provider","rpm":250}`, provider.URL))
	topUpWallet(t, app, `{"amount_irr":1500000}`)

	modelStore.virtualModels = append(modelStore.virtualModels, platform.VirtualModel{
		ID:              "vm-weighted",
		TenantID:        "tenant-1",
		Alias:           "smart-router",
		RoutingStrategy: "weighted",
		Enabled:         true,
		Targets: []platform.VirtualModelTarget{
			{
				Order: 1, Weight: 100, Provider: "openai", Model: "gpt-4o-primary",
				Policy: platform.VirtualModelPolicy{
					Retry:    platform.VirtualModelRetryPolicy{Attempts: 1, DelayMS: 0, StatusCodes: []int{429, 500, 502, 503}},
					Fallback: platform.VirtualModelFallbackPolicy{StatusCodes: []int{429, 500, 502, 503, 504}, Candidate: true},
				},
			},
			{
				Order: 2, Weight: 0, Provider: "openai", Model: "gpt-4o-backup",
				Policy: platform.VirtualModelPolicy{
					Retry:    platform.VirtualModelRetryPolicy{Attempts: 1, DelayMS: 0, StatusCodes: []int{429, 500, 502, 503}},
					Fallback: platform.VirtualModelFallbackPolicy{StatusCodes: []int{429, 500, 502, 503, 504}, Candidate: true},
				},
			},
		},
	})

	resp := performJSONRequest(t, app, http.MethodPost, "http://localhost/v1/chat/completions", "Bearer sk-master", `{"model":"smart-router","messages":[{"role":"user","content":"hello"}]}`)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var completion contractsopenai.ChatCompletionResponse
	decodeJSONBody(t, resp, &completion)
	assert.Equal(t, "backup selected", completion.Choices[0].Message.Content)
	assert.Equal(t, "gpt-4o-backup", completion.Model)
}

func TestIntegration_PersonalAccessTokenLifecycle(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer provider.Close()

	app, keyStore, _, _ := newIntegrationApp(t, logger, memCache, provider.URL)

	createResp := performJSONRequestWithHeaders(
		t,
		app,
		http.MethodPost,
		"http://localhost/api/v1/personal-access-tokens",
		"Bearer sk-master",
		map[string]string{"x-godarai-user-id": "tenant-user-1"},
		`{"name":"ci-token","expires_in":"30d"}`,
	)
	assert.Equal(t, http.StatusCreated, createResp.StatusCode)

	var created platform.CreatedKey
	decodeJSONBody(t, createResp, &created)
	assert.NotEmpty(t, created.Key)
	assert.Equal(t, platform.KeyTypePAT, created.KeyType)

	listResp := performJSONRequestWithHeaders(
		t,
		app,
		http.MethodGet,
		"http://localhost/api/v1/personal-access-tokens?limit=25&offset=0",
		"Bearer sk-master",
		map[string]string{"x-godarai-user-id": "tenant-user-1"},
		"",
	)
	assert.Equal(t, http.StatusOK, listResp.StatusCode)

	var listed platform.ListPATResponse
	decodeJSONBody(t, listResp, &listed)
	require.Len(t, listed.Data, 1)
	assert.Empty(t, listed.Data[0].Key)

	revokeResp := performJSONRequestWithHeaders(
		t,
		app,
		http.MethodDelete,
		"http://localhost/api/v1/personal-access-tokens/"+created.ID,
		"Bearer sk-master",
		map[string]string{"x-godarai-user-id": "tenant-user-1"},
		"",
	)
	assert.Equal(t, http.StatusNoContent, revokeResp.StatusCode)

	revokedAuthInfo, err := keyStore.GetKeyByHash(context.Background(), hashToken(created.Key))
	require.NoError(t, err)
	assert.True(t, revokedAuthInfo.Blocked)

	for _, name := range []string{"all-token-a", "all-token-b"} {
		resp := performJSONRequestWithHeaders(
			t,
			app,
			http.MethodPost,
			"http://localhost/api/v1/personal-access-tokens",
			"Bearer sk-master",
			map[string]string{"x-godarai-user-id": "tenant-user-1"},
			fmt.Sprintf(`{"name":%q}`, name),
		)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	revokeAllResp := performJSONRequestWithHeaders(
		t,
		app,
		http.MethodDelete,
		"http://localhost/api/v1/personal-access-tokens/revoke/all",
		"Bearer sk-master",
		map[string]string{"x-godarai-user-id": "tenant-user-1"},
		"",
	)
	assert.Equal(t, http.StatusOK, revokeAllResp.StatusCode)

	var revokeAll platform.RevokeAllPATResponse
	decodeJSONBody(t, revokeAllResp, &revokeAll)
	assert.EqualValues(t, 2, revokeAll.Revoked)

	listAfterAllResp := performJSONRequestWithHeaders(
		t,
		app,
		http.MethodGet,
		"http://localhost/api/v1/personal-access-tokens?limit=25&offset=0",
		"Bearer sk-master",
		map[string]string{"x-godarai-user-id": "tenant-user-1"},
		"",
	)
	assert.Equal(t, http.StatusOK, listAfterAllResp.StatusCode)

	var listedAfterAll platform.ListPATResponse
	decodeJSONBody(t, listAfterAllResp, &listedAfterAll)
	assert.Empty(t, listedAfterAll.Data)
}

func TestIntegration_InviteTenantUserViaCanonicalInvitationEndpoint(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer provider.Close()

	app, _, _, _ := newIntegrationApp(t, logger, memCache, provider.URL)
	resp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/users/invitations", "Bearer sk-master", `{"email":"new-user@tenant.test","role":"member"}`)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var out platform.InviteTenantAdminUserResponse
	decodeJSONBody(t, resp, &out)
	assert.Equal(t, "new-user@tenant.test", out.Invitation.Email)
	assert.Equal(t, "member", out.Invitation.Role)
}

func TestIntegration_OrgUnitCRUD(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer provider.Close()

	app, _, _, _ := newIntegrationApp(t, logger, memCache, provider.URL)

	createResp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/org-units", "Bearer sk-master", `{"alias":"engineering","display_name":"Engineering"}`)
	assert.Equal(t, http.StatusCreated, createResp.StatusCode)

	var created platform.OrgUnit
	decodeJSONBody(t, createResp, &created)
	assert.Equal(t, "engineering", created.Alias)
	assert.Equal(t, "Engineering", created.DisplayName)

	listResp := performJSONRequest(t, app, http.MethodGet, "http://localhost/api/v1/org-units?limit=20&offset=0", "Bearer sk-master", "")
	assert.Equal(t, http.StatusOK, listResp.StatusCode)

	var listed platform.OrgUnitListResponse
	decodeJSONBody(t, listResp, &listed)
	require.Len(t, listed.Data, 1)

	updateResp := performJSONRequest(t, app, http.MethodPatch, "http://localhost/api/v1/org-units/"+created.ID, "Bearer sk-master", `{"display_name":"Platform Engineering"}`)
	assert.Equal(t, http.StatusOK, updateResp.StatusCode)

	var updated platform.OrgUnit
	decodeJSONBody(t, updateResp, &updated)
	assert.Equal(t, "Platform Engineering", updated.DisplayName)

	deleteResp := performJSONRequest(t, app, http.MethodDelete, "http://localhost/api/v1/org-units/"+created.ID, "Bearer sk-master", "")
	assert.Equal(t, http.StatusNoContent, deleteResp.StatusCode)
}

func TestIntegration_VirtualUserCRUD(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer provider.Close()

	app, _, _, _ := newIntegrationApp(t, logger, memCache, provider.URL)

	createResp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/virtual-users", "Bearer sk-master", `{"alias":"ci-cd","display_name":"CI/CD Service","team_id":"team-ops"}`)
	assert.Equal(t, http.StatusCreated, createResp.StatusCode)

	var created platform.VirtualUser
	decodeJSONBody(t, createResp, &created)
	assert.Equal(t, "ci-cd", created.Alias)
	assert.Equal(t, "team-ops", created.TeamID)

	getResp := performJSONRequest(t, app, http.MethodGet, "http://localhost/api/v1/virtual-users/"+created.ID, "Bearer sk-master", "")
	assert.Equal(t, http.StatusOK, getResp.StatusCode)

	updateResp := performJSONRequest(t, app, http.MethodPatch, "http://localhost/api/v1/virtual-users/"+created.ID, "Bearer sk-master", `{"display_name":"CI Service","active":false}`)
	assert.Equal(t, http.StatusOK, updateResp.StatusCode)

	var updated platform.VirtualUser
	decodeJSONBody(t, updateResp, &updated)
	assert.Equal(t, "CI Service", updated.DisplayName)
	assert.False(t, updated.Active)

	listResp := performJSONRequest(t, app, http.MethodGet, "http://localhost/api/v1/virtual-users?limit=20&offset=0", "Bearer sk-master", "")
	assert.Equal(t, http.StatusOK, listResp.StatusCode)

	var listed platform.VirtualUserListResponse
	decodeJSONBody(t, listResp, &listed)
	require.Len(t, listed.Data, 1)

	deleteResp := performJSONRequest(t, app, http.MethodDelete, "http://localhost/api/v1/virtual-users/"+created.ID, "Bearer sk-master", "")
	assert.Equal(t, http.StatusNoContent, deleteResp.StatusCode)
}

func TestIntegration_VirtualUserTokenLifecycle(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer provider.Close()

	app, keyStore, _, _ := newIntegrationApp(t, logger, memCache, provider.URL)

	vuResp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/virtual-users", "Bearer sk-master", `{"alias":"deploy-bot","display_name":"Deploy Bot","team_id":"team-ops","org_unit_id":"org-platform"}`)
	assert.Equal(t, http.StatusCreated, vuResp.StatusCode)

	var virtualUser platform.VirtualUser
	decodeJSONBody(t, vuResp, &virtualUser)

	createResp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/virtual-users/"+virtualUser.ID+"/tokens", "Bearer sk-master", `{"key_alias":"deploy-token","allowed_models":["openai:default:gpt-4o-mini"]}`)
	assert.Equal(t, http.StatusCreated, createResp.StatusCode)

	var created platform.CreatedKey
	decodeJSONBody(t, createResp, &created)
	assert.NotEmpty(t, created.Key)
	assert.Equal(t, platform.KeyTypeVAT, created.KeyType)
	assert.Equal(t, virtualUser.ID, created.VirtualUserID)
	assert.Equal(t, "team-ops", created.TeamID)
	assert.Equal(t, "org-platform", created.OrgUnitID)

	authInfo, err := keyStore.GetKeyByHash(context.Background(), hashToken(created.Key))
	require.NoError(t, err)
	assert.Equal(t, virtualUser.ID, authInfo.VirtualUserID)
	assert.False(t, authInfo.Blocked)

	listResp := performJSONRequest(t, app, http.MethodGet, "http://localhost/api/v1/virtual-users/"+virtualUser.ID+"/tokens?limit=25&offset=0", "Bearer sk-master", "")
	assert.Equal(t, http.StatusOK, listResp.StatusCode)

	var listed platform.ListVATTokensResponse
	decodeJSONBody(t, listResp, &listed)
	require.Len(t, listed.Data, 1)
	assert.Empty(t, listed.Data[0].Key)
	assert.Equal(t, created.ID, listed.Data[0].ID)

	rotateResp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/virtual-users/"+virtualUser.ID+"/tokens/"+created.ID+"/rotate", "Bearer sk-master", `{}`)
	assert.Equal(t, http.StatusCreated, rotateResp.StatusCode)

	var rotated platform.CreatedKey
	decodeJSONBody(t, rotateResp, &rotated)
	assert.NotEmpty(t, rotated.Key)
	assert.NotEqual(t, created.Key, rotated.Key)
	assert.Equal(t, virtualUser.ID, rotated.VirtualUserID)

	oldAuthInfo, err := keyStore.GetKeyByHash(context.Background(), hashToken(created.Key))
	require.NoError(t, err)
	assert.True(t, oldAuthInfo.Blocked)

	newAuthInfo, err := keyStore.GetKeyByHash(context.Background(), hashToken(rotated.Key))
	require.NoError(t, err)
	assert.Equal(t, virtualUser.ID, newAuthInfo.VirtualUserID)
	assert.False(t, newAuthInfo.Blocked)

	revokeResp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/virtual-users/"+virtualUser.ID+"/tokens/"+rotated.ID+"/revoke", "Bearer sk-master", `{"reason":"compromised"}`)
	assert.Equal(t, http.StatusNoContent, revokeResp.StatusCode)

	revokedAuthInfo, err := keyStore.GetKeyByHash(context.Background(), hashToken(rotated.Key))
	require.NoError(t, err)
	assert.True(t, revokedAuthInfo.Blocked)
}

func TestIntegration_OrgUnitMembershipCRUD(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer provider.Close()

	app, _, _, _ := newIntegrationApp(t, logger, memCache, provider.URL)
	orgResp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/org-units", "Bearer sk-master", `{"alias":"engineering"}`)
	assert.Equal(t, http.StatusCreated, orgResp.StatusCode)

	var org platform.OrgUnit
	decodeJSONBody(t, orgResp, &org)

	addResp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/org-units/"+org.ID+"/members", "Bearer sk-master", `{"user_id":"user-1","role":"org_admin"}`)
	assert.Equal(t, http.StatusCreated, addResp.StatusCode)

	var added platform.OrgUnitMembership
	decodeJSONBody(t, addResp, &added)
	assert.Equal(t, "org_admin", added.Role)

	listResp := performJSONRequest(t, app, http.MethodGet, "http://localhost/api/v1/org-units/"+org.ID+"/members", "Bearer sk-master", "")
	assert.Equal(t, http.StatusOK, listResp.StatusCode)

	var listed platform.OrgUnitMembershipListResponse
	decodeJSONBody(t, listResp, &listed)
	require.Len(t, listed.Data, 1)

	updateResp := performJSONRequest(t, app, http.MethodPatch, "http://localhost/api/v1/org-units/"+org.ID+"/members/user-1", "Bearer sk-master", `{"role":"member"}`)
	assert.Equal(t, http.StatusOK, updateResp.StatusCode)

	var updated platform.OrgUnitMembership
	decodeJSONBody(t, updateResp, &updated)
	assert.Equal(t, "member", updated.Role)

	deleteResp := performJSONRequest(t, app, http.MethodDelete, "http://localhost/api/v1/org-units/"+org.ID+"/members/user-1", "Bearer sk-master", "")
	assert.Equal(t, http.StatusNoContent, deleteResp.StatusCode)
}

func TestIntegration_TeamMembershipCRUD(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer provider.Close()

	app, _, _, _ := newIntegrationApp(t, logger, memCache, provider.URL)

	addResp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/teams/team-ops/members", "Bearer sk-master", `{"user_id":"user-1","role":"team_member","team_member_permissions":["/key/generate"]}`)
	assert.Equal(t, http.StatusCreated, addResp.StatusCode)

	var added platform.TeamMembership
	decodeJSONBody(t, addResp, &added)
	assert.Equal(t, "team_member", added.Role)
	assert.Equal(t, []string{"/key/generate"}, added.TeamMemberPermissions)

	listResp := performJSONRequest(t, app, http.MethodGet, "http://localhost/api/v1/teams/team-ops/members", "Bearer sk-master", "")
	assert.Equal(t, http.StatusOK, listResp.StatusCode)

	var listed platform.TeamMembershipListResponse
	decodeJSONBody(t, listResp, &listed)
	require.Len(t, listed.Data, 1)

	updateResp := performJSONRequest(t, app, http.MethodPatch, "http://localhost/api/v1/teams/team-ops/members/user-1", "Bearer sk-master", `{"role":"team_admin","team_member_permissions":["/key/generate","/key/update"]}`)
	assert.Equal(t, http.StatusOK, updateResp.StatusCode)

	var updated platform.TeamMembership
	decodeJSONBody(t, updateResp, &updated)
	assert.Equal(t, "team_admin", updated.Role)
	assert.Equal(t, []string{"/key/generate", "/key/update"}, updated.TeamMemberPermissions)

	deleteResp := performJSONRequest(t, app, http.MethodDelete, "http://localhost/api/v1/teams/team-ops/members/user-1", "Bearer sk-master", "")
	assert.Equal(t, http.StatusNoContent, deleteResp.StatusCode)
}

func TestIntegration_RBACDeniesTeamMemberOnVirtualUsers(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer provider.Close()

	app, keyStore, _, _ := newIntegrationApp(t, logger, memCache, provider.URL)
	teamMemberHash := hashToken("sk-team-member")
	keyStore.byHash[teamMemberHash] = &platform.AuthContext{
		TenantID:  "tenant-1",
		KeyID:     "tm-1",
		KeyAlias:  "team-member",
		Role:      "team_member",
		TokenHash: teamMemberHash,
	}

	resp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/virtual-users", "Bearer sk-team-member", `{"alias":"blocked-user"}`)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestIntegration_RBACAllowsTeamAdminOnTeamMembers(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer provider.Close()

	app, keyStore, _, _ := newIntegrationApp(t, logger, memCache, provider.URL)
	teamAdminHash := hashToken("sk-team-admin")
	keyStore.byHash[teamAdminHash] = &platform.AuthContext{
		TenantID:  "tenant-1",
		KeyID:     "ta-1",
		KeyAlias:  "team-admin",
		Role:      "team_admin",
		TokenHash: teamAdminHash,
	}

	resp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/teams/team-ops/members", "Bearer sk-team-admin", `{"user_id":"user-2","role":"team_member"}`)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
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
	resp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/users/invite", "Bearer sk-master", `{"email":"new-user@tenant.test","role":"member"}`)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var out platform.InviteTenantAdminUserResponse
	decodeJSONBody(t, resp, &out)
	assert.Equal(t, "new-user@tenant.test", out.Invitation.Email)
	assert.Equal(t, "member", out.Invitation.Role)
	assert.NotEmpty(t, out.Link)
}

func TestIntegration_AcceptTenantUserInvitation(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer provider.Close()

	app, _, _, _ := newIntegrationApp(t, logger, memCache, provider.URL)

	inviteResp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/users/invite", "Bearer sk-master", `{"email":"accept-me@tenant.test","role":"member"}`)
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

	deactivateResp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/users/tenant-user-2/deactivate", "Bearer sk-master", "")
	assert.Equal(t, http.StatusOK, deactivateResp.StatusCode)

	var deactivated platform.TenantAdminUser
	decodeJSONBody(t, deactivateResp, &deactivated)
	assert.False(t, deactivated.Active)
	assert.Equal(t, "member", deactivated.Role)

	activateResp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/users/tenant-user-2/activate", "Bearer sk-master", "")
	assert.Equal(t, http.StatusOK, activateResp.StatusCode)

	var activated platform.TenantAdminUser
	decodeJSONBody(t, activateResp, &activated)
	assert.True(t, activated.Active)
	assert.Equal(t, "member", activated.Role)
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
	createdKey := createVirtualKey(t, app, `{"key_alias":"restricted","models":["openai:default:gpt-4o-mini"]}`)
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
	createdKey := createVirtualKey(t, app, `{"key_alias":"legacy","models":["openai:default:gpt-4o-mini"]}`)
	createModel(t, app, fmt.Sprintf(`{"model_name":"openai:default:gpt-4o-mini","provider":"openai","model":"gpt-4o-mini","api_base":"%s/v1","api_key":"sk-provider","rpm":100}`, provider.URL))
	topUpWallet(t, app, `{"amount_irr":2000000}`)

	resp := performJSONRequest(t, app, http.MethodPost, "http://localhost/v1/completions", "Bearer "+createdKey.Key, `{"model":"openai:default:gpt-4o-mini","prompt":"hello"}`)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, resp.Header.Get(apiheaders.HeaderCostUSD))
	assert.NotEmpty(t, resp.Header.Get(apiheaders.HeaderCostIRR))
	assert.Equal(t, "provider-req-1", resp.Header.Get("X-Request-Id"))
}

type integrationTenantStore struct {
	tenant *platform.Tenant
}

func (s integrationTenantStore) GetTenantByHost(ctx context.Context, host string) (*platform.Tenant, error) {
	return s.tenant, nil
}

type integrationModelStore struct {
	deployments   []platform.Deployment
	virtualModels []platform.VirtualModel
	users         []platform.TenantAdminUser
	invitations   []platform.TenantAdminInvitation
	inviteToken   string
}

func (s *integrationModelStore) CreateDeployment(ctx context.Context, dep platform.Deployment) (*platform.Deployment, error) {
	s.deployments = append(s.deployments, dep)
	return &dep, nil
}

func (s *integrationModelStore) ListDeployments(ctx context.Context, tenantID string) ([]platform.Deployment, error) {
	return s.deployments, nil
}

func (s *integrationModelStore) CreateVirtualModel(ctx context.Context, tenantID string, req platform.CreateVirtualModelRequest, actor string) (*platform.VirtualModel, error) {
	vm := platform.VirtualModel{
		ID:              fmt.Sprintf("vm-%d", len(s.virtualModels)+1),
		TenantID:        tenantID,
		Alias:           req.Alias,
		RoutingStrategy: req.RoutingStrategy,
		Enabled:         req.Enabled == nil || *req.Enabled,
		Version:         1,
		Targets:         req.Targets,
		CreatedBy:       actor,
		UpdatedBy:       actor,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
	s.virtualModels = append(s.virtualModels, vm)

	return &vm, nil
}

func (s *integrationModelStore) ListVirtualModels(ctx context.Context, tenantID string) ([]platform.VirtualModel, error) {
	out := make([]platform.VirtualModel, 0, len(s.virtualModels))
	for _, vm := range s.virtualModels {
		if vm.TenantID == tenantID {
			out = append(out, vm)
		}
	}

	return out, nil
}

func (s *integrationModelStore) ListVirtualModelsForPrincipal(ctx context.Context, tenantID string, userID string, teamID string) ([]platform.VirtualModel, error) {
	return s.ListVirtualModels(ctx, tenantID)
}

func (s *integrationModelStore) GetVirtualModel(ctx context.Context, tenantID string, id string) (*platform.VirtualModel, error) {
	for _, vm := range s.virtualModels {
		if vm.TenantID == tenantID && vm.ID == id {
			cloned := vm
			return &cloned, nil
		}
	}

	return nil, storage.ErrVirtualModelNotFound
}

func (s *integrationModelStore) UpdateVirtualModel(ctx context.Context, tenantID string, id string, req platform.UpdateVirtualModelRequest, actor string) (*platform.VirtualModel, error) {
	for i := range s.virtualModels {
		vm := &s.virtualModels[i]
		if vm.TenantID != tenantID || vm.ID != id {
			continue
		}

		if req.Alias != nil {
			vm.Alias = *req.Alias
		}

		if req.RoutingStrategy != nil {
			vm.RoutingStrategy = *req.RoutingStrategy
		}

		if req.Enabled != nil {
			vm.Enabled = *req.Enabled
		}

		if len(req.Targets) > 0 {
			vm.Targets = req.Targets
		}

		vm.UpdatedBy = actor
		vm.Version++
		vm.UpdatedAt = time.Now().UTC()
		cloned := *vm

		return &cloned, nil
	}

	return nil, storage.ErrVirtualModelNotFound
}

func (s *integrationModelStore) GetVirtualModelByAlias(ctx context.Context, tenantID string, alias string) (*platform.VirtualModel, error) {
	for _, vm := range s.virtualModels {
		if vm.TenantID == tenantID && vm.Alias == alias && vm.Enabled {
			cloned := vm
			return &cloned, nil
		}
	}

	return nil, storage.ErrVirtualModelNotFound
}

func (s *integrationModelStore) DeleteVirtualModel(ctx context.Context, tenantID string, id string) error {
	for i := range s.virtualModels {
		vm := s.virtualModels[i]
		if vm.TenantID == tenantID && vm.ID == id {
			s.virtualModels = append(s.virtualModels[:i], s.virtualModels[i+1:]...)
			return nil
		}
	}

	return storage.ErrVirtualModelNotFound
}

func (s *integrationModelStore) CreateRateLimitRule(ctx context.Context, tenantID string, req platform.CreateRateLimitRuleRequest) (*platform.RateLimitRule, error) {
	return nil, errors.New("not implemented")
}

func (s *integrationModelStore) ListRateLimitRules(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.RateLimitRule, int, error) {
	return []platform.RateLimitRule{}, 0, nil
}

func (s *integrationModelStore) GetRateLimitRuleByRuleID(ctx context.Context, tenantID string, ruleID string) (*platform.RateLimitRule, error) {
	return nil, pgx.ErrNoRows
}

func (s *integrationModelStore) UpdateRateLimitRuleByRuleID(ctx context.Context, tenantID string, ruleID string, req platform.UpdateRateLimitRuleRequest) (*platform.RateLimitRule, error) {
	return nil, pgx.ErrNoRows
}

func (s *integrationModelStore) UpdateRateLimitRuleStatusByRuleID(ctx context.Context, tenantID string, ruleID string, enabled bool) (*platform.RateLimitRule, error) {
	return nil, pgx.ErrNoRows
}

func (s *integrationModelStore) DeleteRateLimitRuleByRuleID(ctx context.Context, tenantID string, ruleID string) error {
	return pgx.ErrNoRows
}

func (s *integrationModelStore) CreateBudgetLimitRule(ctx context.Context, tenantID string, req platform.CreateBudgetLimitRuleRequest) (*platform.BudgetLimitRule, error) {
	return nil, errors.New("not implemented")
}

func (s *integrationModelStore) ListBudgetLimitRules(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.BudgetLimitRule, int, error) {
	return []platform.BudgetLimitRule{}, 0, nil
}

func (s *integrationModelStore) GetBudgetLimitRuleByRuleID(ctx context.Context, tenantID string, ruleID string) (*platform.BudgetLimitRule, error) {
	return nil, pgx.ErrNoRows
}

func (s *integrationModelStore) UpdateBudgetLimitRuleByRuleID(ctx context.Context, tenantID string, ruleID string, req platform.UpdateBudgetLimitRuleRequest) (*platform.BudgetLimitRule, error) {
	return nil, pgx.ErrNoRows
}

func (s *integrationModelStore) UpdateBudgetLimitRuleStatusByRuleID(ctx context.Context, tenantID string, ruleID string, enabled bool) (*platform.BudgetLimitRule, error) {
	return nil, pgx.ErrNoRows
}

func (s *integrationModelStore) DeleteBudgetLimitRuleByRuleID(ctx context.Context, tenantID string, ruleID string) error {
	return pgx.ErrNoRows
}

func (s *integrationModelStore) CreateTenantGuardrail(ctx context.Context, tenantID string, req platform.CreateTenantGuardrailRequest, actor string) (*platform.TenantGuardrail, error) {
	return nil, errors.New("not implemented")
}

func (s *integrationModelStore) ListTenantGuardrails(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.TenantGuardrail, int, error) {
	return []platform.TenantGuardrail{}, 0, nil
}

func (s *integrationModelStore) GetTenantGuardrailBySlug(ctx context.Context, tenantID string, slug string) (*platform.TenantGuardrail, error) {
	return nil, pgx.ErrNoRows
}

func (s *integrationModelStore) UpdateTenantGuardrailBySlug(ctx context.Context, tenantID string, slug string, req platform.UpdateTenantGuardrailRequest, actor string) (*platform.TenantGuardrail, error) {
	return nil, pgx.ErrNoRows
}

func (s *integrationModelStore) DeleteTenantGuardrailBySlug(ctx context.Context, tenantID string, slug string) error {
	return pgx.ErrNoRows
}

func (s *integrationModelStore) ListEnabledTenantGuardrailsBySlugs(ctx context.Context, tenantID string, slugs []string) ([]platform.TenantGuardrail, error) {
	return nil, nil
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

func (s *integrationModelStore) GetTenantAdminUserRecord(ctx context.Context, tenantSlug string, userID string) (*platform.TenantAdminUser, error) {
	for i := range s.users {
		if s.users[i].TenantSlug == tenantSlug && s.users[i].ID == userID {
			cloned := s.users[i]
			return &cloned, nil
		}
	}

	return nil, pgx.ErrNoRows
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
	amount := int64(math.Round(amountIRR))
	s.wallet.BalanceIRR += amount

	return &platform.Transaction{TenantID: tenantID, AmountIRR: amount, Type: "wallet_topup", Status: "successful"}, nil
}

func (s *integrationWalletStore) DebitWallet(ctx context.Context, tenantID string, breakdown platform.UsageCostBreakdown, requestID string) (*platform.Transaction, error) {
	amount := int64(math.Round(breakdown.TotalCostIRR))
	if s.wallet.BalanceIRR < amount {
		return nil, storage.ErrInsufficientWalletBalance
	}

	s.debits = append(s.debits, breakdown)
	s.wallet.BalanceIRR -= amount
	usd := breakdown.TotalCostUSD
	rate := breakdown.ExchangeRate

	return &platform.Transaction{
		TenantID:     tenantID,
		Type:         "usage_debit",
		AmountIRR:    amount,
		AmountUSD:    &usd,
		ExchangeRate: &rate,
		Status:       "successful",
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

func (s *integrationWalletStore) CreateBankTransferRequest(ctx context.Context, tenantID string, req platform.CreateBankTransferRequest, actorID string) (*platform.BankTransferRequest, error) {
	now := time.Now().UTC()

	return &platform.BankTransferRequest{
		ID:                "bank-transfer-1",
		PublicID:          "bank-transfer-public-1",
		TenantID:          tenantID,
		AmountIRR:         req.AmountIRR,
		TrackingNumber:    req.TrackingNumber,
		BankTransactionAt: req.BankTransactionAt,
		Status:            "pending_review",
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}

func (s *integrationWalletStore) ListBankTransferRequests(ctx context.Context, tenantID string, limit int) ([]platform.BankTransferRequest, error) {
	return []platform.BankTransferRequest{}, nil
}

func (s *integrationWalletStore) ApproveBankTransferRequest(ctx context.Context, tenantID, requestID string, req platform.ReviewBankTransferRequest, actor platform.SupportActor) (*platform.BankTransferRequest, error) {
	now := time.Now().UTC()
	return &platform.BankTransferRequest{ID: requestID, TenantID: tenantID, Status: "approved", ReviewedBy: actor.ID, ReviewedAt: &now, CreatedAt: now, UpdatedAt: now}, nil
}

func (s *integrationWalletStore) RejectBankTransferRequest(ctx context.Context, tenantID, requestID string, req platform.ReviewBankTransferRequest, actor platform.SupportActor) (*platform.BankTransferRequest, error) {
	now := time.Now().UTC()
	return &platform.BankTransferRequest{ID: requestID, TenantID: tenantID, Status: "rejected", RejectionReason: req.Reason, ReviewedBy: actor.ID, ReviewedAt: &now, CreatedAt: now, UpdatedAt: now}, nil
}

func (s *integrationWalletStore) RedeemGiftCode(ctx context.Context, tenantID, code, actorID string) (*platform.GiftCodeRedemption, error) {
	now := time.Now().UTC()
	return &platform.GiftCodeRedemption{ID: "gift-redemption-1", GiftCodeID: "gift-code-1", TenantID: tenantID, AmountIRR: 1000, Status: "used", CreatedAt: now}, nil
}

func (s *integrationWalletStore) RedeemDiscountCode(ctx context.Context, tenantID, code, actorID string) (*platform.DiscountRedemption, error) {
	now := time.Now().UTC()
	return &platform.DiscountRedemption{ID: "discount-redemption-1", TenantID: tenantID, Code: code, DiscountType: "fixed", Status: "active", CreatedAt: now, UpdatedAt: now}, nil
}

func (s *integrationWalletStore) GetBillingProfile(ctx context.Context, tenantID string) (*platform.BillingProfile, error) {
	now := time.Now().UTC()
	return &platform.BillingProfile{ID: "billing-profile-1", TenantID: tenantID, ProfileType: "individual", Data: map[string]interface{}{"display_name": "Test Tenant"}, CreatedAt: now, UpdatedAt: now}, nil
}

func (s *integrationWalletStore) UpsertBillingProfile(ctx context.Context, tenantID string, req platform.UpsertBillingProfileRequest, actorID string) (*platform.BillingProfile, error) {
	now := time.Now().UTC()
	return &platform.BillingProfile{ID: "billing-profile-1", TenantID: tenantID, ProfileType: req.ProfileType, Data: req.Data, CreatedBy: actorID, UpdatedBy: actorID, CreatedAt: now, UpdatedAt: now}, nil
}

func (s *integrationWalletStore) ListInvoices(ctx context.Context, tenantID string, limit int) ([]platform.Invoice, error) {
	return []platform.Invoice{}, nil
}

func (s *integrationWalletStore) ListDiscountRedemptions(ctx context.Context, tenantID string, limit int) ([]platform.DiscountRedemption, error) {
	return []platform.DiscountRedemption{}, nil
}

func (s *integrationWalletStore) GetBillingUsageSummary(ctx context.Context, tenantID, preset string) (*platform.BillingUsageSummary, error) {
	now := time.Now().UTC()
	return &platform.BillingUsageSummary{Calendar: "jalali", Preset: preset, Start: now, End: now, Providers: []platform.BillingProviderUsage{}}, nil
}

func (s *integrationWalletStore) GetMonthlyUsageDetail(ctx context.Context, tenantID string, jalaliYear int, jalaliMonth int) (*platform.MonthlyUsageDetail, error) {
	now := time.Now().UTC()
	return &platform.MonthlyUsageDetail{Calendar: "jalali", JalaliYear: jalaliYear, JalaliMonth: jalaliMonth, Start: now, End: now, Providers: []platform.BillingProviderUsage{}}, nil
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

func (noopKeyStore) ListAllowedModelFQNs(ctx context.Context, tenantID string, userID string, teamID string) ([]string, error) {
	return nil, nil
}

func (noopKeyStore) TouchKeyLastUsed(ctx context.Context, keyID string, usedAt time.Time) error {
	return nil
}

func (noopKeyStore) ListPATs(ctx context.Context, tenantID string, userID string, query string, limit int, offset int) ([]platform.CreatedKey, int, error) {
	return []platform.CreatedKey{}, 0, nil
}

func (noopKeyStore) PATExists(ctx context.Context, tenantID string, userID string, name string) (bool, error) {
	return false, nil
}

func (noopKeyStore) RevokePAT(ctx context.Context, tenantID string, userID string, keyID string, revokedByUserID string) (bool, error) {
	return false, nil
}

func (noopKeyStore) RevokeAllPATs(ctx context.Context, tenantID string, userID string, revokedByUserID string) (int64, error) {
	return 0, nil
}

func (noopKeyStore) ListVATTokens(ctx context.Context, tenantID string, virtualUserID string, query string, limit int, offset int) ([]platform.CreatedKey, int, error) {
	return []platform.CreatedKey{}, 0, nil
}

func (noopKeyStore) UpdateVATToken(ctx context.Context, tenantID string, virtualUserID string, keyID string, req platform.UpdateVATTokenRequest) (*platform.CreatedKey, error) {
	return nil, pgx.ErrNoRows
}

func (noopKeyStore) RevokeVATToken(ctx context.Context, tenantID string, virtualUserID string, keyID string, revokedByUserID string, reason string) (bool, error) {
	return false, nil
}

func (noopKeyStore) RotateVATToken(ctx context.Context, tenantID string, virtualUserID string, keyID string, rotatedByUserID string, gracePeriod string) (*platform.CreatedKey, error) {
	return nil, pgx.ErrNoRows
}

func (noopKeyStore) ResolveTenantUserAuth(ctx context.Context, tenantID string, userID string) (*platform.AuthContext, error) {
	return &platform.AuthContext{
		TenantID: tenantID,
		UserID:   userID,
		Role:     "tenant_admin",
	}, nil
}

type integrationKeyStore struct {
	mu                 sync.RWMutex
	nextID             int
	byHash             map[string]*platform.AuthContext
	created            []platform.CreatedKey
	orgUnits           []platform.OrgUnit
	orgUnitMemberships []platform.OrgUnitMembership
	teamMemberships    []platform.TeamMembership
	virtualUsers       []platform.VirtualUser
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

	cloned := *info

	return &cloned, nil
}

func (s *integrationKeyStore) ResolveTenantUserAuth(ctx context.Context, tenantID string, userID string) (*platform.AuthContext, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	role := "member"
	if strings.TrimSpace(userID) == "tenant-user-1" {
		role = "tenant_admin"
	}

	return &platform.AuthContext{
		TenantID: tenantID,
		UserID:   strings.TrimSpace(userID),
		Role:     role,
	}, nil
}

func (s *integrationKeyStore) CreateKey(ctx context.Context, params platform.CreateKeyParams) (*platform.CreatedKey, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	params.Normalize()

	s.nextID++
	rawKey := fmt.Sprintf("sk-managed-%03d", s.nextID)
	keyHash := hashToken(rawKey)
	created := platform.CreatedKey{
		ID:              fmt.Sprintf("key-%03d", s.nextID),
		Key:             rawKey,
		KeyAlias:        params.KeyAlias,
		MaskedKey:       maskToken(rawKey),
		TenantID:        params.TenantID,
		Role:            params.Role,
		KeyType:         params.KeyType,
		Models:          append([]string(nil), params.Models...),
		AllowedModels:   append([]string(nil), params.AllowedModels...),
		TeamID:          params.TeamID,
		OrgID:           params.OrgID,
		OrgUnitID:       params.OrgUnitID,
		UserID:          params.UserID,
		VirtualUserID:   params.VirtualUserID,
		CreatedByUserID: params.UserID,
		TokenHash:       keyHash,
	}
	s.created = append(s.created, created)
	s.byHash[keyHash] = &platform.AuthContext{
		TenantID:       params.TenantID,
		OrganizationID: params.OrgID,
		OrgUnitID:      params.OrgUnitID,
		TeamID:         params.TeamID,
		UserID:         params.UserID,
		VirtualUserID:  params.VirtualUserID,
		KeyID:          created.ID,
		KeyAlias:       params.KeyAlias,
		Role:           params.Role,
		KeyType:        params.KeyType,
		AllowedModels:  append([]string(nil), params.AllowedModels...),
		TokenHash:      keyHash,
	}

	return &created, nil
}

func (s *integrationKeyStore) TouchKeyLastUsed(ctx context.Context, keyID string, usedAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.created {
		if s.created[i].ID == keyID {
			ts := usedAt
			s.created[i].LastUsedAt = &ts
		}
	}

	return nil
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

func (s *integrationKeyStore) ListAllowedModelFQNs(ctx context.Context, tenantID string, userID string, teamID string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	models := make([]string, 0)
	seen := make(map[string]struct{})

	for _, item := range s.created {
		if item.TenantID != tenantID {
			continue
		}

		if userID != "" && item.UserID != userID {
			continue
		}

		if teamID != "" && item.TeamID != teamID {
			continue
		}

		for _, model := range item.AllowedModels {
			model = strings.TrimSpace(model)
			if model == "" {
				continue
			}

			if _, ok := seen[model]; ok {
				continue
			}

			seen[model] = struct{}{}
			models = append(models, model)
		}
	}

	if len(models) == 0 {
		// Integration fallback for bootstrap flows that create keys before provider-account policy is seeded.
		return []string{
			"openai:default:gpt-3.5-turbo",
			"openai:default:gpt-4",
			"openai:default:gpt-4o-mini",
		}, nil
	}

	return models, nil
}

func (s *integrationKeyStore) ListPATs(ctx context.Context, tenantID string, userID string, query string, limit int, offset int) ([]platform.CreatedKey, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filtered := make([]platform.CreatedKey, 0)

	for _, item := range s.created {
		if item.TenantID != tenantID || item.KeyType != platform.KeyTypePAT || item.CreatedByUserID != userID || item.RevokedAt != nil {
			continue
		}

		if query != "" && !strings.Contains(strings.ToLower(item.KeyAlias), strings.ToLower(query)) {
			continue
		}

		listItem := item
		listItem.Key = ""
		filtered = append(filtered, listItem)
	}

	total := len(filtered)
	if offset >= total {
		return []platform.CreatedKey{}, total, nil
	}

	if limit <= 0 {
		limit = 25
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return filtered[offset:end], total, nil
}

func (s *integrationKeyStore) PATExists(ctx context.Context, tenantID string, userID string, name string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, item := range s.created {
		if item.TenantID == tenantID && item.KeyType == platform.KeyTypePAT && item.CreatedByUserID == userID && item.KeyAlias == name && item.RevokedAt == nil {
			return true, nil
		}
	}

	return false, nil
}

func (s *integrationKeyStore) RevokePAT(ctx context.Context, tenantID string, userID string, keyID string, revokedByUserID string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.created {
		if s.created[i].ID == keyID && s.created[i].TenantID == tenantID && s.created[i].KeyType == platform.KeyTypePAT && s.created[i].CreatedByUserID == userID && s.created[i].RevokedAt == nil {
			now := time.Now().UTC()

			s.created[i].RevokedAt = &now
			if info, ok := s.byHash[s.created[i].TokenHash]; ok {
				info.Blocked = true
			}

			return true, nil
		}
	}

	return false, nil
}

func (s *integrationKeyStore) RevokeAllPATs(ctx context.Context, tenantID string, userID string, revokedByUserID string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var count int64

	for i := range s.created {
		if s.created[i].TenantID == tenantID && s.created[i].KeyType == platform.KeyTypePAT && s.created[i].CreatedByUserID == userID && s.created[i].RevokedAt == nil {
			now := time.Now().UTC()

			s.created[i].RevokedAt = &now
			if info, ok := s.byHash[s.created[i].TokenHash]; ok {
				info.Blocked = true
			}

			count++
		}
	}

	return count, nil
}

func (s *integrationKeyStore) ListVATTokens(ctx context.Context, tenantID string, virtualUserID string, query string, limit int, offset int) ([]platform.CreatedKey, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filtered := make([]platform.CreatedKey, 0)

	for _, item := range s.created {
		if item.TenantID != tenantID || item.KeyType != platform.KeyTypeVAT || item.VirtualUserID != virtualUserID {
			continue
		}

		if query != "" && !strings.Contains(strings.ToLower(item.KeyAlias), strings.ToLower(query)) {
			continue
		}

		listItem := item
		listItem.Key = ""
		filtered = append(filtered, listItem)
	}

	total := len(filtered)
	if offset >= total {
		return []platform.CreatedKey{}, total, nil
	}

	if limit <= 0 {
		limit = 25
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return filtered[offset:end], total, nil
}

func (s *integrationKeyStore) UpdateVATToken(ctx context.Context, tenantID string, virtualUserID string, keyID string, req platform.UpdateVATTokenRequest) (*platform.CreatedKey, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.created {
		if s.created[i].ID != keyID || s.created[i].TenantID != tenantID || s.created[i].VirtualUserID != virtualUserID || s.created[i].KeyType != platform.KeyTypeVAT {
			continue
		}

		if strings.TrimSpace(req.KeyAlias) != "" {
			s.created[i].KeyAlias = strings.TrimSpace(req.KeyAlias)
		}

		models := req.AllowedModels
		if len(models) == 0 {
			models = req.Models
		}

		if models != nil {
			s.created[i].Models = append([]string(nil), models...)
			s.created[i].AllowedModels = append([]string(nil), models...)
		}

		if req.Blocked != nil {
			s.created[i].Blocked = *req.Blocked
			if info, ok := s.byHash[s.created[i].TokenHash]; ok {
				info.Blocked = *req.Blocked
			}
		}

		cloned := s.created[i]
		cloned.Key = ""

		return &cloned, nil
	}

	return nil, pgx.ErrNoRows
}

func (s *integrationKeyStore) RevokeVATToken(ctx context.Context, tenantID string, virtualUserID string, keyID string, revokedByUserID string, reason string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.created {
		if s.created[i].ID != keyID || s.created[i].TenantID != tenantID || s.created[i].VirtualUserID != virtualUserID || s.created[i].KeyType != platform.KeyTypeVAT || s.created[i].RevokedAt != nil {
			continue
		}

		now := time.Now().UTC()
		s.created[i].RevokedAt = &now

		s.created[i].Blocked = true
		if info, ok := s.byHash[s.created[i].TokenHash]; ok {
			info.Blocked = true
		}

		return true, nil
	}

	return false, nil
}

func (s *integrationKeyStore) RotateVATToken(ctx context.Context, tenantID string, virtualUserID string, keyID string, rotatedByUserID string, gracePeriod string) (*platform.CreatedKey, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.created {
		if s.created[i].ID != keyID || s.created[i].TenantID != tenantID || s.created[i].VirtualUserID != virtualUserID || s.created[i].KeyType != platform.KeyTypeVAT {
			continue
		}

		s.nextID++
		rawKey := fmt.Sprintf("sk-managed-%03d", s.nextID)
		keyHash := hashToken(rawKey)
		created := s.created[i]
		created.ID = fmt.Sprintf("key-%03d", s.nextID)
		created.Key = rawKey
		created.MaskedKey = maskToken(rawKey)
		created.TokenHash = keyHash
		created.Spend = 0
		created.Blocked = false
		created.RevokedAt = nil
		s.created = append(s.created, created)
		s.byHash[keyHash] = &platform.AuthContext{
			TenantID:       created.TenantID,
			OrganizationID: created.OrgID,
			OrgUnitID:      created.OrgUnitID,
			TeamID:         created.TeamID,
			UserID:         created.UserID,
			VirtualUserID:  created.VirtualUserID,
			KeyID:          created.ID,
			KeyAlias:       created.KeyAlias,
			Role:           created.Role,
			KeyType:        created.KeyType,
			AllowedModels:  append([]string(nil), created.AllowedModels...),
			TokenHash:      keyHash,
		}

		if strings.TrimSpace(gracePeriod) == "" {
			now := time.Now().UTC()
			s.created[i].RevokedAt = &now

			s.created[i].Blocked = true
			if info, ok := s.byHash[s.created[i].TokenHash]; ok {
				info.Blocked = true
			}
		}

		return &created, nil
	}

	return nil, pgx.ErrNoRows
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

func (s *integrationKeyStore) CreateOrgUnit(ctx context.Context, tenantID string, req platform.CreateOrgUnitRequest) (*platform.OrgUnit, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item := platform.OrgUnit{
		ID:          fmt.Sprintf("org-%03d", len(s.orgUnits)+1),
		TenantID:    tenantID,
		Alias:       req.Alias,
		DisplayName: req.DisplayName,
		Metadata:    req.Metadata,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	s.orgUnits = append(s.orgUnits, item)

	return &item, nil
}

func (s *integrationKeyStore) ListOrgUnits(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.OrgUnit, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filtered := make([]platform.OrgUnit, 0, len(s.orgUnits))
	for _, item := range s.orgUnits {
		if item.TenantID != tenantID {
			continue
		}

		if query != "" && !strings.Contains(strings.ToLower(item.Alias), strings.ToLower(query)) {
			continue
		}

		filtered = append(filtered, item)
	}

	total := len(filtered)
	if offset >= total {
		return []platform.OrgUnit{}, total, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return filtered[offset:end], total, nil
}

func (s *integrationKeyStore) GetOrgUnit(ctx context.Context, tenantID string, orgUnitID string) (*platform.OrgUnit, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, item := range s.orgUnits {
		if item.TenantID == tenantID && item.ID == orgUnitID {
			cloned := item
			return &cloned, nil
		}
	}

	return nil, pgx.ErrNoRows
}

func (s *integrationKeyStore) UpdateOrgUnit(ctx context.Context, tenantID string, orgUnitID string, req platform.UpdateOrgUnitRequest) (*platform.OrgUnit, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.orgUnits {
		if s.orgUnits[i].TenantID != tenantID || s.orgUnits[i].ID != orgUnitID {
			continue
		}

		if req.Alias != "" {
			s.orgUnits[i].Alias = req.Alias
		}

		if req.DisplayName != "" {
			s.orgUnits[i].DisplayName = req.DisplayName
		}

		if req.Metadata != nil {
			s.orgUnits[i].Metadata = req.Metadata
		}

		s.orgUnits[i].UpdatedAt = time.Now().UTC()
		cloned := s.orgUnits[i]

		return &cloned, nil
	}

	return nil, pgx.ErrNoRows
}

func (s *integrationKeyStore) DeleteOrgUnit(ctx context.Context, tenantID string, orgUnitID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.orgUnits {
		if s.orgUnits[i].TenantID == tenantID && s.orgUnits[i].ID == orgUnitID {
			s.orgUnits = append(s.orgUnits[:i], s.orgUnits[i+1:]...)
			return nil
		}
	}

	return pgx.ErrNoRows
}

func (s *integrationKeyStore) ListOrgUnitMemberships(ctx context.Context, tenantID string, orgUnitID string) ([]platform.OrgUnitMembership, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]platform.OrgUnitMembership, 0)

	for _, item := range s.orgUnitMemberships {
		if item.TenantID == tenantID && item.OrgUnitID == orgUnitID {
			items = append(items, item)
		}
	}

	return items, nil
}

func (s *integrationKeyStore) AddOrgUnitMembership(ctx context.Context, tenantID string, orgUnitID string, req platform.CreateOrgUnitMembershipRequest) (*platform.OrgUnitMembership, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.orgUnitMemberships {
		if s.orgUnitMemberships[i].TenantID == tenantID && s.orgUnitMemberships[i].OrgUnitID == orgUnitID && s.orgUnitMemberships[i].UserID == req.UserID {
			s.orgUnitMemberships[i].Role = req.Role
			s.orgUnitMemberships[i].UpdatedAt = time.Now().UTC()
			cloned := s.orgUnitMemberships[i]

			return &cloned, nil
		}
	}

	item := platform.OrgUnitMembership{
		ID:        fmt.Sprintf("oum-%03d", len(s.orgUnitMemberships)+1),
		TenantID:  tenantID,
		OrgUnitID: orgUnitID,
		UserID:    req.UserID,
		Role:      req.Role,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	s.orgUnitMemberships = append(s.orgUnitMemberships, item)

	return &item, nil
}

func (s *integrationKeyStore) UpdateOrgUnitMembership(ctx context.Context, tenantID string, orgUnitID string, userID string, req platform.UpdateOrgUnitMembershipRequest) (*platform.OrgUnitMembership, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.orgUnitMemberships {
		if s.orgUnitMemberships[i].TenantID == tenantID && s.orgUnitMemberships[i].OrgUnitID == orgUnitID && s.orgUnitMemberships[i].UserID == userID {
			s.orgUnitMemberships[i].Role = req.Role
			s.orgUnitMemberships[i].UpdatedAt = time.Now().UTC()
			cloned := s.orgUnitMemberships[i]

			return &cloned, nil
		}
	}

	return nil, pgx.ErrNoRows
}

func (s *integrationKeyStore) DeleteOrgUnitMembership(ctx context.Context, tenantID string, orgUnitID string, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.orgUnitMemberships {
		if s.orgUnitMemberships[i].TenantID == tenantID && s.orgUnitMemberships[i].OrgUnitID == orgUnitID && s.orgUnitMemberships[i].UserID == userID {
			s.orgUnitMemberships = append(s.orgUnitMemberships[:i], s.orgUnitMemberships[i+1:]...)
			return nil
		}
	}

	return pgx.ErrNoRows
}

func (s *integrationKeyStore) ListTeamMemberships(ctx context.Context, tenantID string, teamID string) ([]platform.TeamMembership, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]platform.TeamMembership, 0)

	for _, item := range s.teamMemberships {
		if item.TenantID == tenantID && item.TeamID == teamID {
			items = append(items, item)
		}
	}

	return items, nil
}

func (s *integrationKeyStore) AddTeamMembership(ctx context.Context, tenantID string, teamID string, req platform.CreateTeamMembershipRequest) (*platform.TeamMembership, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.teamMemberships {
		if s.teamMemberships[i].TenantID != tenantID || s.teamMemberships[i].TeamID != teamID || s.teamMemberships[i].UserID != req.UserID {
			continue
		}

		s.teamMemberships[i].Role = req.Role
		s.teamMemberships[i].TeamMemberPermissions = append([]string(nil), req.TeamMemberPermissions...)
		s.teamMemberships[i].UpdatedAt = time.Now().UTC()
		cloned := s.teamMemberships[i]

		return &cloned, nil
	}

	item := platform.TeamMembership{
		ID:                    fmt.Sprintf("tm-%03d", len(s.teamMemberships)+1),
		TenantID:              tenantID,
		TeamID:                teamID,
		UserID:                req.UserID,
		Role:                  req.Role,
		TeamMemberPermissions: append([]string(nil), req.TeamMemberPermissions...),
		CreatedAt:             time.Now().UTC(),
		UpdatedAt:             time.Now().UTC(),
	}
	s.teamMemberships = append(s.teamMemberships, item)

	return &item, nil
}

func (s *integrationKeyStore) UpdateTeamMembership(ctx context.Context, tenantID string, teamID string, userID string, req platform.UpdateTeamMembershipRequest) (*platform.TeamMembership, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.teamMemberships {
		if s.teamMemberships[i].TenantID != tenantID || s.teamMemberships[i].TeamID != teamID || s.teamMemberships[i].UserID != userID {
			continue
		}

		if req.Role != "" {
			s.teamMemberships[i].Role = req.Role
		}

		if req.TeamMemberPermissions != nil {
			s.teamMemberships[i].TeamMemberPermissions = append([]string(nil), req.TeamMemberPermissions...)
		}

		s.teamMemberships[i].UpdatedAt = time.Now().UTC()
		cloned := s.teamMemberships[i]

		return &cloned, nil
	}

	return nil, pgx.ErrNoRows
}

func (s *integrationKeyStore) DeleteTeamMembership(ctx context.Context, tenantID string, teamID string, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.teamMemberships {
		if s.teamMemberships[i].TenantID == tenantID && s.teamMemberships[i].TeamID == teamID && s.teamMemberships[i].UserID == userID {
			s.teamMemberships = append(s.teamMemberships[:i], s.teamMemberships[i+1:]...)
			return nil
		}
	}

	return pgx.ErrNoRows
}

func (s *integrationKeyStore) CreateVirtualUser(ctx context.Context, tenantID string, req platform.CreateVirtualUserRequest) (*platform.VirtualUser, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item := platform.VirtualUser{
		ID:          fmt.Sprintf("vu-%03d", len(s.virtualUsers)+1),
		TenantID:    tenantID,
		OrgUnitID:   req.OrgUnitID,
		TeamID:      req.TeamID,
		Alias:       req.Alias,
		DisplayName: req.DisplayName,
		Active:      true,
		Metadata:    req.Metadata,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	s.virtualUsers = append(s.virtualUsers, item)

	return &item, nil
}

func (s *integrationKeyStore) ListVirtualUsers(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.VirtualUser, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filtered := make([]platform.VirtualUser, 0, len(s.virtualUsers))
	for _, item := range s.virtualUsers {
		if item.TenantID != tenantID {
			continue
		}

		if query != "" && !strings.Contains(strings.ToLower(item.Alias), strings.ToLower(query)) {
			continue
		}

		filtered = append(filtered, item)
	}

	total := len(filtered)
	if offset >= total {
		return []platform.VirtualUser{}, total, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return filtered[offset:end], total, nil
}

func (s *integrationKeyStore) GetVirtualUser(ctx context.Context, tenantID string, virtualUserID string) (*platform.VirtualUser, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, item := range s.virtualUsers {
		if item.TenantID == tenantID && item.ID == virtualUserID {
			cloned := item
			return &cloned, nil
		}
	}

	return nil, pgx.ErrNoRows
}

func (s *integrationKeyStore) UpdateVirtualUser(ctx context.Context, tenantID string, virtualUserID string, req platform.UpdateVirtualUserRequest) (*platform.VirtualUser, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.virtualUsers {
		if s.virtualUsers[i].TenantID != tenantID || s.virtualUsers[i].ID != virtualUserID {
			continue
		}

		if req.Alias != "" {
			s.virtualUsers[i].Alias = req.Alias
		}

		if req.DisplayName != "" {
			s.virtualUsers[i].DisplayName = req.DisplayName
		}

		if req.TeamID != "" {
			s.virtualUsers[i].TeamID = req.TeamID
		}

		if req.OrgUnitID != "" {
			s.virtualUsers[i].OrgUnitID = req.OrgUnitID
		}

		if req.Active != nil {
			s.virtualUsers[i].Active = *req.Active
		}

		if req.Metadata != nil {
			s.virtualUsers[i].Metadata = req.Metadata
		}

		s.virtualUsers[i].UpdatedAt = time.Now().UTC()
		cloned := s.virtualUsers[i]

		return &cloned, nil
	}

	return nil, pgx.ErrNoRows
}

func (s *integrationKeyStore) DeleteVirtualUser(ctx context.Context, tenantID string, virtualUserID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.virtualUsers {
		if s.virtualUsers[i].TenantID == tenantID && s.virtualUsers[i].ID == virtualUserID {
			s.virtualUsers = append(s.virtualUsers[:i], s.virtualUsers[i+1:]...)
			return nil
		}
	}

	return pgx.ErrNoRows
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
			{
				ID:         "tenant-user-2",
				TenantSlug: "default",
				Email:      "member@tenant.test",
				Role:       "member",
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
	configureOpenAITestCatalog(t, billingSvc)
	llmRouter, err := router.NewRouter(nil, config.RouterSettings{
		RoutingStrategy: "simple-shuffle",
		CooldownTime:    60,
	}, memCache, modelStore, logger)
	require.NoError(t, err)

	baseProxyDeps := proxy.BaseHandlerDependencies{
		Router:        llmRouter,
		Tunnel:        tunneltransport.NewHTTPTransport(tunnelClient.URL, 30*time.Second),
		Billing:       billingSvc,
		Wallets:       walletStore,
		RequestLogger: walletStore,
		SpendLogger:   walletStore,
		Logger:        logger,
	}
	app := controlserver.NewFiberApp(&config.Config{}, controlserver.Dependencies{
		AuthMiddleware: auth.NewAuthMiddleware(keyStore, "sk-master", logger),
		TenantResolver: tenantmw.NewMiddlewareWithCache(tenantStore, nil, tenantmw.MiddlewareConfig{
			AllowHeaderHint: true,
		}),
		AdminHandlers: admin.New(admin.Dependencies{
			Keys:         keyStore,
			Models:       modelStore,
			Wallets:      walletStore,
			OrgUnits:     keyStore,
			Memberships:  keyStore,
			VirtualUsers: keyStore,
			Users:        modelStore,
		}),
		ProxyHandler: proxy.NewChatCompletionHandlerWithDependencies(proxy.ChatCompletionDependencies{
			Base:              baseProxyDeps,
			VirtualModelStore: modelStore,
		}),
		CompletionsHandler: proxy.NewCompletionsHandlerWithDependencies(baseProxyDeps),
		ResponsesHandler:   proxy.NewResponsesHandlerWithDependencies(baseProxyDeps),
		ModelsHandler:      proxy.NewModelsHandler(modelStore),
		Logger:             logger,
	})

	walletStore.onSpend = keyStore.addSpend

	return app, keyStore, modelStore, walletStore
}

func configureOpenAITestCatalog(t *testing.T, service *billing.Service) {
	t.Helper()

	t.Setenv("GODARAI_OPENAI_CONTRACT_READY", "true")
	now := time.Now().UTC()
	discovery := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/models/runtime-snapshot/openai", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(map[string]interface{}{
			"provider":    "openai",
			"revision":    "integration-openai-1",
			"generatedAt": now,
			"models": []interface{}{
				map[string]interface{}{
					"canonicalId": "openai/gpt-4o-mini",
					"slug":        "gpt-4o-mini",
					"status":      "active",
					"runtimeProfile": map[string]interface{}{
						"version": 1,
						"profile": map[string]interface{}{
							"version":    1,
							"endpoints":  []string{"responses"},
							"transports": []string{"http", "sse"},
							"thinking": map[string]interface{}{
								"defaultMode":  "disabled",
								"allowedModes": []string{"disabled"},
								"effortValues": []string{},
							},
							"assistantPrefill": false,
							"sampling": map[string]interface{}{
								"parameters":             []string{"temperature", "top_p"},
								"disallowedWithThinking": []string{},
							},
							"tools": map[string]interface{}{
								"custom":           true,
								"server":           true,
								"parallel":         true,
								"structuredOutput": true,
								"supported":        []string{"function_calling"},
							},
							"modalities": map[string]interface{}{
								"input":  []string{"text", "image"},
								"output": []string{"text"},
							},
							"requestControls": map[string]interface{}{
								"serviceTierValues":  []string{"standard"},
								"inferenceGeoValues": []string{"global"},
								"speedValues":        []string{"standard"},
							},
							"betaHeaders": []string{},
							"limits": map[string]interface{}{
								"contextWindow":   128000,
								"maxOutputTokens": 16384,
								"messagesBytes":   1048576,
								"batchBytes":      1048576,
								"fileBytes":       1048576,
							},
						},
						"sourceUrl":     "https://developers.openai.com/api/docs/models/gpt-4o-mini",
						"effectiveFrom": now.Add(-time.Hour),
					},
					"pricingRates": []interface{}{
						map[string]interface{}{
							"dimension":     "input_tokens",
							"unit":          "token",
							"unitPriceUsd":  "0.00000015",
							"currency":      "USD",
							"region":        "global",
							"sourceUrl":     "https://developers.openai.com/api/docs/pricing",
							"effectiveFrom": now.Add(-time.Hour),
						},
						map[string]interface{}{
							"dimension":     "output_tokens",
							"unit":          "token",
							"unitPriceUsd":  "0.0000006",
							"currency":      "USD",
							"region":        "global",
							"sourceUrl":     "https://developers.openai.com/api/docs/pricing",
							"effectiveFrom": now.Add(-time.Hour),
						},
					},
				},
			},
		}))
	}))
	t.Cleanup(discovery.Close)
	t.Setenv("SHAHRAH_MODEL_DISCOVERY_URL", discovery.URL)

	manager, err := billing.NewCatalogPricingSnapshotManagerFromEnv(
		"openai",
		discovery.Client(),
	)
	require.NoError(t, err)
	require.NoError(t, manager.Reload(context.Background()))
	service.SetCatalogPricingSnapshotManager("openai", manager)
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

func createVirtualModel(t *testing.T, app *fiber.App, body string) platform.VirtualModel {
	t.Helper()
	resp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/virtual-models", "Bearer sk-master", body)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var out platform.VirtualModel
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

func performJSONRequest(t *testing.T, app *fiber.App, method string, url string, authz string, body string) *testResponse {
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
	t.Cleanup(func() { closeResponseBody(t, resp) })

	return &testResponse{Response: resp}
}

func performJSONRequestWithHeaders(t *testing.T, app *fiber.App, method string, url string, authz string, headers map[string]string, body string) *testResponse {
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

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := app.Test(req, int((30 * time.Second).Milliseconds()))
	require.NoError(t, err)
	t.Cleanup(func() { closeResponseBody(t, resp) })

	return &testResponse{Response: resp}
}

func decodeJSONBody(t *testing.T, resp *testResponse, target interface{}) {
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
