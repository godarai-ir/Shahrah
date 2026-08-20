// Package tokenization owns the tokenizer pinned by the managed RAG profile.
package tokenization

import (
	"fmt"

	gotokenizer "github.com/tiktoken-go/tokenizer"
)

// CL100K counts tokens with the profile's embedded cl100k_base vocabulary.
type CL100K struct {
	codec gotokenizer.Codec
}

// NewCL100K initializes the local deterministic embedding tokenizer.
func NewCL100K() (*CL100K, error) {
	codec, err := gotokenizer.Get(gotokenizer.Cl100kBase)
	if err != nil {
		return nil, fmt.Errorf("load cl100k_base tokenizer: %w", err)
	}

	return &CL100K{codec: codec}, nil
}

// Count returns the exact cl100k_base token count for text.
func (t *CL100K) Count(input string) (int, error) {
	count, err := t.codec.Count(input)
	if err != nil {
		return 0, fmt.Errorf("count cl100k_base tokens: %w", err)
	}

	return count, nil
}
