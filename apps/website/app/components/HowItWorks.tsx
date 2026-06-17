'use client';

import { motion } from 'motion/react';
import { UserPlus, Key, Send, Cpu, BarChart } from 'lucide-react';

const steps = [
  { icon: <UserPlus />, title: 'ثبت‌نام', desc: 'ساخت اکانت در پنل گُدار AI' },
  { icon: <Key />, title: 'دریافت API', desc: 'تولید کلید امن برای پروژه' },
  { icon: <Send />, title: 'ارسال درخواست', desc: 'اتصال به اندپوینت واحد' },
  { icon: <Cpu />, title: 'مدیریت هوشمند', desc: 'مسیریابی و اجرای درخواست' },
  { icon: <BarChart />, title: 'مشاهده داده‌ها', desc: 'مانیتورینگ و بررسی هزینه‌ها' },
];

export function HowItWorks() {
  return (
    <section className="py-14 sm:py-20 lg:py-24 bg-slate-50 dark:bg-[#111827]/30 border-y border-gray-200 dark:border-gray-800">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-10 sm:mb-16">
          <h2 className="text-2xl sm:text-4xl font-bold text-slate-900 dark:text-white mb-4 leading-tight">
            چگونه کار می‌کند؟
          </h2>
          <p className="text-sm sm:text-lg text-slate-600 dark:text-gray-400 leading-7 sm:leading-relaxed">
            شروع کار با گُدار AI تنها چند دقیقه زمان می‌برد.
          </p>
        </div>

        <div className="relative">
          {/* Connecting line for desktop */}
          <div className="hidden md:block absolute top-12 lg:top-14 left-[10%] right-[10%] h-[2px] bg-gray-200 dark:bg-gray-800 z-0"></div>

          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-5 gap-6 md:gap-4 lg:gap-6 relative z-10">
            {steps.map((step, idx) => (
              <motion.div
                key={idx}
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.4, delay: idx * 0.1 }}
                className="mx-auto flex max-w-xs flex-col items-center text-center"
              >
                <div className="relative mb-4 flex h-20 w-20 items-center justify-center rounded-full border-4 border-slate-50 bg-white text-indigo-500 shadow-lg dark:border-[#111827]/30 dark:bg-[#0B0F19] md:h-24 md:w-24 lg:h-28 lg:w-28">
                  <div className="absolute -top-3 right-0 flex h-8 w-8 items-center justify-center rounded-full border-2 border-white bg-indigo-500 text-sm font-black text-white dark:border-[#0B0F19] md:-top-4 md:h-10 md:w-10 md:text-base">
                    {(idx + 1).toLocaleString('fa-IR')}
                  </div>
                  <div className="scale-100 md:scale-125 lg:scale-150">
                    {step.icon}
                  </div>
                </div>
                <h3 className="mb-2 text-base font-bold text-slate-900 dark:text-white md:text-lg">{step.title}</h3>
                <p className="text-sm leading-7 text-slate-500 dark:text-gray-400 md:text-base">{step.desc}</p>
              </motion.div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}

