'use client';

import Cal, { getCalApi } from '@calcom/embed-react';
import { CalendarDays, Clock3, MessagesSquare } from 'lucide-react';
import { useEffect } from 'react';

import { useTheme } from './ThemeProvider';

const calUiConfig = {
  cssVarsPerTheme: {
    light: { 'cal-brand': '#00C9EB' },
    dark: { 'cal-brand': '#012651' },
  },
  hideEventTypeDetails: false,
  layout: 'month_view' as const,
};

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
    <main className="relative overflow-hidden px-4 py-16 sm:px-6 sm:py-20 lg:px-8">
      <div className="pointer-events-none absolute inset-x-0 top-0 -z-10 h-96 bg-gradient-to-b from-indigo-500/10 to-transparent" />
      <div className="mx-auto max-w-7xl">
        <header className="mx-auto mb-10 max-w-3xl text-center sm:mb-12">
          <div className="mb-5 inline-flex items-center gap-2 rounded-full border border-indigo-200 bg-indigo-50 px-4 py-2 text-sm font-semibold text-indigo-700 dark:border-indigo-500/20 dark:bg-indigo-500/10 dark:text-indigo-300">
            <CalendarDays aria-hidden="true" />
            رزرو جلسه خصوصی
          </div>
          <h1 className="mb-5 text-3xl font-black leading-relaxed text-slate-950 dark:text-white sm:text-4xl lg:text-5xl">
            دموی خصوصی گدارAI را رزرو کنید
          </h1>
          <p className="text-base leading-8 text-slate-600 dark:text-slate-300 sm:text-lg">
            در یک جلسه ۶۰ دقیقه‌ای، نیازهای فنی تیم شما را بررسی می‌کنیم و قابلیت‌های گدارAI را متناسب با سناریوی واقعی کسب‌وکارتان نمایش می‌دهیم.
          </p>
        </header>

        <div className="mb-8 grid gap-4 sm:grid-cols-2">
          <div className="flex items-start gap-3 rounded-xl border border-slate-200 bg-white p-4 shadow-sm dark:border-slate-800 dark:bg-slate-900">
            <Clock3 aria-hidden="true" className="mt-0.5 shrink-0 text-indigo-500" />
            <div>
              <h2 className="font-bold text-slate-950 dark:text-white">۶۰ دقیقه گفت‌وگوی متمرکز</h2>
              <p className="mt-1 text-sm leading-6 text-slate-600 dark:text-slate-400">بدون ارائه عمومی؛ جلسه بر اساس نیازها و پرسش‌های تیم شما پیش می‌رود.</p>
            </div>
          </div>
          <div className="flex items-start gap-3 rounded-xl border border-slate-200 bg-white p-4 shadow-sm dark:border-slate-800 dark:bg-slate-900">
            <MessagesSquare aria-hidden="true" className="mt-0.5 shrink-0 text-indigo-500" />
            <div>
              <h2 className="font-bold text-slate-950 dark:text-white">بررسی سناریوی واقعی شما</h2>
              <p className="mt-1 text-sm leading-6 text-slate-600 dark:text-slate-400">از معماری اتصال مدل‌ها تا هزینه، پایش، امنیت و مسیر استقرار را مرور می‌کنیم.</p>
            </div>
          </div>
        </div>

        <section
          aria-label="تقویم رزرو دموی گدارAI"
          className="min-h-[720px] overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-xl shadow-slate-950/5 dark:border-slate-800 dark:bg-slate-950"
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
      </div>
    </main>
  );
}
