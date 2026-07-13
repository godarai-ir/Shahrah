package server

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
)

const anthropicDefaultVersion = "2023-06-01"

type anthropicAdapter struct{}

type anthropicRequest struct {
	Model       string                  `json:"model"`
	Messages    []anthropicMessage      `json:"messages"`
	MaxTokens   int                     `json:"max_tokens,omitempty"`
	Temperature *float64                `json:"temperature,omitempty"`
	TopP        *float64                `json:"top_p,omitempty"`
	Stream      bool                    `json:"stream,omitempty"`
	System      []anthropicContentBlock `json:"system,omitempty"`
}

type anthropicMessage struct {
	Role    string                  `json:"role"`
	Content []anthropicContentBlock `json:"content"`
}

type anthropicContentBlock struct {
	Type   string           `json:"type"`
	Text   string           `json:"text,omitempty"`
	Source *anthropicSource `json:"source,omitempty"`
}

type anthropicSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type anthropicResponse struct {
	ID         string                  `json:"id"`
	Type       string                  `json:"type"`
	Role       string                  `json:"role"`
	Content    []anthropicContentBlock `json:"content"`
	Model      string                  `json:"model"`
	Usage      anthropicUsage          `json:"usage"`
	StopReason string                  `json:"stop_reason,omitempty"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type anthropicStreamEvent struct {
	Type    string `json:"type"`
	Index   int    `json:"index,omitempty"`
	Message struct {
		ID    string `json:"id"`
		Model string `json:"model"`
	} `json:"message,omitempty"`
	Delta struct {
		Text       string `json:"text,omitempty"`
		StopReason string `json:"stop_reason,omitempty"`
	} `json:"delta,omitempty"`
}

func newAnthropicAdapter() providerAdapter {
	return &anthropicAdapter{}
}

func (a *anthropicAdapter) Dispatch(ctx context.Context, client *http.Client, payload platform.TunnelRequest) (*adapterResponse, error) {
	target, err := resolveBaseURL(payload.Deployment.APIBase, "https://api.anthropic.com")
	if err != nil {
		return nil, err
	}

	var openAIReq contractsopenai.ChatCompletionRequest
	if err := json.Unmarshal(payload.Body, &openAIReq); err != nil {
		return nil, fmt.Errorf("decode openai request: %w", err)
	}

	anthropicReq, err := convertToAnthropicRequest(openAIReq)
	if err != nil {
		return nil, fmt.Errorf("convert anthropic request: %w", err)
	}
	body, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("encode anthropic request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, joinURL(target, "/v1/messages"), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", payload.Deployment.APIKey)
	req.Header.Set("anthropic-version", anthropicVersion(payload.Deployment.APIVersion))
	if payload.Stream {
		req.Header.Set("Accept", "text/event-stream")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return &adapterResponse{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header.Clone(),
			Body:       resp.Body,
		}, nil
	}

	if payload.Stream {
		return adaptAnthropicStreamResponse(resp, openAIReq.Model), nil
	}
	return adaptAnthropicSyncResponse(resp, openAIReq.Model)
}

func adaptAnthropicSyncResponse(resp *http.Response, model string) (*adapterResponse, error) {
	defer resp.Body.Close()

	var anthResp anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthResp); err != nil {
		return nil, fmt.Errorf("decode anthropic response: %w", err)
	}
	openAIResp := convertToOpenAIResponse(anthResp, model)
	body, err := json.Marshal(openAIResp)
	if err != nil {
		return nil, fmt.Errorf("encode openai response: %w", err)
	}

	headers := resp.Header.Clone()
	headers.Set("Content-Type", "application/json")
	headers.Set("Content-Length", fmt.Sprintf("%d", len(body)))
	return &adapterResponse{
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       io.NopCloser(bytes.NewReader(body)),
	}, nil
}

func adaptAnthropicStreamResponse(resp *http.Response, model string) *adapterResponse {
	headers := resp.Header.Clone()
	headers.Set("Content-Type", "text/event-stream")
	headers.Del("Content-Length")

	pipeReader, pipeWriter := io.Pipe()
	go func() {
		defer resp.Body.Close()
		defer pipeWriter.Close()

		reader := bufio.NewReader(resp.Body)
		var currentEvent string
		streamID := ""
		streamModel := model

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					_, _ = pipeWriter.Write([]byte("data: [DONE]\n\n"))
					return
				}
				_ = pipeWriter.CloseWithError(err)
				return
			}

			trimmed := strings.TrimSpace(line)
			switch {
			case trimmed == "":
				currentEvent = ""
			case strings.HasPrefix(trimmed, "event:"):
				currentEvent = strings.TrimSpace(strings.TrimPrefix(trimmed, "event:"))
			case strings.HasPrefix(trimmed, "data:"):
				data := strings.TrimSpace(strings.TrimPrefix(trimmed, "data:"))
				if data == "" || data == "[DONE]" {
					continue
				}
				var evt anthropicStreamEvent
				if err := json.Unmarshal([]byte(data), &evt); err != nil {
					continue
				}
				if evt.Message.ID != "" {
					streamID = evt.Message.ID
				}
				if evt.Message.Model != "" {
					streamModel = evt.Message.Model
				}
				switch eventType(currentEvent, evt.Type) {
				case "content_block_delta":
					if evt.Delta.Text == "" {
						continue
					}
					chunk := contractsopenai.StreamChunk{
						ID:      streamID,
						Object:  "chat.completion.chunk",
						Created: time.Now().Unix(),
						Model:   streamModel,
						Choices: []contractsopenai.Choice{{
							Index: evt.Index,
							Delta: &contractsopenai.ChatMessage{Content: evt.Delta.Text},
						}},
					}
					writeSSEChunk(pipeWriter, chunk)
				case "message_delta":
					if evt.Delta.StopReason == "" {
						continue
					}
					chunk := contractsopenai.StreamChunk{
						ID:      streamID,
						Object:  "chat.completion.chunk",
						Created: time.Now().Unix(),
						Model:   streamModel,
						Choices: []contractsopenai.Choice{{
							Index:        0,
							Delta:        &contractsopenai.ChatMessage{},
							FinishReason: evt.Delta.StopReason,
						}},
					}
					writeSSEChunk(pipeWriter, chunk)
				}
			}
		}
	}()

	return &adapterResponse{
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       pipeReader,
	}
}

func writeSSEChunk(w io.Writer, chunk contractsopenai.StreamChunk) {
	encoded, err := json.Marshal(chunk)
	if err != nil {
		return
	}
	_, _ = w.Write([]byte("data: "))
	_, _ = w.Write(encoded)
	_, _ = w.Write([]byte("\n\n"))
}

func eventType(explicit string, fallback string) string {
	if strings.TrimSpace(explicit) != "" {
		return strings.TrimSpace(explicit)
	}
	return strings.TrimSpace(fallback)
}

func anthropicVersion(version string) string {
	if strings.TrimSpace(version) == "" {
		return anthropicDefaultVersion
	}
	return strings.TrimSpace(version)
}

func convertToAnthropicRequest(req contractsopenai.ChatCompletionRequest) (anthropicRequest, error) {
	anthReq := anthropicRequest{
		Model:       canonicalAnthropicModel(req.Model),
		MaxTokens:   4096,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stream:      req.Stream,
	}

	for _, msg := range req.Messages {
		blocks, err := anthropicBlocksFromContent(msg.Content)
		if err != nil {
			return anthropicRequest{}, err
		}
		if msg.Role == "system" {
			for _, block := range blocks {
				if block.Type != "text" {
					return anthropicRequest{}, fmt.Errorf("system messages only support text content")
				}
				anthReq.System = append(anthReq.System, block)
			}
			continue
		}
		anthReq.Messages = append(anthReq.Messages, anthropicMessage{
			Role:    msg.Role,
			Content: blocks,
		})
	}

	if req.MaxTokens != nil && *req.MaxTokens > 0 {
		anthReq.MaxTokens = *req.MaxTokens
	}

	return anthReq, nil
}

func anthropicBlocksFromContent(content interface{}) ([]anthropicContentBlock, error) {
	parts, ok, err := contractsopenai.ChatContentParts(content)
	if err != nil {
		return nil, err
	}
	if !ok {
		text := contractsopenai.ChatContentText(content)
		if text == "" {
			text = " "
		}
		return []anthropicContentBlock{{Type: "text", Text: text}}, nil
	}
	blocks := make([]anthropicContentBlock, 0, len(parts))
	for _, part := range parts {
		switch part.Type {
		case "text":
			blocks = append(blocks, anthropicContentBlock{Type: "text", Text: part.Text})
		case "image_url":
			mediaType, data, err := contractsopenai.ParseImageDataURL(part.ImageURL.URL)
			if err != nil {
				return nil, err
			}
			blocks = append(blocks, anthropicContentBlock{
				Type: "image",
				Source: &anthropicSource{
					Type:      "base64",
					MediaType: mediaType,
					Data:      data,
				},
			})
		case "file":
			if strings.TrimSpace(part.File.FileID) != "" {
				return nil, fmt.Errorf("file.file_id is not supported for anthropic")
			}
			mediaType, data, err := contractsopenai.ParseFileDataURL(part.File.FileData)
			if err != nil {
				return nil, err
			}
			blocks = append(blocks, anthropicContentBlock{
				Type: "document",
				Source: &anthropicSource{
					Type:      "base64",
					MediaType: mediaType,
					Data:      data,
				},
			})
		}
	}
	if len(blocks) == 0 {
		blocks = append(blocks, anthropicContentBlock{Type: "text", Text: " "})
	}
	return blocks, nil
}

func canonicalAnthropicModel(model string) string {
	model = strings.TrimSpace(model)
	if model == "" {
		return ""
	}
	if strings.Count(model, ":") == 2 {
		parts := strings.SplitN(model, ":", 3)
		model = strings.TrimSpace(parts[2])
	}
	const providerPrefix = "anthropic/"
	if strings.HasPrefix(strings.ToLower(model), providerPrefix) {
		return strings.TrimSpace(model[len(providerPrefix):])
	}
	return model
}

func convertToOpenAIResponse(resp anthropicResponse, model string) *contractsopenai.ChatCompletionResponse {
	content := ""
	for _, block := range resp.Content {
		content += block.Text
	}

	return &contractsopenai.ChatCompletionResponse{
		ID:      resp.ID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []contractsopenai.Choice{{
			Index: 0,
			Message: &contractsopenai.ChatMessage{
				Role:    "assistant",
				Content: content,
			},
			FinishReason: resp.StopReason,
		}},
		Usage: &contractsopenai.Usage{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.InputTokens + resp.Usage.OutputTokens,
		},
	}
}
