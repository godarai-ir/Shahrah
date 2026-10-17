package admin

import (
	"errors"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	billingdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/gofiber/fiber/v2"
)

// GetCreditUsageSummary returns shadow credit usage for the tenant over a report preset.
func (h *Handlers) GetCreditUsageSummary(c *fiber.Ctx) error {
	if err := requireBillingRead(c); err != nil {
		return err
	}

	currentTenant, err := currentTenantOrErr(c)
	if err != nil {
		return err
	}

	preset := strings.TrimSpace(c.Query("preset", "current_month"))

	start, end, normalizedPreset, err := billingdomain.BillingReportWindow(time.Now(), preset)
	if err != nil {
		if errors.Is(err, billingdomain.ErrBillingReportPresetInvalid) {
			return fiber.NewError(fiber.StatusBadRequest, "preset is invalid")
		}

		return err
	}

	if h.creditUsage == nil {
		return c.JSON(emptyCreditUsageSummary(normalizedPreset, start, end))
	}

	summary, err := h.creditUsage.GetCreditUsageSummary(c.Context(), currentTenant.ID, start, end)
	if err != nil {
		return err
	}

	summary.Preset = normalizedPreset

	return c.JSON(summary)
}

func emptyCreditUsageSummary(preset string, start, end time.Time) platform.CreditUsageSummary {
	return platform.CreditUsageSummary{
		Preset:            preset,
		PeriodStart:       start,
		PeriodEnd:         end,
		TotalCredits:      credits.Format(0),
		CacheSavedCredits: credits.Format(0),
		ByModel:           []platform.CreditUsageItem{},
		ByTeam:            []platform.CreditUsageItem{},
		ByProduct:         []platform.CreditUsageItem{},
	}
}
