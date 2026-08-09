package server

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestAnthropicInterruptedStreamRemainsAnError(t *testing.T) {
	t.Parallel()

	upstream := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body: io.NopCloser(strings.NewReader(strings.Join([]string{
			"event: message_start",
			`data: {"type":"message_start","message":{"id":"msg_interrupted","model":"claude-sonnet-4-6"}}`,
			"",
			"event: content_block_delta",
			`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"partial"}}`,
			"",
		}, "\n"))),
	}
	response := adaptAnthropicStreamResponse(upstream, "claude-sonnet-4-6")

	body, err := io.ReadAll(response.Body)
	if err == nil {
		t.Fatal("interrupted stream must surface a reader error")
	}

	text := string(body)
	if strings.Contains(text, "data: [DONE]") {
		t.Fatalf("interrupted stream must not be marked successful: %s", text)
	}
}

func TestAnthropic500MBUploadUsesBoundedMemory(t *testing.T) {
	if testing.Short() || os.Getenv("ANTHROPIC_READINESS_LOAD_TEST") != "1" {
		t.Skip("set ANTHROPIC_READINESS_LOAD_TEST=1 to run the 500MB release gate")
	}

	const fileBytes = int64(500 * 1024 * 1024)

	var received atomic.Int64

	upstream := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		request *http.Request,
	) {
		count, err := io.Copy(io.Discard, request.Body)
		if err != nil {
			t.Errorf("upstream copy: %v", err)
			return
		}

		received.Store(count)

		_, _ = io.WriteString(w, `{"id":"file_upstream","type":"file"}`)
	}))
	defer upstream.Close()

	metadata, err := json.Marshal(platform.TunnelRequest{
		Operation: platform.ProviderOperationAnthropicFilesCreate,
		Headers: map[string]string{
			"content-type": "application/octet-stream",
		},
		Deployment: platform.Deployment{
			Provider: "anthropic",
			APIBase:  upstream.URL,
			APIKey:   "sk-provider",
			ProviderHeaders: map[string]string{
				"Anthropic-Beta-Allowlist": "files-api-2025-04-14",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	var prefix [4]byte
	binary.BigEndian.PutUint32(prefix[:], uint32(len(metadata)))
	body := io.MultiReader(
		bytes.NewReader(prefix[:]),
		bytes.NewReader(metadata),
		io.LimitReader(zeroReleaseReader{}, fileBytes),
	)

	runtime.GC()

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	var peak atomic.Uint64
	peak.Store(before.Alloc)

	done := make(chan struct{})

	go func() {
		ticker := time.NewTicker(2 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				var current runtime.MemStats
				runtime.ReadMemStats(&current)

				for {
					known := peak.Load()
					if current.Alloc <= known || peak.CompareAndSwap(known, current.Alloc) {
						break
					}
				}
			}
		}
	}()

	request := httptest.NewRequest(http.MethodPost, "/dispatch/raw", body)
	request.ContentLength = int64(4+len(metadata)) + fileBytes
	recorder := httptest.NewRecorder()
	NewHandler(nil).ServeHTTP(recorder, request)
	close(done)

	if recorder.Code != http.StatusOK || received.Load() != fileBytes {
		t.Fatalf(
			"500MB stream failed: status=%d received=%d body=%s",
			recorder.Code,
			received.Load(),
			recorder.Body.String(),
		)
	}

	if growth := peak.Load() - before.Alloc; growth > 96*1024*1024 {
		t.Fatalf("heap grew by %d bytes while streaming 500MB", growth)
	}
}

type zeroReleaseReader struct{}

func (zeroReleaseReader) Read(buffer []byte) (int, error) {
	for index := range buffer {
		buffer[index] = 0
	}

	return len(buffer), nil
}
