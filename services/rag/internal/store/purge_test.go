package store

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseVectorPurgeObjectKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		objectKey  string
		scope      string
		identifier string
		ok         bool
	}{
		{name: "tenant", objectKey: TenantVectorPurgeObjectKey, scope: "tenant", ok: true},
		{name: "knowledge base", objectKey: KnowledgeBaseVectorPurgeObjectKey("kb-id"), scope: "knowledge_base", identifier: "kb-id", ok: true},
		{name: "document", objectKey: DocumentVectorPurgeObjectKey("document-id"), scope: "document", identifier: "document-id", ok: true},
		{name: "private object", objectKey: "rag/tenant/object", ok: false},
		{name: "empty internal identifier", objectKey: documentVectorPurgePrefix, ok: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			scope, identifier, ok := ParseVectorPurgeObjectKey(test.objectKey)
			require.Equal(t, test.scope, scope)
			require.Equal(t, test.identifier, identifier)
			require.Equal(t, test.ok, ok)
		})
	}
}
