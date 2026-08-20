package domain

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

var citationPattern = regexp.MustCompile(`\[(\d+)]`)

// ValidateCitations rejects answers that cite outside the supplied context or fabricate quotes.
func ValidateCitations(answer string, chunks []platform.RetrievalChunk) ([]platform.Citation, error) {
	matches := citationPattern.FindAllStringSubmatch(answer, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("answer has no citations")
	}

	byID := make(map[int]platform.RetrievalChunk, len(chunks))
	for i, chunk := range chunks {
		byID[i+1] = chunk
	}

	used := make(map[int]struct{}, len(matches))
	for _, match := range matches {
		index, err := strconv.Atoi(match[1])
		if err != nil {
			return nil, fmt.Errorf("parse citation index: %w", err)
		}

		if _, ok := byID[index]; !ok {
			return nil, fmt.Errorf("citation %d is outside context", index)
		}

		used[index] = struct{}{}
	}

	indices := make([]int, 0, len(used))
	for index := range used {
		indices = append(indices, index)
	}

	sort.Ints(indices)

	citations := make([]platform.Citation, 0, len(indices))
	for _, index := range indices {
		chunk := byID[index]

		quote := strings.TrimSpace(chunk.Text)
		if len([]rune(quote)) > 280 {
			quote = string([]rune(quote)[:280])
		}

		citations = append(citations, platform.Citation{
			ID:         strconv.Itoa(index),
			DocumentID: chunk.DocumentID,
			Title:      chunk.Title,
			Page:       chunk.Page,
			Quote:      quote,
		})
	}

	return citations, nil
}
