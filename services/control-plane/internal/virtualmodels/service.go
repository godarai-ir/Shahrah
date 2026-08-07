package virtualmodels

import (
	"context"
	"errors"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

var errRepositoryRequired = errors.New("virtual models: repository is required")

// ManagementRepository is the persistence port used by virtual-model management use cases.
type ManagementRepository interface {
	CreateDeployment(context.Context, platform.Deployment) (*platform.Deployment, error)
	ListDeployments(context.Context, string) ([]platform.Deployment, error)
	CreateVirtualModel(context.Context, string, platform.CreateVirtualModelRequest, string) (*platform.VirtualModel, error)
	ListVirtualModels(context.Context, string) ([]platform.VirtualModel, error)
	GetVirtualModel(context.Context, string, string) (*platform.VirtualModel, error)
	GetVirtualModelByAlias(context.Context, string, string) (*platform.VirtualModel, error)
	UpdateVirtualModel(context.Context, string, string, platform.UpdateVirtualModelRequest, string) (*platform.VirtualModel, error)
	DeleteVirtualModel(context.Context, string, string) error
	ListVirtualModelsForPrincipal(context.Context, string, string, string) ([]platform.VirtualModel, error)
}

// Service is the application boundary for model deployment and virtual-model use cases.
type Service struct {
	ManagementRepository
}

// NewService creates the virtual-model application service.
func NewService(repository ManagementRepository) (*Service, error) {
	if repository == nil {
		return nil, errRepositoryRequired
	}

	return &Service{ManagementRepository: repository}, nil
}
