package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

var (
	persianPrefixes = map[string]struct{}{"می": {}, "نمی": {}}
	persianSuffixes = map[string]struct{}{
		"ها": {}, "های": {}, "هایی": {}, "تر": {}, "ترین": {},
		"ام": {}, "ات": {}, "اش": {}, "مان": {}, "تان": {}, "شان": {},
	}
)

// NormalizedText keeps source-faithful display content separate from retrieval content.
type NormalizedText struct {
	Display string
	Index   string
}

// NormalizePersian applies the profile's stable Persian retrieval normalization.
func NormalizePersian(input string) string {
	return NormalizePersianText(input).Index
}

// NormalizePersianText returns a minimally changed display form and a canonical index form.
func NormalizePersianText(input string) NormalizedText {
	display := normalizeDisplayText(input)
	index := norm.NFKC.String(display)
	index = strings.NewReplacer(
		"ي", "ی",
		"ى", "ی",
		"ك", "ک",
		"\u00a0", " ",
	).Replace(index)
	index = strings.Map(normalizePersianRune, index)
	index = normalizeIndexWhitespace(strings.ToLower(index))

	return NormalizedText{Display: display, Index: index}
}

func normalizeDisplayText(input string) string {
	input = strings.ReplaceAll(input, "\r\n", "\n")
	input = strings.ReplaceAll(input, "\r", "\n")

	lines := strings.Split(input, "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], "\t ")
	}

	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func normalizePersianRune(value rune) rune {
	switch {
	case value >= '۰' && value <= '۹':
		return '0' + value - '۰'
	case value >= '٠' && value <= '٩':
		return '0' + value - '٠'
	case value == '\u066b':
		return '.'
	case value == '\u066c':
		return ','
	case value == '\u066a':
		return '%'
	case value == '،':
		return ','
	case value == '؛':
		return ';'
	case value == '؟':
		return '?'
	case value == '\u06d4':
		return '.'
	case value == '\n':
		return '\n'
	case value == '\u0640', value == '\u200b', value == '\u200c', value == '\u200d', value == '\u2060', value == '\ufeff':
		return -1
	case isArabicDiacritic(value), isBidirectionalControl(value):
		return -1
	case unicode.IsSpace(value):
		return ' '
	default:
		return value
	}
}

func normalizeIndexWhitespace(input string) string {
	lines := strings.Split(input, "\n")
	normalized := make([]string, 0, len(lines))

	for _, line := range lines {
		line = normalizePersianAffixes(strings.Join(strings.Fields(line), " "))
		if line == "" {
			continue
		}

		normalized = append(normalized, line)
	}

	return strings.Join(normalized, "\n")
}

func isArabicDiacritic(value rune) bool {
	return value >= '\u0610' && value <= '\u061a' ||
		value >= '\u064b' && value <= '\u065f' ||
		value == '\u0670' ||
		value >= '\u06d6' && value <= '\u06ed'
}

func isBidirectionalControl(value rune) bool {
	return value == '\u061c' || value == '\u200e' || value == '\u200f' ||
		value >= '\u202a' && value <= '\u202e' ||
		value >= '\u2066' && value <= '\u2069'
}

func normalizePersianAffixes(input string) string {
	words := strings.Fields(input)
	if len(words) < 2 {
		return input
	}

	result := make([]string, 0, len(words))
	for index := 0; index < len(words); index++ {
		word := words[index]
		if _, ok := persianPrefixes[word]; ok && index+1 < len(words) {
			result = append(result, word+words[index+1])
			index++

			continue
		}

		if _, ok := persianSuffixes[word]; ok && len(result) > 0 {
			result[len(result)-1] += word

			continue
		}

		result = append(result, word)
	}

	return strings.Join(result, " ")
}

// ChunkText creates overlapping chunks while preferring paragraph boundaries.
// Token counts use a deterministic Unicode-word approximation; provider token usage remains authoritative for billing.
func ChunkText(input string, targetTokens, overlapTokens int) []Chunk {
	normalized := NormalizePersianText(input)
	if normalized.Index == "" || targetTokens < 1 || overlapTokens < 0 || overlapTokens >= targetTokens {
		return nil
	}

	words := strings.FieldsFunc(normalized.Display, unicode.IsSpace)
	if len(words) == 0 {
		return nil
	}

	step := targetTokens - overlapTokens

	chunks := make([]Chunk, 0, (len(words)+step-1)/step)
	for start := 0; start < len(words); start += step {
		end := min(start+targetTokens, len(words))
		chunkText := strings.Join(words[start:end], " ")
		indexText := NormalizePersian(chunkText)
		hash := sha256.Sum256([]byte(indexText))
		chunks = append(chunks, Chunk{
			Index:      len(chunks),
			Text:       chunkText,
			IndexText:  indexText,
			TokenCount: end - start,
			SHA256:     hex.EncodeToString(hash[:]),
		})

		if end == len(words) {
			break
		}
	}

	return chunks
}
