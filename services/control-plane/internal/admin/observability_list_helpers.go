package admin

import (
	"errors"

	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
)

func listObservabilityItems[TItem, TParams any](
	c *fiber.Ctx,
	authorize func(*fiber.Ctx) error,
	configured bool,
	notConfiguredMessage string,
	parse func(string) (TParams, error),
	list func(TParams) ([]TItem, int, error),
	unavailableError error,
	pagination func(TParams) (int, int),
	respond func([]TItem, int, int, int, *int) error,
) error {
	if err := authorize(c); err != nil {
		return err
	}

	if !configured {
		return fiber.NewError(fiber.StatusNotImplemented, notConfiguredMessage)
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	params, err := parse(currentTenant.ID)
	if err != nil {
		return err
	}

	items, total, err := list(params)
	if err != nil {
		if errors.Is(err, unavailableError) {
			return fiber.NewError(fiber.StatusNotImplemented, notConfiguredMessage)
		}

		return err
	}

	limit, offset := pagination(params)

	var nextOffset *int

	if offset+len(items) < total {
		next := offset + len(items)
		nextOffset = &next
	}

	return respond(items, total, limit, offset, nextOffset)
}
