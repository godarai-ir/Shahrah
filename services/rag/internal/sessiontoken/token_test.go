package sessiontoken

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSignAndVerify(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Truncate(time.Second)
	secret := []byte("01234567890123456789012345678901")
	token, _, err := Sign(secret, Claims{
		ChatbotPublicID: "bot_1", ChatbotVersionID: "version_1",
		TenantID: "tenant", ExpiresAt: now.Add(time.Hour).Unix(),
	}, now)
	require.NoError(t, err)

	claims, err := Verify(secret, token, "bot_1", "version_1", "tenant", now)
	require.NoError(t, err)
	require.Equal(t, "tenant", claims.TenantID)

	_, err = Verify(secret, token, "bot_1", "version_2", "tenant", now)
	require.ErrorContains(t, err, "mismatched")
}
