package proxy

import (
	"encoding/json"
	"strings"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/services/control-plane/internal/billing"
	"go.uber.org/zap"
)

const googleGeminiProvider = "google-gemini"

type googleGeminiUsageRepair struct {
	applied                  bool
	promptTokens             int
	originalCompletionTokens int
	normalizedCompletion     int
	totalTokens              int
	inferredReasoningTokens  int
}

func normalizeGoogleGeminiUsage(provider string, usage *contractsopenai.Usage) googleGeminiUsageRepair {
	if !strings.EqualFold(strings.TrimSpace(provider), googleGeminiProvider) || usage == nil {
		return googleGeminiUsageRepair{}
	}

	repair := googleGeminiUsageRepair{
		promptTokens:             usage.PromptTokens,
		originalCompletionTokens: usage.CompletionTokens,
		totalTokens:              usage.TotalTokens,
	}

	inferredReasoningTokens, changed := billing.NormalizeGoogleGeminiOpenAIUsage(usage)
	if !changed {
		return googleGeminiUsageRepair{}
	}

	repair.applied = true
	repair.normalizedCompletion = usage.CompletionTokens
	repair.inferredReasoningTokens = inferredReasoningTokens

	return repair
}

func logGoogleGeminiUsageRepair(logger *zap.Logger, requestID string, repair googleGeminiUsageRepair) {
	if logger == nil || !repair.applied {
		return
	}

	logger.Warn(
		"normalized inconsistent Google Gemini OpenAI-compatible usage",
		zap.String("provider", googleGeminiProvider),
		zap.String("request_id", requestID),
		zap.Int("prompt_tokens", repair.promptTokens),
		zap.Int("original_completion_tokens", repair.originalCompletionTokens),
		zap.Int("normalized_completion_tokens", repair.normalizedCompletion),
		zap.Int("total_tokens", repair.totalTokens),
		zap.Int("inferred_reasoning_tokens", repair.inferredReasoningTokens),
	)
}

type googleGeminiStreamUsageNormalizer struct {
	buffer string
}

func newGoogleGeminiStreamUsageNormalizer(provider string) *googleGeminiStreamUsageNormalizer {
	if !strings.EqualFold(strings.TrimSpace(provider), googleGeminiProvider) {
		return nil
	}

	return &googleGeminiStreamUsageNormalizer{}
}

func (normalizer *googleGeminiStreamUsageNormalizer) Add(data []byte) ([]byte, []googleGeminiUsageRepair) {
	if normalizer == nil {
		return data, nil
	}

	normalizer.buffer += string(data)

	var output strings.Builder

	repairs := make([]googleGeminiUsageRepair, 0, 1)

	for {
		frame, rest, ok := takeSSEFrame(normalizer.buffer)
		if !ok {
			break
		}

		normalizer.buffer = rest
		normalizedFrame, repair := normalizeGoogleGeminiSSEFrame(frame)
		output.WriteString(normalizedFrame)
		output.WriteString("\n\n")

		if repair.applied {
			repairs = append(repairs, repair)
		}
	}

	return []byte(output.String()), repairs
}

func (normalizer *googleGeminiStreamUsageNormalizer) Flush() []byte {
	if normalizer == nil || normalizer.buffer == "" {
		return nil
	}

	pending := []byte(normalizer.buffer)
	normalizer.buffer = ""

	return pending
}

func normalizeGoogleGeminiSSEFrame(frame string) (string, googleGeminiUsageRepair) {
	payload := streamFrameData(frame)
	if payload == "" || payload == "[DONE]" {
		return frame, googleGeminiUsageRepair{}
	}

	var envelope map[string]json.RawMessage
	if err := json.Unmarshal([]byte(payload), &envelope); err != nil {
		return frame, googleGeminiUsageRepair{}
	}

	rawUsage, ok := envelope["usage"]
	if !ok {
		return frame, googleGeminiUsageRepair{}
	}

	var usage contractsopenai.Usage
	if err := json.Unmarshal(rawUsage, &usage); err != nil {
		return frame, googleGeminiUsageRepair{}
	}

	repair := normalizeGoogleGeminiUsage(googleGeminiProvider, &usage)
	if !repair.applied {
		return frame, googleGeminiUsageRepair{}
	}

	normalizedUsage, err := json.Marshal(usage)
	if err != nil {
		return frame, googleGeminiUsageRepair{}
	}

	envelope["usage"] = normalizedUsage

	normalizedPayload, err := json.Marshal(envelope)
	if err != nil {
		return frame, googleGeminiUsageRepair{}
	}

	return replaceSSEFrameData(frame, string(normalizedPayload)), repair
}

func replaceSSEFrameData(frame string, payload string) string {
	lines := strings.Split(strings.ReplaceAll(frame, "\r\n", "\n"), "\n")
	output := make([]string, 0, len(lines))
	replaced := false

	for _, line := range lines {
		if strings.HasPrefix(line, "data:") {
			if !replaced {
				output = append(output, "data: "+payload)
				replaced = true
			}

			continue
		}

		output = append(output, line)
	}

	if replaced {
		return strings.Join(output, "\n")
	}

	if strings.HasPrefix(strings.TrimSpace(frame), "{") {
		return payload
	}

	return frame
}
