package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	"github.com/Delpak/shahrah/services/control-plane/internal/platformcredentials"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"go.uber.org/zap"
)

type platformCredentialRuntime struct {
	mode    string
	manager *platformcredentials.Manager
	service *platformcredentials.Service
	schemas platformcredentials.SchemaCatalog
	bus     *platformcredentials.RedisInvalidationBus
}

func newPlatformCredentialRuntime(
	ctx context.Context,
	store *storage.PlatformCredentialRepository,
	sharedCache cache.Cache,
	metrics platformcredentials.Metrics,
	prober platformcredentials.CredentialProber,
	logger *zap.Logger,
	settings config.CredentialSettings,
	discovery config.AdminDiscoverySettings,
) (platformCredentialRuntime, error) {
	mode := strings.TrimSpace(settings.PlatformMode)
	if mode == "" {
		mode = platform.ProviderCredentialModeEnv
	}

	switch mode {
	case platform.ProviderCredentialModeEnv,
		platform.ProviderCredentialModeStore,
		platform.ProviderCredentialModeBreakGlass:
	default:
		return platformCredentialRuntime{}, fmt.Errorf("unsupported platform provider credential mode %q", mode)
	}

	runtime := platformCredentialRuntime{mode: mode}
	schemas, schemaErr := platformcredentials.NewModelDiscoverySchemaCatalog(
		discovery.BaseURL,
		discovery.Token,
		nil,
	)
	runtime.schemas = schemas

	if schemaErr != nil {
		logger.Warn("Provider connection schema catalog is unavailable", zap.Error(schemaErr))
	}

	credentialsConfigured := mode != platform.ProviderCredentialModeEnv || settings.PlatformKEKsJSON != ""
	if !credentialsConfigured {
		return runtime, nil
	}

	keks, kekErr := platformcredentials.NewEnvironmentKEKProvider(
		settings.PlatformActiveVersion,
		settings.PlatformKEKsJSON,
	)

	fingerprintKey, fingerprintErr := platformcredentials.DecodeFingerprintKey(settings.PlatformFingerprintKey)
	if kekErr != nil || fingerprintErr != nil {
		if mode == platform.ProviderCredentialModeStore {
			return platformCredentialRuntime{}, errors.New("platform provider credential bootstrap keys are invalid")
		}

		logger.Warn("Platform provider credential store is unavailable; environment break-glass remains active")

		return runtime, nil
	}

	secretStore, err := platformcredentials.NewEncryptedDBSecretStore(keks, fingerprintKey)
	if err != nil {
		return platformCredentialRuntime{}, fmt.Errorf("initialize platform credential encryption: %w", err)
	}

	runtime.manager = platformcredentials.NewManager(store, secretStore, metrics)
	if err := runtime.manager.Reload(ctx); err != nil {
		if mode == platform.ProviderCredentialModeStore {
			return platformCredentialRuntime{}, fmt.Errorf("pre-warm platform provider credential snapshot: %w", err)
		}

		logger.Warn(
			"Initial platform credential snapshot load failed; environment break-glass remains active",
			zap.Error(err),
		)
	}

	if redisCache, ok := sharedCache.(*cache.DualCache); ok {
		runtime.bus = platformcredentials.NewRedisInvalidationBus(redisCache)
	}

	runtime.service = platformcredentials.NewService(
		store,
		secretStore,
		runtime.manager,
		runtime.bus,
		runtime.schemas,
		prober,
	)

	return runtime, nil
}

func (b *builder) registerPlatformCredentialWorkers() error {
	runtime := b.credentialRuntime
	if runtime.manager == nil {
		return nil
	}

	if err := b.lifecycle.AddLoop("platform credential reconciler", func(ctx context.Context) {
		runtime.manager.Reconcile(ctx, 30*time.Second)
	}); err != nil {
		return err
	}

	if runtime.bus == nil {
		return nil
	}

	return b.lifecycle.AddLoop("platform credential invalidation consumer", func(ctx context.Context) {
		if err := runtime.bus.Consume(ctx, runtime.manager); err != nil && !errors.Is(err, context.Canceled) {
			b.logger.Warn("Platform credential invalidation consumer stopped", zap.Error(err))
		}
	})
}
