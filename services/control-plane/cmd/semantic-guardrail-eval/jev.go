package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	systemOnePath       = "/v1/systemone"
	statusOverloaded    = 529
	maxResponseBodySize = 1 << 20
)

var errJevStatus = errors.New("jev returned a non-success status")

// noulQuestion is a TypeSafe yes/no question.
type noulQuestion struct {
	Type         string            `json:"type"`
	Instructions string            `json:"instructions"`
	Criteria     map[string]string `json:"criteria,omitempty"`
}

// systemOneRequest is the body of POST /v1/systemone.
type systemOneRequest struct {
	State     map[string]any          `json:"state"`
	Model     string                  `json:"model"`
	Questions map[string]noulQuestion `json:"questions"`
}

// systemOneAnswer is one typed answer. Only Noul answers are requested by this tool.
type systemOneAnswer struct {
	Type string   `json:"type"`
	Noul *float64 `json:"noul"`
}

// systemOneResponse is the body returned by POST /v1/systemone.
type systemOneResponse struct {
	Model   string                     `json:"model"`
	Answers map[string]systemOneAnswer `json:"answers"`
	Usage   struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// jevClient calls the TypeSafe System One API with bounded retries on 429 and 529.
type jevClient struct {
	baseURL     string
	apiKey      string
	httpClient  *http.Client
	maxAttempts int
	backoff     time.Duration
}

// buildRequest batches every rule of one target into a single request, the same way the gateway will.
func buildRequest(model, lang string, rules []rule, s sample) systemOneRequest {
	questions := make(map[string]noulQuestion, len(rules))

	for _, r := range rules {
		questions[r.Key] = noulQuestion{
			Type:         "noul",
			Instructions: r.Statement.in(lang),
			Criteria: map[string]string{
				"true":  r.TrueWhen.in(lang),
				"false": r.FalseWhen.in(lang),
			},
		}
	}

	role := "user"
	if s.Target == targetOutput {
		role = "assistant"
	}

	state := map[string]any{"role": role, "untrusted_content": s.Text}
	if len(s.Turns) > 0 {
		state["prior_turns"] = s.Turns
	}

	return systemOneRequest{
		State:     state,
		Model:     model,
		Questions: questions,
	}
}

func (c *jevClient) evaluate(ctx context.Context, request systemOneRequest) (systemOneResponse, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return systemOneResponse{}, fmt.Errorf("encode request: %w", err)
	}

	var lastErr error

	for attempt := range c.maxAttempts {
		if attempt > 0 {
			if sleepErr := sleepContext(ctx, c.backoff*time.Duration(1<<(attempt-1))); sleepErr != nil {
				return systemOneResponse{}, sleepErr
			}
		}

		response, retryable, callErr := c.call(ctx, body)
		if callErr == nil {
			return response, nil
		}

		lastErr = callErr

		if !retryable {
			break
		}
	}

	return systemOneResponse{}, lastErr
}

func (c *jevClient) call(ctx context.Context, body []byte) (systemOneResponse, bool, error) {
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.baseURL, "/")+systemOnePath, bytes.NewReader(body))
	if err != nil {
		return systemOneResponse{}, false, fmt.Errorf("build request: %w", err)
	}

	httpRequest.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpRequest.Header.Set("Content-Type", "application/json")

	httpResponse, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return systemOneResponse{}, ctx.Err() == nil, fmt.Errorf("call jev: %w", err)
	}

	responseBody, err := io.ReadAll(io.LimitReader(httpResponse.Body, maxResponseBodySize))
	if closeErr := httpResponse.Body.Close(); closeErr != nil {
		err = errors.Join(err, fmt.Errorf("close response: %w", closeErr))
	}

	if err != nil {
		return systemOneResponse{}, true, fmt.Errorf("read response: %w", err)
	}

	if httpResponse.StatusCode != http.StatusOK {
		retryable := httpResponse.StatusCode == http.StatusTooManyRequests || httpResponse.StatusCode == statusOverloaded

		return systemOneResponse{}, retryable, fmt.Errorf("%w: %d %s", errJevStatus, httpResponse.StatusCode, truncate(string(responseBody), 300))
	}

	var decoded systemOneResponse
	if err = json.Unmarshal(responseBody, &decoded); err != nil {
		return systemOneResponse{}, false, fmt.Errorf("decode response: %w", err)
	}

	return decoded, false, nil
}

func sleepContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func truncate(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}

	return string(runes[:limit]) + "..."
}
