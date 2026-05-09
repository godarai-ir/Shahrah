package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"
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
	var storedAnswer string
	var expiresAt time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT answer, "expiresAt" FROM tenant_admin_captchas WHERE id = $1
	`, id).Scan(&storedAnswer, &expiresAt)
	if err != nil {
		return false, nil
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
	var record TenantAdminOTPRequest
	var otpCode string

	err := s.pool.QueryRow(ctx, `
		SELECT id, "tenantSlug", purpose, channel, identifier, "userID",
		       "pendingValue", "otpCode", attempts, "expiresAt", "createdAt"
		FROM tenant_admin_otp_requests WHERE id = $1
	`, params.RequestID).Scan(
		&record.ID, &record.TenantSlug, &record.Purpose, &record.Channel,
		&record.Identifier, &record.UserID, &record.PendingValue,
		&otpCode, &record.Attempts, &record.ExpiresAt, &record.CreatedAt,
	)
	if err != nil {
		return map[string]interface{}{"ok": false, "error": "درخواست OTP معتبر نیست یا منقضی شده است."}, nil
	}

	if record.TenantSlug != params.TenantSlug {
		return map[string]interface{}{"ok": false, "error": "این درخواست OTP مربوط به Tenant فعلی نیست."}, nil
	}
	if record.Purpose != params.Purpose {
		return map[string]interface{}{"ok": false, "error": "کاربرد OTP با درخواست فعلی هم‌خوانی ندارد."}, nil
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

	err := s.pool.QueryRow(ctx, `
		SELECT id, "tenantSlug", "userID", "expiresAt", "createdAt"
		FROM tenant_admin_sessions WHERE id = $1
	`, id).Scan(&session.ID, &session.TenantSlug, &session.UserID, &session.ExpiresAt, &session.CreatedAt)
	if err != nil {
		return nil, nil
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

// FindOrCreateTenantAdminUser finds or creates a user.
func (s *Store) FindOrCreateTenantAdminUser(ctx context.Context, params FindOrCreateTenantAdminUserParams) (map[string]interface{}, error) {
	var id, email, alias, role string
	var createdAt, updatedAt time.Time

	if params.Channel == "mobile" {
		err := s.pool.QueryRow(ctx, `
			SELECT user_id::text, COALESCE(email, ''), COALESCE(user_alias, ''), COALESCE(role, 'tenant_member'),
			       created_at, updated_at
			FROM users WHERE COALESCE(user_alias, '') = $1 LIMIT 1
		`, params.Identifier).Scan(&id, &email, &alias, &role, &createdAt, &updatedAt)
		if err == nil {
			return buildTenantAdminUserMap(params.TenantSlug, id, email, alias, role, createdAt, updatedAt), nil
		}
	} else {
		err := s.pool.QueryRow(ctx, `
			SELECT user_id::text, COALESCE(email, ''), COALESCE(user_alias, ''), COALESCE(role, 'tenant_member'),
			       created_at, updated_at
			FROM users WHERE LOWER(COALESCE(email, '')) = LOWER($1) LIMIT 1
		`, params.Identifier).Scan(&id, &email, &alias, &role, &createdAt, &updatedAt)
		if err == nil {
			return buildTenantAdminUserMap(params.TenantSlug, id, email, alias, role, createdAt, updatedAt), nil
		}
	}

	// Create new user
	id = generateTenantAdminID()
	now := time.Now()
	role = "tenant_member"
	emailValue := ""
	aliasValue := ""
	if params.Channel == "mobile" {
		aliasValue = params.Identifier
	} else {
		emailValue = params.Identifier
		aliasValue = params.Identifier
	}

	err := s.pool.QueryRow(ctx, `
		INSERT INTO users (user_id, user_alias, email, role, created_at, updated_at)
		VALUES ($1, NULLIF($2, ''), NULLIF($3, ''), $4, $5, $6)
		RETURNING user_id::text, COALESCE(email, ''), COALESCE(user_alias, ''), COALESCE(role, 'tenant_member'), created_at, updated_at
	`, id, aliasValue, emailValue, role, now, now).Scan(&id, &email, &alias, &role, &createdAt, &updatedAt)
	if err != nil {
		// Retry in case of race condition
		if params.Channel == "mobile" {
			err = s.pool.QueryRow(ctx, `
				SELECT user_id::text, COALESCE(email, ''), COALESCE(user_alias, ''), COALESCE(role, 'tenant_member'),
				       created_at, updated_at
				FROM users WHERE COALESCE(user_alias, '') = $1 LIMIT 1
			`, params.Identifier).Scan(&id, &email, &alias, &role, &createdAt, &updatedAt)
		} else {
			err = s.pool.QueryRow(ctx, `
				SELECT user_id::text, COALESCE(email, ''), COALESCE(user_alias, ''), COALESCE(role, 'tenant_member'),
				       created_at, updated_at
				FROM users WHERE LOWER(COALESCE(email, '')) = LOWER($1) LIMIT 1
			`, params.Identifier).Scan(&id, &email, &alias, &role, &createdAt, &updatedAt)
		}
		if err != nil {
			return nil, fmt.Errorf("create user: %w", err)
		}
	}

	return buildTenantAdminUserMap(params.TenantSlug, id, email, alias, role, createdAt, updatedAt), nil
}

// GetTenantAdminUser retrieves a user by ID.
func (s *Store) GetTenantAdminUser(ctx context.Context, tenantID string, userID string) (map[string]interface{}, error) {
	var id, email, alias, role string
	var createdAt, updatedAt time.Time

	err := s.pool.QueryRow(ctx, `
		SELECT user_id::text, COALESCE(email, ''), COALESCE(user_alias, ''),
		       COALESCE(role, 'tenant_member'), created_at, updated_at
		FROM users WHERE user_id::text = $1 LIMIT 1
	`, userID).Scan(&id, &email, &alias, &role, &createdAt, &updatedAt)
	if err != nil {
		return nil, nil
	}

	return buildTenantAdminUserMap(tenantID, id, email, alias, role, createdAt, updatedAt), nil
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
	var id, email, alias, role string
	var createdAt, updatedAt time.Time

	query := ""
	args := []interface{}{params.Value, params.UserID}
	if params.Field == "email" {
		query = `UPDATE users SET email = $1, updated_at = NOW() WHERE user_id::text = $2
			RETURNING user_id::text, COALESCE(email, ''), COALESCE(user_alias, ''), COALESCE(role, 'tenant_member'), created_at, updated_at`
	} else {
		// Current users schema has no `mobile` column. Store mobile as user_alias.
		query = `UPDATE users SET user_alias = $1, updated_at = NOW() WHERE user_id::text = $2
			RETURNING user_id::text, COALESCE(email, ''), COALESCE(user_alias, ''), COALESCE(role, 'tenant_member'), created_at, updated_at`
	}

	err := s.pool.QueryRow(ctx, query, args...).Scan(&id, &email, &alias, &role, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("update user contact: %w", err)
	}

	return buildTenantAdminUserMap(params.TenantSlug, id, email, alias, role, createdAt, updatedAt), nil
}

// TouchTenantAdminLastLogin updates the user's last login timestamp.
func (s *Store) TouchTenantAdminLastLogin(ctx context.Context, tenantID string, userID string) (map[string]interface{}, error) {
	var id, email, alias, role string
	var createdAt, updatedAt time.Time

	err := s.pool.QueryRow(ctx, `
		UPDATE users SET updated_at = NOW() WHERE user_id::text = $1
		RETURNING user_id::text, COALESCE(email, ''), COALESCE(user_alias, ''),
		          COALESCE(role, 'tenant_member'), created_at, updated_at
	`, userID).Scan(&id, &email, &alias, &role, &createdAt, &updatedAt)
	if err != nil {
		return nil, nil
	}

	result := buildTenantAdminUserMap(tenantID, id, email, alias, role, createdAt, updatedAt)
	result["last_login_at_iso"] = time.Now().Format(time.RFC3339)
	return result, nil
}

func buildTenantAdminUserMap(
	tenantSlug string,
	id string,
	email string,
	alias string,
	role string,
	createdAt time.Time,
	updatedAt time.Time,
) map[string]interface{} {
	result := map[string]interface{}{
		"id":             id,
		"tenant_slug":    tenantSlug,
		"email":          email,
		"mobile":         "",
		"role":           role,
		"created_at_iso": createdAt.Format(time.RFC3339),
		"updated_at_iso": updatedAt.Format(time.RFC3339),
	}

	if mobile := mobileFromAlias(alias); mobile != "" {
		result["mobile"] = mobile
	}
	return result
}

var mobileAliasPattern = regexp.MustCompile(`^\+?\d{10,15}$`)

func mobileFromAlias(alias string) string {
	value := strings.TrimSpace(alias)
	if value == "" {
		return ""
	}
	if !mobileAliasPattern.MatchString(value) {
		return ""
	}
	return value
}

// Helper functions
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
