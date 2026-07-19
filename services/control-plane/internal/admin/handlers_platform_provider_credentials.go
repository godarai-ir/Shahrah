package admin

import (
	"errors"
	"reflect"
	"strconv"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/platformcredentials"
	"github.com/gofiber/fiber/v2"
)

func (h *Handlers) GetPlatformProviderConnectionSchemas(c *fiber.Ctx) error {
	service, err := h.requirePlatformProviderCredentials()
	if err != nil {
		return err
	}
	items, err := service.Schemas(c.Context(), c.Params("provider"))
	if err != nil {
		return platformCredentialError(c, err)
	}
	return c.JSON(fiber.Map{"data": items})
}

func (h *Handlers) ListPlatformProviderConnections(c *fiber.Ctx) error {
	service, err := h.requirePlatformProviderCredentials()
	if err != nil {
		return err
	}
	items, err := service.ListConnections(c.Context(), c.Query("provider"), c.Query("environment"))
	if err != nil {
		return platformCredentialError(c, err)
	}
	return c.JSON(fiber.Map{"data": items})
}

func (h *Handlers) GetPlatformProviderConnection(c *fiber.Ctx) error {
	service, err := h.requirePlatformProviderCredentials()
	if err != nil {
		return err
	}
	item, err := service.GetConnection(c.Context(), c.Params("id"))
	if err != nil {
		return platformCredentialError(c, err)
	}
	return c.JSON(fiber.Map{"data": item})
}

func (h *Handlers) CreatePlatformProviderConnection(c *fiber.Ctx) error {
	service, err := h.requirePlatformProviderCredentials()
	if err != nil {
		return err
	}
	var request platform.CreatePlatformProviderConnectionRequest
	if err := c.BodyParser(&request); err != nil {
		return platformCredentialError(c, platformcredentials.ErrInvalidRequest)
	}
	item, err := service.CreateConnection(c.Context(), request, platformCredentialActor(c), strings.TrimSpace(c.Get("Idempotency-Key")))
	if err != nil {
		return platformCredentialError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": item})
}

func (h *Handlers) UpdatePlatformProviderConnection(c *fiber.Ctx) error {
	service, err := h.requirePlatformProviderCredentials()
	if err != nil {
		return err
	}
	var request platform.UpdatePlatformProviderConnectionRequest
	if err := c.BodyParser(&request); err != nil {
		return platformCredentialError(c, platformcredentials.ErrInvalidRequest)
	}
	item, err := service.UpdateConnection(c.Context(), c.Params("id"), request, platformCredentialActor(c))
	if err != nil {
		return platformCredentialError(c, err)
	}
	return c.JSON(fiber.Map{"data": item})
}

func (h *Handlers) CreatePlatformProviderCredentialVersion(c *fiber.Ctx) error {
	service, err := h.requirePlatformProviderCredentials()
	if err != nil {
		return err
	}
	var request platform.CreatePlatformProviderCredentialVersionRequest
	if err := c.BodyParser(&request); err != nil {
		return platformCredentialError(c, platformcredentials.ErrInvalidRequest)
	}
	item, err := service.CreateCredentialVersion(c.Context(), c.Params("id"), request, platformCredentialActor(c), strings.TrimSpace(c.Get("Idempotency-Key")))
	if err != nil {
		return platformCredentialError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": item})
}

func (h *Handlers) TestPlatformProviderCredentialVersion(c *fiber.Ctx) error {
	service, version, err := h.platformCredentialVersionParams(c)
	if err != nil {
		return platformCredentialError(c, err)
	}
	result, err := service.TestCredential(c.Context(), c.Params("id"), version, platformCredentialActor(c))
	if err != nil {
		return platformCredentialError(c, err)
	}
	return c.JSON(fiber.Map{"data": result})
}

func (h *Handlers) ActivatePlatformProviderCredentialVersion(c *fiber.Ctx) error {
	service, version, err := h.platformCredentialVersionParams(c)
	if err != nil {
		return platformCredentialError(c, err)
	}
	var request platform.PlatformProviderCredentialActionRequest
	if err := c.BodyParser(&request); err != nil {
		return platformCredentialError(c, platformcredentials.ErrInvalidRequest)
	}
	item, err := service.Activate(c.Context(), c.Params("id"), version, request.Revision, platformCredentialActor(c), strings.TrimSpace(c.Get("Idempotency-Key")), request.Reason)
	if err != nil {
		return platformCredentialError(c, err)
	}
	return c.JSON(fiber.Map{"data": item})
}

func (h *Handlers) RollbackPlatformProviderCredentialVersion(c *fiber.Ctx) error {
	service, version, err := h.platformCredentialVersionParams(c)
	if err != nil {
		return platformCredentialError(c, err)
	}
	var request platform.PlatformProviderCredentialActionRequest
	if err := c.BodyParser(&request); err != nil {
		return platformCredentialError(c, platformcredentials.ErrInvalidRequest)
	}
	item, err := service.Rollback(c.Context(), c.Params("id"), version, request.Revision, platformCredentialActor(c), request.Reason)
	if err != nil {
		return platformCredentialError(c, err)
	}
	return c.JSON(fiber.Map{"data": item})
}

func (h *Handlers) DisablePlatformProviderConnection(c *fiber.Ctx) error {
	service, err := h.requirePlatformProviderCredentials()
	if err != nil {
		return err
	}
	var request platform.PlatformProviderCredentialActionRequest
	if err := c.BodyParser(&request); err != nil {
		return platformCredentialError(c, platformcredentials.ErrInvalidRequest)
	}
	item, err := service.Disable(c.Context(), c.Params("id"), request.Revision, platformCredentialActor(c), request.Reason)
	if err != nil {
		return platformCredentialError(c, err)
	}
	return c.JSON(fiber.Map{"data": item})
}

func (h *Handlers) RevokePlatformProviderCredentialVersion(c *fiber.Ctx) error {
	service, version, err := h.platformCredentialVersionParams(c)
	if err != nil {
		return platformCredentialError(c, err)
	}
	var request platform.PlatformProviderCredentialActionRequest
	if err := c.BodyParser(&request); err != nil {
		return platformCredentialError(c, platformcredentials.ErrInvalidRequest)
	}
	item, err := service.Revoke(c.Context(), c.Params("id"), version, request.Revision, platformCredentialActor(c), request.Reason)
	if err != nil {
		return platformCredentialError(c, err)
	}
	return c.JSON(fiber.Map{"data": item})
}

func (h *Handlers) platformCredentialVersionParams(c *fiber.Ctx) (platformProviderCredentialService, int64, error) {
	service, err := h.requirePlatformProviderCredentials()
	if err != nil {
		return nil, 0, err
	}
	version, err := strconv.ParseInt(c.Params("version"), 10, 64)
	if err != nil || version <= 0 {
		return service, 0, platformcredentials.ErrInvalidRequest
	}
	return service, version, nil
}

func (h *Handlers) requirePlatformProviderCredentials() (platformProviderCredentialService, error) {
	if h.providerCredentials == nil || isNilPlatformProviderCredentialService(h.providerCredentials) {
		return nil, fiber.NewError(fiber.StatusServiceUnavailable, "platform provider credential service is not configured")
	}
	return h.providerCredentials, nil
}

func isNilPlatformProviderCredentialService(service platformProviderCredentialService) bool {
	value := reflect.ValueOf(service)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func platformCredentialActor(c *fiber.Ctx) platformcredentials.Actor {
	serviceAccountID, _ := c.Locals("platform_service_account_id").(string)
	return platformcredentials.Actor{ID: strings.TrimSpace(c.Get("x-shahrah-actor-id")), Role: strings.TrimSpace(c.Get("x-shahrah-actor-role")), ServiceAccountID: serviceAccountID, RequestID: strings.TrimSpace(c.Get("x-request-id")), SourceIP: c.IP()}
}

func platformCredentialError(c *fiber.Ctx, err error) error {
	status, code, message := fiber.StatusInternalServerError, "platform_provider_credential_internal_error", "The provider credential operation failed."
	switch {
	case errors.Is(err, platformcredentials.ErrInvalidRequest):
		status, code, message = fiber.StatusUnprocessableEntity, err.Error(), "The provider connection request is invalid."
	case errors.Is(err, platformcredentials.ErrRevisionConflict):
		status, code, message = fiber.StatusConflict, err.Error(), "The connection changed; refresh and retry."
	case errors.Is(err, platformcredentials.ErrNotFound):
		status, code, message = fiber.StatusNotFound, err.Error(), "Provider connection or credential version not found."
	case errors.Is(err, platformcredentials.ErrDecryptFailed):
		status, code, message = fiber.StatusServiceUnavailable, err.Error(), "Credential material could not be loaded."
	case errors.Is(err, platformcredentials.ErrRuntimeUnsupported):
		status, code, message = fiber.StatusUnprocessableEntity, err.Error(), "This documented provider authentication mode is not implemented by the runtime yet."
	}
	return c.Status(status).JSON(fiber.Map{"error": fiber.Map{"code": code, "message": message, "request_id": strings.TrimSpace(c.Get("x-request-id"))}})
}
