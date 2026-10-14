// Package chatprovisioning registers every tenant with GodarChat (LibreChat): it creates the
// tenant's OIDC client, then hands the client credentials and slug to the chat provisioning API.
// A reconciler drives it, so a tenant created by any path gets a chat without a hook in that path.
package chatprovisioning

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/oidc"
)

// ClientName is the display name of the OIDC client created for a tenant's chat.
const ClientName = "GodarChat"

// Provisioning states stored per tenant.
const (
	// StatePending means the tenant still needs work.
	StatePending = "pending"
	// StateReady means the chat accepted the tenant and its sign-in works.
	StateReady = "ready"
	// StateSkipped means the tenant cannot have a chat (for example a reserved slug).
	StateSkipped = "skipped"
)

// Tenant statuses the chat provisioning API understands.
const (
	chatStatusActive    = "active"
	chatStatusSuspended = "suspended"
)

// Work is one tenant a reconciler pass must bring up to date.
type Work struct {
	TenantID string
	Slug     string
	// Status is the tenant's own status; anything but "active" suspends its chat.
	Status string
	// CredentialsDelivered reports that the chat already holds the client secret.
	CredentialsDelivered bool
	OIDCClientID         string
	// Attempts counts consecutive failures, which drives the retry backoff.
	Attempts     int
	SyncedSlug   string
	SyncedStatus string
}

// Outcome is what a pass learned about one tenant.
type Outcome struct {
	TenantID             string
	State                string
	SyncedSlug           string
	SyncedStatus         string
	OIDCClientID         string
	CredentialsDelivered bool
	// Err is a short message without secrets; empty after a success.
	Err        string
	RetryAfter time.Duration
}

// Store persists provisioning progress and hands out work.
type Store interface {
	// ClaimDue leases up to limit tenants that need work. A lease keeps other replicas away.
	ClaimDue(ctx context.Context, limit int, lease time.Duration) ([]Work, error)
	// Complete records an outcome and releases the lease.
	Complete(ctx context.Context, outcome Outcome) error
}

// ClientDirectory is the part of the OIDC service the provisioner uses.
type ClientDirectory interface {
	ListClients(ctx context.Context, tenantID string) ([]oidc.Client, error)
	CreateClient(ctx context.Context, tenantID, actor string, request oidc.CreateClientRequest) (*oidc.ClientWithSecret, error)
	UpdateClient(ctx context.Context, tenantID, id string, request oidc.UpdateClientRequest) (*oidc.Client, error)
	RotateClientSecret(ctx context.Context, tenantID, id string) (*oidc.ClientWithSecret, error)
}

// RegisterRequest describes a tenant to the chat provisioning API.
type RegisterRequest struct {
	Slug     string
	ClientID string
	// ClientSecret is sent only while the chat does not hold the secret yet.
	ClientSecret string
}

// Registration is the chat's answer to a registration.
type Registration struct {
	// Ready reports that the chat can sign users of this tenant in.
	Ready bool
	// Reason says why it cannot yet, when Ready is false.
	Reason string
}

// Chat is the chat provisioning API.
type Chat interface {
	Register(ctx context.Context, tenantID string, request RegisterRequest) (Registration, error)
	SetStatus(ctx context.Context, tenantID, status string) error
}

// APIError is an error answer from the chat provisioning API.
type APIError struct {
	Status int
	Code   string
}

// Error implements error.
func (e *APIError) Error() string {
	return "chat provisioning API answered " + strconv.Itoa(e.Status) + " " + e.Code
}

// Permanent reports that retrying the same request cannot succeed.
func (e *APIError) Permanent() bool {
	return e.Status >= 400 && e.Status < 500 && e.Status != 408 && e.Status != 429
}

// AsAPIError returns the APIError in err's chain.
func AsAPIError(err error) (*APIError, bool) {
	var apiErr *APIError

	if errors.As(err, &apiErr) {
		return apiErr, true
	}

	return nil, false
}
