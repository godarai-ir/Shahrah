// Package tunnel provides HTTP tunnel helpers for outbound request streaming.
package tunnel

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

// HTTPTransport sends provider operations directly to Silk Road.
type HTTPTransport struct {
	baseURL      string
	client       *http.Client
	streamClient *http.Client
}

const (
	defaultMaxIdleConns        = 1024
	defaultMaxIdleConnsPerHost = 512
	defaultStreamBufferSize    = 8
	defaultStreamReadBuffer    = 32 * 1024
)

// HTTPTransportConfig controls connection reuse between Control Plane and Silk Road.
type HTTPTransportConfig struct {
	RequestTimeout      time.Duration
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	MaxConnsPerHost     int
	IdleConnTimeout     time.Duration
}

// NewHTTPTransport creates a transport client for Silk Road.
func NewHTTPTransport(baseURL string, timeout time.Duration) *HTTPTransport {
	return NewHTTPTransportWithConfig(baseURL, HTTPTransportConfig{RequestTimeout: timeout})
}

// NewHTTPTransportWithConfig creates a transport with an explicit shared connection pool.
func NewHTTPTransportWithConfig(baseURL string, config HTTPTransportConfig) *HTTPTransport {
	config = normalizeHTTPTransportConfig(config)
	transport := newHTTPRoundTripper(config)

	return &HTTPTransport{
		baseURL:      baseURL,
		client:       &http.Client{Transport: transport, Timeout: config.RequestTimeout},
		streamClient: &http.Client{Transport: transport},
	}
}

func normalizeHTTPTransportConfig(config HTTPTransportConfig) HTTPTransportConfig {
	if config.MaxIdleConns <= 0 {
		config.MaxIdleConns = defaultMaxIdleConns
	}

	if config.MaxIdleConnsPerHost <= 0 {
		config.MaxIdleConnsPerHost = defaultMaxIdleConnsPerHost
	}

	if config.IdleConnTimeout <= 0 {
		config.IdleConnTimeout = 90 * time.Second
	}

	return config
}

func newHTTPRoundTripper(config HTTPTransportConfig) *http.Transport {
	transport, ok := http.DefaultTransport.(*http.Transport)
	if ok {
		transport = transport.Clone()
	} else {
		transport = &http.Transport{Proxy: http.ProxyFromEnvironment}
	}

	transport.MaxIdleConns = config.MaxIdleConns
	transport.MaxIdleConnsPerHost = config.MaxIdleConnsPerHost
	transport.MaxConnsPerHost = config.MaxConnsPerHost
	transport.IdleConnTimeout = config.IdleConnTimeout
	transport.ForceAttemptHTTP2 = true

	return transport
}

// Proxy sends a non-streaming tunnel request.
func (t *HTTPTransport) Proxy(ctx context.Context, req platform.TunnelRequest) (*platform.TunnelResponse, error) {
	body := req.Body

	rawResponse, err := t.dispatchRaw(ctx, req, bytes.NewReader(body), int64(len(body)), false)
	if err != nil {
		return nil, err
	}
	defer closeBody(rawResponse.Body)

	out := &platform.TunnelResponse{
		StatusCode: rawResponse.StatusCode,
		Headers:    rawResponse.Headers,
	}

	out.Body, err = io.ReadAll(rawResponse.Body)
	if err != nil {
		return nil, err
	}

	return out, nil
}

// ProxyRaw sends a length-prefixed metadata envelope followed by the raw body.
func (t *HTTPTransport) ProxyRaw(ctx context.Context, req platform.TunnelRequest, body io.Reader, contentLength int64) (*platform.TunnelRawResponse, error) {
	return t.dispatchRaw(ctx, req, body, contentLength, true)
}

func (t *HTTPTransport) dispatchRaw(
	ctx context.Context,
	req platform.TunnelRequest,
	body io.Reader,
	contentLength int64,
	disableClientTimeout bool,
) (*platform.TunnelRawResponse, error) {
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
	httpReq.Header.Set("X-Shahrah-Tunnel-Stream", strconv.FormatBool(req.Stream))
	httpReq.Header.Set("X-GodarAI-Trace-ID", req.TraceID)
	httpReq.Header.Set("X-Shahrah-Request-ID", req.RequestID)

	client := t.client
	if disableClientTimeout {
		client = t.streamClient
	}

	//nolint:bodyclose // The caller owns the returned response body through TunnelRawResponse.
	resp, err := client.Do(httpReq)
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

	streamCtx, cancel := context.WithCancel(ctx)
	body := req.Body

	resp, err := t.dispatchRaw(streamCtx, req, bytes.NewReader(body), int64(len(body)), true)
	if err != nil {
		cancel()

		return nil, err
	}

	// Check for error status from upstream (e.g., 4xx/5xx from provider)
	if resp.StatusCode >= http.StatusBadRequest {
		defer cancel()

		respBody, readErr := io.ReadAll(resp.Body)
		closeBody(resp.Body)

		if readErr != nil {
			return nil, readErr
		}

		return &errorStream{
			statusCode: resp.StatusCode,
			headers:    resp.Headers,
			body:       respBody,
		}, nil
	}

	stream := &streamRelay{
		body:       resp.Body,
		cancel:     cancel,
		statusCode: resp.StatusCode,
		headers:    resp.Headers,
		chunks:     make(chan platform.TunnelChunk, defaultStreamBufferSize),
		errors:     make(chan error, 1),
	}

	go func() {
		defer func() {
			if closeErr := stream.Close(); closeErr != nil {
				select {
				case stream.errors <- closeErr:
				default:
				}
			}

			close(stream.chunks)
			close(stream.errors)
		}()

		reader := bufio.NewReaderSize(resp.Body, defaultStreamReadBuffer)
		for {
			line, err := reader.ReadSlice('\n')
			if len(line) > 0 {
				data := bytes.Clone(line)
				if !stream.sendChunk(streamCtx, platform.TunnelChunk{Data: data}) {
					return
				}
			}

			if err != nil {
				if errors.Is(err, bufio.ErrBufferFull) {
					continue
				}

				if err == io.EOF {
					stream.sendChunk(streamCtx, platform.TunnelChunk{IsDone: true})

					return
				}

				select {
				case stream.errors <- err:
				case <-streamCtx.Done():
				}

				return
			}
		}
	}()

	return stream, nil
}

// Healthy checks the Silk Road health endpoint.
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
		return fmt.Errorf("silk road unhealthy: %d", resp.StatusCode)
	}

	return nil
}

type streamRelay struct {
	body       io.Closer
	cancel     context.CancelFunc
	statusCode int
	headers    map[string]string
	chunks     chan platform.TunnelChunk
	errors     chan error
	closeOnce  sync.Once
	closeErr   error
}

func (s *streamRelay) Chunks() <-chan platform.TunnelChunk { return s.chunks }
func (s *streamRelay) Errors() <-chan error                { return s.errors }
func (s *streamRelay) Close() error {
	s.closeOnce.Do(func() {
		if s.cancel != nil {
			s.cancel()
		}

		if s.body != nil {
			s.closeErr = s.body.Close()
		}
	})

	return s.closeErr
}
func (s *streamRelay) Headers() map[string]string { return s.headers }
func (s *streamRelay) StatusCode() int            { return s.statusCode }

func (s *streamRelay) sendChunk(ctx context.Context, chunk platform.TunnelChunk) bool {
	select {
	case s.chunks <- chunk:
		return true
	case <-ctx.Done():
		return false
	}
}

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

func closeBody(body io.Closer) {
	if body == nil {
		return
	}

	if err := body.Close(); err != nil {
		return
	}
}
