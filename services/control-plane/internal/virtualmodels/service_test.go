package virtualmodels

import (
	"context"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/require"
)

type virtualModelRepositoryStub struct {
	ManagementRepository
	model *platform.VirtualModel
}

func (s virtualModelRepositoryStub) GetVirtualModel(context.Context, string, string) (*platform.VirtualModel, error) {
	return s.model, nil
}

func TestNewServiceRequiresRepository(t *testing.T) {
	t.Parallel()

	service, err := NewService(nil)

	require.ErrorIs(t, err, errRepositoryRequired)
	require.Nil(t, service)
}

func TestServiceDelegatesVirtualModelUseCase(t *testing.T) {
	t.Parallel()

	want := &platform.VirtualModel{ID: "vm-1"}
	service, err := NewService(virtualModelRepositoryStub{model: want})
	require.NoError(t, err)

	got, err := service.GetVirtualModel(context.Background(), "tenant-1", "vm-1")
	require.NoError(t, err)
	require.Equal(t, want, got)
}
