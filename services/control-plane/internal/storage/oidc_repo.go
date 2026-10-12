package storage

import (
	"context"
	"errors"
	"strings"

	"github.com/Delpak/shahrah/services/control-plane/internal/oidc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OIDCRepository persists OIDC clients, authorization requests, and refresh tokens.
type OIDCRepository struct {
	pool *pgxpool.Pool
}

func newOIDCRepository(pool *pgxpool.Pool) *OIDCRepository {
	return &OIDCRepository{pool: pool}
}

const oidcClientColumns = `
	id::text, tenant_id::text, client_id, name, secret_hash, redirect_uris, post_logout_redirect_uris,
	active, credit_product, created_by, created_at, updated_at, secret_rotated_at`

func scanOIDCClient(row pgx.Row) (*oidc.Client, error) {
	var client oidc.Client

	err := row.Scan(
		&client.ID, &client.TenantID, &client.ClientID, &client.Name, &client.SecretHash,
		&client.RedirectURIs, &client.PostLogoutRedirectURIs, &client.Active, &client.CreditProduct, &client.CreatedBy,
		&client.CreatedAt, &client.UpdatedAt, &client.SecretRotatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, oidc.ErrClientNotFound
		}

		return nil, err
	}

	return &client, nil
}

// GetOIDCClient loads a client by its public identifier.
func (r *OIDCRepository) GetOIDCClient(ctx context.Context, clientID string) (*oidc.Client, error) {
	return scanOIDCClient(r.pool.QueryRow(ctx, `SELECT `+oidcClientColumns+` FROM oidc_clients WHERE client_id = $1`, clientID))
}

// ListOIDCClients lists a tenant's clients, newest first.
func (r *OIDCRepository) ListOIDCClients(ctx context.Context, tenantID string) ([]oidc.Client, error) {
	if _, err := uuid.Parse(tenantID); err != nil {
		return []oidc.Client{}, nil
	}

	rows, err := r.pool.Query(ctx, `
		SELECT `+oidcClientColumns+`
		FROM oidc_clients
		WHERE tenant_id = $1::uuid
		ORDER BY created_at DESC
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	clients := []oidc.Client{}

	for rows.Next() {
		client, scanErr := scanOIDCClient(rows)
		if scanErr != nil {
			return nil, scanErr
		}

		clients = append(clients, *client)
	}

	return clients, rows.Err()
}

// CreateOIDCClient inserts a client registration.
func (r *OIDCRepository) CreateOIDCClient(ctx context.Context, client oidc.Client) (*oidc.Client, error) {
	created, err := scanOIDCClient(r.pool.QueryRow(ctx, `
		INSERT INTO oidc_clients (
			tenant_id, client_id, name, secret_hash, redirect_uris, post_logout_redirect_uris, active, created_by, credit_product
		)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8, COALESCE(NULLIF($9, ''), 'oidc_app'))
		RETURNING `+oidcClientColumns,
		client.TenantID, client.ClientID, client.Name, client.SecretHash,
		client.RedirectURIs, client.PostLogoutRedirectURIs, client.Active, client.CreatedBy, client.CreditProduct,
	))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, oidc.ErrClientIDTaken
		}

		return nil, err
	}

	return created, nil
}

// UpdateOIDCClient applies the non-nil fields of update to a tenant's client.
func (r *OIDCRepository) UpdateOIDCClient(ctx context.Context, tenantID, id string, update oidc.UpdateClientRequest) (*oidc.Client, error) {
	if !validUUIDs(tenantID, id) {
		return nil, oidc.ErrClientNotFound
	}

	return scanOIDCClient(r.pool.QueryRow(ctx, `
		UPDATE oidc_clients
		SET name = COALESCE($3, name),
		    redirect_uris = COALESCE($4, redirect_uris),
		    post_logout_redirect_uris = COALESCE($5, post_logout_redirect_uris),
		    active = COALESCE($6, active),
		    credit_product = COALESCE($7, credit_product),
		    updated_at = NOW()
		WHERE tenant_id = $1::uuid AND id = $2::uuid
		RETURNING `+oidcClientColumns,
		tenantID, id, update.Name, update.RedirectURIs, update.PostLogoutRedirectURIs, update.Active, update.CreditProduct,
	))
}

// RotateOIDCClientSecret replaces a client's secret hash.
func (r *OIDCRepository) RotateOIDCClientSecret(ctx context.Context, tenantID, id, secretHash string) (*oidc.Client, error) {
	if !validUUIDs(tenantID, id) {
		return nil, oidc.ErrClientNotFound
	}

	return scanOIDCClient(r.pool.QueryRow(ctx, `
		UPDATE oidc_clients
		SET secret_hash = $3, secret_rotated_at = NOW(), updated_at = NOW()
		WHERE tenant_id = $1::uuid AND id = $2::uuid
		RETURNING `+oidcClientColumns,
		tenantID, id, secretHash,
	))
}

// DeleteOIDCClient removes a client; requests and refresh tokens cascade.
func (r *OIDCRepository) DeleteOIDCClient(ctx context.Context, tenantID, id string) error {
	if !validUUIDs(tenantID, id) {
		return oidc.ErrClientNotFound
	}

	tag, err := r.pool.Exec(ctx, `DELETE FROM oidc_clients WHERE tenant_id = $1::uuid AND id = $2::uuid`, tenantID, id)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return oidc.ErrClientNotFound
	}

	return nil
}

// GetOIDCIdentity loads the member a token is issued for.
func (r *OIDCRepository) GetOIDCIdentity(ctx context.Context, tenantID, userID string) (*oidc.Identity, error) {
	if !validUUIDs(tenantID, userID) {
		return nil, oidc.ErrIdentityNotFound
	}

	var identity oidc.Identity

	err := r.pool.QueryRow(ctx, `
		SELECT user_id::text, tenant_id::text, COALESCE(email, ''), COALESCE(mobile, ''),
		       COALESCE(user_alias, ''), COALESCE(role, 'member'), COALESCE(active, true)
		FROM users
		WHERE tenant_id = $1::uuid AND user_id = $2::uuid
		LIMIT 1
	`, tenantID, userID).Scan(
		&identity.UserID, &identity.TenantID, &identity.Email, &identity.Mobile,
		&identity.Name, &identity.Role, &identity.Active,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, oidc.ErrIdentityNotFound
		}

		return nil, err
	}

	if strings.EqualFold(identity.Role, "deactivated") {
		identity.Active = false
	}

	return &identity, nil
}

func validUUIDs(values ...string) bool {
	for _, value := range values {
		if _, err := uuid.Parse(strings.TrimSpace(value)); err != nil {
			return false
		}
	}

	return true
}
