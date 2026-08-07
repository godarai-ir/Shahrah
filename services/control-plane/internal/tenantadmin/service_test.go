package tenantadmin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type tenantAdminRepositoryStub struct {
	repository
	availability *TenantSlugAvailability
}

func (s tenantAdminRepositoryStub) CheckTenantSlugAvailability(context.Context, string) (*TenantSlugAvailability, error) {
	return s.availability, nil
}

func TestNewServiceRejectsNilRepository(t *testing.T) {
	service, err := NewService(nil)

	require.ErrorIs(t, err, errRepositoryRequired)
	require.Nil(t, service)
}

func TestServiceDelegatesTenantAdminUseCase(t *testing.T) {
	want := &TenantSlugAvailability{Slug: "acme", Available: true}
	service, err := NewService(tenantAdminRepositoryStub{availability: want})
	require.NoError(t, err)

	got, err := service.CheckTenantSlugAvailability(context.Background(), "acme")
	require.NoError(t, err)
	require.Equal(t, want, got)
}
