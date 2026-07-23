'use client';

import { motion } from 'motion/react';
import { AppWindow, Globe2 } from 'lucide-react';
import Image from 'next/image';

export function Solution() {
  return (
    <section className="py-14 sm:py-20 lg:py-24 relative overflow-hidden">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center max-w-3xl mx-auto mb-10 sm:mb-16">
          <h2 className="text-2xl sm:text-4xl font-bold text-slate-900 dark:text-white mb-4 leading-tight">
            گدار، جای کم‌عمق رودخانه که می‌توان پهنای آن را بدون شنا کردن پیمود.
            <cite className="mt-3 block text-sm font-medium not-italic text-indigo-600 dark:text-indigo-400">
              (فرهنگ معین)
            </cite>
          </h2>
          <p className="text-sm sm:text-lg text-slate-600 dark:text-gray-400 leading-7 sm:leading-relaxed">
            به بیش از ۲۷۰۰ مدل از ۲۱ ارائه‌دهنده معتبر جهانی دسترسی بی‌دغدغه دارید. از کنترل مصرف و امنیت و پایداری تا نظارت و هشدار و عیب‌یابی.
          </p>
        </div>

        <div className="relative max-w-6xl mx-auto py-4 sm:py-12">
          {/* Connecting Lines (Desktop) */}
          <div className="hidden md:block absolute top-1/2 left-0 w-full h-[2px] -translate-y-1/2 z-0">
            <div className="w-full h-full bg-gray-200 dark:bg-gray-800 relative">
              <motion.div 
                animate={{ x: ["0%", "100%"] }}
                transition={{ repeat: Infinity, duration: 3, ease: "linear" }}
                className="absolute top-0 bottom-0 w-32 bg-gradient-to-r from-transparent via-indigo-500 to-transparent opacity-50"
              />
            </div>
          </div>

          <div className="relative z-10 flex flex-col items-center gap-10 sm:gap-12 md:grid md:grid-cols-[minmax(280px,1fr)_auto_minmax(280px,1fr)] md:items-center md:gap-10 lg:gap-14">
            
            {/* App Node */}
            <div className="flex flex-col items-center md:justify-self-center">
              <div className="w-20 h-20 rounded-2xl bg-white dark:bg-[#111827] border-2 border-gray-200 dark:border-gray-700 shadow-lg flex items-center justify-center relative group">
                <AppWindow className="w-10 h-10 text-slate-600 dark:text-gray-400 group-hover:text-indigo-500 transition-colors" />
                <div className="absolute -bottom-8 whitespace-nowrap font-medium text-slate-900 dark:text-white">
                  اپلیکیشن‌های شما
                </div>
              </div>
            </div>

            {/* GodarAI Gateway */}
            <motion.div 
              animate={{ y: [-5, 5, -5] }}
              transition={{ repeat: Infinity, duration: 4, ease: "easeInOut" }}
              className="flex flex-col items-center md:justify-self-center"
            >
              <div className="relative flex h-28 w-28 items-center justify-center rounded-3xl border-4 border-white bg-gradient-to-br from-indigo-500 to-purple-600 shadow-[0_0_30px_rgba(99,102,241,0.4)] dark:border-[#0B0F19]">
                <div className="relative z-10 flex h-20 w-20 items-center justify-center rounded-2xl border border-white/80 bg-white shadow-lg ring-1 ring-slate-900/10">
                  <Image
                    src="/mini_logo.svg"
                    alt="لوگوی گدارAI"
                    width={64}
                    height={64}
                    className="h-16 w-16"
                  />
                </div>
                <div className="absolute -bottom-10 z-10 whitespace-nowrap text-lg font-bold text-indigo-600 dark:text-indigo-400">
                  درگاه گدارAI
                </div>
                
                {/* Ping rings */}
                <div className="absolute inset-0 rounded-3xl border-2 border-indigo-400 dark:border-indigo-500 animate-ping opacity-20"></div>
              </div>
            </motion.div>

            {/* AI Providers */}
            <div className="flex w-full max-w-sm flex-col gap-4 sm:gap-6 md:w-[420px] md:max-w-none md:items-stretch md:justify-self-center md:gap-5 lg:w-[460px] lg:gap-6">
              <div className="flex items-center justify-between gap-4 bg-white dark:bg-[#111827] p-3 pr-4 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm md:min-h-16">
                <Image
                  src="/openai-light.png"
                  alt="OpenAI"
                  width={32}
                  height={32}
                  className="h-8 w-8 rounded-lg"
                />
                <span className="font-medium text-slate-900 dark:text-white" dir="ltr">Open AI</span>
              </div>
              <div className="flex items-center justify-between gap-4 bg-white dark:bg-[#111827] p-3 pr-4 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm md:min-h-16">
                <Image
                  src="/providers/anthropic.png"
                  alt="Anthropic"
                  width={32}
                  height={32}
                  className="h-8 w-8 rounded-lg"
                />
                <span className="font-medium text-slate-900 dark:text-white" dir="ltr">Anthropic (Claude)</span>
              </div>
              <div className="flex items-center justify-between gap-4 bg-white dark:bg-[#111827] p-3 pr-4 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm md:min-h-16">
                <Image
                  src="/Google-Gemini-Logo-Transparent.png"
                  alt="Google Gemini"
                  width={32}
                  height={32}
                  className="h-8 w-8 rounded-lg"
                />
                <span className="font-medium text-slate-900 dark:text-white" dir="ltr">Google Gemini</span>
              </div>
              <div className="flex items-center justify-between gap-4 bg-white dark:bg-[#111827] p-3 pr-4 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm md:min-h-16">
                <Globe2 className="h-8 w-8 shrink-0 text-indigo-500" aria-hidden="true" />
                <span className="font-medium text-slate-900 dark:text-white">
                  و بیش از ۲۰ ارائه‌دهنده معتبر جهان
                </span>
              </div>
            </div>

          </div>
        </div>
      </div>
    </section>
  );
}

