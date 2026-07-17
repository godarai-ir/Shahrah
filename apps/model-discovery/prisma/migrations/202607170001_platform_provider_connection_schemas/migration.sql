-- Connection schemas contain field definitions only. The legacy
-- ProviderConnectionProfile table remains temporarily for a redacted audit and
-- operator-led credential rotation; application APIs no longer select it.
CREATE TABLE "ProviderConnectionSchema" (
    "id" TEXT NOT NULL,
    "providerId" TEXT NOT NULL,
    "authMode" TEXT NOT NULL,
    "version" INTEGER NOT NULL,
    "runtimeReady" BOOLEAN NOT NULL DEFAULT false,
    "fieldsJson" JSONB NOT NULL,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "ProviderConnectionSchema_pkey" PRIMARY KEY ("id")
);

CREATE UNIQUE INDEX "ProviderConnectionSchema_providerId_authMode_version_key"
    ON "ProviderConnectionSchema"("providerId", "authMode", "version");
CREATE INDEX "ProviderConnectionSchema_providerId_authMode_version_idx"
    ON "ProviderConnectionSchema"("providerId", "authMode", "version");
ALTER TABLE "ProviderConnectionSchema"
    ADD CONSTRAINT "ProviderConnectionSchema_providerId_fkey"
    FOREIGN KEY ("providerId") REFERENCES "Provider"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- Every row below is an explicit provider contract derived from that provider's
-- official documentation. Similar bearer transports deliberately remain
-- separate rows so endpoint requirements and runtime readiness cannot silently
-- drift together.
WITH schemas(slug, auth_mode, runtime_ready, fields_json) AS (
  VALUES
  ('ai21', 'api_key', false, '[
    {"key":"api_key","label":"API key","type":"password","secret":true,"required":true,"scope":"credential","runtime_mapping":"authorization_bearer","documentation_url":"https://docs.ai21.com/reference/authentication","redaction_policy":"write_only"},
    {"key":"base_url","label":"AI21 Studio base URL","type":"url","secret":false,"required":false,"scope":"connection","runtime_mapping":"base_url","validation":{"default":"https://api.ai21.com/studio/v1"},"documentation_url":"https://docs.ai21.com/reference/jamba-1-6-api-ref"}
  ]'::jsonb),
  ('anthropic', 'api_key', true, '[
    {"key":"api_key","label":"API key","type":"password","secret":true,"required":true,"scope":"credential","runtime_mapping":"x_api_key","documentation_url":"https://platform.claude.com/docs/en/manage-claude/authentication","redaction_policy":"write_only"},
    {"key":"base_url","label":"Anthropic base URL","type":"url","secret":false,"required":false,"scope":"connection","runtime_mapping":"base_url","validation":{"default":"https://api.anthropic.com"}},
    {"key":"api_version","label":"Anthropic API version","type":"text","secret":false,"required":false,"scope":"connection","runtime_mapping":"anthropic_version","validation":{"default":"2023-06-01"},"documentation_url":"https://platform.claude.com/docs/en/api/versioning"}
  ]'::jsonb),
  ('aws-bedrock', 'bedrock_api_key', false, '[
    {"key":"api_key","label":"Amazon Bedrock API key","type":"password","secret":true,"required":true,"scope":"credential","runtime_mapping":"authorization_bearer","documentation_url":"https://docs.aws.amazon.com/bedrock/latest/userguide/api-keys-use.html","redaction_policy":"write_only"},
    {"key":"region","label":"AWS Region","type":"text","secret":false,"required":true,"scope":"connection","runtime_mapping":"aws_region","documentation_url":"https://docs.aws.amazon.com/bedrock/latest/userguide/bedrock-mantle.html"}
  ]'::jsonb),
  ('aws-bedrock', 'aws_sigv4', false, '[
    {"key":"access_key_id","label":"AWS access key ID","type":"text","secret":false,"required":true,"scope":"connection","runtime_mapping":"aws_access_key_id"},
    {"key":"secret_access_key","label":"AWS secret access key","type":"password","secret":true,"required":true,"scope":"credential","runtime_mapping":"aws_secret_access_key","redaction_policy":"write_only"},
    {"key":"session_token","label":"AWS session token","type":"password","secret":true,"required":false,"scope":"credential","runtime_mapping":"aws_session_token","redaction_policy":"write_only"},
    {"key":"region","label":"AWS Region","type":"text","secret":false,"required":true,"scope":"connection","runtime_mapping":"aws_region","documentation_url":"https://docs.aws.amazon.com/bedrock/latest/APIReference/CommonParameters.html"}
  ]'::jsonb),
  ('azure-ai-foundry', 'api_key', false, '[
    {"key":"api_key","label":"Azure AI Foundry API key","type":"password","secret":true,"required":true,"scope":"credential","runtime_mapping":"api_key_header","documentation_url":"https://learn.microsoft.com/en-us/azure/foundry/concepts/authentication-authorization-foundry?view=foundry-classic","redaction_policy":"write_only"},
    {"key":"endpoint","label":"Foundry project/model endpoint","type":"url","secret":false,"required":true,"scope":"connection","runtime_mapping":"base_url"}
  ]'::jsonb),
  ('azure-ai-foundry', 'entra_client_credentials', false, '[
    {"key":"tenant_id","label":"Microsoft Entra tenant ID","type":"text","secret":false,"required":true,"scope":"connection","runtime_mapping":"entra_tenant_id"},
    {"key":"client_id","label":"Microsoft Entra client ID","type":"text","secret":false,"required":true,"scope":"connection","runtime_mapping":"entra_client_id"},
    {"key":"client_secret","label":"Microsoft Entra client secret","type":"password","secret":true,"required":true,"scope":"credential","runtime_mapping":"entra_client_secret","redaction_policy":"write_only"},
    {"key":"endpoint","label":"Foundry project/model endpoint","type":"url","secret":false,"required":true,"scope":"connection","runtime_mapping":"base_url","documentation_url":"https://learn.microsoft.com/en-us/azure/ai-foundry/model-inference/how-to/configure-entra-id"}
  ]'::jsonb),
  ('azure-open-ai', 'api_key', true, '[
    {"key":"api_key","label":"Azure OpenAI API key","type":"password","secret":true,"required":true,"scope":"credential","runtime_mapping":"api_key_header","documentation_url":"https://learn.microsoft.com/en-us/azure/ai-foundry/openai/reference","redaction_policy":"write_only"},
    {"key":"endpoint","label":"Azure OpenAI endpoint","type":"url","secret":false,"required":true,"scope":"connection","runtime_mapping":"base_url"},
    {"key":"api_version","label":"API version","type":"text","secret":false,"required":false,"scope":"connection","runtime_mapping":"api_version","help_text":"Required when the endpoint uses the deployment-style Azure OpenAI API."}
  ]'::jsonb),
  ('azure-open-ai', 'entra_client_credentials', false, '[
    {"key":"tenant_id","label":"Microsoft Entra tenant ID","type":"text","secret":false,"required":true,"scope":"connection","runtime_mapping":"entra_tenant_id"},
    {"key":"client_id","label":"Microsoft Entra client ID","type":"text","secret":false,"required":true,"scope":"connection","runtime_mapping":"entra_client_id"},
    {"key":"client_secret","label":"Microsoft Entra client secret","type":"password","secret":true,"required":true,"scope":"credential","runtime_mapping":"entra_client_secret","redaction_policy":"write_only"},
    {"key":"endpoint","label":"Azure OpenAI endpoint","type":"url","secret":false,"required":true,"scope":"connection","runtime_mapping":"base_url","documentation_url":"https://learn.microsoft.com/en-us/azure/ai-foundry/openai/reference"}
  ]'::jsonb),
  ('cerebras', 'api_key', true, '[
    {"key":"api_key","label":"API key","type":"password","secret":true,"required":true,"scope":"credential","runtime_mapping":"authorization_bearer","documentation_url":"https://inference-docs.cerebras.ai/api-reference/authentication","redaction_policy":"write_only"},
    {"key":"base_url","label":"Cerebras API base URL","type":"url","secret":false,"required":false,"scope":"connection","runtime_mapping":"base_url","validation":{"default":"https://api.cerebras.ai/v1"}}
  ]'::jsonb),
  ('cohere', 'api_key', true, '[
    {"key":"api_key","label":"API key","type":"password","secret":true,"required":true,"scope":"credential","runtime_mapping":"authorization_bearer","documentation_url":"https://docs.cohere.com/reference/chat","redaction_policy":"write_only"},
    {"key":"base_url","label":"Cohere API base URL","type":"url","secret":false,"required":false,"scope":"connection","runtime_mapping":"base_url","validation":{"default":"https://api.cohere.com"}},
    {"key":"client_name","label":"Client name","type":"text","secret":false,"required":false,"scope":"connection","runtime_mapping":"cohere_client_name"}
  ]'::jsonb),
  ('databricks', 'personal_access_token', true, '[
    {"key":"access_token","label":"Databricks personal access token","type":"password","secret":true,"required":true,"scope":"credential","runtime_mapping":"authorization_bearer","documentation_url":"https://docs.databricks.com/aws/en/machine-learning/model-serving/score-custom-model-endpoints","redaction_policy":"write_only"},
    {"key":"serving_endpoint_url","label":"Serving endpoint URL","type":"url","secret":false,"required":true,"scope":"connection","runtime_mapping":"base_url","help_text":"Route-optimized endpoints are not supported by this auth mode."}
  ]'::jsonb),
  ('databricks', 'oauth_m2m', false, '[
    {"key":"client_id","label":"OAuth service-principal client ID","type":"text","secret":false,"required":true,"scope":"connection","runtime_mapping":"oauth_client_id"},
    {"key":"client_secret","label":"OAuth service-principal client secret","type":"password","secret":true,"required":true,"scope":"credential","runtime_mapping":"oauth_client_secret","redaction_policy":"write_only"},
    {"key":"workspace_url","label":"Databricks workspace URL","type":"url","secret":false,"required":true,"scope":"connection","runtime_mapping":"workspace_url","documentation_url":"https://docs.databricks.com/aws/en/dev-tools/auth/oauth-m2m"},
    {"key":"serving_endpoint_url","label":"Serving endpoint URL","type":"url","secret":false,"required":true,"scope":"connection","runtime_mapping":"base_url"}
  ]'::jsonb),
  ('deepgram', 'api_key', false, '[
    {"key":"api_key","label":"Deepgram API key","type":"password","secret":true,"required":true,"scope":"credential","runtime_mapping":"authorization_token","documentation_url":"https://developers.deepgram.com/reference/authentication","redaction_policy":"write_only"},
    {"key":"base_url","label":"Deepgram API base URL","type":"url","secret":false,"required":false,"scope":"connection","runtime_mapping":"base_url","validation":{"default":"https://api.deepgram.com"}}
  ]'::jsonb),
  ('deepinfra', 'api_key', true, '[
    {"key":"api_key","label":"API key or scoped JWT","type":"password","secret":true,"required":true,"scope":"credential","runtime_mapping":"authorization_bearer","documentation_url":"https://deepinfra.com/docs/account/authentication","redaction_policy":"write_only"},
    {"key":"base_url","label":"DeepInfra OpenAI API base URL","type":"url","secret":false,"required":false,"scope":"connection","runtime_mapping":"base_url","validation":{"default":"https://api.deepinfra.com/v1/openai"}}
  ]'::jsonb),
  ('elevenlabs', 'api_key', false, '[
    {"key":"api_key","label":"ElevenLabs API key","type":"password","secret":true,"required":true,"scope":"credential","runtime_mapping":"xi_api_key","documentation_url":"https://elevenlabs.io/docs/api-reference/authentication","redaction_policy":"write_only"},
    {"key":"base_url","label":"ElevenLabs API base URL","type":"url","secret":false,"required":false,"scope":"connection","runtime_mapping":"base_url","validation":{"default":"https://api.elevenlabs.io/v1"}}
  ]'::jsonb),
  ('google-gemini', 'api_key', true, '[
    {"key":"api_key","label":"Gemini API key","type":"password","secret":true,"required":true,"scope":"credential","runtime_mapping":"x_goog_api_key","documentation_url":"https://ai.google.dev/gemini-api/docs/api-key","redaction_policy":"write_only"},
    {"key":"base_url","label":"Gemini API base URL","type":"url","secret":false,"required":false,"scope":"connection","runtime_mapping":"base_url","validation":{"default":"https://generativelanguage.googleapis.com/v1beta"}},
    {"key":"api_version","label":"Gemini API version","type":"text","secret":false,"required":false,"scope":"connection","runtime_mapping":"api_version","validation":{"default":"v1beta"}}
  ]'::jsonb),
  ('google-vertex', 'application_default_credentials', false, '[
    {"key":"project_id","label":"Google Cloud project ID","type":"text","secret":false,"required":true,"scope":"connection","runtime_mapping":"gcp_project_id"},
    {"key":"location","label":"Vertex AI location","type":"text","secret":false,"required":true,"scope":"connection","runtime_mapping":"gcp_location","documentation_url":"https://cloud.google.com/docs/authentication"}
  ]'::jsonb),
  ('google-vertex', 'service_account', false, '[
    {"key":"service_account_json","label":"Service-account JSON","type":"textarea","secret":true,"required":true,"scope":"credential","runtime_mapping":"google_service_account_json","documentation_url":"https://cloud.google.com/iam/docs/keys-create-delete","redaction_policy":"write_only"},
    {"key":"project_id","label":"Google Cloud project ID","type":"text","secret":false,"required":true,"scope":"connection","runtime_mapping":"gcp_project_id"},
    {"key":"location","label":"Vertex AI location","type":"text","secret":false,"required":true,"scope":"connection","runtime_mapping":"gcp_location"}
  ]'::jsonb),
  ('groq', 'api_key', true, '[
    {"key":"api_key","label":"API key","type":"password","secret":true,"required":true,"scope":"credential","runtime_mapping":"authorization_bearer","documentation_url":"https://console.groq.com/docs/api-reference","redaction_policy":"write_only"},
    {"key":"base_url","label":"Groq API base URL","type":"url","secret":false,"required":false,"scope":"connection","runtime_mapping":"base_url","validation":{"default":"https://api.groq.com/openai/v1"}}
  ]'::jsonb),
  ('mistral-ai', 'api_key', true, '[
    {"key":"api_key","label":"API key","type":"password","secret":true,"required":true,"scope":"credential","runtime_mapping":"authorization_bearer","documentation_url":"https://docs.mistral.ai/admin/identity-access/api-keys","redaction_policy":"write_only"},
    {"key":"base_url","label":"Mistral API base URL","type":"url","secret":false,"required":false,"scope":"connection","runtime_mapping":"base_url","validation":{"default":"https://api.mistral.ai/v1"}}
  ]'::jsonb),
  ('openai', 'api_key', true, '[
    {"key":"api_key","label":"API key","type":"password","secret":true,"required":true,"scope":"credential","runtime_mapping":"authorization_bearer","documentation_url":"https://platform.openai.com/docs/api-reference/authentication","redaction_policy":"write_only"},
    {"key":"base_url","label":"OpenAI API base URL","type":"url","secret":false,"required":false,"scope":"connection","runtime_mapping":"base_url","validation":{"default":"https://api.openai.com/v1"}},
    {"key":"organization","label":"OpenAI organization","type":"text","secret":false,"required":false,"scope":"connection","runtime_mapping":"openai_organization"},
    {"key":"project","label":"OpenAI project","type":"text","secret":false,"required":false,"scope":"connection","runtime_mapping":"openai_project"}
  ]'::jsonb),
  ('openrouter', 'api_key', true, '[
    {"key":"api_key","label":"API key","type":"password","secret":true,"required":true,"scope":"credential","runtime_mapping":"authorization_bearer","documentation_url":"https://openrouter.ai/docs/api/reference/authentication","redaction_policy":"write_only"},
    {"key":"base_url","label":"OpenRouter API base URL","type":"url","secret":false,"required":false,"scope":"connection","runtime_mapping":"base_url","validation":{"default":"https://openrouter.ai/api/v1"}},
    {"key":"http_referer","label":"HTTP Referer","type":"url","secret":false,"required":false,"scope":"connection","runtime_mapping":"http_referer"},
    {"key":"app_title","label":"Application title","type":"text","secret":false,"required":false,"scope":"connection","runtime_mapping":"x_openrouter_title"}
  ]'::jsonb),
  ('perplexity-ai', 'api_key', false, '[
    {"key":"api_key","label":"API key","type":"password","secret":true,"required":true,"scope":"credential","runtime_mapping":"authorization_bearer","documentation_url":"https://docs.perplexity.ai/docs/getting-started/api-groups","redaction_policy":"write_only"},
    {"key":"base_url","label":"Perplexity API base URL","type":"url","secret":false,"required":false,"scope":"connection","runtime_mapping":"base_url","validation":{"default":"https://api.perplexity.ai"}}
  ]'::jsonb),
  ('sambanova', 'api_key', false, '[
    {"key":"api_key","label":"API key","type":"password","secret":true,"required":true,"scope":"credential","runtime_mapping":"authorization_bearer","documentation_url":"https://docs.sambanova.ai/docs/en/get-started/quickstart","redaction_policy":"write_only"},
    {"key":"base_url","label":"SambaCloud or SambaStack API base URL","type":"url","secret":false,"required":true,"scope":"connection","runtime_mapping":"base_url","documentation_url":"https://docs.sambanova.ai/docs/en/api-reference/overview"}
  ]'::jsonb),
  ('together-ai', 'api_key', true, '[
    {"key":"api_key","label":"Project API key","type":"password","secret":true,"required":true,"scope":"credential","runtime_mapping":"authorization_bearer","documentation_url":"https://docs.together.ai/docs/api-keys-authentication","redaction_policy":"write_only"},
    {"key":"base_url","label":"Together API base URL","type":"url","secret":false,"required":false,"scope":"connection","runtime_mapping":"base_url","validation":{"default":"https://api.together.xyz/v1"}}
  ]'::jsonb),
  ('xai', 'api_key', true, '[
    {"key":"api_key","label":"Team API key","type":"password","secret":true,"required":true,"scope":"credential","runtime_mapping":"authorization_bearer","documentation_url":"https://docs.x.ai/developers/rest-api-reference/management/auth","redaction_policy":"write_only"},
    {"key":"base_url","label":"xAI API base URL","type":"url","secret":false,"required":false,"scope":"connection","runtime_mapping":"base_url","validation":{"default":"https://api.x.ai/v1"}}
  ]'::jsonb)
)
INSERT INTO "ProviderConnectionSchema"
  ("id", "providerId", "authMode", "version", "runtimeReady", "fieldsJson", "createdAt", "updatedAt")
SELECT
  CONCAT('schema_', p."id", '_', replace(s.auth_mode, '-', '_'), '_v1'),
  p."id",
  s.auth_mode,
  1,
  s.runtime_ready,
  s.fields_json,
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP
FROM schemas s
JOIN "Provider" p ON p."slug" = s.slug
WHERE p."enabled" = true
ON CONFLICT ("providerId", "authMode", "version") DO UPDATE SET
  "runtimeReady" = EXCLUDED."runtimeReady",
  "fieldsJson" = EXCLUDED."fieldsJson",
  "updatedAt" = CURRENT_TIMESTAMP;
