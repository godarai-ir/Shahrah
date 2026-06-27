package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// TenantAdminCaptcha represents a captcha challenge.
type TenantAdminCaptcha struct {
	ID        string
	Answer    string
	ExpiresAt time.Time
}

// TenantAdminOTPRequest represents an OTP request.
type TenantAdminOTPRequest struct {
	ID           string
	TenantSlug   string
	Purpose      string
	Channel      string
	Identifier   string
	UserID       *string
	PendingValue *string
	OTPCode      string
	Attempts     int
	ExpiresAt    time.Time
	CreatedAt    time.Time
}

// TenantAdminSession represents a user session.
type TenantAdminSession struct {
	ID         string
	TenantSlug string
	UserID     string
	ExpiresAt  time.Time
	CreatedAt  time.Time
}

const (
	tenantAdminCaptchaTTL     = 2 * time.Minute
	tenantAdminOTPTTL         = 5 * time.Minute
	tenantAdminSessionTTL     = 7 * 24 * time.Hour
	tenantAdminMaxOTPAttempts = 5
)

// CreateCaptcha creates a captcha challenge and returns the ID, answer, and expiry.
func (s *Store) CreateTenantAdminCaptcha(ctx context.Context) (string, string, time.Time, error) {
	id := generateTenantAdminID()
	answer := generateCaptchaAnswer(5)
	expiresAt := time.Now().Add(tenantAdminCaptchaTTL)

	_, err := s.pool.Exec(ctx, `
		INSERT INTO tenant_admin_captchas (id, answer, "expiresAt")
		VALUES ($1, $2, $3)
	`, id, answer, expiresAt)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("create captcha: %w", err)
	}

	return id, answer, expiresAt, nil
}

// ConsumeTenantAdminCaptcha verifies and deletes the captcha.
func (s *Store) ConsumeTenantAdminCaptcha(ctx context.Context, id string, answer string) (bool, error) {
	var (
		storedAnswer string
		expiresAt    time.Time
	)

	queryErr := s.pool.QueryRow(ctx, `
		SELECT answer, "expiresAt" FROM tenant_admin_captchas WHERE id = $1
	`, id).Scan(&storedAnswer, &expiresAt)

	if hasInvalidOTPRequest(queryErr) {
		if queryErr == pgx.ErrNoRows {
			return false, nil
		}

		return false, queryErr
	}

	_, _ = s.pool.Exec(ctx, `DELETE FROM tenant_admin_captchas WHERE id = $1`, id)

	if time.Now().After(expiresAt) {
		return false, nil
	}

	return storedAnswer == answer, nil
}

// TenantAdminOTPRequestParams holds parameters for creating an OTP request.
type TenantAdminOTPRequestParams struct {
	TenantSlug   string
	Purpose      string
	Channel      string
	Identifier   string
	UserID       *string
	PendingValue *string
}

// CreateTenantAdminOTPRequest creates an OTP request.
func (s *Store) CreateTenantAdminOTPRequest(ctx context.Context, params TenantAdminOTPRequestParams) (map[string]interface{}, error) {
	id := generateTenantAdminID()
	otpCode := generateOTPCode()
	expiresAt := time.Now().Add(tenantAdminOTPTTL)

	channel := params.Channel
	if channel == "" {
		channel = "email"
	}

	_, err := s.pool.Exec(ctx, `
		INSERT INTO tenant_admin_otp_requests
			(id, "tenantSlug", purpose, channel, identifier, "userID", "pendingValue", "otpCode", attempts, "expiresAt")
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 0, $9)
	`, id, params.TenantSlug, params.Purpose, channel, params.Identifier, params.UserID, params.PendingValue, otpCode, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("create OTP request: %w", err)
	}

	return map[string]interface{}{
		"request_id":         id,
		"channel":            channel,
		"destination_masked": maskTenantAdminIdentifier(channel, params.Identifier),
		"expires_in_seconds": int(tenantAdminOTPTTL.Seconds()),
		"otp_code":           otpCode,
	}, nil
}

// VerifyTenantAdminOTPRequestParams holds parameters for verifying an OTP.
type VerifyTenantAdminOTPRequestParams struct {
	TenantSlug string
	RequestID  string
	OTPCode    string
	Purpose    string
	UserID     *string
}

// VerifyTenantAdminOTPRequest verifies an OTP request.
func (s *Store) VerifyTenantAdminOTPRequest(ctx context.Context, params VerifyTenantAdminOTPRequestParams) (map[string]interface{}, error) {
	var (
		record  TenantAdminOTPRequest
		otpCode string
	)

	queryErr := s.pool.QueryRow(ctx, `
		SELECT id, "tenantSlug", purpose, channel, identifier, "userID",
		       "pendingValue", "otpCode", attempts, "expiresAt", "createdAt"
		FROM tenant_admin_otp_requests WHERE id = $1
	`, params.RequestID).Scan(
		&record.ID, &record.TenantSlug, &record.Purpose, &record.Channel,
		&record.Identifier, &record.UserID, &record.PendingValue,
		&otpCode, &record.Attempts, &record.ExpiresAt, &record.CreatedAt,
	)

	invalidRequest := queryErr != nil
	if invalidRequest {
		return map[string]interface{}{"ok": false, "error": "درخواست OTP معتبر نیست یا منقضی شده است."}, nil
	}

	if record.TenantSlug != params.TenantSlug {
		return map[string]interface{}{"ok": false, "error": "این درخواست OTP مربوط به Tenant فعلی نیست."}, nil
	}

	if record.Purpose != params.Purpose {
		return map[string]interface{}{"ok": false, "error": "کاربرد OTP با درخواست فعلی هم\u200cخوانی ندارد."}, nil
	}

	if params.UserID != nil && (record.UserID == nil || *record.UserID != *params.UserID) {
		return map[string]interface{}{"ok": false, "error": "این درخواست OTP متعلق به کاربر فعلی نیست."}, nil
	}

	if time.Now().After(record.ExpiresAt) {
		_, _ = s.pool.Exec(ctx, `DELETE FROM tenant_admin_otp_requests WHERE id = $1`, params.RequestID)
		return map[string]interface{}{"ok": false, "error": "کد OTP منقضی شده است."}, nil
	}

	if params.OTPCode != otpCode {
		nextAttempts := record.Attempts + 1
		if nextAttempts >= tenantAdminMaxOTPAttempts {
			_, _ = s.pool.Exec(ctx, `DELETE FROM tenant_admin_otp_requests WHERE id = $1`, params.RequestID)
			return map[string]interface{}{"ok": false, "error": "تعداد تلاش بیش از حد مجاز است. دوباره OTP بگیرید."}, nil
		}

		_, _ = s.pool.Exec(ctx, `UPDATE tenant_admin_otp_requests SET attempts = $1 WHERE id = $2`, nextAttempts, params.RequestID)

		return map[string]interface{}{"ok": false, "error": "کد OTP اشتباه است."}, nil
	}

	// OTP is correct - delete the record
	_, _ = s.pool.Exec(ctx, `DELETE FROM tenant_admin_otp_requests WHERE id = $1`, params.RequestID)

	return map[string]interface{}{
		"ok": true,
		"record": map[string]interface{}{
			"id":             record.ID,
			"tenant_slug":    record.TenantSlug,
			"purpose":        record.Purpose,
			"channel":        record.Channel,
			"identifier":     record.Identifier,
			"user_id":        record.UserID,
			"pending_value":  record.PendingValue,
			"otp_code":       otpCode,
			"attempts":       record.Attempts,
			"expires_at_iso": record.ExpiresAt.Format(time.RFC3339),
			"created_at_iso": record.CreatedAt.Format(time.RFC3339),
		},
	}, nil
}

// CreateTenantAdminSessionParams holds parameters for creating a session.
type CreateTenantAdminSessionParams struct {
	TenantSlug string
	UserID     string
}

// CreateTenantAdminSession creates a session.
func (s *Store) CreateTenantAdminSession(ctx context.Context, params CreateTenantAdminSessionParams) (map[string]interface{}, error) {
	id, err := generateTenantAdminSessionID()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(tenantAdminSessionTTL)
	now := time.Now()

	_, err = s.pool.Exec(ctx, `
		INSERT INTO tenant_admin_sessions (id, "tenantSlug", "userID", "expiresAt")
		VALUES ($1, $2, $3, $4)
	`, id, params.TenantSlug, params.UserID, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return map[string]interface{}{
		"id":             id,
		"tenant_slug":    params.TenantSlug,
		"user_id":        params.UserID,
		"created_at_iso": now.Format(time.RFC3339),
		"expires_at_iso": expiresAt.Format(time.RFC3339),
	}, nil
}

// GetTenantAdminSession retrieves a session by ID.
func (s *Store) GetTenantAdminSession(ctx context.Context, id string) (map[string]interface{}, error) {
	var session TenantAdminSession

	queryErr := s.pool.QueryRow(ctx, `
		SELECT id, "tenantSlug", "userID", "expiresAt", "createdAt"
		FROM tenant_admin_sessions WHERE id = $1
	`, id).Scan(&session.ID, &session.TenantSlug, &session.UserID, &session.ExpiresAt, &session.CreatedAt)
	if queryErr != nil {
		if queryErr == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, queryErr
	}

	if time.Now().After(session.ExpiresAt) {
		_, _ = s.pool.Exec(ctx, `DELETE FROM tenant_admin_sessions WHERE id = $1`, id)
		return nil, nil
	}

	return map[string]interface{}{
		"id":             session.ID,
		"tenant_slug":    session.TenantSlug,
		"user_id":        session.UserID,
		"created_at_iso": session.CreatedAt.Format(time.RFC3339),
		"expires_at_iso": session.ExpiresAt.Format(time.RFC3339),
	}, nil
}

// RevokeTenantAdminSession deletes a session.
func (s *Store) RevokeTenantAdminSession(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM tenant_admin_sessions WHERE id = $1`, id)
	return err
}

// FindOrCreateTenantAdminUserParams holds parameters for finding or creating a user.
type FindOrCreateTenantAdminUserParams struct {
	TenantSlug string
	Channel    string
	Identifier string
}

// FindOrCreateTenantAdminUser finds an active tenant user for login.
func (s *Store) FindOrCreateTenantAdminUser(ctx context.Context, params FindOrCreateTenantAdminUserParams) (map[string]interface{}, error) {
	var (
		id, slug, email, mobile, role string
		active                        bool
		createdAt, updatedAt          time.Time
	)

	tenantSlug := strings.TrimSpace(strings.ToLower(params.TenantSlug))
	channel := strings.TrimSpace(strings.ToLower(params.Channel))
	identifier := strings.TrimSpace(params.Identifier)

	if channel == "mobile" {
		err := s.pool.QueryRow(ctx, `
			SELECT u.user_id::text, t.slug, COALESCE(u.email, ''), COALESCE(u.mobile, ''), COALESCE(u.role, 'member'), COALESCE(u.active, true),
			       u.created_at, u.updated_at
			FROM users u
			JOIN tenants t ON t.tenant_id = u.tenant_id
			WHERE t.slug = $1
			  AND COALESCE(u.mobile, '') = $2
			  AND COALESCE(u.active, true) = true
			  AND LOWER(COALESCE(u.role, '')) NOT IN ('deactivated', 'removed', 'suspended')
			LIMIT 1
		`, tenantSlug, identifier).Scan(&id, &slug, &email, &mobile, &role, &active, &createdAt, &updatedAt)
		if err == nil {
			return buildTenantAdminUserMap(slug, id, email, mobile, role, active, createdAt, updatedAt), nil
		}
	} else {
		err := s.pool.QueryRow(ctx, `
			SELECT u.user_id::text, t.slug, COALESCE(u.email, ''), COALESCE(u.mobile, ''), COALESCE(u.role, 'member'), COALESCE(u.active, true),
			       u.created_at, u.updated_at
			FROM users u
			JOIN tenants t ON t.tenant_id = u.tenant_id
			WHERE t.slug = $1
			  AND LOWER(COALESCE(u.email, '')) = LOWER($2)
			  AND COALESCE(u.active, true) = true
			  AND LOWER(COALESCE(u.role, '')) NOT IN ('deactivated', 'removed', 'suspended')
			LIMIT 1
		`, tenantSlug, identifier).Scan(&id, &slug, &email, &mobile, &role, &active, &createdAt, &updatedAt)
		if err == nil {
			return buildTenantAdminUserMap(slug, id, email, mobile, role, active, createdAt, updatedAt), nil
		}
	}

	return nil, fmt.Errorf("tenant member not found")
}

// GetTenantAdminUser retrieves a user by ID.
func (s *Store) GetTenantAdminUser(ctx context.Context, tenantID string, userID string) (map[string]interface{}, error) {
	var (
		id, slug, email, mobile, role string
		active                        bool
		createdAt, updatedAt          time.Time
	)

	queryErr := s.pool.QueryRow(ctx, `
		SELECT u.user_id::text, t.slug, COALESCE(u.email, ''), COALESCE(u.mobile, ''),
		       COALESCE(u.role, 'member'), COALESCE(u.active, true), u.created_at, u.updated_at
		FROM users u
		JOIN tenants t ON t.tenant_id = u.tenant_id
		WHERE t.slug = $1 AND u.user_id::text = $2
		LIMIT 1
	`, strings.TrimSpace(strings.ToLower(tenantID)), strings.TrimSpace(userID)).Scan(&id, &slug, &email, &mobile, &role, &active, &createdAt, &updatedAt)
	if queryErr != nil {
		if queryErr == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, queryErr
	}

	return buildTenantAdminUserMap(slug, id, email, mobile, role, active, createdAt, updatedAt), nil
}

// UpdateTenantAdminUserContactParams holds parameters for updating user contact.
type UpdateTenantAdminUserContactParams struct {
	TenantSlug string
	UserID     string
	Field      string
	Value      string
}

// UpdateTenantAdminUserContact updates a user's contact field.
func (s *Store) UpdateTenantAdminUserContact(ctx context.Context, params UpdateTenantAdminUserContactParams) (map[string]interface{}, error) {
	var (
		id, slug, email, mobile, role string
		active                        bool
		createdAt, updatedAt          time.Time
	)

	var query string

	args := []interface{}{params.Value, strings.TrimSpace(strings.ToLower(params.TenantSlug)), params.UserID}
	if params.Field == "email" {
		query = `UPDATE users u
			SET email = $1, updated_at = NOW()
			FROM tenants t
			WHERE t.slug = $2 AND u.tenant_id = t.tenant_id AND u.user_id::text = $3
			RETURNING u.user_id::text, t.slug, COALESCE(u.email, ''), COALESCE(u.mobile, ''), COALESCE(u.role, 'member'), COALESCE(u.active, true), u.created_at, u.updated_at`
	} else {
		query = `UPDATE users u
			SET mobile = $1, updated_at = NOW()
			FROM tenants t
			WHERE t.slug = $2 AND u.tenant_id = t.tenant_id AND u.user_id::text = $3
			RETURNING u.user_id::text, t.slug, COALESCE(u.email, ''), COALESCE(u.mobile, ''), COALESCE(u.role, 'member'), COALESCE(u.active, true), u.created_at, u.updated_at`
	}

	err := s.pool.QueryRow(ctx, query, args...).Scan(&id, &slug, &email, &mobile, &role, &active, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("update user contact: %w", err)
	}

	return buildTenantAdminUserMap(slug, id, email, mobile, role, active, createdAt, updatedAt), nil
}

// TouchTenantAdminLastLogin updates the user's last login timestamp.
func (s *Store) TouchTenantAdminLastLogin(ctx context.Context, tenantID string, userID string) (map[string]interface{}, error) {
	var (
		id, slug, email, mobile, role string
		active                        bool
		createdAt, updatedAt          time.Time
	)

	queryErr := s.pool.QueryRow(ctx, `
		UPDATE users u
		SET updated_at = NOW()
		FROM tenants t
		WHERE t.slug = $1 AND u.tenant_id = t.tenant_id AND u.user_id::text = $2
		RETURNING u.user_id::text, t.slug, COALESCE(u.email, ''), COALESCE(u.mobile, ''),
		          COALESCE(u.role, 'member'), COALESCE(u.active, true), u.created_at, u.updated_at
	`, strings.TrimSpace(strings.ToLower(tenantID)), strings.TrimSpace(userID)).Scan(&id, &slug, &email, &mobile, &role, &active, &createdAt, &updatedAt)
	if queryErr != nil {
		if queryErr == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, queryErr
	}

	result := buildTenantAdminUserMap(slug, id, email, mobile, role, active, createdAt, updatedAt)
	result["last_login_at_iso"] = time.Now().Format(time.RFC3339)

	return result, nil
}

func buildTenantAdminUserMap(
	tenantSlug string,
	id string,
	email string,
	mobile string,
	role string,
	active bool,
	createdAt time.Time,
	updatedAt time.Time,
) map[string]interface{} {
	return map[string]interface{}{
		"id":             id,
		"tenant_slug":    tenantSlug,
		"email":          email,
		"mobile":         mobile,
		"role":           role,
		"active":         active,
		"created_at_iso": createdAt.Format(time.RFC3339),
		"updated_at_iso": updatedAt.Format(time.RFC3339),
	}
}

func hasInvalidOTPRequest(err error) bool {
	return err != nil
}

// Helper functions.
func generateTenantAdminID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)

	return hex.EncodeToString(b)
}

func generateTenantAdminSessionID() (string, error) {
	b := make([]byte, 24)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func generateCaptchaAnswer(length int) string {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

	result := make([]byte, length)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		result[i] = alphabet[n.Int64()]
	}

	return string(result)
}

func generateOTPCode() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(900000))
	return fmt.Sprintf("%06d", n.Int64()+100000)
}

func maskTenantAdminIdentifier(channel, identifier string) string {
	if channel == "email" {
		parts := strings.SplitN(strings.ToLower(identifier), "@", 2)
		local := parts[0]

		domain := ""
		if len(parts) > 1 {
			domain = parts[1]
		}

		if local == "" {
			return "***@" + domain
		}

		if len(local) <= 2 {
			return string(local[0]) + "***@" + domain
		}

		return local[:2] + "***@" + domain
	}

	compact := strings.ReplaceAll(identifier, " ", "")
	if len(compact) <= 4 {
		return "***"
	}

	return compact[:3] + "****" + compact[len(compact)-3:]
}
