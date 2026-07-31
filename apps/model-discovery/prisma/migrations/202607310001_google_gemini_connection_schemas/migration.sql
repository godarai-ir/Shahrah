-- Route-aware Google Gemini connection schemas.
-- Model reconciliation is intentionally owned by the dedicated YAML importer.

-- API keys use different wire authentication on the two documented surfaces:
-- x-goog-api-key for native Gemini and Bearer for OpenAI compatibility.
INSERT INTO "ProviderConnectionSchema" (
  "id", "providerId", "authMode", "version", "runtimeReady",
  "fieldsJson", "createdAt", "updatedAt"
)
SELECT
  'google-gemini-api-key-v2', p."id", 'api_key', 2, true,
  '[
    {"key":"api_key","label":"Gemini API key","type":"password","secret":true,"required":true,"scope":"credential","runtime_mapping":"google_gemini.api_key","documentation_url":"https://ai.google.dev/gemini-api/docs/api-key","redaction_policy":"write_only"},
    {"key":"project_id","label":"Google Cloud project ID","type":"text","secret":false,"required":false,"scope":"connection","runtime_mapping":"project_id","validation":{"pattern":"^[a-z][a-z0-9-]{4,61}[a-z0-9]$"},"redaction_policy":"identifier"},
    {"key":"native_base_url","label":"Native Gemini API base URL","type":"url","secret":false,"required":false,"scope":"connection","runtime_mapping":"google_gemini.native_base_url","validation":{"default":"https://generativelanguage.googleapis.com","allowed_hosts":["generativelanguage.googleapis.com"]},"documentation_url":"https://ai.google.dev/api","redaction_policy":"public"},
    {"key":"openai_compatible_base_url","label":"OpenAI-compatible base URL","type":"url","secret":false,"required":false,"scope":"connection","runtime_mapping":"google_gemini.openai_base_url","validation":{"default":"https://generativelanguage.googleapis.com/v1beta/openai","allowed_hosts":["generativelanguage.googleapis.com"]},"documentation_url":"https://ai.google.dev/gemini-api/docs/openai","redaction_policy":"public"}
  ]'::jsonb,
  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM "Provider" p
WHERE p."slug" = 'google-gemini'
ON CONFLICT ("providerId", "authMode", "version") DO UPDATE SET
  "runtimeReady" = EXCLUDED."runtimeReady",
  "fieldsJson" = EXCLUDED."fieldsJson",
  "updatedAt" = CURRENT_TIMESTAMP;

-- Service-account credentials are a durable token source, not an OAuth
-- access token. Keep this schema non-routable until token refresh is wired.
INSERT INTO "ProviderConnectionSchema" (
  "id", "providerId", "authMode", "version", "runtimeReady",
  "fieldsJson", "createdAt", "updatedAt"
)
SELECT
  'google-gemini-service-account-v1', p."id", 'google_service_account', 1, false,
  '[
    {"key":"service_account_json","label":"Service account JSON","type":"password","secret":true,"required":true,"scope":"credential","runtime_mapping":"google_service_account.source","documentation_url":"https://ai.google.dev/gemini-api/docs/oauth","redaction_policy":"write_only"},
    {"key":"project_id","label":"Google Cloud project ID","type":"text","secret":false,"required":true,"scope":"connection","runtime_mapping":"project_id","redaction_policy":"identifier"},
    {"key":"oauth_scopes","label":"OAuth scopes","type":"multi_select","secret":false,"required":true,"scope":"connection","runtime_mapping":"google_service_account.scopes","validation":{"default":["https://www.googleapis.com/auth/generative-language"]},"documentation_url":"https://ai.google.dev/gemini-api/docs/oauth","redaction_policy":"public"},
    {"key":"native_base_url","label":"Native Gemini API base URL","type":"url","secret":false,"required":false,"scope":"connection","runtime_mapping":"google_gemini.native_base_url","validation":{"default":"https://generativelanguage.googleapis.com","allowed_hosts":["generativelanguage.googleapis.com"]},"redaction_policy":"public"}
  ]'::jsonb,
  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM "Provider" p
WHERE p."slug" = 'google-gemini'
ON CONFLICT ("providerId", "authMode", "version") DO UPDATE SET
  "runtimeReady" = EXCLUDED."runtimeReady",
  "fieldsJson" = EXCLUDED."fieldsJson",
  "updatedAt" = CURRENT_TIMESTAMP;
