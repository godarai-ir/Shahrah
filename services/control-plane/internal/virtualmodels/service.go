package virtualmodels

import (
	"context"
	"errors"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

var (
	errRepositoryRequired = errors.New("virtual models: repository is required")
	// ErrNotFound indicates that a virtual model does not exist in the tenant scope.
	ErrNotFound = errors.New("virtual model not found")
	// ErrInvalidModes indicates an invalid virtual-model routing mode configuration.
	ErrInvalidModes = errors.New("invalid virtual model modes")
)

type deploymentRepository interface {
	CreateDeployment(context.Context, platform.Deployment) (*platform.Deployment, error)
	ListDeployments(context.Context, string) ([]platform.Deployment, error)
}

type virtualModelRepository interface {
	CreateVirtualModel(context.Context, string, platform.CreateVirtualModelRequest, string) (*platform.VirtualModel, error)
	ListVirtualModels(context.Context, string) ([]platform.VirtualModel, error)
	GetVirtualModel(context.Context, string, string) (*platform.VirtualModel, error)
	GetVirtualModelByAlias(context.Context, string, string) (*platform.VirtualModel, error)
	UpdateVirtualModel(context.Context, string, string, platform.UpdateVirtualModelRequest, string) (*platform.VirtualModel, error)
	DeleteVirtualModel(context.Context, string, string) error
}

type principalModelRepository interface {
	ListVirtualModelsForPrincipal(context.Context, string, string, string) ([]platform.VirtualModel, error)
}

// ManagementRepository composes focused persistence ports used by virtual-model management.
type ManagementRepository interface {
	deploymentRepository
	virtualModelRepository
	principalModelRepository
}

// Service is the application boundary for model deployment and virtual-model use cases.
type Service struct {
	deployments deploymentRepository
	models      virtualModelRepository
	principals  principalModelRepository
}

// NewService creates the virtual-model application service.
func NewService(repository ManagementRepository) (*Service, error) {
	if repository == nil {
		return nil, errRepositoryRequired
	}

	return &Service{deployments: repository, models: repository, principals: repository}, nil
}

// CreateDeployment creates a provider deployment available to virtual-model routing.
func (s *Service) CreateDeployment(ctx context.Context, deployment platform.Deployment) (*platform.Deployment, error) {
	return s.deployments.CreateDeployment(ctx, deployment)
}

// ListDeployments lists deployments in a tenant scope.
func (s *Service) ListDeployments(ctx context.Context, tenantID string) ([]platform.Deployment, error) {
	return s.deployments.ListDeployments(ctx, tenantID)
}

// CreateVirtualModel creates a virtual model after application-level input has been accepted.
func (s *Service) CreateVirtualModel(ctx context.Context, tenantID string, request platform.CreateVirtualModelRequest, actor string) (*platform.VirtualModel, error) {
	return s.models.CreateVirtualModel(ctx, tenantID, request, actor)
}

// ListVirtualModels lists virtual models in a tenant scope.
func (s *Service) ListVirtualModels(ctx context.Context, tenantID string) ([]platform.VirtualModel, error) {
	return s.models.ListVirtualModels(ctx, tenantID)
}

// GetVirtualModel resolves a virtual model by identifier.
func (s *Service) GetVirtualModel(ctx context.Context, tenantID, modelID string) (*platform.VirtualModel, error) {
	return s.models.GetVirtualModel(ctx, tenantID, modelID)
}

// GetVirtualModelByAlias resolves a virtual model by its tenant-scoped alias.
func (s *Service) GetVirtualModelByAlias(ctx context.Context, tenantID, alias string) (*platform.VirtualModel, error) {
	return s.models.GetVirtualModelByAlias(ctx, tenantID, alias)
}

// UpdateVirtualModel updates a virtual model.
func (s *Service) UpdateVirtualModel(ctx context.Context, tenantID, modelID string, request platform.UpdateVirtualModelRequest, actor string) (*platform.VirtualModel, error) {
	return s.models.UpdateVirtualModel(ctx, tenantID, modelID, request, actor)
}

// DeleteVirtualModel deletes a virtual model.
func (s *Service) DeleteVirtualModel(ctx context.Context, tenantID, modelID string) error {
	return s.models.DeleteVirtualModel(ctx, tenantID, modelID)
}

// ListVirtualModelsForPrincipal lists virtual models visible to a principal.
func (s *Service) ListVirtualModelsForPrincipal(ctx context.Context, tenantID, userID, teamID string) ([]platform.VirtualModel, error) {
	return s.principals.ListVirtualModelsForPrincipal(ctx, tenantID, userID, teamID)
}
