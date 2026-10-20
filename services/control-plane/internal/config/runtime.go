package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// StorageSettings is the process-level storage configuration snapshot.
type StorageSettings struct {
	PostgresMaxConnections       int32
	PostgresMinConnections       int32
	PostgresMaxLifetime          time.Duration
	PostgresMaxIdleTime          time.Duration
	PostgresHealthCheckPeriod    time.Duration
	ClickHouseDSN                string
	ClickHouseRequired           bool
	ClickHouseAttempts           int
	ClickHouseRetryDelay         time.Duration
	RequestLoggingModeTTL        time.Duration
	TokenUsageEnabled            bool
	TokenUsageFailOpen           bool
	TokenUsageKeyPrefix          string
	TokenUsageFlushInterval      time.Duration
	TokenUsageFlushBatchSize     int
	TokenUsageDeltaTTL           time.Duration
	LastUsedFlushInterval        time.Duration
	LastUsedTTL                  time.Duration
	ModelCacheTTL                time.Duration
	ModelCacheStaleTTL           time.Duration
	AuthorizationCacheTTL        time.Duration
	AuthorizationCacheMaxEntries int
	EntitlementCacheTTL          time.Duration
	TenantCredentialCacheTTL     time.Duration
	GatewayDatabaseTimeout       time.Duration
	SubscriptionDatabaseTimeout  time.Duration
	BudgetWindowTTLGrace         time.Duration
	// CreditOverageCeilingCredits bounds overage for credit tenants that
	// enabled it without a cap.
	CreditOverageCeilingCredits int64
}

// RuntimeSettings is the process-level non-storage runtime configuration snapshot.
type RuntimeSettings struct {
	RuntimeProfileVersionCheck time.Duration
	RuntimeProfileStaleGrace   time.Duration
	RuntimeProfileL2TTL        time.Duration
	RouterDeploymentCacheTTL   time.Duration
	PlatformCredentialEnv      string
	RedisSentinelAddresses     []string
	RedisMasterName            string
	RedisSentinelPassword      string
	GatewayDetachedWorkTimeout time.Duration
	VirtualModelAliasCacheTTL  time.Duration
	VirtualModelNegativeTTL    time.Duration
	BillingExchangeRateTTL     time.Duration
	BillingExchangeRateStale   time.Duration
	BillingPricingCacheEnabled bool
	BillingPricingCacheTTL     time.Duration
	BillingCreditRateCardTTL   time.Duration
	BillingCreditLedger        CreditLedgerSettings
	// SeatsEnabled turns on member seats: personal seat credits for chat usage
	// and the seat lifecycle worker.
	SeatsEnabled                  bool
	RateLimitPolicyFailOpen       bool
	NobitexExchangeRateEnabled    bool
	NobitexExchangeRateConfigured bool
	NobitexExchangeRateBaseURL    string
	NobitexExchangeRateSymbol     string
	NobitexExchangeRateSource     string
	NobitexExchangeRateInterval   time.Duration
	NobitexExchangeRateTimeout    time.Duration
	NobitexExchangeRateMinTrades  int
	GatewayMaxInFlight            int
}

// LoadStorageSettings reads storage environment once at the configuration edge.
func LoadStorageSettings() StorageSettings {
	maxConnections := int32(positiveIntEnv("POSTGRES_POOL_MAX_CONNS", 32))
	minConnections := int32(positiveIntEnv("POSTGRES_POOL_MIN_CONNS", 4))

	if minConnections > maxConnections {
		minConnections = maxConnections
	}

	return StorageSettings{
		PostgresMaxConnections:       maxConnections,
		PostgresMinConnections:       minConnections,
		PostgresMaxLifetime:          secondsEnv("POSTGRES_POOL_MAX_CONN_LIFETIME_SECONDS", 30*time.Minute),
		PostgresMaxIdleTime:          secondsEnv("POSTGRES_POOL_MAX_CONN_IDLE_TIME_SECONDS", 5*time.Minute),
		PostgresHealthCheckPeriod:    secondsEnv("POSTGRES_POOL_HEALTH_CHECK_PERIOD_SECONDS", 30*time.Second),
		ClickHouseDSN:                strings.TrimSpace(os.Getenv("CLICKHOUSE_DSN")),
		ClickHouseRequired:           booleanEnv("CLICKHOUSE_REQUIRED", false),
		ClickHouseAttempts:           positiveIntEnv("CLICKHOUSE_INIT_ATTEMPTS", 10),
		ClickHouseRetryDelay:         secondsEnv("CLICKHOUSE_INIT_RETRY_SECONDS", time.Second),
		RequestLoggingModeTTL:        secondsEnv("REQUEST_LOG_MODE_CACHE_TTL_SECONDS", time.Minute),
		TokenUsageEnabled:            booleanEnv("TOKEN_USAGE_AGGREGATION_ENABLED", true),
		TokenUsageFailOpen:           booleanEnv("TOKEN_USAGE_REDIS_FAIL_OPEN", true),
		TokenUsageKeyPrefix:          stringEnv("TOKEN_USAGE_REDIS_KEY_PREFIX", "usage"),
		TokenUsageFlushInterval:      secondsEnv("TOKEN_USAGE_FLUSH_INTERVAL_SECONDS", 5*time.Second),
		TokenUsageFlushBatchSize:     positiveIntEnv("TOKEN_USAGE_FLUSH_BATCH_SIZE", 500),
		TokenUsageDeltaTTL:           secondsEnv("TOKEN_USAGE_DELTA_TTL_SECONDS", 24*time.Hour),
		LastUsedFlushInterval:        secondsEnv("TOKEN_LAST_USED_FLUSH_INTERVAL_SECONDS", time.Minute),
		LastUsedTTL:                  secondsEnv("TOKEN_LAST_USED_TTL_SECONDS", 24*time.Hour),
		ModelCacheTTL:                secondsEnv("MODEL_LIST_CACHE_TTL_SECONDS", 30*time.Second),
		ModelCacheStaleTTL:           secondsEnv("MODEL_LIST_CACHE_STALE_SECONDS", 10*time.Second),
		AuthorizationCacheTTL:        secondsEnv("AUTHORIZATION_CACHE_TTL_SECONDS", time.Minute),
		AuthorizationCacheMaxEntries: positiveIntEnv("AUTHORIZATION_CACHE_MAX_ENTRIES", 5000),
		EntitlementCacheTTL:          secondsEnv("ENTITLEMENT_CACHE_TTL_SECONDS", time.Minute),
		TenantCredentialCacheTTL:     secondsEnv("TENANT_PROVIDER_CREDENTIAL_CACHE_TTL_SECONDS", time.Minute),
		GatewayDatabaseTimeout:       secondsEnv("GATEWAY_DB_TIMEOUT_SECONDS", 3*time.Second),
		SubscriptionDatabaseTimeout:  secondsEnv("SUBSCRIPTION_LIFECYCLE_DB_TIMEOUT_SECONDS", 5*time.Second),
		BudgetWindowTTLGrace:         secondsEnv("BILLING_BUDGET_WINDOW_GRACE_SECONDS", 24*time.Hour),
		CreditOverageCeilingCredits:  positiveInt64Env("BILLING_CREDIT_OVERAGE_DEFAULT_CEILING_CREDITS", 100_000),
	}
}

// LoadRuntimeSettings reads composition settings once at the configuration edge.
func LoadRuntimeSettings() RuntimeSettings {
	nobitexEnabled, nobitexConfigured := optionalBooleanEnv("NOBITEX_EXCHANGE_RATE_ENABLED")

	return RuntimeSettings{
		RuntimeProfileVersionCheck:    secondsEnv("TENANT_RUNTIME_PROFILE_VERSION_CHECK_SECONDS", time.Second),
		RuntimeProfileStaleGrace:      secondsEnv("TENANT_RUNTIME_PROFILE_STALE_GRACE_SECONDS", 5*time.Minute),
		RuntimeProfileL2TTL:           secondsEnv("TENANT_RUNTIME_PROFILE_L2_TTL_SECONDS", time.Hour),
		RouterDeploymentCacheTTL:      nonNegativeSecondsEnv("ROUTER_DEPLOYMENT_CACHE_TTL_SECONDS", 30*time.Second),
		PlatformCredentialEnv:         stringEnv("PLATFORM_PROVIDER_CREDENTIAL_ENVIRONMENT", ""),
		RedisSentinelAddresses:        csvEnv("REDIS_SENTINEL_ADDRS"),
		RedisMasterName:               stringEnv("REDIS_MASTER_NAME", "mymaster"),
		RedisSentinelPassword:         stringEnv("REDIS_SENTINEL_PASSWORD", ""),
		GatewayDetachedWorkTimeout:    secondsEnv("GATEWAY_DETACHED_WORK_TIMEOUT_SECONDS", 5*time.Second),
		VirtualModelAliasCacheTTL:     secondsEnv("VIRTUAL_MODEL_ALIAS_CACHE_TTL_SECONDS", time.Minute),
		VirtualModelNegativeTTL:       secondsEnv("VIRTUAL_MODEL_ALIAS_NEGATIVE_CACHE_TTL_SECONDS", 10*time.Second),
		BillingExchangeRateTTL:        secondsEnv("BILLING_EXCHANGE_RATE_CACHE_TTL_SECONDS", time.Minute),
		BillingExchangeRateStale:      secondsEnv("BILLING_EXCHANGE_RATE_CACHE_MAX_STALE_SECONDS", 10*time.Minute),
		BillingPricingCacheEnabled:    booleanEnv("BILLING_PRICING_CACHE_ENABLED", true),
		BillingPricingCacheTTL:        secondsEnv("BILLING_PRICING_CACHE_TTL_SECONDS", 5*time.Minute),
		BillingCreditRateCardTTL:      secondsEnv("BILLING_CREDIT_RATE_CARD_TTL_SECONDS", 5*time.Minute),
		BillingCreditLedger:           loadCreditLedgerSettings(),
		SeatsEnabled:                  booleanEnv("SEATS_ENABLED", false),
		RateLimitPolicyFailOpen:       booleanEnv("RATE_LIMIT_POLICY_REDIS_FAIL_OPEN", false),
		NobitexExchangeRateEnabled:    nobitexEnabled,
		NobitexExchangeRateConfigured: nobitexConfigured,
		NobitexExchangeRateBaseURL:    stringEnv("NOBITEX_EXCHANGE_RATE_BASE_URL", "https://apiv2.nobitex.ir"),
		NobitexExchangeRateSymbol:     stringEnv("NOBITEX_EXCHANGE_RATE_SYMBOL", "USDTIRT"),
		NobitexExchangeRateSource:     stringEnv("NOBITEX_EXCHANGE_RATE_SOURCE", ""),
		NobitexExchangeRateInterval:   secondsEnv("NOBITEX_EXCHANGE_RATE_INTERVAL_SECONDS", 5*time.Minute),
		NobitexExchangeRateTimeout:    secondsEnv("NOBITEX_EXCHANGE_RATE_TIMEOUT_SECONDS", 10*time.Second),
		NobitexExchangeRateMinTrades:  positiveIntEnv("NOBITEX_EXCHANGE_RATE_MIN_TRADES", 1),
		GatewayMaxInFlight:            nonNegativeIntEnv("GATEWAY_MAX_IN_FLIGHT_REQUESTS", 512),
	}
}

func optionalBooleanEnv(key string) (bool, bool) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return false, false
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, false
	}

	return value, true
}

func secondsEnv(key string, fallback time.Duration) time.Duration {
	value := positiveIntEnv(key, int(fallback/time.Second))

	return time.Duration(value) * time.Second
}

func nonNegativeSecondsEnv(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return fallback
	}

	return time.Duration(value) * time.Second
}

func positiveIntEnv(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}

	return value
}

func nonNegativeIntEnv(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return fallback
	}

	return value
}

func booleanEnv(key string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}

	return value
}

func stringEnv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func csvEnv(key string) []string {
	parts := strings.Split(os.Getenv(key), ",")
	values := make([]string, 0, len(parts))

	for _, part := range parts {
		if value := strings.TrimSpace(part); value != "" {
			values = append(values, value)
		}
	}

	return values
}

// CreditLedgerSettings configures the credit ledger workers (pricing PRD, Phase 2).
type CreditLedgerSettings struct {
	SettlementInterval time.Duration
	SettleDelay        time.Duration
	BatchSize          int
	ExpiryInterval     time.Duration
	ReconcileInterval  time.Duration
	PriceInterval      time.Duration
	PriceAutoApply     bool
	Timeout            time.Duration
}

func loadCreditLedgerSettings() CreditLedgerSettings {
	return CreditLedgerSettings{
		SettlementInterval: secondsEnv("BILLING_CREDIT_SETTLEMENT_INTERVAL_SECONDS", 30*time.Second),
		SettleDelay:        secondsEnv("BILLING_CREDIT_SETTLE_DELAY_SECONDS", 30*time.Second),
		BatchSize:          positiveIntEnv("BILLING_CREDIT_SETTLEMENT_BATCH_SIZE", 500),
		ExpiryInterval:     secondsEnv("BILLING_CREDIT_EXPIRY_INTERVAL_SECONDS", 5*time.Minute),
		ReconcileInterval:  secondsEnv("BILLING_CREDIT_RECONCILE_INTERVAL_SECONDS", 24*time.Hour),
		PriceInterval:      secondsEnv("BILLING_CREDIT_PRICE_INTERVAL_SECONDS", time.Hour),
		PriceAutoApply:     booleanEnv("BILLING_CREDIT_PRICE_AUTO_APPLY", false),
		Timeout:            secondsEnv("BILLING_CREDIT_JOB_TIMEOUT_SECONDS", 2*time.Minute),
	}
}
