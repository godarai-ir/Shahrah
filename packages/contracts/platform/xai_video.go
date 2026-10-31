package platform

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strings"
)

// xaiVideoBodyFields lists every xAI video request field GodarAI forwards.
// `output` (caller-signed upload URLs) and `storage_options` are deliberately
// absent: both are gated until their delivery and ownership paths exist.
var xaiVideoBodyFields = map[string][]string{
	VideoOperationGenerate: {
		"model", "prompt", "duration", "aspect_ratio", "resolution", "image",
		"last_frame", "keyframes", "reference_images", "reference_audios", "user",
	},
	VideoOperationEdit:   {"model", "prompt", "video", "user"},
	VideoOperationExtend: {"model", "prompt", "video", "duration", "user"},
}

type xaiVideoMedia struct {
	URL      string `json:"url"`
	ImageURL string `json:"image_url"`
	FileID   string `json:"file_id"`
}

// ParseXAIVideoRequest strictly decodes an xAI video request body and returns
// its canonical facts. Unknown fields are rejected, never passed through.
func ParseXAIVideoRequest(operation string, body []byte) (VideoRequestFacts, error) {
	facts := VideoRequestFacts{Operation: operation}

	allowed, ok := xaiVideoBodyFields[operation]
	if !ok {
		return facts, fmt.Errorf("unsupported video operation %q", operation)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(body, &fields); err != nil || fields == nil {
		return facts, fmt.Errorf("request body must be a JSON object")
	}

	for field := range fields {
		if !containsRuntimeValue(allowed, field) {
			return facts, fmt.Errorf("parameter %q is not supported for video %s", field, operation)
		}
	}

	if err := parseXAIVideoScalars(fields, &facts); err != nil {
		return facts, err
	}

	if operation != VideoOperationGenerate {
		var source xaiVideoMedia
		if err := decodeStrictJSON(fields["video"], &source); err != nil || validateXAIMedia(source, "video") != nil {
			return facts, fmt.Errorf("video must provide exactly one of url or file_id")
		}

		return facts, nil
	}

	err := parseXAIVideoReferences(fields, &facts)

	return facts, err
}

func parseXAIVideoScalars(fields map[string]json.RawMessage, facts *VideoRequestFacts) error {
	var prompt string
	if raw, ok := fields["prompt"]; ok {
		if err := json.Unmarshal(raw, &prompt); err != nil {
			return fmt.Errorf("prompt must be a string")
		}
	}

	facts.PromptLength = len([]rune(prompt))
	facts.HasPrompt = strings.TrimSpace(prompt) != ""

	if raw, ok := fields["duration"]; ok && string(raw) != "null" {
		seconds, err := ParseVideoSeconds(raw)
		if err != nil {
			return err
		}

		facts.DurationSeconds = seconds
	}

	for field, target := range map[string]*string{"aspect_ratio": &facts.AspectRatio, "resolution": &facts.Resolution} {
		if raw, ok := fields[field]; ok && json.Unmarshal(raw, target) != nil {
			return fmt.Errorf("%s must be a string", field)
		}
	}

	return nil
}

func parseXAIVideoReferences(fields map[string]json.RawMessage, facts *VideoRequestFacts) error {
	for field, flag := range map[string]*bool{"image": &facts.FirstFrame, "last_frame": &facts.LastFrame} {
		raw, ok := fields[field]
		if !ok || string(raw) == "null" {
			continue
		}

		var media xaiVideoMedia
		if err := decodeStrictJSON(raw, &media); err != nil || validateXAIMedia(media, "image") != nil {
			return fmt.Errorf("%s must provide exactly one of url or file_id", field)
		}

		*flag = true

		facts.trackInlineImage(media)
	}

	var references []xaiVideoMedia
	if err := decodeOptionalArray(fields["reference_images"], &references); err != nil {
		return fmt.Errorf("reference_images must be an array of images")
	}

	for _, reference := range references {
		if err := validateXAIMedia(reference, "image"); err != nil {
			return fmt.Errorf("reference_images entries must provide exactly one of url or file_id")
		}

		facts.trackInlineImage(reference)
	}

	facts.ReferenceImages = len(references)

	if err := parseXAIKeyframes(fields["keyframes"], facts); err != nil {
		return err
	}

	var voices []struct {
		VoiceID string  `json:"voice_id"`
		URL     *string `json:"url"`
	}
	if err := decodeOptionalArray(fields["reference_audios"], &voices); err != nil {
		return fmt.Errorf("reference_audios must be an array")
	}

	for _, voice := range voices {
		// Caller-supplied voice clips are a trusted-partner feature upstream.
		if voice.URL != nil || strings.TrimSpace(voice.VoiceID) == "" {
			return fmt.Errorf("reference_audios entries must use a supported voice_id")
		}

		facts.ReferenceVoices = append(facts.ReferenceVoices, strings.TrimSpace(voice.VoiceID))
	}

	return nil
}

func parseXAIKeyframes(raw json.RawMessage, facts *VideoRequestFacts) error {
	var keyframes []struct {
		Image      xaiVideoMedia `json:"image"`
		TimestampS float64       `json:"timestamp_s"`
	}
	if err := decodeOptionalArray(raw, &keyframes); err != nil {
		return fmt.Errorf("keyframes must be an array of {image, timestamp_s}")
	}

	for _, keyframe := range keyframes {
		if validateXAIMedia(keyframe.Image, "image") != nil || keyframe.TimestampS <= 0 ||
			math.IsNaN(keyframe.TimestampS) ||
			(facts.DurationSeconds > 0 && keyframe.TimestampS >= float64(facts.DurationSeconds)) {
			return fmt.Errorf("keyframes need an image and a timestamp_s strictly inside the clip")
		}

		facts.trackInlineImage(keyframe.Image)
	}

	facts.Keyframes = len(keyframes)

	return nil
}

// ParseVideoSeconds accepts a duration as a JSON integer or an integer string,
// matching both the OpenAI (`"8"`) and xAI (`8`) request shapes.
func ParseVideoSeconds(raw json.RawMessage) (int, error) {
	var text string
	if json.Unmarshal(raw, &text) == nil {
		raw = json.RawMessage(strings.TrimSpace(text))
	}

	var number float64
	if err := json.Unmarshal(raw, &number); err != nil || number != math.Trunc(number) || number < 1 || number > 3600 {
		return 0, fmt.Errorf("duration must be a whole number of seconds")
	}

	return int(number), nil
}

// trackInlineImage records the decoded size of a base64 data-URL image.
func (facts *VideoRequestFacts) trackInlineImage(media xaiVideoMedia) {
	location := strings.TrimSpace(media.URL)
	if location == "" {
		location = strings.TrimSpace(media.ImageURL)
	}

	_, payload, found := strings.Cut(location, ";base64,")
	if !strings.HasPrefix(location, "data:") || !found {
		return
	}

	if size := base64.StdEncoding.DecodedLen(len(payload)); size > facts.LargestInlineImageBytes {
		facts.LargestInlineImageBytes = size
	}
}

func validateXAIMedia(media xaiVideoMedia, kind string) error {
	location := strings.TrimSpace(media.URL)
	if location == "" {
		location = strings.TrimSpace(media.ImageURL)
	}

	fileID := strings.TrimSpace(media.FileID)
	if (location == "") == (fileID == "") {
		return fmt.Errorf("exactly one of url or file_id is required")
	}

	if location == "" {
		return nil
	}

	if strings.HasPrefix(location, "data:") {
		allowed := []string{"data:image/jpeg;", "data:image/png;", "data:image/webp;"}
		if kind == "video" {
			allowed = []string{"data:video/mp4;"}
		}

		for _, prefix := range allowed {
			if strings.HasPrefix(location, prefix) {
				return nil
			}
		}

		return fmt.Errorf("unsupported %s data URL type", kind)
	}

	parsed, err := url.Parse(location)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil {
		return fmt.Errorf("%s url must be https or a data URL", kind)
	}

	return nil
}

func decodeStrictJSON(raw json.RawMessage, target interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()

	return decoder.Decode(target)
}

func decodeOptionalArray(raw json.RawMessage, target interface{}) error {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}

	return decodeStrictJSON(raw, target)
}
