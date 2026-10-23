import { BookOpen, Bot, Brain, Cpu, FileText, MessageSquare, Plug, Wrench } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import Link from 'next/link';
import {
  ClosingBand,
  FeatureRow,
  MockRows,
  PageHero,
  PillarGrid,
  ProductStage,
  ProductWindow,
  Section,
  StepList,
  TrustStrip,
} from '@/app/components/patterns';
import { iconWell, sectionCopy, sectionHeading, textLink } from '@/app/components/sectionType';
import { isReady } from '@/app/lib/navigation';

type Part = { icon: LucideIcon; title: string; text: string; href?: string };

// Integrations and the assistants page link only once their routes exist.
const parts: Part[] = [
  { icon: MessageSquare, title: 'گفتگو', text: 'فضای گفتگوی سازمان با هر مدل، برای همه‌ی همکاران.' },
  { icon: Bot, title: 'دستیارها', text: 'برای هر تیم دستیار خودش را بسازید و با تیم به اشتراک بگذارید.', href: '/features/agents' },
  { icon: BookOpen, title: 'دانش', text: 'فایل‌ها و پایگاه دانش تیم را به دستیار بدهید.' },
  { icon: Plug, title: 'اتصال‌ها', text: 'Jira، Notion، Git، CRM و سرورهای MCP، و پیام‌رسان‌ها مانند تلگرام، واتساپ، بله و …', href: '/integrations' },
];

function PartTile({ part }: { part: Part }) {
  const href = part.href && isReady(part.href) ? part.href : undefined;
  const body = (
    <>
      <span className={`size-10 ${iconWell}`}>
        <part.icon aria-hidden="true" className="size-5" />
      </span>
      <span className="text-h3 font-semibold text-site-ink">{part.title}</span>
      <span className="text-body text-site-ink-soft">{part.text}</span>
    </>
  );
  const frame = 'flex h-full flex-col gap-3 rounded-card border border-site-line bg-site-surface p-6';

  return href ? (
    <Link
      href={href}
      className={`${frame} transition-colors duration-150 hover:bg-site-ground focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-site-focus`}
    >
      {body}
    </Link>
  ) : (
    <div className={frame}>{body}</div>
  );
}

export function Workplace() {
  return (
    <main className="bg-site-ground">
      <PageHero
        tone="night"
        title="فضای کاری هوش مصنوعی. برای همه، نه فقط تیم فنی."
        lead="همه‌ی همکاران با هر مدلی که لازم دارند گفتگو می‌کنند و هر تیم دستیار خودش را دارد. مدیر فناوری می‌بیند و تنظیم می‌کند."
        primary={{ href: '/onboarding', label: 'شروع با اعتبار هدیه' }}
        secondary={{ href: '/demo', label: 'درخواست دمو' }}
        visual={
          <ProductStage layout="size-full p-6 sm:p-10">
            <ProductWindow
              title="دستیارهای سازمان"
              meta="نمونه"
              label="نمونه‌ای از فهرست دستیارهای یک سازمان: دستیار پشتیبانی، فروش و منابع انسانی، هر کدام با تیم و ابزارهای خودش."
            >
              <MockRows
                rows={[
                  { icon: Bot, title: 'دستیار پشتیبانی', text: 'تیم پشتیبانی · CRM و پایگاه دانش', badge: { label: 'فعال', tone: 'success' } },
                  { icon: Bot, title: 'دستیار فروش', text: 'تیم فروش · CRM', badge: { label: 'فعال', tone: 'success' } },
                  { icon: Bot, title: 'دستیار منابع انسانی', text: 'تیم منابع انسانی · آیین‌نامه‌ها', badge: { label: 'پیش‌نویس', tone: 'neutral' } },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <Section id="suite" layout="flex flex-col gap-8">
        <div className="flex flex-col gap-4">
          <h2 className={`max-w-3xl ${sectionHeading}`}>چهار تکه، یک فضای کاری.</h2>
          <p className={`max-w-2xl ${sectionCopy}`}>گفتگو، دستیار، دانش و اتصال؛ همه در یک جا و همه زیر یک کنترل.</p>
        </div>
        <ul className="m-0 grid list-none grid-cols-1 gap-3 p-0 sm:gap-6 md:grid-cols-2">
          {parts.map((part) => (
            <li key={part.title}>
              <PartTile part={part} />
            </li>
          ))}
        </ul>
      </Section>

      <FeatureRow
        id="builder"
        tone="panel"
        reverse
        title="دستیار را تیم می‌سازد. نه تیم فنی."
        copy="بدون کد: دستورالعمل، مدل، دانش و ابزارها را انتخاب کنید. دستیار برای همه‌ی تیم آماده است."
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="ساخت دستیار" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از صفحه‌ی ساخت دستیار با چهار بخش: دستورالعمل، مدل، دانش و ابزارها.">
              <MockRows
                rows={[
                  { icon: FileText, title: 'دستورالعمل', text: 'به سؤال‌های مشتری مؤدب و کوتاه جواب بده.' },
                  { icon: Cpu, title: 'مدل', text: 'یکی از مدل‌های مجاز تیم' },
                  { icon: Brain, title: 'دانش', text: 'راهنمای محصول، سیاست بازگشت وجه' },
                  { icon: Wrench, title: 'ابزارها', text: 'CRM از راه MCP' },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <Section id="control" layout="flex flex-col gap-8">
        <div className="flex flex-col gap-4">
          <h2 className={`max-w-3xl ${sectionHeading}`}>هر قابلیت با کنترلش می‌آید.</h2>
          <p className={`max-w-2xl ${sectionCopy}`}>دستیار جدید یعنی پرسش جدید برای مدیر فناوری. جوابش از قبل اینجاست.</p>
        </div>
        <PillarGrid columns={4} size="compact" />
      </Section>

      <FeatureRow
        id="models"
        tone="panel"
        title="۱۲۰۰+ مدل. یک انتخاب درست."
        copy="بیش از ۱۲۰۰ مدل از ۲۱ ارائه‌دهنده. برای هر کار مدل مناسبش را بگذارید؛ به یک مدل یا یک ارائه‌دهنده بسته نیستید."
        link={{ href: '/features/model-discovery', label: 'کشف و مقایسه‌ی مدل‌ها' }}
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="مدل‌های مجاز تیم پشتیبانی" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از فهرست مدل‌های مجاز یک تیم از چند ارائه‌دهنده.">
              <MockRows
                rows={[
                  { title: 'GPT', text: 'OpenAI', badge: { label: 'مجاز', tone: 'success' } },
                  { title: 'Claude', text: 'Anthropic', badge: { label: 'مجاز', tone: 'success' } },
                  { title: 'Gemini', text: 'Google', badge: { label: 'مجاز', tone: 'success' } },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <Section id="local" layout="flex flex-col gap-8">
        <h2 className={`max-w-3xl ${sectionHeading}`}>ساخته برای کار در ایران.</h2>
        <TrustStrip />
      </Section>

      <Section id="start" tone="panel" layout="flex flex-col gap-8">
        <div className="flex flex-col gap-4">
          <h2 className={`max-w-3xl ${sectionHeading}`}>از کجا شروع کنیم؟ از همین سه قدم.</h2>
        </div>
        <StepList
          steps={[
            { title: 'فضای کاری بسازید', text: 'ثبت‌نام می‌کنید و فضای کاری سازمانتان همان لحظه آماده است.' },
            { title: 'یک دستیار بسازید', text: 'دستورالعمل، مدل، دانش و یک اتصال؛ برای شروع همین کافی است.' },
            { title: 'تیم را دعوت کنید', text: 'سقف مصرف بگذارید و بگذارید همکارها امتحان کنند.' },
          ]}
        />
        <a href="/docs" className={textLink}>
          راهنمای قدم‌به‌قدم در مستندات
        </a>
      </Section>

      <ClosingBand
        title="فضای کاری را بسازید."
        titleMuted="بقیه‌اش با ماست."
        copy="با اعتبار هدیه وارد پنل شوید؛ تاریخ انقضا ندارد. اول امتحان کنید، بعد تصمیم بگیرید."
        primary={{ href: '/onboarding', label: 'شروع با اعتبار هدیه' }}
        secondary={{ href: '/demo', label: 'درخواست دمو' }}
      />
    </main>
  );
}
