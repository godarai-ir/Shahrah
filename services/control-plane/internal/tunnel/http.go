package tunnel

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

// HTTPTransport sends requests to the local tunnel-client service.
type HTTPTransport struct {
	baseURL string
	client  *http.Client
}

// NewHTTPTransport creates a transport client for the tunnel-client service.
func NewHTTPTransport(baseURL string, timeout time.Duration) *HTTPTransport {
	return &HTTPTransport{
		baseURL: baseURL,
		client:  &http.Client{Timeout: timeout},
	}
}

// Proxy sends a non-streaming tunnel request.
func (t *HTTPTransport) Proxy(ctx context.Context, req platform.TunnelRequest) (*platform.TunnelResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, t.baseURL+"/dispatch", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Shahrah-Trace-ID", req.TraceID)

	resp, err := t.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	out := &platform.TunnelResponse{
		StatusCode: resp.StatusCode,
		Headers:    flattenHeaders(resp.Header),
	}
	out.Body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Stream sends a streaming tunnel request.
func (t *HTTPTransport) Stream(ctx context.Context, req platform.TunnelRequest) (platform.TunnelStream, error) {
	req.Stream = true
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, t.baseURL+"/dispatch", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := t.client.Do(httpReq)
	if err != nil {
		return nil, err
	}

	stream := &streamRelay{
		body:   resp.Body,
		chunks: make(chan platform.TunnelChunk),
		errors: make(chan error, 1),
	}

	go func() {
		defer resp.Body.Close()
		defer close(stream.chunks)
		defer close(stream.errors)

		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadBytes('\n')
			if len(line) > 0 {
				stream.chunks <- platform.TunnelChunk{Data: append([]byte(nil), line...)}
			}
			if err != nil {
				if err == io.EOF {
					stream.chunks <- platform.TunnelChunk{IsDone: true}
					return
				}
				stream.errors <- err
				return
			}
		}
	}()

	return stream, nil
}

// Healthy checks the tunnel-client health endpoint.
func (t *HTTPTransport) Healthy(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.baseURL+"/health", nil)
	if err != nil {
		return err
	}
	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("tunnel client unhealthy: %d", resp.StatusCode)
	}
	return nil
}

type streamRelay struct {
	body   io.Closer
	chunks chan platform.TunnelChunk
	errors chan error
}

func (s *streamRelay) Chunks() <-chan platform.TunnelChunk { return s.chunks }
func (s *streamRelay) Errors() <-chan error                { return s.errors }
func (s *streamRelay) Close() error                        { return s.body.Close() }

func flattenHeaders(in http.Header) map[string]string {
	out := make(map[string]string, len(in))
	for key, values := range in {
		if len(values) > 0 {
			out[key] = values[0]
		}
	}
	return out
}
