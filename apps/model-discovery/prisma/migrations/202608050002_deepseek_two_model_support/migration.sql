-- DeepSeek direct provider identity and write-only API-key connection schema.
-- Model reconciliation remains owned by the exact-allowlist YAML importer.
INSERT INTO "Provider" (
  "id", "slug", "name", "docsUrl", "websiteUrl", "status", "enabled",
  "supportsByok", "supportsRegionalRouting", "createdAt", "updatedAt"
)
VALUES (
  'provider_deepseek', 'deepseek', 'DeepSeek',
  'https://api-docs.deepseek.com/', 'https://www.deepseek.com/',
  'active', true, true, false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
)
ON CONFLICT ("slug") DO UPDATE SET
  "name" = EXCLUDED."name",
  "docsUrl" = EXCLUDED."docsUrl",
  "websiteUrl" = EXCLUDED."websiteUrl",
  "status" = EXCLUDED."status",
  "enabled" = EXCLUDED."enabled",
  "supportsByok" = EXCLUDED."supportsByok",
  "supportsRegionalRouting" = EXCLUDED."supportsRegionalRouting",
  "updatedAt" = CURRENT_TIMESTAMP;

INSERT INTO "ProviderConnectionSchema" (
  "id", "providerId", "authMode", "version", "runtimeReady",
  "fieldsJson", "createdAt", "updatedAt"
)
SELECT
  'deepseek-api-key-v1', p."id", 'api_key', 1, true,
  '[
    {"key":"api_key","label":"DeepSeek API key","type":"password","secret":true,"required":true,"scope":"credential","runtime_mapping":"authorization_bearer","documentation_url":"https://api-docs.deepseek.com/api/deepseek-api","redaction_policy":"write_only"},
    {"key":"base_url","label":"DeepSeek API base URL","type":"url","secret":false,"required":false,"scope":"connection","runtime_mapping":"base_url","validation":{"default":"https://api.deepseek.com","allowed_hosts":["api.deepseek.com"]},"documentation_url":"https://api-docs.deepseek.com/quick_start/pricing","redaction_policy":"public"}
  ]'::jsonb,
  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM "Provider" p
WHERE p."slug" = 'deepseek'
ON CONFLICT ("providerId", "authMode", "version") DO UPDATE SET
  "runtimeReady" = EXCLUDED."runtimeReady",
  "fieldsJson" = EXCLUDED."fieldsJson",
  "updatedAt" = CURRENT_TIMESTAMP;
