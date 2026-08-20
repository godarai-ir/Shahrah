package auth

import (
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/require"
)

func TestRAGRouteAuthorizationUsesScopedPermissions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		method      string
		route       string
		permissions []string
		wantError   bool
	}{
		{name: "knowledge base list", method: "GET", route: "/api/v1/knowledge-bases", permissions: []string{"knowledge_base:list"}},
		{name: "knowledge base write denied", method: "POST", route: "/api/v1/knowledge-bases", permissions: []string{"knowledge_base:list"}, wantError: true},
		{name: "chatbot analytics", method: "GET", route: "/api/v1/chatbots/bot/analytics", permissions: []string{"chatbot:analytics"}},
		{name: "chatbot publish", method: "POST", route: "/api/v1/chatbots/bot/publish", permissions: []string{"chatbot:publish"}},
		{name: "chatbot preview", method: "POST", route: "/api/v1/chatbots/bot/preview-responses", permissions: []string{"chatbot:manage"}},
		{name: "chatbot preview denied", method: "POST", route: "/api/v1/chatbots/bot/preview-responses", permissions: []string{"chatbot:list"}, wantError: true},
		{name: "chatbot publish denied", method: "POST", route: "/api/v1/chatbots/bot/publish", permissions: []string{"chatbot:list"}, wantError: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			handled, err := checkAPIV1RAGRouteAuthorization(
				test.method, test.route, &platform.AuthContext{Permissions: test.permissions},
			)
			require.True(t, handled)

			if test.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
