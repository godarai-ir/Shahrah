package platformcredentials

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type recordingCoordination struct{ message string }

func (r *recordingCoordination) Publish(_ context.Context, _ string, message string) error {
	r.message = message
	return nil
}
func (r *recordingCoordination) Subscribe(context.Context, string) (<-chan string, error) {
	return nil, nil
}

func TestInvalidationEventContainsMetadataOnly(t *testing.T) {
	client := &recordingCoordination{}
	bus := NewRedisInvalidationBus(client)
	require.NoError(t, bus.PublishPlatformCredentialInvalidation(context.Background(), InvalidationEvent{ProviderSlug: "openai", Environment: "production", ConnectionID: "connection-1", CredentialVersionID: "version-1", Action: "activate"}))
	require.Contains(t, client.message, "connection-1")
	require.False(t, strings.Contains(strings.ToLower(client.message), "api_key"))
	require.NotContains(t, client.message, "sk-secret")
}
