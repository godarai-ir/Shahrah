package guardrails

import (
	"container/list"
	"crypto/sha256"
	"sync"
)

// defaultExtractionCacheBytes bounds the extracted text kept in memory. Chat clients resend every earlier attachment on
// each turn, so without the cache a long conversation parses the same documents again and again.
const defaultExtractionCacheBytes = 64 << 20

// extractionEntryOverhead charges each entry for its key and bookkeeping, so cached failures, which have no text, still
// count against the budget.
const extractionEntryOverhead = 256

type extractionKey [sha256.Size]byte

type extractionEntry struct {
	key  extractionKey
	text string
	err  error
}

// extractionCache is a size-bounded LRU of extraction results keyed by the SHA-256 of the attachment's data URL.
// Failures are cached too, so a malformed file is not parsed again on every turn.
type extractionCache struct {
	mu       sync.Mutex
	maxBytes int
	size     int
	order    *list.List
	entries  map[extractionKey]*list.Element
}

func newExtractionCache(maxBytes int) *extractionCache {
	return &extractionCache{maxBytes: maxBytes, order: list.New(), entries: map[extractionKey]*list.Element{}}
}

func extractionCacheKey(dataURL string) extractionKey {
	return sha256.Sum256([]byte(dataURL))
}

func (c *extractionCache) get(key extractionKey) (string, error, bool) {
	if c == nil {
		return "", nil, false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	element, ok := c.entries[key]
	if !ok {
		return "", nil, false
	}

	entry, ok := element.Value.(*extractionEntry)
	if !ok {
		return "", nil, false
	}

	c.order.MoveToFront(element)

	return entry.text, entry.err, true
}

func (c *extractionCache) put(key extractionKey, text string, err error) {
	if c == nil || len(text)+extractionEntryOverhead > c.maxBytes {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if element, ok := c.entries[key]; ok {
		c.order.MoveToFront(element)

		return
	}

	c.entries[key] = c.order.PushFront(&extractionEntry{key: key, text: text, err: err})
	c.size += len(text) + extractionEntryOverhead

	for c.size > c.maxBytes {
		oldest := c.order.Back()
		c.order.Remove(oldest)

		if entry, ok := oldest.Value.(*extractionEntry); ok {
			delete(c.entries, entry.key)
			c.size -= len(entry.text) + extractionEntryOverhead
		}
	}
}
