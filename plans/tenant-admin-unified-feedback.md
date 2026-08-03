# ممیزی و پلن یکپارچه‌سازی بازخورد کاربر در tenant-admin

## دامنه و روش ممیزی

این ممیزی تمام استفاده‌های مستقیم از `Alert`، `OperationNotice`، `sonner`،
ناحیه‌های زنده (`role="alert"` و `role="status"`)، پیام‌های عملیات CRUD، خطاهای
بارگذاری، بازخورد داخل Dialog/Sheet و تأییدهای مخرب را در `apps/tenant-admin`
پوشش می‌دهد. خطاهای سطح فیلد نیز به‌عنوان یک سناریوی مستقل ثبت شده‌اند، اما
نباید به Toast تبدیل شوند.

یافتهٔ اصلی: `Toaster` در `app/layout.tsx` نصب شده است، ولی پیش از این تغییر هیچ
فراخوانی‌ای از API مربوط به `toast` در کد محصول وجود نداشت. در عوض، پیام‌های
کوتاه عملیات و خطاهای پایدار بارگذاری، هر دو غالباً در یک state مشترک و در بالای
صفحه رندر می‌شدند؛ بنابراین پیام عملیاتی ممکن بود خارج از viewport باشد و خطای
مسدودکننده نیز گاهی مانند یک پیام موقتی رفتار می‌کرد.

## قرارداد UX پیشنهادی

| وضعیت | مؤلفه | رفتار مورد انتظار |
| --- | --- | --- |
| نتیجهٔ کوتاه و غیرمسدودکنندهٔ عملیات، مانند ذخیره، حذف، کپی یا اتصال | `Toast` | مستقل از scroll، یک‌بار نمایش داده شود و تمرکز را جابه‌جا نکند. |
| خطای عملیات صفحه‌ای که کاربر می‌تواند همان عملیات را دوباره اجرا کند | `Toast` | متن دقیق و قابل اقدام؛ در صورت وجود retry/undo، action داخل Toast. |
| خطای بارگذاری که محتوا یا تصمیم بعدی را مسدود می‌کند | `Alert` | نزدیک محتوای ناموفق، پایدار و همراه با retry در صورت امکان. |
| هشدار یا راهنمای ماندگار دامنه، مانند منسوخ‌شدن مدل یا محدودیت امنیتی | `Alert` | در همان context باقی بماند و صرفاً به رنگ متکی نباشد. |
| اعتبارسنجی فرم | `FieldError` یا `Alert` داخل همان فرم | نزدیک فیلد/فرم؛ هرگز Toast تنها نباشد. |
| نتیجهٔ عملیات در Dialog/Sheet باز | `OperationNotice` داخل همان overlay | overlay در خطا باز بماند؛ پیام پشت overlay رندر نشود. |
| تأیید مخرب یا پرریسک ساده | `AlertDialog` از طریق `ConfirmDialog` | تمرکز modal، لغو صریح و اقدام مخرب صریح. |
| تأیید پرریسک همراه با ورود شناسه | `TypedConfirmDialog` | به‌دلیل داشتن فرم ورودی، Dialog باقی می‌ماند و خطا داخل آن نمایش داده می‌شود. |
| اعلان دسترس‌پذیری کپی/حالت کوتاه در همان کنترل | ناحیهٔ زندهٔ محلی | برای کنترل‌های کوچک مانند `CopyAction` حفظ شود؛ Toast فقط در صورت نیاز محصولی اضافه شود. |

## جدول وضعیت فعلی و مقصد هر سناریو

| بخش / فرایند | نوع پیام | الگوی فعلی | مؤلفهٔ فعلی | مشکل UX | الگوی مقصد |
| --- | --- | --- | --- | --- | --- |
| پوستهٔ سراسری برنامه | زیرساخت | Toaster بدون مصرف‌کننده | `Toaster` در `app/layout.tsx` | زیرساخت نصب است ولی هیچ API مشترک و هیچ Toast عملیاتی وجود ندارد. | API متمرکز Toast با tokenهای معنایی، RTL و تست. |
| Alert پایه | همهٔ toneها | همهٔ Alertها `role="alert"` دارند | `components/ui/alert.tsx` | موفقیت/اطلاع‌رسانی غیرضروری به‌صورت assertive اعلام می‌شود. | نقش بر اساس context؛ `OperationNotice` خطا را assertive و سایر toneها را polite اعلام کند. |
| ورود | اعتبارسنجی و خطای احراز هویت | پیام نزدیک فرم | `Alert` و متن خطای فیلد | context درست است؛ تبدیل به Toast باعث دورشدن پیام از فیلد می‌شود. | حفظ Alert/FieldError داخل فرم. |
| پذیرش دعوت | وضعیت در حال پردازش و نتیجه | status محلی و Alert نتیجه | `role="status"` و `Alert` | در صفحهٔ focused مناسب است. | حفظ؛ tone موفقیت/خطا و live-region اصلاح‌شده. |
| داشبورد tenant | خطای بارگذاری | پنل بالای محتوا | `Alert` | خطا محتوا را مسدود می‌کند و باید پایدار باشد. | Alert پایدار نزدیک داده. |
| فهرست Trace | خطای بارگذاری/فیلتر | پنل بالای جدول | `Alert` | خطای داده است، نه نتیجهٔ کوتاه عملیات. | Alert پایدار با retry. |
| داشبورد متریک مدل | خطای بارگذاری | پنل بالای صفحه | `Alert` | مسدودکنندهٔ داده است. | Alert پایدار. |
| مرکز فرمان Log | بارگذاری و عملیات Saved View | یک state خطا برای صفحه و Dialog | `Alert` بالای صفحه | خطای ذخیره/حذف view پشت Dialog دیده نمی‌شود و با خطای load مخلوط است. | خطای load در Alert؛ ذخیره/حذف در overlay یا Toast پس از بسته‌شدن. |
| جزئیات Log | خطای صفحهٔ server-rendered | paragraph دست‌ساز | markup سفارشی | با design system یکسان نیست و semantics صریح ندارد. | `Alert` استاندارد. |
| تیم‌ها: بارگذاری فهرست/جزئیات | خطا | `OperationNotice` بالای صفحه | Alert محصولی | باید پایدار باشد، اما با پیام عملیات مشترک است. | state مستقل و Alert پایدار. |
| تیم‌ها: ایجاد/ویرایش/عضو/حذف | موفقیت و خطای عملیات | banner بالای صفحه | `OperationNotice` | خارج viewport؛ نتیجهٔ عملیات کوتاه است. | Toast؛ خطای فرم باز داخل Sheet. |
| تیم‌ها: سطح دسترسی مدیر | توضیح پایدار | Alert داخل Sheet | `Alert` | context و ماندگاری درست است. | حفظ Alert. |
| تیم‌ها: حذف دارای اثر | هشدار داخل تأیید تایپی | Alert داخل Dialog | `TypedConfirmDialog` + `Alert` | رفتار درست است. | حفظ؛ خطای async نیز داخل Dialog. |
| واحدهای سازمانی: بارگذاری/usage | خطا | banner صفحه و خطای جزئیات | `OperationNotice`/متن | state عملیات و load مخلوط است. | Alert پایدار در context مربوط. |
| واحدهای سازمانی: CRUD و اعضا | موفقیت/خطای عملیات | banner بالای صفحه | `OperationNotice` | ممکن است خارج viewport باشد. | Toast؛ validation داخل Sheet. |
| PAT: بارگذاری | خطا | notice بالای صفحه | `OperationNotice` | خطای مسدودکننده با نتیجهٔ عملیات مخلوط است. | Alert پایدار. |
| PAT: ایجاد/لغو/لغو همه | موفقیت/خطای عملیات | notice بالای صفحه | `OperationNotice` | بازخورد موقتی و خارج viewport. | Toast؛ خطای ایجاد داخل Sheet تا زمان بسته‌شدن. |
| PAT: نمایش secret یک‌بارمصرف | هشدار و خطای copy | Alert داخل Dialog و banner صفحه | `OneTimeSecretDialog`/notice | هشدار context درست دارد؛ خطای copy ممکن است پشت Dialog باشد. | هشدار داخل Dialog؛ خطای copy به Toast یا notice داخل Dialog. |
| کاربر مجازی و token | بارگذاری و CRUD | یک banner برای همه | `Alert` بالای صفحه | پیام‌های عملیات از context دور و load با action مخلوط است. | load در Alert؛ نتیجهٔ CRUD در Toast؛ validation داخل Dialog/Sheet. |
| مدل مجازی | load/detail/validation/CRUD/copy | یک notice داخل پنل جزئیات | `Alert` | موفقیت پس از بسته‌شدن پنل ممکن است دیده نشود؛ load باید پایدار بماند. | load/detail در Alert؛ عملیات بسته‌شده و copy در Toast؛ validation داخل فرم. |
| کاربران و دسترسی | load، دعوت، نقش | banner صفحه + خطای Dialog | `Alert` صفحه و `Alert` داخل Dialog | خطای دعوت داخل Dialog درست است؛ موفقیت و خطاهای صفحه خارج viewport هستند. | خطای Dialog حفظ؛ عملیات صفحه Toast؛ لینک دعوت به‌عنوان Alert پایدار تا زمان مصرف. |
| نقش‌ها: فهرست | خطای load | Alert با retry | `Alert` | درست و قابل اقدام است. | حفظ Alert. |
| نقش‌ها: ایجاد/ویرایش | validation و submit | Alert داخل فرم | `Alert` | context درست است؛ برخی متن‌ها hardcoded هستند. | حفظ Alert/FieldError و انتقال متن hardcoded به i18n. |
| اتصال‌های LLM / حساب provider | load، ایجاد، قطع اتصال، جست‌وجو | banner در بالای viewport داخلی | `Alert` | یک state برای load، redirect-success، جست‌وجو و حذف؛ خطای overlay ممکن است دور باشد. | load در Alert؛ redirect/قطع اتصال Toast؛ جست‌وجوی overlay داخل overlay. |
| wizard حساب provider: ایجاد و ویرایش | validation، ذخیره، تست، فعال‌سازی | Alertهای متعدد در صفحه و فرم | `Alert` و `OperationNotice` | برخی پیام‌ها در context درست‌اند، banner عمومی برای چند سطح استفاده می‌شود. | validation/schema/test داخل مرحله یا Dialog؛ موفقیت ذخیره/کپی Toast؛ load Alert. |
| تست اتصال provider | موفقیت/خطا در Dialog | notice داخل Dialog | `OperationNotice` | context درست است و باید تا تصمیم بعدی باقی بماند. | حفظ OperationNotice داخل Dialog. |
| تست دسترسی مؤثر | خطای اجرای تست | notice در پنل ابزار | `OperationNotice` | نتیجه بخشی از خروجی ابزار است، نه notification عمومی. | حفظ OperationNotice. |
| کاتالوگ مدل | load و عملیات account | notice بالای صفحه | `OperationNotice` | load و action مخلوط هستند. | load Alert؛ عملیات Toast. |
| جزئیات مدل در Drawer/Sheet | deprecation | notice داخل Sheet | `OperationNotice` | هشدار دامنه‌ای و ماندگار است. | Alert/OperationNotice داخل Sheet حفظ شود. |
| Playground: بارگذاری مدل | خطا و unavailable | Alert و empty/error surface | `Alert`/markup context | محتوا را مسدود می‌کند. | Alert یا error state پایدار. |
| Playground: validation اجرا | خطا | notice بالای workbench | `OperationNotice` | validation باید نزدیک ستون/تنظیم مشکل‌دار باشد؛ Toast تنها کافی نیست. | Alert محلی و focus به context؛ Toast فقط برای نتایج کلی. |
| Playground: ذخیره prompt و copy | موفقیت/خطا | notice بالای workbench | `OperationNotice` | پیام کوتاه و خارج viewport. | Toast. |
| Playground: نتیجه هر ستون | خطای اجرا/تصویر | Alert داخل ستون | `Alert` | context صحیح و بخشی از خروجی اجراست. | حفظ Alert. |
| Playground Settings Sheet | conflict و safety hint | Alert داخل Sheet | `Alert` | راهنمای ماندگار و قابل اقدام است. | حفظ Alert. |
| Quick Start: ساخت PAT | پیام صفحه یا lab | Alert بالای هر بخش | `Alert` | حالت Dialog و حالت lab با هم مخلوط‌اند. | خطای Dialog داخل Dialog؛ موفقیت بسته‌شدن Toast؛ خروجی lab داخل lab. |
| Quick Start: اجرای درخواست نمونه | موفقیت/خطا در lab | Alert داخل lab | `Alert` | بخشی از خروجی آموزشی است. | حفظ Alert. |
| Prompt Registry list | load + delete | دو OperationNotice پشت سر هم | `OperationNotice` | تکرار بصری و اختلاط خطای پایدار با نتیجهٔ حذف. | load Alert؛ حذف Toast. |
| Prompt detail/builder | load، validation، save/copy | Alert در header/form | `Alert` | load/validation باید بماند؛ save/copy موقتی است. | load و validation Alert؛ save/copy Toast. |
| Guardrail list | load + save/delete + validation Sheet | یک notice داخل Card صفحه | `OperationNotice` | خطای فرم پشت Sheet نمایش داده می‌شود. | load Alert؛ validation/save-error داخل Sheet؛ success/delete Toast. |
| Rate/Budget policy list | load/delete | notice در PageContainer | `Alert` | load و delete مخلوط است. | load Alert؛ delete Toast؛ خطای confirm داخل Dialog. |
| Rate/Budget rule form | save | Alert داخل فرم embedded/page | `Alert` | خطا در Sheet درست است؛ success بلافاصله با بسته‌شدن Sheet ناپدید می‌شود. | خطا داخل فرم؛ success Toast پس از `onSaved`. |
| صفحهٔ edit policy | load rule | Alert | `Alert` | مسدودکننده و پایدار است؛ fallback انگلیسی hardcoded وجود دارد. | حفظ Alert و تکمیل i18n. |
| تنظیمات privacy و request logging | ذخیره و خطا | Alert بالای card/section | `Alert` | در Sheet ممکن است پیام خارج context باشد؛ در صفحه نتیجه کوتاه است. | Toast برای موفقیت؛ خطا داخل section/Sheet تا اصلاح. |
| Profile settings | ارسال OTP، تأیید، اتصال social | Alert بالای صفحه | `Alert` | پیام OTP باید داخل فرم بماند؛ نتیجهٔ تغییر موفق و خطای redirect کوتاه‌اند. | OTP/validation داخل فرم؛ نتیجهٔ موفق Toast؛ خطای عملیات در context. |
| Billing plan | خطای load | Alert | `Alert` | مسدودکنندهٔ داده است. | حفظ Alert. |
| Top-up | validation/API error و هشدار gateway | دو Alert صفحه | `Alert` | هشدار gateway پایدار است؛ validation باید نزدیک فیلد باشد. | هشدار Alert؛ validation FieldError/Alert فرم؛ خطای redirect/API Alert قابل اقدام. |
| پنل‌های مالی | خطای load | Alert؛ بخشی با متن hardcoded فارسی | `Alert` | الگو درست، ولی i18n و عنوان‌ها یکسان نیستند. | حفظ Alert و انتقال copy به i18n. |
| تیکت پشتیبانی: list/detail/new | load/validation/reply/rating | paragraph دست‌ساز و status متن | markup سفارشی | tone و composition ناهماهنگ؛ rating موفق/خطا تفکیک نشده است. | Alert استاندارد برای load/form؛ Toast برای rating ذخیره‌شده؛ FieldError برای validation. |
| `ConfirmDialog` عمومی | تأیید مخرب ساده | shadcn `Dialog` | `ConfirmDialog` | از primitive تخصصی AlertDialog استفاده نمی‌کند. | بازنویسی با `AlertDialog` و حفظ API محصولی. |
| `TypedConfirmDialog` | تأیید با ورود شناسه | Dialog فرم‌دار | `Dialog` | استفاده از AlertDialog برای فرم تایپی مناسب نیست. | حفظ Dialog؛ feedback داخل خود Dialog. |
| Unsaved changes | تأیید ترک فرم | Dialog | `UnsavedChangesDialog` | نیازمند پاسخ صریح است و API محصولی دارد. | بررسی composition؛ در صورت بدون ورودی بودن، استفاده از ConfirmDialog/AlertDialog. |
| CopyAction/CopyableValue | موفقیت کپی | متن sr-only محلی | `role="status"` | دسترس‌پذیر و نزدیک کنترل است؛ ولی الگوی بصری در featureها متفاوت است. | live region حفظ؛ Toast فقط وقتی workflow انتظار نتیجهٔ عمومی دارد. |
| Field validation | خطای فیلد | FieldError و چند متن دستی | `FieldError`/`p` | استفاده‌های دستی و متن hardcoded پراکنده‌اند. | استانداردسازی روی Field family؛ هرگز Toast تنها نباشد. |

## پلن اجرایی

1. یک API متمرکز و typed برای Toast بسازیم که `success`، `error`، `warning` و
   `info` را با عنوان/توضیح/action/id پشتیبانی کند و مستقیماً از `sonner` در
   featureها استفاده نشود.
2. `Toaster` را با tokenهای معنایی پروژه، جهت RTL، دکمهٔ بستن و تنظیمات یکسان
   پیکربندی کنیم.
3. primitive رسمی `alert-dialog` را اضافه و `ConfirmDialog` را بدون شکستن API
   مصرف‌کننده‌ها به آن منتقل کنیم؛ Dialogهای فرم‌دار تایپی را حفظ کنیم.
4. `OperationNotice` را برای feedback پایدار و context-local نگه داریم و نقش
   live-region را بر اساس شدت پیام اصلاح کنیم.
5. پیام‌های کوتاه عملیات صفحه‌ای را به API Toast منتقل کنیم؛ خطاهای load را از
   state عملیات جدا و به Alert پایدار محدود کنیم؛ خطاهای Sheet/Dialog را داخل
   همان overlay نگه داریم.
6. markupهای alert دست‌ساز در تیکت‌ها و صفحهٔ Log را با مؤلفهٔ مشترک جایگزین
   کنیم.
7. مثال‌های Alert/Toast/AlertDialog و ماتریس انتخاب را به playground و
   `COMPONENT_USAGE.md` اضافه کنیم.
8. تست‌های واحد زیرساخت، تست‌های مرتبط featureها، i18n validation، ESLint،
   TypeScript، Turbo lint و build تولید را اجرا کنیم.

## معیار پذیرش

- هیچ feature نباید مستقیماً از `sonner` import کند؛ فقط API متمرکز مجاز است.
- نتیجهٔ کوتاه عملیات پس از scroll نیز قابل مشاهده باشد.
- هیچ خطای Dialog/Sheet که overlay را باز نگه می‌دارد، فقط پشت overlay نمایش
  داده نشود.
- خطاهای مسدودکنندهٔ load فقط به Toast وابسته نباشند.
- تأییدهای مخرب ساده از AlertDialog و متن/دکمه‌های قابل دسترس استفاده کنند.
- تمام copyهای جدید در `fa` و `en` هم‌ساختار باشند و RTL حفظ شود.

## نتیجهٔ اجرای پلن

- adapter متمرکز `showFeedbackToast` روی Sonner ایجاد و تنها نقطهٔ مجاز مصرف
  مستقیم Toast در featureها تعریف شد.
- Toaster سراسری با toneهای معنایی، جهت RTL/LTR، دکمهٔ بستن، مدت نمایش و سقف
  تعداد یکسان پیکربندی شد.
- `ConfirmDialog` از Dialog عمومی به primitive تخصصی `AlertDialog` منتقل شد و
  خطای async می‌تواند با prop `feedback` داخل همان تأیید باقی بماند.
- عملیات تیم، واحد سازمانی، PAT، کاربر مجازی، مدل مجازی، کاربران و نقش‌ها، اتصال
  LLM، Guardrail، Rate/Budget Policy، Prompt Registry، Prompt Builder،
  Playground، Model Catalog، Provider Account Wizard، Quick Start و تنظیمات
  حساب/حریم خصوصی به قرارداد جدید منتقل شدند.
- پیام‌های موقتی موفقیت/کپی/حذف به Toast، خطاهای فرم به overlay مربوط، خطاهای
  load به notice پایدار و تأییدهای مخرب ساده به AlertDialog منتقل شدند.
- alertهای دست‌ساز صفحهٔ Log و تیکت‌های پشتیبانی با `Alert` استاندارد جایگزین
  شدند و مثال‌های Design System به‌روزرسانی شدند.

## بدهی UI باقی‌مانده

- برخی خطاهای پایدار load در صفحات قدیمی هنوز action استاندارد Retry ندارند؛
  نوع نمایش آن‌ها درست است، اما قابلیت بازیابی باید در یک تغییر جداگانه تکمیل
  شود.
- بازخوردهای داخل ابزارهای تعاملی مانند تست اتصال Provider، خروجی آزمایشگاهی
  Quick Start، اجرای Prompt و ذخیرهٔ ویرایشگر بازِ Virtual Model عمداً inline
  باقی مانده‌اند؛ این پیام‌ها بخشی از خروجی یا context جاری‌اند و Toast جایگزین
  مناسبی برای آن‌ها نیست.
- اعتبارسنجی‌های قدیمی که هنوز با متن دستی به‌جای Field family ساخته شده‌اند در
  این تغییر بازطراحی نشدند؛ ممیزی بالا محل آن‌ها را برای مهاجرت بعدی ثبت کرده
  است.

## نتیجهٔ اعتبارسنجی

- `pnpm run lint`: موفق.
- `pnpm run validate:i18n`: موفق برای ۲۸ namespace.
- تست‌های هدفمند `feedback-toast`، `OperationNotice` و `ConfirmDialog`: ۱۳ تست
  موفق در ۳ فایل.
- `pnpm run build`: موفق؛ شامل کامپایل production، TypeScript و تولید routeها.
- اجرای `pnpm run test:full` در محدودیت ۵ دقیقه‌ای محیط بدون نتیجهٔ نهایی
  timeout شد؛ بنابراین نتیجهٔ suite کامل در این تغییر تأیید نشده است.
- اجرای مستقل `tsc --noEmit` به خطاهای ازپیش‌موجود در چند فایل تست برخورد کرد؛
  TypeScript مرحلهٔ production build برای سورس اپلیکیشن موفق است.
