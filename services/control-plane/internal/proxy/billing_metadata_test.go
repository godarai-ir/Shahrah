package proxy

import (
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestDeploymentBillingModeTreatsMockProviderAsExternalBilled(t *testing.T) {
	t.Parallel()

	deployment := platform.Deployment{
		Provider:    "mock",
		BillingMode: platform.BillingModeWalletBilled,
	}

	if got := deploymentBillingMode(deployment); got != platform.BillingModeExternalBilled {
		t.Fatalf("expected mock provider billing mode %q, got %q", platform.BillingModeExternalBilled, got)
	}
}
