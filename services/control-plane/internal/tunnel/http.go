// Package tunnel provides HTTP tunnel helpers for outbound request streaming.
package tunnel

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
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
	httpReq.Header.Set("X-GodarAI-Trace-ID", req.TraceID)

	resp, err := t.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp.Body)

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

// ProxyRaw sends a length-prefixed metadata envelope followed by the raw body.
func (t *HTTPTransport) ProxyRaw(ctx context.Context, req platform.TunnelRequest, body io.Reader, contentLength int64) (*platform.TunnelRawResponse, error) {
	req.Body = nil
	metadata, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	if len(metadata) == 0 || len(metadata) > 256*1024 {
		return nil, fmt.Errorf("raw tunnel metadata is invalid")
	}
	prefix := make([]byte, 4)
	binary.BigEndian.PutUint32(prefix, uint32(len(metadata)))
	if body == nil {
		body = http.NoBody
	}
	reader := io.MultiReader(bytes.NewReader(prefix), bytes.NewReader(metadata), body)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, t.baseURL+"/dispatch/raw", reader)
	if err != nil {
		return nil, err
	}
	if contentLength >= 0 {
		httpReq.ContentLength = int64(len(prefix)+len(metadata)) + contentLength
	}
	httpReq.Header.Set("Content-Type", "application/vnd.godarai.tunnel-stream")
	resp, err := streamingHTTPClient(t.client).Do(httpReq)
	if err != nil {
		return nil, err
	}
	return &platform.TunnelRawResponse{
		StatusCode: resp.StatusCode,
		Headers:    flattenHeaders(resp.Header),
		Body:       resp.Body,
	}, nil
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

	//nolint:bodyclose // streamRelay takes ownership of resp.Body and closes it.
	resp, err := streamingHTTPClient(t.client).Do(httpReq)
	if err != nil {
		return nil, err
	}

	// Check for error status from upstream (e.g., 4xx/5xx from provider)
	if resp.StatusCode >= http.StatusBadRequest {
		respBody, readErr := io.ReadAll(resp.Body)
		closeBody(resp.Body)

		if readErr != nil {
			return nil, readErr
		}

		return &errorStream{
			statusCode: resp.StatusCode,
			headers:    flattenHeaders(resp.Header),
			body:       respBody,
		}, nil
	}

	stream := &streamRelay{
		body:       resp.Body,
		statusCode: resp.StatusCode,
		headers:    flattenHeaders(resp.Header),
		chunks:     make(chan platform.TunnelChunk),
		errors:     make(chan error, 1),
	}

	go func() {
		defer closeBody(resp.Body)
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.baseURL+"/health", http.NoBody)
	if err != nil {
		return err
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer closeBody(resp.Body)

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("tunnel client unhealthy: %d", resp.StatusCode)
	}

	return nil
}

type streamRelay struct {
	body       io.Closer
	statusCode int
	headers    map[string]string
	chunks     chan platform.TunnelChunk
	errors     chan error
}

func (s *streamRelay) Chunks() <-chan platform.TunnelChunk { return s.chunks }
func (s *streamRelay) Errors() <-chan error                { return s.errors }
func (s *streamRelay) Close() error                        { return s.body.Close() }
func (s *streamRelay) Headers() map[string]string          { return s.headers }
func (s *streamRelay) StatusCode() int                     { return s.statusCode }

// errorStream implements TunnelStream for immediate upstream errors.
type errorStream struct {
	statusCode int
	headers    map[string]string
	body       []byte
}

func (e *errorStream) Chunks() <-chan platform.TunnelChunk {
	ch := make(chan platform.TunnelChunk, 1)
	ch <- platform.TunnelChunk{Data: e.body, IsDone: true}

	close(ch)

	return ch
}
func (e *errorStream) Errors() <-chan error       { return nil }
func (e *errorStream) Close() error               { return nil }
func (e *errorStream) Headers() map[string]string { return e.headers }
func (e *errorStream) StatusCode() int            { return e.statusCode }

func flattenHeaders(in http.Header) map[string]string {
	out := make(map[string]string, len(in))
	for key, values := range in {
		if len(values) > 0 {
			out[key] = values[0]
		}
	}

	return out
}

func streamingHTTPClient(base *http.Client) *http.Client {
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

func closeBody(body io.Closer) {
	if body == nil {
		return
	}

	if err := body.Close(); err != nil {
		return
	}
}
