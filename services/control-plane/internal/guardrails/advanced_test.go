package guardrails

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/require"
)

func TestRegexReplaceGuardrailMutatesInput(t *testing.T) {
	g := NewFromTenantGuardrail(platform.TenantGuardrail{
		Slug:        "mask-card",
		Type:        "regex_replace",
		Operation:   platform.GuardrailOperationMutate,
		Enforcement: platform.GuardrailEnforcementEnforce,
		Config:      map[string]interface{}{"pattern": `\d{4}-\d{4}`, "replacement": "[***]"},
	})
	engine := NewDefaultEngine()
	req := contractsopenai.ChatCompletionRequest{Messages: []contractsopenai.ChatMessage{{Role: "user", Content: "card 1234-5678"}}}
	result, err := engine.ApplyInputWithGuardrails(req, RequestSelection{LLMInput: []string{"guardrails/mask-card"}, Scope: ScopeAll}, map[string]Guardrail{"mask_card": g})
	require.NoError(t, err)
	require.Equal(t, "card [***]", result.Request.Messages[0].Content)
}

func TestCountGuardrailsBlockOutOfRange(t *testing.T) {
	g := NewFromTenantGuardrail(platform.TenantGuardrail{
		Slug:        "short-answer",
		Type:        "word_count",
		Operation:   platform.GuardrailOperationValidate,
		Enforcement: platform.GuardrailEnforcementEnforce,
		Config:      map[string]interface{}{"min": float64(1), "max": float64(2)},
	})
	engine := NewDefaultEngine()
	req := contractsopenai.ChatCompletionRequest{Messages: []contractsopenai.ChatMessage{{Role: "user", Content: "one two three"}}}
	_, err := engine.ApplyInputWithGuardrails(req, RequestSelection{LLMInput: []string{"guardrails/short-answer"}, Scope: ScopeAll}, map[string]Guardrail{"short_answer": g})
	require.Error(t, err)
}

func TestJSONSchemaAndKeysGuardrails(t *testing.T) {
	schemaGuardrail := NewFromTenantGuardrail(platform.TenantGuardrail{
		Slug:        "schema",
		Type:        "json_schema",
		Operation:   platform.GuardrailOperationValidate,
		Enforcement: platform.GuardrailEnforcementEnforce,
		Config: map[string]interface{}{"schema": map[string]interface{}{
			"type":     "object",
			"required": []interface{}{"answer"},
			"properties": map[string]interface{}{
				"answer": map[string]interface{}{"type": "string"},
			},
		}},
	})
	keysGuardrail := NewFromTenantGuardrail(platform.TenantGuardrail{
		Slug:        "keys",
		Type:        "json_keys",
		Operation:   platform.GuardrailOperationValidate,
		Enforcement: platform.GuardrailEnforcementEnforce,
		Config:      map[string]interface{}{"mode": "all", "keys": []interface{}{"answer"}},
	})
	engine := NewDefaultEngine()
	resp := contractsopenai.ChatCompletionResponse{Choices: []contractsopenai.Choice{{Message: &contractsopenai.ChatMessage{Content: `{"answer":"ok"}`}}}}
	result, err := engine.ApplyOutputWithGuardrails(resp, RequestSelection{LLMOutput: []string{"guardrails/schema", "guardrails/keys"}, Scope: ScopeAll}, map[string]Guardrail{"schema": schemaGuardrail, "keys": keysGuardrail})
	require.NoError(t, err)
	require.Equal(t, "passed", result.Checks["llm_output_guardrails"][0].Status)
	require.Equal(t, "passed", result.Checks["llm_output_guardrails"][1].Status)
}

func TestContainsCodeAndNotNullGuardrails(t *testing.T) {
	codeGuardrail := NewFromTenantGuardrail(platform.TenantGuardrail{
		Slug:        "needs-code",
		Type:        "contains_code",
		Operation:   platform.GuardrailOperationValidate,
		Enforcement: platform.GuardrailEnforcementEnforce,
		Config:      map[string]interface{}{"languages": []interface{}{"sql"}, "must_contain": true},
	})
	notNullGuardrail := NewFromTenantGuardrail(platform.TenantGuardrail{
		Slug:        "not-null",
		Type:        "not_null",
		Operation:   platform.GuardrailOperationValidate,
		Enforcement: platform.GuardrailEnforcementEnforce,
		Config:      map[string]interface{}{},
	})
	engine := NewDefaultEngine()
	resp := contractsopenai.ChatCompletionResponse{Choices: []contractsopenai.Choice{{Message: &contractsopenai.ChatMessage{Content: "select * from users"}}}}
	_, err := engine.ApplyOutputWithGuardrails(resp, RequestSelection{LLMOutput: []string{"guardrails/needs-code", "guardrails/not-null"}, Scope: ScopeAll}, map[string]Guardrail{"needs_code": codeGuardrail, "not_null": notNullGuardrail})
	require.NoError(t, err)
}

func TestJWTValidationGuardrailValidatesHMAC(t *testing.T) {
	token := signedTestJWT(t, map[string]interface{}{"alg": "HS256", "typ": "JWT"}, map[string]interface{}{"exp": time.Now().Add(time.Hour).Unix(), "iss": "issuer-a", "aud": "api"}, "secret")
	g := NewFromTenantGuardrail(platform.TenantGuardrail{
		Slug:        "jwt",
		Type:        "jwt_validation",
		Operation:   platform.GuardrailOperationValidate,
		Enforcement: platform.GuardrailEnforcementEnforce,
		Config:      map[string]interface{}{"shared_secret": "secret", "issuer": "issuer-a", "audience": "api"},
	})
	engine := NewDefaultEngine()
	req := contractsopenai.ChatCompletionRequest{Messages: []contractsopenai.ChatMessage{{Role: "user", Content: "Bearer " + token}}}
	_, err := engine.ApplyInputWithGuardrails(req, RequestSelection{LLMInput: []string{"guardrails/jwt"}, Scope: ScopeAll}, map[string]Guardrail{"jwt": g})
	require.NoError(t, err)
}

func TestInlineImageURLsGuardrailMutatesTextURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "image/png")
		_, _ = w.Write([]byte{0x89, 'P', 'N', 'G'})
	}))
	t.Cleanup(server.Close)
	g := NewFromTenantGuardrail(platform.TenantGuardrail{
		Slug:        "inline-images",
		Type:        "inline_image_urls",
		Operation:   platform.GuardrailOperationMutate,
		Enforcement: platform.GuardrailEnforcementEnforce,
		Config:      map[string]interface{}{"allow_private_networks": true, "max_bytes": float64(64)},
	})
	engine := NewDefaultEngine()
	req := contractsopenai.ChatCompletionRequest{Messages: []contractsopenai.ChatMessage{{Role: "user", Content: "look " + server.URL + "/image.png"}}}
	result, err := engine.ApplyInputWithGuardrails(req, RequestSelection{LLMInput: []string{"guardrails/inline-images"}, Scope: ScopeAll}, map[string]Guardrail{"inline_images": g})
	require.NoError(t, err)
	require.Contains(t, result.Request.Messages[0].Content, "data:image/png;base64,")
}

func signedTestJWT(t *testing.T, header map[string]interface{}, claims map[string]interface{}, secret string) string {
	t.Helper()
	encode := func(value interface{}) string {
		raw, err := json.Marshal(value)
		require.NoError(t, err)
		return base64.RawURLEncoding.EncodeToString(raw)
	}
	signed := encode(header) + "." + encode(claims)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(signed))
	return signed + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
