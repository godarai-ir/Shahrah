package proxy

import "testing"

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
