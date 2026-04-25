package unit_test

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func httptestRequest(t *testing.T, method string, url string, body string) *http.Request {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	return req
}
