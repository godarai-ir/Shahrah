package httpapi

import (
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/require"
)

func TestLastOpenAIUserMessageSupportsTextParts(t *testing.T) {
	t.Parallel()

	messages := []openAIChatMessage{
		{Role: "system", Content: "ignored"},
		{Role: "user", Content: "older question"},
		{Role: "assistant", Content: "older answer"},
		{Role: "user", Content: []any{
			map[string]any{"type": "text", "text": "  بخش اول  "},
			map[string]any{"type": "image_url", "image_url": "ignored"},
			map[string]any{"type": "text", "text": "بخش دوم"},
		}},
	}

	message, err := lastOpenAIUserMessage(messages)
	require.NoError(t, err)
	require.Equal(t, "بخش اول\nبخش دوم", message)
}

func TestLastOpenAIUserMessageRequiresUserText(t *testing.T) {
	t.Parallel()

	_, err := lastOpenAIUserMessage([]openAIChatMessage{{Role: "assistant", Content: "answer"}})
	require.ErrorContains(t, err, "user message")
}

func TestOpenAISessionIDUsesBoundedStableIdentifier(t *testing.T) {
	t.Parallel()

	require.Equal(t, "openai_chat-42", openAISessionID(openAIChatCompletionRequest{
		ChatID: "chat-42", User: "user-1", Metadata: map[string]any{"conversation_id": "conversation-42"},
	}))
	require.Empty(t, openAISessionID(openAIChatCompletionRequest{User: string(make([]byte, 201))}))
}

func TestAnswerWithSourcesAppendsValidatedCitationList(t *testing.T) {
	t.Parallel()

	page := 7
	answer := answerWithSources("پاسخ مستند [1].", []platform.Citation{{
		ID: "1", Title: "راهنمای محصول", Page: &page,
	}})

	require.Contains(t, answer, "پاسخ مستند [1].")
	require.Contains(t, answer, "### Sources")
	require.Contains(t, answer, "[1] راهنمای محصول — page 7")
}
