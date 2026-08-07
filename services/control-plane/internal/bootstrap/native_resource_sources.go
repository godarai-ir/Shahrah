package bootstrap

import (
	"context"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/runtimeprofile"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
)

// nativeResourceSources composes provider resources with runtime deployments.
type nativeResourceSources struct {
	*storage.ProviderResourceRepository
	deployments *runtimeprofile.DeploymentStore
}

func newNativeResourceSources(
	resources *storage.ProviderResourceRepository,
	deployments *runtimeprofile.DeploymentStore,
) nativeResourceSources {
	return nativeResourceSources{
		ProviderResourceRepository: resources,
		deployments:                deployments,
	}
}

func (s nativeResourceSources) ListDeployments(ctx context.Context, tenantID string) ([]platform.Deployment, error) {
	return s.deployments.ListDeployments(ctx, tenantID)
}
