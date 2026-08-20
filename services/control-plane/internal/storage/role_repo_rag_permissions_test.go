package storage

import "testing"

func TestAdministrativeSystemRolesIncludeManagedRAGPermissions(t *testing.T) {
	t.Parallel()

	required := []string{
		"knowledge_base:list",
		"knowledge_base:manage",
		"knowledge_base:query",
		"chatbot:list",
		"chatbot:manage",
		"chatbot:publish",
		"chatbot:analytics",
	}

	for _, role := range []string{"proxy_admin", "tenant_admin"} {
		permissions := make(map[string]struct{}, len(systemRolePermissions[role]))
		for _, permission := range systemRolePermissions[role] {
			permissions[permission] = struct{}{}
		}

		for _, permission := range required {
			if _, ok := permissions[permission]; !ok {
				t.Errorf("system role %q is missing managed RAG permission %q", role, permission)
			}
		}
	}
}
