package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	prommetrics "github.com/Delpak/shahrah/services/control-plane/internal/observability/prometheus"
	"github.com/Delpak/shahrah/services/control-plane/internal/platformcredentials"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

var ErrTenantCredentialEncryptionUnavailable = errors.New(
	"tenant_credential_encryption_unavailable",
)

const maxTenantCredentialCacheEntries = 1024

type tenantCredentialCacheEntry struct {
	apiKey    string
	expiresAt time.Time
}

func (s *Store) writeTenantCredentialVersion(
	ctx context.Context,
	tx pgx.Tx,
	tenantID string,
	accountID string,
	provider string,
	authMode string,
	secrets map[string]string,
) (string, error) {
	if s.tenantCredentialStore == nil {
		return "", ErrTenantCredentialEncryptionUnavailable
	}

	var version int64
	if err := tx.QueryRow(ctx, `
		SELECT COALESCE(MAX(version), 0) + 1
		FROM tenant_provider_account_credential_versions
		WHERE account_id = $1
	`, accountID).Scan(&version); err != nil {
		return "", err
	}

	versionID := uuid.NewString()
	binding := platformcredentials.Binding{
		TenantID:            tenantID,
		AccountID:           accountID,
		ProviderSlug:        strings.ToLower(strings.TrimSpace(provider)),
		CredentialVersionID: versionID,
		CredentialVersion:   version,
		AuthMode:            strings.TrimSpace(authMode),
		SchemaVersion:       1,
	}
	encrypted, err := s.tenantCredentialStore.Encrypt(binding, secrets)
	if err != nil {
		return "", err
	}
	verified, err := s.tenantCredentialStore.Decrypt(encrypted)
	fingerprintCheck, fingerprintErr := s.tenantCredentialStore.Encrypt(
		binding,
		secrets,
	)
	if err != nil ||
		fingerprintErr != nil ||
		!s.tenantCredentialStore.FingerprintsEqual(
			encrypted.Fingerprint,
			fingerprintCheck.Fingerprint,
		) ||
		verified["api_key"] != secrets["api_key"] {
		clear(verified)
		return "", platformcredentials.ErrDecryptFailed
	}
	clear(verified)

	if _, err := tx.Exec(ctx, `
		UPDATE tenant_provider_account_credential_versions
		SET status = 'retiring'
		WHERE account_id = $1 AND status = 'active'
	`, accountID); err != nil {
		return "", err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO tenant_provider_account_credential_versions (
			credential_version_id, tenant_id, account_id, provider, auth_mode,
			schema_version, version, ciphertext, data_nonce, wrapped_dek,
			wrap_nonce, kek_version, fingerprint, display_hint, status
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, 'active'
		)
	`, versionID, tenantID, accountID, binding.ProviderSlug, binding.AuthMode,
		binding.SchemaVersion, version, encrypted.Ciphertext, encrypted.DataNonce,
		encrypted.WrappedDEK, encrypted.WrapNonce, encrypted.KEKVersion,
		encrypted.Fingerprint, encrypted.DisplayHint); err != nil {
		return "", err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE tenant_provider_accounts
		SET active_credential_version_id = $3,
		    credential_hint = $4,
		    credential_rotation_required = false,
		    credential_data = '{}'::jsonb,
		    has_credential = true
		WHERE tenant_id = $1 AND account_id = $2
	`, tenantID, accountID, versionID, encrypted.DisplayHint); err != nil {
		return "", err
	}

	return encrypted.DisplayHint, nil
}

type LegacyTenantCredentialMigrationResult struct {
	Migrated  int
	Remaining int
}

// MigrateLegacyTenantProviderCredentials is resumable and concurrency-safe.
// Rows are locked with SKIP LOCKED so multiple workers can process disjoint
// batches. Migrated credentials remain rotation-required.
func (s *Store) MigrateLegacyTenantProviderCredentials(
	ctx context.Context,
	batchSize int,
) (LegacyTenantCredentialMigrationResult, error) {
	if s.tenantCredentialStore == nil {
		return LegacyTenantCredentialMigrationResult{},
			ErrTenantCredentialEncryptionUnavailable
	}
	if batchSize <= 0 || batchSize > 1000 {
		batchSize = 100
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return LegacyTenantCredentialMigrationResult{}, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	type legacyRow struct {
		tenantID  string
		accountID string
		provider  string
		authMode  string
		apiKey    string
	}
	rows, err := tx.Query(ctx, `
		SELECT tenant_id::text, account_id::text, provider, auth_mode,
		       credential_data->>'api_key'
		FROM tenant_provider_accounts
		WHERE active_credential_version_id IS NULL
		  AND COALESCE(credential_data->>'api_key', '') <> ''
		ORDER BY created_at
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`, batchSize)
	if err != nil {
		return LegacyTenantCredentialMigrationResult{}, err
	}
	batch := make([]legacyRow, 0, batchSize)
	for rows.Next() {
		var row legacyRow
		if err := rows.Scan(
			&row.tenantID,
			&row.accountID,
			&row.provider,
			&row.authMode,
			&row.apiKey,
		); err != nil {
			rows.Close()
			return LegacyTenantCredentialMigrationResult{}, err
		}
		batch = append(batch, row)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return LegacyTenantCredentialMigrationResult{}, err
	}
	rows.Close()

	for _, row := range batch {
		if _, err := s.writeTenantCredentialVersion(
			ctx,
			tx,
			row.tenantID,
			row.accountID,
			row.provider,
			row.authMode,
			map[string]string{"api_key": strings.TrimSpace(row.apiKey)},
		); err != nil {
			return LegacyTenantCredentialMigrationResult{}, err
		}
		if _, err := tx.Exec(ctx, `
			UPDATE tenant_provider_accounts
			SET credential_rotation_required = true
			WHERE tenant_id = $1 AND account_id = $2
		`, row.tenantID, row.accountID); err != nil {
			return LegacyTenantCredentialMigrationResult{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return LegacyTenantCredentialMigrationResult{}, err
	}

	var remaining int
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM tenant_provider_accounts
		WHERE active_credential_version_id IS NULL
		  AND COALESCE(credential_data->>'api_key', '') <> ''
	`).Scan(&remaining); err != nil {
		return LegacyTenantCredentialMigrationResult{}, err
	}
	return LegacyTenantCredentialMigrationResult{
		Migrated:  len(batch),
		Remaining: remaining,
	}, nil
}

func (s *Store) clearTenantCredentialVersion(
	ctx context.Context,
	tx pgx.Tx,
	tenantID string,
	accountID string,
) error {
	if _, err := tx.Exec(ctx, `
		UPDATE tenant_provider_accounts
		SET active_credential_version_id = NULL,
		    credential_hint = NULL,
		    credential_data = '{}'::jsonb,
		    has_credential = false
		WHERE tenant_id = $1 AND account_id = $2
	`, tenantID, accountID); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `
		UPDATE tenant_provider_account_credential_versions
		SET status = 'revoked', revoked_at = NOW()
		WHERE account_id = $1 AND status IN ('active', 'retiring')
	`, accountID)
	return err
}

func (s *Store) resolveTenantProviderAccountAPIKey(
	ctx context.Context,
	tenantID string,
	accountID string,
) (string, error) {
	type encryptedRow struct {
		versionID   string
		version     int64
		provider    string
		authMode    string
		schema      int
		ciphertext  []byte
		dataNonce   []byte
		wrappedDEK  []byte
		wrapNonce   []byte
		kekVersion  string
		fingerprint string
	}

	var row encryptedRow
	err := s.pool.QueryRow(ctx, `
		SELECT v.credential_version_id::text, v.version, v.provider, v.auth_mode,
		       v.schema_version, v.ciphertext, v.data_nonce, v.wrapped_dek,
		       v.wrap_nonce, v.kek_version, v.fingerprint
		FROM tenant_provider_accounts a
		JOIN tenant_provider_account_credential_versions v
		  ON v.credential_version_id = a.active_credential_version_id
		 AND v.tenant_id = a.tenant_id
		 AND v.account_id = a.account_id
		 AND v.status = 'active'
		WHERE a.tenant_id = $1 AND a.account_id = $2
	`, tenantID, accountID).Scan(
		&row.versionID, &row.version, &row.provider, &row.authMode, &row.schema,
		&row.ciphertext, &row.dataNonce, &row.wrappedDEK, &row.wrapNonce,
		&row.kekVersion, &row.fingerprint,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return s.resolveLegacyTenantProviderAccountAPIKey(ctx, tenantID, accountID)
	}
	if err != nil {
		return "", err
	}
	if s.tenantCredentialStore == nil {
		return "", ErrTenantCredentialEncryptionUnavailable
	}

	cacheKey := fmt.Sprintf("%s:%s:%s:%d", tenantID, accountID, row.versionID, row.version)
	now := time.Now()
	s.tenantCredentialCacheMu.Lock()
	cached, ok := s.tenantCredentialCache[cacheKey]
	s.tenantCredentialCacheMu.Unlock()
	if ok && cached.expiresAt.After(now) {
		return cached.apiKey, nil
	}

	value, err, _ := s.tenantCredentialLoadGroup.Do(cacheKey, func() (interface{}, error) {
		record := platformcredentials.EncryptedCredential{
			Binding: platformcredentials.Binding{
				TenantID:            tenantID,
				AccountID:           accountID,
				ProviderSlug:        row.provider,
				CredentialVersionID: row.versionID,
				CredentialVersion:   row.version,
				AuthMode:            row.authMode,
				SchemaVersion:       row.schema,
			},
			Ciphertext:  row.ciphertext,
			DataNonce:   row.dataNonce,
			WrappedDEK:  row.wrappedDEK,
			WrapNonce:   row.wrapNonce,
			KEKVersion:  row.kekVersion,
			Fingerprint: row.fingerprint,
		}
		secrets, decryptErr := s.tenantCredentialStore.Decrypt(record)
		if decryptErr != nil {
			return "", decryptErr
		}
		apiKey := strings.TrimSpace(secrets["api_key"])
		clear(secrets)
		if apiKey == "" {
			return "", platformcredentials.ErrDecryptFailed
		}

		s.tenantCredentialCacheMu.Lock()
		if len(s.tenantCredentialCache) >= maxTenantCredentialCacheEntries {
			s.tenantCredentialCache = make(map[string]tenantCredentialCacheEntry)
		}
		s.tenantCredentialCache[cacheKey] = tenantCredentialCacheEntry{
			apiKey:    apiKey,
			expiresAt: now.Add(s.tenantCredentialCacheTTL),
		}
		s.tenantCredentialCacheMu.Unlock()
		return apiKey, nil
	})
	if err != nil {
		return "", err
	}
	return value.(string), nil
}

func (s *Store) resolveLegacyTenantProviderAccountAPIKey(
	ctx context.Context,
	tenantID string,
	accountID string,
) (string, error) {
	var apiKey string
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(credential_data->>'api_key', '')
		FROM tenant_provider_accounts
		WHERE tenant_id = $1 AND account_id = $2
		  AND active_credential_version_id IS NULL
	`, tenantID, accountID).Scan(&apiKey)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(apiKey) != "" {
		prommetrics.RecordAnthropicLegacyCredentialRead()
		if s.logger != nil {
			s.logger.Warn(
				"tenant_provider_credential.legacy_read",
				zap.String("provider", "redacted"),
			)
		}
	}
	return strings.TrimSpace(apiKey), nil
}
