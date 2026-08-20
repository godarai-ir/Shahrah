package domain

import (
	"strings"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/require"
)

func TestNormalizePersian(t *testing.T) {
	t.Parallel()

	require.Equal(t, "یک متن فارسی\nخط دوم", NormalizePersian("يك  متن فارسي\r\n خط دوم "))
}

func TestNormalizePersianCultureVariants(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"Arabic keyboard and digits":  "کتابهای 123",
		"spacing around prefix":       "نمیرود",
		"spacing around suffix":       "کتابها",
		"Persian decimal and percent": "12.5%",
		"Persian punctuation":         "سلام, خوب هستید? بله;",
		"invisible controls":          "متن فارسی",
	}

	inputs := map[string]string{
		"Arabic keyboard and digits":  "كِتاب\u200cهای ١٢٣",
		"spacing around prefix":       "نمی رود",
		"spacing around suffix":       "کتاب ها",
		"Persian decimal and percent": "۱۲٫۵٪",
		"Persian punctuation":         "سلام، خوب هستید؟ بله؛",
		"invisible controls":          "متن\u200f فارسی",
	}

	for name, expected := range tests {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, expected, NormalizePersian(inputs[name]))
		})
	}
}

func TestNormalizePersianPreservesDisplayText(t *testing.T) {
	t.Parallel()

	input := "كِتاب\u200cهای ١٢٣"
	normalized := NormalizePersianText(input)

	require.Equal(t, input, normalized.Display)
	require.Equal(t, "کتابهای 123", normalized.Index)
}

func TestChunkTextUsesConfiguredOverlap(t *testing.T) {
	t.Parallel()

	chunks := ChunkText(strings.Repeat("word ", 12), 5, 2)
	require.Len(t, chunks, 4)
	require.Equal(t, "word word word word word", chunks[0].Text)
	require.Equal(t, "word word word word word", chunks[0].IndexText)
	require.Equal(t, 5, chunks[1].TokenCount)
	require.NotEmpty(t, chunks[0].SHA256)
}

func TestValidateCitationsRejectsOutsideContext(t *testing.T) {
	t.Parallel()

	chunks := []platform.RetrievalChunk{{DocumentID: "doc-1", Title: "Guide", Text: "Grounded evidence."}}
	_, err := ValidateCitations("Answer [2]", chunks)
	require.ErrorContains(t, err, "outside context")
}

func TestValidateCitationsBuildsSources(t *testing.T) {
	t.Parallel()

	chunks := []platform.RetrievalChunk{{DocumentID: "doc-1", Title: "Guide", Text: "Grounded evidence."}}
	citations, err := ValidateCitations("Answer [1]", chunks)
	require.NoError(t, err)
	require.Equal(t, "doc-1", citations[0].DocumentID)
}
