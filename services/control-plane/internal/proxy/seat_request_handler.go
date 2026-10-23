package proxy

import (
	"context"
	"errors"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/seats"
	"github.com/gofiber/fiber/v2"
)

// SeatRequestCreator records a member's seat upgrade request.
type SeatRequestCreator interface {
	CreateSeatUpgradeRequest(ctx context.Context, tenantID, userID string, req platform.CreateSeatUpgradeRequest) (*platform.SeatUpgradeRequest, error)
}

// NewSeatRequestHandler serves POST /v1/seat-requests: a member asks their
// organization for a Pro or Max seat, typically from GodarChat after a seat
// refusal. It needs a credential that names a member, such as an OIDC token.
func NewSeatRequestHandler(requests SeatRequestCreator) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authInfo := auth.GetAuthInfo(c)
		if authInfo == nil || strings.TrimSpace(authInfo.TenantID) == "" {
			return writeAPIErrorDetail(c, fiber.StatusUnauthorized, "Authentication required", "authentication_error", "", "unauthorized")
		}

		if strings.TrimSpace(authInfo.UserID) == "" {
			return writeAPIErrorDetail(c, fiber.StatusForbidden, "Seat requests need a member credential", "permission_error", "", "member_required")
		}

		var req platform.CreateSeatUpgradeRequest
		if err := c.BodyParser(&req); err != nil {
			return writeAPIErrorDetail(c, fiber.StatusBadRequest, "Invalid request body", "invalid_request_error", "", "invalid_body")
		}

		created, err := requests.CreateSeatUpgradeRequest(c.Context(), authInfo.TenantID, authInfo.UserID, req)
		if err != nil {
			status, code := seatRequestErrorStatus(err)

			return writeAPIErrorDetail(c, status, err.Error(), "invalid_request_error", "", code)
		}

		return c.Status(fiber.StatusCreated).JSON(created)
	}
}

func seatRequestErrorStatus(err error) (int, string) {
	switch {
	case errors.Is(err, seats.ErrInvalidSeatType):
		return fiber.StatusBadRequest, "invalid_seat_type"
	case errors.Is(err, seats.ErrRequestNotUpgrade):
		return fiber.StatusBadRequest, "seat_request_not_upgrade"
	case errors.Is(err, seats.ErrRequestsDisabled):
		return fiber.StatusForbidden, "seat_requests_disabled"
	case errors.Is(err, seats.ErrRequestOpen):
		return fiber.StatusConflict, "seat_request_open"
	case errors.Is(err, seats.ErrMemberNotFound):
		return fiber.StatusNotFound, "member_not_found"
	default:
		return fiber.StatusServiceUnavailable, "seat_requests_unavailable"
	}
}
