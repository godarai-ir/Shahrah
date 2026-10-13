package proxy

import (
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
)

func enrichPricingMetadata(input map[string]interface{}, authInfo *platform.AuthContext, model string) map[string]interface{} {
	out := map[string]interface{}{}
	for k, v := range input {
		out[k] = v
	}

	if authInfo == nil {
		return out
	}

	if mode, ok := authInfo.AllowedModelPricingMode[model]; ok && mode != "" {
		out["pricing_mode"] = mode
	}

	if resolved, ok := authInfo.AllowedModelPricingResolved[model]; ok {
		out["pricing_resolved"] = resolved
		out["pricing_unresolved"] = !resolved
	}

	if product := strings.TrimSpace(authInfo.CreditProduct); product != "" {
		out[credits.ProductMetadataKey] = credits.TrustedProduct(product)
	}

	return out
}
