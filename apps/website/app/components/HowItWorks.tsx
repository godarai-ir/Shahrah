'use client';

import { motion } from 'motion/react';
import { UserPlus, Key, Send, Cpu, BarChart } from 'lucide-react';

const steps = [
  { icon: <UserPlus />, title: 'ثبت‌نام', desc: 'ساخت اکانت در پنل شاهراه' },
  { icon: <Key />, title: 'دریافت API', desc: 'تولید کلید امن برای پروژه' },
  { icon: <Send />, title: 'ارسال درخواست', desc: 'اتصال به اندپوینت واحد' },
  { icon: <Cpu />, title: 'مدیریت هوشمند', desc: 'مسیریابی و اجرای درخواست' },
  { icon: <BarChart />, title: 'مشاهده داده‌ها', desc: 'مانیتورینگ و بررسی هزینه‌ها' },
];

export function HowItWorks() {
  return (
    <section className="py-24 bg-slate-50 dark:bg-[#111827]/30 border-y border-gray-200 dark:border-gray-800">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-16">
          <h2 className="text-3xl sm:text-4xl font-bold text-slate-900 dark:text-white mb-4">
            چگونه کار می‌کند؟
          </h2>
          <p className="text-lg text-slate-600 dark:text-gray-400">
            شروع کار با شاهراه تنها چند دقیقه زمان می‌برد.
          </p>
        </div>

        <div className="relative">
          {/* Connecting line for desktop */}
          <div className="hidden md:block absolute top-10 left-[10%] right-[10%] h-[2px] bg-gray-200 dark:bg-gray-800 z-0"></div>

          <div className="grid grid-cols-2 md:grid-cols-5 gap-6 md:gap-4 relative z-10">
            {steps.map((step, idx) => (
              <motion.div
                key={idx}
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.4, delay: idx * 0.1 }}
                className="flex flex-col items-center text-center"
              >
                <div className="w-20 h-20 rounded-full bg-white dark:bg-[#0B0F19] border-4 border-slate-50 dark:border-[#111827]/30 shadow-lg flex items-center justify-center text-indigo-500 mb-4 relative">
                  <div className="absolute -top-2 -right-2 w-6 h-6 rounded-full bg-indigo-500 text-white text-xs font-bold flex items-center justify-center border-2 border-white dark:border-[#0B0F19]">
                    {idx + 1}
                  </div>
                  {step.icon}
                </div>
                <h3 className="font-bold text-slate-900 dark:text-white mb-2">{step.title}</h3>
                <p className="text-sm text-slate-500 dark:text-gray-400">{step.desc}</p>
              </motion.div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}

