package server

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

// NewHandler builds the Silk Road HTTP handler.
func NewHandler(client *http.Client) http.Handler {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Minute}
	}
	registry := newProviderRegistry()

	mux := http.NewServeMux()
	metrics := newHTTPMetrics("silk-road")
	mux.Handle("/metrics", metrics.handler())
	localHealth := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"healthy","service":"silk-road"}`))
	}
	mux.HandleFunc("/health", localHealth)
	mux.HandleFunc("/live", localHealth)
	mux.HandleFunc("/ready", localHealth)

	mux.HandleFunc("/dispatch", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var payload platform.TunnelRequest
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		adapter, err := registry.adapterFor(payload.Deployment.Provider)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp, err := adapter.Dispatch(r.Context(), clientForStream(client, payload.Stream), payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		copyHeaders(w.Header(), resp.Headers)
		w.WriteHeader(resp.StatusCode)
		relayBody(w, resp.Body, payload.Stream || responseIsEventStream(resp.Headers))
	})

	return metrics.wrap(mux)
}

func copyHeaders(dst http.Header, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

func clientForStream(base *http.Client, stream bool) *http.Client {
	if !stream {
		return base
	}
	if base == nil {
		return &http.Client{}
	}
	clone := *base
	clone.Timeout = 0
	if clone.Transport == nil {
		clone.Transport = http.DefaultTransport
	}
	return &clone
}

func responseIsEventStream(headers http.Header) bool {
	return strings.Contains(strings.ToLower(headers.Get("Content-Type")), "text/event-stream")
}

func relayBody(w http.ResponseWriter, body io.Reader, flush bool) {
	if !flush {
		_, _ = io.Copy(w, body)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		_, _ = io.Copy(w, body)
		return
	}

	buf := make([]byte, 32*1024)
	for {
		n, err := body.Read(buf)
		if n > 0 {
			if _, writeErr := w.Write(buf[:n]); writeErr != nil {
				return
			}
			flusher.Flush()
		}
		if err != nil {
			return
		}
	}
}
