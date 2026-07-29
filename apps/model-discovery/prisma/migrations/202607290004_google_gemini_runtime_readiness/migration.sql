-- Google Gemini runtime and pricing readiness.
--
-- Sources:
--   https://ai.google.dev/gemini-api/docs/openai
--   https://ai.google.dev/gemini-api/docs/pricing
--   https://ai.google.dev/gemini-api/docs/deprecations
--   https://ai.google.dev/gemini-api/docs/interactions-overview
--   https://ai.google.dev/gemini-api/docs/live-api
--
-- The OpenAI compatibility endpoint uses bearer authentication. Native
-- Interactions and Live transports continue to use the same stored API key;
-- the runtime adapter selects the documented transport-specific header.
UPDATE "ProviderConnectionSchema" pcs
SET "fieldsJson" = (
      SELECT jsonb_agg(
        CASE
          WHEN field->>'key' = 'api_key'
            THEN jsonb_set(field, '{runtime_mapping}', '"authorization_bearer"'::jsonb)
          WHEN field->>'key' = 'base_url'
            THEN jsonb_set(
              field,
              '{validation,default}',
              '"https://generativelanguage.googleapis.com/v1beta/openai"'::jsonb,
              true
            )
          ELSE field
        END
        ORDER BY ordinal
      )
      FROM jsonb_array_elements(pcs."fieldsJson")
           WITH ORDINALITY AS fields(field, ordinal)
    ),
    "updatedAt" = CURRENT_TIMESTAMP
FROM "Provider" p
WHERE pcs."providerId" = p."id"
  AND p."slug" = 'google-gemini'
  AND pcs."authMode" = 'api_key';

-- These shutdowns and replacements are explicitly listed in Google's
-- deprecation schedule. A shutdown model must not remain routable.
WITH lifecycle(slug, retirement_at, replacement) AS (
  VALUES
    ('gemini-3.1-flash-image-preview', TIMESTAMP '2026-06-25 00:00:00', 'google-gemini/gemini-3.1-flash-image'),
    ('gemini-3-pro-image-preview', TIMESTAMP '2026-06-25 00:00:00', 'google-gemini/gemini-3-pro-image'),
    ('gemini-3.1-flash-lite-preview', TIMESTAMP '2026-05-25 00:00:00', 'google-gemini/gemini-3.1-flash-lite')
)
UPDATE "Model" m
SET "status" = 'unavailable',
    "enabled" = false,
    "lifecycleStatus" = 'archived',
    "retirementDate" = lifecycle.retirement_at,
    "replacementCanonicalId" = lifecycle.replacement,
    "updatedAt" = CURRENT_TIMESTAMP
FROM lifecycle
JOIN "Provider" p ON p."slug" = 'google-gemini'
WHERE m."providerId" = p."id"
  AND m."slug" = lifecycle.slug;

UPDATE "Model" m
SET "status" = 'deprecated',
    "lifecycleStatus" = 'deprecated',
    "deprecationDate" = COALESCE(m."deprecationDate", TIMESTAMP '2026-06-15 00:00:00'),
    "retirementDate" = TIMESTAMP '2026-08-17 00:00:00',
    "replacementCanonicalId" = 'google-gemini/gemini-3.1-flash-image',
    "officialDocsUrl" = 'https://ai.google.dev/gemini-api/docs/imagen',
    "updatedAt" = CURRENT_TIMESTAMP
FROM "Provider" p
WHERE m."providerId" = p."id"
  AND p."slug" = 'google-gemini'
  AND m."slug" = 'imagen-4.0-generate-001';

-- Keep officialDocsUrl on Google-owned documentation for the models served by
-- the Gemini API.
UPDATE "Model" m
SET "officialDocsUrl" = 'https://ai.google.dev/gemma/docs/core/gemma_on_gemini_api',
    "updatedAt" = CURRENT_TIMESTAMP
FROM "Provider" p
WHERE m."providerId" = p."id"
  AND p."slug" = 'google-gemini'
  AND m."slug" LIKE 'gemma-%';

-- The current hosted Gemma documentation explicitly lists only the two Gemma
-- 4 model codes below as supported by the Gemini API. Keep older Gemma 3/3n
-- catalog records for history, but do not advertise them as routable.
UPDATE "Model" m
SET "status" = 'unavailable',
    "enabled" = false,
    "lifecycleStatus" = 'archived',
    "updatedAt" = CURRENT_TIMESTAMP
FROM "Provider" p
WHERE m."providerId" = p."id"
  AND p."slug" = 'google-gemini'
  AND (
    m."slug" LIKE 'gemma-3-%'
    OR m."slug" LIKE 'gemma-3n-%'
  );

-- Runtime profiles are generated from the catalog's explicit mode,
-- capability, modality, parameter, and limit records. Endpoint selection is a
-- stable gateway mapping; no undocumented model capability is introduced.
INSERT INTO "ModelRuntimeProfile" (
  "id", "modelId", "version", "profileJson", "sourceUrl",
  "effectiveFrom", "createdAt"
)
SELECT
  'gemini_runtime_v1_' || m."id",
  m."id",
  1,
  jsonb_build_object(
    'version', 1,
    'endpoints', (
      SELECT jsonb_agg(endpoint ORDER BY endpoint)
      FROM (
        SELECT DISTINCT endpoint
        FROM unnest(
          ARRAY['models']::text[] ||
          CASE
            WHEN m."slug" = 'aqa' THEN ARRAY['generate_answer']::text[]
            WHEN m."slug" LIKE 'deep-research-%' THEN ARRAY['interactions']::text[]
            WHEN m."slug" LIKE 'gemma-%' THEN ARRAY['generate_content']::text[]
            WHEN m."slug" = 'gemini-embedding-2-preview' THEN ARRAY['embeddings']::text[]
            WHEN m."slug" = 'gemini-embedding-2' THEN ARRAY['embed_content']::text[]
            WHEN 'image' = ANY(m."mode") THEN
              CASE
                WHEN m."slug" IN (
                  'gemini-2.5-flash-image',
                  'gemini-3-pro-image-preview'
                ) THEN ARRAY['image_generations']::text[]
                WHEN m."slug" LIKE 'imagen-%' THEN ARRAY['imagen_generate']::text[]
                ELSE ARRAY['generate_content']::text[]
              END
            ELSE
              CASE WHEN m."mode" && ARRAY['chat', 'completion']::text[] THEN ARRAY['chat_completions']::text[] ELSE ARRAY[]::text[] END ||
              CASE WHEN 'embedding' = ANY(m."mode") THEN ARRAY['embeddings']::text[] ELSE ARRAY[]::text[] END ||
              CASE WHEN 'text_to_speech' = ANY(m."mode") THEN ARRAY['interactions']::text[] ELSE ARRAY[]::text[] END ||
              CASE WHEN 'video' = ANY(m."mode") THEN ARRAY['videos']::text[] ELSE ARRAY[]::text[] END ||
              CASE WHEN 'responses' = ANY(m."mode") THEN ARRAY['interactions']::text[] ELSE ARRAY[]::text[] END ||
              CASE WHEN 'realtime' = ANY(m."mode") THEN ARRAY['live']::text[] ELSE ARRAY[]::text[] END
          END ||
          CASE WHEN m."slug" IN (
            'gemini-3.1-flash-lite',
            'gemini-3.1-pro-preview',
            'gemini-3.1-pro-preview-customtools',
            'gemini-3-flash-preview',
            'gemini-flash-latest',
            'gemini-2.5-flash-image',
            'gemini-2.5-flash-lite',
            'gemini-flash-lite-latest',
            'gemini-robotics-er-1.6-preview',
            'gemini-2.5-flash-preview-tts',
            'gemini-2.5-pro-preview-tts',
            'gemini-3.1-flash-tts-preview',
            'gemini-embedding-2',
            'gemini-embedding-2-preview'
          ) THEN ARRAY['batches']::text[] ELSE ARRAY[]::text[] END
        ) endpoint
      ) endpoints
    ),
    'thinking', jsonb_build_object(
      'defaultMode', CASE WHEN EXISTS (
        SELECT 1
        FROM "ModelSupportedParameter" msp
        JOIN "SupportedParameter" sp ON sp."id" = msp."supportedParameterId"
        WHERE msp."modelId" = m."id"
          AND sp."slug" IN ('thinking', 'reasoning_effort')
      ) THEN 'manual' ELSE 'disabled' END,
      'allowedModes', CASE WHEN EXISTS (
        SELECT 1
        FROM "ModelSupportedParameter" msp
        JOIN "SupportedParameter" sp ON sp."id" = msp."supportedParameterId"
        WHERE msp."modelId" = m."id"
          AND sp."slug" IN ('thinking', 'reasoning_effort')
      ) THEN '["manual"]'::jsonb ELSE '["disabled"]'::jsonb END,
      'manualDeprecated', false,
      'effortValues', '[]'::jsonb
    ),
    'assistantPrefill', EXISTS (
      SELECT 1
      FROM "ModelCapability" mc
      JOIN "Capability" c ON c."id" = mc."capabilityId"
      WHERE mc."modelId" = m."id" AND c."slug" = 'assistant_prefill'
    ),
    'sampling', jsonb_build_object(
      'parameters', COALESCE((
        SELECT jsonb_agg(sp."slug" ORDER BY sp."slug")
        FROM "ModelSupportedParameter" msp
        JOIN "SupportedParameter" sp ON sp."id" = msp."supportedParameterId"
        WHERE msp."modelId" = m."id"
          AND sp."slug" IN ('temperature', 'top_p', 'top_k')
      ), '[]'::jsonb),
      'disallowedWithThinking', '[]'::jsonb
    ),
    'tools', jsonb_build_object(
      'custom', EXISTS (
        SELECT 1 FROM "ModelCapability" mc JOIN "Capability" c ON c."id" = mc."capabilityId"
        WHERE mc."modelId" = m."id" AND c."slug" = 'function_calling'
      ),
      'server', EXISTS (
        SELECT 1 FROM "ModelCapability" mc JOIN "Capability" c ON c."id" = mc."capabilityId"
        WHERE mc."modelId" = m."id" AND c."slug" = 'code_execution'
      ),
      'parallel', EXISTS (
        SELECT 1 FROM "ModelCapability" mc JOIN "Capability" c ON c."id" = mc."capabilityId"
        WHERE mc."modelId" = m."id" AND c."slug" = 'parallel_function_calling'
      ),
      'structuredOutput', EXISTS (
        SELECT 1 FROM "ModelCapability" mc JOIN "Capability" c ON c."id" = mc."capabilityId"
        WHERE mc."modelId" = m."id" AND c."slug" IN ('structured_output', 'json_output')
      )
    ),
    'modalities', jsonb_build_object(
      'input', COALESCE((
        SELECT jsonb_agg(
          CASE WHEN mo."slug" = 'doc' THEN 'pdf' ELSE mo."slug" END
          ORDER BY mo."slug"
        )
        FROM "ModelInputModality" mim
        JOIN "Modality" mo ON mo."id" = mim."modalityId"
        WHERE mim."modelId" = m."id"
      ), '["unknown"]'::jsonb),
      'output', COALESCE((
        SELECT jsonb_agg(mo."slug" ORDER BY mo."slug")
        FROM "ModelOutputModality" mom
        JOIN "Modality" mo ON mo."id" = mom."modalityId"
        WHERE mom."modelId" = m."id"
      ), CASE
        WHEN 'embedding' = ANY(m."mode") THEN '["embedding"]'::jsonb
        WHEN 'image' = ANY(m."mode") THEN '["image"]'::jsonb
        WHEN 'video' = ANY(m."mode") THEN '["video"]'::jsonb
        WHEN 'text_to_speech' = ANY(m."mode") OR 'realtime' = ANY(m."mode") THEN '["audio"]'::jsonb
        ELSE '["text"]'::jsonb
      END)
    ),
    'requestControls', jsonb_build_object(
      'serviceTierValues', CASE
        WHEN m."slug" IN (
          'gemini-3.1-flash-lite',
          'gemini-3.1-pro-preview',
          'gemini-3.1-pro-preview-customtools',
          'gemini-3-flash-preview',
          'gemini-flash-latest',
          'gemini-2.5-flash-image',
          'gemini-2.5-flash-lite'
        ) THEN '["standard","flex","priority"]'::jsonb
        ELSE '["standard"]'::jsonb
      END,
      'inferenceGeoValues', '["global"]'::jsonb,
      'speedValues', '[]'::jsonb,
      'effectiveSpeed', '{}'::jsonb
    ),
    'betaHeaders', '[]'::jsonb,
    'betaCapabilities', '[]'::jsonb,
    'limits', jsonb_build_object(
      'contextWindow', COALESCE(m."contextWindow", 0),
      'maxOutputTokens', COALESCE(m."maxOutputTokens", 0),
      'messagesBytes', 33554432,
      'batchBytes', 268435456,
      'fileBytes', 524288000
    )
  ),
  COALESCE(m."officialDocsUrl", 'https://ai.google.dev/gemini-api/docs/models'),
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP
FROM "Model" m
JOIN "Provider" p ON p."id" = m."providerId"
WHERE p."slug" = 'google-gemini'
ON CONFLICT ("modelId", "version") DO UPDATE SET
  "profileJson" = EXCLUDED."profileJson",
  "sourceUrl" = EXCLUDED."sourceUrl";

-- Google explicitly moved gemini-flash-latest to Gemini 3.5 Flash on
-- 2026-05-19. Replace the imported Gemini 3 Flash prices with the current
-- alias target's official Standard rates before normalizing the flat table.
UPDATE "ModelPricing" mp
SET "inputPricePerMillionTokens" = 1.50,
    "outputPricePerMillionTokens" = 9.00,
    "cachedInputPricePerMillionTokens" = 0.15,
    "audioInputPrice" = 0.00000150,
    "sourceUrl" = 'https://ai.google.dev/gemini-api/docs/pricing',
    "updatedAt" = CURRENT_TIMESTAMP
FROM "Model" m
JOIN "Provider" p ON p."id" = m."providerId"
WHERE mp."modelId" = m."id"
  AND p."slug" = 'google-gemini'
  AND m."slug" = 'gemini-flash-latest'
  AND mp."effectiveTo" IS NULL;

-- Replace imported flat rates with auditable, unit-normalized official rates.
DELETE FROM "ModelPricingRate" r
USING "Model" m, "Provider" p
WHERE r."modelId" = m."id"
  AND m."providerId" = p."id"
  AND p."slug" = 'google-gemini';

-- Standard rates already present in the legacy flat table are normalized to
-- per-unit rates. Agent models are excluded because Google documents their
-- cost as underlying-model plus tool usage, not as a fixed token tariff.
WITH latest AS (
  SELECT DISTINCT ON (mp."modelId") mp.*
  FROM "ModelPricing" mp
  ORDER BY mp."modelId", mp."effectiveFrom" DESC NULLS LAST, mp."createdAt" DESC
), rates AS (
  SELECT l."modelId", 'input_tokens' dimension, 'token' unit,
         l."inputPricePerMillionTokens" / 1000000.0 price
  FROM latest l WHERE l."inputPricePerMillionTokens" IS NOT NULL
  UNION ALL
  SELECT l."modelId", 'output_tokens', 'token',
         l."outputPricePerMillionTokens" / 1000000.0
  FROM latest l WHERE l."outputPricePerMillionTokens" IS NOT NULL
  UNION ALL
  SELECT l."modelId", 'cache_read_tokens', 'token',
         l."cachedInputPricePerMillionTokens" / 1000000.0
  FROM latest l WHERE l."cachedInputPricePerMillionTokens" IS NOT NULL
  UNION ALL
  SELECT l."modelId", 'input_images', 'image', l."imageInputPrice"
  FROM latest l WHERE l."imageInputPrice" IS NOT NULL
  UNION ALL
  SELECT l."modelId", 'output_images', 'image', l."imageOutputPrice"
  FROM latest l WHERE l."imageOutputPrice" IS NOT NULL
  UNION ALL
  SELECT l."modelId", 'audio_input_tokens', 'token', l."audioInputPrice"
  FROM latest l WHERE l."audioInputPrice" IS NOT NULL
  UNION ALL
  SELECT l."modelId", 'audio_output_tokens', 'token', l."audioOutputPrice"
  FROM latest l WHERE l."audioOutputPrice" IS NOT NULL
  UNION ALL
  SELECT l."modelId", 'requests', 'request', l."requestPrice"
  FROM latest l WHERE l."requestPrice" IS NOT NULL
)
INSERT INTO "ModelPricingRate" (
  "id", "modelId", "dimension", "unit", "unitPriceUsd", "currency",
  "conditionJson", "region", "effectiveFrom", "sourceUrl",
  "createdAt", "updatedAt"
)
SELECT
  'gemini_rate_' || md5(r."modelId" || ':' || r.dimension || ':standard'),
  r."modelId", r.dimension, r.unit, r.price, 'USD',
  '{"service_tier":"standard"}'::jsonb, 'global',
  CURRENT_TIMESTAMP, 'https://ai.google.dev/gemini-api/docs/pricing',
  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM rates r
JOIN "Model" m ON m."id" = r."modelId"
JOIN "Provider" p ON p."id" = m."providerId"
WHERE m."slug" NOT LIKE 'deep-research-%'
  AND m."slug" <> 'aqa'
  -- Google documents this as a moving alias but does not currently publish
  -- an explicit alias target on the pricing page. Do not reuse a historical
  -- target's tariff.
  AND m."slug" <> 'gemini-flash-lite-latest'
  AND p."slug" = 'google-gemini';

-- Correct/add non-flat Standard pricing dimensions.
WITH official(slug, dimension, unit, price, condition_json) AS (
  VALUES
    ('gemini-2.5-flash-image', 'output_images', 'image', 0.039::numeric, '{"service_tier":"standard","max_resolution":"1024x1024"}'::jsonb),
    ('gemini-3.1-flash-image-preview', 'image_output_tokens', 'token', 0.000060::numeric, '{"service_tier":"standard"}'::jsonb),
    ('gemini-3.1-flash-image-preview', 'output_images_0_5k', 'image', 0.045::numeric, '{"service_tier":"standard","resolution":"0.5k"}'::jsonb),
    ('gemini-3.1-flash-image-preview', 'output_images_1k', 'image', 0.067::numeric, '{"service_tier":"standard","resolution":"1k"}'::jsonb),
    ('gemini-3.1-flash-image-preview', 'output_images_2k', 'image', 0.101::numeric, '{"service_tier":"standard","resolution":"2k"}'::jsonb),
    ('gemini-3.1-flash-image-preview', 'output_images_4k', 'image', 0.151::numeric, '{"service_tier":"standard","resolution":"4k"}'::jsonb),
    ('gemini-3-pro-image-preview', 'image_output_tokens', 'token', 0.000120::numeric, '{"service_tier":"standard"}'::jsonb),
    ('gemini-3-pro-image-preview', 'output_images_1k_2k', 'image', 0.134::numeric, '{"service_tier":"standard","resolution":"1k_or_2k"}'::jsonb),
    ('gemini-3-pro-image-preview', 'output_images_4k', 'image', 0.240::numeric, '{"service_tier":"standard","resolution":"4k"}'::jsonb),
    ('nano-banana-pro-preview', 'image_output_tokens', 'token', 0.000120::numeric, '{"service_tier":"standard"}'::jsonb),
    ('nano-banana-pro-preview', 'output_images_1k_2k', 'image', 0.134::numeric, '{"service_tier":"standard","resolution":"1k_or_2k"}'::jsonb),
    ('nano-banana-pro-preview', 'output_images_4k', 'image', 0.240::numeric, '{"service_tier":"standard","resolution":"4k"}'::jsonb),
    ('lyria-3-clip-preview', 'requests', 'request', 0.040::numeric, '{"duration_seconds":30}'::jsonb),
    ('lyria-3-pro-preview', 'requests', 'request', 0.080::numeric, '{"output":"full_song"}'::jsonb),
    ('veo-3.1-generate-preview', 'output_video_seconds_720p_1080p', 'second', 0.400::numeric, '{"resolution":["720p","1080p"],"audio":true}'::jsonb),
    ('veo-3.1-generate-preview', 'output_video_seconds_4k', 'second', 0.600::numeric, '{"resolution":"4k","audio":true}'::jsonb),
    ('veo-3.1-fast-generate-preview', 'output_video_seconds_720p', 'second', 0.100::numeric, '{"resolution":"720p","audio":true}'::jsonb),
    ('veo-3.1-fast-generate-preview', 'output_video_seconds_1080p', 'second', 0.120::numeric, '{"resolution":"1080p","audio":true}'::jsonb),
    ('veo-3.1-fast-generate-preview', 'output_video_seconds_4k', 'second', 0.300::numeric, '{"resolution":"4k","audio":true}'::jsonb),
    ('veo-3.1-lite-generate-preview', 'output_video_seconds_720p', 'second', 0.050::numeric, '{"resolution":"720p","audio":true}'::jsonb),
    ('veo-3.1-lite-generate-preview', 'output_video_seconds_1080p', 'second', 0.080::numeric, '{"resolution":"1080p","audio":true}'::jsonb),
    ('gemini-embedding-2', 'input_image_tokens', 'token', 0.00000045::numeric, '{"modality":"image"}'::jsonb),
    ('gemini-embedding-2', 'input_audio_seconds', 'second', 0.000160::numeric, '{"modality":"audio"}'::jsonb),
    ('gemini-embedding-2', 'input_video_tokens', 'token', 0.00001200::numeric, '{"modality":"video"}'::jsonb),
    ('gemini-embedding-2', 'input_video_frames', 'frame', 0.000790::numeric, '{"modality":"video"}'::jsonb),
    ('gemini-embedding-2-preview', 'input_image_tokens', 'token', 0.00000045::numeric, '{"modality":"image"}'::jsonb),
    ('gemini-embedding-2-preview', 'input_audio_seconds', 'second', 0.000160::numeric, '{"modality":"audio"}'::jsonb),
    ('gemini-embedding-2-preview', 'input_video_tokens', 'token', 0.00001200::numeric, '{"modality":"video"}'::jsonb),
    ('gemini-embedding-2-preview', 'input_video_frames', 'frame', 0.000790::numeric, '{"modality":"video"}'::jsonb),
    ('gemini-3.1-flash-lite', 'audio_cache_read_tokens', 'token', 0.000000050::numeric, '{"service_tier":"standard","modality":"audio"}'::jsonb),
    ('gemini-3.1-flash-lite', 'cache_storage_token_hours', 'token_hour', 0.000001000::numeric, '{"service_tier":"standard"}'::jsonb),
    ('gemini-3.1-flash-lite-preview', 'audio_cache_read_tokens', 'token', 0.000000050::numeric, '{"service_tier":"standard","modality":"audio"}'::jsonb),
    ('gemini-3.1-flash-lite-preview', 'cache_storage_token_hours', 'token_hour', 0.000001000::numeric, '{"service_tier":"standard"}'::jsonb),
    ('gemini-3.1-pro-preview', 'cache_storage_token_hours', 'token_hour', 0.000004500::numeric, '{"service_tier":"standard"}'::jsonb),
    ('gemini-3.1-pro-preview-customtools', 'cache_storage_token_hours', 'token_hour', 0.000004500::numeric, '{"service_tier":"standard"}'::jsonb),
    ('gemini-3-flash-preview', 'audio_cache_read_tokens', 'token', 0.000000100::numeric, '{"service_tier":"standard","modality":"audio"}'::jsonb),
    ('gemini-3-flash-preview', 'cache_storage_token_hours', 'token_hour', 0.000001000::numeric, '{"service_tier":"standard"}'::jsonb),
    ('gemini-flash-latest', 'cache_storage_token_hours', 'token_hour', 0.000001000::numeric, '{"service_tier":"standard"}'::jsonb),
    ('gemini-2.5-flash-lite', 'audio_cache_read_tokens', 'token', 0.000000030::numeric, '{"service_tier":"standard","modality":"audio"}'::jsonb),
    ('gemini-2.5-flash-lite', 'cache_storage_token_hours', 'token_hour', 0.000001000::numeric, '{"service_tier":"standard"}'::jsonb),
    ('gemini-3.1-flash-live-preview', 'image_video_input_tokens', 'token', 0.000001000::numeric, '{"service_tier":"standard","modality":["image","video"]}'::jsonb),
    ('gemini-2.5-flash-native-audio-preview-12-2025', 'video_input_tokens', 'token', 0.000003000::numeric, '{"service_tier":"standard","modality":"video"}'::jsonb)
)
INSERT INTO "ModelPricingRate" (
  "id", "modelId", "dimension", "unit", "unitPriceUsd", "currency",
  "conditionJson", "region", "effectiveFrom", "sourceUrl",
  "createdAt", "updatedAt"
)
SELECT
  'gemini_rate_' || md5(m."id" || ':' || o.dimension || ':' || o.condition_json::text),
  m."id", o.dimension, o.unit, o.price, 'USD', o.condition_json,
  'global', CURRENT_TIMESTAMP,
  'https://ai.google.dev/gemini-api/docs/pricing',
  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM official o
JOIN "Provider" p ON p."slug" = 'google-gemini'
JOIN "Model" m ON m."providerId" = p."id" AND m."slug" = o.slug;

-- Tiered token rates. Dimension names remain unique so an atomic snapshot can
-- retain every conditional price without collapsing rows.
WITH tiered(slug, tier, input_price_m, audio_input_price_m, output_price_m, cache_price_m, audio_cache_price_m, storage_price_m_h) AS (
  VALUES
    ('gemini-3.1-flash-lite', 'batch', 0.125, 0.25, 0.75, 0.0125, 0.025, 0.50),
    ('gemini-3.1-flash-lite', 'flex', 0.125, 0.25, 0.75, 0.0125, 0.025, 0.50),
    ('gemini-3.1-flash-lite', 'priority', 0.45, 0.90, 2.70, 0.045, 0.09, 1.80),
    ('gemini-3.1-flash-lite-preview', 'batch', 0.125, 0.25, 0.75, 0.0125, 0.025, 0.50),
    ('gemini-3.1-flash-lite-preview', 'flex', 0.125, 0.25, 0.75, 0.0125, 0.025, 0.50),
    ('gemini-3.1-flash-lite-preview', 'priority', 0.45, 0.90, 2.70, 0.045, 0.09, 1.80),
    ('gemini-3.1-pro-preview', 'batch', 1.00, NULL, 6.00, 0.20, NULL, 4.50),
    ('gemini-3.1-pro-preview', 'flex', 1.00, NULL, 6.00, 0.20, NULL, 4.50),
    ('gemini-3.1-pro-preview', 'priority', 3.60, NULL, 21.60, 0.36, NULL, 8.10),
    ('gemini-3.1-pro-preview-customtools', 'batch', 1.00, NULL, 6.00, 0.20, NULL, 4.50),
    ('gemini-3.1-pro-preview-customtools', 'flex', 1.00, NULL, 6.00, 0.20, NULL, 4.50),
    ('gemini-3.1-pro-preview-customtools', 'priority', 3.60, NULL, 21.60, 0.36, NULL, 8.10),
    ('gemini-3-flash-preview', 'batch', 0.25, 0.50, 1.50, 0.05, 0.10, 1.00),
    ('gemini-3-flash-preview', 'flex', 0.25, 0.50, 1.50, 0.05, 0.10, 1.00),
    ('gemini-3-flash-preview', 'priority', 0.90, 1.80, 5.40, 0.09, 0.18, 1.80),
    ('gemini-flash-latest', 'batch', 0.75, 0.75, 4.50, 0.075, 0.075, 1.00),
    ('gemini-flash-latest', 'flex', 0.75, 0.75, 4.50, 0.080, 0.080, 1.00),
    ('gemini-flash-latest', 'priority', 2.70, 2.70, 16.20, 0.27, 0.27, 1.00),
    ('gemini-2.5-flash-lite', 'batch', 0.05, 0.15, 0.20, 0.01, 0.03, 1.00),
    ('gemini-2.5-flash-lite', 'flex', 0.05, 0.15, 0.20, 0.01, 0.03, 1.00),
    ('gemini-2.5-flash-lite', 'priority', 0.18, 0.54, 0.72, 0.018, 0.054, 1.80),
    ('gemini-robotics-er-1.6-preview', 'batch', 0.50, 1.00, 2.50, NULL, NULL, NULL)
), expanded AS (
  SELECT slug, tier, tier || '_input_tokens' dimension, 'token' unit, input_price_m / 1000000.0 price FROM tiered WHERE input_price_m IS NOT NULL
  UNION ALL SELECT slug, tier, tier || '_audio_input_tokens', 'token', audio_input_price_m / 1000000.0 FROM tiered WHERE audio_input_price_m IS NOT NULL
  UNION ALL SELECT slug, tier, tier || '_output_tokens', 'token', output_price_m / 1000000.0 FROM tiered WHERE output_price_m IS NOT NULL
  UNION ALL SELECT slug, tier, tier || '_cache_read_tokens', 'token', cache_price_m / 1000000.0 FROM tiered WHERE cache_price_m IS NOT NULL
  UNION ALL SELECT slug, tier, tier || '_audio_cache_read_tokens', 'token', audio_cache_price_m / 1000000.0 FROM tiered WHERE audio_cache_price_m IS NOT NULL
  UNION ALL SELECT slug, tier, tier || '_cache_storage_token_hours', 'token_hour', storage_price_m_h / 1000000.0 FROM tiered WHERE storage_price_m_h IS NOT NULL
)
INSERT INTO "ModelPricingRate" (
  "id", "modelId", "dimension", "unit", "unitPriceUsd", "currency",
  "conditionJson", "region", "effectiveFrom", "sourceUrl",
  "createdAt", "updatedAt"
)
SELECT
  'gemini_rate_' || md5(m."id" || ':' || e.dimension),
  m."id", e.dimension, e.unit, e.price, 'USD',
  CASE
    WHEN e.tier = 'batch' THEN jsonb_build_object('api', 'batches')
    ELSE jsonb_build_object('service_tier', e.tier)
  END,
  'global',
  CURRENT_TIMESTAMP, 'https://ai.google.dev/gemini-api/docs/pricing',
  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM expanded e
JOIN "Provider" p ON p."slug" = 'google-gemini'
JOIN "Model" m ON m."providerId" = p."id" AND m."slug" = e.slug;

-- Endpoint-specific Batch/Flex/Priority prices use distinct dimensions so the
-- runtime can select them without interpreting free-form conditions.
WITH endpoint_rate(slug, dimension, unit, price, condition_json) AS (
  VALUES
    ('gemini-2.5-flash-preview-tts', 'batch_input_tokens', 'token', 0.000000250::numeric, '{"api":"batches"}'::jsonb),
    ('gemini-2.5-flash-preview-tts', 'batch_audio_output_tokens', 'token', 0.000005000::numeric, '{"api":"batches"}'::jsonb),
    ('gemini-2.5-pro-preview-tts', 'batch_input_tokens', 'token', 0.000000500::numeric, '{"api":"batches"}'::jsonb),
    ('gemini-2.5-pro-preview-tts', 'batch_audio_output_tokens', 'token', 0.000010000::numeric, '{"api":"batches"}'::jsonb),
    ('gemini-3.1-flash-tts-preview', 'batch_input_tokens', 'token', 0.000000500::numeric, '{"api":"batches"}'::jsonb),
    ('gemini-3.1-flash-tts-preview', 'batch_audio_output_tokens', 'token', 0.000010000::numeric, '{"api":"batches"}'::jsonb),
    ('gemini-embedding-2', 'batch_input_tokens', 'token', 0.000000100::numeric, '{"api":"batches","modality":"text"}'::jsonb),
    ('gemini-embedding-2', 'batch_input_image_tokens', 'token', 0.000000225::numeric, '{"api":"batches","modality":"image"}'::jsonb),
    ('gemini-embedding-2', 'batch_input_images', 'image', 0.000060000::numeric, '{"api":"batches","modality":"image"}'::jsonb),
    ('gemini-embedding-2', 'batch_input_audio_tokens', 'token', 0.000003250::numeric, '{"api":"batches","modality":"audio"}'::jsonb),
    ('gemini-embedding-2', 'batch_input_audio_seconds', 'second', 0.000080000::numeric, '{"api":"batches","modality":"audio"}'::jsonb),
    ('gemini-embedding-2', 'batch_input_video_tokens', 'token', 0.000006000::numeric, '{"api":"batches","modality":"video"}'::jsonb),
    ('gemini-embedding-2', 'batch_input_video_frames', 'frame', 0.000395000::numeric, '{"api":"batches","modality":"video"}'::jsonb),
    ('gemini-2.5-flash-image', 'batch_input_tokens', 'token', 0.000000150::numeric, '{"api":"batches"}'::jsonb),
    ('gemini-2.5-flash-image', 'batch_output_tokens', 'token', 0.000001250::numeric, '{"api":"batches","modality":"text"}'::jsonb),
    ('gemini-2.5-flash-image', 'batch_output_images', 'image', 0.019500000::numeric, '{"api":"batches"}'::jsonb),
    ('gemini-2.5-flash-image', 'flex_input_tokens', 'token', 0.000000150::numeric, '{"service_tier":"flex"}'::jsonb),
    ('gemini-2.5-flash-image', 'flex_output_tokens', 'token', 0.000001250::numeric, '{"service_tier":"flex","modality":"text"}'::jsonb),
    ('gemini-2.5-flash-image', 'flex_output_images', 'image', 0.019500000::numeric, '{"service_tier":"flex"}'::jsonb),
    ('gemini-2.5-flash-image', 'priority_input_tokens', 'token', 0.000000540::numeric, '{"service_tier":"priority"}'::jsonb),
    ('gemini-2.5-flash-image', 'priority_output_tokens', 'token', 0.000004500::numeric, '{"service_tier":"priority","modality":"text"}'::jsonb),
    ('gemini-2.5-flash-image', 'priority_output_images', 'image', 0.070200000::numeric, '{"service_tier":"priority"}'::jsonb)
)
INSERT INTO "ModelPricingRate" (
  "id", "modelId", "dimension", "unit", "unitPriceUsd", "currency",
  "conditionJson", "region", "effectiveFrom", "sourceUrl",
  "createdAt", "updatedAt"
)
SELECT
  'gemini_rate_' || md5(m."id" || ':' || e.dimension),
  m."id", e.dimension, e.unit, e.price, 'USD', e.condition_json,
  'global', CURRENT_TIMESTAMP,
  'https://ai.google.dev/gemini-api/docs/pricing',
  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM endpoint_rate e
JOIN "Provider" p ON p."slug" = 'google-gemini'
JOIN "Model" m ON m."providerId" = p."id" AND m."slug" = e.slug;

WITH combined(slug, tier, input_price_m, output_price_m, cache_price_m) AS (
  VALUES
    ('gemini-3.1-pro-preview', 'batch', 2.00, 9.00, 0.40),
    ('gemini-3.1-pro-preview', 'flex', 2.00, 9.00, 0.40),
    ('gemini-3.1-pro-preview', 'priority', 7.20, 32.40, 0.72),
    ('gemini-3.1-pro-preview-customtools', 'batch', 2.00, 9.00, 0.40),
    ('gemini-3.1-pro-preview-customtools', 'flex', 2.00, 9.00, 0.40),
    ('gemini-3.1-pro-preview-customtools', 'priority', 7.20, 32.40, 0.72)
), expanded AS (
  SELECT slug, tier, tier || '_long_context_input_tokens' dimension, input_price_m / 1000000.0 price FROM combined
  UNION ALL SELECT slug, tier, tier || '_long_context_output_tokens', output_price_m / 1000000.0 FROM combined
  UNION ALL SELECT slug, tier, tier || '_long_context_cache_read_tokens', cache_price_m / 1000000.0 FROM combined
)
INSERT INTO "ModelPricingRate" (
  "id", "modelId", "dimension", "unit", "unitPriceUsd", "currency",
  "conditionJson", "region", "effectiveFrom", "sourceUrl",
  "createdAt", "updatedAt"
)
SELECT
  'gemini_rate_' || md5(m."id" || ':' || e.dimension),
  m."id", e.dimension, 'token', e.price, 'USD',
  jsonb_build_object('service_tier', e.tier, 'min_prompt_tokens', 200001),
  'global', CURRENT_TIMESTAMP,
  'https://ai.google.dev/gemini-api/docs/pricing',
  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM expanded e
JOIN "Provider" p ON p."slug" = 'google-gemini'
JOIN "Model" m ON m."providerId" = p."id" AND m."slug" = e.slug;

-- Long-context rates for the two officially documented >200k models present
-- in this catalog.
WITH long_context(slug, input_price_m, output_price_m, cache_price_m) AS (
  VALUES
    ('gemini-3.1-pro-preview', 4.00, 18.00, 0.40),
    ('gemini-3.1-pro-preview-customtools', 4.00, 18.00, 0.40),
    ('gemini-2.5-computer-use-preview-10-2025', 2.50, 15.00, NULL)
), expanded AS (
  SELECT slug, 'long_context_input_tokens' dimension, input_price_m / 1000000.0 price FROM long_context
  UNION ALL SELECT slug, 'long_context_output_tokens', output_price_m / 1000000.0 FROM long_context
  UNION ALL SELECT slug, 'long_context_cache_read_tokens', cache_price_m / 1000000.0 FROM long_context WHERE cache_price_m IS NOT NULL
)
INSERT INTO "ModelPricingRate" (
  "id", "modelId", "dimension", "unit", "unitPriceUsd", "currency",
  "conditionJson", "region", "effectiveFrom", "sourceUrl",
  "createdAt", "updatedAt"
)
SELECT
  'gemini_rate_' || md5(m."id" || ':' || e.dimension),
  m."id", e.dimension, 'token', e.price, 'USD',
  '{"min_prompt_tokens":200001}'::jsonb, 'global',
  CURRENT_TIMESTAMP, 'https://ai.google.dev/gemini-api/docs/pricing',
  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM expanded e
JOIN "Provider" p ON p."slug" = 'google-gemini'
JOIN "Model" m ON m."providerId" = p."id" AND m."slug" = e.slug;
