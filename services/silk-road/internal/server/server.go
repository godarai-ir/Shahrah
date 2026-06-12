package server

import (
	"net/http"

	pkgserver "github.com/Delpak/shahrah/services/silk-road/pkg/server"
)

// NewHandler delegates to the shared pkg/server implementation.
func NewHandler(client *http.Client) http.Handler {
	return pkgserver.NewHandler(client)
}
