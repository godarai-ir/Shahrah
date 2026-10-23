-- +goose Up
ALTER TABLE public.tenant_guardrails
    DROP CONSTRAINT IF EXISTS tenant_guardrails_type_chk,
    ADD CONSTRAINT tenant_guardrails_type_chk CHECK (guardrail_type IN (
        'keyword_blocklist',
        'regex',
        'secrets_detection',
        'basic_pii',
        'jwt_validation',
        'regex_replace',
        'sentence_count',
        'word_count',
        'character_count',
        'json_schema',
        'json_keys',
        'contains_code',
        'not_null',
        'inline_image_urls',
        'semantic_policy',
        'document_confidentiality'
    ));

-- +goose Down
-- Rolling back removes document confidentiality guardrails, because the older constraint cannot hold them.
DELETE FROM public.tenant_guardrails WHERE guardrail_type = 'document_confidentiality';

ALTER TABLE public.tenant_guardrails
    DROP CONSTRAINT IF EXISTS tenant_guardrails_type_chk,
    ADD CONSTRAINT tenant_guardrails_type_chk CHECK (guardrail_type IN (
        'keyword_blocklist',
        'regex',
        'secrets_detection',
        'basic_pii',
        'jwt_validation',
        'regex_replace',
        'sentence_count',
        'word_count',
        'character_count',
        'json_schema',
        'json_keys',
        'contains_code',
        'not_null',
        'inline_image_urls',
        'semantic_policy'
    ));
