-- Moonshot/Kimi provider identity and write-only API-key connection schema.
INSERT INTO "Provider" (
  "id", "slug", "name", "docsUrl", "websiteUrl", "status", "enabled",
  "supportsByok", "supportsRegionalRouting", "createdAt", "updatedAt"
)
VALUES (
  'provider_moonshot', 'moonshot', 'Moonshot',
  'https://platform.kimi.ai/docs/llms.txt', 'https://platform.kimi.ai',
  'active', true, true, false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
)
ON CONFLICT ("slug") DO UPDATE SET
  "name" = EXCLUDED."name",
  "docsUrl" = EXCLUDED."docsUrl",
  "websiteUrl" = EXCLUDED."websiteUrl",
  "status" = EXCLUDED."status",
  "enabled" = EXCLUDED."enabled",
  "supportsByok" = EXCLUDED."supportsByok",
  "updatedAt" = CURRENT_TIMESTAMP;

INSERT INTO "ProviderConnectionSchema" (
  "id", "providerId", "authMode", "version", "runtimeReady",
  "fieldsJson", "createdAt", "updatedAt"
)
SELECT
  'moonshot-api-key-v1', p."id", 'api_key', 1, true,
  '[
    {"key":"api_key","label":"Moonshot API key","type":"password","secret":true,"required":true,"scope":"credential","runtime_mapping":"authorization_bearer","documentation_url":"https://platform.kimi.ai/docs/api/overview","redaction_policy":"write_only"},
    {"key":"base_url","label":"Moonshot API base URL","type":"url","secret":false,"required":false,"scope":"connection","runtime_mapping":"base_url","validation":{"default":"https://api.moonshot.ai/v1","allowed_hosts":["api.moonshot.ai"]},"documentation_url":"https://platform.kimi.ai/docs/api/overview","redaction_policy":"public"}
  ]'::jsonb,
  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM "Provider" p
WHERE p."slug" = 'moonshot'
ON CONFLICT ("providerId", "authMode", "version") DO UPDATE SET
  "runtimeReady" = EXCLUDED."runtimeReady",
  "fieldsJson" = EXCLUDED."fieldsJson",
  "updatedAt" = CURRENT_TIMESTAMP;
