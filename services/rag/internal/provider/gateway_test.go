package provider

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/stretchr/testify/require"
)

func TestGatewayOCRUsesPrivateDataURLAndPreservesPersianPages(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v1/ocr" {
			t.Errorf("unexpected path: %q", request.URL.Path)
		}

		if request.Header.Get("x-godarai-tenant") != "tenant" {
			t.Errorf("unexpected tenant header: %q", request.Header.Get("x-godarai-tenant"))
		}

		var payload map[string]any
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Errorf("decode OCR request: %v", err)
		}

		document, ok := payload["document"].(map[string]any)

		documentURL, urlOK := document["document_url"].(string)
		if !ok || !urlOK || !strings.HasPrefix(documentURL, "data:application/pdf;base64,") {
			t.Errorf("source was not sent as a private PDF data URL: %#v", payload["document"])
		}

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{
			"model":"mistral-ocr-4-0",
			"pages":[{"index":0,"markdown":"# رسید\nمبلغ ۱۰۰٬۰۰۰ ریال"}],
			"usage_info":{"pages_processed":1}
		}`))
	}))
	defer server.Close()

	client, err := NewGatewayClient(server.URL, "secret", "cohere:platform:rerank-multilingual-v3.0", time.Second, SemanticEmbeddingConfig{
		BaseURL: server.URL, APIKey: "embedding-secret", Model: "Bge-m3", Revision: "v1",
		Dimensions: 1024, Timeout: time.Second, MaxBatchSize: 16,
	})
	require.NoError(t, err)

	document, pages, err := client.OCR(context.Background(), "tenant", "ocr-model", "application/pdf", []byte("PDF"))
	require.NoError(t, err)
	require.Equal(t, 1, pages)
	require.Equal(t, "رسید", document.Blocks[0].Text)
	require.Equal(t, "مبلغ ۱۰۰٬۰۰۰ ریال", document.Blocks[1].Text)
}

func TestGatewayRerankUsesConfiguredDeploymentInsteadOfProfileAlias(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var payload contractsopenai.RerankRequest
		require.NoError(t, json.NewDecoder(request.Body).Decode(&payload))
		require.Equal(t, "cohere:platform:rerank-multilingual-v3.0", payload.Model)

		writer.Header().Set("Content-Type", "application/json")
		_, err := writer.Write([]byte(`{"results":[{"index":0,"relevance_score":0.9}]}`))
		require.NoError(t, err)
	}))
	defer server.Close()

	client, err := NewGatewayClient(
		server.URL, "secret", "cohere:platform:rerank-multilingual-v3.0", time.Second,
		SemanticEmbeddingConfig{
			BaseURL: server.URL, APIKey: "embedding-secret", Model: "Bge-m3", Revision: "v1",
			Dimensions: 1024, Timeout: time.Second, MaxBatchSize: 16,
		},
	)
	require.NoError(t, err)

	results, _, err := client.Rerank(context.Background(), "tenant", "profile-alias", "query", []string{"document"}, 1)
	require.NoError(t, err)
	require.Len(t, results, 1)
}

func TestValidateEmbeddingVectorRejectsNonFiniteValues(t *testing.T) {
	t.Parallel()

	require.Error(t, validateEmbeddingVector([]float32{1, float32(math.Inf(1))}))
}

func TestValidateRerankResultsRejectsDuplicateIndices(t *testing.T) {
	t.Parallel()

	err := validateRerankResults([]contractsopenai.RerankResult{
		{Index: 0, RelevanceScore: 0.9},
		{Index: 0, RelevanceScore: 0.8},
	}, 2, 2)
	require.ErrorContains(t, err, "duplicated")
}

func TestValidateRerankResultsRejectsNonFiniteScore(t *testing.T) {
	t.Parallel()

	err := validateRerankResults([]contractsopenai.RerankResult{{
		Index: 0, RelevanceScore: math.NaN(),
	}}, 1, 1)
	require.ErrorContains(t, err, "not finite")
}
