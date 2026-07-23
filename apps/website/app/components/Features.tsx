'use client';

import { motion } from 'motion/react';
import { Layers, Route, RefreshCcw, TrendingDown, Eye, Globe2, CreditCard, ShieldCheck } from 'lucide-react';

const features = [
  {
    icon: <Layers className="w-6 h-6" />,
    title: 'API سازگار با OpenAI',
    description: 'پس از دریافت کلید دسترسی گدارAI می‌توانید به سادگی به بیش از ۱۲۰۰ مدل هوش مصنوعی از ۲۱ ارائه‌دهنده معتبر متصل شوید.'
  },
  {
    icon: <Route className="w-6 h-6" />,
    title: 'مسیریابی قابل تنظیم',
    description: 'می‌توانید مدل مجازی بسازید، مدل‌ها را اولویت‌بندی کنید و برای هر مدل وزن یا اولویت متفاوتی تعریف کنید.'
  },
  {
    icon: <RefreshCcw className="w-6 h-6" />,
    title: 'تلاش دوباره و مسیر جایگزین',
    description: 'برای خطا یا کندی یک ارائه‌دهنده، ارائه‌دهنده‌های جایگزین تعریف کنید و رفتار هر درخواست را در گزارش رخدادها ببینید.'
  },
  {
    icon: <TrendingDown className="w-6 h-6" />,
    title: 'کنترل هزینه و مصرف',
    description: 'مصرف توکن، سقف هزینه و حافظه نهان را در سطح تیم یا سرویس پایش کنید تا هزینه‌ها قابل پیش‌بینی‌تر بمانند.'
  },
  {
    icon: <Eye className="w-6 h-6" />,
    title: 'پایش و مشاهده‌پذیری',
    description: 'گزارش رخداد، رد درخواست، سنجه مصرف و تاخیر را کنار هم ببینید و علت خطاها را سریع‌تر پیدا کنید.'
  },
  {
    icon: <Globe2 className="w-6 h-6" />,
    title: 'چند حساب ارائه‌دهنده',
    description: 'اتصال به ارائه‌دهنده‌های مختلف را از یک فضای کاری مدیریت کنید و مدل مناسب هر سناریو را جدا انتخاب کنید.'
  },
  {
    icon: <CreditCard className="w-6 h-6" />,
    title: 'پرداخت ریالی',
    description: 'کیف پول را به تومان شارژ کنید و مصرف مدل‌ها را با گزارش مالی و سقف‌های قابل تنظیم کنترل کنید.'
  },
  {
    icon: <ShieldCheck className="w-6 h-6" />,
    title: 'کاهش ریسک اتصال مستقیم',
    description: 'به‌جای پخش کلیدها در چند سرویس، اتصال مدل‌ها را از مسیر درگاه و با سیاست‌های دسترسی مشخص انجام دهید.'
  }
];

export function Features() {
  return (
    <section id="features" className="py-14 sm:py-20 lg:py-24 bg-slate-50 dark:bg-[#111827]/30 border-y border-gray-200 dark:border-gray-800">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center max-w-3xl mx-auto mb-10 sm:mb-16">
          <h2 className="text-2xl sm:text-4xl font-bold text-slate-900 dark:text-white mb-4 leading-tight">
            امکانات کلیدی
          </h2>
          <p className="text-sm sm:text-lg text-slate-600 dark:text-gray-400 leading-7 sm:leading-relaxed">
            ابزارهای اصلی برای اتصال، مسیریابی، کنترل هزینه و پایش مصرف مدل‌ها در یک مسیر واحد.
          </p>
        </div>

        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4 sm:gap-6">
          {features.map((feature, idx) => (
            <motion.div
              key={idx}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.4, delay: idx * 0.05 }}
              className="bg-white dark:bg-[#0B0F19] p-5 sm:p-6 rounded-2xl border border-gray-100 dark:border-gray-800 hover:border-indigo-500/50 dark:hover:border-indigo-500/50 transition-colors group"
            >
              <div className="mb-5 flex items-center gap-3">
                <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-indigo-50 text-indigo-600 transition-transform group-hover:scale-110 dark:bg-indigo-500/10 dark:text-indigo-400">
                  {feature.icon}
                </div>
                <h3 className="text-lg font-bold text-slate-900 dark:text-white">
                  {feature.title}
                </h3>
              </div>
              <p className="text-slate-600 dark:text-gray-400 text-sm leading-relaxed">
                {feature.description}
              </p>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}

