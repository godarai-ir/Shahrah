'use client';

import Cal, { getCalApi } from '@calcom/embed-react';
import { Bot, Gauge, ShieldCheck } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import Link from 'next/link';
import { useEffect } from 'react';

import { iconWell, textLink } from './sectionType';
import { useTheme } from './ThemeProvider';

const calUiConfig = {
  cssVarsPerTheme: {
    // Cal.com takes literal colours: the design system's ink in each theme (DESIGN.md).
    light: { 'cal-brand': '#0f1b33' },
    dark: { 'cal-brand': '#f4f1ea' },
  },
  hideEventTypeDetails: false,
  layout: 'month_view' as const,
};

const demoPoints: { icon: LucideIcon; title: string; text: string }[] = [
  { icon: Bot, title: 'دستیاری برای یکی از تیم‌هایتان', text: 'روی کار واقعی همان تیم.' },
  { icon: ShieldCheck, title: 'گاردریل معنایی', text: 'روی نمونه‌ی داده‌ی خودتان.' },
  { icon: Gauge, title: 'مصرف و سقف هزینه', text: 'داشبورد و گزارش‌ها روی محیط واقعی.' },
];

export function DemoBooking() {
  const { resolvedTheme } = useTheme();

  useEffect(() => {
    async function configureCalendar() {
      const cal = await getCalApi({ namespace: '60min' });
      cal('ui', calUiConfig);
    }

    void configureCalendar();
  }, [resolvedTheme]);

  return (
    <main className="bg-site-ground px-4 py-10 sm:px-6 lg:py-20">
      <div className="mx-auto max-w-content">
        <header className="mx-auto mb-8 flex max-w-3xl flex-col items-center gap-4 text-center">
          <h1 className="m-0 text-balance text-h1 font-extrabold text-site-ink lg:text-display">گدارAI را روی کار خودتان ببینید.</h1>
          <p className="m-0 text-lead font-light text-site-ink-soft">
            یک جلسه‌ی ۶۰ دقیقه‌ای با تیم ما؛ فضای کاری، گاردریل و گزارش‌ها را روی سناریوی سازمانتان نشانتان می‌دهیم.
          </p>
        </header>

        <ul className="m-0 mb-8 grid list-none gap-3 p-0 sm:grid-cols-3 sm:gap-6">
          {demoPoints.map((point) => (
            <li key={point.title} className="flex items-start gap-3 rounded-card border border-site-line bg-site-surface p-5">
              <span className={`size-10 ${iconWell}`}>
                <point.icon aria-hidden="true" className="size-5" />
              </span>
              <span className="flex flex-col gap-1">
                <span className="text-label font-semibold text-site-ink">{point.title}</span>
                <span className="text-caption text-site-ink-soft">{point.text}</span>
              </span>
            </li>
          ))}
        </ul>

        <section
          aria-label="تقویم رزرو دموی گدارAI"
          className="min-h-180 overflow-hidden rounded-card border border-site-line bg-site-surface"
          dir="ltr"
        >
          <Cal
            namespace="60min"
            calLink="godarai/60min"
            style={{ width: '100%', height: '100%', minHeight: '720px', overflow: 'auto' }}
            config={{
              layout: 'month_view',
              useSlotsViewOnSmallScreen: 'true',
              theme: resolvedTheme,
            }}
          />
        </section>
        <p className="mt-6 text-center text-label text-site-ink-soft">
          ترجیح می‌دهید خودتان امتحان کنید؟{' '}
          <Link href="/onboarding" className={textLink}>
            شروع ۱۵ روز رایگان، با اعتبار اولیه
          </Link>
        </p>
      </div>
    </main>
  );
}
