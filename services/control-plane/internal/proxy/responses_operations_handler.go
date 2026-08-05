package proxy

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apiheaders"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
)

const responsesResourceKind = "response"

type providerNativeResourceStore interface {
	UpsertProviderNativeResourceMapping(
		ctx context.Context,
		mapping platform.ProviderNativeResourceMapping,
	) (*platform.ProviderNativeResourceMapping, error)
	GetProviderNativeResourceMapping(
		ctx context.Context,
		tenantID string,
		provider string,
		resourceKind string,
		publicID string,
	) (*platform.ProviderNativeResourceMapping, error)
	UpdateProviderNativeResourceStatus(
		ctx context.Context,
		tenantID string,
		provider string,
		resourceKind string,
		publicID string,
		status string,
	) error
}

type responseDeploymentLister interface {
	ListDeployments(ctx context.Context, tenantID string) ([]platform.Deployment, error)
}

// HandleRetrieve implements GET /v1/responses/{response_id}.
func (h *ResponsesHandler) HandleRetrieve(c *fiber.Ctx) error {
	return h.handleResponseResourceOperation(c, platform.ProviderOperationOpenAIResponsesGet)
}

// HandleDelete implements DELETE /v1/responses/{response_id}.
func (h *ResponsesHandler) HandleDelete(c *fiber.Ctx) error {
	return h.handleResponseResourceOperation(c, platform.ProviderOperationOpenAIResponsesDelete)
}

// HandleCancel implements POST /v1/responses/{response_id}/cancel.
func (h *ResponsesHandler) HandleCancel(c *fiber.Ctx) error {
	return h.handleResponseResourceOperation(c, platform.ProviderOperationOpenAIResponsesCancel)
}

// HandleInputItems implements GET /v1/responses/{response_id}/input_items.
func (h *ResponsesHandler) HandleInputItems(c *fiber.Ctx) error {
	return h.handleResponseResourceOperation(c, platform.ProviderOperationOpenAIResponsesInputItems)
}

// HandleCompact implements POST /v1/responses/compact.
func (h *ResponsesHandler) HandleCompact(c *fiber.Ctx) error {
	return h.handleResponseBodyOperation(
		c,
		platform.ProviderOperationOpenAIResponsesCompact,
		"/responses/compact",
		false,
	)
}

// HandleInputTokens implements POST /v1/responses/input_tokens.
func (h *ResponsesHandler) HandleInputTokens(c *fiber.Ctx) error {
	return h.handleResponseBodyOperation(
		c,
		platform.ProviderOperationOpenAIResponsesInputTokens,
		"/responses/input_tokens",
		false,
	)
}

//nolint:govet // Short-lived branch errors are checked immediately and intentionally scoped with short declarations.
func (h *ResponsesHandler) handleResponseResourceOperation(
	c *fiber.Ctx,
	operation platform.ProviderOperation,
) error {
	authInfo, currentTenant, ok := responsesRequestContext(c)
	if !ok {
		return nil
	}

	publicID := strings.TrimSpace(c.Params("response_id"))
	if publicID == "" {
		return writeAPIError(c, fiber.StatusBadRequest, "response_id is required", "invalid_request_error")
	}

	operationInput, stream, err := parseResponseOperationQuery(c, operation)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error")
	}

	operationInput.ResourceID = publicID

	accountSelector := strings.TrimSpace(c.Get("x-godarai-provider-account"))
	providerSelector := ""
	resourceModel := ""

	if h.resources != nil {
		mapping, mappingErr := h.responseResourceMapping(c.Context(), currentTenant.ID, publicID)
		switch {
		case mappingErr == nil && mapping != nil:
			if model, _ := mapping.Metadata["model"].(string); strings.TrimSpace(model) != "" { //nolint:errcheck // Optional metadata uses the zero value when absent or mismatched.
				resourceModel = normalizeModelIdentifier(model)

				modelAccess := modelAccessForType(authInfo, normalizeModelIdentifier(model), platform.ModelTypeResponses)
				if !modelAccess.Allowed {
					return writeAPIError(c, fiber.StatusForbidden, "Model access denied or unsupported for responses", "authorization_error")
				}
			}

			accountSelector = mapping.ProviderAccountID
			providerSelector = mapping.Provider
			operationInput.ResourceID = mapping.UpstreamID
		case mappingErr != nil && !errors.Is(mappingErr, storage.ErrProviderNativeResourceNotFound):
			return writeAPIError(c, fiber.StatusServiceUnavailable, "Response routing state is unavailable", "routing_error")
		}
	}

	deployment, err := h.resolveResponseDeployment(c.Context(), currentTenant.ID, providerSelector, accountSelector, "")
	if err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, err.Error(), "routing_error")
	}

	requestID := responseRequestID(c)

	tunnelReq := platform.TunnelRequest{
		RequestID:      requestID,
		TenantID:       currentTenant.ID,
		TraceID:        requestID,
		Stream:         stream,
		Operation:      operation,
		OperationInput: operationInput,
		Deployment:     deployment,
		Headers:        responseProviderHeaders(c),
		RequestMetadata: enrichRequestMetadataForDeployment(map[string]interface{}{
			"tenant_slug":            apiheaders.GetTenant(c),
			"provider_endpoint_path": responseOperationPath(operation, publicID),
			"response_id":            publicID,
		}, deployment),
	}
	if strings.EqualFold(deployment.Provider, "xai") {
		if resourceModel == "" {
			return writeAPIError(c, fiber.StatusServiceUnavailable, "xAI response model binding is unavailable", "routing_error")
		}

		estimated, estimateErr := h.billing.EstimateCost(c.Context(), contractsopenai.ChatCompletionRequest{Model: billingModelForDeployment(resourceModel, deployment)})
		if estimateErr != nil {
			return writeBillingEstimateError(c, estimateErr)
		}

		runtimeModel := strings.TrimSpace(deployment.Model)
		if runtimeModel == "" || runtimeModel == "*" {
			runtimeModel = providerModelID(resourceModel)
		}

		tunnelReq.ModelRuntimeProfiles, err = runtimeProfilesFromEstimate(deployment.Provider, runtimeModel, estimated)
		if err != nil {
			return writeAPIError(c, fiber.StatusServiceUnavailable, err.Error(), "runtime_profile_unavailable")
		}

		if err := validateXAIResponseResourceOperation(operation, tunnelReq.ModelRuntimeProfiles[0].Profile); err != nil {
			return writeAPIError(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error")
		}
	}

	if err := h.tunnel.Healthy(c.Context()); err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Tunnel path unavailable", "tunnel_unavailable")
	}

	if stream {
		return h.relayResponseOperationStream(c, tunnelReq)
	}

	resp, err := h.tunnel.Proxy(c.Context(), tunnelReq)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, fmt.Sprintf("Tunnel error: %v", err), "api_error")
	}

	if resp.StatusCode >= httpErrorFloor {
		return writeOpenAIProviderError(c, resp.StatusCode, resp.Body, resp.Headers)
	}

	switch operation {
	case platform.ProviderOperationOpenAIResponsesDelete:
		h.updateResponseMappingStatus(c.Context(), currentTenant.ID, deployment.Provider, publicID, "deleted")
	case platform.ProviderOperationOpenAIResponsesCancel:
		h.updateResponseMappingStatusFromBody(c.Context(), currentTenant.ID, deployment.Provider, publicID, resp.Body)
	}

	forwardProviderHeaders(c, resp.Headers)
	setTraceHeaders(c, requestID)
	c.Status(resp.StatusCode)
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	return c.Send(resp.Body)
}

//nolint:govet // Short-lived branch errors are checked immediately and intentionally scoped with short declarations.
func (h *ResponsesHandler) handleResponseBodyOperation(
	c *fiber.Ctx,
	operation platform.ProviderOperation,
	endpoint string,
	requireModel bool,
) error {
	authInfo, currentTenant, ok := responsesRequestContext(c)
	if !ok {
		return nil
	}

	var err error

	body := append([]byte(nil), c.Body()...)

	var object map[string]json.RawMessage
	if err := json.Unmarshal(body, &object); err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, "Invalid request body", "invalid_request_error")
	}

	var requestedModel string
	if rawModel, ok := object["model"]; ok {
		if err := json.Unmarshal(rawModel, &requestedModel); err != nil {
			return writeAPIError(c, fiber.StatusBadRequest, "model must be a string", "invalid_request_error")
		}

		requestedModel = normalizeModelIdentifier(requestedModel)
	}

	if requireModel && requestedModel == "" {
		return writeAPIError(c, fiber.StatusBadRequest, "Model is required", "invalid_request_error")
	}

	var deployment platform.Deployment

	if requestedModel != "" {
		modelAccess := modelAccessForType(authInfo, requestedModel, platform.ModelTypeResponses)
		if !modelAccess.Allowed {
			if modelAccess.UnsupportedVMMode {
				return writeUnsupportedModelModeError(c, requestedModel, platform.ModelTypeResponses, modelAccess.ExposedModes)
			}

			return writeAPIError(c, fiber.StatusForbidden, "Model access denied or unsupported for responses", "authorization_error")
		}

		var routedErr error

		for _, candidate := range modelRoutingCandidates(requestedModel) {
			routed, routeErr := h.llmRouter.Route(c.Context(), currentTenant.ID, candidate)
			if routeErr == nil && routed != nil {
				deployment = toContractDeployment(routed)
				routedErr = nil

				break
			}

			routedErr = routeErr
		}

		if routedErr != nil || deployment.Provider == "" {
			return writeAPIError(c, fiber.StatusServiceUnavailable, fmt.Sprintf("No available deployment for model: %s", requestedModel), "routing_error")
		}

		if !supportsResponsesProvider(deployment.Provider) {
			return writeAPIError(c, fiber.StatusNotImplemented, fmt.Sprintf("Responses API is not enabled for provider: %s", deployment.Provider), "invalid_request_error")
		}

		outboundModel := strings.TrimSpace(deployment.Model)
		if outboundModel == "" || outboundModel == "*" {
			outboundModel = providerModelID(requestedModel)
		}

		body, err = rewriteOpenAIJSONModel(body, outboundModel)
		if err != nil {
			return writeAPIError(c, fiber.StatusBadRequest, "Failed to encode request", "invalid_request_error")
		}
	} else {
		providerSelector := strings.ToLower(strings.TrimSpace(c.Get("x-godarai-provider")))
		if providerSelector == "" {
			providerSelector = "openai"
		}

		if providerSelector != "openai" && providerSelector != "xai" {
			return writeAPIError(c, fiber.StatusBadRequest, "x-godarai-provider must be openai or xai", "invalid_request_error")
		}

		deployment, err = h.resolveResponseDeployment(
			c.Context(),
			currentTenant.ID,
			providerSelector,
			strings.TrimSpace(c.Get("x-godarai-provider-account")),
			"",
		)
		if err != nil {
			return writeAPIError(c, fiber.StatusServiceUnavailable, err.Error(), "routing_error")
		}
	}

	requestID := responseRequestID(c)

	tunnelReq := platform.TunnelRequest{
		RequestID:  requestID,
		TenantID:   currentTenant.ID,
		TraceID:    requestID,
		Operation:  operation,
		Deployment: deployment,
		Headers:    responseProviderHeaders(c),
		Body:       body,
		RequestMetadata: enrichRequestMetadataForDeployment(map[string]interface{}{
			"tenant_slug":            apiheaders.GetTenant(c),
			"model":                  requestedModel,
			"provider_endpoint_path": endpoint,
		}, deployment),
	}
	if strings.EqualFold(strings.TrimSpace(deployment.Provider), "xai") {
		var responseRequest contractsopenai.ResponseRequest
		if err := json.Unmarshal(body, &responseRequest); err != nil {
			return writeAPIError(c, fiber.StatusBadRequest, "Invalid request body", "invalid_request_error")
		}

		estimateRequest := responseRequestToChatCompletion(
			responseRequest,
			billingModelForDeployment(requestedModel, deployment),
		)

		estimated, estimateErr := h.billing.EstimateCost(c.Context(), estimateRequest)
		if estimateErr != nil {
			return writeBillingEstimateError(c, estimateErr)
		}

		runtimeModel := strings.TrimSpace(deployment.Model)
		if runtimeModel == "" || runtimeModel == "*" {
			runtimeModel = providerModelID(requestedModel)
		}

		tunnelReq.ModelRuntimeProfiles, err = runtimeProfilesFromEstimate(
			deployment.Provider,
			runtimeModel,
			estimated,
		)
		if err != nil {
			return writeAPIError(c, fiber.StatusServiceUnavailable, err.Error(), "runtime_profile_unavailable")
		}

		profile := tunnelReq.ModelRuntimeProfiles[0].Profile
		if operation == platform.ProviderOperationOpenAIResponsesCompact && !profile.ResponseResources.Compact {
			return writeAPIError(c, fiber.StatusBadRequest, "model does not support Responses compaction", "invalid_request_error")
		}

		if operation == platform.ProviderOperationOpenAIResponsesInputTokens && !profile.ResponseResources.CountInputTokens {
			return writeAPIError(c, fiber.StatusBadRequest, "model does not support Responses input token counting", "invalid_request_error")
		}
	}

	if err := h.tunnel.Healthy(c.Context()); err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Tunnel path unavailable", "tunnel_unavailable")
	}

	resp, err := h.tunnel.Proxy(c.Context(), tunnelReq)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, fmt.Sprintf("Tunnel error: %v", err), "api_error")
	}

	if resp.StatusCode >= httpErrorFloor {
		return writeOpenAIProviderError(c, resp.StatusCode, resp.Body, resp.Headers)
	}

	forwardProviderHeaders(c, resp.Headers)
	setTraceHeaders(c, requestID)
	c.Status(resp.StatusCode)
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	return c.Send(resp.Body)
}

func responsesRequestContext(c *fiber.Ctx) (*platform.AuthContext, *platform.Tenant, bool) {
	authInfo := auth.GetAuthInfo(c)
	if authInfo == nil {
		_ = writeAPIError(c, fiber.StatusUnauthorized, "Authentication required", "authentication_error") //nolint:errcheck // Optional metadata uses the zero value when absent or mismatched.
		return nil, nil, false
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		_ = writeAPIError(c, fiber.StatusBadRequest, "Tenant resolution required", "invalid_request_error") //nolint:errcheck // Optional metadata uses the zero value when absent or mismatched.
		return nil, nil, false
	}

	authInfo.TenantID = currentTenant.ID
	tracing.SetTenantForTrace(c, currentTenant.ID)

	return authInfo, currentTenant, true
}

func (h *ResponsesHandler) resolveResponseDeployment(
	ctx context.Context,
	tenantID string,
	providerSelector string,
	accountSelector string,
	requestedModel string,
) (platform.Deployment, error) {
	if h.deployments == nil {
		return platform.Deployment{}, fmt.Errorf("responses deployment routing is unavailable")
	}

	deployments, err := h.deployments.ListDeployments(ctx, tenantID)
	if err != nil {
		return platform.Deployment{}, fmt.Errorf("responses deployment routing is unavailable")
	}

	accountSelector = strings.TrimSpace(accountSelector)
	providerSelector = strings.ToLower(strings.TrimSpace(providerSelector))
	requestedModel = normalizeModelIdentifier(requestedModel)

	candidates := make([]platform.Deployment, 0, len(deployments))
	for _, deployment := range deployments {
		provider := strings.ToLower(strings.TrimSpace(deployment.Provider))
		if !supportsResponsesProvider(provider) || provider == "mock" {
			continue
		}

		if providerSelector != "" && provider != providerSelector {
			continue
		}

		if accountSelector != "" &&
			!strings.EqualFold(accountSelector, strings.TrimSpace(deployment.ProviderAccountID)) &&
			!strings.EqualFold(accountSelector, strings.TrimSpace(deployment.ProviderAccountName)) {
			continue
		}

		if requestedModel != "" &&
			normalizeModelIdentifier(deployment.ModelName) != requestedModel &&
			normalizeModelIdentifier(deployment.Model) != requestedModel &&
			providerModelID(deployment.Model) != providerModelID(requestedModel) {
			continue
		}

		candidates = append(candidates, deployment)
	}

	if len(candidates) == 0 {
		if accountSelector != "" {
			return platform.Deployment{}, fmt.Errorf("responses provider account %q was not found", accountSelector)
		}

		return platform.Deployment{}, fmt.Errorf("no Responses provider account is available")
	}

	if accountSelector == "" {
		accountIDs := make(map[string]struct{})

		for _, candidate := range candidates {
			key := strings.TrimSpace(candidate.ProviderAccountID)
			if key == "" {
				key = strings.TrimSpace(candidate.ProviderAccountName)
			}

			if key == "" {
				key = candidate.ID
			}

			accountIDs[key] = struct{}{}
		}

		if len(accountIDs) > 1 {
			return platform.Deployment{}, fmt.Errorf("response provider account is ambiguous")
		}
	}

	return h.hydrateResponseProviderAccountDeployment(ctx, tenantID, candidates[0])
}

func (h *ResponsesHandler) hydrateResponseProviderAccountDeployment(
	ctx context.Context,
	tenantID string,
	deployment platform.Deployment,
) (platform.Deployment, error) {
	if !strings.HasPrefix(strings.TrimSpace(deployment.ID), "provider-account:") {
		return deployment, nil
	}

	if h.llmRouter == nil {
		return platform.Deployment{}, fmt.Errorf("responses provider account routing is unavailable")
	}

	modelName := strings.TrimSpace(deployment.ModelName)
	if modelName == "" || modelName == "*" {
		return platform.Deployment{}, fmt.Errorf("responses provider account model is unavailable")
	}

	routed, err := h.llmRouter.Route(ctx, tenantID, modelName)
	if err != nil {
		return platform.Deployment{}, fmt.Errorf("responses provider account routing is unavailable: %w", err)
	}

	if routed == nil {
		return platform.Deployment{}, fmt.Errorf("responses provider account routing is unavailable")
	}

	hydrated := toContractDeployment(routed)
	if !responseProviderAccountMatches(deployment, hydrated) {
		return platform.Deployment{}, fmt.Errorf("responses provider account routing changed account affinity")
	}

	return hydrated, nil
}

func responseProviderAccountMatches(expected platform.Deployment, actual platform.Deployment) bool {
	if !strings.EqualFold(strings.TrimSpace(expected.Provider), strings.TrimSpace(actual.Provider)) {
		return false
	}

	if accountID := strings.TrimSpace(expected.ProviderAccountID); accountID != "" {
		return strings.EqualFold(accountID, strings.TrimSpace(actual.ProviderAccountID))
	}

	if accountName := strings.TrimSpace(expected.ProviderAccountName); accountName != "" {
		return strings.EqualFold(accountName, strings.TrimSpace(actual.ProviderAccountName))
	}

	return strings.EqualFold(strings.TrimSpace(expected.ID), strings.TrimSpace(actual.ID))
}

func (h *ResponsesHandler) responseResourceMapping(
	ctx context.Context,
	tenantID string,
	publicID string,
) (*platform.ProviderNativeResourceMapping, error) {
	if h.resources == nil {
		return nil, storage.ErrProviderNativeResourceNotFound
	}

	var found *platform.ProviderNativeResourceMapping

	for _, provider := range []string{"openai", "xai"} {
		mapping, err := h.resources.GetProviderNativeResourceMapping(
			ctx,
			tenantID,
			provider,
			responsesResourceKind,
			publicID,
		)
		if errors.Is(err, storage.ErrProviderNativeResourceNotFound) {
			continue
		}

		if err != nil {
			return nil, err
		}

		if mapping == nil {
			continue
		}

		if found != nil {
			return nil, fmt.Errorf("ambiguous provider response resource")
		}

		found = mapping
	}

	if found == nil {
		return nil, storage.ErrProviderNativeResourceNotFound
	}

	return found, nil
}

func responseProviderFromModel(model string) string {
	value := strings.ToLower(strings.TrimSpace(model))
	switch {
	case strings.HasPrefix(value, "xai/"), strings.HasPrefix(value, "xai:"), strings.HasPrefix(providerModelID(value), "grok-"):
		return "xai"
	case strings.HasPrefix(value, "openai/"), strings.HasPrefix(value, "openai:"):
		return "openai"
	default:
		return ""
	}
}

// resolveResponseCreateDeployment pins continuations to the account that owns
// the previous response. It also supplies a model for authorization, limits,
// and billing when OpenAI permits the client to omit model.
func (h *ResponsesHandler) resolveResponseCreateDeployment(
	ctx context.Context,
	tenantID string,
	requestedModel string,
	previousResponseID string,
	accountSelector string,
) (*platform.Deployment, string, string, error) {
	requestedModel = normalizeModelIdentifier(requestedModel)
	previousResponseID = strings.TrimSpace(previousResponseID)
	accountSelector = strings.TrimSpace(accountSelector)
	upstreamPreviousResponseID := previousResponseID
	mappedModel := ""
	providerSelector := responseProviderFromModel(requestedModel)

	if previousResponseID != "" && h.resources != nil {
		mapping, err := h.responseResourceMapping(ctx, tenantID, previousResponseID)
		switch {
		case err == nil && mapping != nil:
			accountSelector = strings.TrimSpace(mapping.ProviderAccountID)
			providerSelector = strings.ToLower(strings.TrimSpace(mapping.Provider))

			upstreamPreviousResponseID = strings.TrimSpace(mapping.UpstreamID)
			if upstreamPreviousResponseID == "" {
				upstreamPreviousResponseID = previousResponseID
			}

			if model, ok := mapping.Metadata["model"].(string); ok {
				mappedModel = normalizeModelIdentifier(model)
			}
		case err != nil && !errors.Is(err, storage.ErrProviderNativeResourceNotFound):
			return nil, "", "", fmt.Errorf("response routing state is unavailable")
		}
	}

	if requestedModel != "" && accountSelector == "" {
		return nil, requestedModel, upstreamPreviousResponseID, nil
	}

	deployment, err := h.resolveResponseDeployment(ctx, tenantID, providerSelector, accountSelector, requestedModel)
	if err != nil && accountSelector != "" && requestedModel != "" {
		// Account affinity is mandatory, while a wildcard provider deployment
		// may not advertise a concrete model in its deployment row.
		deployment, err = h.resolveResponseDeployment(ctx, tenantID, providerSelector, accountSelector, "")
	}

	if err != nil {
		return nil, "", "", err
	}

	effectiveModel := requestedModel
	if effectiveModel == "" {
		effectiveModel = mappedModel
	}

	if effectiveModel == "" {
		effectiveModel = normalizeModelIdentifier(deployment.ModelName)
	}

	if effectiveModel == "" || effectiveModel == "*" {
		effectiveModel = normalizeModelIdentifier(deployment.Model)
	}

	if effectiveModel == "*" {
		effectiveModel = ""
	}

	return &deployment, effectiveModel, upstreamPreviousResponseID, nil
}

func parseResponseOperationQuery(
	c *fiber.Ctx,
	operation platform.ProviderOperation,
) (platform.ProviderOperationInput, bool, error) {
	var input platform.ProviderOperationInput

	c.Context().QueryArgs().VisitAll(func(key, value []byte) {
		name := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(string(key))), "[]")
		if name == "include" && strings.TrimSpace(string(value)) != "" {
			input.Include = append(input.Include, strings.TrimSpace(string(value)))
		}
	})

	switch operation {
	case platform.ProviderOperationOpenAIResponsesGet:
		if raw := strings.TrimSpace(c.Query("include_obfuscation")); raw != "" {
			value, err := strconv.ParseBool(raw)
			if err != nil {
				return input, false, fmt.Errorf("include_obfuscation must be a boolean")
			}

			input.IncludeObfuscation = &value
		}

		if raw := strings.TrimSpace(c.Query("starting_after")); raw != "" {
			value, err := strconv.ParseInt(raw, 10, 64)
			if err != nil || value < 0 {
				return input, false, fmt.Errorf("starting_after must be a non-negative integer")
			}

			input.StartingAfter = &value
		}

		if raw := strings.TrimSpace(c.Query("stream")); raw != "" {
			value, err := strconv.ParseBool(raw)
			if err != nil {
				return input, false, fmt.Errorf("stream must be a boolean")
			}

			input.Stream = &value

			return input, value, nil
		}
	case platform.ProviderOperationOpenAIResponsesInputItems:
		input.AfterID = strings.TrimSpace(c.Query("after"))
		if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
			value, err := strconv.Atoi(raw)
			if err != nil || value < 1 || value > 100 {
				return input, false, fmt.Errorf("limit must be between 1 and 100")
			}

			input.Limit = value
		}

		input.Order = strings.ToLower(strings.TrimSpace(c.Query("order")))
		if input.Order != "" && input.Order != "asc" && input.Order != "desc" {
			return input, false, fmt.Errorf("order must be asc or desc")
		}
	}

	return input, false, nil
}

func responseProviderHeaders(c *fiber.Ctx) map[string]string {
	headers := map[string]string{"Content-Type": "application/json"}
	if value := strings.TrimSpace(c.Get("X-Client-Request-Id")); value != "" {
		headers["X-Client-Request-Id"] = value
	}

	return headers
}

func responseRequestID(c *fiber.Ctx) string {
	if traceID := strings.TrimSpace(tracing.TraceIDFromContext(c)); traceID != "" {
		return traceID
	}

	return fmt.Sprintf("req_%d", c.Context().Time().UnixNano())
}

func responseOperationPath(operation platform.ProviderOperation, responseID string) string {
	base := "/responses/" + responseID

	switch operation {
	case platform.ProviderOperationOpenAIResponsesCancel:
		return base + "/cancel"
	case platform.ProviderOperationOpenAIResponsesInputItems:
		return base + "/input_items"
	default:
		return base
	}
}

func (h *ResponsesHandler) relayResponseOperationStream(
	c *fiber.Ctx,
	tunnelReq platform.TunnelRequest,
) error {
	stream, err := h.tunnel.Stream(c.Context(), tunnelReq)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, fmt.Sprintf("Tunnel error: %v", err), "api_error")
	}

	if stream.StatusCode() >= httpErrorFloor {
		return writeOpenAITunnelStreamHTTPError(c, h.logger, stream)
	}

	c.Set(fiber.HeaderContentType, "text/event-stream")
	c.Set(fiber.HeaderCacheControl, "no-cache")
	c.Set(fiber.HeaderConnection, "keep-alive")
	forwardProviderHeaders(c, stream.Headers())
	setTraceHeaders(c, tunnelReq.TraceID)
	c.Context().SetBodyStreamWriter(func(writer *bufio.Writer) {
		defer closeStreamBestEffort(h.logger, stream)

		relayTunnelStreamWithErrorFramer(
			stream,
			writer,
			func() {},
			h.logger,
			tunnelReq.RequestID,
			"responses_retrieve",
			frameOpenAIResponseStreamError,
		)
	})

	return nil
}

func (h *ResponsesHandler) recordResponseMapping(
	ctx context.Context,
	tenantID string,
	deployment platform.Deployment,
	responseID string,
	status string,
	model string,
) {
	if h.resources == nil ||
		strings.TrimSpace(responseID) == "" ||
		strings.TrimSpace(deployment.ProviderAccountID) == "" ||
		!supportsResponsesProvider(deployment.Provider) {
		return
	}

	if strings.TrimSpace(status) == "" {
		status = "active"
	}

	_, err := h.resources.UpsertProviderNativeResourceMapping(ctx, platform.ProviderNativeResourceMapping{
		PublicID:          responseID,
		TenantID:          tenantID,
		Provider:          strings.ToLower(strings.TrimSpace(deployment.Provider)),
		ProviderAccountID: deployment.ProviderAccountID,
		ResourceKind:      responsesResourceKind,
		UpstreamID:        responseID,
		Status:            status,
		Metadata: map[string]interface{}{
			"deployment_id": deployment.ID,
			"model":         model,
		},
	})
	if err != nil && h.logger != nil {
		h.logger.Warn("failed to persist OpenAI response routing state")
	}
}

func (h *ResponsesHandler) updateResponseMappingStatus(
	ctx context.Context,
	tenantID string,
	provider string,
	responseID string,
	status string,
) {
	if h.resources == nil {
		return
	}

	if err := h.resources.UpdateProviderNativeResourceStatus(
		ctx,
		tenantID,
		strings.ToLower(strings.TrimSpace(provider)),
		responsesResourceKind,
		responseID,
		status,
	); err != nil && !errors.Is(err, storage.ErrProviderNativeResourceNotFound) && h.logger != nil {
		h.logger.Warn("failed to update OpenAI response routing state")
	}
}

func (h *ResponsesHandler) updateResponseMappingStatusFromBody(
	ctx context.Context,
	tenantID string,
	provider string,
	responseID string,
	body []byte,
) {
	var result struct {
		Status string `json:"status"`
	}
	if json.Unmarshal(body, &result) == nil && strings.TrimSpace(result.Status) != "" {
		h.updateResponseMappingStatus(ctx, tenantID, provider, responseID, result.Status)
	}
}

func (h *ResponsesHandler) newResponseStreamObserver(
	_ context.Context,
	tenantID string,
	deployment platform.Deployment,
	onResponse ...func(contractsopenai.Response),
) func([]byte) {
	canRecordMapping := h.resources != nil &&
		strings.TrimSpace(deployment.ProviderAccountID) != ""
	if !canRecordMapping && len(onResponse) == 0 {
		return nil
	}

	var pending []byte

	recorded := false

	return func(chunk []byte) {
		if len(chunk) == 0 {
			return
		}

		pending = append(pending, chunk...)
		for {
			end, delimiterWidth := responseSSEFrameBoundary(pending)
			if end < 0 {
				if len(pending) > 256*1024 {
					pending = append([]byte(nil), pending[len(pending)-64*1024:]...)
				}

				return
			}

			frame := pending[:end]
			pending = pending[end+delimiterWidth:]

			for _, line := range bytes.Split(frame, []byte("\n")) {
				line = bytes.TrimSpace(line)
				if !bytes.HasPrefix(line, []byte("data:")) {
					continue
				}

				var event struct {
					Response contractsopenai.Response `json:"response"`
				}
				if json.Unmarshal(bytes.TrimSpace(bytes.TrimPrefix(line, []byte("data:"))), &event) != nil ||
					strings.TrimSpace(event.Response.ID) == "" {
					continue
				}

				if canRecordMapping && !recorded {
					recordCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					h.recordResponseMapping(
						recordCtx,
						tenantID,
						deployment,
						event.Response.ID,
						event.Response.Status,
						event.Response.Model,
					)
					cancel()

					recorded = true
				} else if canRecordMapping && strings.TrimSpace(event.Response.Status) != "" {
					statusCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					h.updateResponseMappingStatus(statusCtx, tenantID, deployment.Provider, event.Response.ID, event.Response.Status)
					cancel()
				}

				for _, observe := range onResponse {
					if observe != nil {
						observe(event.Response)
					}
				}
			}
		}
	}
}

func responseSSEFrameBoundary(data []byte) (int, int) {
	lf := bytes.Index(data, []byte("\n\n"))

	crlf := bytes.Index(data, []byte("\r\n\r\n"))
	switch {
	case lf < 0:
		return crlf, 4
	case crlf < 0:
		return lf, 2
	case crlf < lf:
		return crlf, 4
	default:
		return lf, 2
	}
}
