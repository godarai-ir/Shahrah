package main

import (
	"net/http"
	"testing"
	"time"
)

func TestProviderHTTPClientRequiresExplicitProxy(t *testing.T) {
	t.Setenv("SILK_ROAD_EGRESS_PROXY_REQUIRED", "true")
	t.Setenv("SILK_ROAD_EGRESS_PROXY_URL", "")

	if _, _, _, err := providerHTTPClient(time.Second); err == nil {
		t.Fatal("expected an error when the required proxy URL is missing")
	}
}

func TestProviderHTTPClientUsesExplicitProxy(t *testing.T) {
	t.Setenv("SILK_ROAD_EGRESS_PROXY_REQUIRED", "true")
	t.Setenv("SILK_ROAD_EGRESS_PROXY_URL", "http://10.10.0.190:1443")

	client, configured, required, err := providerHTTPClient(time.Second)
	if err != nil {
		t.Fatalf("providerHTTPClient: %v", err)
	}

	if !configured || !required {
		t.Fatalf("configured=%t required=%t, want true/true", configured, required)
	}

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("transport type = %T", client.Transport)
	}

	request, err := http.NewRequest(http.MethodGet, "https://example.com", http.NoBody)
	if err != nil {
		t.Fatal(err)
	}

	proxyURL, err := transport.Proxy(request)
	if err != nil {
		t.Fatal(err)
	}

	if got := proxyURL.String(); got != "http://10.10.0.190:1443" {
		t.Fatalf("proxy URL = %q", got)
	}
}

func TestProviderHTTPClientRejectsInvalidProxy(t *testing.T) {
	t.Setenv("SILK_ROAD_EGRESS_PROXY_URL", "socks5://127.0.0.1:1080")

	if _, _, _, err := providerHTTPClient(time.Second); err == nil {
		t.Fatal("expected invalid proxy URL error")
	}
}
