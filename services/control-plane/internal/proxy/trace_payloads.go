package proxy

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
)

func buildTraceChatRequestPayload(req contractsopenai.ChatCompletionRequest) map[string]interface{} {
	messages := make([]map[string]interface{}, 0, len(req.Messages))
	for _, msg := range req.Messages {
		messages = append(messages, map[string]interface{}{
			"role":              msg.Role,
			"content":           contractsopenai.RedactChatContent(msg.Content),
			"name":              msg.Name,
			"reasoning_content": reasoningContentSummary(msg.ReasoningContent),
			"thinking_blocks":   summarizeThinkingBlocks(msg.ThinkingBlocks),
			"tool_calls":        summarizeToolCalls(msg.ToolCalls),
			"tool_call_id":      msg.ToolCallID,
		})
	}

	return map[string]interface{}{
		"model":               req.Model,
		"messages":            messages,
		"max_tokens":          req.MaxTokens,
		"reasoning_effort":    req.ReasoningEffort,
		"thinking":            summarizeThinkingConfig(req.Thinking),
		"temperature":         req.Temperature,
		"top_p":               req.TopP,
		"stream":              req.Stream,
		"response_format":     summarizeResponseFormat(req.ResponseFormat),
		"tools":               summarizeToolDefinitions(req.Tools),
		"tool_choice":         req.ToolChoice,
		"parallel_tool_calls": req.ParallelToolCalls,
	}
}

func buildTraceChatResponsePayload(resp contractsopenai.ChatCompletionResponse) map[string]interface{} {
	choices := make([]map[string]interface{}, 0, len(resp.Choices))
	for _, choice := range resp.Choices {
		content := ""
		role := ""

		if choice.Message != nil {
			content = contractsopenai.ChatContentText(choice.Message.Content)
			role = choice.Message.Role
		}

		item := map[string]interface{}{
			"index":         choice.Index,
			"role":          role,
			"content":       content,
			"finish_reason": choice.FinishReason,
		}
		if choice.Message != nil && len(choice.Message.ToolCalls) > 0 {
			item["tool_calls"] = summarizeToolCalls(choice.Message.ToolCalls)
		}
		if choice.Message != nil && choice.Message.Refusal != "" {
			item["refusal"] = true
		}
		if choice.Message != nil && choice.Message.ReasoningContent != "" {
			item["reasoning_content"] = reasoningContentSummary(choice.Message.ReasoningContent)
		}
		if choice.Message != nil && len(choice.Message.ThinkingBlocks) > 0 {
			item["thinking_blocks"] = summarizeThinkingBlocks(choice.Message.ThinkingBlocks)
		}
		choices = append(choices, item)
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

func summarizeToolDefinitions(tools []contractsopenai.ToolDefinition) []map[string]interface{} {
	if len(tools) == 0 {
		return nil
	}

	out := make([]map[string]interface{}, 0, len(tools))
	for _, tool := range tools {
		item := map[string]interface{}{"type": tool.Type}
		if tool.Function != nil {
			item["function"] = map[string]interface{}{
				"name":   tool.Function.Name,
				"strict": tool.Function.Strict,
			}
		}
		if tool.Custom != nil {
			item["custom"] = map[string]interface{}{"name": tool.Custom.Name}
		}
		out = append(out, item)
	}

	return out
}

func summarizeResponseFormat(value interface{}) interface{} {
	format, ok := value.(map[string]interface{})
	if !ok {
		return value
	}

	out := map[string]interface{}{}
	if formatType, ok := format["type"].(string); ok {
		out["type"] = formatType
	}

	if rawSchema, ok := format["json_schema"].(map[string]interface{}); ok {
		schemaSummary := map[string]interface{}{}
		if name, ok := rawSchema["name"].(string); ok {
			schemaSummary["name"] = name
		}
		if strict, ok := rawSchema["strict"].(bool); ok {
			schemaSummary["strict"] = strict
		}
		if _, ok := rawSchema["schema"]; ok {
			schemaSummary["schema_present"] = true
		}
		out["json_schema"] = schemaSummary
	}

	return out
}

func summarizeThinkingConfig(value *contractsopenai.ThinkingConfig) interface{} {
	if value == nil {
		return nil
	}

	return map[string]interface{}{
		"type":          value.Type,
		"budget_tokens": value.BudgetTokens,
		"display":       value.Display,
	}
}

func reasoningContentSummary(value string) interface{} {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	return map[string]interface{}{
		"present": true,
		"length":  len(value),
	}
}

func summarizeThinkingBlocks(blocks []contractsopenai.ThinkingBlock) []map[string]interface{} {
	if len(blocks) == 0 {
		return nil
	}

	out := make([]map[string]interface{}, 0, len(blocks))
	for _, block := range blocks {
		item := map[string]interface{}{}
		if blockType, ok := block["type"].(string); ok {
			item["type"] = blockType
		}
		if thinking, ok := block["thinking"].(string); ok {
			item["thinking_length"] = len(thinking)
		}
		if data, ok := block["data"].(string); ok {
			item["data_length"] = len(data)
		}
		if signature, ok := block["signature"].(string); ok && strings.TrimSpace(signature) != "" {
			item["signature_present"] = true
		}
		out = append(out, item)
	}

	return out
}

func summarizeToolCalls(toolCalls []contractsopenai.ToolCall) []map[string]interface{} {
	if len(toolCalls) == 0 {
		return nil
	}

	out := make([]map[string]interface{}, 0, len(toolCalls))
	for _, call := range toolCalls {
		item := map[string]interface{}{
			"id":   call.ID,
			"type": call.Type,
		}
		if call.Function != nil {
			item["function"] = map[string]interface{}{
				"name":             call.Function.Name,
				"arguments_length": len(call.Function.Arguments),
			}
		}
		if call.Custom != nil {
			item["custom"] = map[string]interface{}{
				"name":         call.Custom.Name,
				"input_length": len(call.Custom.Input),
			}
		}
		out = append(out, item)
	}

	return out
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

func buildTraceResponseRequestPayload(req contractsopenai.ResponseRequest) map[string]interface{} {
	return map[string]interface{}{
		"model":                req.Model,
		"input":                redactTraceValue(req.Input),
		"instructions":         req.Instructions,
		"max_output_tokens":    req.MaxOutputTokens,
		"temperature":          req.Temperature,
		"top_p":                req.TopP,
		"text":                 summarizeResponseTextConfig(req.Text),
		"tools":                req.Tools,
		"tool_choice":          req.ToolChoice,
		"previous_response_id": req.PreviousResponseID,
		"stream":               req.Stream,
		"user":                 req.User,
	}
}

func summarizeResponseTextConfig(text *contractsopenai.ResponseTextConfig) interface{} {
	if text == nil {
		return nil
	}

	return map[string]interface{}{
		"format": summarizeResponseFormat(text.Format),
	}
}

func redactTraceValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		return contractsopenai.RedactDataURL(v)
	case []interface{}:
		out := make([]interface{}, 0, len(v))
		for _, item := range v {
			out = append(out, redactTraceValue(item))
		}

		return out
	case map[string]interface{}:
		out := make(map[string]interface{}, len(v))
		for key, item := range v {
			if strings.EqualFold(key, "image_url") {
				out[key] = redactTraceValue(item)
				continue
			}

			if strings.EqualFold(key, "url") {
				if s, ok := item.(string); ok {
					out[key] = contractsopenai.RedactDataURL(s)
					continue
				}
			}

			out[key] = redactTraceValue(item)
		}

		return out
	default:
		return value
	}
}

func buildTraceResponsePayload(resp contractsopenai.Response) map[string]interface{} {
	output := make([]map[string]interface{}, 0, len(resp.Output))
	for _, item := range resp.Output {
		content := make([]map[string]interface{}, 0, len(item.Content))
		for _, block := range item.Content {
			content = append(content, map[string]interface{}{
				"type": block.Type,
				"text": block.Text,
			})
		}

		output = append(output, map[string]interface{}{
			"id":      item.ID,
			"type":    item.Type,
			"role":    item.Role,
			"status":  item.Status,
			"content": content,
		})
	}

	usage := map[string]interface{}{}
	if resp.Usage != nil {
		usage["input_tokens"] = resp.Usage.InputTokens
		usage["output_tokens"] = resp.Usage.OutputTokens
		usage["total_tokens"] = resp.Usage.TotalTokens
	}

	return map[string]interface{}{
		"id":     resp.ID,
		"model":  resp.Model,
		"status": resp.Status,
		"output": output,
		"usage":  usage,
	}
}

func buildTraceEmbeddingRequestPayload(req contractsopenai.EmbeddingRequest) map[string]interface{} {
	inputType, inputCount, inputLengths, inputHashes := summarizeEmbeddingInput(req.Input)

	return map[string]interface{}{
		"model":           req.Model,
		"input_type":      inputType,
		"input_count":     inputCount,
		"input_lengths":   inputLengths,
		"input_hashes":    inputHashes,
		"encoding_format": req.EncodingFormat,
		"dimensions":      req.Dimensions,
		"user":            req.User,
	}
}

func buildTraceEmbeddingResponsePayload(resp contractsopenai.EmbeddingResponse) map[string]interface{} {
	vectorDims := make([]int, 0, len(resp.Data))
	for _, item := range resp.Data {
		var dims []float64
		if err := json.Unmarshal(item.Embedding, &dims); err == nil {
			vectorDims = append(vectorDims, len(dims))
			continue
		}

		vectorDims = append(vectorDims, 0)
	}

	usage := map[string]interface{}{}
	if resp.Usage != nil {
		usage["prompt_tokens"] = resp.Usage.PromptTokens
		usage["total_tokens"] = resp.Usage.TotalTokens
	}

	return map[string]interface{}{
		"model":         resp.Model,
		"object":        resp.Object,
		"embedding_cnt": len(resp.Data),
		"vector_dims":   vectorDims,
		"usage":         usage,
	}
}

func buildTraceModerationRequestPayload(req contractsopenai.ModerationRequest) map[string]interface{} {
	inputType, inputCount, inputLengths, inputHashes := summarizeModerationInput(req.Input)

	return map[string]interface{}{
		"model":         req.Model,
		"input_type":    inputType,
		"input_count":   inputCount,
		"input_lengths": inputLengths,
		"input_hashes":  inputHashes,
		"user":          req.User,
	}
}

func buildTraceModerationResponsePayload(resp contractsopenai.ModerationResponse) map[string]interface{} {
	results := make([]map[string]interface{}, 0, len(resp.Results))
	flaggedCount := 0

	for _, result := range resp.Results {
		if result.Flagged {
			flaggedCount++
		}

		results = append(results, map[string]interface{}{
			"flagged":    result.Flagged,
			"categories": result.Categories,
		})
	}

	return map[string]interface{}{
		"id":            resp.ID,
		"model":         resp.Model,
		"result_count":  len(resp.Results),
		"flagged_count": flaggedCount,
		"results":       results,
	}
}

func buildTraceImageRequestPayload(operation string, body []byte, contentType string, metadata map[string]interface{}) map[string]interface{} {
	out := map[string]interface{}{
		"operation":    operation,
		"content_type": safeContentTypeFamily(contentType),
	}
	for k, v := range metadata {
		out[k] = v
	}

	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(contentType)), "application/json") {
		return out
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return out
	}

	if prompt, ok := raw["prompt"].(string); ok {
		out["prompt_length"] = len(prompt)
		out["prompt_hash"] = hashTracePreview(prompt)
	}

	if image, ok := raw["image"]; ok {
		out["image_source_type"] = traceImageSourceType(image)
	}

	return out
}

func buildTraceImageResponsePayload(body []byte) map[string]interface{} {
	out := map[string]interface{}{}

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		out["decode"] = "unavailable"
		return out
	}

	if created, ok := raw["created"]; ok {
		out["created"] = created
	}

	if usage, ok := raw["usage"].(map[string]interface{}); ok {
		safeUsage := map[string]interface{}{}

		for _, key := range []string{"cost_in_usd_ticks", "total_tokens", "input_tokens", "output_tokens"} {
			if v, exists := usage[key]; exists {
				safeUsage[key] = v
			}
		}

		out["usage"] = safeUsage
	}

	if data, ok := raw["data"].([]interface{}); ok {
		out["data_count"] = len(data)

		formats := make([]string, 0, len(data))
		for _, item := range data {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			switch {
			case m["b64_json"] != nil:
				formats = append(formats, "b64_json")
			case m["url"] != nil:
				formats = append(formats, "url")
			}

			if mt, ok := m["mime_type"].(string); ok && mt != "" {
				out["mime_type"] = mt
			}
		}

		out["data_formats"] = formats
	}

	return out
}

func safeContentTypeFamily(contentType string) string {
	ct := strings.ToLower(strings.TrimSpace(contentType))
	switch {
	case strings.HasPrefix(ct, "multipart/form-data"):
		return "multipart/form-data"
	case strings.HasPrefix(ct, "application/json"):
		return "application/json"
	default:
		if i := strings.Index(ct, ";"); i >= 0 {
			return strings.TrimSpace(ct[:i])
		}

		return ct
	}
}

func traceImageSourceType(value interface{}) string {
	switch v := value.(type) {
	case string:
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(v)), "http") {
			return "url"
		}

		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(v)), "data:") {
			return "data_url"
		}

		return "string"
	case map[string]interface{}:
		if t, ok := v["type"].(string); ok && strings.TrimSpace(t) != "" {
			return strings.TrimSpace(t)
		}

		return "object"
	case []interface{}:
		return "array"
	default:
		if value == nil {
			return ""
		}

		return "unknown"
	}
}

func summarizeModerationInput(input interface{}) (string, int, []int, []string) {
	return summarizeTextLikeInput(input)
}

func summarizeEmbeddingInput(input interface{}) (string, int, []int, []string) {
	return summarizeTextLikeInput(input)
}

func summarizeTextLikeInput(input interface{}) (string, int, []int, []string) {
	switch v := input.(type) {
	case string:
		return "single", 1, []int{len(v)}, []string{hashTracePreview(v)}
	case []string:
		lengths := make([]int, 0, len(v))

		hashes := make([]string, 0, len(v))
		for _, item := range v {
			lengths = append(lengths, len(item))
			hashes = append(hashes, hashTracePreview(item))
		}

		return "batch", len(v), lengths, hashes
	case []interface{}:
		lengths := make([]int, 0, len(v))

		hashes := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				lengths = append(lengths, len(s))
				hashes = append(hashes, hashTracePreview(s))
			}
		}

		return "batch", len(lengths), lengths, hashes
	default:
		return "unknown", 0, nil, nil
	}
}

func hashTracePreview(input string) string {
	sum := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", sum[:6])
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
