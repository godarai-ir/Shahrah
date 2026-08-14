package platform

import "strings"

// SupportedGoogleGeminiModelIDs is the intentionally curated provider
// boundary shared by Control Plane and Silk Road.
var SupportedGoogleGeminiModelIDs = []string{
	"gemini-3.1-pro-preview",
	"gemini-3.5-flash",
	"gemini-3.5-flash-lite",
	"gemini-3.6-flash",
	"gemini-3.7-flash",
	"gemini-3-pro-image",
	"gemini-3.1-flash-image",
	"gemini-3.1-flash-lite-image",
	"gemini-3.1-flash-tts-preview",
	"gemini-embedding-2",
}

// IsSupportedGoogleGeminiModel checks a resolved canonical upstream model ID.
func IsSupportedGoogleGeminiModel(model string) bool {
	model = strings.TrimPrefix(strings.TrimSpace(model), "models/")
	for _, candidate := range SupportedGoogleGeminiModelIDs {
		if model == candidate {
			return true
		}
	}

	return false
}
