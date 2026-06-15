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
			"role":    msg.Role,
			"content": contractsopenai.RedactChatContent(msg.Content),
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
			content = contractsopenai.ChatContentText(choice.Message.Content)
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

func buildTraceResponseRequestPayload(req contractsopenai.ResponseRequest) map[string]interface{} {
	return map[string]interface{}{
		"model":                req.Model,
		"input":                redactTraceValue(req.Input),
		"instructions":         req.Instructions,
		"max_output_tokens":    req.MaxOutputTokens,
		"temperature":          req.Temperature,
		"top_p":                req.TopP,
		"tools":                req.Tools,
		"tool_choice":          req.ToolChoice,
		"previous_response_id": req.PreviousResponseID,
		"stream":               req.Stream,
		"user":                 req.User,
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

func summarizeModerationInput(input interface{}) (string, int, []int, []string) {
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

func summarizeEmbeddingInput(input interface{}) (string, int, []int, []string) {
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
