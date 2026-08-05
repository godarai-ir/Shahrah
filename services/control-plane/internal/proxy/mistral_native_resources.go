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
		if mapping, err := h.resources.GetProviderNativeResourceMappingByUpstream(
			ctx,
			tenantID,
			mistralProviderName,
			request.Deployment.ProviderAccountID,
			mistralConversationResourceKind,
			request.OperationInput.ResourceID,
		); err == nil && mapping != nil {
			known[mapping.UpstreamID] = mapping.PublicID
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
			prompt, _, _ := numberFromMap(raw, "prompt_tokens")         //nolint:errcheck // Missing provider usage fields are zero.
			completion, _, _ := numberFromMap(raw, "completion_tokens") //nolint:errcheck // Missing provider usage fields are zero.
			total, _, _ := numberFromMap(raw, "total_tokens")           //nolint:errcheck // Missing provider usage fields are zero.

			return &contractsopenai.Usage{PromptTokens: int(prompt), CompletionTokens: int(completion), TotalTokens: int(total)}
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

	rewriteOne := func(field, kind string) error {
		publicID := strings.TrimSpace(stringFromMap(object, field))
		if publicID == "" {
			return nil
		}

		mapping, err := h.resources.GetProviderNativeResourceMapping(ctx, tenantID, mistralProviderName, kind, publicID)
		if err != nil || mapping == nil {
			return fmt.Errorf("%s references a missing or inaccessible resource", field)
		}
		if accountID != "" && !strings.EqualFold(accountID, mapping.ProviderAccountID) {
			return fmt.Errorf("all referenced resources must belong to the selected provider account")
		}

		accountID = mapping.ProviderAccountID
		object[field] = mapping.UpstreamID
		if model == "" && kind == mistralAgentResourceKind {
			model, _ = mapping.Metadata["model"].(string)
		}

		return nil
	}
	if err := rewriteOne("file_id", mistralFileResourceKind); err != nil {
		return nil, "", "", err
	}
	if err := rewriteOne("agent_id", mistralAgentResourceKind); err != nil {
		return nil, "", "", err
	}

	for field, kind := range map[string]string{
		"input_files": mistralFileResourceKind,
		"handoffs":    mistralAgentResourceKind,
	} {
		values, ok := object[field].([]interface{})
		if !ok {
			continue
		}

		for index, value := range values {
			publicID, ok := value.(string)
			if !ok || strings.TrimSpace(publicID) == "" {
				return nil, "", "", fmt.Errorf("%s must contain resource IDs", field)
			}

			mapping, err := h.resources.GetProviderNativeResourceMapping(ctx, tenantID, mistralProviderName, kind, publicID)
			if err != nil || mapping == nil {
				return nil, "", "", fmt.Errorf("%s references a missing or inaccessible resource", field)
			}
			if accountID != "" && !strings.EqualFold(accountID, mapping.ProviderAccountID) {
				return nil, "", "", fmt.Errorf("all referenced resources must belong to the selected provider account")
			}

			accountID = mapping.ProviderAccountID
			values[index] = mapping.UpstreamID
		}

		object[field] = values
	}

	rewritten, err := json.Marshal(object)
	if err != nil {
		return nil, "", "", err
	}

	return rewritten, model, accountID, nil
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
) ([]byte, error) {
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

	known := mappingsByUpstream(mappings)
	if mistralOperationIsCreate(operation) {
		upstreamID, status, err := mistralResponseIdentity(body)
		if err != nil {
			return nil, err
		}

		metadata := map[string]interface{}{"deployment_id": deployment.ID}
		if model != "" {
			metadata["model"] = model
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
		if reservation != nil {
			created.ReservationID = reservation.ID
			created.PricingRevision = estimated.PricingRevision
			created.EstimatedUsage = &estimated
		}

		persisted, persistErr := h.resources.UpsertProviderNativeResourceMapping(ctx, created)
		if persistErr != nil {
			return nil, persistErr
		}

		known[upstreamID] = persisted.PublicID
	}
	if resource != nil {
		known[resource.UpstreamID] = resource.PublicID
		if status := mistralResponseStatus(body); status != "" && status != resource.Status {
			_ = h.resources.UpdateProviderNativeResourceStatus(ctx, tenantID, mistralProviderName, kind, resource.PublicID, status) //nolint:errcheck // The provider response remains authoritative; the worker retries reconciliation.
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

	return rewriteMistralResponseJSON(body, known, mistralOperationIsList(operation))
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
		return "", "", fmt.Errorf("Mistral resource response is missing id")
	}

	status := strings.ToLower(strings.TrimSpace(stringFromMap(response, "status")))
	if status == "" {
		status = "active"
	}

	return upstreamID, status, nil
}

func mappingsByUpstream(mappings []platform.ProviderNativeResourceMapping) map[string]string {
	result := make(map[string]string, len(mappings))
	for _, mapping := range mappings {
		result[mapping.UpstreamID] = mapping.PublicID
	}

	return result
}

func rewriteMistralResponseJSON(body []byte, known map[string]string, filterList bool) ([]byte, error) {
	var value interface{}
	if err := json.Unmarshal(body, &value); err != nil {
		return nil, err
	}

	value = rewriteMistralResponseValue(value, known, filterList)

	return json.Marshal(value)
}

func rewriteMistralResponseValue(value interface{}, known map[string]string, filterList bool) interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		for key, child := range typed {
			if array, ok := child.([]interface{}); ok && filterList && mistralListField(key) {
				typed[key] = filterMistralOwnedList(array, known)

				continue
			}

			typed[key] = rewriteMistralResponseValue(child, known, false)
		}

		return typed
	case []interface{}:
		if filterList {
			return filterMistralOwnedList(typed, known)
		}
		for index, child := range typed {
			typed[index] = rewriteMistralResponseValue(child, known, false)
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

func filterMistralOwnedList(values []interface{}, known map[string]string) []interface{} {
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
		if _, ok := known[upstreamID]; !ok {
			continue
		}

		filtered = append(filtered, rewriteMistralResponseValue(object, known, false))
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
