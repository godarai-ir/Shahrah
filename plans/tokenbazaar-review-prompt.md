# TokenBazaar — دستورالعمل ریویو (۴ کامیت push‌نشده)

## نقش ریویوئر

یک مهندس نرم‌افزار ارشد و متخصص LLM Gateway که هم به معماری Go در پروژه‌های
production مسلط است و هم مستندات API مدل‌های LLM اصلی (Anthropic، Google Gemini،
OpenAI، Zhipu/GLM، Moonshot/Kimi، DeepSeek) را به‌خوبی می‌شناسد. هدف: ریویوی
عمیق و بی‌رحمانه (اما منصفانه) روی چهار کامیت اخیر و هنوز push‌نشده.

## محدوده‌ی دقیق ریویو

چهار کامیت پشت‌سرهم روی `main`، هنوز push نشده‌اند:

| کامیت | عنوان |
| --- | --- |
| `7ce91ed6` | Add new model configurations for GLM-5.2, GLM-5V-Turbo, GPT-5.5, GPT-5.6 (Luna, Sol, Terra), Kimi models (K2.6, K2.7-Code, K3), GPT-Image-2, Text-Embedding-3-Large |
| `8d1d8be9` | feat: add TokenBazaar provider support |
| `a64ae8a1` | feat: add TokenBazaarMessagesHandler for handling /v1/messages |
| `e1dd7d21` | feat: Implement TokenBazaarGenerateContentHandler for generateContent و streamGenerateContent |

دیف کامل: `git diff 7ce91ed6~1..e1dd7d21`

## هدف بیزینسی این چهار کامیت

اضافه‌کردن پروایدر جدید «TokenBazaar» به گدار (Gadar/گدار)، شامل:

- **۲۶ مدل در کاتالوگ** (فایل‌های yaml زیر `apps/model-discovery/providers/tokenbazaar/`):
  - خانواده Claude: `fable-5`, `haiku-4-5`, `opus-5`, `sonnet-5`
  - خانواده DeepSeek: `v4-flash`, `v4-pro`
  - خانواده Gemini: `2.5-flash`, `2.5-pro`, `3-pro-image`, `3.1-flash-image`, `3.1-flash-lite-image`, `3.1-pro-preview`, `3.5-flash`, `3.6-flash`, `3.7-flash`
  - خانواده GLM: `5.2`, `5v-turbo`
  - خانواده GPT: `5.5`, `5.6-luna`, `5.6-sol`, `5.6-terra`, `image-2`
  - خانواده Kimi: `k2.6`, `k2.7-code`, `k3`
  - و `text-embedding-3-large`
  - (`default.yaml` قالب/دیفالت است و جزو ۲۶ مدل شمرده نمی‌شود — درستی این فرض را هم راستی‌آزمایی کن.)
- **سه اندپوینت رایج و پرکاربرد** پشت پروکسی گدار:
  - `/v1/messages` (سازگار با Anthropic Messages API) — `services/control-plane/internal/proxy/tokenbazaar_messages_handler.go`
  - `generateContent` و `streamGenerateContent` (سازگار با Gemini native API) — `services/control-plane/internal/proxy/tokenbazaar_generate_content_handler.go`
  - مسیرهای embeddings/images/responses موجود که برای TokenBazaar گسترش داده شده‌اند (`embeddings_handler.go`، `images_handler.go`، `responses_handler.go`، `anthropic_native_handler.go`، `gemini_native_handler.go`)
- **ثبت پروایدر و روتینگ**: `internal/providerregistry/registry.go`، `internal/server/routes_proxy.go`، `internal/server/server.go`، `internal/bootstrap/build_gateway.go`
- **قیمت‌گذاری**: `internal/billing/provider_pricing_strategy.go` و تست‌های `tokenbazaar_embeddings_pricing_test.go` / `tokenbazaar_image_pricing_test.go`
- **شبیه‌ساز/mock سرویس TokenBazaar** در `services/silk-road/pkg/server/` (`tokenbazaar.go`، `tokenbazaar_image_discovery.go` و تست‌هایشان) که برای تست‌ها استفاده می‌شود
- **ادمین/فرانت**: `provider_catalog.go`، `provider-connection-test-dialog.tsx`، catalog importer، فایل‌های i18n (`en.json`/`fa.json`)
- سند `plans/tokenbazaar-production-readiness.md` که محدوده و چک‌لیست production-readiness این فیچر را از قبل تعریف کرده است.

---

## مرحله‌ی ۱ — راستی‌آزمایی در برابر مستندات رسمی (source of truth)

هیچ ادعایی درباره‌ی درستی پیاده‌سازی را بدون رجوع به مستندات رسمی همان مدل/پروایدر
نپذیر. برای هر خانواده مدل، مرجع رسمی مربوطه را باز کن و فیلدهای زیر را در yaml
کاتالوگ + منطق هندلرها/pricing مقایسه کن:

- **Anthropic** (Claude fable-5/opus-5/sonnet-5/haiku-4-5) → مستندات رسمی Anthropic API (Messages API، مدل‌ها، pricing، prompt caching، tool use، vision، extended thinking در صورت وجود)
- **Google** (خانواده Gemini، از جمله نسخه‌های image) → مستندات رسمی Gemini API گوگل (generateContent/streamGenerateContent، function calling، context caching، مدالیتی‌های ورودی/خروجی هر مدل، محدودیت توکن، pricing)
- **OpenAI** (GPT-5.5، GPT-5.6 Luna/Sol/Terra، GPT-Image-2، text-embedding-3-large) → مستندات رسمی OpenAI (Responses/Chat API، Images API، Embeddings API، pricing، ابعاد embedding، اندازه/فرمت تصویر)
- **Zhipu/GLM** (GLM-5.2، GLM-5V-Turbo) → مستندات رسمی Zhipu AI
- **Moonshot/Kimi** (K2.6، K2.7-Code، K3) → مستندات رسمی Moonshot AI
- **DeepSeek** (v4-flash، v4-pro) → مستندات رسمی DeepSeek API

برای هر مدل این‌ها را چک کن و هر مغایرت را با ذکر «فیلد در کد / مقدار در مستندات رسمی / لینک منبع» گزارش کن:

1. context window و max output tokens
2. قیمت‌گذاری input/output (و اگر cached input/output دارد) — واحد و نرخ
3. پشتیبانی از streaming
4. پشتیبانی از tool/function calling، parallel tool calls، structured output / JSON mode
5. مدالیتی‌های ورودی و خروجی (متن، تصویر، صدا، ویدیو) و محدودیت‌های آن‌ها (حداکثر تعداد تصویر، حداکثر سایز فایل، فرمت‌های مجاز)
6. پشتیبانی از prompt caching / context caching (در صورت وجود در آن مدل)
7. برای مدل‌های image: اندازه‌ها، نسبت تصویر، تعداد خروجی، فرمت پاسخ (base64/url)
8. برای embedding: ابعاد بردار، حداکثر طول ورودی، پشتیبانی از dimensions دلخواه
9. stop reasons/finish reasons و نگاشت صحیح آن‌ها بین فرمت TokenBazaar و فرمت native (Anthropic/Gemini) که هندلر ادعا می‌کند پیاده کرده

> **نکته‌ی مهم:** TokenBazaar احتمالاً خودش یک لایه‌ی aggregator/reseller روی این مدل‌هاست،
> نه ارائه‌دهنده‌ی اصلی. اگر مستندات خود TokenBazaar در دسترس است، آن را هم به‌عنوان یک
> لایه‌ی میانی چک کن، ولی مرجع نهایی برای صحت فیچرها/قابلیت‌های ذاتی هر مدل همان مستندات
> رسمی سازنده است — نه فرض یا حدس. اگر جایی TokenBazaar قابلیتی را که مدل اصلی پشتیبانی
> نمی‌کند ادعا کرده، یا برعکس قابلیتی موجود را در کاتالوگ گدار غایب گذاشته، صریح فلگ کن.

## مرحله‌ی ۲ — درستی پیاده‌سازی هندلرها (Go)

برای هر دو هندلر جدید (`tokenbazaar_messages_handler.go` و `tokenbazaar_generate_content_handler.go`)
و برای هندلرهای موجودی که تغییر کرده‌اند (`anthropic_native_handler.go`، `gemini_native_handler.go`،
`embeddings_handler.go`، `images_handler.go`، `responses_handler.go`):

- آیا ترجمه‌ی request/response بین فرمت native (Anthropic/Gemini) و فرمت واقعی TokenBazaar کامل و درست است؟ (نقشه‌ی فیلدها، nested content blocks، system prompt، multi-turn history، tool_use/tool_result blocks، multimodal content یعنی image/file parts)
- مدیریت streaming (SSE) درست پیاده شده؟ chunk‌های جزئی، event‌های خطا، پایان صحیح استریم، و صحت usage/token count نهایی؟
- مدیریت خطا: خطاهای TokenBazaar (rate limit، invalid request، auth) به درستی به فرمت خطای native نگاشت می‌شوند؟ context cancellation/timeout درست هندل شده؟
- آیا منطق مشترک بین این دو هندلر جدید و هندلرهای قبلی (OpenAI/Anthropic/Gemini native) کپی-پیست شده یا به‌درستی reuse شده؟ (این مستقیماً به ماژولاریتی مربوط است — پایین را ببین)
- Concurrency/race safety، عدم لو رفتن API key در لاگ‌ها یا پیام خطا، و صحت context propagation (cancellation, deadlines)

## مرحله‌ی ۳ — Registry، Routing، Pricing، Bootstrap

- `registry.go`: آیا ثبت پروایدر TokenBazaar و ۲۶ مدلش با الگوی پروایدرهای دیگر سازگار است؟ تغییرات در `registry_test.go` (۴۲۰+ خط) واقعاً رفتار جدید را پوشش می‌دهد یا صرفاً تست‌های موجود را تطبیق داده؟
- `routes_proxy.go` / `server.go` / `build_gateway.go`: مسیرهای جدید درست wiring شده‌اند؟ auth middleware، rate limiting، و سایر cross-cutting concernها برای این پروایدر هم اعمال می‌شود؟
- `provider_pricing_strategy.go`: فرمول محاسبه‌ی هزینه برای embeddings و image models درست است (واحد صحیح، گرد کردن صحیح، عدم دابل-کانت کش)؟ تست‌های pricing واقعاً edge caseها (صفر توکن، حجم خیلی بزرگ، تصاویر چندگانه) را پوشش می‌دهند یا فقط happy path؟

## مرحله‌ی ۴ — شبیه‌ساز silk-road و کیفیت تست‌ها

- `tokenbazaar.go` (۱۴۵۶ خط) در silk-road: آیا این mock service واقعاً رفتار API واقعی TokenBazaar/مدل‌های بالادستی را به اندازه‌ی کافی دقیق شبیه‌سازی می‌کند تا تست‌ها معنادار باشند، یا صرفاً حالت‌هایی را برمی‌گرداند که تست‌ها را سبز می‌کند (تست فریب‌دهنده / tautological test)؟
- تست‌های جدید (`tokenbazaar_test.go` با ۲۸۸۰ خط، handler testها، pricing testها) را از نظر **کیفیت — نه فقط حجم** — بررسی کن:
  - آیا assertionها معنادارند یا فقط status code/عدم پنیک را چک می‌کنند؟
  - پوشش مسیرهای خطا (4xx/5xx از TokenBazaar، timeout، پاسخ ناقص/بدشکل)
  - پوشش streaming واقعی (نه فقط non-stream)
  - پوشش tool calling و multimodal ورودی/خروجی
  - تست‌های تکراری/redundant که فقط حجم اضافه می‌کنند بدون افزایش پوشش
- مقایسه‌ی پوشش تست فعلی با چک‌لیست `plans/tokenbazaar-production-readiness.md` — هر آیتمی که در آن سند به‌عنوان الزامی ذکر شده ولی در کد/تست غایب است را صریح لیست کن.

## مرحله‌ی ۵ — کیفیت کد، ماژولاریتی، نگهداری‌پذیری

- اصول پایه (SRP، DRY، عدم duplication بین هندلرهای TokenBazaar و هندلرهای مشابه سایر پروایدرها)
- آیا افزودن مدل بیست‌وهفتم یا اندپوینت جدید در آینده، بدون دست‌زدن به منطق core، ممکن است؟ (extensibility)
- نام‌گذاری، خوانایی، error wrapping idiomatic در Go (`fmt.Errorf` با `%w`، سلسله‌مراتب خطا)
- سازگاری تغییرات `contracts/platform/tenant_provider_account.go` با بقیه‌ی سیستم (breaking change احتمالی؟)
- کامل بودن i18n (`en.json`/`fa.json`) و سازگاری `provider-connection-test-dialog`

---

## خروجی مورد انتظار

یک لیست از یافته‌ها، هرکدام شامل:

- شدت (Critical/High/Medium/Low/Nit)
- فایل:خط دقیق
- توضیح مشکل
- در صورت مغایرت با مستندات رسمی: نقل‌قول کوتاه یا اشاره‌ی دقیق از مستندات رسمی + لینک منبع
- پیشنهاد اصلاح مشخص (نه صرفاً «بهتره درست بشه»)

مرتب‌شده از شدیدترین به کم‌اهمیت‌ترین. اگر جایی مطمئن نیستی مستندات رسمی چه می‌گوید،
به‌جای حدس زدن صریح بنویس «نیاز به راستی‌آزمایی دستی» و دلیلش را بگو.
