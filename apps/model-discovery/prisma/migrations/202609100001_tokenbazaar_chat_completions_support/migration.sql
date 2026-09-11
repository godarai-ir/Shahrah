-- TokenBazaar (third-party model aggregator/reseller) provider identity and
-- write-only API-key connection schema, scoped to the documented
-- POST /v1/chat/completions envelope only.
--
-- Unlike Z.ai's connection schema, this one does NOT expose a user-editable
-- base_url field: TokenBazaar's base URL is a single fixed, code-level
-- allowlisted host (https://zx1.deepwl.net; see
-- services/silk-road/pkg/server/tokenbazaar.go and
-- plans/tokenbazaar-production-readiness.md, "Base-URL reachability
-- re-check"). Letting tenants supply an arbitrary base_url here would
-- reopen exactly the host-confusion risk that reachability check exists to
-- close, for a host this session could not independently verify end to end.
--
-- The provider is enabled/selectable here so account creation and the
-- credential connection-test flow can be exercised end to end; live
-- dispatch is independently gated off by the SHAHRAH_TOKENBAZAAR_ENABLED
-- kill switch (defaults to false; see
-- services/control-plane/internal/config/provider_features.go), which is
-- unaffected by this row.
INSERT INTO "Provider" (
  "id", "slug", "name", "docsUrl", "websiteUrl", "status", "enabled",
  "supportsByok", "supportsRegionalRouting", "createdAt", "updatedAt"
)
VALUES (
  'provider_tokenbazaar', 'tokenbazaar', 'TokenBazaar',
  'https://docs.deepwl.cn/zx1/en', 'https://docs.deepwl.cn/zx1/en',
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
  'tokenbazaar-api-key-v1', p."id", 'api_key', 1, true,
  '[
    {"key":"api_key","label":"TokenBazaar API key","type":"password","secret":true,"required":true,"scope":"credential","runtime_mapping":"authorization_bearer","documentation_url":"https://docs.deepwl.cn/zx1/en","redaction_policy":"write_only"}
  ]'::jsonb,
  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM "Provider" p
WHERE p."slug" = 'tokenbazaar'
ON CONFLICT ("providerId", "authMode", "version") DO UPDATE SET
  "runtimeReady" = EXCLUDED."runtimeReady",
  "fieldsJson" = EXCLUDED."fieldsJson",
  "updatedAt" = CURRENT_TIMESTAMP;
