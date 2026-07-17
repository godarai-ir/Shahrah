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
	Model       string                          `json:"model"`
	Messages    []anthropicMessage              `json:"messages"`
	MaxTokens   int                             `json:"max_tokens,omitempty"`
	Temperature *float64                        `json:"temperature,omitempty"`
	TopP        *float64                        `json:"top_p,omitempty"`
	Stream      bool                            `json:"stream,omitempty"`
	System      []anthropicContentBlock         `json:"system,omitempty"`
	Thinking    *contractsopenai.ThinkingConfig `json:"thinking,omitempty"`
}

type anthropicMessage struct {
	Role    string                  `json:"role"`
	Content []anthropicContentBlock `json:"content"`
}

type anthropicContentBlock struct {
	Type      string           `json:"type"`
	Text      string           `json:"text,omitempty"`
	Thinking  string           `json:"thinking,omitempty"`
	Signature string           `json:"signature,omitempty"`
	Data      string           `json:"data,omitempty"`
	Source    *anthropicSource `json:"source,omitempty"`
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
		Type       string `json:"type,omitempty"`
		Text       string `json:"text,omitempty"`
		Thinking   string `json:"thinking,omitempty"`
		Signature  string `json:"signature,omitempty"`
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
	req.Header.Set("x-api-key", deploymentAPIKey(payload.Deployment))
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
					if evt.Delta.Text == "" && evt.Delta.Thinking == "" && evt.Delta.Signature == "" {
						continue
					}
					delta := &contractsopenai.ChatMessage{}
					if evt.Delta.Thinking != "" {
						delta.ReasoningContent = evt.Delta.Thinking
					}
					if evt.Delta.Signature != "" {
						delta.ThinkingBlocks = []contractsopenai.ThinkingBlock{{
							"type":      "thinking",
							"signature": evt.Delta.Signature,
						}}
					}
					if evt.Delta.Text != "" {
						delta.Content = evt.Delta.Text
					}
					chunk := contractsopenai.StreamChunk{
						ID:      streamID,
						Object:  "chat.completion.chunk",
						Created: time.Now().Unix(),
						Model:   streamModel,
						Choices: []contractsopenai.Choice{{
							Index: evt.Index,
							Delta: delta,
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
		if strings.EqualFold(strings.TrimSpace(msg.Role), "assistant") && len(msg.ThinkingBlocks) > 0 {
			thinkingBlocks := anthropicBlocksFromThinkingBlocks(msg.ThinkingBlocks)
			blocks = append(thinkingBlocks, blocks...)
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

	anthReq.Thinking = anthropicThinkingConfig(req, anthReq.MaxTokens)

	return anthReq, nil
}

func anthropicBlocksFromThinkingBlocks(blocks []contractsopenai.ThinkingBlock) []anthropicContentBlock {
	out := make([]anthropicContentBlock, 0, len(blocks))
	for _, block := range blocks {
		blockType, _ := block["type"].(string)
		blockType = strings.TrimSpace(blockType)
		switch blockType {
		case "thinking":
			item := anthropicContentBlock{Type: "thinking"}
			if thinking, ok := block["thinking"].(string); ok {
				item.Thinking = thinking
			}
			if signature, ok := block["signature"].(string); ok {
				item.Signature = signature
			}
			out = append(out, item)
		case "redacted_thinking":
			item := anthropicContentBlock{Type: "redacted_thinking"}
			if data, ok := block["data"].(string); ok {
				item.Data = data
			}
			out = append(out, item)
		}
	}
	return out
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

func anthropicThinkingConfig(req contractsopenai.ChatCompletionRequest, maxTokens int) *contractsopenai.ThinkingConfig {
	if req.Thinking != nil {
		if strings.EqualFold(strings.TrimSpace(req.Thinking.Type), "disabled") {
			return nil
		}
		cfg := *req.Thinking
		if strings.TrimSpace(cfg.Type) == "" {
			cfg.Type = "enabled"
		}
		if cfg.BudgetTokens <= 0 {
			cfg.BudgetTokens = defaultAnthropicThinkingBudget(req.ReasoningEffort, maxTokens)
		}
		return &cfg
	}

	effort := strings.ToLower(strings.TrimSpace(req.ReasoningEffort))
	if effort == "" || effort == "none" {
		return nil
	}

	return &contractsopenai.ThinkingConfig{
		Type:         "enabled",
		BudgetTokens: defaultAnthropicThinkingBudget(effort, maxTokens),
	}
}

func defaultAnthropicThinkingBudget(effort string, maxTokens int) int {
	budget := 2048
	switch strings.ToLower(strings.TrimSpace(effort)) {
	case "minimal", "low":
		budget = 1024
	case "medium", "":
		budget = 2048
	case "high":
		budget = 3200
	case "xhigh", "max":
		budget = 8192
	}
	if maxTokens > 1024 && budget >= maxTokens {
		return maxTokens - 1
	}
	return budget
}

func convertToOpenAIResponse(resp anthropicResponse, model string) *contractsopenai.ChatCompletionResponse {
	content := ""
	reasoningContent := ""
	thinkingBlocks := make([]contractsopenai.ThinkingBlock, 0)
	for _, block := range resp.Content {
		switch block.Type {
		case "text":
			content += block.Text
		case "thinking":
			thinkingBlock := contractsopenai.ThinkingBlock{"type": block.Type}
			if block.Thinking != "" {
				thinkingBlock["thinking"] = block.Thinking
				reasoningContent += block.Thinking
			}
			if block.Signature != "" {
				thinkingBlock["signature"] = block.Signature
			}
			thinkingBlocks = append(thinkingBlocks, thinkingBlock)
		case "redacted_thinking":
			thinkingBlock := contractsopenai.ThinkingBlock{"type": block.Type}
			if block.Data != "" {
				thinkingBlock["data"] = block.Data
			}
			thinkingBlocks = append(thinkingBlocks, thinkingBlock)
		default:
			content += block.Text
		}
	}

	message := &contractsopenai.ChatMessage{
		Role:    "assistant",
		Content: content,
	}
	if reasoningContent != "" {
		message.ReasoningContent = reasoningContent
	}
	if len(thinkingBlocks) > 0 {
		message.ThinkingBlocks = thinkingBlocks
	}

	return &contractsopenai.ChatCompletionResponse{
		ID:      resp.ID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []contractsopenai.Choice{{
			Index:        0,
			Message:      message,
			FinishReason: resp.StopReason,
		}},
		Usage: &contractsopenai.Usage{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.InputTokens + resp.Usage.OutputTokens,
		},
	}
}
