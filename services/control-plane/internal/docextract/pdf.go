package docextract

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/ledongthuc/pdf"
	"golang.org/x/text/unicode/norm"
)

// extractPDF reads the text layer of every page. Scanned pages without a text layer yield no text; OCR is out of scope
// here. The parser reports some malformed input by panicking, so the whole read runs under recover.
func extractPDF(data []byte, limits Limits) (text string, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			text = ""
			err = fmt.Errorf("%w: invalid PDF: %v", ErrMalformed, recovered)
		}
	}()

	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		if errors.Is(err, pdf.ErrInvalidPassword) {
			return "", fmt.Errorf("%w: PDF requires a password", ErrEncrypted)
		}

		return "", fmt.Errorf("%w: invalid PDF: %w", ErrMalformed, err)
	}

	pages := reader.NumPage()
	if pages > limits.MaxPDFPages {
		return "", fmt.Errorf("%w: PDF has %d pages, limit is %d", ErrTooLarge, pages, limits.MaxPDFPages)
	}

	var out strings.Builder

	for number := 1; number <= pages; number++ {
		pageText, err := reader.Page(number).GetPlainText(nil)
		if err != nil {
			return "", fmt.Errorf("%w: PDF page %d: %w", ErrMalformed, number, err)
		}

		pageText = strings.TrimSpace(pageText)
		if pageText == "" {
			continue
		}

		if out.Len() > 0 {
			out.WriteString("\n\n")
		}

		out.WriteString(pageText)

		if out.Len() > limits.MaxTextBytes {
			return "", fmt.Errorf("%w: extracted text exceeds %d bytes", ErrTooLarge, limits.MaxTextBytes)
		}
	}

	// PDF generators often emit Arabic presentation forms for Persian glyphs. NFKC maps them back to base letters, so
	// keyword and pattern guardrails match the same spelling users type.
	return norm.NFKC.String(out.String()), nil
}
