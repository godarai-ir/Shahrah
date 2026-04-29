'use client';

import { motion } from 'motion/react';
import { User, Server, Cpu, Database, Cloud } from 'lucide-react';

export function Solution() {
  return (
    <section className="py-24 relative overflow-hidden">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center max-w-3xl mx-auto mb-16">
          <h2 className="text-3xl sm:text-4xl font-bold text-slate-900 dark:text-white mb-4">
            یک درگاه واحد برای همه نیازهای شما
          </h2>
          <p className="text-lg text-slate-600 dark:text-gray-400">
            با شاهراه، پیچیدگی‌ها را حذف کنید. تمام درخواست‌های خود را به یک اندپوینت ارسال کنید و ما بقیه کارها را انجام می‌دهیم.
          </p>
        </div>

        <div className="relative max-w-5xl mx-auto py-12">
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

          <div className="flex flex-col md:flex-row items-center justify-between relative z-10 gap-12 md:gap-0">
            
            {/* User Node */}
            <div className="flex flex-col items-center">
              <div className="w-20 h-20 rounded-2xl bg-white dark:bg-[#111827] border-2 border-gray-200 dark:border-gray-700 shadow-lg flex items-center justify-center relative group">
                <User className="w-10 h-10 text-slate-600 dark:text-gray-400 group-hover:text-indigo-500 transition-colors" />
                <div className="absolute -bottom-8 whitespace-nowrap font-medium text-slate-900 dark:text-white">
                  برنامه شما
                </div>
              </div>
            </div>

            {/* Shahrah Gateway */}
            <motion.div 
              animate={{ y: [-5, 5, -5] }}
              transition={{ repeat: Infinity, duration: 4, ease: "easeInOut" }}
              className="flex flex-col items-center"
            >
              <div className="w-28 h-28 rounded-3xl bg-gradient-to-br from-indigo-500 to-purple-600 shadow-[0_0_30px_rgba(99,102,241,0.4)] flex items-center justify-center relative border-4 border-white dark:border-[#0B0F19]">
                <Server className="w-12 h-12 text-white" />
                <div className="absolute -bottom-10 whitespace-nowrap font-bold text-indigo-600 dark:text-indigo-400 text-lg">
                  شاهراه Gateway
                </div>
                
                {/* Ping rings */}
                <div className="absolute inset-0 rounded-3xl border-2 border-indigo-400 dark:border-indigo-500 animate-ping opacity-20"></div>
              </div>
            </motion.div>

            {/* AI Providers */}
            <div className="flex flex-col gap-6 md:gap-8">
              <div className="flex items-center gap-4 bg-white dark:bg-[#111827] p-3 pr-4 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm">
                <Cpu className="w-6 h-6 text-emerald-500" />
                <span className="font-medium text-slate-900 dark:text-white" dir="ltr">OpenAI Models</span>
              </div>
              <div className="flex items-center gap-4 bg-white dark:bg-[#111827] p-3 pr-4 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm ml-6 md:ml-8">
                <Database className="w-6 h-6 text-blue-500" />
                <span className="font-medium text-slate-900 dark:text-white" dir="ltr">Anthropic (Claude)</span>
              </div>
              <div className="flex items-center gap-4 bg-white dark:bg-[#111827] p-3 pr-4 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm">
                <Cloud className="w-6 h-6 text-orange-500" />
                <span className="font-medium text-slate-900 dark:text-white" dir="ltr">Google Gemini</span>
              </div>
            </div>

          </div>
        </div>
      </div>
    </section>
  );
}

