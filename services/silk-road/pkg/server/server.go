package server

import (
	"encoding/json"
	"io"
	"net/http"
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
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"healthy","service":"silk-road"}`))
	})

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

		resp, err := adapter.Dispatch(r.Context(), client, payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		copyHeaders(w.Header(), resp.Headers)
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	})

	return mux
}

func copyHeaders(dst http.Header, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}
