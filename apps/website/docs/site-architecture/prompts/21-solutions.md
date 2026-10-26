# پرامپت ۲۱: قالب راهکارها `/solutions/[slug]`

اول `apps/website/docs/site-architecture/prompts/_shared-brief.md` را کامل بخوان. پرامپت‌های `00` و `01` باید اجرا شده باشند.

## نقش و هدف

تو مهندس فرانت‌اند هستی. شش صفحهٔ راهکار بر اساس نوع کسب‌وکار (`startups-freelancers`، `b2c-ai-products`، `b2b-ai-products`، `ai-education`، `platform-teams`، `banking-insurance-fintech`) با یک قالب (`SolutionPage`، ۴۶۱ خط) و دادهٔ `lib/solutions.ts` (۷۲۰ خط) ساخته می‌شوند. کار: قالب را روی سیستم جدید ببر. **محتوای داده حفظ می‌شود** مگر با `PRODUCT.md` تعارض داشته باشد.

## فایل‌های مخصوص

`app/pages/solutions/SolutionPage.tsx`، `app/lib/solutions.ts` (انواع `usage`، `routing`، `audit` برای پنجره‌ها)، `app/(site)/solutions/[slug]/page.tsx`.

## کارها

1. ساختار قالب را با الگوهای مشترک بازسازی کن: `PageHero` (عنوان و Hero فعلی) · ProductWindow از دادهٔ `usage` / `routing` / `audit` با برچسب «نمونه» · FeatureRow برای ویژگی‌ها · `ClosingBand` با `finalCta` · «راهکارهای دیگر گدارAI».
2. تیند: اگر یک راهکار به‌طور روشن به یک تمایز تکیه دارد (مثلاً `banking-insurance-fintech` ← گاردریل و حاکمیت)، فیلد اختیاری `pillar` در داده و تیند همان؛ وگرنه `site-panel`.
3. **بازبینی محتوا** با `PRODUCT.md`: «پرداخت ریالی» → «پرداخت تومانی» اگر منظور همان است؛ واژه‌های انگلیسی پراکنده مثل «endpoint» طبق واژه‌نامه («نقطهٔ پایانی» یا «یک نشانی»)؛ هر عدد ساختگی در پنجره‌ها برچسب «نمونه» بگیرد و در متن اطراف ادعا نشود؛ هر ادعای ⚠️ حذف.
4. پیوند به صفحه‌های تمایز به‌جای توضیح تکراری.

## معیار پذیرش

طبق بریف مشترک، برای هر شش slug (اسکرین‌شات دو slug در هر چهار حالت). در گزارش فهرست تغییرهای متنی داده با دلیل هر کدام.
