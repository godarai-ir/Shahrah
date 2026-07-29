# گزارش بازبینی آمادگی ارائه‌دهنده Anthropic در گدارAI

> تاریخ بازبینی: ۲۹ ژوئیه ۲۰۲۶
>
> دامنه: ارائه‌دهنده مستقیم `anthropic` و ۱۱ مدل آماده‌ای که در اجرای مبنا از دیتابیس Model Discovery استخراج شدند
>
> نتیجه نهایی: **آمادگی کامل تأیید نمی‌شود؛ عرضه فقط به‌صورت کنترل‌شده در Staging یا برای حساب‌های Allowlist‌شده مجاز است.**

## خلاصه مدیریتی

اصلاحات اصلی برنامه پیاده‌سازی شده‌اند: قیمت‌گذاری Anthropic از Snapshot نسخه‌دار Model Discovery خوانده می‌شود، Credentialهای جدید سازمان رمزگذاری می‌شوند، WIF در قرارداد و Silk Road وجود دارد، Adapter از Tools و Thinking و Structured Output پشتیبانی می‌کند و API بومی Messages، Models، Files و Message Batches عرضه شده است.

بااین‌حال، بررسی دوباره کد و تطبیق آن با مستندات رسمی جاری Anthropic نشان می‌دهد که هنوز نمی‌توان عبارت «پشتیبانی کامل» را برای این ۱۱ مدل به‌کار برد. موانع اصلی انتشار عبارت‌اند از:

1. ماتریس زنده API Key و WIF برای تمام مدل‌ها و قابلیت‌ها اجرا نشده است.
2. Beta خروجی 300k، Server Toolهای مسیر OpenAI-compatible و سقف واقعی Sonnet 4.6 هنوز تعیین تکلیف نشده‌اند.
3. مجموعه آزمون‌های Release Readiness و تأیید انسانی اجرا نشده است.

در این بازبینی، ۹ مغایرت باقی‌مانده ثبت شده است: ۵ مورد P1 و ۴ مورد P2. مغایرت‌های `ANT-REM-001`، `ANT-REM-003`، `ANT-REM-004` و `ANT-REM-005` بسته شده‌اند.

## مبنا و روش ارزیابی

| مورد | مقدار |
|---|---|
| معیار مدل آماده | Provider فعال و Model فعال، روشن و `published` در Snapshot اجرای مبنا |
| تعداد مدل‌های دامنه | ۱۱ |
| مسیرهای Runtime | Control Plane، Silk Road، Tunnel Client و Tunnel Server |
| قراردادها | Anthropic-native و OpenAI-compatible |
| کاتالوگ | `D:\source\models\providers\anthropic` و Model Discovery |
| آزمون زنده | اجرا نشده؛ Credential واقعی API Key و WIF در دسترس این بازبینی نبود |
| وضعیت مدل‌های Claude 5 | خارج از دامنه؛ در Snapshot مبنا جزو ۱۱ مدل آماده نبودند |

این سند وضعیت فعلی را نشان می‌دهد و جایگزین نسخه قبلی است. شرح خط مبنای پیش از اصلاح و چک‌لیست عمومی اجرا‌نشده حذف شده‌اند تا وضعیت‌های قدیمی با نتیجه فعلی اشتباه نشوند.

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
- [Tool Use](https://platform.claude.com/docs/en/agents-and-tools/tool-use/overview)
- [Code Execution](https://platform.claude.com/docs/en/agents-and-tools/tool-use/code-execution-tool)
- [Usage and Cost API](https://platform.claude.com/docs/en/manage-claude/usage-cost-api)
- [Prompt Caching](https://platform.claude.com/docs/en/build-with-claude/prompt-caching)
- [Rate Limits](https://platform.claude.com/docs/en/api/rate-limits)
- [Rate Limit API](https://platform.claude.com/docs/en/manage-claude/rate-limits-api)
- [Errors](https://platform.claude.com/docs/en/api/errors)
- [Workload Identity Federation](https://platform.claude.com/docs/en/manage-claude/workload-identity-federation)
- [WIF API Reference](https://platform.claude.com/docs/en/manage-claude/wif-reference)
- [Fast Mode](https://platform.claude.com/docs/en/build-with-claude/fast-mode)

## تعریف وضعیت و شدت

| وضعیت | تعریف |
|---|---|
| منطبق | قرارداد، پیاده‌سازی و آزمون متناسب وجود دارد. |
| منطبق در کد | پیاده‌سازی و تست خودکار وجود دارد، اما آزمون زنده یا عملیات انتشار انجام نشده است. |
| ناقص | بخشی از قرارداد رسمی قابل استفاده نیست یا منبع حقیقت در Runtime اعمال نمی‌شود. |
| مسدود عملیاتی | کد آماده است، اما Migration، Credential، زیرساخت یا تأیید عملیاتی تکمیل نشده است. |
| نامرتبط | قابلیت برای مدل‌های این دامنه پشتیبانی رسمی ندارد. |

| شدت | تعریف |
|---|---|
| P0 | خطر امنیتی، افشای Secret، محاسبه مالی سراسری نادرست یا نقض جداسازی سازمان‌ها |
| P1 | شکست یک مدل، Endpoint یا قابلیت اصلی اعلام‌شده |
| P2 | شکاف محدود، مشاهده‌پذیری، تاب‌آوری یا تجربه مدیریتی |
| P3 | نقص مستندات یا Metadata بدون اثر مستقیم Runtime |

## موجودی مدل‌ها و نتیجه فعلی

تمام مدل‌های فعال این جدول تحت تأثیر دروازه‌های مشترک آزمون زنده، Credential و Runtime Profile هستند. بنابراین وضعیت «آماده مشروط» به معنی آمادگی کد برای آزمون Staging است، نه تأیید عرضه عمومی.

| مدل | Context | سقف خروجی ثبت‌شده | وضعیت رسمی در تاریخ بازبینی | نتیجه |
|---|---:|---:|---|---|
| `anthropic/claude-haiku-4-5` | 200k | 64k | Alias فعال | آماده مشروط |
| `anthropic/claude-haiku-4-5-20251001` | 200k | 64k | فعال؛ Retirement نه زودتر از ۱۵ اکتبر ۲۰۲۶ | آماده مشروط |
| `anthropic/claude-opus-4-1` | 200k | 32k | Deprecated | فقط عرضه محدود با هشدار مهاجرت |
| `anthropic/claude-opus-4-1-20250805` | 200k | 32k | Deprecated؛ Retirement در ۵ اوت ۲۰۲۶ | فقط عرضه محدود تا Retirement |
| `anthropic/claude-opus-4-5` | 200k | 64k | Alias فعال | آماده مشروط |
| `anthropic/claude-opus-4-5-20251101` | 200k | 64k | فعال؛ Retirement نه زودتر از ۲۴ نوامبر ۲۰۲۶ | آماده مشروط |
| `anthropic/claude-opus-4-6` | 1M | 128k | فعال | آماده مشروط |
| `anthropic/claude-opus-4-7` | 1M | 128k | فعال | آماده مشروط |
| `anthropic/claude-sonnet-4-5` | 200k | 64k | Alias فعال | آماده مشروط |
| `anthropic/claude-sonnet-4-5-20250929` | 200k | 64k | فعال؛ Retirement نه زودتر از ۲۹ سپتامبر ۲۰۲۶ | آماده مشروط |
| `anthropic/claude-sonnet-4-6` | 1M | 128k در YAML فعلی | فعال | ناقص تا تثبیت سقف با Models API |

### نکات چرخه عمر

- Anthropic مدل Opus 4.1 را در ۵ ژوئن ۲۰۲۶ Deprecated کرده و تاریخ Retirement آن را ۵ اوت ۲۰۲۶ اعلام کرده است.
- Importer اکنون `deprecationDate` و `retirementDate` را جدا نگه می‌دارد و پس از Retirement وضعیت مدل را `unavailable` می‌کند.
- پاسخ صحیح پس از Retirement برای Opus 4.1 باید `model_retired` باشد؛ قابل فراخوانی نگه‌داشتن آن معیار قبولی نیست.
- مدل‌های Claude 5 موجود در مخزن YAML تنها پس از Import و احراز معیار «آماده» باید در ارزیابی مستقل بعدی وارد دامنه شوند.

## وضعیت حوزه‌های اصلی

| حوزه | وضعیت فعلی | توضیح |
|---|---|---|
| کاتالوگ و چرخه عمر | منطبق در کد | Lifecycle، منبع رسمی، Runtime Profile و نرخ‌های نرمال‌شده افزوده شده‌اند. |
| API Key پلتفرم | منطبق در کد | Envelope Encryption و Probe وجود دارد. |
| Credential سازمان | منطبق در کد و Migration | ستون Plaintext و Dual-read حذف شده‌اند؛ Runtime فقط نسخه رمزگذاری‌شده را می‌پذیرد. |
| WIF | منطبق در کد | Exchange، Cache، Singleflight و Scope Probe پیاده‌سازی شده‌اند؛ آزمون زنده انجام نشده است. |
| قیمت Token و Cache | منطبق در کد | Snapshot نسخه‌دار، Decimal، Pin در Reserve و Fail Closed پس از ۲۴ ساعت وجود دارد. |
| Web Search Billing | منطبق در کد | بعد `web_search_requests` محاسبه و ممیزی می‌شود. |
| Code Execution Billing | منطبق در کد با Fail Closed | حالت رایگان همراه `web_search_20260209+` یا `web_fetch_20260209+` ممیزی می‌شود؛ حالت مستقلِ قابل‌صورتحساب پیش از Upstream غیرفعال است. |
| Tools سفارشی | منطبق در کد | Function، Tool Choice، Tool Result و Tool Delta نگاشت می‌شوند. |
| Server Toolها | ناقص در OpenAI-compatible | مسیر بومی Body رسمی را عبور می‌دهد؛ Adapter فقط Function Tool را می‌پذیرد. |
| Runtime Profile | منطبق در کد | Profile با Snapshot قیمت از یک Revision اتمیک Pin می‌شود و نبود، ناسازگاری یا کهنگی آن Fail Closed است. |
| Thinking و Sampling | منطبق در کد | قواعد Thinking، Effort، Sampling conflict و پارامترهای مجاز از Runtime Profile خوانده و در Control Plane و Silk Road اعتبارسنجی می‌شوند. |
| Structured Output | منطبق در کد | JSON Schema به `output_config.format` نگاشت می‌شود. |
| Streaming | منطبق در کد | Eventها، Usage نهایی، Tool Delta و خطای میانه Stream حفظ می‌شوند. |
| Image و PDF | منطبق در کد | Base64، URL و `file_id` پشتیبانی می‌شوند. |
| Audio و Video | نامرتبط | برای این ۱۱ مدل ورودی یا خروجی رسمی نیستند. |
| API بومی | منطبق در کد | Messages، Count Tokens، Models، Files و Batches عرضه شده‌اند. |
| Tunnel جریانی | منطبق در کد | Upload و Download بزرگ به‌صورت Stream عبور می‌کنند. |
| Rate Limit upstream | ناقص | Headerها منتقل می‌شوند، اما Rate Limit API در سیاست Runtime وارد نشده است. |
| Retry | ناقص | Retry ناامن انجام نمی‌شود، اما سیاست مستند Backoff و Retry برای عملیات Idempotent وجود ندارد. |
| پنل WIF | ناقص | Auth Mode و Probe دیده می‌شوند؛ Expiry، Scope مؤثر و آخرین Refresh نمایش داده نمی‌شوند. |

## ماتریس Endpoint و Mode

| قابلیت | مسیر رسمی | مسیر گدارAI | وضعیت |
|---|---|---|---|
| Messages همگام | `POST /v1/messages` | همان مسیر | منطبق در کد |
| Messages جریانی | `POST /v1/messages` با `stream=true` | همان مسیر | منطبق در کد |
| OpenAI-compatible Chat | نامرتبط در Anthropic | `POST /v1/chat/completions` | منطبق برای Text و Custom Tool؛ ناقص برای Server Tool |
| Count Tokens | `POST /v1/messages/count_tokens` | همان مسیر | منطبق در کد |
| Models | `GET /v1/models` و `GET /v1/models/{id}` | همان مسیر با تشخیص `anthropic-version` | منطبق در کد |
| Files | خانواده `/v1/files` | همان مسیر | منطبق در کد |
| Message Batches | خانواده `/v1/messages/batches` | همان مسیر | منطبق پایه؛ Beta خروجی 300k ناقص |
| Fast | پارامتر `speed` | Adapter و قرارداد بومی | Metadata مدل‌های 4.6 نیازمند اصلاح معنایی |
| Priority/Service Tier | `service_tier` | عبور و ثبت Usage | منطبق در کد؛ وابسته به قرارداد حساب |

### محدودیت اندازه

| عملیات | سقف اعمال‌شده |
|---|---:|
| Messages و Count Tokens | 32MB |
| Message Batches | 256MB |
| Files | 500MB |

## پارامترها، محتوا و پاسخ

موارد زیر در Adapter فعلی پیاده‌سازی شده‌اند:

- `max_tokens`، `temperature`، `top_p`، `top_k` و `stop_sequences`
- `system`، Metadata و `user_id`
- Function Tools، `tool_choice`، `parallel_tool_calls=false` و `tool_result`
- Prompt Cache و `cache_control`
- `response_format.json_schema` و `output_config.format`
- `service_tier`، `speed` و `inference_geo`
- Text، Image و PDF از Base64، URL و `file_id`
- Thinking، Redacted Thinking، Citation، Tool Use، Stop Reason و Usage

قواعد Thinking ثبت‌شده در Runtime Profile:

- Opus 4.7 فقط Adaptive Thinking را می‌پذیرد و Manual Thinking رد می‌شود.
- Opus 4.6 و Sonnet 4.6 به‌صورت پیش‌فرض Adaptive هستند؛ Manual Thinking با وضعیت Deprecated قابل استفاده است.
- مدل‌های 4.5 و پایین‌تر از Manual Extended Thinking استفاده می‌کنند.
- تعارض Sampling با Adaptive Thinking و مجازبودن `effort=max` اعتبارسنجی می‌شود.

Model Discovery مقدار `profileJson` را با Zod اعتبارسنجی می‌کند. Control Plane همان Profile و قیمت را در یک Snapshot اتمیک نگه می‌دارد و Revision آن را هنگام Reserve در `UsageCostBreakdown` Pin می‌کند. Profile Pin‌شده در قرارداد Tunnel به Silk Road منتقل می‌شود و Validator و Adapter برای Endpoint، Thinking، Effort، Sampling، Prefill، Tools، Structured Output، Modality، Request Control، Beta Header و سقف‌ها فقط از همان Profile استفاده می‌کنند. نبود Profile، اختلاف Revision یا Profile نامعتبر پیش از تماس با Upstream متوقف می‌شود.

## قیمت‌گذاری

نرخ‌های پایه‌ای که باید در Snapshot معتبر مدل‌ها وجود داشته باشند:

| خانواده | Input / MTok | Cache write 5m | Cache write 1h | Cache read | Output / MTok |
|---|---:|---:|---:|---:|---:|
| Claude Haiku 4.5 | $1 | $1.25 | $2 | $0.10 | $5 |
| Claude Sonnet 4.5 و 4.6 | $3 | $3.75 | $6 | $0.30 | $15 |
| Claude Opus 4.5، 4.6 و 4.7 | $5 | $6.25 | $10 | $0.50 | $25 |
| Claude Opus 4.1 | $15 | $18.75 | $30 | $1.50 | $75 |

قواعد تکمیلی:

- Message Batches برای Token و Prompt Cache تخفیف ۵۰ درصدی دارد.
- `inference_geo=us` برای نسل 4.6 و جدیدتر ضریب ۱٫۱ روی ابعاد Token و Cache دارد.
- Web Search برابر ۱۰ دلار به‌ازای هر ۱۰۰۰ درخواست است.
- Code Execution مستقل پس از سهمیه رایگان ماهانه سازمان، بر حسب Container-hour محاسبه می‌شود و حداقل زمان قابل‌صورتحساب دارد.
- Code Execution در همان درخواست همراه `web_search_20260209+` یا `web_fetch_20260209+` هزینه افزوده ندارد.
- نبود Snapshot معتبر، Rate یا Dimension برای Anthropic باید با `pricing_unavailable` و HTTP 503 متوقف شود.

### سیاست ایمن Code Execution

پاسخ Messages مقدار `server_tool_use.code_execution_requests` را گزارش می‌کند، اما کاتالوگ گدارAI نرخ `code_execution_container_hours` را نگه می‌دارد. مدت Container و مانده سهمیه رایگان سازمان از پاسخ Messages به‌دست نمی‌آید. بنابراین نگاشت مستقیم و دقیق بین Usage و Rate وجود ندارد.

Cost API رسمی Anthropic هزینه را به‌صورت تجمیعی با شرح `Code Execution Usage` گزارش می‌کند، آن را در Usage API قرار نمی‌دهد و داده ممکن است با چند دقیقه تأخیر ظاهر شود. این داده برای Finalize قطعی همان درخواست کافی نیست. برای جلوگیری از هزینه صفر یا تسویه حدسی، سیاست زیر پیاده‌سازی شده است:

1. در Messages و تک‌تک آیتم‌های Message Batch، وجود `code_execution_*` بدون `web_search_20260209+` یا `web_fetch_20260209+` با خطای بومی Anthropic و HTTP 400، پیش از Route، Reserve و Tunnel رد می‌شود.
2. حضور Web Tool واجد شرایط به‌صورت شرط قیمت‌گذاری در Reservation و Snapshot همان درخواست Pin می‌شود.
3. اگر پاسخ واجد شرایط `code_execution_requests` داشته باشد، یک مؤلفه ممیزی با Quantity واقعی، Unit برابر `request`، نرخ صفر و ارجاع به بعد کاتالوگ `code_execution_container_hours` ثبت می‌شود.
4. هر Usage اجرای کد بدون شرط Pin‌شده همچنان با `pricing_unavailable` Fail Closed می‌شود؛ Missing Dimension هیچ‌گاه به‌طور ضمنی صفر محاسبه نمی‌شود.
5. کاتالوگ صریحاً `billable_gateway_mode=disabled`، منبع Reconciliation برابر `anthropic_cost_api` و حداقل نسخه Web Tool را ثبت می‌کند.

در نتیجه، اجرای کد رایگان رسمی قابل استفاده است و اجرای مستقل تا زمانی که Anthropic امکان انتساب هزینه قطعی به درخواست را فراهم کند، به‌درستی غیرقابل استفاده است.

## Credential و WIF

موارد پیاده‌سازی‌شده:

- API Keyهای جدید سازمان فقط به‌صورت رمزگذاری‌شده نوشته می‌شوند.
- AAD شامل Tenant، Account، Provider، Auth Mode، Schema Version و Credential Version است.
- Runtime Cache محدود، TTLدار، Singleflight و Tenant-scoped است.
- Runtime فقط نسخه فعال جدول `tenant_provider_account_credential_versions` را می‌خواند و هیچ fallback به Plaintext ندارد.
- Migration شماره 054 تمام Credentialهای Rotation-required را بی‌اعتبار و حذف می‌کند، حساب‌های ناسازگار را Fail Closed می‌کند و ستون `credential_data` را حذف می‌کند.
- ابزار Migrator و متریک Legacy-read حذف شده‌اند؛ نبود نسخه رمزگذاری‌شده به‌معنی نبود Credential است.
- قرار دادن Secret در `provider_config`، حتی در ساختارهای تو‌در‌تو، با HTTP 400 رد می‌شود و Secret فقط از فیلد Write-only مربوط به Credential پذیرفته می‌شود.
- WIF برای Platform و Tenant از Token Source شناسه‌دار و Allowlist‌شده استفاده می‌کند.
- Token Exchange در Silk Road انجام می‌شود و Token پیش از انقضا با Singleflight نوسازی می‌شود.
- Probe قابلیت‌های `workspace:inference` و `workspace:developer` را جدا گزارش می‌کند.

موارد باز:

- آزمون زنده Token Exchange، Refresh هم‌زمان، Scope ناکافی و Rule نامعتبر
- ثبت و نمایش Expiry، Scope مؤثر و آخرین Refresh در پنل

## مغایرت بسته‌شده در این بازبینی

| شناسه | وضعیت | شواهد |
|---|---|---|
| `ANT-REM-001` | بسته | Migration شماره 054، حذف `credential_data` از Prisma و SQLهای Runtime، حذف Dual-read و Migrator، تست PostgreSQL با رکورد Legacy و حفظ Credential رمزگذاری‌شده سالم |
| `ANT-REM-003` | بسته | کنترل پیش از Upstream برای Messages و هر آیتم Batch، Pin شرط حالت رایگان در Reservation، مؤلفه ممیزی صفر فقط برای Web Toolهای واجد شرایط، Fail Closed برای Usage بدون انتساب، Metadata کاتالوگ و تست‌های رگرسیون Billing/Proxy |
| `ANT-REM-004` | بسته | اعتبارسنجی Zod در Model Discovery، Profile نسخه‌دار در Snapshot اتمیک Control Plane، Pin قیمت و Profile با Revision یکسان، عبور قرارداد نوع‌دار از Tunnel، Validator و Adapter مبتنی بر Profile و Fail Closed پیش از Upstream |
| `ANT-REM-005` | بسته | تقاطع `requested ∩ account allowlist ∩ model runtime profile` در Silk Road برای درخواست‌های مدل‌محور و تست رد Header مجاز حساب اما نامجاز مدل |

## مغایرت‌های باقی‌مانده

| شناسه | شدت | مغایرت | اثر | اقدام اصلاحی |
|---|---|---|---|---|
| `ANT-REM-002` | P1 | هیچ ماتریس زنده‌ای برای ۱۱ مدل با API Key و WIF اجرا نشده است | دسترسی واقعی مدل، Scope، Beta و رفتار upstream اثبات نشده است | Smoke و E2E زنده را با Credential کم‌دسترسی اجرا و Fixture پاک‌سازی‌شده ثبت کنید |
| `ANT-REM-006` | P1 | Beta رسمی `output-300k-2026-03-24` برای Batch مدل‌های 4.6 و 4.7 در Runtime Profile ثبت نشده است | سقف رسمی 300k در Batch قابل استفاده مطمئن نیست | Header را فقط برای مدل‌های مجاز به Profile و Connection Schema اضافه و تست مرزی 300k ایجاد کنید |
| `ANT-REM-007` | P1 | OpenAI-compatible فقط Function Tool را می‌پذیرد و Server Toolهای Anthropic را رد می‌کند | Web Search، Web Fetch و Code Execution فقط از API بومی قابل استفاده‌اند | یا قرارداد نوع‌دار Server Tool به OpenAI-compatible اضافه شود، یا این محدودیت در کاتالوگ و مستندات مسیر OpenAI شفاف شود |
| `ANT-REM-008` | P2 | `speed=fast` برای مدل‌های 4.6 در Profile به‌عنوان Mode مجاز نمایش داده می‌شود، درحالی‌که Anthropic آن را با سرعت و قیمت استاندارد اجرا می‌کند | کاربر تصور می‌کند Fast واقعی دریافت کرده است | Mode پذیرفته‌شده را از Mode مؤثر جدا و مقدار `effective_speed=standard` را نمایش و ثبت کنید |
| `ANT-REM-009` | P1 | سقف خروجی Sonnet 4.6 بین Snapshot مبنا، YAML فعلی و صفحات رسمی در حال تغییر یکدست نیست | Validator می‌تواند درخواست معتبر را رد یا درخواست نامعتبر را عبور دهد | `GET /v1/models/claude-sonnet-4-6` را با حساب واقعی اجرا و مقدار، Account، Revision و تاریخ مشاهده را ثبت کنید |
| `ANT-REM-010` | P2 | Rate Limit API Anthropic در Runtime مصرف نمی‌شود | سیاست داخلی از Tier و محدودیت واقعی Workspace آگاه نیست | Snapshot محدودیت‌ها را با Admin API دریافت و Headerهای پاسخ را برای اصلاح پویا حفظ کنید |
| `ANT-REM-011` | P2 | سیاست Retry/Backoff ویژه Anthropic وجود ندارد | خطاهای گذرای Idempotent بدون بازیابی می‌مانند | Retry محدود برای خطاهای مستند و عملیات اثبات‌شده Idempotent با رعایت `Retry-After` پیاده‌سازی کنید |
| `ANT-REM-012` | P2 | Expiry، Scope مؤثر و آخرین Refresh WIF در پنل نمایش داده نمی‌شود | عیب‌یابی و پیشگیری از انقضا دشوار است | وضعیت Sanitized Token را بدون Assertion یا Access Token ذخیره و نمایش دهید |
| `ANT-REM-013` | P1 انتشار | آزمون 500MB، چرخه کامل Batch، قطع Stream، تست امنیت و تأیید انسانی انجام نشده است | آمادگی عملیاتی زیر بار و خطا اثبات نشده است | مجموعه Release Readiness را در Staging اجرا و گزارش مستقل ثبت کنید |

## وضعیت ۱۷ مغایرت اولیه

| شناسه اولیه | وضعیت فعلی | توضیح |
|---|---|---|
| `ANT-FIND-001` | بسته | Pricing Snapshot، LKG 24h، Pin Revision و Fail Closed پیاده‌سازی شده‌اند. |
| `ANT-FIND-002` | بسته | Plaintext، Dual-read و Migrator حذف شده‌اند؛ آزمون Migration تخریبی و مسیر رمزگذاری‌شده موفق است. |
| `ANT-FIND-003` | بسته ایمن | Token، Cache و Web Search تسویه می‌شوند؛ Code Execution رایگان ممیزی و حالت مستقلِ غیرقابل انتساب پیش از Upstream رد می‌شود. |
| `ANT-FIND-004` | بسته برای Custom Tool | Server Tool مسیر OpenAI در `ANT-REM-007` جدا ثبت شده است. |
| `ANT-FIND-005` | بسته | Thinking و Effort از Runtime Profile همان Revision قیمت در Validator و Adapter اعمال می‌شوند. |
| `ANT-FIND-006` | بسته در کد | تمام خانواده‌های API بومی عرضه شده‌اند. |
| `ANT-FIND-007` | بسته در کد | State machine و خطای میانه Stream پیاده‌سازی شده‌اند. |
| `ANT-FIND-008` | بسته در کد | Lifecycle و Retirement خودکار وجود دارد؛ آزمون روز Retirement لازم است. |
| `ANT-FIND-009` | بسته | پارامترهای مدل‌محور و تعارض‌های Sampling از Runtime Profile نسخه‌دار اعتبارسنجی می‌شوند. |
| `ANT-FIND-010` | بسته در کد | آزمون زنده WIF در `ANT-REM-002` باز است. |
| `ANT-FIND-011` | بسته در کد | Probe قابلیت‌محور وجود دارد؛ آزمون Scope واقعی باز است. |
| `ANT-FIND-012` | بسته در کد | URL، Base64 و `file_id` پشتیبانی می‌شوند. |
| `ANT-FIND-013` | بسته | منابع، Lifecycle و Runtime Profile در Importer ثبت می‌شوند. |
| `ANT-FIND-014` | بسته با Fail Closed | Schema چندبعدی حفظ شده و تا فراهم‌شدن انتساب قطعی Cost API، حالت مستقل غیرفعال است. |
| `ANT-FIND-015` | بسته | `assistant_prefill` از 4.6 و 4.7 حذف شده است. |
| `ANT-FIND-016` | باز | در `ANT-REM-009` پیگیری می‌شود. |
| `ANT-FIND-017` | بسته | Source و تاریخ‌های قابل استخراج ثبت می‌شوند. |

## شواهد پیاده‌سازی

| قابلیت | شاهد |
|---|---|
| Snapshot و قیمت‌گذاری | `services/control-plane/internal/billing/anthropic_pricing_snapshot.go` |
| Usage چندبعدی و Decimal | `services/control-plane/internal/billing/service.go` |
| سیاست Code Execution | `services/control-plane/internal/proxy/anthropic_native_request.go` |
| تست رگرسیون Code Execution | `services/control-plane/internal/proxy/anthropic_native_handler_test.go` و `services/control-plane/internal/billing/anthropic_pricing_snapshot_test.go` |
| Adapter Anthropic | `services/silk-road/pkg/server/anthropic.go` |
| WIF | `services/silk-road/pkg/server/anthropic_wif.go` |
| API بومی | `services/control-plane/internal/proxy/anthropic_native_handler.go` |
| Files و Batch | `services/control-plane/internal/proxy/anthropic_native_resources.go` |
| Worker Batch | `services/control-plane/internal/proxy/anthropic_batch_worker.go` |
| Credential سازمان | `services/control-plane/internal/storage/tenant_provider_credential_crypto.go` |
| حذف Plaintext | `db/main/migrations/054_drop_plaintext_tenant_provider_credentials.sql` |
| تست Schema و Credential رمزگذاری‌شده | `services/control-plane/tests/integration/tenant_provider_credential_schema_integration_test.go` |
| جلوگیری از Secret در Provider Config | `services/control-plane/internal/storage/provider_config_secret_validation.go` |
| Operationهای نوع‌دار Tunnel | `packages/contracts/platform/types.go` |
| Runtime Profile و Pricing Import | `apps/model-discovery/scripts/import-models-from-yaml.ts` |
| اعتبارسنجی Profile در Model Discovery | `apps/model-discovery/src/catalog/catalog.service.ts` و `apps/model-discovery/src/catalog/runtime-profile.schema.ts` |
| قرارداد مشترک Runtime Profile | `packages/contracts/platform/model_runtime_profile.go` |
| Pin اتمیک قیمت و Profile | `services/control-plane/internal/billing/anthropic_pricing_snapshot.go` |
| Validator مبتنی بر Profile | `services/control-plane/internal/proxy/anthropic_runtime_profile.go` |
| Adapter مبتنی بر Profile و Beta intersection | `services/silk-road/pkg/server/anthropic_runtime_profile.go` و `services/silk-road/pkg/server/anthropic.go` |
| Schema کاتالوگ | `apps/model-discovery/prisma/schema.prisma` |

## آزمون‌های ثبت‌شده

علاوه بر تست‌های خودکار ثبت‌شده قبلی، Migration شماره 054 روی PostgreSQL موقت با سه حالت Plaintext، نسخه مهاجرت‌شده Rotation-required و نسخه رمزگذاری‌شده سالم اجرا شد. دو حالت ناامن حذف و Credential رمزگذاری‌شده سالم حفظ شد. تست Integration حذف ستون Plaintext نیز موفق بود.

برای `ANT-REM-003` تست‌های مستقل زیر ثبت و اجرا شدند:

- رد Code Execution مستقل و رد Web Tool قدیمی
- پذیرش ترکیب رایگان رسمی
- اعتبارسنجی تک‌تک آیتم‌های Batch و جلوگیری از تعمیم شرط یک آیتم به آیتم دیگر
- ثبت مؤلفه صفر قابل ممیزی برای `code_execution_requests`
- Fail Closed برای Usage اجرای کد بدون شرط Pin‌شده
- اجرای موفق `go test ./...` و `go vet ./...` در Control Plane
- اجرای موفق ۴۱ تست، Build، Typecheck و Import خشک ۲۷ مدل Anthropic در Model Discovery

برای `ANT-REM-004` و `ANT-REM-005` تست‌های مستقل زیر ثبت و اجرا شدند:

- رد Snapshot فاقد Runtime Profile یا دارای Profile نامعتبر
- اثبات یکسان‌بودن Revision قیمت و Profile Pin‌شده در Reserve و Finalize
- Round-trip قرارداد Runtime Profile در `TunnelRequest`
- اعمال رفتار مدل آینده بدون اتکا به نام Hard-coded مدل
- رد Assistant Prefill، Sampling، Thinking، Effort، Modality، Tool، Request Control و سقف نامجاز بر اساس Profile
- رد درخواست Anthropic فاقد Profile پیش از هر تماس Upstream
- تقاطع Beta Header در سطح درخواست، Allowlist حساب و تمام Profileهای مدل‌های Batch
- اجرای موفق ۴۲ تست، Build، Typecheck و Import خشک ۲۷ مدل Anthropic در Model Discovery
- اجرای موفق `go test ./...` و `go vet ./...` در Control Plane و Silk Road

این موفقیت جایگزین موارد زیر نیست:

- آزمون زنده تمام ۱۱ مدل با API Key
- آزمون زنده تمام ۱۱ مدل با WIF
- آزمون Scopeهای `workspace:inference` و `workspace:developer`
- آزمون زنده حالت رایگان Code Execution و تأیید رد حالت مستقل پیش از Upstream
- آزمون Batch شامل Success، Error، Cancel، Expiry و خروجی 300k
- آزمون فایل مصنوعی 500MB با Backpressure
- آزمون قطع Upload، Download و SSE
- آزمون Opus 4.1 در تاریخ Retirement
- تأیید انسانی گزارش نهایی

## چک‌لیست اقدامات باقی‌مانده

### امنیت و Credential

- [x] Dual-read و ابزار Migrator حذف شوند.
- [x] Credentialهای Rotation-required بدون انتقال Secret قدیمی بی‌اعتبار و حذف شوند.
- [x] حساب API Key بدون نسخه رمزگذاری‌شده Fail Closed شود.
- [x] `credential_data` از دیتابیس، Prisma و SQLهای Runtime حذف شود.
- [x] Credential رمزگذاری‌شده سالم در Migration حفظ شود.
- [x] ذخیره Secret در `provider_config` و ساختارهای تو‌در‌تو رد شود.
- [ ] هیچ Secret، Assertion یا Access Token در Log، Trace و پاسخ مشاهده نشود.

### Runtime و قابلیت‌ها

- [x] Runtime Profile همراه Pricing Snapshot در Control Plane دریافت و Pin شود.
- [x] Silk Road قواعد Thinking، Sampling، Prefill، Speed، Beta و سقف‌ها را از Profile دریافت کند.
- [x] Beta Header با Allowlist حساب و Profile مدل تقاطع داده شود.
- [ ] `output-300k-2026-03-24` برای Batch مدل‌های مجاز اضافه و تست شود.
- [ ] محدودیت Server Tool در OpenAI-compatible رفع یا به‌صورت صریح مستند شود.
- [ ] رفتار مؤثر `speed=fast` برای مدل‌های 4.6 به‌عنوان Standard ثبت شود.

### Billing و محدودیت‌ها

- [x] Code Execution مستقل تا فراهم‌شدن انتساب قطعی Cost API پیش از Upstream غیرفعال شود.
- [x] حالت رایگان Code Execution فقط با Web Tool واجد شرایط Pin و ممیزی شود.
- [x] هیچ Missing Dimension با هزینه صفر ضمنی ادامه پیدا نکند.
- [ ] سقف Sonnet 4.6 با Models API حساب واقعی ثبت شود.
- [ ] Rate Limitهای واقعی Workspace در Snapshot عملیاتی وارد شوند.
- [ ] Retry فقط برای عملیات Idempotent و خطاهای مستند اعمال شود.

### آزمون زنده و انتشار

- [ ] Credential واقعی API Key کم‌دسترسی در Staging فراهم شود.
- [ ] WIF واقعی با Scopeهای حداقلی فراهم شود.
- [ ] ماتریس Text Sync/Stream، Image، PDF، Tools، Structured Output و Thinking برای هر ۱۱ مدل اجرا شود.
- [ ] Count Tokens، Files و Message Batches به‌صورت E2E اجرا شوند.
- [ ] Tenant isolation و Account ambiguity آزموده شوند.
- [ ] تست 500MB و Backpressure بدون رشد حافظه متناسب با فایل موفق باشد.
- [ ] نتیجه Retirement مدل Opus 4.1 بررسی شود.
- [ ] تمام P0 و P1ها بسته و گزارش توسط بازبین انسانی تأیید شود.

## معیار اعلام آمادگی کامل

ارائه‌دهنده Anthropic فقط زمانی «کاملاً آماده» است که تمام شرایط زیر هم‌زمان برقرار باشند:

- هیچ مغایرت P0 یا P1 باز نباشد.
- هیچ Credential قدیمی قابل خواندن باقی نمانده باشد.
- Snapshot قیمت کمتر از ۲۴ ساعت عمر داشته باشد.
- Runtime Profile همان Revision کاتالوگ در Validator و Adapter اعمال شود.
- تمام ۱۱ مدل نتیجه موفق زنده یا نتیجه چرخه عمر مورد انتظار داشته باشند.
- Code Execution هزینه قابل ممیزی داشته باشد یا رسماً غیرفعال باشد.
- API Key و WIF هر دو در Staging موفق باشند.
- Load، Stream، Batch، Security و Tenant isolation موفق باشند.
- بازبین انسانی گزارش و شواهد را تأیید کرده باشد.

تا پیش از تحقق این شروط، وضعیت صحیح ارائه‌دهنده Anthropic در گدارAI **«آماده برای آزمون کنترل‌شده، نه آماده برای پشتیبانی کامل عمومی»** است.
