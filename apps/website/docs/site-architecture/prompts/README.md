# ترتیب اجرای پرامپت‌ها

هر پرامپت را در یک جلسهٔ جدا از ریشهٔ ریپو اجرا کنید. همه به `_shared-brief.md` تکیه دارند. پرسش‌های `../04-open-questions.md` پاسخ گرفته‌اند؛ سه مورد «هنوز باز» آن (عدد تأخیر گاردریل، همگام‌سازی کاربران در ابر، ایتا) تا پاسخ ادعا نمی‌شوند و جلوی اجرای هیچ پرامپتی را نمی‌گیرند.

> **۱ آبان ۱۴۰۵:** همهٔ پرامپت‌های ۰۰ تا ۲۴ روی دیزاین سیستم جدید (`DESIGN.md`) اجرا شدند. ۰۰ تا ۰۲ کامیت شده‌اند؛ بقیه هنوز کامیت نشده‌اند.

| موج | پرامپت | صفحه | پیش‌نیاز |
| --- | --- | --- | --- |
| پایه | `00-foundation` | توکن‌ها (تغییر نام به `site-*`)، پوسته، الگوها، `/_patterns` | — |
| پایه | `01-navigation-footer` | ناوبری و فوتر | ۰۰ |
| P0 | `02-home` | `/` | ۰۰، ۰۱ |
| P0 | `03-workplace` | `/workplace` | ۰۰، ۰۱ |
| P0 | `04-features-guardrails` | `/features/guardrails` | ۰۰، ۰۱ |
| P0 | `05-features-governance` | `/features/governance` (جدید) | ۰۰، ۰۱ |
| P0 | `06-features-observability` | `/features/observability` | ۰۰، ۰۱ |
| P0 | `07-features-cost-control` | `/features/cost-control` (جدید) | ۰۰، ۰۱ |
| P0 | `08-features-security` | `/features/security` | ۰۴، ۰۵ |
| P0 | `09-pricing` | `/pricing` | ۰۰، ۰۱ |
| P1 | `10-teams` | `/teams/*` (جدید) | ۰۲ |
| P1 | `16-features-ai-gateway` | `/features/ai-gateway` | ۰۰، ۰۱ — **پیش از ۱۱** |
| P1 | `11-features-agents` | `/features/agents` | ۱۶ |
| P1 | `12-integrations` | `/integrations` (جدید) | ۰۰، ۰۱ |
| P1 | `13-vs-chatgpt` | `/vs/chatgpt` (جدید) | ۰۴، ۰۷ |
| P1 | `14-demo` | `/demo` | ۰۰، ۰۱ |
| P1 | `15-on-premise` | `/on-premise` | ۰۸ |
| P2 | `17-features-model-discovery` | `/features/model-discovery` | ۱۶ |
| P2 | `18-features-prompt-engineering` | `/features/prompt-engineering` | ۱۶ |
| P2 | `19-features-mcp-gateway` | `/features/mcp-gateway` | ۱۲، ۱۶ |
| P2 | `20-dashboard` | `/dashboard` | ۰۶ |
| P2 | `21-solutions` | `/solutions/*` | ۰۰، ۰۱ |
| P2 | `22-vs-litellm-portkey` | `/vs/litellm`، `/vs/portkey` | ۱۳، ۱۶ |
| P2 | `23-about-contact` | `/about`، `/contact` | ۰۰، ۰۱ |
| P2 | `24-legal-build-guild` | `/privacy`، `/terms`، `/build-guild` | ۰۰ |

پرامپت‌های هم‌موجی که پیش‌نیاز مشترک ندارند و فایل مشترک ویرایش نمی‌کنند (مثلاً ۰۴، ۰۶، ۰۷، ۰۹) می‌توانند موازی اجرا شوند؛ هر صفحهٔ جدید فقط مسیر خودش را از `pendingRoutes` در `app/lib/navigation.ts` حذف می‌کند، پس برای ادغام، آن فایل را آخر از همه یکی کنید.
