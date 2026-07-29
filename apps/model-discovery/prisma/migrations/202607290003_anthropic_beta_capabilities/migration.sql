UPDATE "ProviderConnectionSchema" schema
SET "fieldsJson" = CASE
  WHEN schema."fieldsJson" @> '[{"key":"beta_allowlist"}]'::jsonb
    THEN (
      SELECT jsonb_agg(
        CASE
          WHEN field->>'key' = 'beta_allowlist'
            THEN field || jsonb_build_object(
              'validation',
              jsonb_build_object(
                'options',
                jsonb_build_array(
                  'files-api-2025-04-14',
                  'output-300k-2026-03-24'
                )
              ),
              'documentation_url',
              'https://platform.claude.com/docs/en/build-with-claude/batch-processing'
            )
          ELSE field
        END
        ORDER BY ordinal
      )
      FROM jsonb_array_elements(schema."fieldsJson")
        WITH ORDINALITY AS entries(field, ordinal)
    )
  ELSE schema."fieldsJson" || jsonb_build_array(
    jsonb_build_object(
      'key', 'beta_allowlist',
      'label', 'Allowed Anthropic beta capabilities',
      'type', 'multi_select',
      'secret', false,
      'required', false,
      'scope', 'connection',
      'runtime_mapping', 'anthropic_beta_allowlist',
      'validation', jsonb_build_object(
        'options',
        jsonb_build_array(
          'files-api-2025-04-14',
          'output-300k-2026-03-24'
        )
      ),
      'documentation_url',
      'https://platform.claude.com/docs/en/build-with-claude/batch-processing',
      'redaction_policy', 'public'
    )
  )
END,
"updatedAt" = now()
FROM "Provider" provider
WHERE schema."providerId" = provider."id"
  AND provider."slug" = 'anthropic'
  AND schema."authMode" IN ('api_key', 'anthropic_wif');
