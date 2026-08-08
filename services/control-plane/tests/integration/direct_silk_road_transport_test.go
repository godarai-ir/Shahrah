package integration_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	tunneltransport "github.com/Delpak/shahrah/services/control-plane/internal/tunnel"
	silkserver "github.com/Delpak/shahrah/services/silk-road/pkg/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type directRoundTripFunc func(*http.Request) (*http.Response, error)

func (f directRoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestDirectSilkRoadTransportStreamsProviderResponse(t *testing.T) {
	const providerEvents = "data: {\"delta\":\"one\"}\n\ndata: [DONE]\n\n"

	transport := newDirectSilkRoadTransport(t, directRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		assert.Equal(t, "/v1/chat/completions", request.URL.Path)

		return directProviderResponse(request, http.StatusOK, "text/event-stream", providerEvents), nil
	}), time.Second)

	stream, err := transport.Stream(t.Context(), directSilkRoadRequest(platform.ProviderOperationChatCompletions))

	require.NoError(t, err)

	defer func() {
		require.NoError(t, stream.Close())
	}()

	var received bytes.Buffer

	for chunk := range stream.Chunks() {
		if !chunk.IsDone {
			_, writeErr := received.Write(chunk.Data)
			require.NoError(t, writeErr)
		}
	}

	for streamErr := range stream.Errors() {
		require.NoError(t, streamErr)
	}

	assert.Equal(t, http.StatusOK, stream.StatusCode())
	assert.Equal(t, providerEvents, received.String())
}

func TestDirectSilkRoadTransportPropagatesCancellation(t *testing.T) {
	providerBody := newCancellationProbeBody()
	transport := newDirectSilkRoadTransport(t, directRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       providerBody,
			Request:    request,
		}, nil
	}), time.Second)
	t.Cleanup(func() { require.NoError(t, providerBody.Close()) })

	stream, err := transport.Stream(t.Context(), directSilkRoadRequest(platform.ProviderOperationChatCompletions))
	require.NoError(t, err)

	select {
	case <-stream.Chunks():
	case <-time.After(time.Second):
		t.Fatal("provider stream did not start")
	}

	require.NoError(t, stream.Close())

	select {
	case <-providerBody.closed:
	case <-time.After(5 * time.Second):
		t.Fatal("provider response was not closed after direct stream cancellation")
	}
}

func TestDirectSilkRoadTransportStreamsLargeUpload(t *testing.T) {
	const uploadSize = 8 * 1024 * 1024

	upload := bytes.Repeat([]byte("s"), uploadSize)
	receivedLength := make(chan int, 1)
	transport := newDirectSilkRoadTransport(t, directRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			return nil, err
		}

		receivedLength <- len(body)

		assert.Equal(t, int64(uploadSize), request.ContentLength)
		assert.Equal(t, "/v1/files", request.URL.Path)

		return directProviderResponse(request, http.StatusCreated, "application/json", `{"id":"file-test"}`), nil
	}), time.Second)

	request := directSilkRoadRequest(platform.ProviderOperationOpenAIFilesCreate)
	request.Headers = map[string]string{"Content-Type": "application/octet-stream"}
	response, err := transport.ProxyRaw(t.Context(), request, bytes.NewReader(upload), int64(len(upload)))

	require.NoError(t, err)

	defer func() {
		require.NoError(t, response.Body.Close())
	}()

	_, err = io.Copy(io.Discard, response.Body)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, response.StatusCode)
	assert.Equal(t, uploadSize, <-receivedLength)
}

func TestDirectSilkRoadTransportEnforcesTimeout(t *testing.T) {
	providerRelease := make(chan struct{})
	transport := newDirectSilkRoadTransport(t, directRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		select {
		case <-request.Context().Done():
			return nil, request.Context().Err()
		case <-providerRelease:
			return nil, context.DeadlineExceeded
		}
	}), 25*time.Millisecond)
	t.Cleanup(func() { close(providerRelease) })

	_, err := transport.Proxy(t.Context(), directSilkRoadRequest(platform.ProviderOperationChatCompletions))
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded), "expected deadline error, got %v", err)
}

func TestDirectSilkRoadTransportRunsCredentialProbe(t *testing.T) {
	type capturedProbe struct {
		method        string
		url           string
		authorization string
	}

	captured := make(chan capturedProbe, 1)
	transport := newDirectSilkRoadTransport(t, directRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		captured <- capturedProbe{
			method:        request.Method,
			url:           request.URL.String(),
			authorization: request.Header.Get("Authorization"),
		}

		return directProviderResponse(request, http.StatusOK, "application/json", `{"data":[]}`), nil
	}), time.Second)

	request := directSilkRoadRequest(platform.ProviderOperationCredentialProbe)
	request.Deployment = platform.Deployment{
		Provider: "deepseek",
		ProviderAuth: &platform.ResolvedProviderAuth{
			Mode:   string(platform.ProviderAuthModeAPIKey),
			APIKey: "credential-test-key",
		},
	}
	response, err := transport.Proxy(t.Context(), request)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.JSONEq(t, `{"data":[]}`, string(response.Body))
	assert.Equal(t, capturedProbe{
		method:        http.MethodGet,
		url:           "https://api.deepseek.com/models",
		authorization: "Bearer credential-test-key",
	}, <-captured)
}

func newDirectSilkRoadTransport(t *testing.T, roundTripper http.RoundTripper, timeout time.Duration) *tunneltransport.HTTPTransport {
	t.Helper()

	silkRoad := httptest.NewServer(silkserver.NewHandler(&http.Client{Transport: roundTripper}))
	t.Cleanup(silkRoad.Close)

	return tunneltransport.NewHTTPTransport(silkRoad.URL, timeout)
}

func directSilkRoadRequest(operation platform.ProviderOperation) platform.TunnelRequest {
	return platform.TunnelRequest{
		RequestID: "direct-request",
		TraceID:   "direct-trace",
		Operation: operation,
		Deployment: platform.Deployment{
			Provider: "openai",
			Model:    "gpt-test",
			APIBase:  "https://provider.test/v1",
			APIKey:   "provider-test-key",
		},
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    []byte(`{"model":"gpt-test"}`),
	}
}

func directProviderResponse(request *http.Request, status int, contentType, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{contentType}},
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Request:    request,
	}
}

type cancellationProbeBody struct {
	closed chan struct{}
	once   sync.Once
}

func newCancellationProbeBody() *cancellationProbeBody {
	return &cancellationProbeBody{closed: make(chan struct{})}
}

func (b *cancellationProbeBody) Read(output []byte) (int, error) {
	select {
	case <-b.closed:
		return 0, io.EOF
	case <-time.After(10 * time.Millisecond):
		return copy(output, "data: ping\n\n"), nil
	}
}

func (b *cancellationProbeBody) Close() error {
	b.once.Do(func() { close(b.closed) })

	return nil
}
