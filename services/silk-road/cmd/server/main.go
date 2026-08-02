package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	appserver "github.com/Delpak/shahrah/services/silk-road/pkg/server"
)

func main() {
	addr := getenv("SILK_ROAD_ADDR", ":8083")
	client, proxyConfigured, proxyRequired, err := providerHTTPClient(5 * time.Minute)
	if err != nil {
		log.Fatalf("configure provider HTTP client: %v", err)
	}

	log.Printf("silk-road listening on %s (egress proxy configured=%t required=%t)", addr, proxyConfigured, proxyRequired)
	log.Fatal(http.ListenAndServe(addr, appserver.NewHandler(client)))
}

func providerHTTPClient(timeout time.Duration) (*http.Client, bool, bool, error) {
	required, err := strconv.ParseBool(getenv("SILK_ROAD_EGRESS_PROXY_REQUIRED", "false"))
	if err != nil {
		return nil, false, false, fmt.Errorf("SILK_ROAD_EGRESS_PROXY_REQUIRED: %w", err)
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	rawProxyURL := os.Getenv("SILK_ROAD_EGRESS_PROXY_URL")
	if rawProxyURL == "" {
		if required {
			return nil, false, true, fmt.Errorf("SILK_ROAD_EGRESS_PROXY_URL is required")
		}
		transport.Proxy = http.ProxyFromEnvironment
		return &http.Client{Timeout: timeout, Transport: transport}, false, false, nil
	}

	proxyURL, err := url.Parse(rawProxyURL)
	if err != nil || proxyURL.Host == "" || (proxyURL.Scheme != "http" && proxyURL.Scheme != "https") {
		return nil, false, required, fmt.Errorf("SILK_ROAD_EGRESS_PROXY_URL must be an absolute HTTP(S) URL")
	}
	transport.Proxy = http.ProxyURL(proxyURL)
	return &http.Client{Timeout: timeout, Transport: transport}, true, required, nil
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
