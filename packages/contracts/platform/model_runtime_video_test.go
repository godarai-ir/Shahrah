package platform

import (
	"strings"
	"testing"
)

func videoContractProfile() ModelRuntimeProfile {
	return ModelRuntimeProfile{
		Version: 4, Endpoints: []string{"videos"},
		VideoControls: ModelRuntimeVideoControls{
			Operations:         []string{VideoOperationGenerate, VideoOperationExtend},
			MinDurationSeconds: 1, MaxDurationSeconds: 15, DefaultDurationSeconds: 8,
			MinExtensionSeconds: 2, MaxExtensionSeconds: 10, DefaultExtensionSeconds: 6,
			AspectRatios: []string{"16:9", "9:16"}, Resolutions: []string{"480p", "720p", "1080p"},
			DefaultResolution: "480p", ReferenceResolutions: []string{"480p", "720p"},
			Sizes:            map[string]ModelRuntimeVideoSize{"1280x720": {AspectRatio: "16:9", Resolution: "720p"}},
			ReferenceModes:   []string{VideoReferenceFirstFrame, VideoReferenceKeyframes, VideoReferenceImages, VideoReferenceVoices},
			MaximumKeyframes: 4, MaximumReferenceImages: 2, MaximumReferenceVoices: 1,
			ReferenceVoiceIDs: []string{"eve"}, MaxInputImageBytes: 64,
			ContentVariants: []string{"video"}, ContentHosts: []string{"vidgen.x.ai"},
		},
	}
}

func TestVideoControlsValidation(t *testing.T) {
	t.Parallel()

	tests := map[string]func(*ModelRuntimeProfile){
		"controls without endpoint": func(p *ModelRuntimeProfile) { p.Endpoints = []string{"image_generations"} },
		"unknown operation":         func(p *ModelRuntimeProfile) { p.VideoControls.Operations = []string{"remix"} },
		"default outside range":     func(p *ModelRuntimeProfile) { p.VideoControls.DefaultDurationSeconds = 20 },
		"missing extension range":   func(p *ModelRuntimeProfile) { p.VideoControls.MaxExtensionSeconds = 0 },
		"edit without source limit": func(p *ModelRuntimeProfile) {
			p.VideoControls.Operations = append(p.VideoControls.Operations, VideoOperationEdit)
		},
		"size with unknown format": func(p *ModelRuntimeProfile) {
			p.VideoControls.Sizes["1920x1080"] = ModelRuntimeVideoSize{AspectRatio: "16:9", Resolution: "4k"}
		},
		"voices without roster":   func(p *ModelRuntimeProfile) { p.VideoControls.ReferenceVoiceIDs = nil },
		"host with scheme":        func(p *ModelRuntimeProfile) { p.VideoControls.ContentHosts = []string{"https://vidgen.x.ai"} },
		"no content variant":      func(p *ModelRuntimeProfile) { p.VideoControls.ContentVariants = nil },
		"unknown content variant": func(p *ModelRuntimeProfile) { p.VideoControls.ContentVariants = []string{"gif"} },
	}

	if err := (videoContractProfile()).validateVideoControls(); err != nil {
		t.Fatalf("valid profile rejected: %v", err)
	}

	bare := ModelRuntimeProfile{Endpoints: []string{"videos"}}
	if err := bare.validateVideoControls(); err != nil || bare.AllowsVideoOperation(VideoOperationGenerate) {
		t.Fatalf("a bare videos endpoint must stay inert, got err=%v", err)
	}

	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			profile := videoContractProfile()
			profile.VideoControls.Sizes = map[string]ModelRuntimeVideoSize{"1280x720": {AspectRatio: "16:9", Resolution: "720p"}}
			mutate(&profile)

			if err := profile.validateVideoControls(); err == nil {
				t.Fatal("invalid video controls were accepted")
			}
		})
	}
}

func TestParseAndValidateXAIVideoRequests(t *testing.T) {
	t.Parallel()

	image := `{"url":"data:image/png;base64,` + strings.Repeat("A", 40) + `"}`
	tests := []struct {
		name, operation, body, wantError string
	}{
		{name: "text to video", operation: VideoOperationGenerate, body: `{"model":"m","prompt":"waves","duration":"8","resolution":"720p","aspect_ratio":"16:9"}`},
		{name: "image only", operation: VideoOperationGenerate, body: `{"model":"m","image":` + image + `}`},
		{name: "references and voice", operation: VideoOperationGenerate, body: `{"model":"m","prompt":"x","resolution":"720p","reference_images":[{"file_id":"f1"}],"reference_audios":[{"voice_id":"eve"}]}`},
		{name: "keyframe", operation: VideoOperationGenerate, body: `{"model":"m","prompt":"x","duration":6,"keyframes":[{"image":{"url":"https://a.example/k.png"},"timestamp_s":3}]}`},
		{name: "extension", operation: VideoOperationExtend, body: `{"model":"m","prompt":"next","duration":4,"video":{"url":"https://vidgen.x.ai/v.mp4"}}`},
		{name: "unknown field", operation: VideoOperationGenerate, body: `{"model":"m","prompt":"x","n":2}`, wantError: `"n"`},
		{name: "caller upload url", operation: VideoOperationGenerate, body: `{"model":"m","prompt":"x","output":{"upload_url":"https://evil"}}`, wantError: `"output"`},
		{name: "storage options gated", operation: VideoOperationGenerate, body: `{"model":"m","prompt":"x","storage_options":{"filename":"a"}}`, wantError: "storage_options"},
		{name: "fractional duration", operation: VideoOperationGenerate, body: `{"model":"m","prompt":"x","duration":2.5}`, wantError: "whole number"},
		{name: "duration out of range", operation: VideoOperationGenerate, body: `{"model":"m","prompt":"x","duration":16}`, wantError: "between 1 and 15"},
		{name: "unsupported resolution", operation: VideoOperationGenerate, body: `{"model":"m","prompt":"x","resolution":"4k"}`, wantError: "resolution"},
		{name: "references above 720p", operation: VideoOperationGenerate, body: `{"model":"m","prompt":"x","resolution":"1080p","reference_images":[{"file_id":"f1"}]}`, wantError: "reference inputs"},
		{name: "too many references", operation: VideoOperationGenerate, body: `{"model":"m","prompt":"x","reference_images":[{"file_id":"a"},{"file_id":"b"},{"file_id":"c"}]}`, wantError: "at most 2"},
		{name: "unlisted voice", operation: VideoOperationGenerate, body: `{"model":"m","prompt":"x","reference_audios":[{"voice_id":"zed"}]}`, wantError: "zed"},
		{name: "caller voice clip", operation: VideoOperationGenerate, body: `{"model":"m","prompt":"x","reference_audios":[{"url":"https://a/x.wav"}]}`, wantError: "voice_id"},
		{name: "keyframe outside clip", operation: VideoOperationGenerate, body: `{"model":"m","prompt":"x","duration":4,"keyframes":[{"image":{"file_id":"k"},"timestamp_s":4}]}`, wantError: "strictly inside"},
		{name: "plain http image", operation: VideoOperationGenerate, body: `{"model":"m","prompt":"x","image":{"url":"http://a/i.png"}}`, wantError: "url"},
		{name: "image with both refs", operation: VideoOperationGenerate, body: `{"model":"m","prompt":"x","image":{"url":"https://a/i.png","file_id":"f"}}`, wantError: "exactly one"},
		{name: "no prompt with references", operation: VideoOperationGenerate, body: `{"model":"m","reference_images":[{"file_id":"a"}]}`, wantError: "prompt is required"},
		{name: "inline image too large", operation: VideoOperationGenerate, body: `{"model":"m","prompt":"x","image":{"url":"data:image/png;base64,` + strings.Repeat("A", 200) + `"}}`, wantError: "at most 64 bytes"},
		{name: "edit not enabled", operation: VideoOperationEdit, body: `{"model":"m","prompt":"x","video":{"file_id":"v"}}`, wantError: "does not support video edit"},
		{name: "extension too long", operation: VideoOperationExtend, body: `{"model":"m","prompt":"x","duration":11,"video":{"file_id":"v"}}`, wantError: "between 2 and 10"},
		{name: "extension needs source", operation: VideoOperationExtend, body: `{"model":"m","prompt":"x"}`, wantError: "video must provide"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			facts, err := ParseXAIVideoRequest(test.operation, []byte(test.body))
			if err == nil {
				err = videoContractProfile().ValidateVideoRequest(facts)
			}

			switch {
			case test.wantError == "" && err != nil:
				t.Fatalf("unexpected error: %v", err)
			case test.wantError != "" && (err == nil || !strings.Contains(err.Error(), test.wantError)):
				t.Fatalf("error = %v, want it to contain %q", err, test.wantError)
			}
		})
	}
}

func TestParseVideoSecondsAcceptsBothShapes(t *testing.T) {
	t.Parallel()

	for raw, want := range map[string]int{`8`: 8, `"8"`: 8, `" 12 "`: 12} {
		if got, err := ParseVideoSeconds([]byte(raw)); err != nil || got != want {
			t.Fatalf("ParseVideoSeconds(%s) = %d, %v", raw, got, err)
		}
	}

	for _, raw := range []string{`0`, `-1`, `"abc"`, `true`, `1.5`} {
		if _, err := ParseVideoSeconds([]byte(raw)); err == nil {
			t.Fatalf("ParseVideoSeconds(%s) accepted an invalid value", raw)
		}
	}
}
