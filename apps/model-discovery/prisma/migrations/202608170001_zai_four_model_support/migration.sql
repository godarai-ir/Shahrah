-- Z.ai direct provider identity and write-only API-key connection schema.
-- Exact model reconciliation remains owned by the curated YAML importer.
INSERT INTO "Provider" (
  "id", "slug", "name", "docsUrl", "websiteUrl", "status", "enabled",
  "supportsByok", "supportsRegionalRouting", "createdAt", "updatedAt"
)
VALUES (
  'provider_zai', 'zai', 'Z.ai',
  'https://docs.z.ai/', 'https://z.ai/',
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
  'zai-api-key-v1', p."id", 'api_key', 1, true,
  '[
    {"key":"api_key","label":"Z.ai API key","type":"password","secret":true,"required":true,"scope":"credential","runtime_mapping":"authorization_bearer","documentation_url":"https://docs.z.ai/guides/develop/http/introduction","redaction_policy":"write_only"},
    {"key":"base_url","label":"Z.ai general API base URL","type":"url","secret":false,"required":false,"scope":"connection","runtime_mapping":"base_url","validation":{"default":"https://api.z.ai/api/paas/v4","allowed_hosts":["api.z.ai"]},"documentation_url":"https://docs.z.ai/guides/develop/http/introduction","redaction_policy":"public"}
  ]'::jsonb,
  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM "Provider" p
WHERE p."slug" = 'zai'
ON CONFLICT ("providerId", "authMode", "version") DO UPDATE SET
  "runtimeReady" = EXCLUDED."runtimeReady",
  "fieldsJson" = EXCLUDED."fieldsJson",
  "updatedAt" = CURRENT_TIMESTAMP;
