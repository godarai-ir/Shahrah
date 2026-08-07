package storage

import (
	"context"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	billingdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
)

func (s *PlanCatalogRepository) GetCommercialSignupSettings(ctx context.Context) (*platform.CommercialSignupSettings, error) {
	var out platform.CommercialSignupSettings

	err := s.pool.QueryRow(ctx, `
		SELECT trial_days, signup_credit_irr::bigint, COALESCE(updated_by, ''), updated_at
		FROM commercial_signup_settings WHERE settings_id = true
	`).Scan(&out.TrialDays, &out.SignupCreditIRR, &out.UpdatedBy, &out.UpdatedAt)

	return &out, err
}

func (s *PlanCatalogRepository) UpdateCommercialSignupSettings(ctx context.Context, credit int64, actor string) (*platform.CommercialSignupSettings, error) {
	if credit < 0 || credit > 999999999999999999 {
		return nil, billingdomain.ErrInvalidSignupCredit
	}

	var out platform.CommercialSignupSettings

	err := s.pool.QueryRow(ctx, `
		INSERT INTO commercial_signup_settings (settings_id, trial_days, signup_credit_irr, updated_by)
		VALUES (true, 15, $1, NULLIF($2, ''))
		ON CONFLICT (settings_id) DO UPDATE
		SET signup_credit_irr = EXCLUDED.signup_credit_irr, updated_by = EXCLUDED.updated_by, updated_at = NOW()
		RETURNING trial_days, signup_credit_irr::bigint, COALESCE(updated_by, ''), updated_at
	`, credit, actor).Scan(&out.TrialDays, &out.SignupCreditIRR, &out.UpdatedBy, &out.UpdatedAt)

	return &out, err
}
