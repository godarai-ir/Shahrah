'use client';

import { motion } from 'motion/react';
import { CreditCard, Network, DollarSign, Activity } from 'lucide-react';

const problems = [
  {
    icon: <CreditCard className="w-6 h-6" />,
    title: 'چه کاری با چه هزینه‌ای؟',
    description: 'کنترل هزینه‌ها، بودجه‌بندی، گزارش مصرف ارزی و پرداخت بین‌المللی کارهای ساده‌ای نیستند. هرچه مقیاس بزرگ‌تر، مشکل هم بزرگ‌تر.',
    color: 'text-red-500',
    bg: 'bg-red-50 dark:bg-red-500/10'
  },
  {
    icon: <Network className="w-6 h-6" />,
    title: 'چه کسی؟ چه مدلی؟ چقدر؟',
    description: 'افراد و تیم‌ها برای نیازهای متنوع به مدل‌های متنوع با الگوهای مصرف متفاوت نیاز دارند. تنظیم حدود دسترسی ضروری است.',
    color: 'text-orange-500',
    bg: 'bg-orange-50 dark:bg-orange-500/10'
  },
  {
    icon: <DollarSign className="w-6 h-6" />,
    title: 'چه مدلی با چه قابلیتی؟',
    description: 'مدل‌ها و ارائه‌دهندگان متنوع‌اند. شیوه استفاده همیشه یکسان نیست. مدیریت این تنوع هم آسان نیست.',
    color: 'text-rose-500',
    bg: 'bg-rose-50 dark:bg-rose-500/10'
  },
  {
    icon: <Activity className="w-6 h-6" />,
    title: 'مسئولیت پایداری با کیست؟',
    description: 'هوا همیشه آفتابی نیست. برای پایداری در شرایط بحران، از الان باید به زیرساخت قابل اتکا فکر کرد.',
    color: 'text-purple-500',
    bg: 'bg-purple-50 dark:bg-purple-500/10'
  }
];

export function Problem() {
  return (
    <section className="py-14 sm:py-20 lg:py-24 bg-slate-50 dark:bg-[#111827]/50 border-y border-gray-200 dark:border-gray-800 transition-colors">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center max-w-3xl mx-auto mb-10 sm:mb-16">
          <h2 className="text-2xl sm:text-4xl font-bold text-slate-900 dark:text-white mb-4 leading-tight">
            در استفاده از AI،{' '}
            <span className="bg-gradient-to-l from-indigo-500 to-purple-500 bg-clip-text text-transparent">
              هرگز
            </span>{' '}
            بی‌گدار به آب نزنید!
          </h2>
          <p className="text-sm sm:text-lg text-slate-600 dark:text-gray-400 leading-7 sm:leading-relaxed">
            مدل‌های هوش مصنوعی در محیط عملیاتی؟! بسی افتاده مشکل‌ها!
          </p>
        </div>

        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4 sm:gap-6">
          {problems.map((item, index) => (
            <motion.div
              key={index}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: index * 0.1 }}
              className="bg-white dark:bg-[#0B0F19] p-5 sm:p-6 rounded-2xl border border-gray-100 dark:border-gray-800 shadow-sm hover:shadow-md transition-shadow"
            >
              <div className={`w-12 h-12 rounded-xl flex items-center justify-center mb-6 ${item.bg} ${item.color}`}>
                {item.icon}
              </div>
              <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-3">
                {item.title}
              </h3>
              <p className="text-slate-600 dark:text-gray-400 text-sm leading-relaxed">
                {item.description}
              </p>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}

