package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestAnthropicWIFExchangesOnceAndUsesBearerToken(t *testing.T) {
	t.Parallel()
	assertionPath := filepath.Join(t.TempDir(), "assertion.jwt")
	if err := os.WriteFile(assertionPath, []byte("signed-assertion"), 0o600); err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	exchanges := 0
	messages := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		switch r.URL.Path {
		case "/v1/oauth/token":
			if err := r.ParseForm(); err != nil {
				t.Error(err)
			}
			if r.Form.Get("assertion") != "signed-assertion" ||
				r.Form.Get("federation_rule_id") != "rule-1" {
				t.Errorf("unexpected token exchange form")
			}
			mu.Lock()
			exchanges++
			mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"access_token":"access-secret","expires_in":3600,"scope":"workspace:inference workspace:developer"}`)
		case "/v1/messages":
			if r.Header.Get("Authorization") != "Bearer access-secret" {
				t.Errorf("unexpected authorization header")
			}
			if r.Header.Get("x-api-key") != "" {
				t.Errorf("virtual or provider API key must not be sent in WIF mode")
			}
			mu.Lock()
			messages++
			mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"id":"msg_1","type":"message","role":"assistant","content":[{"type":"text","text":"ok"}],"model":"claude-sonnet-4-6","usage":{"input_tokens":1,"output_tokens":1},"stop_reason":"end_turn"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer upstream.Close()

	manager := &anthropicWIFManager{
		sources: map[string]string{
			"workload": assertionPath,
		},
		cache:    map[string]anthropicWIFToken{},
		inflight: map[string]chan struct{}{},
		now:      time.Now,
	}
	adapter := &anthropicAdapter{wif: manager}
	payload := platform.TunnelRequest{
		Deployment: platform.Deployment{
			Provider:          "anthropic",
			Model:             "claude-sonnet-4-6",
			APIBase:           upstream.URL,
			ProviderAccountID: "account-1",
			ProviderAuth: &platform.ResolvedProviderAuth{
				Mode: string(platform.ProviderAuthModeAnthropicWIF),
				AnthropicWIF: &platform.ResolvedAnthropicWIFAuth{
					FederationRuleID: "rule-1",
					OrganizationID:   "org-1",
					ServiceAccountID: "service-1",
					WorkspaceID:      "workspace-1",
					TokenSourceID:    "workload",
				},
			},
		},
		ModelRuntimeProfiles: []platform.ModelRuntimeProfileBinding{
			anthropicTestRuntimeProfile("claude-sonnet-4-6"),
		},
		Body: []byte(`{"model":"claude-sonnet-4-6","messages":[{"role":"user","content":"hi"}]}`),
	}

	for range 2 {
		response, err := adapter.Dispatch(
			context.Background(),
			upstream.Client(),
			payload,
		)
		if err != nil {
			t.Fatal(err)
		}
		_ = response.Body.Close()
	}

	mu.Lock()
	defer mu.Unlock()
	if exchanges != 1 || messages != 2 {
		t.Fatalf("exchanges=%d messages=%d", exchanges, messages)
	}
}
