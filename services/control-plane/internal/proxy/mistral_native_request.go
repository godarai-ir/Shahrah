package proxy

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/gofiber/fiber/v2"
)

const (
	mistralProviderName             = "mistral-ai"
	mistralFileResourceKind         = "mistral_file"
	mistralBatchResourceKind        = "mistral_batch"
	mistralAgentResourceKind        = "mistral_agent"
	mistralConversationResourceKind = "mistral_conversation"
)

func mistralOperationFromRequest(c *fiber.Ctx) (platform.ProviderOperation, platform.ProviderOperationInput, error) {
	page := c.QueryInt("page", 0)
	pageSize := c.QueryInt("page_size", 0)
	createdByMe := strings.TrimSpace(c.Query("created_by_me"))

	resourceID := strings.TrimSpace(c.Params("id"))
	if resourceID == "" {
		resourceID = strings.TrimSpace(c.Params("file_id"))
	}

	input := platform.ProviderOperationInput{
		ResourceID: resourceID,
		OrderBy:    strings.TrimSpace(c.Query("order_by")),
		Status:     strings.TrimSpace(c.Query("status")),
		Model:      strings.TrimSpace(c.Query("model")),
	}
	if page > 0 {
		input.Page = &page
	}

	if pageSize > 0 {
		input.PageSize = &pageSize
	}

	if createdByMe != "" {
		value := strings.EqualFold(createdByMe, "true")
		if !value && !strings.EqualFold(createdByMe, "false") {
			return "", input, fmt.Errorf("created_by_me must be true or false")
		}

		input.CreatedByMe = &value
	}

	method := c.Method()

	path := strings.TrimSuffix(c.Path(), "/")
	for _, resolve := range []func(string, string) (platform.ProviderOperation, bool){
		mistralFileOperation,
		mistralBatchOperation,
		mistralAgentOperation,
		mistralConversationOperation,
	} {
		if operation, ok := resolve(method, path); ok {
			return operation, input, nil
		}
	}

	return "", input, fmt.Errorf("unsupported Mistral operation")
}

func mistralFileOperation(method, path string) (platform.ProviderOperation, bool) {
	switch {
	case method == fiber.MethodPost && path == "/v1/files":
		return platform.ProviderOperationMistralFilesCreate, true
	case method == fiber.MethodGet && path == "/v1/files":
		return platform.ProviderOperationMistralFilesList, true
	case method == fiber.MethodGet && strings.HasSuffix(path, "/content"):
		return platform.ProviderOperationMistralFilesContent, true
	case method == fiber.MethodGet && strings.HasSuffix(path, "/url"):
		return platform.ProviderOperationMistralFilesURL, true
	case method == fiber.MethodDelete && strings.HasPrefix(path, "/v1/files/"):
		return platform.ProviderOperationMistralFilesDelete, true
	case method == fiber.MethodGet && strings.HasPrefix(path, "/v1/files/"):
		return platform.ProviderOperationMistralFilesGet, true
	default:
		return "", false
	}
}

func mistralBatchOperation(method, path string) (platform.ProviderOperation, bool) {
	switch {
	case method == fiber.MethodPost && path == "/v1/batch/jobs":
		return platform.ProviderOperationMistralBatchCreate, true
	case method == fiber.MethodGet && path == "/v1/batch/jobs":
		return platform.ProviderOperationMistralBatchList, true
	case method == fiber.MethodPost && strings.HasSuffix(path, "/cancel"):
		return platform.ProviderOperationMistralBatchCancel, true
	case method == fiber.MethodDelete && strings.HasPrefix(path, "/v1/batch/jobs/"):
		return platform.ProviderOperationMistralBatchDelete, true
	case method == fiber.MethodGet && strings.HasPrefix(path, "/v1/batch/jobs/"):
		return platform.ProviderOperationMistralBatchGet, true
	default:
		return "", false
	}
}

func mistralAgentOperation(method, path string) (platform.ProviderOperation, bool) {
	switch {
	case method == fiber.MethodPost && path == "/v1/agents":
		return platform.ProviderOperationMistralAgentsCreate, true
	case method == fiber.MethodGet && path == "/v1/agents":
		return platform.ProviderOperationMistralAgentsList, true
	case method == fiber.MethodGet && path == "/v1/agents/pages":
		return platform.ProviderOperationMistralAgentsPages, true
	case method == fiber.MethodPatch && strings.HasPrefix(path, "/v1/agents/"):
		return platform.ProviderOperationMistralAgentsUpdate, true
	case method == fiber.MethodDelete && strings.HasPrefix(path, "/v1/agents/"):
		return platform.ProviderOperationMistralAgentsDelete, true
	case method == fiber.MethodGet && strings.HasPrefix(path, "/v1/agents/"):
		return platform.ProviderOperationMistralAgentsGet, true
	default:
		return "", false
	}
}

func mistralConversationOperation(method, path string) (platform.ProviderOperation, bool) {
	switch {
	case method == fiber.MethodPost && path == "/v1/conversations":
		return platform.ProviderOperationMistralConversationsCreate, true
	case method == fiber.MethodGet && path == "/v1/conversations":
		return platform.ProviderOperationMistralConversationsList, true
	case method == fiber.MethodPost && strings.HasSuffix(path, "/restart"):
		return platform.ProviderOperationMistralConversationsRestart, true
	case method == fiber.MethodGet && strings.HasSuffix(path, "/history"):
		return platform.ProviderOperationMistralConversationsHistory, true
	case method == fiber.MethodGet && strings.HasSuffix(path, "/messages"):
		return platform.ProviderOperationMistralConversationsMessages, true
	case method == fiber.MethodPost && strings.HasPrefix(path, "/v1/conversations/"):
		return platform.ProviderOperationMistralConversationsAppend, true
	case method == fiber.MethodDelete && strings.HasPrefix(path, "/v1/conversations/"):
		return platform.ProviderOperationMistralConversationsDelete, true
	case method == fiber.MethodGet && strings.HasPrefix(path, "/v1/conversations/"):
		return platform.ProviderOperationMistralConversationsGet, true
	default:
		return "", false
	}
}

func mistralRequestModel(operation platform.ProviderOperation, body []byte) (string, bool, error) {
	if len(body) == 0 {
		return "", false, nil
	}

	var request map[string]interface{}
	if err := json.Unmarshal(body, &request); err != nil {
		return "", false, fmt.Errorf("invalid JSON request body")
	}

	model := normalizeModelIdentifier(stringFromMap(request, "model"))
	if model == "" && operation == platform.ProviderOperationMistralBatchCreate {
		model = mistralNestedRequestModel(request["requests"])
	}

	stream, ok := request["stream"].(bool)
	if !ok {
		stream = false
	}

	switch operation {
	case platform.ProviderOperationMistralAgentsCreate,
		platform.ProviderOperationMistralConversationsCreate,
		platform.ProviderOperationMistralBatchCreate:
		if model == "" && strings.TrimSpace(stringFromMap(request, "agent_id")) == "" {
			return "", false, fmt.Errorf("model or agent_id is required")
		}
	}

	return model, stream, nil
}

func mistralNestedRequestModel(value interface{}) string {
	requests, ok := value.([]interface{})
	if !ok {
		return ""
	}

	model := ""

	for _, raw := range requests {
		item, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}

		candidate := normalizeModelIdentifier(stringFromMap(item, "model"))
		if candidate == "" {
			if body, ok := item["body"].(map[string]interface{}); ok {
				candidate = normalizeModelIdentifier(stringFromMap(body, "model"))
			}
		}

		if candidate == "" {
			continue
		}

		if model != "" && !strings.EqualFold(model, candidate) {
			return ""
		}

		model = candidate
	}

	return model
}

func mistralOperationResourceKind(operation platform.ProviderOperation) string {
	switch operation {
	case platform.ProviderOperationMistralFilesCreate,
		platform.ProviderOperationMistralFilesList,
		platform.ProviderOperationMistralFilesGet,
		platform.ProviderOperationMistralFilesDelete,
		platform.ProviderOperationMistralFilesContent,
		platform.ProviderOperationMistralFilesURL:
		return mistralFileResourceKind
	case platform.ProviderOperationMistralBatchCreate,
		platform.ProviderOperationMistralBatchList,
		platform.ProviderOperationMistralBatchGet,
		platform.ProviderOperationMistralBatchCancel,
		platform.ProviderOperationMistralBatchDelete:
		return mistralBatchResourceKind
	case platform.ProviderOperationMistralAgentsCreate,
		platform.ProviderOperationMistralAgentsList,
		platform.ProviderOperationMistralAgentsPages,
		platform.ProviderOperationMistralAgentsGet,
		platform.ProviderOperationMistralAgentsUpdate,
		platform.ProviderOperationMistralAgentsDelete:
		return mistralAgentResourceKind
	case platform.ProviderOperationMistralConversationsCreate,
		platform.ProviderOperationMistralConversationsList,
		platform.ProviderOperationMistralConversationsGet,
		platform.ProviderOperationMistralConversationsAppend,
		platform.ProviderOperationMistralConversationsDelete,
		platform.ProviderOperationMistralConversationsHistory,
		platform.ProviderOperationMistralConversationsMessages,
		platform.ProviderOperationMistralConversationsRestart:
		return mistralConversationResourceKind
	default:
		return ""
	}
}

func mistralOperationIsCreate(operation platform.ProviderOperation) bool {
	return operation == platform.ProviderOperationMistralFilesCreate ||
		operation == platform.ProviderOperationMistralBatchCreate ||
		operation == platform.ProviderOperationMistralAgentsCreate ||
		operation == platform.ProviderOperationMistralConversationsCreate
}

func mistralOperationIsList(operation platform.ProviderOperation) bool {
	return operation == platform.ProviderOperationMistralFilesList ||
		operation == platform.ProviderOperationMistralBatchList ||
		operation == platform.ProviderOperationMistralAgentsList ||
		operation == platform.ProviderOperationMistralAgentsPages ||
		operation == platform.ProviderOperationMistralConversationsList
}

func mistralOperationIsInference(operation platform.ProviderOperation) bool {
	return operation == platform.ProviderOperationMistralConversationsCreate ||
		operation == platform.ProviderOperationMistralConversationsAppend ||
		operation == platform.ProviderOperationMistralConversationsRestart
}
