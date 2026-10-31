package platform

import (
	"fmt"
	"strings"
)

// Video operations a model profile can enable.
const (
	VideoOperationGenerate = "generate"
	VideoOperationEdit     = "edit"
	VideoOperationExtend   = "extend"
)

// Video reference modes a model profile can enable.
const (
	VideoReferenceFirstFrame       = "first_frame"
	VideoReferenceLastFrame        = "last_frame"
	VideoReferenceKeyframes        = "keyframes"
	VideoReferenceImages           = "reference_images"
	VideoReferenceVoices           = "reference_voices"
	videoEndpointName              = "videos"
	videoContentVariantVideo       = "video"
	videoContentVariantThumbnail   = "thumbnail"
	videoContentVariantSpritesheet = "spritesheet"
)

// ModelRuntimeVideoControls is the provider-neutral video generation contract
// of one model. Request validation reads only from this block, so adding a
// provider is data rather than code.
type ModelRuntimeVideoControls struct {
	Operations              []string                         `json:"operations,omitempty"`
	MinDurationSeconds      int                              `json:"minDurationSeconds,omitempty"`
	MaxDurationSeconds      int                              `json:"maxDurationSeconds,omitempty"`
	DefaultDurationSeconds  int                              `json:"defaultDurationSeconds,omitempty"`
	MinExtensionSeconds     int                              `json:"minExtensionSeconds,omitempty"`
	MaxExtensionSeconds     int                              `json:"maxExtensionSeconds,omitempty"`
	DefaultExtensionSeconds int                              `json:"defaultExtensionSeconds,omitempty"`
	MaxEditSourceSeconds    float64                          `json:"maxEditSourceSeconds,omitempty"`
	AspectRatios            []string                         `json:"aspectRatios,omitempty"`
	Resolutions             []string                         `json:"resolutions,omitempty"`
	DefaultResolution       string                           `json:"defaultResolution,omitempty"`
	ReferenceResolutions    []string                         `json:"referenceResolutions,omitempty"`
	Sizes                   map[string]ModelRuntimeVideoSize `json:"sizes,omitempty"`
	ReferenceModes          []string                         `json:"referenceModes,omitempty"`
	MaximumKeyframes        int                              `json:"maximumKeyframes,omitempty"`
	MaximumReferenceImages  int                              `json:"maximumReferenceImages,omitempty"`
	MaximumReferenceVoices  int                              `json:"maximumReferenceVoices,omitempty"`
	ReferenceVoiceIDs       []string                         `json:"referenceVoiceIds,omitempty"`
	MaxInputImageBytes      int                              `json:"maxInputImageBytes,omitempty"`
	MaxPromptCharacters     int                              `json:"maxPromptCharacters,omitempty"`
	ContentVariants         []string                         `json:"contentVariants,omitempty"`
	ContentHosts            []string                         `json:"contentHosts,omitempty"`
	ExtensionFields         []string                         `json:"extensionFields,omitempty"`
	Audio                   bool                             `json:"audio,omitempty"`
}

// ModelRuntimeVideoSize maps an OpenAI-style `size` value onto the provider's
// aspect ratio and resolution.
type ModelRuntimeVideoSize struct {
	AspectRatio string `json:"aspectRatio"`
	Resolution  string `json:"resolution"`
}

// VideoRequestFacts is the canonical, provider-neutral view of one video
// request that both the control plane and the provider egress validate.
type VideoRequestFacts struct {
	Operation       string
	DurationSeconds int
	AspectRatio     string
	Resolution      string
	PromptLength    int
	HasPrompt       bool
	FirstFrame      bool
	LastFrame       bool
	Keyframes       int
	ReferenceImages int
	ReferenceVoices []string
	ExtensionFields []string
	// LargestInlineImageBytes is the decoded size of the largest data-URL image.
	LargestInlineImageBytes int
}

// AllowsVideoOperation reports whether the profile enables one video operation.
func (profile ModelRuntimeProfile) AllowsVideoOperation(operation string) bool {
	return profile.AllowsEndpoint(videoEndpointName) &&
		containsRuntimeValue(profile.VideoControls.Operations, operation)
}

func (profile ModelRuntimeProfile) validateVideoControls() error {
	controls := profile.VideoControls
	if !profile.AllowsEndpoint(videoEndpointName) {
		if len(controls.Operations) > 0 {
			return fmt.Errorf("video controls require the videos endpoint")
		}

		return nil
	}

	// A bare videos endpoint advertises nothing until operations are declared.
	if len(controls.Operations) == 0 {
		return nil
	}

	for _, operation := range controls.Operations {
		if !containsRuntimeValue([]string{VideoOperationGenerate, VideoOperationEdit, VideoOperationExtend}, operation) {
			return fmt.Errorf("unsupported video operation %q", operation)
		}
	}

	if err := controls.validateDurations(); err != nil {
		return err
	}

	if err := controls.validateFormats(); err != nil {
		return err
	}

	if err := controls.validateReferences(); err != nil {
		return err
	}

	return controls.validateDelivery()
}

func (controls ModelRuntimeVideoControls) validateDurations() error {
	if containsRuntimeValue(controls.Operations, VideoOperationGenerate) &&
		!validDurationRange(controls.MinDurationSeconds, controls.MaxDurationSeconds, controls.DefaultDurationSeconds) {
		return fmt.Errorf("video generation requires a valid duration range and default")
	}

	if containsRuntimeValue(controls.Operations, VideoOperationExtend) &&
		!validDurationRange(controls.MinExtensionSeconds, controls.MaxExtensionSeconds, controls.DefaultExtensionSeconds) {
		return fmt.Errorf("video extension requires a valid duration range and default")
	}

	if containsRuntimeValue(controls.Operations, VideoOperationEdit) && controls.MaxEditSourceSeconds <= 0 {
		return fmt.Errorf("video edits require a positive maximum source duration")
	}

	return nil
}

func validDurationRange(minimum, maximum, fallback int) bool {
	return minimum >= 1 && maximum >= minimum && fallback >= minimum && fallback <= maximum
}

func (controls ModelRuntimeVideoControls) validateFormats() error {
	if !containsRuntimeValue(controls.Operations, VideoOperationGenerate) {
		return nil
	}

	if len(controls.AspectRatios) == 0 || len(controls.Resolutions) == 0 {
		return fmt.Errorf("video generation requires aspect ratios and resolutions")
	}

	if !containsRuntimeValue(controls.Resolutions, controls.DefaultResolution) {
		return fmt.Errorf("default video resolution %q is not allowed", controls.DefaultResolution)
	}

	for _, resolution := range controls.ReferenceResolutions {
		if !containsRuntimeValue(controls.Resolutions, resolution) {
			return fmt.Errorf("reference video resolution %q is not allowed", resolution)
		}
	}

	for size, mapping := range controls.Sizes {
		if strings.TrimSpace(size) == "" ||
			!containsRuntimeValue(controls.AspectRatios, mapping.AspectRatio) ||
			!containsRuntimeValue(controls.Resolutions, mapping.Resolution) {
			return fmt.Errorf("video size %q maps to an unsupported format", size)
		}
	}

	return nil
}

func (controls ModelRuntimeVideoControls) validateReferences() error {
	allowed := []string{
		VideoReferenceFirstFrame, VideoReferenceLastFrame, VideoReferenceKeyframes,
		VideoReferenceImages, VideoReferenceVoices,
	}

	for _, mode := range controls.ReferenceModes {
		if !containsRuntimeValue(allowed, mode) {
			return fmt.Errorf("unsupported video reference mode %q", mode)
		}
	}

	if controls.MaximumKeyframes < 0 || controls.MaximumReferenceImages < 0 ||
		controls.MaximumReferenceVoices < 0 || controls.MaxInputImageBytes < 0 ||
		controls.MaxPromptCharacters < 0 {
		return fmt.Errorf("video reference limits cannot be negative")
	}

	required := map[string]bool{
		VideoReferenceKeyframes:  controls.MaximumKeyframes > 0,
		VideoReferenceImages:     controls.MaximumReferenceImages > 0,
		VideoReferenceVoices:     controls.MaximumReferenceVoices > 0 && len(controls.ReferenceVoiceIDs) > 0,
		VideoReferenceFirstFrame: controls.MaxInputImageBytes > 0,
	}
	for mode, configured := range required {
		if containsRuntimeValue(controls.ReferenceModes, mode) && !configured {
			return fmt.Errorf("video reference mode %q requires its limit", mode)
		}
	}

	return nil
}

func (controls ModelRuntimeVideoControls) validateDelivery() error {
	if len(controls.ContentVariants) == 0 {
		return fmt.Errorf("videos endpoint requires at least one content variant")
	}

	for _, variant := range controls.ContentVariants {
		if !containsRuntimeValue([]string{videoContentVariantVideo, videoContentVariantThumbnail, videoContentVariantSpritesheet}, variant) {
			return fmt.Errorf("unsupported video content variant %q", variant)
		}
	}

	for _, host := range controls.ContentHosts {
		if host == "" || host != strings.ToLower(strings.TrimSpace(host)) || strings.ContainsAny(host, ":/?#@ ") {
			return fmt.Errorf("video content host %q must be a bare lowercase hostname", host)
		}
	}

	return nil
}

// ValidateVideoRequest checks one canonical request against the profile. It is
// shared by the control plane and the provider egress so both fail the same way.
func (profile ModelRuntimeProfile) ValidateVideoRequest(facts VideoRequestFacts) error {
	controls := profile.VideoControls
	if !profile.AllowsVideoOperation(facts.Operation) {
		return fmt.Errorf("model does not support video %s", facts.Operation)
	}

	for _, field := range facts.ExtensionFields {
		if !containsRuntimeValue(controls.ExtensionFields, field) {
			return fmt.Errorf("parameter %q is not supported by this video model", field)
		}
	}

	if controls.MaxInputImageBytes > 0 && facts.LargestInlineImageBytes > controls.MaxInputImageBytes {
		return fmt.Errorf("inline images must be at most %d bytes", controls.MaxInputImageBytes)
	}

	if controls.MaxPromptCharacters > 0 && facts.PromptLength > controls.MaxPromptCharacters {
		return fmt.Errorf("prompt exceeds %d characters", controls.MaxPromptCharacters)
	}

	switch facts.Operation {
	case VideoOperationGenerate:
		return controls.validateGenerateRequest(facts)
	case VideoOperationExtend:
		if !facts.HasPrompt {
			return fmt.Errorf("prompt is required for video extensions")
		}

		return validateDuration(facts.DurationSeconds, controls.MinExtensionSeconds, controls.MaxExtensionSeconds)
	default:
		if !facts.HasPrompt {
			return fmt.Errorf("prompt is required for video edits")
		}

		if facts.DurationSeconds != 0 || facts.Resolution != "" || facts.AspectRatio != "" {
			return fmt.Errorf("video edits keep the source duration and format")
		}

		return nil
	}
}

func (controls ModelRuntimeVideoControls) validateGenerateRequest(facts VideoRequestFacts) error {
	if err := validateDuration(facts.DurationSeconds, controls.MinDurationSeconds, controls.MaxDurationSeconds); err != nil {
		return err
	}

	if facts.AspectRatio != "" && !containsRuntimeValue(controls.AspectRatios, facts.AspectRatio) {
		return fmt.Errorf("aspect_ratio %q is not supported by this video model", facts.AspectRatio)
	}

	if facts.Resolution != "" && !containsRuntimeValue(controls.Resolutions, facts.Resolution) {
		return fmt.Errorf("resolution %q is not supported by this video model", facts.Resolution)
	}

	usesReferences := facts.ReferenceImages > 0 || len(facts.ReferenceVoices) > 0
	if !facts.HasPrompt && (!facts.FirstFrame || usesReferences) {
		return fmt.Errorf("prompt is required unless only a first-frame image is provided")
	}

	if usesReferences && facts.Resolution != "" && len(controls.ReferenceResolutions) > 0 &&
		!containsRuntimeValue(controls.ReferenceResolutions, facts.Resolution) {
		return fmt.Errorf("resolution %q is not supported with reference inputs", facts.Resolution)
	}

	checks := []struct {
		mode    string
		count   int
		maximum int
	}{
		{VideoReferenceFirstFrame, boolCount(facts.FirstFrame), 1},
		{VideoReferenceLastFrame, boolCount(facts.LastFrame), 1},
		{VideoReferenceKeyframes, facts.Keyframes, controls.MaximumKeyframes},
		{VideoReferenceImages, facts.ReferenceImages, controls.MaximumReferenceImages},
		{VideoReferenceVoices, len(facts.ReferenceVoices), controls.MaximumReferenceVoices},
	}
	for _, check := range checks {
		if check.count == 0 {
			continue
		}

		if !containsRuntimeValue(controls.ReferenceModes, check.mode) {
			return fmt.Errorf("video reference mode %q is not supported by this model", check.mode)
		}

		if check.count > check.maximum {
			return fmt.Errorf("at most %d %s inputs are allowed", check.maximum, check.mode)
		}
	}

	for _, voice := range facts.ReferenceVoices {
		if !containsRuntimeValue(controls.ReferenceVoiceIDs, voice) {
			return fmt.Errorf("voice_id %q is not supported by this video model", voice)
		}
	}

	return nil
}

func validateDuration(value, minimum, maximum int) error {
	if value != 0 && (value < minimum || value > maximum) {
		return fmt.Errorf("duration must be between %d and %d seconds", minimum, maximum)
	}

	return nil
}

func boolCount(value bool) int {
	if value {
		return 1
	}

	return 0
}
