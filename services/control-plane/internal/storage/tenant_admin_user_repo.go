package storage

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5"
)

const invitationTTL = 7 * 24 * time.Hour

var (
	ErrInvalidInvitationToken = errors.New("invalid tenant admin invitation token")
	ErrInvitationExpired      = errors.New("tenant admin invitation expired")
	ErrInvitationAlreadyUsed  = errors.New("tenant admin invitation already used")
	ErrInvitationContactTaken = errors.New("tenant admin invitation contact already used")
)

func (s *Store) ListTenantAdminUsers(ctx context.Context, tenantSlug string, query string, limit int, offset int) ([]platform.TenantAdminUser, int, error) {
	tenantSlug = strings.TrimSpace(strings.ToLower(tenantSlug))
	query = strings.TrimSpace(strings.ToLower(query))
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	like := "%" + query + "%"
	var total int
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(1)
		FROM users u
		JOIN tenants t ON t.tenant_id = u.tenant_id
		WHERE t.slug = $1
		  AND ($2 = '' OR LOWER(COALESCE(u.email, '')) LIKE $3 OR LOWER(COALESCE(u.mobile, '')) LIKE $3 OR LOWER(COALESCE(u.user_alias, '')) LIKE $3)
	`, tenantSlug, query, like).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := s.pool.Query(ctx, `
		SELECT u.user_id::text, t.slug, COALESCE(u.email, ''), COALESCE(u.mobile, ''), COALESCE(u.role, 'tenant_member'),
		       COALESCE(u.active, true), u.created_at, u.updated_at
		FROM users u
		JOIN tenants t ON t.tenant_id = u.tenant_id
		WHERE t.slug = $1
		  AND ($2 = '' OR LOWER(COALESCE(u.email, '')) LIKE $3 OR LOWER(COALESCE(u.mobile, '')) LIKE $3 OR LOWER(COALESCE(u.user_alias, '')) LIKE $3)
		ORDER BY u.updated_at DESC, u.created_at DESC
		LIMIT $4 OFFSET $5
	`, tenantSlug, query, like, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]platform.TenantAdminUser, 0, limit)
	for rows.Next() {
		var item platform.TenantAdminUser
		var email, mobile, role string
		var active bool
		if err := rows.Scan(&item.ID, &item.TenantSlug, &email, &mobile, &role, &active, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, err
		}
		if email != "" {
			item.Email = email
		}
		if mobile != "" {
			item.Mobile = mobile
		}
		item.Role = role
		item.Active = active
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *Store) InviteTenantAdminUser(ctx context.Context, tenantID string, tenantSlug string, req platform.InviteTenantAdminUserRequest) (*platform.TenantAdminInvitation, string, error) {
	tenantID = strings.TrimSpace(tenantID)
	tenantSlug = strings.TrimSpace(strings.ToLower(tenantSlug))
	email := strings.TrimSpace(strings.ToLower(req.Email))
	mobile := strings.TrimSpace(req.Mobile)
	role := strings.TrimSpace(strings.ToLower(req.Role))
	if role == "" {
		role = "tenant_member"
	}

	now := time.Now().UTC()
	expiresAt := now.Add(invitationTTL)
	invitationID := uuid.NewString()
	token, tokenHash, err := generateInviteToken()
	if err != nil {
		return nil, "", err
	}

	var acceptedAt *time.Time
	if err := s.pool.QueryRow(ctx, `
		INSERT INTO "tenant_admin_invitations" (
			"invitation_id", "tenant_id", "tenantSlug", "email", "mobile", "role", "status",
			"token_hash", "expires_at", "created_at", "updated_at"
		) VALUES (
			$1, $2::uuid, $3, NULLIF($4, ''), NULLIF($5, ''), $6, 'pending',
			$7, $8, $9, $9
		)
		RETURNING "accepted_at"
	`, invitationID, tenantID, tenantSlug, email, mobile, role, tokenHash, expiresAt, now).Scan(&acceptedAt); err != nil {
		return nil, "", err
	}

	invitation := &platform.TenantAdminInvitation{
		InvitationID: invitationID,
		TenantID:     tenantID,
		TenantSlug:   tenantSlug,
		Email:        email,
		Mobile:       mobile,
		Role:         role,
		Status:       "pending",
		ExpiresAt:    expiresAt,
		AcceptedAt:   acceptedAt,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	link := buildInviteAcceptLink(req.AcceptInviteClientURL, tenantSlug, token)
	return invitation, link, nil
}

func (s *Store) AcceptTenantAdminInvitation(ctx context.Context, tenantID string, tenantSlug string, token string) (*platform.TenantAdminInvitation, *platform.TenantAdminUser, error) {
	tenantID = strings.TrimSpace(tenantID)
	tenantSlug = strings.TrimSpace(strings.ToLower(tenantSlug))
	tokenHash, err := tokenHashFromInviteToken(token)
	if err != nil {
		return nil, nil, ErrInvalidInvitationToken
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var invitation platform.TenantAdminInvitation
	var status string
	var acceptedAt *time.Time
	var updatedAt time.Time
	if err := tx.QueryRow(ctx, `
		SELECT "invitation_id"::text, "tenant_id"::text, "tenantSlug",
		       COALESCE("email", ''), COALESCE("mobile", ''), COALESCE("role", 'tenant_member'),
		       "status", "expires_at", "accepted_at", "created_at", "updated_at"
		FROM "tenant_admin_invitations"
		WHERE "token_hash" = $1
		FOR UPDATE
	`, tokenHash).Scan(
		&invitation.InvitationID,
		&invitation.TenantID,
		&invitation.TenantSlug,
		&invitation.Email,
		&invitation.Mobile,
		&invitation.Role,
		&status,
		&invitation.ExpiresAt,
		&acceptedAt,
		&invitation.CreatedAt,
		&updatedAt,
	); err != nil {
		return nil, nil, err
	}

	if invitation.TenantID != tenantID || invitation.TenantSlug != tenantSlug {
		return nil, nil, pgx.ErrNoRows
	}
	if status != "pending" {
		return nil, nil, ErrInvitationAlreadyUsed
	}

	now := time.Now().UTC()
	if invitation.ExpiresAt.Before(now) {
		_, _ = tx.Exec(ctx, `
			UPDATE "tenant_admin_invitations"
			SET "status" = 'expired', "updated_at" = $2
			WHERE "invitation_id" = $1::uuid
		`, invitation.InvitationID, now)
		return nil, nil, ErrInvitationExpired
	}

	user, err := upsertTenantAdminUserFromInvitation(ctx, tx, invitation.TenantID, invitation.TenantSlug, invitation.Email, invitation.Mobile, invitation.Role)
	if err != nil {
		return nil, nil, err
	}

	if err := tx.QueryRow(ctx, `
		UPDATE "tenant_admin_invitations"
		SET "status" = 'accepted', "accepted_at" = $2, "updated_at" = $2
		WHERE "invitation_id" = $1::uuid
		RETURNING "accepted_at", "updated_at"
	`, invitation.InvitationID, now).Scan(&acceptedAt, &updatedAt); err != nil {
		return nil, nil, err
	}

	invitation.Status = "accepted"
	invitation.AcceptedAt = acceptedAt
	invitation.UpdatedAt = updatedAt

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, err
	}
	return &invitation, user, nil
}

func (s *Store) SetTenantAdminUserActive(ctx context.Context, tenantSlug string, userID string, active bool) (*platform.TenantAdminUser, error) {
	tenantSlug = strings.TrimSpace(strings.ToLower(tenantSlug))
	userID = strings.TrimSpace(userID)
	if tenantSlug == "" || userID == "" {
		return nil, pgx.ErrNoRows
	}

	rows, err := s.pool.Query(ctx, `
		UPDATE users u
		SET active = $3, updated_at = NOW()
		FROM tenants t
		WHERE t.slug = $1
		  AND u.user_id::text = $2
		  AND u.tenant_id = t.tenant_id
		RETURNING u.user_id::text, t.slug, COALESCE(u.email, ''), COALESCE(u.mobile, ''), COALESCE(u.role, 'tenant_member'),
		          COALESCE(u.active, true), u.created_at, u.updated_at
	`, tenantSlug, userID, active)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, pgx.ErrNoRows
	}

	var user platform.TenantAdminUser
	var email, mobile, role string
	if err := rows.Scan(&user.ID, &user.TenantSlug, &email, &mobile, &role, &user.Active, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return nil, err
	}
	user.Role = role
	if email != "" {
		user.Email = email
	}
	if mobile != "" {
		user.Mobile = mobile
	}
	return &user, nil
}

func upsertTenantAdminUserFromInvitation(ctx context.Context, tx pgx.Tx, tenantID string, tenantSlug string, email string, mobile string, role string) (*platform.TenantAdminUser, error) {
	var user platform.TenantAdminUser
	var currentRole string
	var currentActive bool
	err := tx.QueryRow(ctx, `
		SELECT u.user_id::text, t.slug, COALESCE(u.email, ''), COALESCE(u.mobile, ''), COALESCE(u.role, 'tenant_member'),
		       COALESCE(u.active, true), u.created_at, u.updated_at
		FROM users u
		JOIN tenants t ON t.tenant_id = u.tenant_id
		WHERE t.tenant_id = $1::uuid
		  AND (
		    ($2 <> '' AND LOWER(COALESCE(u.email, '')) = LOWER($2))
		    OR ($3 <> '' AND COALESCE(u.mobile, '') = $3)
		  )
		LIMIT 1
		FOR UPDATE
	`, tenantID, email, mobile).Scan(&user.ID, &user.TenantSlug, &user.Email, &user.Mobile, &currentRole, &currentActive, &user.CreatedAt, &user.UpdatedAt)
	switch {
	case err == nil:
		nextRole := currentRole
		if strings.TrimSpace(nextRole) == "" || strings.EqualFold(nextRole, "internal_user") || strings.EqualFold(nextRole, "deactivated") {
			nextRole = role
		}
		if err := tx.QueryRow(ctx, `
			UPDATE users
			SET role = $2, active = true, updated_at = NOW()
			WHERE user_id::text = $1
			RETURNING updated_at
		`, user.ID, nextRole).Scan(&user.UpdatedAt); err != nil {
			return nil, err
		}
		user.Role = nextRole
		user.Active = true
		return &user, nil
	case errors.Is(err, pgx.ErrNoRows):
		newUserID := uuid.NewString()
		if err := tx.QueryRow(ctx, `
			INSERT INTO users (user_id, tenant_id, user_alias, email, mobile, role, active, created_at, updated_at)
			VALUES ($1::uuid, $2::uuid, NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''), $6, true, NOW(), NOW())
			RETURNING created_at, updated_at
		`, newUserID, tenantID, strings.TrimSpace(email), strings.TrimSpace(email), strings.TrimSpace(mobile), role).Scan(&user.CreatedAt, &user.UpdatedAt); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return nil, ErrInvitationContactTaken
			}
			return nil, err
		}
		user.ID = newUserID
		user.TenantSlug = tenantSlug
		user.Email = strings.TrimSpace(email)
		user.Mobile = strings.TrimSpace(mobile)
		user.Role = role
		user.Active = true
		return &user, nil
	default:
		return nil, err
	}
}

func buildInviteAcceptLink(rawBaseURL string, tenantSlug string, token string) string {
	baseURL := strings.TrimSpace(rawBaseURL)
	if baseURL == "" {
		baseURL = fmt.Sprintf("https://%s.app.sharah.ir/invite/accept", tenantSlug)
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		parsed = &url.URL{
			Scheme: "https",
			Host:   fmt.Sprintf("%s.app.sharah.ir", tenantSlug),
			Path:   "/invite/accept",
		}
	}
	q := parsed.Query()
	q.Set("token", token)
	parsed.RawQuery = q.Encode()
	return parsed.String()
}

func generateInviteToken() (token string, hash string, err error) {
	raw := uuid.NewString() + ":" + uuid.NewString()
	sum := sha256.Sum256([]byte(raw))
	hash = fmt.Sprintf("%x", sum[:])
	token = base64.RawURLEncoding.EncodeToString([]byte(raw))
	return token, hash, nil
}

func tokenHashFromInviteToken(token string) (string, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(token))
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(decoded)
	return fmt.Sprintf("%x", sum[:]), nil
}
