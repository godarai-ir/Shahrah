package proxy

import (
	"context"
	"strings"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
)

type deploymentLister interface {
	ListDeployments(ctx context.Context, tenantID string) ([]platform.Deployment, error)
}

// ModelsHandler serves OpenAI-compatible model listing.
type ModelsHandler struct {
	models deploymentLister
}

func NewModelsHandler(models deploymentLister) *ModelsHandler {
	return &ModelsHandler{models: models}
}

func (h *ModelsHandler) Handle(c *fiber.Ctx) error {
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
		model := byAlias[id]
		model.ID = id
		model.Object = "model"
		model.OwnedBy = "shahrah"
		if model.Provider == "" {
			model.Provider = dep.Provider
		}
		model.Aliases = appendUniqueModelAlias(model.Aliases, dep.ModelName)
		model.Aliases = appendUniqueModelAlias(model.Aliases, dep.Model)
		byAlias[id] = model
	}

	data := make([]contractsopenai.Model, 0, len(byAlias))
	for _, item := range byAlias {
		data = append(data, contractsopenai.Model{
			ID:       item.ID,
			Object:   "model",
			OwnedBy:  item.OwnedBy,
			Provider: item.Provider,
			Aliases:  item.Aliases,
		})
	}

	return c.JSON(contractsopenai.ModelListResponse{Object: "list", Data: data})
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
