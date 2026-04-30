'use client';

import { motion } from 'motion/react';
import { ArrowLeft, Terminal, Activity, TrendingDown } from 'lucide-react';
import Link from 'next/link';

export function Hero() {
  return (
    <section className="relative overflow-hidden pt-10 pb-12 sm:pt-16 sm:pb-16 lg:pt-32 lg:pb-40">
      {/* Background Gradients */}
      <div className="absolute top-0 left-1/2 -translate-x-1/2 w-[1000px] h-[500px] opacity-30 dark:opacity-20 pointer-events-none">
        <div className="absolute inset-0 bg-gradient-to-r from-indigo-500 to-purple-500 blur-[100px] rounded-full mix-blend-multiply dark:mix-blend-screen" />
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
        <div className="grid gap-10 lg:grid-cols-2 lg:gap-8 items-center">
          
          {/* Text Content */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
            className="max-w-2xl text-center lg:text-right"
          >
            <div className="inline-flex max-w-full items-center gap-2 px-3 py-1 rounded-full bg-indigo-50 dark:bg-indigo-500/10 text-indigo-600 dark:text-indigo-400 text-[11px] sm:text-sm font-medium mb-4 sm:mb-8 border border-indigo-100 dark:border-indigo-500/20 whitespace-normal leading-6">
              <span className="flex h-2 w-2 rounded-full bg-indigo-500 animate-pulse"></span>
              دسترسی بدون محدودیت به مدل‌های روز دنیا
            </div>
            <h1 className="text-3xl sm:text-4xl lg:text-6xl font-bold text-slate-900 dark:text-white leading-[1.35] sm:leading-[1.2] mb-4 sm:mb-6 font-sans text-balance">شاهراه؛ مسیر ساده و به‌صرفه برای دسترسی به <span className="text-transparent bg-clip-text bg-gradient-to-r from-indigo-500 to-purple-500">مدل‌های هوش مصنوعی</span></h1>
            <p className="mx-auto max-w-xl text-sm sm:text-lg text-slate-600 dark:text-gray-400 mb-6 sm:mb-8 leading-8 sm:leading-relaxed lg:mx-0">
              یک LLM Gateway برای کاربران و کسب‌وکارهای ایرانی — از درگاه پرداخت ایرانی تا مدیریت، بهینه‌سازی و استفاده از مدل‌های AI در مقیاس با کمترین هزینه و بالاترین پایداری.
            </p>
            <div className="flex flex-col sm:flex-row sm:flex-wrap items-stretch sm:items-center gap-3 sm:gap-4">
              <Link href="/signup" className="justify-center flex w-full sm:w-auto items-center gap-2 px-6 py-3.5 min-h-12 rounded-xl bg-indigo-500 hover:bg-indigo-600 text-white font-medium transition-all shadow-[0_0_20px_rgba(99,102,241,0.4)] hover:shadow-[0_0_30px_rgba(99,102,241,0.6)]">
                شروع کنید
                <ArrowLeft className="w-5 h-5" />
              </Link>
              <Link href="/docs" className="justify-center flex w-full sm:w-auto items-center gap-2 px-6 py-3.5 min-h-12 rounded-xl bg-white dark:bg-gray-800 text-slate-700 dark:text-gray-300 font-medium hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors border border-gray-200 dark:border-gray-700 shadow-sm">
                مشاهده مستندات
                <Terminal className="w-5 h-5" />
              </Link>
            </div>
          </motion.div>

          {/* Code / Visuals */}
          <motion.div
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.5, delay: 0.2 }}
            className="relative flex items-center justify-center lg:h-[500px]"
          >
            {/* Dashboard / Code Mockup */}
            <div className="relative w-full max-w-lg rounded-2xl bg-slate-900 shadow-2xl border border-gray-800 overflow-hidden" dir="ltr">
              <div className="flex items-center px-4 py-3 border-b border-gray-800 bg-slate-900/50">
                <div className="flex space-x-2">
                  <div className="w-3 h-3 rounded-full bg-red-500"></div>
                  <div className="w-3 h-3 rounded-full bg-yellow-500"></div>
                  <div className="w-3 h-3 rounded-full bg-green-500"></div>
                </div>
                <div className="mx-auto text-xs text-gray-500 font-mono">request.py</div>
              </div>
              <div className="overflow-x-auto p-3 sm:p-6 text-xs sm:text-sm font-mono text-gray-300">
                <pre>
                  <code className="language-python">
<span className="text-purple-400">from</span> openai <span className="text-purple-400">import</span> OpenAI{'\n\n'}
client = OpenAI({'\n'}
{'  '}api_key=<span className="text-green-400">"sh_live_xxx"</span>,{'\n'}
{'  '}base_url=<span className="text-green-400">"https://api.sharah.ir/v1"</span>{'\n'}
){'\n\n'}
response = client.chat.completions.create({'\n'}
{'  '}model=<span className="text-green-400">"gpt-4o"</span>,{'\n'}
{'  '}messages=[{'{'}{'\n'}
{'    '}<span className="text-blue-400">"role"</span>: <span className="text-green-400">"user"</span>,{'\n'}
{'    '}<span className="text-blue-400">"content"</span>: <span className="text-green-400">"سلام! شاهراه چطور کار میکنه؟"</span>{'\n'}
{'  '}{'}'}]{'\n'}
){'\n\n'}
<span className="text-yellow-400">print</span>(response.choices[<span className="text-orange-400">0</span>].message.content)
                  </code>
                </pre>
              </div>
            </div>

            {/* Floating Badges */}
            <motion.div 
              animate={{ y: [0, -10, 0] }}
              transition={{ repeat: Infinity, duration: 4, ease: "easeInOut" }}
              className="hidden sm:flex absolute -right-4 top-12 bg-white dark:bg-[#111827] p-4 rounded-xl shadow-xl border border-gray-100 dark:border-gray-800 items-center gap-3"
            >
              <div className="p-2 rounded-lg bg-green-100 dark:bg-green-900/30 text-green-600 dark:text-green-400">
                <TrendingDown className="w-5 h-5" />
              </div>
              <div>
                <p className="text-xs text-slate-500 dark:text-gray-400 font-medium">کاهش هزینه</p>
                <p className="text-sm font-bold text-slate-900 dark:text-white">تا ۴۰٪ بهینه‌سازی</p>
              </div>
            </motion.div>

            <motion.div 
              animate={{ y: [0, 10, 0] }}
              transition={{ repeat: Infinity, duration: 5, ease: "easeInOut", delay: 1 }}
              className="hidden sm:flex absolute -left-4 bottom-12 bg-white dark:bg-[#111827] p-4 rounded-xl shadow-xl border border-gray-100 dark:border-gray-800 items-center gap-3"
            >
              <div className="p-2 rounded-lg bg-blue-100 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400">
                <Activity className="w-5 h-5" />
              </div>
              <div>
                <p className="text-xs text-slate-500 dark:text-gray-400 font-medium">وضعیت سرویس</p>
                <p className="text-sm font-bold text-slate-900 dark:text-white">۹۹.۹۹٪ آپتایم</p>
              </div>
            </motion.div>

          </motion.div>
        </div>
      </div>
    </section>
  );
}


