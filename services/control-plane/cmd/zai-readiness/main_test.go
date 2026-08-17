package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestZAIReadinessRequiresExplicitLiveApproval(t *testing.T) {
	t.Setenv("GODARAI_ZAI_ALLOW_LIVE", "")

	_, err := parseConfig(nil)
	require.ErrorContains(t, err, "ALLOW_LIVE")
}

func TestZAIReadinessRunsThroughGodarAIAndSanitizesReport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		require.Equal(t, "Bearer virtual-secret", request.Header.Get("Authorization"))
		require.Equal(t, "zai", request.Header.Get("x-godarai-provider"))
		writer.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(writer, `{"usage":{"prompt_tokens":3,"prompt_tokens_details":{"cached_tokens":1},"completion_tokens":1,"total_tokens":4}}`)
		require.NoError(t, err)
	}))
	defer server.Close()

	t.Setenv("GODARAI_ZAI_ALLOW_LIVE", "true")
	t.Setenv("GODARAI_BASE_URL", server.URL)
	t.Setenv("GODARAI_VIRTUAL_KEY", "virtual-secret")
	t.Setenv("GODARAI_ZAI_PROVIDER_ACCOUNT", "account-1")
	t.Setenv("GODARAI_ZAI_MAX_SPEND_USD", "0.02")

	for _, key := range []string{"GODARAI_ZAI_IMAGE_URL", "GODARAI_ZAI_VIDEO_URL", "GODARAI_ZAI_DOCUMENT_URL", "GODARAI_ZAI_OCR_FILE_URL"} {
		t.Setenv(key, "https://fixtures.example/test")
	}

	reportPath := filepath.Join(t.TempDir(), "report.json")
	require.NoError(t, run([]string{"-report", reportPath}, io.Discard))
	report, err := os.ReadFile(reportPath)
	require.NoError(t, err)
	require.NotContains(t, string(report), "virtual-secret")
	require.Contains(t, string(report), `"production_ready": false`)
}
