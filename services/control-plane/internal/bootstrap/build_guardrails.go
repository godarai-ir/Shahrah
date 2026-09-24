package bootstrap

import (
	"net/http"
	"strings"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/guardrails"
	"github.com/Delpak/shahrah/services/control-plane/internal/semanticjudge"
	"go.uber.org/zap"
)

// guardrailEngine returns the one engine shared by the gateway and the management test endpoint, so both judge
// semantic policies with the same client, cache, and circuit breaker.
func (b *builder) guardrailEngine() *guardrails.Engine {
	if b.guardrailsEngine != nil {
		return b.guardrailsEngine
	}

	var options []guardrails.EngineOption

	if judge := b.buildSemanticJudge(); judge != nil {
		options = append(options, guardrails.WithSemanticJudge(judge))
	}

	b.guardrailsEngine = guardrails.NewDefaultEngine(options...)

	return b.guardrailsEngine
}

func (b *builder) buildSemanticJudge() guardrails.SemanticJudge {
	settings := b.cfg.SemanticJudge
	if !settings.Enabled || strings.TrimSpace(settings.APIKey) == "" {
		b.logger.Info("Semantic guardrails are disabled because no judge API key is configured")

		return nil
	}

	timeout := time.Duration(settings.TimeoutMS) * time.Millisecond

	client, err := semanticjudge.New(&http.Client{Timeout: timeout}, b.cache, semanticjudge.Config{
		BaseURL:          settings.BaseURL,
		APIKey:           settings.APIKey,
		Model:            settings.Model,
		Timeout:          timeout,
		CacheTTL:         time.Duration(settings.CacheTTLSeconds) * time.Second,
		CacheKeyPrefix:   settings.CacheKeyPrefix,
		FailureThreshold: settings.FailureThreshold,
		Cooldown:         time.Duration(settings.CooldownSeconds) * time.Second,
	})
	if err != nil {
		b.logger.Warn("Semantic guardrail judge is unavailable", zap.Error(err))

		return nil
	}

	return client
}
