# GodarAI Postman Collections

این پوشه برای تست API layer، دیباگ و دموی GodarAI/Shahrah gateway ساخته شده است. کالکشن‌ها از providerهای زنده `model-discovery` در `http://localhost:3002/api/v1/providers` تولید شده‌اند و هیچ secret یا credential داخل فایل‌ها ذخیره نشده است.

## فایل‌ها

- `GodarAI.local.postman_environment.json`: template محیط محلی بدون secret.
- `provider-manifest.json`: snapshot providerها، endpointها و منبع تولید.
- `collections/GodarAI - ai21.postman_collection.json`
- `collections/GodarAI - anthropic.postman_collection.json`
- `collections/GodarAI - aws-bedrock.postman_collection.json`
- `collections/GodarAI - azure-ai-foundry.postman_collection.json`
- `collections/GodarAI - azure-open-ai.postman_collection.json`
- `collections/GodarAI - cerebras.postman_collection.json`
- `collections/GodarAI - cohere.postman_collection.json`
- `collections/GodarAI - databricks.postman_collection.json`
- `collections/GodarAI - deepgram.postman_collection.json`
- `collections/GodarAI - deepinfra.postman_collection.json`
- `collections/GodarAI - elevenlabs.postman_collection.json`
- `collections/GodarAI - google-gemini.postman_collection.json`
- `collections/GodarAI - google-vertex.postman_collection.json`
- `collections/GodarAI - groq.postman_collection.json`
- `collections/GodarAI - mistral-ai.postman_collection.json`
- `collections/GodarAI - openai.postman_collection.json`
- `collections/GodarAI - openrouter.postman_collection.json`
- `collections/GodarAI - perplexity-ai.postman_collection.json`
- `collections/GodarAI - sambanova.postman_collection.json`
- `collections/GodarAI - together-ai.postman_collection.json`
- `collections/GodarAI - xai.postman_collection.json`

## اجرای سریع OpenAI

1. در Postman، environment و کالکشن `GodarAI - OpenAI` را import کنید.
2. مقدار `godarai_api_key` را با PAT tenant یا credential معتبر gateway پر کنید. برای local bootstrap می‌توانید مقدار `MASTER_KEY` را فقط در محیط شخصی خودتان به عنوان bearer token بگذارید، اما آن را در فایل‌ها ذخیره نکنید.
3. `tenant_slug` به صورت پیش‌فرض `tenant1` است.
4. اگر از tenant UI نیاز به ساخت Account-Provider یا خواندن FQNها داشتید، از `http://tenant1.localhost:3000/` استفاده کنید. در صورت نیاز به OTP، صندوق محلی روی `http://localhost:8025/` است.
5. درخواست `Discovery / List Gateway Models` را اجرا کنید و در صورت تفاوت، متغیرهای `openai_*_model_fqn` را با مدل‌های expose شده جایگزین کنید.

## متغیرهای مهم

- `base_url`: آدرس control-plane، پیش‌فرض `http://localhost:4000`.
- `godarai_api_key`: bearer token برای gateway.
- `tenant_slug`: tenant resolver header.
- `openai_account_provider_id` و `openai_account_provider_name`: account-provider فعلی OpenAI در local DB.
- `openai_chat_model_fqn`, `openai_responses_model_fqn`, `openai_embedding_model_fqn`: FQNهای مدل‌های قابل اجرا.
- `sample_image_path` و `sample_audio_path`: فقط برای requestهای multipart.

## Cache

کالکشن OpenAI دو سناریو دارد:

- Simple cache: دو request کاملا یکسان با header `x-godarai-cache-config` و namespace مشترک.
- Smart cache: دو prompt مشابه با `type=smart` و `similarity_threshold=0.85`.

assertionها عمدا diagnostic هستند: اگر headerهای `x-godarai-cache-status`, `x-godarai-cached-trace-id` یا `x-godarai-cache-similarity-score` وجود داشته باشند اعتبارسنجی می‌شوند؛ اگر gateway cache را bypass کند تست فقط وضعیت را ثبت می‌کند.

## Providerهای بدون credential

برای providerهای غیر OpenAI کالکشن placeholder ساخته شده است. تا وقتی `provider_enabled=true`، `provider_account_name` و FQNهای مدل پر نشوند، requestها با pre-request script skip می‌شوند و secretی hard-code نشده است.

## Provider Snapshot

| slug | name | status | enabled | modelCount |
| --- | --- | --- | --- | --- |
| ai21 | Ai21 | active | true | 1 |
| anthropic | Anthropic | active | true | 11 |
| aws-bedrock | Aws Bedrock | active | true | 137 |
| azure-ai-foundry | Azure Ai Foundry | active | true | 68 |
| azure-open-ai | Azure Open Ai | active | true | 124 |
| cerebras | Cerebras | active | true | 4 |
| cohere | Cohere | active | true | 28 |
| databricks | Databricks | active | true | 21 |
| deepgram | Deepgram | active | true | 143 |
| deepinfra | Deepinfra | active | true | 163 |
| elevenlabs | Elevenlabs | active | true | 10 |
| google-gemini | Google Gemini | active | true | 41 |
| google-vertex | Google Vertex | active | true | 301 |
| groq | Groq | active | true | 16 |
| mistral-ai | Mistral Ai | active | true | 55 |
| openai | Openai | active | true | 107 |
| openrouter | Openrouter | active | true | 696 |
| perplexity-ai | Perplexity Ai | active | true | 5 |
| sambanova | Sambanova | active | true | 13 |
| together-ai | Together Ai | active | true | 238 |
| xai | Xai | active | true | 72 |

## بازتولید

```powershell
node scripts/postman/generate-godarai-postman.mjs
```

اگر `model-discovery` در دسترس نباشد، generator از snapshot fallback همین فایل استفاده می‌کند.
