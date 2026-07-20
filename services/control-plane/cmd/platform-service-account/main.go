package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	name := flag.String("name", "", "service account name")
	scopesCSV := flag.String("scopes", "", "comma-separated scopes")
	profile := flag.String("profile", "", "predefined scope profile (support-admin)")
	expiresAtValue := flag.String("expires-at", "", "optional RFC3339 expiry")
	revokeKeyID := flag.String("revoke-key-id", "", "revoke an existing credential instead of creating one")

	flag.Parse()

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		fatal("DATABASE_URL is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		fatal("connect to database: %v", err)
	}
	defer pool.Close()

	if keyID := strings.TrimSpace(*revokeKeyID); keyID != "" {
		result, execErr := pool.Exec(ctx, `
			UPDATE platform_service_account_credentials
			SET active = false, revoked_at = NOW()
			WHERE credential_id = $1 AND revoked_at IS NULL
		`, keyID)
		if execErr != nil {
			fatal("revoke credential: %v", execErr)
		}

		if result.RowsAffected() != 1 {
			fatal("active credential was not found")
		}

		fmt.Printf("Platform service credential %s was revoked.\n", keyID)

		return
	}

	accountName, scopes, err := resolveCredentialConfig(*name, *profile, *scopesCSV)
	if err != nil {
		fatal("%v", err)
	}

	if !validName(accountName) {
		fatal("--name must contain 3-64 lowercase letters, numbers, or hyphens")
	}

	var expiresAt *time.Time

	if value := strings.TrimSpace(*expiresAtValue); value != "" {
		parsed, parseErr := time.Parse(time.RFC3339, value)
		if parseErr != nil || !parsed.After(time.Now()) {
			fatal("--expires-at must be a future RFC3339 timestamp")
		}

		expiresAt = &parsed
	}

	keyID := randomToken(12)
	secret := randomToken(32)
	token := "cp_sa_" + keyID + "." + secret
	hash := sha256.Sum256([]byte(token))

	_, err = pool.Exec(ctx, `
		INSERT INTO platform_service_account_credentials
			(credential_id, account_name, token_hash, scopes, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`, keyID, accountName, hex.EncodeToString(hash[:]), scopes, expiresAt)
	if err != nil {
		fatal("create credential: %v", err)
	}

	fmt.Printf("CONTROL_PLANE_SERVICE_TOKEN=%s\n", token)
	fmt.Fprintln(os.Stderr, "Store this credential now; only its hash is persisted. Re-running creates a rotation credential.")
}

const supportAdminProfile = "support-admin"

var supportAdminScopes = []string{
	auth.PlatformScopeRead,
	auth.PlatformScopeCatalogWrite,
	auth.PlatformScopeTenantWrite,
	auth.PlatformScopeProviderCredentialRead,
	auth.PlatformScopeProviderCredentialCreateVersion,
	auth.PlatformScopeProviderCredentialTest,
	auth.PlatformScopeProviderCredentialActivate,
	auth.PlatformScopeProviderCredentialDisable,
}

func resolveCredentialConfig(name, profile, rawScopes string) (string, []string, error) {
	accountName := strings.TrimSpace(name)
	profile = strings.TrimSpace(profile)
	if profile == "" {
		if accountName == supportAdminProfile || accountName == "platform-support" {
			return "", nil, fmt.Errorf("%s credentials must use --profile=%s", accountName, supportAdminProfile)
		}
		scopes, err := parseScopes(rawScopes)
		return accountName, scopes, err
	}
	if strings.TrimSpace(rawScopes) != "" {
		return "", nil, fmt.Errorf("--profile and --scopes cannot be used together")
	}
	if profile != supportAdminProfile {
		return "", nil, fmt.Errorf("unsupported profile %q", profile)
	}
	if accountName == "" {
		accountName = supportAdminProfile
	}

	scopes, err := parseScopes(strings.Join(supportAdminScopes, ","))
	return accountName, scopes, err
}

func parseScopes(raw string) ([]string, error) {
	allowed := map[string]struct{}{
		auth.PlatformScopeRead: {}, auth.PlatformScopeCatalogWrite: {},
		auth.PlatformScopeTenantWrite: {}, auth.PlatformScopeAll: {},
		auth.PlatformScopeProviderCredentialRead:          {},
		auth.PlatformScopeProviderCredentialCreateVersion: {},
		auth.PlatformScopeProviderCredentialTest:          {},
		auth.PlatformScopeProviderCredentialActivate:      {},
		auth.PlatformScopeProviderCredentialDisable:       {},
	}
	seen := map[string]struct{}{}

	for _, value := range strings.Split(raw, ",") {
		scope := strings.TrimSpace(value)
		if scope == "" {
			continue
		}

		if _, ok := allowed[scope]; !ok {
			return nil, fmt.Errorf("unsupported scope %q", scope)
		}

		seen[scope] = struct{}{}
	}

	if len(seen) == 0 {
		return nil, fmt.Errorf("--scopes must contain at least one supported scope")
	}

	out := make([]string, 0, len(seen))
	for scope := range seen {
		out = append(out, scope)
	}

	sort.Strings(out)

	return out, nil
}

func validName(value string) bool {
	if len(value) < 3 || len(value) > 64 {
		return false
	}

	for _, char := range value {
		if (char < 'a' || char > 'z') && (char < '0' || char > '9') && char != '-' {
			return false
		}
	}

	return true
}

func randomToken(size int) string {
	value := make([]byte, size)
	if _, err := rand.Read(value); err != nil {
		fatal("generate credential: %v", err)
	}

	return base64.RawURLEncoding.EncodeToString(value)
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
