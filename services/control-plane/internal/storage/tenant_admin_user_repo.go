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
	"github.com/Delpak/shahrah/services/control-plane/internal/apperrors"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/seats"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenantadmin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const invitationTTL = 7 * 24 * time.Hour

var (
	ErrInvalidInvitationToken    = tenantadmin.ErrInvalidInvitationToken
	ErrInvitationExpired         = tenantadmin.ErrInvitationExpired
	ErrInvitationAlreadyUsed     = tenantadmin.ErrInvitationAlreadyUsed
	ErrInvitationContactTaken    = tenantadmin.ErrInvitationContactTaken
	ErrInvitationAlreadyPending  = tenantadmin.ErrInvitationAlreadyPending
	ErrTenantMemberAlreadyExists = tenantadmin.ErrTenantMemberAlreadyExists
	ErrTenantMemberDeactivated   = tenantadmin.ErrTenantMemberDeactivated
	ErrInvalidInvitationEmail    = tenantadmin.ErrInvalidInvitationEmail
)

func (s *TenantAdminRepository) ListTenantAdminUsers(ctx context.Context, tenantSlug string, query string, limit int, offset int) ([]platform.TenantAdminUser, int, error) {
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
		  AND ($2 = '' OR LOWER(COALESCE(u.email, '')) LIKE $3 OR LOWER(COALESCE(u.mobile, '')) LIKE $3 OR LOWER(COALESCE(u.user_alias, '')) LIKE $3 OR LOWER(COALESCE(u.first_name, '') || ' ' || COALESCE(u.last_name, '')) LIKE $3)
	`, tenantSlug, query, like).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := s.pool.Query(ctx, `
		SELECT u.user_id::text, t.slug, COALESCE(u.email, ''), COALESCE(u.mobile, ''), COALESCE(u.role, 'member'),
		       COALESCE(u.first_name, ''), COALESCE(u.last_name, ''),
		       COALESCE(u.active, true), u.created_at, u.updated_at, u.last_login_at,
		       COALESCE(
		         ARRAY_AGG(r.name ORDER BY r.name) FILTER (WHERE urb.is_active = true AND r.name IS NOT NULL),
		         '{}'
		       ) AS roles
		FROM users u
		JOIN tenants t ON t.tenant_id = u.tenant_id
		LEFT JOIN user_role_bindings urb ON urb.tenant_id = u.tenant_id AND urb.user_id = u.user_id
		LEFT JOIN roles r ON r.tenant_id = u.tenant_id AND r.role_id = urb.role_id
		WHERE t.slug = $1
		  AND ($2 = '' OR LOWER(COALESCE(u.email, '')) LIKE $3 OR LOWER(COALESCE(u.mobile, '')) LIKE $3 OR LOWER(COALESCE(u.user_alias, '')) LIKE $3 OR LOWER(COALESCE(u.first_name, '') || ' ' || COALESCE(u.last_name, '')) LIKE $3)
		GROUP BY u.user_id, t.slug, u.email, u.mobile, u.role, u.first_name, u.last_name, u.active, u.created_at, u.updated_at, u.last_login_at
		ORDER BY u.updated_at DESC, u.created_at DESC
		LIMIT $4 OFFSET $5
	`, tenantSlug, query, like, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]platform.TenantAdminUser, 0, limit)

	for rows.Next() {
		var (
			item                platform.TenantAdminUser
			email, mobile, role string
			active              bool
			roles               []string
		)
		if err := rows.Scan(&item.ID, &item.TenantSlug, &email, &mobile, &role, &item.FirstName, &item.LastName, &active, &item.CreatedAt, &item.UpdatedAt, &item.LastLoginAt, &roles); err != nil {
			return nil, 0, err
		}

		if email != "" {
			item.Email = email
		}

		if mobile != "" {
			item.Mobile = mobile
		}

		item.Role = role
		if len(roles) > 0 {
			item.Roles = roles
		} else if role != "" {
			item.Roles = []string{role}
		}

		item.Active = active
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (s *TenantAdminRepository) GetTenantAdminUserRecord(ctx context.Context, tenantSlug string, userID string) (*platform.TenantAdminUser, error) {
	tenantSlug = strings.TrimSpace(strings.ToLower(tenantSlug))

	userID = strings.TrimSpace(userID)
	if tenantSlug == "" || userID == "" {
		return nil, apperrors.ErrNotFound
	}

	var (
		item                platform.TenantAdminUser
		email, mobile, role string
		roles               []string
	)
	if err := s.pool.QueryRow(ctx, `
		SELECT u.user_id::text, t.slug, COALESCE(u.email, ''), COALESCE(u.mobile, ''), COALESCE(u.role, 'member'),
		       COALESCE(u.first_name, ''), COALESCE(u.last_name, ''),
		       COALESCE(u.active, true), u.created_at, u.updated_at, u.last_login_at,
		       COALESCE(
		         ARRAY_AGG(r.name ORDER BY r.name) FILTER (WHERE urb.is_active = true AND r.name IS NOT NULL),
		         '{}'
		       ) AS roles
		FROM users u
		JOIN tenants t ON t.tenant_id = u.tenant_id
		LEFT JOIN user_role_bindings urb ON urb.tenant_id = u.tenant_id AND urb.user_id = u.user_id
		LEFT JOIN roles r ON r.tenant_id = u.tenant_id AND r.role_id = urb.role_id
		WHERE t.slug = $1
		  AND u.user_id::text = $2
		GROUP BY u.user_id, t.slug, u.email, u.mobile, u.role, u.first_name, u.last_name, u.active, u.created_at, u.updated_at, u.last_login_at
		LIMIT 1
	`, tenantSlug, userID).Scan(
		&item.ID, &item.TenantSlug, &email, &mobile, &role, &item.FirstName, &item.LastName, &item.Active, &item.CreatedAt, &item.UpdatedAt, &item.LastLoginAt, &roles,
	); err != nil {
		return nil, err
	}

	if email != "" {
		item.Email = email
	}

	if mobile != "" {
		item.Mobile = mobile
	}

	item.Role = role
	if len(roles) > 0 {
		item.Roles = roles
	} else if role != "" {
		item.Roles = []string{role}
	}

	return &item, nil
}

func (s *TenantAdminRepository) InviteTenantAdminUser(ctx context.Context, tenantID string, tenantSlug string, req platform.InviteTenantAdminUserRequest) (*platform.TenantAdminInvitation, string, error) {
	tenantID = strings.TrimSpace(tenantID)
	tenantSlug = strings.TrimSpace(strings.ToLower(tenantSlug))
	email := strings.TrimSpace(strings.ToLower(req.Email))
	mobile := strings.TrimSpace(req.Mobile)
	firstName := strings.TrimSpace(req.FirstName)
	lastName := strings.TrimSpace(req.LastName)

	role := strings.TrimSpace(strings.ToLower(req.Role))
	if role == "" {
		role = "member"
	}

	if email != "" && !IsValidEmailAddress(email) {
		return nil, "", ErrInvalidInvitationEmail
	}

	seatType := ""

	if strings.TrimSpace(req.SeatType) != "" {
		parsed, seatErr := seats.ParseType(req.SeatType)
		if seatErr != nil {
			return nil, "", seatErr
		}

		seatType = string(parsed)
	}

	now := time.Now().UTC()

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, "", err
	}
	defer rollbackTxBestEffort(ctx, tx)

	var (
		existingActive bool
		existingRole   string
	)

	err = tx.QueryRow(ctx, `
		SELECT COALESCE(u.active, true), COALESCE(u.role, 'member')
		FROM users u
		WHERE u.tenant_id = $1::uuid
		  AND (
			($2 <> '' AND LOWER(COALESCE(u.email, '')) = LOWER($2))
			OR ($3 <> '' AND COALESCE(u.mobile, '') = $3)
		  )
		LIMIT 1
		FOR UPDATE
	`, tenantID, email, mobile).Scan(&existingActive, &existingRole)
	switch {
	case err == nil:
		if !existingActive || strings.EqualFold(existingRole, "deactivated") {
			return nil, "", ErrTenantMemberDeactivated
		}

		return nil, "", ErrTenantMemberAlreadyExists
	case !errors.Is(err, pgx.ErrNoRows):
		return nil, "", err
	}

	_, cleanupErr := tx.Exec(ctx, `
		DELETE FROM tenant_admin_invitations
		WHERE tenant_id = $1::uuid
		  AND status = 'pending'
		  AND expires_at <= $2
		  AND (
			($3 <> '' AND LOWER(COALESCE(email, '')) = LOWER($3))
			OR ($4 <> '' AND COALESCE(mobile, '') = $4)
		  )
	`, tenantID, now, email, mobile)
	if cleanupErr != nil {
		return nil, "", cleanupErr
	}

	var pendingID string

	err = tx.QueryRow(ctx, `
		SELECT invitation_id::text
		FROM tenant_admin_invitations
		WHERE tenant_id = $1::uuid
		  AND status = 'pending'
		  AND expires_at > $2
		  AND (
			($3 <> '' AND LOWER(COALESCE(email, '')) = LOWER($3))
			OR ($4 <> '' AND COALESCE(mobile, '') = $4)
		  )
		LIMIT 1
		FOR UPDATE
	`, tenantID, now, email, mobile).Scan(&pendingID)
	switch {
	case err == nil:
		return nil, "", ErrInvitationAlreadyPending
	case !errors.Is(err, pgx.ErrNoRows):
		return nil, "", err
	}

	// A pending invitation holds a seat until it is accepted or expires.
	if capErr := enforceTenantUserCapTx(ctx, tx, tenantID, 1); capErr != nil {
		return nil, "", capErr
	}

	expiresAt := now.Add(invitationTTL)
	invitationID := uuid.NewString()

	token, tokenHash, err := generateInviteToken()
	if err != nil {
		return nil, "", err
	}

	var acceptedAt *time.Time
	if err := tx.QueryRow(ctx, `
		INSERT INTO "tenant_admin_invitations" (
			"invitation_id", "tenant_id", "tenantSlug", "email", "mobile", "role", "status",
			"token_hash", "expires_at", "created_at", "updated_at", "first_name", "last_name", "seat_type"
		) VALUES (
			$1, $2::uuid, $3, NULLIF($4, ''), NULLIF($5, ''), $6, 'pending',
			$7, $8, $9, $9, NULLIF($10, ''), NULLIF($11, ''), NULLIF($12, '')
		)
		RETURNING "accepted_at"
	`, invitationID, tenantID, tenantSlug, email, mobile, role, tokenHash, expiresAt, now, firstName, lastName, seatType).Scan(&acceptedAt); err != nil {
		return nil, "", err
	}

	if err := tx.Commit(ctx); err != nil {
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

func (s *TenantAdminRepository) AcceptTenantAdminInvitation(ctx context.Context, tenantID string, tenantSlug string, token string) (*platform.TenantAdminInvitation, *platform.TenantAdminUser, error) {
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
	defer rollbackTxBestEffort(ctx, tx)

	var (
		invitation platform.TenantAdminInvitation
		status     string
		acceptedAt *time.Time
		updatedAt  time.Time
		firstName  string
		lastName   string
		seatType   string
	)

	queryErr := tx.QueryRow(ctx, `
		SELECT "invitation_id"::text, "tenant_id"::text, "tenantSlug",
		       COALESCE("email", ''), COALESCE("mobile", ''), COALESCE("role", 'member'),
		       "status", "expires_at", "accepted_at", "created_at", "updated_at",
		       COALESCE("first_name", ''), COALESCE("last_name", ''), COALESCE("seat_type", '')
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
		&firstName,
		&lastName,
		&seatType,
	)
	if queryErr != nil {
		return nil, nil, queryErr
	}

	if invitation.TenantID != tenantID || invitation.TenantSlug != tenantSlug {
		return nil, nil, apperrors.ErrNotFound
	}

	if status != "pending" {
		return nil, nil, ErrInvitationAlreadyUsed
	}

	now := time.Now().UTC()
	if invitation.ExpiresAt.Before(now) {
		if _, updateErr := tx.Exec(ctx, `
			UPDATE "tenant_admin_invitations"
			SET "status" = 'expired', "updated_at" = $2
			WHERE "invitation_id" = $1::uuid
		`, invitation.InvitationID, now); updateErr != nil {
			return nil, nil, updateErr
		}

		return nil, nil, ErrInvitationExpired
	}

	// The pending invitation already holds a seat, so accepting it adds none.
	if capErr := enforceTenantUserCapTx(ctx, tx, invitation.TenantID, 0); capErr != nil {
		return nil, nil, capErr
	}

	user, err := upsertTenantAdminUserFromInvitation(ctx, tx, invitation.TenantID, invitation.TenantSlug, invitation.Email, invitation.Mobile, invitation.Role, firstName, lastName)
	if err != nil {
		return nil, nil, err
	}

	if err := assignInvitedSeatTx(ctx, tx, invitation.TenantID, user.ID, seatType, now); err != nil {
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

	s.invalidateAuthzCacheForUser(invitation.TenantID, user.ID)
	s.bumpAuthzTenantEpoch(invitation.TenantID)

	return &invitation, user, nil
}

// UpdateTenantAdminUserProfile sets the first and last name of one tenant member.
func (s *TenantAdminRepository) UpdateTenantAdminUserProfile(ctx context.Context, tenantSlug string, userID string, req platform.UpdateTenantAdminUserProfileRequest) (*platform.TenantAdminUser, error) {
	tenantSlug = strings.TrimSpace(strings.ToLower(tenantSlug))

	userID = strings.TrimSpace(userID)
	if tenantSlug == "" || userID == "" {
		return nil, apperrors.ErrNotFound
	}

	tag, err := s.pool.Exec(ctx, `
		UPDATE users u
		SET first_name = NULLIF($3, ''), last_name = NULLIF($4, ''), updated_at = NOW()
		FROM tenants t
		WHERE t.tenant_id = u.tenant_id
		  AND t.slug = $1
		  AND u.user_id::text = $2
	`, tenantSlug, userID, strings.TrimSpace(req.FirstName), strings.TrimSpace(req.LastName))
	if err != nil {
		return nil, err
	}

	if tag.RowsAffected() == 0 {
		return nil, apperrors.ErrNotFound
	}

	return s.GetTenantAdminUserRecord(ctx, tenantSlug, userID)
}

func (s *TenantAdminRepository) SetTenantAdminUserActive(ctx context.Context, tenantSlug string, userID string, active bool) (*platform.TenantAdminUser, error) {
	tenantSlug = strings.TrimSpace(strings.ToLower(tenantSlug))

	userID = strings.TrimSpace(userID)
	if tenantSlug == "" || userID == "" {
		return nil, apperrors.ErrNotFound
	}

	if active {
		if err := s.reactivateWithinUserCap(ctx, tenantSlug, userID); err != nil {
			return nil, err
		}
	}

	rows, err := s.pool.Query(ctx, `
		UPDATE users u
		SET active = $3, updated_at = NOW()
		FROM tenants t
		WHERE t.slug = $1
		  AND u.user_id::text = $2
		  AND u.tenant_id = t.tenant_id
		RETURNING u.user_id::text, t.tenant_id::text, t.slug, COALESCE(u.email, ''), COALESCE(u.mobile, ''), COALESCE(u.role, 'member'),
		          COALESCE(u.active, true), u.created_at, u.updated_at
	`, tenantSlug, userID, active)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, apperrors.ErrNotFound
	}

	var (
		user                platform.TenantAdminUser
		email, mobile, role string
		tenantID            string
	)
	if err := rows.Scan(&user.ID, &tenantID, &user.TenantSlug, &email, &mobile, &role, &user.Active, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return nil, err
	}

	user.Role = role
	if email != "" {
		user.Email = email
	}

	if mobile != "" {
		user.Mobile = mobile
	}

	if !active {
		if _, revokeErr := s.pool.Exec(ctx, `
			DELETE FROM tenant_admin_sessions
			WHERE "tenantSlug" = $1 AND "userID" = $2
		`, tenantSlug, userID); revokeErr != nil {
			return nil, revokeErr
		}

		// A departing member's paid seat goes back to the tenant's stock.
		if _, releaseErr := s.pool.Exec(ctx, `
			DELETE FROM tenant_seat_assignments WHERE tenant_id = $1::uuid AND user_id::text = $2
		`, tenantID, userID); releaseErr != nil {
			return nil, releaseErr
		}
	}

	if strings.TrimSpace(tenantID) != "" {
		s.invalidateAuthzCacheForUser(tenantID, user.ID)
		s.bumpAuthzTenantEpoch(tenantID)
	}

	return &user, nil
}

func upsertTenantAdminUserFromInvitation(ctx context.Context, tx pgx.Tx, tenantID string, tenantSlug string, email string, mobile string, role string, firstName string, lastName string) (*platform.TenantAdminUser, error) {
	var (
		user          platform.TenantAdminUser
		currentRole   string
		currentActive bool
	)

	err := tx.QueryRow(ctx, `
		SELECT u.user_id::text, t.slug, COALESCE(u.email, ''), COALESCE(u.mobile, ''), COALESCE(u.role, 'member'),
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

		updateErr := tx.QueryRow(ctx, `
			UPDATE users
			SET role = $2, active = true, updated_at = NOW(),
			    first_name = COALESCE(NULLIF($3, ''), first_name),
			    last_name = COALESCE(NULLIF($4, ''), last_name)
			WHERE user_id::text = $1
			RETURNING updated_at, COALESCE(first_name, ''), COALESCE(last_name, '')
		`, user.ID, nextRole, firstName, lastName).Scan(&user.UpdatedAt, &user.FirstName, &user.LastName)
		if updateErr != nil {
			return nil, updateErr
		}

		user.Role = nextRole
		user.Active = true

		return &user, nil
	case errors.Is(err, pgx.ErrNoRows):
		newUserID := uuid.NewString()

		insertErr := tx.QueryRow(ctx, `
			INSERT INTO users (user_id, tenant_id, user_alias, email, mobile, role, first_name, last_name, active, created_at, updated_at)
			VALUES ($1::uuid, $2::uuid, NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''), $6, NULLIF($7, ''), NULLIF($8, ''), true, NOW(), NOW())
			RETURNING created_at, updated_at
		`, newUserID, tenantID, strings.TrimSpace(email), strings.TrimSpace(email), strings.TrimSpace(mobile), role, firstName, lastName).Scan(&user.CreatedAt, &user.UpdatedAt)
		if insertErr != nil {
			var pgErr *pgconn.PgError
			if errors.As(insertErr, &pgErr) && pgErr.Code == "23505" {
				return nil, ErrInvitationContactTaken
			}

			return nil, insertErr
		}

		user.ID = newUserID
		user.TenantSlug = tenantSlug
		user.Email = strings.TrimSpace(email)
		user.Mobile = strings.TrimSpace(mobile)
		user.Role = role
		user.FirstName = firstName
		user.LastName = lastName
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

// reactivateWithinUserCap activates a deactivated member only when the tenant
// has a free seat. Activating an already active member is a no-op.
func (s *TenantAdminRepository) reactivateWithinUserCap(ctx context.Context, tenantSlug, userID string) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer rollbackTxBestEffort(ctx, tx)

	var (
		tenantID      string
		alreadyActive bool
	)

	err = tx.QueryRow(ctx, `
		SELECT u.tenant_id::text, COALESCE(u.active, true)
		FROM users u
		JOIN tenants t ON t.tenant_id = u.tenant_id
		WHERE t.slug = $1 AND u.user_id::text = $2
		FOR UPDATE OF u
	`, tenantSlug, userID).Scan(&tenantID, &alreadyActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return apperrors.ErrNotFound
	}

	if err != nil || alreadyActive {
		return err
	}

	if capErr := enforceTenantUserCapTx(ctx, tx, tenantID, 1); capErr != nil {
		return capErr
	}

	if _, err := tx.Exec(ctx, `UPDATE users SET active = true, updated_at = NOW() WHERE user_id::text = $1`, userID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// assignInvitedSeatTx gives a new member the seat their invitation named, or
// the free seat when it named none. A seat that is no longer available leaves
// the member without one; the seat worker gives them the free seat later if
// one frees up. The savepoint keeps a refused seat from failing the acceptance.
func assignInvitedSeatTx(ctx context.Context, tx pgx.Tx, tenantID, userID, seatType string, now time.Time) error {
	target := seats.Free
	if seatType != "" {
		target = seats.Type(seatType)
	}

	if target == seats.Unassigned {
		return nil
	}

	if _, err := tx.Exec(ctx, `SAVEPOINT invited_seat`); err != nil {
		return err
	}

	if _, err := setMemberSeatTx(ctx, tx, tenantID, userID, target, "invitation", now); err != nil {
		_, rollbackErr := tx.Exec(ctx, `ROLLBACK TO SAVEPOINT invited_seat`)

		return rollbackErr
	}

	_, err := tx.Exec(ctx, `RELEASE SAVEPOINT invited_seat`)

	return err
}
