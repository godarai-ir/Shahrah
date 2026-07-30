package server

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

// resolveOpenAIOperation is the only method/path compiler accepted by the
// OpenAI-compatible adapter. Control Plane sends an enum plus opaque IDs; it
// cannot make Silk Road call an arbitrary upstream path or URL.
func resolveOpenAIOperation(payload platform.TunnelRequest) (string, string, error) {
	if payload.Operation == "" {
		return resolveLegacyOpenAICompatibleOperation(payload.RequestMetadata)
	}

	input := payload.OperationInput
	id := func(value, name string) (string, error) {
		value = strings.TrimSpace(value)
		if value == "" {
			return "", fmt.Errorf("%s is required for %s", name, payload.Operation)
		}
		return url.PathEscape(value), nil
	}
	resource := func() (string, error) { return id(input.ResourceID, "resource_id") }
	parent := func() (string, error) { return id(input.ParentResourceID, "parent_resource_id") }
	switch payload.Operation {
	case platform.ProviderOperationChatCompletions:
		return http.MethodPost, "/chat/completions", nil
	case platform.ProviderOperationOpenAICompletions:
		return http.MethodPost, "/completions", nil
	case platform.ProviderOperationOpenAIResponsesCreate:
		return http.MethodPost, "/responses", nil
	case platform.ProviderOperationOpenAIResponsesGet,
		platform.ProviderOperationOpenAIResponsesDelete,
		platform.ProviderOperationOpenAIResponsesCancel,
		platform.ProviderOperationOpenAIResponsesInputItems:
		value, err := resource()
		if err != nil {
			return "", "", err
		}
		method, suffix := http.MethodGet, ""
		switch payload.Operation {
		case platform.ProviderOperationOpenAIResponsesDelete:
			method = http.MethodDelete
		case platform.ProviderOperationOpenAIResponsesCancel:
			method, suffix = http.MethodPost, "/cancel"
		case platform.ProviderOperationOpenAIResponsesInputItems:
			suffix = "/input_items"
		}
		return method, withListQuery("/responses/"+value+suffix, input), nil
	case platform.ProviderOperationOpenAIResponsesCompact:
		return http.MethodPost, "/responses/compact", nil
	case platform.ProviderOperationOpenAIResponsesInputTokens:
		return http.MethodPost, "/responses/input_tokens", nil
	case platform.ProviderOperationOpenAIConversationsCreate:
		return http.MethodPost, "/conversations", nil
	case platform.ProviderOperationOpenAIConversationsList:
		return http.MethodGet, withListQuery("/conversations", input), nil
	case platform.ProviderOperationOpenAIConversationsGet,
		platform.ProviderOperationOpenAIConversationsUpdate,
		platform.ProviderOperationOpenAIConversationsDelete:
		value, err := resource()
		if err != nil {
			return "", "", err
		}
		method := http.MethodGet
		if payload.Operation == platform.ProviderOperationOpenAIConversationsUpdate {
			method = http.MethodPost
		} else if payload.Operation == platform.ProviderOperationOpenAIConversationsDelete {
			method = http.MethodDelete
		}
		return method, "/conversations/" + value, nil
	case platform.ProviderOperationOpenAIConversationItemsCreate,
		platform.ProviderOperationOpenAIConversationItemsList:
		value, err := parent()
		if err != nil {
			return "", "", err
		}
		method := http.MethodPost
		if payload.Operation == platform.ProviderOperationOpenAIConversationItemsList {
			method = http.MethodGet
		}
		return method, withListQuery("/conversations/"+value+"/items", input), nil
	case platform.ProviderOperationOpenAIConversationItemsGet,
		platform.ProviderOperationOpenAIConversationItemsDelete:
		parentID, err := parent()
		if err != nil {
			return "", "", err
		}
		itemID, err := resource()
		if err != nil {
			return "", "", err
		}
		method := http.MethodGet
		if payload.Operation == platform.ProviderOperationOpenAIConversationItemsDelete {
			method = http.MethodDelete
		}
		return method, "/conversations/" + parentID + "/items/" + itemID, nil
	case platform.ProviderOperationOpenAIEmbeddings:
		return http.MethodPost, "/embeddings", nil
	case platform.ProviderOperationOpenAIModerations:
		return http.MethodPost, "/moderations", nil
	case platform.ProviderOperationOpenAIImagesGenerations:
		return http.MethodPost, "/images/generations", nil
	case platform.ProviderOperationOpenAIImagesEdits:
		return http.MethodPost, "/images/edits", nil
	case platform.ProviderOperationOpenAIImagesVariations:
		return http.MethodPost, "/images/variations", nil
	case platform.ProviderOperationOpenAIAudioSpeech:
		return http.MethodPost, "/audio/speech", nil
	case platform.ProviderOperationOpenAIAudioTranscriptions:
		return http.MethodPost, "/audio/transcriptions", nil
	case platform.ProviderOperationOpenAIAudioTranslations:
		return http.MethodPost, "/audio/translations", nil
	case platform.ProviderOperationOpenAIBatchesCreate:
		return http.MethodPost, "/batches", nil
	case platform.ProviderOperationOpenAIBatchesList:
		return http.MethodGet, withListQuery("/batches", input), nil
	case platform.ProviderOperationOpenAIBatchesGet,
		platform.ProviderOperationOpenAIBatchesCancel:
		value, err := resource()
		if err != nil {
			return "", "", err
		}
		if payload.Operation == platform.ProviderOperationOpenAIBatchesCancel {
			return http.MethodPost, "/batches/" + value + "/cancel", nil
		}
		return http.MethodGet, "/batches/" + value, nil
	case platform.ProviderOperationOpenAIFilesCreate:
		return http.MethodPost, "/files", nil
	case platform.ProviderOperationOpenAIFilesList:
		return http.MethodGet, withListQuery("/files", input), nil
	case platform.ProviderOperationOpenAIFilesGet,
		platform.ProviderOperationOpenAIFilesDelete,
		platform.ProviderOperationOpenAIFilesContent:
		value, err := resource()
		if err != nil {
			return "", "", err
		}
		method, suffix := http.MethodGet, ""
		if payload.Operation == platform.ProviderOperationOpenAIFilesDelete {
			method = http.MethodDelete
		} else if payload.Operation == platform.ProviderOperationOpenAIFilesContent {
			suffix = "/content"
		}
		return method, "/files/" + value + suffix, nil
	case platform.ProviderOperationOpenAIUploadsCreate:
		return http.MethodPost, "/uploads", nil
	case platform.ProviderOperationOpenAIUploadsAddPart,
		platform.ProviderOperationOpenAIUploadsComplete,
		platform.ProviderOperationOpenAIUploadsCancel:
		value, err := resource()
		if err != nil {
			return "", "", err
		}
		suffix := "/parts"
		if payload.Operation == platform.ProviderOperationOpenAIUploadsComplete {
			suffix = "/complete"
		} else if payload.Operation == platform.ProviderOperationOpenAIUploadsCancel {
			suffix = "/cancel"
		}
		return http.MethodPost, "/uploads/" + value + suffix, nil
	case platform.ProviderOperationOpenAIVectorStoresCreate:
		return http.MethodPost, "/vector_stores", nil
	case platform.ProviderOperationOpenAIVectorStoresList:
		return http.MethodGet, withListQuery("/vector_stores", input), nil
	case platform.ProviderOperationOpenAIVectorStoresGet,
		platform.ProviderOperationOpenAIVectorStoresUpdate,
		platform.ProviderOperationOpenAIVectorStoresDelete,
		platform.ProviderOperationOpenAIVectorStoresSearch:
		value, err := resource()
		if err != nil {
			return "", "", err
		}
		method, suffix := http.MethodGet, ""
		switch payload.Operation {
		case platform.ProviderOperationOpenAIVectorStoresUpdate:
			method = http.MethodPost
		case platform.ProviderOperationOpenAIVectorStoresDelete:
			method = http.MethodDelete
		case platform.ProviderOperationOpenAIVectorStoresSearch:
			method, suffix = http.MethodPost, "/search"
		}
		return method, "/vector_stores/" + value + suffix, nil
	case platform.ProviderOperationOpenAIVectorStoreFilesCreate,
		platform.ProviderOperationOpenAIVectorStoreFilesList:
		value, err := parent()
		if err != nil {
			return "", "", err
		}
		method := http.MethodPost
		if payload.Operation == platform.ProviderOperationOpenAIVectorStoreFilesList {
			method = http.MethodGet
		}
		return method, withListQuery("/vector_stores/"+value+"/files", input), nil
	case platform.ProviderOperationOpenAIVectorStoreFilesGet,
		platform.ProviderOperationOpenAIVectorStoreFilesUpdate,
		platform.ProviderOperationOpenAIVectorStoreFilesDelete,
		platform.ProviderOperationOpenAIVectorStoreFilesContent:
		storeID, err := parent()
		if err != nil {
			return "", "", err
		}
		fileID, err := resource()
		if err != nil {
			return "", "", err
		}
		method, suffix := http.MethodGet, ""
		switch payload.Operation {
		case platform.ProviderOperationOpenAIVectorStoreFilesUpdate:
			method = http.MethodPost
		case platform.ProviderOperationOpenAIVectorStoreFilesDelete:
			method = http.MethodDelete
		case platform.ProviderOperationOpenAIVectorStoreFilesContent:
			suffix = "/content"
		}
		return method, "/vector_stores/" + storeID + "/files/" + fileID + suffix, nil
	case platform.ProviderOperationOpenAIVectorStoreFileBatchesCreate,
		platform.ProviderOperationOpenAIVectorStoreFileBatchesList:
		value, err := parent()
		if err != nil {
			return "", "", err
		}
		method := http.MethodPost
		if payload.Operation == platform.ProviderOperationOpenAIVectorStoreFileBatchesList {
			method = http.MethodGet
		}
		return method, withListQuery("/vector_stores/"+value+"/file_batches", input), nil
	case platform.ProviderOperationOpenAIVectorStoreFileBatchesGet,
		platform.ProviderOperationOpenAIVectorStoreFileBatchesCancel:
		storeID, err := parent()
		if err != nil {
			return "", "", err
		}
		batchID, err := resource()
		if err != nil {
			return "", "", err
		}
		suffix := ""
		if payload.Operation == platform.ProviderOperationOpenAIVectorStoreFileBatchesCancel {
			suffix = "/cancel"
			return http.MethodPost, "/vector_stores/" + storeID + "/file_batches/" + batchID + suffix, nil
		}
		return http.MethodGet, "/vector_stores/" + storeID + "/file_batches/" + batchID, nil
	case platform.ProviderOperationOpenAIContainersCreate:
		return http.MethodPost, "/containers", nil
	case platform.ProviderOperationOpenAIContainersList:
		return http.MethodGet, withListQuery("/containers", input), nil
	case platform.ProviderOperationOpenAIContainersGet,
		platform.ProviderOperationOpenAIContainersDelete:
		value, err := resource()
		if err != nil {
			return "", "", err
		}
		if payload.Operation == platform.ProviderOperationOpenAIContainersDelete {
			return http.MethodDelete, "/containers/" + value, nil
		}
		return http.MethodGet, "/containers/" + value, nil
	case platform.ProviderOperationOpenAIContainerFilesCreate,
		platform.ProviderOperationOpenAIContainerFilesList:
		value, err := parent()
		if err != nil {
			return "", "", err
		}
		method := http.MethodPost
		if payload.Operation == platform.ProviderOperationOpenAIContainerFilesList {
			method = http.MethodGet
		}
		return method, withListQuery("/containers/"+value+"/files", input), nil
	case platform.ProviderOperationOpenAIContainerFilesGet,
		platform.ProviderOperationOpenAIContainerFilesDelete,
		platform.ProviderOperationOpenAIContainerFilesContent:
		containerID, err := parent()
		if err != nil {
			return "", "", err
		}
		fileID, err := resource()
		if err != nil {
			return "", "", err
		}
		method, suffix := http.MethodGet, ""
		if payload.Operation == platform.ProviderOperationOpenAIContainerFilesDelete {
			method = http.MethodDelete
		} else if payload.Operation == platform.ProviderOperationOpenAIContainerFilesContent {
			suffix = "/content"
		}
		return method, "/containers/" + containerID + "/files/" + fileID + suffix, nil
	case platform.ProviderOperationOpenAIVideosCreate:
		return http.MethodPost, "/videos", nil
	case platform.ProviderOperationOpenAIVideosList:
		return http.MethodGet, withListQuery("/videos", input), nil
	case platform.ProviderOperationOpenAIVideosGet,
		platform.ProviderOperationOpenAIVideosDelete,
		platform.ProviderOperationOpenAIVideosContent,
		platform.ProviderOperationOpenAIVideosRemix:
		value, err := resource()
		if err != nil {
			return "", "", err
		}
		method, suffix := http.MethodGet, ""
		switch payload.Operation {
		case platform.ProviderOperationOpenAIVideosDelete:
			method = http.MethodDelete
		case platform.ProviderOperationOpenAIVideosContent:
			suffix = "/content"
		case platform.ProviderOperationOpenAIVideosRemix:
			method, suffix = http.MethodPost, "/remix"
		}
		return method, "/videos/" + value + suffix, nil
	case platform.ProviderOperationOpenAIVideosEdit:
		return http.MethodPost, "/videos/edits", nil
	case platform.ProviderOperationOpenAIVideosExtend:
		return http.MethodPost, "/videos/extensions", nil
	case platform.ProviderOperationOpenAIModelsList:
		return http.MethodGet, "/models", nil
	case platform.ProviderOperationOpenAIModelsGet:
		value, err := resource()
		if err != nil {
			return "", "", err
		}
		return http.MethodGet, "/models/" + value, nil
	default:
		return "", "", fmt.Errorf("unsupported or missing fixed provider operation %q", payload.Operation)
	}
}

func resolveLegacyOpenAICompatibleOperation(metadata map[string]interface{}) (string, string, error) {
	path := "/chat/completions"
	if raw, ok := metadata["provider_endpoint_path"].(string); ok && strings.TrimSpace(raw) != "" {
		path = strings.TrimSpace(raw)
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
	}
	// Compatibility is limited to the finite endpoint set used by older
	// Control Plane releases. Any other value is rejected fail-closed.
	allowed := map[string]struct{}{
		"/chat/completions": {}, "/completions": {}, "/responses": {},
		"/embeddings": {}, "/moderations": {}, "/rerank": {}, "/v2/rerank": {},
		"/images/generations": {}, "/images/edits": {}, "/images/variations": {},
		"/audio/speech": {}, "/audio/transcriptions": {}, "/audio/translations": {},
		"/videos": {},
	}
	if _, ok := allowed[path]; !ok {
		return "", "", fmt.Errorf("unsupported legacy provider endpoint %q", path)
	}
	return http.MethodPost, path, nil
}

func withListQuery(path string, input platform.ProviderOperationInput) string {
	query := url.Values{}
	if input.BeforeID != "" {
		query.Set("before", input.BeforeID)
	}
	if input.AfterID != "" {
		query.Set("after", input.AfterID)
	}
	if input.Limit > 0 {
		query.Set("limit", strconv.Itoa(input.Limit))
	}
	if input.Order != "" {
		query.Set("order", input.Order)
	}
	for _, include := range input.Include {
		if value := strings.TrimSpace(include); value != "" {
			query.Add("include", value)
		}
	}
	if input.IncludeObfuscation != nil {
		query.Set("include_obfuscation", strconv.FormatBool(*input.IncludeObfuscation))
	}
	if input.StartingAfter != nil {
		query.Set("starting_after", strconv.FormatInt(*input.StartingAfter, 10))
	}
	if input.Stream != nil {
		query.Set("stream", strconv.FormatBool(*input.Stream))
	}
	if encoded := query.Encode(); encoded != "" {
		return path + "?" + encoded
	}
	return path
}
