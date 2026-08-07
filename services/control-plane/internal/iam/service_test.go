package iam

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewServiceRejectsNilDirectoryRepository(t *testing.T) {
	service, err := NewService(nil, nil)

	require.ErrorIs(t, err, errDirectoryRepositoryRequired)
	require.Nil(t, service)
}
