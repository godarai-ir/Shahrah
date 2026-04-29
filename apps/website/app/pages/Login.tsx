'use client';

import { ArrowLeft } from 'lucide-react';

export function Login() {
  return (
    <div className="min-h-[calc(100vh-4rem)] flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full">
        <div className="text-center mb-8">
          <div className="w-16 h-16 rounded-2xl bg-indigo-500 flex items-center justify-center text-white font-bold text-3xl mx-auto mb-4">
            ش
          </div>
          <h1 className="text-3xl font-bold mb-2">ورود به شاهراه</h1>
          <p className="text-slate-600 dark:text-gray-400">
            به حساب کاربری خود وارد شوید
          </p>
        </div>

        <div className="bg-white dark:bg-gray-900 p-8 rounded-2xl border border-gray-200 dark:border-gray-800">
          <form className="space-y-6">
            <div>
              <label className="block text-sm font-medium mb-2">ایمیل</label>
              <input
                type="email"
                className="w-full px-4 py-3 rounded-lg border border-gray-200 dark:border-gray-800 bg-white dark:bg-[#0B0F19] focus:outline-none focus:ring-2 focus:ring-indigo-500"
                placeholder="email@example.com"
                dir="ltr"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-2">رمز عبور</label>
              <input
                type="password"
                className="w-full px-4 py-3 rounded-lg border border-gray-200 dark:border-gray-800 bg-white dark:bg-[#0B0F19] focus:outline-none focus:ring-2 focus:ring-indigo-500"
                placeholder="رمز عبور"
              />
            </div>
            <div className="flex items-center justify-between">
              <label className="flex items-center gap-2">
                <input type="checkbox" className="rounded border-gray-300" />
                <span className="text-sm">مرا به خاطر بسپار</span>
              </label>
              <a href="#" className="text-sm text-indigo-500 hover:text-indigo-600">
                فراموشی رمز عبور
              </a>
            </div>
            <button
              type="submit"
              className="w-full flex items-center justify-center gap-2 py-3 rounded-lg bg-indigo-500 hover:bg-indigo-600 text-white font-medium transition-all shadow-[0_0_15px_rgba(99,102,241,0.3)]"
            >
              ورود
              <ArrowLeft className="w-4 h-4" />
            </button>
          </form>

          <div className="mt-6 text-center">
            <p className="text-sm text-slate-600 dark:text-gray-400">
              حساب کاربری ندارید؟{' '}
              <a href="#" className="text-indigo-500 hover:text-indigo-600 font-medium">
                ثبت‌نام کنید
              </a>
            </p>
          </div>
        </div>

        <div className="mt-8 text-center">
          <p className="text-sm text-slate-600 dark:text-gray-400 mb-4">یا ادامه دهید با</p>
          <div className="flex gap-4">
            <button className="flex-1 py-3 px-4 rounded-lg border border-gray-200 dark:border-gray-800 hover:border-gray-300 dark:hover:border-gray-700 transition-colors">
              Google
            </button>
            <button className="flex-1 py-3 px-4 rounded-lg border border-gray-200 dark:border-gray-800 hover:border-gray-300 dark:hover:border-gray-700 transition-colors">
              GitHub
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

