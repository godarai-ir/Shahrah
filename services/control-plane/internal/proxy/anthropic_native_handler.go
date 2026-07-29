package proxy

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type anthropicDeploymentLister interface {
	ListDeployments(ctx context.Context, tenantID string) ([]platform.Deployment, error)
}

type anthropicNativeResourceStore interface {
	UpsertAnthropicFileMapping(ctx context.Context, mapping platform.AnthropicNativeResourceMapping) (*platform.AnthropicNativeResourceMapping, error)
	GetAnthropicFileMapping(ctx context.Context, tenantID, publicID string) (*platform.AnthropicNativeResourceMapping, error)
	UpsertAnthropicBatchMapping(ctx context.Context, mapping platform.AnthropicNativeResourceMapping) (*platform.AnthropicNativeResourceMapping, error)
	GetAnthropicBatchMapping(ctx context.Context, tenantID, publicID string) (*platform.AnthropicNativeResourceMapping, error)
	UpdateAnthropicBatchStatus(ctx context.Context, tenantID, publicID, status string) error
	ListPendingAnthropicBatchMappings(ctx context.Context, limit int) ([]platform.AnthropicNativeResourceMapping, error)
}

// AnthropicNativeHandler exposes Anthropic-shaped operations while preserving
// GodarAI authentication, account routing, tunnel isolation and billing.
type AnthropicNativeHandler struct {
	router           *router.Router
	deployments      anthropicDeploymentLister
	resources        anthropicNativeResourceStore
	tunnel           platform.TunnelTransport
	billing          budgeter
	wallets          walletStore
	runtimeProfiles  anthropicModelRuntimeProfileResolver
	logger           *zap.Logger
	nativeEnabled    bool
	filesEnabled     bool
	batchesEnabled   bool
	accountAllowlist map[string]struct{}
}

func NewAnthropicNativeHandler(deps BaseHandlerDependencies, deployments anthropicDeploymentLister) *AnthropicNativeHandler {
	resources, _ := deployments.(anthropicNativeResourceStore)
	return &AnthropicNativeHandler{
		router: deps.Router, deployments: deployments, tunnel: deps.Tunnel,
		resources: resources, billing: deps.Billing, wallets: deps.Wallets,
		runtimeProfiles: deps.AnthropicRuntimeProfiles, logger: deps.Logger,
		nativeEnabled:    anthropicFeatureEnabled("ANTHROPIC_NATIVE_API_ENABLED", true),
		filesEnabled:     anthropicFeatureEnabled("ANTHROPIC_FILES_API_ENABLED", true),
		batchesEnabled:   anthropicFeatureEnabled("ANTHROPIC_BATCH_API_ENABLED", true),
		accountAllowlist: anthropicAccountAllowlist(),
	}
}

func (h *AnthropicNativeHandler) Handle(c *fiber.Ctx) error {
	if strings.TrimSpace(c.Get("anthropic-version")) == "" {
		return writeAnthropicError(c, fiber.StatusBadRequest, "invalid_request_error", "anthropic-version is required")
	}
	authInfo := auth.GetAuthInfo(c)
	currentTenant := tenant.FromContext(c)
	if authInfo == nil {
		return writeAnthropicError(c, fiber.StatusUnauthorized, "authentication_error", "Authentication required")
	}
	if currentTenant == nil {
		return writeAnthropicError(c, fiber.StatusBadRequest, "invalid_request_error", "Tenant resolution required")
	}
	authInfo.TenantID = currentTenant.ID

	operation, input, err := anthropicOperationFromRequest(c)
	if err != nil {
		return writeAnthropicError(c, fiber.StatusBadRequest, "invalid_request_error", err.Error())
	}
	if !h.operationEnabled(operation) {
		return writeAnthropicError(c, fiber.StatusServiceUnavailable, "api_error", "Anthropic operation is disabled")
	}
	if operation == platform.ProviderOperationAnthropicFilesCreate ||
		operation == platform.ProviderOperationAnthropicFilesContent ||
		operation == platform.ProviderOperationAnthropicBatchesResults {
		return h.handleRawOperation(c, currentTenant.ID, operation, input)
	}
	body := append([]byte(nil), c.Body()...)
	models, stream, err := anthropicRequestModels(operation, input, body)
	if err != nil {
		return writeAnthropicError(c, fiber.StatusBadRequest, "invalid_request_error", err.Error())
	}
	pricingConditions, err := anthropicCodeExecutionPricingPolicy(body, operation)
	if err != nil {
		return writeAnthropicError(
			c,
			fiber.StatusBadRequest,
			"invalid_request_error",
			err.Error(),
		)
	}

	accountSelector := c.Get("x-godarai-provider-account")
	var resourceMapping *platform.AnthropicNativeResourceMapping
	if isAnthropicFileResourceOperation(operation) || isAnthropicBatchResourceOperation(operation) {
		if h.resources == nil {
			return writeAnthropicError(c, fiber.StatusServiceUnavailable, "api_error", "Native resource mapping is unavailable")
		}
		publicID := input.ResourceID
		if isAnthropicFileResourceOperation(operation) {
			resourceMapping, err = h.resources.GetAnthropicFileMapping(c.Context(), currentTenant.ID, publicID)
		} else {
			resourceMapping, err = h.resources.GetAnthropicBatchMapping(c.Context(), currentTenant.ID, publicID)
		}
		if err != nil || resourceMapping == nil {
			return writeAnthropicError(c, fiber.StatusNotFound, "not_found_error", "Resource not found")
		}
		accountSelector = resourceMapping.ProviderAccountID
		input.ResourceID = resourceMapping.UpstreamID
	}

	deployment, officialModels, err := h.resolveDeployment(c.Context(), currentTenant.ID, models, accountSelector)
	if err != nil {
		code := "routing_error"
		status := fiber.StatusServiceUnavailable
		if strings.Contains(err.Error(), "ambiguous") || strings.Contains(err.Error(), "same provider account") {
			code, status = "ambiguous_provider_account", fiber.StatusBadRequest
		}
		return writeAnthropicError(c, status, code, err.Error())
	}
	if deployment == nil {
		return writeAnthropicError(c, fiber.StatusServiceUnavailable, "routing_error", "No Anthropic provider account is available")
	}
	if !h.accountEnabled(deployment.ProviderAccountID) {
		return writeAnthropicError(c, fiber.StatusForbidden, "permission_error", "Anthropic native API is not enabled for this provider account")
	}
	if len(officialModels) > 0 {
		body, err = replaceAnthropicNativeModels(operation, body, officialModels)
		if err != nil {
			return writeAnthropicError(c, fiber.StatusBadRequest, "invalid_request_error", err.Error())
		}
	}
	if operation == platform.ProviderOperationAnthropicModelsGet && len(officialModels) == 1 {
		input.ResourceID = officialModels[0]
	}

	requestID := strings.TrimSpace(c.GetRespHeader("request-id"))
	if requestID == "" {
		requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}
	tunnelRequest := platform.TunnelRequest{
		RequestID: requestID, TenantID: currentTenant.ID, Stream: stream,
		Operation: operation, OperationInput: input, Deployment: *deployment,
		Headers: map[string]string{
			"content-type":   c.Get(fiber.HeaderContentType),
			"anthropic-beta": c.Get("anthropic-beta"),
		},
		Body: body,
	}

	var estimated platform.UsageCostBreakdown
	var reservation *platform.BudgetReservation
	if operation == platform.ProviderOperationAnthropicMessages ||
		operation == platform.ProviderOperationAnthropicBatchesCreate {
		if h.billing == nil {
			return writeAnthropicError(c, fiber.StatusServiceUnavailable, "billing_error", "Billing is unavailable")
		}
		estimateRequest := contractsopenai.ChatCompletionRequest{
			Model:    officialModels[0],
			Messages: []contractsopenai.ChatMessage{{Role: "user", Content: string(body)}},
		}
		if max := anthropicMaximumOutputTokens(body, operation); max > 0 {
			estimateRequest.MaxTokens = &max
		}
		estimated, err = h.billing.EstimateCost(c.Context(), estimateRequest)
		if err != nil {
			return writeAnthropicBillingError(c, err)
		}
		if len(pricingConditions) > 0 {
			if estimated.PricingConditions == nil {
				estimated.PricingConditions = make(map[string]string)
			}
			for key, value := range pricingConditions {
				estimated.PricingConditions[key] = value
			}
		}
		var handled bool
		reservation, handled, err = reserveBudgetForRequest(c, h.billing, authInfo, estimated, requestID, deployment.BillingMode)
		if err != nil || handled {
			return err
		}
	}
	if anthropicOperationNeedsModelRuntimeProfile(operation) {
		if h.runtimeProfiles == nil {
			releaseBudgetBestEffort(
				c.Context(),
				h.logger,
				h.billing,
				reservation,
				"runtime_profile_unavailable",
			)
			return writeAnthropicError(
				c,
				fiber.StatusServiceUnavailable,
				"runtime_profile_unavailable",
				"Anthropic model runtime profile is unavailable",
			)
		}
		expectedRevision := estimated.PricingRevision
		tunnelRequest.ModelRuntimeProfiles, err =
			h.runtimeProfiles.ResolveModelRuntimeProfiles(
				officialModels,
				expectedRevision,
			)
		if err != nil {
			releaseBudgetBestEffort(
				c.Context(),
				h.logger,
				h.billing,
				reservation,
				"runtime_profile_unavailable",
			)
			return writeAnthropicError(
				c,
				fiber.StatusServiceUnavailable,
				"runtime_profile_unavailable",
				"Anthropic model runtime profile is unavailable",
			)
		}
	}

	if stream {
		return h.proxyStream(c, tunnelRequest, reservation, estimated, authInfo, officialModels[0])
	}
	response, err := h.tunnel.Proxy(c.Context(), tunnelRequest)
	if err != nil {
		releaseBudgetBestEffort(c.Context(), h.logger, h.billing, reservation, "tunnel_error")
		return writeAnthropicError(c, fiber.StatusBadGateway, "api_error", "Tunnel request failed")
	}
	if response.StatusCode >= fiber.StatusBadRequest {
		releaseBudgetBestEffort(c.Context(), h.logger, h.billing, reservation, "provider_http_error")
		forwardProviderHeaders(c, response.Headers)
		return c.Status(response.StatusCode).Send(response.Body)
	}
	response.Body, err = h.rewriteNativeResponse(c.Context(), operation, currentTenant.ID, deployment.ProviderAccountID, response.Body, reservation, estimated, resourceMapping)
	if err != nil {
		releaseBudgetBestEffort(c.Context(), h.logger, h.billing, reservation, "resource_mapping_error")
		return writeAnthropicError(c, fiber.StatusBadGateway, "api_error", "Failed to persist native resource mapping")
	}
	if operation == platform.ProviderOperationAnthropicMessages && reservation != nil {
		usage := parseAnthropicNativeUsage(response.Body)
		finalized, finalizeErr := h.billing.FinalizeFromUsage(c.Context(), officialModels[0], usage, estimated)
		if finalizeErr != nil {
			releaseBudgetBestEffort(c.Context(), h.logger, h.billing, reservation, "usage_finalize_error")
			return writeAnthropicBillingError(c, finalizeErr)
		}
		if err := commitBudgetOrDebit(c.Context(), h.logger, h.billing, h.wallets, reservation, currentTenant.ID, finalized, requestID, deployment.BillingMode); err != nil {
			return writeAnthropicError(c, fiber.StatusPaymentRequired, "billing_error", err.Error())
		}
	}
	if operation == platform.ProviderOperationAnthropicBatchesCancel && resourceMapping != nil {
		releaseBudgetBestEffort(c.Context(), h.logger, h.billing, resourceMapping.Reservation, "batch_canceled")
		_ = h.resources.UpdateAnthropicBatchStatus(c.Context(), currentTenant.ID, resourceMapping.PublicID, "canceled")
	}
	forwardProviderHeaders(c, response.Headers)
	if contentType := response.Headers["Content-Type"]; contentType != "" {
		c.Set(fiber.HeaderContentType, contentType)
	}
	return c.Status(response.StatusCode).Send(response.Body)
}

func anthropicOperationNeedsModelRuntimeProfile(
	operation platform.ProviderOperation,
) bool {
	switch operation {
	case platform.ProviderOperationAnthropicMessages,
		platform.ProviderOperationAnthropicCountTokens,
		platform.ProviderOperationAnthropicBatchesCreate:
		return true
	default:
		return false
	}
}

func anthropicFeatureEnabled(name string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false
	}
	return value
}

func anthropicAccountAllowlist() map[string]struct{} {
	values := map[string]struct{}{}
	for _, value := range strings.Split(os.Getenv("ANTHROPIC_NATIVE_ACCOUNT_ALLOWLIST"), ",") {
		if id := strings.TrimSpace(value); id != "" {
			values[id] = struct{}{}
		}
	}
	return values
}

func (h *AnthropicNativeHandler) accountEnabled(accountID string) bool {
	if len(h.accountAllowlist) == 0 {
		return true
	}
	_, ok := h.accountAllowlist[strings.TrimSpace(accountID)]
	return ok
}

func (h *AnthropicNativeHandler) operationEnabled(operation platform.ProviderOperation) bool {
	if !h.nativeEnabled {
		return false
	}
	switch operation {
	case platform.ProviderOperationAnthropicFilesCreate,
		platform.ProviderOperationAnthropicFilesList,
		platform.ProviderOperationAnthropicFilesGet,
		platform.ProviderOperationAnthropicFilesDelete,
		platform.ProviderOperationAnthropicFilesContent:
		return h.filesEnabled
	case platform.ProviderOperationAnthropicBatchesCreate,
		platform.ProviderOperationAnthropicBatchesList,
		platform.ProviderOperationAnthropicBatchesGet,
		platform.ProviderOperationAnthropicBatchesCancel,
		platform.ProviderOperationAnthropicBatchesResults:
		return h.batchesEnabled
	default:
		return true
	}
}

func (h *AnthropicNativeHandler) resolveDeployment(ctx context.Context, tenantID string, requestedModels []string, accountSelector string) (*platform.Deployment, []string, error) {
	if h == nil || h.router == nil || h.deployments == nil {
		return nil, nil, fmt.Errorf("Anthropic routing is unavailable")
	}
	deployments, err := h.deployments.ListDeployments(ctx, tenantID)
	if err != nil {
		return nil, nil, err
	}
	anthropic := make([]platform.Deployment, 0)
	for _, deployment := range deployments {
		if strings.EqualFold(deployment.Provider, "anthropic") {
			anthropic = append(anthropic, deployment)
		}
	}
	if len(anthropic) == 0 {
		return nil, nil, fmt.Errorf("No Anthropic provider account is available")
	}
	selector := strings.TrimSpace(accountSelector)
	var selected *platform.Deployment
	official := make([]string, 0, len(requestedModels))
	for _, requested := range requestedModels {
		requested = strings.TrimSpace(requested)
		var matches []platform.Deployment
		for _, deployment := range anthropic {
			if selector != "" && !strings.EqualFold(selector, deployment.ProviderAccountID) &&
				!strings.EqualFold(selector, deployment.ProviderAccountName) {
				continue
			}
			if requested == "" || strings.EqualFold(requested, deployment.Model) ||
				strings.EqualFold(requested, deployment.ModelName) ||
				strings.EqualFold(strings.TrimPrefix(requested, "anthropic/"), deployment.Model) {
				matches = append(matches, deployment)
			}
		}
		accounts := uniqueAnthropicAccounts(matches)
		if len(accounts) == 0 {
			return nil, nil, fmt.Errorf("Anthropic model %q is not configured", requested)
		}
		if len(accounts) > 1 {
			return nil, nil, fmt.Errorf("ambiguous provider account for model %q", requested)
		}
		match := matches[0]
		if selected != nil && selected.ProviderAccountID != match.ProviderAccountID {
			return nil, nil, fmt.Errorf("all batch items must use the same provider account")
		}
		selected = &match
		official = append(official, match.Model)
	}
	if selected == nil {
		if selector == "" {
			accounts := uniqueAnthropicAccounts(anthropic)
			if len(accounts) > 1 {
				return nil, nil, fmt.Errorf("ambiguous provider account; x-godarai-provider-account is required")
			}
		}
		for _, deployment := range anthropic {
			if selector == "" || strings.EqualFold(selector, deployment.ProviderAccountID) ||
				strings.EqualFold(selector, deployment.ProviderAccountName) {
				selected = &deployment
				break
			}
		}
	}
	if selected == nil {
		return nil, nil, fmt.Errorf("Anthropic provider account %q was not found", selector)
	}
	resolved, err := h.router.Route(ctx, tenantID, selected.ModelName)
	if err != nil {
		return nil, nil, err
	}
	contract := toContractDeployment(resolved)
	return &contract, official, nil
}

func uniqueAnthropicAccounts(deployments []platform.Deployment) map[string]struct{} {
	out := map[string]struct{}{}
	for _, deployment := range deployments {
		key := deployment.ProviderAccountID
		if key == "" {
			key = deployment.ProviderAccountName
		}
		if key == "" {
			key = deployment.ID
		}
		out[key] = struct{}{}
	}
	return out
}

func (h *AnthropicNativeHandler) proxyStream(c *fiber.Ctx, request platform.TunnelRequest, reservation *platform.BudgetReservation, estimated platform.UsageCostBreakdown, authInfo *platform.AuthContext, model string) error {
	stream, err := h.tunnel.Stream(c.Context(), request)
	if err != nil {
		releaseBudgetBestEffort(c.Context(), h.logger, h.billing, reservation, "stream_open_failed")
		return writeAnthropicError(c, fiber.StatusBadGateway, "api_error", "Tunnel stream failed")
	}
	if stream.StatusCode() >= fiber.StatusBadRequest {
		releaseBudgetBestEffort(c.Context(), h.logger, h.billing, reservation, "provider_http_error")
		return writeTunnelStreamHTTPError(c, h.logger, stream, nil)
	}
	c.Set(fiber.HeaderContentType, "text/event-stream")
	forwardProviderHeaders(c, stream.Headers())
	requestContext := c.Context()
	requestContext.SetBodyStreamWriter(func(writer *bufio.Writer) {
		defer closeStreamBestEffort(h.logger, stream)
		defer writer.Flush()
		var usage *contractsopenai.Usage
		failed := false
		for {
			select {
			case chunk, ok := <-stream.Chunks():
				if !ok {
					if failed || usage == nil {
						releaseBudgetBestEffort(context.Background(), h.logger, h.billing, reservation, "stream_incomplete")
						return
					}
					finalized, finalizeErr := h.billing.FinalizeFromUsage(context.Background(), model, usage, estimated)
					if finalizeErr != nil {
						releaseBudgetBestEffort(context.Background(), h.logger, h.billing, reservation, "usage_finalize_error")
						return
					}
					_ = commitBudgetOrDebit(context.Background(), h.logger, h.billing, h.wallets, reservation, authInfo.TenantID, finalized, request.RequestID, request.Deployment.BillingMode)
					return
				}
				if len(chunk.Data) > 0 {
					_, _ = writer.Write(chunk.Data)
					_ = writer.Flush()
					if parsed := parseAnthropicSSEUsage(chunk.Data); parsed != nil {
						usage = parsed
					}
					if anthropicSSEIsError(chunk.Data) {
						failed = true
					}
				}
			case streamErr, ok := <-stream.Errors():
				if ok && streamErr != nil {
					failed = true
					releaseBudgetBestEffort(context.Background(), h.logger, h.billing, reservation, "stream_error")
				}
				return
			}
		}
	})
	return nil
}

func writeAnthropicError(c *fiber.Ctx, status int, errorType, message string) error {
	requestID := strings.TrimSpace(c.GetRespHeader("request-id"))
	return c.Status(status).JSON(fiber.Map{
		"type":       "error",
		"error":      fiber.Map{"type": errorType, "message": message},
		"request_id": requestID,
	})
}

func writeAnthropicBillingError(c *fiber.Ctx, err error) error {
	if strings.Contains(err.Error(), "pricing_unavailable") {
		return writeAnthropicError(c, fiber.StatusServiceUnavailable, "pricing_unavailable", "Anthropic pricing is unavailable")
	}
	return writeAnthropicError(c, fiber.StatusBadGateway, "billing_error", "Failed to calculate usage cost")
}

func parseAnthropicNativeUsage(body []byte) *contractsopenai.Usage {
	var response struct {
		Usage struct {
			InputTokens              int `json:"input_tokens"`
			OutputTokens             int `json:"output_tokens"`
			CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
			CacheReadInputTokens     int `json:"cache_read_input_tokens"`
			CacheCreation            struct {
				FiveMinutes int `json:"ephemeral_5m_input_tokens"`
				OneHour     int `json:"ephemeral_1h_input_tokens"`
			} `json:"cache_creation"`
			ServerToolUse map[string]int `json:"server_tool_use"`
		} `json:"usage"`
	}
	if json.Unmarshal(body, &response) != nil {
		return nil
	}
	return &contractsopenai.Usage{
		PromptTokens: response.Usage.InputTokens, CompletionTokens: response.Usage.OutputTokens,
		TotalTokens:                response.Usage.InputTokens + response.Usage.OutputTokens,
		CacheCreationInputTokens:   response.Usage.CacheCreationInputTokens,
		CacheCreation5mInputTokens: response.Usage.CacheCreation.FiveMinutes,
		CacheCreation1hInputTokens: response.Usage.CacheCreation.OneHour,
		CacheReadInputTokens:       response.Usage.CacheReadInputTokens,
		ServerToolUse:              response.Usage.ServerToolUse,
	}
}

func parseAnthropicSSEUsage(data []byte) *contractsopenai.Usage {
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		if usage := parseAnthropicNativeUsage([]byte(strings.TrimSpace(strings.TrimPrefix(line, "data:")))); usage != nil &&
			(usage.TotalTokens > 0 || usage.CacheCreationInputTokens > 0 || usage.CacheReadInputTokens > 0) {
			return usage
		}
	}
	return nil
}

func anthropicSSEIsError(data []byte) bool {
	text := string(data)
	return strings.Contains(text, `"type":"error"`) || strings.Contains(text, `"type": "error"`)
}
