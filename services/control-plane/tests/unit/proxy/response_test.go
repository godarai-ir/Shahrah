package proxy_test

import (
	"net/http"
	"testing"
)

type testResponse struct {
	*http.Response
}

func closeResponseBody(tb testing.TB, resp *http.Response) {
	tb.Helper()

	if resp == nil || resp.Body == nil {
		return
	}

	_ = resp.Body.Close()
}
