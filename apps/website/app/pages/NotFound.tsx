'use client';

import { Home } from 'lucide-react';
import Link from 'next/link';

export function NotFound() {
  return (
    <div className="min-h-[calc(100vh-4rem)] flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
      <div className="text-center">
        <h1 className="text-9xl font-bold bg-gradient-to-l from-indigo-400 to-purple-500 bg-clip-text text-transparent mb-4">
          ۴۰۴
        </h1>
        <h2 className="text-3xl font-bold mb-4">صفحه پیدا نشد</h2>
        <p className="text-xl text-slate-600 dark:text-gray-400 mb-8 max-w-md mx-auto">
          متأسفانه صفحه‌ای که به دنبال آن هستید وجود ندارد یا منتقل شده است.
        </p>
        <div className="flex items-center justify-center gap-4">
          <Link
            href="/"
            className="flex items-center gap-2 px-6 py-3 rounded-lg bg-indigo-500 hover:bg-indigo-600 text-white font-medium transition-all shadow-[0_0_15px_rgba(99,102,241,0.3)]"
          >
            بازگشت به صفحه اصلی
            <Home className="w-5 h-5" />
          </Link>
        </div>
      </div>
    </div>
  );
}


