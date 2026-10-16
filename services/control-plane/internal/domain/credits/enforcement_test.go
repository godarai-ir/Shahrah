package credits

import (
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/require"
)

func TestPoolAllowanceAddsBackSettledConsumption(t *testing.T) {
	t.Parallel()

	// 10 credits left after 4 were settled this month; the window has
	// committed those 4 too, so its allowance must be 14 to leave 10.
	pool := PoolAllowance{AvailableMC: 10_000, SettledMC: 4_000}
	require.Equal(t, int64(14_000), pool.AllowanceMC())

	pool.DebtMC = 3_000
	require.Equal(t, int64(11_000), pool.AllowanceMC(), "debt reduces what is left")

	pool = PoolAllowance{DebtMC: 9_000, SettledMC: 1_000}
	require.Zero(t, pool.AllowanceMC(), "an allowance is never negative")
}

func TestPoolAllowanceOverageHeadroom(t *testing.T) {
	t.Parallel()

	capMC := int64(5_000)

	pool := PoolAllowance{DebtMC: 2_000, OverageEnabled: true, OverageCapMC: &capMC, OverageCeilingMC: 100_000_000}
	require.Equal(t, int64(1_000), pool.AllowanceMC(), "a cap of 5 credits with 2 owed leaves 3, minus the 2 owed")

	pool = PoolAllowance{AvailableMC: 1_000, OverageEnabled: true, OverageCeilingMC: 50_000}
	require.Equal(t, int64(51_000), pool.AllowanceMC(), "no cap falls back to the ceiling")

	pool = PoolAllowance{AvailableMC: 1_000, OverageCapMC: &capMC}
	require.Equal(t, int64(1_000), pool.AllowanceMC(), "a cap means nothing while overage is off")
}

func TestPoolWindowKeysNeverMixWithIRRWindows(t *testing.T) {
	t.Parallel()

	windowID, end := PoolWindow(time.Date(2026, 10, 31, 23, 0, 0, 0, time.FixedZone("x", -3*3600)))
	require.Equal(t, "202611", windowID, "the window is chosen in UTC")
	require.Equal(t, time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC), end)

	require.Equal(t, "credit:tenant:t1:window:202611", PoolWindowKey("t1", windowID))
	require.NotContains(t, PoolWindowKey("t1", windowID), "budget:tenant:")
	require.Equal(t, "credit_policy:t1:rule:r1:scope:abc:window:202611", PolicyWindowKey("t1", "r1", "abc", windowID))
}

func TestParseCreditAmountMC(t *testing.T) {
	t.Parallel()

	amount, err := ParseCreditAmountMC(" 12.5 ")
	require.NoError(t, err)
	require.Equal(t, int64(12_500), amount)

	for _, raw := range []string{"", "0", "-1", "1.005", "abc"} {
		_, err := ParseCreditAmountMC(raw)
		require.ErrorIs(t, err, ErrInvalidCreditAmount, raw)
	}
}

func TestNormalizeAlertThresholds(t *testing.T) {
	t.Parallel()

	thresholds, err := NormalizeAlertThresholds(nil)
	require.NoError(t, err)
	require.Equal(t, []int{80, 100}, thresholds)

	thresholds, err = NormalizeAlertThresholds([]int{100, 50, 100})
	require.NoError(t, err)
	require.Equal(t, []int{50, 100}, thresholds)

	_, err = NormalizeAlertThresholds([]int{0})
	require.ErrorIs(t, err, ErrInvalidAlertThreshold)
}

func TestCrossedThresholds(t *testing.T) {
	t.Parallel()

	thresholds := []int{80, 100}
	require.Equal(t, []int{80}, CrossedThresholds(thresholds, 10_000, 7_000, 8_000))
	require.Equal(t, []int{80, 100}, CrossedThresholds(thresholds, 10_000, 0, 12_000))
	require.Empty(t, CrossedThresholds(thresholds, 10_000, 8_000, 9_000), "80% was crossed before")
	require.Empty(t, CrossedThresholds(thresholds, 0, 0, 1), "no limit, no alert")
}

func TestResolveOnExceed(t *testing.T) {
	t.Parallel()

	yes, no := true, false
	block, alert, downgrade := platform.BudgetOnExceedBlock, platform.BudgetOnExceedAlert, platform.BudgetOnExceedDowngrade
	unknown := platform.BudgetOnExceed("pause")

	cases := []struct {
		name     string
		onExceed *platform.BudgetOnExceed
		block    *bool
		current  platform.BudgetOnExceed
		want     platform.BudgetOnExceed
		err      error
	}{
		{name: "default is block", want: block},
		{name: "legacy true", block: &yes, want: block},
		{name: "legacy false keeps not blocking", block: &no, want: alert},
		{name: "keep stored value", current: downgrade, want: downgrade},
		{name: "explicit", onExceed: &downgrade, want: downgrade},
		{name: "consistent pair", onExceed: &alert, block: &no, want: alert},
		{name: "inconsistent pair", onExceed: &alert, block: &yes, err: ErrInconsistentOnExceed},
		{name: "downgrade can block", onExceed: &downgrade, block: &no, err: ErrInconsistentOnExceed},
		{name: "unknown", onExceed: &unknown, err: ErrInvalidOnExceed},
	}

	for _, tc := range cases {
		got, err := ResolveOnExceed(tc.onExceed, tc.block, tc.current)
		if tc.err != nil {
			require.ErrorIs(t, err, tc.err, tc.name)

			continue
		}

		require.NoError(t, err, tc.name)
		require.Equal(t, tc.want, got, tc.name)
	}
}
