# PRD تجربه‌ی Model Catalog شاهراه

**وضعیت:** پیشنهادی برای تصویب  
**تاریخ:** ۱۴۰۵/۰۵/۰۱ (2026-07-23)  
**مالک محصول:** Tenant Admin / AI Gateway  
**دامنه:** مدیریت `TenantProviderAccount` و مدل‌های قابل‌فراخوانی آن  
**الهام اصلی:** مشاهده‌ی مستقیم Model Catalog در Portkey و تطبیق آن با قراردادها و UI فعلی شاهراه

---

## 1. خلاصه‌ی تصمیم

صفحه‌ی فعلی Provider Accounts باید به یک تجربه‌ی واحد با عنوان «کاتالوگ مدل‌ها» تبدیل شود که دو نمای اصلی و هم‌ارز دارد:

1. **ارائه‌دهنده‌های AI**: مدیریت اتصال‌های واقعی هر tenant؛
2. **مدل‌ها**: نمای تجمیعی مدل‌های قابل‌استفاده روی همه‌ی اتصال‌ها.

این تغییر یک بازطراحی تجربه و معماری اطلاعات است، نه بازطراحی دامنه. در MVP هیچ entity تازه‌ای ساخته نمی‌شود:

- «AI Provider» در UI همان `TenantProviderAccount` موجود است؛
- «Model availability» از `TenantProviderAccountModel`های موجود محاسبه می‌شود؛
- اطلاعات پایه‌ی provider و model از Provider Catalog موجود خوانده می‌شود؛
- مجوزها از `TenantProviderAccountAccess` موجود می‌آیند؛
- روش اتصال از `auth_mode`، `provider_config`، `has_credential` و
  `use_platform_managed_credentials` موجود خوانده می‌شود.

بیشترین شباهت به Portkey در این موارد ایجاد می‌شود: دو تب Providers/Models، جست‌وجو و فیلتر چندانتخابی، مسیر دو مرحله‌ای افزودن provider، صفحه‌ی جزئیات provider، مدیریت مدل‌ها، drawer جزئیات مدل، نمونه‌کد و اکشن‌های Playground/Logs. قابلیت‌های متمایز شاهراه مانند دسترسی مدیر/کاربر، BYOK در برابر credential مدیریت‌شده، قیمت ریالی و تست دسترسی مؤثر حفظ می‌شوند.

---

## 2. مسئله

شاهراه اجزای دامنه‌ای لازم را دارد، اما تجربه‌ی فعلی پراکنده و account-centric است:

- تب‌های فعلی «همه‌ی providerها / اتصال‌های فعال» با مدل ذهنی «من چه مدل‌هایی می‌توانم استفاده کنم؟» هم‌راستا نیستند؛
- کاربر برای دیدن مدل‌های قابل‌استفاده باید وارد تک‌تک accountها شود؛
- مسیر ساخت account یک wizard چهارمرحله‌ای کامل اما سنگین است؛
- صفحه‌ی manage بیشتر شبیه فرم تنظیمات است تا صفحه‌ی هویت و عملیات یک provider؛
- شناسه‌ی ناقص `provider:account` با عنوان FQN کپی می‌شود، درحالی‌که runtime شناسه‌ی کامل
  `provider:account:model` می‌خواهد؛
- کشف مدل، provisioning، نمونه‌کد، تست، Playground و Logs در یک مسیر پیوسته کنار هم نیستند.

در Portkey، کاربر از یک کاتالوگ مرکزی شروع می‌کند، provider متصل را می‌بیند، مدل‌های حاصل از آن را در نمای تجمیعی پیدا می‌کند و از همان‌جا به نمونه‌کد، تست، Playground یا Logs می‌رود. هدف این PRD انتقال همین منطق به شاهراه، با حفظ دامنه و محدودیت‌های محصول خودمان است.

---

## 3. شواهد و روش بررسی

### 3.1 مشاهده‌ی مستقیم Portkey

در تاریخ سند، مسیرهای زیر در محیط واقعی Portkey بررسی شدند:

- `/model-catalog/`، تب **AI Providers**
- `/model-catalog/`، تب **Models**
- `/model-catalog/models?integration=<id>`
- `/model-catalog/providers`
- `/model-catalog/providers/<provider-id>/details`
- `/integrations/update-provider/<provider-type-id>/<integration-id>`
- drawer جزئیات یک model
- modalهای Edit Provider، Delete Provider، Add Model و Set Pricing

فرم‌ها تا مرحله‌ی اعتبارسنجی و تولید slug با داده‌ی آزمایشی پر شدند، اما هیچ provider/model ساخته یا حذف نشد و هیچ test request هزینه‌دار ارسال نشد.

### 3.2 منابع داخلی Portkey

منابع مرتبط زیر از `.portkey-docs` بررسی شدند:

- Model Catalog
- Integrations
- Custom Models
- Model Overrides
- Portkey Models
- Provider API: create/list/retrieve/update/delete
- Integration Model Access API: list/update/delete
- Bedrock assumed role

### 3.3 وضعیت فعلی شاهراه

منابع اصلی داخلی:

- `plans/README.md`
- `plans/phase0-tenant-provider-account-contract.md`
- `plans/phase1-tenant-provider-account-backend-foundation.md`
- `plans/phase3-provider-model-catalogue.md`
- `plans/provider-account-models-ux-flow.md`
- `packages/contracts/platform/tenant_provider_account.go`
- `packages/database/schema.prisma`
- `services/control-plane/internal/modelid/modelid.go`
- routeها و handlerهای `tenant-provider-accounts`
- `apps/tenant-admin/app/components/llm-integrations-management.tsx`
- `apps/tenant-admin/app/components/provider-account-wizard.tsx`
- `apps/tenant-admin/app/components/provider-account-wizard/models-editor.tsx`
- Design System و Component Usage پنل tenant

اثر اسناد قبلی بر این تصمیم:

- قرارداد Phase 0 ریشه‌ی دامنه را روی `TenantProviderAccount` قفل کرده و این PRD همان تصمیم را حفظ می‌کند.
- Phase 1 APIهای account، models، access و credential را فراهم کرده و مبنای MVP است.
- `provider-account-models-ux-flow.md` قابلیت‌های دسترسی، credential و readiness فعلی را حفظ می‌کند، اما معماری اطلاعات آن با الگوی کاتالوگ جایگزین می‌شود.
- Phase 3 قدیمی، entityها و versioning تازه برای کاتالوگ پیشنهاد می‌کند. این بخش با قید صریح این PRD ناسازگار است و در دامنه‌ی این اجرا قرار نمی‌گیرد.
- `plans/README.md` باعث می‌شود قیمت ریالی/تومانی، tenant isolation، مسیر تونل ایران–اروپا و RTL جزو نیازمندی‌های اصلی باشند، نه جزئیات تزئینی.

---

## 4. واقعیت‌های مشاهده‌شده در Portkey

این بخش «مشاهده» است، نه الزام خودکار برای شاهراه.

### 4.1 معماری اطلاعات

- عنوان صفحه `Model Catalog` و اکشن اصلی `Add Provider` است.
- دو تب هم‌سطح وجود دارد: `AI Providers` و `Models`.
- Provider یک اتصال workspace-level به یک integration credential است.
- یک integration می‌تواند پشت چند provider قرار بگیرد.
- مدل قابل‌فراخوانی با قالب `@provider_slug/model_name` نمایش داده می‌شود.

### 4.2 تب AI Providers

Toolbar شامل:

- جست‌وجوی فوری؛
- فیلتر چندانتخابی `Org Integrations`؛
- فیلتر چندانتخابی `AI Providers`؛
- نمایش تعداد فیلترهای فعال؛
- دکمه‌ی Apply داخل منوی فیلتر.

جدول شامل:

- Name
- Slug با اکشن Copy
- Created On
- Created By
- منوی عملیات هر ردیف

عملیات هر provider:

- View Models
- Edit Provider
- Delete Provider

`View Models` کاربر را به تب Models با integration filter فعال می‌برد.

### 4.3 تب Models

Toolbar شامل:

- جست‌وجوی model؛
- فیلتر `Org Integrations`؛
- فیلتر `Workspace Provider`.

جدول شامل:

- Name
- Provider family
- Availability به‌شکل «روی چند provider در دسترس است»

با انتخاب یک model، drawer جزئیات باز می‌شود و route اصلی صفحه حفظ می‌شود.

### 4.4 drawer جزئیات model

drawer شامل:

- نام مدل و provider سازنده؛
- انتخاب زبان نمونه‌کد: NodeJS، Python، cURL و SDKهای OpenAI؛
- انتخاب provider instance؛
- نمونه‌کد با model identifier نهایی؛
- API key ماسک‌شده؛
- Copy code؛
- Run Test Request؛
- Copy to Integrate؛
- وضعیت/پاسخ test request؛
- Max Tokens؛
- model identifier قابل‌کپی؛
- modalities؛
- لینک Try in Playground؛
- لینک View Logs با فیلتر model.

### 4.5 صفحه‌ی جزئیات provider

ساختار صفحه یک detail page بلند با sectionهای زیر است:

- Overview
- API Setup & Code
- Models
- More Details

Overview:

- آخرین زمان استفاده؛
- کل requestها؛
- requestهای هفته‌ی جاری و قبل؛
- درصد تغییر؛
- View Analytics.

API Setup & Code:

- انتخاب زبان؛
- انتخاب model؛
- کپی نمونه‌کد؛
- Run Test Request؛
- Copy to Integrate.

Models:

- تعداد مدل‌ها؛
- نام، تاریخ ایجاد و modalities؛
- کپی model identifier؛
- Manage Models.

More Details:

- status؛
- slug؛
- creator؛
- creation time؛
- credential/integration backing؛
- Edit Model Access؛
- Go to Integration.

منوی More options:

- Edit
- Delete

### 4.6 ویرایش و حذف provider

Edit modal:

- Name
- Short Description
- نمایش integration پشت provider به‌صورت read-only
- Manage Integration
- Save/Cancel

Delete modal:

- هشدار توقف فوری requestهای وابسته؛
- checkbox اجباری برای تأیید درک اثر؛
- Delete تا قبل از تأیید غیرفعال است.

### 4.7 افزودن provider

فلو دو مرحله دارد:

#### مرحله 1: Select Service

- انتخاب AI service از فهرست کامل؛
- quick choice برای providerهای محبوب؛
- انتخاب credential/integration موجود؛
- Add New Integration در صورت نبود credential.

#### مرحله 2: Provider Details

اگر integration موجود انتخاب شود:

- مرور service و credential؛
- Name؛
- Description؛
- Slug که از name تولید و قابل‌ویرایش است؛
- Back/Create.

اگر integration جدید ساخته شود:

- scope: کل سازمان یا فقط workspace؛
- name/description/slug؛
- API key و فیلدهای provider-specific؛
- optional fields؛
- advanced options؛
- pricing adjustments؛
- Create.

### 4.8 مدیریت integration و model provisioning

صفحه‌ی integration دو بخش اصلی دارد:

- Integration Details
- Model Provisioning

Integration Details:

- name، description و slug غیرقابل‌تغییر؛
- credential ماسک‌شده و Edit؛
- optionهای provider-specific؛
- Custom Host؛
- Custom Headers؛
- Pricing Adjustment با multiplier ساده یا JSON سفارشی.

Model Provisioning:

- search؛
- Select All/Deselect All؛
- toggle فعال‌بودن هر model؛
- Set Pricing؛
- Add Model؛
- فعال‌کردن خودکار modelهای جدید provider.

Add Model:

- slug و description؛
- نوع Fine-tuned یا Custom؛
- base model برای fine-tuned؛
- custom pricing؛
- custom host.

Set Pricing:

- inherit یا custom pricing؛
- input/output/cache pricing؛
- custom host.

---

## 5. چه چیزی از Portkey کپی نمی‌شود

شباهت موردنظر، شباهت در مدل ذهنی و flow است، نه کپی کور UI یا دامنه.

- entity مستقل Integration ساخته نمی‌شود؛ account فعلی شاهراه اتصال و credential را مالک است.
- scope جدید «organization/workspace integration» اضافه نمی‌شود؛ tenant scope و access rule فعلی منبع حقیقت است.
- قالب `@slug/model` به runtime تحمیل نمی‌شود؛ قالب canonical فعلی شاهراه
  `provider:account:model` حفظ می‌شود.
- custom model آزاد در MVP اضافه نمی‌شود، چون backend فعلی انتخاب model را در برابر catalog معتبرسازی می‌کند.
- advanced header، host override و pricing multiplier عمومی بدون قرارداد امنیتی و billing روشن اضافه نمی‌شوند.
- breadth فهرست providerهای Portkey معیار MVP نیست؛ فقط providerهای موجود در catalog شاهراه نمایش داده می‌شوند.
- secret خام در UI نمایش یا persist نمی‌شود.

---

## 6. کاربران و use caseها

### 6.1 مدیر tenant

- provider جدید متصل می‌کند؛
- credential اختصاصی یا مدیریت‌شده را انتخاب می‌کند؛
- مدل‌ها و قیمت را provision می‌کند؛
- دسترسی مدیران و کاربران را تعیین می‌کند؛
- سلامت اتصال را تست می‌کند؛
- provider را ویرایش یا حذف می‌کند.

### 6.2 توسعه‌دهنده

- مدل قابل‌استفاده را پیدا می‌کند؛
- می‌فهمد مدل روی کدام accountها در دسترس است؛
- FQN کامل و نمونه‌کد را کپی می‌کند؛
- مدل را در Playground باز می‌کند؛
- logهای همان مدل را می‌بیند.

### 6.3 مدیر مالی/عملیاتی

- قیمت public/private و معادل ریالی/تومانی را می‌بیند؛
- مدل‌های فاقد قیمت معتبر را تشخیص می‌دهد؛
- health/readiness provider را می‌بیند؛
- مصرف و روند requestها را از detail page دنبال می‌کند.

### 6.4 پشتیبان/عیب‌یاب

- credential را تست می‌کند؛
- effective access یک principal را می‌آزماید؛
- دلیل عدم دسترسی model را می‌بیند؛
- از model detail به logهای فیلترشده می‌رود.

---

## 7. اهداف

### P0

1. کاربر از یک route واحد به providerها و modelهای tenant دسترسی داشته باشد.
2. تمام accountهای موجود بدون migration داده در نمای Providers دیده شوند.
3. تمام modelهای provision‌شده در نمای Models به‌شکل تجمیعی دیده شوند.
4. کاربر از model به account قابل‌استفاده، FQN کامل، نمونه‌کد، Playground و Logs برسد.
5. create/manage/delete قابلیت‌های فعلی account از دست نرود.
6. مسیر افزودن provider از چهار گام به دو مرحله‌ی قابل‌فهم تبدیل شود.
7. UI فارسی/RTL، دسترس‌پذیر و مبتنی بر design system موجود باشد.

### P1

1. provider detail خلاصه‌ی مصرف و readiness داشته باشد.
2. model drawer امکان test request کنترل‌شده داشته باشد.
3. bulk provisioning، deep link و filter persistence کامل شود.

### P2

1. custom/fine-tuned model با همان `TenantProviderAccountModel` و پس از تغییر قرارداد validation پشتیبانی شود.
2. advanced connection options با قرارداد امنیتی روشن اضافه شود.
3. model recommendation و مقایسه‌ی هزینه/کیفیت اضافه شود.

---

## 8. Non-goals

- طراحی entity تازه برای Provider، Integration یا Model Catalog؛
- تغییر tenancy یا مجوزهای پایه؛
- ساخت billing engine تازه؛
- تغییر canonical runtime identifier؛
- جایگزینی Playground یا Logs؛
- افزودن provider upstream جدید؛
- پیاده‌سازی model routing یا Virtual Models در این صفحه؛
- ساخت marketplace عمومی providerها.

---

## 9. معماری اطلاعات پیشنهادی

### 9.1 route اصلی

```text
/provider-accounts
  ?tab=providers             نمای پیش‌فرض
  ?tab=models                کاتالوگ مدل‌های tenant
  &q=...                     جست‌وجو
  &provider=...              provider family filter
  &account=...               account filter یا detail selection
  &model=...                 model drawer selection
```

نام route برای سازگاری حفظ می‌شود؛ عنوان navigation و صفحه به «کاتالوگ مدل‌ها» تغییر می‌کند.

### 9.2 route افزودن

```text
/provider-accounts/new
  ?provider=openai           provider از قبل انتخاب‌شده
  &source=byok|managed       روش اتصال از قبل انتخاب‌شده
```

### 9.3 جزئیات provider

برای MVP route فعلی حفظ می‌شود:

```text
/provider-accounts?account=<account-id>&section=overview
```

sectionهای قابل deep link:

- `overview`
- `models`
- `access`
- `credentials`

در صورت تصمیم فنی برای route مستقل، redirect دائمی یا canonical mapping لازم است و bookmarkهای فعلی نباید بشکنند.

### 9.4 جزئیات model

```text
/provider-accounts?tab=models&model=<url-encoded-official-model>
```

اگر model روی چند account موجود باشد:

```text
&account=<account-id>
```

drawer باید با Back/Forward مرورگر باز و بسته شود و URL قابل اشتراک باشد.

---

## 10. صفحه‌ی اصلی

### 10.1 header

- عنوان: «کاتالوگ مدل‌ها»
- توضیح: «ارائه‌دهنده‌ها را متصل کنید و مدل‌های قابل‌استفاده‌ی این tenant را مدیریت کنید.»
- اکشن اصلی در toolbar: «افزودن ارائه‌دهنده»
- help link: مستند provider account و BYOK

### 10.2 تب‌ها

- «ارائه‌دهنده‌های AI»
- «مدل‌ها»

تب انتخاب‌شده در query string ثبت می‌شود. focus، keyboard navigation و `aria-selected` الزامی است.

تب‌های فعلی «همه / متصل» حذف نمی‌شوند، بلکه کارکردشان بازتوزیع می‌شود:

- catalog providerهای قابل اتصال در داخل فلو Add Provider و empty state قرار می‌گیرد؛
- accountهای متصل در تب Providers نمایش داده می‌شوند.

---

## 11. تب «ارائه‌دهنده‌های AI»

### 11.1 toolbar

اجزا:

- جست‌وجوی debounced روی name، provider، auth mode و alias؛
- فیلتر چندانتخابی «سرویس AI»؛
- فیلتر چندانتخابی «روش اتصال»: credential اختصاصی / مدیریت‌شده؛
- فیلتر «وضعیت»: آماده / نیازمند credential / نیازمند access / خطای اتصال / قیمت ناقص؛
- ActiveFilterChips؛
- Refresh؛
- Add Provider.

فیلترها در URL ذخیره می‌شوند و reset all دارند.

### 11.2 جدول دسکتاپ

ستون‌ها:

1. **نام**: account name + provider logo/name
2. **شناسه‌ی اتصال**: `provider:account`، قابل‌کپی و با توضیح اینکه FQN کامل نیست
3. **مدل‌های فعال**: count و لینک به تب Models با account filter
4. **روش credential**: BYOK / مدیریت‌شده
5. **آمادگی**: readiness badge
6. **آخرین تغییر**: `updated_at`
7. **عملیات**

در MVP از table استفاده می‌شود، نه grid کارت‌ها؛ چون scan، فیلتر، مقایسه و row actions در تعداد زیاد بهتر است. در mobile هر row به card فشرده تبدیل یا table به‌صورت افقی scroll می‌شود.

### 11.3 عملیات ردیف

ترتیب و grouping:

- «مشاهده‌ی مدل‌ها»
- «جزئیات و ویرایش»
- «تست دسترسی مؤثر»
- «تست اتصال» فقط برای credential قابل‌تست
- «مستندات provider» اگر موجود است
- separator
- «حذف اتصال»

کلیک روی خود ردیف detail page را باز می‌کند؛ منوی عملیات باید `stopPropagation` داشته باشد.

### 11.4 empty stateها

سه حالت مجزا:

1. **tenant هیچ account ندارد**  
   CTA اصلی: Add Provider، CTA دوم: مستند شروع سریع.
2. **فیلتر نتیجه ندارد**  
   CTA: پاک‌کردن فیلترها.
3. **catalog provider در دسترس نیست**  
   پیام خطا + Retry؛ accountهای موجود در صورت امکان همچنان نمایش داده شوند.

### 11.5 loading/error

- header و toolbar ثابت می‌مانند؛
- table skeleton با تعداد ستون واقعی؛
- refresh خطا نباید داده‌ی قبلی را پاک کند؛
- OperationNotice برای خطای operation؛
- inline retry برای خطای initial load.

---

## 12. تب «مدل‌ها»

### 12.1 تعریف availability

یک model در این تب یک projection تجمیعی است، نه entity تازه:

```text
key = canonical(provider, official_model)
availability = accountهای tenant که:
  - provider آنها با model سازگار است؛
  - TenantProviderAccountModel متناظر فعال است؛
  - pricing/config معتبر است.
```

برای مدیر tenant، availability همه‌ی accountها نمایش داده می‌شود. برای کاربر بدون مجوز مدیریت، فقط accountهای مجاز طبق effective access نمایش داده می‌شوند.

### 12.2 toolbar

- جست‌وجو روی display name، official model ID، provider و FQN؛
- فیلتر چندانتخابی «سرویس AI»؛
- فیلتر چندانتخابی «اتصال»؛
- فیلتر «قابلیت»: chat، completion، embedding، image generation، audio، moderation و قابلیت‌های موجود در catalog؛
- فیلتر «ورودی/خروجی»: text، image، audio؛
- فیلتر «وضعیت قیمت»: عمومی، خصوصی، حل‌نشده؛
- sort: نام، context، قیمت ورودی، قیمت خروجی، تعداد اتصال‌های در دسترس؛
- ActiveFilterChips؛
- Refresh.

فیلترهای قابلیت و قیمت می‌توانند P1 باشند؛ جست‌وجو، provider و account در P0 الزامی‌اند.

### 12.3 جدول

ستون‌های P0:

1. **مدل**: display name + official ID قابل‌کپی
2. **سرویس AI**
3. **قابلیت‌ها/Modalities**
4. **دسترسی**: «روی N اتصال فعال»
5. **Context**
6. **قیمت ورودی/خروجی** با واحد انتخابی USD یا تومان

کلیک ردیف drawer را باز می‌کند.

### 12.4 availability interaction

متن «روی N اتصال فعال» popover دارد:

- account name؛
- readiness؛
- credential type؛
- قیمت مؤثر؛
- FQN کامل `provider:account:model`؛
- Copy؛
- انتخاب به‌عنوان account فعال drawer.

اگر فقط یک account موجود باشد همان مورد خودکار انتخاب می‌شود.

### 12.5 empty stateها

1. **هیچ مدل provision نشده**  
   CTA: Add Provider یا Manage Models اولین account.
2. **account وجود دارد ولی مدل فعال نیست**  
   CTA: Manage Models.
3. **جست‌وجو/فیلتر نتیجه ندارد**  
   CTA: Clear filters.
4. **مدل برای principal فعلی مجاز نیست**  
   پیام دسترسی؛ مدیر می‌تواند Effective Access Tester را باز کند.

---

## 13. drawer جزئیات model

### 13.1 header

- display name؛
- official provider؛
- status catalog مانند stable/preview/deprecated؛
- replacement model در صورت deprecated؛
- close button با keyboard support.

### 13.2 انتخاب account

Combobox فقط accountهایی را نشان می‌دهد که model روی آن‌ها فعال و برای principal قابل‌دسترسی است.

هر option:

- account name؛
- readiness؛
- pricing mode؛
- credential type.

انتخاب account در URL ثبت می‌شود.

### 13.3 شناسه و نمونه‌کد

FQN canonical:

```text
provider:account:model
```

مثال:

```text
anthropic:production:claude-3-5-sonnet
```

الزامات:

- هیچ‌جا `provider:account` به‌عنوان model FQN معرفی نشود؛
- FQN از helper مشترک هم‌معنای `modelid.BuildModelFQN` ساخته شود؛
- identifierها LTR و monospace باشند؛
- Copy failure قابل مشاهده باشد.

زبان‌های P0:

- cURL Bash
- PowerShell

زبان‌های P1:

- Python SDK
- Node.js/TypeScript SDK
- OpenAI-compatible Python
- OpenAI-compatible Node.js

API key در نمونه‌کد placeholder است و secret واقعی هرگز در DOM یا clipboard وارد نمی‌شود.

### 13.4 اکشن‌ها

P0:

- Copy FQN
- Copy code
- Open in Playground با model و account از قبل انتخاب‌شده
- View Logs با model/account filter
- Manage Model برای مدیر

P1:

- Run Test Request
- نمایش status، latency، token usage، resolved model و trace ID
- Copy to Integrate

Run Test باید:

- از API key نشست یا مکانیزم امن tenant استفاده کند؛
- قبل از ارسال، endpoint و هزینه‌ی احتمالی را توضیح دهد؛
- timeout و cancel داشته باشد؛
- response خام حساس را log نکند؛
- از مسیر تونل و policyهای runtime واقعی عبور کند تا تست نمایشی نباشد.

### 13.5 metadata

- model ID؛
- FQN؛
- context window؛
- max output؛
- supported modes؛
- input/output modalities؛
- capabilities؛
- pricing public/private و واحد؛
- release date/knowledge cutoff در صورت موجودبودن؛
- deprecation/replacement؛
- regions و rate-limit notes در صورت موجودبودن.

### 13.6 حالت‌ها

- loading drawer skeleton؛
- account unavailable؛
- pricing unresolved؛
- access denied؛
- deprecated warning؛
- connection unhealthy؛
- test request waiting/success/error/cancelled.

---

## 14. فلو «افزودن ارائه‌دهنده»

فلو ظاهری دو مرحله دارد، اما تمام guardrailهای wizard فعلی حفظ می‌شوند.

### 14.1 مرحله 1: انتخاب سرویس و روش اتصال

#### انتخاب سرویس

- searchable provider picker از Provider Catalog فعلی؛
- logo، display name، تعداد model و auth modeهای پشتیبانی‌شده؛
- quick choices برای providerهای محبوب tenant/بازار؛
- providerهای unavailable نمایش داده نمی‌شوند یا disabled با دلیل هستند.

#### انتخاب روش اتصال

به‌جای integration entity در Portkey:

- **Credential مدیریت‌شده توسط شاهراه**
- **Credential اختصاصی tenant (BYOK)**

فقط روش‌های مجاز طبق provider connection schema و entitlement نمایش داده می‌شوند.

اگر provider از قبل account دارد، کاربر می‌تواند:

- account موجود را باز کند؛
- یا account دیگری با نام یکتا بسازد.

### 14.2 مرحله 2: تنظیم و فعال‌سازی

صفحه‌ی مرحله 2 سه section قابل‌جمع‌شدن دارد:

#### الف) جزئیات اتصال

- Name اجباری؛
- preview شناسه‌ی اتصال `provider:account`؛
- auth mode در صورت چند schema؛
- provider-specific config fields؛
- credential write-only برای BYOK؛
- لینک مستند provider؛
- Test Connection.

`name` همان alias runtime موجود است و به slug/entity تازه نیاز ندارد. تغییر نام account موجود در MVP پشتیبانی نمی‌شود، چون بخشی از FQN runtime است و rename می‌تواند requestهای موجود را بشکند.

#### ب) Model Provisioning

- recommendation یک model شروع؛
- search و filter؛
- انتخاب مدل‌ها؛
- capabilityهای مؤثر؛
- public/private pricing؛
- حداقل یک model معتبر برای ادامه.

#### ج) Access

- managerها؛
- userها؛
- access scope account/model؛
- اعتبارسنجی اینکه creator دسترسی مدیریت خود را ناخواسته حذف نکند.

### 14.3 footer

- Back
- Discard
- Create and Activate

Review جداگانه به‌صورت step مستقل حذف می‌شود؛ summary در ستون/پنل کناری sticky نمایش داده می‌شود:

- provider؛
- connection method؛
- account alias؛
- تعداد model؛
- pricing readiness؛
- access summary.

Create تنها وقتی فعال است که:

- name معتبر و در tenant یکتا باشد؛
- credential/schema معتبر باشد یا managed credential مجاز باشد؛
- test اجباری طبق policy موفق شده باشد؛
- حداقل یک model فعال و pricing resolved باشد؛
- access معتبر باشد.

در backend همان ترتیب فعلی account → models → access اجرا می‌شود. اگر API اتمیک نیست، UI باید rollback/repair state روشن داشته باشد و account نیمه‌کاره را به‌عنوان Draft/Needs setup نشان دهد؛ status تازه در DB لازم نیست و از readiness مشتق می‌شود.

---

## 15. صفحه‌ی جزئیات provider

### 15.1 header هویت

- provider logo/name؛
- account name؛
- شناسه‌ی اتصال `provider:account`؛
- readiness؛
- credential type؛
- More actions: Test access، Test connection، Docs، Delete.

### 15.2 navigation درون صفحه

- Overview
- Models
- Access
- Credentials

در desktop می‌تواند sticky anchor/tab باشد؛ در mobile select یا tab scrollable.

### 15.3 Overview

P0:

- readiness با دلیل و CTA اصلاح؛
- credential mode؛
- auth mode؛
- provider config غیرحساس؛
- تعداد مدل فعال؛
- تعداد manager/user؛
- created/updated؛
- نمونه‌کد سریع با اولین model فعال؛
- لینک Playground و Logs.

P1:

- last used؛
- total requests؛
- requests this week/last week؛
- trend؛
- cost؛
- latency/error rate؛
- View Analytics.

متریک‌های P1 از trace/spend موجود محاسبه می‌شوند و ستون تازه‌ای در account لازم ندارند.

### 15.4 Models

از `ModelsEditor` فعلی reuse می‌شود، با تغییر composition:

- search همیشه visible؛
- filterهای بیشتر در popover؛
- «فقط انتخاب‌شده‌ها»؛
- Select/Deselect visible results؛
- compact table برای حجم بالا؛
- model card/detail drawer برای inspect؛
- enable/disable؛
- effective capability selection؛
- public/private pricing؛
- Copy FQN؛
- code sample.

Save bulk explicit است. خروج با تغییر ذخیره‌نشده `UnsavedChangesDialog` دارد.

### 15.5 Access

قابلیت فعلی بدون حذف:

- manager/user؛
- principal picker؛
- account/model scope؛
- effective access tester؛
- دلیل allow/deny.

Access در UI مشابه Portkey افزوده نمی‌شود، بلکه مزیت دامنه‌ای شاهراه است و باید در detail page باقی بماند.

### 15.6 Credentials

- نوع credential؛
- schema-driven fields؛
- write-only secret؛
- credential hint؛
- test connection؛
- provider docs؛
- managed credential حالت read-only.

### 15.7 Delete

به‌دلیل اثر بالا از `TypedConfirmDialog` فعلی استفاده می‌شود، نه checkbox ساده‌ی Portkey.

Dialog باید:

- نام account را برای تایپ نشان دهد؛
- تعداد modelهای فعال را نشان دهد؛
- هشدار دهد FQNها و requestهای وابسته از کار می‌افتند؛
- اگر active API key/reference وجود دارد، conflict backend را با dependency قابل‌اقدام نمایش دهد؛
- در failure باز بماند؛
- در success به تب Providers برگردد و toast/notice بدهد.

---

## 16. منطق UI و state

### 16.1 readiness مشتق‌شده

بدون افزودن field:

```text
connection_failed
  test اخیر ناموفق است یا schema/config ناسازگار است

needs_credentials
  نه credential اختصاصی معتبر است و نه managed credential مجاز

needs_access
  access rule قابل‌استفاده وجود ندارد

pricing_incomplete
  یک model فعال pricing resolved ندارد

ready
  حداقل یک model فعال + credential usable + access usable + pricing resolved
```

اگر test result persist نمی‌شود، `connection_failed` فقط session-local است و بعد از reload به `unknown/not recently tested` برمی‌گردد. persist کردن health یک تصمیم P1 جداست و entity تازه نمی‌خواهد، اما قرارداد و storage مشخص می‌خواهد.

### 16.2 permissions

- مشاهده‌ی catalog عمومی provider/model طبق مجوز tenant؛
- مشاهده‌ی account/model availability طبق effective access؛
- create/edit/delete/model provisioning فقط manager مجاز؛
- credential view/update فقط role مجاز و بدون secret readback؛
- Logs/Analytics طبق permission همان محصولات؛
- UI هرگز تنها guard نباشد؛ backend source of truth است.

### 16.3 URL state

در URL:

- active tab؛
- search؛
- filterها؛
- sort؛
- selected provider account؛
- selected model drawer؛
- selected account داخل drawer.

در local/session state:

- draft form؛
- unsaved pricing/model selection؛
- test request response؛
- transient notices.

### 16.4 refresh

- manual refresh داده‌ی قبلی را حفظ می‌کند؛
- catalog refresh با account refresh اشتباه گرفته نمی‌شود؛
- cache Provider Catalog می‌تواند stale-while-revalidate باشد؛
- mutation موفق فقط queryهای مرتبط را invalidate می‌کند.

---

## 17. قرارداد داده و API

### 17.1 APIهای موجود و قابل‌استفاده در P0

- `GET /api/v1/provider-catalog/providers`
- `GET /api/v1/provider-catalog/providers/:provider/models`
- connection schema و display settings موجود
- `POST /api/v1/tenant-provider-accounts`
- `GET /api/v1/tenant-provider-accounts`
- `GET /api/v1/tenant-provider-accounts/:accountID`
- `DELETE /api/v1/tenant-provider-accounts/:accountID`
- `PUT /api/v1/tenant-provider-accounts/:accountID/models`
- `PUT /api/v1/tenant-provider-accounts/:accountID/access`
- `PUT /api/v1/tenant-provider-accounts/:accountID/credentials`
- `POST /api/v1/tenant-provider-accounts/:accountID/test-credentials`
- `GET /api/v1/tenant-provider-accounts/:accountID/effective-models`

### 17.2 projection پیشنهادی برای تب Models

برای جلوگیری از N+1، یک read projection اضافه شود؛ این API entity یا table تازه نیست:

```http
GET /api/v1/tenant-provider-account-models
  ?q=
  &provider=
  &account_id=
  &capability=
  &pricing_status=
  &page=
  &page_size=
  &sort=
```

پاسخ پیشنهادی:

```json
{
  "data": [
    {
      "provider": "anthropic",
      "model": "claude-3-5-sonnet",
      "display_name": "Claude 3.5 Sonnet",
      "catalog": {
        "context_window": 200000,
        "supported_modes": ["chat"],
        "input_modalities": ["text", "image"],
        "output_modalities": ["text"]
      },
      "availability": [
        {
          "account_id": "uuid",
          "account_name": "production",
          "fqn": "anthropic:production:claude-3-5-sonnet",
          "pricing_mode": "public",
          "pricing_resolved": true,
          "readiness": "ready"
        }
      ]
    }
  ],
  "page": 1,
  "page_size": 25,
  "total": 1
}
```

قواعد:

- projection با join/read از جدول‌های موجود و Provider Catalog ساخته می‌شود؛
- credential data هرگز در پاسخ نیست؛
- tenant و effective access در query enforce می‌شود؛
- FQN با helper backend ساخته می‌شود، نه string concatenation پراکنده در frontend؛
- official model ID canonical می‌شود؛
- pagination server-side است.

اگر API جدید در MVP پذیرفته نشود، frontend می‌تواند برای تعداد account محدود از detail endpointها compose کند، اما این فقط راه‌حل موقت با سقف مشخص است و برای production scale توصیه نمی‌شود.

### 17.3 بهبود list account

بدون تغییر entity، response فهرست account می‌تواند summary مشتق‌شده داشته باشد:

```json
{
  "model_count": 12,
  "manager_count": 2,
  "user_count": 8,
  "readiness": "ready",
  "pricing_unresolved_count": 0
}
```

این فیلدها optional و backward-compatible باشند.

### 17.4 update account name

قرارداد فعلی update برای name ندارد. چون name بخشی از FQN است:

- P0: rename وجود ندارد؛
- UI به‌جای Edit Name، تنظیمات قابل‌تغییر را edit می‌کند؛
- P1/P2: rename فقط با impact analysis، alias migration و grace period بررسی شود.

### 17.5 custom models

handler فعلی model selection را با catalog validate می‌کند. بنابراین:

- P0/P1: فقط catalog modelها؛
- P2: همان `TenantProviderAccountModel` می‌تواند custom/fine-tuned را نگه دارد، اما validation، base model، pricing، custom host و امنیت SSRF باید contract مشخص داشته باشد؛
- table/entity تازه پیش‌فرض این قابلیت نیست.

---

## 18. نیازمندی‌های عملکردی

### P0 — الزامی

#### MC-P0-01: دو تب اصلی

صفحه دو تب Providers و Models داشته باشد و state آن در URL ثبت شود.

#### MC-P0-02: Providers list

تمام accountهای tenant با search، provider filter، credential filter، readiness و row actions نمایش داده شوند.

#### MC-P0-03: View Models

از هر provider بتوان با یک اکشن به Models tab با account filter فعال رفت.

#### MC-P0-04: Models aggregation

modelهای فعال همه‌ی accountها به‌شکل deduplicated و با availability count نمایش داده شوند.

#### MC-P0-05: Model drawer

انتخاب model drawer قابل deep link با metadata، account selection، FQN، code، Playground و Logs باز کند.

#### MC-P0-06: FQN صحیح

تمام model identifierهای قابل‌کپی مطابق `provider:account:model` باشند. `provider:account` فقط «شناسه‌ی اتصال» نامیده شود.

#### MC-P0-07: Add Provider دو مرحله‌ای

انتخاب provider/credential method و سپس details/models/access در یک مرحله‌ی composition شده انجام شود.

#### MC-P0-08: حفظ capabilityهای فعلی

BYOK/managed credential، schema-driven config، test credential، model provisioning، pricing، access و effective access tester حذف نشوند.

#### MC-P0-09: Provider detail

Overview، Models، Access و Credentials با deep link و unsaved-change protection در دسترس باشند.

#### MC-P0-10: حذف امن

Typed confirmation، dependency conflict و feedback نتیجه پیاده شود.

#### MC-P0-11: permissions

تمام read/writeها tenant-scoped و role-enforced باشند.

#### MC-P0-12: وضعیت‌ها

loading، empty، no-results، partial error، operation error و permission denied طراحی مجزا داشته باشند.

### P1 — بعد از MVP

#### MC-P1-01: Analytics summary

last used، requests، cost، latency/error و trend در provider overview.

#### MC-P1-02: Inline test request

اجرای request واقعی از model drawer با telemetry نتیجه.

#### MC-P1-03: SDK examples

Python، Node.js و OpenAI-compatible samples.

#### MC-P1-04: فیلترهای پیشرفته

capabilities، modalities، pricing، context و sortهای بیشتر.

#### MC-P1-05: Bulk provisioning

select/deselect visible results و save اتمیک.

#### MC-P1-06: Health persistence

last tested و نتیجه‌ی سلامت با retention مشخص.

### P2 — مشروط

#### MC-P2-01: Custom/fine-tuned model

پس از threat model برای custom host و تغییر catalog validation.

#### MC-P2-02: Advanced connection overrides

custom host/header و pricing adjustment فقط با audit، secret handling و SSRF controls.

#### MC-P2-03: Auto-enable new models

فقط با policy روشن درباره‌ی هزینه، capability و دسترسی.

#### MC-P2-04: Recommendation

پیشنهاد model براساس قیمت، کیفیت، latency، region و availability.

---

## 19. نیازمندی‌های غیرعملکردی

### 19.1 کارایی

- P95 بارگذاری shell و داده‌ی اولیه‌ی هر تب در شبکه‌ی معمول ایران: حداکثر 2.5 ثانیه؛
- تغییر tab از cache محلی: حداکثر 200ms تا feedback؛
- search debounce: 250–350ms؛
- server pagination پیش‌فرض 25 و حداکثر 100؛
- تب Models نباید N+1 بدون سقف ایجاد کند؛
- drawer metadata پس از بازشدن حداکثر 1 ثانیه تا skeleton replacement.

### 19.2 پایداری شبکه

- stale data هنگام refresh حفظ شود؛
- mutationها idempotency/duplicate-submit guard داشته باشند؛
- timeout، retry کنترل‌شده و پیام فارسی قابل‌اقدام؛
- قطع تونل ایران–اروپا به‌عنوان connection/runtime error از خطای credential تفکیک شود.

### 19.3 امنیت

- secret readback ممنوع؛
- credential fieldها write-only؛
- test request و credential test audit شوند؛
- custom URL/host در صورت اضافه‌شدن SSRF validation داشته باشد؛
- access در backend enforce شود؛
- log و analytics داده‌ی حساس prompt/credential را در این flow expose نکنند؛
- حذف account قبل از اجرا dependency check داشته باشد.

### 19.4 دسترس‌پذیری

- keyboard navigation کامل tab/table/menu/drawer/dialog؛
- focus trap drawer/dialog؛
- focus restoration بعد از close؛
- accessible name برای icon buttonها؛
- status فقط با رنگ منتقل نشود؛
- table header و empty row semantic؛
- contrast مطابق الگوی موجود.

### 19.5 RTL و i18n

- layout و alignment RTL؛
- identifier/code/URL/email همواره LTR؛
- نام‌های فارسی با `DisplayNameField`؛
- اعداد، زمان و پول localizable؛
- تومان واحد نمایشی اصلی قابل‌انتخاب و USD قابل مشاهده؛
- متن UI از message catalog، نه hard-code.

### 19.6 design system

reuse:

- `PageContainer`
- `DataListPanel`
- `DebouncedTextFilter`
- `FilterCombobox`
- `ActiveFilterChips`
- `RowActionsMenu`
- `StatusBadge`
- `ModelNameValue`
- `CopyAction` / `CopyableValue`
- `EntityIdentityHeader`
- `DefinitionList`
- `MetricCard`
- `CodePanel`
- `FormSection` / `FormDialog` / `FormSheet`
- `TypedConfirmDialog`
- `UnsavedChangesDialog`
- `TableLoadingRows` / `TableEmptyRow`
- `OperationNotice`

pattern reusable تازه فقط در صورت تکرار واقعی ساخته و در playground طراحی ثبت شود. semantic tokenها و RTL حفظ شوند؛ raw color/arbitrary value جدید پذیرفته نیست.

---

## 20. تصمیم‌های Include / Defer / Exclude

| قابلیت | تصمیم | دلیل |
|---|---|---|
| تب Providers و Models | Include P0 | هسته‌ی مدل ذهنی Portkey |
| جست‌وجو و multi-filter | Include P0 | کشف سریع و مقیاس‌پذیری |
| availability روی چند account | Include P0 | ارزش اصلی نمای Models |
| model detail drawer | Include P0 | اتصال کشف به استفاده |
| FQN کامل | Include P0 | اصلاح خطای عملیاتی فعلی |
| Add Provider دو مرحله‌ای | Include P0 | کاهش بار شناختی |
| Access و Effective Access | Include P0 | قابلیت ضروری و متمایز شاهراه |
| BYOK / managed credential | Include P0 | قرارداد دامنه‌ی فعلی |
| قیمت تومان/USD | Include P0 | نیاز بازار ایران |
| Analytics provider | Defer P1 | نیازمند aggregation trace/spend |
| Inline test request | Defer P1 | هزینه و امنیت بیشتر از credential test |
| SDK samples متعدد | Defer P1 | cURL/PowerShell برای MVP کافی است |
| health persistence | Defer P1 | نیازمند contract retention |
| custom/fine-tuned model | Defer P2 | catalog validation و SSRF |
| custom headers/host | Defer P2 | امنیت و audit |
| auto-enable modelهای جدید | Defer P2 | ریسک هزینه و دسترسی |
| entity مستقل Integration | Exclude | تکرار مسئولیت account |
| entity تازه Model Catalog | Exclude | catalog و account model موجودند |
| تغییر FQN به `@slug/model` | Exclude | ناسازگار با runtime فعلی |
| rename آزاد account | Exclude P0 | شکستن FQNهای موجود |
| نمایش secret واقعی | Exclude | ریسک امنیتی |

---

## 21. MVP cutline

MVP کامل است وقتی:

1. route فعلی با دو تب Providers/Models عرضه شود؛
2. Providers list table و تمام عملیات فعلی کار کند؛
3. Models tab projection تجمیعی، search و provider/account filter داشته باشد؛
4. model drawer با account selection، FQN صحیح، cURL/PowerShell، Playground و Logs کار کند؛
5. Add Provider دو مرحله‌ای تمام validationهای فعلی را پوشش دهد؛
6. provider detail چهار section Overview/Models/Access/Credentials را داشته باشد؛
7. delete conflict و typed confirmation امن باشد؛
8. تست‌های permission، RTL، keyboard و empty/error state پاس شوند؛
9. هیچ migration/entity تازه‌ای اضافه نشده باشد.

Analytics، inline inference test، SDKهای بیشتر، custom model و advanced connection options جزو MVP نیستند.

---

## 22. معیارهای موفقیت

### محصول

- حداقل 30٪ کاهش median زمان از ورود به صفحه تا Copy FQN یک model؛
- حداقل 25٪ کاهش رهاکردن فلو ایجاد provider؛
- حداقل 40٪ کاهش navigation بین Provider Accounts و Playground/Logs؛
- حداقل 80٪ create flowهای شروع‌شده بدون نیاز به بازگشت به مرحله‌ی provider picker کامل شوند؛
- کاهش ticketهای «model identifier نامعتبر» مرتبط با FQN ناقص.

### عملیاتی

- کمتر از 1٪ خطای frontend برای load دو تب؛
- کمتر از 0.5٪ duplicate mutation؛
- صفر exposure credential در telemetry/DOM snapshotهای کنترل‌شده؛
- P95 projection تب Models مطابق هدف کارایی.

### UX research

کاربر در تست usability بدون راهنمایی بتواند:

1. provider متصل کند؛
2. یک model را فعال کند؛
3. FQN کامل را کپی کند؛
4. همان model را در Playground باز کند؛
5. دلیل عدم دسترسی یک principal را پیدا کند.

---

## 23. instrumentation

eventهای پیشنهادی بدون داده‌ی حساس:

- `model_catalog_tab_viewed`
- `model_catalog_filter_applied`
- `model_catalog_search_used`
- `provider_add_started`
- `provider_add_method_selected`
- `provider_add_completed`
- `provider_add_failed`
- `provider_details_viewed`
- `provider_model_provisioning_saved`
- `model_details_opened`
- `model_account_selected`
- `model_fqn_copied`
- `model_code_copied`
- `model_opened_in_playground`
- `model_logs_opened`
- `provider_connection_tested`
- `provider_effective_access_tested`
- `provider_delete_blocked`
- `provider_deleted`

ممنوع:

- API key؛
- credential config حساس؛
- prompt/response کامل؛
- principal PII بدون نیاز؛
- raw provider error حاوی secret.

---

## 24. وابستگی‌ها

- صحت و refresh Provider Catalog؛
- endpointهای TenantProviderAccount فعلی؛
- permission و effective access؛
- exchange rate برای نمایش تومان؛
- Playground با preselection از query؛
- Logs با model/account filter؛
- trace/spend aggregation برای P1؛
- docs routeهای provider و BYOK؛
- entitlement managed credentials؛
- تونل inference ایران–اروپا برای test request واقعی.

---

## 25. ریسک‌ها و کاهش ریسک

### ریسک 1: اشتباه‌گرفتن provider family با account

**کاهش:** در copy فارسی همواره «سرویس AI» برای family و «اتصال» برای account استفاده شود. Tooltip کوتاه مدل دامنه را توضیح دهد.

### ریسک 2: سنگین‌شدن تب Models

**کاهش:** read projection server-side، pagination، filter، cache و عدم N+1.

### ریسک 3: create flow دو مرحله‌ای بیش‌ازحد متراکم شود

**کاهش:** sectionهای progressive disclosure، summary sticky، validation محلی و حفظ draft.

### ریسک 4: readiness بدون health persisted گمراه‌کننده باشد

**کاهش:** «تست‌نشده/آخرین تست این نشست» صریح نمایش داده شود؛ status قطعی جعل نشود.

### ریسک 5: تغییر نام و FQN

**کاهش:** rename در P0 حذف و account alias هنگام ساخت شفاف شود.

### ریسک 6: نمایش قیمت اشتباه در ایران

**کاهش:** واحد، نرخ تبدیل و timestamp نرخ آشکار؛ public/private pricing تفکیک؛ unresolved هرگز صفر نشان داده نشود.

### ریسک 7: حذف account وابسته

**کاهش:** dependency check backend، typed confirmation و لینک به منابع وابسته.

### ریسک 8: تقلید advanced options Portkey بدون کنترل امنیتی

**کاهش:** custom host/header و custom model تا threat model و audit به P2 موکول شوند.

---

## 26. معیار پذیرش

### Providers

- [ ] تب Providers پیش‌فرض است و URL آن قابل share است.
- [ ] search حداقل name/provider/auth/connection identifier را پیدا می‌کند.
- [ ] provider، credential و readiness filter قابل ترکیب‌اند.
- [ ] View Models تب Models را با account filter باز می‌کند.
- [ ] row actions با keyboard قابل استفاده‌اند.
- [ ] account بدون credential یا model حالت readiness درست دارد.
- [ ] delete نیازمند تایپ نام دقیق است.
- [ ] conflict حذف active reference پیام قابل‌اقدام دارد.

### Models

- [ ] مدل تکراری روی چند account یک ردیف با availability count دارد.
- [ ] accountهای غیرمجاز در availability کاربر عادی نمایش داده نمی‌شوند.
- [ ] search official ID و FQN را پیدا می‌کند.
- [ ] انتخاب ردیف drawer و URL را هم‌زمان به‌روز می‌کند.
- [ ] Back drawer را می‌بندد و filter state را حفظ می‌کند.
- [ ] FQN کپی‌شده دقیقاً `provider:account:model` و runtime-valid است.
- [ ] قیمت unresolved با `—/نامشخص` نمایش داده می‌شود، نه صفر.
- [ ] تومان و USD unit واضح دارند.

### Model drawer

- [ ] اگر یک account موجود است خودکار انتخاب می‌شود.
- [ ] اگر چند account موجود است انتخاب کاربر در URL ثبت می‌شود.
- [ ] code sample از FQN همان account استفاده می‌کند.
- [ ] secret واقعی در نمونه‌کد نیست.
- [ ] Playground با model/account درست preselect می‌شود.
- [ ] Logs filter درست دارد.
- [ ] deprecated model warning و replacement نمایش داده می‌شود.

### Add Provider

- [ ] provider picker فقط از catalog موجود تغذیه می‌شود.
- [ ] روش BYOK/managed براساس schema/entitlement محدود می‌شود.
- [ ] duplicate account name قبل از submit پیام روشن دارد.
- [ ] credential test error با network/tunnel error اشتباه نمی‌شود.
- [ ] حداقل یک model و access معتبر لازم است.
- [ ] submit چندباره account تکراری نمی‌سازد.
- [ ] شکست مرحله‌ی models/access account نیمه‌کاره را با recovery CTA نشان می‌دهد.

### Provider detail

- [ ] Overview/Models/Access/Credentials deep link دارند.
- [ ] unsaved changes هنگام navigation محافظت می‌شود.
- [ ] model provisioning قابلیت فعلی pricing/capability را حفظ می‌کند.
- [ ] effective access tester همچنان در دسترس است.
- [ ] managed credential قابل reveal/edit نیست.

### کیفیت

- [ ] RTL در desktop/mobile و identifierهای LTR بررسی شده‌اند.
- [ ] focus order، focus trap و focus restoration پاس می‌شوند.
- [ ] loading/empty/error/no-results جدا هستند.
- [ ] tenant isolation و role enforcement تست integration دارند.
- [ ] lint، typecheck، unit test و build با `pnpm` پاس می‌شوند.
- [ ] pattern reusable تازه در design-system playground ثبت شده است.

---

## 27. طرح تحویل مرحله‌ای

### فاز A — shell و read model

- tab/URL architecture؛
- Providers table؛
- projection API تب Models؛
- Models table و filters پایه؛
- اصلاح helper مشترک FQN.

### فاز B — detail surfaces

- model drawer؛
- Playground/Logs deep links؛
- provider detail composition؛
- stateهای loading/error/permission.

### فاز C — create/manage

- Add Provider دو مرحله‌ای؛
- reuse model/access/credential editors؛
- delete/dependency UX؛
- telemetry.

### فاز D — hardening

- performance و network resilience؛
- RTL/a11y؛
- permission/integration tests؛
- migration redirects و docs؛
- design-system playground برای patternهای واقعاً reusable.

هر فاز باید feature flag مستقل داشته باشد. تب Models می‌تواند ابتدا read-only عرضه شود؛ اما تا قبل از عرضه‌ی عمومی، FQN، Playground و Logs باید end-to-end معتبر باشند.

---

## 28. پرسش‌های باز برای تصویب محصول

این پرسش‌ها اجرای MVP را متوقف نمی‌کنند مگر خلاف آن تصمیم گرفته شود:

1. آیا عنوان navigation «کاتالوگ مدل‌ها» باشد یا «ارائه‌دهنده‌ها و مدل‌ها»؟
2. آیا کاربر عادی مدل‌های فعال اما غیرمجاز را با حالت lock می‌بیند یا اصلاً نمی‌بیند؟ پیشنهاد: پیش‌فرض عدم نمایش، مدیر نمای کامل.
3. آیا Create نیازمند موفقیت اجباری credential test است یا مدیر می‌تواند با هشدار ادامه دهد؟ پیشنهاد: BYOK اجباری، managed credential بدون test دستی.
4. حد مجاز frontend composition در نبود projection API چند account است؟ پیشنهاد: حداکثر 10، سپس الزام endpoint.
5. نام account در آینده rename می‌شود؟ پیشنهاد: فقط پس از طراحی alias migration.
6. test request از drawer در P1 با PAT کاربر اجرا شود یا service credential داخلی؟ پیشنهاد: مکانیزم runtime عادی tenant با audit روشن.

---

## 29. تصمیم نهایی پیشنهادی

این PRD پیشنهاد می‌کند:

- معماری اطلاعات Portkey اقتباس شود؛
- دامنه‌ی Shahrah حفظ شود؛
- `TenantProviderAccount` در UI نقش AI Provider متصل را بازی کند؛
- `TenantProviderAccountModel` منبع availability باشد؛
- نمای تجمیعی Models به‌عنوان read projection ساخته شود؛
- FQN canonical موجود شاهراه محور تمام copy/code/playground/log flowها باشد؛
- قابلیت‌های دسترسی، BYOK، managed credential و قیمت ریالی شاهراه به تجربه‌ی Portkey افزوده شوند؛
- entity و migration تازه در MVP ساخته نشود.

نتیجه‌ی مورد انتظار یک تجربه‌ی catalog-first است: کاربر از «چه providerهایی دارم؟» و «چه modelهایی واقعاً قابل‌استفاده‌اند؟» شروع می‌کند و بدون پرش مفهومی به تنظیم، تست، کدنویسی، Playground و عیب‌یابی می‌رسد.
