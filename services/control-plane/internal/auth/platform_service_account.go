package auth

import (
	"context"
	"errors"
)

const (
	PlatformScopeRead                            = "platform:read"
	PlatformScopeCatalogWrite                    = "platform:catalog:write"
	PlatformScopeTenantWrite                     = "platform:tenant:write"
	PlatformScopeProviderCredentialRead          = "platform:provider-credentials:read-metadata"
	PlatformScopeProviderCredentialCreateVersion = "platform:provider-credentials:create-version"
	PlatformScopeProviderCredentialTest          = "platform:provider-credentials:test"
	PlatformScopeProviderCredentialActivate      = "platform:provider-credentials:activate"
	PlatformScopeProviderCredentialDisable       = "platform:provider-credentials:disable"
	PlatformScopeSupportTicketRead               = "platform:support-tickets:read"
	PlatformScopeSupportTicketWrite              = "platform:support-tickets:write"
	PlatformScopeAll                             = "platform:*"
)

var ErrInvalidPlatformServiceCredential = errors.New("invalid platform service credential")

type PlatformServiceAccountCredential struct {
	CredentialID string
	AccountName  string
	Scopes       []string
}

type PlatformServiceAccountStore interface {
	GetPlatformServiceAccountByTokenHash(ctx context.Context, tokenHash string) (*PlatformServiceAccountCredential, error)
}

func HasPlatformScope(scopes []string, required string) bool {
	for _, scope := range scopes {
		if scope == required || scope == PlatformScopeAll {
			return true
		}
	}

	return false
}
