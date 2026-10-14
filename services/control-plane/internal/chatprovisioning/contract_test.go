package chatprovisioning

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestHTTPClientAgainstTheChatProvisioningAPI runs the client against a real LibreChat provisioning
// API. Set CHAT_PROVISIONING_CONTRACT_URL and CHAT_PROVISIONING_CONTRACT_TOKEN to run it; the API
// must be a disposable instance, because the test creates and suspends a tenant.
func TestHTTPClientAgainstTheChatProvisioningAPI(t *testing.T) {
	baseURL := os.Getenv("CHAT_PROVISIONING_CONTRACT_URL")
	token := os.Getenv("CHAT_PROVISIONING_CONTRACT_TOKEN")

	if baseURL == "" || token == "" {
		t.Skip("set CHAT_PROVISIONING_CONTRACT_URL and CHAT_PROVISIONING_CONTRACT_TOKEN to run the contract test")
	}

	ctx := context.Background()
	tenantID := "cccccccc-0000-4000-8000-0000000000aa"

	client, err := NewHTTPClient(baseURL, token, 10*time.Second)
	require.NoError(t, err)

	// Without a secret an unknown tenant cannot be created: the chat answers INVALID_OIDC.
	_, err = client.Register(ctx, tenantID, RegisterRequest{Slug: "contract", ClientID: "gdc_contract"})
	apiErr, ok := AsAPIError(err)
	require.True(t, ok, "expected an API error, got %v", err)
	require.Equal(t, errInvalidOIDC, apiErr.Code)
	require.True(t, apiErr.Permanent())

	created, err := client.Register(ctx, tenantID, RegisterRequest{
		Slug: "contract", ClientID: "gdc_contract", ClientSecret: "gds_contract_secret",
	})
	require.NoError(t, err)
	t.Logf("created: ready=%v reason=%q", created.Ready, created.Reason)

	// The same call again is idempotent, and a call without the secret keeps the stored one.
	again, err := client.Register(ctx, tenantID, RegisterRequest{Slug: "contract", ClientID: "gdc_contract"})
	require.NoError(t, err)
	require.Equal(t, created.Ready, again.Ready)

	require.NoError(t, client.SetStatus(ctx, tenantID, chatStatusSuspended))
	require.NoError(t, client.SetStatus(ctx, tenantID, chatStatusActive))

	// A reserved slug is refused for good.
	_, err = client.Register(ctx, "cccccccc-0000-4000-8000-0000000000bb", RegisterRequest{
		Slug: "admin", ClientID: "c", ClientSecret: "s",
	})
	apiErr, ok = AsAPIError(err)
	require.True(t, ok)
	require.True(t, apiErr.Permanent())

	// A wrong token is refused.
	wrong, err := NewHTTPClient(baseURL, token+"x", 10*time.Second)
	require.NoError(t, err)

	_, err = wrong.Register(ctx, tenantID, RegisterRequest{Slug: "contract", ClientID: "c"})
	apiErr, ok = AsAPIError(err)
	require.True(t, ok)
	require.Equal(t, 401, apiErr.Status)
}
