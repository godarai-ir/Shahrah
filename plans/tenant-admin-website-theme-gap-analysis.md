# Tenant Admin / Website Theme Gap Analysis

> هدف: هم‌راستا کردن ظاهر `apps/tenant-admin` با زبان بصری موفق `apps/website` بدون تخریب خوانایی، رفتار admin-oriented، یا maintainability.

---

## 1. جمع‌بندی اجرایی

وضعیت فعلی برای migration مناسب است، چون `tenant-admin` از قبل روی design tokenها و semantic colorها سوار شده و بیشتر کار در سطح refinement است، نه بازنویسی کامل.

چیزی که هنوز مانع reuse مستقیم تم می‌شود این است که `website` بیشتر یک **visual reference** است تا یک **shared design source**. هویت موفق آن در تعداد زیادی کلاس Tailwind hardcoded پخش شده و فقط بخشی از آن در tokenها متمرکز شده است.

### نتیجه

- `tenant-admin` برای دریافت تم `website` آماده است.
- `website` هنوز برای export مستقیم theme آماده نیست.
- بهترین مسیر، **sync کردن tokenها + حذف hardcodeهای کلیدی tenant-admin + تعریف چند primitive مشترک** است.

---

## 2. شواهد اصلی

### 2.1. زیرساخت تم در tenant-admin

- Design tokenهای مستقل و website-aligned در `apps/tenant-admin/design-system/tokens.css`
- map شدن tokenها به Tailwind theme در `apps/tenant-admin/app/globals.css`
- wiring تم و فونت در `apps/tenant-admin/app/layout.tsx`
- وجود documentation داخلی برای design system در:
  - `apps/tenant-admin/design-system/IMPLEMENTATION.md`
  - `apps/tenant-admin/design-system/COMPONENTS.md`

### 2.2. وضعیت واقعی website

- tokenهای پایه در `apps/website/app/styles/theme.css` وجود دارند.
- اما ظاهر نهایی سایت عمدتاً در سطح component/page با رنگ‌ها و surfaceهای hardcoded ساخته شده است.
- `apps/website/app/components/SiteShell.tsx` هنوز از کلاس‌های مستقیم مثل `dark:bg-[#0B0F19]` استفاده می‌کند.

### 2.3. شاخص‌های تقریبی

- `tenant-admin`: حدود `1037` استفاده از semantic utilityها
- `tenant-admin`: حدود `131` استفاده از رنگ‌های palette-based hardcoded
- `tenant-admin`: فقط `5` مورد رنگ hex-style مستقیم
- `website`: حدود `4533` استفاده از رنگ‌های palette-based hardcoded

### برداشت

- در `tenant-admin` نسبت hardcode به semantic پایین است و migration کنترل‌پذیر خواهد بود.
- در `website` منبع واقعی هویت بصری، token file نیست؛ بلکه مجموعه‌ای از الگوهای تکرارشونده در componentهاست.

---

## 3. شکاف‌های اصلی

### Gap A: منبع واحد حقیقت برای برند هنوز مشترک نیست

`tenant-admin` token-centric است، اما `website` هنوز brand expression را در فایل‌های UI نگه می‌دارد. بنابراین اگر امروز رنگ یا surface برند تغییر کند، sync دستی لازم می‌شود.

### Gap B: semantic tokenهای tenant-admin هنوز همه نیازهای website language را پوشش نمی‌دهند

چند الگوی موفق در `website` هنوز primitive رسمی در `tenant-admin` ندارند:

- gradient accent surface
- tinted status surfaces
- dark elevated panels
- marketing-style indigo glow
- code/demo surfaces با tone ثابت

### Gap C: تعدادی hotspot در tenant-admin هنوز hardcoded هستند

بیشترین فاصله در چند فایل محدود متمرکز است و اگر همان‌ها اصلاح شوند، بخش زیادی از ناهماهنگی بصری حذف می‌شود.

### Gap D: رفتار تم و first impression یکی نیست

- `website`: پیش‌فرض `dark`
- `tenant-admin`: پیش‌فرض `light`

این تفاوت لزوماً بد نیست، ولی اگر هدف "هارمونی یکسان" باشد، باید آگاهانه درباره‌اش تصمیم بگیریم.

### Gap E: typography implementation یکسان نیست

- `website`: `@font-face` محلی + inline font family
- `tenant-admin`: `next/font/google`

از نظر ظاهری نزدیک‌اند، ولی از نظر پیاده‌سازی هنوز unified نیستند.

---

## 4. اولویت فایل‌های tenant-admin برای migration

### P0 — بیشترین بازده با کمترین پراکندگی

این فایل‌ها بیشترین hardcode را دارند و بهتر است اولین موج migration روی آن‌ها انجام شود:

1. `apps/tenant-admin/app/components/provider-account-wizard/models-editor.tsx`
2. `apps/tenant-admin/app/components/llm-integrations-management.tsx`
3. `apps/tenant-admin/app/components/trace-privacy-settings.tsx`
4. `apps/tenant-admin/app/components/request-logging-settings.tsx`
5. `apps/tenant-admin/app/components/model-card.tsx`

### P1 — surfaceها و demo blockها

1. `apps/tenant-admin/app/components/provider-account-wizard/code-example-block.tsx`
2. `apps/tenant-admin/app/components/profile-settings.tsx`
3. `apps/tenant-admin/app/components/prompt-detail-page.tsx`
4. `apps/tenant-admin/app/components/provider-account-wizard/overview.tsx`

### P2 — cleanup و consistency

1. `apps/tenant-admin/features/gateway-policies/rate-limiting/RateLimitRulesPage.tsx`
2. `apps/tenant-admin/app/components/stat-card.tsx`
3. `apps/tenant-admin/app/components/code-block.tsx`
4. `apps/tenant-admin/app/components/users-access-management.tsx`

---

## 5. migration به چه چیزهایی باید تبدیل شود

### 5.1. token additions پیشنهادی

برای اینکه visual language `website` به شکل سیستماتیک وارد `tenant-admin` شود، پیشنهاد می‌شود tokenهای زیر به design system اضافه شوند:

- `--primary-soft`
- `--primary-soft-hover`
- `--primary-subtle-border`
- `--surface-elevated`
- `--surface-elevated-foreground`
- `--surface-demo`
- `--surface-demo-foreground`
- `--success-soft`
- `--warning-soft`
- `--danger-soft`
- `--info-soft`
- `--brand-gradient-start`
- `--brand-gradient-end`
- `--shadow-brand-glow`

### 5.2. primitiveهای مشترک پیشنهادی

به جای تکرار کلاس‌ها، این primitiveها بیشترین سود را می‌دهند:

- `BrandPanel`
  - برای کارت‌ها و containerهایی با surface نزدیک به website
- `StatusTint`
  - برای stateهایی مثل success/warning/info/danger
- `CodeSurface`
  - برای code/demo blockهای تیره
- `MetricCard`
  - برای stat/summary blockهای یکپارچه
- `FeatureBadge`
  - برای badgeهای soft-brand

---

## 6. تصمیم‌های UX که باید آگاهانه گرفته شوند

### تصمیم 1: default theme

دو مسیر معتبر داریم:

- `tenant-admin` روی `light` بماند و فقط brand expression به website نزدیک شود.
- `tenant-admin` هم مثل `website` به صورت پیش‌فرض `dark` شود.

پیشنهاد:
برای پنل مدیریتی فعلاً `light` حفظ شود، مگر اینکه تصمیم محصول بر هویت dark-first باشد. این کار هارمونی برند را حفظ می‌کند بدون اینکه کارایی پنل آسیب ببیند.

### تصمیم 2: شدت migration

دو سطح وجود دارد:

- `Brand Alignment`
  - فقط tokenها، surfaceها، badgeها، panelها و accentها هم‌راستا شوند.
- `Visual Convergence`
  - علاوه بر موارد بالا، کارت‌ها، headerها، gradients و demo patternها هم به website نزدیک شوند.

پیشنهاد:
اول `Brand Alignment`، بعد اگر نتیجه رضایت‌بخش بود، سراغ `Visual Convergence` برویم.

---

## 7. برنامه اجرایی پیشنهادی

### فاز 1: Token Hardening

هدف:
گسترش `tokens.css` برای پوشش دادن الگوهای موفق website

خروجی:

- tokenهای soft surface و gradient و demo surface
- حذف نیاز به hardcode در componentهای اصلی

### فاز 2: Primitive Extraction

هدف:
ساخت چند wrapper یا variant جدید برای panelها، badgeها و code surfaceها

خروجی:

- primitiveهای reusable
- کاهش تکرار کلاس‌های visual

### فاز 3: P0 Migration

هدف:
اصلاح فایل‌های پرریسک و پرتکرار

خروجی:

- بیشترین بهبود بصری با کمترین ریسک

### فاز 4: Theme Behavior Unification

هدف:
یکسان‌سازی یا deliberate کردن تصمیم‌های مربوط به:

- default theme
- typography loading
- dark mode surface rules

### فاز 5: P1/P2 Cleanup

هدف:
جمع کردن hardcodeهای باقی‌مانده و رسیدن به consistency بالا

---

## 8. ریسک‌ها

### ریسک 1: over-marketing شدن پنل

ظاهر `website` نمایشی‌تر است. اگر بدون فیلتر وارد `tenant-admin` شود، ممکن است readability و density مناسب پنل مدیریتی افت کند.

### ریسک 2: semantic drift

اگر فقط hardcodeها را با رنگ‌های جدید عوض کنیم ولی primitive نسازیم، چند ماه بعد دوباره divergence اتفاق می‌افتد.

### ریسک 3: dark mode inconsistency

بخش‌هایی از `tenant-admin` الان برای dark mode local override دارند. اگر token coverage کامل نشود، ظاهر dark ناهمگون می‌شود.

---

## 9. پیشنهاد نهایی

قدم بعدی باید اجرای یک migration محدود ولی با اثر بالا باشد:

1. گسترش tokenهای `apps/tenant-admin/design-system/tokens.css`
2. ساخت 2 تا 3 primitive سبک برای surface و status
3. migration فایل‌های P0

این مسیر، سریع‌ترین راه برای رسیدن به هارمونی واقعی با `website` است بدون اینکه مجبور شویم ابتدا خود `website` را بازسازی کنیم.

---

## 10. پیشنهاد اجرای بعدی

اگر ادامه بدهیم، بهترین تسک بعدی این است:

**اجرای فاز 1 و فاز 2 روی `tenant-admin`**

یعنی:

- افزودن tokenهای missing
- ساخت primitiveهای surface/status/code
- migrate کردن 1 یا 2 فایل از گروه P0 به عنوان pilot

این کار هم ریسک را پایین نگه می‌دارد، هم خیلی سریع نشان می‌دهد که direction بصری انتخاب‌شده درست است یا نه.
