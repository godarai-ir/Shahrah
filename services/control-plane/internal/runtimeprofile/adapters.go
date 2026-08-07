package runtimeprofile

import (
	"context"
	"sort"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

type VirtualModelStore struct {
	manager *Manager
}

func NewVirtualModelStore(manager *Manager) *VirtualModelStore {
	return &VirtualModelStore{manager: manager}
}

func (s *VirtualModelStore) GetVirtualModelByAlias(ctx context.Context, tenantID string, alias string) (*platform.VirtualModel, error) {
	profile, err := s.manager.GetSnapshot(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	vm, ok := profile.VirtualModels[strings.TrimSpace(alias)]
	if !ok {
		return nil, nil
	}

	out := cloneVirtualModel(vm)

	return &out, nil
}

type GuardrailResolver struct {
	manager *Manager
}

func NewGuardrailResolver(manager *Manager) *GuardrailResolver {
	return &GuardrailResolver{manager: manager}
}

func (r *GuardrailResolver) ListEnabledTenantGuardrailsBySlugs(ctx context.Context, tenantID string, slugs []string) ([]platform.TenantGuardrail, error) {
	profile, err := r.manager.GetSnapshot(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	out := make([]platform.TenantGuardrail, 0, len(slugs))
	seen := map[string]bool{}

	for _, slug := range slugs {
		slug = strings.TrimSpace(slug)
		if slug == "" || seen[slug] {
			continue
		}

		seen[slug] = true
		if guardrail, ok := profile.Guardrails[slug]; ok && guardrail.Enabled {
			out = append(out, cloneGuardrail(guardrail))
		}
	}

	return out, nil
}

type DeploymentStore struct {
	manager      *Manager
	authResolver interface {
		ResolveProviderAccountAuth(
			ctx context.Context,
			tenantID string,
			modelFQN string,
		) (*platform.ResolvedProviderAuth, error)
	}
}

func NewDeploymentStore(
	manager *Manager,
	resolvers ...interface {
		ResolveProviderAccountAuth(
			ctx context.Context,
			tenantID string,
			modelFQN string,
		) (*platform.ResolvedProviderAuth, error)
	},
) *DeploymentStore {
	store := &DeploymentStore{manager: manager}
	if len(resolvers) > 0 {
		store.authResolver = resolvers[0]
	}

	return store
}

func (s *DeploymentStore) ResolveProviderAccountAuth(
	ctx context.Context,
	tenantID string,
	modelFQN string,
) (*platform.ResolvedProviderAuth, error) {
	if s.authResolver == nil {
		apiKey, err := s.ResolveProviderAccountAPIKey(ctx, tenantID, modelFQN)
		if err != nil || apiKey == "" {
			return nil, err
		}

		return &platform.ResolvedProviderAuth{
			Mode:   string(platform.ProviderAuthModeAPIKey),
			APIKey: apiKey,
		}, nil
	}

	return s.authResolver.ResolveProviderAccountAuth(ctx, tenantID, modelFQN)
}

func (s *DeploymentStore) ListDeployments(ctx context.Context, tenantID string) ([]platform.Deployment, error) {
	profile, err := s.manager.GetSnapshot(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	return cloneDeployments(profile.Deployments), nil
}

// DeploymentSnapshot returns the immutable deployment slice owned by the active runtime profile.
// Callers must not mutate the returned slice or any nested values.
func (s *DeploymentStore) DeploymentSnapshot(
	ctx context.Context,
	tenantID string,
) ([]platform.Deployment, int64, error) {
	profile, err := s.manager.GetSnapshot(ctx, tenantID)
	if err != nil {
		return nil, 0, err
	}

	return profile.Deployments, profile.Version, nil
}

func (s *DeploymentStore) ListVirtualModels(ctx context.Context, tenantID string) ([]platform.VirtualModel, error) {
	profile, err := s.manager.GetSnapshot(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	out := make([]platform.VirtualModel, 0, len(profile.VirtualModels))
	for _, vm := range profile.VirtualModels {
		out = append(out, cloneVirtualModel(vm))
	}

	sort.Slice(out, func(i, j int) bool {
		if !out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].CreatedAt.Before(out[j].CreatedAt)
		}

		return out[i].Alias < out[j].Alias
	})

	return out, nil
}

func (s *DeploymentStore) ResolveProviderAccountAPIKey(ctx context.Context, tenantID, modelFQN string) (string, error) {
	profile, err := s.manager.GetSnapshot(ctx, tenantID)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(profile.ProviderAPIKeys[strings.TrimSpace(modelFQN)]), nil
}

type EntitlementResolver struct {
	manager *Manager
}

func NewEntitlementResolver(manager *Manager) *EntitlementResolver {
	return &EntitlementResolver{manager: manager}
}

func (r *EntitlementResolver) ResolveEffectiveEntitlementLimits(ctx context.Context, tenantID string) (*platform.EffectiveEntitlementLimits, error) {
	profile, err := r.manager.GetSnapshot(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	if profile.Entitlements == nil {
		return nil, ErrProfileUnavailable
	}

	out := *profile.Entitlements

	out.Features = cloneInterfaceMap(profile.Entitlements.Features)
	if profile.Entitlements.Subscription != nil {
		sub := *profile.Entitlements.Subscription
		sub.PriceSnapshot = cloneInterfaceMap(sub.PriceSnapshot)
		sub.Metadata = cloneInterfaceMap(sub.Metadata)
		out.Subscription = &sub
	}

	return &out, nil
}
