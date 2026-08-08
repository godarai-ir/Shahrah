package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

	streamClient := clientForStream(client, true)

	mux := http.NewServeMux()
	metrics := newHTTPMetrics("tunnel-server")
	mux.Handle("/metrics", metrics.handler())

	localHealth := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if _, err := w.Write([]byte(`{"status":"healthy","service":"tunnel-server"}`)); err != nil {
			return
		}
	}
	mux.HandleFunc("/health", localHealth)
	mux.HandleFunc("/live", localHealth)
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, silkRoadURL+"/ready", http.NoBody)
		if err != nil {
			http.Error(w, "dependency unavailable", http.StatusServiceUnavailable)
			return
		}

		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "dependency unavailable", http.StatusServiceUnavailable)
			return
		}
		defer closeBody(resp.Body)

		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			http.Error(w, "dependency unavailable", http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)

		if _, copyErr := io.Copy(w, resp.Body); copyErr != nil {
			return
		}
	})

	mux.HandleFunc("/dispatch", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var payload platform.TunnelRequest
		if unmarshalErr := json.Unmarshal(body, &payload); unmarshalErr != nil {
			http.Error(w, unmarshalErr.Error(), http.StatusBadRequest)
			return
		}

		req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, silkRoadURL+"/dispatch", bytes.NewReader(body))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-GodarAI-Trace-ID", payload.TraceID)
		req.Header.Set("X-Shahrah-Request-ID", payload.RequestID)

		resp, err := selectHTTPClient(client, streamClient, payload.Stream).Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer closeBody(resp.Body)

		copyHeaders(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)
		relayBodyOrAbort(
			w,
			resp.Body,
			payload.Stream || responseIsEventStream(resp.Header),
			payload.RequestID,
			payload.TraceID,
		)
	})
	mux.HandleFunc("/dispatch/raw", func(w http.ResponseWriter, r *http.Request) {
		req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, silkRoadURL+"/dispatch/raw", r.Body)
		if err != nil {
			http.Error(w, "invalid raw tunnel request", http.StatusBadRequest)
			return
		}

		req.ContentLength = r.ContentLength
		req.Header.Set("Content-Type", "application/vnd.godarai.tunnel-stream")
		stream := strings.EqualFold(strings.TrimSpace(r.Header.Get("X-Shahrah-Tunnel-Stream")), "true")
		req.Header.Set("X-Shahrah-Tunnel-Stream", r.Header.Get("X-Shahrah-Tunnel-Stream"))
		req.Header.Set("X-GodarAI-Trace-ID", r.Header.Get("X-GodarAI-Trace-ID"))
		req.Header.Set("X-Shahrah-Request-ID", r.Header.Get("X-Shahrah-Request-ID"))

		resp, err := streamClient.Do(req)
		if err != nil {
			http.Error(w, "raw tunnel dependency unavailable", http.StatusBadGateway)
			return
		}
		defer closeBody(resp.Body)

		copyHeaders(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)
		relayBodyOrAbort(
			w,
			resp.Body,
			stream || responseIsEventStream(resp.Header),
			r.Header.Get("X-Shahrah-Request-ID"),
			r.Header.Get("X-GodarAI-Trace-ID"),
		)
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

func selectHTTPClient(regular, streaming *http.Client, stream bool) *http.Client {
	if stream {
		return streaming
	}

	return regular
}

func responseIsEventStream(headers http.Header) bool {
	return strings.Contains(strings.ToLower(headers.Get("Content-Type")), "text/event-stream")
}

func relayBodyOrAbort(
	w http.ResponseWriter,
	body io.Reader,
	flush bool,
	requestID string,
	traceID string,
) {
	if err := relayBody(w, body, flush); err != nil {
		log.Printf("tunnel response relay failed request_id=%q trace_id=%q: %v", requestID, traceID, err)
		panic(http.ErrAbortHandler)
	}
}

func relayBody(w http.ResponseWriter, body io.Reader, flush bool) error {
	if !flush {
		if _, err := io.Copy(w, body); err != nil {
			return fmt.Errorf("relay response body: %w", err)
		}

		return nil
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		if _, err := io.Copy(w, body); err != nil {
			return fmt.Errorf("relay streaming response body: %w", err)
		}

		return nil
	}

	buf := make([]byte, 32*1024)
	for {
		n, err := body.Read(buf)
		if n > 0 {
			if _, writeErr := w.Write(buf[:n]); writeErr != nil {
				return fmt.Errorf("write streaming response body: %w", writeErr)
			}

			flusher.Flush()
		}

		if err == io.EOF {
			return nil
		}

		if err != nil {
			return fmt.Errorf("read streaming response body: %w", err)
		}
	}
}

func closeBody(body io.Closer) {
	if body == nil {
		return
	}

	if err := body.Close(); err != nil {
		return
	}
}
