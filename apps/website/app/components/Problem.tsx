'use client';

import { motion } from 'motion/react';
import { CreditCard, Network, DollarSign, Activity } from 'lucide-react';

const problems = [
  {
    icon: <CreditCard className="w-6 h-6" />,
    title: 'پرداخت و حسابداری پیچیده',
    description: 'مدیریت پرداخت ارزی، شارژ حساب‌های خارجی و تطبیق آن با هزینه ریالی تیم محصول زمان‌بر است.',
    color: 'text-red-500',
    bg: 'bg-red-50 dark:bg-red-500/10'
  },
  {
    icon: <Network className="w-6 h-6" />,
    title: 'تنوع ارائه‌دهندگان و مدل‌ها',
    description: 'هر ارائه‌دهنده مدل، کلید، نشانی و محدودیت خودش را دارد و تغییر مدل در کد محصول پرهزینه می‌شود.',
    color: 'text-orange-500',
    bg: 'bg-orange-50 dark:bg-orange-500/10'
  },
  {
    icon: <DollarSign className="w-6 h-6" />,
    title: 'مصرف نامشخص در تیم‌ها',
    description: 'وقتی چند تیم از یک مدل استفاده می‌کنند، بدون سقف هزینه و گزارش مصرف نمی‌توان هزینه هر سرویس را دقیق دید.',
    color: 'text-rose-500',
    bg: 'bg-rose-50 dark:bg-rose-500/10'
  },
  {
    icon: <Activity className="w-6 h-6" />,
    title: 'نبود پایش قابل اتکا',
    description: 'برای عیب‌یابی پاسخ کند یا خطای مدل، تیم به گزارش رخداد، رد درخواست و سنجه‌های مصرف نیاز دارد.',
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
            در استفاده از AI بی‌گدار به آب نزنید!
          </h2>
          <p className="text-sm sm:text-lg text-slate-600 dark:text-gray-400 leading-7 sm:leading-relaxed">
            استفاده از AI در محیط عملیاتی به سطح متفاوتی از مشاهده‌پذیری، کنترل و امنیت نیازمند است. می‌توانید آنها را به گدارAI بسپارید.
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

