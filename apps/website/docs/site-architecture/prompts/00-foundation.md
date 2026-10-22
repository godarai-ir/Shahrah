# پرامپت ۰۰: پایهٔ سیستم طراحی و الگوهای مشترک

اول `apps/website/docs/site-architecture/prompts/_shared-brief.md` را کامل بخوان و قیدهایش را رعایت کن.

## نقش و هدف

تو مهندس فرانت‌اند ارشد سیستم طراحی وب‌سایت گدارAI هستی. پیش از بازطراحی هر صفحه، پایهٔ مشترک باید آماده باشد: توکن‌ها، پوستهٔ سایت و الگوهای بخش. خروجی این کار صفحهٔ جدید نیست؛ ابزارهایی است که همهٔ پرامپت‌های بعدی از آن‌ها استفاده می‌کنند. هیچ صفحه‌ای جز آن‌چه برای حفظ کارکرد لازم است عوض نشود.

## فایل‌های مخصوص این کار

- `apps/website/app/styles/theme.css` (توکن‌های `--ford-*` و `--wp-*`، کلاس‌های حرکت)
- `apps/website/app/components/sectionType.ts`، `Reveal.tsx`، `HeroWindow.tsx`، `Hero.tsx`، `Pillars.tsx`، `Enterprise.tsx`، `TeamTabs.tsx`، `GetStarted.tsx`، `HomeFAQ.tsx`، `ui/FAQAccordion.tsx`، `ClosingCta.tsx`، `MediaSlot.tsx`، `lib/home-media.ts`، `SiteShell.tsx`، `FordArt.tsx`
- `apps/website/app/components/workplace/**` (برای فهم توکن‌های `wp-*`)
- `apps/website/app/pages/compare/ComparisonPage.tsx` (جدول مقایسه)
- `apps/website/.impeccable/design.json`

## کارها به ترتیب

1. **تغییر نام توکن‌ها (تصمیم مالک، د.۲):** همهٔ `--ford-*` و `ford-*` را در یک گذر به `--site-*` و `site-*` تغییر بده (در `theme.css`، `@theme inline` و همهٔ کامپوننت‌ها). `DESIGN.md`، `.impeccable/design.json`، `03-shared-patterns.md` و `_shared-brief.md` را هم به‌روز کن. پس از کار، `grep -r "ford-" apps/website/app` خالی باشد.
2. **توکن‌های تیند:** چهار توکن `tone-guard`، `tone-govern`، `tone-observe`، `tone-cost` با مقادیر روشن و تاریک جدول `03-shared-patterns.md` در `theme.css` و `@theme inline`. کنتراست `ford-ink` و `ford-ink-soft` روی هر تیند را در هر دو حالت محاسبه کن؛ اگر زیر AA بود، روشنایی تیند را در همان خانواده تنظیم کن و مقدار نهایی را در `DESIGN.md` (بخش Differentiator panels، حذف برچسب provisional) و `.impeccable/design.json` ثبت کن.
3. **سایه‌ها:** سه سایهٔ `DESIGN.md` (Window lift، Card rest، Hairline) به‌صورت توکن یا کلاس مشترک تا هیچ صفحه‌ای مقدار دلبخواه ننویسد.
4. **نقشهٔ تمایزها:** `app/lib/pillars.ts` با چهار تمایز (شناسه، عنوان، توضیح یک‌خطی، مسیر، آیکون lucide، کلاس تیند):
   - `guard`: «گاردریل معنایی» · «دادهٔ محرمانه را از معنایش می‌شناسد.» · `/features/guardrails` · `ShieldCheck`
   - `govern`: «حاکمیت» · «هر تیم به همان چیزی می‌رسد که باید.» · `/features/governance` · `Users`
   - `observe`: «مشاهده‌پذیری» · «مصرف و هزینهٔ هر تیم، و رد هر درخواست.» · `/features/observability` · `Activity`
   - `cost`: «مدیریت هزینه» · «هزینه زیر سقفی که خودتان می‌گذارید.» · `/features/cost-control` · `Wallet`
5. **پوستهٔ سایت:** `SiteShell` روی `bg-ford-ground text-ford-ink` و `selection` معنایی؛ حذف `bg-white`، `#0B0F19`، indigo و `style` فونت درون‌خطی اگر فونت از `layout.tsx` می‌آید.
6. **الگوها** (جدول «الگوها» در `03-shared-patterns.md`): `Section`، `PageHero`، `ProductWindow` (+ `ToolSteps` استخراج‌شده از `HeroWindow`)، `FeatureRow`، `PillarGrid`، `ControlChecklist`، `StepList`، `ComparisonTable`، `TrustStrip`، `PillarCrossLinks`، گروه‌بندی در `FAQ`، و `ClosingBand` پارامتری. مسیر پیشنهادی: `app/components/patterns/`. کامپوننت‌های فعلی Home را طوری بازنویسی کن که از این الگوها استفاده کنند **بدون تغییر محتوای دیدنی صفحهٔ اصلی** (محتوا در پرامپت ۰۲ عوض می‌شود).
7. **حرکت:** کلاس‌های `ford-pending` / `ford-complete` به نام خنثی (`step-pending` / `step-complete`) و `ford-ripple` و `FordArt` حذف، اگر جای دیگری استفاده نمی‌شوند (با جست‌وجو بررسی کن).
8. **`wp-*`:** فهرست جاهای استفاده را در گزارش بیاور؛ حذف آن‌ها در پرامپت ۰۳ (`/workplace`) انجام می‌شود، نه این‌جا.
9. **زمین بازی (تصمیم مالک، د.۶):** نشانی `/_patterns` با `robots: { index: false, follow: false }`، بیرون از `sitemap.ts` و ناوبری، با نمونهٔ هر الگو در هر چهار تیند. حالت تاریک با کلید تم سایت بررسی می‌شود. توجه: پوشه‌ای که با `_` شروع شود در App Router خصوصی است و مسیر نمی‌سازد؛ پوشه را `app/(site)/%5Fpatterns/` بنام تا نشانی `/_patterns` ساخته شود، و در مرورگر بررسی کن.

## قیدهای ویژه

- API الگوها کوچک و صریح؛ بدون `className` آزاد برای رنگ (فقط `tone`).
- `PageHero` در موبایل: متن، دکمه‌ها، سپس تصویر. `ProductWindow` با `overflow-hidden` و گوشهٔ ۲۴ پیکسل؛ اسکرین‌شات با `next/image` و ابعاد مشخص (بدون جابه‌جایی چیدمان).
- `TeamTabs` و `FAQ` کیبوردپذیر و با ARIA درست.

## معیار پذیرش

- `lint` و `build` بدون خطا؛ صفحهٔ اصلی در مرورگر از نظر ظاهر و محتوا با پیش از تغییر یکسان (اسکرین‌شات قبل و بعد، روشن و تاریک، ۳۷۵ و ۱۴۴۰).
- `grep` برای `indigo-|slate-|bg-white|#0B0F19` در `SiteShell.tsx` و `components/patterns/` خالی.
- گزارش پایانی طبق بریف مشترک، به‌علاوهٔ فهرست props هر الگو.
