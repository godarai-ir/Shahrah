package tracing

import (
	"net/http"
	"testing"
)

func closeResponseBody(tb testing.TB, resp *http.Response) {
	tb.Helper()

	if resp == nil || resp.Body == nil {
		return
	}

	_ = resp.Body.Close()
}
