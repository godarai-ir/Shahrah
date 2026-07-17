# Platform-managed provider credentials: database-derived provider matrix

Date checked: 2026-07-17

## Source of provider scope

The scope is not a hand-written allow-list. It was obtained from the configured
`model-discovery` PostgreSQL database with this value-free query:

```sql
SELECT slug, name, docs_url
FROM "Provider"
WHERE enabled = true AND status = 'active'
ORDER BY slug;
```

It returned 21 providers: `ai21`, `anthropic`, `aws-bedrock`,
`azure-ai-foundry`, `azure-open-ai`, `cerebras`, `cohere`, `databricks`,
`deepgram`, `deepinfra`, `elevenlabs`, `google-gemini`, `google-vertex`, `groq`,
`mistral-ai`, `openai`, `openrouter`, `perplexity-ai`, `sambanova`,
`together-ai`, and `xai`.

No credential value was selected or printed by the query.

## Official-document contract matrix

`Runtime ready` means the current GodarAI request format, auth injection, token
lifetime, and endpoint construction are implemented. A schema can be available
for configuration while remaining non-ready; this prevents the UI or router
from claiming compatibility that does not exist.

| Provider | Stored auth mode | Exact documented injection | Connection-specific fields | Runtime ready | Official evidence |
|---|---|---|---|---:|---|
| AI21 | `api_key` | `Authorization: Bearer <api-key>` | optional `base_url`; default Studio base is provider-owned | no: no documented non-billable credential probe | [Authentication](https://docs.ai21.com/reference/authentication), [Jamba chat endpoint](https://docs.ai21.com/reference/jamba-1-6-api-ref) |
| Anthropic | `api_key` | `x-api-key`; `anthropic-version` is a separate required API header | optional `base_url`, optional `api_version` (default `2023-06-01`) | yes | [Authentication](https://platform.claude.com/docs/en/manage-claude/authentication), [Versions](https://platform.claude.com/docs/en/api/versioning) |
| AWS Bedrock | `bedrock_api_key` | `Authorization: Bearer <Bedrock-api-key>` | required `region`; Mantle base is derived from region | no | [API keys](https://docs.aws.amazon.com/bedrock/latest/userguide/api-keys.html), [Use an API key](https://docs.aws.amazon.com/bedrock/latest/userguide/api-keys-use.html), [Responses API](https://docs.aws.amazon.com/bedrock/latest/userguide/bedrock-mantle.html) |
| Azure AI Foundry | `api_key` | `api-key: <key>` | required `endpoint` | no: inference is implemented, but no documented non-billable credential probe | [Foundry authentication and authorization](https://learn.microsoft.com/en-us/azure/foundry/concepts/authentication-authorization-foundry?view=foundry-classic), [Model inference endpoints](https://learn.microsoft.com/en-us/azure/foundry/foundry-models/concepts/endpoints) |
| Azure OpenAI | `api_key` | `api-key: <key>` | required `endpoint`; optional `api_version` for deployment-style endpoints | yes | [Azure OpenAI REST reference](https://learn.microsoft.com/en-us/azure/ai-foundry/openai/reference), [Chat REST API](https://learn.microsoft.com/en-us/rest/api/microsoft-foundry/azureopenai/chat) |
| Cerebras | `api_key` | `Authorization: Bearer <api-key>` | optional `base_url`; default `https://api.cerebras.ai/v1` | yes | [Authentication](https://inference-docs.cerebras.ai/api-reference/authentication), [API reference](https://inference-docs.cerebras.ai/api-reference) |
| Cohere | `api_key` | `Authorization: Bearer <api-key>` | optional `base_url`; optional `client_name` is non-secret | yes | [Chat API](https://docs.cohere.com/reference/chat), [API keys](https://docs.cohere.com/docs/going-to-production) |
| Databricks | `personal_access_token` | `Authorization: Bearer <workspace-token>` | required `serving_endpoint_url` | yes, non-route-optimized endpoints | [Use foundation models](https://docs.databricks.com/aws/en/machine-learning/model-serving/score-foundation-models), [Query custom endpoints](https://docs.databricks.com/aws/en/machine-learning/model-serving/score-custom-model-endpoints) |
| Deepgram | `api_key` | `Authorization: Token <api-key>` (not Bearer) | optional `base_url` | no | [Authentication](https://developers.deepgram.com/reference/authentication) |
| DeepInfra | `api_key` | `Authorization: Bearer <api-key>` | optional `base_url`; OpenAI base defaults to `https://api.deepinfra.com/v1/openai` | yes | [Authentication](https://deepinfra.com/docs/account/authentication), [API introduction](https://deepinfra.com/docs/api-reference/introduction) |
| ElevenLabs | `api_key` | `xi-api-key: <api-key>` | optional `base_url`; default `https://api.elevenlabs.io/v1` | no | [Authentication](https://elevenlabs.io/docs/api-reference/authentication) |
| Google Gemini API | `api_key` | `x-goog-api-key: <api-key>`; query-key compatibility is not used for new platform connections | optional `base_url`, optional `api_version` | yes | [API keys](https://ai.google.dev/gemini-api/docs/api-key), [REST quickstart](https://ai.google.dev/gemini-api/docs/quickstart?lang=rest) |
| Google Vertex AI | `application_default_credentials` | OAuth 2 access token in `Authorization: Bearer`; ADC/WIF refreshes short-lived tokens | required `project_id`, required `location` | no | [Google Cloud authentication](https://cloud.google.com/docs/authentication), [Vertex OpenAI-compatible endpoint](https://cloud.google.com/vertex-ai/generative-ai/docs/multimodal/call-vertex-using-openai-library) |
| Groq | `api_key` | `Authorization: Bearer <api-key>` | optional `base_url`; default `https://api.groq.com/openai/v1` | yes | [API reference](https://console.groq.com/docs/api-reference), [Models endpoint](https://console.groq.com/docs/api-reference#models) |
| Mistral AI | `api_key` | `Authorization: Bearer <api-key>` | optional `base_url`; default `https://api.mistral.ai/v1` | yes | [API keys](https://docs.mistral.ai/admin/identity-access/api-keys), [Models endpoint](https://docs.mistral.ai/api/endpoint/models) |
| OpenAI | `api_key` | `Authorization: Bearer <api-key>` | optional `base_url`, `organization`, and `project` | yes | [Authentication](https://platform.openai.com/docs/api-reference/authentication), [Models](https://platform.openai.com/docs/api-reference/models/list) |
| OpenRouter | `api_key` | `Authorization: Bearer <api-key>` | optional `base_url`, `http_referer`, and `app_title` | yes | [Authentication](https://openrouter.ai/docs/api/reference/authentication), [API overview](https://openrouter.ai/docs/api/reference/overview) |
| Perplexity | `api_key` | `Authorization: Bearer <api-key>` | optional `base_url`; default `https://api.perplexity.ai` | no: no documented non-billable credential probe | [API groups](https://docs.perplexity.ai/docs/getting-started/api-groups), [API-key management](https://docs.perplexity.ai/docs/admin/api-key-management) |
| SambaNova | `api_key` | `Authorization: Bearer <api-key>` | required `base_url` because SambaCloud and SambaStack URLs differ | no: no documented non-billable credential probe across both products | [Quickstart](https://docs.sambanova.ai/docs/en/get-started/quickstart), [API overview](https://docs.sambanova.ai/docs/en/api-reference/overview) |
| Together AI | `api_key` | `Authorization: Bearer <api-key>` | optional `base_url`; default `https://api.together.xyz/v1` | yes | [API-key authentication](https://docs.together.ai/docs/api-keys-authentication), [API reference](https://docs.together.ai/reference) |
| xAI | `api_key` | `Authorization: Bearer <api-key>` | optional `base_url`; default `https://api.x.ai/v1` | yes | [API-key authentication](https://docs.x.ai/developers/rest-api-reference/management/auth), [REST API](https://docs.x.ai/developers/rest-api-reference) |

## Explicitly represented but not falsely enabled

- AWS SigV4/temporary-role credentials are an official production path, but the
  current stateless tunnel does not sign requests. `aws_sigv4` is a non-ready
  schema until signing and refresh are implemented.
- Azure Entra client credentials are official, but require token acquisition and
  refresh. `entra_client_credentials` is a non-ready schema for both Azure
  providers.
- Databricks recommends OAuth M2M for production, and route-optimized endpoints
  reject PATs. `oauth_m2m` is non-ready until token exchange and refresh exist.
- Vertex ADC/WIF and service-account credentials require short-lived OAuth token
  refresh. Storing one access token as a long-lived secret would be incorrect,
  so Vertex stays non-ready.
- Deepgram and ElevenLabs use provider-specific media protocols that the current
  Silk Road request translators do not implement. Their exact schemas are
  available, but activation is disabled.
- AI21, Azure AI Foundry API-key, Perplexity, and SambaNova inference/auth
  contracts are represented explicitly, but activation stays disabled until a
  provider-documented, non-billable credential-validation operation is
  implemented. GodarAI does not assume that every OpenAI-compatible endpoint
  supports `GET /models`.

## High-throughput boundary

Provider schemas are read from model-discovery only during operator workflows.
Active encrypted credentials are loaded and decrypted by a background manager,
then atomically swapped into a process-local immutable snapshot. Gateway request
resolution reads only that snapshot: no model-discovery request, PostgreSQL
query, Redis call, decryption, or provider probe occurs on the hot path.
