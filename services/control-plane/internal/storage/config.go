package storage

import "time"

// Config contains immutable storage settings supplied by the composition root.
type Config struct {
	Postgres        PostgresConfig
	ClickHouse      ClickHouseConfig
	TokenUsage      TokenUsageConfig
	HotPath         HotPathConfig
	Plans           PlanRepositoryConfig
	Wallets         WalletRepositoryConfig
	ProviderAccount ProviderAccountRepositoryConfig
}

// PostgresConfig controls the PostgreSQL connection pool.
type PostgresConfig struct {
	MaxConnections    int32
	MinConnections    int32
	MaxLifetime       time.Duration
	MaxIdleTime       time.Duration
	HealthCheckPeriod time.Duration
}

// ClickHouseConfig controls the optional ClickHouse connection.
type ClickHouseConfig struct {
	DSN        string
	Required   bool
	Attempts   int
	RetryDelay time.Duration
}

// TokenUsageConfig controls asynchronous token-usage aggregation.
type TokenUsageConfig struct {
	Enabled          bool
	FailOpen         bool
	KeyPrefix        string
	FlushInterval    time.Duration
	FlushBatchSize   int
	DeltaTTL         time.Duration
	LastUsedInterval time.Duration
	LastUsedTTL      time.Duration
}

// HotPathConfig controls request-path caches and database timeouts.
type HotPathConfig struct {
	RequestLoggingModeTTL        time.Duration
	ModelCacheTTL                time.Duration
	ModelCacheStaleTTL           time.Duration
	DatabaseTimeout              time.Duration
	SubscriptionLifecycleTimeout time.Duration
	AuthorizationCacheTTL        time.Duration
	AuthorizationCacheMaxEntries int
}

// PlanRepositoryConfig controls plan and entitlement persistence behavior.
type PlanRepositoryConfig struct {
	EntitlementCacheTTL time.Duration
}

// WalletRepositoryConfig controls wallet cache synchronization behavior.
type WalletRepositoryConfig struct {
	BudgetWindowTTLGrace time.Duration
	// CreditOverageCeilingMC bounds overage headroom when a tenant enabled
	// overage without a cap.
	CreditOverageCeilingMC int64
}

// ProviderAccountRepositoryConfig controls tenant credential caching.
type ProviderAccountRepositoryConfig struct {
	CredentialCacheTTL time.Duration
}

// DefaultConfig returns safe storage defaults without reading process state.
func DefaultConfig() Config {
	return Config{
		Postgres: PostgresConfig{
			MaxConnections:    32,
			MinConnections:    4,
			MaxLifetime:       30 * time.Minute,
			MaxIdleTime:       5 * time.Minute,
			HealthCheckPeriod: 30 * time.Second,
		},
		ClickHouse: ClickHouseConfig{
			Attempts:   10,
			RetryDelay: time.Second,
		},
		TokenUsage: TokenUsageConfig{
			Enabled:          true,
			FailOpen:         true,
			KeyPrefix:        "usage",
			FlushInterval:    5 * time.Second,
			FlushBatchSize:   500,
			DeltaTTL:         24 * time.Hour,
			LastUsedInterval: time.Minute,
			LastUsedTTL:      24 * time.Hour,
		},
		HotPath: HotPathConfig{
			RequestLoggingModeTTL:        time.Minute,
			ModelCacheTTL:                30 * time.Second,
			ModelCacheStaleTTL:           10 * time.Second,
			DatabaseTimeout:              defaultHotPathDBTimeout,
			SubscriptionLifecycleTimeout: defaultSubscriptionLifecycleDBTimeout,
			AuthorizationCacheTTL:        time.Minute,
			AuthorizationCacheMaxEntries: 5000,
		},
		Plans: PlanRepositoryConfig{
			EntitlementCacheTTL: time.Minute,
		},
		Wallets: WalletRepositoryConfig{
			BudgetWindowTTLGrace: 24 * time.Hour,
		},
		ProviderAccount: ProviderAccountRepositoryConfig{
			CredentialCacheTTL: time.Minute,
		},
	}
}
