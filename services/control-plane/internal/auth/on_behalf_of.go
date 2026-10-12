package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apperrors"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

const (
	onBehalfOfNotAllowedType   = "on_behalf_of_not_allowed"
	onBehalfOfUserNotFoundType = "on_behalf_of_user_not_found"
	onBehalfOfRoutePrefix      = "/v1/"
)

// OnBehalfOfResolver resolves the tenant member an on-behalf-of service key acts for.
type OnBehalfOfResolver interface {
	ResolveOnBehalfOfUserAuth(ctx context.Context, tenantID, userID string) (*platform.AuthContext, error)
}

// resolveOnBehalfOf validates an x-godarai-on-behalf-of request and returns the
// merged auth context. Only gateway routes accept the header, only VATs with the
// capability may send it, and the member must be an active user of the key's tenant.
func (m *AuthMiddleware) resolveOnBehalfOf(c *fiber.Ctx, key *platform.AuthContext, userID string) (*platform.AuthContext, *authRejection) {
	if !strings.HasPrefix(c.Path(), onBehalfOfRoutePrefix) || key == nil || !key.OnBehalfOfEnabled || m.onBehalfOf == nil {
		return nil, rejectAuth(fiber.StatusForbidden, "This key is not allowed to act on behalf of users", onBehalfOfNotAllowedType)
	}

	if _, err := uuid.Parse(userID); err != nil {
		return nil, rejectAuth(fiber.StatusForbidden, "The on-behalf-of user was not found", onBehalfOfUserNotFoundType)
	}

	user, err := m.onBehalfOf.ResolveOnBehalfOfUserAuth(c.Context(), key.TenantID, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) || errors.Is(err, pgx.ErrNoRows) {
			return nil, rejectAuth(fiber.StatusForbidden, "The on-behalf-of user was not found", onBehalfOfUserNotFoundType)
		}

		m.logger.Error("on-behalf-of user lookup failed", zap.Error(err), zap.String("tenant_id", key.TenantID))

		return nil, rejectAuth(fiber.StatusInternalServerError, "Authentication lookup failed", "server_error")
	}

	if user == nil || !strings.EqualFold(user.TenantID, key.TenantID) || !strings.EqualFold(user.UserID, userID) {
		return nil, rejectAuth(fiber.StatusForbidden, "The on-behalf-of user was not found", onBehalfOfUserNotFoundType)
	}

	return mergeOnBehalfOf(key, user), nil
}

// mergeOnBehalfOf keeps the service key's identity, lifecycle, role, and
// permissions, attributes usage to the member, and allows only the models both
// the key and the member may use.
func mergeOnBehalfOf(key, user *platform.AuthContext) *platform.AuthContext {
	out := *key
	out.UserID = user.UserID
	out.TeamID = user.TeamID
	out.OrgUnitID = user.OrgUnitID
	out.OrganizationID = user.OrganizationID
	out.VirtualUserID = ""
	out.OnBehalfOfKeyID = key.KeyID
	applyModelIntersection(&out, key, user)

	return &out
}

// applyModelIntersection sets the model policy of out to the intersection of
// the key and member policies. An explicit intersection is always enforced so
// that an empty result denies every model instead of allowing all of them.
func applyModelIntersection(out, key, user *platform.AuthContext) {
	switch {
	case modelPolicyUnrestricted(key) && modelPolicyUnrestricted(user):
		return
	case modelPolicyUnrestricted(key):
		copyModelPolicy(out, user)

		return
	case modelPolicyUnrestricted(user):
		return
	}

	out.EnforceModelAllowlist = true
	out.AllowedModels = make([]string, 0, len(key.AllowedModels))
	out.AllowedModelTypes = map[string][]string{}
	out.AllowedModelPricingMode = map[string]string{}
	out.AllowedModelPricingResolved = map[string]bool{}

	for _, model := range key.AllowedModels {
		if !modelPolicyAllows(user, model) {
			continue
		}

		out.AllowedModels = append(out.AllowedModels, model)
		copyModelEntry(out, key, user, model)
	}
}

func copyModelEntry(out, key, user *platform.AuthContext, model string) {
	if types, ok := key.AllowedModelTypes[model]; ok {
		out.AllowedModelTypes[model] = append([]string(nil), types...)
	} else if types, ok := user.AllowedModelTypes[model]; ok {
		out.AllowedModelTypes[model] = append([]string(nil), types...)
	}

	if mode, ok := key.AllowedModelPricingMode[model]; ok {
		out.AllowedModelPricingMode[model] = mode
	} else if mode, ok := user.AllowedModelPricingMode[model]; ok {
		out.AllowedModelPricingMode[model] = mode
	}

	if resolved, ok := key.AllowedModelPricingResolved[model]; ok {
		out.AllowedModelPricingResolved[model] = resolved
	} else if resolved, ok := user.AllowedModelPricingResolved[model]; ok {
		out.AllowedModelPricingResolved[model] = resolved
	}
}

func copyModelPolicy(out, from *platform.AuthContext) {
	out.EnforceModelAllowlist = from.EnforceModelAllowlist
	out.AllowedModels = append([]string(nil), from.AllowedModels...)
	out.AllowedModelTypes = make(map[string][]string, len(from.AllowedModelTypes))

	for model, types := range from.AllowedModelTypes {
		out.AllowedModelTypes[model] = append([]string(nil), types...)
	}

	out.AllowedModelPricingMode = make(map[string]string, len(from.AllowedModelPricingMode))
	for model, mode := range from.AllowedModelPricingMode {
		out.AllowedModelPricingMode[model] = mode
	}

	out.AllowedModelPricingResolved = make(map[string]bool, len(from.AllowedModelPricingResolved))
	for model, resolved := range from.AllowedModelPricingResolved {
		out.AllowedModelPricingResolved[model] = resolved
	}
}

// modelPolicyUnrestricted mirrors the gateway allowlist check: a wildcard, or
// an empty list that is not enforced, allows every model.
func modelPolicyUnrestricted(auth *platform.AuthContext) bool {
	if containsModelName(auth.AllowedModels, "*") {
		return true
	}

	return !auth.EnforceModelAllowlist && len(auth.AllowedModels) == 0
}

func modelPolicyAllows(auth *platform.AuthContext, model string) bool {
	return modelPolicyUnrestricted(auth) || containsModelName(auth.AllowedModels, model)
}

// containsModelName matches exactly, like the gateway allowlist check, so the
// intersection is never more permissive than either side.
func containsModelName(models []string, model string) bool {
	for _, candidate := range models {
		if candidate == model {
			return true
		}
	}

	return false
}
