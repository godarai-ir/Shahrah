package integration_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/Delpak/shahrah/services/control-plane/internal/testcache"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newCreditIntegrationStore(t *testing.T) (*storage.Store, *pgxpool.Pool) {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(databaseURL) == "" {
		t.Skip("DATABASE_URL not set")
	}

	ctx := context.Background()
	logger := zap.NewNop()

	mem, err := testcache.New(t, logger)
	require.NoError(t, err)

	store, err := storage.NewStore(ctx, databaseURL, logger, storage.StoreDependencies{Cache: mem})
	require.NoError(t, err)
	t.Cleanup(store.Close)

	pool, err := pgxpool.New(ctx, databaseURL)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	return store, pool
}

func TestIntegration_CreditRateCardPublishedVersionAndMultipliers(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()

	card, err := store.GetPublishedCreditRateCard(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, card.Version)
	require.Equal(t, "rc-1", card.Label())

	model := "xai/grok-integration-" + uuid.NewString()[:8]
	_, err = pool.Exec(ctx, `
		INSERT INTO credit_model_multipliers (version, model, multiplier, reason, created_by)
		VALUES (1, $1, 1.25, 'integration test egress cost', 'test')
	`, model)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, cleanupErr := pool.Exec(context.Background(), `DELETE FROM credit_model_multipliers WHERE model = $1`, model)
		require.NoError(t, cleanupErr)
	})

	card, err = store.GetPublishedCreditRateCard(ctx)
	require.NoError(t, err)
	require.Equal(t, "1.25", card.Multiplier(model).String())
	require.Equal(t, "1", card.Multiplier("anthropic/unknown").String())
}

func TestIntegration_SpendLogPersistsShadowCredits(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()

	withCredits := "credit-shadow-" + uuid.NewString()
	withoutCredits := "credit-none-" + uuid.NewString()

	t.Cleanup(func() {
		_, cleanupErr := pool.Exec(context.Background(), `DELETE FROM spend_logs WHERE request_id = ANY($1)`, []string{withCredits, withoutCredits})
		require.NoError(t, cleanupErr)
	})

	now := time.Now().UTC()
	err := store.LogSpends(ctx, []platform.SpendLog{
		{
			RequestID: withCredits, Model: "anthropic/claude-sonnet-5", CostUSD: 0.009, CostIRR: 2430,
			Status: "success", CompletionStart: now, CompletionEnd: now,
			Credits: &platform.SpendLogCredits{
				Millicredits: 4500, RateCardVersion: 1, Product: platform.CreditProductLibreChat,
				BillingMode: platform.BillingModeWalletBilled, Source: "components",
			},
		},
		{RequestID: withoutCredits, Model: "m", Status: "success", CompletionStart: now, CompletionEnd: now},
	})
	require.NoError(t, err)

	var (
		creditsMC       int64
		rateCardVersion int
		product, source string
	)

	err = pool.QueryRow(ctx, `
		SELECT credits_mc, rate_card_version, product, credit_source
		FROM spend_logs WHERE request_id = $1
	`, withCredits).Scan(&creditsMC, &rateCardVersion, &product, &source)
	require.NoError(t, err)
	require.EqualValues(t, 4500, creditsMC)
	require.Equal(t, 1, rateCardVersion)
	require.Equal(t, platform.CreditProductLibreChat, product)
	require.Equal(t, "components", source)

	var nullCredits bool

	err = pool.QueryRow(ctx, `SELECT credits_mc IS NULL FROM spend_logs WHERE request_id = $1`, withoutCredits).Scan(&nullCredits)
	require.NoError(t, err)
	require.True(t, nullCredits, "spend logs without shadow credits must store NULL credit columns")
}
