package storage

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	onboardingOTPTTL      = 5 * time.Minute
	maxOnboardingAttempts = 5
)

var (
	ErrTenantSlugInvalid   = errors.New("tenant slug is invalid or reserved")
	ErrTenantSlugTaken     = errors.New("tenant slug is already taken")
	ErrOnboardingNotFound  = errors.New("onboarding request not found")
	ErrOnboardingExpired   = errors.New("onboarding request expired")
	ErrOnboardingOTPFailed = errors.New("invalid onboarding otp")
)

type TenantSlugAvailability struct {
	Slug      string `json:"slug"`
	Available bool   `json:"available"`
	Reason    string `json:"reason,omitempty"`
}

type CreateOnboardingOTPParams struct {
	TenantSlug   string                 `json:"tenant_slug"`
	TenantDomain string                 `json:"tenant_domain"`
	PlanSlug     string                 `json:"plan_slug"`
	Email        string                 `json:"email"`
	Mobile       string                 `json:"mobile"`
	CustomerName string                 `json:"customer_name"`
	Metadata     map[string]interface{} `json:"metadata"`
}

type OnboardingOTP struct {
	RequestID         string `json:"request_id"`
	TenantSlug        string `json:"tenant_slug"`
	Channel           string `json:"channel"`
	DestinationMasked string `json:"destination_masked"`
	ExpiresInSeconds  int    `json:"expires_in_seconds"`
	AlreadyPending    bool   `json:"already_pending,omitempty"`
	DebugOTP          string `json:"debug_otp,omitempty"`
}

type VerifyOnboardingOTPParams struct {
	RequestID string `json:"request_id"`
	OTPCode   string `json:"otp_code"`
}

type ProvisionedTenant struct {
	Tenant   platform.Tenant `json:"tenant"`
	LoginURL string          `json:"login_url"`
}

func (s *Store) CheckTenantSlugAvailability(ctx context.Context, slug string) (*TenantSlugAvailability, error) {
	slug = normalizeOnboardingSlug(slug)
	if !tenant.ValidateTenantSlug(slug) {
		return &TenantSlugAvailability{Slug: slug, Available: false, Reason: "invalid_or_reserved"}, nil
	}

	if err := s.expirePendingOnboardingRequests(ctx); err != nil {
		return nil, err
	}

	var exists bool
	if err := s.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM tenants WHERE slug = $1
			UNION ALL
			SELECT 1 FROM tenant_onboarding_requests
			WHERE tenant_slug = $1 AND status = 'pending' AND expires_at > NOW()
		)
	`, slug).Scan(&exists); err != nil {
		return nil, err
	}

	if exists {
		return &TenantSlugAvailability{Slug: slug, Available: false, Reason: "taken"}, nil
	}

	return &TenantSlugAvailability{Slug: slug, Available: true}, nil
}

func (s *Store) CreateOnboardingOTP(ctx context.Context, p CreateOnboardingOTPParams) (*OnboardingOTP, error) {
	p.TenantSlug = normalizeOnboardingSlug(p.TenantSlug)
	p.TenantDomain = strings.TrimSpace(strings.ToLower(p.TenantDomain))
	p.PlanSlug = strings.TrimSpace(strings.ToLower(p.PlanSlug))
	p.Email = strings.TrimSpace(strings.ToLower(p.Email))
	p.Mobile = strings.TrimSpace(p.Mobile)

	p.CustomerName = strings.TrimSpace(p.CustomerName)
	if p.PlanSlug == "" {
		p.PlanSlug = "developer"
	}

	if p.PlanSlug == "free" {
		p.PlanSlug = "developer"
	}

	if p.TenantDomain == "" {
		p.TenantDomain = p.TenantSlug
	}

	if !tenant.ValidateTenantSlug(p.TenantSlug) {
		return nil, ErrTenantSlugInvalid
	}

	if p.Email == "" && p.Mobile == "" {
		return nil, errors.New("email or mobile is required")
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var planExists bool

	planErr := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM commercial_plans p
			JOIN plan_versions pv ON pv.plan_id = p.plan_id AND pv.status = 'published'
			WHERE p.slug = $1 AND p.active = true
		)
	`, p.PlanSlug).Scan(&planExists)
	if planErr != nil {
		return nil, planErr
	}

	if !planExists {
		return nil, ErrPlanNotFound
	}

	_, lockErr := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, p.TenantSlug)
	if lockErr != nil {
		return nil, lockErr
	}

	_, expireErr := tx.Exec(ctx, `
		UPDATE tenant_onboarding_requests
		SET status = 'expired', updated_at = NOW()
		WHERE status = 'pending' AND expires_at <= NOW()
	`)
	if expireErr != nil {
		return nil, expireErr
	}

	var tenantExists bool

	tenantErr := tx.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM tenants WHERE slug = $1)
	`, p.TenantSlug).Scan(&tenantExists)
	if tenantErr != nil {
		return nil, tenantErr
	}

	if tenantExists {
		return nil, ErrTenantSlugTaken
	}

	existing, err := findActiveOnboardingRequestForSlug(ctx, tx, p)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		commitErr := tx.Commit(ctx)
		if commitErr != nil {
			return nil, commitErr
		}

		return existing, nil
	}

	code, err := numericOTP(6)
	if err != nil {
		return nil, err
	}

	metadata, err := json.Marshal(p.Metadata)
	if err != nil {
		return nil, err
	}

	requestID := uuid.NewString()

	expiresAt := time.Now().UTC().Add(onboardingOTPTTL)
	if _, err := tx.Exec(ctx, `
		INSERT INTO tenant_onboarding_requests (
			request_id, tenant_slug, tenant_domain, plan_slug, email, mobile, customer_name,
			otp_code, expires_at, metadata
		) VALUES ($1::uuid, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''), $8, $9, $10::jsonb)
	`, requestID, p.TenantSlug, p.TenantDomain, p.PlanSlug, p.Email, p.Mobile, p.CustomerName, code, expiresAt, string(metadata)); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	channel := "email"

	destination := p.Email
	if destination == "" {
		channel = "mobile"
		destination = p.Mobile
	}

	s.logger.Info("onboarding otp generated")

	return &OnboardingOTP{
		RequestID:         requestID,
		TenantSlug:        p.TenantSlug,
		Channel:           channel,
		DestinationMasked: maskDestination(channel, destination),
		ExpiresInSeconds:  int(onboardingOTPTTL.Seconds()),
		DebugOTP:          code,
	}, nil
}

func (s *Store) expirePendingOnboardingRequests(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE tenant_onboarding_requests
		SET status = 'expired', updated_at = NOW()
		WHERE status = 'pending' AND expires_at <= NOW()
	`)

	return err
}

func findActiveOnboardingRequestForSlug(ctx context.Context, tx pgx.Tx, p CreateOnboardingOTPParams) (*OnboardingOTP, error) {
	rows, err := tx.Query(ctx, `
		SELECT request_id::text, tenant_slug, COALESCE(email, ''), COALESCE(mobile, ''), expires_at
		FROM tenant_onboarding_requests
		WHERE tenant_slug = $1 AND status = 'pending' AND expires_at > NOW()
		FOR UPDATE
	`, p.TenantSlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	now := time.Now().UTC()

	var taken bool

	for rows.Next() {
		var (
			requestID, tenantSlug, email, mobile string
			expiresAt                            time.Time
		)
		if err := rows.Scan(&requestID, &tenantSlug, &email, &mobile, &expiresAt); err != nil {
			return nil, err
		}

		if sameOnboardingRequester(email, mobile, p.Email, p.Mobile) {
			channel := "email"

			destination := email
			if destination == "" {
				channel = "mobile"
				destination = mobile
			}

			remaining := int(time.Until(expiresAt).Seconds())
			if remaining < 1 && expiresAt.After(now) {
				remaining = 1
			}

			return &OnboardingOTP{
				RequestID:         requestID,
				TenantSlug:        tenantSlug,
				Channel:           channel,
				DestinationMasked: maskDestination(channel, destination),
				ExpiresInSeconds:  remaining,
				AlreadyPending:    true,
			}, rows.Err()
		}

		taken = true
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if taken {
		return nil, ErrTenantSlugTaken
	}

	return nil, nil
}

func sameOnboardingRequester(existingEmail, existingMobile, requestedEmail, requestedMobile string) bool {
	if existingEmail != "" && requestedEmail != "" && existingEmail == requestedEmail {
		return true
	}

	if existingMobile != "" && requestedMobile != "" && existingMobile == requestedMobile {
		return true
	}

	return false
}

func (s *Store) VerifyOnboardingOTPAndProvision(ctx context.Context, p VerifyOnboardingOTPParams) (*ProvisionedTenant, error) {
	requestID := strings.TrimSpace(p.RequestID)

	otpCode := strings.TrimSpace(p.OTPCode)
	if requestID == "" || otpCode == "" {
		return nil, ErrOnboardingNotFound
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var (
		tenantSlug, tenantDomain, planSlug, email, mobile string
		otp                                               string
		attempts                                          int
		expiresAt                                         time.Time
	)

	requestErr := tx.QueryRow(ctx, `
		SELECT tenant_slug, tenant_domain, COALESCE(plan_slug, 'developer'), COALESCE(email, ''), COALESCE(mobile, ''),
		       otp_code, attempts, expires_at
		FROM tenant_onboarding_requests
		WHERE request_id = $1::uuid AND status = 'pending'
		FOR UPDATE
	`, requestID).Scan(&tenantSlug, &tenantDomain, &planSlug, &email, &mobile, &otp, &attempts, &expiresAt)
	if requestErr != nil {
		if errors.Is(requestErr, pgx.ErrNoRows) {
			return nil, ErrOnboardingNotFound
		}

		return nil, requestErr
	}

	if time.Now().UTC().After(expiresAt) {
		_, _ = tx.Exec(ctx, `
			UPDATE tenant_onboarding_requests
			SET status = 'expired', updated_at = NOW()
			WHERE request_id = $1::uuid
		`, requestID)

		return nil, ErrOnboardingExpired
	}

	if otpCode != otp {
		nextAttempts := attempts + 1

		status := "pending"
		if nextAttempts >= maxOnboardingAttempts {
			status = "canceled"
		}

		_, _ = tx.Exec(ctx, `
			UPDATE tenant_onboarding_requests
			SET attempts = $2, status = $3, updated_at = NOW()
			WHERE request_id = $1::uuid
		`, requestID, nextAttempts, status)

		return nil, ErrOnboardingOTPFailed
	}

	var created platform.Tenant

	tenantCreateErr := tx.QueryRow(ctx, `
		INSERT INTO tenants (slug, domain, status)
		VALUES ($1, $2, 'active')
		RETURNING tenant_id::text, slug, domain, status, created_at
	`, tenantSlug, tenantDomain).Scan(&created.ID, &created.Slug, &created.Domain, &created.Status, &created.CreatedAt)
	if tenantCreateErr != nil {
		return nil, tenantCreateErr
	}

	var walletID string

	walletErr := tx.QueryRow(ctx, `
		INSERT INTO wallets (tenant_id, balance_irr, pending_irr)
		SELECT $1::uuid, signup_credit_irr::double precision, 0
		FROM commercial_signup_settings WHERE settings_id = true
		RETURNING wallet_id::text
	`, created.ID).Scan(&walletID)
	if walletErr != nil {
		return nil, walletErr
	}

	requestedPlanSlug := strings.TrimSpace(strings.ToLower(planSlug))
	if requestedPlanSlug == "" || requestedPlanSlug == "free" {
		requestedPlanSlug = "developer"
	}

	subscriptionResult, err := tx.Exec(ctx, `
		INSERT INTO tenant_subscriptions (
			tenant_id, plan_id, version_id, status, billing_interval,
			current_period_start, current_period_end, trial_started_at, trial_ends_at,
			price_snapshot_json, metadata
		)
		SELECT $1::uuid, p.plan_id, pv.version_id, 'trialing', 'monthly',
		       t.created_at, t.created_at + make_interval(days => settings.trial_days),
		       t.created_at, t.created_at + make_interval(days => settings.trial_days),
		       jsonb_build_object(
		           'currency', 'IRR',
		           'amount_minor', 0,
		           'billing_interval', 'monthly',
		           'discount_percent', 100,
		           'regular_amount_minor', COALESCE(pp.amount_minor, 0)
		       ),
		       jsonb_build_object('source', 'onboarding', 'trial_days', settings.trial_days)
		FROM commercial_plans p
		JOIN plan_versions pv ON pv.plan_id = p.plan_id AND pv.status = 'published'
		LEFT JOIN plan_prices pp ON pp.version_id = pv.version_id AND pp.billing_interval = 'monthly'
		JOIN tenants t ON t.tenant_id = $1::uuid
		CROSS JOIN commercial_signup_settings settings
		WHERE p.slug = $2 AND p.active = true AND settings.settings_id = true
	`, created.ID, requestedPlanSlug)
	if err != nil {
		return nil, err
	}

	if subscriptionResult.RowsAffected() == 0 {
		return nil, ErrPlanNotFound
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO subscription_events (tenant_id, subscription_id, event_type, actor, payload)
		SELECT tenant_id, subscription_id, 'trial_started', 'onboarding',
		       jsonb_build_object('plan_slug', $2::text, 'trial_ends_at', trial_ends_at)
		FROM tenant_subscriptions
		WHERE tenant_id = $1::uuid AND status = 'trialing'
	`, created.ID, requestedPlanSlug); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO wallet_transactions (
			wallet_id, tenant_id, tx_type, amount_irr, status, description, metadata
		)
		SELECT $1::uuid, $2::uuid, 'SIGNUP_CREDIT', signup_credit_irr::double precision,
		       'COMPLETED', 'Promotional signup credit',
		       jsonb_build_object('source', 'onboarding', 'plan_slug', $3::text)
		FROM commercial_signup_settings
		WHERE settings_id = true AND signup_credit_irr > 0
	`, walletID, created.ID, requestedPlanSlug); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO users (
			user_id, tenant_id, user_alias, email, mobile, role, active, created_at, updated_at
		)
		VALUES (
			$1::uuid, $2::uuid, NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''), 'tenant_admin', true, NOW(), NOW()
		)
		ON CONFLICT DO NOTHING
	`, uuid.NewString(), created.ID, strings.TrimSpace(email), strings.TrimSpace(email), strings.TrimSpace(mobile)); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `
		UPDATE tenant_onboarding_requests
		SET status = 'verified', verified_at = NOW(), tenant_id = $2::uuid, updated_at = NOW()
		WHERE request_id = $1::uuid
	`, requestID, created.ID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &ProvisionedTenant{Tenant: created, LoginURL: fmt.Sprintf("https://%s/login", created.Domain)}, nil
}

func normalizeOnboardingSlug(slug string) string {
	return strings.Trim(strings.ToLower(strings.TrimSpace(slug)), ".")
}

func numericOTP(length int) (string, error) {
	var b strings.Builder

	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}

		b.WriteString(n.String())
	}

	return b.String(), nil
}

func maskDestination(channel string, value string) string {
	if channel == "email" {
		parts := strings.Split(value, "@")
		if len(parts) != 2 {
			return "***"
		}

		local := parts[0]
		if len(local) <= 2 {
			return local[:1] + "***@" + parts[1]
		}

		return local[:2] + "***@" + parts[1]
	}

	if len(value) <= 6 {
		return "***"
	}

	return value[:3] + "****" + value[len(value)-3:]
}
