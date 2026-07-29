package proxy

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/gofiber/fiber/v2"
)

func anthropicOperationFromRequest(c *fiber.Ctx) (platform.ProviderOperation, platform.ProviderOperationInput, error) {
	input := platform.ProviderOperationInput{
		ResourceID: strings.TrimSpace(c.Params("id")),
		BeforeID:   strings.TrimSpace(c.Query("before_id")),
		AfterID:    strings.TrimSpace(c.Query("after_id")),
		Limit:      c.QueryInt("limit", 0),
	}
	path := strings.TrimSuffix(c.Path(), "/")
	switch {
	case c.Method() == fiber.MethodPost && path == "/v1/messages":
		return platform.ProviderOperationAnthropicMessages, input, nil
	case c.Method() == fiber.MethodPost && path == "/v1/messages/count_tokens":
		return platform.ProviderOperationAnthropicCountTokens, input, nil
	case c.Method() == fiber.MethodPost && path == "/v1/messages/batches":
		return platform.ProviderOperationAnthropicBatchesCreate, input, nil
	case c.Method() == fiber.MethodGet && path == "/v1/messages/batches":
		return platform.ProviderOperationAnthropicBatchesList, input, nil
	case c.Method() == fiber.MethodGet && strings.HasPrefix(path, "/v1/messages/batches/") && strings.HasSuffix(path, "/results"):
		return platform.ProviderOperationAnthropicBatchesResults, input, nil
	case c.Method() == fiber.MethodPost && strings.HasPrefix(path, "/v1/messages/batches/") && strings.HasSuffix(path, "/cancel"):
		return platform.ProviderOperationAnthropicBatchesCancel, input, nil
	case c.Method() == fiber.MethodGet && strings.HasPrefix(path, "/v1/messages/batches/"):
		return platform.ProviderOperationAnthropicBatchesGet, input, nil
	case c.Method() == fiber.MethodPost && path == "/v1/files":
		return platform.ProviderOperationAnthropicFilesCreate, input, nil
	case c.Method() == fiber.MethodGet && path == "/v1/files":
		return platform.ProviderOperationAnthropicFilesList, input, nil
	case c.Method() == fiber.MethodGet && strings.HasPrefix(path, "/v1/files/") && strings.HasSuffix(path, "/content"):
		return platform.ProviderOperationAnthropicFilesContent, input, nil
	case c.Method() == fiber.MethodDelete && strings.HasPrefix(path, "/v1/files/"):
		return platform.ProviderOperationAnthropicFilesDelete, input, nil
	case c.Method() == fiber.MethodGet && strings.HasPrefix(path, "/v1/files/"):
		return platform.ProviderOperationAnthropicFilesGet, input, nil
	case c.Method() == fiber.MethodGet && path == "/v1/models":
		return platform.ProviderOperationAnthropicModelsList, input, nil
	case c.Method() == fiber.MethodGet && strings.HasPrefix(path, "/v1/models/"):
		return platform.ProviderOperationAnthropicModelsGet, input, nil
	default:
		return "", input, fmt.Errorf("unsupported Anthropic operation")
	}
}

func anthropicRequestModels(operation platform.ProviderOperation, input platform.ProviderOperationInput, body []byte) ([]string, bool, error) {
	switch operation {
	case platform.ProviderOperationAnthropicMessages, platform.ProviderOperationAnthropicCountTokens:
		var request struct {
			Model  string `json:"model"`
			Stream bool   `json:"stream"`
		}
		if err := json.Unmarshal(body, &request); err != nil {
			return nil, false, fmt.Errorf("Invalid request body")
		}
		if strings.TrimSpace(request.Model) == "" {
			return nil, false, fmt.Errorf("model is required")
		}
		return []string{request.Model}, request.Stream && operation == platform.ProviderOperationAnthropicMessages, nil
	case platform.ProviderOperationAnthropicModelsGet:
		if input.ResourceID == "" {
			return nil, false, fmt.Errorf("model id is required")
		}
		return []string{input.ResourceID}, false, nil
	case platform.ProviderOperationAnthropicBatchesCreate:
		var request struct {
			Requests []struct {
				Params struct {
					Model string `json:"model"`
				} `json:"params"`
			} `json:"requests"`
		}
		if err := json.Unmarshal(body, &request); err != nil {
			return nil, false, fmt.Errorf("Invalid request body")
		}
		if len(request.Requests) == 0 {
			return nil, false, fmt.Errorf("requests must not be empty")
		}
		models := make([]string, 0, len(request.Requests))
		for _, item := range request.Requests {
			if strings.TrimSpace(item.Params.Model) == "" {
				return nil, false, fmt.Errorf("every batch item requires a model")
			}
			models = append(models, item.Params.Model)
		}
		return models, false, nil
	default:
		return nil, false, nil
	}
}

func replaceAnthropicNativeModels(operation platform.ProviderOperation, body []byte, official []string) ([]byte, error) {
	var object map[string]interface{}
	if err := json.Unmarshal(body, &object); err != nil {
		return nil, err
	}
	switch operation {
	case platform.ProviderOperationAnthropicMessages, platform.ProviderOperationAnthropicCountTokens:
		object["model"] = official[0]
	case platform.ProviderOperationAnthropicBatchesCreate:
		requests, ok := object["requests"].([]interface{})
		if !ok || len(requests) != len(official) {
			return nil, fmt.Errorf("invalid batch requests")
		}
		for index, raw := range requests {
			item, ok := raw.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid batch request at index %d", index)
			}
			params, ok := item["params"].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid batch params at index %d", index)
			}
			params["model"] = official[index]
		}
	}
	return json.Marshal(object)
}

func anthropicMaximumOutputTokens(body []byte, operation platform.ProviderOperation) int {
	if operation == platform.ProviderOperationAnthropicMessages {
		var request struct {
			MaxTokens int `json:"max_tokens"`
		}
		_ = json.Unmarshal(body, &request)
		return request.MaxTokens
	}
	if operation == platform.ProviderOperationAnthropicBatchesCreate {
		var request struct {
			Requests []struct {
				Params struct {
					MaxTokens int `json:"max_tokens"`
				} `json:"params"`
			} `json:"requests"`
		}
		_ = json.Unmarshal(body, &request)
		total := 0
		for _, item := range request.Requests {
			total += item.Params.MaxTokens
		}
		return total
	}
	return 0
}

type anthropicNativeToolPricingInput struct {
	Type string `json:"type"`
}

type anthropicNativeMessagePricingInput struct {
	Tools []anthropicNativeToolPricingInput `json:"tools"`
}

func anthropicCodeExecutionPricingPolicy(
	body []byte,
	operation platform.ProviderOperation,
) (map[string]string, error) {
	var requests []anthropicNativeMessagePricingInput
	switch operation {
	case platform.ProviderOperationAnthropicMessages:
		var request anthropicNativeMessagePricingInput
		if err := json.Unmarshal(body, &request); err != nil {
			return nil, fmt.Errorf("Invalid request body")
		}
		requests = []anthropicNativeMessagePricingInput{request}
	case platform.ProviderOperationAnthropicBatchesCreate:
		var batch struct {
			Requests []struct {
				Params anthropicNativeMessagePricingInput `json:"params"`
			} `json:"requests"`
		}
		if err := json.Unmarshal(body, &batch); err != nil {
			return nil, fmt.Errorf("Invalid request body")
		}
		requests = make([]anthropicNativeMessagePricingInput, 0, len(batch.Requests))
		for _, item := range batch.Requests {
			requests = append(requests, item.Params)
		}
	default:
		return nil, nil
	}

	freeCodeExecutionPossible := false
	for index, request := range requests {
		hasCodeExecution := false
		hasCurrentWebTool := false
		for _, tool := range request.Tools {
			toolType := strings.ToLower(strings.TrimSpace(tool.Type))
			if strings.HasPrefix(toolType, "code_execution_") {
				hasCodeExecution = true
			}
			if anthropicVersionedToolAtLeast(
				toolType,
				"web_search",
				"20260209",
			) || anthropicVersionedToolAtLeast(
				toolType,
				"web_fetch",
				"20260209",
			) {
				hasCurrentWebTool = true
			}
		}
		if hasCodeExecution && !hasCurrentWebTool {
			prefix := ""
			if operation == platform.ProviderOperationAnthropicBatchesCreate {
				prefix = fmt.Sprintf("batch item %d: ", index)
			}
			return nil, fmt.Errorf(
				"%sbillable Anthropic code execution is unavailable; "+
					"include web_search_20260209 or later, or "+
					"web_fetch_20260209 or later",
				prefix,
			)
		}
		freeCodeExecutionPossible = freeCodeExecutionPossible ||
			hasCurrentWebTool
	}
	if !freeCodeExecutionPossible {
		return nil, nil
	}
	return map[string]string{
		platform.AnthropicCodeExecutionPricingCondition: platform.AnthropicCodeExecutionFreeWithCurrentWebTools,
	}, nil
}

func anthropicVersionedToolAtLeast(
	toolType, name, minimumVersion string,
) bool {
	prefix := name + "_"
	if !strings.HasPrefix(toolType, prefix) {
		return false
	}
	version := strings.TrimPrefix(toolType, prefix)
	if len(version) != len(minimumVersion) {
		return false
	}
	for _, value := range version {
		if value < '0' || value > '9' {
			return false
		}
	}
	return version >= minimumVersion
}
