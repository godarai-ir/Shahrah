package iam

import (
	"context"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/require"
)

type directoryRepositoryStub struct {
	directoryRepository
	team *platform.Team
}

func (s directoryRepositoryStub) GetTeam(context.Context, string, string) (*platform.Team, error) {
	return s.team, nil
}

type accessRepositoryStub struct{ accessRepository }

func TestNewServiceRejectsNilDirectoryRepository(t *testing.T) {
	service, err := NewService(nil, nil)

	require.ErrorIs(t, err, errDirectoryRepositoryRequired)
	require.Nil(t, service)
}

func TestServiceDelegatesIAMUseCase(t *testing.T) {
	want := &platform.Team{ID: "team-1"}
	service, err := NewService(directoryRepositoryStub{team: want}, accessRepositoryStub{})
	require.NoError(t, err)

	got, err := service.GetTeam(context.Background(), "tenant-1", "team-1")
	require.NoError(t, err)
	require.Equal(t, want, got)
}
