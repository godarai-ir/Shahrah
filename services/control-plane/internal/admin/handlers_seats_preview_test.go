package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/seats"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// seatPreviewStoreFake answers renewal previews; the other seat methods are not used.
type seatPreviewStoreFake struct {
	seatStore
	gotKeep map[seats.Type]int
	err     error
}

func (f *seatPreviewStoreFake) PreviewSeatRenewal(_ context.Context, _ string, keep map[seats.Type]int, _ time.Time) (*platform.SeatRenewalPreview, error) {
	f.gotKeep = keep
	if f.err != nil {
		return nil, f.err
	}

	assigned := time.Date(2026, time.October, 5, 9, 0, 0, 0, time.UTC)

	return &platform.SeatRenewalPreview{Lines: []platform.SeatRenewalPreviewLine{{
		Seat: "pro", Held: 3, Keep: keep[seats.Pro], Assigned: 3,
		Losing: []platform.SeatRenewalPreviewMember{{UserID: "u3", Email: "sara@example.com", Name: "Sara", AssignedAt: &assigned, Fallback: "unassigned"}},
	}}}, nil
}

func seatPreviewApp(store *seatPreviewStoreFake, permissions ...string) *fiber.App {
	h := &Handlers{billingState: billingState{seats: store}}
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(auth.AuthContextKey{}, &platform.AuthContext{Role: "member", UserID: "user-1", Permissions: permissions})
		c.Locals(tenant.ContextKey{}, &platform.Tenant{ID: "tenant_a", Slug: "a", Status: "active"})

		return c.Next()
	})
	app.Get("/api/v1/billing/seats/renewal/preview", h.PreviewSeatRenewal)

	return app
}

func getSeatPreview(t *testing.T, app *fiber.App, query string) (int, platform.SeatRenewalPreview) {
	t.Helper()

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/api/v1/billing/seats/renewal/preview"+query, http.NoBody))
	require.NoError(t, err)

	var preview platform.SeatRenewalPreview

	_ = json.NewDecoder(resp.Body).Decode(&preview)

	require.NoError(t, resp.Body.Close())

	return resp.StatusCode, preview
}

func TestPreviewSeatRenewalPassesTheRequestedCounts(t *testing.T) {
	store := &seatPreviewStoreFake{}
	app := seatPreviewApp(store, "billing:read", "user:list")

	status, preview := getSeatPreview(t, app, "?pro=2&max=0")
	require.Equal(t, fiber.StatusOK, status)
	require.Equal(t, map[seats.Type]int{seats.Pro: 2, seats.Max: 0}, store.gotKeep)
	require.Equal(t, "sara@example.com", preview.Lines[0].Losing[0].Email)

	_, _ = getSeatPreview(t, app, "")
	require.Empty(t, store.gotKeep, "a missing count is left to the saved one")
}

func TestPreviewSeatRenewalRejectsBadCounts(t *testing.T) {
	app := seatPreviewApp(&seatPreviewStoreFake{}, "billing:read")

	for _, query := range []string{"?pro=-1", "?pro=abc", "?max=1.5"} {
		status, _ := getSeatPreview(t, app, query)
		require.Equal(t, fiber.StatusBadRequest, status, query)
	}
}

func TestPreviewSeatRenewalHidesMemberIdentityWithoutUserAccess(t *testing.T) {
	app := seatPreviewApp(&seatPreviewStoreFake{}, "billing:read")

	status, preview := getSeatPreview(t, app, "?pro=2")
	require.Equal(t, fiber.StatusOK, status)

	member := preview.Lines[0].Losing[0]
	require.Equal(t, "u3", member.UserID)
	require.Empty(t, member.Email)
	require.Empty(t, member.Name)
	require.Equal(t, "unassigned", member.Fallback, "what happens to the member is still shown")
}

func TestPreviewSeatRenewalNeedsSeatReadAndMapsErrors(t *testing.T) {
	status, _ := getSeatPreview(t, seatPreviewApp(&seatPreviewStoreFake{}), "")
	require.Equal(t, fiber.StatusForbidden, status)

	status, _ = getSeatPreview(t, seatPreviewApp(&seatPreviewStoreFake{err: seats.ErrRenewalNotDue}, "billing:read"), "")
	require.Equal(t, fiber.StatusConflict, status)

	status, _ = getSeatPreview(t, seatPreviewApp(&seatPreviewStoreFake{err: seats.ErrInvalidRenewal}, "billing:read"), "?pro=99")
	require.Equal(t, fiber.StatusBadRequest, status)
}
