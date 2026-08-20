package store

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRateLimitKeyHashDoesNotExposeIdentifier(t *testing.T) {
	t.Parallel()

	first := hashRateLimitValue([]byte("secret-one"), "192.0.2.1")
	second := hashRateLimitValue([]byte("secret-one"), "192.0.2.1")
	third := hashRateLimitValue([]byte("secret-two"), "192.0.2.1")

	require.Equal(t, first, second)
	require.NotEqual(t, first, third)
	require.NotContains(t, first, "192.0.2.1")
	require.Len(t, first, 24)
}
