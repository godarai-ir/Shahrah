package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestOriginAllowedUsesExactNormalizedOrigins(t *testing.T) {
	t.Parallel()

	require.True(t, originAllowed("https://shop.example/", []string{"https://shop.example"}))
	require.False(t, originAllowed("https://evil.example", []string{"https://shop.example"}))
	require.False(t, originAllowed("https://shop.example.evil", []string{"https://shop.example"}))
	require.True(t, originAllowed("", nil))
}

func TestValidateChatbotRequestRejectsOriginPaths(t *testing.T) {
	t.Parallel()

	err := validateChatbotRequest(platform.UpsertChatbotRequest{
		Name: "bot", AnswerModel: "model", KnowledgeBaseIDs: []string{"11111111-1111-4111-8111-111111111111"},
		AccessMode: "public", RetentionDays: 30, AllowedOrigins: []string{"https://example.com/path"},
	})
	require.ErrorContains(t, err, "origins")
}

func TestValidateChatbotRequestValidatesLocalePolicy(t *testing.T) {
	t.Parallel()

	base := platform.UpsertChatbotRequest{
		Name: "bot", AnswerModel: "model", KnowledgeBaseIDs: []string{"11111111-1111-4111-8111-111111111111"},
		DefaultLocale: "fa-IR", AnswerLanguage: "query", RetentionDays: 30,
	}
	require.NoError(t, validateChatbotRequest(base))

	base.DefaultLocale = "fa-AF"
	require.ErrorContains(t, validateChatbotRequest(base), "default_locale")

	base.DefaultLocale = "fa-IR"
	base.AnswerLanguage = "auto"
	require.ErrorContains(t, validateChatbotRequest(base), "answer_language")
}

func TestValidateChatbotRequestRejectsDuplicateKnowledgeBasesAndOrigins(t *testing.T) {
	t.Parallel()

	const knowledgeBaseID = "11111111-1111-4111-8111-111111111111"

	request := platform.UpsertChatbotRequest{
		Name: "bot", AnswerModel: "model", KnowledgeBaseIDs: []string{knowledgeBaseID, knowledgeBaseID},
		RetentionDays: 30,
	}
	require.ErrorContains(t, validateChatbotRequest(request), "duplicates")

	request.KnowledgeBaseIDs = []string{knowledgeBaseID}
	request.AllowedOrigins = []string{"https://example.com", "https://example.com/"}
	require.ErrorContains(t, validateChatbotRequest(request), "duplicates")
}

func TestValidateChatMessageRejectsOversizedInput(t *testing.T) {
	t.Parallel()

	require.ErrorContains(t, validateChatMessage(strings.Repeat("پ", platform.RAGMaxQueryCharacters+1)), "exceeds")
}

func TestRuntimeProxyAuthenticationFailsClosed(t *testing.T) {
	t.Parallel()

	const token = "01234567890123456789012345678901"

	h := &handler{dependencies: Dependencies{InternalToken: token}}
	server := fiber.New()
	server.Get("/runtime", h.requireProxy, func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusNoContent)
	})

	unauthorized, err := server.Test(httptest.NewRequest(fiber.MethodGet, "/runtime", http.NoBody))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, unauthorized.StatusCode)
	require.NoError(t, unauthorized.Body.Close())

	request := httptest.NewRequest(fiber.MethodGet, "/runtime", http.NoBody)
	request.Header.Set(headerInternalToken, token)
	authorized, err := server.Test(request)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNoContent, authorized.StatusCode)
	require.NoError(t, authorized.Body.Close())
}
