// Package provider implements the model-provider boundary used by managed RAG.
package provider

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"strings"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/services/rag/internal/app"
	"github.com/Delpak/shahrah/services/rag/internal/domain"
	"github.com/Delpak/shahrah/services/rag/internal/parser"
)

const maxProviderResponseBytes = 16 << 20

const maxOCRInputBytes = 25 << 20

// GatewayClient calls the existing Godar gateway so routing, wallet settlement, and provider metering stay centralized.
type GatewayClient struct {
	baseURL     string
	apiKey      string
	rerankModel string
	client      *http.Client
	ocrClient   *http.Client
	embedder    *semanticEmbedder
}

// NewGatewayClient creates a fail-closed provider client.
func NewGatewayClient(
	baseURL, apiKey, rerankModel string,
	timeout time.Duration,
	embeddingConfig SemanticEmbeddingConfig,
) (*GatewayClient, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")

	apiKey = strings.TrimSpace(apiKey)

	rerankModel = strings.TrimSpace(rerankModel)
	if baseURL == "" || apiKey == "" || rerankModel == "" {
		return nil, errors.New("RAG gateway base URL, API key, and rerank model are required")
	}

	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	ocrTimeout := max(timeout, 2*time.Minute)

	embedder, err := newSemanticEmbedder(embeddingConfig)
	if err != nil {
		return nil, err
	}

	return &GatewayClient{
		baseURL:     baseURL,
		apiKey:      apiKey,
		rerankModel: rerankModel,
		client:      &http.Client{Timeout: timeout},
		ocrClient:   &http.Client{Timeout: ocrTimeout},
		embedder:    embedder,
	}, nil
}

// Embed returns one vector per input and the authoritative provider token count.
func (c *GatewayClient) Embed(ctx context.Context, tenantSlug, model string, inputs []string, dimensions int) ([][]float32, int, error) {
	if c.embedder.matches(model, dimensions) {
		return c.embedder.embed(ctx, model, inputs, dimensions)
	}

	payload := map[string]any{
		"model":           model,
		"input":           inputs,
		"encoding_format": "float",
		"dimensions":      dimensions,
	}

	var response contractsopenai.EmbeddingResponse
	if err := c.post(ctx, tenantSlug, "/v1/embeddings", payload, &response); err != nil {
		return nil, 0, fmt.Errorf("embedding request: %w", err)
	}

	if len(response.Data) != len(inputs) {
		return nil, 0, fmt.Errorf("embedding count mismatch: got %d, want %d", len(response.Data), len(inputs))
	}

	vectors := make([][]float32, len(response.Data))
	seen := make(map[int]struct{}, len(response.Data))

	for _, item := range response.Data {
		if item.Index < 0 || item.Index >= len(vectors) {
			return nil, 0, fmt.Errorf("embedding index %d is out of range", item.Index)
		}

		if _, exists := seen[item.Index]; exists {
			return nil, 0, fmt.Errorf("embedding index %d is duplicated", item.Index)
		}

		var vector []float32
		if err := json.Unmarshal(item.Embedding, &vector); err != nil {
			return nil, 0, fmt.Errorf("decode embedding %d: %w", item.Index, err)
		}

		if len(vector) != dimensions {
			return nil, 0, fmt.Errorf("embedding %d has %d dimensions, want %d", item.Index, len(vector), dimensions)
		}

		if err := validateEmbeddingVector(vector); err != nil {
			return nil, 0, fmt.Errorf("embedding %d is invalid: %w", item.Index, err)
		}

		vectors[item.Index] = vector
		seen[item.Index] = struct{}{}
	}

	tokens := 0
	if response.Usage != nil {
		tokens = response.Usage.TotalTokens
	}

	return vectors, tokens, nil
}

// Rerank returns provider-ranked document indices and fails closed on malformed results.
func (c *GatewayClient) Rerank(ctx context.Context, tenantSlug, model, query string, documents []string, topN int) ([]contractsopenai.RerankResult, int, error) {
	if strings.TrimSpace(model) == "" {
		return nil, 0, errors.New("RAG profile rerank model is required")
	}

	payload := contractsopenai.RerankRequest{
		Model:     c.rerankModel,
		Query:     query,
		Documents: documents,
		TopN:      &topN,
	}

	var response contractsopenai.RerankResponse
	if err := c.post(ctx, tenantSlug, "/v1/rerank", payload, &response); err != nil {
		return nil, 0, fmt.Errorf("rerank request: %w", err)
	}

	if err := validateRerankResults(response.Results, len(documents), topN); err != nil {
		return nil, 0, err
	}

	units := 1
	if response.Usage != nil && response.Usage.TotalTokens > 0 {
		units = response.Usage.TotalTokens
	}

	return response.Results, units, nil
}

// OCR extracts page-scoped Persian-capable text through the metered Godar OCR route.
func (c *GatewayClient) OCR(
	ctx context.Context,
	tenantSlug, model, mediaType string,
	content []byte,
) (domain.SourceDocument, int, error) {
	if mediaType != "application/pdf" {
		return domain.SourceDocument{}, 0, errors.New("managed OCR currently supports PDF only")
	}

	if len(content) == 0 || len(content) > maxOCRInputBytes {
		return domain.SourceDocument{}, 0, errors.New("OCR input must be between 1 byte and 25 MiB")
	}

	payload := map[string]any{
		"model": model,
		"document": map[string]string{
			"type":         "document_url",
			"document_url": "data:" + mediaType + ";base64," + base64.StdEncoding.EncodeToString(content),
		},
		"include_image_base64":          false,
		"include_blocks":                true,
		"table_format":                  "markdown",
		"confidence_scores_granularity": "page",
	}

	var response struct {
		Model string `json:"model"`
		Pages []struct {
			Index    int    `json:"index"`
			Markdown string `json:"markdown"`
		} `json:"pages"`
		UsageInfo struct {
			PagesProcessed int `json:"pages_processed"`
		} `json:"usage_info"`
	}
	if err := c.postWithClient(ctx, c.ocrClient, tenantSlug, "/v1/ocr", payload, &response); err != nil {
		return domain.SourceDocument{}, 0, fmt.Errorf("OCR request: %w", err)
	}

	if strings.TrimSpace(response.Model) == "" {
		return domain.SourceDocument{}, 0, errors.New("OCR provider returned no model identity")
	}

	pages := make([]parser.OCRPage, len(response.Pages))
	for index, page := range response.Pages {
		pages[index] = parser.OCRPage{Index: page.Index, Markdown: page.Markdown}
	}

	document, err := parser.ParseOCRMarkdown(pages)
	if err != nil {
		return domain.SourceDocument{}, response.UsageInfo.PagesProcessed, fmt.Errorf("parse OCR response: %w", err)
	}

	processedPages := response.UsageInfo.PagesProcessed
	if processedPages < 1 {
		processedPages = len(response.Pages)
	}

	return document, processedPages, nil
}

// Generate creates a non-streaming answer. Citation validation is owned by the caller.
func (c *GatewayClient) Generate(
	ctx context.Context,
	tenantSlug, model, systemPrompt, userPrompt string,
) (string, string, app.GenerationUsage, error) {
	payload := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"stream":      false,
		"temperature": 0,
	}

	var response contractsopenai.ChatCompletionResponse
	if err := c.post(ctx, tenantSlug, "/v1/chat/completions", payload, &response); err != nil {
		return "", "", app.GenerationUsage{}, fmt.Errorf("generation request: %w", err)
	}

	if len(response.Choices) == 0 || response.Choices[0].Message == nil {
		return "", "", app.GenerationUsage{}, errors.New("generation provider returned no answer")
	}

	answer, ok := response.Choices[0].Message.Content.(string)
	if !ok || strings.TrimSpace(answer) == "" {
		return "", "", app.GenerationUsage{}, errors.New("generation provider returned an invalid answer")
	}

	if strings.TrimSpace(response.Model) == "" {
		return "", "", app.GenerationUsage{}, errors.New("generation provider returned no model identity")
	}

	usage := app.GenerationUsage{}
	if response.Usage != nil {
		usage.InputTokens = response.Usage.PromptTokens
		usage.OutputTokens = response.Usage.CompletionTokens
		usage.ProviderCostUSD = float64(response.Usage.CostInUSDTicks) / 10_000_000_000
	}

	return strings.TrimSpace(answer), strings.TrimSpace(response.Model), usage, nil
}

func validateEmbeddingVector(vector []float32) error {
	for _, value := range vector {
		if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
			return errors.New("vector contains a non-finite value")
		}
	}

	return nil
}

func validateRerankResults(results []contractsopenai.RerankResult, documentCount, topN int) error {
	if len(results) == 0 && documentCount > 0 {
		return errors.New("rerank provider returned no results")
	}

	if len(results) > topN {
		return fmt.Errorf("rerank provider returned %d results, maximum is %d", len(results), topN)
	}

	seen := make(map[int]struct{}, len(results))
	for _, result := range results {
		if result.Index < 0 || result.Index >= documentCount {
			return fmt.Errorf("rerank index %d is out of range", result.Index)
		}

		if _, exists := seen[result.Index]; exists {
			return fmt.Errorf("rerank index %d is duplicated", result.Index)
		}

		if math.IsNaN(result.RelevanceScore) || math.IsInf(result.RelevanceScore, 0) {
			return fmt.Errorf("rerank score for index %d is not finite", result.Index)
		}

		seen[result.Index] = struct{}{}
	}

	return nil
}

func (c *GatewayClient) post(ctx context.Context, tenantSlug, path string, payload, output any) error {
	return c.postWithClient(ctx, c.client, tenantSlug, path, payload, output)
}

func (c *GatewayClient) postWithClient(
	ctx context.Context,
	client *http.Client,
	tenantSlug, path string,
	payload, output any,
) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-godarai-tenant", strings.TrimSpace(tenantSlug))
	req.Header.Set("x-godarai-rag-internal", "1")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer closeResponseBody(resp.Body)

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, maxProviderResponseBytes+1))
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if len(responseBody) > maxProviderResponseBytes {
		return errors.New("provider response exceeds size limit")
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("provider returned HTTP %d", resp.StatusCode)
	}

	if err := json.Unmarshal(responseBody, output); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

func closeResponseBody(body io.Closer) {
	if err := body.Close(); err != nil {
		slog.Warn("managed RAG provider response close failed", "error_type", "response_close")
	}
}
