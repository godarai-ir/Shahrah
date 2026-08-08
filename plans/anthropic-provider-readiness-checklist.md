# گزارش نهایی آمادگی ارائه‌دهنده Anthropic در گدارAI

> تاریخ بازبینی: ۲۹ ژوئیه ۲۰۲۶
>
> دامنه: ارائه‌دهنده مستقیم `anthropic` و ۱۱ مدل آماده ثبت‌شده در Model Discovery
>
> نتیجه: **پیاده‌سازی در وضعیت Release Candidate است، اما پشتیبانی کامل Production هنوز تأیید نشده است.**

## خلاصه مدیریتی

از ۹ مغایرتی که در بازبینی قبل باز مانده بود:

- ۶ مورد در کد و کاتالوگ بسته شد.
- ۲ مورد تا حد ممکن در کد بسته شده، اما برای بسته‌شدن عملیاتی به Credential واقعی API Key/WIF و اجرای Staging نیاز دارد.
- ۱ مورد فقط تا سطح انتقال Header و ابزار دریافت Snapshot تکمیل شده و ادغام Rate Limit Admin API با سیاست Runtime همچنان باز است.

آزمون واقعی Upload فایل 500MB و آزمون قطع SSE با موفقیت اجرا شدند. فرمان مستقل Readiness نیز برای اجرای ماتریس زنده ۱۱ مدل، Files، Batches، خروجی 300k و Rate Limit Snapshot ایجاد شده است. چون Credential واقعی Anthropic در این محیط وجود نداشت، هیچ نتیجه‌ای به‌عنوان آزمون زنده ثبت نشده و عنوان «پشتیبانی کامل» صادر نمی‌شود.

در بررسی مستندات جاری یک شکاف جدید نیز شناسایی شد: مسیر OpenAI-compatible سه Server Tool مورد توافق برنامه، یعنی Web Search، Web Fetch و Code Execution را پشتیبانی می‌کند؛ Server Toolهای جدیدتر Anthropic مانند Tool Search، MCP Connector و Advisor هنوز جزو قرارداد و قیمت‌گذاری تضمین‌شده گدارAI نیستند.

## تعریف وضعیت

| وضعیت | تعریف |
|---|---|
| بسته | پیاده‌سازی و آزمون خودکار متناسب وجود دارد و مغایرت کدی شناخته‌شده‌ای باقی نمانده است. |
| آماده در کد | پیاده‌سازی و ابزار آزمون وجود دارد، اما آزمون زنده یا اقدام عملیاتی لازم انجام نشده است. |
| ناقص | بخشی از قرارداد رسمی هنوز در Runtime اعمال نمی‌شود. |
| مسدود عملیاتی | ادامه کار به Credential، دسترسی حساب، زیرساخت Staging یا تأیید انسانی نیاز دارد. |
| نامرتبط | Anthropic این قابلیت را برای مدل‌های دامنه عرضه نمی‌کند. |

## منابع رسمی

- [Models Overview](https://platform.claude.com/docs/en/about-claude/models/overview)
- [Model IDs and Versioning](https://platform.claude.com/docs/en/about-claude/models/model-ids-and-versions)
- [Model Deprecations](https://platform.claude.com/docs/en/about-claude/model-deprecations)
- [Pricing](https://platform.claude.com/docs/en/about-claude/pricing)
- [Messages API](https://platform.claude.com/docs/en/api/messages)
- [Models API](https://platform.claude.com/docs/en/api/models)
- [Message Batches](https://platform.claude.com/docs/en/build-with-claude/batch-processing)
- [Files API](https://platform.claude.com/docs/en/build-with-claude/files)
- [Token Counting](https://platform.claude.com/docs/en/build-with-claude/token-counting)
- [Streaming](https://platform.claude.com/docs/en/build-with-claude/streaming)
- [Extended Thinking](https://platform.claude.com/docs/en/build-with-claude/extended-thinking)
- [Effort](https://platform.claude.com/docs/en/build-with-claude/effort)
- [Tool Reference](https://platform.claude.com/docs/en/agents-and-tools/tool-use/tool-reference)
- [Web Search](https://platform.claude.com/docs/en/agents-and-tools/tool-use/web-search-tool)
- [Code Execution](https://platform.claude.com/docs/en/agents-and-tools/tool-use/code-execution-tool)
- [Prompt Caching](https://platform.claude.com/docs/en/build-with-claude/prompt-caching)
- [Rate Limits](https://platform.claude.com/docs/en/api/rate-limits)
- [Rate Limit API](https://platform.claude.com/docs/en/manage-claude/rate-limits-api)
- [Errors](https://platform.claude.com/docs/en/api/errors)
- [Workload Identity Federation](https://platform.claude.com/docs/en/manage-claude/workload-identity-federation)
- [WIF API Reference](https://platform.claude.com/docs/en/manage-claude/wif-reference)
- [Fast Mode](https://platform.claude.com/docs/en/build-with-claude/fast-mode)

## موجودی ۱۱ مدل

| مدل | Context | سقف خروجی عادی | خروجی Batch با Beta | وضعیت |
|---|---:|---:|---:|---|
| `anthropic/claude-haiku-4-5` | 200k | 64k | — | آماده در کد |
| `anthropic/claude-haiku-4-5-20251001` | 200k | 64k | — | آماده در کد |
| `anthropic/claude-opus-4-1` | 200k | 32k | — | Deprecated |
| `anthropic/claude-opus-4-1-20250805` | 200k | 32k | — | Deprecated؛ Retirement در ۵ اوت ۲۰۲۶ |
| `anthropic/claude-opus-4-5` | 200k | 64k | — | آماده در کد |
| `anthropic/claude-opus-4-5-20251101` | 200k | 64k | — | آماده در کد |
| `anthropic/claude-opus-4-6` | 1M | 128k | 300k | آماده در کد |
| `anthropic/claude-opus-4-7` | 1M | 128k | 300k | آماده در کد |
| `anthropic/claude-sonnet-4-5` | 200k | 64k | — | آماده در کد |
| `anthropic/claude-sonnet-4-5-20250929` | 200k | 64k | — | آماده در کد |
| `anthropic/claude-sonnet-4-6` | 1M | 128k | 300k | آماده در کد |

نکات:

- سقف عادی Sonnet 4.6 مطابق مستندات جاری Anthropic برابر 128k ثبت شده است.
- Beta `output-300k-2026-03-24` فقط برای `message_batches` و فقط سه مدل Opus 4.6، Opus 4.7 و Sonnet 4.6 فعال می‌شود.
- Opus 4.1 پس از تاریخ Retirement باید با `model_retired` غیرقابل استفاده شود. این رفتار صحیح چرخه عمر است، نه نقص پشتیبانی.
- مدل‌های Claude 5 و مدل‌های Anthropic روی Bedrock، Vertex یا Foundry خارج از این دامنه‌اند.

## نتیجه ۹ مغایرت باقی‌مانده

| شناسه | نتیجه نهایی | شواهد و محدودیت باقی‌مانده |
|---|---|---|
| `ANT-REM-002` | مسدود عملیاتی | فرمان ماتریس زنده ۱۱ مدل پیاده‌سازی و با Mock آزموده شد؛ اجرای واقعی API Key و WIF به Credential و Staging نیاز دارد. |
| `ANT-REM-006` | بسته | Beta خروجی 300k در Runtime Profile، Zod، Connection Schema، Deployment، Allowlist و Validator اضافه شد؛ فقط در Batch و مدل‌های مجاز تزریق می‌شود. |
| `ANT-REM-007` | بسته در دامنه برنامه | OpenAI-compatible اکنون Web Search، Web Fetch و Code Execution نوع‌دار، Result Block، Delta و Usage آن‌ها را حفظ می‌کند. |
| `ANT-REM-008` | بسته | `speed` درخواستی از سرعت مؤثر جدا شد؛ Opus 4.6 مقدار `fast → standard` دارد و Opus 4.7/Sonnet 4.6 فقط `standard` می‌پذیرند. |
| `ANT-REM-009` | بسته در کاتالوگ | سقف Sonnet 4.6 برابر 128k ثبت و تست شد؛ مشاهده حساب واقعی Models API زیر `ANT-REM-002` باقی می‌ماند. |
| `ANT-REM-010` | ناقص | همه Headerهای رسمی Rate Limit و `Retry-After` منتقل می‌شوند و Readiness Command می‌تواند Snapshot سازمان/Workspace را با Admin Key دریافت کند؛ Snapshot هنوز سیاست Runtime را تغذیه نمی‌کند. |
| `ANT-REM-011` | بسته | Retry محدود با حداکثر سه Attempt، Backoff نمایی، رعایت `Retry-After` و Allowlist عملیات Idempotent پیاده‌سازی شد. |
| `ANT-REM-012` | آماده در کد | Expiry، Scopeها و آخرین Refresh به‌شکل Sanitized از Silk Road تا Probe، Repository و پنل Tenant منتقل می‌شوند؛ آزمون WIF واقعی زیر `ANT-REM-002` باقی است. |
| `ANT-REM-013` | آماده در کد / مسدود عملیاتی | تست واقعی 500MB و قطع SSE موفق است و فرمان چرخه Files/Batches آماده است؛ اجرای زنده Batch، Security، Tenant isolation و تأیید انسانی به Staging نیاز دارد. |

بنابراین از ۹ مورد قبلی، تنها `ANT-REM-010` یک مغایرت کدی مستقل و باز است. `ANT-REM-002` و بخش عملیاتی `ANT-REM-013` بدون دسترسی بیرونی قابل بستن نیستند.

## وضعیت قابلیت‌ها

| حوزه | وضعیت | توضیح |
|---|---|---|
| کاتالوگ و Lifecycle | بسته | Lifecycle، Replacement، منابع، Runtime Profile و قیمت‌های نسخه‌دار ثبت می‌شوند. |
| API Key پلتفرم و Tenant | بسته در کد | Secret فقط Write-only و رمزگذاری‌شده است؛ Plaintext و Dual-read حذف شده‌اند. |
| WIF پلتفرم و Tenant | آماده در کد | OIDC allowlist، Exchange، Cache، Singleflight، Refresh پیش از Expiry و Scope Probe وجود دارد. |
| قیمت Token و Cache | بسته | Snapshot اتمیک، Decimal، Pin Revision، LKG حداکثر ۲۴ ساعت و Fail Closed وجود دارد. |
| Web Search Billing | بسته | `web_search_requests` محاسبه و در Breakdown ممیزی می‌شود. |
| Code Execution Billing | بسته ایمن | حالت رایگان همراه Web Tool واجد شرایط ممیزی می‌شود؛ حالت مستقلِ غیرقابل انتساب پیش از Upstream رد می‌شود. |
| Custom Tools | بسته | Function، Tool Choice، Tool Result و Parallel Tool Calls نگاشت می‌شوند. |
| Server Tools توافق‌شده | بسته | Web Search، Web Fetch و Code Execution در Native و OpenAI-compatible پشتیبانی می‌شوند. |
| Structured Output | بسته | `response_format.json_schema` با حفظ معنا به `output_config.format` نگاشت می‌شود. |
| Thinking و Effort | بسته | قواعد نسل‌های 4.1 تا 4.7 از Runtime Profile اعمال می‌شوند. |
| Text، Image و PDF | بسته | Base64، URL و `file_id` پشتیبانی می‌شوند. |
| Audio و Video | نامرتبط | برای این ۱۱ مدل Modality رسمی نیستند. |
| Streaming | بسته در کد | Eventها، Thinking، Tool Delta، Usage، Request ID و خطای میانه Stream حفظ می‌شوند. |
| API بومی | بسته در کد | Messages، Count Tokens، Models، Files و Message Batches عرضه شده‌اند. |
| Provider-egress payload بزرگ | بسته در کد | Upload/Download بدون `io.ReadAll` و با Backpressure عبور می‌کند. |
| Retry | بسته | فقط عملیات اثبات‌شده Idempotent Retry می‌شوند. |
| Rate Limit response | بسته | Headerهای درخواست، Token، Input/Output، Priority، Fast و `Retry-After` منتقل می‌شوند. |
| Rate Limit Admin Snapshot در Runtime | ناقص | دریافت در ابزار Readiness وجود دارد، اما منبع سیاست Runtime نشده است. |

## Endpoint و Mode

| قابلیت | مسیر | وضعیت |
|---|---|---|
| Messages Sync/Stream | `POST /v1/messages` | آماده در کد |
| Count Tokens | `POST /v1/messages/count_tokens` | آماده در کد |
| Models List/Get | `GET /v1/models` و `GET /v1/models/{id}` | آماده در کد |
| Files Create/List/Get/Content/Delete | خانواده `/v1/files` | آماده در کد |
| Batches Create/List/Get/Results/Cancel | خانواده `/v1/messages/batches` | آماده در کد |
| OpenAI Chat Completions | `POST /v1/chat/completions` | آماده برای Text، Custom Tool و سه Server Tool توافق‌شده |
| Fast | `speed` | رفتار Requested/Effective مدل‌محور است |
| Priority | `service_tier` | عبور، پاسخ و Usage حفظ می‌شود؛ وابسته به قرارداد حساب |

محدودیت Body:

| عملیات | سقف |
|---|---:|
| Messages و Count Tokens | 32MB |
| Message Batches | 256MB |
| Files | 500MB |

## پارامترها و پاسخ

موارد زیر در Adapter و Runtime Validator پوشش داده می‌شوند:

- `max_tokens`، `temperature`، `top_p`، `top_k` و `stop_sequences`
- `system`، Metadata و `user_id`
- Tools، `tool_choice`، `disable_parallel_tool_use` و `tool_result`
- Prompt Cache و `cache_control`
- `response_format.json_schema` و `output_config.format`
- `service_tier`، `speed` و `inference_geo`
- Text، Image و PDF از Base64، URL و `file_id`
- Text، Thinking، Redacted Thinking، Citation، Tool Use، Server Tool Result و Stop Reason
- Streaming Usage، Tool Delta، Request ID، Rate Limit Header و `Retry-After`

قواعد Thinking:

- Opus 4.7 فقط Adaptive Thinking و `output_config.effort` را می‌پذیرد.
- Opus 4.6 و Sonnet 4.6 به‌صورت پیش‌فرض Adaptive هستند؛ Manual در صورت پشتیبانی Profile با هشدار Deprecated پذیرفته می‌شود.
- مدل‌های 4.5 و پایین‌تر Manual Extended Thinking دارند.
- تعارض Sampling و مجازبودن `effort=max` از Runtime Profile اعتبارسنجی می‌شود.

## جزئیات اصلاحات جدید

### خروجی 300k

- `ModelRuntimeProfile` شامل `betaCapabilities` شد.
- هر Capability شامل Header، Endpointهای مجاز و سقف خروجی است.
- Zod Importer Profile ناسازگار یا Header بدون Capability را رد می‌کند.
- Header فقط از تقاطع درخواست، Allowlist حساب و Profile همه آیتم‌های Batch عبور می‌کند.
- استفاده در Messages عادی یا مدل نامجاز با خطای اعتبارسنجی رد می‌شود.

### Server Toolهای OpenAI-compatible

- ابزار Function همچنان به Custom Tool Anthropic نگاشت می‌شود.
- Web Search، Web Fetch و Code Execution با نسخه‌های Allowlist‌شده نوع‌دار پذیرفته می‌شوند.
- نسخه تاریخی ناشناخته رد می‌شود تا قابلیت یا Billing مبهم به Upstream نرسد.
- `server_tool_use`، Result Blockها، Citation، Delta و Usage بدون حذف حفظ می‌شوند.
- Smoke زنده از `web_search_20250305` استفاده می‌کند تا روی مدل‌های قدیمی دامنه به Programmatic Tool Calling وابسته نباشد.

### Speed

- `speedValues` فقط مقادیر قابل درخواست را نگه می‌دارد.
- `effectiveSpeed` نتیجه واقعی را ثبت می‌کند.
- Opus 4.6 درخواست `fast` را می‌پذیرد، اما Runtime و Metadata آن را `standard` مؤثر ثبت می‌کنند.
- Opus 4.7 و Sonnet 4.6 مقدار `fast` را رد می‌کنند.

### Retry

حداکثر سه Attempt با تأخیر 250 و 500 میلی‌ثانیه انجام می‌شود. مقدار معتبر `Retry-After` بر Backoff داخلی اولویت دارد.

Retry فقط برای این عملیات مجاز است:

- Count Tokens
- Models List/Get
- Files List/Get/Content
- Batches List/Get/Results

Messages، File Create/Delete و Batch Create/Cancel بدون Idempotency قابل اثبات Retry نمی‌شوند.

### وضعیت WIF

پنل Tenant اکنون بدون نمایش Assertion یا Access Token موارد زیر را نشان می‌دهد:

- زمان انقضا
- زمان آخرین Refresh
- Scopeهای مؤثر
- نتیجه قابلیت `workspace:inference`
- نتیجه قابلیت `workspace:developer`
- وضعیت Rotation و Probe

### Release Readiness

فرمان `services/control-plane/cmd/anthropic-readiness` برای هر مدل این موارد را اجرا می‌کند:

- Models Get
- Count Tokens
- Text Sync و Stream
- Structured Output
- Thinking
- قرارداد OpenAI-compatible Server Tool

در حالت `-resources` نیز چرخه Files و Batches اجرا می‌شود. در حالت `-test-300k` سقف Batch بررسی می‌شود. در صورت وجود `ANTHROPIC_ADMIN_KEY` و Workspace ID، Snapshot محدودیت سازمان و Workspace نیز در گزارش Sanitized ثبت می‌شود.

## آزمون‌های اجراشده

| آزمون | نتیجه |
|---|---|
| `go test ./...` در Control Plane | موفق |
| `go vet ./...` در Control Plane | موفق |
| `go test ./...` در Silk Road | موفق |
| `go vet ./...` در Silk Road | موفق |
| `go test ./...` در Contracts | موفق |
| `go vet ./...` در Contracts | موفق |
| Model Discovery unit tests | ۱۳ Suite و ۴۴ Test موفق |
| Model Discovery build | موفق |
| Import خشک Anthropic | ۲۷ مدل موفق |
| Tenant Admin i18n validation | موفق؛ ۲۸ Namespace |
| Tenant Admin ESLint هدفمند | موفق |
| WIF status DOM tests | ۲ Test موفق |
| Upload واقعی 500MB با سنجش حافظه | موفق؛ رشد Heap کمتر از حد 96MB |
| قطع SSE و جلوگیری از `[DONE]` موفق کاذب | موفق |
| Readiness Command با Anthropic Mock | موفق |

Typecheck کامل Tenant Admin به‌علت خطاهای از قبل موجود و نامرتبط در تست‌های Wallet Receipt، Model Card، Rate Limit Form و Auth Import موفق نشد. فایل‌های تغییرکرده WIF در ESLint و تست DOM مستقل موفق بودند.

## شواهد پیاده‌سازی

| قابلیت | مسیر |
|---|---|
| قرارداد Runtime Profile | `packages/contracts/platform/model_runtime_profile.go` |
| قرارداد Tool و Provider Content | `packages/contracts/openai/types.go` |
| اعتبارسنجی Zod | `apps/model-discovery/src/catalog/runtime-profile.schema.ts` |
| Connection Schema Beta | `apps/model-discovery/prisma/migrations/202607290003_anthropic_beta_capabilities/migration.sql` |
| Adapter و Server Tools | `services/silk-road/pkg/server/anthropic.go` |
| Runtime Profile در Silk Road | `services/silk-road/pkg/server/anthropic_runtime_profile.go` |
| Retry | `services/silk-road/pkg/server/anthropic_retry.go` |
| WIF | `services/silk-road/pkg/server/anthropic_wif.go` |
| Rate Limit Header forwarding | `services/control-plane/internal/proxy/errors.go` |
| WIF Probe | `services/control-plane/internal/platformcredentials/tunnel_prober.go` |
| WIF UI | `apps/tenant-admin/app/components/provider-account-wizard/wif-credential-status.tsx` |
| تست 500MB و قطع Stream | `services/silk-road/pkg/server/anthropic_release_readiness_test.go` |
| فرمان ماتریس زنده | `services/control-plane/cmd/anthropic-readiness/main.go` |
| راهنمای اجرای ماتریس | `services/control-plane/cmd/anthropic-readiness/README.md` |
| کاتالوگ ۱۱ مدل | `apps/model-discovery/providers/anthropic` |

## شکاف جدید ناشی از تغییر مستندات

### `ANT-NEW-001` — Server Toolهای جدید خارج از قرارداد تضمین‌شده

مستندات جاری Anthropic علاوه بر سه Server Tool برنامه، موارد زیر را فهرست می‌کند:

- Advisor
- Tool Search
- MCP Connector

API بومی Body رسمی را به‌صورت شفاف عبور می‌دهد، اما مسیر OpenAI-compatible برای این سه ابزار قرارداد نوع‌دار و Billing کامل ندارد. Advisor همچنین ممکن است استنتاج مدل دوم و هزینه جدا ایجاد کند. بنابراین این قابلیت‌ها در ادعای آمادگی فعلی قرار ندارند و تا تعریف Runtime Profile، Beta Allowlist، Usage/Billing و Contract Test نباید از مسیر OpenAI-compatible عرضه شوند.

وضعیت: **P2 و خارج از ۹ مغایرت برنامه قبلی؛ برای ادعای پوشش همه قابلیت‌های جدید Anthropic باید در PR مستقل حل شود.**

## موارد باقی‌مانده

### کار کدی

- [ ] `ANT-REM-010`: قرارداد ذخیره Credential جداگانه Admin API و نگاشت آن به Organization/Workspace طراحی شود.
- [ ] Snapshot Rate Limit سازمان/Workspace نسخه‌دار، TTLدار و Tenant-scoped شود.
- [ ] Snapshot در سیاست Runtime فقط به‌عنوان سقف بالادست اعمال شود؛ Headerهای هر پاسخ منبع اصلاح پویا باقی بمانند.
- [ ] `ANT-NEW-001`: Advisor، Tool Search و MCP Connector برای مدل‌های واقعاً مجاز، قرارداد و Billing مشخص داشته باشند یا صریحاً Fail Closed شوند.

### کار عملیاتی

- [ ] Credential واقعی API Key کم‌دسترسی در Staging فراهم شود.
- [ ] WIF واقعی با Scopeهای حداقلی فراهم شود.
- [ ] ماتریس کامل ۱۱ مدل یک‌بار با API Key و یک‌بار با WIF اجرا شود.
- [ ] Text Sync/Stream، Image URL/Base64/File، PDF، Tools، Structured Output و Thinking بررسی شوند.
- [ ] Files و Batches شامل Success، Partial Failure، Cancel، Expiry و Results اجرا شوند.
- [ ] Batch خروجی 300k برای هر سه مدل مجاز اجرا شود.
- [ ] Scopeهای `workspace:inference` و `workspace:developer`، Refresh هم‌زمان و Expiry واقعی بررسی شوند.
- [ ] Tenant isolation و Account ambiguity در Staging بررسی شوند.
- [ ] Opus 4.1 در روز یا پس از Retirement پاسخ `model_retired` بدهد.
- [ ] گزارش Sanitized توسط بازبین انسانی تأیید شود.

## روش اجرای ماتریس زنده

راهنمای کامل در `services/control-plane/cmd/anthropic-readiness/README.md` قرار دارد. حداقل اجرای API Key:

```powershell
$env:GODARAI_VIRTUAL_KEY = "<virtual-key>"
go run ./cmd/anthropic-readiness `
  -base-url "https://staging.example.com" `
  -auth-label "api_key" `
  -resources `
  -test-300k `
  -report "anthropic-api-key-readiness.json"
```

همین فرمان باید برای Connection مبتنی بر WIF با `-auth-label "wif"` دوباره اجرا شود. برای Snapshot محدودیت‌ها، `ANTHROPIC_ADMIN_KEY` و شناسه Workspace لازم است. Admin Key نباید با API Key استنتاج یکسان فرض شود و در گزارش نیز ذخیره نمی‌شود.

## معیار اعلام پشتیبانی کامل

پشتیبانی Anthropic فقط زمانی «کامل» اعلام می‌شود که:

- `ANT-REM-002` و `ANT-REM-013` با شواهد زنده بسته شوند.
- تصمیم نهایی درباره `ANT-REM-010` اجرا یا محدودیت آن رسماً پذیرفته شود.
- هیچ P0 یا P1 باز نباشد.
- Snapshot قیمت کمتر از ۲۴ ساعت عمر داشته باشد.
- هیچ Credential قدیمی یا Secret قابل خواندن باقی نمانده باشد.
- تمام ۱۱ مدل نتیجه زنده موفق یا نتیجه چرخه عمر مورد انتظار داشته باشند.
- API Key و WIF هر دو در Staging موفق باشند.
- Files، Batches، 300k، Stream، Security و Tenant isolation موفق باشند.
- بازبین انسانی گزارش Sanitized را تأیید کرده باشد.

تا آن زمان، وضعیت دقیق Anthropic در گدارAI:

**«Release Candidate با پوشش کامل دامنه برنامه در کد، اما فاقد تأیید زنده Production و دارای یک شکاف Runtime Rate Limit و یک شکاف جدید در Server Toolهای تازه Anthropic.»**
