package admin

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	billingdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/seats"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// seatStore manages member seats, seat orders, and upgrade requests.
type seatStore interface {
	GetSeatSummary(ctx context.Context, tenantID string, now time.Time) (*platform.SeatSummary, error)
	ListSeatMembers(ctx context.Context, tenantID, query, seatFilter string, limit, offset int, now time.Time) (*platform.SeatMemberList, error)
	SetMemberSeat(ctx context.Context, tenantID, userID string, target seats.Type, actor, requestID string, now time.Time) (*platform.SetMemberSeatResult, error)
	CancelPendingSeatChange(ctx context.Context, tenantID, userID string, now time.Time) (*platform.SeatMember, error)
	QuoteSeats(ctx context.Context, tenantID string, req platform.SeatQuoteRequest, now time.Time) (*platform.SeatQuote, error)
	QuoteSeatRenewal(ctx context.Context, tenantID string, now time.Time) (*platform.SeatRenewalQuote, error)
	UpdateSeatRenewal(ctx context.Context, tenantID string, req platform.UpdateSeatRenewalRequest, now time.Time) (*platform.SeatSummary, error)
	CreateSeatOrder(ctx context.Context, tenantID string, req platform.CreateSeatOrderRequest, callbackURL, returnURL, actor string, now time.Time) (*platform.SeatOrder, error)
	AttachSeatOrderGatewayData(ctx context.Context, publicID, authority, paymentURL string, raw map[string]interface{}) error
	GetSeatOrderByAuthority(ctx context.Context, authority string) (*platform.SeatOrder, error)
	ListSeatOrders(ctx context.Context, tenantID string, limit int) ([]platform.SeatOrder, error)
	CompleteSeatOrderByAuthority(ctx context.Context, authority string, verified billingdomain.PaymentVerification, now time.Time) (*platform.SeatOrder, bool, error)
	MarkSeatOrderFailedByAuthority(ctx context.Context, authority, reason string, raw map[string]interface{}) error
	ListSeatUpgradeRequests(ctx context.Context, tenantID, status string, limit int) ([]platform.SeatUpgradeRequest, error)
	GetSeatUpgradeRequest(ctx context.Context, tenantID, requestID string) (*platform.SeatUpgradeRequest, error)
	DenySeatUpgradeRequest(ctx context.Context, tenantID, requestID, actor, note string, now time.Time) (*platform.SeatUpgradeRequest, error)
	ListSeatTypes(ctx context.Context) ([]platform.SeatTypeOffer, error)
	UpdateSeatType(ctx context.Context, seatType string, req platform.SeatTypeUpdate, actor string) ([]platform.SeatTypeOffer, error)
}

const seatCallbackPath = "/api/v1/seats/payment/callback/"

func (h *Handlers) requireSeatStore() (seatStore, error) {
	if h.seats == nil {
		return nil, fiber.NewError(fiber.StatusNotImplemented, "seats are not configured")
	}

	return h.seats, nil
}

func (h *Handlers) seatContext(c *fiber.Ctx) (seatStore, *platform.Tenant, error) {
	store, err := h.requireSeatStore()
	if err != nil {
		return nil, nil, err
	}

	currentTenant, err := currentTenantOrErr(c)
	if err != nil {
		return nil, nil, err
	}

	return store, currentTenant, nil
}

// requireSeatRead lets billing readers and member managers see seats.
func requireSeatRead(c *fiber.Ctx) error {
	return requirePermissionAnyOrTenantAdmin(c, "seat read permission required",
		[]string{"billing:read", "billing:manage", "settings:list", "settings:manage", "user:list", "user:manage"})
}

// GetSeatSummary returns the tenant's seat overview.
func (h *Handlers) GetSeatSummary(c *fiber.Ctx) error {
	if err := requireSeatRead(c); err != nil {
		return err
	}

	store, currentTenant, err := h.seatContext(c)
	if err != nil {
		return err
	}

	summary, err := store.GetSeatSummary(c.Context(), currentTenant.ID, time.Now())
	if err != nil {
		return seatError(c, err)
	}

	return c.JSON(summary)
}

// ListSeatMembers lists members with their seats and personal credit usage.
func (h *Handlers) ListSeatMembers(c *fiber.Ctx) error {
	if err := requireSeatRead(c); err != nil {
		return err
	}

	store, currentTenant, err := h.seatContext(c)
	if err != nil {
		return err
	}

	members, err := store.ListSeatMembers(c.Context(), currentTenant.ID, c.Query("q"), c.Query("seat"),
		c.QueryInt("limit", 100), c.QueryInt("offset", 0), time.Now())
	if err != nil {
		return seatError(c, err)
	}

	return c.JSON(members)
}

// SetMemberSeat changes one member's seat.
func (h *Handlers) SetMemberSeat(c *fiber.Ctx) error {
	if err := requireUserManage(c); err != nil {
		return err
	}

	store, currentTenant, err := h.seatContext(c)
	if err != nil {
		return err
	}

	var req platform.SetMemberSeatRequest
	if parseErr := c.BodyParser(&req); parseErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	target, err := seats.ParseType(req.Seat)
	if err != nil {
		return seatError(c, err)
	}

	userID := strings.TrimSpace(c.Params("userID"))
	if _, parseErr := uuid.Parse(userID); parseErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "userID must be a UUID")
	}

	result, err := store.SetMemberSeat(c.Context(), currentTenant.ID, userID, target, billingActorID(c), req.RequestID, time.Now())
	if err != nil {
		return h.seatChangeError(c, store, currentTenant.ID, target, err)
	}

	return c.JSON(result)
}

// seatChangeError answers a paid upgrade that has no seat in stock with the
// price of buying one, so the client can offer checkout directly.
func (h *Handlers) seatChangeError(c *fiber.Ctx, store seatStore, tenantID string, target seats.Type, err error) error {
	if !errors.Is(err, seats.ErrSeatPurchaseRequired) {
		return seatError(c, err)
	}

	body := fiber.Map{"code": "seat_purchase_required", "type": "seats", "message": err.Error()}

	if quote, quoteErr := store.QuoteSeats(c.Context(), tenantID, platform.SeatQuoteRequest{Seat: string(target), Quantity: 1}, time.Now()); quoteErr == nil {
		body["quote"] = quote
	}

	return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": body})
}

// CancelPendingSeatChange keeps a member's seat instead of the scheduled change.
func (h *Handlers) CancelPendingSeatChange(c *fiber.Ctx) error {
	if err := requireUserManage(c); err != nil {
		return err
	}

	store, currentTenant, err := h.seatContext(c)
	if err != nil {
		return err
	}

	member, err := store.CancelPendingSeatChange(c.Context(), currentTenant.ID, c.Params("userID"), time.Now())
	if err != nil {
		return seatError(c, err)
	}

	return c.JSON(member)
}

// QuoteSeats prices buying paid seats for the rest of the billing period.
func (h *Handlers) QuoteSeats(c *fiber.Ctx) error {
	if err := requireSeatRead(c); err != nil {
		return err
	}

	store, currentTenant, err := h.seatContext(c)
	if err != nil {
		return err
	}

	var req platform.SeatQuoteRequest
	if parseErr := c.BodyParser(&req); parseErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	quote, err := store.QuoteSeats(c.Context(), currentTenant.ID, req, time.Now())
	if err != nil {
		return seatError(c, err)
	}

	return c.JSON(quote)
}

// QuoteSeatRenewal prices renewing the paid seats that lapse before the
// subscription period ends.
func (h *Handlers) QuoteSeatRenewal(c *fiber.Ctx) error {
	if err := requireSeatRead(c); err != nil {
		return err
	}

	store, currentTenant, err := h.seatContext(c)
	if err != nil {
		return err
	}

	quote, err := store.QuoteSeatRenewal(c.Context(), currentTenant.ID, time.Now())
	if err != nil {
		return seatError(c, err)
	}

	return c.JSON(quote)
}

// UpdateSeatRenewal sets how many paid seats of a type the next renewal keeps.
func (h *Handlers) UpdateSeatRenewal(c *fiber.Ctx) error {
	if err := requireBillingManage(c); err != nil {
		return err
	}

	store, currentTenant, err := h.seatContext(c)
	if err != nil {
		return err
	}

	var req platform.UpdateSeatRenewalRequest
	if parseErr := c.BodyParser(&req); parseErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	summary, err := store.UpdateSeatRenewal(c.Context(), currentTenant.ID, req, time.Now())
	if err != nil {
		return seatError(c, err)
	}

	return c.JSON(summary)
}

// ListSeatOrders returns the tenant's recent seat orders.
func (h *Handlers) ListSeatOrders(c *fiber.Ctx) error {
	if err := requireSeatRead(c); err != nil {
		return err
	}

	store, currentTenant, err := h.seatContext(c)
	if err != nil {
		return err
	}

	orders, err := store.ListSeatOrders(c.Context(), currentTenant.ID, c.QueryInt("limit", 20))
	if err != nil {
		return seatError(c, err)
	}

	return c.JSON(fiber.Map{"data": orders})
}

// CreateSeatOrder starts a gateway payment for buying or renewing seats.
func (h *Handlers) CreateSeatOrder(c *fiber.Ctx) error {
	if err := requireBillingManage(c); err != nil {
		return err
	}

	store, currentTenant, err := h.seatContext(c)
	if err != nil {
		return err
	}

	var req platform.CreateSeatOrderRequest
	if parseErr := c.BodyParser(&req); parseErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	req.Provider = strings.ToLower(strings.TrimSpace(req.Provider))
	if req.Provider == "" {
		req.Provider = h.payment.DefaultSubscriptionProvider
	}

	// Seats are paid through a gateway; bank transfers are for credit packs.
	if req.Provider == platform.CreditPackProviderBankTransfer || !h.creditPackProviderAllowed(req.Provider) {
		return creditPackError(c, billingdomain.ErrCreditPaymentProviderUnavailable)
	}

	baseURL := strings.TrimSuffix(resolveExternalBaseURL(c), "/")

	returnURL := strings.TrimSpace(req.ReturnURL)
	if returnURL == "" {
		returnURL = baseURL + "/billing/seats"
	}

	order, err := store.CreateSeatOrder(c.Context(), currentTenant.ID, req, baseURL+seatCallbackPath+req.Provider, returnURL, billingActorID(c), time.Now())
	if err != nil {
		return seatError(c, err)
	}

	payment, err := h.selectPaymentProvider(order.Provider).CreatePayment(c, &platform.WalletTopUpIntent{
		PublicID: order.PublicID, TenantID: order.TenantID, Provider: order.Provider,
		PayableIRR: float64(order.PayableIRR), CallbackURL: order.CallbackURL, ReturnURL: order.ReturnURL,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, "payment gateway is unavailable")
	}

	if err := store.AttachSeatOrderGatewayData(c.Context(), order.PublicID, payment.Authority, payment.PaymentURL, payment.Raw); err != nil {
		return seatError(c, err)
	}

	order.Authority = payment.Authority
	order.PaymentURL = payment.PaymentURL

	return c.Status(fiber.StatusCreated).JSON(order)
}

// SeatPaymentCallback completes a gateway payment for a seat order. The
// order's own provider verifies the payment; the URL's provider is not trusted.
func (h *Handlers) SeatPaymentCallback(c *fiber.Ctx) error {
	store, err := h.requireSeatStore()
	if err != nil {
		return err
	}

	authority := strings.TrimSpace(c.Query("authority"))
	if authority == "" {
		authority = strings.TrimSpace(c.FormValue("authority"))
	}

	if authority == "" {
		return fiber.NewError(fiber.StatusBadRequest, "authority is required")
	}

	order, err := store.GetSeatOrderByAuthority(c.Context(), authority)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "failed", "reason": "order_not_found"})
	}

	setTopUpIntent(c, &platform.WalletTopUpIntent{
		PublicID: order.PublicID, TenantID: order.TenantID, Provider: order.Provider,
		PayableIRR: float64(order.PayableIRR), CallbackURL: order.CallbackURL, ReturnURL: order.ReturnURL,
	})

	verified, err := h.selectPaymentProvider(order.Provider).VerifyPayment(c, authority)
	if err != nil || !verified.Success {
		reason := "payment failed"

		var raw map[string]interface{}

		if verified != nil {
			raw = verified.Raw
			if verified.Failure != "" {
				reason = verified.Failure
			}
		}

		if markErr := store.MarkSeatOrderFailedByAuthority(c.Context(), authority, reason, raw); markErr != nil {
			return markErr
		}

		return c.Redirect(subscriptionPaymentReturnURL(order.ReturnURL, "failed", order.PublicID, authority, "", "payment_failed"), fiber.StatusFound)
	}

	completed, _, err := store.CompleteSeatOrderByAuthority(c.Context(), authority, billingdomain.PaymentVerification{
		Raw: verified.Raw, RefID: verified.RefID, GatewayTxnID: verified.GatewayTxnID,
		CardMask: verified.CardMask, PayerPhone: verified.PayerPhone,
	}, time.Now())
	if err != nil {
		return c.Redirect(subscriptionPaymentReturnURL(order.ReturnURL, "failed", order.PublicID, authority, "", "commit_failed"), fiber.StatusFound)
	}

	return c.Redirect(subscriptionPaymentReturnURL(completed.ReturnURL, "success", completed.PublicID, authority, completed.RefID, ""), fiber.StatusFound)
}

// ListSeatRequests returns the tenant's seat upgrade requests.
func (h *Handlers) ListSeatRequests(c *fiber.Ctx) error {
	if err := requireUserListOrManage(c); err != nil {
		return err
	}

	store, currentTenant, err := h.seatContext(c)
	if err != nil {
		return err
	}

	requests, err := store.ListSeatUpgradeRequests(c.Context(), currentTenant.ID, c.Query("status"), c.QueryInt("limit", 50))
	if err != nil {
		return seatError(c, err)
	}

	return c.JSON(fiber.Map{"data": requests})
}

// ApproveSeatRequest gives the member the seat they asked for. Without a seat
// in stock it answers 409 seat_purchase_required with the price of one.
func (h *Handlers) ApproveSeatRequest(c *fiber.Ctx) error {
	if err := requireUserManage(c); err != nil {
		return err
	}

	store, currentTenant, err := h.seatContext(c)
	if err != nil {
		return err
	}

	request, err := store.GetSeatUpgradeRequest(c.Context(), currentTenant.ID, c.Params("requestID"))
	if err != nil {
		return seatError(c, err)
	}

	if request.Status != platform.SeatRequestPending {
		return seatError(c, seats.ErrRequestClosed)
	}

	target := seats.Type(request.RequestedSeat)

	result, err := store.SetMemberSeat(c.Context(), currentTenant.ID, request.UserID, target, billingActorID(c), request.ID, time.Now())
	if err != nil {
		return h.seatChangeError(c, store, currentTenant.ID, target, err)
	}

	return c.JSON(result)
}

// DenySeatRequest closes a request without changing the member's seat.
func (h *Handlers) DenySeatRequest(c *fiber.Ctx) error {
	if err := requireUserManage(c); err != nil {
		return err
	}

	store, currentTenant, err := h.seatContext(c)
	if err != nil {
		return err
	}

	var req platform.DecideSeatUpgradeRequest
	if len(c.Body()) > 0 {
		if parseErr := c.BodyParser(&req); parseErr != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}
	}

	if len(req.Note) > 1000 {
		return fiber.NewError(fiber.StatusBadRequest, "note is too long")
	}

	request, err := store.DenySeatUpgradeRequest(c.Context(), currentTenant.ID, c.Params("requestID"), billingActorID(c), req.Note, time.Now())
	if err != nil {
		return seatError(c, err)
	}

	return c.JSON(request)
}

// ListPlatformSeatTypes returns every seat type with its current price.
func (h *Handlers) ListPlatformSeatTypes(c *fiber.Ctx) error {
	store, err := h.requireSeatStore()
	if err != nil {
		return err
	}

	types, err := store.ListSeatTypes(c.Context())
	if err != nil {
		return seatError(c, err)
	}

	return c.JSON(fiber.Map{"data": types})
}

// UpdatePlatformSeatType changes a seat type's quota, discount, or availability.
func (h *Handlers) UpdatePlatformSeatType(c *fiber.Ctx) error {
	store, err := h.requireSeatStore()
	if err != nil {
		return err
	}

	var req platform.SeatTypeUpdate
	if parseErr := c.BodyParser(&req); parseErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	types, err := store.UpdateSeatType(c.Context(), c.Params("seatType"), req, platformActor(c))
	if err != nil {
		return seatError(c, err)
	}

	return c.JSON(fiber.Map{"data": types})
}

// GetPlatformTenantSeats returns one tenant's seat overview and members.
func (h *Handlers) GetPlatformTenantSeats(c *fiber.Ctx) error {
	store, err := h.requireSeatStore()
	if err != nil {
		return err
	}

	tenantID := strings.TrimSpace(c.Params("tenantID"))
	if _, parseErr := uuid.Parse(tenantID); parseErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenantID must be a UUID")
	}

	now := time.Now()

	summary, err := store.GetSeatSummary(c.Context(), tenantID, now)
	if err != nil {
		return seatError(c, err)
	}

	members, err := store.ListSeatMembers(c.Context(), tenantID, c.Query("q"), c.Query("seat"), c.QueryInt("limit", 200), c.QueryInt("offset", 0), now)
	if err != nil {
		return seatError(c, err)
	}

	return c.JSON(fiber.Map{"summary": summary, "members": members})
}

// seatError maps seat errors to client-safe responses with a stable code.
func seatError(c *fiber.Ctx, err error) error {
	status, code := seatErrorStatus(err)
	if status == 0 {
		return creditPackError(c, err)
	}

	return c.Status(status).JSON(fiber.Map{"error": fiber.Map{"code": code, "type": "seats", "message": err.Error()}})
}

func seatErrorStatus(err error) (int, string) {
	for _, mapping := range []struct {
		target error
		status int
		code   string
	}{
		{seats.ErrInvalidSeatType, fiber.StatusBadRequest, "invalid_seat_type"},
		{seats.ErrInvalidQuantity, fiber.StatusBadRequest, "invalid_seat_quantity"},
		{seats.ErrIntervalNotSupported, fiber.StatusConflict, "seat_interval_not_supported"},
		{seats.ErrIntervalLocked, fiber.StatusConflict, "seat_interval_locked"},
		{seats.ErrSameSeat, fiber.StatusConflict, "seat_unchanged"},
		{seats.ErrBackToFree, fiber.StatusConflict, "seat_back_to_free"},
		{seats.ErrFreeSeatUsed, fiber.StatusConflict, "free_seat_used"},
		{seats.ErrFreeSeatsFull, fiber.StatusConflict, "free_seats_full"},
		{seats.ErrSeatsUnavailable, fiber.StatusConflict, "seats_need_active_subscription"},
		{seats.ErrSeatPurchaseRequired, fiber.StatusConflict, "seat_purchase_required"},
		{seats.ErrMemberNotFound, fiber.StatusNotFound, "member_not_found"},
		{seats.ErrNoPendingChange, fiber.StatusNotFound, "no_pending_seat_change"},
		{seats.ErrRequestNotFound, fiber.StatusNotFound, "seat_request_not_found"},
		{seats.ErrRequestClosed, fiber.StatusConflict, "seat_request_closed"},
		{seats.ErrRequestsDisabled, fiber.StatusForbidden, "seat_requests_disabled"},
		{seats.ErrRequestOpen, fiber.StatusConflict, "seat_request_open"},
		{seats.ErrRequestNotUpgrade, fiber.StatusBadRequest, "seat_request_not_upgrade"},
		{seats.ErrOrderNotFound, fiber.StatusNotFound, "seat_order_not_found"},
		{seats.ErrOrderNotPayable, fiber.StatusConflict, "seat_order_not_payable"},
		{seats.ErrRenewalNotDue, fiber.StatusConflict, "seat_renewal_not_due"},
		{seats.ErrInvalidRenewal, fiber.StatusBadRequest, "invalid_seat_renewal"},
		{billingdomain.ErrInvalidSeatPoolLimit, fiber.StatusBadRequest, "invalid_seat_pool_limit"},
	} {
		if errors.Is(err, mapping.target) {
			return mapping.status, mapping.code
		}
	}

	return 0, ""
}
