package evaluation

import (
	"context"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

type evaluationClientStub struct{}

func (evaluationClientStub) Retrieve(_ context.Context, testCase Case) (platform.RetrievalResponse, error) {
	if !testCase.Answerable {
		return platform.RetrievalResponse{Chunks: []platform.RetrievalChunk{}}, nil
	}

	return platform.RetrievalResponse{Chunks: []platform.RetrievalChunk{{
		Title: testCase.ExpectedDocumentTitles[0], Text: "مهلت هفت روز است.",
	}}}, nil
}

func (evaluationClientStub) Answer(_ context.Context, testCase Case) (platform.ChatbotResponse, error) {
	if !testCase.Answerable {
		answer := "اطلاعات کافی ندارم."
		if testCase.ExpectedLanguage == "en" {
			answer = "I do not have enough information."
		}

		return platform.ChatbotResponse{Answer: answer, Refusal: true, Citations: []platform.Citation{}}, nil
	}

	return platform.ChatbotResponse{
		Answer: "مهلت بازگشت هفت روز است [1].",
		Citations: []platform.Citation{{
			ID: "1", DocumentID: "document", Title: testCase.ExpectedDocumentTitles[0], Quote: "مهلت هفت روز است.",
		}},
	}, nil
}

func TestRunPassesCompletePersianGate(t *testing.T) {
	t.Parallel()

	cases := []Case{
		{
			ID: "answerable", Mode: "both", Query: "مهلت بازگشت چقدر است؟",
			KnowledgeBaseKeys: []string{"fixture"}, PublicBotKey: "bot", Answerable: true,
			ExpectedDocumentTitles: []string{"راهنما"}, ExpectedLanguage: "fa", RequiredTerms: []string{"هفت روز"},
		},
		{
			ID: "unanswerable", Mode: "both", Query: "مدیرعامل کیست؟",
			KnowledgeBaseKeys: []string{"fixture"}, PublicBotKey: "bot", Answerable: false,
			ExpectedLanguage: "fa",
		},
	}

	report, err := Run(context.Background(), evaluationClientStub{}, cases)
	if err != nil {
		t.Fatalf("run evaluation: %v", err)
	}

	if !report.Passed {
		t.Fatalf("expected gates to pass: %#v", report)
	}

	if report.CitationValidity != 1 || report.RefusalAccuracy != 1 || report.GroundedAccuracy != 1 {
		t.Fatalf("unexpected metrics: %#v", report)
	}
}
