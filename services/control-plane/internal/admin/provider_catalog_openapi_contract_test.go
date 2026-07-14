package admin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type openAPIParameter struct {
	Name     string `json:"name"`
	In       string `json:"in"`
	Required bool   `json:"required"`
	Schema   struct {
		Type    string   `json:"type"`
		Maximum *float64 `json:"maximum"`
	} `json:"schema"`
}

type openAPIOperation struct {
	Parameters []openAPIParameter `json:"parameters"`
}

type openAPIDocument struct {
	OpenAPI string                                 `json:"openapi"`
	Paths   map[string]map[string]openAPIOperation `json:"paths"`
}

func TestModelDiscoveryOpenAPIContract_MatchesControlPlaneUsage(t *testing.T) {
	spec := loadModelDiscoveryOpenAPI(t)
	require.NotEmpty(t, spec.OpenAPI)
	assert.Contains(t, spec.OpenAPI, "3.")

	assertOpenAPIOperationExists(t, spec, "/api/v1/providers", "get")
	assertOpenAPIOperationExists(t, spec, "/api/v1/providers/{slug}/models", "get")
	assertOpenAPIOperationExists(t, spec, "/api/v1/models", "get")
	assertOpenAPIOperationExists(t, spec, "/api/v1/settings/display", "get")

	modelsGet := spec.Paths["/api/v1/models"]["get"]
	assertOpenAPIQueryParam(t, modelsGet, "provider", "")
	assertOpenAPIQueryParam(t, modelsGet, "mode", "")
	assertOpenAPIQueryParam(t, modelsGet, "page", "")
	assertOpenAPIQueryParam(t, modelsGet, "limit", "100")
}

func loadModelDiscoveryOpenAPI(t *testing.T) openAPIDocument {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	require.True(t, ok)

	specPath := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "..", "apps", "model-discovery", "openapi", "model-discovery.openapi.json"))
	body, err := os.ReadFile(specPath)
	require.NoError(t, err, "run `pnpm --dir apps/model-discovery openapi:export` to refresh the spec")

	var spec openAPIDocument
	require.NoError(t, json.Unmarshal(body, &spec))

	return spec
}

func assertOpenAPIOperationExists(t *testing.T, spec openAPIDocument, path string, method string) {
	t.Helper()

	methods, ok := spec.Paths[path]
	require.Truef(t, ok, "expected path %s in model-discovery OpenAPI spec", path)
	_, ok = methods[method]
	require.Truef(t, ok, "expected %s %s in model-discovery OpenAPI spec", method, path)
}

func assertOpenAPIQueryParam(t *testing.T, operation openAPIOperation, name string, maximum string) {
	t.Helper()

	for _, parameter := range operation.Parameters {
		if parameter.In != "query" || parameter.Name != name {
			continue
		}

		if maximum != "" {
			require.NotNilf(t, parameter.Schema.Maximum, "expected query parameter %s to define a maximum", name)
			assert.Equal(t, maximum, trimFloat(*parameter.Schema.Maximum))
		}

		return
	}

	t.Fatalf("expected query parameter %s in OpenAPI operation", name)
}

func trimFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
