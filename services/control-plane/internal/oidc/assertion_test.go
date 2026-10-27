package oidc

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/require"
)

const (
	testAssertionKeyID    = "librechat-user-assertion"
	testAssertionIssuer   = "librechat"
	testAssertionAudience = "godarai-gateway"
)

type assertionFixture struct {
	*fixture

	signer ed25519.PrivateKey
}

func newAssertionFixture(t *testing.T) *assertionFixture {
	t.Helper()

	public, private, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	verifier, err := NewUserAssertionVerifier(UserAssertionSettings{
		PublicKeyPEM: publicKeyPEM(t, public),
		KeyID:        testAssertionKeyID,
		Issuer:       testAssertionIssuer,
		Audience:     testAssertionAudience,
	})
	require.NoError(t, err)

	f := newFixture(t)
	f.service.assertions = verifier

	return &assertionFixture{fixture: f, signer: private}
}

func publicKeyPEM(t *testing.T, public ed25519.PublicKey) string {
	t.Helper()

	der, err := x509.MarshalPKIXPublicKey(public)
	require.NoError(t, err)

	return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}))
}

func (f *assertionFixture) claims() map[string]any {
	iat := f.now().Unix()

	return map[string]any{
		"iss": testAssertionIssuer, "aud": testAssertionAudience, "sub": testUserID, "tid": testTenantID,
		"uid": "6ad130bf2e66ad6859368e16", "jti": "jti-1", "iat": iat, "nbf": iat, "exp": iat + 600,
	}
}

func (f *assertionFixture) sign(claims map[string]any, signer ed25519.PrivateKey) string {
	f.t.Helper()

	header, err := json.Marshal(map[string]string{"alg": "EdDSA", "typ": "JWT", "kid": testAssertionKeyID})
	require.NoError(f.t, err)

	payload, err := json.Marshal(claims)
	require.NoError(f.t, err)

	input := encodeSegment(header) + "." + encodeSegment(payload)

	return input + "." + encodeSegment(ed25519.Sign(signer, []byte(input)))
}

func TestUserAssertionAuthenticatesTheMemberAsLibreChatUsage(t *testing.T) {
	t.Parallel()

	f := newAssertionFixture(t)

	resolved, err := f.service.AuthenticateAccessToken(context.Background(), testTenantID, f.sign(f.claims(), f.signer))
	require.NoError(t, err)
	require.Equal(t, testUserID, resolved.UserID)
	require.Equal(t, testTenantID, resolved.TenantID)
	require.Equal(t, platform.CreditProductLibreChat, resolved.CreditProduct, "chat must reach seat admission")
	require.Equal(t, UserAssertionKeyType, resolved.KeyType)
	require.Equal(t, UserAssertionKeyType+":"+testUserID, resolved.KeyID, "cache and audit scopes stay per member")
}

func TestUserAssertionRejectsInvalidAssertions(t *testing.T) {
	t.Parallel()

	_, otherSigner, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	cases := map[string]struct {
		tenantID string
		mutate   func(claims map[string]any)
		other    bool
	}{
		"another tenant's host":   {tenantID: otherTenantID},
		"foreign signing key":     {other: true},
		"wrong audience":          {mutate: func(c map[string]any) { c["aud"] = "someone-else" }},
		"wrong issuer":            {mutate: func(c map[string]any) { c["iss"] = "mallory" }},
		"missing tenant":          {mutate: func(c map[string]any) { delete(c, "tid") }},
		"missing subject":         {mutate: func(c map[string]any) { c["sub"] = "" }},
		"expired":                 {mutate: func(c map[string]any) { c["exp"] = c["iat"].(int64) - 120 }},
		"lifetime over one hour":  {mutate: func(c map[string]any) { c["exp"] = c["iat"].(int64) + 7200 }},
		"not yet valid":           {mutate: func(c map[string]any) { c["nbf"] = c["iat"].(int64) + 300 }},
		"unknown member":          {mutate: func(c map[string]any) { c["sub"] = "99999999-9999-4999-8999-999999999999" }},
		"tenant claim of another": {mutate: func(c map[string]any) { c["tid"] = otherTenantID }},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f := newAssertionFixture(t)

			claims := f.claims()
			if tc.mutate != nil {
				tc.mutate(claims)
			}

			signer := f.signer
			if tc.other {
				signer = otherSigner
			}

			tenantID := testTenantID
			if tc.tenantID != "" {
				tenantID = tc.tenantID
			}

			_, err := f.service.AuthenticateAccessToken(context.Background(), tenantID, f.sign(claims, signer))
			require.Error(t, err)
			require.True(t, IsProtocolError(err), "a bad assertion is a 401, not a server error: %v", err)
		})
	}
}

func TestUserAssertionRejectsDeactivatedMember(t *testing.T) {
	t.Parallel()

	f := newAssertionFixture(t)
	f.repo.setIdentity(Identity{UserID: testUserID, TenantID: testTenantID, Role: "member", Active: false})

	_, err := f.service.AuthenticateAccessToken(context.Background(), testTenantID, f.sign(f.claims(), f.signer))
	require.True(t, IsProtocolError(err))
}

func TestAccessTokensStillWorkWithUserAssertionsEnabled(t *testing.T) {
	t.Parallel()

	f := newAssertionFixture(t)

	tokens, err := f.exchange(f.login())
	require.NoError(t, err)

	resolved, err := f.service.AuthenticateAccessToken(context.Background(), testTenantID, tokens.AccessToken)
	require.NoError(t, err)
	require.Equal(t, KeyType, resolved.KeyType)
}

func TestNewUserAssertionVerifierRequiresEverySetting(t *testing.T) {
	t.Parallel()

	public, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	key := publicKeyPEM(t, public)

	for name, settings := range map[string]UserAssertionSettings{
		"no audience": {PublicKeyPEM: key, KeyID: testAssertionKeyID, Issuer: testAssertionIssuer},
		"no key id":   {PublicKeyPEM: key, Issuer: testAssertionIssuer, Audience: testAssertionAudience},
		"not PEM":     {PublicKeyPEM: "nope", KeyID: testAssertionKeyID, Issuer: testAssertionIssuer, Audience: testAssertionAudience},
	} {
		_, settingsErr := NewUserAssertionVerifier(settings)
		require.Error(t, settingsErr, name)
	}

	_, err = NewUserAssertionVerifier(UserAssertionSettings{
		PublicKeyPEM: key, KeyID: testAssertionKeyID, Issuer: testAssertionIssuer, Audience: testAssertionAudience,
	})
	require.NoError(t, err)
}
