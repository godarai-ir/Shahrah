package domain_test

import (
	"strings"
	"testing"

	"github.com/Delpak/shahrah/services/rag/internal/domain"
	"github.com/Delpak/shahrah/services/rag/internal/tokenization"
)

func TestChunkDocumentPreservesPersianStructure(t *testing.T) {
	t.Parallel()

	counter, err := tokenization.NewCL100K()
	if err != nil {
		t.Fatalf("create tokenizer: %v", err)
	}

	page := 2
	document := domain.SourceDocument{Blocks: []domain.SourceBlock{
		{Text: "شرایط بازگشت", Heading: "شرایط بازگشت", Page: &page, Kind: domain.BlockHeading},
		{
			Text:    "کالای خریدار مي\u200cرود و مبلغ ۱۲٫۵٪ بدون تغییر نمایشی بازگردانده می\u200cشود.",
			Heading: "شرایط بازگشت", Page: &page, Kind: domain.BlockParagraph,
		},
	}}

	chunks, err := domain.ChunkDocument(document, 32, 5, counter)
	if err != nil {
		t.Fatalf("chunk document: %v", err)
	}

	if len(chunks) == 0 {
		t.Fatal("expected at least one chunk")
	}

	foundParagraph := false

	for _, chunk := range chunks {
		if chunk.TokenCount > 32 {
			t.Fatalf("chunk exceeds token target: %d", chunk.TokenCount)
		}

		if chunk.Page == nil || *chunk.Page != page {
			t.Fatalf("page was not preserved: %#v", chunk.Page)
		}

		if chunk.Heading != "شرایط بازگشت" {
			t.Fatalf("heading was not preserved: %q", chunk.Heading)
		}

		if strings.Contains(chunk.Text, "مي\u200cرود") {
			foundParagraph = true

			if !strings.Contains(chunk.IndexText, "میرود") || !strings.Contains(chunk.IndexText, "12.5%") {
				t.Fatalf("Persian index form was not canonicalized: %q", chunk.IndexText)
			}
		}
	}

	if !foundParagraph {
		t.Fatal("source-faithful Persian paragraph was not preserved")
	}
}

func TestChunkDocumentSplitsOversizedPersianTerm(t *testing.T) {
	t.Parallel()

	counter, err := tokenization.NewCL100K()
	if err != nil {
		t.Fatalf("create tokenizer: %v", err)
	}

	input := strings.Repeat("آ", 200)
	document := domain.PlainTextDocument(input)

	chunks, err := domain.ChunkDocument(document, 12, 2, counter)
	if err != nil {
		t.Fatalf("chunk oversized term: %v", err)
	}

	if len(chunks) < 2 {
		t.Fatalf("expected an oversized term to be split, got %d chunks", len(chunks))
	}

	var reconstructed strings.Builder

	for _, chunk := range chunks {
		if chunk.TokenCount > 12 {
			t.Fatalf("chunk exceeds token target: %d", chunk.TokenCount)
		}

		reconstructed.WriteString(chunk.Text)
	}

	if reconstructed.String() != input {
		t.Fatal("splitting an oversized term changed visible source text")
	}
}
