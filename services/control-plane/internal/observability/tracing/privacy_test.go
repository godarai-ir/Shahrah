package tracing

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCapturePolicy(t *testing.T) {
	assert.Equal(t, CapturePolicyNone, ParseCapturePolicy("none"))
	assert.Equal(t, CapturePolicyMetadataOnly, ParseCapturePolicy("metadata-only"))
	assert.Equal(t, CapturePolicySampledRedacted, ParseCapturePolicy("sampled-redacted"))
	assert.Equal(t, CapturePolicyFullSecure, ParseCapturePolicy("full-secure"))
	assert.Equal(t, CapturePolicyMetadataOnly, ParseCapturePolicy("unknown"))
}

func TestPrivacyManager_MetadataOnly(t *testing.T) {
	m := NewPrivacyManager(Config{
		CapturePolicy:   CapturePolicyMetadataOnly,
		PayloadMaxBytes: 4096,
	})
	out := m.Apply("trace_1", "span_1", "provider.call", map[string]interface{}{
		"provider": "openai",
		"request_payload": map[string]interface{}{
			"prompt": "hello world",
			"email":  "user@example.com",
		},
	})
	_, hasRaw := out["request_payload"]
	assert.False(t, hasRaw)
	assert.Equal(t, "openai", out["provider"])
	meta, ok := out["request_payload_meta"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, true, meta["present"])
}

func TestPrivacyManager_SampledRedacted(t *testing.T) {
	m := NewPrivacyManager(Config{
		CapturePolicy:   CapturePolicySampledRedacted,
		PayloadMaxBytes: 4096,
		PIIHashSalt:     "salt",
	})
	out := m.Apply("trace_1", "span_2", "provider.call", map[string]interface{}{
		"request_payload": map[string]interface{}{
			"prompt": "hello world",
			"email":  "user@example.com",
		},
	})

	redacted, ok := out["request_payload_redacted"].(map[string]interface{})
	require.True(t, ok)
	prompt, ok := redacted["prompt"].(string)
	require.True(t, ok)
	assert.True(t, strings.Contains(prompt, "[redacted sha256="))

	email, ok := redacted["email"].(string)
	require.True(t, ok)
	assert.True(t, strings.Contains(email, "[redacted sha256="))
}

func TestPrivacyManager_FullSecureStoresEncryptedBlob(t *testing.T) {
	dir := t.TempDir()
	key := base64.StdEncoding.EncodeToString([]byte("12345678901234567890123456789012"))
	m := NewPrivacyManager(Config{
		CapturePolicy:     CapturePolicyFullSecure,
		PayloadBlobDir:    dir,
		PayloadEncryptKey: key,
		PayloadMaxBytes:   4096,
	})

	out := m.Apply("trace_1", "span_3", "provider.call", map[string]interface{}{
		"request_payload": map[string]interface{}{
			"prompt": "very sensitive prompt",
		},
	})
	ref, ok := out["request_payload_blob_ref"].(string)
	require.True(t, ok)
	require.NotEmpty(t, ref)
	raw, err := os.ReadFile(ref)
	require.NoError(t, err)
	assert.NotEmpty(t, raw)
	assert.False(t, strings.Contains(string(raw), "very sensitive prompt"))
}

func TestRecordStepSpan_AppliesPrivacyPolicy(t *testing.T) {
	app := fiber.New()
	queue := NewQueue(16)
	sampler := NewSampler(1.0, 1500)
	privacy := NewPrivacyManager(Config{
		CapturePolicy:   CapturePolicyMetadataOnly,
		PayloadMaxBytes: 4096,
	})
	app.Use(Middleware(queue, sampler, 1500, privacy))
	app.Get("/x", func(c *fiber.Ctx) error {
		start := time.Now().Add(-5 * time.Millisecond)
		RecordStepSpan(c, QueueFromContext(c), "", "provider.call", "client", start, fiber.StatusOK, map[string]interface{}{
			"request_payload": map[string]interface{}{"prompt": "hello"},
		})

		return c.SendStatus(200)
	})

	req := httptest.NewRequest("GET", "/x", http.NoBody)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)

	batch := queue.DequeueBatch(10)
	require.Len(t, batch, 2)
	step := batch[1]
	_, hasRaw := step.Attributes["request_payload"]
	assert.False(t, hasRaw)

	_, hasMeta := step.Attributes["request_payload_meta"]
	assert.True(t, hasMeta)
}
