package main

import (
	"log"
	"net/http"
	"os"
	"time"

	appserver "github.com/Delpak/shahrah/services/tunnel-client/pkg/server"
)

func main() {
	serverURL := getenv("TUNNEL_SERVER_URL", "http://127.0.0.1:8082")
	addr := getenv("TUNNEL_CLIENT_ADDR", ":8081")
	client := tunnelHTTPClient(5 * time.Minute)

	log.Printf("tunnel-client listening on %s -> %s", addr, serverURL)
	log.Fatal(http.ListenAndServe(addr, appserver.NewHandler(serverURL, client)))
}

func tunnelHTTPClient(timeout time.Duration) *http.Client {
	transport, ok := http.DefaultTransport.(*http.Transport)
	if ok {
		transport = transport.Clone()
	} else {
		transport = &http.Transport{Proxy: http.ProxyFromEnvironment}
	}

	transport.MaxIdleConns = 1024
	transport.MaxIdleConnsPerHost = 512
	transport.IdleConnTimeout = 90 * time.Second
	transport.ForceAttemptHTTP2 = true

	return &http.Client{Timeout: timeout, Transport: transport}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
