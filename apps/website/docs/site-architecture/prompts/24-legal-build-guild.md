# پرامپت ۲۴: حریم خصوصی، قوانین و Build Guild (فقط مهاجرت توکن)

اول `apps/website/docs/site-architecture/prompts/_shared-brief.md` را کامل بخوان. پرامپت `00` باید اجرا شده باشد.

## نقش و هدف

تو مهندس فرانت‌اند هستی. سه صفحهٔ `/privacy`، `/terms` و `/build-guild` طراحی جدید نمی‌گیرند. فقط ظاهرشان با سیستم جدید هم‌خوان می‌شود تا سایت یکدست بماند. **هیچ متنی تغییر نمی‌کند** (متن حقوقی از `privacy-policy.md` و `terms.md` می‌آید؛ نظرسنجی از `lib/build-guild-survey.ts`).

## فایل‌های مخصوص

`app/pages/Privacy.tsx`، `Terms.tsx`، `components/MarkdownDocument.tsx`، `app/pages/BuildGuildSurvey.tsx`، `components/BuildGuildGiftNote.tsx`.

## کارها

1. `MarkdownDocument`: تایپوگرافی خواندنی روی توکن‌ها (Body با ارتفاع خط ۲، پهنای ۶۰ تا ۷۰ نویسه، عنوان‌ها طبق Headline/Title، پیوندها `site-accent`، جدول‌ها با `site-line`). حالت Read، نه Persuade: بدون Hero بزرگ، بدون ClosingBand.
2. `BuildGuildSurvey`: جایگزینی رنگ‌های خام با توکن‌ها؛ منطق فرم دست نخورد؛ فیلدها برچسب دیدنی و فوکوس روشن.
3. `grep` رنگ‌های خام در این فایل‌ها پس از کار خالی باشد.

## معیار پذیرش

طبق بریف مشترک. مقایسهٔ متن پیش و پس (`git diff` فقط کلاس‌ها را نشان دهد، نه متن).
