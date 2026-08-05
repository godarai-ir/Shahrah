package proxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"regexp"
	"strings"

	"github.com/Delpak/shahrah/services/control-plane/internal/apiheaders"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/gofiber/fiber/v2"
)

const (
	playgroundObservabilitySource   = "tenant_admin_playground"
	maxPlaygroundObservabilityBytes = 8 * 1024
	maxPlaygroundTags               = 20
)

var (
	playgroundIDPattern  = regexp.MustCompile(`^[A-Za-z0-9_-]{1,100}$`)
	playgroundTagPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,64}$`)
)

type playgroundObservabilityContextKey struct{}

type playgroundObservabilityContext struct {
	Source       string                 `json:"source"`
	ClientRunID  string                 `json:"client_run_id,omitempty"`
	ComparisonID string                 `json:"comparison_id,omitempty"`
	ColumnID     string                 `json:"column_id,omitempty"`
	Tags         map[string]interface{} `json:"tags,omitempty"`
}

func parsePlaygroundObservability(c *fiber.Ctx) error {
	raw := apiheaders.GetPlaygroundObservability(c)
	if raw == "" || !auth.IsDelegatedRequest(c) {
		return nil
	}

	if len(raw) > maxPlaygroundObservabilityBytes {
		return errors.New("playground observability context is too large")
	}

	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.DisallowUnknownFields()

	var context playgroundObservabilityContext
	if err := decoder.Decode(&context); err != nil {
		return fmt.Errorf("invalid playground observability context: %w", err)
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("invalid playground observability context")
	}

	if context.Source != playgroundObservabilitySource {
		return errors.New("invalid playground observability source")
	}

	for name, value := range map[string]string{
		"client_run_id": context.ClientRunID,
		"comparison_id": context.ComparisonID,
		"column_id":     context.ColumnID,
	} {
		if value != "" && !playgroundIDPattern.MatchString(value) {
			return fmt.Errorf("invalid playground observability %s", name)
		}
	}

	if len(context.Tags) > maxPlaygroundTags {
		return errors.New("playground observability contains too many tags")
	}

	for key, value := range context.Tags {
		normalizedKey := strings.ToLower(key)
		if !playgroundTagPattern.MatchString(key) || strings.Contains(normalizedKey, "authorization") || isReservedPlaygroundTag(normalizedKey) {
			return fmt.Errorf("playground observability tag %s is not allowed", key)
		}

		if !validPlaygroundTagValue(value) {
			return fmt.Errorf("playground observability tag %s has an invalid value", key)
		}
	}

	c.Locals(playgroundObservabilityContextKey{}, context)

	return nil
}

func isReservedPlaygroundTag(key string) bool {
	switch key {
	case "source", "client_run_id", "comparison_id", "column_id", "__proto__", "prototype", "constructor":
		return true
	default:
		return false
	}
}

func validPlaygroundTagValue(value interface{}) bool {
	switch typed := value.(type) {
	case string:
		return len(typed) <= 256
	case float64:
		return !math.IsNaN(typed) && !math.IsInf(typed, 0)
	case bool:
		return true
	default:
		return false
	}
}

func enrichPlaygroundObservability(c *fiber.Ctx, metadata map[string]interface{}) map[string]interface{} {
	if c == nil {
		return metadata
	}

	context, ok := c.Locals(playgroundObservabilityContextKey{}).(playgroundObservabilityContext)
	if !ok {
		return metadata
	}

	out := make(map[string]interface{}, len(metadata)+5)
	for key, value := range metadata {
		out[key] = value
	}

	out["playground_source"] = context.Source
	if context.ClientRunID != "" {
		out["playground_client_run_id"] = context.ClientRunID
	}

	if context.ComparisonID != "" {
		out["playground_comparison_id"] = context.ComparisonID
	}

	if context.ColumnID != "" {
		out["playground_column_id"] = context.ColumnID
	}

	if len(context.Tags) > 0 {
		out["playground_tags"] = context.Tags
	}

	return out
}

func copyPlaygroundObservability(target map[string]interface{}, source map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(target)+5)
	for key, value := range target {
		out[key] = value
	}

	for _, key := range []string{
		"playground_source",
		"playground_client_run_id",
		"playground_comparison_id",
		"playground_column_id",
		"playground_tags",
	} {
		if value, ok := source[key]; ok {
			out[key] = value
		}
	}

	return out
}
