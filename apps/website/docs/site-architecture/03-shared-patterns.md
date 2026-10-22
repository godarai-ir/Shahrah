# الگوهای مشترک و کامپوننت‌ها

این الگوها پیش از هر صفحه ساخته یا بازسازی می‌شوند (پرامپت‌های `00` و `01`). مبنا: `apps/website/DESIGN.md` («The Lit Workshop»). همهٔ کامپوننت‌ها RTL، با توکن‌های معنایی، در حالت روشن و تاریک، و با `prefers-reduced-motion`.

## وضعیت فعلی پایه

| موضوع | وضعیت | اقدام |
| --- | --- | --- |
| توکن‌های رنگ | ✅ پرامپت ۰۰: `--ford-*` به `--site-*` تغییر نام داد؛ چهار تیند `--tone-*` و `--site-night-strong` اضافه شد. `/workplace` هنوز توکن‌های جدای `--wp-*` دارد (۹ فایل در `components/workplace/` و `pages/Workplace.tsx`). | `wp-*` در پرامپت ۰۳ حذف می‌شود. |
| پوستهٔ سایت | ✅ `SiteShell` روی `bg-site-ground text-site-ink`. `Navbar` و فوتر (`Trust.tsx`) هنوز رنگ خام دارند. | ناوبری و فوتر در پرامپت ۰۱. |
| کلاس‌های مشترک بخش | `components/sectionType.ts`: `sectionHeading`، `sectionCopy`، `sectionSpace`، `buttonPrimary`، `buttonQuiet` | مبنای همهٔ صفحه‌ها؛ گسترش با `sectionPanel` و تیندها. |
| حرکت | ✅ `step-row` / `step-pending` / `step-complete`: ردیف‌های `ToolSteps` یکی‌یکی می‌آیند و نتیجه یک بار کامل می‌شود؛ `PlayOnView` آن را تا ورود به دید نگه می‌دارد و بدون اسکریپت یا با reduced-motion حالت نهایی دیده می‌شود. ✅ پرامپت ۰۲: `FordArt` و `water-ripple` حذف شدند. | — |
| رسانه | ✅ `MediaSlot` روی توکن‌ها و `cn()`؛ `ProductWindow` با prop `media` از آن استفاده می‌کند. | — |

## توکن‌های تیند تمایز

✅ در `theme.css` پیاده شد؛ کنتراست هر چهار در هر دو حالت AA است (متن `site-ink` بالای ۱۳:۱، `site-ink-soft` بالای ۶:۱، `site-accent` بالای ۵:۱). کلاس‌ها: `bg-tone-guard`، `bg-tone-govern`، `bg-tone-observe`، `bg-tone-cost`.

مقادیر پیشنهادی `DESIGN.md`، با هم‌روشنایی (The Matched Lightness Rule). پیش از استفاده، کنتراست متن `ink` روی هر تیند در هر دو حالت بررسی شود.

| نقش | نام پیشنهادی توکن | روشن | تاریک |
| --- | --- | --- | --- |
| گاردریل معنایی | `tone-guard` | `oklch(0.955 0.03 160)` | `oklch(0.26 0.03 160)` |
| حاکمیت | `tone-govern` | `oklch(0.955 0.025 295)` | `oklch(0.26 0.03 295)` |
| مشاهده‌پذیری | `tone-observe` | `oklch(0.955 0.03 230)` | `oklch(0.26 0.03 230)` |
| هزینه | `tone-cost` | `oklch(0.96 0.035 80)` | `oklch(0.27 0.03 80)` |

✅ نقشهٔ مرکزی `app/lib/pillars.ts` (با `toneSurface` و `getPillar`) نگه می‌دارد: شناسه، عنوان فارسی، توضیح یک‌خطی، مسیر، آیکون lucide، و کلاس تیند. ناوبری، PillarGrid و PillarCrossLinks همه از همین نقشه می‌خوانند تا رنگ هر تمایز همه‌جا یکی بماند (The Color Means a Promise Rule).

## الگوها

✅ همه در `app/components/patterns/` با خروجی واحد `index.ts` ساخته شدند. `PageHero`، `ProductStage`/`ProductWindow`، `ToolSteps`، `FeatureRow`، `ControlChecklist`، `StepList`، `FAQ`، `ClosingBand` و `Section` اکنون در Home استفاده می‌شوند؛ `PillarGrid`، `PillarCrossLinks`، `ComparisonTable` و `TrustStrip` فقط در `/_patterns`. `TeamTabs` پیمایش کیبوردی (فلش در جهت RTL، Home/End) گرفت؛ داده‌اش در پرامپت ۱۰ به `lib/teams.ts` می‌رود.

افزوده‌های پرامپت ۰۲: `ProductStage` گزینهٔ `overhang` دارد تا پنجره از لبهٔ پنل بیرون بزند؛ `FAQItem` پیوند اختیاری `link` دارد؛ `PillarGrid` و هر محتوایی که به صفحه‌ای ساخته‌نشده اشاره می‌کند از `isReady()` در `lib/navigation.ts` استفاده می‌کند تا پیوند ۴۰۴ ساخته نشود.

| الگو | کامپوننت فعلی | ساخت یا بازسازی | Props و حالت‌ها |
| --- | --- | --- | --- |
| **HeroProduct**: تیتر Display، یک جمله، دو CTA، ProductWindow کنار یا زیر | `Hero` (فقط صفحهٔ اصلی) | عمومی‌سازی به `PageHero` | `title`، `emphasis?` (یک واژهٔ تأکیدی)، `lead`، `primary`، `secondary?`، `visual?: ReactNode`، `tone?`؛ چیدمان دوستونه در دسکتاپ، تک‌ستونه با تصویر بعد از متن در موبایل. |
| **ProductWindow**: قاب پنجره با عنوان، سایهٔ Window lift، روی پنل | `HeroWindow` (محتوای ثابت) | عمومی‌سازی | `title`، `meta?` (مثلاً «نمونه»)، `children` یا `media` (شناسهٔ `MediaSlot` برای اسکرین‌شات)، `tone?` برای پنل زیر، `overhang?`؛ اسکرین‌شات با `next/image`، `alt` فارسی، نسخهٔ تاریک اختیاری. |
| **ToolSteps** (درون پنجره): ردیف ابزار که از «در حال انجام» به «تکمیل» می‌رود | منطق داخل `HeroWindow` | استخراج | `steps: {icon, name, text}[]`، `result`؛ اجرا یک‌باره با ورود به دید؛ در reduced-motion حالت نهایی مستقیم. |
| **FeatureRow**: متن + تصویر، طرف‌ها متناوب | ندارد (تکرار دستی در صفحه‌ها) | جدید | `title`، `copy`، `bullets?`، `visual`، `reverse?`، `tone?`، `link?`. |
| **PillarGrid**: کارت‌های تمایز روی تیند خودشان | `Pillars` (محتوای ثابت، شماره‌دار) | بازسازی داده‌محور | `items` از `pillars.ts` یا سفارشی، `columns: 2 \| 3 \| 4`، `numbered?`؛ کارت قابل کلیک با سایهٔ Card rest. |
| **ControlChecklist**: فهرست کنترل‌های سازمانی، هر ردیف به صفحهٔ خود | `Enterprise` | بازسازی داده‌محور | `groups: {title, items: {label, href?, pillar?}[]}[]`؛ فقط کنترل‌های تأییدشده. |
| **TeamTabs**: تب‌های قرصی تیم، پنجرهٔ کنار عوض می‌شود | `TeamTabs` | داده از `lib/teams.ts` | `teams`، `defaultTeam`، `linkToTeamPage`؛ پیمایش کیبورد (فلش‌ها با جهت RTL)، `role="tablist"`. |
| **StepList**: سه قدم شماره‌دار | `GetStarted` | پارامتری | `steps`، `aside?` (مثلاً «برای تیم فنی»). |
| **ComparisonTable**: مقایسهٔ منصفانه | داخل `ComparisonPage` | استخراج | `columns`، `rows: {label, values, note?}`؛ در موبایل به کارت‌های پشته‌ای تبدیل شود. |
| **TrustStrip**: سه مزیت محلی | بخش بالایی `WorkplaceCta` | استخراج | ثابت: پرداخت تومانی، ابر یا سرور خودتان، فارسی؛ بالای فوتر و در صورت نیاز در صفحه. |
| **PillarCrossLinks**: سه تمایز دیگر در انتهای صفحهٔ هر تمایز | ندارد | جدید | `current: PillarId`؛ سه کارت کوچک تیندی. |
| **FAQ**: پرسش‌های گروه‌بندی‌شده | `HomeFAQ` + `ui/FAQAccordion` | گروه‌بندی | `groups?: {title, items}[]`؛ JSON-LD `FAQPage` در صورت وجود در کد فعلی حفظ. |
| **ClosingBand**: باند تیره، یک بار در صفحه | `ClosingCta` | پارامتری | `title`، `copy`، `primary`، `secondary?`. |
| **Section**: فاصله و عرض استاندارد | `sectionSpace` + `max-w-7xl` | کامپوننت نازک | `tone?`، `id?`، `as?`. |

## الگوهایی که فعلاً ساخته نمی‌شوند

- **LogoStrip و QuoteCard:** کد بماند، استفاده نشود تا محتوای واقعی (`PRODUCT.md`).
- **کاروسل داستان مشتری** (`WorkplaceGuide` بخش داستان‌ها): حذف از صفحه تا محتوای واقعی.
- **تصویرسازی آب و گدار**: ✅ در پرامپت ۰۲ همراه `Collab` و `WorkflowCard` حذف شد.

## زمین بازی سیستم طراحی

طبق `AGENTS.md` ریشه، الگوهای قابل بازاستفاده باید در زمین بازی ثبت شوند. زمین بازی فعلی متعلق به `tenant-admin` است (`apps/tenant-admin/app/design-system`). تصمیم مالک: وب‌سایت زمین بازی خودش را در `/_patterns` (پوشهٔ `app/(site)/%5Fpatterns/`، با `robots: noindex`، بیرون از نقشهٔ سایت) دارد؛ در پرامپت ۰۰ ساخته می‌شود.

## تغییر نام توکن‌ها

✅ انجام شد در پرامپت ۰۰: همهٔ `ford-*` به `site-*` تغییر نام داد.
