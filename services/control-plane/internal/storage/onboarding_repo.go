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
		p.PlanSlug = "free"
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

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, p.TenantSlug); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE tenant_onboarding_requests
		SET status = 'expired', updated_at = NOW()
		WHERE status = 'pending' AND expires_at <= NOW()
	`); err != nil {
		return nil, err
	}

	var tenantExists bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM tenants WHERE slug = $1)
	`, p.TenantSlug).Scan(&tenantExists); err != nil {
		return nil, err
	}
	if tenantExists {
		return nil, ErrTenantSlugTaken
	}

	existing, err := findActiveOnboardingRequestForSlug(ctx, tx, p)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
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
		var requestID, tenantSlug, email, mobile string
		var expiresAt time.Time
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

	var tenantSlug, tenantDomain, email, mobile string
	var otp string
	var attempts int
	var expiresAt time.Time
	if err := tx.QueryRow(ctx, `
		SELECT tenant_slug, tenant_domain, COALESCE(email, ''), COALESCE(mobile, ''),
		       otp_code, attempts, expires_at
		FROM tenant_onboarding_requests
		WHERE request_id = $1::uuid AND status = 'pending'
		FOR UPDATE
	`, requestID).Scan(&tenantSlug, &tenantDomain, &email, &mobile, &otp, &attempts, &expiresAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOnboardingNotFound
		}
		return nil, err
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
	if err := tx.QueryRow(ctx, `
		INSERT INTO tenants (slug, domain, status)
		VALUES ($1, $2, 'active')
		RETURNING tenant_id::text, slug, domain, status, created_at
	`, tenantSlug, tenantDomain).Scan(&created.ID, &created.Slug, &created.Domain, &created.Status, &created.CreatedAt); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO wallets (tenant_id, balance_irr, pending_irr)
		VALUES ($1::uuid, 0, 0)
	`, created.ID); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO users (
			user_id, tenant_id, user_alias, email, mobile, role, active, created_at, updated_at
		)
		VALUES (
			$1::uuid, $2::uuid, NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''), 'tenant_admin', true, NOW(), NOW()
		)
		ON CONFLICT (email) DO NOTHING
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
