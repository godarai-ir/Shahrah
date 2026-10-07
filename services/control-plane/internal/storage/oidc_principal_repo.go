package storage

import (
	"context"
	"errors"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apperrors"
	"github.com/Delpak/shahrah/services/control-plane/internal/oidc"
	"github.com/jackc/pgx/v5"
)

// ResolveOIDCUserAuth builds the gateway auth context for a member signed in through OIDC.
// It applies the same user-scoped model policy as a personal access token, so a member reaches
// exactly the models a PAT of theirs would.
func (s *AuthRepository) ResolveOIDCUserAuth(ctx context.Context, tenantID, userID string) (*platform.AuthContext, error) {
	member, err := s.ResolveTenantUserAuth(ctx, tenantID, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) || errors.Is(err, pgx.ErrNoRows) {
			return nil, oidc.ErrIdentityNotFound
		}

		return nil, err
	}

	auth := platform.AuthContext{
		TenantID:    member.TenantID,
		UserID:      member.UserID,
		Role:        deriveKeyRole(""),
		Permissions: append([]string(nil), member.Permissions...),
		KeyType:     oidc.KeyType,
	}

	if err := s.applyRuntimeModelPolicy(ctx, &auth, nil); err != nil {
		return nil, err
	}

	return &auth, nil
}
