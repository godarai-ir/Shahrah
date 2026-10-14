package chatprovisioning

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/oidc"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"go.uber.org/zap"
)

const (
	provisionerActor = "chat-provisioner"
	slugPlaceholder  = "{slug}"
	callbackPath     = "/oauth/openid/callback"

	// retryBase is the first retry delay after a failure; it doubles each time up to retryMax.
	retryBase = 15 * time.Second
	retryMax  = time.Hour
	// maxBackoffShift keeps the doubling from overflowing before the cap applies.
	maxBackoffShift = 8
	// defaultLease is how long a claimed tenant stays away from other replicas.
	defaultLease = 5 * time.Minute

	// errInvalidOIDC is the chat's code for a registration that lacks the client secret.
	errInvalidOIDC = "INVALID_OIDC"
)

// Dependencies wires a Provisioner.
type Dependencies struct {
	Store   Store
	Clients ClientDirectory
	Chat    Chat
	// ChatURLTemplate is a tenant's chat origin with a {slug} placeholder.
	ChatURLTemplate string
	Interval        time.Duration
	BatchSize       int
	Logger          *zap.Logger
}

// Provisioner registers each tenant with the chat and keeps the registration current.
type Provisioner struct {
	store    Store
	clients  ClientDirectory
	chat     Chat
	template string
	interval time.Duration
	batch    int
	lease    time.Duration
	logger   *zap.Logger
}

// New validates the dependencies and creates a Provisioner.
func New(deps Dependencies) (*Provisioner, error) {
	if deps.Store == nil || deps.Clients == nil || deps.Chat == nil {
		return nil, errors.New("chat provisioning: store, client directory and chat are required")
	}

	if strings.Count(deps.ChatURLTemplate, slugPlaceholder) != 1 {
		return nil, fmt.Errorf("chat provisioning: chat URL template must contain %s exactly once", slugPlaceholder)
	}

	if deps.Interval <= 0 || deps.BatchSize <= 0 {
		return nil, errors.New("chat provisioning: interval and batch size must be positive")
	}

	logger := deps.Logger
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Provisioner{
		store:    deps.Store,
		clients:  deps.Clients,
		chat:     deps.Chat,
		template: strings.TrimRight(deps.ChatURLTemplate, "/"),
		interval: deps.Interval,
		batch:    deps.BatchSize,
		lease:    defaultLease,
		logger:   logger,
	}, nil
}

// Run reconciles until ctx is canceled: once at once, then every interval.
func (p *Provisioner) Run(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		if _, err := p.Reconcile(ctx); err != nil && ctx.Err() == nil {
			p.logger.Error("chat provisioning pass failed", zap.Error(err))
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// Reconcile handles one batch of tenants that need work and returns how many it handled.
func (p *Provisioner) Reconcile(ctx context.Context) (int, error) {
	work, err := p.store.ClaimDue(ctx, p.batch, p.lease)
	if err != nil {
		return 0, fmt.Errorf("claim tenants for chat provisioning: %w", err)
	}

	for _, item := range work {
		outcome := p.process(ctx, item)

		// A canceled pass must still release its lease, so the write ignores the cancellation.
		if err := p.store.Complete(context.WithoutCancel(ctx), outcome); err != nil {
			return 0, fmt.Errorf("record chat provisioning outcome: %w", err)
		}
	}

	return len(work), nil
}

func (p *Provisioner) process(ctx context.Context, work Work) Outcome {
	outcome := Outcome{
		TenantID:             work.TenantID,
		State:                StatePending,
		SyncedSlug:           work.SyncedSlug,
		SyncedStatus:         work.SyncedStatus,
		OIDCClientID:         work.OIDCClientID,
		CredentialsDelivered: work.CredentialsDelivered,
	}

	if !tenant.ValidateTenantSlug(work.Slug) {
		outcome.State = StateSkipped
		outcome.SyncedSlug = work.Slug
		outcome.SyncedStatus = desiredStatus(work.Status)

		p.logger.Info("tenant cannot have a chat", zap.String("tenant_id", work.TenantID), zap.String("slug", work.Slug))

		return outcome
	}

	if err := p.sync(ctx, work, &outcome); err != nil {
		return p.failed(work, outcome, err)
	}

	return outcome
}

// sync does the work for one tenant and fills outcome. A returned error leaves outcome partly
// filled with whatever already succeeded, so the next pass does not repeat it.
func (p *Provisioner) sync(ctx context.Context, work Work, outcome *Outcome) error {
	client, secret, err := p.ensureClient(ctx, work)
	if err != nil {
		return err
	}

	outcome.OIDCClientID = client.ClientID

	// The secret is shown once, when a client is created or rotated. A client the chat has never
	// received credentials for gets a fresh one, so nothing sensitive is ever stored here.
	if secret == "" && !work.CredentialsDelivered {
		rotated, rotateErr := p.clients.RotateClientSecret(ctx, work.TenantID, client.ID)
		if rotateErr != nil {
			return fmt.Errorf("rotate chat client secret: %w", rotateErr)
		}

		secret = rotated.ClientSecret
	}

	registration, err := p.chat.Register(ctx, work.TenantID, RegisterRequest{
		Slug:         work.Slug,
		ClientID:     client.ClientID,
		ClientSecret: secret,
	})
	if err != nil {
		// The chat lost the tenant (for example a wiped database) and needs the secret again.
		if apiErr, ok := AsAPIError(err); ok && apiErr.Code == errInvalidOIDC && secret == "" {
			outcome.CredentialsDelivered = false
		}

		return fmt.Errorf("register tenant with chat: %w", err)
	}

	if secret != "" {
		outcome.CredentialsDelivered = true
	}

	desired := desiredStatus(work.Status)
	if work.SyncedStatus != desired {
		if err := p.chat.SetStatus(ctx, work.TenantID, desired); err != nil {
			return fmt.Errorf("set chat status: %w", err)
		}
	}

	outcome.SyncedSlug = work.Slug
	outcome.SyncedStatus = desired

	if !registration.Ready {
		return fmt.Errorf("chat is not ready for this tenant yet (%s)", registration.Reason)
	}

	outcome.State = StateReady

	p.logger.Info("tenant chat is ready", zap.String("tenant_id", work.TenantID), zap.String("slug", work.Slug))

	return nil
}

// ensureClient finds or creates the tenant's chat OIDC client and keeps its URIs on the tenant's
// chat origin. The secret is non-empty only when the client was just created.
func (p *Provisioner) ensureClient(ctx context.Context, work Work) (*oidc.Client, string, error) {
	origin := strings.Replace(p.template, slugPlaceholder, work.Slug, 1)
	redirectURIs := []string{origin + callbackPath}
	logoutURIs := []string{origin + "/"}

	clients, err := p.clients.ListClients(ctx, work.TenantID)
	if err != nil {
		return nil, "", fmt.Errorf("list tenant OIDC clients: %w", err)
	}

	for i := range clients {
		if clients[i].Name != ClientName {
			continue
		}

		existing := clients[i]
		if existing.Active && slices.Equal(existing.RedirectURIs, redirectURIs) &&
			slices.Equal(existing.PostLogoutRedirectURIs, logoutURIs) {
			return &existing, "", nil
		}

		active := true

		updated, updateErr := p.clients.UpdateClient(ctx, work.TenantID, existing.ID, oidc.UpdateClientRequest{
			RedirectURIs:           &redirectURIs,
			PostLogoutRedirectURIs: &logoutURIs,
			Active:                 &active,
		})
		if updateErr != nil {
			return nil, "", fmt.Errorf("update chat OIDC client: %w", updateErr)
		}

		return updated, "", nil
	}

	created, err := p.clients.CreateClient(ctx, work.TenantID, provisionerActor, oidc.CreateClientRequest{
		Name:                   ClientName,
		RedirectURIs:           redirectURIs,
		PostLogoutRedirectURIs: logoutURIs,
		CreditProduct:          platform.CreditProductLibreChat,
	})
	if err != nil {
		return nil, "", fmt.Errorf("create chat OIDC client: %w", err)
	}

	return &created.Client, created.ClientSecret, nil
}

// failed turns an error into an outcome that retries later, with a delay that grows per failure.
func (p *Provisioner) failed(work Work, outcome Outcome, err error) Outcome {
	outcome.State = StatePending
	outcome.Err = err.Error()
	outcome.RetryAfter = backoff(work.Attempts + 1)

	fields := []zap.Field{
		zap.String("tenant_id", work.TenantID),
		zap.String("slug", work.Slug),
		zap.Int("attempt", work.Attempts+1),
		zap.Duration("retry_after", outcome.RetryAfter),
		zap.Error(err),
	}

	if apiErr, ok := AsAPIError(err); ok && apiErr.Permanent() {
		p.logger.Error("chat provisioning rejected, needs attention", fields...)

		return outcome
	}

	p.logger.Warn("chat provisioning will retry", fields...)

	return outcome
}

func desiredStatus(tenantStatus string) string {
	if tenantStatus == "active" {
		return chatStatusActive
	}

	return chatStatusSuspended
}

// backoff doubles the delay with each failed attempt up to retryMax.
func backoff(attempt int) time.Duration {
	shift := min(max(attempt-1, 0), maxBackoffShift)

	return min(retryBase<<shift, retryMax)
}
