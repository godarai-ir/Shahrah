# بسته دارایی‌های وب GodarAI

این بسته از تصویر ارسالی شما ساخته شده و برای استفاده در وب‌سایت، وب‌اپلیکیشن، PWA و شبکه‌های اجتماعی آماده است.

## فایل‌های اصلی

- `svg/godarai-mark-light.svg`: نسخه اصلی برای پس‌زمینه روشن
- `svg/godarai-mark-dark.svg`: نسخه خوانا روی پس‌زمینه تیره؛ بخش سرمه‌ای به سفید یخی تبدیل شده است
- `svg/godarai-mark-adaptive.svg`: نسخه تطبیقی با `prefers-color-scheme`
- `svg/godarai-mark-current-color.svg`: نسخه تک‌رنگ برای استفاده Inline و کنترل رنگ با CSS
- `svg/godarai-mark-mono-*`: نسخه‌های تک‌رنگ سرمه‌ای، مشکی و سفید
- `png/natural/*`: PNG شفاف با نسبت اصلی در عرض‌های 128 تا 2048 پیکسل
- `png/square/*`: PNG شفاف مربعی در اندازه‌های رایج 16 تا 1024 پیکسل
- `icons/favicon.ico`: فاوآیکن چندرزولوشنی 16/32/48/64
- `icons/favicon.svg`: فاوآیکن SVG تطبیقی برای لایت/دارک مود
- `icons/apple-touch-icon.png`: آیکن استاندارد iOS در اندازه 180×180
- `icons/android-chrome-*`: آیکن‌های PWA
- `icons/maskable-icon-*`: آیکن‌های Maskable با محدوده امن
- `social/og-image-*`: تصاویر 1200×630 برای Open Graph
- `social/style-guide/STYLE-GUIDE-fa.md`: مرجع اجرایی هویت بصری شبکه‌های اجتماعی
- `social/style-guide/GodarAI-Social-Media-Style-Guide-fa.pptx`: نسخه بصری و قابل‌ویرایش راهنما
- `social/style-guide/GodarAI-Social-Media-Style-Guide-fa.pdf`: نسخه نهایی مناسب مطالعه و اشتراک‌گذاری
- `social/style-guide/social-tokens.json`: توکن‌های رنگ، تایپوگرافی، گرید، ابعاد و قالب‌ها
- `web/site.webmanifest`: مانیفست آماده PWA
- `web/html-head-snippet.html`: تگ‌های پیشنهادی `<head>`
- `web/GodarAILogo.tsx`: نمونه استفاده در Next.js

## رنگ‌های برند استخراج‌شده

- Cyan: `#00C9EB`
- Navy: `#012651`
- Dark background: `#06182D`
- Dark-mode secondary: `#F6FBFF` تا `#D9EEFF`

## توصیه استفاده

برای رابط روشن از `godarai-mark-light.svg` و برای رابط تیره از `godarai-mark-dark.svg` استفاده کنید. نسخه اصلی روی پس‌زمینه بسیار تیره توصیه نمی‌شود، چون بخش سرمه‌ای کنتراست کافی ندارد.

برای فاوآیکن، فایل `favicon.svg` انتخاب اول است و `favicon.ico` به‌عنوان fallback نگه داشته شود. فایل ICO روی یک کاشی سرمه‌ای ساخته شده تا در نوار مرورگر روشن و تیره خوانا بماند.

برای PWA، آیکن‌های `maskable` عمداً پس‌زمینه غیرشفاف و فاصله امن بیشتری دارند. PNGهای پوشه `png/` شفاف هستند.

## راهنمای شبکه‌های اجتماعی

پیش از ساخت پست، کاروسل، کاور ویدئو یا محتوای داده‌محور، راهنمای پوشه
`social/style-guide/` را مبنا قرار دهید. منبع نشان همچنان همین بسته است و راهنما
اجازه بازطراحی، تغییر رنگ یا تغییر نسبت نشان را نمی‌دهد.

## نکته درباره SVG

تصویر اولیه Raster بود. SVGهای این بسته با بازسازی تمیز مسیرهای Bézier ساخته شده‌اند؛ بنابراین واقعاً مقیاس‌پذیر هستند و تصویر PNG را داخل SVG جاسازی نکرده‌اند.
