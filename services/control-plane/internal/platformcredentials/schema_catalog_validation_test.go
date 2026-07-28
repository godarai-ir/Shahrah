package platformcredentials

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSchemaCatalogDerivesCredentialValidationCapability(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/v1/providers/google-gemini/connection-schemas" {
			http.NotFound(writer, request)

			return
		}

		writer.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(writer, `{
			"slug": "google-gemini",
			"connectionSchemas": [
				{
					"authMode": "api_key",
					"version": 1,
					"runtimeReady": true,
					"fieldsJson": []
				},
				{
					"authMode": "oauth",
					"version": 1,
					"runtimeReady": false,
					"fieldsJson": []
				}
			]
		}`)
	}))
	defer server.Close()

	catalog := &ModelDiscoverySchemaCatalog{
		baseURL: server.URL + "/api/v1",
		client:  server.Client(),
		ttl:     time.Minute,
		cache:   make(map[string]schemaCacheEntry),
	}

	schemas, err := catalog.Schemas(context.Background(), "gemini")
	if err != nil {
		t.Fatalf("Schemas: %v", err)
	}
	if len(schemas) != 2 {
		t.Fatalf("expected 2 schemas, got %d", len(schemas))
	}
	if !schemas[0].CredentialValidationSupported {
		t.Fatal("expected Gemini API-key validation to be advertised")
	}
	if schemas[1].CredentialValidationSupported {
		t.Fatal("unsupported OAuth validation must not be advertised")
	}
}
