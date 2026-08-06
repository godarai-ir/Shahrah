package storage

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	billingdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	socialLoginHandoffTTL = 90 * time.Second
	socialLinkTicketTTL   = 5 * time.Minute
)

var (
	ErrSocialAccountInvalid = errors.New("social account is missing or its email is not verified")
	ErrSocialAccessDenied   = errors.New("social account is not linked to this tenant")
	ErrSocialHandoffInvalid = errors.New("social handoff is invalid, expired, or already consumed")
	ErrSocialLinkConflict   = errors.New("user or social account is already linked to another account")
)

type ProvisionSocialTenantParams struct {
	AccountID    string `json:"account_id"`
	TenantSlug   string `json:"tenant_slug"`
	TenantDomain string `json:"tenant_domain"`
	PlanSlug     string `json:"plan_slug"`
}

type SocialHandoff struct {
	Code       string    `json:"code"`
	TenantSlug string    `json:"tenant_slug"`
	UserID     string    `json:"user_id"`
	ExpiresAt  time.Time `json:"expires_at"`
}

type RedeemedSocialHandoff struct {
	Session map[string]interface{} `json:"session"`
	User    map[string]interface{} `json:"user"`
}

// ProvisionSocialTenant atomically provisions a tenant for a Better Auth user
// whose provider has supplied a verified email, then issues a short-lived
// one-time handoff. OAuth tokens never leave the central auth service.
//
//nolint:govet // Short-lived branch errors are checked immediately and intentionally scoped with short declarations.
func (s *Store) ProvisionSocialTenant(ctx context.Context, p ProvisionSocialTenantParams) (*ProvisionedTenant, error) {
	p.AccountID = strings.TrimSpace(p.AccountID)
	p.TenantSlug = normalizeOnboardingSlug(p.TenantSlug)
	p.TenantDomain = strings.TrimSpace(strings.ToLower(p.TenantDomain))

	p.PlanSlug = strings.TrimSpace(strings.ToLower(p.PlanSlug))
	if p.PlanSlug == "" || p.PlanSlug == "free" {
		p.PlanSlug = "developer"
	}

	if p.TenantDomain == "" {
		p.TenantDomain = p.TenantSlug
	}

	if p.AccountID == "" || !tenant.ValidateTenantSlug(p.TenantSlug) {
		return nil, ErrTenantSlugInvalid
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	var (
		email, name   string
		emailVerified bool
	)
	if err := tx.QueryRow(ctx, `
		SELECT LOWER(email), name, "emailVerified"
		FROM social_auth_users WHERE id = $1
		FOR UPDATE
	`, p.AccountID).Scan(&email, &name, &emailVerified); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSocialAccountInvalid
		}

		return nil, err
	}

	if !emailVerified || email == "" {
		return nil, ErrSocialAccountInvalid
	}

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, p.TenantSlug); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `
		UPDATE tenant_onboarding_requests SET status = 'expired', updated_at = NOW()
		WHERE status = 'pending' AND expires_at <= NOW()
	`); err != nil {
		return nil, err
	}

	var exists bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM tenants WHERE slug = $1
			UNION ALL
			SELECT 1 FROM tenant_onboarding_requests
			WHERE tenant_slug = $1 AND status = 'pending' AND expires_at > NOW()
		)
	`, p.TenantSlug).Scan(&exists); err != nil {
		return nil, err
	}

	if exists {
		return nil, ErrTenantSlugTaken
	}

	var created platform.Tenant
	if err := tx.QueryRow(ctx, `
		INSERT INTO tenants (slug, domain, status) VALUES ($1, $2, 'active')
		RETURNING tenant_id::text, slug, domain, status, created_at
	`, p.TenantSlug, p.TenantDomain).Scan(&created.ID, &created.Slug, &created.Domain, &created.Status, &created.CreatedAt); err != nil {
		return nil, err
	}

	var walletID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO wallets (tenant_id, balance_irr, pending_irr)
		SELECT $1::uuid, signup_credit_irr::double precision, 0
		FROM commercial_signup_settings WHERE settings_id = true
		RETURNING wallet_id::text
	`, created.ID).Scan(&walletID); err != nil {
		return nil, err
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
		       jsonb_build_object('currency', 'IRR', 'amount_minor', 0, 'billing_interval', 'monthly',
		          'discount_percent', 100, 'regular_amount_minor', COALESCE(pp.amount_minor, 0)),
		       jsonb_build_object('source', 'social_onboarding', 'trial_days', settings.trial_days)
		FROM commercial_plans p
		JOIN plan_versions pv ON pv.plan_id = p.plan_id AND pv.status = 'published'
		LEFT JOIN plan_prices pp ON pp.version_id = pv.version_id AND pp.billing_interval = 'monthly'
		JOIN tenants t ON t.tenant_id = $1::uuid
		CROSS JOIN commercial_signup_settings settings
		WHERE p.slug = $2 AND p.active = true AND settings.settings_id = true
	`, created.ID, p.PlanSlug)
	if err != nil {
		return nil, err
	}

	if subscriptionResult.RowsAffected() == 0 {
		return nil, billingdomain.ErrPlanNotFound
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO subscription_events (tenant_id, subscription_id, event_type, actor, payload)
		SELECT tenant_id, subscription_id, 'trial_started', 'social_onboarding',
		       jsonb_build_object('plan_slug', $2::text, 'trial_ends_at', trial_ends_at)
		FROM tenant_subscriptions WHERE tenant_id = $1::uuid AND status = 'trialing'
	`, created.ID, p.PlanSlug); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO wallet_transactions (wallet_id, tenant_id, tx_type, amount_irr, status, description, metadata)
		SELECT $1::uuid, $2::uuid, 'SIGNUP_CREDIT', signup_credit_irr::double precision,
		       'COMPLETED', 'Promotional signup credit',
		       jsonb_build_object('source', 'social_onboarding', 'plan_slug', $3::text)
		FROM commercial_signup_settings WHERE settings_id = true AND signup_credit_irr > 0
	`, walletID, created.ID, p.PlanSlug); err != nil {
		return nil, err
	}

	userID := uuid.NewString()
	if _, err := tx.Exec(ctx, `
		INSERT INTO users (user_id, tenant_id, account_id, user_alias, email, role, active, created_at, updated_at)
		VALUES ($1::uuid, $2::uuid, $3, NULLIF($4, ''), $5, 'tenant_admin', true, NOW(), NOW())
	`, userID, created.ID, p.AccountID, strings.TrimSpace(name), email); err != nil {
		return nil, err
	}

	handoff, err := createSocialHandoffTx(ctx, tx, "login", p.AccountID, created.ID, created.Slug, userID, socialLoginHandoffTTL)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &ProvisionedTenant{
		Tenant:   created,
		LoginURL: fmt.Sprintf("https://%s/auth/handoff?code=%s", created.Domain, handoff.Code),
	}, nil
}

//nolint:govet // Short-lived branch errors are checked immediately and intentionally scoped with short declarations.
func (s *Store) CreateSocialLoginHandoff(ctx context.Context, tenantSlug, accountID string) (*SocialHandoff, error) {
	tenantSlug = normalizeOnboardingSlug(tenantSlug)
	accountID = strings.TrimSpace(accountID)

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	var tenantID, userID string
	if err := tx.QueryRow(ctx, `
		SELECT t.tenant_id::text, u.user_id::text
		FROM tenants t JOIN users u ON u.tenant_id = t.tenant_id
		WHERE t.slug = $1 AND u.account_id = $2 AND COALESCE(u.active, true) = true
		  AND LOWER(COALESCE(u.role, '')) NOT IN ('deactivated', 'removed', 'suspended')
	`, tenantSlug, accountID).Scan(&tenantID, &userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSocialAccessDenied
		}

		return nil, err
	}

	handoff, err := createSocialHandoffTx(ctx, tx, "login", accountID, tenantID, tenantSlug, userID, socialLoginHandoffTTL)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return handoff, nil
}

//nolint:govet // Short-lived branch errors are checked immediately and intentionally scoped with short declarations.
func (s *Store) CreateSocialLinkTicket(ctx context.Context, tenantSlug, userID string) (*SocialHandoff, error) {
	tenantSlug = normalizeOnboardingSlug(tenantSlug)
	userID = strings.TrimSpace(userID)

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	var tenantID string
	if err := tx.QueryRow(ctx, `
		SELECT t.tenant_id::text FROM tenants t JOIN users u ON u.tenant_id = t.tenant_id
		WHERE t.slug = $1 AND u.user_id::text = $2 AND COALESCE(u.active, true) = true
	`, tenantSlug, userID).Scan(&tenantID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSocialAccessDenied
		}

		return nil, err
	}

	ticket, err := createSocialHandoffTx(ctx, tx, "link", "", tenantID, tenantSlug, userID, socialLinkTicketTTL)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return ticket, nil
}

//nolint:govet // Short-lived branch errors are checked immediately and intentionally scoped with short declarations.
func (s *Store) CompleteSocialLink(ctx context.Context, ticketCode, accountID string) (*SocialHandoff, error) {
	accountID = strings.TrimSpace(accountID)

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	var verified bool
	if err := tx.QueryRow(ctx, `SELECT "emailVerified" FROM social_auth_users WHERE id = $1`, accountID).Scan(&verified); err != nil || !verified {
		return nil, ErrSocialAccountInvalid
	}

	hash := hashSocialHandoffCode(ticketCode)

	var tenantID, tenantSlug, userID string
	if err := tx.QueryRow(ctx, `
		SELECT h.tenant_id::text, t.slug, h.user_id::text
		FROM social_auth_handoffs h JOIN tenants t ON t.tenant_id = h.tenant_id
		WHERE h.code_hash = $1 AND h.purpose = 'link' AND h.consumed_at IS NULL AND h.expires_at > NOW()
		FOR UPDATE OF h
	`, hash).Scan(&tenantID, &tenantSlug, &userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSocialHandoffInvalid
		}

		return nil, err
	}

	result, err := tx.Exec(ctx, `
		UPDATE users SET account_id = $1, updated_at = NOW()
		WHERE user_id = $2::uuid AND tenant_id = $3::uuid AND (account_id IS NULL OR account_id = $1)
	`, accountID, userID, tenantID)
	if err != nil || result.RowsAffected() != 1 {
		return nil, ErrSocialLinkConflict
	}

	if _, err := tx.Exec(ctx, `UPDATE social_auth_handoffs SET consumed_at = NOW() WHERE code_hash = $1`, hash); err != nil {
		return nil, err
	}

	handoff, err := createSocialHandoffTx(ctx, tx, "login", accountID, tenantID, tenantSlug, userID, socialLoginHandoffTTL)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return handoff, nil
}

//nolint:govet // Short-lived branch errors are checked immediately and intentionally scoped with short declarations.
func (s *Store) RedeemSocialLoginHandoff(ctx context.Context, tenantSlug, code string) (*RedeemedSocialHandoff, error) {
	tenantSlug = normalizeOnboardingSlug(tenantSlug)

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	var userID string
	if err := tx.QueryRow(ctx, `
		SELECT h.user_id::text
		FROM social_auth_handoffs h JOIN tenants t ON t.tenant_id = h.tenant_id
		JOIN users u ON u.user_id = h.user_id AND u.tenant_id = h.tenant_id
		WHERE h.code_hash = $1 AND h.purpose = 'login' AND t.slug = $2
		  AND h.consumed_at IS NULL AND h.expires_at > NOW() AND COALESCE(u.active, true) = true
		FOR UPDATE OF h
	`, hashSocialHandoffCode(code), tenantSlug).Scan(&userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSocialHandoffInvalid
		}

		return nil, err
	}

	if _, err := tx.Exec(ctx, `UPDATE social_auth_handoffs SET consumed_at = NOW() WHERE code_hash = $1`, hashSocialHandoffCode(code)); err != nil {
		return nil, err
	}

	sessionID, err := randomHex(24)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	expiresAt := now.Add(tenantAdminSessionTTL)
	if _, err := tx.Exec(ctx, `
		INSERT INTO tenant_admin_sessions (id, "tenantSlug", "userID", "expiresAt") VALUES ($1, $2, $3, $4)
	`, sessionID, tenantSlug, userID, expiresAt); err != nil {
		return nil, err
	}

	var (
		email, mobile, role  string
		active               bool
		createdAt, updatedAt time.Time
	)
	if err := tx.QueryRow(ctx, `
		SELECT COALESCE(email,''), COALESCE(mobile,''), COALESCE(role,'member'), COALESCE(active,true), created_at, updated_at
		FROM users WHERE user_id = $1::uuid
	`, userID).Scan(&email, &mobile, &role, &active, &createdAt, &updatedAt); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &RedeemedSocialHandoff{
		Session: map[string]interface{}{"id": sessionID, "tenant_slug": tenantSlug, "user_id": userID, "created_at_iso": now.Format(time.RFC3339), "expires_at_iso": expiresAt.Format(time.RFC3339)},
		User:    buildTenantAdminUserMap(tenantSlug, userID, email, mobile, role, active, createdAt, updatedAt),
	}, nil
}

func createSocialHandoffTx(ctx context.Context, tx pgx.Tx, purpose, accountID, tenantID, tenantSlug, userID string, ttl time.Duration) (*SocialHandoff, error) {
	codeBytes := make([]byte, 32)
	if _, err := rand.Read(codeBytes); err != nil {
		return nil, err
	}

	code := base64.RawURLEncoding.EncodeToString(codeBytes)
	expiresAt := time.Now().UTC().Add(ttl)

	var nullableAccount interface{}
	if accountID != "" {
		nullableAccount = accountID
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO social_auth_handoffs (code_hash, purpose, account_id, tenant_id, user_id, expires_at)
		VALUES ($1, $2, $3, $4::uuid, $5::uuid, $6)
	`, hashSocialHandoffCode(code), purpose, nullableAccount, tenantID, userID, expiresAt); err != nil {
		return nil, err
	}

	return &SocialHandoff{Code: code, TenantSlug: tenantSlug, UserID: userID, ExpiresAt: expiresAt}, nil
}

func hashSocialHandoffCode(code string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(code)))
	return hex.EncodeToString(sum[:])
}

func randomHex(size int) (string, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}
