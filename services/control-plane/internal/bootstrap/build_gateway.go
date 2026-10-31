package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/proxy"
	"github.com/Delpak/shahrah/services/control-plane/internal/responsecache"
	"github.com/Delpak/shahrah/services/control-plane/internal/runtimeprofile"
	"github.com/Delpak/shahrah/services/control-plane/internal/server"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func (b *builder) buildGateway() error {
	responseCache, err := b.buildResponseCache()
	if err != nil {
		return err
	}

	billingCapabilities := proxy.BillingCapabilities{
		MeteredUsage:             b.billing,
		MeteredUsagePolicy:       proxy.CapabilityFailClosed,
		DeploymentPricing:        b.billing,
		RequireDeploymentPricing: true,
		BudgetReservations:       b.billing,
		BudgetReservationPolicy:  proxy.CapabilityFailClosed,
		BatchUsage:               b.billing,
		BatchUsagePolicy:         proxy.CapabilityFailClosed,
		VideoUsage:               b.billing,
	}
	pipeline := proxy.NewRequestPipelineWithConfig(
		b.tunnel,
		b.tunnel,
		b.billing,
		billingCapabilities,
		b.logger,
		proxy.RequestPipelineConfig{DetachedWorkTimeout: b.cfg.Runtime.GatewayDetachedWorkTimeout},
	)

	baseDependencies := proxy.BaseHandlerDependencies{
		Pipeline:                  pipeline,
		Router:                    b.router,
		Tunnel:                    b.tunnel,
		RawTunnel:                 b.tunnel,
		Billing:                   b.billing,
		BillingCapabilities:       billingCapabilities,
		AnthropicRuntimeProfiles:  b.anthropicPricing,
		ProviderResources:         b.store.ProviderResourceRepository,
		AnthropicResources:        b.store.ProviderResourceRepository,
		MistralResources:          b.store.ProviderResourceRepository,
		Deployments:               b.runtimeDeploymentStore,
		ResponseProviderAuth:      b.runtimeDeploymentStore,
		RequestLogger:             b.requestLogger,
		SpendLogger:               b.spendLogger,
		Logger:                    b.logger,
		EmitMetric:                b.emitMetric,
		VirtualModelStore:         runtimeprofile.NewVirtualModelStore(b.runtimeProfiles),
		VirtualModelAliasCacheTTL: b.cfg.Runtime.VirtualModelAliasCacheTTL,
		VirtualModelNegativeTTL:   b.cfg.Runtime.VirtualModelNegativeTTL,
		ResponseCache:             responseCache,
		ProviderFeatures: proxy.ProviderFeatureConfig{
			Configured:                   true,
			MistralFilesEnabled:          b.cfg.ProviderFeatures.MistralFilesEnabled,
			MistralBatchEnabled:          b.cfg.ProviderFeatures.MistralBatchEnabled,
			MistralAgentsEnabled:         b.cfg.ProviderFeatures.MistralAgentsEnabled,
			MistralConversationsEnabled:  b.cfg.ProviderFeatures.MistralConversationsEnabled,
			MistralConversationStreaming: b.cfg.ProviderFeatures.MistralConversationStreaming,
			AnthropicAccountIDs:          b.cfg.ProviderFeatures.AnthropicAccountIDs,
			VideosEnabled:                b.cfg.ProviderFeatures.VideosEnabled,
			VideoCanaryTenantIDs:         b.cfg.ProviderFeatures.VideoCanaryTenantIDs,
		},
	}
	if err := baseDependencies.BillingCapabilities.Validate(); err != nil {
		return fmt.Errorf("validate gateway billing capabilities: %w", err)
	}

	chatHandler := proxy.NewChatCompletionHandlerWithDependencies(proxy.ChatCompletionDependencies{
		Base:                      baseDependencies,
		GuardrailResolver:         runtimeprofile.NewGuardrailResolver(b.runtimeProfiles),
		PromptResolver:            b.promptApplication,
		Guardrails:                b.guardrailEngine(),
		ResponseCache:             baseDependencies.ResponseCache,
		VirtualModelAliasCacheTTL: b.cfg.Runtime.VirtualModelAliasCacheTTL,
		VirtualModelNegativeTTL:   b.cfg.Runtime.VirtualModelNegativeTTL,
	})
	nativeResources := newNativeResourceSources(
		b.store.ProviderResourceRepository,
		b.runtimeDeploymentStore,
	)
	completionsHandler := proxy.NewCompletionsHandlerWithDependencies(baseDependencies)
	responsesHandler := proxy.NewResponsesHandlerWithDependencies(baseDependencies)
	embeddingsHandler := proxy.NewEmbeddingsHandlerWithDependencies(baseDependencies)
	moderationsHandler := proxy.NewModerationsHandlerWithDependencies(baseDependencies)
	rerankHandler := proxy.NewRerankHandlerWithDependencies(baseDependencies)
	imagesHandler := proxy.NewImagesHandlerWithDependencies(baseDependencies)
	audioTranscriptionsHandler := proxy.NewAudioTranscriptionsHandlerWithDependencies(baseDependencies)
	audioSpeechHandler := proxy.NewAudioSpeechHandlerWithDependencies(baseDependencies)
	mistralOCRHandler := proxy.NewMistralOCRHandlerWithDependencies(baseDependencies)
	zaiLayoutParsingHandler := proxy.NewZAILayoutParsingHandlerWithDependencies(baseDependencies)
	mistralNativeHandler := proxy.NewMistralNativeHandler(baseDependencies, nativeResources)
	modelsHandler := proxy.NewModelsHandler(b.runtimeDeploymentStore)
	anthropicNativeHandler := proxy.NewAnthropicNativeHandler(baseDependencies, nativeResources)
	tokenBazaarMessagesHandler := proxy.NewTokenBazaarMessagesHandlerWithDependencies(baseDependencies)
	tokenBazaarGenerateContentHandler := proxy.NewTokenBazaarGenerateContentHandlerWithDependencies(baseDependencies)
	geminiNativeHandler := proxy.NewGeminiNativeHandler(baseDependencies, nativeResources)
	videosHandler := proxy.NewVideosHandler(baseDependencies, b.store.ProviderResourceRepository)

	if err := b.lifecycle.AddLoop("Anthropic batch reconciler", func(ctx context.Context) {
		anthropicNativeHandler.RunBatchWorker(ctx, 15*time.Second)
	}); err != nil {
		return err
	}

	if err := b.lifecycle.AddLoop("Mistral batch reconciler", func(ctx context.Context) {
		mistralNativeHandler.RunBatchWorker(ctx, 15*time.Second)
	}); err != nil {
		return err
	}

	if err := b.lifecycle.AddLoop("Video job reconciler", func(ctx context.Context) {
		videosHandler.RunVideoWorker(ctx, 5*time.Second)
	}); err != nil {
		return err
	}

	b.serverDependencies = server.Dependencies{
		AuthMiddleware:                    b.authMiddleware,
		TenantResolver:                    b.tenantMiddleware,
		TracingMiddleware:                 b.tracingMiddleware,
		TraceHealth:                       b.traceHealth,
		ClickHouseHealth:                  b.store.ObservabilityHealthRepository,
		ManagementRoutes:                  b.managementRoutes,
		ProxyHandler:                      chatHandler,
		CompletionsHandler:                completionsHandler,
		ResponsesHandler:                  responsesHandler,
		EmbeddingsHandler:                 embeddingsHandler,
		ModerationsHandler:                moderationsHandler,
		RerankHandler:                     rerankHandler,
		ImagesHandler:                     imagesHandler,
		VideosHandler:                     videosHandler,
		AudioTranscriptionsHandler:        audioTranscriptionsHandler,
		AudioSpeechHandler:                audioSpeechHandler,
		MistralOCRHandler:                 mistralOCRHandler,
		ZAILayoutParsingHandler:           zaiLayoutParsingHandler,
		MistralNativeHandler:              mistralNativeHandler,
		ModelsHandler:                     modelsHandler,
		AnthropicNativeHandler:            anthropicNativeHandler,
		TokenBazaarMessagesHandler:        tokenBazaarMessagesHandler,
		TokenBazaarGenerateContentHandler: tokenBazaarGenerateContentHandler,
		GeminiNativeHandler:               geminiNativeHandler,
		PoolStats:                         b.store.DatabaseHealthRepository,
		HTTPMetrics:                       b.httpMetrics,
		Logger:                            b.logger,
		PlatformServiceAccounts:           b.store.PlatformServiceAccountRepository,
		ExchangeRates:                     b.store.ExchangeRateRepository,
		GatewayMaxInFlight:                &b.cfg.Runtime.GatewayMaxInFlight,
		CreditSurface:                     b.creditSurfaceMiddleware(),
		RateCardHandler:                   proxy.NewRateCardHandler(b.billing, b.creditRateCards),
		BudgetMeHandler:                   proxy.NewBudgetMeHandler(b.billing),
		SeatRequestHandler:                b.seatRequestHandler(),
		PricingChangelogHandler:           b.pricingChangelogHandler(),
	}

	return nil
}

func (b *builder) buildResponseCache() (*responsecache.Service, error) {
	var embedder responsecache.Embedder

	_, sharedResponseCacheAvailable := b.cache.(*cache.DualCache)

	embeddingConfigured := strings.TrimSpace(b.cfg.ResponseCache.EmbeddingBaseURL) != "" &&
		strings.TrimSpace(b.cfg.ResponseCache.EmbeddingAPIKey) != ""
	if b.cfg.ResponseCache.Enabled && sharedResponseCacheAvailable && embeddingConfigured {
		candidate, err := responsecache.NewArvanCloudBGEEmbedder(
			&http.Client{Timeout: time.Duration(b.cfg.ResponseCache.EmbeddingTimeoutSec) * time.Second},
			responsecache.ArvanCloudConfig{
				BaseURL:      b.cfg.ResponseCache.EmbeddingBaseURL,
				APIKey:       b.cfg.ResponseCache.EmbeddingAPIKey,
				Model:        b.cfg.ResponseCache.EmbeddingModel,
				Revision:     b.cfg.ResponseCache.EmbeddingRevision,
				Dimensions:   b.cfg.ResponseCache.EmbeddingDimensions,
				MaxBatchSize: b.cfg.ResponseCache.EmbeddingMaxBatchSize,
			},
		)
		if err != nil {
			b.logger.Warn("Smart response cache embedder is unavailable", zap.Error(err))
		} else {
			embedder = candidate
		}
	} else if b.cfg.ResponseCache.Enabled && !sharedResponseCacheAvailable {
		b.logger.Warn("Smart response cache disabled because shared Redis is unavailable")
	}

	service, err := responsecache.NewService(b.cache, embedder, responsecache.Config{
		Enabled:               b.cfg.ResponseCache.Enabled,
		KeyPrefix:             b.cfg.ResponseCache.KeyPrefix,
		DefaultTTL:            time.Duration(b.cfg.ResponseCache.DefaultTTLSeconds) * time.Second,
		MinTTL:                time.Duration(b.cfg.ResponseCache.MinTTLSeconds) * time.Second,
		MaxTTL:                time.Duration(b.cfg.ResponseCache.MaxTTLSeconds) * time.Second,
		MaxResponseBytes:      b.cfg.ResponseCache.MaxResponseBytes,
		MaxSemanticCandidates: b.cfg.ResponseCache.MaxSemanticCandidates,
	})
	if err != nil {
		return nil, fmt.Errorf("initialize response cache: %w", err)
	}

	return service, nil
}

func (b *builder) buildHTTP() error {
	app, err := server.NewValidatedFiberApp(b.cfg, b.serverDependencies)
	if err != nil {
		return err
	}

	b.httpApp = app

	return nil
}

// seatRequestHandler lets members ask for a better seat from the gateway.
func (b *builder) seatRequestHandler() fiber.Handler {
	if b.store == nil || b.store.SeatRepository == nil {
		return nil
	}

	return proxy.NewSeatRequestHandler(b.store.SeatRepository)
}
