# PRD: تجربه Analytics گدارAI

## وضعیت سند

- وضعیت: آماده برای بازبینی محصول و شروع طراحی فنی
- تاریخ: ۱۴۰۵/۰۵/۰۱ (2026-07-23)
- مالک پیشنهادی: Product + Tenant Admin + Control Plane
- نام قابلیت در رابط فارسی: «تحلیل‌ها»
- نام قابلیت در رابط انگلیسی: `Analytics`
- مسیر پیشنهادی: `/analytics`

## خلاصه تصمیم

گدارAI باید در لایه `tenant-admin` یک تجربه تحلیلی چندتب در مسیر مستقل `/analytics` اضافه کند. صفحه فعلی «متریک‌های مدل» در مسیر `/metrics` فعلاً بدون Redirect، تغییر نام یا حذف باقی می‌ماند. تجربه جدید از الگوی Portkey برای تب‌بندی، نوار فیلتر تراشه‌ای، بازه زمانی مستقل و Saved Filters الهام می‌گیرد، اما از زیرساخت ClickHouse، قرارداد Metrics، کش، Logs و Design System موجود گدارAI استفاده می‌کند.

تصمیم‌های اصلی:

1. `/analytics` به‌عنوان صفحه جدید و مستقل در `tenant-admin` اضافه می‌شود؛ `/metrics` فعلاً به همان شکل موجود قابل استفاده می‌ماند.
2. تب‌های MVP عبارت‌اند از: `Overview`، `Subjects`، `Errors`، `Cache` و `Summary`.
3. تب `Feedback` و همه نمودارها یا فیلترهای وابسته به Feedback در دامنه این قابلیت نیستند.
4. فیلترها در همه تب‌ها مشترک‌اند و پس از تأیید هر فیلتر، بدون دکمه کلی «اعمال فیلترها» اجرا می‌شوند.
5. Saved Filter فقط Predicateهای فیلتر را ذخیره می‌کند؛ بازه زمانی، تب فعال، صدک Latency و Group By ذخیره نمی‌شوند.
6. API و ذخیره‌سازی Metrics موجود پایه محاسبات است؛ یک API دامنه‌ای Analytics روی آن ساخته می‌شود تا UI مجبور به ساختن چندین Query سطح پایین نباشد.
7. هزینه ریالی در کنار هزینه دلاری یک الزام P0 است.
8. دسترسی اولیه Tenant-wide باقی می‌ماند و فقط نقش‌ها/مجوزهای مجاز می‌توانند داده تحلیلی یا Saved Filter را ببینند و تغییر دهند.
9. مدل Subject گدارAI چهار نوع دارد: کاربر حقیقی، کاربر مجازی، تیم و واحد سازمانی.
10. هر درخواست دقیقاً یک Actor اصلی—کاربر حقیقی یا کاربر مجازی—و حداکثر یک تیم و یک واحد سازمانی resolved دارد تا شمارش مضاعف رخ ندهد.

## زمینه و مسئله

صفحه فعلی `/metrics` داده‌های ارزشمندی مانند تعداد درخواست، توکن، هزینه، Latency، خطا و اثربخشی کش را دارد و فعلاً برای کاربران فعلی و سناریوهای فنی متراکم حفظ می‌شود. بااین‌حال تجربه کاربری آن یک صفحه طولانی با فیلترهای فرم‌مانند، دکمه «اعمال فیلترها» و تعداد زیادی کارت و نمودار هم‌زمان است. صفحه جدید `/analytics` باید در کنار آن، برای پاسخ‌دادن سریع به سؤال‌های عملیاتی زیر بهینه شود:

- «چرا نرخ خطا بالا رفته؟»
- «کدام کاربر حقیقی، کاربر مجازی، تیم یا واحد سازمانی بیشترین مصرف را دارد؟»
- «کش واقعاً چقدر سرعت و هزینه را بهبود داده است؟»
- «این ترکیب فیلتر پرتکرار را چطور دوباره استفاده یا با همکاران به اشتراک بگذارم؟»
- «چطور از یک نمای کلی به Logs یا Traceهای مرتبط بروم؟»

الگوی Portkey این مسئله را با پنج ایده خوب حل می‌کند:

- جداسازی سؤال‌های تحلیلی در تب‌های متمرکز؛
- نوار فیلتر مشترک و کم‌ارتفاع؛
- نمایش هر فیلتر به‌صورت Chip قابل ویرایش و حذف؛
- بازه زمانی مستقل از Saved Filter؛
- Saved Filter با جریان‌های روشن Create، Apply، Update، Rename، Edit و Delete.

هدف این PRD انتقال همین منطق طراحی به گدارAI است، بدون کپی‌کردن معماری یا محدودیت‌های Portkey.

## ورودی‌های بررسی‌شده

### مشاهده مستقیم UI Portkey

صفحه Analytics در نشست واقعی کاربر بررسی شد. موارد مشاهده‌شده:

- تب‌ها: `Overview`، `Users`، `Errors`، `Cache`، `Feedback` و `Summary`.
- نوار مشترک شامل Refresh، فیلترها، Saved Filter و Time Range.
- گزینه‌های Time Range: ۱۵ دقیقه، ۱ ساعت، ۲۴ ساعت، دیروز، ۳، ۷، ۱۴، ۳۰، ۹۰ روز و Custom.
- انتخابگر Custom شامل تقویم، زمان شروع/پایان و دکمه Update است.
- انتخاب Latency بین `Mean`، `p50`، `p90` و `p99`.
- فیلترهای مشاهده‌شده در حساب فعلی: Cost، Tokens، Cache، Status، Avg. Weighted Feedback، Trace ID و Saved Filters.
- Cache و Status از Multi-select با Update استفاده می‌کنند.
- Cost و Tokens از Min/Max و Slider دوطرفه استفاده می‌کنند.
- Trace ID از ورودی چندمقداری، Update و Reset استفاده می‌کند.
- هنگام ویرایش Saved Filter، Predicateها به Chipهای عادی تبدیل و دکمه‌های Update/Cancel نمایش داده می‌شوند.
- هنگام تغییر Predicateهای یک Saved Filter فعال، دکمه `Update` ظاهر می‌شود.
- عملیات Saved Filter فعال: Rename، Edit و Delete.
- Rename و Create به‌صورت Inline در همان نوار فیلتر انجام می‌شوند.
- Saved Filter در URL با شناسه پایدار Deep-link می‌شود.
- Summary امکان Group By بر اساس AI Service، Model، Status Code، Workspace، API Key، Config، Provider و Prompt دارد.

### مستندات Portkey

- `.portkey-docs/markdown/product/docs__product__observability.md`
- `.portkey-docs/markdown/product/docs__product__observability__analytics.md`
- `.portkey-docs/markdown/product/docs__product__observability__filters.md`
- `.portkey-docs/markdown/product/docs__product__administration__configure-analytics-access-permissions.md`

### اسناد و کد گدارAI

- `plans/README.md`
- `plans/observability-platform-implementation-plan.md`
- `plans/model-metrics-clickhouse-implementation-plan.md`
- `plans/model-metrics-contract-v2.md`
- `plans/model-metrics-runbook.md`
- `plans/tenant-organization-team-design.md`
- `plans/organization-team-virtualkey-design.md`
- `plans/phase2-tenant-admin-iam-rbac.md`
- `plans/database-conceptual-model.md`
- `apps/tenant-admin/app/components/model-metrics-dashboard.tsx`
- `apps/tenant-admin/app/components/model-metrics-utils.ts`
- `apps/tenant-admin/app/metrics/page.tsx`
- `apps/tenant-admin/app/components/logs-command-center.tsx`
- `packages/contracts/platform/metrics.go`
- `packages/contracts/platform/logs_v2.go`
- `services/control-plane/internal/admin/metrics_handlers.go`
- `services/control-plane/internal/storage/metrics_repo_clickhouse.go`
- `services/control-plane/internal/storage/log_saved_views_repo.go`
- `infra/k8s/local/migrations/036_log_command_center.sql`
- `docs/design-system/DESIGN.md`
- `docs/design-system/COMPONENT_USAGE.md`
- اجزای موجود `apps/tenant-admin/components/ui` و `apps/tenant-admin/components/app-ui`

### اثر اسناد گدارAI بر تصمیم

`plans/README.md` و قراردادهای Metrics توصیه را در این موارد تغییر دادند:

- Cost باید IRR و USD را پوشش دهد، نه فقط USD.
- ClickHouse و Metrics event موجود باید حفظ و توسعه داده شوند.
- Tenant isolation و دسترسی Tenant-wide الزام امنیتی‌اند.
- Retention و Query SLO باید با زیرساخت ۹۰روزه فعلی هماهنگ باشند.
- Cache savings موجود یک مزیت گدارAI است و باید در تب Cache برجسته‌تر از Portkey نمایش داده شود.
- Saved Views موجود در Logs باید به‌عنوان الگوی پیاده‌سازی استفاده شود، ولی در MVP با Analytics یک جدول واحد نشود.

### Portkey Source

هیچ Source Code از Portkey بررسی یا کپی نشده است. مستندات و رفتار مشاهده‌شده UI فقط به‌عنوان Evidence محصولی استفاده شده‌اند.

## تفکیک Fact، Recommendation و Assumption

### Factهای مستند یا مشاهده‌شده

- Portkey Analytics تب‌بندی و نوار فیلتر مشترک دارد.
- Saved Filter در Portkey میان اعضای سازمان قابل استفاده و مدیریت است.
- Time Range در UI مشاهده‌شده جدا از محتوای Saved Filter باقی می‌ماند.
- گدارAI اکنون Metrics Query API، ClickHouse raw events، aggregate table، Facet مدل/Provider و Cache savings دارد.
- UI فعلی گدارAI فقط Model، Provider، Request ID، Time Range و View By را مستقیماً عرضه می‌کند.
- Backend فعلی فیلترها و Group Byهای بیشتری از UI پشتیبانی می‌کند، اما مسیر Aggregate برای همه ابعاد بهینه نیست.
- Logs گدارAI از قبل CRUD برای Saved Views و scopeهای `tenant` و `private` دارد.
- مدل دامنه گدارAI میان User انسانی، Virtual User سرویس‌محور، Team و Organization Unit تفکیک می‌کند.
- کاربر حقیقی می‌تواند عضو چند تیم یا واحد سازمانی باشد؛ تیم نیز ممکن است مستقیم زیر Tenant یا زیر یک واحد سازمانی قرار گیرد.
- PAT به کاربر حقیقی و VAT به کاربر مجازی تعلق دارد؛ بنابراین Credential context منبع قابل اتکای Actor attribution هر درخواست است.

### Recommendationهای این PRD

- Analytics به‌صورت Route مستقل کنار UI فعلی Metrics اضافه شود.
- API دامنه‌ای Analytics روی Metrics Query ساخته شود.
- Saved Filters مختص Analytics در MVP پیاده‌سازی شود.
- Feedback به‌طور کامل حذف شود.
- مجوزهای اختصاصی Analytics اضافه شوند.
- IRR، Trace/Fallback و Facetهای لازم به قرارداد Metric افزوده شوند.
- تب Portkey-like `Users` در گدارAI به `Subjects` تطبیق داده شود و چهار Subject بومی محصول را پوشش دهد.
- انتساب Subject در لحظه درخواست Snapshot شود تا گزارش تاریخی به عضویت یا نام فعلی موجودیت‌ها وابسته نباشد.

### Assumptionها

- نام نهایی فارسی قابلیت «تحلیل‌ها» است.
- Retention قابل انتخاب در MVP حداکثر ۹۰ روز است.
- داده Analytics در MVP Tenant-wide است و Team-scoped authorization به فاز بعد منتقل می‌شود.
- Saved Filterهای Tenant-shared برای همکاری مهم‌تر از Saved Filterهای خصوصی‌اند؛ هر دو در مدل داده پیش‌بینی می‌شوند.
- `Rescued Request` به درخواستی گفته می‌شود که حداقل یک Fallback داشته و نتیجه نهایی موفق بوده است.
- در MVP هر درخواست یک `resolved_team_id` و یک `resolved_org_unit_id` اختیاری دارد؛ عضویت‌های متعدد فعلی User به‌طور هم‌زمان به یک Request نسبت داده نمی‌شوند.

## اهداف

1. کاربر در کمتر از ۳۰ ثانیه وضعیت هزینه، مصرف، Latency، خطا، Subjectها و کش را بفهمد.
2. هر تب فقط به یک دسته سؤال عملیاتی پاسخ دهد.
3. یک ترکیب فیلتر پرتکرار با حداکثر دو تعامل اعمال شود.
4. Saved Filter قابل Deep-link، Rename، Edit، Update و Delete باشد.
5. فیلترها، Time Range و تب فعال هنگام جابه‌جایی میان تب‌ها حفظ شوند.
6. هر نمودار یا Breakdown مهم بتواند کاربر را با همان Context به Logs ببرد.
7. UI در فارسی/RTL، انگلیسی/LTR، دسکتاپ و موبایل رفتار قابل پیش‌بینی داشته باشد.
8. از Data و Query engine فعلی Metrics استفاده شود و مسیر Gateway در صورت خرابی Analytics مختل نشود.
9. مصرف چهار Subject گدارAI بدون شمارش مضاعف و با قابلیت Filter، Ranking و Group By قابل تحلیل باشد.
10. صفحه `/metrics` در طول ساخت و عرضه MVP بدون Regression باقی بماند.

## Non-Goals

- تب یا API مربوط به Feedback.
- جمع‌آوری Feedback یا Weighted Feedback.
- جایگزینی Grafana، Prometheus یا داشبوردهای عملیاتی زیرساخت.
- Export گزارش، Report Builder یا Scheduled Report در MVP.
- ویرایش آزاد Query/SQL توسط کاربر.
- Real-time streaming chart؛ Refresh دستی و Query معمولی کافی است.
- دسترسی مستقیم UI به ClickHouse.
- اشتراک Saved Filter بین Logs و Analytics در MVP.
- ذخیره Tab، Time Range، Latency percentile یا Summary Group By داخل Saved Filter.
- تحلیل فعالیت اعضا داخل پنل `tenant-admin`؛ Subject analytics فقط به درخواست‌های LLM منتسب‌شده مربوط است.
- Redirect، حذف یا تغییر نام Route فعلی `/metrics` در این فاز.
- ادغام کامل کد UI دو صفحه `/metrics` و `/analytics`؛ اشتراک در این فاز فقط در قرارداد، Query و Componentهای واقعاً reusable انجام می‌شود.

## کاربران و Use Caseها

| کاربر | Use case | نشانه موفقیت |
|---|---|---|
| Tenant Admin | دیدن سلامت مصرف و هزینه کل Tenant | Overview در کمتر از ۲.۵ ثانیه قابل استفاده است |
| Team Admin | یافتن مدل، Provider یا تیم مسئله‌دار | با Filter و Summary علت در کمتر از ۳ دقیقه محدود می‌شود |
| FinOps / Billing Operator | مقایسه هزینه IRR/USD و صرفه‌جویی کش | عددهای Analytics با Spend/Wallet در محدوده خطای تعریف‌شده منطبق‌اند |
| Developer / SRE | تحلیل خطا و Latency و رفتن به Logs | Context فیلتر و زمان در Deep-link حفظ می‌شود |
| Product Owner | دیدن Subjectهای فعال و مصرف هر Subject | Active Subjects و Top Subjects قابل اتکا هستند |
| مدیر واحد سازمانی | مقایسه مصرف تیم‌ها و Actorهای منتسب به واحد | Breakdown واحد سازمانی بدون شمارش مضاعف است |

## نقشه قابلیت Portkey به گدارAI

| قابلیت | مسئله‌ای که حل می‌کند | تصمیم | انطباق گدارAI | دلیل |
|---|---|---|---|---|
| Overview | درک سریع سلامت کلی | Include now | IRR + USD و Cache savings گدارAI | ارزش روز اول |
| Users | تحلیل مصرف‌کننده درخواست | Adapt now | تب `Subjects` با چهار نوع بومی گدارAI و بدون User Feedback | Portkey End-user با مدل IAM گدارAI یکسان نیست |
| Errors | تشخیص نرخ/نوع خطا | Include now | Error taxonomy گدارAI و Deep-link به Logs | عملیات روزمره |
| Cache | سنجش سرعت و صرفه‌جویی | Include now | Simple/Smart و avoided cost فعلی | مزیت محصولی گدارAI |
| Summary | مقایسه ابعاد | Include now | چهار بُعد کاربر حقیقی، کاربر مجازی، تیم و واحد سازمانی | تصمیم‌گیری عملیاتی |
| Feedback | کیفیت پاسخ | Exclude | هیچ تب، Widget یا Filter مرتبط ساخته نمی‌شود | درخواست صریح محصول |
| Saved Filters | تکرار و همکاری | Include now | Tenant-shared با RBAC و optimistic version | UX کلیدی |
| Custom Time Range | تحلیل بازه دقیق | Include now | Local input، UTC query | نیاز عملیاتی |
| Analytics access override | کنترل دسترسی ریزدانه | Defer | مجوزهای اختصاصی ابتدا؛ override سطح Workspace بعداً | کاهش سطح عملیات MVP |
| Cross-surface saved filters | یک فیلتر برای Logs و Analytics | Defer | Schemaهای دو سطح متفاوت‌اند | جلوگیری از Migration پرریسک |
| Export/Reports | تحلیل خارج از محصول | Defer | از Logs export استفاده شود | خارج از MVP |

## معماری اطلاعات و هم‌زیستی Routeها

### Navigation

- آیتم فعلی «متریک مدل‌ها» با Route `/metrics` بدون تغییر باقی بماند.
- آیتم جدید «تحلیل‌ها» با Route `/analytics` در گروه Observability و کنار Metrics، Logs و Traces اضافه شود.
- هیچ Redirect میان `/metrics` و `/analytics` ساخته نشود.
- لینک‌های موجود «Open in Metrics» شکسته یا تغییر داده نشوند؛ در محل‌هایی که تجربه تحلیلی مناسب است، Action مستقل «Open in Analytics» اضافه شود.
- دو صفحه می‌توانند قرارداد و Query engine مشترک داشته باشند، اما state، URL و تجربه کاربری مستقل دارند.
- تصمیم درباره همگرایی یا بازنشستگی احتمالی `/metrics` فقط پس از Telemetry استفاده و یک PRD جداگانه گرفته می‌شود.

مرزبندی پیشنهادی:

| Route | نقش محصولی | تجربه غالب |
|---|---|---|
| `/metrics` | نمای فنی و متراکم متریک‌های مدل | صفحه فعلی با Query و نمودارهای موجود |
| `/analytics` | تحلیل سؤال‌محور مصرف و عملیات | تب‌های متمرکز، فیلتر تراشه‌ای، Saved Filters و Drill-down |

### URL State

Query Parameterهای پیشنهادی:

```text
/analytics
  ?tab=overview
  &range=last24h
  &from=2026-07-22T08:00:00.000Z
  &to=2026-07-23T08:00:00.000Z
  &saved_filter_id=<uuid>
  &filters=<versioned-base64url-json>
  &latency=p50
  &subject_type=realUser
  &group_by=modelName
```

قواعد:

- فقط stateهای Commit‌شده وارد URL شوند.
- `saved_filter_id` شناسه اصلی Deep-link است.
- اگر Saved Filter یافت نشد یا دسترسی وجود نداشت، صفحه باز شود، فیلتر اعمال نشود و Error قابل اقدام نشان داده شود.
- `filters` برای Deep-link فیلترهای موقت است و باید Size limit داشته باشد؛ در صورت عبور از حد، UI کاربر را به Save Filter هدایت کند.
- Time Range در URL است ولی جزو Saved Filter نیست.
- تغییر Tab نباید Predicateهای فیلتر را پاک کند.
- `subject_type` فقط انتخاب نمای تب Subjects است و جزو Saved Filter نیست؛ Predicateهای Subject همچنان در `filters` ذخیره می‌شوند.

## مدل Subject و قواعد انتساب

در این سند Subject یک اصطلاح تحلیلی مشترک برای Actor درخواست و Scope سازمانی آن است. انواع canonical عبارت‌اند از:

| Subject | کلید API | نقش در انتساب | تعریف |
|---|---|---|---|
| کاربر حقیقی | `realUser` | Actor | عضو انسانی Tenant که درخواست با PAT یا Credential منتسب به او اجرا شده است |
| کاربر مجازی | `virtualUser` | Actor | هویت غیرانسانی برای سرویس، Agent یا Automation که درخواست با VAT منتسب به او اجرا شده است |
| تیم | `team` | Scope | گروه کاری resolved درخواست؛ تیم می‌تواند مستقیم زیر Tenant یا زیر یک واحد سازمانی باشد |
| واحد سازمانی | `orgUnit` | Scope | واحد، دپارتمان یا Cost center اختیاری Tenant که Scope سازمانی resolved درخواست را مشخص می‌کند |

قواعد P0:

1. هر Request دقیقاً یک Actor اصلی از نوع `realUser` یا `virtualUser` دارد؛ اگر انتساب Actor در داده قدیمی ممکن نباشد، در Bucket «نامشخص» قرار می‌گیرد.
2. هر Request حداکثر یک `resolved_team_id` و حداکثر یک `resolved_org_unit_id` دارد.
3. User یا Virtual User ممکن است در مدل IAM به چند تیم متصل باشد، اما Analytics همه عضویت‌های جاری را به یک Request الصاق نمی‌کند. تیم و واحد سازمانی از Scope همان Credential/Request resolve می‌شوند.
4. برای جلوگیری از تغییر تاریخچه، ID و Label نمایشی Subject در لحظه درخواست داخل Metric event Snapshot می‌شوند.
5. حذف، غیرفعال‌سازی، تغییر نام یا جابه‌جایی Subject نباید داده تاریخی را حذف یا به Scope جدید بازنویسی کند.
6. هر Request در گزارش یک نوع Subject فقط یک بار شمرده می‌شود. جمع ستون‌های چهار نوع Subject معادل Requests کل نیست، چون هر Request هم Actor و هم Scope دارد.
7. Org Unit aggregation شامل Requestهایی است که `resolved_org_unit_id` یکسان دارند؛ Requestهای Tenant-level بدون واحد در Bucket «بدون واحد سازمانی» قرار می‌گیرند.
8. Raw PAT، VAT، Secret یا Token hash در Event، Facet، URL یا جدول نمایش ذخیره نمی‌شود.

## ساختار صفحه

ترتیب ثابت صفحه:

1. Shell Page Heading با عنوان، توضیح و Help link؛
2. Tab list؛
3. نوار Sticky ابزار شامل Refresh، Filter chips، Search/Add filter، Saved Filter actions و Time Range؛
4. محتوای تب؛
5. خطا یا Warning سطح Widget، نه فقط خطای تمام‌صفحه.

اصول:

- ارتفاع نوار ابزار تا جای ممکن ثابت بماند و در چند خط Wrap کنترل‌شده داشته باشد.
- در Desktop تب‌ها و نوار ابزار Sticky زیر Shell header باشند.
- در Mobile تب‌ها Horizontal scroll و Filter editor در Popover/Sheet سازگار باز شود.
- مقادیر فنی مانند UUID، Model name، Trace ID و Request ID همیشه LTR و Monospace باشند.
- از Semantic tokenها و اجزای موجود Design System استفاده شود.

## الزامات تب‌ها

### P0 — Overview

Overview نمای شروع و پاسخ به سؤال «الان چه خبر است؟» است.

Widgetها:

1. Cost
   - عدد کل با Currency اصلی Tenant؛ برای بازار ایران IRR.
   - مقدار USD به‌عنوان مقدار ثانویه یا Tooltip دقیق.
   - Time series هزینه.
2. Tokens Used
   - Total، Input و Output.
   - نمودار Stacked برای Input/Output.
3. Latency
   - مقدار اصلی انتخاب‌شدنی: `Mean`، `p50`، `p90`، `p99`.
   - انتخاب فقط همین Widget را تغییر دهد و در URL ذخیره شود.
4. Requests
   - Total request count و Time series.
5. Active Actors
   - تعداد Actorهای یکتای فعال، با Breakdown کاربر حقیقی و کاربر مجازی.
6. Top Subjects
   - Selector نوع Subject: کاربر حقیقی، کاربر مجازی، تیم یا واحد سازمانی.
   - پیش‌فرض نوع Subject: کاربر حقیقی؛ مرتب‌سازی پیش‌فرض بر اساس Requests.
   - امکان تغییر مرتب‌سازی به Cost در P1.
7. Top Models
   - Model، Requests، Cost و Error rate.

تعریف Empty state:

- اگر هیچ Request در بازه نیست، Cardها صفر یا `—` معنادار نشان دهند و یک Empty State مشترک توضیح دهد که داده‌ای مطابق Filter وجود ندارد.
- نبود داده نباید به شکل Chart خالی بدون توضیح رها شود.

### P0 — Subjects

این تب اقتباس گدارAI از تب `Users` در Portkey است. انتخابگر نوع Subject در بالای تب چهار گزینه دارد: «کاربر حقیقی»، «کاربر مجازی»، «تیم» و «واحد سازمانی». انتخاب در URL با `subject_type` حفظ می‌شود.

Widgetها:

1. Active Subjects؛
2. Top Subjects؛
3. Requests per Subject؛
4. Cost per Subject؛
5. Subject trend؛
6. Breakdown موفقیت، خطا و Cache برای Subject منتخب.

تعاریف:

- `active_subject` یعنی Subjectی که در Time Range و Predicateهای فعلی حداقل یک Request منتسب دارد.
- شمارش Unique بر اساس ID پایدار همان نوع Subject انجام می‌شود؛ Label فقط برای نمایش است.
- رویداد Legacy بدون ID قابل اتکا در Bucket «نامشخص» نمایش داده می‌شود و در Active Subjects شمرده نمی‌شود.
- کاربر حقیقی و کاربر مجازی دو Actor مستقل‌اند و با Email یا نام مشابه Merge نمی‌شوند.
- تیم و واحد سازمانی از Snapshot زمان درخواست خوانده می‌شوند، نه از عضویت فعلی IAM.
- کلیک روی هر ردیف، Filter همان Subject را به نوار فیلتر اضافه می‌کند و Drill-down به Logs را نیز ارائه می‌دهد.

خارج از دامنه:

- `User Feedback` نمایش داده نمی‌شود.
- تحلیل Login، Session یا فعالیت کاربر داخل پنل نمایش داده نمی‌شود.

### P0 — Errors

Widgetها:

1. Error Rate؛
2. Error Count؛
3. Error Types؛
4. Rescued Requests؛
5. Error trend.

تعاریف:

```text
error_count = count(status_code >= 400)
error_rate = error_count / request_count
rescued_request = final_status_code < 400 AND fallback_count > 0
```

رفتار Drill-down:

- کلیک روی Error type یا Status code، Logs را با همان Time Range و Filterهای قابل تبدیل باز کند.
- اگر یک Predicate در Logs معادل ندارد، حذف آن باید در یک Warning کوچک اعلام شود.

### P0 — Cache

Widgetها:

1. Cache Hits؛
2. Cache Hit Rate؛
3. Cache Speedup؛
4. Cache Savings؛
5. Breakdown بر اساس `simple` و `smart`؛
6. Breakdown تطبیق `exact` و `semantic`.

تعاریف:

```text
cache_hit_rate = hits / (hits + misses)
```

`none`، `bypass` و `error` از مخرج حذف می‌شوند.

```text
cost_without_cache = actual_cost + avoided_cost
savings_rate = avoided_cost / cost_without_cache
```

Cache Speedup:

```text
speedup_factor = average_miss_latency / average_hit_latency
```

- فقط وقتی هر دو گروه Sample کافی دارند نمایش داده شود.
- در نبود Sample کافی `—` نمایش داده شود، نه `0x`.
- UI باید Avg Cache Latency و روش محاسبه را در Tooltip توضیح دهد.

### P0 — Summary

Summary یک Table قابل Sort و Paginate است و یک Group By انتخابی دارد.

Group Byهای P0:

- Request Type؛
- Model؛
- Virtual Model؛
- Provider؛
- Status Code؛
- Error Type؛
- کاربر حقیقی؛
- کاربر مجازی؛
- تیم؛
- واحد سازمانی.

Group Byهای P1:

- Access Key alias/type؛
- Metadata key؛
- Provider Model.

Group Byهای Deferred:

- Config، تا زمانی که Config ID به‌طور پایدار در Metric event ثبت شود؛
- Prompt، تا زمانی که Prompt Registry analytics قرارداد روشن داشته باشد.

ستون‌ها:

- Group value؛
- Requests؛
- Success rate؛
- Input/Output/Total tokens؛
- Cost IRR؛
- Cost USD؛
- Mean/p50/p90/p99 Latency؛
- Cache hit rate.

قواعد:

- Sort پیش‌فرض بر اساس Requests نزولی.
- Pagination server-side.
- مقدارهای خالی با «نامشخص» گروه شوند.
- Technical values با `ModelNameValue` یا نمایش LTR مناسب ارائه شوند.

## نوار فیلتر

### مدل منطقی

- Predicateهای مربوط به Fieldهای متفاوت با `AND` ترکیب می‌شوند.
- چند مقدار در یک Field چندانتخابی با `IN` یا منطق `OR` ترکیب می‌شوند.
- Rangeها Inclusive هستند.
- یک Field در MVP فقط یک Chip فعال دارد؛ ویرایش Chip جایگزین مقدار همان Field می‌شود.
- حداکثر ۲۰ Chip و حداکثر ۵۰ مقدار در هر Multi-select.
- فیلترهای چهار Subject با سایر Fieldها و با یکدیگر `AND` می‌شوند؛ مثلاً کاربر حقیقی A در تیم B فقط Requestهایی را برمی‌گرداند که هر دو انتساب را در Snapshot خود دارند.
- انتخاب `actorType=real_user` همراه با `virtualUserID` ناسازگار است و Client و Server باید آن را با خطای Validation روشن رد کنند.

نمونه:

```text
provider IN (openai, anthropic)
AND status_code IN (429, 500)
AND total_tokens BETWEEN 1000 AND 5000
```

### کاتالوگ فیلتر

| فیلتر | Priority | نوع UI | Backend mapping | وضعیت فعلی |
|---|---|---|---|---|
| Model | P0 | searchable multi-select | `modelName IN` | Backend دارد؛ UI فعلی Single |
| Virtual Model | P0 | searchable multi-select | `virtualModelName IN` | Backend دارد |
| Provider | P0 | searchable multi-select | `providerName IN` | Backend دارد؛ UI فعلی Single |
| Provider Model | P1 | searchable multi-select | `providerModelName IN` | Backend دارد |
| Request Type | P0 | multi-select | `requestType IN` | Backend دارد |
| Status Code | P0 | dynamic multi-select | `statusCode IN` | Backend دارد |
| Error Type | P0 | dynamic multi-select | `errorType IN` | Backend دارد |
| نوع Actor | P0 | multi-select | `actorType IN (real_user, virtual_user)` | جدید |
| کاربر حقیقی | P0 | searchable multi-select | `realUserID IN` | فیلدهای User فعلی نیاز به canonical mapping دارند |
| کاربر مجازی | P0 | searchable multi-select | `virtualUserID IN` | Virtual Account فعلی نیاز به canonical mapping دارد |
| تیم | P0 | searchable multi-select | `resolvedTeamID IN` | Team array فعلی باید به Scope یکتای Request تبدیل شود |
| واحد سازمانی | P0 | searchable multi-select | `resolvedOrgUnitID IN` | نیاز به event/column و Facet جدید |
| Total Tokens | P0 | Min/Max + range slider | `input_tokens + output_tokens BETWEEN` | نیاز به mapping جدید |
| Cost USD | P0 | Min/Max + range slider | `costInUSD BETWEEN` | نیاز به mapping جدید |
| Cost IRR | P0 | Min/Max + range slider | `costInIRR BETWEEN` | نیاز به event/column جدید |
| Latency | P1 | Min/Max + range slider | `latencyMs BETWEEN` | Backend دارد |
| Cache Status | P0 | multi-select | `cacheStatus IN` | Backend دارد |
| Cache Mode | P0 | multi-select | `cacheMode IN` | Backend دارد |
| Cache Match | P1 | multi-select | `cacheMatch IN` | Backend دارد |
| Request ID | P0 | exact multi-value input | `requestID IN` | Backend دارد |
| Trace ID | P1 | exact multi-value input | `traceID IN` | نیاز به event/column جدید |
| Metadata | P1 | key + operator + value | `metadataKey` | Raw query دارد؛ Facet و whitelist لازم است |
| Saved Filters | P0 | searchable list | Saved filter CRUD | جدید |
| Feedback | Excluded | — | — | ساخته نمی‌شود |

### رفتار Search/Add Filter

1. Focus روی Search همه فیلترهای مجاز را نشان می‌دهد.
2. تایپ، نام فیلتر و گزینه‌های مرتبط را جست‌وجو می‌کند.
3. انتخاب نوع فیلتر Editor مناسب را باز می‌کند.
4. دکمه `Update` داخل Editor فقط وقتی مقدار معتبر انتخاب شده فعال می‌شود.
5. پس از Update، Editor بسته، Chip اضافه و Query اجرا می‌شود.
6. Text filterهای Exact فقط با Enter/Update Commit می‌شوند.
7. Numeric range هنگام Drag Query نمی‌زند؛ فقط `onValueCommit`.
8. حذف Chip Query را بلافاصله دوباره اجرا می‌کند.
9. Chip با Enter/Click قابل ویرایش و با Delete/Backspace یا دکمه حذف قابل پاک‌کردن است.
10. Focus بعد از حذف یا بستن Editor به جای منطقی برگردد.

### بازه زمانی

Presetهای P0:

- ۱۵ دقیقه اخیر؛
- ۱ ساعت اخیر؛
- ۲۴ ساعت اخیر؛
- دیروز؛
- ۳ روز اخیر؛
- ۷ روز اخیر؛
- ۱۴ روز اخیر؛
- ۳۰ روز اخیر؛
- ۹۰ روز اخیر؛
- سفارشی.

قواعد:

- ورودی Custom در Timezone مرورگر/کاربر نمایش داده می‌شود.
- API همیشه RFC3339/UTC دریافت می‌کند.
- `from < to`.
- آینده قابل انتخاب نیست.
- طول بازه MVP از Retention مجاز Tenant بیشتر نمی‌شود.
- تغییر Time Range، Saved Filter را Dirty نمی‌کند.
- Interval خودکار:
  - تا ۱ ساعت: ۱ دقیقه؛
  - تا ۲۴ ساعت: ۵ دقیقه؛
  - تا ۷ روز: ۱ ساعت؛
  - بیش از ۷ روز: ۱ روز.

### Refresh

- Refresh همه Widgetهای تب فعال و Facetهای وابسته را تازه می‌کند.
- Filter، Tab، Time Range، Group By و Latency metric حفظ می‌شوند.
- هنگام Refresh از `RefreshButton` و Spinner موجود استفاده شود.
- تکرار کلیک هنگام Query جاری غیرفعال باشد.

## Saved Filters

### دامنه ذخیره

Saved Filter شامل این موارد است:

- نسخه Schema؛
- Predicateهای Commit‌شده؛
- نام؛
- Scope؛
- Creator و Audit metadata.

Saved Filter شامل این موارد نیست:

- Time Range؛
- Tab؛
- Subject type انتخاب‌شده در تب Subjects؛
- Latency metric؛
- Summary Group By؛
- Sort/Pagination؛
- Currency display preference.

این جداسازی باعث می‌شود یک Filter مانند «خطاهای Provider X» در بازه‌های زمانی و تب‌های مختلف قابل استفاده باشد.

### State machine

#### حالت A — بدون Predicate

- فقط Search/Add Filter و Time Range دیده می‌شوند.
- Save Filter وجود ندارد.

#### حالت B — Predicate موقت

- Chipها نمایش داده می‌شوند.
- دکمه `Save Filter` ظاهر می‌شود.

#### حالت C — Saved Filter فعال و بدون تغییر

- Chip نام Saved Filter نمایش داده می‌شود.
- منوی `Rename / Edit / Delete` در دسترس است.

#### حالت D — Saved Filter فعال و Dirty

- Predicate جدید/تغییریافته کنار نام Saved Filter دیده می‌شود.
- دکمه `Update` ظاهر می‌شود.
- تغییر Time Range، Tab، Subject type، Latency metric یا Group By باعث Dirty شدن نمی‌شود.

#### حالت E — Edit

- Predicateهای Saved Filter به Chipهای قابل ویرایش تبدیل می‌شوند.
- `Update` و `Cancel` نمایش داده می‌شوند.
- Cancel نسخه Persist‌شده را برمی‌گرداند.

#### حالت F — Rename/Create

- Form نام‌گذاری در همان Toolbar باز می‌شود.
- نام خالی یا فقط Whitespace معتبر نیست.
- `Save/Update` فقط در حالت معتبر فعال است.
- Escape یا Cancel بدون Mutation خارج می‌شود.

### عملیات

#### Create

- نام ۱ تا ۸۰ کاراکتر.
- نام در Tenant و Scope به‌صورت Case-insensitive یکتا است.
- سقف پیشنهادی: ۵۰ Saved Filter برای هر Tenant در MVP.
- Scope پیش‌فرض برای نقش‌های مدیریتی: `tenant`.

#### Apply

- از فهرست searchable Saved Filters.
- Predicateهای فعلی را جایگزین می‌کند.
- Time Range و Tab را حفظ می‌کند.
- URL با `saved_filter_id` به‌روزرسانی می‌شود.

#### Update

- فقط با مجوز مدیریت.
- Optimistic concurrency با `version` یا `updated_at`.
- Conflict با پاسخ `409` و پیام «این فیلتر توسط کاربر دیگری تغییر کرده است».

#### Rename

- Inline و بدون تغییر Predicateها.
- Duplicate name با `409`.

#### Delete

- با `ConfirmDialog`.
- پس از حذف، Predicateهای Applied می‌توانند به‌صورت موقت در صفحه باقی بمانند، اما `saved_filter_id` از URL حذف شود.
- حذف باید Audit شود.

### دسترسی

- کاربر دارای `analytics:read` می‌تواند Saved Filterهای Tenant-shared و Private متعلق به خودش را ببیند و Apply کند.
- `analytics:saved_filters_manage` برای Create/Update/Rename/Delete لازم است.
- Private filter فقط توسط Creator یا Tenant Admin قابل تغییر/حذف است.
- Tenant-shared filter فقط توسط نقش دارای مجوز مدیریت قابل تغییر/حذف است.
- رفتار «همه اعضا می‌توانند حذف کنند» از Portkey عیناً کپی نمی‌شود.

### مدل داده پیشنهادی

در MVP جدول جداگانه ساخته شود:

```sql
CREATE TABLE tenant_analytics_saved_filters (
  saved_filter_id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  schema_version INTEGER NOT NULL DEFAULT 1,
  filters_json JSONB NOT NULL,
  shared_scope TEXT NOT NULL CHECK (shared_scope IN ('tenant', 'private')),
  created_by_user_id UUID REFERENCES users(user_id) ON DELETE SET NULL,
  version INTEGER NOT NULL DEFAULT 1,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

Indexها:

- Unique case-insensitive روی `(tenant_id, shared_scope, lower(name))`.
- List index روی `(tenant_id, updated_at DESC)`.
- Private visibility index روی `(tenant_id, created_by_user_id, updated_at DESC)`.

دلیل جدول جدا:

- `tenant_log_saved_views` علاوه بر Filters، `visible_columns` و Schema متفاوت دارد.
- Migration به مدل عمومی در MVP ریسک غیرضروری ایجاد می‌کند.
- Repository، Handler و تست‌های Logs باید به‌عنوان Pattern بازاستفاده شوند.

## API پیشنهادی

### 1. Analytics Query

```http
POST /api/v1/analytics/query
```

Request:

```json
{
  "tab": "overview",
  "startTs": "2026-07-22T08:00:00.000Z",
  "endTs": "2026-07-23T08:00:00.000Z",
  "filters": [
    {
      "fieldName": "providerName",
      "operator": "IN",
      "value": ["openai", "anthropic"]
    },
    {
      "fieldName": "statusCode",
      "operator": "IN",
      "value": [429, 500]
    }
  ],
  "intervalInSeconds": 300,
  "options": {
    "latencyMetric": "p50",
    "subjectType": "realUser",
    "groupBy": "modelName",
    "sortBy": "requests",
    "sortOrder": "desc",
    "limit": 25,
    "cursor": null
  }
}
```

Response:

```json
{
  "tab": "overview",
  "range": {
    "startTs": "2026-07-22T08:00:00.000Z",
    "endTs": "2026-07-23T08:00:00.000Z",
    "intervalInSeconds": 300
  },
  "summary": {},
  "series": {},
  "rankings": {},
  "table": null,
  "warnings": [],
  "generatedAt": "2026-07-23T08:00:01.120Z"
}
```

قواعد:

- Endpoint بر اساس Tab فقط داده لازم همان Tab را برمی‌گرداند.
- مقدار `tab` فقط یکی از `overview|subjects|errors|cache|summary` است.
- `subjectType` فقط برای Tab/Widgetهای Subjectمحور لازم است و یکی از `realUser|virtualUser|team|orgUnit` است.
- Control Plane می‌تواند Queryهای سطح پایین را موازی اجرا کند.
- UI نباید قراردادهای Aggregate داخلی ClickHouse را بداند.
- پاسخ می‌تواند Warning سطح Widget داشته باشد؛ شکست یک Widget نباید همیشه کل Tab را Fail کند.
- Validation و Field/operator allow-list در Backend اجباری است.

### 2. Facets

```http
GET /api/v1/analytics/facets
  ?startTs=...
  &endTs=...
  &field=statusCode
  &q=50
  &limit=50
```

Response:

```json
{
  "field": "statusCode",
  "items": [
    { "value": "500", "label": "500", "count": 42 }
  ],
  "range": null,
  "hasMore": false
}
```

برای Numeric field:

```json
{
  "field": "totalTokens",
  "items": [],
  "range": { "min": 0, "max": 5620 },
  "hasMore": false
}
```

قواعد:

- Facetها به Tenant، Time Range و Predicateهای سازگار فعلی Scope شوند.
- پاسخ حداکثر ۵۰ گزینه؛ Search server-side.
- High-cardinality fields بدون Query خالی کامل Load نشوند.
- مقدار Secret، Raw token یا Token hash هرگز Facet نشود.
- Facetهای Subject از Eventهای موجود در بازه زمانی ساخته شوند تا Subject حذف‌شده یا تغییرنام‌یافته همچنان با Label snapshot قابل انتخاب باشد؛ IAM فقط برای enrichment و وضعیت فعلی استفاده شود.

### 3. Saved Filter CRUD

```http
GET    /api/v1/analytics/saved-filters
POST   /api/v1/analytics/saved-filters
PATCH  /api/v1/analytics/saved-filters/:id
DELETE /api/v1/analytics/saved-filters/:id
```

Create:

```json
{
  "name": "خطاهای Providerهای اصلی",
  "sharedScope": "tenant",
  "schemaVersion": 1,
  "filters": [
    {
      "fieldName": "statusCode",
      "operator": "IN",
      "value": [429, 500, 502]
    }
  ]
}
```

Update:

```json
{
  "expectedVersion": 3,
  "name": "خطاهای Providerهای اصلی",
  "filters": []
}
```

### 4. Metrics Query موجود

`POST /api/v1/metrics/query` و مصرف آن توسط صفحه `/metrics` بدون تغییر ناسازگارکننده باقی می‌ماند. صفحه جدید `/analytics` از `/api/v1/analytics/query` استفاده می‌کند؛ Backend Analytics می‌تواند برای Queryهای سازگار از Metrics query/repository موجود reuse کند. حذف یا Deprecate کردن Endpoint فعلی در دامنه این PRD نیست.

## تغییر قرارداد Metric event

فیلدهای P0 پیشنهادی:

| Field | Type | دلیل |
|---|---|---|
| `cost_irr` | decimal | نمایش و محاسبه بومی ایران |
| `total_tokens` | uint یا derived | Range filter و Summary |
| `fallback_count` | uint | Rescued Requests |
| `fallback_status` | string | Error/Fallback analysis |
| `final_status_code` | uint16 | روشن‌کردن معنای نتیجه نهایی |
| `actor_type` | enum(`real_user`,`virtual_user`,`unknown`) | تفکیک Actor انسانی و سرویس‌محور |
| `actor_id` | nullable UUID | شناسه پایدار Actor اصلی |
| `actor_label_snapshot` | nullable string | نمایش تاریخچه بدون وابستگی به IAM جاری |
| `resolved_team_id` | nullable UUID | Scope تیم یکتای Request و جلوگیری از دوباره‌شماری |
| `resolved_team_label_snapshot` | nullable string | نمایش تاریخچه تیم |
| `resolved_org_unit_id` | nullable UUID | Scope واحد سازمانی یکتای Request |
| `resolved_org_unit_label_snapshot` | nullable string | نمایش تاریخچه واحد سازمانی |
| `credential_type` | enum(`pat`,`vat`,`other`) | منشأ resolve شدن Actor بدون ثبت Secret |
| `credential_id` | nullable UUID | Group/Debug امن بدون افشای Credential |

فیلد P1:

| Field | Type | دلیل |
|---|---|---|
| `trace_id` | nullable string | Deep-link و Exact filter |
| `access_key_id` | nullable UUID | Group By بدون افشای Secret |
| `access_key_alias` | nullable string | نمایش Summary |

### الگوریتم P0 برای Subject attribution

1. Auth context با PAT، Actor را `real_user` و با VAT، Actor را `virtual_user` تعیین می‌کند.
2. Scope صریح Credential یا Request برای `resolved_team_id` مبنا است؛ عضویت‌های دیگر Actor به Event افزوده نمی‌شوند.
3. `resolved_org_unit_id` از Scope صریح یا Parent تیم resolved می‌شود. نبود هر دو به معنی Tenant-level است.
4. قبل از enqueue رویداد، ID و Label snapshot چهار Subject ثبت می‌شوند؛ خطای enrichment نباید مسیر Inference را Fail کند و با `unknown`/`null` ادامه می‌دهد.
5. فیلدهای قدیمی `user_id`، `user_email`، `virtual_account_*` و `team_ids` در دوره سازگاری خوانده می‌شوند، اما API Analytics فقط نام‌های canonical جدید را expose می‌کند.
6. Backfill فقط در صورت وجود evidence قطعی انجام می‌شود؛ Event قدیمی مبهم به Subject حدس‌زده‌شده منتسب نمی‌شود.

قواعد امنیتی:

- Raw API key، Token hash، Prompt body، Response body و PII غیرضروری در Metrics ذخیره نشوند.
- Email کاربر حقیقی در Snapshot فقط در صورت نیاز محصول ذخیره و در نمایش بر اساس مجوز Mask شود؛ ترجیح با نام نمایشی یا alias غیرحساس است.
- Metadata فقط از Whitelist Tenant و Whitelist سیستمی عبور کند.

## ClickHouse و Query plan

### استفاده مجدد

- `gateway_model_events` منبع Raw اصلی باقی می‌ماند.
- `gateway_model_metrics_1m` و Materialized View فعلی برای Queryهای Model/Virtual Model استفاده می‌شوند.
- Raw fallback موجود برای Group Byها و Filterهای پشتیبانی‌نشده حفظ می‌شود.

### تغییرات لازم

1. Migration برای ستون‌های P0 Metric event.
2. Mappingهای جدید:
   - `costInIRR`;
   - `totalTokens`;
   - `fallbackCount`;
   - `fallbackStatus`.
3. Aggregationهای `mean` برای Latency.
4. Facet query زمان‌مند و Tenant-safe.
5. Summary pagination و sort server-side.
6. Benchmark برای Group Byهای کاربر حقیقی، کاربر مجازی، تیم، واحد سازمانی، Provider، Status و Error.
7. در صورت نقض SLO، Materialized Viewهای جدا و کم‌کاردینالیتی اضافه شوند؛ یک MV چندبعدی عظیم ساخته نشود.
8. Queryهای Subject باید از ستون‌های snapshot استفاده کنند و در Hot path به PostgreSQL/IAM Join نزنند.

### Retention

- MVP selectable range: حداکثر ۹۰ روز.
- Retention فعلی Raw event: ۹۰ روز.
- Aggregate retention طولانی‌تر فعلی حفظ می‌شود، ولی UI تا تعریف Product tier از ۹۰ روز عبور نمی‌کند.
- اگر Tenant Retention کوتاه‌تری دارد، Preset و Custom picker باید آن را رعایت کنند.

## مجوزها و Isolation

### P0

- `analytics:read`
- `analytics:saved_filters_manage`

Default پیشنهادی:

- Tenant Admin: هر دو مجوز.
- Team Admin: هر دو مجوز، تا زمان پیاده‌سازی Team-scoped data فقط اگر سیاست Tenant اجازه دهد.
- Member/Readonly: بدون مجوز پیش‌فرض.

قواعد:

- `settings:list` دیگر مجوز معنایی اصلی Analytics نباشد.
- Frontend nav، Server page و Control Plane همگی یک مجوز مشترک را enforce کنند.
- تمام Queryها Tenant ID را فقط از Auth context بگیرند، نه Request body.
- Saved Filter ID باید همراه Tenant ID Query شود.

### Deferred

- Team-scoped authorization برای محدودکردن کل Dataset قابل مشاهده؛
- Organization Unit-scoped authorization برای محدودکردن کل Dataset قابل مشاهده؛
- Workspace-level access override مشابه Portkey.

## رفتار Loading، Error و Empty

### Loading

- Skeleton هر Widget؛
- Tab shell و Filter bar در Refresh ناپدید نشوند؛
- داده قبلی هنگام Refresh قابل مشاهده بماند و حالت Updating مشخص باشد.

### Partial Error

- هر Widget وضعیت Error مستقل و Retry دارد.
- Failure Facet نباید Chartهای قبلی را پاک کند.
- Failure ClickHouse نباید Gateway traffic را متوقف کند.

### Empty

- «هیچ ترافیکی در این بازه ثبت نشده» با «هیچ نتیجه‌ای برای این فیلتر پیدا نشد» متفاوت باشد.
- Empty Table header حفظ شود.
- CTA پیشنهادی: Reset filters یا بازکردن بازه زمانی بزرگ‌تر.

### Stale data

- `generatedAt` و در صورت امکان Data freshness نمایش داده شود.
- اگر Queue/ingest lag از آستانه بیشتر شد Warning نمایش داده شود.

## RTL، Accessibility و Design System

الزامات:

- استفاده از `Tabs`، `AppCard`، `MetricCard`، `CompactValue`، `ActiveFilterChips`، `FilterCombobox`، `RangeFilterField`، `ChipInput`، `DateTimeRangeField`، `RefreshButton`، `EmptyState`، `TablePagination`، `ConfirmDialog` و `UnsavedChangesDialog` موجود.
- ساخت Component جدید فقط وقتی رفتار Saved Filter/Filter editor تکرارشونده است.
- اگر `AnalyticsFilterBar` به Component reusable تبدیل شد، Playground Design System به‌روزرسانی شود.
- هیچ Raw color utility یا Arbitrary spacing جدید استفاده نشود.
- Focus ring، Keyboard navigation و Screen reader label برای همه Chipها و Icon actionها الزامی است.
- ترتیب منطقی در RTL حفظ شود؛ Technical valueها LTR باقی بمانند.
- Chartها فقط با Color معنا ندهند؛ Legend، Tooltip و Label لازم است.
- Tab panelها `aria-labelledby` و Loadingها `role=status` مناسب داشته باشند.

## Non-Functional Requirements

### Performance

- Tab initial load p95 برای ۲۴ ساعت: کمتر از ۲.۵ ثانیه.
- Analytics API p95:
  - ۱ ساعت: کمتر از ۸۰۰ms؛
  - ۲۴ ساعت: کمتر از ۱.۵s؛
  - ۷ روز: کمتر از ۲s؛
  - ۳۰/۹۰ روز: کمتر از ۳s.
- Facet search p95 کمتر از ۵۰۰ms.
- Saved Filter CRUD p95 کمتر از ۵۰۰ms.
- UI نباید با هر keystroke Query کامل اجرا کند.

### Reliability

- Query API availability حداقل 99.9%.
- Metrics ingest loss کمتر از 0.1%.
- Analytics/ClickHouse failure روی Inference traffic اثر نگذارد.
- Query timeout سمت Backend و Cancellation هنگام تغییر سریع Filter الزامی است.

### Security

- Strict allow-list برای Field، Operator، Aggregation و Group By.
- عدم ساخت SQL از نام Field خام.
- Limit برای Filter count، Array size، Metadata key و Range.
- Audit برای Create/Update/Rename/Delete Saved Filter.
- عدم نمایش Secret در Facet، Summary یا URL.

### Data correctness

- Cost Analytics با Spend/Wallet روزانه Reconcile شود.
- Cache savings مطابق `model-metrics-contract-v2.md`.
- Time bucket و Timezone در تست‌های DST/مرز روز پوشش داده شوند.
- Unknown/missing values در همه تب‌ها Semantic یکسان داشته باشند.

## Metrics موفقیت محصول

### Activation

- درصد Tenantهای دارای ترافیک که Analytics را در ۷ روز اول باز می‌کنند.
- درصد Sessionهایی که حداقل یک Tab غیر Overview را باز می‌کنند.

### Usage

- تعداد Filter apply در هر Session.
- تعداد Saved Filter ایجادشده و Applyشده.
- نرخ بازگشت به Saved Filter طی ۳۰ روز.
- نرخ Drill-down از Analytics به Logs/Traces.

### UX

- زمان متوسط رسیدن از Overview به یک Breakdown فیلترشده.
- نرخ Reset پس از Filter apply ناموفق.
- نرخ Query cancellation و تکرار سریع درخواست.

### Reliability

- Query latency p50/p95/p99.
- Partial widget failure rate.
- ClickHouse timeout rate.
- Saved Filter conflict/error rate.

## وابستگی‌ها

- ClickHouse schema و migration.
- Metrics processor و event contract.
- Cost IRR finalized snapshot.
- Fallback count/status از Request log lifecycle.
- RBAC catalog و Role migration.
- Design System components.
- Translationهای `fa` و `en`.
- لینک‌دهی Logs/Traces.
- مستندات کاربری `observability-analytics`.

## ریسک‌ها و کاهش ریسک

| ریسک | اثر | کاهش ریسک |
|---|---|---|
| Queryهای متعدد هر Tab | فشار ClickHouse | Analytics BFF، parallel محدود، cancellation، cache کوتاه |
| Group By پرکاردینالیتی | Latency و مصرف حافظه | limit، pagination، benchmark و MV جدا در صورت نیاز |
| هم‌زیستی دو UI Metrics/Analytics | بدهی و سردرگمی | جایگاه و Job-to-be-done روشن در Navigation، Component/Backend reuse و Telemetry مقایسه‌ای |
| انتساب چند تیم به یک Actor | دوباره‌شماری Request و Cost | یک Scope resolved در لحظه Request و Snapshot تاریخی |
| تغییر نام/عضویت Subject | بازنویسی یا گم‌شدن تاریخچه | ذخیره ID و Label snapshot و عدم Join اجباری با IAM جاری |
| Saved Filter مشترک قابل تغییر | overwrite هم‌زمان | version و 409 conflict |
| نشت داده Tenant | بحرانی | Tenant ID فقط از Auth context و تست isolation |
| نمایش Secret در API Key group | بحرانی | فقط alias/type/id مجاز |
| اختلاف Cost | اعتماد مالی | Reconciliation و نمایش Estimated/Final |
| Chart overload | کاهش خوانایی | تب‌بندی و Widgetهای محدود هر تب |
| ناسازگاری RTL | UX ضعیف | استفاده از Componentهای موجود و RTL E2E |
| Event schema ناقص برای Errors | Widget نادرست | افزودن fallback fields پیش از فعال‌سازی Rescued Requests |

## MVP Cutline

### P0 — باید در Launch باشد

- Route و Navigation جدید؛
- حفظ کامل Route و صفحه فعلی `/metrics` در کنار `/analytics`؛
- پنج Tab بدون Feedback؛
- تب Subjects و چهار نوع کاربر حقیقی، کاربر مجازی، تیم و واحد سازمانی؛
- Subject attribution snapshot بدون شمارش مضاعف؛
- Filter bar تراشه‌ای؛
- Time Range کامل تا ۹۰ روز و Custom؛
- فیلترهای P0 جدول Filter catalog؛
- Saved Filter Create/Apply/Update/Rename/Edit/Delete؛
- Cost IRR و USD؛
- Analytics domain API؛
- Permissions اختصاصی؛
- Drill-down به Logs؛
- Loading/Error/Empty قابل استفاده؛
- RTL و Accessibility؛
- تست و Telemetry.

### P1 — پس از پایدارشدن MVP

- Trace ID filter؛
- Metadata filter/group؛
- Provider Model و Access Key group؛
- Latency range؛
- Top Subjects by Cost؛
- قابلیت Pin/Favorite کردن Saved Filter؛
- Team/Organization Unit-scoped authorization؛
- مقایسه با بازه قبلی.

### P2 / Deferred

- Cross-surface Saved Filters با Logs؛
- Export/Reports؛
- Scheduled email/Telegram reports؛
- Config/Prompt group؛
- Workspace-level analytics permission overrides؛
- Retention tierهای تجاری؛
- Real-time auto-refresh.

## فازبندی پیاده‌سازی

### Phase 0 — Contract و Migration

- [ ] قرارداد Analytics query/facets/saved filters نهایی شود.
- [ ] قرارداد چهار Subject و الگوریتم resolve شدن Actor/Team/Org Unit نهایی شود.
- [ ] Permissionهای Analytics اضافه شوند.
- [ ] Migration PostgreSQL برای Saved Filters اضافه شود.
- [ ] Metric event با IRR و Fallback fields توسعه یابد.
- [ ] Metric event با فیلدهای Subject attribution و snapshot توسعه یابد.
- [ ] ClickHouse migration و Backward compatibility تست شود.

### Phase 1 — Query و Saved Filters Backend

- [ ] `internal/analyticsquery` یا package معادل برای orchestration ساخته شود.
- [ ] Analytics query handler و Facet handler اضافه شوند.
- [ ] Saved Filter repository/handlers با الگوی Logs پیاده‌سازی شوند.
- [ ] Validation، limits، tenant isolation و conflict تست شوند.
- [ ] Benchmark بازه‌های ۱h/۲۴h/۷d/۳۰d/۹۰d اجرا شود.

### Phase 2 — Filter bar و Route

- [ ] `/analytics` و Navigation اضافه شود.
- [ ] Regression guard برای حفظ `/metrics` و قرارداد Query فعلی اضافه شود.
- [ ] Filter state و URL codec پیاده‌سازی شود.
- [ ] Filter editorهای Range/Multi-select/Text ساخته شوند.
- [ ] Saved Filter state machine کامل شود.
- [ ] Playground برای pattern reusable به‌روزرسانی شود.

### Phase 3 — Tabها

- [ ] Overview.
- [ ] Subjects برای چهار نوع و بدون Feedback.
- [ ] Errors و Rescued Requests.
- [ ] Cache و Savings/Speedup.
- [ ] Summary Table و Group By.
- [ ] Deep-link به Logs/Traces.

### Phase 4 — Hardening و Rollout

- [ ] Unit، Integration، E2E، RTL و Accessibility tests.
- [ ] Load test و Query SLO.
- [ ] Cost reconciliation.
- [ ] Feature flag و Tenant داخلی.
- [ ] Selected tenants rollout.
- [ ] GA در کنار Metrics و مقایسه Telemetry استفاده دو صفحه.

## مسیرهای کد پیشنهادی

Frontend:

- `apps/tenant-admin/app/analytics/page.tsx`
- `apps/tenant-admin/app/components/analytics-dashboard.tsx`
- `apps/tenant-admin/app/components/analytics-filter-bar.tsx`
- `apps/tenant-admin/app/api/analytics/*`
- `apps/tenant-admin/messages/modules/analytics/{fa,en}.json`
- `apps/tenant-admin/app/components/admin-shell-nav.ts`
- `apps/tenant-admin/app/design-system/page.tsx` در صورت reusable pattern

Contracts/Backend:

- `packages/contracts/platform/analytics.go`
- `packages/contracts/platform/metrics.go`
- `services/control-plane/internal/admin/analytics_handlers.go`
- `services/control-plane/internal/storage/analytics_saved_filters_repo.go`
- `services/control-plane/internal/storage/metrics_repo_clickhouse.go`
- `services/control-plane/internal/observability/metrics/processor.go`
- `services/control-plane/internal/server/routes_tenant_admin.go`

Migrations:

- `infra/k8s/local/migrations/*_analytics_saved_filters.sql`
- `infra/clickhouse/migrations/*_analytics_event_fields.sql`

Docs:

- `apps/docs/docs/observability-analytics.mdx`
- افزودن Cross-link غیرمخرب به `apps/docs/docs/observability-model-metrics.mdx`

## استراتژی تست

### Unit

- Filter state reducer و URL codec؛
- Saved Filter dirty detection؛
- Time range/interval selection؛
- Metric definitions؛
- Filter-to-query mapping؛
- Permissions و scope.

### Backend integration

- ClickHouse fixture برای هر Tab؛
- IRR/USD reconciliation؛
- Cache hit-rate/savings/speedup؛
- Rescued request؛
- Summary pagination/sort؛
- Saved Filter CRUD، duplicate، conflict و tenant isolation.

### Frontend integration

- تغییر Tab و حفظ Filter/Time Range؛
- Create/Edit/Rename/Delete/Cancel Saved Filter؛
- Loading/partial error/empty؛
- Deep-link با Saved Filter معتبر، حذف‌شده و غیرمجاز.

### E2E

- فارسی RTL و انگلیسی LTR؛
- Keyboard-only filter flow؛
- Mobile tab/filter behavior؛
- Analytics → Logs drill-down؛
- دسترسی مستقل به `/metrics` و `/analytics` و نبود Redirect میان آن‌ها؛
- Regression جریان‌های اصلی صفحه موجود `/metrics`.

### Performance

- Query concurrency برای پنج Widget؛
- ۹۰ روز داده؛
- کاربر حقیقی/کاربر مجازی/تیم/واحد سازمانی با cardinality بالا؛
- Facet search؛
- cancellation هنگام تغییر سریع Filter.

## معیارهای پذیرش

### صفحه و تب‌ها

- [ ] `/analytics` برای کاربر مجاز باز می‌شود.
- [ ] `/metrics` همچنان صفحه فعلی را با Query Parameterها و رفتار قبلی باز می‌کند.
- [ ] هر دو آیتم Metrics و Analytics در Navigation مستقل و قابل تشخیص‌اند.
- [ ] پنج Tab موردنیاز وجود دارند.
- [ ] هیچ Tab، Widget یا Filter مربوط به Feedback وجود ندارد.
- [ ] Filter و Time Range هنگام تغییر Tab حفظ می‌شوند.
- [ ] Refresh state صفحه را حفظ می‌کند.

### Filter

- [ ] فیلترهای Fieldهای مختلف با AND و Multi-valueها با IN اجرا می‌شوند.
- [ ] Chip قابل ویرایش، حذف و Keyboard access است.
- [ ] Numeric slider فقط هنگام Commit Query می‌زند.
- [ ] Time Range در Local time نمایش و با UTC ارسال می‌شود.
- [ ] URL فقط state Commit‌شده را نگه می‌دارد.
- [ ] محدودیت تعداد/اندازه Filter در Client و Server enforce می‌شود.

### Saved Filters

- [ ] Create بدون نام معتبر ممکن نیست.
- [ ] نام تکراری پاسخ 409 و پیام روشن دارد.
- [ ] Apply، Time Range، Tab، Subject type، Latency metric و Group By را حفظ می‌کند.
- [ ] تغییر Predicate، Saved Filter را Dirty می‌کند.
- [ ] تغییر Time Range، Tab، Subject type، Latency metric یا Group By، Saved Filter را Dirty نمی‌کند.
- [ ] Update با version قدیمی پاسخ 409 می‌دهد.
- [ ] Rename Predicateها را تغییر نمی‌دهد.
- [ ] Delete نیاز به Confirm دارد.
- [ ] کاربر Tenant دیگر به Saved Filter دسترسی ندارد.

### داده و محاسبات

- [ ] Overview عدد Cost، Tokens، Latency، Requests و Active Actors درست دارد.
- [ ] تب Subjects هر چهار نوع کاربر حقیقی، کاربر مجازی، تیم و واحد سازمانی را عرضه می‌کند.
- [ ] هر Request در Breakdown هر نوع Subject حداکثر یک بار شمرده می‌شود.
- [ ] تغییر عضویت، نام یا Parent یک Subject، نتیجه تاریخی ثبت‌شده را بازنویسی نمی‌کند.
- [ ] Eventهای Legacy مبهم به Subject حدس‌زده‌شده منتسب نمی‌شوند و در Bucket «نامشخص» قرار می‌گیرند.
- [ ] Error rate و Rescued Requests مطابق تعریف محاسبه می‌شوند.
- [ ] Cache hit-rate از `none/bypass/error` صرف‌نظر می‌کند.
- [ ] Cache savings با Contract v2 منطبق است.
- [ ] Cost IRR و USD با داده مالی Reconcile می‌شوند.
- [ ] Summary Group By و Sort/Pagination server-side است.

### کیفیت

- [ ] Query SLOهای تعریف‌شده پاس می‌شوند.
- [ ] خرابی Analytics روی Inference traffic اثر ندارد.
- [ ] RTL، Accessibility و Mobile E2E پاس می‌شوند.
- [ ] هیچ Secret یا Raw token در URL، Facet، Table یا Log مرورگر ظاهر نمی‌شود.
- [ ] Lint، Typecheck، Go tests، Frontend tests و Buildهای مرتبط پاس می‌شوند.

## بدهی UI مجاز پس از MVP

- Cross-surface Saved Filters با Logs؛
- مقایسه بازه فعلی و قبلی؛
- تنظیمات شخصی Currency/Default tab؛
- قابلیت Pin کردن Saved Filter؛
- Export و Share link کوتاه؛
- محدودسازی کل Dataset بر اساس سطح دسترسی Team/Organization Unit.

هیچ بدهی مربوط به RTL، Keyboard accessibility، Tenant isolation یا Correctness مالی قابل انتقال به پس از MVP نیست.

## سؤال‌های باز برای تأیید محصول

این موارد مانع شروع طراحی فنی نیستند، ولی پیش از GA باید نهایی شوند:

1. نام نهایی Navigation در فارسی «تحلیل‌ها» باشد یا «تحلیل و پایش»؟
2. آیا Saved Filter پیش‌فرض Tenant-shared باشد یا Private؟
3. آیا Team Admin به‌صورت پیش‌فرض Analytics Tenant-wide می‌بیند؟
4. Currency اصلی بر اساس Tenant billing currency انتخاب شود یا همیشه IRR؟
5. Retention تجاری در Launch ثابت ۹۰ روز باشد یا از Plan entitlement خوانده شود؟
6. مرزبندی توضیحی Navigation میان «متریک مدل‌ها» و «تحلیل‌ها» با چه Microcopyای برای کاربر روشن‌تر می‌شود؟
7. آیا Label snapshot کاربر حقیقی باید نام نمایشی باشد یا Email mask‌شده؟
8. در Credentialهای قدیمی بدون Scope صریح، Team/Org Unit نامشخص بماند یا Migration قطعی جداگانه تعریف شود؟

## Definition of Done

قابلیت زمانی Done است که:

- پنج Tab با داده واقعی و بدون Feedback در Production فعال باشند؛
- Filter bar و Saved Filters تمام Stateهای این سند را پوشش دهند؛
- داده مالی، Cache و Error با Fixture و Reconciliation تأیید شده باشد؛
- مجوزها و Tenant isolation تست شده باشند؛
- SLO، RTL، Accessibility و Mobile پذیرفته شده باشند؛
- `/metrics` و لینک‌های قبلی آن بدون Regression باقی مانده باشند؛
- `/analytics` در `tenant-admin` به‌صورت Route و Navigation مستقل عرضه شده باشد؛
- Subject attribution برای چهار نوع محصول با Fixture و آزمون عدم دوباره‌شماری تأیید شده باشد؛
- مستندات کاربری و Runbook عملیاتی منتشر شده باشند.
