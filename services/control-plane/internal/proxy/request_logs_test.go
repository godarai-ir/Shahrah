package proxy

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/Delpak/shahrah/services/control-plane/internal/observability"
)

func TestBuildProviderHTTPErrorLogPayloadPreservesNonJSONBody(t *testing.T) {
	body := []byte("event: error\ndata: provider permission denied\n\n")
	payload := buildProviderHTTPErrorLogPayload(body, true)

	if got := payload["raw_body"]; got != string(body) {
		t.Fatalf("raw provider body was not preserved: got %q", got)
	}

	if got := payload["outcome"]; got != "provider_http_error" {
		t.Fatalf("unexpected outcome: %q", got)
	}

	if got := payload["stream"]; got != true {
		t.Fatalf("unexpected stream marker: %v", got)
	}
}

func BenchmarkBuildRequestLogPayload(b *testing.B) {
	body := append([]byte(`{"model":"gpt-test","prompt":"`), bytes.Repeat([]byte("x"), 64*1024)...)
	body = append(body, []byte(`"}`)...)

	b.Run("eager-json-decode", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(body)))

		for b.Loop() {
			payload := make(map[string]interface{})
			if err := json.Unmarshal(body, &payload); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("lazy-worker-decode", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(body)))

		for b.Loop() {
			payload := buildJSONRequestLogPayload(body, map[string]interface{}{"model": "gpt-test"})
			if len(payload) == 0 {
				b.Fatal("expected lazy payload")
			}
		}
	})
}

func TestBuildJSONRequestLogPayloadDefersDecoding(t *testing.T) {
	body := []byte(`{"model":"gpt-test","prompt":"large payload"}`)
	payload := buildJSONRequestLogPayload(body, map[string]interface{}{"model": "override"})

	raw, ok := payload[observability.LazyJSONPayloadField].(json.RawMessage)
	if !ok || string(raw) != string(body) {
		t.Fatalf("expected lazy raw JSON payload, got %#v", payload)
	}

	if payload["model"] != "override" {
		t.Fatalf("expected override to remain available")
	}
}

func TestMaterializeLazyLogPayloadHandlesJSONNull(t *testing.T) {
	payload := buildJSONRequestLogPayload([]byte("null"), map[string]interface{}{"model": "override"})

	materialized := materializeLazyLogPayload(payload)
	if materialized["model"] != "override" {
		t.Fatalf("expected override after null payload, got %#v", materialized)
	}
}
