// Package docextract turns user-supplied document attachments into plain text so content guardrails can inspect them.
package docextract

import (
	"bytes"
	"errors"
	"fmt"
	"path"
	"strings"
	"unicode/utf8"
)

// Kind is the document format detected from the file bytes.
type Kind string

const (
	KindText Kind = "text"
	KindDOCX Kind = "docx"
	KindPDF  Kind = "pdf"
)

const (
	// DefaultMaxFileBytes bounds the decoded attachment size the extractor accepts.
	DefaultMaxFileBytes = 32 << 20
	// DefaultMaxTextBytes bounds the extracted text of one attachment. Text past this bound is never truncated, because
	// content hidden after the cut would bypass inspection.
	DefaultMaxTextBytes = 2 << 20
	// DefaultMaxPDFPages bounds the pages read from one PDF.
	DefaultMaxPDFPages = 500
	// DefaultMaxUnzippedBytes bounds the total uncompressed bytes read from one DOCX, so a zip bomb cannot exhaust memory.
	DefaultMaxUnzippedBytes = 64 << 20
	// DefaultMaxZipEntries bounds the number of parts a DOCX archive may declare.
	DefaultMaxZipEntries = 2000

	textSniffBytes = 8 << 10
)

var (
	// ErrUnsupported reports a non-document format the extractor does not read, such as images, audio, spreadsheets,
	// or unknown binaries.
	ErrUnsupported = errors.New("unsupported attachment format")
	// ErrTooLarge reports an attachment or its text beyond the configured limits.
	ErrTooLarge = errors.New("attachment exceeds inspection limits")
	// ErrEncrypted reports a document that needs a password to read, or an Office compound file that cannot be told apart
	// from one.
	ErrEncrypted = errors.New("attachment is encrypted")
	// ErrMalformed reports a document whose structure could not be parsed.
	ErrMalformed = errors.New("attachment is malformed")
)

// Limits bounds the work one extraction may do.
type Limits struct {
	MaxFileBytes     int
	MaxTextBytes     int
	MaxPDFPages      int
	MaxUnzippedBytes int64
	MaxZipEntries    int
}

// DefaultLimits returns the limits the gateway applies to inline attachments.
func DefaultLimits() Limits {
	return Limits{
		MaxFileBytes:     DefaultMaxFileBytes,
		MaxTextBytes:     DefaultMaxTextBytes,
		MaxPDFPages:      DefaultMaxPDFPages,
		MaxUnzippedBytes: DefaultMaxUnzippedBytes,
		MaxZipEntries:    DefaultMaxZipEntries,
	}
}

// Document is the text extracted from one attachment.
type Document struct {
	Kind Kind
	Text string
}

// Extract detects the format from the leading bytes and returns the document text. The declared media type and file
// name are hints only: a renamed binary is still detected by its content.
func Extract(data []byte, mediaType string, filename string, limits Limits) (Document, error) {
	limits = limits.withDefaults()
	if len(data) > limits.MaxFileBytes {
		return Document{}, fmt.Errorf("%w: file is %d bytes, limit is %d", ErrTooLarge, len(data), limits.MaxFileBytes)
	}

	kind, err := detect(data, mediaType, filename)
	if err != nil {
		return Document{}, err
	}

	var text string

	switch kind {
	case KindPDF:
		text, err = extractPDF(data, limits)
	case KindDOCX:
		text, err = extractDOCX(data, limits)
	default:
		text, err = decodeText(data)
	}

	if err != nil {
		return Document{}, err
	}

	if len(text) > limits.MaxTextBytes {
		return Document{}, fmt.Errorf("%w: extracted text is %d bytes, limit is %d", ErrTooLarge, len(text), limits.MaxTextBytes)
	}

	return Document{Kind: kind, Text: strings.ToValidUTF8(text, "")}, nil
}

func (l Limits) withDefaults() Limits {
	defaults := DefaultLimits()
	if l.MaxFileBytes <= 0 {
		l.MaxFileBytes = defaults.MaxFileBytes
	}

	if l.MaxTextBytes <= 0 {
		l.MaxTextBytes = defaults.MaxTextBytes
	}

	if l.MaxPDFPages <= 0 {
		l.MaxPDFPages = defaults.MaxPDFPages
	}

	if l.MaxUnzippedBytes <= 0 {
		l.MaxUnzippedBytes = defaults.MaxUnzippedBytes
	}

	if l.MaxZipEntries <= 0 {
		l.MaxZipEntries = defaults.MaxZipEntries
	}

	return l
}

func detect(data []byte, mediaType string, filename string) (Kind, error) {
	switch {
	case bytes.HasPrefix(data, []byte("%PDF-")):
		return KindPDF, nil
	case bytes.HasPrefix(data, []byte("PK\x03\x04")):
		if isDOCX(data) {
			return KindDOCX, nil
		}

		return "", fmt.Errorf("%w: zip archive is not a Word document", ErrUnsupported)
	case bytes.HasPrefix(data, []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}):
		// Password-protected DOCX files and legacy .doc files share the OLE compound file container. Both are documents
		// that cannot be inspected, so they are reported as encrypted rather than as a harmless unsupported format.
		return "", fmt.Errorf("%w: password-protected or legacy Office document", ErrEncrypted)
	case looksLikeText(data):
		return KindText, nil
	}

	hint := strings.ToLower(strings.TrimSpace(mediaType))
	if ext := strings.ToLower(path.Ext(filename)); ext != "" {
		hint += " " + ext
	}

	return "", fmt.Errorf("%w: %s", ErrUnsupported, strings.TrimSpace(hint))
}

func looksLikeText(data []byte) bool {
	if hasUTF16BOM(data) {
		return true
	}

	sample := data
	if len(sample) > textSniffBytes {
		sample = sample[:textSniffBytes]
	}

	if bytes.IndexByte(sample, 0) >= 0 {
		return false
	}

	if utf8.Valid(trimPartialRune(sample)) {
		return true
	}

	// Legacy single-byte text, such as Windows-1256 Persian, has no NUL bytes and few control characters.
	control := 0

	for _, b := range sample {
		if b < 0x20 && b != '\n' && b != '\r' && b != '\t' && b != '\f' {
			control++
		}
	}

	return control*100 <= len(sample)
}

// trimPartialRune drops a multi-byte rune cut by the sniff window, so valid UTF-8 is not misread as binary.
func trimPartialRune(sample []byte) []byte {
	for i := 0; i < utf8.UTFMax && len(sample) > 0; i++ {
		if utf8.Valid(sample) {
			return sample
		}

		sample = sample[:len(sample)-1]
	}

	return sample
}
