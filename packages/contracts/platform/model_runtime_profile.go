package platform

import (
	"fmt"
	"strings"
	"time"
)

// ModelRuntimeProfile is the versioned provider behavior contract imported
// from Model Discovery. Field names intentionally match the catalog JSON.
type ModelRuntimeProfile struct {
	Version          int                          `json:"version"`
	Endpoints        []string                     `json:"endpoints"`
	Transports       []string                     `json:"transports,omitempty"`
	Lifecycle        string                       `json:"lifecycle,omitempty"`
	Thinking         ModelRuntimeThinkingProfile  `json:"thinking"`
	AssistantPrefill bool                         `json:"assistantPrefill"`
	Sampling         ModelRuntimeSamplingProfile  `json:"sampling"`
	Tools            ModelRuntimeToolsProfile     `json:"tools"`
	Modalities       ModelRuntimeModalities       `json:"modalities"`
	RequestControls  ModelRuntimeRequestControls  `json:"requestControls"`
	BetaHeaders      []string                     `json:"betaHeaders"`
	BetaCapabilities []ModelRuntimeBetaCapability `json:"betaCapabilities,omitempty"`
	Limits           ModelRuntimeLimits           `json:"limits"`
}

type ModelRuntimeThinkingProfile struct {
	DefaultMode      string   `json:"defaultMode"`
	AllowedModes     []string `json:"allowedModes"`
	ManualDeprecated bool     `json:"manualDeprecated"`
	EffortValues     []string `json:"effortValues"`
	ReasoningModes   []string `json:"reasoningModes,omitempty"`
	ContextValues    []string `json:"contextValues,omitempty"`
}

type ModelRuntimeSamplingProfile struct {
	Parameters             []string `json:"parameters"`
	DisallowedWithThinking []string `json:"disallowedWithThinking"`
}

type ModelRuntimeToolsProfile struct {
	Custom           bool     `json:"custom"`
	Server           bool     `json:"server"`
	Parallel         bool     `json:"parallel"`
	StructuredOutput bool     `json:"structuredOutput"`
	Supported        []string `json:"supported,omitempty"`
}

type ModelRuntimeModalities struct {
	Input  []string `json:"input"`
	Output []string `json:"output"`
}

type ModelRuntimeRequestControls struct {
	ServiceTierValues  []string          `json:"serviceTierValues"`
	InferenceGeoValues []string          `json:"inferenceGeoValues"`
	SpeedValues        []string          `json:"speedValues"`
	EffectiveSpeed     map[string]string `json:"effectiveSpeed,omitempty"`
}

type ModelRuntimeBetaCapability struct {
	Header          string   `json:"header"`
	Endpoints       []string `json:"endpoints"`
	MaxOutputTokens int      `json:"maxOutputTokens,omitempty"`
}

type ModelRuntimeLimits struct {
	ContextWindow   int `json:"contextWindow"`
	MaxOutputTokens int `json:"maxOutputTokens"`
	MessagesBytes   int `json:"messagesBytes"`
	BatchBytes      int `json:"batchBytes"`
	FileBytes       int `json:"fileBytes"`
}

// ModelRuntimeProfileBinding pins one model profile to the exact atomic
// catalog revision used for request pricing and dispatch.
type ModelRuntimeProfileBinding struct {
	Provider        string              `json:"provider"`
	CatalogRevision string              `json:"catalogRevision"`
	CanonicalID     string              `json:"canonicalId"`
	Slug            string              `json:"slug"`
	ProfileVersion  int                 `json:"profileVersion"`
	Profile         ModelRuntimeProfile `json:"profile"`
	SourceURL       string              `json:"sourceUrl,omitempty"`
	EffectiveFrom   time.Time           `json:"effectiveFrom"`
}

func (binding ModelRuntimeProfileBinding) Validate() error {
	if strings.TrimSpace(binding.Provider) == "" {
		return fmt.Errorf("model runtime profile provider is required")
	}
	if strings.TrimSpace(binding.CatalogRevision) == "" {
		return fmt.Errorf("model runtime profile catalog revision is required")
	}
	if strings.TrimSpace(binding.CanonicalID) == "" ||
		strings.TrimSpace(binding.Slug) == "" {
		return fmt.Errorf("model runtime profile identity is required")
	}
	if binding.ProfileVersion <= 0 ||
		binding.Profile.Version != binding.ProfileVersion {
		return fmt.Errorf("model runtime profile version mismatch")
	}
	if binding.EffectiveFrom.IsZero() {
		return fmt.Errorf("model runtime profile effective time is required")
	}
	return binding.Profile.Validate()
}

func (profile ModelRuntimeProfile) Validate() error {
	if profile.Version <= 0 {
		return fmt.Errorf("model runtime profile version must be positive")
	}
	if len(profile.Endpoints) == 0 {
		return fmt.Errorf("model runtime profile endpoints are required")
	}
	for _, endpoint := range profile.Endpoints {
		if !containsRuntimeValue(
			[]string{
				"messages",
				"count_tokens",
				"message_batches",
				"files",
				"models",
				"chat_completions",
				"completions",
				"responses",
				"responses_input_tokens",
				"responses_compact",
				"conversations",
				"embeddings",
				"moderations",
				"image_generations",
				"image_edits",
				"image_variations",
				"audio_speech",
				"audio_transcriptions",
				"audio_translations",
				"realtime",
				"realtime_transcription",
				"realtime_translation",
				"videos",
				"uploads",
				"vector_stores",
				"containers",
				"interactions",
				"live",
				"batches",
				"generate_content",
				"generate_answer",
				"embed_content",
				"imagen_generate",
			},
			endpoint,
		) {
			return fmt.Errorf("unsupported model endpoint %q", endpoint)
		}
	}
	for _, transport := range profile.Transports {
		if !containsRuntimeValue([]string{"http", "sse", "websocket"}, transport) {
			return fmt.Errorf("unsupported model transport %q", transport)
		}
	}
	if profile.Lifecycle != "" && !containsRuntimeValue(
		[]string{"active", "preview", "deprecated", "retired", "unavailable"},
		profile.Lifecycle,
	) {
		return fmt.Errorf("unsupported model lifecycle %q", profile.Lifecycle)
	}
	if !containsRuntimeValue(
		profile.Thinking.AllowedModes,
		profile.Thinking.DefaultMode,
	) {
		return fmt.Errorf("default thinking mode must be allowed")
	}
	for _, mode := range profile.Thinking.AllowedModes {
		if !containsRuntimeValue(
			[]string{"adaptive", "manual", "disabled"},
			mode,
		) {
			return fmt.Errorf("unsupported thinking mode %q", mode)
		}
	}
	for _, effort := range profile.Thinking.EffortValues {
		if !containsRuntimeValue(
			[]string{"none", "minimal", "low", "medium", "high", "xhigh", "max"},
			effort,
		) {
			return fmt.Errorf("unsupported effort value %q", effort)
		}
	}
	for _, mode := range profile.Thinking.ReasoningModes {
		if !containsRuntimeValue([]string{"standard", "pro"}, mode) {
			return fmt.Errorf("unsupported reasoning mode %q", mode)
		}
	}
	for _, value := range profile.Thinking.ContextValues {
		if !containsRuntimeValue([]string{"auto", "current_turn", "all_turns"}, value) {
			return fmt.Errorf("unsupported reasoning context %q", value)
		}
	}
	for _, parameter := range append(
		append(
			[]string(nil),
			profile.Sampling.Parameters...,
		),
		profile.Sampling.DisallowedWithThinking...,
	) {
		if !containsRuntimeValue(
			[]string{"temperature", "top_p", "top_k"},
			parameter,
		) {
			return fmt.Errorf("unsupported sampling parameter %q", parameter)
		}
	}
	for _, tool := range profile.Tools.Supported {
		if !containsRuntimeValue(
			[]string{
				"function_calling",
				"web_search",
				"file_search",
				"tool_search",
				"image_generation",
				"code_interpreter",
				"hosted_shell",
				"apply_patch",
				"skills",
				"computer_use",
				"mcp",
			},
			tool,
		) {
			return fmt.Errorf("unsupported model tool %q", tool)
		}
	}
	if len(profile.Modalities.Input) == 0 ||
		len(profile.Modalities.Output) == 0 {
		return fmt.Errorf("model runtime profile modalities are required")
	}
	for _, modality := range profile.Modalities.Input {
		if !containsRuntimeValue(
			[]string{"unknown", "text", "image", "pdf", "audio", "video", "code"},
			modality,
		) {
			return fmt.Errorf("unsupported input modality %q", modality)
		}
	}
	for _, modality := range profile.Modalities.Output {
		if !containsRuntimeValue(
			[]string{"unknown", "text", "image", "audio", "video", "embedding", "code"},
			modality,
		) {
			return fmt.Errorf("unsupported output modality %q", modality)
		}
	}
	for _, value := range profile.RequestControls.ServiceTierValues {
		if !containsRuntimeValue(
			[]string{"auto", "standard_only", "standard", "batch", "flex", "priority", "fast"},
			value,
		) {
			return fmt.Errorf("unsupported service tier %q", value)
		}
	}
	for _, value := range profile.RequestControls.InferenceGeoValues {
		if !containsRuntimeValue([]string{"global", "us"}, value) {
			return fmt.Errorf("unsupported inference geo %q", value)
		}
	}
	for _, value := range profile.RequestControls.SpeedValues {
		if !containsRuntimeValue([]string{"standard", "fast"}, value) {
			return fmt.Errorf("unsupported speed %q", value)
		}
	}
	for requested, effective := range profile.RequestControls.EffectiveSpeed {
		if !profile.AllowsSpeed(requested) {
			return fmt.Errorf(
				"effective speed mapping uses unsupported request %q",
				requested,
			)
		}
		if !containsRuntimeValue([]string{"standard", "fast"}, effective) {
			return fmt.Errorf("unsupported effective speed %q", effective)
		}
	}
	for _, header := range profile.BetaHeaders {
		if strings.TrimSpace(header) == "" ||
			strings.ContainsAny(header, ",\r\n") {
			return fmt.Errorf("invalid beta header %q", header)
		}
	}
	for _, capability := range profile.BetaCapabilities {
		if !profile.AllowsBetaHeader(capability.Header) {
			return fmt.Errorf(
				"beta capability header %q is not allowlisted",
				capability.Header,
			)
		}
		if len(capability.Endpoints) == 0 {
			return fmt.Errorf(
				"beta capability %q requires endpoints",
				capability.Header,
			)
		}
		for _, endpoint := range capability.Endpoints {
			if !profile.AllowsEndpoint(endpoint) {
				return fmt.Errorf(
					"beta capability %q uses unsupported endpoint %q",
					capability.Header,
					endpoint,
				)
			}
		}
		if capability.MaxOutputTokens < 0 {
			return fmt.Errorf(
				"beta capability %q max output tokens cannot be negative",
				capability.Header,
			)
		}
	}
	if profile.Limits.ContextWindow < 0 ||
		profile.Limits.MaxOutputTokens < 0 ||
		profile.Limits.MessagesBytes <= 0 ||
		profile.Limits.BatchBytes <= 0 ||
		profile.Limits.FileBytes <= 0 {
		return fmt.Errorf("model runtime profile limits must be positive")
	}
	return nil
}

func (binding ModelRuntimeProfileBinding) MatchesModel(model string) bool {
	normalized := normalizeRuntimeModelID(model, binding.Provider)
	return normalized != "" &&
		(normalized == normalizeRuntimeModelID(binding.CanonicalID, binding.Provider) ||
			normalized == normalizeRuntimeModelID(binding.Slug, binding.Provider))
}

func (profile ModelRuntimeProfile) AllowsEndpoint(endpoint string) bool {
	return containsRuntimeValue(profile.Endpoints, endpoint)
}

func (profile ModelRuntimeProfile) AllowsThinkingMode(mode string) bool {
	return containsRuntimeValue(profile.Thinking.AllowedModes, mode)
}

func (profile ModelRuntimeProfile) AllowsEffort(effort string) bool {
	return containsRuntimeValue(profile.Thinking.EffortValues, effort)
}

func (profile ModelRuntimeProfile) AllowsSamplingParameter(
	parameter string,
) bool {
	return containsRuntimeValue(profile.Sampling.Parameters, parameter)
}

func (profile ModelRuntimeProfile) AllowsInputModality(modality string) bool {
	return containsRuntimeValue(profile.Modalities.Input, modality)
}

func (profile ModelRuntimeProfile) AllowsBetaHeader(header string) bool {
	return containsRuntimeValue(profile.BetaHeaders, header)
}

func (profile ModelRuntimeProfile) AllowsServiceTier(value string) bool {
	return containsRuntimeValue(
		profile.RequestControls.ServiceTierValues,
		value,
	)
}

func (profile ModelRuntimeProfile) AllowsInferenceGeo(value string) bool {
	return containsRuntimeValue(
		profile.RequestControls.InferenceGeoValues,
		value,
	)
}

func (profile ModelRuntimeProfile) AllowsSpeed(value string) bool {
	return containsRuntimeValue(profile.RequestControls.SpeedValues, value)
}

func (profile ModelRuntimeProfile) EffectiveSpeed(value string) string {
	requested := strings.ToLower(strings.TrimSpace(value))
	if requested == "" {
		return ""
	}
	for key, effective := range profile.RequestControls.EffectiveSpeed {
		if strings.EqualFold(strings.TrimSpace(key), requested) {
			return strings.ToLower(strings.TrimSpace(effective))
		}
	}
	if profile.AllowsSpeed(requested) {
		return requested
	}
	return ""
}

func (profile ModelRuntimeProfile) MaxOutputTokensFor(
	endpoint string,
	betaHeaders []string,
) int {
	limit := profile.Limits.MaxOutputTokens
	for _, capability := range profile.BetaCapabilities {
		if capability.MaxOutputTokens <= limit ||
			!containsRuntimeValue(capability.Endpoints, endpoint) ||
			!containsRuntimeValue(betaHeaders, capability.Header) {
			continue
		}
		limit = capability.MaxOutputTokens
	}
	return limit
}

func (profile ModelRuntimeProfile) RequiredBetaHeaderForOutput(
	endpoint string,
	maxOutputTokens int,
) string {
	if maxOutputTokens <= profile.Limits.MaxOutputTokens {
		return ""
	}
	for _, capability := range profile.BetaCapabilities {
		if capability.MaxOutputTokens >= maxOutputTokens &&
			containsRuntimeValue(capability.Endpoints, endpoint) {
			return capability.Header
		}
	}
	return ""
}

func FindModelRuntimeProfile(
	bindings []ModelRuntimeProfileBinding,
	model string,
) (*ModelRuntimeProfileBinding, bool) {
	for index := range bindings {
		if bindings[index].MatchesModel(model) {
			binding := bindings[index]
			return &binding, true
		}
	}
	return nil, false
}

func containsRuntimeValue(values []string, expected string) bool {
	expected = strings.ToLower(strings.TrimSpace(expected))
	if expected == "" {
		return false
	}
	for _, value := range values {
		if strings.ToLower(strings.TrimSpace(value)) == expected {
			return true
		}
	}
	return false
}

func normalizeRuntimeModelID(model string, provider string) string {
	value := strings.ToLower(strings.TrimSpace(model))
	if strings.Count(value, ":") == 2 {
		parts := strings.SplitN(value, ":", 3)
		value = strings.TrimSpace(parts[2])
	}
	prefix := strings.ToLower(strings.TrimSpace(provider)) + "/"
	return strings.TrimPrefix(value, prefix)
}
