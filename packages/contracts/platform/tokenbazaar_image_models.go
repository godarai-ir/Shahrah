package platform

// TokenBazaarImageModelFamilies is the single source of truth for which
// TokenBazaar image models exist and which direct provider's image cost
// dimensions/billing shape each one follows ("openai" or "google-gemini").
//
// Both services/control-plane's billing package (FinalizeImage, via
// TokenBazaarImageModelFamily below) and services/silk-road's tokenbazaar
// adapter (dispatch gating) key off this map instead of each maintaining its
// own independent copy: those two services do not otherwise share a
// dependency edge, but both already import this contracts package, so this
// is the one place a 5th TokenBazaar image model needs to be added.
var TokenBazaarImageModelFamilies = map[string]string{
	"gpt-image-2":                 "openai",
	"gemini-3-pro-image":          "google-gemini",
	"gemini-3.1-flash-image":      "google-gemini",
	"gemini-3.1-flash-lite-image": "google-gemini",
}

// TokenBazaarImageModelFamily returns the billing family for a bare
// (already-normalized) TokenBazaar image model id, or "" if model is not one
// of TokenBazaar's in-scope image models.
func TokenBazaarImageModelFamily(model string) string {
	return TokenBazaarImageModelFamilies[model]
}
