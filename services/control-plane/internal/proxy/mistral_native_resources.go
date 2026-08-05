package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func (h *MistralNativeHandler) rewriteMistralSSEChunk(ctx context.Context, tenantID string, request platform.TunnelRequest, chunk []byte, model string) ([]byte, *contractsopenai.Usage, error) {
	lines := bytes.Split(chunk, []byte("\n"))
	var latestUsage *contractsopenai.Usage

	for index, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if !bytes.HasPrefix(trimmed, []byte("data:")) {
			continue
		}

		payload := bytes.TrimSpace(bytes.TrimPrefix(trimmed, []byte("data:")))
		if bytes.Equal(payload, []byte("[DONE]")) || len(payload) == 0 {
			continue
		}

		var value interface{}
		if json.Unmarshal(payload, &value) != nil {
			continue
		}
		if usage := mistralUsageFromValue(value); usage != nil {
			latestUsage = usage
		}

		known := make(map[string]string)
		if request.OperationInput.ResourceID != "" {
			mapping, err := h.resources.GetProviderNativeResourceMappingByUpstream(
				ctx,
				tenantID,
				mistralProviderName,
				request.Deployment.ProviderAccountID,
				mistralConversationResourceKind,
				request.OperationInput.ResourceID,
			)
			if err == nil && mapping != nil {
				known[mapping.UpstreamID] = mapping.PublicID
			}
		}

		if request.Operation == platform.ProviderOperationMistralConversationsCreate {
			if upstreamID := mistralConversationIDFromValue(value); upstreamID != "" {
				mapping, err := h.resources.GetProviderNativeResourceMappingByUpstream(
					ctx,
					tenantID,
					mistralProviderName,
					request.Deployment.ProviderAccountID,
					mistralConversationResourceKind,
					upstreamID,
				)
				if err != nil || mapping == nil {
					mapping, err = h.resources.UpsertProviderNativeResourceMapping(ctx, platform.ProviderNativeResourceMapping{
						PublicID: newMistralPublicID(mistralConversationResourceKind), TenantID: tenantID,
						Provider: mistralProviderName, ProviderAccountID: request.Deployment.ProviderAccountID,
						ResourceKind: mistralConversationResourceKind, UpstreamID: upstreamID, Status: "active",
						Metadata: map[string]interface{}{"deployment_id": request.Deployment.ID, "model": model},
					})
					if err != nil {
						return nil, nil, err
					}
				}

				known[upstreamID] = mapping.PublicID
			}
		}

		value = rewriteMistralResponseValue(value, known, false)
		rewritten, err := json.Marshal(value)
		if err != nil {
			return nil, nil, err
		}

		prefixLength := bytes.Index(line, []byte("data:"))
		prefix := line[:prefixLength]
		lines[index] = append(append(prefix, []byte("data: ")...), rewritten...)
	}

	return bytes.Join(lines, []byte("\n")), latestUsage, nil
}

func mistralConversationIDFromValue(value interface{}) string {
	object, ok := value.(map[string]interface{})
	if !ok {
		return ""
	}
	if id := strings.TrimSpace(stringFromMap(object, "conversation_id")); id != "" {
		return id
	}
	if eventType := strings.ToLower(strings.TrimSpace(stringFromMap(object, "type"))); strings.Contains(eventType, "conversation") {
		return strings.TrimSpace(stringFromMap(object, "id"))
	}

	return ""
}

func mistralUsageFromValue(value interface{}) *contractsopenai.Usage {
	switch typed := value.(type) {
	case map[string]interface{}:
		if raw, ok := typed["usage"].(map[string]interface{}); ok {
			prompt, promptOK := mistralNonnegativeTokenCount(raw, "prompt_tokens")
			completion, completionOK := mistralNonnegativeTokenCount(raw, "completion_tokens")
			if promptOK && completionOK {
				total, totalOK := mistralNonnegativeTokenCount(raw, "total_tokens")
				if !totalOK {
					maxInt := int(^uint(0) >> 1)
					if prompt > maxInt-completion {
						return nil
					}

					total = prompt + completion
				}

				return &contractsopenai.Usage{PromptTokens: prompt, CompletionTokens: completion, TotalTokens: total}
			}
		}
		for _, child := range typed {
			if usage := mistralUsageFromValue(child); usage != nil {
				return usage
			}
		}
	case []interface{}:
		for _, child := range typed {
			if usage := mistralUsageFromValue(child); usage != nil {
				return usage
			}
		}
	}

	return nil
}

func mistralNonnegativeTokenCount(values map[string]interface{}, key string) (int, bool) {
	number, exists, err := numberFromMap(values, key)
	if err != nil || !exists || number < 0 {
		return 0, false
	}

	maxInt := int(^uint(0) >> 1)
	if number > float64(maxInt) || number != float64(int(number)) {
		return 0, false
	}

	return int(number), true
}

func mistralSSEIsError(data []byte) bool {
	return bytes.Contains(bytes.ToLower(data), []byte(`"type":"error"`)) ||
		bytes.Contains(bytes.ToLower(data), []byte(`"error":`))
}

func (h *MistralNativeHandler) rewriteRequestReferences(
	ctx context.Context,
	tenantID string,
	accountID string,
	model string,
	body []byte,
) ([]byte, string, string, error) {
	if len(body) == 0 {
		return body, model, accountID, nil
	}

	var object map[string]interface{}
	if err := json.Unmarshal(body, &object); err != nil {
		return nil, "", "", fmt.Errorf("invalid JSON request body")
	}

	resolveReference := func(field, kind, publicID string) (string, error) {
		mapping, err := h.resources.GetProviderNativeResourceMapping(ctx, tenantID, mistralProviderName, kind, publicID)
		if err != nil || mapping == nil {
			return "", fmt.Errorf("%s references a missing or inaccessible resource", field)
		}
		if accountID != "" && !strings.EqualFold(accountID, mapping.ProviderAccountID) {
			return "", fmt.Errorf("all referenced resources must belong to the selected provider account")
		}

		accountID = mapping.ProviderAccountID
		if model == "" && kind == mistralAgentResourceKind {
			model, _ = mapping.Metadata["model"].(string)
		}

		return mapping.UpstreamID, nil
	}

	var rewriteValue func(value interface{}) error
	rewriteValue = func(value interface{}) error {
		switch typed := value.(type) {
		case map[string]interface{}:
			for field, child := range typed {
				canonical := strings.ToLower(strings.TrimSpace(field))
				kind := ""
				switch canonical {
				case "file_id":
					kind = mistralFileResourceKind
				case "agent_id":
					kind = mistralAgentResourceKind
				}
				if kind != "" {
					publicID, ok := child.(string)
					if !ok || strings.TrimSpace(publicID) == "" {
						return fmt.Errorf("%s must be a resource ID", field)
					}

					upstreamID, err := resolveReference(field, kind, strings.TrimSpace(publicID))
					if err != nil {
						return err
					}

					typed[field] = upstreamID
					continue
				}

				if canonical == "input_files" || canonical == "handoffs" {
					values, ok := child.([]interface{})
					if !ok {
						return fmt.Errorf("%s must be an array of resource IDs", field)
					}

					kind = mistralFileResourceKind
					if canonical == "handoffs" {
						kind = mistralAgentResourceKind
					}
					for index, rawID := range values {
						publicID, ok := rawID.(string)
						if !ok || strings.TrimSpace(publicID) == "" {
							return fmt.Errorf("%s must contain resource IDs", field)
						}

						upstreamID, err := resolveReference(field, kind, strings.TrimSpace(publicID))
						if err != nil {
							return err
						}

						values[index] = upstreamID
					}

					typed[field] = values
					continue
				}

				if err := rewriteValue(child); err != nil {
					return err
				}
			}
		case []interface{}:
			for _, child := range typed {
				if err := rewriteValue(child); err != nil {
					return err
				}
			}
		}

		return nil
	}
	if err := rewriteValue(object); err != nil {
		return nil, "", "", err
	}

	rewritten, err := json.Marshal(object)
	if err != nil {
		return nil, "", "", err
	}

	return rewritten, model, accountID, nil
}

func (h *MistralNativeHandler) mistralBatchRequestCount(ctx context.Context, tenantID string, body []byte) (int, error) {
	var request struct {
		Endpoint   string            `json:"endpoint"`
		InputFiles []string          `json:"input_files"`
		Requests   []json.RawMessage `json:"requests"`
	}
	if err := json.Unmarshal(body, &request); err != nil {
		return 0, fmt.Errorf("invalid Mistral batch request")
	}
	if len(request.Requests) > 0 {
		return len(request.Requests), nil
	}
	if len(request.InputFiles) == 0 {
		return 0, fmt.Errorf("input_files or requests is required")
	}

	target := mistralBatchRuntimeEndpoint(request.Endpoint)
	if target == "ocr" || target == "audio_transcriptions" {
		return len(request.InputFiles), nil
	}

	count := 0
	maxInt := int(^uint(0) >> 1)
	for _, publicID := range request.InputFiles {
		mapping, err := h.resources.GetProviderNativeResourceMapping(
			ctx,
			tenantID,
			mistralProviderName,
			mistralFileResourceKind,
			strings.TrimSpace(publicID),
		)
		if err != nil || mapping == nil {
			return 0, fmt.Errorf("input_files references a missing or inaccessible resource")
		}
		policyValidated, _ := mapping.Metadata["batch_policy_validated"].(bool)
		if !policyValidated {
			return 0, fmt.Errorf("input file was not validated for Mistral batch policy; upload it again")
		}

		lines, ok := mistralMetadataPositiveInt(mapping.Metadata, "num_lines")
		if !ok {
			return 0, fmt.Errorf("input file line count is unavailable; retrieve the file after processing and retry")
		}
		if count > maxInt-lines {
			return 0, fmt.Errorf("Mistral batch request count is too large")
		}

		count += lines
	}

	return count, nil
}

func mistralMetadataPositiveInt(metadata map[string]interface{}, key string) (int, bool) {
	value, exists := metadata[key]
	if !exists {
		return 0, false
	}

	var number int
	switch typed := value.(type) {
	case int:
		number = typed
	case int64:
		if typed > int64(^uint(0)>>1) {
			return 0, false
		}

		number = int(typed)
	case float64:
		if typed != float64(int(typed)) {
			return 0, false
		}

		number = int(typed)
	default:
		return 0, false
	}

	return number, number > 0
}

func (h *MistralNativeHandler) persistAndRewriteResponse(
	ctx context.Context,
	tenantID string,
	operation platform.ProviderOperation,
	deployment platform.Deployment,
	body []byte,
	resource *platform.ProviderNativeResourceMapping,
	reservation *platform.BudgetReservation,
	estimated platform.UsageCostBreakdown,
	model string,
	extraMetadata map[string]interface{},
) ([]byte, error) {
	if len(bytes.TrimSpace(body)) == 0 {
		return body, nil
	}

	kind := mistralOperationResourceKind(operation)
	mappings, err := h.resources.ListProviderNativeResourceMappings(
		ctx,
		tenantID,
		mistralProviderName,
		deployment.ProviderAccountID,
		kind,
		nil,
		500,
	)
	if err != nil {
		return nil, err
	}

	owned := mappingsByUpstream(mappings)
	known := cloneMistralIDMap(owned)
	if mistralOperationIsCreate(operation) {
		upstreamID, status, err := mistralResponseIdentity(body)
		if err != nil {
			return nil, err
		}

		metadata := map[string]interface{}{"deployment_id": deployment.ID}
		if model != "" {
			metadata["model"] = model
		}
		if operation == platform.ProviderOperationMistralFilesCreate {
			for key, value := range mistralResponseResourceMetadata(body) {
				metadata[key] = value
			}
			for key, value := range extraMetadata {
				metadata[key] = value
			}
		}
		if operation == platform.ProviderOperationMistralBatchCreate {
			if endpoint, ok := estimated.Metadata["batch_endpoint"].(string); ok && endpoint != "" {
				metadata["endpoint"] = endpoint
			}
			if endpoint, ok := estimated.Metadata["batch_runtime_endpoint"].(string); ok && endpoint != "" {
				metadata["runtime_endpoint"] = endpoint
			}
		}

		created := platform.ProviderNativeResourceMapping{
			PublicID: newMistralPublicID(kind), TenantID: tenantID,
			Provider: mistralProviderName, ProviderAccountID: deployment.ProviderAccountID,
			ResourceKind: kind, UpstreamID: upstreamID, Status: status,
			Metadata: metadata, Reservation: reservation, EstimatedUsage: nil,
		}
		if operation == platform.ProviderOperationMistralBatchCreate {
			created.PricingRevision = estimated.PricingRevision
			created.EstimatedUsage = &estimated
			if requestID, ok := estimated.Metadata["request_id"].(string); ok {
				created.ReservationID = requestID
			}
		}
		if reservation != nil {
			created.ReservationID = reservation.ID
			created.Reservation = reservation
		}

		persisted, persistErr := h.resources.UpsertProviderNativeResourceMapping(ctx, created)
		if persistErr != nil {
			return nil, persistErr
		}

		known[upstreamID] = persisted.PublicID
	}
	if resource != nil {
		known[resource.UpstreamID] = resource.PublicID
		status := mistralResponseStatus(body)
		workerMustSettle := kind == mistralBatchResourceKind && mistralBatchStatusIsTerminal(status)
		if status != "" && status != resource.Status && !workerMustSettle {
			_ = h.resources.UpdateProviderNativeResourceStatus(ctx, tenantID, mistralProviderName, kind, resource.PublicID, status) //nolint:errcheck // The provider response remains authoritative; the worker retries reconciliation.
			resource.Status = status
		}
		if operation == platform.ProviderOperationMistralAgentsUpdate && model != "" {
			updated := *resource
			updated.Metadata = cloneMistralMetadata(resource.Metadata)
			updated.Metadata["model"] = model
			if _, updateErr := h.resources.UpsertProviderNativeResourceMapping(ctx, updated); updateErr != nil {
				return nil, updateErr
			}
		}
		if operation == platform.ProviderOperationMistralFilesGet {
			updated := *resource
			updated.Metadata = cloneMistralMetadata(resource.Metadata)
			for key, value := range mistralResponseResourceMetadata(body) {
				updated.Metadata[key] = value
			}
			if _, updateErr := h.resources.UpsertProviderNativeResourceMapping(ctx, updated); updateErr != nil {
				return nil, updateErr
			}
		}
	}

	if operation == platform.ProviderOperationMistralBatchCreate ||
		operation == platform.ProviderOperationMistralBatchGet ||
		operation == platform.ProviderOperationMistralBatchCancel {
		for _, upstreamID := range mistralResponseFileReferences(body) {
			if _, exists := known[upstreamID]; exists {
				continue
			}

			created, persistErr := h.resources.UpsertProviderNativeResourceMapping(ctx, platform.ProviderNativeResourceMapping{
				PublicID: newMistralPublicID(mistralFileResourceKind), TenantID: tenantID,
				Provider: mistralProviderName, ProviderAccountID: deployment.ProviderAccountID,
				ResourceKind: mistralFileResourceKind, UpstreamID: upstreamID, Status: "active",
				Metadata: map[string]interface{}{"deployment_id": deployment.ID, "source": "batch_output"},
			})
			if persistErr != nil {
				return nil, persistErr
			}

			known[upstreamID] = created.PublicID
		}
	}

	for _, relatedKind := range []string{mistralFileResourceKind, mistralAgentResourceKind, mistralConversationResourceKind} {
		related, listErr := h.resources.ListProviderNativeResourceMappings(
			ctx,
			tenantID,
			mistralProviderName,
			deployment.ProviderAccountID,
			relatedKind,
			nil,
			500,
		)
		if listErr != nil {
			return nil, listErr
		}
		for upstreamID, publicID := range mappingsByUpstream(related) {
			known[upstreamID] = publicID
		}
	}

	if mistralOperationIsList(operation) {
		return rewriteMistralOwnedListResponseJSON(body, known, owned, true)
	}

	return rewriteMistralResponseJSON(body, known, false)
}

func mistralResponseStatus(body []byte) string {
	var response map[string]interface{}
	if json.Unmarshal(body, &response) != nil {
		return ""
	}

	return strings.ToLower(strings.TrimSpace(stringFromMap(response, "status")))
}

func mistralResponseFileReferences(body []byte) []string {
	var value interface{}
	if json.Unmarshal(body, &value) != nil {
		return nil
	}

	seen := make(map[string]struct{})
	collectMistralFileReferences(value, seen)

	result := make([]string, 0, len(seen))
	for id := range seen {
		result = append(result, id)
	}

	return result
}

func collectMistralFileReferences(value interface{}, seen map[string]struct{}) {
	switch typed := value.(type) {
	case map[string]interface{}:
		for key, child := range typed {
			switch strings.ToLower(strings.TrimSpace(key)) {
			case "output_file", "output_file_id", "error_file", "error_file_id":
				if id, ok := child.(string); ok && strings.TrimSpace(id) != "" {
					seen[id] = struct{}{}
				}
			}

			collectMistralFileReferences(child, seen)
		}
	case []interface{}:
		for _, child := range typed {
			collectMistralFileReferences(child, seen)
		}
	}
}

func mistralResponseIdentity(body []byte) (string, string, error) {
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", "", fmt.Errorf("invalid Mistral resource response")
	}

	upstreamID := strings.TrimSpace(stringFromMap(response, "id"))
	if upstreamID == "" {
		upstreamID = strings.TrimSpace(stringFromMap(response, "batch_id"))
	}
	if upstreamID == "" {
		upstreamID = strings.TrimSpace(stringFromMap(response, "conversation_id"))
	}
	if upstreamID == "" {
		return "", "", fmt.Errorf("Mistral resource response is missing id")
	}

	status := strings.ToLower(strings.TrimSpace(stringFromMap(response, "status")))
	if status == "" {
		status = "active"
	}

	return upstreamID, status, nil
}

func cloneMistralMetadata(source map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(source)+1)
	for key, value := range source {
		result[key] = value
	}

	return result
}

func mistralResponseResourceMetadata(body []byte) map[string]interface{} {
	var response map[string]interface{}
	if json.Unmarshal(body, &response) != nil {
		return nil
	}

	metadata := make(map[string]interface{})
	for _, key := range []string{"bytes", "size", "num_lines", "purpose", "filename", "file_name", "sample_type"} {
		if value, exists := response[key]; exists {
			metadata[key] = value
		}
	}

	return metadata
}

func mappingsByUpstream(mappings []platform.ProviderNativeResourceMapping) map[string]string {
	result := make(map[string]string, len(mappings))
	for _, mapping := range mappings {
		result[mapping.UpstreamID] = mapping.PublicID
	}

	return result
}

func cloneMistralIDMap(source map[string]string) map[string]string {
	result := make(map[string]string, len(source))
	for upstreamID, publicID := range source {
		result[upstreamID] = publicID
	}

	return result
}

func rewriteMistralResponseJSON(body []byte, known map[string]string, filterList bool) ([]byte, error) {
	return rewriteMistralOwnedListResponseJSON(body, known, known, filterList)
}

func rewriteMistralOwnedListResponseJSON(body []byte, known, owned map[string]string, filterList bool) ([]byte, error) {
	var value interface{}
	if err := json.Unmarshal(body, &value); err != nil {
		return nil, err
	}

	value = rewriteMistralResponseValueOwned(value, known, owned, filterList)

	return json.Marshal(value)
}

func rewriteMistralResponseValue(value interface{}, known map[string]string, filterList bool) interface{} {
	return rewriteMistralResponseValueOwned(value, known, known, filterList)
}

func rewriteMistralResponseValueOwned(value interface{}, known, owned map[string]string, filterList bool) interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		ownedCount := -1
		for key, child := range typed {
			if array, ok := child.([]interface{}); ok && filterList && mistralListField(key) {
				filtered := filterMistralOwnedList(array, known, owned)
				typed[key] = filtered
				ownedCount = len(filtered)

				continue
			}

			typed[key] = rewriteMistralResponseValueOwned(child, known, owned, false)
		}
		if filterList {
			if ownedCount < 0 {
				ownedCount = len(owned)
			}
			for _, key := range []string{"total", "total_count", "count"} {
				if _, exists := typed[key]; exists {
					typed[key] = ownedCount
				}
			}
		}

		return typed
	case []interface{}:
		if filterList {
			return filterMistralOwnedList(typed, known, owned)
		}
		for index, child := range typed {
			typed[index] = rewriteMistralResponseValueOwned(child, known, owned, false)
		}

		return typed
	case string:
		if publicID, ok := known[typed]; ok {
			return publicID
		}

		return typed
	default:
		return typed
	}
}

func filterMistralOwnedList(values []interface{}, known, owned map[string]string) []interface{} {
	filtered := make([]interface{}, 0, len(values))
	for _, value := range values {
		object, ok := value.(map[string]interface{})
		if !ok {
			continue
		}

		upstreamID := strings.TrimSpace(stringFromMap(object, "id"))
		if upstreamID == "" {
			upstreamID = strings.TrimSpace(stringFromMap(object, "batch_id"))
		}
		if _, ok := owned[upstreamID]; !ok {
			continue
		}

		filtered = append(filtered, rewriteMistralResponseValueOwned(object, known, owned, false))
	}

	return filtered
}

func mistralListField(field string) bool {
	switch strings.ToLower(strings.TrimSpace(field)) {
	case "data", "files", "jobs", "agents", "conversations":
		return true
	default:
		return false
	}
}
