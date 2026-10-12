package admin

import (
	"context"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apiheaders"
	"github.com/Delpak/shahrah/services/control-plane/internal/apperrors"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// onBehalfOfKeyStore implements only the method under test; the embedded
// interface panics if the handler reaches any other key-store method.
type onBehalfOfKeyStore struct {
	keyStore
	calls  int
	req    platform.UpdateVATOnBehalfOfRequest
	actor  string
	result error
}

func (s *onBehalfOfKeyStore) SetVATOnBehalfOf(
	_ context.Context,
	_ string,
	_ string,
	_ string,
	req platform.UpdateVATOnBehalfOfRequest,
	actor string,
) (*platform.VATOnBehalfOf, error) {
	s.calls++
	s.req, s.actor = req, actor

	if s.result != nil {
		return nil, s.result
	}

	return &platform.VATOnBehalfOf{Enabled: req.Enabled, Product: req.Product, UpdatedBy: actor}, nil
}

func putOnBehalfOf(t *testing.T, store *onBehalfOfKeyStore, role, body string) (int, string) {
	t.Helper()

	h := &Handlers{iamState: iamState{keys: store}}
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(auth.AuthContextKey{}, &platform.AuthContext{Role: role})
		c.Locals(tenant.ContextKey{}, &platform.Tenant{ID: "tenant_a", Slug: "a", Status: "active"})

		return c.Next()
	})
	app.Put("/api/v1/virtual-users/:virtualUserID/tokens/:tokenID/on-behalf-of", h.UpdateVirtualUserTokenOnBehalfOf)

	req := httptest.NewRequest(fiber.MethodPut, "/api/v1/virtual-users/vu-1/tokens/key-1/on-behalf-of", strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	req.Header.Set(apiheaders.HeaderUserID, "admin-user")

	resp, err := app.Test(req)
	require.NoError(t, err)

	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	return resp.StatusCode, string(raw)
}

func TestUpdateVirtualUserTokenOnBehalfOfRequiresTenantAdmin(t *testing.T) {
	t.Parallel()

	for _, role := range []string{"team_admin", "member", "org_admin"} {
		store := &onBehalfOfKeyStore{}

		status, _ := putOnBehalfOf(t, store, role, `{"enabled":true}`)
		require.Equal(t, fiber.StatusForbidden, status, role)
		require.Zero(t, store.calls, "%s must not reach the store", role)
	}
}

func TestUpdateVirtualUserTokenOnBehalfOfDefaultsAndRecordsActor(t *testing.T) {
	t.Parallel()

	store := &onBehalfOfKeyStore{}

	status, body := putOnBehalfOf(t, store, "tenant_admin", `{"enabled":true}`)
	require.Equal(t, fiber.StatusOK, status, body)
	require.Equal(t, platform.CreditProductLibreChat, store.req.Product)
	require.Equal(t, "admin-user", store.actor)
	require.Contains(t, body, `"enabled":true`)
}

func TestUpdateVirtualUserTokenOnBehalfOfValidation(t *testing.T) {
	t.Parallel()

	store := &onBehalfOfKeyStore{}

	status, _ := putOnBehalfOf(t, store, "tenant_admin", `{"enabled":true,"product":"guardrail"}`)
	require.Equal(t, fiber.StatusBadRequest, status)
	require.Zero(t, store.calls)

	status, _ = putOnBehalfOf(t, store, "tenant_admin", `{"enabled":false,"product":"anything"}`)
	require.Equal(t, fiber.StatusOK, status, "disabling ignores the product")

	notFound := &onBehalfOfKeyStore{result: apperrors.ErrNotFound}
	status, _ = putOnBehalfOf(t, notFound, "tenant_admin", `{"enabled":true}`)
	require.Equal(t, fiber.StatusNotFound, status)
}
