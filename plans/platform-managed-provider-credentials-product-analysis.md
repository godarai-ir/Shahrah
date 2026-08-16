# تحلیل محصولی مدیریت اعتبارنامه‌های پلتفرمی ارائه‌دهندگان

> وضعیت: پیشنهاد محصولی برای تصمیم‌گیری  
> تاریخ: ۱۴۰۵/۰۴/۲۶  
> محدوده: اعتبارنامه‌ای که گدارAI هنگام نبود BYOK کاربر برای فراخوانی ارائه‌دهنده استفاده می‌کند

## خلاصه تصمیم

گدارAI باید credentialهای پیش‌فرض ارائه‌دهندگان را از متغیرهای محیطی پراکنده به مفهوم مستقل **Platform Provider Connection** منتقل کند. هر Connection یک اتصال عملیاتی متعلق به خود پلتفرم است که provider، محیط، روش احراز هویت، تنظیمات غیرمحرمانه، نسخه credential، وضعیت سلامت و دامنه مدل/منطقه را مشخص می‌کند.

`support-admin` باید سطح کنترل این اتصال‌ها باشد، اما مقدار secret پس از ثبت هرگز دوباره به مرورگر برنگردد. `model-discovery` فقط تعریف provider و schema فیلدهای اتصال را نگه می‌دارد؛ secret و تنظیمات runtime در `control-plane` و یک secret store امن نگه‌داری می‌شوند.

پیشنهاد MVP:

- مدیریت چند Connection پلتفرمی برای هر provider، نه یک env ثابت برای هر provider؛
- فرم پویا و provider-specific بر اساس schema؛
- ذخیره رمز‌شده و نسخه‌دار credential، تست قبل از فعال‌سازی و rollback؛
- انتخاب صریح BYOK یا credential پلتفرم در runtime؛
- RBAC و audit اختصاصی در `support-admin`؛
- مهاجرت تدریجی با dual-read کنترل‌شده از DB و env؛
- توقف فوری ذخیره secret در `model-discovery.ProviderConnectionProfile.configJson`.

## تعریف دقیق مسئله

این سند سه نوع secret را از هم جدا می‌کند:

| نوع | مالک | کاربرد | در محدوده این سند |
|---|---|---|---|
| Platform provider credential | گدارAI | فراخوانی upstream وقتی کاربر BYOK ندارد | بله |
| Tenant BYOK credential | tenant یا کاربر | فراخوانی upstream با حساب خود مشتری | خیر؛ فقط مرزبندی runtime بررسی می‌شود |
| Internal service credential | سرویس‌های گدارAI | ارتباط `support-admin`، `control-plane` و `model-discovery` | خیر |

عبارت «Platform Managed» در این سند به معنی مدیریت credential متعلق به گدارAI است؛ ارتباطی با PAT/VAT گدارAI ندارد.

## وضعیت موجود

### ۱. credential مشترک runtime از env وارد می‌شود

در `services/control-plane/configs/proxy_config.yaml` مدل‌های پیش‌فرض OpenAI و Anthropic به `${OPENAI_API_KEY}` و `${ANTHROPIC_API_KEY}` متصل‌اند. `LoadConfig` نیز referenceهای `${...}` را در startup resolve می‌کند. بنابراین تغییر key به تغییر secret استقرار و معمولاً restart یا rollout سرویس نیاز دارد.

### ۲. fallback به credential مشترک وجود دارد

Router برای Provider Accountهایی که BYOK ندارند، تنظیمات اتصال و API key را از deployment مشترک همان provider/model برمی‌دارد. این همان مسیر مصرف wallet-billed و Godar-managed است، اما منبع آن هنوز config/env است.

### ۳. مدل داده‌های هم‌پوشان و ناامن وجود دارد

در دیتابیس فعلی سه محل نزدیک به credential دیده می‌شود:

- `model_deployments.provider_config` که می‌تواند `api_key` را داخل JSON نگه دارد؛
- جدول عمومی `credentials` با `credential_values` از نوع JSON؛
- `tenant_provider_accounts.credential_data` برای BYOK tenant.

در مسیرهای بررسی‌شده، رمزنگاری در سطح application برای این مقادیر وجود ندارد. استفاده از این ساختارها به همان شکل برای platform credential، مسئله env را با ذخیره secret خام در DB جایگزین می‌کند و وضعیت مطلوب نیست.

### ۴. abstraction فعلی همه providerها را واقعاً پوشش نمی‌دهد

`ProviderEndpointConfig` تعدادی فیلد provider-specific دارد، ولی تبدیل runtime عمدتاً روی `api_key`، `api_base`، `api_version` و `region` متمرکز است. برای نمونه:

- OpenAI معمولاً `api_key` و در صورت نیاز `base_url` می‌خواهد؛
- Azure OpenAI/Foundry به endpoint، API version، deployment و یکی از API Key یا Entra credential نیاز دارد؛
- Bedrock ممکن است access key/secret، session token، SigV4، AssumeRole و region بخواهد؛ adapter فعلی Bedrock هنوز SigV4 واقعی را پیاده نمی‌کند و مسیر ساده `x-api-key` دارد.

پس «یک JSON آزاد» یا «یک API Key برای هر provider» قرارداد کافی برای محصول نیست.

### ۵. `support-admin` بخشی از UI مورد نیاز را دارد، ولی boundary اشتباه است

صفحه `/admin/providers` هم‌اکنون برای OpenAI-like، Azure، Bedrock، Vertex و Databricks فیلدهای متفاوت نشان می‌دهد. این نقطه شروع UX مفیدی است، اما در وضع فعلی:

- schema فیلدها در frontend hardcode شده است؛
- secretها همراه `configJson` به `model-discovery` فرستاده می‌شوند؛
- `ProviderConnectionProfile.configJson` در model-discovery به‌صورت JSONB ذخیره می‌شود؛
- connection profile در read API دوباره با provider برگردانده می‌شود؛
- فیلدهایی مانند Azure deployment و Bedrock model ID با credential account مخلوط شده‌اند.

این مسیر نباید برای credential واقعی production استفاده شود.

## محدودیت‌های محصولی روش env

| محدودیت | اثر محصولی/عملیاتی |
|---|---|
| یک مقدار سراسری و بدون هویت مستقل | چند حساب، region یا قرارداد برای یک provider قابل مدیریت نیست |
| تغییر وابسته به deployment | زمان چرخش بالا و نیازمند هماهنگی DevOps است |
| نبود workflow تست و فعال‌سازی | key خراب می‌تواند مستقیماً traffic production را مختل کند |
| نبود version و rollback | بازگشت سریع به credential سالم دشوار است |
| نبود تاریخچه و مالک تغییر | پاسخ‌گویی پشتیبانی و audit ناقص است |
| نبود وضعیت سلامت | اپراتور نمی‌داند credential منقضی، محدود یا فاقد دسترسی مدل است |
| schema ثابت | تفاوت Azure، AWS، Google و API-key providerها به JSON یا envهای پراکنده منتقل می‌شود |
| restart برای اعمال تغییر | عملیات rotation می‌تواند availability را کاهش دهد |
| fallback ضمنی | تشخیص اینکه هزینه با BYOK یا wallet گدارAI محاسبه شده دشوار می‌شود |

## وضعیت مطلوب محصول

### مفهوم اصلی: Platform Provider Connection

هر Connection نماینده یک حساب یا مسیر اتصال واقعی متعلق به گدارAI است؛ برای مثال:

- `OpenAI / Production / Primary`
- `Azure Foundry / Sweden Central / Contract A`
- `AWS Bedrock / eu-central-1 / Primary Role`

ویژگی‌های اصلی:

- provider و environment؛
- نام عملیاتی یکتا؛
- auth mode صریح؛
- config غیرمحرمانه؛
- reference به نسخه فعال secret؛
- region و محدوده مدل‌ها/capabilityها؛
- وضعیت lifecycle و سلامت؛
- priority/weight برای routing؛
- timestamps و actor آخرین تغییر.

وجود چند Connection برای یک provider باید از روز اول در مدل داده مجاز باشد، حتی اگر MVP routing فقط یک Connection فعال primary را انتخاب کند.

### جداسازی چهار لایه داده

| لایه | نمونه | محل مالکیت پیشنهادی |
|---|---|---|
| Provider definition | نام، لوگو، auth modeهای مجاز | `model-discovery` |
| Connection schema | فیلدها، نوع، required rule، secret flag | provider registry/catalog؛ بدون مقدار واقعی |
| Connection config | endpoint، region، API version | `control-plane` |
| Secret material | API key، client secret، AWS secret | secret store یا ciphertext رمز‌شده تحت کنترل `control-plane` |

`model-discovery` نباید secret store یا runtime source of truth باشد.

### فرم schema-driven

برای هر provider/auth mode یک schema نسخه‌دار تعریف شود. هر field حداقل این metadata را دارد:

- `key`, `label`, `type`؛
- `secret: true|false`؛
- `required` یا `required_when`؛
- validation مانند URL، enum، region یا JSON؛
- scope: `connection` یا `model_deployment`؛
- runtime mapping؛
- help text و documentation URL؛
- redaction policy.

نمونه تفکیک:

| Provider | Connection/Auth fields | Deployment/Model fields |
|---|---|---|
| OpenAI | API key، organization/project اختیاری، base URL اختیاری | model ID |
| Azure Foundry | endpoint/resource، API version، API key یا Entra tenant/client/secret | deployment name و model mapping |
| Bedrock | region، access key/secret/session یا role ARN/external ID | model ID یا inference profile ARN |
| Vertex AI | project، location، service account/ADC mode | publisher/model/endpoint |

Azure deployment و Bedrock model ID نباید داخل credential bundle ذخیره شوند، چون با یک credential می‌توان به چند deployment/model دسترسی داشت.

## تجربه مطلوب در `support-admin`

### فهرست Connectionها

برای هر مورد نشان داده شود:

- provider، نام Connection، environment و region؛
- auth mode؛
- وضعیت `Draft`, `Testing`, `Active`, `Degraded`, `Disabled`؛
- نسخه فعال credential و زمان آخرین rotation؛
- آخرین تست موفق/ناموفق و خطای خلاصه‌شده؛
- مدل‌ها یا deploymentهای وابسته؛
- primary/fallback بودن.

### ایجاد یا تغییر

1. انتخاب provider و environment؛
2. انتخاب auth mode؛
3. نمایش فرم provider-specific از schema؛
4. ثبت config و secret به‌عنوان draft version؛
5. تست اتصال و دسترسی حداقلی؛
6. نمایش نتیجه قابل اقدام؛
7. فعال‌سازی نسخه جدید؛
8. نگه‌داشتن نسخه قبلی در grace period برای rollback.

«Save» نباید مستقیماً credential فعال را overwrite کند. update یک نسخه جدید می‌سازد و activation یک عملیات جداگانه و audited است.

### نمایش secret

- secret موجود هرگز در GET response یا HTML/JS مرورگر برنگردد؛
- مقدار جدید فقط در input وارد و از browser به backend ارسال شود؛
- پاسخ فقط `configured=true`، hint ماسک‌شده، version و timestamps بدهد؛
- گزینه «نمایش secret فعلی» وجود نداشته باشد؛
- input خالی در edit یعنی «بدون تغییر»، نه حذف credential؛
- حذف یا revoke عملیات مستقل و نیازمند تأیید باشد.

### RBAC پیشنهادی

مجوز عمومی `provider:write` برای این حوزه بیش از حد وسیع است. مجوزهای جدا پیشنهاد می‌شود:

- `platform_credentials:read_metadata`
- `platform_credentials:create_version`
- `platform_credentials:test`
- `platform_credentials:activate`
- `platform_credentials:disable`

`support_agent` فقط metadata و health را ببیند. ایجاد/فعال‌سازی به نقش محدودتر مانند `credential_admin` یا `super_admin` تعلق داشته باشد. هیچ نقشی مجوز reveal secret ذخیره‌شده نداشته باشد. فعال‌سازی production در P1 می‌تواند step-up authentication یا تأیید نفر دوم داشته باشد.

## رفتار مطلوب runtime

### انتخاب منبع credential

قانون باید صریح و قابل مشاهده باشد:

1. اگر target یک Provider Account با BYOK معتبر است، فقط همان BYOK استفاده شود و billing برابر `external_billed` باشد.
2. اگر target فاقد BYOK و برای platform-managed usage مجاز است، Connection فعال پلتفرم resolve شود و billing برابر `wallet_billed` باشد.
3. اگر Connection فعال یا credential سالم وجود ندارد، target پیش از تماس upstream unavailable شود.
4. شکست credential BYOK نباید به‌طور ضمنی به credential پلتفرم fallback کند؛ این کار مالک هزینه را تغییر می‌دهد. fallback فقط با policy صریح مجاز است.

در trace و billing metadata باید `credential_source`, `connection_id` و `credential_version_id` ثبت شوند، اما هیچ secret یا header احراز هویت ثبت نشود.

### اعمال تغییر بدون restart

runtime resolver باید نسخه فعال Connection را با cache کوتاه‌عمر یا event-driven invalidation دریافت کند. activation باید cache را invalid کند. هدف پذیرش تغییر در کمتر از ۶۰ ثانیه بدون restart است.

### مرز ایران/اروپا

برای MVP می‌توان secret را در `control-plane` resolve و فقط برای همان dispatch از تونل mTLS به egress فرستاد؛ secret در اروپا persist نشود. در فاز بعد می‌توان secret broker یا reference resolution در egress را بررسی کرد. در هر دو حالت باید log/tracing redaction و memory lifetime محدود باشد.

## مدل داده پیشنهادی

### `platform_provider_connections`

- `connection_id`
- `provider_slug`
- `name`
- `environment`
- `auth_mode`
- `schema_version`
- `config_json` — فقط non-secret
- `status`
- `region`
- `priority`, `weight`
- `active_credential_version_id`
- `last_test_status`, `last_tested_at`, `last_success_at`
- `created_by`, `updated_by`, timestamps

### `platform_provider_credential_versions`

- `credential_version_id`
- `connection_id`
- `version`
- `secret_backend` (`encrypted_db`, `vault`, `aws_sm`, ...)
- `secret_ref` یا `ciphertext`, `nonce`, `wrapped_dek`, `kek_version`
- `fingerprint` غیرقابل برگشت برای تشخیص تغییر/تکرار
- `status` (`draft`, `validated`, `active`, `retiring`, `revoked`)
- `valid_from`, `expires_at`, `retire_at`
- `created_by`, `activated_by`, timestamps

### `platform_provider_connection_scopes`

- `connection_id`
- provider model/deployment pattern
- capability/region constraints
- enabled flag

### `platform_provider_credential_events`

- actor و request/correlation ID؛
- action؛
- connection و version؛
- نتیجه و reason؛
- metadata redacted؛
- timestamp و source IP/session metadata در حد سیاست امنیتی.

## ذخیره امن

### MVP قابل قبول

Envelope encryption در application:

- یک DEK تصادفی برای هر credential version؛
- رمزنگاری authenticated مانند AES-256-GCM؛
- wrap شدن DEK با KEK خارج از DB؛
- versioning برای KEK؛
- ciphertext و metadata در DB؛
- KEK یا دسترسی KMS تنها bootstrap secret باقی‌مانده در env/deployment باشد.

هدف حذف کامل env نیست؛ env باید فقط bootstrap اعتماد را نگه دارد، نه credential تک‌تک providerها.

### فاز بعد

پشتیبانی از HashiCorp Vault یا secret manager ابری پشت interface واحد. UI و runtime نباید به backend ذخیره‌سازی وابسته باشند و گدارAI باید implementation کوچک و Go-native خود را داشته باشد.

## API داخلی پیشنهادی

### Support/control APIs

- `GET /api/v1/platform/provider-connection-schemas/:provider`
- `GET /api/v1/platform/provider-connections`
- `POST /api/v1/platform/provider-connections`
- `PATCH /api/v1/platform/provider-connections/:id` — فقط metadata/config غیرمحرمانه
- `POST /api/v1/platform/provider-connections/:id/credential-versions`
- `POST /api/v1/platform/provider-connections/:id/credential-versions/:version/test`
- `POST /api/v1/platform/provider-connections/:id/credential-versions/:version/activate`
- `POST /api/v1/platform/provider-connections/:id/rollback`
- `POST /api/v1/platform/provider-connections/:id/disable`

تمام read responseها redacted هستند. `support-admin` این endpointها را فقط از route handler سروری و با service credential scoped فراخوانی می‌کند.

### Runtime contract

runtime به جای دریافت map آزاد، یک `ResolvedProviderAuth` typed دریافت کند که auth mode و material لازم برای adapter را مشخص می‌کند. adapterهای OpenAI، Azure، Bedrock و Vertex باید قرارداد typed مربوط به خود را validate کنند. این تغییر پیش‌نیاز پشتیبانی واقعی SigV4/Entra است.

## تست credential

تست نباید الزاماً یک completion هزینه‌دار اجرا کند. سطح‌های تست:

1. validation محلی schema و URL؛
2. authentication probe کم‌هزینه مثل list models در صورت پشتیبانی؛
3. authorization check برای model/deploymentهای scope؛
4. optional inference smoke test با مدل تعیین‌شده.

نتیجه باید خطاها را به دسته‌های قابل اقدام تبدیل کند: `invalid_credential`, `expired`, `forbidden`, `quota_exhausted`, `region_mismatch`, `deployment_not_found`, `network_error`. متن خام upstream فقط در technical details محدود و redacted نمایش داده شود.

## الزامات غیرعملکردی و امنیتی

- secret در log، trace، audit payload، error، cache distributed و analytics ظاهر نشود؛
- TLS برای browser-to-support-admin و service-to-service؛ محدودسازی شبکه برای provider egress؛
- read APIها فقط metadata redacted برگردانند؛
- rotation بدون downtime و با rollback؛
- optimistic locking برای جلوگیری از overwrite هم‌زمان؛
- idempotency key برای create/activate؛
- backup رمز‌شده و retention مشخص برای نسخه‌های revoked؛
- SSRF protection برای endpointهای قابل تنظیم؛
- health checkها rate-limited و دارای timeout؛
- public/provider catalog API هیچ connection value یا secret metadata حساس منتشر نکند.

## شکاف‌های فوری و اولویت‌ها

### P0 — قبل از استفاده production

- جلوگیری از ذخیره `apiKey`, `clientSecret`, `secretAccessKey`, `sessionToken` و service-account JSON در `model-discovery.configJson`؛
- حذف connection profile حاوی مقدار از public/read responses یا redaction قطعی آن؛
- بررسی اینکه آیا credential واقعی تاکنون در این جدول ذخیره شده است؛ در صورت وجود، rotation و پاک‌سازی امن؛
- ممنوع کردن بازگرداندن secret موجود به `support-admin`؛
- تعریف boundary: model-discovery = schema/catalog، control-plane = runtime connection.

### P1 — MVP مدیریت credential پلتفرم

- schema و APIهای Platform Provider Connection؛
- encrypted credential versions؛
- RBAC و audit؛
- فرم پویا در `support-admin`؛
- تست، activate، disable و rollback؛
- resolver runtime برای OpenAI/Anthropic و سپس Azure/Bedrock؛
- metric و alert برای authentication failure و expiry.

### P2 — بلوغ عملیاتی

- چند Connection و weighted/fallback routing؛
- external secret manager؛
- approval دو نفره برای production؛
- auto-rotation در providerهای پشتیبان؛
- credential-level budget/usage dashboard؛
- egress-side secret resolution.

## برنامه مهاجرت پیشنهادی

### فاز ۰: مهار ریسک

- secret inputهای فعلی model-discovery را غیرفعال یا به backend جدید هدایت کنید؛
- داده موجود را audit و credentialهای افشاشده احتمالی را rotate کنید؛
- response redaction را اعمال کنید.

### فاز ۱: foundation

- مدل داده، envelope encryption، schema registry، RBAC و audit؛
- APIهای redacted و unit/security tests؛
- OpenAI به‌عنوان provider اول.

### فاز ۲: عملیات در `support-admin`

- list/create/edit/test/activate/rollback؛
- status و health؛
- UX صریح برای تفاوت connection config و model deployment config.

### فاز ۳: runtime cutover

- حالت‌های source: `env`, `platform_store`, `platform_store_with_env_break_glass`؛
- shadow validation credential جدید؛
- انتقال provider به provider؛
- ثبت source/version در trace و billing؛
- حذف fallback ضمنی پس از تثبیت.

### فاز ۴: جمع‌کردن envهای provider

- حذف `OPENAI_API_KEY`, `ANTHROPIC_API_KEY` و مشابه از deployment؛
- نگه‌داشتن فقط KMS/KEK bootstrap و service credentials؛
- runbook rotation و disaster recovery.

## معیارهای موفقیت

- ۱۰۰٪ تغییرهای credential از مسیر audited انجام شوند؛
- هیچ secret در GET response، log، trace یا audit payload دیده نشود؛
- rotation بدون restart و بدون downtime انجام شود؛
- زمان اعمال تغییر production کمتر از ۶۰ ثانیه باشد؛
- rollback نسخه قبلی کمتر از ۵ دقیقه انجام شود؛
- نرخ خطای `invalid_credential` پس از rotation از آستانه تعریف‌شده عبور نکند؛
- اپراتور بتواند source هزینه (`BYOK` یا `Godar-managed`) را برای هر request تشخیص دهد.

## تصمیم‌های Include / Defer / Exclude

| قابلیت | تصمیم | دلیل |
|---|---|---|
| Provider-specific schema | Include now | شرط لازم برای Azure/Bedrock/Vertex |
| Encrypted versioned storage | Include now | حداقل امنیت و rollback production |
| support-admin workflow | Include now | نیاز صریح عملیات پشتیبانی |
| Test-before-activate | Include now | کاهش outage ناشی از rotation |
| چند Connection در data model | Include now | جلوگیری از بن‌بست معماری |
| weighted multi-credential routing | Defer | MVP می‌تواند یک primary فعال داشته باشد |
| external secret-manager integrations | Defer | interface را اکنون تعریف، integration را بعداً اضافه کنید |
| auto-rotation | Defer | وابسته به provider و ops-heavy است |
| reveal secret | Exclude | نیاز عملیاتی قابل دفاع ندارد و ریسک بالاست |
| ذخیره secret در catalog/model-discovery | Exclude | نقض boundary و احتمال exposure |
| fallback خاموش BYOK به کلید پلتفرم | Exclude | مالک billing را بدون رضایت تغییر می‌دهد |

## ریسک‌ها و پرسش‌های باز

### تصمیم محصولی لازم

- آیا credential پلتفرم در MVP برای هر environment یک primary دارد یا per-region primary؟
- چه نقش سازمانی مجاز به activate production است؟
- grace period استاندارد rotation چقدر باشد؟
- اگر BYOK کاربر fail شد، آیا policy اختیاری و شفاف برای fallback wallet-billed ارائه می‌شود یا کاملاً ممنوع می‌ماند؟

### تصمیم فنی لازم

- KEK با KMS نگه‌داری شود یا در MVP با deployment secret؟
- secret در control-plane resolve و با درخواست provider-egress حمل شود یا در egress resolve شود؟
- invalidation runtime با event bus انجام شود یا cache TTL کوتاه برای MVP کافی است؟
- adapterهای Azure Entra و AWS SigV4 در چه فازی production-ready می‌شوند؟

## شواهد بررسی‌شده

### اسناد گدارAI

- `plans/README.md`
- `plans/phase0-tenant-provider-account-contract.md`
- `plans/phase1-tenant-provider-account-backend-foundation.md`
- `plans/provider-account-models-ux-flow.md`
- `plans/truefoundry-provider-account-management-reference.md`

### کد و schema گدارAI

- `services/control-plane/configs/proxy_config.yaml`
- `services/control-plane/internal/config/config.go`
- `services/control-plane/internal/router/router.go`
- `services/control-plane/internal/storage/deployment_repo.go`
- `services/control-plane/internal/storage/provider_account_runtime_repo.go`
- `services/control-plane/internal/providerregistry/`
- `services/control-plane/internal/proxy/`
- `packages/database/schema.prisma`
- `apps/support-admin/app/admin/providers/page.tsx`
- `apps/support-admin/lib/admin-auth.ts`
- `apps/model-discovery/src/providers/providers.service.ts`
- `apps/model-discovery/prisma/schema.prisma`

## تفکیک fact، recommendation و assumption

**Fact:** envهای OpenAI/Anthropic در config پیش‌فرض resolve می‌شوند؛ Router برای مسیر non-BYOK از deployment مشترک استفاده می‌کند؛ secret-like values در چند JSON column قابل ذخیره‌اند؛ support-admin اکنون secret-like fields را به connection profile در model-discovery می‌فرستد؛ encryption application-level برای این مسیرها پیدا نشد.

**Recommendation:** Platform Provider Connection نسخه‌دار، رمز‌شده و تحت مالکیت control-plane ایجاد شود؛ support-admin فقط control surface باشد؛ model-discovery فقط schema/catalog نگه دارد؛ runtime source و billing mode صریح باشند.

**Assumption:** نیاز کوتاه‌مدت یک primary credential فعال برای هر provider/environment است و multi-credential balancing می‌تواند بعد از MVP عرضه شود. همچنین فرض شده است تغییر production باید بدون restart و با rollback قابل انجام باشد.
