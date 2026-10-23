-- +goose Up
-- The official Gemini API has no OpenAI Responses endpoint. Its native
-- generateContent surface is exposed as the generate_content model type
-- instead. Account models copied the catalog's earlier [chat, responses]
-- modes when they were enabled, so this rewrites those stored types to match
-- the corrected catalog.
UPDATE public.tenant_provider_account_models AS m
SET types = (
        SELECT array_agg(DISTINCT model_type ORDER BY model_type)
        FROM unnest(array_remove(m.types, 'responses') || ARRAY['generate_content']::text[]) AS model_type
    ),
    updated_at = now()
FROM public.tenant_provider_accounts AS a
WHERE a.account_id = m.account_id
  AND lower(a.provider) IN ('google-gemini', 'gemini', 'tokenbazaar')
  AND lower(regexp_replace(m.model, '^.*[/:]', '')) IN (
        'gemini-2.5-flash-lite',
        'gemini-2.5-flash',
        'gemini-2.5-pro',
        'gemini-3.1-pro-preview',
        'gemini-3.5-flash-lite',
        'gemini-3.5-flash',
        'gemini-3.6-flash',
        'gemini-3.7-flash'
    )
  AND 'chat' = ANY(m.types);

-- +goose Down
UPDATE public.tenant_provider_account_models AS m
SET types = (
        SELECT array_agg(DISTINCT model_type ORDER BY model_type)
        FROM unnest(array_remove(m.types, 'generate_content') || ARRAY['responses']::text[]) AS model_type
    ),
    updated_at = now()
FROM public.tenant_provider_accounts AS a
WHERE a.account_id = m.account_id
  AND lower(a.provider) IN ('google-gemini', 'gemini', 'tokenbazaar')
  AND lower(regexp_replace(m.model, '^.*[/:]', '')) IN (
        'gemini-2.5-flash-lite',
        'gemini-2.5-flash',
        'gemini-2.5-pro',
        'gemini-3.1-pro-preview',
        'gemini-3.5-flash-lite',
        'gemini-3.5-flash',
        'gemini-3.6-flash',
        'gemini-3.7-flash'
    )
  AND 'generate_content' = ANY(m.types);
