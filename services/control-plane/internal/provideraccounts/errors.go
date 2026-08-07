// Package provideraccounts owns tenant provider-account application contracts.
package provideraccounts

import "errors"

var (
	// ErrInvalidAccessPrincipal indicates an unsupported provider-account principal.
	ErrInvalidAccessPrincipal = errors.New("invalid access principal")
	// ErrInvalidAPIBase indicates an invalid provider API base URL.
	ErrInvalidAPIBase = errors.New("invalid provider api_base")
	// ErrConfigContainsSecret indicates credential material in provider configuration.
	ErrConfigContainsSecret = errors.New("provider_config must not contain credentials")
	// ErrPrivatePricingRequiresCredential indicates missing user-owned credentials.
	ErrPrivatePricingRequiresCredential = errors.New("private pricing requires a provider account with user-owned credentials")
	// ErrDependencyConflict indicates active dependencies on a provider account.
	ErrDependencyConflict = errors.New("provider account has active dependencies")
	// ErrAlreadyExists indicates a conflicting provider-account resource.
	ErrAlreadyExists = errors.New("provider account resource already exists")
	// ErrInvalidPayload indicates a payload rejected by provider-account constraints.
	ErrInvalidPayload = errors.New("invalid provider account payload")
)
