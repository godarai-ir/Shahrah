package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
)

type geminiAdapter struct{}

type geminiRequest struct {
	Contents         []geminiContent        `json:"contents"`
	GenerationConfig geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text       string            `json:"text,omitempty"`
	InlineData *geminiInlineData `json:"inline_data,omitempty"`
}

type geminiInlineData struct {
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

type geminiGenerationConfig struct {
	Temperature     *float64 `json:"temperature,omitempty"`
	MaxOutputTokens *int     `json:"maxOutputTokens,omitempty"`
	TopP            *float64 `json:"topP,omitempty"`
}

type geminiResponse struct {
	Candidates    []geminiCandidate `json:"candidates"`
	UsageMetadata geminiUsage       `json:"usageMetadata,omitempty"`
}

type geminiCandidate struct {
	Content      geminiContent `json:"content"`
	FinishReason string        `json:"finishReason,omitempty"`
}

type geminiUsage struct {
	PromptTokenCount     int `json:"promptTokenCount,omitempty"`
	CandidatesTokenCount int `json:"candidatesTokenCount,omitempty"`
	TotalTokenCount      int `json:"totalTokenCount,omitempty"`
}

func newGeminiAdapter() providerAdapter {
	return &geminiAdapter{}
}

func (a *geminiAdapter) Dispatch(ctx context.Context, client *http.Client, payload platform.TunnelRequest) (*adapterResponse, error) {
	if payload.Stream {
		return nil, fmt.Errorf("gemini streaming is not supported through Shahrah chat completions")
	}
	target, err := resolveBaseURL(payload.Deployment.APIBase, "https://generativelanguage.googleapis.com/v1")
	if err != nil {
		return nil, err
	}

	var openAIReq contractsopenai.ChatCompletionRequest
	if err := json.Unmarshal(payload.Body, &openAIReq); err != nil {
		return nil, fmt.Errorf("decode openai request: %w", err)
	}

	geminiReq, err := convertToGeminiRequest(openAIReq)
	if err != nil {
		return nil, fmt.Errorf("convert gemini request: %w", err)
	}
	body, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, fmt.Errorf("encode gemini request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/models/%s:generateContent", strings.TrimRight(target, "/"), url.PathEscape(canonicalGeminiModel(openAIReq.Model)))
	if payload.Deployment.APIKey != "" {
		endpoint += "?key=" + url.QueryEscape(payload.Deployment.APIKey)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return &adapterResponse{StatusCode: resp.StatusCode, Headers: resp.Header.Clone(), Body: resp.Body}, nil
	}
	return adaptGeminiSyncResponse(resp, openAIReq.Model)
}

func convertToGeminiRequest(req contractsopenai.ChatCompletionRequest) (geminiRequest, error) {
	out := geminiRequest{Contents: make([]geminiContent, 0, len(req.Messages))}
	for _, msg := range req.Messages {
		role := msg.Role
		switch role {
		case "assistant":
			role = "model"
		case "system":
			role = "user"
		}
		parts, err := geminiPartsFromContent(msg.Content)
		if err != nil {
			return geminiRequest{}, err
		}
		out.Contents = append(out.Contents, geminiContent{Role: role, Parts: parts})
	}
	out.GenerationConfig = geminiGenerationConfig{
		Temperature:     req.Temperature,
		TopP:            req.TopP,
		MaxOutputTokens: req.MaxTokens,
	}
	return out, nil
}

func geminiPartsFromContent(content interface{}) ([]geminiPart, error) {
	parts, ok, err := contractsopenai.ChatContentParts(content)
	if err != nil {
		return nil, err
	}
	if !ok {
		text := contractsopenai.ChatContentText(content)
		if text == "" {
			text = " "
		}
		return []geminiPart{{Text: text}}, nil
	}
	out := make([]geminiPart, 0, len(parts))
	for _, part := range parts {
		switch part.Type {
		case "text":
			out = append(out, geminiPart{Text: part.Text})
		case "image_url":
			mediaType, data, err := contractsopenai.ParseImageDataURL(part.ImageURL.URL)
			if err != nil {
				return nil, err
			}
			out = append(out, geminiPart{InlineData: &geminiInlineData{MimeType: mediaType, Data: data}})
		}
	}
	if len(out) == 0 {
		out = append(out, geminiPart{Text: " "})
	}
	return out, nil
}

func adaptGeminiSyncResponse(resp *http.Response, model string) (*adapterResponse, error) {
	defer resp.Body.Close()
	var gemResp geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&gemResp); err != nil {
		return nil, fmt.Errorf("decode gemini response: %w", err)
	}
	openAIResp := convertGeminiToOpenAIResponse(gemResp, model)
	body, err := json.Marshal(openAIResp)
	if err != nil {
		return nil, fmt.Errorf("encode openai response: %w", err)
	}
	headers := resp.Header.Clone()
	headers.Set("Content-Type", "application/json")
	headers.Set("Content-Length", fmt.Sprintf("%d", len(body)))
	return &adapterResponse{StatusCode: resp.StatusCode, Headers: headers, Body: io.NopCloser(bytes.NewReader(body))}, nil
}

func convertGeminiToOpenAIResponse(resp geminiResponse, model string) *contractsopenai.ChatCompletionResponse {
	content := ""
	finishReason := "stop"
	for _, candidate := range resp.Candidates {
		for _, part := range candidate.Content.Parts {
			content += part.Text
		}
		if candidate.FinishReason != "" {
			finishReason = candidate.FinishReason
		}
	}
	total := resp.UsageMetadata.TotalTokenCount
	if total == 0 {
		total = resp.UsageMetadata.PromptTokenCount + resp.UsageMetadata.CandidatesTokenCount
	}
	return &contractsopenai.ChatCompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-%d", time.Now().Unix()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []contractsopenai.Choice{{
			Index: 0,
			Message: &contractsopenai.ChatMessage{
				Role:    "assistant",
				Content: content,
			},
			FinishReason: finishReason,
		}},
		Usage: &contractsopenai.Usage{
			PromptTokens:     resp.UsageMetadata.PromptTokenCount,
			CompletionTokens: resp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      total,
		},
	}
}

func canonicalGeminiModel(model string) string {
	model = strings.TrimSpace(model)
	if strings.Count(model, ":") == 2 {
		parts := strings.SplitN(model, ":", 3)
		model = strings.TrimSpace(parts[2])
	}
	const providerPrefix = "gemini/"
	if strings.HasPrefix(strings.ToLower(model), providerPrefix) {
		return strings.TrimSpace(model[len(providerPrefix):])
	}
	return model
}
