package prompts

import (
	"context"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/require"
)

type repositoryStub struct {
	version *platform.PromptVersion
	created bool
}

func (r *repositoryStub) CreatePrompt(_ context.Context, _ string, _ string, _ platform.CreatePromptRequest, _ string) (*platform.PromptVersion, error) {
	r.created = true

	return r.version, nil
}

func (r *repositoryStub) CreatePromptVersion(context.Context, string, string, string, platform.CreatePromptVersionRequest, string) (*platform.PromptVersion, error) {
	return r.version, nil
}

func (r *repositoryStub) ListPrompts(context.Context, string, string, int, int) ([]platform.Prompt, int, error) {
	return nil, 0, nil
}

func (r *repositoryStub) GetPrompt(context.Context, string, string) (*platform.Prompt, error) {
	return nil, ErrNotFound
}

func (r *repositoryStub) DeletePrompt(context.Context, string, string) error {
	return nil
}

func (r *repositoryStub) ListPromptVersions(context.Context, string, string, int, int) ([]platform.PromptVersion, int, error) {
	return nil, 0, nil
}

func (r *repositoryStub) GetPromptVersion(context.Context, string, string) (*platform.PromptVersion, error) {
	return r.version, nil
}

func (r *repositoryStub) GetPromptVersionByFQN(context.Context, string, string) (*platform.PromptVersion, error) {
	return r.version, nil
}

func TestServiceRejectsInvalidManifestBeforePersistence(t *testing.T) {
	t.Parallel()

	repository := &repositoryStub{}
	service, err := NewService(repository)
	require.NoError(t, err)

	_, err = service.CreatePrompt(context.Background(), "tenant", "acme", platform.CreatePromptRequest{}, "actor")
	require.ErrorIs(t, err, ErrInvalidManifest)
	require.False(t, repository.created)
}

func TestServiceRendersResolvedPromptVersion(t *testing.T) {
	t.Parallel()

	repository := &repositoryStub{version: &platform.PromptVersion{
		ID:       "version-id",
		PromptID: "prompt-id",
		FQN:      "chat_prompt:acme/welcome:1",
		Messages: nil,
	}}
	service, err := NewService(repository)
	require.NoError(t, err)

	rendered, err := service.RenderPromptVersion(context.Background(), "tenant", "version-id", platform.RenderPromptVersionRequest{})
	require.NoError(t, err)
	require.Equal(t, "version-id", rendered.PromptVersionID)
}
