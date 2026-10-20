# شواهد پشتیبانی Mistral AI در گدارAI

تاریخ بررسی: ۳۱ ژوئیهٔ ۲۰۲۶
ارائه‌دهندهٔ اصلی: `mistral-ai`
نام سازگار قدیمی: `mistral`

این سند وضعیت همین شاخه را ثبت می‌کند. مستندات رسمی Mistral مرجع شناسه مدل، نقطه پایانی، قرارداد درخواست و پاسخ و قیمت است. سازگاری ظاهری با OpenAI به‌تنهایی مجوز فعال‌کردن قابلیت نیست.

## مرز کاتالوگ

Importer اختصاصی فقط ۱۰ مدل زیر را می‌پذیرد و در صورت مدل اضافه، مدل کم، شناسه نادرست، منبع غیراصلی، نقطه پایانی ناسازگار یا ابعاد هزینهٔ ناقص، ساخت را متوقف می‌کند.

| مدل | ورودی ← خروجی | نقطه‌های پایانی مجاز |
| --- | --- | --- |
| `codestral-2508` | متن/کد ← متن/کد | `chat_completions`، `fim_completions`، `conversations`، `batches` |
| `codestral-embed-2505` | متن/کد ← بردار | `embeddings`، `batches` |
| `mistral-embed-2312` | متن ← بردار | `embeddings`، `batches` |
| `mistral-large-2512` | متن/تصویر/سند ← متن | `chat_completions`، `conversations`، `agents`، `files`، `batches` |
| `mistral-medium-3-5` | متن/تصویر/سند ← متن | `chat_completions`، `conversations`، `agents`، `files`، `batches` |
| `mistral-moderation-2603` | متن/گفت‌وگو ← نتیجهٔ پالایش | `moderations`، `chat_moderations`، `batches` |
| `mistral-ocr-4-0` | تصویر/PDF/سند ← سند ساخت‌یافته | `ocr`، `files`، `batches` |
| `mistral-small-2603` | متن/تصویر/سند ← متن | `chat_completions`، `conversations`، `agents`، `files`، `batches` |
| `voxtral-mini-2602` | صوت ← متن/زمان‌بندی | `audio_transcriptions`، `batches` |
| `voxtral-mini-tts-2603` | متن/صدای مرجع ← صوت | `audio_speech` |

شناسهٔ اصلی OCR برابر `mistral-ocr-4-0` است. `mistral-ocr-4` و `mistral-ocr-latest` فقط نام مستعار سازگاری‌اند. مدل‌های خارج از جدول از مسیر runtime این ارائه‌دهنده قیمت نمی‌گیرند و نباید تبلیغ یا مسیریابی شوند.

## وضعیت نقطه‌های پایانی

| مسیر گدارAI | قرارداد upstream | وضعیت این شاخه |
| --- | --- | --- |
| `POST /v1/chat/completions` | `POST /v1/chat/completions` | پیاده‌سازی شده؛ پارامترها و modality با پروفایل همان مدل محدود می‌شوند. Function calling فعال است؛ ابزارهای داخلی که تسویه یا API مکمل لازم دارند fail-closed هستند. |
| `POST /v1/fim/completions` | `POST /v1/fim/completions` | پیاده‌سازی شده؛ `suffix`، `min_tokens`، `random_seed` و فیلدهای native حفظ می‌شوند. |
| `POST /v1/embeddings` | `POST /v1/embeddings` | پیاده‌سازی شده؛ `output_dimension`/`output_dtype` فقط برای `codestral-embed-2505` پذیرفته و حفظ می‌شوند. `mistral-embed-2312` مطابق سند رسمی بردار ثابت ۱۰۲۴بعدی دارد. |
| `POST /v1/moderations` | `POST /v1/moderations` | پیاده‌سازی شده. |
| `POST /v1/chat/moderations` | `POST /v1/chat/moderations` | پیاده‌سازی شده. |
| `POST /v1/ocr` | `POST /v1/ocr` | پیاده‌سازی شده؛ union رسمی `document`، انتخاب صفحه به‌شکل آرایه یا range string، enumها و وابستگی annotation prompt اعتبارسنجی و هزینه بر پایهٔ صفحه تسویه می‌شود. |
| `POST /v1/audio/transcriptions` | `POST /v1/audio/transcriptions` | پیاده‌سازی شده؛ upload با multipart یا ارجاع `file_url`/`file_id` با JSON پذیرفته می‌شود. مصرف رسمی `prompt_audio_seconds` مبنای تسویه است. سقف فعلی ۶۰ دقیقه و ۵۰۰ MB اعمال‌شده در رزرو/پروفایل است. |
| `POST /v1/audio/speech` | `POST /v1/audio/speech` | پیاده‌سازی شده؛ `voice_id` یا `ref_audio` و enum رسمی قالب صوتی اعمال می‌شود و پاسخ native JSON/base64 یا SSE بدون تبدیل قرارداد عبور می‌کند. |
| Files، Batch، Agents و Conversations | مسیرهای native رسمی | handlerهای عمومی، allowlist مسیر، ID opaque، مالکیت tenant/account، ارجاع متقابل منابع و reconciliation/تسویهٔ Batch پیاده‌سازی شده‌اند. هر سطح تا عبور contract/smoke زنده با feature flag مستقل پیش‌فرض خاموش است. ابزارهای داخلی پولی در Agents/Conversations تا وجود usage قابل اعتماد fail-closed می‌مانند. |

## اعتبارنامه و اتصال

- گدارAI نام `mistral-ai` را canonical نگه می‌دارد و `mistral` را فقط در مرز سازگاری می‌پذیرد.
- کلید دسترسی با `Authorization: Bearer ...` ارسال می‌شود.
- نشانی پیش‌فرض `https://api.mistral.ai/v1` است.
- آزمون اعتبارنامه از `GET /v1/models` استفاده می‌کند.
- اعتبارسنجی BYOK و Platform-managed هر دو نام ارائه‌دهنده را به یک قرارداد متصل می‌کنند.
- میزبان نشانی پایه فقط از allowlist ارائه‌دهنده پذیرفته می‌شود.

سطوح OpenAI-compatible میسترال دیگر پرچم env ندارند و همیشه فعال‌اند؛ دسترسی از طریق deploymentها و مدل‌های تننت کنترل می‌شود.

## محاسبه هزینه

قیمت از snapshot کاتالوگ `mistral-ai` خوانده می‌شود. snapshot گم‌شده، قدیمی، دارای مدل ناشناخته یا فاقد بُعد هزینهٔ الزامی با خطای `pricing_unavailable` رد می‌شود؛ fallback هزینهٔ صفر وجود ندارد.

ابعاد پیاده‌سازی‌شده:

- `input_tokens` و `output_tokens`؛
- `cache_read_tokens`؛
- نرخ شرطی ۵۰٪ Batch با شرط `api=batches`؛
- `audio_input_seconds`؛
- `speech_input_characters`؛
- `ocr_input_pages` و `ocr_annotated_pages`؛
- نرخ‌های مستقل ابزارهای داخلی مانند `code_execution_calls`، `image_generation_outputs`، Web Search و Library در کاتالوگ ثبت می‌شوند، ولی تا تکمیل endpoint و استخراج مصرف فعال نمی‌شوند.

تسویهٔ نهایی Chat، FIM، Embeddings، Moderation، OCR، Transcription و TTS از میزان مصرف پاسخ یا واحد واقعی همان modality استفاده می‌کند. مسیرهای صوتی token billing را اجرا نمی‌کنند: Transcription با ثانیه و TTS با تعداد نویسه تسویه می‌شوند. Batch با reservation پایدار، polling، claim اتمیک و خواندن JSONL خروجی تسویه می‌شود؛ OCR و Transcription به‌ترتیب با صفحه و ثانیه و سایر endpointهای مجاز با usage توکنی نهایی می‌شوند. نگاشت کامل مصرف ابزارهای داخلی پولی در Agents همچنان شرط فعال‌سازی آن ابزارهاست و درخواست آن‌ها fail-closed است.

## بررسی نتیجه

این فرمان‌ها در worktree اجرا شده‌اند:

```text
pnpm --filter model-discovery test -- --runInBand mistral-ai-provider-importer.spec.ts
pnpm --filter model-discovery run import:providers:yaml -- --providers mistral-ai --dry-run
pnpm --filter model-discovery run build
go test ./platform
go test ./pkg/server
go test ./internal/proxy ./internal/billing
go test ./cmd/server ./internal/admin ./internal/storage
```

آزمون زنده با کلید staging اجرا نشده است. تا پیش از تکمیل smoke matrix برای هر ۱۰ مدل، rollout باید canary و پشت feature flag بماند.

## منابع رسمی

- [فهرست مدل‌های Mistral](https://docs.mistral.ai/models/overview)
- [OpenAPI رسمی Mistral](https://docs.mistral.ai/openapi.yaml)
- [قیمت API](https://mistral.ai/pricing/api/)
- [Chat API](https://docs.mistral.ai/api/endpoint/chat)
- [Embeddings API](https://docs.mistral.ai/api/endpoint/embeddings)
- [راهنمای رسمی Mistral Embed و خروجی ثابت ۱۰۲۴بعدی](https://docs.mistral.ai/resources/cookbooks/mistral-embeddings-embeddings)
- [معرفی رسمی Codestral Embed و ابعاد/precision قابل تنظیم](https://mistral.ai/news/codestral-embed/)
- [FIM API](https://docs.mistral.ai/api/endpoint/fim)
- [OCR API](https://docs.mistral.ai/api/endpoint/ocr)
- [Audio Transcriptions API](https://docs.mistral.ai/api/endpoint/audio/transcriptions)
- [Audio Speech API](https://docs.mistral.ai/api/endpoint/audio/speech)
- [محدودیت‌های جاری](https://docs.mistral.ai/resources/known-limitations)
- [ابزارهای Agents و محدودیت endpoint](https://docs.mistral.ai/studio-api/agents/agent-tools)
- [Batch API](https://docs.mistral.ai/api/endpoint/batch)
- [Files API](https://docs.mistral.ai/api/endpoint/files)
- [Agents API](https://docs.mistral.ai/api/endpoint/beta/agents)
- [Conversations API](https://docs.mistral.ai/api/endpoint/beta/conversations)
- [مدیریت کلید دسترسی](https://docs.mistral.ai/admin/identity-access/api-keys)

## بدهی باقی‌مانده پیش از اعلام پشتیبانی کامل

1. مصرف هر ابزار داخلی Agents به بُعد هزینهٔ رسمی همان ابزار نگاشت و با پاسخ usage واقعی آزمون شود.
2. migration کاتالوگ روی staging اجرا و رکوردهای قدیمی فقط پس از dry-run غیرفعال شوند.
3. ماتریس زندهٔ مدل × نقطه پایانی × نوع اعتبارنامه × انتقال با کلید staging سبز شود.
