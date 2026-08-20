package app

import (
	"context"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/rag/internal/store"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestAnswerRejectsPublishedProfileMismatch(t *testing.T) {
	t.Parallel()

	knowledgeBaseID := uuid.NewString()
	chatbot := store.PublishedChatbot{
		Chatbot: platform.Chatbot{
			ID: uuid.NewString(), TenantID: uuid.NewString(), AnswerModel: "answer",
			KnowledgeBaseIDs: []string{knowledgeBaseID},
		},
		Version: platform.ChatbotPublication{RAGProfileRevision: "different-profile"},
	}
	service, err := New(runtimeStoreStub{}, &answerProviderStub{answers: []string{"unused"}})
	require.NoError(t, err)

	_, err = service.Answer(
		context.Background(), chatbot, platform.ChatbotResponseRequest{Message: "question"},
	)
	require.ErrorIs(t, err, store.ErrProfileMismatch)
}
