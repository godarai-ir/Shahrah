# TrueFoundry Provider Account & Model Access Management

## هدف سند
این سند یک مرجع محصولی برای بخش «مدیریت اکانت پروایدر و مدل‌ها» در TrueFoundry است تا مبنای تحلیل شکاف (پیاده‌سازی شده/نشده) در GodarAI باشد.

## محدوده
این سند روی AI Gateway متمرکز است و شامل این محورهاست:
- ساختار مدیریت Provider Account و Integrations (مدل‌ها)
- کنترل دسترسی (Collaborators, Roles)
- نوع مدل‌ها (Model Types) و ارتباط با Endpointها
- مدیریت Cost (Public/Private)
- نام‌گذاری و Display Name
- کشف مدل‌ها (`/models`) و لیست Integrations
- مرزبندی Tenant / Account / Team / Virtual Account

## خلاصه اجرایی
- در TrueFoundry، «Provider Account» ظرف اصلی اتصال به یک vendor (مثل OpenAI/Anthropic/Bedrock/...) است.
- هر Provider Account می‌تواند چندین «Model Integration» داشته باشد؛ در عمل همان گروه مدل‌های قابل استفاده زیر یک حساب پروایدر است.
- دسترسی با Collaborators در سطح Provider Account مدیریت می‌شود و دو نقش کلیدی دارد: `provider-account-manager` و `provider-account-access`.
- مصرف در Gateway با کلیدهای TrueFoundry انجام می‌شود (PAT/VAT)، نه با کلید خام پروایدر.
- هر مدل با قابلیت/نوع مشخص ثبت می‌شود (مثل chat, embedding, realtime, rerank, audio transcription) و همین نوع، endpointهای مجاز را تعیین می‌کند.
- برای Cost دو رویکرد وجود دارد: `public_cost` (کاتالوگ عمومی) و `private` (تعرفه سفارشی).

## مدل مفهومی

### 1) Provider Account
Provider Account نماینده یک اتصال منطقی به یک vendor است و معمولا شامل این موارد می‌شود:
- `type`: مثل `provider-account/openai`
- `name`: نام یکتا در tenant
- `auth_data`: اطلاعات احراز هویت پروایدر (API key یا روش‌های معادل)
- `collaborators`: کاربران/تیم‌ها و نقش دسترسی
- `integrations`: لیست مدل‌های اضافه‌شده زیر همان حساب

### 2) Model Integration
هر Integration یک مدل قابل مصرف در Gateway است و شامل:
- نام داخلی/نمایشی مدل در gateway
- شناسه مدل upstream (مثل `gpt-4o`)
- `model_types` یا `model_type` برای اعلام قابلیت
- پیکربندی cost

### 3) رابطه Provider Account و Model Group
برداشت شما درست است: در تجربه عملی UI، وقتی یک Provider Account می‌سازید و چند مدل به آن اضافه می‌کنید، همان عملا یک «گروه مدل» قابل کنترل از نظر دسترسی و مدیریت است. در مستندات API این گروه با مفهوم `provider account + integrations` نمایش داده می‌شود.

## جریان عملیاتی استاندارد

### گام 1: ساخت Provider Account
- انتخاب provider
- ورود credential اختصاصی همان provider
- تعیین collaborators

### گام 2: افزودن مدل‌ها (Integrations)
- انتخاب یا ثبت دستی مدل
- تعیین نوع مدل (Model Type/Capabilities)
- تعیین cost metric
- ذخیره و فعال‌سازی

### گام 3: تخصیص دسترسی
- از طریق collaboratorهای همان provider account
- نقش Manager برای مدیریت حساب/مدل‌ها
- نقش User/Access برای مصرف مدل‌ها بدون دسترسی مدیریتی

### گام 4: مصرف توسط کاربران/اپلیکیشن‌ها
- با PAT یا VAT
- فراخوانی endpointهای OpenAI-compatible
- مشاهده مدل‌های مجاز از طریق `/models`

## کنترل دسترسی (Collaborators)

### نقش‌ها
- `Provider Account Manager`
  - تغییر تنظیمات provider account
  - افزودن/حذف/ویرایش مدل‌ها
  - مدیریت دسترسی دیگران
- `Provider Account User` (یا access)
  - مصرف تمام مدل‌های همان account
  - بدون دسترسی به تغییر تنظیمات

### نکات مهم
- دسترسی معمولا در سطح Provider Account تعریف می‌شود؛ در نتیجه کاربر دارای نقش access، به همه مدل‌های همان account دسترسی می‌گیرد.
- Tenant Admin طبق مستندات دسترسی سراسری دارد.

## احراز هویت مصرف‌کننده
- Gateway با توکن‌های TrueFoundry کار می‌کند:
  - `PAT` برای کاربر انسانی/توسعه
  - `VAT` برای اپلیکیشن/سرویس
- کلید پروایدر در backend نگه‌داری می‌شود و مصرف‌کننده نهایی با آن مستقیم کار نمی‌کند.

## Model Types و رفتار Endpoint
ثبت نوع مدل، رفتار runtime را تعیین می‌کند. مثال‌ها:
- `chat` -> `/chat/completions`
- `embedding` -> `/embeddings`
- `rerank` -> `/rerank`
- `realtime` -> `/live/...` (یا realtime interface)
- `audio transcription` -> `/audio/transcriptions`

نتیجه محصولی:
- اگر قابلیت لازم روی integration فعال نشده باشد، endpoint متناظر خطا می‌دهد (به‌خصوص در APIهای صوتی/ترجمه).
- برای مدل‌های چندوجهی، نوع/قابلیت مدل باید هنگام افزودن integration درست انتخاب شود.

## Display Name و نام مدل در مصرف
- در سناریوهای خاص (خصوصا audio/realtime طبق docs)، شناسه مدل در کد باید با display name ثبت‌شده در TrueFoundry منطبق باشد.
- در خروجی `/models`، مدل‌ها به‌شکل OpenAI-compatible و با `id` قابل مصرف ارائه می‌شوند.

## Cost Management

### Public Cost
- اتکا به کاتالوگ عمومی قیمت‌ها
- auto-populate و به‌روزرسانی‌شونده
- مناسب مدل‌های رایج

### Private Cost
- تعریف تعرفه سفارشی
- مناسب قراردادهای اختصاصی، مدل‌های fine-tuned یا مدل‌های فاقد قیمت عمومی

### محل مصرف داده هزینه
- در metrics/analytics برای breakdown بر اساس user/model/team
- پایه‌ای برای budget limiting و گزارش‌گیری مالی

## APIها و داده‌های کلیدی برای مدیریت

### Apply API (ایجاد/به‌روزرسانی provider account + models)
- endpoint: `/api/svc/v1/apply`
- manifest شامل provider-account + integrations + collaborators

### List Provider Integrations
- endpoint: `/api/svc/v1/provider-integrations?type=model`
- مناسب inventory، audit، و کنترل وضعیت مدل‌ها

### Model Discovery
- endpoint: `/models`
- خروجی فقط مدل‌هایی را نشان می‌دهد که API key فعلی مجاز به مصرف آن‌هاست

## مرزبندی سلسله‌مراتبی (Tenant/Account/Team/Virtual Account)
- Tenant: مرز ایزولیشن اصلی
- Account داخل Tenant: برای جداسازی سازمانی/مالی داخلی
- Team: گروه‌بندی دسترسی
- Virtual Account: هویت غیرانسانی برای اپلیکیشن‌ها (توکن پایدار برای production)

این لایه‌بندی مستقیما روی نحوه واگذاری access به provider accountها اثر می‌گذارد.

## جمع‌بندی اعتبار برداشت شما
برداشت شما عمدتا دقیق است. اصلاح/تکمیل‌ها:
1. «گروه مدل» در مستندات API با `provider account + integrations` مدل‌سازی می‌شود.
2. دسترسی به‌صورت collaborator در سطح provider account اعمال می‌شود؛ بنابراین access معمولا برای همه مدل‌های آن account است.
3. انتخاب model type واقعا تعیین‌کننده endpointهای مجاز است.
4. Cost metric هر مدل قابل تنظیم است (public یا private).
5. Display name در برخی APIها (به‌ویژه audio/realtime) نقش عملیاتی مستقیم در model id مصرفی دارد.

## چک‌لیست پیشنهادی برای Gap Analysis در Shahrah
- [ ] آیا موجودیت `TenantProviderAccount` با `auth_data`, `collaborators`, `integrations` کامل است؟
- [ ] آیا نقش‌های `manager` و `user/access` در سطح account enforce می‌شوند؟
- [ ] آیا model types به‌صورت capability-based ثبت و enforce می‌شوند؟
- [ ] آیا endpointها براساس capability خطای استاندارد می‌دهند؟
- [ ] آیا public/private cost metric برای هر مدل وجود دارد؟
- [ ] آیا `/v1/models` یا معادل، فقط مدل‌های مجاز caller را نشان می‌دهد؟
- [ ] آیا PAT/VAT و مصرف production-ready برای non-human identities کامل است؟
- [ ] آیا mapping display name/model id در مسیرهای audio/realtime روشن و تست‌شده است؟

## Documented Facts
- Provider account ظرف اصلی مدل‌هاست و integrations زیر آن ذخیره می‌شوند.
- Collaborator-based access control با نقش manager/access روی provider account وجود دارد.
- Model discovery از `/models` بر اساس دسترسی API key فیلتر می‌شود.
- Cost tracking دارای public و private mode است.

## Product Recommendations for Shahrah
- واژه‌نامه داخلی را روی `provider account + model integrations + collaborators + capabilities + pricing mode` ثابت نگه دارید.
- enforcement capability-to-endpoint را P0 در نظر بگیرید تا رفتار API قابل پیش‌بینی بماند.
- از روز اول، گزارش هزینه را بر اساس user/team/model/account قابل تفکیک نگه دارید.

## Assumptions
- برخی تفاوت‌های UX بین providerها (مثلا naming در realtime/audio) در API contract قابل نرمال‌سازی است.
- GodarAI برای MVP همان مدل دسترسی دو-نقشی (manager/user) را کافی می‌داند.

## Sources Used
- `.truefoundry-docs/markdown/docs__apply-api-create-models.md`
- `.truefoundry-docs/markdown/docs__using-tfy-apply.md`
- `.truefoundry-docs/markdown/docs__ai-gateway__gateway-access-control.md`
- `.truefoundry-docs/markdown/docs__ai-gateway__cost-tracking.md`
- `.truefoundry-docs/markdown/docs__ai-gateway__model-discovery.md`
- `.truefoundry-docs/markdown/docs__ai-gateway__intro-to-llm-gateway.md`
- `.truefoundry-docs/markdown/docs__ai-gateway__audio-transcription.md`
- `.truefoundry-docs/markdown/docs__ai-gateway__live-api.md`
- `.truefoundry-docs/markdown/docs__api-reference__provider-integrations__list-provider-integrations.md`
- `.truefoundry-docs/markdown/docs__platform__user-team-account-management.md`

## Impact of GodarAI Plans
اسناد GodarAI (خصوصا `plans/readme.md` و `plans/phase0-tenant-provider-account-contract.md`) روی این جمع‌بندی اثر گذاشتند:
- تاکید روی MVP-first و قرارداد شفاف provider account
- هم‌راستایی نقش‌های manager/user با قرارداد فاز صفر
- نیاز به قابلیت gap analysis اجرایی برای فاز بعدی
