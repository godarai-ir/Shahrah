package integration_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestIntegration_TenantProviderCredentialPlaintextStorageIsRemoved(
	t *testing.T,
) {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set")
	}

	pool, err := pgxpool.New(context.Background(), databaseURL)
	require.NoError(t, err)

	defer pool.Close()

	var plaintextColumns int
	require.NoError(t, pool.QueryRow(context.Background(), `
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_schema = 'public'
		  AND table_name = 'tenant_provider_accounts'
		  AND column_name = 'credential_data'
	`).Scan(&plaintextColumns))
	require.Zero(t, plaintextColumns, "plaintext credential column must not exist")

	var inconsistentAPIKeyAccounts int
	require.NoError(t, pool.QueryRow(context.Background(), `
		SELECT COUNT(*)
		FROM tenant_provider_accounts
		WHERE auth_mode = 'api_key'
		  AND use_platform_managed_credentials = false
		  AND has_credential = true
		  AND active_credential_version_id IS NULL
	`).Scan(&inconsistentAPIKeyAccounts))
	require.Zero(
		t,
		inconsistentAPIKeyAccounts,
		"user-owned API-key accounts must reference an encrypted credential version",
	)

	var encryptedVersionColumns int
	require.NoError(t, pool.QueryRow(context.Background(), `
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_schema = 'public'
		  AND table_name = 'tenant_provider_account_credential_versions'
		  AND column_name IN (
		      'ciphertext', 'data_nonce', 'wrapped_dek', 'wrap_nonce',
		      'kek_version', 'fingerprint'
		  )
	`).Scan(&encryptedVersionColumns))
	require.Equal(t, 6, encryptedVersionColumns)
}
