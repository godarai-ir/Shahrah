package integration_test

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/oidc"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenantadmin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type integrationSessions struct {
	store *storage.Store
}

func (s integrationSessions) LookupSession(ctx context.Context, sessionID string) (*oidc.Session, error) {
	session, err := s.store.GetTenantAdminSession(ctx, sessionID)
	if err != nil || session == nil {
		return nil, err
	}

	createdAt, err := time.Parse(time.RFC3339, session.CreatedAtISO)
	if err != nil {
		return nil, err
	}

	return &oidc.Session{ID: session.ID, TenantSlug: session.TenantSlug, UserID: session.UserID, CreatedAt: createdAt}, nil
}

// TestIntegration_OIDCProviderPersistsTheFullGrantLifecycle runs the provider against the
// PostgreSQL adapter: registration, login challenge, code exchange, rotation, replay
// revocation, and gateway principal resolution.
func TestIntegration_OIDCProviderPersistsTheFullGrantLifecycle(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(databaseURL) == "" {
		t.Skip("DATABASE_URL not set")
	}

	ctx := context.Background()
	logger := zap.NewNop()

	store, err := storage.NewStore(ctx, databaseURL, logger, storage.StoreDependencies{})
	require.NoError(t, err)
	t.Cleanup(store.Close)

	pool, err := pgxpool.New(ctx, databaseURL)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	slug := "oidc-" + strings.ReplaceAll(uuid.NewString(), "-", "")[:12]
	tenantID := uuid.NewString()
	userID := uuid.NewString()

	_, err = pool.Exec(ctx, `INSERT INTO tenants (tenant_id, slug, domain, status) VALUES ($1::uuid, $2, $2, 'active')`, tenantID, slug)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, cleanupErr := pool.Exec(context.Background(), `DELETE FROM tenants WHERE tenant_id = $1::uuid`, tenantID)
		require.NoError(t, cleanupErr)
	})

	_, err = pool.Exec(ctx, `
		INSERT INTO users (user_id, tenant_id, user_alias, email, role, active, created_at, updated_at)
		VALUES ($1::uuid, $2::uuid, 'Sara', 'sara@oidc.example', 'member', true, NOW(), NOW())
	`, userID, tenantID)
	require.NoError(t, err)

	session, err := store.CreateTenantAdminSession(ctx, tenantadmin.CreateTenantAdminSessionParams{TenantSlug: slug, UserID: userID})
	require.NoError(t, err)

	keys, err := oidc.GenerateKeySet()
	require.NoError(t, err)

	service, err := oidc.NewService(oidc.Dependencies{
		Repository: store.OIDCRepository,
		Sessions:   integrationSessions{store: store},
		Principals: store.AuthRepository,
		Keys:       keys,
	})
	require.NoError(t, err)

	client, err := service.CreateClient(ctx, tenantID, "admin", oidc.CreateClientRequest{
		Name: "LibreChat", RedirectURIs: []string{"https://chat.oidc.example/oauth/openid/callback"},
	})
	require.NoError(t, err)

	listed, err := service.ListClients(ctx, tenantID)
	require.NoError(t, err)
	require.Len(t, listed, 1)
	require.Equal(t, client.ClientID, listed[0].ClientID)

	tenant := oidc.Tenant{ID: tenantID, Slug: slug, Host: slug + ".godar.example"}
	verifier := strings.Repeat("v", 64)
	digest := sha256.Sum256([]byte(verifier))

	authorized, err := service.Authorize(ctx, tenant, oidc.AuthorizeParams{
		ClientID:            client.ClientID,
		RedirectURI:         "https://chat.oidc.example/oauth/openid/callback",
		ResponseType:        oidc.ResponseTypeCode,
		Scope:               "openid profile email",
		State:               "s",
		Nonce:               "n",
		CodeChallenge:       base64.RawURLEncoding.EncodeToString(digest[:]),
		CodeChallengeMethod: oidc.CodeChallengeMethodS256,
	})
	require.NoError(t, err)

	loginURL, err := url.Parse(authorized.RedirectURL)
	require.NoError(t, err)

	decision, err := service.AcceptLogin(ctx, slug, loginURL.Query().Get("oidc_request"), session.ID)
	require.NoError(t, err)

	callback, err := url.Parse(decision.RedirectTo)
	require.NoError(t, err)

	code := callback.Query().Get("code")
	exchange := oidc.TokenParams{
		GrantType: oidc.GrantTypeAuthorizationCode, ClientID: client.ClientID, ClientSecret: client.ClientSecret,
		Code: code, RedirectURI: "https://chat.oidc.example/oauth/openid/callback", CodeVerifier: verifier,
	}

	tokens, err := service.Token(ctx, tenant, exchange)
	require.NoError(t, err)

	_, err = service.Token(ctx, tenant, exchange)
	require.True(t, oidc.IsProtocolError(err), "a code is single use")

	principal, err := service.AuthenticateAccessToken(ctx, tenantID, tokens.AccessToken)
	require.NoError(t, err)
	require.Equal(t, userID, principal.UserID)
	require.Equal(t, oidc.KeyType, principal.KeyType)

	refresh := oidc.TokenParams{
		GrantType: oidc.GrantTypeRefreshToken, ClientID: client.ClientID, ClientSecret: client.ClientSecret,
		RefreshToken: tokens.RefreshToken,
	}

	_, err = service.Token(ctx, tenant, refresh)
	require.True(t, oidc.IsProtocolError(err), "replaying the code revoked its refresh token")

	_, err = pool.Exec(ctx, `UPDATE users SET active = false WHERE user_id = $1::uuid`, userID)
	require.NoError(t, err)

	_, err = service.AuthenticateAccessToken(ctx, tenantID, tokens.AccessToken)
	require.True(t, oidc.IsProtocolError(err), "deactivated members lose gateway access")
}

func TestIntegration_OIDCRefreshRotationRevokesTheGrantOnReplay(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(databaseURL) == "" {
		t.Skip("DATABASE_URL not set")
	}

	ctx := context.Background()

	store, err := storage.NewStore(ctx, databaseURL, zap.NewNop(), storage.StoreDependencies{})
	require.NoError(t, err)
	t.Cleanup(store.Close)

	pool, err := pgxpool.New(ctx, databaseURL)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	tenantID := uuid.NewString()
	slug := "oidc-" + strings.ReplaceAll(uuid.NewString(), "-", "")[:12]

	_, err = pool.Exec(ctx, `INSERT INTO tenants (tenant_id, slug, domain, status) VALUES ($1::uuid, $2, $2, 'active')`, tenantID, slug)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, cleanupErr := pool.Exec(context.Background(), `DELETE FROM tenants WHERE tenant_id = $1::uuid`, tenantID)
		require.NoError(t, cleanupErr)
	})

	_, err = pool.Exec(ctx, `
		INSERT INTO oidc_clients (tenant_id, client_id, name, secret_hash, redirect_uris)
		VALUES ($1::uuid, $2, 'c', 'h', ARRAY['https://c.example/cb'])
	`, tenantID, "gdc-"+slug)
	require.NoError(t, err)

	now := time.Now().UTC()
	grantID := uuid.NewString()
	first := oidc.RefreshToken{
		ID: uuid.NewString(), GrantID: grantID, TenantID: tenantID, ClientID: "gdc-" + slug,
		UserID: uuid.NewString(), TokenHash: "first-" + slug, Scope: "openid", AuthTime: now,
		CreatedAt: now, ExpiresAt: now.Add(time.Hour), GrantExpiresAt: now.Add(30 * time.Minute),
	}
	require.NoError(t, store.CreateOIDCRefreshToken(ctx, first))

	accept := func(oidc.RefreshToken) error { return nil }
	second, err := store.RotateOIDCRefreshToken(ctx, first.TokenHash, oidc.RefreshToken{
		ID: uuid.NewString(), TokenHash: "second-" + slug, CreatedAt: now, ExpiresAt: now.Add(time.Hour),
	}, accept)
	require.NoError(t, err)
	require.Equal(t, grantID, second.GrantID)
	require.WithinDuration(t, first.GrantExpiresAt, second.ExpiresAt, time.Second, "a rotation never outlives its grant")

	_, err = store.RotateOIDCRefreshToken(ctx, first.TokenHash, oidc.RefreshToken{
		ID: uuid.NewString(), TokenHash: "third-" + slug, CreatedAt: now, ExpiresAt: now.Add(time.Hour),
	}, accept)
	require.ErrorIs(t, err, oidc.ErrRefreshTokenReused)

	_, err = store.RotateOIDCRefreshToken(ctx, second.TokenHash, oidc.RefreshToken{
		ID: uuid.NewString(), TokenHash: "fourth-" + slug, CreatedAt: now, ExpiresAt: now.Add(time.Hour),
	}, accept)
	require.ErrorIs(t, err, oidc.ErrRefreshTokenNotFound, "replay revoked the successor")
}
