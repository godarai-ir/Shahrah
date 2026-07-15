'use client';

import { Activity, KeyRound, Route, ShieldCheck, WalletCards } from 'lucide-react';
import { motion } from 'motion/react';

const cases = [
  {
    icon: <KeyRound className="w-8 h-8" />,
    title: 'پراکنده‌شدن کلیدها و دسترسی‌ها',
    desc: 'وقتی هر محصول یا تیم مستقیم به ارائه‌دهنده وصل می‌شود، کلیدها در سرویس‌ها و مخزن‌ها پخش می‌شوند. LLM Gateway مسیر دسترسی را متمرکز و قابل‌کنترل می‌کند.'
  },
  {
    icon: <WalletCards className="w-8 h-8" />,
    title: 'هزینه و مصرف نامشخص',
    desc: 'بدون سقف مصرف، گزارش توکن و تفکیک هزینه، تشخیص سهم هر تیم سخت می‌شود. گدارAI مصرف مدل‌ها را قابل‌پایش و قابل‌بودجه‌بندی می‌کند.'
  },
  {
    icon: <Route className="w-8 h-8" />,
    title: 'وابستگی به یک مدل یا ارائه‌دهنده',
    desc: 'اگر یک مدل کند، گران یا ناپایدار شود، تغییر مسیر در چند کدبیس پرهزینه است. Gateway امکان مسیریابی، مدل مجازی و تغییر مسیر کنترل‌شده را فراهم می‌کند.'
  },
  {
    icon: <Activity className="w-8 h-8" />,
    title: 'عیب‌یابی دشوار درخواست‌ها',
    desc: 'وقتی پاسخ کند یا خطا رخ می‌دهد، تیم باید بداند کدام مدل، کاربر، کلید یا مسیر درگیر بوده است. گزارش رخداد و Trace این مسیر را روشن می‌کند.'
  },
  {
    icon: <ShieldCheck className="w-8 h-8" />,
    title: 'کنترل‌نشدن داده و سیاست‌ها',
    desc: 'ورودی و خروجی مدل می‌تواند شامل داده حساس، محتوای ناخواسته یا فرمت نامعتبر باشد. گاردریل‌ها و سیاست‌های Gateway ریسک را قبل از رسیدن به محصول کم می‌کنند.'
  }
];

export function UseCases() {
  return (
    <section id="use-cases" className="py-14 sm:py-20 lg:py-24">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-10 sm:mb-16">
          <h2 className="text-2xl sm:text-4xl font-bold text-slate-900 dark:text-white mb-4 leading-tight">
            مشکلات رایج تیم‌ها و سازمان‌ها در استفاده از مدل‌های AI
          </h2>
          <p className="text-sm sm:text-lg text-slate-600 dark:text-gray-400 leading-7 sm:leading-relaxed">
            چالش‌هایی که با یک LLM Gateway می‌توان آن‌ها را از مسیر اتصال، کنترل مصرف، پایش و امنیت مدیریت کرد.
          </p>
        </div>

        <div className="flex flex-wrap justify-center gap-4 sm:gap-6">
          {cases.map((item, idx) => (
            <motion.div
              key={idx}
              initial={{ opacity: 0, scale: 0.95 }}
              whileInView={{ opacity: 1, scale: 1 }}
              viewport={{ once: true }}
              transition={{ duration: 0.4, delay: idx * 0.1 }}
              className="w-full sm:w-[calc(50%-12px)] lg:w-[calc(33.333%-16px)] bg-slate-50 dark:bg-[#111827] p-5 sm:p-8 rounded-3xl border border-gray-100 dark:border-gray-800 hover:shadow-xl transition-all group overflow-hidden relative"
            >
              <div className="absolute top-0 right-0 w-32 h-32 bg-indigo-500/5 rounded-full blur-3xl group-hover:bg-indigo-500/10 transition-colors pointer-events-none"></div>
              
              <div className="text-indigo-500 mb-6 group-hover:scale-110 transition-transform origin-right">
                {item.icon}
              </div>
              <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-3">
                {item.title}
              </h3>
              <p className="text-slate-600 dark:text-gray-400 leading-relaxed">
                {item.desc}
              </p>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}

