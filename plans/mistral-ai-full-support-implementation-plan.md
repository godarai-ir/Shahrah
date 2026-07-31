# پلن پشتیبانی کامل از مدل‌های Mistral AI در گدارAI

> وضعیت: پیشنهادی — آماده‌ی تبدیل به Epic و تسک اجرایی  
> دامنه: تمام مدل‌های موجود در `apps/model-discovery/providers/mistral-ai/`  
> تاریخ بررسی منابع: ۲۰۲۶-۰۷-۳۱  
> پیش‌نیاز تحلیلی: `plans/mistral-provider-readiness-checklist.md`

## ۱. نتیجه‌ی مورد انتظار

پس از اجرای این پلن، هر ۱۰ مدل Mistral AI موجود در پوشه‌ی کاتالوگ باید فقط زمانی در گدارAI «پشتیبانی‌شده» و قابل انتخاب باشد که این زنجیره برای آن کامل و آزموده شده باشد:

`YAML کاتالوگ → مدل و alias رسمی → endpoint مجاز → قرارداد request/response → credential → فراخوانی upstream → استخراج usage → محاسبه و تسویه‌ی هزینه → ثبت مشاهده‌پذیری و آزمون زنده`

هیچ مدل، endpoint یا نرخ هزینه‌ای نباید صرفاً با اتکا به سازگاری ظاهری OpenAI یا مقدار ثابت در کد فعال شود. مدل‌های خارج از این پوشه در دامنه‌ی این پلن نیستند و تا زمان اضافه‌شدن به کاتالوگ تأییدشده نباید به‌عنوان پشتیبانی‌شده منتشر شوند.

## ۲. منبع حقیقت و قاعده‌ی حل تعارض

اطلاعات اجرایی و دیتابیس باید از منابع رسمی زیر استخراج و با تاریخ و نشانی منبع ذخیره شوند:

1. [Mistral OpenAPI](https://docs.mistral.ai/openapi.yaml): مسیرها، روش HTTP، content type، فیلدهای request/response، enumها و قرارداد خطا.
2. [Models Overview](https://docs.mistral.ai/models/overview): وضعیت عرضه، شناسه‌ها و aliasهای مدل.
3. model card رسمی هر مدل: context، modality، capability و endpointهای اعلام‌شده.
4. [API Pricing](https://mistral.ai/pricing/api/): تنها مرجع قیمت عمومی و تخفیف‌های batch/cache.
5. [API Keys](https://docs.mistral.ai/admin/identity-access/api-keys): روش احراز هویت و مدیریت کلید.
6. [Changelog](https://docs.mistral.ai/resources/changelogs): تغییر قرارداد، مدل، قیمت یا lifecycle.

ترتیب حل تعارض:

- قرارداد API: OpenAPI رسمی مقدم است.
- قیمت: صفحه‌ی API Pricing مقدم است.
- capability و lifecycle: model card و Models Overview مقدم‌اند.
- اگر دو منبع رسمی همچنان ناسازگار باشند، انتشار آن مدل یا بُعد هزینه متوقف می‌شود؛ تعارض با `source_url`، زمان مشاهده و داده‌ی خام ثبت و پس از تأیید دستی رفع می‌شود.
- نمونه‌ی واقعی فعلی: model card مدل `mistral-moderation-2603` قیمت صفر نشان می‌دهد، اما API Pricing ورودی را `0.1 USD / 1M tokens` اعلام می‌کند. مقدار قابل استفاده برای billing همان API Pricing است و هشدار اختلاف منبع نیز باید ثبت شود.

## ۳. تعریف «پشتیبانی کامل» و دروازه‌ی انتشار

هر ردیف مدل باید تمام وضعیت‌های زیر را `ready` داشته باشد:

| محور | شرط پذیرش |
|---|---|
| هویت | شناسه‌ی canonical، alias رسمی، lifecycle و context درست |
| modality | ورودی و خروجی دقیق و سازگار با schema و UI/API |
| capability | فقط قابلیت‌های اعلام‌شده در model card رسمی |
| endpoint | تمام endpointهای رسمی مرتبط با همان مدل؛ sync، stream و batch هرجا اعلام شده |
| contract | request/response و خطا مطابق OpenAPI همان endpoint |
| credential | Platform-managed و BYOK با probe موفق و بدون افشای secret |
| routing | مسیر upstream، content type، SSE/multipart و timeout صحیح |
| billing | نرخ معتبر، واحد مصرف، رزرو، تسویه و refund/reconciliation آزموده‌شده |
| observability | usage، هزینه، endpoint، model و credential قابل ردیابی بدون ثبت محتوای حساس |
| verification | unit، contract، integration، tunnel و smoke زنده موفق |

پرچم عمومی `supported` نباید مستقل از این readiness matrix باشد. نبود runtime profile یا نرخ لازم باید به خطای روشن و fail-closed منجر شود، نه هزینه‌ی صفر یا ارسال درخواست ناقص.

## ۴. ماتریس هدف مدل‌ها

| فایل فعلی | شناسه‌ی canonical هدف | modality | endpointها و حالت‌های الزامی | نکته‌ی اصلاحی اصلی |
|---|---|---|---|---|
| `codestral-2508.yaml` | `codestral-2508` | text/code → text/code | Chat Completions، FIM، Conversations، Batch | context از `256k` به `128k` اصلاح شود؛ FIM روی `/v1/fim/completions` است. |
| `codestral-embed-2505.yaml` | `codestral-embed-2505` | text/code → embedding | Embeddings، Batch | خروجی‌های embedding و پارامترهای dimension/dtype/encoding کامل شوند. |
| `mistral-embed-2312.yaml` | `mistral-embed-2312` | text → embedding | Embeddings، Batch | deprecated فعلی حذف شود؛ مدل در منبع رسمی GA و فعال است. |
| `mistral-large-2512.yaml` | `mistral-large-2512` | text/image/document → text | Chat، Conversations، Agents، Built-in Tools، Batch | Document QnA، agents و هزینه‌ی tool call به پروفایل اضافه شود. |
| `mistral-medium-3-5.yaml` | `mistral-medium-3-5` | text/image/document → text | Chat، Conversations، Agents، Built-in Tools، Batch | reasoning، multimodal و endpointهای agent به‌صورت مدل‌محور محدود شوند. |
| `mistral-moderation-2603.yaml` | `mistral-moderation-2603` | text/conversation → moderation result | Moderations، Chat Moderations، Batch | خروجی `text` نیست؛ قیمت billing از API Pricing خوانده شود. |
| `mistral-ocr-4.yaml` | `mistral-ocr-4-0` | image/PDF/document → OCR/structured document | OCR، BBox، Structured Annotations، Batch | شناسه‌ی فعلی `mistral-ocr-4` رسمی نیست؛ `temperature` حذف شود. |
| `mistral-small-2603.yaml` | `mistral-small-2603` | text/image/document → text | Chat، Conversations، Agents، Built-in Tools، Batch | document، reasoning و هزینه‌ی ابزارها پوشش داده شود. |
| `voxtral-mini-2602.yaml` | `voxtral-mini-2602` | audio → transcription | Audio Transcriptions، streaming، Batch در صورت تأیید OpenAPI/model availability | هزینه از ثانیه به واحد canonical دقیق نگاشت شود؛ `structured_output` بدون مدرک حذف شود. |
| `voxtral-mini-tts-2603.yaml` | `voxtral-mini-tts-2603` | text + voice reference → audio | Audio Speech، streaming، voice cloning | `function_calling` و `temperature` حذف؛ هزینه‌ی character و فرمت‌های رسمی اعمال شود. |

نکته‌ها:

- aliasها فقط از `GET /v1/models`، Models Overview یا model card رسمی وارد شوند. aliasها برای routing قابل قبول‌اند، اما در usage و صورتحساب باید به شناسه‌ی canonical و catalog revision resolve شوند.
- برای مهاجرت OCR، اگر شناسه‌ی اشتباه `mistral-ocr-4` قبلاً عمومی شده باشد، فقط به‌عنوان alias منسوخ داخلی با هشدار migration نگه داشته شود؛ upstream همواره `mistral-ocr-4-0` دریافت کند.
- endpointهای batch برای هر مدل باید از فهرست endpointهای مجاز OpenAPI و smoke واقعی تأیید شوند؛ وجود عمومی `/v1/batch` به‌تنهایی مجوز batch برای هر مدل نیست.

## ۵. شکاف‌های قطعی فعلی

### ۵.۱. کاتالوگ و دیتابیس

- importer عمومی YAML مقدارهای `input_cost_per_second`، `input_cost_per_page` و `input_cost_per_annotated_page` را به rateهای billing تبدیل نمی‌کند.
- schema حالت مدل `ocr` و endpointهای Mistral مانند FIM، OCR و Chat Moderations را کامل نمی‌شناسد.
- modality سند در همه‌ی لایه‌های runtime profile یکدست نیست.
- هیچ importer و validator اختصاصی برای `mistral-ai` وجود ندارد.
- runtime profile صریح در فایل‌های این provider وجود ندارد؛ بنابراین endpoint و پارامترهای قابل قبول از کاتالوگ قابل اثبات نیست.
- اجرای dry-run فعلی ۱۰ مدل را parse می‌کند، اما صحت معنایی داده‌ها را در برابر منابع رسمی تضمین نمی‌کند.

### ۵.۲. API و routing

- مسیر فعلی completions، FIM را به `/completions` می‌فرستد؛ مسیر رسمی `/v1/fim/completions` است.
- Embeddings در بخشی از کد فقط alias قدیمی `mistral` را می‌پذیرد و canonical provider یعنی `mistral-ai` را رد می‌کند.
- Moderation، OCR و TTS برای Mistral مسیر end-to-end کامل ندارند.
- transcription فعلی فقط upload فایل فیزیکی را پوشش می‌دهد؛ قرارداد رسمی `file`، `file_url` و `file_id` را دارد.
- adapter عمومی پارامترهای native را ممکن است حذف کند و تضمینی برای multipart، SSE، binary audio و قرارداد خطای Mistral ندارد.
- endpointهای Agents، Conversations، Files و Batch برای این provider کامل و model-aware نیستند.

### ۵.۳. credential

- schema پایه‌ی `api_key` و Bearer مناسب است، اما aliasهای `mistral` و `mistral-ai` در BYOK یکدست نیستند.
- probe باید هم برای Platform-managed و هم BYOK روی endpoint رسمی models و با base URL مجاز اجرا شود.
- انتساب هزینه به credential/account و rotation بدون downtime باید به‌صورت صریح آزموده شود.

### ۵.۴. billing و pricing

- registry ثابت فعلی فقط چند مدل قدیمی Mistral را می‌شناسد و canonical provider می‌تواند به قیمت صفر سقوط کند.
- billing برای token، cached token، batch token، audio duration، character، OCR page و annotated page به ابعاد مستقل نیاز دارد.
- فراخوانی native pass-through بدون استخراج usage می‌تواند billing را دور بزند. مستندات LiteLLM نیز برای passthrough خام Mistral، cost tracking کامل تضمین نمی‌کند؛ گدارAI نباید این محدودیت را تکرار کند.
- عملیات async مانند Batch نیازمند reservation و settlement چندمرحله‌ای و idempotent است.

## ۶. معماری هدف

### ۶.۱. لایه‌ی کاتالوگ

برای هر مدل یک runtime profile نسخه‌دار ذخیره شود که حداقل این داده‌ها را دارد:

- provider canonical برابر `mistral-ai` و alias ورودی `mistral`؛
- canonical model ID، aliasها، lifecycle و source URLs؛
- input/output modalities و محدودیت context؛
- endpointهای مجاز، روش HTTP، content type و پشتیبانی stream/batch؛
- capabilityهای مدل و پارامترهای مجاز به تفکیک endpoint؛
- pricing rates با unit، currency، source، effective time و revision؛
- policyهای max body/file/page/audio duration، timeout و retry؛
- وضعیت readiness هر محور.

یک importer اختصاصی در `apps/model-discovery/scripts/model-import/providers/mistral-ai.ts` اضافه شود. این importer باید علاوه بر parse عمومی، قواعد مختص Mistral را validate کند و در صورت شناسه، endpoint، modality یا pricing ناقص build را fail کند.

### ۶.۲. سطح API گدارAI

دو سطح قرارداد ارائه شود:

1. **OpenAI-compatible** برای endpointهای واقعاً هم‌پوشان: `/v1/chat/completions`، `/v1/embeddings`، `/v1/moderations`، `/v1/audio/transcriptions` و `/v1/audio/speech`.
2. **Mistral-native** برای حفظ قرارداد رسمی: `/v1/fim/completions`، `/v1/chat/moderations`، `/v1/ocr`، Batch، Files، Agents و Conversations.

مسیر native نباید یک reverse proxy آزاد باشد. allowlist مسیر، مدل، content type، پارامتر و اندازه‌ی payload باید از runtime profile خوانده شود. تمام مسیرهای native نیز باید از auth، policy، reservation، usage capture و settlement گدارAI عبور کنند.

### ۶.۳. قراردادهای typed

یک package قرارداد Mistral ایجاد شود و schemaهای OpenAPI رسمی به fixture و typeهای قابل آزمون تبدیل شوند. حداقل خانواده‌های زیر باید قرارداد مستقل داشته باشند:

- Chat/Conversations/Agents؛
- FIM؛
- Embeddings؛
- Moderations و Chat Moderations؛
- OCR؛
- Audio Transcriptions؛
- Audio Speech؛
- Batch و Files؛
- SSE events، binary responses، usage و errors.

پارامترهای native نباید در generic map رها یا بی‌صدا حذف شوند. در مسیر compatible، نگاشت‌های صریح زیر لازم است:

- `seed` گدارAI/OpenAI → `random_seed` Mistral؛ ارسال همزمان هر دو، خطای validation بدهد.
- `dimensions` → `output_dimension` برای embeddings؛ تعارض رد شود.
- metadata داخلی گدارAI از `metadata` قابل ارسال به Mistral جدا شود.
- `file`، `file_url` و `file_id` در transcription دقیقاً طبق قرارداد رسمی و به‌صورت mutually exclusive اعتبارسنجی شوند.
- `voice_id` و `ref_audio` در TTS حفظ شوند؛ مقدار OpenAI-style `voice` نباید با حدس نگاشت شود.

### ۶.۴. adapter و tunnel

در Silk Road یک adapter یا profile اختصاصی Mistral ایجاد شود که این موارد را تضمین کند:

- base URL رسمی و regional endpointهای allowlist‌شده؛
- `Authorization: Bearer <api_key>` بدون forwardشدن headerهای حساس کاربر؛
- path صحیح هر endpoint؛
- JSON، multipart، SSE و binary audio؛
- cancellation و backpressure برای stream؛
- timeout و retry فقط برای عملیات idempotent؛
- استخراج model و usage پیش از پایان stream یا job؛
- نگاشت خطاهای network/upstream بدون نشت credential.

### ۶.۵. credential

اتصال provider باید این قرارداد را داشته باشد:

- `api_key`: اجباری، write-only، رمزنگاری‌شده و redacted؛
- `base_url`: اختیاری، با مقدار پیش‌فرض رسمی و allowlist برای endpointهای منطقه‌ای؛
- probe: درخواست حداقلی به `GET /v1/models` و ثبت timestamp/status، بدون ذخیره‌ی پاسخ حساس؛
- scope: tenant/account و در صورت نیاز project، با انتخاب قطعی و قابل audit؛
- دو حالت Platform-managed و BYOK با pipeline یکسان پس از resolve credential؛
- rotation: نسخه‌ی credential، تست نسخه‌ی جدید، cutover اتمیک و امکان rollback؛
- canonicalization: ورودی `mistral` در مرز API به `mistral-ai` تبدیل شود و از آن پس فقط canonical استفاده شود.

### ۶.۶. نقشه‌ی تغییرات در مخزن

| ناحیه | نقطه‌ی شروع | مسئولیت تغییر |
|---|---|---|
| YAML و runtime schema | `apps/model-discovery/providers/mistral-ai/` و `apps/model-discovery/src/catalog/runtime-profile.schema.ts` | اصلاح داده‌ی ۱۰ مدل و افزودن endpoint/modality/rate dimensionهای لازم |
| import | `apps/model-discovery/scripts/import-models-from-yaml.ts` و importer جدید `scripts/model-import/providers/mistral-ai.ts` | validation اختصاصی، alias/lifecycle، runtime profile و pricing rates |
| route registration | `services/control-plane/internal/server/routes_proxy.go` | ثبت مسیرهای native و compatible با feature flag مستقل |
| Chat و multimodal validation | `services/control-plane/internal/proxy/handler.go` | حفظ metadata، validation مدل‌محور و usage capture |
| FIM | `services/control-plane/internal/proxy/completions_handler.go` | جداسازی قرارداد legacy completions از `/v1/fim/completions` |
| Embeddings | `services/control-plane/internal/proxy/embeddings_handler.go` | پذیرش canonical provider و پارامترهای native embedding |
| Moderation | `services/control-plane/internal/proxy/moderations_handler.go` | Mistral raw/chat moderation و قرارداد پاسخ صحیح |
| Audio | `services/control-plane/internal/proxy/audio_transcriptions_handler.go` و `audio_speech_handler.go` | file reference، streaming، binary، voice cloning و usage |
| قراردادها | `packages/contracts/openai/` به‌همراه package اختصاصی Mistral | typeهای native و mappingهای compatible بدون حذف فیلد |
| tunnel | `services/silk-road/pkg/server/adapters.go` و `openai_compatible.go` یا adapter اختصاصی | path/content-type/SSE/multipart/binary و header policy |
| credential | `services/control-plane/internal/platformcredentials/` و `internal/storage/provider_credential_validation.go` | canonicalization، probe، BYOK و Platform-managed |
| billing | `services/control-plane/internal/billing/pricing.go`، `service.go` و `catalog_pricing_snapshot.go` | حذف fallback صفر، snapshot کاتالوگ Mistral و settlement چندبعدی |

نام و محل نهایی فایل‌های جدید می‌تواند با conventions هر package تنظیم شود، اما مرز مسئولیت‌های جدول نباید با یک handler یا proxy عمومی مبهم جایگزین شود.

## ۷. قرارداد پارامترها و پاسخ‌ها

فهرست نهایی enum و required/optional باید در زمان پیاده‌سازی مستقیماً از OpenAPI pinشده تولید شود. محدوده‌ی لازم برای هر خانواده:

### Chat

پوشش کامل message contentهای text/image/document، tools، tool choice، parallel tool calls، structured response، predicted output، prefix/prefill، `random_seed`، penalties، `reasoning_effort`، `prompt_mode`، `guardrails`، streaming و usage نهایی. مجازبودن هر پارامتر باید با مدل و endpoint تقاطع داده شود.

### FIM

`model` و `prompt` اجباری؛ `suffix`، `min_tokens`، sampling، stop، stream، metadata و prompt cache key طبق OpenAPI. این قرارداد از legacy `/v1/completions` جدا باشد.

### Embeddings

ورودی تکی و آرایه‌ای؛ `output_dimension`، `output_dtype` و `encoding_format` با enum رسمی؛ پاسخ float یا base64 بدون تغییر مخرب؛ usage توکن برای billing.

### Moderation

raw text و chat conversation، category scores و jailbreak detection مطابق پاسخ رسمی. نوع خروجی در کاتالوگ `moderation_result` باشد، نه `text`.

### OCR

document input، page selection، base64 images، image limits، bbox/structured annotation format، document annotation prompt، table format، header/footer extraction، blocks و confidence granularity. پاسخ page-level و usage تعداد صفحه بدون تبدیل از دست‌دهنده حفظ شود.

### Transcription

multipart و reference input؛ language، temperature در صورت وجود در OpenAPI، streaming، diarization، context bias و timestamp granularities. duration رسمی پاسخ مبنای تسویه است.

### TTS

input، model، voice ID یا reference audio، response format رسمی، stream و prompt cache key در صورت وجود. پاسخ binary/stream بدون JSON wrapping اجباری عبور کند.

### Agents، Conversations، Files و Batch

چرخه‌ی ایجاد/خواندن/فهرست/لغو یا حذف طبق OpenAPI، با tenant ownership، idempotency key و مدل مجاز. Batch input file، endpoint target، model/agent، metadata، timeout، status و result/error files باید end-to-end پشتیبانی شوند.

## ۸. مدل billing و pricing

### ۸.۱. ابعاد نرخ

`ModelPricingRate` باید حداقل این dimensionها را پشتیبانی کند:

| dimension | واحد canonical | کاربرد |
|---|---|---|
| `input_tokens` | token | Chat، FIM، Embeddings، Moderation |
| `output_tokens` | token | Chat و FIM |
| `cache_read_tokens` | token | cached input با نرخ رسمی |
| `batch_input_tokens` | token | batch token input |
| `batch_output_tokens` | token | batch token output |
| `audio_input_seconds` | second | Transcription؛ نرخ رسمی دقیقه‌ای با Decimal دقیق تبدیل شود |
| `speech_input_characters` | Unicode character | TTS |
| `ocr_input_pages` | page | OCR عادی |
| `ocr_annotated_pages` | page | OCR با annotation/Document AI |
| `tool_calls.<tool>` | call/image/page/token | built-in tools و libraries مطابق API Pricing |

تمام قیمت‌ها با Decimal، currency USD، source URL، زمان اثرگذاری و revision ذخیره شوند. float باینری برای محاسبه‌ی پول ممنوع است.

### ۸.۲. نرخ‌های baseline فعلی

baseline زمان تدوین پلن، برای validation اولیه است و باید در import مجدد از API Pricing تأیید شود:

- Codestral: ورودی `$0.3/M`، خروجی `$0.9/M`.
- Codestral Embed: ورودی `$0.15/M`.
- Mistral Embed: ورودی `$0.1/M`.
- Mistral Large 3: ورودی `$0.5/M`، خروجی `$1.5/M`.
- Mistral Medium 3.5: ورودی `$1.5/M`، خروجی `$7.5/M`.
- Mistral Small 4: ورودی `$0.15/M`، خروجی `$0.6/M`.
- Mistral Moderation: ورودی `$0.1/M`.
- OCR 4: `$4/1000 pages` و Document AI/annotated برابر `$5/1000 pages`.
- Voxtral Mini Transcribe 2: `$0.003/minute`.
- Voxtral TTS: `$0.016/1000 characters`.
- Batch عمومی: ۵۰٪ تخفیف؛ cached input: ۹۰٪ تخفیف، فقط برای model/endpointهایی که رسماً مشمول‌اند.

### ۸.۳. رزرو و تسویه

- Chat/FIM/Embedding/Moderation: رزرو با برآورد توکن و سقف خروجی؛ تسویه با usage رسمی.
- TTS: رزرو قطعی بر پایه‌ی شمارش Unicode character ورودی؛ تسویه با همان canonical count و usage upstream در صورت ارائه.
- Transcription: برای upload، duration فایل پیش از ارسال استخراج شود؛ برای URL/file ID سقف رزرو امن اعمال و پس از دریافت `usage.seconds` تسویه شود.
- OCR: رزرو بر پایه‌ی تعداد صفحات انتخاب‌شده یا سقف فایل؛ تسویه با pages processed و نرخ عادی/annotated مطابق گزینه‌های درخواست.
- Batch: هنگام ایجاد job رزرو، هنگام completion تسویه‌ی هر item/result، هنگام cancel یا failure آزادسازی باقی‌مانده؛ همه‌ی transitionها idempotent.
- Agents/Built-in Tools: هزینه‌ی مدل و tool call جدا ثبت و سپس در یک ledger transaction گروه‌بندی شود.

catalog revision، exchange-rate revision و credential ID باید همراه reservation pin شوند تا تغییر قیمت در میانه‌ی request/job محاسبه را عوض نکند. نبود rate، usage یا profile معتبر باید اجرای پولی را متوقف کند.

## ۹. فازهای اجرایی

### فاز ۰ — قفل‌کردن قرارداد رسمی و baseline

خروجی‌ها:

- snapshot نسخه‌دار از OpenAPI، model cards، Models Overview و API Pricing با checksum و observed-at؛
- manifest رسمی ۱۰ مدل و aliasهای تأییدشده؛
- endpoint/model/parameter/pricing matrix ماشینی؛
- ثبت تعارض‌های منبع و rule حل آن‌ها؛
- golden fixtures request/response/error/SSE از schema رسمی.

معیار خروج: هیچ مقدار model ID، price، modality یا endpoint بدون source رسمی باقی نماند.

### فاز ۱ — اصلاح schema، migration و importer

تغییرات:

- افزودن `ocr`، `fim_completion`، `chat_moderation`، `agents`، `conversations` و endpointهای لازم به taxonomy/runtime schema؛
- یکدست‌کردن modality سند؛
- تکمیل `ModelRuntimeProfile` و `ModelPricingRate` و اجرای migrationهای محیط‌ها؛
- ساخت importer/validator اختصاصی Mistral؛
- افزودن rate dimensionهای audio/page/annotated-page/character/tool؛
- اصلاح ۱۰ YAML طبق ماتریس بخش ۴ و افزودن runtime profile/source/effective dates؛
- سیاست پاک‌سازی رکوردهای قدیمی Mistral در DB: dry-run، گزارش diff، سپس disable رکوردهای خارج از manifest؛ حذف فیزیکی در این فاز انجام نشود.

معیار خروج: import در دیتابیس خالی و دیتابیس دارای داده‌ی قدیمی، خروجی یکسان و idempotent بدهد؛ هیچ مدل خارج از manifest به‌صورت enabled باقی نماند.

### فاز ۲ — credential و provider identity

تغییرات:

- canonicalization سراسری `mistral` → `mistral-ai`؛
- schema اتصال، secret storage، redaction، probe و rotation؛
- اتصال Platform-managed و BYOK به یک credential resolver؛
- attribution مصرف و هزینه به credential/account؛
- validation و allowlist برای base URL و regional endpoints.

معیار خروج: probe، rotation، revoke و failover برای هر دو نوع credential در integration test موفق باشد و secret در log/error/trace دیده نشود.

### فاز ۳ — مسیرهای synchronous و streaming اصلی

ترتیب پیاده‌سازی:

1. Chat و FIM؛
2. Embeddings؛
3. Moderations و Chat Moderations؛
4. OCR؛
5. Audio Transcriptions؛
6. Audio Speech/TTS.

برای هر خانواده همزمان contract، handler، adapter، billing extractor و tests پیاده شود؛ endpoint بدون billing کامل پشت feature flag باقی بماند.

معیار خروج: هر ۱۰ مدل حداقل endpoint اصلی خود را با Platform-managed و BYOK، sync/stream مربوط و صورتحساب دقیق در staging اجرا کند.

### فاز ۴ — Batch، Files، Conversations و Agents

تغییرات:

- proxy امن چرخه‌ی Files و Batch؛
- state mapping بین tenant و upstream resource ID؛
- Conversations برای مدل‌های اعلام‌شده؛
- Agents و Built-in Tools فقط برای Large/Medium/Small و مدل‌هایی که model card تأیید می‌کند؛
- reservation/settlement async، polling/webhook داخلی، cancellation و result retrieval؛
- rateهای جداگانه‌ی built-in tools.

معیار خروج: job و resource متعلق به tenant دیگر قابل دسترسی نباشد؛ duplicate callback/poll هزینه را دوباره ثبت نکند؛ Batch برای هر endpoint مجاز smoke شود.

### فاز ۵ — سازگاری، مستندسازی و rollout

تغییرات:

- مستند endpoint/model matrix عمومی گدارAI؛
- نمونه‌ی curl/SDK برای native و compatible؛
- راهنمای migration برای OCR ID، `seed`/`random_seed` و TTS voice؛
- dashboard readiness، billing reconciliation و provider health؛
- feature flag سراسری provider و flag مجزای هر endpoint؛
- canary ابتدا Platform-managed، سپس BYOK؛
- فعال‌سازی مدل به مدل فقط پس از عبور کامل از GA gate.

معیار خروج: تمام ردیف‌های readiness برای ۱۰ مدل سبز، runbook rollback آزموده و اختلاف هزینه‌ی محاسبه‌شده با usage رسمی در tolerance تعریف‌شده باشد.

## ۱۰. طرح آزمون

### آزمون‌های build و داده

- schema validation تمام YAMLها؛
- مقایسه‌ی manifest با ۱۰ فایل مورد انتظار؛
- validator منبع رسمی برای ID، context، modality، endpoint، alias و pricing؛
- migration up/down در دیتابیس آزمایشی و import idempotency؛
- جلوگیری از قیمت صفر ناخواسته و مدل enabled بدون runtime profile.

### آزمون‌های قرارداد

- golden request/response برای تمام endpointها؛
- required/optional، enum، mutually-exclusive fields و unknown field policy؛
- mappingهای compatible و conflict errors؛
- SSE chunk/usage termination، multipart و binary audio؛
- parity دوره‌ای با OpenAPI checksum و هشدار drift در CI.

### آزمون‌های routing و امنیت

- model-endpoint deny matrix؛
- جلوگیری از arbitrary path و SSRF از base URL/file URL؛
- محدودیت size/page/duration و MIME sniffing؛
- header allowlist و credential redaction؛
- tenant ownership برای files/batches/agents/conversations؛
- cancellation، timeout، retry safety و stream disconnect.

### آزمون‌های billing

- Decimal exactness برای token/minute/second/page/character؛
- cache و batch discount فقط در حالت مجاز؛
- reservation، success settlement، upstream error، client cancel و refund؛
- async idempotency و replay؛
- alias → canonical pricing؛
- pinشدن catalog و exchange-rate revision؛
- fail-closed برای rate/usage/profile گمشده.

### آزمون زنده

با کلید staging واقعی Mistral، برای هر مدل حداقل یک درخواست کوچک و قابل پیش‌بینی اجرا شود. smoke testها محتوا و credential را ذخیره نکنند و نتیجه‌ی زیر را ثبت کنند: model resolved، endpoint، HTTP status، response shape، usage، هزینه‌ی upstream، هزینه‌ی گدارAI و catalog revision.

## ۱۱. مشاهده‌پذیری و عملیات

متریک‌های لازم:

- request count/error/latency به تفکیک provider، model، endpoint و credential type؛
- upstream status و retry count؛
- stream first-token/first-audio latency و disconnect؛
- usage به تفکیک dimension؛
- reserved/settled/refunded amount و reconciliation delta؛
- catalog/pricing revision و source freshness؛
- credential probe health و آخرین موفقیت؛
- batch queue time، completion time و duplicate-event count.

هشدارها:

- pricing یا OpenAPI drift؛
- مدل enabled با readiness ناقص؛
- zero-cost invocation برای مدل غیررایگان؛
- usage response گمشده؛
- افزایش reconciliation delta؛
- خطای مداوم یک endpoint یا credential.

runbook باید شامل خاموش‌کردن endpoint مستقل، freezeکردن catalog revision، برگشت به snapshot قبلی، تعویض credential و تسویه‌ی jobهای درحال اجرا باشد.

## ۱۲. rollout و rollback

1. migration دیتابیس و schemaها بدون فعال‌سازی provider؛
2. dry-run importer و بررسی diff؛
3. import کاتالوگ با همه‌ی مدل‌ها در حالت hidden؛
4. probe credential و smoke مستقیم upstream؛
5. canary endpointهای sync با Platform-managed؛
6. canary BYOK؛
7. فعال‌سازی streaming و سپس async resources؛
8. انتشار مدل‌به‌مدل پس از readiness کامل؛
9. حذف نمایش رکوردهای قدیمی خارج از manifest پس از دوره‌ی سازگاری.

rollback با feature flag endpoint، بازگردانی catalog snapshot و توقف ایجاد job جدید انجام می‌شود. jobهای پذیرفته‌شده باید تا terminal state مدیریت و تسویه شوند؛ rollback نباید ledger را بازنویسی کند.

## ۱۳. خارج از دامنه

- اضافه‌کردن خانواده‌های دیگر Mistral که در پوشه‌ی مشخص‌شده نیستند؛
- reverse proxy عمومی و بدون allowlist برای همه‌ی endpointهای Mistral؛
- حدس‌زدن capability، alias یا قیمت از LiteLLM یا منابع غیررسمی؛
- حذف فیزیکی فوری رکوردهای قدیمی دیتابیس؛
- پشتیبانی از self-hosted/open-weight deploymentهای غیر Mistral API؛ این موضوع به provider profile جدا نیاز دارد.

LiteLLM در این پلن صرفاً مرجع الگوی محصول برای credential reuse، spend attribution و route organization است؛ هیچ داده‌ی مدل، قیمت یا قرارداد Mistral از LiteLLM به‌عنوان منبع حقیقت وارد نمی‌شود.

## ۱۴. Definition of Done نهایی

پشتیبانی Mistral AI فقط وقتی کامل است که:

- هر ۱۰ مدل بخش ۴ canonical و enabled باشد و هیچ مدل خارج از manifest به اشتباه تبلیغ نشود؛
- endpoint، modality و capability هر مدل با منابع رسمی یکسان باشد؛
- native request/response بدون حذف داده و compatible mapping با خطای شفاف کار کند؛
- Platform-managed و BYOK هر دو probe و smoke موفق داشته باشند؛
- تمام ابعاد هزینه، از token تا page/character/audio/tool، رزرو و تسویه شوند؛
- هیچ مسیر native یا async نتواند auth، policy یا billing را دور بزند؛
- contract drift، pricing drift و zero-cost fallback در CI/runtime مسدود شوند؛
- ماتریس آزمون مدل × endpoint × credential × transport سبز باشد؛
- مستندات عمومی، migration guide، dashboard و rollback runbook منتشر شده باشند.

## ۱۵. منابع رسمی مدل‌ها

- [Codestral 25.08](https://docs.mistral.ai/models/model-cards/codestral-25-08)
- [Codestral Embed 25.05](https://docs.mistral.ai/models/model-cards/codestral-embed-25-05)
- [Mistral Embed 23.12](https://docs.mistral.ai/models/model-cards/mistral-embed-23-12)
- [Mistral Large 3 / 25.12](https://docs.mistral.ai/models/model-cards/mistral-large-3-25-12)
- [Mistral Medium 3.5 / 26.04](https://docs.mistral.ai/models/model-cards/mistral-medium-3-5-26-04)
- [Mistral Moderation 2 / 26.03](https://docs.mistral.ai/models/model-cards/mistral-moderation-26-03)
- [OCR 4](https://docs.mistral.ai/models/model-cards/ocr-4-0)
- [Mistral Small 4 / 26.03](https://docs.mistral.ai/models/model-cards/mistral-small-4-0-26-03)
- [Voxtral Mini Transcribe 2 / 26.02](https://docs.mistral.ai/models/model-cards/voxtral-mini-transcribe-26-02)
- [Voxtral TTS / 26.03](https://docs.mistral.ai/models/model-cards/voxtral-tts-26-03)
