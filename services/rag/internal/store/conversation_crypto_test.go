package store

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConversationEncryptionRoundTripAndTamperDetection(t *testing.T) {
	t.Parallel()

	secret := []byte("01234567890123456789012345678901")
	aad := conversationAAD("tenant", "session", "assistant")
	ciphertext, err := encryptConversation(secret, aad, "پاسخ محرمانه")
	require.NoError(t, err)
	require.NotContains(t, string(ciphertext), "محرمانه")

	plaintext, err := decryptConversation(secret, aad, ciphertext)
	require.NoError(t, err)
	require.Equal(t, "پاسخ محرمانه", plaintext)

	ciphertext[len(ciphertext)-1] ^= 1
	_, err = decryptConversation(secret, aad, ciphertext)
	require.Error(t, err)

	untampered, err := encryptConversation(secret, aad, "پاسخ محرمانه")
	require.NoError(t, err)
	_, err = decryptConversation(secret, conversationAAD("other-tenant", "session", "assistant"), untampered)
	require.Error(t, err)
}
