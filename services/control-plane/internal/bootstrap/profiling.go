package bootstrap

import (
	"runtime"

	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	"go.uber.org/zap"
)

func configureRuntimeProfiling(cfg *config.Config, logger *zap.Logger) {
	if cfg.Profiling.MutexProfileFraction > 0 {
		runtime.SetMutexProfileFraction(cfg.Profiling.MutexProfileFraction)
		logger.Info("Enabled mutex profiling", zap.Int("fraction", cfg.Profiling.MutexProfileFraction))
	}

	if cfg.Profiling.BlockProfileRate > 0 {
		runtime.SetBlockProfileRate(cfg.Profiling.BlockProfileRate)
		logger.Info("Enabled block profiling", zap.Int("rate", cfg.Profiling.BlockProfileRate))
	}
}
