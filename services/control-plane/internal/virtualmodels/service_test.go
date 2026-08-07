package virtualmodels

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewServiceRequiresRepository(t *testing.T) {
	t.Parallel()

	service, err := NewService(nil)

	require.ErrorIs(t, err, errRepositoryRequired)
	require.Nil(t, service)
}
