'use client';

import { motion } from 'motion/react';
import { Layers, Route, RefreshCcw, TrendingDown, Eye, Globe2, CreditCard, ShieldCheck } from 'lucide-react';

const features = [
  {
    icon: <Layers className="w-6 h-6" />,
    title: 'یکپارچه‌سازی آسان (Unified API)',
    description: 'تنها با یک API Key و تغییر Base URL، به ده‌ها مدل زبانی مختلف متصل شوید بدون نیاز به تغییر کدهای قبلی.'
  },
  {
    icon: <Route className="w-6 h-6" />,
    title: 'مسیریابی هوشمند (Smart Routing)',
    description: 'ارسال خودکار درخواست‌ها به سریع‌ترین و ارزان‌ترین پروایدر در دسترس بر اساس تنظیمات شما.'
  },
  {
    icon: <RefreshCcw className="w-6 h-6" />,
    title: 'تلاش مجدد و جایگزینی (Fallback)',
    description: 'در صورت قطعی یک پروایدر، شاهراه به‌طور خودکار درخواست را به مدل مشابه در سرویس‌دهنده دیگر ارسال می‌کند.'
  },
  {
    icon: <TrendingDown className="w-6 h-6" />,
    title: 'بهینه‌سازی هزینه‌ها (Cost Optimization)',
    description: 'ذخیره نتایج (Caching) و مدیریت هوشمند توکن‌ها برای کاهش چشمگیر هزینه‌های ارزی.'
  },
  {
    icon: <Eye className="w-6 h-6" />,
    title: 'نظارت و لاگ‌گیری (Observability)',
    description: 'داشبورد پیشرفته برای مشاهده دقیق مصرف، هزینه‌ها، زمان پاسخگویی و لاگ کامل درخواست‌ها.'
  },
  {
    icon: <Globe2 className="w-6 h-6" />,
    title: 'پشتیبانی از چندین سرویس‌دهنده',
    description: 'دسترسی همزمان به OpenAI, Anthropic, Google, Mistral و مدل‌های Open Source.'
  },
  {
    icon: <CreditCard className="w-6 h-6" />,
    title: 'پرداخت ریالی، مصرف ارزی',
    description: 'بدون نیاز به کارت‌های اعتباری بین‌المللی، حساب خود را به ریال شارژ کنید و از سرویس‌های ارزی استفاده کنید.'
  },
  {
    icon: <ShieldCheck className="w-6 h-6" />,
    title: 'ارتباط پایدار و بدون تحریم',
    description: 'تونل‌های امن و اختصاصی برای دور زدن محدودیت‌های تحریمی و اطمینان از پایداری ۱۰۰ درصدی سرویس.'
  }
];

export function Features() {
  return (
    <section id="features" className="py-24 bg-slate-50 dark:bg-[#111827]/30 border-y border-gray-200 dark:border-gray-800">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center max-w-3xl mx-auto mb-16">
          <h2 className="text-3xl sm:text-4xl font-bold text-slate-900 dark:text-white mb-4">
            امکانات و ویژگی‌های کلیدی
          </h2>
          <p className="text-lg text-slate-600 dark:text-gray-400">
            هرآنچه برای مدیریت یکپارچه، امن و مقرون‌به‌صرفه زیرساخت هوش مصنوعی نیاز دارید.
          </p>
        </div>

        <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6">
          {features.map((feature, idx) => (
            <motion.div
              key={idx}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.4, delay: idx * 0.05 }}
              className="bg-white dark:bg-[#0B0F19] p-6 rounded-2xl border border-gray-100 dark:border-gray-800 hover:border-indigo-500/50 dark:hover:border-indigo-500/50 transition-colors group"
            >
              <div className="w-12 h-12 rounded-xl bg-indigo-50 dark:bg-indigo-500/10 text-indigo-600 dark:text-indigo-400 flex items-center justify-center mb-6 group-hover:scale-110 transition-transform">
                {feature.icon}
              </div>
              <h3 className="text-lg font-bold text-slate-900 dark:text-white mb-2">
                {feature.title}
              </h3>
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

