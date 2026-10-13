package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAttributionUUIDDropsMalformedIDs(t *testing.T) {
	t.Parallel()

	require.Nil(t, attributionUUID(""))
	require.Nil(t, attributionUUID("virtual-user-7"), "a malformed id must not fail the spend-log batch")
	require.Equal(t, "8f14e45f-ceea-467f-a9f0-5b6b7b8c9d0e", attributionUUID(" 8F14E45F-CEEA-467F-A9F0-5B6B7B8C9D0E "))
}
