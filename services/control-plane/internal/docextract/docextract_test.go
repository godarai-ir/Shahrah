package docextract

import (
	"archive/zip"
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
)

func TestExtractPlainTextEncodings(t *testing.T) {
	t.Parallel()

	persian := "سند محرمانه"

	windows1256, err := charmap.Windows1256.NewEncoder().String(persian)
	require.NoError(t, err)

	utf16, err := unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewEncoder().String(persian)
	require.NoError(t, err)

	cases := map[string][]byte{
		"utf8":         []byte(persian),
		"utf8 bom":     append([]byte{0xEF, 0xBB, 0xBF}, persian...),
		"utf16 bom":    []byte(utf16),
		"windows-1256": []byte(windows1256),
	}

	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			doc, err := Extract(data, "text/plain", "note.txt", Limits{})
			require.NoError(t, err)
			assert.Equal(t, KindText, doc.Kind)
			assert.Equal(t, persian, doc.Text)
		})
	}
}

func TestExtractDetectsFormatFromContentNotName(t *testing.T) {
	t.Parallel()

	doc, err := Extract(minimalPDF(t, "Confidential budget"), "text/plain", "notes.txt", Limits{})
	require.NoError(t, err)
	assert.Equal(t, KindPDF, doc.Kind)
	assert.Contains(t, doc.Text, "Confidential budget")

	_, err = Extract([]byte{0x89, 'P', 'N', 'G', 0, 0, 0, 0}, "text/plain", "notes.txt", Limits{})
	require.ErrorIs(t, err, ErrUnsupported)
}

func TestExtractPDFPages(t *testing.T) {
	t.Parallel()

	doc, err := Extract(minimalPDF(t, "Quarterly report", "INTERNAL ONLY"), "application/pdf", "report.pdf", Limits{})
	require.NoError(t, err)
	assert.Contains(t, doc.Text, "Quarterly report")
	assert.Contains(t, doc.Text, "INTERNAL ONLY")

	_, err = Extract(minimalPDF(t, "one", "two"), "application/pdf", "report.pdf", Limits{MaxPDFPages: 1})
	require.ErrorIs(t, err, ErrTooLarge)
}

func TestExtractMalformedPDF(t *testing.T) {
	t.Parallel()

	_, err := Extract([]byte("%PDF-1.4\nnot really a pdf"), "application/pdf", "broken.pdf", Limits{})
	require.ErrorIs(t, err, ErrMalformed)
}

func TestExtractDOCXReadsBodyHeadersCommentsAndProperties(t *testing.T) {
	t.Parallel()

	data := buildDOCX(t, map[string]string{
		"word/document.xml":   wordXML(`<w:p><w:r><w:t>Project plan</w:t></w:r><w:r><w:tab/><w:t xml:space="preserve">draft </w:t></w:r><w:del><w:r><w:delText>removed secret</w:delText></w:r></w:del></w:p><w:p><w:r><w:t>Second paragraph</w:t></w:r></w:p>`),
		"word/header1.xml":    wordXML(`<w:p><w:r><w:t>خیلی محرمانه</w:t></w:r></w:p>`),
		"word/comments.xml":   wordXML(`<w:comment><w:p><w:r><w:t>do not share</w:t></w:r></w:p></w:comment>`),
		"docProps/custom.xml": `<?xml version="1.0"?><Properties xmlns:vt="http://schemas.openxmlformats.org/officeDocument/2006/docPropsVTypes"><property name="MSIP_Label_Name"><vt:lpwstr>Highly Confidential</vt:lpwstr></property></Properties>`,
		"word/styles.xml":     wordXML(`<w:p><w:r><w:t>style noise</w:t></w:r></w:p>`),
	})

	doc, err := Extract(data, "application/octet-stream", "plan.docx", Limits{})
	require.NoError(t, err)
	assert.Equal(t, KindDOCX, doc.Kind)

	for _, want := range []string{"Project plan\tdraft removed secret\nSecond paragraph", "خیلی محرمانه", "do not share", "Highly Confidential"} {
		assert.Contains(t, doc.Text, want)
	}

	assert.NotContains(t, doc.Text, "style noise")
	assert.Less(t, strings.Index(doc.Text, "Highly Confidential"), strings.Index(doc.Text, "Project plan"), "properties come first")
}

func TestExtractRejectsZipBombAndOtherArchives(t *testing.T) {
	t.Parallel()

	bomb := buildDOCX(t, map[string]string{"word/document.xml": wordXML(`<w:p><w:r><w:t>` + strings.Repeat("a", 4096) + `</w:t></w:r></w:p>`)})

	_, err := Extract(bomb, "", "bomb.docx", Limits{MaxUnzippedBytes: 1024})
	require.ErrorIs(t, err, ErrTooLarge)

	xlsx := buildDOCX(t, map[string]string{"xl/workbook.xml": "<workbook/>"})

	_, err = Extract(xlsx, "", "sheet.xlsx", Limits{})
	require.ErrorIs(t, err, ErrUnsupported)
}

func TestExtractRejectsOversizedInputAndText(t *testing.T) {
	t.Parallel()

	_, err := Extract([]byte(strings.Repeat("a", 100)), "text/plain", "", Limits{MaxFileBytes: 10})
	require.ErrorIs(t, err, ErrTooLarge)

	_, err = Extract([]byte(strings.Repeat("a", 100)), "text/plain", "", Limits{MaxTextBytes: 10})
	require.ErrorIs(t, err, ErrTooLarge)
}

func wordXML(body string) string {
	return `<?xml version="1.0"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>` + body + `</w:body></w:document>`
}

func buildDOCX(t *testing.T, parts map[string]string) []byte {
	t.Helper()

	var buf bytes.Buffer

	archive := zip.NewWriter(&buf)

	for name, content := range parts {
		writer, err := archive.Create(name)
		require.NoError(t, err)

		_, err = writer.Write([]byte(content))
		require.NoError(t, err)
	}

	require.NoError(t, archive.Close())

	return buf.Bytes()
}

// minimalPDF writes one Helvetica page per text with a correct cross-reference table.
func minimalPDF(t *testing.T, pages ...string) []byte {
	t.Helper()

	objects := make([]string, 0, 3+2*len(pages))
	objects = append(objects, "<< /Type /Catalog /Pages 2 0 R >>", "")
	kids := make([]string, 0, len(pages))
	fontID := 3 + 2*len(pages)

	for i, text := range pages {
		pageID, contentID := 3+2*i, 4+2*i
		stream := fmt.Sprintf("BT /F1 12 Tf 72 720 Td (%s) Tj ET", text)
		kids = append(kids, fmt.Sprintf("%d 0 R", pageID))
		objects = append(objects,
			fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents %d 0 R /Resources << /Font << /F1 %d 0 R >> >> >>", contentID, fontID),
			fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(stream), stream),
		)
	}

	objects[1] = fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>", strings.Join(kids, " "), len(pages))
	objects = append(objects, "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>")

	var buf bytes.Buffer

	buf.WriteString("%PDF-1.4\n")

	offsets := make([]int, len(objects))
	for i, object := range objects {
		offsets[i] = buf.Len()
		fmt.Fprintf(&buf, "%d 0 obj\n%s\nendobj\n", i+1, object)
	}

	xref := buf.Len()
	fmt.Fprintf(&buf, "xref\n0 %d\n0000000000 65535 f \n", len(objects)+1)

	for _, offset := range offsets {
		fmt.Fprintf(&buf, "%010d 00000 n \n", offset)
	}

	fmt.Fprintf(&buf, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xref)

	return buf.Bytes()
}
