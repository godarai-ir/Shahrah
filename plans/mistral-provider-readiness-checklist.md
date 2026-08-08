# ممیزی آمادگی Mistral AI در گدارAI

> این سند baseline پیش از پیاده‌سازی را نگه می‌دارد. وضعیت جاری شاخه و شواهد آزمون در [گزارش پشتیبانی Mistral AI](../docs/godarai-supported-mistral-ai-models-report.md) ثبت شده است.

تاریخ ممیزی: ۳۰ ژوئیهٔ ۲۰۲۶  
دامنه: کد `control-plane` و `silk-road`، کاتالوگ زندهٔ PostgreSQL و مستندات رسمی Mistral  
مرجع حقیقت: فقط مستندات، OpenAPI، model card، changelog و صفحهٔ رسمی قیمت‌گذاری Mistral

## نتیجهٔ اجرایی

وضعیت Mistral در گدارAI **Partial / Not production-ready** است.

- هیچ مدل Mistral در وضعیت فعلی، زنجیرهٔ کامل «کاتالوگ صحیح + endpoint صحیح + پارامترهای صحیح + اعتبارنامهٔ صحیح + صورتحساب صحیح» را پاس نمی‌کند.
- Chat Completions برای تعدادی از مدل‌های جاری، در زیرمجموعهٔ پایهٔ سازگار با OpenAI قابل عبور است.
- FIM به endpoint اشتباه فرستاده می‌شود.
- Embeddings، Moderation و TTS با slug رسمی دیتابیس یعنی `mistral-ai` در handler رد می‌شوند.
- OCR، Chat Moderation، Batch، Files، Agents، Conversations، Classification و APIهای بومی دیگر route عمومی ندارند.
- ورودی صوتی Voxtral Small در Chat صریحاً رد می‌شود.
- Transcription فقط با فایل آپلودی کار می‌کند؛ `file_url` و `file_id` رسمی پشتیبانی نمی‌شوند.
- دیتابیس زنده ۵۵ رکورد Mistral دارد، ولی چند مدل جاری را ندارد، مدل‌های منسوخ را فعال نگه داشته و برای چند ID هیچ شاهدی در مدل‌های رسمی Mistral وجود ندارد.
- runtime صورتحساب Mistral از قیمت‌های دیتابیس استفاده نمی‌کند و به registry ایستا، ناقص و قدیمی متکی است؛ در مسیر canonical `mistral-ai` عملاً قیمت اکثر یا همهٔ مدل‌ها صفر می‌شود.

این provider نباید تا رفع موارد P0 این گزارش، به‌عنوان «fully supported» یا «billable production provider» عرضه شود.

## منابع رسمی و قاعدهٔ حل تعارض

منابع اصلی:

- [فهرست و وضعیت مدل‌ها](https://docs.mistral.ai/models/overview)
- [OpenAPI رسمی](https://docs.mistral.ai/openapi.yaml)
- [API endpoints](https://docs.mistral.ai/api)
- [قیمت API](https://mistral.ai/pricing/api/)
- [Changelog رسمی](https://docs.mistral.ai/resources/changelogs)
- [API keys](https://docs.mistral.ai/admin/identity-access/api-keys)
- [Models API](https://docs.mistral.ai/api/endpoint/models)
- [Batch API](https://docs.mistral.ai/api/endpoint/batch)
- [Batch processing](https://docs.mistral.ai/studio-api/batch-processing)
- [Audio API](https://docs.mistral.ai/studio-api/audio/overview)
- [OCR](https://docs.mistral.ai/studio-api/document-processing/basic_ocr)

قاعدهٔ پیشنهادی:

1. برای ID، lifecycle، context و capability از model card و Models Overview استفاده شود.
2. برای schema درخواست و endpoint از `openapi.yaml` استفاده شود.
3. برای قیمت، صفحهٔ اختصاصی API Pricing بر model card مقدم باشد.
4. هر رکورد DB باید `sourceUrl`، زمان مشاهده و نسخه/تاریخ مؤثر داشته باشد.
5. اگر دو منبع رسمی تعارض دارند، مقدار نباید بدون ثبت conflict وارد billing production شود.

دو تعارض رسمی مشاهده‌شده:

- API Pricing برای `mistral-moderation-2603` مقدار `$0.1 / M input tokens` نشان می‌دهد، در حالی که model card در زمان ممیزی قیمت صفر نشان می‌داد. برای billing، صفحهٔ API Pricing مرجع مقدم در نظر گرفته شده است.
- API Pricing خروجی متنی Voxtral Small را `$0.4 / M` نشان می‌دهد، ولی model card مقدار `$0.3 / M` دارد. دیتابیس فعلی `$0.3 / M` است و بنابراین بدون policy حل تعارض، قابل تأیید نهایی نیست.
- changelog نام `voxtral-tts-2603` را آورده، ولی model card جاری و pricing از خانوادهٔ `voxtral-mini-tts-*` استفاده می‌کنند. ID جاری model card باید مبنا باشد و این اختلاف در provenance ثبت شود.

## ماتریس endpoint و modality

| قابلیت رسمی Mistral | Endpoint رسمی | modality | وضعیت گدارAI | نتیجه |
|---|---|---|---|---|
| Chat Completions | `POST /v1/chat/completions` | text/image/document/audio بسته به مدل → text | route و adapter عمومی موجود است | ناقص؛ فقط زیرمجموعهٔ پایه |
| FIM / Codestral | `POST /v1/fim/completions` | text/code → code/text | `/v1/completions` گدارAI به `/v1/completions` provider می‌رود | پشتیبانی نمی‌شود |
| Embeddings | `POST /v1/embeddings` | text/code → vector | route موجود، ولی `mistral-ai` در allowlist handler نیست | در مسیر canonical پشتیبانی نمی‌شود |
| Raw Moderation | `POST /v1/moderations` | text → moderation scores | route موجود، ولی Mistral در allowlist نیست | پشتیبانی نمی‌شود |
| Chat Moderation | `POST /v1/chat/moderations` | chat messages → moderation scores | route ندارد | پشتیبانی نمی‌شود |
| OCR / Document AI | `POST /v1/ocr` | PDF/document/image → markdown/annotations/bboxes | route ندارد | پشتیبانی نمی‌شود |
| Audio transcription | `POST /v1/audio/transcriptions` | audio → text/timestamps | route و adapter موجود | ناقص؛ فقط فایل multipart |
| Realtime transcription | `POST /v1/audio/transcriptions` با stream | live audio → streamed text | handler streaming دارد، ولی mode کاتالوگ `realtime` با policy endpoint رد می‌شود | پشتیبانی نمی‌شود |
| Text to speech | `POST /v1/audio/speech` | text/reference audio → audio | route عمومی موجود، ولی Mistral در allowlist نیست | پشتیبانی نمی‌شود |
| Batch | `/v1/batch/jobs...` | درخواست‌های batch برای chat/embedding/FIM/moderation/OCR/classification/conversation/transcription | route ندارد | پشتیبانی نمی‌شود |
| Files | `/v1/files...` | file storage/input | routeهای موجود فقط Anthropic-native هستند | پشتیبانی نمی‌شود |
| Agents | `/v1/agents...`, `/v1/agents/completions` | agent requests → text/tool results | route ندارد | پشتیبانی نمی‌شود |
| Conversations | `/v1/conversations...` | conversation state → text/tool results | route ندارد | پشتیبانی نمی‌شود |
| Classifications | `/v1/classifications`, `/v1/chat/classifications` | text/chat → labels | route ندارد | پشتیبانی نمی‌شود |
| Libraries / Connectors | `/v1/libraries...`, `/v1/connectors...` | documents/search/connectors | route ندارد | پشتیبانی نمی‌شود |
| Model list | `GET /v1/models` | metadata | گدارAI کاتالوگ خودش را برمی‌گرداند؛ passthrough نیست | متفاوت با API رسمی |
| OpenAI Responses | در OpenAPI رسمی Mistral وجود ندارد | — | گدارAI `/responses` را می‌تواند به Mistral بفرستد | قابلیت Mistral محسوب نمی‌شود |
| OpenAI Images | Mistral image generation را به‌صورت built-in agent tool قیمت‌گذاری می‌کند، نه `/v1/images/generations` | text → image | endpoint عمومی OpenAI-style گدارAI معادل رسمی ندارد | پشتیبانی نمی‌شود |

شواهد کد:

- routeهای عمومی: `services/control-plane/internal/server/routes_proxy.go:8-75`
- نگاشت هر دو alias به adapter عمومی: `services/silk-road/pkg/server/adapters.go:45-46`
- مسیر پیش‌فرض Chat و امکان override path: `services/silk-road/pkg/server/openai_compatible.go:97-107`
- مسیر اشتباه FIM: `services/control-plane/internal/proxy/completions_handler.go:245`
- Embeddings فقط alias قدیمی `mistral` را می‌پذیرد: `services/control-plane/internal/proxy/embeddings_handler.go:508-519`
- Moderation فقط OpenAI/OpenRouter/Mock را می‌پذیرد: `services/control-plane/internal/proxy/moderations_handler.go:492-498`
- TTS هیچ alias از Mistral را نمی‌پذیرد: `services/control-plane/internal/proxy/audio_speech_handler.go:789-795`
- Transcription هر دو alias را می‌پذیرد: `services/control-plane/internal/proxy/audio_transcriptions_handler.go:935-953`
- Chat input audio برای Mistral رد می‌شود: `services/control-plane/internal/proxy/handler.go:1919-1925`

## مدل‌های جاری رسمی و وضعیت دقیق

### General chat models

| مدل رسمی و alias | modality رسمی | endpoint | قیمت رسمی | وضعیت DB | وضعیت واقعی گدارAI |
|---|---|---|---|---|---|
| `mistral-medium-3-5`, `mistral-medium-latest` | text/image/document → text | Chat، Conversations، Agents | `$1.5/M` input، `$7.5/M` output | هر دو موجود؛ `latest` هنوز metadata و قیمت Medium 3.1 با context 128k دارد | Chat پایه ناقص؛ قیمت runtime صفر |
| `mistral-small-2603`, `mistral-small-latest` | text/image/document → text | Chat، Conversations، Agents | `$0.15/M` input، `$0.6/M` output | موجود، 256k و قیمت DB صحیح | Chat پایه ناقص؛ قیمت runtime canonical صفر |
| `mistral-large-2512`, `mistral-large-latest` | text/image/document → text | Chat، Conversations، Agents | `$0.5/M` input، `$1.5/M` output | موجود، 256k و قیمت DB صحیح | Chat پایه ناقص؛ registry قدیمی یا صفر |
| `ministral-3b-2512`, `ministral-3b-latest` | text/image → text | Chat، Conversations | `$0.1/M` input/output | موجود؛ context DB برابر 128k است | context رسمی 256k است؛ Chat پایه ناقص |
| `ministral-8b-2512`, `ministral-8b-latest` | text/image → text | Chat، Conversations | `$0.15/M` input/output | موجود، 256k | Chat پایه ناقص؛ قیمت runtime صفر |
| `ministral-14b-2512`, `ministral-14b-latest` | text/image → text | Chat، Conversations | `$0.2/M` input/output | موجود، 256k | Chat پایه ناقص؛ قیمت runtime صفر |
| `labs-leanstral-1-5` | text/code → text/code | Chat، Conversations، Agents | رایگان؛ retirement رسمی 30 Sep 2026 | وجود ندارد؛ DB فقط نسخهٔ منسوخ `labs-leanstral-2603` را دارد | پشتیبانی نمی‌شود |

Chat پایه یعنی text/image، streaming، sampling پایه، tools و structured response در حد contract عمومی. Document input و audio input برای Mistral در validation گدارAI مجاز نیستند.

### Specialist models

| مدل رسمی و alias | modality رسمی | endpoint | قیمت رسمی | وضعیت DB | وضعیت واقعی گدارAI |
|---|---|---|---|---|---|
| `codestral-2508`, `codestral-latest` | text/code → text/code | Chat و FIM | `$0.3/M` input، `$0.9/M` output | موجود؛ context اشتباه 256k | Chat پایه ناقص؛ FIM خراب؛ context رسمی 128k |
| `codestral-embed-2505`, `codestral-embed` | text/code → embedding | Embeddings | `$0.15/M` input | موجود، 8k | handler با `mistral-ai` آن را 501 می‌کند؛ پارامترهای dimension هم ناقص |
| `mistral-embed-2312`, `mistral-embed` | text → embedding | Embeddings | `$0.1/M` input | هیچ‌کدام وجود ندارند | پشتیبانی نمی‌شود |
| `mistral-moderation-2603` | text → moderation scores | Raw/Chat Moderation | طبق API Pricing: `$0.1/M` input | موجود ولی قیمت DB صفر | handler Mistral را رد می‌کند؛ Chat Moderation route ندارد |
| `mistral-ocr-4-0`, `mistral-ocr-latest` | PDF/document/image → markdown/annotations/bboxes | OCR | `$4/1000` pages، `$5/1000` annotated pages | primary وجود ندارد؛ latest موجود ولی هنوز به OCR 3 و 16k اشاره می‌کند | route و billing صفحه‌ای وجود ندارد |
| `mistral-ocr-2512` | PDF/document/image → markdown/annotations/bboxes | OCR | `$2/1000` pages، `$3/1000` annotated pages | موجود با mode=`unknown` و بدون قیمت | قابل route نیست |
| `voxtral-mini-2602`, `voxtral-mini-latest` | audio → text/timestamps | Audio Transcriptions | `$0.003/min` | primary موجود، alias latest وجود ندارد، قیمت ندارد | فقط آپلود فایل؛ URL/ID و billing دقیقه‌ای ناقص |
| `voxtral-mini-transcribe-realtime-2602`, `voxtral-mini-transcribe-realtime-latest` | live audio → streamed text | Audio Transcriptions streaming | `$0.006/min` | primary موجود؛ alias رسمی latest وجود ندارد؛ aliasهای متفاوت و اثبات‌نشده ذخیره شده‌اند | mode=`realtime` به endpoint transcription دسترسی ندارد |
| `voxtral-mini-tts-2603`, `voxtral-mini-tts-latest` | text + optional reference audio → audio | Audio Speech | `$0.016/1000` characters | موجود، بدون قیمت | handler provider را رد می‌کند |
| `voxtral-small-2507`, `voxtral-small-latest` | audio/text → text | Chat | audio `$0.004/min`، text input `$0.1/M`، output طبق API Pricing `$0.4/M` | primary وجود ندارد؛ latest موجود؛ DB فقط text `$0.1/$0.3` دارد | text-only ممکن؛ audio صریحاً رد می‌شود؛ billing audio ندارد |

## پارامترهای خاص Mistral

### Chat Completions

پارامترهای رسمی OpenAPI:

`model`, `messages`, `temperature`, `top_p`, `max_tokens`, `stream`, `stop`,
`random_seed`, `metadata`, `response_format`, `tools`, `tool_choice`,
`presence_penalty`, `frequency_penalty`, `n`, `prediction`,
`parallel_tool_calls`, `reasoning_effort`, `prompt_mode`, `guardrails`,
`prompt_cache_key`, `safe_prompt`.

وضعیت:

- عبور می‌کنند: `model`, `messages`, `temperature`, `top_p`, `max_tokens`, `stream`, `stop`, `response_format`, `tools`, `tool_choice`, `parallel_tool_calls`, `reasoning_effort`.
- حذف می‌شوند چون contract فیلد ندارد: `random_seed`, `presence_penalty`, `frequency_penalty`, `n`, `prediction`, `prompt_mode`, `guardrails`, `prompt_cache_key`, `safe_prompt`.
- `metadata` parse می‌شود، ولی پیش از ارسال به provider عمداً پاک می‌شود.
- DB پارامتر `seed` را تبلیغ می‌کند، در حالی که OpenAPI رسمی `random_seed` می‌خواهد و contract فعلی هیچ‌کدام را به provider نمی‌رساند.
- فیلدهای غیررسمی/عمومی مثل `top_k`, `thinking`, `service_tier`, `speed`, `user` ممکن است وارد payload شوند و برای Mistral قابلیت تأییدشده محسوب نمی‌شوند.
- contract ناشناخته‌های top-level را هنگام unmarshal حذف می‌کند؛ `extra_body` نیز به پارامترهای top-level Mistral expand نمی‌شود.
- file/document content برای provider Mistral توسط allowlist رد می‌شود.
- `input_audio` برای Voxtral Small توسط allowlist رد می‌شود.

شواهد:

- contract: `packages/contracts/openai/types.go:10-36`
- حذف metadata و marshal دوباره: `services/control-plane/internal/proxy/handler.go:535-551`
- validation modality: `services/control-plane/internal/proxy/handler.go:462-503`

### FIM

پارامترهای رسمی:

`model`, `prompt`, `temperature`, `top_p`, `max_tokens`, `stream`, `stop`,
`random_seed`, `metadata`, `suffix`, `min_tokens`, `prompt_cache_key`.

گدارAI endpoint عمومی `/v1/completions` را به `/completions` می‌فرستد، نه `/fim/completions`. در نتیجه حتی اگر `suffix` یا `min_tokens` حفظ شوند، درخواست به endpoint رسمی نمی‌رسد.

### Embeddings

پارامترهای رسمی:

`model`, `input`, `metadata`, `output_dimension`, `output_dtype`,
`encoding_format`.

مشکلات:

- contract گدارAI `dimensions` دارد، نه `output_dimension`.
- custom dimension فقط برای دو مدل OpenAI مجاز است.
- `output_dtype` و `metadata` در contract embeddings وجود ندارند.
- provider allowlist فقط `mistral` را می‌پذیرد، نه slug canonical یعنی `mistral-ai`.

### Audio Transcriptions

پارامترهای رسمی multipart:

`model`, یکی از `file`/`file_url`/`file_id`, `language`, `temperature`,
`stream`, `diarize`, `context_bias`, `timestamp_granularities`.

مشکلات:

- handler الزاماً `multipart/form-data` و فایل فیزیکی `file` می‌خواهد.
- `file_url` و `file_id` به‌تنهایی رد می‌شوند.
- validation عمومی `response_format` و وابستگی `verbose_json` را تحمیل می‌کند؛ این فیلد در schema جاری OpenAPI Mistral وجود ندارد.
- فیلدهای اضافی multipart مانند `diarize` و `context_bias` ممکن است خام عبور کنند، ولی validation و test اختصاصی Mistral ندارند.
- cost estimate از اندازهٔ فایل «توکن صوتی» تخمین می‌زند، در حالی که قیمت رسمی بر دقیقه است.

شواهد: `services/control-plane/internal/proxy/audio_transcriptions_handler.go:533-610,765-805`

### Text to Speech

پارامترهای رسمی:

`input`, `model`, `metadata`, `stream`, `prompt_cache_key`, `voice_id`,
`ref_audio`, `response_format` با enum رسمی `pcm|wav|mp3|flac|opus`.

مشکلات:

- handler فیلد OpenAI-style `voice` را اجباری می‌کند.
- `voice_id` و `ref_audio` رسمی جایگزین requirement گدارAI نمی‌شوند.
- `aac` در گدارAI مجاز است، ولی در enum رسمی Mistral نیست.
- `instructions`, `speed`, `stream_format` مربوط به contract عمومی هستند و پشتیبانی آن‌ها در Mistral تأیید نشده است.
- مهم‌تر از همه، provider Mistral پیش از ارسال 501 می‌شود.

شواهد: `services/control-plane/internal/proxy/audio_speech_handler.go:528-615,679-685,789-795`

### OCR

پارامترهای رسمی:

`model`, `document`, `pages`, `include_image_base64`, `image_limit`,
`image_min_size`, `bbox_annotation_format`, `document_annotation_format`,
`document_annotation_prompt`, `table_format`, `extract_header`,
`extract_footer`, `include_blocks`, `confidence_scores_granularity`.

هیچ contract، route، adapter validation یا billing صفحه‌ای برای این درخواست در گدارAI وجود ندارد. پارامترهای generic Chat که DB برای OCR ثبت کرده (`temperature`, `top_p`, `stream`, `seed`, ...) با schema رسمی OCR منطبق نیستند.

### Batch

OpenAPI رسمی batch این endpointها را می‌پذیرد:

`/v1/chat/completions`, `/v1/embeddings`, `/v1/fim/completions`,
`/v1/moderations`, `/v1/chat/moderations`, `/v1/ocr`,
`/v1/classifications`, `/v1/chat/classifications`,
`/v1/conversations`, `/v1/audio/transcriptions`.

پارامترهای job شامل `endpoint`, `input_files` یا `requests`, `model`,
`agent_id`, `metadata`, `timeout_hours` است. گدارAI route عمومی برای آن ندارد و تخفیف رسمی ۵۰٪ batch را برای Mistral اعمال نمی‌کند.

## قیمت و صورتحساب

### قیمت‌های کلیدی رسمی در زمان ممیزی

| خانواده | قیمت رسمی |
|---|---|
| Mistral Medium 3.5 | `$1.5/M` input، `$7.5/M` output |
| Mistral Small 4 | `$0.15/M` input، `$0.6/M` output |
| Mistral Large 3 | `$0.5/M` input، `$1.5/M` output |
| Ministral 3B / 8B / 14B | به‌ترتیب `$0.1/$0.1`، `$0.15/$0.15`، `$0.2/$0.2` per M tokens |
| Codestral | `$0.3/M` input، `$0.9/M` output |
| Codestral Embed | `$0.15/M` input |
| Mistral Embed | `$0.1/M` input |
| Moderation 2 | طبق API Pricing: `$0.1/M` input |
| OCR 4 | `$4/1000` pages؛ annotated `$5/1000` |
| OCR 3 | `$2/1000` pages؛ annotated `$3/1000` |
| Voxtral Mini Transcribe 2 | `$0.003/min` |
| Voxtral Realtime | `$0.006/min` |
| Voxtral TTS | `$0.016/1000` characters |
| Voxtral Small | audio `$0.004/min`، text input `$0.1/M`، output طبق API Pricing `$0.4/M` |

شرایط رسمی دیگر:

- Batch: پنجاه درصد تخفیف.
- Cached input: نود درصد تخفیف در موارد واجد شرایط.
- Agent API: هزینهٔ مدل به‌علاوهٔ tool call.
- ابزارهای built-in قیمت مستقل دارند؛ از جمله code execution، web search و image generation.

### وضعیت دیتابیس و runtime

- DB زنده فقط `ModelPricing` تخت با ستون‌های token/image/audio/request دارد.
- جدول `ModelPricingRate` و `ModelRuntimeProfile` در DB زنده وجود ندارند.
- آخرین migration اعمال‌شده در DB زنده `202607190001_provider_logo_assets` است؛ migrationهای runtime/pricing مورخ ۲۹ ژوئیه اعمال نشده‌اند.
- مدل‌های transcription، realtime، TTS و OCR قیمت قابل استفاده بر حسب دقیقه/کاراکتر/صفحه ندارند.
- `mistral-medium-latest` در DB هنوز `$0.4/$2` و context 128k دارد؛ این مشخصات Medium 3.1 منسوخ است.
- Codestral در DB با context 256k ثبت شده، در حالی که حد رسمی 128k است.
- Ministral 3B در DB با context 128k ثبت شده، در حالی که حد رسمی 256k است.
- `mistral-moderation-2603` در DB قیمت صفر دارد.

بحرانی‌تر: billing runtime برای Mistral از همین DB هم استفاده نمی‌کند.

- registry ایستا فقط `mistral-large-latest`, `mistral-small-latest`, `mistral-tiny` را دارد.
- قیمت registry برای Large برابر `$2/$6` و برای Small برابر `$0.2/$0.6` است؛ هر دو قدیمی‌اند.
- `mistral-tiny` مدل retired است.
- lookup canonical با provider=`mistral-ai` با registry provider=`mistral` match نمی‌شود.
- modelهای ناشناخته به `priceInfo{}` و هزینهٔ صفر fallback می‌کنند.
- estimate پیش از route برای model ID بدون provider، به provider پیش‌فرض `openai` می‌افتد و برای Mistral قیمت صفر می‌گیرد.

شواهد:

- registry ایستا: `services/control-plane/internal/billing/pricing.go:135-147`
- fallback صفر: `services/control-plane/internal/billing/pricing.go:225-255,302-325`
- billing فقط Anthropic و Google Gemini را از snapshot اختصاصی می‌خواند: `services/control-plane/internal/billing/service.go:1319-1369`
- ساخت model ID نهایی با provider deployment: `services/control-plane/internal/proxy/handler.go:1658-1673`

نتیجه: تا وقتی snapshot/rateهای رسمی Mistral با unitهای token/minute/character/page/request پیاده نشده، صورتحساب Mistral قابل اتکا نیست.

## اعتبارنامه و اتصال provider

### قرارداد رسمی

- secret: یک API key از Mistral Studio/workspace
- header: `Authorization: Bearer <MISTRAL_API_KEY>`
- API base پیش‌فرض: `https://api.mistral.ai/v1`
- probe معتبر: `GET /v1/models`
- API key باید write-only، رمزنگاری‌شده و در logها redacted باشد.

### وضعیت گدارAI

| بخش | وضعیت |
|---|---|
| `ProviderConnectionSchema` DB برای `mistral-ai` | صحیح: `api_key` اجباری، password/write-only/secret، bearer mapping |
| `base_url` | صحیح: optional با default رسمی |
| host allowlist | صحیح: `api.mistral.ai` |
| platform-managed probe | صحیح: canonical `mistral-ai`، `/models`، bearer |
| Silk Road adapter | هر دو alias `mistral` و `mistral-ai` را می‌پذیرد |
| legacy direct adapter | حذف شده؛ مسیر production از Control Plane و Silk Road عبور می‌کند |
| BYOK tenant connection test | ناقص: validator فقط `mistral` را می‌شناسد، نه `mistral-ai` |

شواهد:

- platform credential probe: `services/control-plane/internal/platformcredentials/service.go:394-395,431`
- BYOK validator: `services/control-plane/internal/storage/provider_credential_validation.go:104-118`
- tenant test alias را بدون normalization می‌فرستد: `services/control-plane/internal/storage/tenant_provider_account_repo.go:748,827`
- base URL normalization: `services/control-plane/internal/storage/provider_api_base_validation.go:169-170,204-205`
- legacy direct adapter: removed after confirming it had no production importer

نتیجه: schema اتصال درست است، ولی flow آزمون BYOK canonical یک شکاف اجرایی دارد.

## ممیزی هر ۵۵ رکورد Mistral در DB زنده

### IDهای جاری رسمی که در DB وجود دارند

این رکوردها رسمی و جاری‌اند، ولی وضعیت endpoint/billing آن‌ها مطابق جداول بالا ناقص است:

`codestral-2508`, `codestral-latest`, `codestral-embed-2505`,
`codestral-embed`, `ministral-3b-2512`, `ministral-3b-latest`,
`ministral-8b-2512`, `ministral-8b-latest`, `ministral-14b-2512`,
`ministral-14b-latest`, `mistral-large-2512`, `mistral-large-latest`,
`mistral-medium-3-5`, `mistral-medium-latest`,
`mistral-small-2603`, `mistral-small-latest`,
`mistral-moderation-2603`, `mistral-ocr-2512`,
`mistral-ocr-latest`, `voxtral-mini-2602`,
`voxtral-mini-transcribe-realtime-2602`,
`voxtral-mini-tts-2603`, `voxtral-mini-tts-latest`,
`voxtral-small-latest`.

توضیح مهم: وجود aliasهایی مثل `*-latest` به معنی درستی metadata نیست؛ `mistral-medium-latest` و `mistral-ocr-latest` اکنون به نسخهٔ جدید اشاره می‌کنند، ولی DB metadata نسخهٔ قبلی را نگه داشته است.

### رکوردهای منسوخ یا stale که هنوز enabled هستند

`devstral-2512`, `devstral-latest`, `devstral-medium-2507`,
`devstral-small-2507`, `labs-leanstral-2603`,
`magistral-medium-2509`, `magistral-medium-latest`,
`mistral-large-2411`, `mistral-large-pixtral-2411`,
`mistral-medium-2508`, `mistral-moderation-2411`,
`mistral-moderation-latest`, `mistral-ocr-2505`,
`pixtral-large-2411`, `pixtral-large-latest`,
`voxtral-mini-2507`, `voxtral-mini-transcribe-2507`.

بعض رکوردها `deprecatedAt` دارند، ولی همچنان `status=active` و `enabled=true` هستند. برخی aliasهای `latest` نیز به نسخهٔ منسوخ pin شده‌اند.

### IDهایی که در فهرست/model card رسمی جاری شاهد کافی ندارند

این IDها نباید بدون پاسخ موفق `GET /v1/models` از حساب رسمی و ثبت source رسمی، فعال بمانند:

`mistral-embed-dim128-2510`, `mistral-embed-dim256-2510`,
`mistral-medium-2604`, `mistral-medium-3`, `mistral-medium-3.5`,
`mistral-medium-c21211-r0-75`, `mistral-small`,
`mistral-squarepoint-2602`, `mistral-vibe-cli-fast`,
`mistral-vibe-cli-latest`, `mistral-vibe-cli-with-tools`,
`voxtral-mini-realtime-2602`, `voxtral-mini-realtime-latest`,
`voxtral-mini-tts-260213`.

«بدون شاهد کافی» الزاماً به معنی عدم وجود روی همهٔ حساب‌های Mistral نیست؛ به معنی آن است که طبق شرط این ممیزی، نمی‌توان آن‌ها را با مستندات عمومی رسمی به‌عنوان مدل supported گدارAI تأیید کرد.

### مدل‌های جاری رسمی که در DB نیستند

- `labs-leanstral-1-5`
- `mistral-embed`
- `mistral-embed-2312`
- `mistral-ocr-4-0`
- `voxtral-mini-latest`
- `voxtral-mini-transcribe-realtime-latest`
- `voxtral-small-2507`

## اشکالات اولویت‌بندی‌شده

### P0 — مانع عرضه

1. صورتحساب runtime از قیمت‌های رسمی/DB Mistral استفاده نمی‌کند و در canonical path به صفر می‌افتد.
2. FIM به `/completions` به‌جای `/fim/completions` می‌رود.
3. canonical provider slug بین DB و endpoint handlers ناهماهنگ است.
4. OCR route و contract ندارد.
5. Moderation و TTS با وجود مدل‌های enabled، provider را 501 می‌کنند.
6. lifecycle کاتالوگ مدل‌های منسوخ را فعال نگه داشته است.
7. DB زنده migrationهای unit-normalized pricing/runtime profile را ندارد.

### P1 — پشتیبانی ناقص مهم

1. پارامترهای اختصاصی Chat حذف می‌شوند.
2. Embeddings از نام‌های رسمی `output_dimension` و `output_dtype` پشتیبانی نمی‌کند.
3. Transcription از `file_url` و `file_id` پشتیبانی نمی‌کند.
4. Voxtral Small audio در Chat رد می‌شود.
5. realtime mode به endpoint transcription map نشده است.
6. Batch و تخفیف ۵۰٪ آن وجود ندارد.
7. BYOK credential test برای `mistral-ai` canonical کار نمی‌کند.
8. metadata و context مدل‌های alias جاری stale است.

### P2 — تکمیل پوشش بومی

Agents، Conversations، Files، Chat Moderation، Classifications، Libraries،
Connectors، built-in tool billing و native passthrough رسمی Mistral.

## معیار پذیرش پیشنهادی

Mistral فقط وقتی «supported» اعلام شود که:

1. یک snapshot رسمی model catalog از `GET /v1/models` با provenance ذخیره شود.
2. فقط IDهای دارای model card یا خروجی رسمی account فعال شوند.
3. endpoint capability به‌صورت per-model و نه generic mode map شود.
4. OpenAPI request/response contract برای Chat، FIM، Embeddings، Moderation، OCR، Transcription و TTS تست contract داشته باشد.
5. هر دو alias provider به canonical `mistral-ai` normalize شوند.
6. pricing rateهای unit-normalized برای token/minute/character/page/request و شرایط batch/cache/tool پیاده شوند.
7. مدل بدون قیمت رسمی fail-closed شود؛ هزینهٔ صفر fallback مجاز نباشد.
8. lifecycle aliasها و deprecation با changelog/model overview sync شود.
9. smoke test واقعی با credential رسمی برای هر endpoint و modality پاس شود.
10. خروجی audit نشان دهد کد، DB و منابع رسمی برای model ID، context، endpoint، modality، parameter و price برابرند.

## اعتبارسنجی انجام‌شده

- کاتالوگ زنده PostgreSQL مستقیماً query شد: ۵۵ مدل Mistral.
- migrationهای اعمال‌شدهٔ DB بررسی شد.
- OpenAPI رسمی برای endpoint و schema پارامترها parse شد.
- model cards، overview، pricing و changelog رسمی تطبیق داده شد.
- `go test ./pkg/server` در Silk Road پاس شد.
- تست محدود handlerهای مرتبط در control-plane پاس شد:
  `go test ./internal/proxy -run 'Mistral|Audio|Embedding|Moderation|Completion' -count=1`
- تست‌های pricing/billing مرتبط پاس شدند، ولی این فقط رفتار فعلی ناقص را تأیید می‌کند.
- مجموعهٔ هم‌زمان و کامل‌تر control-plane در مهلت ۶۰ ثانیه تمام نشد؛ failure عملکردی ثبت نشد.

