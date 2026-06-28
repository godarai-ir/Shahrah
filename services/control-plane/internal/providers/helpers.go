// Package providers contains shared helpers for provider adapter implementations.
package providers

import (
	"io"
	"net/http"
)

// Close closes an io.Closer while intentionally ignoring close errors.
func Close(closer io.Closer) {
	if closer == nil {
		return
	}

	err := closer.Close()
	if err != nil {
		return
	}
}

// ReadAllAndClose reads the full response body and closes it.
func ReadAllAndClose(resp *http.Response) ([]byte, error) {
	if resp == nil || resp.Body == nil {
		return nil, nil
	}

	defer Close(resp.Body)

	return io.ReadAll(resp.Body)
}
