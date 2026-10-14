package chatprovisioning

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/oidc"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

const (
	testTenantID = "11111111-1111-4111-8111-111111111111"
	testTemplate = "https://{slug}.chat.example.com"
)

type fakeStore struct {
	work      []Work
	completed []Outcome
	claimErr  error
	// signal receives one value per recorded outcome, for tests that run the loop.
	signal chan struct{}
}

func (s *fakeStore) ClaimDue(context.Context, int, time.Duration) ([]Work, error) {
	return s.work, s.claimErr
}

func (s *fakeStore) Complete(_ context.Context, outcome Outcome) error {
	s.completed = append(s.completed, outcome)

	if s.signal != nil {
		s.signal <- struct{}{}
	}

	return nil
}

type fakeClients struct {
	clients  []oidc.Client
	created  []oidc.CreateClientRequest
	updated  []oidc.UpdateClientRequest
	rotated  int
	listErr  error
	secretOf int
}

func (f *fakeClients) ListClients(context.Context, string) ([]oidc.Client, error) {
	return f.clients, f.listErr
}

func (f *fakeClients) CreateClient(_ context.Context, tenantID, _ string, request oidc.CreateClientRequest) (*oidc.ClientWithSecret, error) {
	f.created = append(f.created, request)
	f.secretOf++

	client := oidc.Client{
		ID: "row-1", TenantID: tenantID, ClientID: "gdc_new", Name: request.Name, Active: true,
		RedirectURIs: request.RedirectURIs, PostLogoutRedirectURIs: request.PostLogoutRedirectURIs,
	}

	return &oidc.ClientWithSecret{Client: client, ClientSecret: fmt.Sprintf("gds_created_%d", f.secretOf)}, nil
}

func (f *fakeClients) UpdateClient(_ context.Context, _, _ string, request oidc.UpdateClientRequest) (*oidc.Client, error) {
	f.updated = append(f.updated, request)

	client := f.clients[0]
	client.RedirectURIs = *request.RedirectURIs
	client.PostLogoutRedirectURIs = *request.PostLogoutRedirectURIs
	client.Active = true

	return &client, nil
}

func (f *fakeClients) RotateClientSecret(_ context.Context, _, id string) (*oidc.ClientWithSecret, error) {
	f.rotated++

	return &oidc.ClientWithSecret{Client: oidc.Client{ID: id, ClientID: "gdc_existing"}, ClientSecret: "gds_rotated"}, nil
}

type fakeChat struct {
	registered []RegisterRequest
	statuses   []string
	result     Registration
	err        error
	statusErr  error
}

func (c *fakeChat) Register(_ context.Context, _ string, request RegisterRequest) (Registration, error) {
	c.registered = append(c.registered, request)

	return c.result, c.err
}

func (c *fakeChat) SetStatus(_ context.Context, _, status string) error {
	c.statuses = append(c.statuses, status)

	return c.statusErr
}

type harness struct {
	store   *fakeStore
	clients *fakeClients
	chat    *fakeChat
	logs    *observer.ObservedLogs
	run     *Provisioner
}

func newHarness(t *testing.T, work ...Work) *harness {
	t.Helper()

	core, logs := observer.New(zap.DebugLevel)
	h := &harness{
		store:   &fakeStore{work: work},
		clients: &fakeClients{},
		chat:    &fakeChat{result: Registration{Ready: true}},
		logs:    logs,
	}

	provisioner, err := New(Dependencies{
		Store: h.store, Clients: h.clients, Chat: h.chat, ChatURLTemplate: testTemplate,
		Interval: time.Hour, BatchSize: 5, Logger: zap.New(core),
	})
	require.NoError(t, err)

	h.run = provisioner

	return h
}

func newTenant(slug string) Work {
	return Work{TenantID: testTenantID, Slug: slug, Status: "active"}
}

func (h *harness) reconcile(t *testing.T) Outcome {
	t.Helper()

	handled, err := h.run.Reconcile(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, handled)
	require.Len(t, h.store.completed, 1)

	return h.store.completed[0]
}

func TestReconcileCreatesTheClientAndDeliversTheSecretForANewTenant(t *testing.T) {
	t.Parallel()

	h := newHarness(t, newTenant("acme"))

	outcome := h.reconcile(t)

	require.Len(t, h.clients.created, 1)
	require.Equal(t, ClientName, h.clients.created[0].Name)
	require.Equal(t, []string{"https://acme.chat.example.com/oauth/openid/callback"}, h.clients.created[0].RedirectURIs)
	require.Equal(t, []string{"https://acme.chat.example.com/"}, h.clients.created[0].PostLogoutRedirectURIs)
	require.Equal(t, "librechat", h.clients.created[0].CreditProduct)
	require.Zero(t, h.clients.rotated)
	require.Equal(t, []RegisterRequest{{Slug: "acme", ClientID: "gdc_new", ClientSecret: "gds_created_1"}}, h.chat.registered)
	require.Equal(t, []string{"active"}, h.chat.statuses)
	require.Equal(t, Outcome{
		TenantID: testTenantID, State: StateReady, SyncedSlug: "acme", SyncedStatus: "active",
		OIDCClientID: "gdc_new", CredentialsDelivered: true,
	}, outcome)
}

func TestReconcileDoesNotResendTheSecretOnceDelivered(t *testing.T) {
	t.Parallel()

	work := newTenant("acme")
	work.CredentialsDelivered = true
	work.SyncedSlug = "acme"
	work.SyncedStatus = "active"

	h := newHarness(t, work)
	h.clients.clients = []oidc.Client{{
		ID: "row-1", ClientID: "gdc_existing", Name: ClientName, Active: true,
		RedirectURIs:           []string{"https://acme.chat.example.com/oauth/openid/callback"},
		PostLogoutRedirectURIs: []string{"https://acme.chat.example.com/"},
	}}

	outcome := h.reconcile(t)

	require.Empty(t, h.clients.created)
	require.Zero(t, h.clients.rotated)
	require.Empty(t, h.clients.updated)
	require.Empty(t, h.chat.statuses, "the status was already in sync")
	require.Equal(t, []RegisterRequest{{Slug: "acme", ClientID: "gdc_existing"}}, h.chat.registered)
	require.Equal(t, StateReady, outcome.State)
}

func TestReconcileRotatesTheSecretForAClientTheChatNeverReceived(t *testing.T) {
	t.Parallel()

	h := newHarness(t, newTenant("acme"))
	h.clients.clients = []oidc.Client{{
		ID: "row-1", ClientID: "gdc_existing", Name: ClientName, Active: true,
		RedirectURIs:           []string{"https://acme.chat.example.com/oauth/openid/callback"},
		PostLogoutRedirectURIs: []string{"https://acme.chat.example.com/"},
	}}

	outcome := h.reconcile(t)

	require.Empty(t, h.clients.created, "a client with the same name already exists")
	require.Equal(t, 1, h.clients.rotated)
	require.Equal(t, "gds_rotated", h.chat.registered[0].ClientSecret)
	require.True(t, outcome.CredentialsDelivered)
}

func TestReconcileMovesTheClientURIsWhenTheSlugChanges(t *testing.T) {
	t.Parallel()

	work := newTenant("acme-new")
	work.CredentialsDelivered = true
	work.SyncedSlug = "acme"
	work.SyncedStatus = "active"

	h := newHarness(t, work)
	h.clients.clients = []oidc.Client{{
		ID: "row-1", ClientID: "gdc_existing", Name: ClientName, Active: true,
		RedirectURIs:           []string{"https://acme.chat.example.com/oauth/openid/callback"},
		PostLogoutRedirectURIs: []string{"https://acme.chat.example.com/"},
	}}

	outcome := h.reconcile(t)

	require.Len(t, h.clients.updated, 1)
	require.Equal(t, []string{"https://acme-new.chat.example.com/oauth/openid/callback"}, *h.clients.updated[0].RedirectURIs)
	require.Equal(t, "acme-new", h.chat.registered[0].Slug)
	require.Empty(t, h.chat.registered[0].ClientSecret)
	require.Equal(t, "acme-new", outcome.SyncedSlug)
}

func TestReconcileSuspendsAndReactivatesTheChatWithTheTenant(t *testing.T) {
	t.Parallel()

	suspended := newTenant("acme")
	suspended.Status = "suspended"
	suspended.CredentialsDelivered = true
	suspended.SyncedSlug = "acme"
	suspended.SyncedStatus = "active"

	h := newHarness(t, suspended)
	h.clients.clients = []oidc.Client{{
		ID: "row-1", ClientID: "gdc_existing", Name: ClientName, Active: true,
		RedirectURIs:           []string{"https://acme.chat.example.com/oauth/openid/callback"},
		PostLogoutRedirectURIs: []string{"https://acme.chat.example.com/"},
	}}

	outcome := h.reconcile(t)

	require.Equal(t, []string{"suspended"}, h.chat.statuses)
	require.Equal(t, "suspended", outcome.SyncedStatus)
}

func TestReconcileKeepsWaitingWhenTheChatIsNotReady(t *testing.T) {
	t.Parallel()

	h := newHarness(t, newTenant("acme"))
	h.chat.result = Registration{Ready: false, Reason: "DISCOVERY_FAILED"}

	outcome := h.reconcile(t)

	require.Equal(t, StatePending, outcome.State)
	require.Contains(t, outcome.Err, "DISCOVERY_FAILED")
	require.Equal(t, retryBase, outcome.RetryAfter)
	require.True(t, outcome.CredentialsDelivered, "the chat holds the secret even though it is not ready")
	require.Equal(t, "acme", outcome.SyncedSlug)
}

func TestReconcileBacksOffAfterRepeatedFailures(t *testing.T) {
	t.Parallel()

	work := newTenant("acme")
	work.Attempts = 3

	h := newHarness(t, work)
	h.chat.err = &APIError{Status: 503}

	outcome := h.reconcile(t)

	require.Equal(t, StatePending, outcome.State)
	require.Equal(t, 8*retryBase, outcome.RetryAfter)
	require.NotEmpty(t, outcome.Err)
}

func TestReconcileAsksForTheSecretAgainWhenTheChatLostTheTenant(t *testing.T) {
	t.Parallel()

	work := newTenant("acme")
	work.CredentialsDelivered = true
	work.SyncedSlug = "acme"
	work.SyncedStatus = "active"

	h := newHarness(t, work)
	h.clients.clients = []oidc.Client{{
		ID: "row-1", ClientID: "gdc_existing", Name: ClientName, Active: true,
		RedirectURIs:           []string{"https://acme.chat.example.com/oauth/openid/callback"},
		PostLogoutRedirectURIs: []string{"https://acme.chat.example.com/"},
	}}
	h.chat.err = &APIError{Status: 400, Code: "INVALID_OIDC"}

	outcome := h.reconcile(t)

	require.False(t, outcome.CredentialsDelivered, "the next pass rotates and sends a secret")
	require.Equal(t, StatePending, outcome.State)
}

func TestReconcileSkipsTenantsThatCannotHaveAChat(t *testing.T) {
	t.Parallel()

	h := newHarness(t, newTenant("default"))

	outcome := h.reconcile(t)

	require.Equal(t, StateSkipped, outcome.State)
	require.Empty(t, h.clients.created)
	require.Empty(t, h.chat.registered)
}

func TestReconcileRecordsAClientDirectoryFailureWithoutCallingTheChat(t *testing.T) {
	t.Parallel()

	h := newHarness(t, newTenant("acme"))
	h.clients.listErr = errors.New("database is down")

	outcome := h.reconcile(t)

	require.Equal(t, StatePending, outcome.State)
	require.Contains(t, outcome.Err, "database is down")
	require.Empty(t, h.chat.registered)
}

func TestReconcileNeverLogsOrStoresTheSecret(t *testing.T) {
	t.Parallel()

	h := newHarness(t, newTenant("acme"))
	h.chat.err = errors.New("connection refused")

	outcome := h.reconcile(t)

	require.NotContains(t, outcome.Err, "gds_created")

	for _, entry := range h.logs.All() {
		require.NotContains(t, fmt.Sprint(entry.Message, entry.ContextMap()), "gds_created")
	}
}

func TestReconcileStillReleasesTheLeaseWhenCancelled(t *testing.T) {
	t.Parallel()

	h := newHarness(t, newTenant("acme"))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	handled, err := h.run.Reconcile(ctx)

	require.NoError(t, err)
	require.Equal(t, 1, handled)
	require.Len(t, h.store.completed, 1)
}

func TestReconcileReportsAClaimFailure(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	h.store.claimErr = errors.New("db down")

	_, err := h.run.Reconcile(context.Background())

	require.ErrorContains(t, err, "db down")
}

func TestBackoffDoublesUpToTheCap(t *testing.T) {
	t.Parallel()

	require.Equal(t, retryBase, backoff(0))
	require.Equal(t, retryBase, backoff(1))
	require.Equal(t, 2*retryBase, backoff(2))
	require.Equal(t, 4*retryBase, backoff(3))
	require.Equal(t, time.Hour, backoff(50))
}

func TestNewRejectsAnUnusableSetup(t *testing.T) {
	t.Parallel()

	valid := Dependencies{
		Store: &fakeStore{}, Clients: &fakeClients{}, Chat: &fakeChat{},
		ChatURLTemplate: testTemplate, Interval: time.Second, BatchSize: 1,
	}

	for name, mutate := range map[string]func(*Dependencies){
		"no store":         func(d *Dependencies) { d.Store = nil },
		"no clients":       func(d *Dependencies) { d.Clients = nil },
		"no chat":          func(d *Dependencies) { d.Chat = nil },
		"no placeholder":   func(d *Dependencies) { d.ChatURLTemplate = "https://chat.example.com" },
		"two placeholders": func(d *Dependencies) { d.ChatURLTemplate = "https://{slug}.{slug}.example.com" },
		"no interval":      func(d *Dependencies) { d.Interval = 0 },
		"no batch":         func(d *Dependencies) { d.BatchSize = 0 },
	} {
		deps := valid

		mutate(&deps)

		_, err := New(deps)
		require.Error(t, err, name)
	}
}

func TestRunReconcilesAtOnceAndStopsWhenCancelled(t *testing.T) {
	t.Parallel()

	h := newHarness(t, newTenant("acme"))
	h.store.signal = make(chan struct{}, 1)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		defer close(done)

		h.run.Run(ctx)
	}()

	select {
	case <-h.store.signal:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not reconcile at once")
	}

	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not stop after cancellation")
	}
}
