package parser

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/Delpak/shahrah/services/rag/internal/domain"
)

var orderedListPrefix = regexp.MustCompile(`^\d+[.)]\s+`)

// OCRPage is one zero-indexed page returned by a document OCR provider.
type OCRPage struct {
	Index    int
	Markdown string
}

// ParseOCRMarkdown converts page-scoped OCR Markdown into citation-safe source blocks.
func ParseOCRMarkdown(input []OCRPage) (domain.SourceDocument, error) {
	if len(input) == 0 {
		return domain.SourceDocument{}, ErrOCRRequired
	}

	pages := append([]OCRPage(nil), input...)
	sort.Slice(pages, func(first, second int) bool { return pages[first].Index < pages[second].Index })

	document := domain.SourceDocument{}
	seen := make(map[int]struct{}, len(pages))
	currentHeading := ""

	var languageSample strings.Builder

	for _, page := range pages {
		if page.Index < 0 {
			return domain.SourceDocument{}, errors.New("OCR page index must not be negative")
		}

		if _, exists := seen[page.Index]; exists {
			return domain.SourceDocument{}, fmt.Errorf("OCR page index %d is duplicated", page.Index)
		}

		seen[page.Index] = struct{}{}
		pageNumber := page.Index + 1
		blocks, heading := parseMarkdownPage(page.Markdown, currentHeading, pageNumber)
		currentHeading = heading

		document.Blocks = append(document.Blocks, blocks...)

		languageSample.WriteString(page.Markdown)
		languageSample.WriteRune('\n')
	}

	pageCount := pages[len(pages)-1].Index + 1
	document.PageCount = &pageCount

	document.Language = domain.ResolveAnswerLanguage("query", "fa-IR", languageSample.String())
	if err := validateExtractedDocument(document); err != nil {
		return domain.SourceDocument{}, err
	}

	return document, nil
}

func parseMarkdownPage(markdown, initialHeading string, page int) ([]domain.SourceBlock, string) {
	lines := strings.Split(strings.ReplaceAll(markdown, "\r\n", "\n"), "\n")
	blocks := make([]domain.SourceBlock, 0, len(lines))
	heading := initialHeading
	paragraph := make([]string, 0, 4)

	flushParagraph := func() {
		text := cleanInlineMarkdown(strings.Join(paragraph, " "))
		paragraph = paragraph[:0]

		if text != "" {
			blocks = append(blocks, sourceBlock(text, heading, page, domain.BlockParagraph))
		}
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			flushParagraph()

			continue
		}

		if level, title := markdownHeading(line); level > 0 {
			flushParagraph()

			heading = cleanInlineMarkdown(title)
			if heading != "" {
				blocks = append(blocks, sourceBlock(heading, heading, page, domain.BlockHeading))
			}

			continue
		}

		if isMarkdownTableRow(line) {
			flushParagraph()

			if row := markdownTableRow(line); row != "" {
				blocks = append(blocks, sourceBlock(row, heading, page, domain.BlockTableRow))
			}

			continue
		}

		if item, ok := markdownListItem(line); ok {
			flushParagraph()

			blocks = append(blocks, sourceBlock(cleanInlineMarkdown(item), heading, page, domain.BlockListItem))

			continue
		}

		paragraph = append(paragraph, line)
	}

	flushParagraph()

	return blocks, heading
}

func sourceBlock(text, heading string, page int, kind domain.SourceBlockKind) domain.SourceBlock {
	pageCopy := page

	return domain.SourceBlock{Text: text, Heading: heading, Page: &pageCopy, Kind: kind}
}

func markdownHeading(line string) (int, string) {
	level := 0
	for level < len(line) && level < 6 && line[level] == '#' {
		level++
	}

	if level == 0 || level >= len(line) || line[level] != ' ' {
		return 0, ""
	}

	return level, strings.TrimSpace(line[level:])
}

func markdownListItem(line string) (string, bool) {
	if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") || strings.HasPrefix(line, "+ ") {
		return strings.TrimSpace(line[2:]), true
	}

	if location := orderedListPrefix.FindStringIndex(line); location != nil {
		return strings.TrimSpace(line[location[1]:]), true
	}

	return "", false
}

func isMarkdownTableRow(line string) bool {
	return strings.HasPrefix(line, "|") && strings.HasSuffix(line, "|")
}

func markdownTableRow(line string) string {
	line = strings.Trim(line, "|")
	cells := strings.Split(line, "|")
	clean := make([]string, 0, len(cells))
	separator := true

	for _, cell := range cells {
		cell = cleanInlineMarkdown(cell)
		if strings.Trim(cell, ":- ") != "" {
			separator = false
		}

		if cell != "" {
			clean = append(clean, cell)
		}
	}

	if separator {
		return ""
	}

	return strings.Join(clean, " | ")
}

func cleanInlineMarkdown(input string) string {
	input = strings.NewReplacer("**", "", "__", "", "`", "").Replace(input)

	return strings.Join(strings.Fields(input), " ")
}
