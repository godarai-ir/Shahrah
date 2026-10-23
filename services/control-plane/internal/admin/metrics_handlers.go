package admin

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability"
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
	if err := requireTenantAdminOrHigher(c); err != nil {
		return err
	}

	if h.metrics == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "metrics store not configured")
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	req, err := parseMetricsQueryRequest(c)
	if err != nil {
		return err
	}

	items, err := h.metrics.QueryModelMetrics(c.Context(), currentTenant.ID, req)
	if err != nil {
		if errors.Is(err, observability.ErrMetricsStoreUnavailable) {
			return fiber.NewError(fiber.StatusNotImplemented, "metrics store not configured")
		}

		return err
	}

	return c.JSON(fiber.Map{"data": items})
}

func parseMetricsQueryRequest(c *fiber.Ctx) (platform.MetricsQueryRequest, error) {
	var raw metricsQueryInput
	if err := c.BodyParser(&raw); err != nil {
		return platform.MetricsQueryRequest{}, fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	startTS, err := time.Parse(time.RFC3339, strings.TrimSpace(raw.StartTS))
	if err != nil {
		return platform.MetricsQueryRequest{}, fiber.NewError(fiber.StatusBadRequest, "startTs must be RFC3339")
	}

	endTS, err := time.Parse(time.RFC3339, strings.TrimSpace(raw.EndTS))
	if err != nil {
		return platform.MetricsQueryRequest{}, fiber.NewError(fiber.StatusBadRequest, "endTs must be RFC3339")
	}

	if !startTS.Before(endTS) {
		return platform.MetricsQueryRequest{}, fiber.NewError(fiber.StatusBadRequest, "startTs must be before endTs")
	}

	if strings.TrimSpace(raw.Datasource) != "modelMetrics" {
		return platform.MetricsQueryRequest{}, fiber.NewError(fiber.StatusBadRequest, "datasource must be modelMetrics")
	}

	queryType := strings.TrimSpace(strings.ToLower(raw.Type))
	if queryType != "distribution" && queryType != "timeseries" {
		return platform.MetricsQueryRequest{}, fiber.NewError(fiber.StatusBadRequest, "type must be distribution or timeseries")
	}

	filters, err := normalizeMetricsFilters(raw.Filters)
	if err != nil {
		return platform.MetricsQueryRequest{}, err
	}

	return platform.MetricsQueryRequest{
		StartTS:           startTS,
		EndTS:             endTS,
		Datasource:        "modelMetrics",
		Type:              queryType,
		Aggregations:      raw.Aggregations,
		GroupBy:           raw.GroupBy,
		Filters:           filters,
		IntervalInSeconds: raw.IntervalInSeconds,
	}, nil
}

func normalizeMetricsFilters(filters []platform.MetricsFilter) ([]platform.MetricsFilter, error) {
	normalized := make([]platform.MetricsFilter, 0, len(filters))
	for _, filter := range filters {
		value, err := normalizeMetricsFilterValue(filter.Value)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "invalid filter value")
		}

		filter.Value = value
		normalized = append(normalized, filter)
	}

	return normalized, nil
}

func normalizeMetricsFilterValue(value interface{}) (interface{}, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	var normalized interface{}
	if err := json.Unmarshal(encoded, &normalized); err != nil {
		return nil, err
	}

	return normalized, nil
}

func (h *Handlers) ListModelMetricsFacets(c *fiber.Ctx) error {
	if err := requireTenantAdminOrHigher(c); err != nil {
		return err
	}

	if h.metrics == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "metrics store not configured")
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	facets, err := h.metrics.ListModelMetricsFacets(c.Context(), currentTenant.ID)
	if err != nil {
		if errors.Is(err, observability.ErrMetricsStoreUnavailable) {
			return fiber.NewError(fiber.StatusNotImplemented, "metrics store not configured")
		}

		return err
	}

	return c.JSON(facets)
}
