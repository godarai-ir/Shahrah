INSERT INTO "ProviderConnectionSchema" (
  "id", "providerId", "authMode", "version", "runtimeReady",
  "fieldsJson", "createdAt", "updatedAt"
)
SELECT
  'anthropic-wif-v1',
  p."id",
  'anthropic_wif',
  1,
  true,
  '[
    {"key":"federation_rule_id","label":"Federation rule ID","type":"text","secret":false,"required":true,"scope":"connection","runtime_mapping":"anthropic_wif.federation_rule_id","documentation_url":"https://platform.claude.com/docs/en/api/admin/workspaces/get-workspace","redaction_policy":"identifier"},
    {"key":"organization_id","label":"Organization ID","type":"text","secret":false,"required":true,"scope":"connection","runtime_mapping":"anthropic_wif.organization_id","redaction_policy":"identifier"},
    {"key":"service_account_id","label":"Service account ID","type":"text","secret":false,"required":true,"scope":"connection","runtime_mapping":"anthropic_wif.service_account_id","redaction_policy":"identifier"},
    {"key":"workspace_id","label":"Workspace ID","type":"text","secret":false,"required":false,"scope":"connection","runtime_mapping":"anthropic_wif.workspace_id","redaction_policy":"identifier"},
    {"key":"token_source_id","label":"Allowlisted token source","type":"text","secret":false,"required":true,"scope":"connection","runtime_mapping":"anthropic_wif.token_source_id","validation":{"pattern":"^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$"},"redaction_policy":"identifier"},
    {"key":"beta_allowlist","label":"Allowed Anthropic beta capabilities","type":"multi_select","secret":false,"required":false,"scope":"connection","runtime_mapping":"anthropic_beta_allowlist","redaction_policy":"public"}
  ]'::jsonb,
  now(),
  now()
FROM "Provider" p
WHERE p."slug" = 'anthropic'
ON CONFLICT ("providerId", "authMode", "version")
DO UPDATE SET
  "runtimeReady" = EXCLUDED."runtimeReady",
  "fieldsJson" = EXCLUDED."fieldsJson",
  "updatedAt" = now();
