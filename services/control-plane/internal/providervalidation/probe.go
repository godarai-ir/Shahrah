package providervalidation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	defaultTimeout  = 6 * time.Second
	maxResponseBody = 4096
	maxErrorDetail  = 180
)

var (
	ErrInvalidConfiguration = errors.New("invalid provider validation configuration")
	ErrMissingCredential    = errors.New("missing provider credential")
	bearerTokenPattern      = regexp.MustCompile(`(?i)Bearer\s+[A-Za-z0-9._~+/=-]+`)
	apiKeyPattern           = regexp.MustCompile(`(?i)(?:sk|gsk|AIza|mistral|together|cohere|xai)[-_][A-Za-z0-9._~+/=-]{8,}`)
)

// Request is the provider-neutral input to a credential probe.
type Request struct {
	Provider string
	AuthMode string
	Config   map[string]interface{}
	Secrets  map[string]string
}

// Result is the stable internal credential-validation result.
type Result struct {
	Success     bool
	Unsupported bool
	Category    string
	Message     string
}

type probeSpec struct {
	url     string
	headers map[string]string
}

// Prober executes provider validation requests.
type Prober struct {
	client  *http.Client
	timeout time.Duration
}

// NewProber returns a credential prober with a bounded HTTP timeout.
func NewProber(client *http.Client) *Prober {
	if client == nil {
		client = &http.Client{Timeout: defaultTimeout}
	}

	safeClient := *client
	safeClient.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		// The validated provider URL is the only authorized network target.
		// Returning ErrUseLastResponse exposes the 3xx to the classifier without
		// following a provider-controlled redirect to an unapproved host.
		return http.ErrUseLastResponse
	}

	return &Prober{client: &safeClient, timeout: defaultTimeout}
}

// Probe validates a credential without issuing a billable inference request.
func (p *Prober) Probe(ctx context.Context, request Request) Result {
	request.Provider = NormalizeProvider(request.Provider)
	request.AuthMode = strings.ToLower(strings.TrimSpace(request.AuthMode))

	spec, supported, err := buildProbe(request)
	if !supported {
		return Result{
			Unsupported: true,
			Category:    "unsupported",
			Message:     "Automatic credential validation is not available for this provider authentication mode.",
		}
	}
	if err != nil {
		switch {
		case errors.Is(err, ErrMissingCredential):
			return Result{Category: "missing_credential", Message: "Required provider credential is missing or masked."}
		default:
			return Result{Category: "configuration", Message: "Provider validation configuration is invalid."}
		}
	}

	timeout := p.timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	probeContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(probeContext, http.MethodGet, spec.url, http.NoBody)
	if err != nil {
		return Result{Category: "configuration", Message: "Provider validation endpoint is invalid."}
	}

	for name, value := range spec.headers {
		req.Header.Set(name, value)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		if probeContext.Err() != nil {
			return Result{Category: "timeout", Message: "Provider validation timed out."}
		}

		return Result{Category: "network", Message: "Provider validation could not reach the upstream service."}
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
	if readErr != nil {
		return Result{Category: "network", Message: "Provider validation response could not be read."}
	}

	return classifyResponse(resp.StatusCode, body, request.Secrets)
}

func classifyResponse(status int, body []byte, secrets map[string]string) Result {
	detail := sanitizedProviderError(body, secrets)

	switch {
	case status >= http.StatusOK && status < http.StatusMultipleChoices:
		return Result{Success: true, Category: "ok", Message: "Provider accepted the credential."}
	case status == http.StatusUnauthorized:
		return Result{Category: "invalid_credential", Message: appendProviderDetail("Provider rejected the credential as invalid.", detail)}
	case status == http.StatusForbidden:
		return Result{Category: "forbidden", Message: appendProviderDetail("The credential authenticated but lacks permission for the validation endpoint.", detail)}
	case status == http.StatusPaymentRequired || status == http.StatusTooManyRequests:
		return Result{
			Success:  true,
			Category: "quota_exhausted",
			Message:  appendProviderDetail("Provider authenticated the credential, but quota, billing, or rate limits blocked the probe.", detail),
		}
	case status == http.StatusNotFound:
		return Result{Category: "configuration", Message: appendProviderDetail("The configured provider validation endpoint was not found.", detail)}
	case status >= http.StatusInternalServerError:
		return Result{Category: "provider", Message: appendProviderDetail("Provider validation endpoint returned a server error.", detail)}
	default:
		return Result{
			Category: "provider",
			Message:  appendProviderDetail(fmt.Sprintf("Provider validation failed with HTTP %d.", status), detail),
		}
	}
}

func appendProviderDetail(message, detail string) string {
	if strings.TrimSpace(detail) == "" {
		return message
	}

	return message + " " + detail
}

func sanitizedProviderError(body []byte, secrets map[string]string) string {
	text := strings.TrimSpace(string(body))
	if text == "" {
		return ""
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err == nil {
		text = firstProviderErrorMessage(payload)
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}

	for _, secret := range secrets {
		secret = strings.TrimSpace(secret)
		if secret != "" {
			text = strings.ReplaceAll(text, secret, "[redacted]")
		}
	}

	text = bearerTokenPattern.ReplaceAllString(text, "Bearer [redacted]")
	text = apiKeyPattern.ReplaceAllString(text, "[redacted]")
	if len(text) > maxErrorDetail {
		text = strings.TrimSpace(text[:maxErrorDetail]) + "..."
	}

	return text
}

func firstProviderErrorMessage(payload map[string]interface{}) string {
	if message, ok := payload["message"].(string); ok {
		return message
	}
	if detail, ok := payload["detail"].(string); ok {
		return detail
	}

	if errorValue, ok := payload["error"]; ok {
		switch typed := errorValue.(type) {
		case string:
			return typed
		case map[string]interface{}:
			if message, ok := typed["message"].(string); ok {
				return message
			}
			if detail, ok := typed["detail"].(string); ok {
				return detail
			}
			if code, ok := typed["code"].(string); ok {
				return code
			}
		}
	}

	return ""
}
