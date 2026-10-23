import { Database, GitBranch, KeyRound, Layers, LayoutDashboard, Network, RefreshCw, Route, ScrollText, Server, TerminalSquare } from 'lucide-react';
import { ClosingBand, CodeWindow, FeatureRow, MockRows, PageHero, ProductStage, ProductWindow, Section, TileGrid } from '@/app/components/patterns';
import { sectionCopy, sectionHeading } from '@/app/components/sectionType';

const code = `from openai import OpenAI

client = OpenAI(
    api_key="YOUR_GODARAI_TOKEN",
    base_url="https://YOUR_WORKSPACE.godarai.ir/v1",
)

response = client.chat.completions.create(
    model="support-reply",
    messages=[{"role": "user", "content": "..."}],
)`;

// The second product's hub, for developers and product teams. It also carries the developer-agent content that
// used to live on /features/agents (now the workplace assistants page), in the #agents section.
export function AIGateway() {
  return (
    <main className="bg-site-ground">
      <PageHero
        tone="night"
        title="یک نشانی برای همه‌ی مدل‌ها. یک کلید برای همه‌ی دردسرها."
        lead="API سازگار با OpenAI برای بیش از ۱۲۰۰ مدل از ۲۱ ارائه‌دهنده. کلید، مسیریابی، هزینه و گزارش در همان مسیر."
        primary={{ href: '/onboarding', label: 'شروع با اعتبار هدیه' }}
        secondary={{ href: '/docs', label: 'مستندات' }}
        visual={
          <ProductStage layout="size-full p-6 sm:p-10">
            <CodeWindow title="اولین درخواست" meta="Python" code={code} />
          </ProductStage>
        }
      />

      <Section id="parts" layout="flex flex-col gap-8">
        <div className="flex flex-col gap-4">
          <h2 className={`max-w-3xl ${sectionHeading}`}>سه چیز برای اتصال عملیاتی.</h2>
          <p className={`max-w-2xl ${sectionCopy}`}>نشانی پایه‌ی درگاه، شناسه‌ی مدل و توکن دسترسی؛ بقیه‌اش با ماست.</p>
        </div>
        <TileGrid
          columns={3}
          items={[
            { key: 'api', icon: Network, title: 'یک API', text: 'سازگار با OpenAI؛ SDK رسمی همان کار می‌کند، فقط نشانی و کلید عوض می‌شود.' },
            { key: 'virtual', icon: Layers, title: 'مدل مجازی', text: 'یک نام پایدار در برنامه، و شما پشت آن مدل را عوض می‌کنید.' },
            { key: 'keys', icon: KeyRound, title: 'کلید جدا', text: 'توکن شخصی برای افراد، کلید مجازی برای سرویس‌ها.' },
          ]}
        />
      </Section>

      <FeatureRow
        id="routing"
        tone="panel"
        reverse
        title="مدل اول جواب نداد؟ مسیر جایگزین دارد."
        copy="مسیریابی اولویتی یا وزنی بین چند مقصد، با تلاش دوباره و مسیر جایگزین هنگام خطا؛ بدون تغییر در کد برنامه."
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="مدل مجازی: support-reply" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از مدل مجازی با دو مقصد و مسیر جایگزین.">
              <MockRows
                rows={[
                  { icon: Route, title: 'مقصد اول', text: 'مدل سبک', badge: { label: 'اولویت ۱', tone: 'neutral' } },
                  { icon: RefreshCw, title: 'مسیر جایگزین', text: 'مدل قوی، هنگام خطا', badge: { label: 'اولویت ۲', tone: 'neutral' } },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <FeatureRow
        id="cache"
        title="درخواست تکراری؟ از حافظه‌ی نهان."
        copy="پاسخ درخواست‌های تکراری، و در پلن‌هایی که دارند درخواست‌های هم‌معنا، از حافظه‌ی نهان می‌آید و دوباره به مدل نمی‌رود."
        link={{ href: '/features/cost-control', label: 'مدیریت هزینه' }}
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="حافظه‌ی نهان" meta="نمونه" size="m-auto max-w-md" label="نمونه: درخواست دوم از حافظه‌ی نهان پاسخ گرفته است.">
              <MockRows
                rows={[
                  { icon: Database, title: 'درخواست اول', text: 'به مدل رفت' },
                  { icon: Database, title: 'درخواست تکراری', text: 'همان پرسش', badge: { label: 'از حافظه‌ی نهان', tone: 'success' } },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <FeatureRow
        id="keys"
        tone="panel"
        reverse
        title="کلید خام ارائه‌دهنده هرگز به برنامه نمی‌رسد."
        copy="برنامه‌ها با کلید دسترسی مجازی کار می‌کنند و برای هر کلید سقف مصرف می‌گذارید. کلید ارائه‌دهنده در گدارAI می‌ماند."
        link={{ href: '/features/security', label: 'امنیت و اعتماد' }}
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="کلیدها" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از دو کلید دسترسی مجازی با سقف مصرف.">
              <MockRows
                rows={[
                  { icon: KeyRound, title: 'سرویس پشتیبانی', text: 'کلید دسترسی مجازی', badge: { label: 'سقف فعال', tone: 'success' } },
                  { icon: KeyRound, title: 'اپ موبایل', text: 'کلید دسترسی مجازی', badge: { label: 'سقف فعال', tone: 'success' } },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <FeatureRow
        id="logs"
        title="هر درخواست، با رد و هزینه‌اش."
        copy="گزارش رخداد، رد درخواست و هزینه‌ی هر فراخوانی؛ با سرآیند x-godarai-trace-id به سامانه‌ی پایش خودتان هم وصل می‌شود."
        link={{ href: '/features/observability', label: 'مشاهده‌پذیری' }}
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="گزارش رخداد" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از گزارش رخداد درگاه با سه درخواست.">
              <MockRows
                rows={[
                  { icon: ScrollText, title: 'support-reply', text: 'سرویس پشتیبانی', badge: { label: '۲۰۰', tone: 'success' } },
                  { icon: ScrollText, title: 'support-reply', text: 'مسیر جایگزین', badge: { label: '۲۰۰', tone: 'success' } },
                  { icon: ScrollText, title: 'summarize', text: 'اپ موبایل', badge: { label: '۴۲۹', tone: 'danger' } },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <Section id="agents" tone="panel" layout="flex flex-col gap-8">
        <div className="flex flex-col gap-4">
          <h2 className={`max-w-3xl ${sectionHeading}`}>ایجنت خودتان را می‌سازید؟ از همین درگاه.</h2>
          <p className={`max-w-2xl ${sectionCopy}`}>مستقل از چارچوب ایجنت؛ هر فراخوانی مدل از درگاه رد می‌شود و کنترل‌هایش را با خودش دارد.</p>
        </div>
        <TileGrid
          columns={2}
          size="compact"
          items={[
            { key: 'framework', icon: GitBranch, title: 'مستقل از چارچوب', text: 'هر کلاینت سازگار با OpenAI API کار می‌کند.' },
            { key: 'retry', icon: RefreshCw, title: 'تلاش دوباره و مسیر جایگزین', text: 'برای هر فراخوانی مدل در ایجنت.' },
            { key: 'trace', icon: ScrollText, title: 'رد هر فراخوانی', text: 'در کنار ردیابی خود ایجنت، با شناسه‌ی رد درخواست.' },
            { key: 'guard', icon: Server, title: 'گاردریل در مسیر', text: 'متن ورودی و خروجی ایجنت هم بررسی می‌شود.', href: '/features/guardrails' },
          ]}
        />
      </Section>

      <Section id="tools" layout="flex flex-col gap-8">
        <div className="flex flex-col gap-4">
          <h2 className={`max-w-3xl ${sectionHeading}`}>ابزارهای درگاه.</h2>
        </div>
        <TileGrid
          columns={4}
          size="compact"
          items={[
            { key: 'discovery', icon: Database, title: 'کشف مدل‌ها', text: 'جست‌وجو و مقایسه‌ی مدل‌ها', href: '/features/model-discovery' },
            { key: 'prompts', icon: TerminalSquare, title: 'رجیستری پرامپت', text: 'ساخت و نسخه‌بندی پرامپت', href: '/features/prompt-engineering' },
            { key: 'mcp', icon: Server, title: 'درگاه MCP', text: 'یک مسیر برای سرورهای MCP', href: '/features/mcp-gateway' },
            { key: 'dashboard', icon: LayoutDashboard, title: 'داشبورد', text: 'سنجه‌های هر مدل و کلید', href: '/dashboard' },
          ]}
        />
        <TileGrid
          columns={2}
          size="compact"
          items={[
            { key: 'litellm', icon: Layers, title: 'گدارAI و LiteLLM', text: 'تفاوت‌ها و زمان مهاجرت', href: '/vs/litellm' },
            { key: 'portkey', icon: Layers, title: 'گدارAI و Portkey', text: 'درگاه مستقل با استقرار داخلی', href: '/vs/portkey' },
          ]}
        />
      </Section>

      <ClosingBand
        title="اولین درخواست را امروز بفرستید."
        titleMuted="بقیه‌اش تکرار است."
        copy="با اعتبار هدیه شروع کنید؛ نشانی و کلید را عوض کنید و تمام."
        primary={{ href: '/onboarding', label: 'شروع با اعتبار هدیه' }}
        secondary={{ href: '/docs', label: 'مستندات' }}
      />
    </main>
  );
}
