package oidc

import (
	"context"
	"strings"
)

// ListClients returns a tenant's registered clients.
func (s *Service) ListClients(ctx context.Context, tenantID string) ([]Client, error) {
	return s.clients.ListOIDCClients(ctx, tenantID)
}

// CreateClient registers a confidential client and returns its secret once.
func (s *Service) CreateClient(ctx context.Context, tenantID, actor string, request CreateClientRequest) (*ClientWithSecret, error) {
	name, err := normalizeClientName(request.Name)
	if err != nil {
		return nil, err
	}

	redirectURIs, err := normalizeRedirectURIs(request.RedirectURIs, true)
	if err != nil {
		return nil, err
	}

	logoutURIs, err := normalizeRedirectURIs(request.PostLogoutRedirectURIs, false)
	if err != nil {
		return nil, err
	}

	product, err := normalizeCreditProduct(request.CreditProduct)
	if err != nil {
		return nil, err
	}

	clientID, err := randomClientID()
	if err != nil {
		return nil, err
	}

	secret, err := randomToken("gds_")
	if err != nil {
		return nil, err
	}

	created, err := s.clients.CreateOIDCClient(ctx, Client{
		TenantID:               tenantID,
		ClientID:               clientID,
		Name:                   name,
		SecretHash:             hashToken(secret),
		RedirectURIs:           redirectURIs,
		PostLogoutRedirectURIs: logoutURIs,
		Active:                 true,
		CreditProduct:          product,
		CreatedBy:              strings.TrimSpace(actor),
	})
	if err != nil {
		return nil, err
	}

	return &ClientWithSecret{Client: *created, ClientSecret: secret}, nil
}

// UpdateClient changes a client's name, redirect URIs, or active flag.
func (s *Service) UpdateClient(ctx context.Context, tenantID, id string, request UpdateClientRequest) (*Client, error) {
	if request.Name != nil {
		name, err := normalizeClientName(*request.Name)
		if err != nil {
			return nil, err
		}

		request.Name = &name
	}

	if request.RedirectURIs != nil {
		uris, err := normalizeRedirectURIs(*request.RedirectURIs, true)
		if err != nil {
			return nil, err
		}

		request.RedirectURIs = &uris
	}

	if request.PostLogoutRedirectURIs != nil {
		uris, err := normalizeRedirectURIs(*request.PostLogoutRedirectURIs, false)
		if err != nil {
			return nil, err
		}

		request.PostLogoutRedirectURIs = &uris
	}

	if request.CreditProduct != nil {
		product, err := normalizeCreditProduct(*request.CreditProduct)
		if err != nil {
			return nil, err
		}

		request.CreditProduct = &product
	}

	return s.clients.UpdateOIDCClient(ctx, tenantID, id, request)
}

// RotateClientSecret replaces a client's secret and returns the new one once.
func (s *Service) RotateClientSecret(ctx context.Context, tenantID, id string) (*ClientWithSecret, error) {
	secret, err := randomToken("gds_")
	if err != nil {
		return nil, err
	}

	client, err := s.clients.RotateOIDCClientSecret(ctx, tenantID, id, hashToken(secret))
	if err != nil {
		return nil, err
	}

	return &ClientWithSecret{Client: *client, ClientSecret: secret}, nil
}

// DeleteClient removes a client together with its pending requests and refresh tokens.
func (s *Service) DeleteClient(ctx context.Context, tenantID, id string) error {
	return s.clients.DeleteOIDCClient(ctx, tenantID, id)
}
