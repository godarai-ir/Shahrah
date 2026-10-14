package chatprovisioning

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// maxResponseBytes bounds how much of an answer is read.
const maxResponseBytes = 64 * 1024

// HTTPClient calls the chat provisioning API over HTTP.
type HTTPClient struct {
	baseURL string
	token   string
	http    *http.Client
}

// NewHTTPClient creates a client for the API at baseURL authenticated with token.
func NewHTTPClient(baseURL, token string, timeout time.Duration) (*HTTPClient, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, errors.New("chat provisioning: base URL must be an http(s) URL")
	}

	if token == "" {
		return nil, errors.New("chat provisioning: token is required")
	}

	return &HTTPClient{
		baseURL: parsed.String(),
		token:   token,
		http: &http.Client{
			Timeout: timeout,
			// The API never redirects; following one could send the token elsewhere.
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
	}, nil
}

type registerBody struct {
	Slug string       `json:"slug"`
	OIDC registerOIDC `json:"oidc"`
}

type registerOIDC struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret,omitempty"`
}

type stateResponse struct {
	Ready  bool   `json:"ready"`
	Reason string `json:"reason"`
}

type errorResponse struct {
	Error struct {
		Code string `json:"code"`
	} `json:"error"`
}

// Register creates the tenant in the chat, or brings an existing one to the described state.
func (c *HTTPClient) Register(ctx context.Context, tenantID string, request RegisterRequest) (Registration, error) {
	body := registerBody{
		Slug: request.Slug,
		OIDC: registerOIDC{ClientID: request.ClientID, ClientSecret: request.ClientSecret},
	}

	var state stateResponse
	if err := c.do(ctx, http.MethodPut, tenantID, body, &state); err != nil {
		return Registration{}, err
	}

	return Registration(state), nil
}

// SetStatus moves the tenant's chat between active and suspended.
func (c *HTTPClient) SetStatus(ctx context.Context, tenantID, status string) error {
	return c.do(ctx, http.MethodPatch, tenantID, map[string]string{"status": status}, nil)
}

func (c *HTTPClient) do(ctx context.Context, method, tenantID string, body, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encode chat provisioning request: %w", err)
	}

	endpoint := c.baseURL + "/tenants/" + url.PathEscape(tenantID)

	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("build chat provisioning request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		// The error text carries the URL but never the body or the token.
		return fmt.Errorf("call chat provisioning API: %w", err)
	}

	raw, readErr := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	closeErr := resp.Body.Close()

	if readErr != nil {
		return fmt.Errorf("read chat provisioning response: %w", readErr)
	}

	if closeErr != nil {
		return fmt.Errorf("close chat provisioning response: %w", closeErr)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		var failure errorResponse
		if json.Unmarshal(raw, &failure) != nil {
			failure = errorResponse{}
		}

		return &APIError{Status: resp.StatusCode, Code: failure.Error.Code}
	}

	if out == nil {
		return nil
	}

	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("decode chat provisioning response: %w", err)
	}

	return nil
}
