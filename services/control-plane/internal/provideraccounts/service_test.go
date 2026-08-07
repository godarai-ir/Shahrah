package provideraccounts

import (
	"context"
	"errors"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/require"
)

var errPersistenceConstraint = errors.New("persistence constraint")

type providerAccountRepositoryStub struct {
	repository
	accounts []platform.TenantProviderAccount
}

func (s providerAccountRepositoryStub) ListTenantProviderAccounts(context.Context, string) ([]platform.TenantProviderAccount, error) {
	return s.accounts, errPersistenceConstraint
}

func (s providerAccountRepositoryStub) TranslateProviderAccountError(err error) error {
	if errors.Is(err, errPersistenceConstraint) {
		return ErrInvalidPayload
	}

	return err
}

func TestNewServiceRequiresRepository(t *testing.T) {
	t.Parallel()

	service, err := NewService(nil)
	require.ErrorIs(t, err, errRepositoryRequired)
	require.Nil(t, service)
}

func TestServiceTranslatesRepositoryErrors(t *testing.T) {
	t.Parallel()

	service, err := NewService(providerAccountRepositoryStub{})
	require.NoError(t, err)

	_, err = service.ListTenantProviderAccounts(context.Background(), "tenant-1")
	require.ErrorIs(t, err, ErrInvalidPayload)
}
