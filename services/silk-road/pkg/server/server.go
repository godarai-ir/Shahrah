package server

import (
	"encoding/binary"
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

	streamClient := clientForStream(client, true)
	registry := newProviderRegistry()

	mux := http.NewServeMux()
	metrics := newHTTPMetrics("silk-road")
	mux.Handle("/metrics", metrics.handler())

	localHealth := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if _, err := w.Write([]byte(`{"status":"healthy","service":"silk-road"}`)); err != nil {
			return
		}
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
		if unmarshalErr := json.Unmarshal(body, &payload); unmarshalErr != nil {
			http.Error(w, unmarshalErr.Error(), http.StatusBadRequest)
			return
		}

		if payload.Operation == platform.ProviderOperationCredentialProbe {
			probeResp, probeErr := dispatchCredentialProbe(r.Context(), client, payload)
			if probeErr != nil {
				http.Error(w, probeErr.Error(), http.StatusBadRequest)
				return
			}
			defer closeBody(probeResp.Body)

			copyHeaders(w.Header(), probeResp.Headers)
			w.WriteHeader(probeResp.StatusCode)
			relayBody(w, probeResp.Body, false)

			return
		}

		adapter, err := registry.adapterFor(payload.Deployment.Provider)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp, err := adapter.Dispatch(r.Context(), selectHTTPClient(client, streamClient, payload.Stream), payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer closeBody(resp.Body)

		copyHeaders(w.Header(), resp.Headers)
		w.WriteHeader(resp.StatusCode)
		relayBody(w, resp.Body, payload.Stream || responseIsEventStream(resp.Headers))
	})
	mux.HandleFunc("/dispatch/raw", func(w http.ResponseWriter, r *http.Request) {
		var lengthPrefix [4]byte
		if _, err := io.ReadFull(r.Body, lengthPrefix[:]); err != nil {
			http.Error(w, "invalid raw tunnel envelope", http.StatusBadRequest)
			return
		}

		metadataLength := int(binary.BigEndian.Uint32(lengthPrefix[:]))
		if metadataLength <= 0 || metadataLength > 256*1024 {
			http.Error(w, "raw tunnel metadata exceeds limit", http.StatusBadRequest)
			return
		}

		metadata := make([]byte, metadataLength)
		if _, err := io.ReadFull(r.Body, metadata); err != nil {
			http.Error(w, "invalid raw tunnel metadata", http.StatusBadRequest)
			return
		}

		var payload platform.TunnelRequest
		if unmarshalErr := json.Unmarshal(metadata, &payload); unmarshalErr != nil {
			http.Error(w, "invalid raw tunnel metadata", http.StatusBadRequest)
			return
		}

		rawLength := int64(-1)
		if r.ContentLength >= int64(4+metadataLength) {
			rawLength = r.ContentLength - int64(4+metadataLength)
		}

		var (
			resp        *adapterResponse
			dispatchErr error
		)
		if payload.Operation == platform.ProviderOperationCredentialProbe {
			payload.Body, dispatchErr = io.ReadAll(r.Body)
			if dispatchErr == nil {
				resp, dispatchErr = dispatchCredentialProbe(r.Context(), streamClient, payload)
			}
		} else {
			adapter, adapterErr := registry.adapterFor(payload.Deployment.Provider)
			if adapterErr != nil {
				http.Error(w, adapterErr.Error(), http.StatusBadRequest)
				return
			}

			if rawAdapter, ok := adapter.(rawProviderAdapter); ok {
				resp, dispatchErr = rawAdapter.DispatchRaw(r.Context(), streamClient, payload, r.Body, rawLength)
			} else {
				payload.Body, dispatchErr = io.ReadAll(r.Body)
				if dispatchErr == nil {
					resp, dispatchErr = adapter.Dispatch(r.Context(), streamClient, payload)
				}
			}
		}

		if dispatchErr != nil {
			http.Error(w, "raw provider operation failed", http.StatusBadGateway)
			return
		}

		if resp == nil || resp.Body == nil {
			http.Error(w, "raw provider returned an empty response", http.StatusBadGateway)
			return
		}

		defer closeBody(resp.Body)

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

func selectHTTPClient(regular, streaming *http.Client, stream bool) *http.Client {
	if stream {
		return streaming
	}

	return regular
}

func responseIsEventStream(headers http.Header) bool {
	return strings.Contains(strings.ToLower(headers.Get("Content-Type")), "text/event-stream")
}

func relayBody(w http.ResponseWriter, body io.Reader, flush bool) {
	if !flush {
		if _, err := io.Copy(w, body); err != nil {
			return
		}

		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		if _, err := io.Copy(w, body); err != nil {
			return
		}

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

func closeBody(body io.Closer) {
	if body == nil {
		return
	}

	if err := body.Close(); err != nil {
		return
	}
}
