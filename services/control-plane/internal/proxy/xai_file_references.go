package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Delpak/shahrah/services/control-plane/internal/providernative"
	"github.com/gofiber/fiber/v2"
)

var (
	errXAIFileRoutingUnavailable = errors.New("xAI file routing state is unavailable")
	errXAIFileNotOwned           = errors.New("xAI file reference was not found")
)

func rewriteXAIFileReferences(
	ctx context.Context,
	resources providerNativeResourceStore,
	tenantID string,
	providerAccountID string,
	body []byte,
) ([]byte, error) {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()

	var document interface{}
	if err := decoder.Decode(&document); err != nil {
		return nil, err
	}

	changed, err := rewriteXAIFileReferenceValue(ctx, resources, tenantID, providerAccountID, document, "")
	if err != nil {
		return nil, err
	}

	if !changed {
		return body, nil
	}

	return json.Marshal(document)
}

func rewriteXAIFileReferenceValue(
	ctx context.Context,
	resources providerNativeResourceStore,
	tenantID string,
	providerAccountID string,
	value interface{},
	parentKey string,
) (bool, error) {
	switch typed := value.(type) {
	case []interface{}:
		changed := false

		for _, item := range typed {
			itemChanged, err := rewriteXAIFileReferenceValue(ctx, resources, tenantID, providerAccountID, item, parentKey)
			if err != nil {
				return false, err
			}

			changed = changed || itemChanged
		}

		return changed, nil
	case map[string]interface{}:
		changed := false

		for key, child := range typed {
			if key == "file_id" {
				publicID, ok := child.(string)
				if !ok || strings.TrimSpace(publicID) == "" {
					return false, fmt.Errorf("%w: file_id must be a non-empty string", errXAIFileNotOwned)
				}

				if resources == nil {
					return false, errXAIFileRoutingUnavailable
				}

				mapping, err := resources.GetProviderNativeResourceMapping(ctx, tenantID, "xai", openAIFileResourceKind, strings.TrimSpace(publicID))
				if errors.Is(err, providernative.ErrResourceNotFound) {
					return false, errXAIFileNotOwned
				}

				if err != nil {
					return false, errXAIFileRoutingUnavailable
				}

				if !strings.EqualFold(strings.TrimSpace(mapping.ProviderAccountID), strings.TrimSpace(providerAccountID)) || strings.TrimSpace(mapping.UpstreamID) == "" {
					return false, errXAIFileNotOwned
				}

				typed[key] = mapping.UpstreamID
				changed = changed || mapping.UpstreamID != publicID

				continue
			}
			// Avoid interpreting example/default values inside function JSON schemas
			// or tenant metadata as provider resource references.
			if key == "schema" || key == "parameters" || key == "metadata" {
				continue
			}

			childChanged, err := rewriteXAIFileReferenceValue(ctx, resources, tenantID, providerAccountID, child, key)
			if err != nil {
				return false, err
			}

			changed = changed || childChanged
		}

		return changed, nil
	default:
		return false, nil
	}
}

func writeXAIFileReferenceError(c *fiber.Ctx, err error) error {
	if errors.Is(err, errXAIFileRoutingUnavailable) {
		return writeAPIError(c, fiber.StatusServiceUnavailable, errXAIFileRoutingUnavailable.Error(), "routing_error")
	}

	return writeAPIError(c, fiber.StatusNotFound, errXAIFileNotOwned.Error(), "invalid_request_error")
}
