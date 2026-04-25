package storage

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Store provides PostgreSQL-backed persistence for the control-plane.
type Store struct {
	pool          *pgxpool.Pool
	logger        *zap.Logger
	defaultTenant string
}

// NewStore creates a new PostgreSQL store.
func NewStore(ctx context.Context, databaseURL, defaultTenant string, logger *zap.Logger) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return &Store{pool: pool, logger: logger, defaultTenant: defaultTenant}, nil
}

// Close closes the store.
func (s *Store) Close() {
	s.pool.Close()
}

// GetTenantByHost resolves a tenant from the request host.
func (s *Store) GetTenantByHost(ctx context.Context, host string) (*platform.Tenant, error) {
	slug := tenantSlugFromHost(host)
	if slug == "" || slug == "localhost" {
		return s.EnsureTenant(ctx, s.defaultTenant)
	}

	var tenant platform.Tenant
	err := s.pool.QueryRow(ctx, `
		SELECT tenant_id::text, slug, domain, status, created_at
		FROM tenants
		WHERE slug = $1 OR domain = $2
		LIMIT 1
	`, slug, host).Scan(&tenant.ID, &tenant.Slug, &tenant.Domain, &tenant.Status, &tenant.CreatedAt)
	if err == nil {
		return &tenant, nil
	}
	return nil, err
}

// EnsureTenant creates the tenant and wallet when missing.
func (s *Store) EnsureTenant(ctx context.Context, slug string) (*platform.Tenant, error) {
	if slug == "" {
		slug = s.defaultTenant
	}

	var tenant platform.Tenant
	err := s.pool.QueryRow(ctx, `
		INSERT INTO tenants (slug, domain, status)
		VALUES ($1, $2, 'active')
		ON CONFLICT (slug)
		DO UPDATE SET slug = EXCLUDED.slug
		RETURNING tenant_id::text, slug, domain, status, created_at
	`, slug, slug).Scan(&tenant.ID, &tenant.Slug, &tenant.Domain, &tenant.Status, &tenant.CreatedAt)
	if err != nil {
		return nil, err
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO wallets (tenant_id, balance_irr, pending_irr)
		VALUES ($1, 0, 0)
		ON CONFLICT (tenant_id) DO NOTHING
	`, tenant.ID)
	if err != nil {
		return nil, err
	}

	return &tenant, nil
}

// GetKeyByHash retrieves a key by its hashed token.
func (s *Store) GetKeyByHash(ctx context.Context, hash string) (*platform.AuthContext, error) {
	var auth platform.AuthContext
	var maxBudget *float64
	var rpmLimit *int64
	var tpmLimit *int64
	var expiresAt *time.Time
	var models []string

	err := s.pool.QueryRow(ctx, `
		SELECT id::text,
		       tenant_id::text,
		       COALESCE(organization_id::text, ''),
		       COALESCE(team_id::text, ''),
		       COALESCE(key_alias, ''),
		       COALESCE(token_hash, ''),
		       COALESCE(models, '{}'),
		       COALESCE(spend, 0),
		       max_budget,
		       rpm_limit,
		       tpm_limit,
		       COALESCE(expires, NULL),
		       COALESCE(blocked, false)
		FROM verification_tokens
		WHERE token_hash = $1
		LIMIT 1
	`, hash).Scan(
		&auth.KeyID,
		&auth.TenantID,
		&auth.OrganizationID,
		&auth.TeamID,
		&auth.KeyAlias,
		&auth.TokenHash,
		&models,
		&auth.Spend,
		&maxBudget,
		&rpmLimit,
		&tpmLimit,
		&expiresAt,
		&auth.Blocked,
	)
	if err != nil {
		return nil, err
	}

	auth.Role = "virtual_key"
	auth.AllowedModels = models
	auth.MaxBudget = maxBudget
	auth.RPMLimit = rpmLimit
	auth.TPMLimit = tpmLimit
	auth.ExpiresAt = expiresAt
	return &auth, nil
}

// CreateKey creates a tenant-scoped virtual key.
func (s *Store) CreateKey(ctx context.Context, params platform.CreateKeyParams) (*platform.CreatedKey, error) {
	rawKey := "sk-" + strings.ReplaceAll(uuid.NewString(), "-", "")
	tokenHash := hashKey(rawKey)
	keyID := uuid.NewString()
	masked := maskKey(rawKey)

	var expiresAt *time.Time
	if params.ExpiresIn != "" {
		duration, err := parseDuration(params.ExpiresIn)
		if err != nil {
			return nil, err
		}
		exp := time.Now().Add(duration)
		expiresAt = &exp
	}

	_, err := s.pool.Exec(ctx, `
		INSERT INTO verification_tokens (
			id, tenant_id, token_hash, key_alias, key_name, models, max_budget,
			rpm_limit, tpm_limit, team_id, organization_id, expires, spend, blocked
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NULLIF($10, ''), NULLIF($11, ''), $12, 0, false)
	`, keyID, params.TenantID, tokenHash, params.KeyAlias, masked, params.Models, params.MaxBudget,
		params.RPMLimit, params.TPMLimit, params.TeamID, params.OrgID, expiresAt)
	if err != nil {
		return nil, err
	}

	return &platform.CreatedKey{
		ID:        keyID,
		Key:       rawKey,
		KeyAlias:  params.KeyAlias,
		MaskedKey: masked,
		TenantID:  params.TenantID,
		Role:      params.Role,
		Models:    params.Models,
		MaxBudget: params.MaxBudget,
		RPMLimit:  params.RPMLimit,
		TPMLimit:  params.TPMLimit,
		TeamID:    params.TeamID,
		OrgID:     params.OrgID,
		TokenHash: tokenHash,
	}, nil
}

// ListKeys returns created keys for the tenant.
func (s *Store) ListKeys(ctx context.Context, tenantID string) ([]platform.CreatedKey, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, COALESCE(key_alias, ''), COALESCE(key_name, ''), tenant_id::text,
		       COALESCE(models, '{}'), COALESCE(max_budget, NULL), COALESCE(rpm_limit, NULL),
		       COALESCE(tpm_limit, NULL), COALESCE(team_id::text, ''), COALESCE(organization_id::text, ''),
		       COALESCE(spend, 0), COALESCE(blocked, false), COALESCE(expires, NULL), token_hash
		FROM verification_tokens
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []platform.CreatedKey
	for rows.Next() {
		var item platform.CreatedKey
		var models []string
		if err := rows.Scan(&item.ID, &item.KeyAlias, &item.MaskedKey, &item.TenantID, &models,
			&item.MaxBudget, &item.RPMLimit, &item.TPMLimit, &item.TeamID, &item.OrgID,
			&item.Spend, &item.Blocked, &item.ExpiresAt, &item.TokenHash); err != nil {
			return nil, err
		}
		item.Models = models
		item.Role = "virtual_key"
		keys = append(keys, item)
	}
	return keys, rows.Err()
}

// CreateDeployment stores a model deployment for a tenant.
func (s *Store) CreateDeployment(ctx context.Context, dep platform.Deployment) (*platform.Deployment, error) {
	if dep.ID == "" {
		dep.ID = uuid.NewString()
	}
	params, err := json.Marshal(map[string]interface{}{
		"model":       dep.Model,
		"api_key":     dep.APIKey,
		"api_base":    dep.APIBase,
		"api_version": dep.APIVersion,
		"region_name": dep.RegionName,
		"rpm":         dep.RPM,
		"tpm":         dep.TPM,
		"weight":      dep.Weight,
		"timeout":     dep.TimeoutSec,
	})
	if err != nil {
		return nil, err
	}
	info, err := json.Marshal(dep.ModelInfo)
	if err != nil {
		return nil, err
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO model_deployments (
			id, tenant_id, model_name, litellm_provider, litellm_model, litellm_params, model_info, rpm, tpm, timeout
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, dep.ID, nullUUID(dep.TenantID), dep.ModelName, dep.Provider, dep.Model, params, info, dep.RPM, dep.TPM, dep.TimeoutSec)
	if err != nil {
		return nil, err
	}

	return &dep, nil
}

// ListDeployments returns tenant-specific deployments.
func (s *Store) ListDeployments(ctx context.Context, tenantID string) ([]platform.Deployment, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, COALESCE(tenant_id::text, ''), model_name, litellm_provider, litellm_model,
		       litellm_params, COALESCE(model_info, '{}')
		FROM model_deployments
		WHERE tenant_id IS NULL OR tenant_id = $1
		ORDER BY created_at ASC
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deps []platform.Deployment
	for rows.Next() {
		var dep platform.Deployment
		var paramsRaw []byte
		var infoRaw []byte
		if err := rows.Scan(&dep.ID, &dep.TenantID, &dep.ModelName, &dep.Provider, &dep.Model, &paramsRaw, &infoRaw); err != nil {
			return nil, err
		}
		var params map[string]interface{}
		if err := json.Unmarshal(paramsRaw, &params); err == nil {
			dep.APIBase = toString(params["api_base"])
			dep.APIKey = toString(params["api_key"])
			dep.APIVersion = toString(params["api_version"])
			dep.RegionName = toString(params["region_name"])
			dep.RPM = toInt64(params["rpm"])
			dep.TPM = toInt64(params["tpm"])
			dep.Weight = toFloat64(params["weight"])
			dep.TimeoutSec = toFloat64(params["timeout"])
		}
		if len(infoRaw) > 0 {
			_ = json.Unmarshal(infoRaw, &dep.ModelInfo)
		}
		deps = append(deps, dep)
	}
	return deps, rows.Err()
}

// GetWallet retrieves or creates the tenant wallet.
func (s *Store) GetWallet(ctx context.Context, tenantID string) (*platform.Wallet, error) {
	var wallet platform.Wallet
	err := s.pool.QueryRow(ctx, `
		INSERT INTO wallets (tenant_id, balance_irr, pending_irr)
		VALUES ($1, 0, 0)
		ON CONFLICT (tenant_id) DO UPDATE SET tenant_id = EXCLUDED.tenant_id
		RETURNING wallet_id::text, tenant_id::text, balance_irr, pending_irr
	`, tenantID).Scan(&wallet.ID, &wallet.TenantID, &wallet.BalanceIRR, &wallet.PendingIRR)
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

// TopUpWallet records a completed stub top-up.
func (s *Store) TopUpWallet(ctx context.Context, tenantID string, amountIRR float64, metadata map[string]interface{}) (*platform.Transaction, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var walletID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO wallets (tenant_id, balance_irr, pending_irr)
		VALUES ($1, 0, 0)
		ON CONFLICT (tenant_id) DO UPDATE SET balance_irr = wallets.balance_irr
		RETURNING wallet_id::text
	`, tenantID).Scan(&walletID); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `UPDATE wallets SET balance_irr = balance_irr + $2, updated_at = NOW() WHERE wallet_id::text = $1`, walletID, amountIRR); err != nil {
		return nil, err
	}

	metaRaw, _ := json.Marshal(metadata)
	var out platform.Transaction
	if err := tx.QueryRow(ctx, `
		INSERT INTO wallet_transactions (
			transaction_id, wallet_id, tenant_id, tx_type, amount_irr, status, description, metadata
		)
		VALUES ($1, $2, $3, 'DEPOSIT', $4, 'COMPLETED', 'stubbed wallet top-up', $5)
		RETURNING transaction_id::text, wallet_id::text, tenant_id::text, tx_type, amount_irr, status, description, created_at
	`, uuid.NewString(), walletID, tenantID, amountIRR, metaRaw).Scan(
		&out.ID, &out.WalletID, &out.TenantID, &out.Type, &out.AmountIRR, &out.Status, &out.Description, &out.CreatedAt,
	); err != nil {
		return nil, err
	}
	out.Metadata = metadata

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &out, nil
}

// DebitWallet debits the tenant wallet for inference usage.
func (s *Store) DebitWallet(ctx context.Context, tenantID string, breakdown platform.UsageCostBreakdown, requestID string) (*platform.Transaction, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var wallet platform.Wallet
	if err := tx.QueryRow(ctx, `
		SELECT wallet_id::text, tenant_id::text, balance_irr, pending_irr
		FROM wallets
		WHERE tenant_id = $1
		FOR UPDATE
	`, tenantID).Scan(&wallet.ID, &wallet.TenantID, &wallet.BalanceIRR, &wallet.PendingIRR); err != nil {
		return nil, err
	}

	if wallet.BalanceIRR < breakdown.TotalCostIRR {
		return nil, errors.New("insufficient wallet balance")
	}

	if _, err := tx.Exec(ctx, `UPDATE wallets SET balance_irr = balance_irr - $2, updated_at = NOW() WHERE wallet_id::text = $1`, wallet.ID, breakdown.TotalCostIRR); err != nil {
		return nil, err
	}

	metaRaw, _ := json.Marshal(map[string]interface{}{
		"request_id": requestID,
		"model":      breakdown.Model,
		"tokens":     breakdown.TotalTokens,
	})

	var txn platform.Transaction
	if err := tx.QueryRow(ctx, `
		INSERT INTO wallet_transactions (
			transaction_id, wallet_id, tenant_id, tx_type, amount_irr, amount_usd, exchange_rate, status, description, metadata
		)
		VALUES ($1, $2, $3, 'CONSUMPTION', $4, $5, $6, 'COMPLETED', 'inference usage', $7)
		RETURNING transaction_id::text, wallet_id::text, tenant_id::text, tx_type, amount_irr, amount_usd, exchange_rate, status, description, created_at
	`, uuid.NewString(), wallet.ID, tenantID, breakdown.TotalCostIRR, breakdown.TotalCostUSD, breakdown.ExchangeRate, metaRaw).Scan(
		&txn.ID, &txn.WalletID, &txn.TenantID, &txn.Type, &txn.AmountIRR, &txn.AmountUSD, &txn.ExchangeRate, &txn.Status, &txn.Description, &txn.CreatedAt,
	); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &txn, nil
}

// GetLatestExchangeRate returns the newest rate snapshot.
func (s *Store) GetLatestExchangeRate(ctx context.Context) (*platform.ExchangeRate, error) {
	var rate platform.ExchangeRate
	err := s.pool.QueryRow(ctx, `
		SELECT rate_id::text, usd_to_irr, source, fetched_at
		FROM exchange_rates
		ORDER BY fetched_at DESC
		LIMIT 1
	`).Scan(&rate.ID, &rate.USDtoIRR, &rate.Source, &rate.FetchedAt)
	if err != nil {
		return nil, err
	}
	return &rate, nil
}

// UpsertExchangeRate stores a rate snapshot.
func (s *Store) UpsertExchangeRate(ctx context.Context, rate platform.ExchangeRate) (*platform.ExchangeRate, error) {
	if rate.ID == "" {
		rate.ID = uuid.NewString()
	}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO exchange_rates (rate_id, usd_to_irr, source, fetched_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING rate_id::text, usd_to_irr, source, fetched_at
	`, rate.ID, rate.USDtoIRR, rate.Source).Scan(&rate.ID, &rate.USDtoIRR, &rate.Source, &rate.FetchedAt)
	if err != nil {
		return nil, err
	}
	return &rate, nil
}

// LogRequest writes the normalized request log entry.
func (s *Store) LogRequest(ctx context.Context, entry platform.RequestLogEntry) error {
	reqRaw, _ := json.Marshal(entry.RequestBody)
	respRaw, _ := json.Marshal(entry.ResponseBody)
	_, err := s.pool.Exec(ctx, `
		INSERT INTO request_logs (request_id, token_hash, tenant_id, model, request_body, response_body, status, cost, response_time_ms)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, entry.RequestID, emptyToNil(entry.TokenHash), nullUUID(entry.TenantID), entry.Model, reqRaw, respRaw, entry.Status, entry.CostUSD, entry.ResponseTimeMS)
	return err
}

// LogSpend writes the spend log entry.
func (s *Store) LogSpend(ctx context.Context, log platform.SpendLog) error {
	metaRaw, _ := json.Marshal(log.RequestMetadata)
	_, err := s.pool.Exec(ctx, `
		INSERT INTO spend_logs (
			request_id, token_hash, tenant_id, organization_id, team_id, model, input_tokens, output_tokens,
			total_tokens, cost_usd, cost_irr, response_time_ms, status, request_metadata,
			completion_start_time, completion_end_time
		)
		VALUES ($1, $2, $3, NULLIF($4, '')::uuid, NULLIF($5, '')::uuid, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`, log.RequestID, emptyToNil(log.TokenHash), nullUUID(log.TenantID), log.OrganizationID, log.TeamID, log.Model, log.InputTokens,
		log.OutputTokens, log.TotalTokens, log.CostUSD, log.CostIRR, log.ResponseTimeMS, log.Status, metaRaw, log.CompletionStart, log.CompletionEnd)
	return err
}

func tenantSlugFromHost(host string) string {
	host = strings.TrimSpace(host)
	if host == "" {
		return ""
	}
	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		host = parsedHost
	}
	if ip := net.ParseIP(host); ip != nil {
		return "default"
	}
	parts := strings.Split(host, ".")
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func maskKey(key string) string {
	if len(key) <= 8 {
		return key
	}
	return key[:5] + "..." + key[len(key)-4:]
}

func parseDuration(input string) (time.Duration, error) {
	if strings.HasSuffix(input, "d") {
		n, err := strconv.Atoi(strings.TrimSuffix(input, "d"))
		if err != nil {
			return 0, err
		}
		return time.Duration(n) * 24 * time.Hour, nil
	}
	return time.ParseDuration(input)
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	default:
		return fmt.Sprintf("%v", v)
	}
}

func toInt64(v interface{}) int64 {
	switch val := v.(type) {
	case float64:
		return int64(val)
	case int64:
		return val
	case int:
		return int64(val)
	case json.Number:
		out, _ := val.Int64()
		return out
	default:
		return 0
	}
}

func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int64:
		return float64(val)
	case int:
		return float64(val)
	case json.Number:
		out, _ := val.Float64()
		return out
	default:
		return 0
	}
}

func emptyToNil(v string) interface{} {
	if v == "" {
		return nil
	}
	return v
}

func nullUUID(v string) interface{} {
	if v == "" {
		return nil
	}
	return v
}

func hashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return fmt.Sprintf("%x", hash[:])
}
