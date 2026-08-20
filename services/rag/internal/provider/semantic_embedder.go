package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
)

// SemanticEmbeddingConfig points RAG at the same BGE endpoint used by Semantic Cache.
type SemanticEmbeddingConfig struct {
	BaseURL      string
	APIKey       string
	Model        string
	Revision     string
	Dimensions   int
	Timeout      time.Duration
	MaxBatchSize int
}

type semanticEmbedder struct {
	client *http.Client
	config SemanticEmbeddingConfig
	url    string
}

func (e *semanticEmbedder) matches(model string, dimensions int) bool {
	return model == e.config.Model && dimensions == e.config.Dimensions
}

func newSemanticEmbedder(config SemanticEmbeddingConfig) (*semanticEmbedder, error) {
	config.BaseURL = strings.TrimRight(strings.TrimSpace(config.BaseURL), "/")
	config.APIKey = strings.TrimSpace(config.APIKey)
	config.Model = strings.TrimSpace(config.Model)
	config.Revision = strings.TrimSpace(config.Revision)

	parsed, err := url.ParseRequestURI(config.BaseURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return nil, errors.New("semantic embedding base URL must be an absolute HTTP(S) URL")
	}

	if config.APIKey == "" || config.Model == "" || config.Revision == "" || config.Dimensions <= 0 {
		return nil, errors.New("semantic embedding credentials and identity are required")
	}

	if config.Timeout <= 0 {
		config.Timeout = 10 * time.Second
	}

	if config.MaxBatchSize <= 0 {
		config.MaxBatchSize = 16
	}

	return &semanticEmbedder{
		client: &http.Client{Timeout: config.Timeout}, config: config, url: config.BaseURL + "/embeddings",
	}, nil
}

func (e *semanticEmbedder) embed(ctx context.Context, model string, inputs []string, dimensions int) ([][]float32, int, error) {
	if model != e.config.Model || dimensions != e.config.Dimensions {
		return nil, 0, fmt.Errorf(
			"RAG profile embedding identity %s/%d does not match Semantic Cache identity %s/%d",
			model, dimensions, e.config.Model, e.config.Dimensions,
		)
	}

	if len(inputs) == 0 {
		return nil, 0, errors.New("embedding input is required")
	}

	vectors := make([][]float32, 0, len(inputs))
	totalTokens := 0

	for start := 0; start < len(inputs); start += e.config.MaxBatchSize {
		end := min(start+e.config.MaxBatchSize, len(inputs))

		batch, tokens, err := e.embedBatch(ctx, inputs[start:end])
		if err != nil {
			return nil, totalTokens, err
		}

		vectors = append(vectors, batch...)
		totalTokens += tokens
	}

	return vectors, totalTokens, nil
}

func (e *semanticEmbedder) embedBatch(ctx context.Context, inputs []string) ([][]float32, int, error) {
	for _, input := range inputs {
		if strings.TrimSpace(input) == "" {
			return nil, 0, errors.New("embedding input must not be empty")
		}
	}

	body, err := json.Marshal(map[string]any{
		"model": e.config.Model, "input": inputs, "encoding_format": "float",
	})
	if err != nil {
		return nil, 0, fmt.Errorf("encode semantic embedding request: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, e.url, bytes.NewReader(body))
	if err != nil {
		return nil, 0, errors.New("create semantic embedding request")
	}

	request.Header.Set("Authorization", "Bearer "+e.config.APIKey)
	request.Header.Set("Content-Type", "application/json")

	response, err := e.client.Do(request)
	if err != nil {
		if ctx.Err() != nil {
			return nil, 0, ctx.Err()
		}

		return nil, 0, errors.New("semantic embedding request failed")
	}
	defer closeResponseBody(response.Body)

	raw, err := io.ReadAll(io.LimitReader(response.Body, maxProviderResponseBytes+1))
	if err != nil {
		return nil, 0, fmt.Errorf("read semantic embedding response: %w", err)
	}

	if len(raw) > maxProviderResponseBytes {
		return nil, 0, errors.New("semantic embedding response exceeds size limit")
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, 0, fmt.Errorf("semantic embedding provider returned HTTP %d", response.StatusCode)
	}

	var payload contractsopenai.EmbeddingResponse
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, 0, fmt.Errorf("decode semantic embedding response: %w", err)
	}

	sort.Slice(payload.Data, func(i, j int) bool { return payload.Data[i].Index < payload.Data[j].Index })

	if len(payload.Data) != len(inputs) {
		return nil, 0, fmt.Errorf("semantic embedding count mismatch: got %d, want %d", len(payload.Data), len(inputs))
	}

	vectors := make([][]float32, len(inputs))

	for index, item := range payload.Data {
		if item.Index != index {
			return nil, 0, errors.New("semantic embedding response contains invalid indexes")
		}

		if err := json.Unmarshal(item.Embedding, &vectors[index]); err != nil {
			return nil, 0, fmt.Errorf("decode semantic embedding %d: %w", index, err)
		}

		if len(vectors[index]) != e.config.Dimensions {
			return nil, 0, fmt.Errorf("semantic embedding %d has unexpected dimensions: %d", index, len(vectors[index]))
		}

		for _, value := range vectors[index] {
			if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
				return nil, 0, fmt.Errorf("semantic embedding %d contains a non-finite value", index)
			}
		}
	}

	tokens := 0
	if payload.Usage != nil {
		tokens = payload.Usage.TotalTokens
	}

	return vectors, tokens, nil
}
