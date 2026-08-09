package proxy

import (
	"fmt"
	"strings"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
)

// validateGoogleGeminiChatRuntimeProfile prevents a model imported for a
// different Gemini API surface (Interactions, Live, embeddings, media, or a
// native-only method) from being dispatched through Chat Completions.
func validateGoogleGeminiChatRuntimeProfile(
	req contractsopenai.ChatCompletionRequest,
	binding *platform.ModelRuntimeProfileBinding,
) error {
	if err := validateGoogleGeminiRuntimeEndpoint(
		binding,
		"chat_completions",
	); err != nil {
		return err
	}

	profile := binding.Profile
	if req.MaxTokens != nil && profile.Limits.MaxOutputTokens > 0 &&
		*req.MaxTokens > profile.Limits.MaxOutputTokens {
		return fmt.Errorf(
			"max_tokens exceeds model limit %d",
			profile.Limits.MaxOutputTokens,
		)
	}

	if chatRequestHasImage(req) && !profile.AllowsInputModality("image") {
		return fmt.Errorf("model does not support image input")
	}

	if chatRequestHasInputAudio(req) &&
		!profile.AllowsInputModality("audio") {
		return fmt.Errorf("model does not support audio input")
	}

	for _, tool := range req.Tools {
		if !strings.EqualFold(strings.TrimSpace(tool.Type), "function") {
			return fmt.Errorf(
				"unsupported Google Gemini tool type %q",
				tool.Type,
			)
		}

		if !profile.Tools.Custom {
			return fmt.Errorf("model does not support custom tools")
		}
	}

	if req.ParallelToolCalls != nil && *req.ParallelToolCalls &&
		!profile.Tools.Parallel {
		return fmt.Errorf("model does not support parallel tool calls")
	}

	if chatRequestHasResponseFormat(req) &&
		!profile.Tools.StructuredOutput {
		return fmt.Errorf("model does not support structured output")
	}

	tier := strings.ToLower(strings.TrimSpace(req.ServiceTier))
	if tier == "" || tier == "default" {
		tier = "standard"
	}

	if !profile.AllowsServiceTier(tier) {
		return fmt.Errorf("unsupported service_tier %q", tier)
	}

	return nil
}

func validateGoogleGeminiRuntimeEndpoint(
	binding *platform.ModelRuntimeProfileBinding,
	endpoint string,
) error {
	if binding == nil {
		return fmt.Errorf("Google Gemini model runtime profile is unavailable") //nolint:staticcheck // Provider name is intentionally capitalized in user-facing validation errors.
	}

	if err := binding.Validate(); err != nil ||
		!isGoogleGeminiProvider(binding.Provider) {
		return fmt.Errorf("Google Gemini model runtime profile is invalid") //nolint:staticcheck // Provider name is intentionally capitalized in user-facing validation errors.
	}

	if !binding.Profile.AllowsEndpoint(endpoint) {
		return fmt.Errorf("model does not support the %s endpoint", endpoint)
	}

	return nil
}
