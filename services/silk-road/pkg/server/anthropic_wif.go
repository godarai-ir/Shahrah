package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

var errAnthropicWIFUnavailable = errors.New("anthropic_wif_unavailable")

type anthropicWIFToken struct {
	accessToken string
	expiresAt   time.Time
	refreshedAt time.Time
	scopes      []string
}

type anthropicWIFManager struct {
	mu       sync.Mutex
	sources  map[string]string
	cache    map[string]anthropicWIFToken
	inflight map[string]chan struct{}
	now      func() time.Time
}

type anthropicWIFExchangeResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	Scope       string `json:"scope"`
}

func newAnthropicWIFManagerFromEnv() (*anthropicWIFManager, error) {
	if !silkRoadFeatureEnabled("ANTHROPIC_WIF_ENABLED", true) {
		return nil, errAnthropicWIFUnavailable
	}

	raw := strings.TrimSpace(os.Getenv("ANTHROPIC_WIF_SOURCES_JSON"))
	if raw == "" {
		return &anthropicWIFManager{
			sources:  map[string]string{},
			cache:    map[string]anthropicWIFToken{},
			inflight: map[string]chan struct{}{},
			now:      time.Now,
		}, nil
	}

	var sources map[string]string
	if err := json.Unmarshal([]byte(raw), &sources); err != nil {
		return nil, fmt.Errorf("%w: invalid source map", errAnthropicWIFUnavailable)
	}

	for id, path := range sources {
		if strings.TrimSpace(id) == "" || !filepath.IsAbs(path) {
			return nil, fmt.Errorf(
				"%w: every source must use an absolute allowlisted path",
				errAnthropicWIFUnavailable,
			)
		}

		sources[id] = filepath.Clean(path)
	}

	return &anthropicWIFManager{
		sources:  sources,
		cache:    map[string]anthropicWIFToken{},
		inflight: map[string]chan struct{}{},
		now:      time.Now,
	}, nil
}

func silkRoadFeatureEnabled(name string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}

	value, err := strconv.ParseBool(raw)

	return err == nil && value
}

func (m *anthropicWIFManager) token(
	ctx context.Context,
	client *http.Client,
	baseURL string,
	connectionID string,
	config *platform.ResolvedAnthropicWIFAuth,
) (anthropicWIFToken, error) {
	if m == nil || config == nil {
		return anthropicWIFToken{}, errAnthropicWIFUnavailable
	}

	cacheKey := anthropicWIFCacheKey(connectionID, config)

	for {
		m.mu.Lock()
		refresh := false

		if cached, ok := m.cache[cacheKey]; ok {
			if cached.expiresAt.After(m.now().Add(60 * time.Second)) {
				m.mu.Unlock()
				return cached, nil
			}

			refresh = true
		}

		if wait, ok := m.inflight[cacheKey]; ok {
			m.mu.Unlock()

			select {
			case <-ctx.Done():
				return anthropicWIFToken{}, ctx.Err()
			case <-wait:
				continue
			}
		}

		wait := make(chan struct{})
		m.inflight[cacheKey] = wait
		m.mu.Unlock()

		token, err := m.exchange(ctx, client, baseURL, config)
		m.mu.Lock()
		if err == nil {
			m.cache[cacheKey] = token

			anthropicWIFExchangeSuccess.Add(1)

			if refresh {
				anthropicWIFRefreshTotal.Add(1)
			}
		} else {
			anthropicWIFExchangeFailure.Add(1)
		}

		close(wait)
		delete(m.inflight, cacheKey)
		m.mu.Unlock()

		return token, err
	}
}

func (m *anthropicWIFManager) exchange(
	ctx context.Context,
	client *http.Client,
	baseURL string,
	config *platform.ResolvedAnthropicWIFAuth,
) (anthropicWIFToken, error) {
	sourcePath, ok := m.sources[config.TokenSourceID]
	if !ok {
		return anthropicWIFToken{}, errAnthropicWIFUnavailable
	}

	file, err := os.Open(sourcePath)
	if err != nil {
		return anthropicWIFToken{}, errAnthropicWIFUnavailable
	}

	assertionBytes, readErr := io.ReadAll(io.LimitReader(file, 1024*1024+1))
	closeErr := file.Close()

	if readErr != nil || closeErr != nil || len(assertionBytes) == 0 || len(assertionBytes) > 1024*1024 {
		return anthropicWIFToken{}, errAnthropicWIFUnavailable
	}

	assertion := strings.TrimSpace(string(assertionBytes))
	clear(assertionBytes)

	if assertion == "" {
		return anthropicWIFToken{}, errAnthropicWIFUnavailable
	}

	form := url.Values{
		"grant_type": {
			"urn:ietf:params:oauth:grant-type:jwt-bearer",
		},
		"assertion":          {assertion},
		"federation_rule_id": {config.FederationRuleID},
		"organization_id":    {config.OrganizationID},
		"service_account_id": {config.ServiceAccountID},
	}
	if config.WorkspaceID != "" {
		form.Set("workspace_id", config.WorkspaceID)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		joinURL(baseURL, "/v1/oauth/token"),
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return anthropicWIFToken{}, errAnthropicWIFUnavailable
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return anthropicWIFToken{}, errAnthropicWIFUnavailable
	}
	defer closeBody(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if _, copyErr := io.Copy(io.Discard, io.LimitReader(resp.Body, 64*1024)); copyErr != nil {
			return anthropicWIFToken{}, errAnthropicWIFUnavailable
		}

		return anthropicWIFToken{}, errAnthropicWIFUnavailable
	}

	var payload anthropicWIFExchangeResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1024*1024)).Decode(&payload); err != nil ||
		strings.TrimSpace(payload.AccessToken) == "" ||
		payload.ExpiresIn <= 60 {
		return anthropicWIFToken{}, errAnthropicWIFUnavailable
	}

	return anthropicWIFToken{
		accessToken: payload.AccessToken,
		expiresAt:   m.now().Add(time.Duration(payload.ExpiresIn) * time.Second),
		refreshedAt: m.now(),
		scopes:      strings.Fields(payload.Scope),
	}, nil
}

func anthropicWIFCacheKey(
	connectionID string,
	config *platform.ResolvedAnthropicWIFAuth,
) string {
	if config == nil {
		return ""
	}

	return strings.Join([]string{
		connectionID,
		config.FederationRuleID,
		config.WorkspaceID,
		config.TokenSourceID,
	}, "\x00")
}

func (m *anthropicWIFManager) status(
	connectionID string,
	config *platform.ResolvedAnthropicWIFAuth,
) (anthropicWIFToken, bool) {
	if m == nil || config == nil {
		return anthropicWIFToken{}, false
	}

	cacheKey := anthropicWIFCacheKey(connectionID, config)

	m.mu.Lock()
	defer m.mu.Unlock()

	token, ok := m.cache[cacheKey]
	if !ok {
		return anthropicWIFToken{}, false
	}

	token.accessToken = ""
	token.scopes = append([]string(nil), token.scopes...)

	return token, true
}
