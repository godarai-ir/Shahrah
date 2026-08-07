package iam

import (
	"context"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/require"
)

type directoryRepositoryStub struct {
	directoryRepository
	orgUnit *platform.OrgUnit
}

func (s directoryRepositoryStub) GetOrgUnit(context.Context, string, string) (*platform.OrgUnit, error) {
	return s.orgUnit, nil
}

type accessRepositoryStub struct{ accessRepository }

func TestNewServiceRejectsNilDirectoryRepository(t *testing.T) {
	service, err := NewService(nil, nil)

	require.ErrorIs(t, err, errDirectoryRepositoryRequired)
	require.Nil(t, service)
}

func TestServiceDelegatesIAMUseCase(t *testing.T) {
	want := &platform.OrgUnit{ID: "org-1"}
	service, err := NewService(directoryRepositoryStub{orgUnit: want}, accessRepositoryStub{})
	require.NoError(t, err)

	got, err := service.GetOrgUnit(context.Background(), "tenant-1", "org-1")
	require.NoError(t, err)
	require.Equal(t, want, got)
}
