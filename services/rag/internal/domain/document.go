package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

// SourceBlockKind identifies a structural unit extracted from a document.
type SourceBlockKind string

const (
	BlockParagraph SourceBlockKind = "paragraph"
	BlockHeading   SourceBlockKind = "heading"
	BlockListItem  SourceBlockKind = "list_item"
	BlockTableRow  SourceBlockKind = "table_row"
)

// SourceBlock preserves the visible source text and its structural location.
type SourceBlock struct {
	Text    string
	Heading string
	Page    *int
	Kind    SourceBlockKind
}

// SourceDocument is the structure-preserving parser output used by chunking.
type SourceDocument struct {
	Blocks    []SourceBlock
	PageCount *int
	Language  string
}

// TokenCounter returns the pinned embedding tokenizer's exact token count.
type TokenCounter interface {
	Count(string) (int, error)
}

// PlainTextDocument converts manual text into paragraph blocks without changing visible content.
func PlainTextDocument(input string) SourceDocument {
	display := NormalizePersianText(input).Display
	paragraphs := strings.Split(display, "\n\n")
	blocks := make([]SourceBlock, 0, len(paragraphs))

	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}

		blocks = append(blocks, SourceBlock{Text: paragraph, Kind: BlockParagraph})
	}

	language := ""
	if display != "" {
		language = ResolveAnswerLanguage("query", "fa-IR", display)
	}

	return SourceDocument{Blocks: blocks, Language: language}
}

// ChunkDocument preserves structural boundaries and measures chunks with the pinned tokenizer.
func ChunkDocument(
	document SourceDocument,
	targetTokens, overlapTokens int,
	tokenizer TokenCounter,
) ([]Chunk, error) {
	if tokenizer == nil {
		return nil, errors.New("document tokenizer is required")
	}

	if targetTokens < 1 || overlapTokens < 0 || overlapTokens >= targetTokens {
		return nil, errors.New("invalid chunk token configuration")
	}

	groups, err := groupSourceBlocks(document.Blocks, targetTokens, tokenizer)
	if err != nil {
		return nil, err
	}

	chunks := make([]Chunk, 0, len(groups))
	for _, group := range groups {
		groupChunks, splitErr := splitSourceGroup(group, targetTokens, overlapTokens, tokenizer)
		if splitErr != nil {
			return nil, splitErr
		}

		for _, chunk := range groupChunks {
			chunk.Index = len(chunks)
			chunks = append(chunks, chunk)
		}
	}

	return chunks, nil
}

type sourceGroup struct {
	text    string
	heading string
	page    *int
}

func groupSourceBlocks(blocks []SourceBlock, targetTokens int, tokenizer TokenCounter) ([]sourceGroup, error) {
	groups := make([]sourceGroup, 0, len(blocks))

	var current sourceGroup

	flush := func() {
		if strings.TrimSpace(current.text) != "" {
			groups = append(groups, current)
		}

		current = sourceGroup{}
	}

	for _, block := range blocks {
		text := strings.TrimSpace(block.Text)
		if text == "" {
			continue
		}

		if current.text == "" {
			current = sourceGroup{text: text, heading: strings.TrimSpace(block.Heading), page: clonePage(block.Page)}

			continue
		}

		if current.heading != strings.TrimSpace(block.Heading) || !samePage(current.page, block.Page) {
			flush()

			current = sourceGroup{text: text, heading: strings.TrimSpace(block.Heading), page: clonePage(block.Page)}

			continue
		}

		candidate := current.text + "\n\n" + text

		count, err := tokenizer.Count(indexContent(current.heading, candidate))
		if err != nil {
			return nil, fmt.Errorf("count grouped document tokens: %w", err)
		}

		if count > targetTokens {
			flush()

			current = sourceGroup{text: text, heading: strings.TrimSpace(block.Heading), page: clonePage(block.Page)}

			continue
		}

		current.text = candidate
	}

	flush()

	return groups, nil
}

func splitSourceGroup(
	group sourceGroup,
	targetTokens, overlapTokens int,
	tokenizer TokenCounter,
) ([]Chunk, error) {
	wholeIndex := indexContent(group.heading, group.text)

	wholeCount, err := tokenizer.Count(wholeIndex)
	if err != nil {
		return nil, fmt.Errorf("count document group tokens: %w", err)
	}

	if wholeCount <= targetTokens {
		return []Chunk{newStructuredChunk(group.text, wholeIndex, group.heading, group.page, wholeCount)}, nil
	}

	words := strings.Fields(group.text)
	chunks := make([]Chunk, 0, wholeCount/targetTokens+1)

	for start := 0; start < len(words); {
		end, tokenCount, findErr := maximumChunkEnd(words, start, group.heading, targetTokens, tokenizer)
		if findErr != nil {
			return nil, findErr
		}

		if end <= start {
			fragments, fragmentErr := splitOversizedWord(words[start], group.heading, targetTokens, tokenizer)
			if fragmentErr != nil {
				return nil, fragmentErr
			}

			for _, fragment := range fragments {
				indexText := indexContent(group.heading, fragment)

				count, countErr := tokenizer.Count(indexText)
				if countErr != nil {
					return nil, fmt.Errorf("count oversized document fragment: %w", countErr)
				}

				chunks = append(chunks, newStructuredChunk(fragment, indexText, group.heading, group.page, count))
			}

			start++

			continue
		}

		displayText := strings.Join(words[start:end], " ")
		indexText := indexContent(group.heading, displayText)
		chunks = append(chunks, newStructuredChunk(displayText, indexText, group.heading, group.page, tokenCount))

		if end == len(words) {
			break
		}

		nextStart, overlapErr := overlapStart(words, start, end, overlapTokens, tokenizer)
		if overlapErr != nil {
			return nil, overlapErr
		}

		if nextStart <= start {
			nextStart = end
		}

		start = nextStart
	}

	return chunks, nil
}

func splitOversizedWord(word, heading string, targetTokens int, tokenizer TokenCounter) ([]string, error) {
	runes := []rune(word)
	fragments := make([]string, 0, 2)

	for start := 0; start < len(runes); {
		low := start + 1
		high := len(runes)
		bestEnd := start

		for low <= high {
			middle := low + (high-low)/2

			count, err := tokenizer.Count(indexContent(heading, string(runes[start:middle])))
			if err != nil {
				return nil, fmt.Errorf("count oversized document word: %w", err)
			}

			if count <= targetTokens {
				bestEnd = middle
				low = middle + 1

				continue
			}

			high = middle - 1
		}

		if bestEnd == start {
			return nil, errors.New("chunk target is too small for the repeated heading")
		}

		fragments = append(fragments, string(runes[start:bestEnd]))
		start = bestEnd
	}

	return fragments, nil
}

func maximumChunkEnd(
	words []string,
	start int,
	heading string,
	targetTokens int,
	tokenizer TokenCounter,
) (int, int, error) {
	low := start + 1
	high := len(words)
	bestEnd := start
	bestCount := 0

	for low <= high {
		middle := low + (high-low)/2
		candidate := indexContent(heading, strings.Join(words[start:middle], " "))

		count, err := tokenizer.Count(candidate)
		if err != nil {
			return 0, 0, fmt.Errorf("count split document tokens: %w", err)
		}

		if count <= targetTokens {
			bestEnd = middle
			bestCount = count
			low = middle + 1

			continue
		}

		high = middle - 1
	}

	return bestEnd, bestCount, nil
}

func overlapStart(words []string, chunkStart, chunkEnd, overlapTokens int, tokenizer TokenCounter) (int, error) {
	if overlapTokens == 0 {
		return chunkEnd, nil
	}

	start := chunkEnd
	for candidate := chunkEnd - 1; candidate > chunkStart; candidate-- {
		count, err := tokenizer.Count(NormalizePersian(strings.Join(words[candidate:chunkEnd], " ")))
		if err != nil {
			return 0, fmt.Errorf("count chunk overlap tokens: %w", err)
		}

		if count > overlapTokens {
			break
		}

		start = candidate
	}

	return start, nil
}

func newStructuredChunk(display, index, heading string, page *int, tokenCount int) Chunk {
	hash := sha256.Sum256([]byte(index))

	return Chunk{
		Text: display, IndexText: index, TokenCount: tokenCount,
		Page: clonePage(page), Heading: heading, SHA256: hex.EncodeToString(hash[:]),
	}
}

func indexContent(heading, display string) string {
	content := NormalizePersian(display)

	heading = NormalizePersian(heading)
	if heading == "" || content == heading || strings.HasPrefix(content, heading+"\n") {
		return content
	}

	return heading + "\n" + content
}

func samePage(first, second *int) bool {
	if first == nil || second == nil {
		return first == nil && second == nil
	}

	return *first == *second
}

func clonePage(input *int) *int {
	if input == nil {
		return nil
	}

	value := *input

	return &value
}
