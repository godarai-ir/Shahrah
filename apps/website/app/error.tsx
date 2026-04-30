'use client';

import { useEffect } from 'react';
import Link from 'next/link';

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error(error);
  }, [error]);

  return (
    <html lang="fa" dir="rtl">
      <body className="min-h-screen bg-white dark:bg-[#0B0F19] text-slate-900 dark:text-white">
        <main className="min-h-screen flex items-center justify-center px-4 py-12">
          <div className="max-w-md text-center">
            <h1 className="text-2xl font-bold mb-3">خطایی رخ داد</h1>
            <p className="text-slate-600 dark:text-slate-400 mb-6">
              در پردازش درخواست مشکلی پیش آمد. دوباره تلاش کنید.
            </p>
            <div className="flex items-center justify-center gap-3">
              <button
                onClick={reset}
                className="px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium"
              >
                تلاش مجدد
              </button>
              <Link
                href="/"
                className="px-4 py-2 rounded-lg border border-slate-300 dark:border-slate-700 text-sm font-medium"
              >
                بازگشت به صفحه اصلی
              </Link>
            </div>
          </div>
        </main>
      </body>
    </html>
  );
}
