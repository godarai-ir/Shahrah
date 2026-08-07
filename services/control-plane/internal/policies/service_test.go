package policies

import (
	"context"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/require"
)

type policyRepositoryStub struct {
	repository
	guardrails []platform.TenantGuardrail
}

func (s policyRepositoryStub) ListEnabledTenantGuardrails(context.Context, string) ([]platform.TenantGuardrail, error) {
	return s.guardrails, nil
}

func TestNewServiceRejectsNilRepository(t *testing.T) {
	service, err := NewService(nil)

	require.ErrorIs(t, err, errRepositoryRequired)
	require.Nil(t, service)
}

func TestServiceDelegatesPolicyUseCase(t *testing.T) {
	want := []platform.TenantGuardrail{{Slug: "safe-output"}}
	service, err := NewService(policyRepositoryStub{guardrails: want})
	require.NoError(t, err)

	got, err := service.ListEnabledTenantGuardrails(context.Background(), "tenant-1")
	require.NoError(t, err)
	require.Equal(t, want, got)
}
