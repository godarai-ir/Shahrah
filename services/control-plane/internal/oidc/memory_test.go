package oidc

import (
	"context"
	"sync"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/google/uuid"
)

// memoryRepository mirrors the storage adapter's contract for service tests.
type memoryRepository struct {
	mu         sync.Mutex
	now        func() time.Time
	clients    map[string]*Client
	requests   map[string]*memoryRequest
	refresh    map[string]*memoryRefresh
	identities map[string]*Identity
}

type memoryRequest struct {
	request       AuthorizationRequest
	codeHash      string
	codeExpiresAt time.Time
	codeUsed      bool
}

type memoryRefresh struct {
	token   RefreshToken
	used    bool
	revoked bool
}

func newMemoryRepository(now func() time.Time) *memoryRepository {
	return &memoryRepository{
		now:        now,
		clients:    map[string]*Client{},
		requests:   map[string]*memoryRequest{},
		refresh:    map[string]*memoryRefresh{},
		identities: map[string]*Identity{},
	}
}

func (m *memoryRepository) GetOIDCClient(_ context.Context, clientID string) (*Client, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	client, ok := m.clients[clientID]
	if !ok {
		return nil, ErrClientNotFound
	}

	copied := *client

	return &copied, nil
}

func (m *memoryRepository) ListOIDCClients(_ context.Context, tenantID string) ([]Client, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	clients := []Client{}

	for _, client := range m.clients {
		if client.TenantID == tenantID {
			clients = append(clients, *client)
		}
	}

	return clients, nil
}

func (m *memoryRepository) CreateOIDCClient(_ context.Context, client Client) (*Client, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.clients[client.ClientID]; exists {
		return nil, ErrClientIDTaken
	}

	client.ID = uuid.NewString()
	client.CreatedAt = m.now()
	client.UpdatedAt = client.CreatedAt
	m.clients[client.ClientID] = &client
	copied := client

	return &copied, nil
}

func (m *memoryRepository) findClient(tenantID, id string) *Client {
	for _, client := range m.clients {
		if client.ID == id && client.TenantID == tenantID {
			return client
		}
	}

	return nil
}

func (m *memoryRepository) UpdateOIDCClient(_ context.Context, tenantID, id string, update UpdateClientRequest) (*Client, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	client := m.findClient(tenantID, id)
	if client == nil {
		return nil, ErrClientNotFound
	}

	if update.Name != nil {
		client.Name = *update.Name
	}

	if update.RedirectURIs != nil {
		client.RedirectURIs = *update.RedirectURIs
	}

	if update.PostLogoutRedirectURIs != nil {
		client.PostLogoutRedirectURIs = *update.PostLogoutRedirectURIs
	}

	if update.Active != nil {
		client.Active = *update.Active
	}

	if update.CreditProduct != nil {
		client.CreditProduct = *update.CreditProduct
	}

	copied := *client

	return &copied, nil
}

func (m *memoryRepository) RotateOIDCClientSecret(_ context.Context, tenantID, id, secretHash string) (*Client, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	client := m.findClient(tenantID, id)
	if client == nil {
		return nil, ErrClientNotFound
	}

	client.SecretHash = secretHash
	copied := *client

	return &copied, nil
}

func (m *memoryRepository) DeleteOIDCClient(_ context.Context, tenantID, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	client := m.findClient(tenantID, id)
	if client == nil {
		return ErrClientNotFound
	}

	delete(m.clients, client.ClientID)

	return nil
}

func (m *memoryRepository) CreateOIDCAuthorizationRequest(_ context.Context, request AuthorizationRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.requests[request.ID] = &memoryRequest{request: request}

	return nil
}

func (m *memoryRepository) GetOIDCAuthorizationRequest(_ context.Context, id string) (*AuthorizationRequest, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	stored, ok := m.requests[id]
	if !ok {
		return nil, ErrRequestNotFound
	}

	copied := stored.request

	return &copied, nil
}

func (m *memoryRepository) AcceptOIDCAuthorizationRequest(_ context.Context, accepted AuthorizationRequest, codeHash string, codeExpiresAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	stored, ok := m.requests[accepted.ID]
	if !ok || stored.request.Status != RequestStatusPending || !m.now().Before(stored.request.ExpiresAt) {
		return ErrRequestNotPending
	}

	stored.request.Status = RequestStatusAccepted
	stored.request.UserID = accepted.UserID
	stored.request.GrantID = accepted.GrantID
	stored.request.AuthTime = accepted.AuthTime
	stored.codeHash = codeHash
	stored.codeExpiresAt = codeExpiresAt

	return nil
}

func (m *memoryRepository) RejectOIDCAuthorizationRequest(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	stored, ok := m.requests[id]
	if !ok || stored.request.Status != RequestStatusPending {
		return ErrRequestNotPending
	}

	stored.request.Status = RequestStatusRejected

	return nil
}

func (m *memoryRepository) ConsumeOIDCAuthorizationCode(_ context.Context, codeHash string) (*AuthorizationRequest, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, stored := range m.requests {
		if stored.codeHash != codeHash || stored.request.Status != RequestStatusAccepted {
			continue
		}

		copied := stored.request
		if stored.codeUsed {
			return &copied, ErrCodeReused
		}

		if !m.now().Before(stored.codeExpiresAt) {
			return nil, ErrCodeNotFound
		}

		stored.codeUsed = true

		return &copied, nil
	}

	return nil, ErrCodeNotFound
}

func (m *memoryRepository) CreateOIDCRefreshToken(_ context.Context, token RefreshToken) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.refresh[token.TokenHash] = &memoryRefresh{token: token}

	return nil
}

func (m *memoryRepository) RotateOIDCRefreshToken(_ context.Context, presentedHash string, next RefreshToken, check func(RefreshToken) error) (*RefreshToken, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	stored, ok := m.refresh[presentedHash]
	if !ok || stored.revoked {
		return nil, ErrRefreshTokenNotFound
	}

	if stored.used {
		m.revokeGrantLocked(stored.token.GrantID)

		return nil, ErrRefreshTokenReused
	}

	if err := check(stored.token); err != nil {
		return nil, err
	}

	next.GrantID = stored.token.GrantID
	next.UserID = stored.token.UserID
	next.Scope = stored.token.Scope
	next.Nonce = stored.token.Nonce
	next.AuthTime = stored.token.AuthTime
	next.GrantExpiresAt = stored.token.GrantExpiresAt
	next.TenantID = stored.token.TenantID
	next.ClientID = stored.token.ClientID

	if next.ExpiresAt.After(next.GrantExpiresAt) {
		next.ExpiresAt = next.GrantExpiresAt
	}

	stored.used = true
	m.refresh[next.TokenHash] = &memoryRefresh{token: next}
	copied := next

	return &copied, nil
}

func (m *memoryRepository) revokeGrantLocked(grantID string) {
	for _, stored := range m.refresh {
		if stored.token.GrantID == grantID {
			stored.revoked = true
		}
	}
}

func (m *memoryRepository) RevokeOIDCGrant(_ context.Context, grantID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.revokeGrantLocked(grantID)

	return nil
}

func (m *memoryRepository) RevokeOIDCUserClientGrants(_ context.Context, tenantID, userID, clientID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, stored := range m.refresh {
		if stored.token.TenantID == tenantID && stored.token.UserID == userID && stored.token.ClientID == clientID {
			stored.revoked = true
		}
	}

	return nil
}

func (m *memoryRepository) GetOIDCIdentity(_ context.Context, tenantID, userID string) (*Identity, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	identity, ok := m.identities[tenantID+":"+userID]
	if !ok {
		return nil, ErrIdentityNotFound
	}

	copied := *identity

	return &copied, nil
}

func (m *memoryRepository) setIdentity(identity Identity) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.identities[identity.TenantID+":"+identity.UserID] = &identity
}

func (m *memoryRepository) activeRefreshTokens(grantID string) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0

	for _, stored := range m.refresh {
		if stored.token.GrantID == grantID && !stored.used && !stored.revoked {
			count++
		}
	}

	return count
}

type memorySessions struct {
	sessions map[string]*Session
}

func (m memorySessions) LookupSession(_ context.Context, sessionID string) (*Session, error) {
	session, ok := m.sessions[sessionID]
	if !ok {
		return nil, nil
	}

	copied := *session

	return &copied, nil
}

type memoryPrincipals struct {
	repo *memoryRepository
}

func (m memoryPrincipals) ResolveOIDCUserAuth(ctx context.Context, tenantID, userID string) (*platform.AuthContext, error) {
	identity, err := m.repo.GetOIDCIdentity(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}

	if !identity.Active {
		return nil, ErrIdentityNotFound
	}

	return &platform.AuthContext{
		TenantID:              tenantID,
		UserID:                userID,
		Role:                  "virtual_key",
		AllowedModels:         []string{"openai:acct:gpt-4o"},
		EnforceModelAllowlist: true,
	}, nil
}
