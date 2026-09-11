package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

const (
	// tokenBazaarBaseURL is the authoritative TokenBazaar host per direct
	// user instruction (see plans/tokenbazaar-production-readiness.md,
	// "Base-URL reachability re-check"). This session independently
	// re-attempted DNS resolution/connection to this host from its own
	// execution environment (nslookup, curl to "/", curl to "/v1/models")
	// and got the same DNS-timeout failure signature already on record for
	// this host and for docs.deepwl.cn. Per the delivery's own fail-closed
	// rule, api.tokenbazaar.ai (this session's own earlier, now-downgraded
	// finding) is NOT substituted here, and no third host is invented.
	// validateTokenBazaarBaseURL below enforces an exact-host allowlist so a
	// misconfigured or malicious api_base can never silently redirect
	// dispatch elsewhere, and any DNS/connect failure at request time
	// surfaces as a plain Go network error from client.Do — never a silent
	// retry loop and never a silent fallback to a different host.
	tokenBazaarBaseURL         = "https://zx1.deepwl.net"
	tokenBazaarHost            = "zx1.deepwl.net"
	tokenBazaarProviderName    = "tokenbazaar"
	tokenBazaarRequestIDHeader = "X-Tokenbazaar-Request-Id"

	// tokenBazaarMaxErrorBodyBytes bounds how much of an upstream error body
	// this adapter will ever read into memory before normalizing/discarding
	// it (see normalizeTokenBazaarErrorResponse).
	tokenBazaarMaxErrorBodyBytes = 64 * 1024

	// tokenBazaarCredentialProbeModel is the model used for the explicit,
	// UI-approval-gated, billable connection test (see credential_probe.go).
	// It is the lowest-cost of the 21 in-scope chat models per this
	// delivery's spot check of the imported TokenBazaar catalog (see
	// plans/tokenbazaar-production-readiness.md, "Pricing anomalies" —
	// deepseek-v4-flash is priced roughly 85% below GodarAI's direct-provider
	// DeepSeek cost, making it the cheapest observed model in scope by a wide
	// margin). This is a fixed fallback for this Go-level probe wiring only;
	// conflict-register item #6 already flags that per-key pricing from the
	// live authenticated GET /v1/models should be the long-term source of
	// truth, and the tenant-admin connection-test dialog re-derives the
	// cheapest model from the imported catalog rather than trusting this
	// constant blindly.
	tokenBazaarCredentialProbeModel = "deepseek-v4-flash"
)

// tokenBazaarMaxResponseBytes is a response-size ceiling applied to every
// TokenBazaar response body (success or pass-through 429) to bound memory
// use against a misbehaving or compromised upstream. It is a package-level
// var (not a const) solely so tests can shrink it and assert the ceiling is
// enforced without allocating tens of megabytes of fixture data.
var tokenBazaarMaxResponseBytes int64 = 32 * 1024 * 1024

// tokenBazaarChatModels is the exhaustive, deliberately narrow set of
// TokenBazaar model IDs this delivery dispatches through the documented
// POST /v1/chat/completions envelope. This is 21 of the 26 models scaffolded
// under apps/model-discovery/providers/tokenbazaar/ — see
// tokenBazaarUnsupportedModels for the other 5, which are explicitly gated
// rather than silently accepted or 404'd.
var tokenBazaarChatModels = map[string]struct{}{
	"claude-fable-5":         {},
	"claude-haiku-4-5":       {},
	"claude-opus-5":          {},
	"claude-sonnet-5":        {},
	"deepseek-v4-flash":      {},
	"deepseek-v4-pro":        {},
	"gemini-2.5-flash":       {},
	"gemini-2.5-pro":         {},
	"gemini-3.1-pro-preview": {},
	"gemini-3.5-flash":       {},
	"gemini-3.6-flash":       {},
	"gemini-3.7-flash":       {},
	"glm-5.2":                {},
	"glm-5v-turbo":           {},
	"gpt-5.5":                {},
	"gpt-5.6-luna":           {},
	"gpt-5.6-sol":            {},
	"gpt-5.6-terra":          {},
	"kimi-k2.6":              {},
	"kimi-k2.7-code":         {},
	"kimi-k3":                {},
}

// tokenBazaarUnsupportedModels previously gated the 5 originally out-of-scope
// TokenBazaar models. All 5 now have real dispatch (the 4 image models via
// tokenBazaarImageModels/validateTokenBazaarImageDispatch, and the last
// remaining one, the embeddings model text-embedding-3-large, via
// tokenBazaarEmbeddingModels/validateTokenBazaarEmbeddingsDispatch below).
// Images are additionally gated by the live per-model parameter confirmation
// cache. Neither capability carries an independent code-level kill switch --
// see plans/tokenbazaar-production-readiness.md for the accepted-risk
// rationale. This map is kept (now empty) rather than deleted so any future
// out-of-scope TokenBazaar model can be gated the same explicit, actionable
// way instead of falling through to a generic 404/misroute.
var tokenBazaarUnsupportedModels = map[string]string{}

// tokenBazaarImageModels is the exhaustive set of TokenBazaar model IDs this
// delivery dispatches through the documented POST /v1/images/generations and
// POST /v1/images/edits envelopes. Per
// plans/tokenbazaar-production-readiness.md conflict #12, TokenBazaar's own
// "Supported Image Models" documentation table does not name the 3 Gemini
// image models, even though the full catalog lists them with
// endpoint_types: ["images"]; all 4 share this code path, but only
// "gpt-image-2" carries first-party documentation-table evidence. The other
// 3 need their own live probe before being considered equally confirmed —
// see the evidence-tier table in the readiness plan.
//
// Derived from platform.TokenBazaarImageModelFamilies, the single source of
// truth for TokenBazaar's in-scope image models also used by
// services/control-plane's billing package (tokenBazaarImageModelFamily) --
// so the two stay in sync without either copying the other's list.
var tokenBazaarImageModels = tokenBazaarImageModelSet()

func tokenBazaarImageModelSet() map[string]struct{} {
	set := make(map[string]struct{}, len(platform.TokenBazaarImageModelFamilies))
	for model := range platform.TokenBazaarImageModelFamilies {
		set[model] = struct{}{}
	}

	return set
}

// tokenBazaarEmbeddingModels is the exhaustive set of TokenBazaar model IDs
// this delivery dispatches through POST /v1/embeddings.
//
// Evidence-tier caveat (see plans/tokenbazaar-production-readiness.md,
// "Embeddings endpoint delivery"): unlike chat and images, TokenBazaar's own
// rendered docs page has no worked embeddings request/response example. This
// contract is inferred from (1) the docs' product description explicitly
// claiming embedding support, and (2) the authenticated GET /v1/models
// catalog listing this model with supported_endpoint_types: embeddings,
// combined with the fact that every other confirmed TokenBazaar capability
// follows an exact OpenAI/Anthropic-compatible path and shape with no
// deviation. It is a strongly-implied, not first-party-confirmed, contract:
// validateTokenBazaarEmbeddingsDispatch and
// validateTokenBazaarEmbeddingsResponseShape below are deliberately
// conservative as a result — see those functions' own comments.
var tokenBazaarEmbeddingModels = map[string]struct{}{
	"text-embedding-3-large": {},
}

// tokenBazaarEmbeddingsMinDimensions and tokenBazaarEmbeddingsMaxDimensions
// bound the optional "dimensions" embeddings request parameter for
// text-embedding-3-large. These are not TokenBazaar-documented values (see
// the evidence-tier caveat above) — they are reused verbatim from the
// already-vetted direct-provider catalog scaffold's own params entry
// (apps/model-discovery/providers/tokenbazaar/text-embedding-3-large.yaml:
// "dimensions", minValue 1, maxValue 3072, itself pinned from OpenAI's
// documented output_vector_size), rather than inventing a new bound or
// leaving the parameter unchecked.
const (
	tokenBazaarEmbeddingsMinDimensions = 1
	tokenBazaarEmbeddingsMaxDimensions = 3072
)

// tokenBazaarEmbeddingsEncodingFormats is the exhaustive set of accepted
// encoding_format values, matching the standard OpenAI-compatible embeddings
// contract and the direct-provider catalog scaffold's own params entry for
// this model.
var tokenBazaarEmbeddingsEncodingFormats = map[string]struct{}{
	"float":  {},
	"base64": {},
}

// tokenBazaarResponsesModels is the exhaustive set of TokenBazaar model IDs
// this delivery dispatches through the documented POST /v1/responses
// envelope. Per plans/tokenbazaar-production-readiness.md's "Confirmed wire
// contract", all 21 chat-capable models in scope can in principle speak
// /v1/responses (TokenBazaar normalizes to whichever of the three text
// envelopes the caller selects), but this delivery is deliberately scoped to
// exactly the 4 OpenAI-family models and the 2 DeepSeek models per the
// executable prompt driving this change -- the other 15 chat-capable models
// (Claude/Gemini/GLM/Kimi) are NOT added here and continue to reject a
// /v1/responses request with the generic tokenbazaar_unsupported_model error,
// exactly like the 4 image models and the 1 embeddings model do for this
// operation.
var tokenBazaarResponsesModels = map[string]struct{}{
	"gpt-5.5":           {},
	"gpt-5.6-luna":      {},
	"gpt-5.6-sol":       {},
	"gpt-5.6-terra":     {},
	"deepseek-v4-flash": {},
	"deepseek-v4-pro":   {},
}

// tokenBazaarMessagesModels is the exhaustive set of TokenBazaar model IDs
// this delivery dispatches through the documented POST /v1/messages
// (Anthropic Messages-compatible) envelope. Per this
// delivery's own executable prompt, this is scoped to exactly the 4 Claude
// models TokenBazaar carries -- create-only, mirroring how the Responses
// delivery was scoped to "create only" for its own 6 models.
var tokenBazaarMessagesModels = map[string]struct{}{
	"claude-fable-5":   {},
	"claude-haiku-4-5": {},
	"claude-opus-5":    {},
	"claude-sonnet-5":  {},
}

// tokenBazaarGenerateContentModels is the exhaustive set of TokenBazaar model
// IDs this delivery dispatches through the Gemini-native
// POST /v1beta/models/{model}:generateContent and
// POST /v1beta/models/{model}:streamGenerateContent envelopes.
//
// Evidence-tier caveat (see plans/tokenbazaar-production-readiness.md,
// "Native Gemini generateContent/streamGenerateContent delivery"): this is a
// WEAKER evidence tier than any prior TokenBazaar capability except
// embeddings -- unlike chat/images/responses/messages, TokenBazaar's own
// rendered docs page never mentions a Gemini-native API style anywhere in
// its "Text Generation (LLM)" prose (which explicitly names exactly three
// styles: Chat Completions, Responses, Anthropic Messages), and unlike
// embeddings, there is not even a product-description sentence claiming the
// capability -- the ONLY evidence this operation exists at all is the
// authenticated GET /v1/models catalog's supported_endpoint_types column
// listing "generateContent" for these models. This delivery implements
// against the standard Gemini-native shape gemini.go already builds for the
// direct google-gemini provider (reusing geminiOperationTarget for the exact
// same /v1beta/models/{model}:action path template) and adds a defensive
// response-shape check (validateTokenBazaarGenerateContentResponseShape)
// before returning success, mirroring the embeddings delivery's own
// discipline for an inferred contract.
//
// This is deliberately scoped to the 6 text models only. The 3 Gemini image
// models (gemini-3-pro-image, gemini-3.1-flash-image,
// gemini-3.1-flash-lite-image) are NOT included here: they are already
// served via the OpenAI-compatible images envelope
// (tokenBazaarImageModels/validateTokenBazaarImageDispatch), and native
// Gemini generateContent for image *generation* (inlineData response parts,
// generationConfig.responseModalities, etc.) is a materially different
// request/response shape than the text models' shape this delivery
// implements -- extending this delivery to the 3 image models would require
// its own, separately evidenced and tested contract, not an assumption that
// "same operation name" means "same shape." See the readiness doc's release
// gates for this explicit follow-up gap.
var tokenBazaarGenerateContentModels = map[string]struct{}{
	"gemini-2.5-flash":       {},
	"gemini-2.5-pro":         {},
	"gemini-3.1-pro-preview": {},
	"gemini-3.5-flash":       {},
	"gemini-3.6-flash":       {},
	"gemini-3.7-flash":       {},
}

// tokenBazaarImageParameterSnapshot records a per-model image size/quality
// support set confirmed live from an authenticated GET /v1/models response
// (TokenBazaar's documented "supported_parameters" field — see conflict #12
// in plans/tokenbazaar-production-readiness.md). This delivery never
// hardcodes a static size/quality table: a snapshot that is absent or
// expired means "unconfirmed", not "unsupported" and not "assumed
// supported" — see validateTokenBazaarImageParameters.
type tokenBazaarImageParameterSnapshot struct {
	// SupportedSizes and SupportedQuality are the exact values TokenBazaar's
	// live /v1/models response reported this model supports.
	SupportedSizes   map[string]struct{}
	SupportedQuality map[string]struct{}
	// SourceURL is the exact endpoint this snapshot was read from
	// (provenance), e.g. "https://zx1.deepwl.net/v1/models".
	SourceURL string
	// FetchedAt is when this snapshot was captured; it expires after
	// tokenBazaarImageParameterCacheTTL and is then treated as absent.
	FetchedAt time.Time
}

// tokenBazaarImageParameterCacheTTL bounds how long a live-confirmed
// per-model parameter snapshot is trusted before dispatch must treat it as
// unconfirmed again. TokenBazaar's own docs describe pricing/parameters as
// potentially account/tier-specific and not a fixed contract (conflict #6),
// so this cache is deliberately short-lived rather than a permanent
// hardcoded table.
const tokenBazaarImageParameterCacheTTL = 6 * time.Hour

var (
	tokenBazaarImageParameterCacheMu sync.RWMutex
	tokenBazaarImageParameterCache   = map[string]tokenBazaarImageParameterSnapshot{}
)

// setTokenBazaarImageParameterSnapshot installs a live-confirmed snapshot for
// a model, stamping the capture time. It is the only way this cache is ever
// populated — no code path pre-seeds it with guessed data. In production
// this is intended to be called by the credential-test/discovery job
// described in plans/tokenbazaar-production-readiness.md conflict #12 (not
// yet wired to a live account); tests call it directly to simulate a
// confirmed discovery result.
func setTokenBazaarImageParameterSnapshot(model string, snapshot tokenBazaarImageParameterSnapshot) {
	snapshot.FetchedAt = time.Now()

	tokenBazaarImageParameterCacheMu.Lock()
	defer tokenBazaarImageParameterCacheMu.Unlock()

	tokenBazaarImageParameterCache[model] = snapshot
}

// clearTokenBazaarImageParameterCache resets the cache. Test-only.
func clearTokenBazaarImageParameterCache() {
	tokenBazaarImageParameterCacheMu.Lock()
	defer tokenBazaarImageParameterCacheMu.Unlock()

	tokenBazaarImageParameterCache = map[string]tokenBazaarImageParameterSnapshot{}
}

func tokenBazaarImageParameterSnapshotFor(model string) (tokenBazaarImageParameterSnapshot, bool) {
	tokenBazaarImageParameterCacheMu.RLock()
	defer tokenBazaarImageParameterCacheMu.RUnlock()

	snapshot, ok := tokenBazaarImageParameterCache[model]
	if !ok || time.Since(snapshot.FetchedAt) > tokenBazaarImageParameterCacheTTL {
		return tokenBazaarImageParameterSnapshot{}, false
	}

	return snapshot, true
}

type tokenBazaarAdapter struct {
	compat *openAICompatibleAdapter
}

func newTokenBazaarAdapter() providerAdapter {
	return &tokenBazaarAdapter{compat: &openAICompatibleAdapter{
		defaultBaseURL: tokenBazaarBaseURL,
		authHeader:     "Authorization",
		// TokenBazaar's documented chat envelope lives at
		// /v1/chat/completions, but the shared OpenAI-compatible operation
		// table resolves platform.ProviderOperationChatCompletions to the
		// path suffix "/chat/completions" (most providers configure their
		// base URL to already include "/v1"). Appending "/v1" here, after
		// api_base validation/defaulting, keeps validateTokenBazaarBaseURL's
		// allowlist scoped to exactly the host+empty-path shape while still
		// producing the documented full URL.
		baseURLNormalizer: func(target string) string { return target + "/v1" },
	}}
}

func (a *tokenBazaarAdapter) Dispatch(
	ctx context.Context,
	client *http.Client,
	payload platform.TunnelRequest,
) (*adapterResponse, error) {
	if err := validateTokenBazaarDispatch(payload); err != nil {
		return nil, err
	}

	// generateContent/streamGenerateContent speak Gemini's own native wire
	// shape (POST /v1beta/models/{model}:action), not the OpenAI-compatible
	// shape every other TokenBazaar operation uses. a.compat.Dispatch cannot
	// produce this: its baseURLNormalizer unconditionally appends "/v1" to
	// the base URL (correct for /v1/chat/completions et al., but it would
	// wrongly produce .../v1/v1beta/models/... here), and
	// resolveOpenAIOperation has no case that embeds a model id into a
	// path. dispatchGenerateContent below builds and sends this request
	// directly instead, reusing gemini.go's own geminiOperationTarget path
	// template (the exact same one the direct google-gemini provider uses)
	// since the two providers speak byte-identical wire shapes here.
	if payload.Operation == platform.ProviderOperationGeminiGenerateContent ||
		payload.Operation == platform.ProviderOperationGeminiStreamGenerateContent {
		return a.dispatchGenerateContent(ctx, client, payload)
	}

	// Single attempt, no adapter-level retry: this is a POST to a
	// non-idempotent completions endpoint, and there is no way for this
	// adapter to prove from a client.Do error alone that the upstream never
	// received/processed the request body. Retrying here could double-bill
	// or double-execute a tool call. A caller-level retry is only safe when
	// it can independently prove no request reached TokenBazaar (e.g. a
	// pre-dispatch DNS/connect failure surfaced before client.Do was ever
	// called), which this adapter does not attempt to paper over.
	response, err := a.compat.Dispatch(ctx, client, payload)
	if err != nil {
		return nil, err
	}

	response, err = normalizeTokenBazaarResponse(response)
	if err != nil {
		return nil, err
	}

	if payload.Operation == platform.ProviderOperationOpenAIEmbeddings {
		// See the evidence-tier caveat on tokenBazaarEmbeddingModels: this
		// contract is inferred, not first-party-confirmed, so a successful
		// response is shape-checked here rather than forwarded blind.
		return validateTokenBazaarEmbeddingsResponseShape(response)
	}

	return response, nil
}

// dispatchGenerateContent builds and sends the Gemini-native
// generateContent/streamGenerateContent request directly, bypassing
// a.compat.Dispatch's "/v1"-suffixed base-URL normalization and
// OpenAI-compatible resolveOpenAIOperation path resolution -- neither
// produces the documented /v1beta/models/{model}:action shape TokenBazaar's
// Gemini-style endpoints use (see the readiness doc's "Auth (compatibility
// aliases)" bullet, which already documents /v1beta/... as a sibling path
// family to /v1/..., not something nested under it). It reuses gemini.go's
// geminiOperationTarget -- the exact same path-template logic the direct
// google-gemini adapter uses for this operation -- since the two providers
// speak byte-identical wire shapes here; only the base URL, host allowlist,
// and error-normalization differ, and those are TokenBazaar's own (already
// validated by validateTokenBazaarDispatch/validateTokenBazaarBaseURL above).
func (a *tokenBazaarAdapter) dispatchGenerateContent(ctx context.Context, client *http.Client, payload platform.TunnelRequest) (*adapterResponse, error) {
	model := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(payload.Deployment.Model)), "models/")

	method, path, err := geminiOperationTarget(payload.Operation, payload.OperationInput, model)
	if err != nil {
		return nil, err
	}

	target, err := resolveBaseURL(payload.Deployment.APIBase, tokenBazaarBaseURL)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, method, joinURL(target, path), bytes.NewReader(payload.Body))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")

	if payload.Operation == platform.ProviderOperationGeminiStreamGenerateContent {
		request.Header.Set("Accept", "text/event-stream")
	}

	// Only the universally-documented Authorization: Bearer scheme is used
	// here -- per the readiness doc's own guidance, the endpoint-scoped
	// x-goog-api-key compatibility alias is deliberately NOT adopted as a
	// generic TokenBazaar auth path, consistent with how the Messages
	// delivery declined to adopt the x-api-key alias for the same reason.
	if apiKey := deploymentAPIKey(payload.Deployment); apiKey != "" {
		request.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := client.Do(request) //nolint:bodyclose // The server handler owns and closes the body transferred through adapterResponse.
	if err != nil {
		return nil, err
	}

	response, err := normalizeTokenBazaarResponse(&adapterResponse{StatusCode: resp.StatusCode, Headers: resp.Header, Body: resp.Body})
	if err != nil {
		return nil, err
	}

	if payload.Operation == platform.ProviderOperationGeminiGenerateContent {
		// See the evidence-tier caveat on tokenBazaarGenerateContentModels:
		// this contract is inferred from catalog metadata alone, so a
		// successful, non-streaming response is shape-checked here rather
		// than forwarded blind. streamGenerateContent is deliberately
		// excluded from this check: its body is an SSE stream consumed
		// chunk-by-chunk by the control plane's own relay
		// (streamGeminiShapedResponse), not a single JSON document this
		// function could safely buffer and re-wrap.
		return validateTokenBazaarGenerateContentResponseShape(response)
	}

	return response, nil
}

func validateTokenBazaarDispatch(payload platform.TunnelRequest) error {
	if err := validateTokenBazaarBaseURL(payload.Deployment.APIBase); err != nil {
		return err
	}

	body, isJSON := tokenBazaarParseJSONBody(payload)

	model := tokenBazaarRequestModel(payload, body)

	if reason, gated := tokenBazaarUnsupportedModels[model]; gated {
		return fmt.Errorf("tokenbazaar_model_not_yet_supported: %s: %s", model, reason)
	}

	switch payload.Operation {
	case platform.ProviderOperationChatCompletions:
		if !isJSON {
			return fmt.Errorf("tokenbazaar_invalid_json")
		}

		return validateTokenBazaarChatDispatch(payload, model, body)
	case platform.ProviderOperationOpenAIImagesGenerations, platform.ProviderOperationOpenAIImagesEdits:
		return validateTokenBazaarImageDispatch(payload, model, body, isJSON)
	case platform.ProviderOperationOpenAIEmbeddings:
		return validateTokenBazaarEmbeddingsDispatch(payload, model, body, isJSON)
	case platform.ProviderOperationOpenAIResponsesCreate:
		return validateTokenBazaarResponsesDispatch(payload, model, body, isJSON)
	case platform.ProviderOperationAnthropicMessages:
		return validateTokenBazaarMessagesDispatch(payload, model, body, isJSON)
	case platform.ProviderOperationGeminiGenerateContent, platform.ProviderOperationGeminiStreamGenerateContent:
		return validateTokenBazaarGenerateContentDispatch(payload, model, body, isJSON)
	default:
		return fmt.Errorf("tokenbazaar_unsupported_operation: %s", payload.Operation)
	}
}

// tokenBazaarParseJSONBody attempts to decode the request body as a JSON
// object. Chat completions are always JSON; images/edits may instead be
// multipart/form-data (an uploaded source image), in which case this returns
// ok=false and callers fall back to model-only validation from
// payload.Deployment.Model, since this delivery does not parse multipart
// fields (see validateTokenBazaarImageDispatch).
func tokenBazaarParseJSONBody(payload platform.TunnelRequest) (map[string]json.RawMessage, bool) {
	var body map[string]json.RawMessage
	if err := json.Unmarshal(payload.Body, &body); err != nil || body == nil {
		return nil, false
	}

	return body, true
}

// tokenBazaarRequireRuntimeProfile enforces the runtime-profile and
// transport gating shared by every TokenBazaar dispatch validator: model
// must resolve to a validated, TokenBazaar-owned runtime profile that allows
// the given endpoint, and that profile must allow the given transport.
func tokenBazaarRequireRuntimeProfile(
	payload platform.TunnelRequest, model, endpoint, transport string,
) (*platform.ModelRuntimeProfileBinding, error) {
	binding, ok := platform.FindModelRuntimeProfile(payload.ModelRuntimeProfiles, model)
	if !ok || binding.Validate() != nil ||
		!strings.EqualFold(binding.Provider, tokenBazaarProviderName) ||
		!binding.Profile.AllowsEndpoint(endpoint) {
		return nil, fmt.Errorf("tokenbazaar_runtime_profile_missing: %s", model)
	}

	if !binding.Profile.AllowsTransport(transport) {
		return nil, fmt.Errorf("tokenbazaar_unsupported_transport: %s", transport)
	}

	return binding, nil
}

func validateTokenBazaarChatDispatch(payload platform.TunnelRequest, model string, body map[string]json.RawMessage) error {
	if _, supported := tokenBazaarChatModels[model]; !supported {
		return fmt.Errorf("tokenbazaar_unsupported_model: %s", model)
	}

	transport := "http"
	if payload.Stream {
		transport = "sse"
	}

	if _, err := tokenBazaarRequireRuntimeProfile(payload, model, "chat_completions", transport); err != nil {
		return err
	}

	if _, ok := body["messages"]; !ok {
		return fmt.Errorf("tokenbazaar_messages_required")
	}

	return nil
}

// validateTokenBazaarImageDispatch validates a POST /v1/images/generations
// or POST /v1/images/edits request against the documented contract (see
// plans/tokenbazaar-production-readiness.md, "Confirmed wire contract" and
// conflict-register item #12) before any HTTP call is attempted. It:
//  1. confirms the model is one of the 4 in-scope image models and its
//     runtime profile allows the requested image endpoint/transport;
//  2. enforces gpt-image-2's base64-only response_format constraint; and
//  3. rejects any requested size/quality that has not been confirmed live
//     from TokenBazaar's own authenticated GET /v1/models
//     "supported_parameters" (never a hardcoded per-model enum — see
//     tokenBazaarImageParameterSnapshotFor).
func validateTokenBazaarImageDispatch(payload platform.TunnelRequest, model string, body map[string]json.RawMessage, isJSON bool) error {
	if _, supported := tokenBazaarImageModels[model]; !supported {
		return fmt.Errorf("tokenbazaar_unsupported_model: %s", model)
	}

	endpoint := "image_generations"
	if payload.Operation == platform.ProviderOperationOpenAIImagesEdits {
		endpoint = "image_edits"
	}

	if _, err := tokenBazaarRequireRuntimeProfile(payload, model, endpoint, "http"); err != nil {
		return err
	}

	if !isJSON {
		// multipart/form-data (e.g. an images/edits request carrying an
		// uploaded source image) is forwarded as-is. This delivery does not
		// parse multipart fields, so the response_format/size/quality checks
		// below apply only to JSON-encoded requests; see
		// plans/tokenbazaar-production-readiness.md for this follow-up.
		return nil
	}

	if err := validateTokenBazaarGPTImage2ResponseFormat(model, body); err != nil {
		return err
	}

	return validateTokenBazaarImageParameters(model, body)
}

// validateTokenBazaarGPTImage2ResponseFormat enforces the documented
// constraint that gpt-image-2 via TokenBazaar returns base64 only:
// response_format must be "b64_json" or omitted. A caller-requested "url" (or
// any other value) is rejected here, before dispatch — never silently
// converted to b64_json and never silently dropped.
func validateTokenBazaarGPTImage2ResponseFormat(model string, body map[string]json.RawMessage) error {
	if model != "gpt-image-2" {
		return nil
	}

	raw, ok := body["response_format"]
	if !ok {
		return nil
	}

	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return fmt.Errorf("tokenbazaar_invalid_json: response_format must be a string")
	}

	if !strings.EqualFold(strings.TrimSpace(value), "b64_json") {
		return fmt.Errorf(
			"tokenbazaar_gpt_image_2_response_format_unsupported: gpt-image-2 via TokenBazaar returns base64 only; "+
				"response_format must be \"b64_json\" or omitted, got %q",
			value,
		)
	}

	return nil
}

// validateTokenBazaarImageParameters implements conflict-register item #12's
// safe-default: a requested size or quality is checked against the
// live-confirmed tokenBazaarImageParameterSnapshotFor cache for this model.
// If the caller did not request a size or quality at all, this delivery lets
// TokenBazaar apply its own model default rather than guessing one — no
// value is invented on GodarAI's side. If the caller DID request a size or
// quality and no live-confirmed snapshot exists for this model (the common
// case until a real credential populates the cache — see
// plans/tokenbazaar-production-readiness.md conflict #12), the request is
// rejected with a distinct, identifiable error rather than passed through
// silently or rejected with the generic out-of-scope error.
func validateTokenBazaarImageParameters(model string, body map[string]json.RawMessage) error {
	size := tokenBazaarStringField(body, "size")
	quality := tokenBazaarStringField(body, "quality")

	if size == "" && quality == "" {
		return nil
	}

	snapshot, ok := tokenBazaarImageParameterSnapshotFor(model)
	if !ok {
		return fmt.Errorf(
			"tokenbazaar_image_parameters_unconfirmed: %s: live supported_parameters for this model have not been "+
				"confirmed from an authenticated GET /v1/models response; omit size/quality to use TokenBazaar's own "+
				"default, or wait for a live credential to populate the confirmation cache before requesting size=%q quality=%q",
			model, size, quality,
		)
	}

	if size != "" {
		if _, allowed := snapshot.SupportedSizes[size]; !allowed {
			return fmt.Errorf(
				"tokenbazaar_image_parameter_not_supported: %s: size %q is not in the live-confirmed supported set",
				model, size,
			)
		}
	}

	if quality != "" {
		if _, allowed := snapshot.SupportedQuality[quality]; !allowed {
			return fmt.Errorf(
				"tokenbazaar_image_parameter_not_supported: %s: quality %q is not in the live-confirmed supported set",
				model, quality,
			)
		}
	}

	return nil
}

func tokenBazaarStringField(body map[string]json.RawMessage, key string) string {
	raw, ok := body[key]
	if !ok {
		return ""
	}

	var value string
	if json.Unmarshal(raw, &value) != nil {
		return ""
	}

	return strings.TrimSpace(value)
}

// validateTokenBazaarEmbeddingsDispatch validates a POST /v1/embeddings
// request against the *inferred* TokenBazaar embeddings contract (see the
// evidence-tier caveat on tokenBazaarEmbeddingModels) before any HTTP call is
// attempted. Because this contract has no first-party worked example (unlike
// chat and images), this function is deliberately conservative: it rejects
// anything it cannot validate against the existing, already-vetted
// direct-provider catalog bounds, rather than passing an unrecognized shape
// through silently.
func validateTokenBazaarEmbeddingsDispatch(payload platform.TunnelRequest, model string, body map[string]json.RawMessage, isJSON bool) error {
	if _, supported := tokenBazaarEmbeddingModels[model]; !supported {
		return fmt.Errorf("tokenbazaar_unsupported_model: %s", model)
	}

	if _, err := tokenBazaarRequireRuntimeProfile(payload, model, "embeddings", "http"); err != nil {
		return err
	}

	// The standard OpenAI-compatible embeddings contract is always a JSON
	// object (never multipart/form-data, unlike images/edits). A non-JSON
	// body is rejected outright rather than forwarded blind.
	if !isJSON {
		return fmt.Errorf("tokenbazaar_invalid_json")
	}

	if _, ok := body["input"]; !ok {
		return fmt.Errorf("tokenbazaar_input_required")
	}

	if err := validateTokenBazaarEmbeddingsEncodingFormat(body); err != nil {
		return err
	}

	return validateTokenBazaarEmbeddingsDimensions(body)
}

// validateTokenBazaarEmbeddingsEncodingFormat rejects any encoding_format
// other than "float"/"base64" (case-insensitively) before dispatch. Omitting
// the field is allowed (TokenBazaar's own default, inferred to be "float" per
// the standard OpenAI-compatible contract, applies).
func validateTokenBazaarEmbeddingsEncodingFormat(body map[string]json.RawMessage) error {
	raw, ok := body["encoding_format"]
	if !ok {
		return nil
	}

	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return fmt.Errorf("tokenbazaar_invalid_json: encoding_format must be a string")
	}

	value = strings.ToLower(strings.TrimSpace(value))
	if _, allowed := tokenBazaarEmbeddingsEncodingFormats[value]; !allowed {
		return fmt.Errorf(
			"tokenbazaar_embeddings_encoding_format_unsupported: encoding_format must be \"float\" or \"base64\", got %q",
			value,
		)
	}

	return nil
}

// validateTokenBazaarEmbeddingsDimensions enforces the
// tokenBazaarEmbeddingsMinDimensions/tokenBazaarEmbeddingsMaxDimensions bound
// on an optional "dimensions" request field before dispatch. Omitting the
// field is allowed (TokenBazaar's own model default applies).
func validateTokenBazaarEmbeddingsDimensions(body map[string]json.RawMessage) error {
	raw, ok := body["dimensions"]
	if !ok {
		return nil
	}

	var value float64
	if err := json.Unmarshal(raw, &value); err != nil || value != float64(int64(value)) {
		return fmt.Errorf("tokenbazaar_invalid_json: dimensions must be an integer")
	}

	dimensions := int64(value)
	if dimensions < tokenBazaarEmbeddingsMinDimensions || dimensions > tokenBazaarEmbeddingsMaxDimensions {
		return fmt.Errorf(
			"tokenbazaar_embeddings_dimensions_out_of_range: dimensions must be between %d and %d, got %d",
			tokenBazaarEmbeddingsMinDimensions, tokenBazaarEmbeddingsMaxDimensions, dimensions,
		)
	}

	return nil
}

// validateTokenBazaarResponsesDispatch validates a POST /v1/responses request
// against the documented, first-party-confirmed contract (see
// plans/tokenbazaar-production-readiness.md, "Confirmed wire contract") before
// any HTTP call is attempted. TokenBazaar's own docs state plainly that this
// route is stateless-only: "previous_response_id, built-in tools, and
// background: true are not supported." This function fails closed on all
// three, each with its own distinguishable error, rather than letting the
// generic control-plane Responses handler's continuation machinery
// (resolveResponseCreateDeployment / rewriteOpenAIJSONStringField) silently
// forward them -- that machinery only tracks response-resource mappings for
// "openai"/"xai" (see responseResourceMapping in
// services/control-plane/internal/proxy/responses_operations_handler.go), so
// it would never populate or rewrite previous_response_id for TokenBazaar; a
// caller-supplied value would otherwise reach TokenBazaar completely
// unvalidated were it not rejected here.
func validateTokenBazaarResponsesDispatch(payload platform.TunnelRequest, model string, body map[string]json.RawMessage, isJSON bool) error {
	if _, supported := tokenBazaarResponsesModels[model]; !supported {
		return fmt.Errorf("tokenbazaar_unsupported_model: %s", model)
	}

	transport := "http"
	if payload.Stream {
		transport = "sse"
	}

	if _, err := tokenBazaarRequireRuntimeProfile(payload, model, "responses", transport); err != nil {
		return err
	}

	// The documented /v1/responses contract is always a JSON object (never
	// multipart/form-data, unlike images/edits).
	if !isJSON {
		return fmt.Errorf("tokenbazaar_invalid_json")
	}

	if err := validateTokenBazaarResponsesPreviousResponseID(body); err != nil {
		return err
	}

	if err := validateTokenBazaarResponsesBackground(body); err != nil {
		return err
	}

	if err := validateTokenBazaarResponsesTools(body); err != nil {
		return err
	}

	return validateTokenBazaarResponsesInputPresence(body)
}

// validateTokenBazaarResponsesPreviousResponseID rejects a non-empty
// previous_response_id before dispatch: TokenBazaar's /v1/responses is
// documented as stateless ("Send the full conversation in input with every
// request"), so continuation against a previous response is never valid for
// this provider. A null, absent, or empty-string value is allowed through
// unchanged (the common shape for a fresh, non-continuation request).
func validateTokenBazaarResponsesPreviousResponseID(body map[string]json.RawMessage) error {
	raw, ok := body["previous_response_id"]
	if !ok || tokenBazaarJSONIsNull(raw) {
		return nil
	}

	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return fmt.Errorf("tokenbazaar_invalid_json: previous_response_id must be a string")
	}

	if strings.TrimSpace(value) == "" {
		return nil
	}

	return fmt.Errorf(
		"tokenbazaar_responses_previous_response_id_unsupported: TokenBazaar's /v1/responses is stateless; " +
			"send the full conversation in input with every request instead of previous_response_id",
	)
}

// validateTokenBazaarResponsesBackground rejects background: true before
// dispatch, per TokenBazaar's own documented "not supported" list. An absent
// field or an explicit background: false both dispatch normally.
func validateTokenBazaarResponsesBackground(body map[string]json.RawMessage) error {
	raw, ok := body["background"]
	if !ok || tokenBazaarJSONIsNull(raw) {
		return nil
	}

	var value bool
	if err := json.Unmarshal(raw, &value); err != nil {
		return fmt.Errorf("tokenbazaar_invalid_json: background must be a boolean")
	}

	if value {
		return fmt.Errorf("tokenbazaar_responses_background_unsupported: TokenBazaar's /v1/responses does not support background: true")
	}

	return nil
}

// validateTokenBazaarResponsesTools rejects any tool in the request's "tools"
// array whose type is not "function" -- i.e. any OpenAI-hosted built-in tool
// (web_search, file_search, image_generation, code_interpreter, and any other
// hosted tool type, present or future) is rejected fail-closed. TokenBazaar's
// docs confirm caller-defined function tools are fine on this route; only the
// built-in/hosted tool types are documented as unsupported.
func validateTokenBazaarResponsesTools(body map[string]json.RawMessage) error {
	raw, ok := body["tools"]
	if !ok || tokenBazaarJSONIsNull(raw) {
		return nil
	}

	var tools []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &tools); err != nil {
		return fmt.Errorf("tokenbazaar_invalid_json: tools must be an array")
	}

	for _, tool := range tools {
		toolType := tokenBazaarStringField(tool, "type")
		if toolType != "" && !strings.EqualFold(toolType, "function") {
			return fmt.Errorf(
				"tokenbazaar_responses_builtin_tool_unsupported: tool type %q is an OpenAI-hosted built-in tool; "+
					"TokenBazaar's /v1/responses supports only caller-defined function tools",
				toolType,
			)
		}
	}

	return nil
}

// validateTokenBazaarResponsesInputPresence requires at least one of "input"
// or "instructions" to be present and non-null. Since previous_response_id
// continuation is always rejected above, a TokenBazaar /v1/responses request
// must always carry its own conversation content.
func validateTokenBazaarResponsesInputPresence(body map[string]json.RawMessage) error {
	input, hasInput := body["input"]

	instructions, hasInstructions := body["instructions"]
	if (!hasInput || tokenBazaarJSONIsNull(input)) && (!hasInstructions || tokenBazaarJSONIsNull(instructions)) {
		return fmt.Errorf("tokenbazaar_responses_input_required: input or instructions is required")
	}

	return nil
}

// tokenBazaarJSONIsNull reports whether a raw JSON value is the literal null.
func tokenBazaarJSONIsNull(raw json.RawMessage) bool {
	return strings.EqualFold(strings.TrimSpace(string(raw)), "null")
}

// validateTokenBazaarMessagesDispatch validates a POST /v1/messages request
// against TokenBazaar's own documented contract (see
// plans/tokenbazaar-production-readiness.md, "Confirmed wire contract" and
// "Anthropic Messages endpoint delivery") before any HTTP call is attempted.
// Unlike chat/responses/embeddings, this route's request body is already
// Anthropic-native JSON (model, messages, max_tokens, system, stream,
// thinking) -- not an OpenAI-compatible shape -- so this function validates
// that native shape directly rather than an OpenAI-compatible one.
func validateTokenBazaarMessagesDispatch(payload platform.TunnelRequest, model string, body map[string]json.RawMessage, isJSON bool) error {
	if _, supported := tokenBazaarMessagesModels[model]; !supported {
		return fmt.Errorf("tokenbazaar_unsupported_model: %s", model)
	}

	transport := "http"
	if payload.Stream {
		transport = "sse"
	}

	if _, err := tokenBazaarRequireRuntimeProfile(payload, model, "messages", transport); err != nil {
		return err
	}

	// TokenBazaar's own documented /v1/messages contract is always a JSON
	// object (never multipart/form-data).
	if !isJSON {
		return fmt.Errorf("tokenbazaar_invalid_json")
	}

	if err := validateTokenBazaarMessagesMaxTokens(body); err != nil {
		return err
	}

	if _, ok := body["messages"]; !ok {
		return fmt.Errorf("tokenbazaar_messages_required")
	}

	if err := validateTokenBazaarMessagesToolsUnsupported(body); err != nil {
		return err
	}

	// "thinking" (extended-thinking / budget_tokens) is deliberately NOT
	// validated or stripped anywhere in this function: TokenBazaar's own docs
	// document thinking: {type: "enabled", budget_tokens: N} as a supported
	// field on this exact route, and the request body is forwarded to
	// TokenBazaar byte-for-byte (aside from the model-id rewrite already
	// applied by proxy.TokenBazaarMessagesHandler before this adapter ever
	// sees the body), so "thinking" passes through unmodified by
	// construction -- see
	// TestTokenBazaarMessagesThinkingPassesThroughUnmodified.
	return validateTokenBazaarMessagesPromptCachingUnsupported(body)
}

// validateTokenBazaarMessagesMaxTokens requires max_tokens to be present and
// a positive integer before dispatch, per TokenBazaar's own documented
// contract ("requires max_tokens", matching Claude's own requirement) and
// this delivery's own instruction not to rely on upstream to reject its
// absence.
func validateTokenBazaarMessagesMaxTokens(body map[string]json.RawMessage) error {
	raw, ok := body["max_tokens"]
	if !ok || tokenBazaarJSONIsNull(raw) {
		return fmt.Errorf("tokenbazaar_messages_max_tokens_required: max_tokens is required for TokenBazaar's /v1/messages")
	}

	var value float64
	if err := json.Unmarshal(raw, &value); err != nil || value != float64(int64(value)) {
		return fmt.Errorf("tokenbazaar_invalid_json: max_tokens must be an integer")
	}

	if value <= 0 {
		return fmt.Errorf("tokenbazaar_messages_max_tokens_required: max_tokens must be a positive integer")
	}

	return nil
}

// validateTokenBazaarMessagesToolsUnsupported rejects any request carrying a
// non-empty "tools" array before dispatch. TokenBazaar's documented
// /v1/messages example (unlike its /v1/chat/completions section, which
// explicitly documents "tools") does not show a tools parameter for this
// specific route -- per this delivery's own fail-closed discipline, an
// unconfirmed field is rejected with a specific, actionable error rather than
// silently forwarded and left to fail (or silently succeed in some
// unconfirmed, unreviewed way) upstream.
func validateTokenBazaarMessagesToolsUnsupported(body map[string]json.RawMessage) error {
	raw, ok := body["tools"]
	if !ok || tokenBazaarJSONIsNull(raw) {
		return nil
	}

	var tools []json.RawMessage
	if err := json.Unmarshal(raw, &tools); err != nil {
		return fmt.Errorf("tokenbazaar_invalid_json: tools must be an array")
	}

	if len(tools) == 0 {
		return nil
	}

	return fmt.Errorf(
		"tokenbazaar_messages_tools_unsupported: TokenBazaar's documented /v1/messages contract does not confirm " +
			"tool-use support on this route (unlike /v1/chat/completions, which documents \"tools\" explicitly); " +
			"omit tools until this is confirmed, or use /v1/chat/completions if tool-use is required",
	)
}

// validateTokenBazaarMessagesPromptCachingUnsupported rejects any
// "cache_control" block before dispatch, wherever the Anthropic Messages
// schema allows one to appear on a request that could otherwise reach
// dispatch: a system block or a message content block. (A tool definition is
// the schema's third location for cache_control, but
// validateTokenBazaarMessagesToolsUnsupported above already rejects any
// non-empty "tools" array outright -- since tool-use itself is unconfirmed on
// this route -- so a cache_control-on-a-tool request can never reach this
// function with a non-empty tools array; scanning tools here would be
// unreachable dead code.) Prompt caching is not shown in TokenBazaar's
// documented /v1/messages example, so -- per the same fail-closed discipline
// as tools -- this delivery rejects it rather than silently forwarding an
// unconfirmed field and letting a cache_control block be silently ignored (or
// worse, billed unexpectedly) by TokenBazaar's own gateway.
func validateTokenBazaarMessagesPromptCachingUnsupported(body map[string]json.RawMessage) error {
	if tokenBazaarBlocksContainCacheControl(body["system"]) {
		return tokenBazaarPromptCachingUnsupportedError("system")
	}

	raw, ok := body["messages"]
	if !ok || tokenBazaarJSONIsNull(raw) {
		return nil
	}

	var messages []struct {
		Content json.RawMessage `json:"content"`
	}
	if err := json.Unmarshal(raw, &messages); err != nil {
		// Not an array of objects with a "content" field (e.g. malformed).
		// This function's only job is the cache_control fail-closed check;
		// deep "messages" shape validation is not otherwise performed by any
		// TokenBazaar dispatch path in this file (the required-field check
		// above only confirms "messages" is present, not its inner shape), so
		// a malformed "messages" value is deliberately left for the upstream
		// response to surface as its own error rather than guessed at here.
		return nil //nolint:nilerr // Intentional: see comment above.
	}

	for _, message := range messages {
		if tokenBazaarBlocksContainCacheControl(message.Content) {
			return tokenBazaarPromptCachingUnsupportedError("messages[].content")
		}
	}

	return nil
}

// tokenBazaarBlocksContainCacheControl reports whether a raw JSON value that
// may be an array of Anthropic content blocks contains a "cache_control" key
// on any block. A non-array value (e.g. a plain string system prompt, which
// the Anthropic schema also allows) never carries cache_control and returns
// false.
func tokenBazaarBlocksContainCacheControl(raw json.RawMessage) bool {
	if len(raw) == 0 || tokenBazaarJSONIsNull(raw) {
		return false
	}

	var blocks []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return false
	}

	for _, block := range blocks {
		if _, ok := block["cache_control"]; ok {
			return true
		}
	}

	return false
}

func tokenBazaarPromptCachingUnsupportedError(location string) error {
	return fmt.Errorf(
		"tokenbazaar_messages_prompt_caching_unsupported: a cache_control block was found in %s, but TokenBazaar's "+
			"documented /v1/messages contract does not confirm prompt-caching support on this route; omit "+
			"cache_control until this is confirmed",
		location,
	)
}

// tokenBazaarEmbeddingsResponseEnvelope is the minimal shape this delivery
// requires of a successful TokenBazaar embeddings response, per the
// evidence-tier caveat on tokenBazaarEmbeddingModels: standard OpenAI
// embeddings shape {object:"list", data:[{object:"embedding",
// embedding:[...], index}], model, usage:{...}}. Only the fields needed to
// confirm the shape are modeled.
type tokenBazaarEmbeddingsResponseEnvelope struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string          `json:"object"`
		Embedding json.RawMessage `json:"embedding"`
		Index     int             `json:"index"`
	} `json:"data"`
}

// validateTokenBazaarEmbeddingsResponseShape reads a successful (< 400)
// embeddings response body in full — embeddings requests are never streamed
// (Stream is always false for this operation) — and confirms it at least
// matches the inferred {object:"list", data:[{object:"embedding",
// embedding:[...], index}]} envelope before handing it back to the caller.
// This matters more here than for the chat/image paths precisely because the
// embeddings contract is inferred, not first-party-confirmed: a response
// that fails this shape check is rejected with a distinct, identifiable
// error instead of being forwarded as-is (which could otherwise be silently
// mis-parsed or mis-billed further up the stack) or causing a panic in a
// downstream json.Unmarshal into a fixed struct. The read respects whatever
// size ceiling normalizeTokenBazaarResponse already applied to the body.
func validateTokenBazaarEmbeddingsResponseShape(response *adapterResponse) (*adapterResponse, error) {
	if response == nil || response.StatusCode >= 400 {
		return response, nil
	}

	raw, readErr := io.ReadAll(response.Body)
	closeErr := response.Body.Close()

	if readErr != nil {
		return nil, fmt.Errorf("tokenbazaar_embeddings_response_read: %w (close: %v)", readErr, closeErr)
	}

	if closeErr != nil {
		return nil, fmt.Errorf("tokenbazaar_embeddings_response_close: %w", closeErr)
	}

	var envelope tokenBazaarEmbeddingsResponseEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, fmt.Errorf("tokenbazaar_embeddings_response_shape_unexpected: response body is not valid JSON: %w", err)
	}

	if !strings.EqualFold(envelope.Object, "list") || len(envelope.Data) == 0 {
		return nil, fmt.Errorf(
			"tokenbazaar_embeddings_response_shape_unexpected: expected {\"object\":\"list\",\"data\":[...]}, "+
				"got object=%q with %d data entries; this delivery's embeddings contract is inferred, not "+
				"first-party-confirmed (see plans/tokenbazaar-production-readiness.md), and this response does not "+
				"match the expected shape",
			envelope.Object, len(envelope.Data),
		)
	}

	for i, entry := range envelope.Data {
		if len(entry.Embedding) == 0 {
			return nil, fmt.Errorf(
				"tokenbazaar_embeddings_response_shape_unexpected: data[%d] is missing an embedding vector", i,
			)
		}
	}

	response.Body = io.NopCloser(bytes.NewReader(raw))

	return response, nil
}

// validateTokenBazaarGenerateContentDispatch validates a
// POST /v1beta/models/{model}:generateContent or
// .../{model}:streamGenerateContent request against the standard Gemini
// generateContent contract (see the evidence-tier caveat on
// tokenBazaarGenerateContentModels) before any HTTP call is attempted. This
// is deliberately conservative given the weak evidence tier: it requires the
// one field Gemini's own schema requires ("contents"), rejects the one
// undocumented feature explicitly called out for this delivery
// (cachedContent), and otherwise defers to the runtime-profile/transport
// gating every other TokenBazaar operation already goes through -- it does
// not attempt to validate "tools"/"systemInstruction"/generationConfig
// shapes, since gemini.go's own direct-dispatch path does not validate those
// either (the direct provider forwards the body as-is beyond
// model/operation-level access control), and there is no stronger evidence
// here to justify TokenBazaar going further than the direct provider does.
func validateTokenBazaarGenerateContentDispatch(payload platform.TunnelRequest, model string, body map[string]json.RawMessage, isJSON bool) error {
	if _, supported := tokenBazaarGenerateContentModels[model]; !supported {
		return fmt.Errorf("tokenbazaar_unsupported_model: %s", model)
	}

	endpoint := "generate_content"
	if payload.Operation == platform.ProviderOperationGeminiStreamGenerateContent {
		endpoint = "stream_generate_content"
	}

	transport := "http"
	if payload.Stream || payload.Operation == platform.ProviderOperationGeminiStreamGenerateContent {
		transport = "sse"
	}

	if _, err := tokenBazaarRequireRuntimeProfile(payload, model, endpoint, transport); err != nil {
		return err
	}

	// Gemini's own generateContent contract is always a JSON object (never
	// multipart/form-data).
	if !isJSON {
		return fmt.Errorf("tokenbazaar_invalid_json")
	}

	if _, ok := body["contents"]; !ok {
		return fmt.Errorf("tokenbazaar_generate_content_contents_required")
	}

	return validateTokenBazaarGenerateContentCachedContentUnsupported(body)
}

// validateTokenBazaarGenerateContentCachedContentUnsupported rejects a
// non-empty "cachedContent" field before dispatch. Gemini's own
// context-caching feature (a reference to a cachedContents/{id} resource
// created via the separate CachesCreate operation) is not claimed anywhere
// in TokenBazaar's documentation -- the rendered docs page's nav sidebar has
// no caching section at all, and this delivery's own evidence-gathering
// found no mention of generateContent itself beyond catalog metadata, let
// alone one of its optional sub-features. Per this delivery's own
// instruction and the fail-closed precedent set by
// validateTokenBazaarMessagesToolsUnsupported/
// validateTokenBazaarMessagesPromptCachingUnsupported above, an unconfirmed
// field is rejected with a specific, actionable error rather than silently
// forwarded (where TokenBazaar might silently ignore it, reject it with an
// opaque upstream error, or -- worse -- bill for context caching GodarAI
// never confirmed it actually performs).
func validateTokenBazaarGenerateContentCachedContentUnsupported(body map[string]json.RawMessage) error {
	raw, ok := body["cachedContent"]
	if !ok || tokenBazaarJSONIsNull(raw) {
		return nil
	}

	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return fmt.Errorf("tokenbazaar_invalid_json: cachedContent must be a string")
	}

	if strings.TrimSpace(value) == "" {
		return nil
	}

	return fmt.Errorf(
		"tokenbazaar_generate_content_cached_content_unsupported: TokenBazaar does not document support for Gemini's " +
			"context-caching feature (cachedContent); omit this field until support is confirmed",
	)
}

// tokenBazaarGenerateContentResponseEnvelope is the minimal shape this
// delivery requires of a successful, non-streaming TokenBazaar
// generateContent response, per the evidence-tier caveat on
// tokenBazaarGenerateContentModels: Gemini's own documented
// {"candidates":[{"content":{...}, ...}], "usageMetadata":{...}} envelope.
// Only the field needed to confirm the shape is modeled.
type tokenBazaarGenerateContentResponseEnvelope struct {
	Candidates []struct {
		Content json.RawMessage `json:"content"`
	} `json:"candidates"`
}

// validateTokenBazaarGenerateContentResponseShape reads a successful (< 400)
// non-streaming generateContent response body in full and confirms it at
// least matches the inferred {"candidates":[{"content":{...}}]} envelope
// before handing it back to the caller. This matters more here than for the
// chat/image paths precisely because this contract is inferred purely from
// catalog metadata, not first-party-confirmed: a response that fails this
// shape check is rejected with a distinct, identifiable error instead of
// being forwarded as-is (which could otherwise be silently mis-parsed or
// mis-billed further up the stack, e.g. by parseGeminiNativeUsage) or
// causing a panic in a downstream json.Unmarshal into a fixed struct. The
// read respects whatever size ceiling normalizeTokenBazaarResponse already
// applied to the body. Mirrors validateTokenBazaarEmbeddingsResponseShape's
// pattern exactly.
func validateTokenBazaarGenerateContentResponseShape(response *adapterResponse) (*adapterResponse, error) {
	if response == nil || response.StatusCode >= 400 {
		return response, nil
	}

	raw, readErr := io.ReadAll(response.Body)
	closeErr := response.Body.Close()

	if readErr != nil {
		return nil, fmt.Errorf("tokenbazaar_generate_content_response_read: %w (close: %v)", readErr, closeErr)
	}

	if closeErr != nil {
		return nil, fmt.Errorf("tokenbazaar_generate_content_response_close: %w", closeErr)
	}

	var envelope tokenBazaarGenerateContentResponseEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, fmt.Errorf("tokenbazaar_generate_content_response_shape_unexpected: response body is not valid JSON: %w", err)
	}

	if len(envelope.Candidates) == 0 {
		return nil, fmt.Errorf(
			"tokenbazaar_generate_content_response_shape_unexpected: expected {\"candidates\":[...]}, got %d candidates; "+
				"this delivery's generateContent contract is inferred from catalog metadata only, not first-party-confirmed "+
				"(see plans/tokenbazaar-production-readiness.md), and this response does not match the expected shape",
			len(envelope.Candidates),
		)
	}

	response.Body = io.NopCloser(bytes.NewReader(raw))

	return response, nil
}

func tokenBazaarRequestModel(payload platform.TunnelRequest, body map[string]json.RawMessage) string {
	model := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(payload.Deployment.Model)), "models/")

	var requestedModel string
	if json.Unmarshal(body["model"], &requestedModel) == nil && strings.TrimSpace(requestedModel) != "" {
		model = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(requestedModel)), "models/")
	}

	return model
}

// validateTokenBazaarBaseURL is the host allowlist described in the package
// comment above. An empty api_base is allowed (the adapter's default is
// used); anything else must be exactly https://zx1.deepwl.net with no path,
// query, fragment, credentials, or non-default port.
func validateTokenBazaarBaseURL(raw string) error {
	value := strings.TrimRight(strings.TrimSpace(raw), "/")
	if value == "" {
		return nil
	}

	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() != tokenBazaarHost ||
		parsed.Port() != "" || parsed.User != nil || parsed.RawQuery != "" ||
		parsed.Fragment != "" || (parsed.EscapedPath() != "" && parsed.EscapedPath() != "/") {
		return fmt.Errorf("tokenbazaar_invalid_api_base")
	}

	return nil
}

// normalizeTokenBazaarResponse implements the readiness doc's documented
// "attribution rule": an upstream 4xx is not passed through raw unless
// locally allow-listed. 429 is allow-listed as-is (with any Retry-After
// header intact) because it is documented as the only real rate-limit
// signal TokenBazaar exposes, and GodarAI's own tenant-plan limiting relies
// on seeing it. Every other 4xx is rewritten to a 503 model_unavailable
// error with a fixed, generic body — never the raw upstream body, which may
// contain upstream-vendor-specific details GodarAI has not vetted for
// leakage. X-Tokenbazaar-Request-Id is preserved on both paths so operators
// can correlate with TokenBazaar support, but it is never treated as (or
// logged as) a GodarAI-internal request ID. Every response body, on every
// path, is capped at tokenBazaarMaxResponseBytes.
func normalizeTokenBazaarResponse(response *adapterResponse) (*adapterResponse, error) {
	if response == nil {
		return nil, nil
	}

	if response.StatusCode < 400 || response.StatusCode >= 500 || response.StatusCode == http.StatusTooManyRequests {
		response.Body = newTokenBazaarSizeCappedBody(response.Body, tokenBazaarMaxResponseBytes)

		return response, nil
	}

	requestID := ""
	if response.Headers != nil {
		requestID = response.Headers.Get(tokenBazaarRequestIDHeader)
	}

	if _, err := io.Copy(io.Discard, io.LimitReader(response.Body, tokenBazaarMaxErrorBodyBytes)); err != nil {
		closeErr := response.Body.Close()

		return nil, fmt.Errorf("tokenbazaar_error_body_drain: %w (close: %v)", err, closeErr)
	}

	if err := response.Body.Close(); err != nil {
		return nil, fmt.Errorf("tokenbazaar_error_body_close: %w", err)
	}

	headers := http.Header{}
	headers.Set("Content-Type", "application/json")

	if requestID != "" {
		headers.Set(tokenBazaarRequestIDHeader, requestID)
	}

	normalizedBody := []byte(
		`{"error":{"code":"model_unavailable","message":"TokenBazaar upstream is currently unavailable for this model."}}`,
	)

	return &adapterResponse{
		StatusCode: http.StatusServiceUnavailable,
		Headers:    headers,
		Body:       io.NopCloser(bytes.NewReader(normalizedBody)),
	}, nil
}

// tokenBazaarSizeCappedBody wraps a response body so that reading past the
// configured ceiling fails closed with a clear error instead of silently
// truncating (which could otherwise produce a corrupted-but-parseable JSON
// or SSE fragment) or growing unbounded in memory.
type tokenBazaarSizeCappedBody struct {
	io.ReadCloser
	remaining int64
	exceeded  bool
}

func newTokenBazaarSizeCappedBody(body io.ReadCloser, ceiling int64) io.ReadCloser {
	return &tokenBazaarSizeCappedBody{ReadCloser: body, remaining: ceiling}
}

// Read allows exactly one byte past the ceiling through to the underlying
// reader so a body that lands precisely at the ceiling (remaining reaches 0
// with the stream legitimately exhausted) can still surface its trailing
// io.EOF instead of being mistaken for an oversized body; only a stream that
// actually produces that extra byte is treated as exceeding the ceiling.
func (b *tokenBazaarSizeCappedBody) Read(p []byte) (int, error) {
	if b.exceeded {
		return 0, fmt.Errorf("tokenbazaar_response_size_ceiling_exceeded")
	}

	limit := b.remaining + 1
	if int64(len(p)) > limit {
		p = p[:limit]
	}

	n, err := b.ReadCloser.Read(p)
	b.remaining -= int64(n)

	if b.remaining < 0 {
		b.exceeded = true
		return 0, fmt.Errorf("tokenbazaar_response_size_ceiling_exceeded")
	}

	return n, err
}
