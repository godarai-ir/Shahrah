'use client';

import { Braces, FileJson, ImageIcon, MessageSquare, Mic, Search } from 'lucide-react';

const endpoints = [
  {
    icon: <MessageSquare className="h-6 w-6" />,
    title: 'Chat Completions',
    description: 'برای چت، تولید متن، ابزارها و خروجی ساختاریافته با قرارداد سازگار با OpenAI.',
  },
  {
    icon: <FileJson className="h-6 w-6" />,
    title: 'Responses API',
    description: 'برای تیم‌هایی که جریان‌های جدیدتر تولید پاسخ، ابزارها و ورودی چندبخشی را ترجیح می‌دهند.',
  },
  {
    icon: <ImageIcon className="h-6 w-6" />,
    title: 'Image APIs',
    description: 'تولید، ویرایش و تغییر تصویر از همان مسیر درگاه، همراه با کنترل مصرف و گزارش رخداد.',
  },
  {
    icon: <Braces className="h-6 w-6" />,
    title: 'Embeddings',
    description: 'تبدیل متن به بردار برای جست‌وجوی معنایی، RAG، دسته‌بندی و تحلیل داده‌های متنی.',
  },
  {
    icon: <Mic className="h-6 w-6" />,
    title: 'Audio APIs',
    description: 'تبدیل گفتار به متن، متن به گفتار و ترجمه صوتی برای تجربه‌های صوتی و پشتیبانی.',
  },
  {
    icon: <Search className="h-6 w-6" />,
    title: 'Rerank و Moderation',
    description: 'بهبود ترتیب نتایج بازیابی و بررسی ایمنی محتوا پیش از نمایش یا پردازش نهایی.',
  },
];

export function Developer() {
  return (
    <section id="developers" className="py-14 sm:py-20 lg:py-24 overflow-hidden">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="mx-auto max-w-3xl text-center">
          <h2 className="mb-4 text-2xl font-bold leading-tight text-slate-900 dark:text-white sm:text-4xl">
            اندپوینت‌های قابل استفاده در گدارAI
          </h2>
          <p className="mb-8 text-sm leading-7 text-slate-600 dark:text-gray-400 sm:text-lg sm:leading-relaxed">
            از تولید متن و پاسخ‌های ساختاریافته تا تصویر، صوت، Embedding و کنترل محتوا را از یک مسیر واحد فراخوانی کنید و همان‌جا مصرف، هزینه و رخدادها را ببینید.
          </p>
        </div>

        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {endpoints.map((endpoint) => (
            <div
              key={endpoint.title}
              className="rounded-2xl border border-gray-100 bg-white p-5 transition-colors hover:border-indigo-500/50 dark:border-gray-800 dark:bg-[#0B0F19] dark:hover:border-indigo-500/50"
            >
              <div className="mb-5 flex h-12 w-12 items-center justify-center rounded-xl bg-indigo-50 text-indigo-600 dark:bg-indigo-500/10 dark:text-indigo-400">
                {endpoint.icon}
              </div>
              <h3 className="mb-2 text-lg font-bold text-slate-900 dark:text-white" dir="ltr">
                {endpoint.title}
              </h3>
              <p className="text-sm leading-7 text-slate-600 dark:text-gray-400">
                {endpoint.description}
              </p>
            </div>
          ))}
        </div>

        <div className="mt-8 flex justify-center">
          <a href="/docs/supported-apis" className="inline-flex min-h-12 w-full items-center justify-center rounded-xl bg-slate-900 px-6 py-3 font-medium text-white transition-colors hover:bg-slate-800 dark:bg-white dark:text-slate-900 dark:hover:bg-gray-100 sm:w-auto">
            مطالعه مستندات API
          </a>
        </div>
      </div>
    </section>
  );
}
