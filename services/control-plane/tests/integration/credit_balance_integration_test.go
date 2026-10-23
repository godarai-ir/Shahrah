package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestIntegration_TenantCreditBalanceListsSpendableLotsAndForecasts(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()

	settleAllCredits(t, store)

	tenantID := newCreditLedgerTenant(t, pool, credits.BillingUnitCredit)
	monthEnd := time.Now().UTC().AddDate(0, 1, 0).Truncate(time.Second)

	grantCredits(t, store, tenantID, credits.LotSourceTrial, 500_000, &monthEnd)

	weekEnd := time.Now().UTC().AddDate(0, 0, 7).Truncate(time.Second)
	// The promo lot expires first, so it is spent first.
	spent := grantCredits(t, store, tenantID, credits.LotSourcePromo, 1_000, &weekEnd)

	insertCreditSpendLog(t, pool, tenantID, "req-"+uuid.NewString(), 3_000, "")
	settleAllCredits(t, store)

	balance, err := store.GetTenantCreditBalance(ctx, tenantID, time.Now().UTC())
	require.NoError(t, err)
	require.Equal(t, "498.00", balance.AvailableCredits)
	require.Len(t, balance.Lots, 1, "the drained promo lot is not spendable")
	require.NotEqual(t, spent.LotID, balance.Lots[0].LotID)
	require.Equal(t, "498.00", balance.Lots[0].RemainingCredits)

	require.Equal(t, platform.CreditForecastUncertain, balance.Forecast.Status)
	require.Equal(t, "3.00", balance.Forecast.ConsumedCredits)
	require.NotNil(t, balance.Forecast.DepletesAt)
	require.Equal(t, credits.BillingUnitCredit, balance.Settings.BillingUnit)
}
