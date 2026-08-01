package proxy

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"strings"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/gofiber/fiber/v2"
)

func validateMistralParameters(binding *platform.ModelRuntimeProfileBinding, endpoint string, parameters map[string]json.RawMessage) error {
	if err := validateMistralRuntimeEndpoint(binding, endpoint); err != nil {
		return err
	}

	for parameter := range parameters {
		canonical := strings.ToLower(strings.TrimSpace(parameter))
		switch canonical {
		case "seed":
			canonical = "random_seed"
		case "dimensions":
			canonical = "output_dimension"
		}

		if !binding.Profile.AllowsParameter(endpoint, canonical) {
			return fmt.Errorf("model does not support parameter %q on %s", parameter, endpoint)
		}
	}

	return nil
}

func validateMistralMultipartParameters(binding *platform.ModelRuntimeProfileBinding, endpoint string, c *fiber.Ctx) error {
	if err := validateMistralRuntimeEndpoint(binding, endpoint); err != nil {
		return err
	}

	form, err := c.MultipartForm()
	if err != nil || form == nil {
		return fmt.Errorf("invalid Mistral AI multipart request")
	}

	return validateMistralMultipartForm(binding, endpoint, form)
}

func validateMistralMultipartForm(binding *platform.ModelRuntimeProfileBinding, endpoint string, form *multipart.Form) error {
	for parameter := range form.Value {
		if !binding.Profile.AllowsParameter(endpoint, parameter) {
			return fmt.Errorf("model does not support parameter %q on %s", parameter, endpoint)
		}
	}

	for parameter := range form.File {
		if !binding.Profile.AllowsParameter(endpoint, parameter) {
			return fmt.Errorf("model does not support parameter %q on %s", parameter, endpoint)
		}
	}

	return nil
}

func validateMistralJSONParameters(binding *platform.ModelRuntimeProfileBinding, endpoint string, body []byte) error {
	var parameters map[string]json.RawMessage
	if err := json.Unmarshal(body, &parameters); err != nil {
		return fmt.Errorf("invalid Mistral AI JSON request")
	}

	return validateMistralParameters(binding, endpoint, parameters)
}

func validateMistralEmbeddingRequest(body []byte) error {
	var request struct {
		OutputDimension *int   `json:"output_dimension"`
		OutputDType     string `json:"output_dtype"`
	}
	if err := json.Unmarshal(body, &request); err != nil {
		return fmt.Errorf("invalid Mistral AI JSON request")
	}

	if request.OutputDimension != nil && *request.OutputDimension <= 0 {
		return fmt.Errorf("output_dimension must be a positive integer")
	}

	if request.OutputDType != "" {
		switch strings.ToLower(strings.TrimSpace(request.OutputDType)) {
		case "float", "int8", "uint8", "binary", "ubinary":
		default:
			return fmt.Errorf("output_dtype must be one of float, int8, uint8, binary, or ubinary")
		}
	}

	return nil
}

func validateMistralRuntimeEndpoint(binding *platform.ModelRuntimeProfileBinding, endpoint string) error {
	if binding == nil {
		return fmt.Errorf("Mistral AI model runtime profile is unavailable") //nolint:staticcheck // Provider name is intentionally capitalized in user-facing validation errors.
	}

	if err := binding.Validate(); err != nil || !isMistralProvider(binding.Provider) {
		return fmt.Errorf("Mistral AI model runtime profile is invalid") //nolint:staticcheck // Provider name is intentionally capitalized in user-facing validation errors.
	}

	if !binding.Profile.AllowsEndpoint(endpoint) {
		return fmt.Errorf("model does not support the %s endpoint", endpoint)
	}

	return nil
}

func validateMistralChatRuntimeProfile(req contractsopenai.ChatCompletionRequest, binding *platform.ModelRuntimeProfileBinding) error {
	if err := validateMistralRuntimeEndpoint(binding, "chat_completions"); err != nil {
		return err
	}

	profile := binding.Profile
	if req.MaxTokens != nil && profile.Limits.MaxOutputTokens > 0 && *req.MaxTokens > profile.Limits.MaxOutputTokens {
		return fmt.Errorf("max_tokens exceeds model limit %d", profile.Limits.MaxOutputTokens)
	}

	if req.Temperature != nil && !profile.AllowsSamplingParameter("temperature") {
		return fmt.Errorf("model does not support temperature")
	}

	if req.TopP != nil && !profile.AllowsSamplingParameter("top_p") {
		return fmt.Errorf("model does not support top_p")
	}

	if req.TopK != nil && !profile.AllowsSamplingParameter("top_k") {
		return fmt.Errorf("model does not support top_k")
	}

	if chatRequestHasImage(req) && !profile.AllowsInputModality("image") {
		return fmt.Errorf("model does not support image input")
	}

	if chatRequestHasFile(req) && !profile.AllowsInputModality("document") && !profile.AllowsInputModality("pdf") {
		return fmt.Errorf("model does not support document input")
	}

	if chatRequestHasTools(req) && !profile.Tools.Custom {
		return fmt.Errorf("model does not support custom tools")
	}

	for _, tool := range req.Tools {
		toolType := strings.ToLower(strings.TrimSpace(tool.Type))

		capability := toolType
		if toolType == "function" {
			capability = "function_calling"
		}

		if !profile.AllowsTool(capability) {
			return fmt.Errorf("model does not support Mistral tool type %q", tool.Type)
		}

		if toolType != "function" {
			if toolType == "image_generation" {
				return fmt.Errorf("mistral tool type %q requires image-output billing and Files API support that are not enabled", tool.Type)
			}

			return fmt.Errorf("mistral tool type %q requires the Conversations or Agents API", tool.Type)
		}
	}

	if req.ParallelToolCalls != nil && *req.ParallelToolCalls && !profile.Tools.Parallel {
		return fmt.Errorf("model does not support parallel tool calls")
	}

	if chatRequestHasResponseFormat(req) && !profile.Tools.StructuredOutput {
		return fmt.Errorf("model does not support structured output")
	}

	return validateMistralParameters(binding, "chat_completions", req.AdditionalFields)
}

func validateMistralFIMRuntimeProfile(req contractsopenai.CompletionRequest, binding *platform.ModelRuntimeProfileBinding) error {
	if err := validateMistralRuntimeEndpoint(binding, "fim_completions"); err != nil {
		return err
	}

	profile := binding.Profile
	if req.MaxTokens != nil && profile.Limits.MaxOutputTokens > 0 && *req.MaxTokens > profile.Limits.MaxOutputTokens {
		return fmt.Errorf("max_tokens exceeds model limit %d", profile.Limits.MaxOutputTokens)
	}

	if req.MinTokens != nil && *req.MinTokens < 0 {
		return fmt.Errorf("min_tokens must be non-negative")
	}

	if req.MaxTokens != nil && req.MinTokens != nil && *req.MinTokens > *req.MaxTokens {
		return fmt.Errorf("min_tokens cannot exceed max_tokens")
	}

	if req.Temperature != nil && !profile.AllowsSamplingParameter("temperature") {
		return fmt.Errorf("model does not support temperature")
	}

	if req.TopP != nil && !profile.AllowsSamplingParameter("top_p") {
		return fmt.Errorf("model does not support top_p")
	}

	return validateMistralParameters(binding, "fim_completions", req.AdditionalFields)
}
