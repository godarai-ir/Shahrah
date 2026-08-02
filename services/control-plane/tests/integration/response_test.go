package integration_test

import (
	"fmt"
	"net/http"
	"sync"
	"testing"
)

type testResponse struct {
	*http.Response
}

type asyncErrorCollector struct {
	mu     sync.Mutex
	errors []error
}

func newAsyncErrorCollector(tb testing.TB) *asyncErrorCollector {
	tb.Helper()

	collector := &asyncErrorCollector{}

	tb.Cleanup(func() {
		for _, err := range collector.all() {
			tb.Errorf("asynchronous test handler failed: %v", err)
		}
	})

	return collector
}

func (c *asyncErrorCollector) Errorf(format string, args ...interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.errors = append(c.errors, fmt.Errorf(format, args...))
}

func (c *asyncErrorCollector) all() []error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return append([]error(nil), c.errors...)
}

func closeResponseBody(tb testing.TB, resp *http.Response) {
	tb.Helper()

	if resp == nil || resp.Body == nil {
		return
	}

	_ = resp.Body.Close()
}
