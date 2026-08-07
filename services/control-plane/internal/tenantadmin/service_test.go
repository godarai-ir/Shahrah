package tenantadmin

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewServiceRejectsNilRepository(t *testing.T) {
	service, err := NewService(nil)

	require.ErrorIs(t, err, errRepositoryRequired)
	require.Nil(t, service)
}
