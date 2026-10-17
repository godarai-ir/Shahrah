package proxy

import (
	"context"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/gofiber/fiber/v2"
)

// BudgetStatusReader reports the caller's credit budgets.
type BudgetStatusReader interface {
	BudgetStatus(ctx context.Context, auth *platform.AuthContext, includeBalance bool) (*platform.BudgetStatus, error)
}

// billingReadPermissions let a caller see the tenant balance on budget/me.
var billingReadPermissions = []string{"billing:read", "billing:manage"}

// NewBudgetMeHandler serves GET /v1/budget/me: the credit budget rules that
// apply to the caller, and the tenant balance for callers who may read billing.
// It works with any gateway credential, including members' OIDC access tokens.
func NewBudgetMeHandler(statuses BudgetStatusReader) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authInfo := auth.GetAuthInfo(c)
		if authInfo == nil || strings.TrimSpace(authInfo.TenantID) == "" {
			return writeAPIErrorDetail(c, fiber.StatusUnauthorized, "Authentication required", "authentication_error", "", "unauthorized")
		}

		status, err := statuses.BudgetStatus(c.Context(), authInfo, canReadBilling(authInfo))
		if err != nil {
			return writeAPIErrorDetail(c, fiber.StatusServiceUnavailable, billingUnavailableReason, billingUnavailableCode, "", billingUnavailableCode)
		}

		c.Set(fiber.HeaderCacheControl, "no-store")

		return c.JSON(status)
	}
}

func canReadBilling(authInfo *platform.AuthContext) bool {
	switch strings.ToLower(strings.TrimSpace(authInfo.Role)) {
	case "admin", "tenant_admin":
		return true
	}

	for _, granted := range authInfo.Permissions {
		for _, permission := range billingReadPermissions {
			if strings.EqualFold(strings.TrimSpace(granted), permission) {
				return true
			}
		}
	}

	return false
}
