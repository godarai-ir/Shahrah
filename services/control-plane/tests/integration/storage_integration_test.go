package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
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

func newPersistentIntegrationApp(t *testing.T, logger *zap.Logger, memCache cache.Cache, databaseURL string) (*fiber.App, *pgxpool.Pool) {
	t.Helper()

	ctx := context.Background()
	store, err := storage.NewStore(ctx, databaseURL, "default", logger)
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
		Logger: logger,
	})

	return app, pool
}

func resetPersistentState(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	ctx := context.Background()
	_, err := pool.Exec(ctx, `
		TRUNCATE TABLE
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
