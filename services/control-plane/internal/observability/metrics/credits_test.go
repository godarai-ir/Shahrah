package metrics

import (
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
)

func TestApplyShadowCredits(t *testing.T) {
	t.Parallel()

	event := platform.ModelMetricEvent{
		ModelName:      "claude-sonnet-5",
		ProviderName:   "anthropic",
		RequestType:    "chat_completions",
		CostUSD:        0.009,
		AvoidedCostUSD: 0.018,
		Metadata:       map[string]interface{}{"playground_source": "tenant_admin_playground"},
	}

	ApplyShadowCredits(&event, credits.RateCard{Version: 5})

	if event.CreditsMC != 4500 || event.CreditsAvoidedMC != 9000 {
		t.Fatalf("credits = %d avoided = %d, want 4500 and 9000", event.CreditsMC, event.CreditsAvoidedMC)
	}

	if event.CreditProduct != platform.CreditProductPlayground || event.RateCardVersion != 5 {
		t.Fatalf("product = %s version = %d", event.CreditProduct, event.RateCardVersion)
	}

	ApplyShadowCredits(nil, credits.IdentityRateCard())
}
