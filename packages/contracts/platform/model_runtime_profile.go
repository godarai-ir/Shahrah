package platform

import (
	"fmt"
	"strings"
	"time"
)

// ModelRuntimeProfile is the versioned provider behavior contract imported
// from Model Discovery. Field names intentionally match the catalog JSON.
type ModelRuntimeProfile struct {
	Version               int                           `json:"version"`
	Aliases               []string                      `json:"aliases,omitempty"`
	Endpoints             []string                      `json:"endpoints"`
	Transports            []string                      `json:"transports,omitempty"`
	Lifecycle             string                        `json:"lifecycle,omitempty"`
	Thinking              ModelRuntimeThinkingProfile   `json:"thinking"`
	AssistantPrefill      bool                          `json:"assistantPrefill"`
	FineTuning            bool                          `json:"fineTuning,omitempty"`
	PredictedOutputs      bool                          `json:"predictedOutputs,omitempty"`
	Sampling              ModelRuntimeSamplingProfile   `json:"sampling"`
	Tools                 ModelRuntimeToolsProfile      `json:"tools"`
	Modalities            ModelRuntimeModalities        `json:"modalities"`
	Parameters            map[string][]string           `json:"parameters,omitempty"`
	RequestControls       ModelRuntimeRequestControls   `json:"requestControls"`
	RequestHeaders        ModelRuntimeRequestHeaders    `json:"requestHeaders,omitempty"`
	ContextPricing        ModelRuntimeContextPricing    `json:"contextPricing,omitempty"`
	Batch                 ModelRuntimeBatchProfile      `json:"batch,omitempty"`
	DeferredCompletions   ModelRuntimeDeferredProfile   `json:"deferredCompletions,omitempty"`
	ResponseResources     ModelRuntimeResponseResources `json:"responseResources,omitempty"`
	ImageControls         ModelRuntimeImageControls     `json:"imageControls,omitempty"`
	UsageAccounting       ModelRuntimeUsageAccounting   `json:"usageAccounting,omitempty"`
	UnsupportedParameters []string                      `json:"unsupportedParameters,omitempty"`
	BetaHeaders           []string                      `json:"betaHeaders"`
	BetaCapabilities      []ModelRuntimeBetaCapability  `json:"betaCapabilities,omitempty"`
	ReferenceRateLimits   []ModelRuntimeRateLimitTier   `json:"referenceRateLimits,omitempty"`
	MinimumRetentionDays  int                           `json:"minimumRetentionDays,omitempty"`
	ZDREligible           bool                          `json:"zdrEligible,omitempty"`
	Limits                ModelRuntimeLimits            `json:"limits"`
}

type ModelRuntimeThinkingProfile struct {
	DefaultMode       string   `json:"defaultMode"`
	AllowedModes      []string `json:"allowedModes"`
	ManualDeprecated  bool     `json:"manualDeprecated"`
	EffortValues      []string `json:"effortValues"`
	ReasoningModes    []string `json:"reasoningModes,omitempty"`
	ContextValues     []string `json:"contextValues,omitempty"`
	AlwaysEnabled     bool     `json:"alwaysEnabled,omitempty"`
	DisabledMaxEffort string   `json:"disabledMaxEffort,omitempty"`
}

type ModelRuntimeSamplingProfile struct {
	Parameters             []string `json:"parameters"`
	DisallowedWithThinking []string `json:"disallowedWithThinking"`
}

type ModelRuntimeToolsProfile struct {
	Custom            bool              `json:"custom"`
	Server            bool              `json:"server"`
	Parallel          bool              `json:"parallel"`
	StructuredOutput  bool              `json:"structuredOutput"`
	Supported         []string          `json:"supported,omitempty"`
	PreferredVersions map[string]string `json:"preferredVersions,omitempty"`
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

type ModelRuntimeRequestHeaders struct {
	Allowed    []string            `json:"allowed,omitempty"`
	ByEndpoint map[string][]string `json:"byEndpoint,omitempty"`
}

type ModelRuntimeContextPricing struct {
	LongContextThresholdTokens int  `json:"longContextThresholdTokens,omitempty"`
	AppliesToAllTokens         bool `json:"appliesToAllTokens"`
}

type ModelRuntimeBatchProfile struct {
	Eligible           bool     `json:"eligible"`
	Discounted         bool     `json:"discounted"`
	SupportedEndpoints []string `json:"supportedEndpoints,omitempty"`
	MaxRequestBytes    int      `json:"maxRequestBytes,omitempty"`
}

type ModelRuntimeDeferredProfile struct {
	Eligible          bool `json:"eligible"`
	RetentionSeconds  int  `json:"retentionSeconds,omitempty"`
	MaximumRetrievals int  `json:"maximumRetrievals,omitempty"`
}

type ModelRuntimeResponseResources struct {
	Store            bool `json:"store"`
	Retrieve         bool `json:"retrieve"`
	Delete           bool `json:"delete"`
	Compact          bool `json:"compact"`
	CountInputTokens bool `json:"countInputTokens"`
}

type ModelRuntimeImageControls struct {
	AspectRatios           []string `json:"aspectRatios,omitempty"`
	Resolutions            []string `json:"resolutions,omitempty"`
	ResponseFormats        []string `json:"responseFormats,omitempty"`
	MaximumOutputs         int      `json:"maximumOutputs,omitempty"`
	MaximumReferenceImages int      `json:"maximumReferenceImages,omitempty"`
	EditTransport          string   `json:"editTransport,omitempty"`
	FilesInput             bool     `json:"filesInput"`
	StorageOptions         bool     `json:"storageOptions"`
	Streaming              bool     `json:"streaming"`
}

type ModelRuntimeUsageAccounting struct {
	RequiredResponseFields []string `json:"requiredResponseFields,omitempty"`
	ExactProviderCost      bool     `json:"exactProviderCost"`
}

type ModelRuntimeBetaCapability struct {
	Header          string   `json:"header"`
	Endpoints       []string `json:"endpoints"`
	MaxOutputTokens int      `json:"maxOutputTokens,omitempty"`
}

// ModelRuntimeRateLimitTier records provider-published reference quota
// metadata. Dynamic/project values remain distinct from GodarAI tenant limits.
type ModelRuntimeRateLimitTier struct {
	Tier                    int     `json:"tier,omitempty"`
	UsageTier               string  `json:"usageTier,omitempty"`
	Scope                   string  `json:"scope,omitempty"`
	Dynamic                 bool    `json:"dynamic,omitempty"`
	RequestsPerMinute       int     `json:"requestsPerMinute,omitempty"`
	RequestsPerDay          int     `json:"requestsPerDay,omitempty"`
	TokensPerMinute         int     `json:"tokensPerMinute,omitempty"`
	TokensPerDay            int     `json:"tokensPerDay,omitempty"`
	ImagesPerMinute         int     `json:"imagesPerMinute,omitempty"`
	BatchQueueTokens        int     `json:"batchQueueTokens,omitempty"`
	SpendUSDPerTenMinutes   float64 `json:"spendUsdPerTenMinutes,omitempty"`
	ConcurrentBatchRequests int     `json:"concurrentBatchRequests,omitempty"`
	InputFileBytes          int64   `json:"inputFileBytes,omitempty"`
	FileStorageBytes        int64   `json:"fileStorageBytes,omitempty"`
}

type ModelRuntimeLimits struct {
	ContextWindow      int `json:"contextWindow"`
	MaxInputTokens     int `json:"maxInputTokens,omitempty"`
	MaxInputCharacters int `json:"maxInputCharacters,omitempty"`
	MaxInputBytes      int `json:"maxInputBytes,omitempty"`
	MaxOutputTokens    int `json:"maxOutputTokens"`
	MessagesBytes      int `json:"messagesBytes"`
	BatchBytes         int `json:"batchBytes"`
	FileBytes          int `json:"fileBytes"`
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
	validators := []func() error{
		profile.validateIdentityAndEndpoints,
		profile.validateThinking,
		profile.validateSamplingToolsAliasesAndHeaders,
		profile.validateResourcePolicies,
		profile.validateModalitiesAndRequestControls,
		profile.validateBetaCapabilitiesAndLimits,
	}

	for _, validate := range validators {
		if err := validate(); err != nil {
			return err
		}
	}

	return nil
}

func (profile ModelRuntimeProfile) validateIdentityAndEndpoints() error {
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
				"language_models",
				"chat_completions",
				"deferred_chat_completions",
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
				"image_generation_models",
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
				"batch_requests",
				"batch_results",
				"file_public_urls",
				"generate_content",
				"stream_generate_content",
				"batch_generate_content",
				"batch_embed_contents",
				"cached_contents",
				"generate_answer",
				"embed_content",
				"imagen_generate",
				"fim_completions",
				"chat_moderations",
				"ocr",
				"agents",
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

	return nil
}

func (profile ModelRuntimeProfile) validateThinking() error {
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

	if profile.Thinking.AlwaysEnabled &&
		(len(profile.Thinking.AllowedModes) != 1 ||
			!profile.AllowsThinkingMode("adaptive")) {
		return fmt.Errorf("always-enabled thinking must allow only adaptive mode")
	}

	if profile.Thinking.DisabledMaxEffort != "" {
		if !profile.AllowsThinkingMode("disabled") {
			return fmt.Errorf("disabled max effort requires disabled thinking mode")
		}

		if !containsRuntimeValue(
			[]string{"low", "medium", "high", "xhigh", "max"},
			profile.Thinking.DisabledMaxEffort,
		) {
			return fmt.Errorf(
				"unsupported disabled max effort %q",
				profile.Thinking.DisabledMaxEffort,
			)
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

	return nil
}

func (profile ModelRuntimeProfile) validateSamplingToolsAliasesAndHeaders() error {
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
				"web_search_premium",
				"web_fetch",
				"file_search",
				"document_library",
				"tool_search",
				"image_generation",
				"code_interpreter",
				"code_execution",
				"hosted_shell",
				"bash",
				"text_editor",
				"apply_patch",
				"skills",
				"computer_use",
				"mcp",
				"memory",
				"advisor",
				"programmatic_tool_calling",
				"google_search",
				"google_maps",
				"url_context",
				"x_search",
				"attachment_search",
				"collections_search",
			},
			tool,
		) {
			return fmt.Errorf("unsupported model tool %q", tool)
		}
	}

	seenAliases := map[string]struct{}{}

	for _, alias := range profile.Aliases {
		alias = strings.TrimSpace(alias)
		if alias == "" {
			return fmt.Errorf("model alias cannot be empty")
		}

		if _, exists := seenAliases[alias]; exists {
			return fmt.Errorf("duplicate model alias %q", alias)
		}

		seenAliases[alias] = struct{}{}
	}

	for _, header := range profile.RequestHeaders.Allowed {
		if strings.TrimSpace(header) == "" || strings.ContainsAny(header, "\r\n") {
			return fmt.Errorf("invalid provider request header %q", header)
		}
	}

	for endpoint, headers := range profile.RequestHeaders.ByEndpoint {
		if !profile.AllowsEndpoint(endpoint) {
			return fmt.Errorf("request header policy uses unsupported endpoint %q", endpoint)
		}

		for _, header := range headers {
			if strings.TrimSpace(header) == "" || strings.ContainsAny(header, "\r\n") {
				return fmt.Errorf("invalid provider request header %q", header)
			}
		}
	}

	return nil
}

func (profile ModelRuntimeProfile) validateResourcePolicies() error {
	if profile.ContextPricing.LongContextThresholdTokens < 0 {
		return fmt.Errorf("long-context threshold cannot be negative")
	}

	if profile.Batch.Eligible && len(profile.Batch.SupportedEndpoints) == 0 {
		return fmt.Errorf("eligible batch profile requires supported endpoints")
	}

	for _, endpoint := range profile.Batch.SupportedEndpoints {
		if !profile.AllowsEndpoint(endpoint) {
			return fmt.Errorf("batch profile uses unsupported endpoint %q", endpoint)
		}
	}

	if profile.Batch.MaxRequestBytes < 0 || profile.DeferredCompletions.RetentionSeconds < 0 || profile.DeferredCompletions.MaximumRetrievals < 0 {
		return fmt.Errorf("runtime retention and request limits cannot be negative")
	}

	if profile.ImageControls.EditTransport != "" && profile.ImageControls.EditTransport != "json" && profile.ImageControls.EditTransport != "multipart" {
		return fmt.Errorf("unsupported image edit transport %q", profile.ImageControls.EditTransport)
	}

	if profile.ImageControls.MaximumOutputs < 0 || profile.ImageControls.MaximumReferenceImages < 0 {
		return fmt.Errorf("image control limits cannot be negative")
	}

	for _, field := range profile.UsageAccounting.RequiredResponseFields {
		if !containsRuntimeValue([]string{"cost_in_usd_ticks", "num_server_side_tools_used", "service_tier"}, field) {
			return fmt.Errorf("unsupported required usage field %q", field)
		}
	}

	for tool, version := range profile.Tools.PreferredVersions {
		if !profile.AllowsTool(tool) {
			return fmt.Errorf(
				"preferred version uses unsupported model tool %q",
				tool,
			)
		}

		if strings.TrimSpace(version) == "" ||
			strings.ContainsAny(version, ",\r\n") {
			return fmt.Errorf(
				"invalid preferred version %q for model tool %q",
				version,
				tool,
			)
		}
	}

	if profile.MinimumRetentionDays < 0 {
		return fmt.Errorf("minimum retention days cannot be negative")
	}

	return nil
}

func (profile ModelRuntimeProfile) validateModalitiesAndRequestControls() error {
	if len(profile.Modalities.Input) == 0 ||
		len(profile.Modalities.Output) == 0 {
		return fmt.Errorf("model runtime profile modalities are required")
	}

	for _, modality := range profile.Modalities.Input {
		if !containsRuntimeValue(
			[]string{"unknown", "text", "image", "pdf", "document", "audio", "video", "code"},
			modality,
		) {
			return fmt.Errorf("unsupported input modality %q", modality)
		}
	}

	for _, modality := range profile.Modalities.Output {
		if !containsRuntimeValue(
			[]string{"unknown", "text", "image", "audio", "video", "embedding", "code", "document", "moderation_result"},
			modality,
		) {
			return fmt.Errorf("unsupported output modality %q", modality)
		}
	}

	for endpoint, parameters := range profile.Parameters {
		if !profile.AllowsEndpoint(endpoint) {
			return fmt.Errorf("parameter policy uses unsupported endpoint %q", endpoint)
		}

		for _, parameter := range parameters {
			if strings.TrimSpace(parameter) == "" {
				return fmt.Errorf("empty parameter in endpoint %q", endpoint)
			}
		}
	}

	for _, value := range profile.RequestControls.ServiceTierValues {
		if !containsRuntimeValue(
			[]string{"auto", "standard_only", "default", "standard", "batch", "flex", "priority", "fast"},
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

	return nil
}

func (profile ModelRuntimeProfile) validateBetaCapabilitiesAndLimits() error {
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

	seenRateLimitTiers := map[int]struct{}{}
	seenUsageTiers := map[string]struct{}{}

	for _, limits := range profile.ReferenceRateLimits {
		if limits.Tier == 0 && strings.TrimSpace(limits.UsageTier) == "" {
			return fmt.Errorf("reference rate-limit tier identity is required")
		}

		if limits.Tier < 0 || limits.Tier > 5 {
			return fmt.Errorf("reference rate-limit tier must be between 1 and 5")
		}

		if limits.Tier > 0 {
			if _, exists := seenRateLimitTiers[limits.Tier]; exists {
				return fmt.Errorf("duplicate reference rate-limit tier %d", limits.Tier)
			}

			seenRateLimitTiers[limits.Tier] = struct{}{}
		}

		usageTier := strings.ToLower(strings.TrimSpace(limits.UsageTier))
		if usageTier != "" {
			if !containsRuntimeValue([]string{"free", "tier_1", "tier_2", "tier_3"}, usageTier) {
				return fmt.Errorf("unsupported reference usage tier %q", limits.UsageTier)
			}

			if _, exists := seenUsageTiers[usageTier]; exists {
				return fmt.Errorf("duplicate reference usage tier %q", usageTier)
			}

			seenUsageTiers[usageTier] = struct{}{}
		}

		if limits.Scope != "" && limits.Scope != "project" {
			return fmt.Errorf("unsupported reference rate-limit scope %q", limits.Scope)
		}

		if limits.RequestsPerMinute < 0 ||
			limits.RequestsPerDay < 0 ||
			limits.TokensPerMinute < 0 ||
			limits.TokensPerDay < 0 ||
			limits.ImagesPerMinute < 0 ||
			limits.BatchQueueTokens < 0 ||
			limits.SpendUSDPerTenMinutes < 0 ||
			limits.ConcurrentBatchRequests < 0 ||
			limits.InputFileBytes < 0 ||
			limits.FileStorageBytes < 0 {
			return fmt.Errorf("reference rate-limit values cannot be negative")
		}
	}

	if profile.Limits.ContextWindow < 0 ||
		profile.Limits.MaxInputTokens < 0 ||
		profile.Limits.MaxInputCharacters < 0 ||
		profile.Limits.MaxInputBytes < 0 ||
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
	if normalized == "" {
		return false
	}

	if normalized == normalizeRuntimeModelID(binding.CanonicalID, binding.Provider) ||
		normalized == normalizeRuntimeModelID(binding.Slug, binding.Provider) {
		return true
	}

	for _, alias := range binding.Profile.Aliases {
		if normalized == normalizeRuntimeModelID(alias, binding.Provider) {
			return true
		}
	}

	return false
}

func (profile ModelRuntimeProfile) AllowsEndpoint(endpoint string) bool {
	return containsRuntimeValue(profile.Endpoints, endpoint)
}

// AllowsTransport reports whether the imported model contract permits the
// requested wire transport (for example HTTP or SSE).
func (profile ModelRuntimeProfile) AllowsTransport(transport string) bool {
	return containsRuntimeValue(profile.Transports, transport)
}

func (profile ModelRuntimeProfile) AllowsThinkingMode(mode string) bool {
	return containsRuntimeValue(profile.Thinking.AllowedModes, mode)
}

func (profile ModelRuntimeProfile) AllowsEffort(effort string) bool {
	return containsRuntimeValue(profile.Thinking.EffortValues, effort)
}

func (profile ModelRuntimeProfile) AllowsTool(tool string) bool {
	return containsRuntimeValue(profile.Tools.Supported, tool)
}

func (profile ModelRuntimeProfile) AllowsDisabledEffort(effort string) bool {
	if profile.Thinking.DisabledMaxEffort == "" || effort == "" {
		return true
	}

	return runtimeEffortRank(effort) <=
		runtimeEffortRank(profile.Thinking.DisabledMaxEffort)
}

func runtimeEffortRank(value string) int {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "low", "minimal":
		return 1
	case "medium":
		return 2
	case "high":
		return 3
	case "xhigh":
		return 4
	case "max":
		return 5
	default:
		return 0
	}
}

func (profile ModelRuntimeProfile) AllowsSamplingParameter(
	parameter string,
) bool {
	return containsRuntimeValue(profile.Sampling.Parameters, parameter)
}

func (profile ModelRuntimeProfile) AllowsInputModality(modality string) bool {
	return containsRuntimeValue(profile.Modalities.Input, modality)
}

// AllowsParameter reports whether a provider-native request field is
// explicitly allowed for the endpoint. Empty endpoint policies deny fields.
func (profile ModelRuntimeProfile) AllowsParameter(endpoint, parameter string) bool {
	for configuredEndpoint, parameters := range profile.Parameters {
		if strings.EqualFold(strings.TrimSpace(configuredEndpoint), strings.TrimSpace(endpoint)) {
			return containsRuntimeValue(parameters, parameter)
		}
	}

	return false
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
