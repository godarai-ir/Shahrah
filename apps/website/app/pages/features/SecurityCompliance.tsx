import { ArrowLeft, Cloud, Cpu, KeyRound, ScrollText, Server, ShieldCheck, User } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import { ClosingBand, ControlChecklist, FAQ, FeatureRow, MockRows, PageHero, ProductStage, ProductWindow, Section, TileGrid } from '@/app/components/patterns';
import { iconWell, sectionCopy, sectionHeading } from '@/app/components/sectionType';

type FlowStep = { icon: LucideIcon; title: string; note: string };

const flow: FlowStep[] = [
  { icon: User, title: 'کاربر', note: 'درخواست می‌فرستد' },
  { icon: ShieldCheck, title: 'گاردریل ورودی', note: 'داده‌ی محرمانه را می‌گیرد' },
  { icon: Cpu, title: 'مدل', note: 'فقط مدل‌های مجاز تیم' },
  { icon: ShieldCheck, title: 'گاردریل خروجی', note: 'پاسخ را بررسی می‌کند' },
  { icon: User, title: 'کاربر', note: 'پاسخ را می‌گیرد' },
];

// The data path, read right to left in RTL and top to bottom on mobile; every step is logged.
function DataFlow() {
  return (
    <ol className="m-0 grid list-none grid-cols-1 gap-3 p-0 md:grid-cols-5">
      {flow.map((step, index) => (
        <li key={`${step.title}-${index}`} className="relative flex flex-col items-center gap-2 rounded-card border border-site-line bg-site-surface p-4 text-center">
          <span className={`size-10 ${iconWell}`}>
            <step.icon aria-hidden="true" className="size-5" />
          </span>
          <span className="text-label font-semibold text-site-ink">{step.title}</span>
          <span className="text-caption text-site-ink-soft">{step.note}</span>
          {index < flow.length - 1 ? (
            <ArrowLeft aria-hidden="true" className="absolute -bottom-3 left-1/2 size-4 -translate-x-1/2 rotate-90 text-site-ink-soft md:-left-3 md:bottom-auto md:top-1/2 md:translate-x-0 md:-translate-y-1/2 md:rotate-0" />
          ) : null}
        </li>
      ))}
    </ol>
  );
}

export function SecurityCompliance() {
  return (
    <main className="bg-site-ground">
      <PageHero
        title="داده و دسترسی‌ها دست شماست. ما فقط نگهبانیم."
        lead="این صفحه برای کسی است که باید پیش از تأیید بداند داده کجا می‌رود، چه کسی آن را می‌بیند و چه چیزی ثبت می‌شود."
        primary={{ href: '/demo', label: 'درخواست دمو' }}
        secondary={{ href: '/on-premise', label: 'استقرار در سازمان' }}
      />

      <Section id="flow" tone="panel" layout="flex flex-col gap-8">
        <div className="flex flex-col gap-4">
          <h2 className={`max-w-3xl ${sectionHeading}`}>مسیر داده، قدم‌به‌قدم.</h2>
          <p className={`max-w-2xl ${sectionCopy}`}>از کاربر تا مدل و برگشت. هر قدم در گزارش رخداد ثبت می‌شود.</p>
        </div>
        <DataFlow />
      </Section>

      <Section id="where" layout="flex flex-col gap-8">
        <div className="flex flex-col gap-4">
          <h2 className={`max-w-3xl ${sectionHeading}`}>کجا اجرا شود؟ شما انتخاب می‌کنید.</h2>
        </div>
        <TileGrid
          columns={2}
          items={[
            { key: 'cloud', icon: Cloud, title: 'ابر گدارAI', text: 'همین امروز و بدون نصب، با اشتراک ماهانه.' },
            { key: 'onprem', icon: Server, title: 'سرورهای خودتان', text: 'نصب داخل زیرساخت سازمان (Kubernetes) با قرارداد سالانه، و ورود با حساب سازمانی.', href: '/on-premise' },
          ]}
        />
      </Section>

      <FeatureRow
        id="credentials"
        tone="panel"
        reverse
        title="کلید خام ارائه‌دهنده از گدارAI بیرون نمی‌رود."
        copy="کلید ارائه‌دهنده‌ی مدل در گدارAI می‌ماند و به هیچ برنامه‌ای داده نمی‌شود. افراد توکن شخصی دارند، سرویس‌ها کلید مجازی؛ هر کدام با دسترسی و سقف خودش."
        link={{ href: '/features/governance', label: 'حاکمیت' }}
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="اعتبارنامه‌ها" meta="نمونه" size="m-auto max-w-md" label="نمونه: کلید ارائه‌دهنده در گدارAI می‌ماند و برنامه‌ها با توکن یا کلید مجازی کار می‌کنند.">
              <MockRows
                rows={[
                  { icon: KeyRound, title: 'کلید ارائه‌دهنده', text: 'فقط در گدارAI', badge: { label: 'پنهان', tone: 'neutral' } },
                  { icon: KeyRound, title: 'توکن دسترسی شخصی', text: 'برای افراد', badge: { label: 'PAT', tone: 'neutral' } },
                  { icon: KeyRound, title: 'کلید دسترسی مجازی', text: 'برای سرویس‌ها', badge: { label: 'VAT', tone: 'neutral' } },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <FeatureRow
        id="logs"
        title="گزارش‌ها هم قاعده دارند."
        copy="هر درخواست، پاسخ و تغییر تنظیمات ثبت می‌شود و مدت نگهداری گزارش‌ها را خودتان تعیین می‌کنید."
        link={{ href: '/features/observability', label: 'مشاهده‌پذیری' }}
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="نگهداری گزارش‌ها" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از تنظیم مدت نگهداری گزارش رخداد و رد درخواست.">
              <MockRows
                rows={[
                  { icon: ScrollText, title: 'گزارش رخداد', text: 'مدت نگهداری به انتخاب شما' },
                  { icon: ScrollText, title: 'رد درخواست', text: 'مدت نگهداری به انتخاب شما' },
                  { icon: ScrollText, title: 'تغییر تنظیمات', text: 'چه کسی، چه چیزی، کی' },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <Section id="checklist" tone="panel" layout="flex flex-col gap-8">
        <div className="flex flex-col gap-4">
          <h2 className={`max-w-3xl ${sectionHeading}`}>فهرست امنیت، یک‌جا.</h2>
        </div>
        <ControlChecklist
          groups={[
            { title: 'محافظت', surface: 'surface', items: [{ label: 'گاردریل معنایی روی ورودی، خروجی و فایل‌ها', href: '/features/guardrails' }, { label: 'گدارAI با داده‌ی شما مدل آموزش نمی‌دهد' }] },
            { title: 'دسترسی', surface: 'sand', items: [{ label: 'نقش و دسترسی هر تیم', href: '/features/governance' }, { label: 'کلید جدا برای انسان و سرویس' }] },
            { title: 'ثبت', surface: 'surface', items: [{ label: 'ثبت هر درخواست، پاسخ و تغییر تنظیمات' }, { label: 'مدت نگهداری قابل تنظیم' }] },
            { title: 'اجرا', surface: 'sand', items: [{ label: 'سقف مصرف هر تیم', href: '/features/cost-control' }, { label: 'اجرا روی سرورهای سازمان', href: '/on-premise' }] },
          ]}
        />
      </Section>

      <FAQ
        title="سؤال‌های مسئول امنیت."
        items={[
          { question: 'آیا با داده‌ی ما مدل آموزش می‌بینید؟', answer: 'نه. گدارAI با داده‌ی شما مدل آموزش نمی‌دهد.' },
          { question: 'گزارش‌ها چه مدت نگه داشته می‌شوند؟', answer: 'مدتش را خودتان تنظیم می‌کنید.' },
          { question: 'ورود با حساب سازمانی (SSO) دارید؟', answer: 'در استقرار داخل سازمان، بله.', link: { href: '/on-premise', label: 'استقرار در سازمان' } },
          { question: 'تغییر تنظیمات ثبت می‌شود؟', answer: 'بله: چه کسی، چه چیزی، کی.' },
        ]}
      />

      <ClosingBand
        title="پیش از تأیید، از نزدیک ببینید."
        titleMuted="نه از روی بروشور."
        copy="در دمو، مسیر داده و گزارش‌ها را روی محیط واقعی نشانتان می‌دهیم."
        primary={{ href: '/demo', label: 'درخواست دمو' }}
        secondary={{ href: '/on-premise', label: 'استقرار در سازمان' }}
      />
    </main>
  );
}
