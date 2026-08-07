package storage

import (
	"context"
	"errors"

	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/jackc/pgx/v5"
)

func (s *PlatformServiceAccountRepository) GetPlatformServiceAccountByTokenHash(ctx context.Context, tokenHash string) (*auth.PlatformServiceAccountCredential, error) {
	var credential auth.PlatformServiceAccountCredential

	err := s.pool.QueryRow(ctx, `
		SELECT credential_id, account_name, scopes
		FROM platform_service_account_credentials
		WHERE token_hash = $1
		  AND active = true
		  AND revoked_at IS NULL
		  AND (expires_at IS NULL OR expires_at > NOW())
		LIMIT 1
	`, tokenHash).Scan(&credential.CredentialID, &credential.AccountName, &credential.Scopes)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, auth.ErrInvalidPlatformServiceCredential
	}

	if err != nil {
		return nil, err
	}

	if _, err := s.pool.Exec(ctx, `
		UPDATE platform_service_account_credentials
		SET last_used_at = NOW()
		WHERE credential_id = $1
		  AND (last_used_at IS NULL OR last_used_at < NOW() - INTERVAL '5 minutes')
	`, credential.CredentialID); err != nil {
		return nil, err
	}

	return &credential, nil
}
