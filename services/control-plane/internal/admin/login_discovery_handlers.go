package admin

import (
	"errors"
	"net/mail"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apperrors"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
)

type loginDiscoveryTenant struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Domain    string `json:"domain,omitempty"`
	LoginURL  string `json:"loginUrl,omitempty"`
	LoginURL2 string `json:"login_url,omitempty"`
}

func (h *Handlers) DiscoverTenantsByIdentifier(c *fiber.Ctx) error {
	store, err := h.requireLoginDiscoveryStore()
	if err != nil {
		return err
	}

	channel, normalized, ok := normalizeDiscoveryIdentifier(c.Query("identifier"))
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, "identifier must be a valid email or mobile number")
	}

	tenants, err := store.ListTenantsByAdminIdentifier(c.Context(), channel, normalized)
	if err != nil {
		return err
	}

	items := mapDiscoveryTenants(tenants)

	status := "none"
	if len(items) == 1 {
		status = "single"
	} else if len(items) > 1 {
		status = "multiple"
	}

	return c.JSON(fiber.Map{
		"status":  status,
		"tenants": items,
	})
}

func (h *Handlers) DiscoverTenantByName(c *fiber.Ctx) error {
	store, err := h.requireLoginDiscoveryStore()
	if err != nil {
		return err
	}

	slug := strings.TrimSpace(strings.ToLower(c.Params("slug")))
	if !tenant.ValidateTenantSlug(slug) {
		return fiber.NewError(fiber.StatusBadRequest, "tenant name is invalid")
	}

	item, err := store.GetTenantBySlug(c.Context(), slug)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return c.JSON(fiber.Map{"exists": false})
		}

		return err
	}

	mapped := mapDiscoveryTenants([]platform.Tenant{*item})
	if len(mapped) == 0 {
		return c.JSON(fiber.Map{"exists": false})
	}

	return c.JSON(fiber.Map{
		"exists": true,
		"tenant": mapped[0],
	})
}

func normalizeDiscoveryIdentifier(value string) (string, string, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", "", false
	}

	email := strings.ToLower(trimmed)
	if _, err := mail.ParseAddress(email); err == nil && !strings.ContainsAny(email, " <>") {
		return "email", email, true
	}

	mobile := strings.ReplaceAll(trimmed, " ", "")

	mobile = strings.ReplaceAll(mobile, "-", "")
	if strings.HasPrefix(mobile, "00") {
		mobile = "+" + strings.TrimPrefix(mobile, "00")
	}

	if len(mobile) >= 10 && len(mobile) <= 15 {
		start := 0
		if strings.HasPrefix(mobile, "+") {
			start = 1
		}

		valid := start < len(mobile)
		for _, c := range mobile[start:] {
			if c < '0' || c > '9' {
				valid = false
				break
			}
		}

		if valid {
			return "mobile", mobile, true
		}
	}

	return "", "", false
}

func mapDiscoveryTenants(tenants []platform.Tenant) []loginDiscoveryTenant {
	items := make([]loginDiscoveryTenant, 0, len(tenants))
	for _, tenant := range tenants {
		slug := strings.TrimSpace(tenant.Slug)
		if slug == "" {
			continue
		}

		items = append(items, loginDiscoveryTenant{
			ID:        tenant.ID,
			Name:      slug,
			Slug:      slug,
			Domain:    strings.TrimSpace(tenant.Domain),
			LoginURL:  "",
			LoginURL2: "",
		})
	}

	return items
}
