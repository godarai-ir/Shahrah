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
			"models", "max_budget", "rpm_limit", "tpm_limit", "spend", "expires", "blocked", "created_at", "updated_at",
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
