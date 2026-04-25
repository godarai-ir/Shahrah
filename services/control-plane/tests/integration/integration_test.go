package integration_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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
	tenantmw "github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	tunneltransport "github.com/Delpak/shahrah/services/control-plane/internal/tunnel"
	silkserver "github.com/Delpak/shahrah/services/silk-road/pkg/server"
	tclientserver "github.com/Delpak/shahrah/services/tunnel-client/pkg/server"
	tserverserver "github.com/Delpak/shahrah/services/tunnel-server/pkg/server"
	"github.com/stretchr/testify/assert"
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
		LiteLLMParams: config.LiteLLMParams{
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
		Logger: logger,
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
}

func (s *integrationModelStore) CreateDeployment(ctx context.Context, dep platform.Deployment) (*platform.Deployment, error) {
	s.deployments = append(s.deployments, dep)
	return &dep, nil
}

func (s *integrationModelStore) ListDeployments(ctx context.Context, tenantID string) ([]platform.Deployment, error) {
	return s.deployments, nil
}

type integrationWalletStore struct {
	wallet      *platform.Wallet
	rate        *platform.ExchangeRate
	debits      []platform.UsageCostBreakdown
	requestLogs []platform.RequestLogEntry
	spendLogs   []platform.SpendLog
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

func (s *integrationWalletStore) LogRequest(ctx context.Context, entry platform.RequestLogEntry) error {
	s.requestLogs = append(s.requestLogs, entry)
	return nil
}

func (s *integrationWalletStore) LogSpend(ctx context.Context, log platform.SpendLog) error {
	s.spendLogs = append(s.spendLogs, log)
	return nil
}

type noopKeyStore struct{}

func (noopKeyStore) CreateKey(ctx context.Context, params platform.CreateKeyParams) (*platform.CreatedKey, error) {
	return &platform.CreatedKey{}, nil
}

func (noopKeyStore) ListKeys(ctx context.Context, tenantID string) ([]platform.CreatedKey, error) {
	return nil, nil
}
