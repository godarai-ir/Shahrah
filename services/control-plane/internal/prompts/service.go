package prompts

import (
	"context"
	"errors"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

// Repository is the persistence port consumed by the prompt application.
type Repository interface {
	CreatePrompt(ctx context.Context, tenantID string, tenantSlug string, req platform.CreatePromptRequest, actor string) (*platform.PromptVersion, error)
	CreatePromptVersion(ctx context.Context, tenantID string, tenantSlug string, promptID string, req platform.CreatePromptVersionRequest, actor string) (*platform.PromptVersion, error)
	ListPrompts(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.Prompt, int, error)
	GetPrompt(ctx context.Context, tenantID string, promptID string) (*platform.Prompt, error)
	DeletePrompt(ctx context.Context, tenantID string, promptID string) error
	ListPromptVersions(ctx context.Context, tenantID string, promptID string, limit int, offset int) ([]platform.PromptVersion, int, error)
	GetPromptVersion(ctx context.Context, tenantID string, versionID string) (*platform.PromptVersion, error)
	GetPromptVersionByFQN(ctx context.Context, tenantID string, fqn string) (*platform.PromptVersion, error)
}

// Service owns prompt use cases independently from HTTP and PostgreSQL.
type Service struct {
	repository Repository
}

// NewService creates a prompt application service.
func NewService(repository Repository) (*Service, error) {
	if repository == nil {
		return nil, ErrRepositoryUnavailable
	}

	return &Service{repository: repository}, nil
}

// CreatePrompt validates and persists the first immutable prompt version.
func (s *Service) CreatePrompt(ctx context.Context, tenantID string, tenantSlug string, req platform.CreatePromptRequest, actor string) (*platform.PromptVersion, error) {
	if err := ValidateManifest(req.Manifest); err != nil {
		return nil, err
	}

	return s.repository.CreatePrompt(ctx, tenantID, tenantSlug, req, actor)
}

// CreatePromptVersion validates and persists a new immutable prompt version.
func (s *Service) CreatePromptVersion(ctx context.Context, tenantID string, tenantSlug string, promptID string, req platform.CreatePromptVersionRequest, actor string) (*platform.PromptVersion, error) {
	if err := ValidateManifest(req.Manifest); err != nil {
		return nil, err
	}

	return s.repository.CreatePromptVersion(ctx, tenantID, tenantSlug, promptID, req, actor)
}

// ListPrompts returns a tenant-scoped prompt page.
func (s *Service) ListPrompts(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.Prompt, int, error) {
	return s.repository.ListPrompts(ctx, tenantID, query, limit, offset)
}

// GetPrompt returns one tenant-scoped prompt.
func (s *Service) GetPrompt(ctx context.Context, tenantID string, promptID string) (*platform.Prompt, error) {
	return s.repository.GetPrompt(ctx, tenantID, promptID)
}

// DeletePrompt removes one tenant-scoped prompt.
func (s *Service) DeletePrompt(ctx context.Context, tenantID string, promptID string) error {
	return s.repository.DeletePrompt(ctx, tenantID, promptID)
}

// ListPromptVersions returns immutable versions for one prompt.
func (s *Service) ListPromptVersions(ctx context.Context, tenantID string, promptID string, limit int, offset int) ([]platform.PromptVersion, int, error) {
	return s.repository.ListPromptVersions(ctx, tenantID, promptID, limit, offset)
}

// GetPromptVersion returns one tenant-scoped prompt version.
func (s *Service) GetPromptVersion(ctx context.Context, tenantID string, versionID string) (*platform.PromptVersion, error) {
	return s.repository.GetPromptVersion(ctx, tenantID, versionID)
}

// GetPromptVersionByFQN validates and resolves a tenant-scoped prompt FQN.
func (s *Service) GetPromptVersionByFQN(ctx context.Context, tenantID string, fqn string) (*platform.PromptVersion, error) {
	if _, err := ParseFQN(fqn); err != nil {
		return nil, err
	}

	return s.repository.GetPromptVersionByFQN(ctx, tenantID, fqn)
}

// RenderPromptVersion resolves and renders a prompt version in one use case.
func (s *Service) RenderPromptVersion(ctx context.Context, tenantID string, versionID string, req platform.RenderPromptVersionRequest) (*platform.RenderPromptVersionResponse, error) {
	version, err := s.repository.GetPromptVersion(ctx, tenantID, versionID)
	if err != nil {
		return nil, err
	}

	rendered, err := Render(*version, req.Variables, req.Messages, req.Model)
	if err != nil {
		return nil, err
	}

	return &rendered, nil
}

// IsUnavailable reports whether prompt persistence cannot serve the use case.
func IsUnavailable(err error) bool {
	return errors.Is(err, ErrStorageUnavailable) || errors.Is(err, ErrRepositoryUnavailable)
}
