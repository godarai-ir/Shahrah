package storage

import (
	"context"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
)

// GetTenantCreditBalance returns a tenant's spendable lots and a depletion
// forecast from its consumption over the forecast window.
func (s *CreditLedgerRepository) GetTenantCreditBalance(ctx context.Context, tenantID string, now time.Time) (*platform.TenantCreditBalance, error) {
	settings, err := (&CreditPackRepository{walletState: s.walletState}).GetCreditSettings(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	balance, err := s.GetCreditBalance(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	lots, err := s.ListTenantCreditLots(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	spendable := make([]platform.CreditLot, 0, len(lots))

	// expiresBy is when the last spendable lot expires, or nil when one never does.
	var expiresBy *time.Time

	neverExpires := false

	for _, lot := range lots {
		if lot.RemainingCredits == credits.Format(0) || (lot.ExpiresAt != nil && !lot.ExpiresAt.After(now)) {
			continue
		}

		spendable = append(spendable, lot)

		switch {
		case lot.ExpiresAt == nil:
			neverExpires = true
		case expiresBy == nil || lot.ExpiresAt.After(*expiresBy):
			expiresBy = lot.ExpiresAt
		}
	}

	if neverExpires {
		expiresBy = nil
	}

	var (
		consumedMC int64
		firstUsage *time.Time
	)

	if err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(-SUM(amount_mc), 0)::bigint, MIN(occurred_at)
		FROM credit_ledger
		WHERE tenant_id = $1::uuid AND kind = 'consume' AND occurred_at >= $2 AND occurred_at <= $3
	`, tenantID, now.Add(-credits.ForecastWindow), now).Scan(&consumedMC, &firstUsage); err != nil {
		return nil, err
	}

	first := time.Time{}
	if firstUsage != nil {
		first = *firstUsage
	}

	return &platform.TenantCreditBalance{
		AvailableCredits: credits.Format(credits.Millicredits(balance.AvailableMC)),
		DebtCredits:      credits.Format(credits.Millicredits(balance.DebtMC)),
		Lots:             spendable,
		Forecast:         credits.Forecast(balance.AvailableMC-balance.DebtMC, consumedMC, first, now, expiresBy),
		Settings:         *settings,
	}, nil
}
