package server

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	"github.com/gofiber/fiber/v2"
)

// ValidateRequired verifies the complete production HTTP surface before routes
// are registered. Partial dependency sets remain available only to focused
// tests through NewFiberApp.
func (d Dependencies) ValidateRequired() error {
	missing := make([]string, 0)

	require := func(name string, available bool) {
		if !available {
			missing = append(missing, name)
		}
	}

	require("auth middleware", d.AuthMiddleware != nil)
	require("tenant resolver", d.TenantResolver != nil)
	require("tracing middleware", d.TracingMiddleware != nil)
	require("trace health reporter", d.TraceHealth != nil)
	require("ClickHouse health reporter", d.ClickHouseHealth != nil)
	require("management routes", d.ManagementRoutes != nil)
	require("chat completions handler", d.ProxyHandler != nil)
	require("completions handler", d.CompletionsHandler != nil)
	require("responses handler", d.ResponsesHandler != nil)
	require("embeddings handler", d.EmbeddingsHandler != nil)
	require("moderations handler", d.ModerationsHandler != nil)
	require("rerank handler", d.RerankHandler != nil)
	require("images handler", d.ImagesHandler != nil)
	require("videos handler", d.VideosHandler != nil)
	require("audio transcriptions handler", d.AudioTranscriptionsHandler != nil)
	require("audio speech handler", d.AudioSpeechHandler != nil)
	require("Mistral OCR handler", d.MistralOCRHandler != nil)
	require("Z.ai layout parsing handler", d.ZAILayoutParsingHandler != nil)
	require("Mistral native handler", d.MistralNativeHandler != nil)
	require("models handler", d.ModelsHandler != nil)
	require("Anthropic native handler", d.AnthropicNativeHandler != nil)
	require("Gemini native handler", d.GeminiNativeHandler != nil)
	require("pool stats provider", d.PoolStats != nil)
	require("HTTP metrics", d.HTTPMetrics != nil)
	require("logger", d.Logger != nil)
	require("platform service accounts", d.PlatformServiceAccounts != nil)
	require("exchange rates", d.ExchangeRates != nil)

	if len(missing) > 0 {
		return fmt.Errorf("server dependencies: missing required %s", strings.Join(missing, ", "))
	}

	return nil
}

// NewValidatedFiberApp validates the production dependency graph before
// constructing the HTTP application.
func NewValidatedFiberApp(cfg *config.Config, deps Dependencies) (*fiber.App, error) {
	if cfg == nil {
		return nil, errors.New("server dependencies: configuration is nil")
	}

	if err := deps.ValidateRequired(); err != nil {
		return nil, err
	}

	return NewFiberApp(cfg, deps), nil
}
