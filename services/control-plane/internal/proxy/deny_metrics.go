package proxy

import "go.uber.org/zap"

const (
	denyReasonKeyAllowlist   = "key_allowlist"
	denyReasonProviderPolicy = "provider_account_policy"
	denyReasonUnknownModel   = "unknown_model"
)

func emitDenyCounter(logger *zap.Logger, reason string, model string, tenantID string) {
	if logger == nil {
		return
	}
	logger.Warn("proxy.request_denied",
		zap.String("counter", "proxy_request_denied_total"),
		zap.String("reason", reason),
		zap.String("model", model),
		zap.String("tenant_id", tenantID),
		zap.Int("count", 1),
	)
}
