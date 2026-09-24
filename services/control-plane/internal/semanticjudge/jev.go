// Package semanticjudge calls TypeSafe Jev to answer the typed questions of semantic_policy guardrails.
package semanticjudge

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/guardrails"
)

const (
	systemOnePath       = "/v1/systemone"
	statusOverloaded    = 529
	maxResponseBodySize = 1 << 20
	minRetryBudget      = 200 * time.Millisecond
	retryBackoff        = 50 * time.Millisecond
)

var (
	// ErrCircuitOpen short-circuits calls after repeated failures so requests do not wait on a failing judge.
	ErrCircuitOpen = errors.New("semantic judge circuit is open")

	errInvalidConfig = errors.New("invalid semantic judge config")
	errStatus        = errors.New("semantic judge returned a non-success status")
	// errRejected marks a 4xx answer about this request's content or rules; it says nothing about judge health.
	errRejected = errors.New("semantic judge rejected the request")
)

// Config is the typed runtime configuration loaded by bootstrap.
type Config struct {
	BaseURL          string
	APIKey           string
	Model            string
	Timeout          time.Duration
	CacheTTL         time.Duration
	CacheKeyPrefix   string
	FailureThreshold int
	Cooldown         time.Duration
}

// Store is the subset of the gateway cache the client needs.
type Store interface {
	Get(ctx context.Context, key string) (interface{}, bool)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration)
}

// Client implements guardrails.SemanticJudge against the TypeSafe System One API.
type Client struct {
	cfg        Config
	httpClient *http.Client
	store      Store
	now        func() time.Time

	mu        sync.Mutex
	failures  int
	openUntil time.Time
}

type apiQuestion struct {
	Type         string `json:"type"`
	Instructions string `json:"instructions"`
	Criteria     any    `json:"criteria,omitempty"`
}

type apiRequest struct {
	State     map[string]any         `json:"state"`
	Model     string                 `json:"model"`
	Questions map[string]apiQuestion `json:"questions"`
}

type apiAnswer struct {
	Noul  *float64 `json:"noul"`
	Score *float64 `json:"score"`
}

type apiResponse struct {
	Model   string               `json:"model"`
	Answers map[string]apiAnswer `json:"answers"`
}

type cachedResponse struct {
	Model   string             `json:"model"`
	Answers map[string]float64 `json:"answers"`
}

// New validates the config. A nil store disables caching.
func New(httpClient *http.Client, store Store, cfg Config) (*Client, error) {
	switch {
	case strings.TrimSpace(cfg.BaseURL) == "":
		return nil, fmt.Errorf("%w: base URL is required", errInvalidConfig)
	case strings.TrimSpace(cfg.APIKey) == "":
		return nil, fmt.Errorf("%w: API key is required", errInvalidConfig)
	case strings.TrimSpace(cfg.Model) == "":
		return nil, fmt.Errorf("%w: model is required", errInvalidConfig)
	case cfg.Timeout <= 0:
		return nil, fmt.Errorf("%w: timeout must be positive", errInvalidConfig)
	}

	if httpClient == nil {
		httpClient = &http.Client{}
	}

	return &Client{cfg: cfg, httpClient: httpClient, store: store, now: time.Now}, nil
}

// Judge answers every question in one call. Identical requests within the cache TTL are served from the cache.
func (c *Client) Judge(ctx context.Context, request guardrails.JudgeRequest) (guardrails.JudgeResponse, error) {
	body, err := json.Marshal(c.apiRequest(request))
	if err != nil {
		return guardrails.JudgeResponse{}, fmt.Errorf("encode judge request: %w", err)
	}

	cacheKey := c.cacheKey(body)
	if cached, ok := c.cached(ctx, cacheKey); ok {
		return cached, nil
	}

	if c.circuitOpen() {
		return guardrails.JudgeResponse{}, ErrCircuitOpen
	}

	callCtx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	response, err := c.callWithRetry(callCtx, body)
	c.recordResult(err)

	if err != nil {
		return guardrails.JudgeResponse{}, err
	}

	c.remember(ctx, cacheKey, response)

	return response, nil
}

func (c *Client) apiRequest(request guardrails.JudgeRequest) apiRequest {
	questions := make(map[string]apiQuestion, len(request.Questions))

	for key, question := range request.Questions {
		if question.Kind == guardrails.JudgeQuestionScore {
			questions[key] = apiQuestion{Type: "score", Instructions: question.Instructions, Criteria: question.Levels}

			continue
		}

		item := apiQuestion{Type: "noul", Instructions: question.Instructions}
		trueSide := criterion(question.TrueWhen, question.ExamplesTrue)
		falseSide := criterion(question.FalseWhen, question.ExamplesFalse)

		if trueSide != nil || falseSide != nil {
			criteria := map[string]any{}
			if trueSide != nil {
				criteria["true"] = trueSide
			}

			if falseSide != nil {
				criteria["false"] = falseSide
			}

			item.Criteria = criteria
		}

		questions[key] = item
	}

	return apiRequest{State: request.State, Model: c.cfg.Model, Questions: questions}
}

// criterion uses TypeSafe's structured criteria form only when there are examples to carry.
func criterion(definition string, examples []string) any {
	switch {
	case len(examples) > 0:
		return map[string]any{"definition": definition, "examples": examples}
	case definition != "":
		return definition
	default:
		return nil
	}
}

func (c *Client) callWithRetry(ctx context.Context, body []byte) (guardrails.JudgeResponse, error) {
	response, retryable, err := c.call(ctx, body)
	if err == nil || !retryable {
		return response, err
	}

	if deadline, ok := ctx.Deadline(); ok && time.Until(deadline) < minRetryBudget {
		return response, err
	}

	timer := time.NewTimer(retryBackoff)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return guardrails.JudgeResponse{}, errors.Join(err, ctx.Err())
	case <-timer.C:
	}

	response, _, err = c.call(ctx, body)

	return response, err
}

func (c *Client) call(ctx context.Context, body []byte) (guardrails.JudgeResponse, bool, error) {
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.cfg.BaseURL, "/")+systemOnePath, bytes.NewReader(body))
	if err != nil {
		return guardrails.JudgeResponse{}, false, fmt.Errorf("build judge request: %w", err)
	}

	httpRequest.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	httpRequest.Header.Set("Content-Type", "application/json")

	httpResponse, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return guardrails.JudgeResponse{}, ctx.Err() == nil, fmt.Errorf("call judge: %w", err)
	}

	responseBody, err := io.ReadAll(io.LimitReader(httpResponse.Body, maxResponseBodySize))
	if closeErr := httpResponse.Body.Close(); closeErr != nil {
		err = errors.Join(err, fmt.Errorf("close judge response: %w", closeErr))
	}

	if err != nil {
		return guardrails.JudgeResponse{}, true, fmt.Errorf("read judge response: %w", err)
	}

	if httpResponse.StatusCode != http.StatusOK {
		retryable := httpResponse.StatusCode == http.StatusTooManyRequests || httpResponse.StatusCode == statusOverloaded ||
			httpResponse.StatusCode >= http.StatusInternalServerError

		if !retryable {
			return guardrails.JudgeResponse{}, false, fmt.Errorf("%w: %w: %d", errStatus, errRejected, httpResponse.StatusCode)
		}

		return guardrails.JudgeResponse{}, true, fmt.Errorf("%w: %d", errStatus, httpResponse.StatusCode)
	}

	var decoded apiResponse
	if err = json.Unmarshal(responseBody, &decoded); err != nil {
		return guardrails.JudgeResponse{}, false, fmt.Errorf("decode judge response: %w", err)
	}

	answers := make(map[string]float64, len(decoded.Answers))

	for key, answer := range decoded.Answers {
		switch {
		case answer.Noul != nil:
			answers[key] = *answer.Noul
		case answer.Score != nil:
			answers[key] = *answer.Score
		}
	}

	return guardrails.JudgeResponse{Answers: answers, Model: decoded.Model}, false, nil
}

func (c *Client) cacheKey(body []byte) string {
	sum := sha256.Sum256(body)

	return c.cfg.CacheKeyPrefix + hex.EncodeToString(sum[:])
}

func (c *Client) cached(ctx context.Context, key string) (guardrails.JudgeResponse, bool) {
	if c.store == nil || c.cfg.CacheTTL <= 0 {
		return guardrails.JudgeResponse{}, false
	}

	value, ok := c.store.Get(ctx, key)
	if !ok {
		return guardrails.JudgeResponse{}, false
	}

	raw, ok := value.(string)
	if !ok {
		return guardrails.JudgeResponse{}, false
	}

	var entry cachedResponse
	if err := json.Unmarshal([]byte(raw), &entry); err != nil {
		return guardrails.JudgeResponse{}, false
	}

	return guardrails.JudgeResponse{Answers: entry.Answers, Model: entry.Model, Cached: true}, true
}

func (c *Client) remember(ctx context.Context, key string, response guardrails.JudgeResponse) {
	if c.store == nil || c.cfg.CacheTTL <= 0 {
		return
	}

	raw, err := json.Marshal(cachedResponse{Model: response.Model, Answers: response.Answers})
	if err != nil {
		return
	}

	c.store.Set(ctx, key, string(raw), c.cfg.CacheTTL)
}

func (c *Client) circuitOpen() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.now().Before(c.openUntil)
}

// recordResult feeds the shared breaker. A rejected request is one tenant's problem, so it must not open the circuit
// for every tenant.
func (c *Client) recordResult(err error) {
	if errors.Is(err, errRejected) {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if err == nil {
		c.failures = 0

		return
	}

	c.failures++
	if c.cfg.FailureThreshold > 0 && c.failures >= c.cfg.FailureThreshold {
		c.openUntil = c.now().Add(c.cfg.Cooldown)
		c.failures = 0
	}
}
