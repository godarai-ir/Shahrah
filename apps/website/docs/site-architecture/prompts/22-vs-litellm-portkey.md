# پرامپت ۲۲: صفحه‌های مقایسهٔ درگاه `/vs/litellm` و `/vs/portkey`

اول `apps/website/docs/site-architecture/prompts/_shared-brief.md` را کامل بخوان. پرامپت‌های `00`، `01`، `13` و `16` باید اجرا شده باشند.

این پرامپت دو صفحه را با هم پوشش می‌دهد، چون هر دو یک قالب (`ComparisonPage`) دارند.

## نقش و هدف

تو مهندس فرانت‌اند هستی. دو صفحهٔ مقایسهٔ درگاه برای توسعه‌دهنده‌ای است که LiteLLM یا Portkey را می‌شناسد و می‌پرسد «چرا مهاجرت کنم؟». قالب (`ComparisonPage`، ۳۸۱ خط) روی سیستم جدید و `ComparisonTable` مشترک با `/vs/chatgpt` برود. **محتوای مقایسه حفظ می‌شود**، با بازبینی درستی.

## فایل‌های مخصوص

`app/pages/compare/ComparisonPage.tsx`، `GodarAIvsLiteLLM.tsx`، `GodarAIvsPortkey.tsx`، `app/(site)/vs/*/page.tsx`، `components/patterns/ComparisonTable`.

## کارها

1. قالب: `PageHero` · `ComparisonTable` · بخش‌های «چه زمانی مهاجرت منطقی است» (FeatureRow) · `ClosingBand` با «۱۵ روز رایگان» و «مستندات».
2. بازبینی محتوا: هر ادعا دربارهٔ رقیب قابل بررسی باشد (اگر ادعایی منبع ندارد، در گزارش بیاور و پیشنهاد حذف بده؛ خودت حذف نکن مگر نادرستی آشکار)؛ هر ادعا دربارهٔ گدارAI با `PRODUCT.md` سازگار باشد.
3. پیوند به `/features/ai-gateway` و `/features/cost-control`.

## معیار پذیرش

طبق بریف مشترک، برای هر دو صفحه. جدول در موبایل کارت پشته‌ای، بدون اسکرول افقی.
