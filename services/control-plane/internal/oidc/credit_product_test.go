package oidc

import (
	"context"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/require"
)

func TestClientCreditProductDefaultsToGenericOIDCApp(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	require.Equal(t, platform.CreditProductOIDCApp, f.client.CreditProduct)

	tokens, err := f.exchange(f.login())
	require.NoError(t, err)

	resolved, err := f.service.AuthenticateAccessToken(context.Background(), testTenantID, tokens.AccessToken)
	require.NoError(t, err)
	require.Equal(t, platform.CreditProductOIDCApp, resolved.CreditProduct)
}

func TestClientCreditProductIsStampedIntoAccessTokens(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	product := " LibreChat "

	updated, err := f.service.UpdateClient(context.Background(), testTenantID, f.client.ID, UpdateClientRequest{CreditProduct: &product})
	require.NoError(t, err)
	require.Equal(t, platform.CreditProductLibreChat, updated.CreditProduct)

	tokens, err := f.exchange(f.login())
	require.NoError(t, err)

	claims, err := f.service.VerifyAccessToken(tokens.AccessToken)
	require.NoError(t, err)
	require.Equal(t, platform.CreditProductLibreChat, claims.Product)

	resolved, err := f.service.AuthenticateAccessToken(context.Background(), testTenantID, tokens.AccessToken)
	require.NoError(t, err)
	require.Equal(t, platform.CreditProductLibreChat, resolved.CreditProduct)
	require.Empty(t, resolved.TeamID, "OIDC usage is attributed to the user and tenant only")
}

func TestClientCreditProductRejectsUnknownLabels(t *testing.T) {
	t.Parallel()

	f := newFixture(t)

	_, err := f.service.CreateClient(context.Background(), testTenantID, "admin", CreateClientRequest{
		Name: "Spoof", RedirectURIs: []string{testRedirect}, CreditProduct: platform.CreditProductGuardrail,
	})
	require.ErrorIs(t, err, ErrInvalidClientConfig)

	product := "gateway"
	_, err = f.service.UpdateClient(context.Background(), testTenantID, f.client.ID, UpdateClientRequest{CreditProduct: &product})
	require.ErrorIs(t, err, ErrInvalidClientConfig)
}

func TestAccessTokenProductFallsBackForUnknownClaims(t *testing.T) {
	t.Parallel()

	require.Equal(t, platform.CreditProductOIDCApp, accessTokenProduct(""))
	require.Equal(t, platform.CreditProductOIDCApp, accessTokenProduct("guardrail"))
	require.Equal(t, platform.CreditProductLibreChat, accessTokenProduct(platform.CreditProductLibreChat))
}
