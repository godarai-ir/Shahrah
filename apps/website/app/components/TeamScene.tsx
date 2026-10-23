import { BookOpen, Database, FileText, Lock, MessageCircle, Scale } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import type { ReactNode } from 'react';
import type { Team } from '@/app/lib/teams';
import { ProductStage, ProductWindow } from './patterns';

// Drawn product scenes for the team tabs, built from the briefs in
// docs/site-architecture/illustrations/team-scenes.md. Each scene carries exactly one Signal Orange marker:
// the row the assistant wants the colleague to look at. Labelled «نمونه» until real screenshots arrive.

type Source = { icon: LucideIcon; label: string };

const barWidth = { full: 'w-full', long: 'w-5/6', mid: 'w-2/3', short: 'w-1/2', tiny: 'w-1/3' } as const;

function Bar({ width = 'full' }: { width?: keyof typeof barWidth }) {
  return <span className={`block h-1.5 rounded-full bg-site-line ${barWidth[width]}`} />;
}

function Marker() {
  return <span className="size-2 shrink-0 rounded-full bg-site-accent" />;
}

function Status({ tone, children }: { tone: 'success' | 'danger'; children: ReactNode }) {
  const color = tone === 'success' ? 'border-site-success text-site-success' : 'border-site-danger text-site-danger';

  return <span className={`shrink-0 rounded-full border px-2 py-0.5 text-caption ${color}`}>{children}</span>;
}

function Scene({ title, label, request, sources, children }: { title: string; label: string; request: ReactNode; sources: Source[]; children: ReactNode }) {
  return (
    <ProductStage layout="flex size-full items-center justify-center p-4 sm:p-8">
      <ProductWindow title={title} meta="نمونه" label={label} size="w-full max-w-md">
        <div className="flex flex-col gap-3 p-4 sm:p-5">
          <div className="ms-auto max-w-xs rounded-card bg-site-panel px-3 py-2 text-caption text-site-ink">{request}</div>
          <ul className="m-0 flex list-none flex-wrap gap-2 p-0">
            {sources.map((source) => (
              <li key={source.label} className="flex items-center gap-1.5 rounded-full border border-site-line px-2.5 py-1 text-caption text-site-ink-soft">
                <source.icon aria-hidden="true" className="size-3.5" />
                {source.label}
              </li>
            ))}
          </ul>
          <div className="rounded-card border border-site-line p-3 sm:p-4">{children}</div>
        </div>
      </ProductWindow>
    </ProductStage>
  );
}

function ArtifactTitle({ children }: { children: ReactNode }) {
  return <p className="m-0 mb-3 text-label font-semibold text-site-ink">{children}</p>;
}

function SalesScene() {
  const rows = [
    { term: 'آخرین مذاکره', value: 'تخفیف حجمی برای سفارش دوم', marked: false },
    { term: 'قول داده‌شده', value: 'ارسال پیش‌فاکتور تا پنجشنبه', marked: true },
    { term: 'آخرین پیام', value: 'پرسش درباره‌ی زمان تحویل', marked: false },
  ];

  return (
    <Scene
      title="دستیار تیم فروش"
      label="نمونه: دستیار فروش پیش از جلسه، سابقه‌ی مشتری را از CRM و پیام‌رسان‌ها خلاصه می‌کند و قول داده‌شده را علامت می‌زند."
      request="قبل از جلسه با شرکت نمونه، خلاصه‌ی سابقه‌اش را بده."
      sources={[
        { icon: Database, label: 'CRM از راه MCP' },
        { icon: MessageCircle, label: 'پیام‌رسان‌ها' },
      ]}
    >
      <ArtifactTitle>خلاصه‌ی پیش از جلسه · شرکت نمونه</ArtifactTitle>
      <dl className="m-0 flex flex-col divide-y divide-site-line">
        {rows.map((row) => (
          <div key={row.term} className="flex items-center gap-3 py-2">
            <dt className="w-24 shrink-0 text-caption text-site-ink-soft">{row.term}</dt>
            <dd className="m-0 flex min-w-0 flex-1 items-center justify-between gap-2 text-caption text-site-ink">
              <span className="truncate">{row.value}</span>
              {row.marked ? <Marker /> : null}
            </dd>
          </div>
        ))}
      </dl>
    </Scene>
  );
}

function SupportScene() {
  return (
    <Scene
      title="دستیار تیم پشتیبانی"
      label="نمونه: دستیار پشتیبانی برای یک تیکت فوری، از پایگاه دانش و تیکت‌های قبلی پیش‌نویس پاسخ می‌نویسد."
      request={
        <span className="flex items-center gap-2">
          <Marker />
          تیکت فوری: «سفارشم هنوز نرسیده.»
        </span>
      }
      sources={[
        { icon: BookOpen, label: 'پایگاه دانش' },
        { icon: MessageCircle, label: 'تیکت‌های قبلی' },
      ]}
    >
      <ArtifactTitle>پیش‌نویس پاسخ</ArtifactTitle>
      <p className="m-0 mb-3 text-caption text-site-ink">سلام، سفارش شما دیروز از انبار خارج شده و تا فردا می‌رسد.</p>
      <div className="flex flex-col gap-2">
        <Bar width="long" />
        <Bar width="mid" />
      </div>
      <p className="m-0 mt-3 border-t border-site-line pt-2 text-caption text-site-ink-soft">همکار پشتیبانی می‌خواند و می‌فرستد.</p>
    </Scene>
  );
}

function FinanceScene() {
  const rows = [
    { invoice: 'فاکتور ۱۰۴', order: 'سفارش ۱۰۴', ok: true, marked: false },
    { invoice: 'فاکتور ۱۰۵', order: 'سفارش ۱۰۵', ok: false, marked: true },
    { invoice: 'فاکتور ۱۰۶', order: 'سفارش ۱۰۶', ok: true, marked: false },
    { invoice: 'فاکتور ۱۰۷', order: 'سفارش ۱۰۷', ok: false, marked: false },
  ];

  return (
    <Scene
      title="دستیار تیم مالی"
      label="نمونه: دستیار مالی فاکتورهای ماه را با سفارش‌ها تطبیق می‌دهد و دو مغایرت را برای بررسی علامت می‌زند."
      request="فاکتورهای این ماه را با سفارش‌ها تطبیق بده."
      sources={[
        { icon: FileText, label: 'فایل‌های تیم' },
        { icon: Database, label: 'CRM از راه MCP' },
      ]}
    >
      <table className="w-full border-collapse text-caption">
        <thead>
          <tr className="text-site-ink-soft">
            <th className="pb-2 text-start font-medium">فاکتور</th>
            <th className="pb-2 text-start font-medium">سفارش</th>
            <th className="pb-2 text-end font-medium">وضعیت</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((row) => (
            <tr key={row.invoice} className="border-t border-site-line text-site-ink">
              <td className="py-2">
                <span className="flex items-center gap-2">
                  {row.marked ? <Marker /> : null}
                  {row.invoice}
                </span>
              </td>
              <td className="py-2">{row.order}</td>
              <td className="py-2 text-end">
                <Status tone={row.ok ? 'success' : 'danger'}>{row.ok ? 'تطبیق' : 'مغایرت'}</Status>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      <p className="m-0 mt-2 border-t border-site-line pt-2 text-caption text-site-ink-soft">دو مغایرت برای بررسی علامت خورد.</p>
    </Scene>
  );
}

function LegalScene() {
  const clauses = [
    { number: '۳', flagged: false, note: '' },
    { number: '۵', flagged: true, note: '' },
    { number: '۷', flagged: true, note: 'مدت فسخ با سیاست شرکت نمی‌خواند.' },
    { number: '۹', flagged: true, note: '' },
  ];

  return (
    <Scene
      title="دستیار تیم حقوقی"
      label="نمونه: دستیار حقوقی قرارداد را با چک‌لیست شرکت مرور می‌کند و سه بند را با دلیل علامت می‌زند."
      request="این قرارداد را با چک‌لیست شرکت مرور کن."
      sources={[
        { icon: FileText, label: 'قرارداد بارگذاری‌شده' },
        { icon: Scale, label: 'چک‌لیست شرکت' },
      ]}
    >
      <ol className="m-0 flex list-none flex-col gap-2.5 p-0">
        {clauses.map((clause) => (
          <li key={clause.number} className={`flex flex-col gap-1.5 border-s-2 ps-3 ${clause.flagged ? 'border-site-danger' : 'border-transparent'}`}>
            <span className="flex items-center gap-2 text-caption font-semibold text-site-ink">
              بند {clause.number}
              {clause.note ? <Marker /> : null}
            </span>
            <Bar width={clause.flagged ? 'long' : 'mid'} />
            {clause.note ? <span className="text-caption text-site-ink">{clause.note}</span> : null}
          </li>
        ))}
      </ol>
    </Scene>
  );
}

function HrScene() {
  return (
    <Scene
      title="دستیار منابع انسانی"
      label="نمونه: دستیار منابع انسانی به پرسش مرخصی از روی آیین‌نامه جواب می‌دهد و فقط اطلاعات خود همکار را می‌بیند."
      request="چند روز مرخصی برایم مانده؟"
      sources={[
        { icon: BookOpen, label: 'آیین‌نامه' },
        { icon: Lock, label: 'دسترسی' },
      ]}
    >
      <p className="m-0 text-caption text-site-ink">طبق آیین‌نامه، مرخصی سالانه‌ی شما تا پایان سال قابل استفاده است. مانده‌ی شما در پرونده‌ی خودتان ثبت شده است.</p>
      <div className="mt-3 flex flex-wrap items-center gap-2">
        <span className="flex items-center gap-1.5 rounded-full border border-site-line-strong px-2.5 py-1 text-caption text-site-ink">
          <Marker />
          آیین‌نامه، ماده‌ی مرخصی
        </span>
      </div>
      <p className="m-0 mt-3 flex items-center gap-2 border-t border-site-line pt-2 text-caption text-site-ink-soft">
        <Lock aria-hidden="true" className="size-3.5" />
        فقط اطلاعات خود شما
      </p>
    </Scene>
  );
}

function MarketingScene() {
  const days = [
    { day: 'شنبه', title: 'معرفی محصول', marked: false },
    { day: 'یکشنبه', title: 'نکته‌ی کاربردی', marked: false },
    { day: 'دوشنبه', title: 'تجربه‌ی مشتری', marked: true },
    { day: 'سه‌شنبه', title: 'پشت صحنه', marked: false },
    { day: 'چهارشنبه', title: 'پیشنهاد هفته', marked: false },
  ];

  return (
    <Scene
      title="دستیار تیم بازاریابی"
      label="نمونه: دستیار بازاریابی با راهنمای برند، پیش‌نویس پست‌های پنج روز هفته را آماده می‌کند و یکی را برای بازبینی علامت می‌زند."
      request="پست‌های این هفته را با راهنمای برند بنویس."
      sources={[
        { icon: BookOpen, label: 'راهنمای برند' },
        { icon: FileText, label: 'نتیجه‌ی کمپین قبلی' },
      ]}
    >
      <ul className="m-0 grid list-none grid-cols-2 gap-2 p-0 sm:grid-cols-3">
        {days.map((item) => (
          <li key={item.day} className={`flex flex-col gap-1.5 rounded-field border p-2 ${item.marked ? 'border-site-line-strong bg-site-panel' : 'border-site-line'}`}>
            <span className="flex items-center justify-between gap-1 text-caption text-site-ink-soft">
              {item.day}
              {item.marked ? <Marker /> : null}
            </span>
            <span className="truncate text-caption font-semibold text-site-ink">{item.title}</span>
            <Bar width="long" />
            {item.marked ? <span className="text-caption text-site-ink">آماده‌ی بازبینی</span> : <Bar width="short" />}
          </li>
        ))}
      </ul>
    </Scene>
  );
}

const scenes: Record<Team['id'], () => ReactNode> = {
  sales: SalesScene,
  support: SupportScene,
  finance: FinanceScene,
  legal: LegalScene,
  hr: HrScene,
  marketing: MarketingScene,
};

export function TeamScene({ id }: { id: Team['id'] }) {
  const Render = scenes[id];

  return <Render />;
}
