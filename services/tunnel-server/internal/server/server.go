package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

// NewHandler builds the tunnel-server HTTP handler.
func NewHandler(silkRoadURL string, client *http.Client) http.Handler {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Minute}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		req, _ := http.NewRequestWithContext(r.Context(), http.MethodGet, silkRoadURL+"/health", nil)
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		defer resp.Body.Close()
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
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

		req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, silkRoadURL+"/dispatch", bytes.NewReader(body))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Shahrah-Trace-ID", payload.TraceID)
		req.Header.Set("X-Shahrah-Request-ID", payload.RequestID)

		resp, err := clientForStream(client, payload.Stream).Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		copyHeaders(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)
		relayBody(w, resp.Body, payload.Stream || responseIsEventStream(resp.Header))
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
