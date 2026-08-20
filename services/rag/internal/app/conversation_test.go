package app

import (
	"testing"

	"github.com/Delpak/shahrah/services/rag/internal/store"
	"github.com/stretchr/testify/require"
)

func TestBoundedConversationKeepsNewestMessages(t *testing.T) {
	t.Parallel()

	messages := []store.ConversationMessage{
		{Role: "user", Content: "one two three"},
		{Role: "assistant", Content: "four five"},
		{Role: "user", Content: "six"},
	}

	result := boundedConversation(messages, 2, 3)
	require.Equal(t, messages[1:], result)
}
