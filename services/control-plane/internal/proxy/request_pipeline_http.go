package proxy

import (
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/gofiber/fiber/v2"
)

// budgetRequestFromHTTP maps HTTP metadata into the transport-agnostic budget stage input.
func budgetRequestFromHTTP(
	c *fiber.Ctx,
	authInfo *platform.AuthContext,
	estimated platform.UsageCostBreakdown,
	requestID string,
	billingMode ...string,
) BudgetRequest {
	if strings.TrimSpace(estimated.Endpoint) == "" {
		estimated.Endpoint = c.Path()
	}

	if strings.TrimSpace(estimated.RequestType) == "" {
		estimated.RequestType = budgetRequestTypeFromPath(c.Path())
	}

	mode := ""
	if len(billingMode) > 0 {
		mode = billingMode[0]
	}

	return BudgetRequest{
		Auth:        authInfo,
		Estimated:   estimated,
		RequestID:   requestID,
		BillingMode: mode,
	}
}

func budgetRequestTypeFromPath(path string) string {
	normalized := strings.Trim(strings.ToLower(path), "/")
	normalized = strings.TrimPrefix(normalized, "v1/")

	return strings.ReplaceAll(normalized, "/", "_")
}
