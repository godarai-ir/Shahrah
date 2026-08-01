package proxy

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRewriteXAIFileReferencesEnforcesTenantAccountOwnership(t *testing.T) {
	resources := &responseOperationResources{mappings: map[string]platform.ProviderNativeResourceMapping{
		"file_public": {
			PublicID: "file_public", TenantID: "tenant_a", Provider: "xai",
			ProviderAccountID: "acct_a", ResourceKind: openAIFileResourceKind,
			UpstreamID: "file_upstream", Status: "active",
		},
	}}
	body, err := rewriteXAIFileReferences(context.Background(), resources, "tenant_a", "acct_a", []byte(`{
		"input":[{"type":"input_image","file_id":"file_public"}],
		"tools":[{"type":"function","parameters":{"properties":{"file_id":{"type":"string","default":"schema-example"}}}}]
	}`))
	require.NoError(t, err)

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &decoded))
	input := decoded["input"].([]interface{})[0].(map[string]interface{})
	assert.Equal(t, "file_upstream", input["file_id"])

	parameters := decoded["tools"].([]interface{})[0].(map[string]interface{})["parameters"].(map[string]interface{})
	properties := parameters["properties"].(map[string]interface{})
	assert.Equal(t, "schema-example", properties["file_id"].(map[string]interface{})["default"])

	_, err = rewriteXAIFileReferences(context.Background(), resources, "tenant_a", "acct_other", []byte(`{"image":{"file_id":"file_public"}}`))
	require.ErrorIs(t, err, errXAIFileNotOwned)

	_, err = rewriteXAIFileReferences(context.Background(), resources, "tenant_b", "acct_a", []byte(`{"image":{"file_id":"file_public"}}`))
	require.ErrorIs(t, err, errXAIFileNotOwned)
}

func TestRewriteXAIFileReferencesRequiresRoutingStateOnlyWhenReferenced(t *testing.T) {
	body := []byte(`{"model":"grok-4.5","input":"hello"}`)
	rewritten, err := rewriteXAIFileReferences(context.Background(), nil, "tenant_a", "acct_a", body)
	require.NoError(t, err)
	assert.Equal(t, string(body), string(rewritten))

	_, err = rewriteXAIFileReferences(context.Background(), nil, "tenant_a", "acct_a", []byte(`{"image":{"file_id":"file_public"}}`))
	require.ErrorIs(t, err, errXAIFileRoutingUnavailable)
}
