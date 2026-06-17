'use client';

import { MessageSquare, FileText, Wrench, Rocket, Database } from 'lucide-react';
import { motion } from 'motion/react';

const cases = [
  {
    icon: <MessageSquare className="w-8 h-8" />,
    title: 'چت‌بات‌ها (Chatbots)',
    desc: 'پشتیبانی مشتریان هوشمند و ربات‌های تلگرامی با سرعت بالا و هزینه کم.'
  },
  {
    icon: <FileText className="w-8 h-8" />,
    title: 'تولید محتوا',
    desc: 'ایجاد خودکار مقالات، توضیحات محصول و پست‌های شبکه‌های اجتماعی.'
  },
  {
    icon: <Wrench className="w-8 h-8" />,
    title: 'ابزارهای داخلی سازمان',
    desc: 'اتوماسیون فرآیندها، خلاصه‌سازی اسناد و دستیار هوشمند کارمندان.'
  },
  {
    icon: <Rocket className="w-8 h-8" />,
    title: 'استارتاپ‌های هوش مصنوعی',
    desc: 'زیرساخت مطمئن برای محصولاتی که هسته آن‌ها بر پایه مدل‌های زبانی است.'
  },
  {
    icon: <Database className="w-8 h-8" />,
    title: 'پردازش داده‌ها',
    desc: 'استخراج اطلاعات از متون بدون ساختار، دسته‌بندی و تحلیل احساسات.'
  }
];

export function UseCases() {
  return (
    <section id="use-cases" className="py-14 sm:py-20 lg:py-24">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-10 sm:mb-16">
          <h2 className="text-2xl sm:text-4xl font-bold text-slate-900 dark:text-white mb-4 leading-tight">
            کاربردهای گُدار AI
          </h2>
          <p className="text-sm sm:text-lg text-slate-600 dark:text-gray-400 leading-7 sm:leading-relaxed">
            برای هر محصولی که به قدرت پردازش زبان طبیعی نیاز دارد.
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

