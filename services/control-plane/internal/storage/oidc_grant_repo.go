package storage

import (
	"context"
	"errors"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/oidc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const oidcRequestColumns = `
	id, tenant_id::text, tenant_slug, issuer, client_id, redirect_uri, scope, state, nonce,
	code_challenge, code_challenge_method, prompt, status, COALESCE(user_id::text, ''),
	COALESCE(grant_id::text, ''), auth_time, created_at, expires_at`

func scanOIDCRequest(row pgx.Row) (*oidc.AuthorizationRequest, error) {
	var (
		request  oidc.AuthorizationRequest
		authTime *time.Time
	)

	err := row.Scan(
		&request.ID, &request.TenantID, &request.TenantSlug, &request.Issuer, &request.ClientID,
		&request.RedirectURI, &request.Scope, &request.State, &request.Nonce, &request.CodeChallenge,
		&request.CodeChallengeMethod, &request.Prompt, &request.Status, &request.UserID,
		&request.GrantID, &authTime, &request.CreatedAt, &request.ExpiresAt,
	)
	if err != nil {
		return nil, err
	}

	if authTime != nil {
		request.AuthTime = authTime.UTC()
	}

	return &request, nil
}

// CreateOIDCAuthorizationRequest stores a pending authorization request.
func (r *OIDCRepository) CreateOIDCAuthorizationRequest(ctx context.Context, request oidc.AuthorizationRequest) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO oidc_authorization_requests (
			id, tenant_id, tenant_slug, issuer, client_id, redirect_uri, scope, state, nonce,
			code_challenge, code_challenge_method, prompt, status, created_at, expires_at
		) VALUES ($1, $2::uuid, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`,
		request.ID, request.TenantID, request.TenantSlug, request.Issuer, request.ClientID,
		request.RedirectURI, request.Scope, request.State, request.Nonce, request.CodeChallenge,
		request.CodeChallengeMethod, request.Prompt, request.Status, request.CreatedAt, request.ExpiresAt,
	)

	return err
}

// GetOIDCAuthorizationRequest loads an authorization request by its challenge identifier.
func (r *OIDCRepository) GetOIDCAuthorizationRequest(ctx context.Context, id string) (*oidc.AuthorizationRequest, error) {
	request, err := scanOIDCRequest(r.pool.QueryRow(ctx, `SELECT `+oidcRequestColumns+` FROM oidc_authorization_requests WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, oidc.ErrRequestNotFound
	}

	return request, err
}

// AcceptOIDCAuthorizationRequest binds a pending request to a member and stores its code.
func (r *OIDCRepository) AcceptOIDCAuthorizationRequest(ctx context.Context, accepted oidc.AuthorizationRequest, codeHash string, codeExpiresAt time.Time) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE oidc_authorization_requests
		SET status = 'accepted', user_id = $2::uuid, grant_id = $3::uuid, auth_time = $4,
		    code_hash = $5, code_expires_at = $6
		WHERE id = $1 AND status = 'pending' AND expires_at > NOW()
	`, accepted.ID, accepted.UserID, accepted.GrantID, accepted.AuthTime, codeHash, codeExpiresAt)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return oidc.ErrRequestNotPending
	}

	return nil
}

// RejectOIDCAuthorizationRequest closes a pending request without issuing a code.
func (r *OIDCRepository) RejectOIDCAuthorizationRequest(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE oidc_authorization_requests SET status = 'rejected'
		WHERE id = $1 AND status = 'pending'
	`, id)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return oidc.ErrRequestNotPending
	}

	return nil
}

// ConsumeOIDCAuthorizationCode marks a code used exactly once. A replayed code returns
// ErrCodeReused together with its request so the caller can revoke the grant.
func (r *OIDCRepository) ConsumeOIDCAuthorizationCode(ctx context.Context, codeHash string) (*oidc.AuthorizationRequest, error) {
	request, err := scanOIDCRequest(r.pool.QueryRow(ctx, `
		UPDATE oidc_authorization_requests
		SET code_used_at = NOW()
		WHERE code_hash = $1 AND status = 'accepted' AND code_used_at IS NULL AND code_expires_at > NOW()
		RETURNING `+oidcRequestColumns,
		codeHash,
	))
	if err == nil {
		return request, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	used, lookupErr := scanOIDCRequest(r.pool.QueryRow(ctx, `
		SELECT `+oidcRequestColumns+` FROM oidc_authorization_requests
		WHERE code_hash = $1 AND code_used_at IS NOT NULL
	`, codeHash))
	if lookupErr == nil {
		return used, oidc.ErrCodeReused
	}

	if errors.Is(lookupErr, pgx.ErrNoRows) {
		return nil, oidc.ErrCodeNotFound
	}

	return nil, lookupErr
}

// CreateOIDCRefreshToken stores the first refresh token of a grant.
func (r *OIDCRepository) CreateOIDCRefreshToken(ctx context.Context, token oidc.RefreshToken) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO oidc_refresh_tokens (
			id, grant_id, tenant_id, client_id, user_id, token_hash, scope, nonce, auth_time,
			created_at, expires_at, grant_expires_at
		) VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5::uuid, $6, $7, $8, $9, $10, $11, $12)
	`,
		token.ID, token.GrantID, token.TenantID, token.ClientID, token.UserID, token.TokenHash,
		token.Scope, token.Nonce, token.AuthTime, token.CreatedAt, token.ExpiresAt, token.GrantExpiresAt,
	)

	return err
}

// RotateOIDCRefreshToken exchanges a refresh token for its successor in one transaction.
// Presenting a token that was already rotated revokes the entire grant (RFC 9700 4.14.2).
func (r *OIDCRepository) RotateOIDCRefreshToken(
	ctx context.Context,
	presentedHash string,
	next oidc.RefreshToken,
	check func(oidc.RefreshToken) error,
) (*oidc.RefreshToken, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	var (
		current   oidc.RefreshToken
		usedAt    *time.Time
		revokedAt *time.Time
	)

	err = tx.QueryRow(ctx, `
		SELECT id::text, grant_id::text, tenant_id::text, client_id, user_id::text, token_hash, scope, nonce,
		       auth_time, created_at, expires_at, grant_expires_at, used_at, revoked_at
		FROM oidc_refresh_tokens
		WHERE token_hash = $1
		FOR UPDATE
	`, presentedHash).Scan(
		&current.ID, &current.GrantID, &current.TenantID, &current.ClientID, &current.UserID,
		&current.TokenHash, &current.Scope, &current.Nonce, &current.AuthTime, &current.CreatedAt,
		&current.ExpiresAt, &current.GrantExpiresAt, &usedAt, &revokedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, oidc.ErrRefreshTokenNotFound
		}

		return nil, err
	}

	if revokedAt != nil {
		return nil, oidc.ErrRefreshTokenNotFound
	}

	if usedAt != nil {
		if err := revokeOIDCGrant(ctx, tx, current.GrantID); err != nil {
			return nil, err
		}

		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}

		return nil, oidc.ErrRefreshTokenReused
	}

	if err := check(current); err != nil {
		return nil, err
	}

	next.GrantID = current.GrantID
	next.UserID = current.UserID
	next.Scope = current.Scope
	next.Nonce = current.Nonce
	next.AuthTime = current.AuthTime
	next.GrantExpiresAt = current.GrantExpiresAt

	if next.ExpiresAt.After(current.GrantExpiresAt) {
		next.ExpiresAt = current.GrantExpiresAt
	}

	if _, err := tx.Exec(ctx, `UPDATE oidc_refresh_tokens SET used_at = NOW() WHERE id = $1::uuid`, current.ID); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO oidc_refresh_tokens (
			id, grant_id, tenant_id, client_id, user_id, token_hash, scope, nonce, auth_time,
			created_at, expires_at, grant_expires_at
		) VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5::uuid, $6, $7, $8, $9, $10, $11, $12)
	`,
		next.ID, next.GrantID, current.TenantID, current.ClientID, next.UserID, next.TokenHash,
		next.Scope, next.Nonce, next.AuthTime, next.CreatedAt, next.ExpiresAt, next.GrantExpiresAt,
	); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	next.TenantID = current.TenantID
	next.ClientID = current.ClientID

	return &next, nil
}

// RevokeOIDCGrant revokes every refresh token of a grant.
func (r *OIDCRepository) RevokeOIDCGrant(ctx context.Context, grantID string) error {
	if !validUUIDs(grantID) {
		return nil
	}

	return revokeOIDCGrant(ctx, r.pool, grantID)
}

// RevokeOIDCUserClientGrants revokes a member's refresh tokens for one client.
func (r *OIDCRepository) RevokeOIDCUserClientGrants(ctx context.Context, tenantID, userID, clientID string) error {
	if !validUUIDs(tenantID, userID) {
		return nil
	}

	_, err := r.pool.Exec(ctx, `
		UPDATE oidc_refresh_tokens SET revoked_at = NOW()
		WHERE tenant_id = $1::uuid AND user_id = $2::uuid AND client_id = $3 AND revoked_at IS NULL
	`, tenantID, userID, clientID)

	return err
}

type oidcExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func revokeOIDCGrant(ctx context.Context, executor oidcExecutor, grantID string) error {
	_, err := executor.Exec(ctx, `
		UPDATE oidc_refresh_tokens SET revoked_at = NOW()
		WHERE grant_id = $1::uuid AND revoked_at IS NULL
	`, grantID)

	return err
}
