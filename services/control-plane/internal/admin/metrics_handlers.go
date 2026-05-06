package admin

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
)

type metricsQueryInput struct {
	StartTS           string                        `json:"startTs"`
	EndTS             string                        `json:"endTs"`
	Datasource        string                        `json:"datasource"`
	Type              string                        `json:"type"`
	Aggregations      []platform.MetricsAggregation `json:"aggregations"`
	GroupBy           []string                      `json:"groupBy"`
	Filters           []platform.MetricsFilter      `json:"filters"`
	IntervalInSeconds int                           `json:"intervalInSeconds"`
}

func (h *Handlers) QueryModelMetrics(c *fiber.Ctx) error {
	if err := requireTeamAdminOrHigher(c); err != nil {
		return err
	}
	if h.metrics == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "metrics store not configured")
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	var raw metricsQueryInput
	if err := c.BodyParser(&raw); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	startTS, err := time.Parse(time.RFC3339, strings.TrimSpace(raw.StartTS))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "startTs must be RFC3339")
	}
	endTS, err := time.Parse(time.RFC3339, strings.TrimSpace(raw.EndTS))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "endTs must be RFC3339")
	}
	if !startTS.Before(endTS) {
		return fiber.NewError(fiber.StatusBadRequest, "startTs must be before endTs")
	}
	if strings.TrimSpace(raw.Datasource) != "modelMetrics" {
		return fiber.NewError(fiber.StatusBadRequest, "datasource must be modelMetrics")
	}
	queryType := strings.TrimSpace(strings.ToLower(raw.Type))
	if queryType != "distribution" && queryType != "timeseries" {
		return fiber.NewError(fiber.StatusBadRequest, "type must be distribution or timeseries")
	}
	req := platform.MetricsQueryRequest{
		StartTS:           startTS,
		EndTS:             endTS,
		Datasource:        "modelMetrics",
		Type:              queryType,
		Aggregations:      raw.Aggregations,
		GroupBy:           raw.GroupBy,
		Filters:           raw.Filters,
		IntervalInSeconds: raw.IntervalInSeconds,
	}

	// Parse generic values from JSON for strict downstream parser behavior.
	normalizedFilters := make([]platform.MetricsFilter, 0, len(req.Filters))
	for _, f := range req.Filters {
		b, _ := json.Marshal(f.Value)
		var generic interface{}
		_ = json.Unmarshal(b, &generic)
		f.Value = generic
		normalizedFilters = append(normalizedFilters, f)
	}
	req.Filters = normalizedFilters

	items, err := h.metrics.QueryModelMetrics(c.Context(), currentTenant.ID, req)
	if err != nil {
		if errors.Is(err, storage.ErrMetricsStoreUnavailable) {
			return fiber.NewError(fiber.StatusNotImplemented, "metrics store not configured")
		}
		return err
	}

	return c.JSON(fiber.Map{"data": items})
}
