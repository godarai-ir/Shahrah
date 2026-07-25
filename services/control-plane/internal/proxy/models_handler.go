// Package proxy hosts OpenAI-compatible request handlers for the control plane.
package proxy

import (
	"context"
	"strings"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
)

type deploymentLister interface {
	ListDeployments(ctx context.Context, tenantID string) ([]platform.Deployment, error)
	ListVirtualModels(ctx context.Context, tenantID string) ([]platform.VirtualModel, error)
}

// ModelsHandler serves OpenAI-compatible model listing.
type ModelsHandler struct {
	models deploymentLister
}

// NewModelsHandler creates a new models handler.
func NewModelsHandler(models deploymentLister) *ModelsHandler {
	return &ModelsHandler{models: models}
}

// Handle serves the model list endpoint.
func (h *ModelsHandler) Handle(c *fiber.Ctx) error {
	authInfo := auth.GetAuthInfo(c)
	if authInfo == nil {
		return writeAPIError(c, fiber.StatusUnauthorized, "authentication required", "authentication_error")
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return writeAPIError(c, fiber.StatusBadRequest, "tenant context missing", "invalid_request_error")
	}

	deployments, err := h.models.ListDeployments(c.Context(), currentTenant.ID)
	if err != nil {
		return err
	}

	byAlias := make(map[string]contractsopenai.Model, len(deployments))
	for _, dep := range deployments {
		id := dep.ModelName
		if id == "" {
			id = dep.ID
		}

		if !isModelAllowed(authInfo, id) {
			continue
		}

		model := byAlias[id]
		model.ID = id
		model.Object = "model"

		model.OwnedBy = "shahrah"
		if model.Provider == "" {
			model.Provider = dep.Provider
		}

		model.Aliases = appendUniqueModelAlias(model.Aliases, dep.ModelName)
		model.Aliases = appendUniqueModelAlias(model.Aliases, dep.Model)

		modes := authInfo.AllowedModelTypes[id]
		if len(modes) == 0 {
			modes = supportedModesFromDeploymentInfo(dep.ModelInfo)
		}

		for _, mode := range modes {
			model.SupportedModes = appendUniqueModelAlias(model.SupportedModes, mode)
		}

		byAlias[id] = model
	}

	virtualModels, err := h.models.ListVirtualModels(c.Context(), currentTenant.ID)
	if err != nil {
		return err
	}

	for _, vm := range virtualModels {
		if !vm.Enabled || !isModelAllowed(authInfo, vm.Alias) {
			continue
		}

		model := byAlias[vm.Alias]
		model.ID = vm.Alias
		model.Object = "model"
		model.OwnedBy = "shahrah"
		model.Provider = "virtual"
		model.Aliases = appendUniqueModelAlias(model.Aliases, vm.Alias)

		modes := authInfo.AllowedModelTypes[vm.Alias]
		if len(modes) == 0 {
			modes = vm.ExposedModes
		}

		for _, mode := range modes {
			model.SupportedModes = appendUniqueModelAlias(model.SupportedModes, mode)
		}

		byAlias[vm.Alias] = model
	}

	data := make([]contractsopenai.Model, 0, len(byAlias))
	for _, item := range byAlias {
		data = append(data, contractsopenai.Model{
			ID:             item.ID,
			Object:         "model",
			OwnedBy:        item.OwnedBy,
			Provider:       item.Provider,
			Aliases:        item.Aliases,
			SupportedModes: item.SupportedModes,
		})
	}

	return c.JSON(contractsopenai.ModelListResponse{Object: "list", Data: data})
}

func supportedModesFromDeploymentInfo(info map[string]interface{}) []string {
	if len(info) == 0 {
		return nil
	}

	switch modes := info["supported_modes"].(type) {
	case []string:
		return append([]string(nil), modes...)
	case []interface{}:
		out := make([]string, 0, len(modes))
		for _, mode := range modes {
			if text, ok := mode.(string); ok && strings.TrimSpace(text) != "" {
				out = append(out, text)
			}
		}

		return out
	default:
		return nil
	}
}

func appendUniqueModelAlias(values []string, candidate string) []string {
	candidate = strings.TrimSpace(candidate)
	if candidate == "" {
		return values
	}

	for _, existing := range values {
		if existing == candidate {
			return values
		}
	}

	return append(values, candidate)
}
