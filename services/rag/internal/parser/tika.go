// Package parser extracts structured content through the internal Apache Tika service.
package parser

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/Delpak/shahrah/services/rag/internal/domain"
)

const maxExtractedTextBytes = 100 << 20

var numericXMLReference = regexp.MustCompile(`&#(?:x[0-9A-Fa-f]+|\d+);`)

var (
	// ErrOCRRequired indicates that the source has no usable embedded text.
	ErrOCRRequired = errors.New("ocr_required")
	// ErrLowExtractionQuality indicates that extracted text is unsafe to index.
	ErrLowExtractionQuality = errors.New("low_extraction_quality")
)

// TikaClient is a bounded internal structured-content extraction client.
type TikaClient struct {
	endpoint string
	client   *http.Client
}

// NewTikaClient creates a Tika client.
func NewTikaClient(endpoint string, timeout time.Duration) (*TikaClient, error) {
	endpoint = strings.TrimRight(strings.TrimSpace(endpoint), "/")
	if endpoint == "" {
		return nil, errors.New("tika endpoint is required")
	}

	if timeout <= 0 {
		timeout = 2 * time.Minute
	}

	return &TikaClient{endpoint: endpoint, client: &http.Client{Timeout: timeout}}, nil
}

// Extract returns source-faithful blocks and structural metadata from Tika XHTML.
func (c *TikaClient) Extract(
	ctx context.Context,
	mediaType string,
	body io.Reader,
) (domain.SourceDocument, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.endpoint+"/tika", body)
	if err != nil {
		return domain.SourceDocument{}, fmt.Errorf("create Tika request: %w", err)
	}

	req.Header.Set("Content-Type", mediaType)
	// Tika Server 3.x exposes its XML-serialized XHTML representation as
	// text/xml. Requesting application/xhtml+xml is rejected with HTTP 406,
	// while text/html is not guaranteed to be XML-safe.
	req.Header.Set("Accept", "text/xml")

	resp, err := c.client.Do(req)
	if err != nil {
		return domain.SourceDocument{}, fmt.Errorf("send Tika request: %w", err)
	}
	defer closeResponseBody(resp.Body)

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return domain.SourceDocument{}, fmt.Errorf("tika returned HTTP %d", resp.StatusCode)
	}

	content, err := io.ReadAll(io.LimitReader(resp.Body, maxExtractedTextBytes+1))
	if err != nil {
		return domain.SourceDocument{}, fmt.Errorf("read Tika response: %w", err)
	}

	if len(content) > maxExtractedTextBytes {
		return domain.SourceDocument{}, errors.New("extracted text exceeds limit")
	}

	document, err := parseXHTML(content, mediaType)
	if err != nil {
		return domain.SourceDocument{}, err
	}

	if err := validateExtractedDocument(document); err != nil {
		return domain.SourceDocument{}, err
	}

	if strings.TrimSpace(document.Language) == "" {
		document.Language = detectedDocumentLanguage(document)
	}

	return document, nil
}

type xhtmlCapture struct {
	tag       string
	kind      domain.SourceBlockKind
	text      strings.Builder
	tableCell bool
}

type xhtmlExtraction struct {
	document       domain.SourceDocument
	bodyText       strings.Builder
	inBody         bool
	currentHeading string
	currentPage    int
	capture        *xhtmlCapture
}

func parseXHTML(content []byte, mediaType string) (domain.SourceDocument, error) {
	content = normalizeTikaXML(content)
	decoder := xml.NewDecoder(strings.NewReader(string(content)))
	decoder.Strict = false
	state := xhtmlExtraction{}

	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return domain.SourceDocument{}, fmt.Errorf("parse Tika XHTML: %w", err)
		}

		switch value := token.(type) {
		case xml.StartElement:
			state.start(value)
		case xml.CharData:
			if state.inBody {
				state.bodyText.Write(value)
			}

			if state.capture != nil {
				state.capture.text.Write(value)
			}
		case xml.EndElement:
			state.end(value)
		}
	}

	if len(state.document.Blocks) == 0 {
		fallback := domain.PlainTextDocument(state.bodyText.String())

		state.document.Blocks = fallback.Blocks
		if state.document.Language == "" {
			state.document.Language = fallback.Language
		}
	}

	if state.currentPage > 0 {
		pageCount := state.currentPage
		state.document.PageCount = &pageCount
	} else if strings.EqualFold(mediaType, "application/pdf") && len(state.document.Blocks) > 0 {
		pageCount := 1
		state.document.PageCount = &pageCount
	}

	return state.document, nil
}

func normalizeTikaXML(content []byte) []byte {
	const (
		xml11 = `<?xml version="1.1"`
		xml10 = `<?xml version="1.0"`
	)

	if bytes.HasPrefix(content, []byte(xml11)) {
		content = bytes.Replace(content, []byte(xml11), []byte(xml10), 1)
	}

	content = numericXMLReference.ReplaceAllFunc(content, func(reference []byte) []byte {
		value := string(reference[2 : len(reference)-1])
		base := 10

		if strings.HasPrefix(value, "x") {
			base = 16
			value = value[1:]
		}

		parsed, err := strconv.ParseInt(value, base, 32)
		if err != nil || validXML10Rune(rune(parsed)) {
			return reference
		}

		return []byte(" ")
	})

	return bytes.Map(func(value rune) rune {
		if validXML10Rune(value) {
			return value
		}

		return ' '
	}, content)
}

func validXML10Rune(value rune) bool {
	return value == '\t' || value == '\n' || value == '\r' ||
		value >= 0x20 && value <= 0xD7FF ||
		value >= 0xE000 && value <= 0xFFFD ||
		value >= 0x10000 && value <= 0x10FFFF
}

func (s *xhtmlExtraction) start(element xml.StartElement) {
	tag := strings.ToLower(element.Name.Local)
	if tag == "html" {
		s.document.Language = attribute(element.Attr, "lang")
	}

	if tag == "body" {
		s.inBody = true

		return
	}

	if !s.inBody {
		return
	}

	if tag == "div" && hasClass(element.Attr, "page") {
		s.currentPage++
	}

	if s.capture != nil {
		if s.capture.kind == domain.BlockTableRow && (tag == "td" || tag == "th") {
			if s.capture.tableCell {
				s.capture.text.WriteRune('\x00')
			}

			s.capture.tableCell = true
		}

		return
	}

	kind, ok := blockKind(tag)
	if ok {
		s.capture = &xhtmlCapture{tag: tag, kind: kind}
	}
}

func (s *xhtmlExtraction) end(element xml.EndElement) {
	tag := strings.ToLower(element.Name.Local)
	if tag == "body" {
		s.inBody = false

		return
	}

	if s.capture == nil || s.capture.tag != tag {
		return
	}

	text := normalizeBlockText(s.capture.text.String(), s.capture.kind)
	kind := s.capture.kind
	s.capture = nil

	if text == "" {
		return
	}

	if kind == domain.BlockHeading {
		s.currentHeading = text
	}

	block := domain.SourceBlock{Text: text, Heading: s.currentHeading, Kind: kind}
	if s.currentPage > 0 {
		page := s.currentPage
		block.Page = &page
	}

	s.document.Blocks = append(s.document.Blocks, block)
}

func blockKind(tag string) (domain.SourceBlockKind, bool) {
	switch tag {
	case "h1", "h2", "h3", "h4", "h5", "h6":
		return domain.BlockHeading, true
	case "p":
		return domain.BlockParagraph, true
	case "li":
		return domain.BlockListItem, true
	case "tr":
		return domain.BlockTableRow, true
	default:
		return "", false
	}
}

func normalizeBlockText(input string, kind domain.SourceBlockKind) string {
	if kind != domain.BlockTableRow {
		return strings.Join(strings.Fields(input), " ")
	}

	cells := strings.Split(input, "\x00")

	cleanCells := make([]string, 0, len(cells))
	for _, cell := range cells {
		cell = strings.Join(strings.Fields(cell), " ")
		if cell != "" {
			cleanCells = append(cleanCells, cell)
		}
	}

	return strings.Join(cleanCells, " | ")
}

func validateExtractedDocument(document domain.SourceDocument) error {
	var content strings.Builder
	for _, block := range document.Blocks {
		content.WriteString(block.Text)
		content.WriteRune('\n')
	}

	text := strings.TrimSpace(content.String())
	if text == "" {
		return ErrOCRRequired
	}

	if !utf8.ValidString(text) {
		return fmt.Errorf("%w: invalid UTF-8", ErrLowExtractionQuality)
	}

	total := 0
	replacement := 0
	controls := 0
	mojibake := 0

	for _, value := range text {
		total++

		switch {
		case value == unicode.ReplacementChar:
			replacement++
		case unicode.IsControl(value) && !unicode.IsSpace(value):
			controls++
		case strings.ContainsRune("ÃÂØÙ", value):
			mojibake++
		}
	}

	if excessiveNoise(total, replacement, 3) || excessiveNoise(total, controls, 1) || excessiveNoise(total, mojibake, 3) {
		return fmt.Errorf("%w: excessive invalid characters", ErrLowExtractionQuality)
	}

	return nil
}

func excessiveNoise(total, noise, minimum int) bool {
	return noise >= minimum && noise*100 > total
}

func detectedDocumentLanguage(document domain.SourceDocument) string {
	var sample strings.Builder
	for _, block := range document.Blocks {
		sample.WriteString(block.Text)
		sample.WriteRune('\n')
	}

	return domain.ResolveAnswerLanguage("query", "fa-IR", sample.String())
}

func attribute(attributes []xml.Attr, name string) string {
	for _, candidate := range attributes {
		if strings.EqualFold(candidate.Name.Local, name) {
			return strings.TrimSpace(candidate.Value)
		}
	}

	return ""
}

func hasClass(attributes []xml.Attr, expected string) bool {
	for _, className := range strings.Fields(attribute(attributes, "class")) {
		if strings.EqualFold(className, expected) {
			return true
		}
	}

	return false
}

func closeResponseBody(body io.Closer) {
	if err := body.Close(); err != nil {
		slog.Warn("managed RAG parser response close failed", "error_type", "response_close")
	}
}
