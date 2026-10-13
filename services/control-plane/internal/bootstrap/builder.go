package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	"github.com/Delpak/shahrah/services/control-plane/internal/guardrails"
	"github.com/Delpak/shahrah/services/control-plane/internal/iam"
	"github.com/Delpak/shahrah/services/control-plane/internal/management"
	metricsobs "github.com/Delpak/shahrah/services/control-plane/internal/observability/metrics"
	prommetrics "github.com/Delpak/shahrah/services/control-plane/internal/observability/prometheus"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/Delpak/shahrah/services/control-plane/internal/oidc"
	"github.com/Delpak/shahrah/services/control-plane/internal/platformcredentials"
	"github.com/Delpak/shahrah/services/control-plane/internal/policies"
	"github.com/Delpak/shahrah/services/control-plane/internal/prompts"
	"github.com/Delpak/shahrah/services/control-plane/internal/provideraccounts"
	"github.com/Delpak/shahrah/services/control-plane/internal/proxy"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/runtimeprofile"
	"github.com/Delpak/shahrah/services/control-plane/internal/server"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenantadmin"
	tunneltransport "github.com/Delpak/shahrah/services/control-plane/internal/tunnel"
	"github.com/Delpak/shahrah/services/control-plane/internal/virtualmodels"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type builder struct {
	ctx       context.Context
	cfg       *config.Config
	logger    *zap.Logger
	lifecycle *Lifecycle

	cache                      cache.Cache
	store                      *storage.Store
	tunnel                     *tunneltransport.HTTPTransport
	tenantProviderProber       *platformcredentials.TunnelCredentialProber
	credentialRuntime          platformCredentialRuntime
	runtimeProfiles            *runtimeprofile.Manager
	runtimeDeploymentStore     *runtimeprofile.DeploymentStore
	router                     *router.Router
	authMiddleware             *auth.AuthMiddleware
	tenantMiddleware           *tenant.Middleware
	billing                    *billing.Service
	anthropicPricing           *billing.AnthropicPricingSnapshotManager
	creditRateCards            *billing.CreditRateCardCache
	httpMetrics                *prommetrics.HTTPMetrics
	emitMetric                 func(metricsobs.BuildInput)
	requestLogger              proxy.RequestLogger
	spendLogger                proxy.SpendLogger
	managementRoutes           *management.Routes
	promptApplication          *prompts.Service
	iamApplication             *iam.Service
	tenantAdminApplication     *tenantadmin.Service
	policyApplication          *policies.Service
	modelApplication           *virtualmodels.Service
	providerAccountApplication *provideraccounts.Service
	oidcApplication            *oidc.Service
	tracingMiddleware          fiber.Handler
	traceHealth                *tracing.ReliableExporter
	serverDependencies         server.Dependencies
	httpApp                    *fiber.App
	guardrailsEngine           *guardrails.Engine
}

// Build assembles and validates the complete control-plane application.
func Build(ctx context.Context, cfg *config.Config, logger *zap.Logger) (_ *Application, err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if cfg == nil {
		return nil, errors.New("bootstrap: configuration is nil")
	}

	if logger == nil {
		return nil, errors.New("bootstrap: logger is nil")
	}

	if strings.TrimSpace(cfg.GeneralSettings.DatabaseURL) == "" {
		return nil, errors.New("bootstrap: general_settings.database_url is required")
	}

	b := &builder{
		ctx:         ctx,
		cfg:         cfg,
		logger:      logger,
		lifecycle:   NewLifecycle(),
		httpMetrics: prommetrics.NewHTTPMetrics(),
	}

	defer func() {
		if err == nil {
			return
		}

		cleanupCtx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
		defer cancel()

		err = errors.Join(err, b.lifecycle.Stop(cleanupCtx))
	}()

	configureRuntimeProfiling(cfg, logger)

	for _, build := range []struct {
		name string
		run  func() error
	}{
		{name: "storage", run: b.buildStorage},
		{name: "workers", run: b.buildWorkers},
		{name: "management", run: b.buildManagement},
		{name: "gateway", run: b.buildGateway},
		{name: "HTTP server", run: b.buildHTTP},
	} {
		if buildErr := build.run(); buildErr != nil {
			return nil, fmt.Errorf("bootstrap: build %s: %w", build.name, buildErr)
		}
	}

	return &Application{
		HTTP:      b.httpApp,
		Address:   fmt.Sprintf("%s:%d", cfg.GeneralSettings.Host, cfg.GeneralSettings.Port),
		lifecycle: b.lifecycle,
	}, nil
}
