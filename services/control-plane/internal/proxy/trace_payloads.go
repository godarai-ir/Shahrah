package proxy

import (
	"encoding/json"
	"strings"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
)

func buildTraceChatRequestPayload(req contractsopenai.ChatCompletionRequest) map[string]interface{} {
	messages := make([]map[string]interface{}, 0, len(req.Messages))
	for _, msg := range req.Messages {
		messages = append(messages, map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
			"name":    msg.Name,
		})
	}
	return map[string]interface{}{
		"model":       req.Model,
		"messages":    messages,
		"max_tokens":  req.MaxTokens,
		"temperature": req.Temperature,
		"top_p":       req.TopP,
		"stream":      req.Stream,
	}
}

func buildTraceCompletionRequestPayload(req contractsopenai.CompletionRequest) map[string]interface{} {
	return map[string]interface{}{
		"model":       req.Model,
		"prompt":      normalizePrompt(req.Prompt),
		"max_tokens":  req.MaxTokens,
		"temperature": req.Temperature,
		"top_p":       req.TopP,
		"stream":      req.Stream,
	}
}

func buildTraceChatResponsePayload(resp contractsopenai.ChatCompletionResponse) map[string]interface{} {
	choices := make([]map[string]interface{}, 0, len(resp.Choices))
	for _, choice := range resp.Choices {
		content := ""
		role := ""
		if choice.Message != nil {
			content = choice.Message.Content
			role = choice.Message.Role
		}
		choices = append(choices, map[string]interface{}{
			"index":         choice.Index,
			"role":          role,
			"content":       content,
			"finish_reason": choice.FinishReason,
		})
	}
	usage := map[string]interface{}{}
	if resp.Usage != nil {
		usage["prompt_tokens"] = resp.Usage.PromptTokens
		usage["completion_tokens"] = resp.Usage.CompletionTokens
		usage["total_tokens"] = resp.Usage.TotalTokens
	}
	return map[string]interface{}{
		"id":      resp.ID,
		"model":   resp.Model,
		"choices": choices,
		"usage":   usage,
	}
}

func buildTraceCompletionResponsePayload(resp contractsopenai.CompletionResponse) map[string]interface{} {
	choices := make([]map[string]interface{}, 0, len(resp.Choices))
	for _, choice := range resp.Choices {
		choices = append(choices, map[string]interface{}{
			"index":         choice.Index,
			"text":          choice.Text,
			"finish_reason": choice.FinishReason,
		})
	}
	usage := map[string]interface{}{}
	if resp.Usage != nil {
		usage["prompt_tokens"] = resp.Usage.PromptTokens
		usage["completion_tokens"] = resp.Usage.CompletionTokens
		usage["total_tokens"] = resp.Usage.TotalTokens
	}
	return map[string]interface{}{
		"id":      resp.ID,
		"model":   resp.Model,
		"choices": choices,
		"usage":   usage,
	}
}

func buildTraceProviderErrorPayload(statusCode int, body []byte) map[string]interface{} {
	payload := map[string]interface{}{
		"status_code": statusCode,
	}
	if len(body) == 0 {
		return payload
	}
	var decoded interface{}
	if err := json.Unmarshal(body, &decoded); err == nil {
		payload["body"] = decoded
		return payload
	}
	bodyText := strings.TrimSpace(string(body))
	if len(bodyText) > 2048 {
		bodyText = bodyText[:2048]
	}
	payload["body"] = bodyText
	return payload
}
