package docextract

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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

		// The parser cannot decrypt AES-256 (V=5) documents, even when they open without a password.
		if strings.Contains(err.Error(), "unsupported PDF: encryption") {
			return "", fmt.Errorf("%w: unsupported PDF encryption: %w", ErrEncrypted, err)
		}

		return "", fmt.Errorf("%w: invalid PDF: %w", ErrMalformed, err)
	}

	pages := reader.NumPage()
	if pages > limits.MaxPDFPages {
		return "", fmt.Errorf("%w: PDF has %d pages, limit is %d", ErrTooLarge, pages, limits.MaxPDFPages)
	}

	var out strings.Builder

	budget := limits.MaxUnzippedBytes

	for number := 1; number <= pages; number++ {
		page := reader.Page(number)

		// Content streams are inflated without bound by the parser, so their decoded size is measured against the budget
		// first; a compression bomb then fails here instead of exhausting memory during text extraction.
		size, err := pdfContentSize(page.V.Key("Contents"), budget)
		if err != nil {
			return "", err
		}

		budget -= size

		pageText, err := page.GetPlainText(nil)
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

// pdfContentSize returns the decoded size of a page's content streams, or ErrTooLarge once it passes budget.
func pdfContentSize(contents pdf.Value, budget int64) (int64, error) {
	streams := []pdf.Value{contents}
	if contents.Kind() == pdf.Array {
		streams = make([]pdf.Value, 0, contents.Len())
		for i := 0; i < contents.Len(); i++ {
			streams = append(streams, contents.Index(i))
		}
	}

	var total int64

	for _, stream := range streams {
		if stream.Kind() != pdf.Stream {
			continue
		}

		reader := stream.Reader()
		n, err := io.Copy(io.Discard, io.LimitReader(reader, budget-total+1))
		closeErr := reader.Close()

		if err != nil {
			return 0, fmt.Errorf("%w: PDF content stream: %w", ErrMalformed, err)
		}

		if closeErr != nil {
			return 0, fmt.Errorf("%w: PDF content stream: %w", ErrMalformed, closeErr)
		}

		total += n
		if total > budget {
			return 0, fmt.Errorf("%w: PDF content expands beyond the inspection limit", ErrTooLarge)
		}
	}

	return total, nil
}
