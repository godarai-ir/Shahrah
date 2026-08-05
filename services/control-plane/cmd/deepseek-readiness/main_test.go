package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeepSeekReadinessRunsThroughGodarAIAndSanitizesReport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		require.Equal(t, "Bearer virtual-secret", request.Header.Get("Authorization"))
		require.Equal(t, "deepseek", request.Header.Get("x-godarai-provider"))
		require.Equal(t, "account-1", request.Header.Get("x-godarai-provider-account"))

		writer.Header().Set("Content-Type", "application/json")

		if request.Method == http.MethodGet {
			_, err := io.WriteString(writer, `{"data":[{"id":"deepseek-v4-flash"},{"id":"deepseek-v4-pro"}]}`)
			require.NoError(t, err)

			return
		}

		var body struct {
			Model string `json:"model"`
		}
		require.NoError(t, json.NewDecoder(request.Body).Decode(&body))

		if request.URL.Path == "/v1/responses" {
			_, err := io.WriteString(writer, `{"model":"`+body.Model+`","usage":{"input_tokens":3,"input_tokens_details":{"cached_tokens":1},"output_tokens":1,"total_tokens":4}}`)
			require.NoError(t, err)

			return
		}

		_, err := io.WriteString(writer, `{"model":"`+body.Model+`","usage":{"prompt_cache_hit_tokens":1,"prompt_cache_miss_tokens":2,"completion_tokens":1}}`)
		require.NoError(t, err)
	}))
	defer server.Close()

	t.Setenv("GODARAI_BASE_URL", server.URL)
	t.Setenv("GODARAI_VIRTUAL_KEY", "virtual-secret")
	t.Setenv("GODARAI_DEEPSEEK_PROVIDER_ACCOUNT", "account-1")
	t.Setenv("GODARAI_DEEPSEEK_MAX_SPEND_USD", "0.01")
	reportPath := filepath.Join(t.TempDir(), "report.json")

	err := run([]string{"-report", reportPath}, io.Discard)
	require.NoError(t, err)
	report, err := os.ReadFile(reportPath)
	require.NoError(t, err)
	require.NotContains(t, string(report), "virtual-secret")
	require.Contains(t, string(report), `"production_ready": false`)
}

func TestDeepSeekReadinessRequiresExplicitSpendCeiling(t *testing.T) {
	t.Setenv("GODARAI_BASE_URL", "https://gateway.godarai.ir")
	t.Setenv("GODARAI_VIRTUAL_KEY", "virtual-secret")
	t.Setenv("GODARAI_DEEPSEEK_PROVIDER_ACCOUNT", "account-1")
	t.Setenv("GODARAI_DEEPSEEK_MAX_SPEND_USD", "")

	_, err := parseConfig(nil)
	require.ErrorContains(t, err, "MAX_SPEND_USD")
}
