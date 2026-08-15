package server

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

const (
	deepSeekDefaultBaseURL = "https://api.deepseek.com"
	deepSeekProviderName   = "deepseek"
	deepSeekFlashModelID   = "deepseek-v4-flash"
	deepSeekProModelID     = "deepseek-v4-pro"
)

var deepSeekUserIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,512}$`)

type deepSeekAdapter struct {
	compat *openAICompatibleAdapter
}

func newDeepSeekAdapter() providerAdapter {
	return &deepSeekAdapter{compat: &openAICompatibleAdapter{
		defaultBaseURL: deepSeekDefaultBaseURL,
		authHeader:     "Authorization",
	}}
}

func (a *deepSeekAdapter) Dispatch(
	ctx context.Context,
	client *http.Client,
	payload platform.TunnelRequest,
) (*adapterResponse, error) {
	decodedBody, err := validateDeepSeekDispatchBody(payload)
	if err != nil {
		return nil, err
	}

	if payload.Operation == platform.ProviderOperationChatCompletions ||
		payload.Operation == platform.ProviderOperationOpenAIResponsesCreate {
		body, err := deepSeekRequestBodyFromDecoded(payload, decodedBody)
		if err != nil {
			return nil, err
		}

		payload.Body = body
	}

	attempts := 1
	if payload.Operation == platform.ProviderOperationOpenAIModelsList {
		attempts = 2
	}

	for attempt := 1; attempt <= attempts; attempt++ {
		response, err := a.compat.Dispatch(ctx, client, payload)
		if err != nil || response == nil || attempt == attempts || !deepSeekRetryableStatus(response.StatusCode) {
			return response, err
		}

		if _, copyErr := io.Copy(io.Discard, io.LimitReader(response.Body, 64*1024)); copyErr != nil {
			return nil, fmt.Errorf("deepseek_retry_response_drain: %w", copyErr)
		}

		if closeErr := response.Body.Close(); closeErr != nil {
			return nil, fmt.Errorf("deepseek_retry_response_close: %w", closeErr)
		}

		timer := time.NewTimer(deepSeekRetryDelay())
		select {
		case <-ctx.Done():
			timer.Stop()

			return nil, ctx.Err()
		case <-timer.C:
		}
	}

	return nil, fmt.Errorf("deepseek_retry_exhausted")
}

func validateDeepSeekDispatch(payload platform.TunnelRequest) error {
	_, err := validateDeepSeekDispatchBody(payload)

	return err
}

func validateDeepSeekDispatchBody(payload platform.TunnelRequest) (map[string]json.RawMessage, error) {
	switch payload.Operation {
	case platform.ProviderOperationChatCompletions,
		platform.ProviderOperationOpenAIResponsesCreate,
		platform.ProviderOperationOpenAIModelsList:
	default:
		return nil, fmt.Errorf("deepseek_unsupported_operation: %s", payload.Operation)
	}

	if err := validateDeepSeekBaseURL(payload.Deployment.APIBase); err != nil {
		return nil, err
	}

	if payload.Operation == platform.ProviderOperationOpenAIModelsList {
		return nil, nil
	}

	body, err := decodeDeepSeekRequestBody(payload.Body)
	if err != nil {
		return nil, err
	}

	model := deepSeekRequestModelFromDecoded(payload, body)
	if payload.Operation == platform.ProviderOperationOpenAIResponsesCreate &&
		model != deepSeekFlashModelID && model != deepSeekProModelID {
		return nil, fmt.Errorf("deepseek_unsupported_responses_model: %s", model)
	}

	if payload.Operation == platform.ProviderOperationChatCompletions &&
		model != deepSeekFlashModelID && model != deepSeekProModelID {
		return nil, fmt.Errorf("deepseek_unsupported_model: %s", model)
	}

	endpoint := "chat_completions"
	if payload.Operation == platform.ProviderOperationOpenAIResponsesCreate {
		endpoint = "responses"
	}

	binding, ok := platform.FindModelRuntimeProfile(payload.ModelRuntimeProfiles, model)
	if !ok || binding.Validate() != nil || !strings.EqualFold(binding.Provider, deepSeekProviderName) ||
		!binding.Profile.AllowsEndpoint(endpoint) {
		return nil, fmt.Errorf("deepseek_runtime_profile_missing: %s", model)
	}

	transport := "http"
	if payload.Stream {
		transport = "sse"
	}

	if !binding.Profile.AllowsTransport(transport) {
		return nil, fmt.Errorf("deepseek_unsupported_transport: %s", transport)
	}

	return body, nil
}

func validateDeepSeekBaseURL(raw string) error {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil
	}

	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() != "api.deepseek.com" ||
		parsed.Port() != "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" ||
		(parsed.EscapedPath() != "" && parsed.EscapedPath() != "/") {
		return fmt.Errorf("deepseek_invalid_api_base")
	}

	return nil
}

func deepSeekRequestBody(payload platform.TunnelRequest) ([]byte, error) {
	body, err := decodeDeepSeekRequestBody(payload.Body)
	if err != nil {
		return nil, err
	}

	return deepSeekRequestBodyFromDecoded(payload, body)
}

func decodeDeepSeekRequestBody(bodyBytes []byte) (map[string]json.RawMessage, error) {
	var body map[string]json.RawMessage
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		return nil, fmt.Errorf("deepseek_invalid_json: %w", err)
	}

	if body == nil {
		return nil, fmt.Errorf("deepseek_invalid_json: request body must be an object")
	}

	return body, nil
}

func deepSeekRequestBodyFromDecoded(payload platform.TunnelRequest, body map[string]json.RawMessage) ([]byte, error) {
	identityField := "user_id"
	reservedFields := []string{"user_id"}

	if payload.Operation == platform.ProviderOperationOpenAIResponsesCreate {
		identityField = "user"
		reservedFields = []string{"user", "user_id"}
	}

	for _, field := range reservedFields {
		if _, exists := body[field]; exists {
			return nil, fmt.Errorf("deepseek_%s_is_gateway_managed", field)
		}
	}

	if payload.Deployment.CredentialSource == platform.CredentialSourceUserBYOK {
		return payload.Body, nil
	}

	userID, err := deepSeekOpaqueUserID(
		payload.TenantID,
		payload.Deployment.ProviderAccountID,
		os.Getenv("DEEPSEEK_USER_ID_HMAC_SECRET"),
	)
	if err != nil {
		return nil, err
	}

	encoded, err := json.Marshal(userID)
	if err != nil {
		return nil, fmt.Errorf("deepseek_user_id_encode: %w", err)
	}

	body[identityField] = encoded

	out, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("deepseek_request_encode: %w", err)
	}

	return out, nil
}

func deepSeekOpaqueUserID(tenantID, accountID, secret string) (string, error) {
	tenantID = strings.TrimSpace(tenantID)
	accountID = strings.TrimSpace(accountID)

	secret = strings.TrimSpace(secret)
	if tenantID == "" || accountID == "" || secret == "" {
		return "", fmt.Errorf("deepseek_platform_managed_isolation_unavailable")
	}

	mac := hmac.New(sha256.New, []byte(secret))
	if _, err := mac.Write([]byte(tenantID + "\x00" + accountID)); err != nil {
		return "", fmt.Errorf("deepseek_user_id_hmac: %w", err)
	}

	userID := "gdr_" + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !deepSeekUserIDPattern.MatchString(userID) {
		return "", fmt.Errorf("deepseek_user_id_invalid")
	}

	return userID, nil
}

func deepSeekRequestModelFromDecoded(payload platform.TunnelRequest, body map[string]json.RawMessage) string {
	model := strings.ToLower(strings.TrimSpace(payload.Deployment.Model))

	var requestedModel string
	if json.Unmarshal(body["model"], &requestedModel) == nil && strings.TrimSpace(requestedModel) != "" {
		model = strings.ToLower(strings.TrimSpace(requestedModel))
	}

	return model
}

func deepSeekRetryableStatus(status int) bool {
	return status == http.StatusTooManyRequests ||
		status == http.StatusInternalServerError ||
		status == http.StatusServiceUnavailable
}

func deepSeekRetryDelay() time.Duration {
	const (
		minimum = 75 * time.Millisecond
		jitter  = 50 * time.Millisecond
	)

	return minimum + time.Duration(time.Now().UnixNano()%int64(jitter))
}
