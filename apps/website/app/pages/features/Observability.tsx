'use client';

import { motion } from 'motion/react';
import Link from 'next/link';
import { MagicCard } from '../../components/ui/MagicCard';
import { FAQAccordion } from '../../components/ui/FAQAccordion';
import { 
  ArrowLeft, Activity, Box, BarChart3, Clock, AlertTriangle, 
  Layers, Table, Network, CheckCircle, Play, DollarSign, 
  ListFilter, Users, Building, HelpCircle, Code, Workflow, 
  Zap, Database, TerminalSquare, ShieldAlert, Cpu, Fingerprint, 
  Search, ChevronDown, ChevronUp, Check, Globe, CreditCard, Crosshair
} from 'lucide-react';
import { useState } from 'react';

const fadeIn = {
  initial: { opacity: 0, y: 20 },
  animate: { opacity: 1, y: 0 },
  transition: { duration: 0.5 }
};

export function Observability() {
  return (
    <main className="flex min-h-screen flex-col bg-white dark:bg-[#0B0F19] text-slate-900 dark:text-white" dir="rtl">
      
      {/* 1. Hero Section */}
      <section className="relative pt-32 pb-20 overflow-hidden">
        <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top,#4F46E5_0%,transparent_50%)] opacity-[0.08] dark:opacity-[0.15] pointer-events-none" />
        
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
          <div className="text-center max-w-4xl mx-auto">
            <motion.div
              initial={{ opacity: 0, scale: 0.9 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ duration: 0.5 }}
              className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-indigo-50 dark:bg-indigo-500/10 border border-indigo-100 dark:border-indigo-500/20 text-indigo-600 dark:text-indigo-400 text-sm font-bold mb-8"
              style={{ fontFamily: 'var(--font-heading)' }}
            >
              <Activity className="w-4 h-4" />
              LLM Observability برای تیم‌های ایرانی
            </motion.div>
            
            <motion.h1 
              className="text-4xl md:text-5xl lg:text-6xl font-bold tracking-tight mb-8 leading-tight text-slate-900 dark:text-white"
              style={{ fontFamily: 'var(--font-heading)' }}
              {...fadeIn}
            >
              دید کامل روی <span className="text-transparent bg-clip-text bg-gradient-to-l from-indigo-500 to-purple-500">تمام تعاملات AI</span> در محصول شما
            </motion.h1>
            
            <motion.p 
              className="text-lg md:text-xl text-slate-600 dark:text-slate-400 mb-10 leading-relaxed max-w-3xl mx-auto"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.1 }}
            >با گدارAI، هر درخواست، پاسخ، هزینه، تأخیر، خطا، مصرف توکن و رفتار مدل‌های زبانی را در یک مسیر پایش کنید؛ با کاهش ریسک اتصال مستقیم، پرداخت ارزی و پیچیدگی اتصال به چند ارائه‌دهنده خارجی.</motion.p>
            
            <motion.div 
              className="flex flex-col sm:flex-row items-center justify-center gap-4"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.2 }}
            >
              <Link href="/signup" className="w-full sm:w-auto px-8 py-4 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl font-bold text-lg shadow-[0_0_20px_rgba(79,70,229,0.3)] hover:shadow-[0_0_30px_rgba(79,70,229,0.5)] transition-all active:scale-95 flex items-center justify-center gap-2 group">
                شروع کنید
                <ArrowLeft className="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
              </Link>
              <Link href="/contact" className="w-full sm:w-auto px-8 py-4 bg-white dark:bg-slate-900 text-slate-900 dark:text-white border-2 border-slate-200 dark:border-slate-800 hover:border-indigo-600 dark:hover:border-indigo-500 rounded-xl font-bold text-lg transition-all active:scale-95 flex items-center justify-center">
                درخواست دمو
              </Link>
            </motion.div>
          </div>

          {/* Hero Visual Direction - Mock Dashboard */}
          <motion.div 
            className="mt-20 relative mx-auto w-full max-w-6xl"
            initial={{ opacity: 0, y: 40 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.7, delay: 0.3 }}
          >
            <div className="absolute inset-0 bg-gradient-to-t from-white via-transparent to-transparent dark:from-[#0B0F19] z-20 pointer-events-none h-full w-full" style={{ top: '60%' }} />
            
            <div className="bg-slate-50 dark:bg-[#121827] border border-slate-200 dark:border-slate-800 rounded-2xl shadow-2xl overflow-hidden backdrop-blur-sm">
              {/* Fake window header */}
              <div className="flex items-center gap-2 px-4 py-3 border-b border-slate-200 dark:border-slate-800 bg-white/50 dark:bg-slate-900/50">
                <div className="w-3 h-3 rounded-full bg-rose-500/80" />
                <div className="w-3 h-3 rounded-full bg-amber-500/80" />
                <div className="w-3 h-3 rounded-full bg-emerald-500/80" />
                <div className="mx-auto flex items-center gap-2 px-3 py-1 rounded-md bg-slate-100 dark:bg-slate-800 text-xs text-slate-500 dark:text-slate-400 font-medium font-sans">
                  <Activity className="w-3 h-3" />
                  GodarAI Observability
                </div>
              </div>
              
              <div className="p-6 grid grid-cols-1 md:grid-cols-4 gap-4">
                {/* Stats Row */}
                <div className="col-span-1 md:col-span-4 grid grid-cols-2 md:grid-cols-4 gap-4">
                  {[
                    { label: 'حجم درخواست (۲۴ ساعت)', value: '۱۲۴,۵۹۲', change: '+۱۲٪', icon: Network, color: 'text-indigo-500' },
                    { label: 'مصرف توکن', value: '۴۵.۲M', change: '+۵٪', icon: Database, color: 'text-emerald-500' },
                    { label: 'میانگین تاخیر (Latency)', value: '240ms', change: '-۱۵ms', icon: Clock, color: 'text-amber-500' },
                    { label: 'نرخ خطا', value: '۰.۱۲٪', change: '-۰.۱٪', icon: AlertTriangle, color: 'text-rose-500' }
                  ].map((stat, i) => (
                    <div key={i} className="p-4 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
                      <div className="flex justify-between items-start mb-4">
                        <span className="text-xs text-slate-500 dark:text-slate-400 font-medium">{stat.label}</span>
                        <stat.icon className={`w-4 h-4 ${stat.color}`} />
                      </div>
                      <div className="text-2xl font-bold font-sans text-slate-900 dark:text-white mb-2">{stat.value}</div>
                      <div className="text-xs font-medium text-emerald-600 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-500/10 w-fit px-2 py-0.5 rounded">
                        {stat.change}
                      </div>
                    </div>
                  ))}
                </div>

                {/* Main Charts Area */}
                <div className="col-span-1 md:col-span-3 space-y-4">
                  <div className="p-5 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 h-64 relative flex flex-col">
                    <div className="flex justify-between items-center mb-6">
                      <h3 className="text-sm font-bold text-slate-900 dark:text-white">رهگیری هزینه و درخواست‌ها</h3>
                      <div className="flex gap-2">
                        <span className="text-[10px] font-medium px-2 py-1 rounded bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300">هزینه (تومان)</span>
                        <span className="text-[10px] font-medium px-2 py-1 rounded bg-indigo-50 dark:bg-indigo-500/10 text-indigo-600 dark:text-indigo-400">تعداد درخواست</span>
                      </div>
                    </div>
                    <div className="flex-1 flex items-end gap-2 px-2 pb-2">
                      {[40, 60, 45, 80, 50, 70, 90, 65, 85, 100, 75, 85].map((h, i) => (
                        <div key={i} className="flex-1 flex flex-col justify-end gap-1 h-full group">
                          <div className="w-full bg-indigo-500/20 group-hover:bg-indigo-500/40 rounded-t-sm transition-all" style={{ height: `${h}%` }}></div>
                          <div className="w-full bg-slate-200 dark:bg-slate-800 group-hover:bg-slate-300 dark:group-hover:bg-slate-700 rounded-sm transition-all" style={{ height: `${h * 0.6}%` }}></div>
                        </div>
                      ))}
                    </div>
                  </div>
                  
                  {/* Logs Table */}
                  <div className="p-5 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 overflow-hidden">
                    <h3 className="text-sm font-bold text-slate-900 dark:text-white mb-4">جدول گزارش رخدادها (Live)</h3>
                    <div className="w-full overflow-x-auto">
                      <table className="w-full text-xs text-right">
                        <thead>
                          <tr className="text-slate-500 dark:text-slate-400 border-b border-slate-100 dark:border-slate-800">
                            <th className="pb-3 font-medium">وضعیت</th>
                            <th className="pb-3 font-medium">زمان</th>
                            <th className="pb-3 font-medium">مدل</th>
                            <th className="pb-3 font-medium">تأمین‌کننده</th>
                            <th className="pb-3 font-medium">تأخیر</th>
                            <th className="pb-3 font-medium">توکن‌ها</th>
                            <th className="pb-3 font-medium">هزینه</th>
                          </tr>
                        </thead>
                        <tbody className="font-sans">
                          {[
                            { status: '200', time: '10:42:01', model: 'gpt-4-turbo', provider: 'OpenAI', latency: '420ms', tokens: '450', cost: '۵۲۰ ت' },
                            { status: '200', time: '10:41:55', model: 'claude-3-opus', provider: 'Anthropic', latency: '850ms', tokens: '1200', cost: '۱,۴۵۰ ت' },
                            { status: '429', time: '10:41:30', model: 'gpt-3.5-turbo', provider: 'OpenAI', latency: '120ms', tokens: '-', cost: '۰ ت' },
                          ].map((log, i) => (
                            <tr key={i} className="border-b border-slate-50 dark:border-slate-800/50">
                              <td className="py-2.5">
                                <span className={`px-2 py-0.5 rounded-full text-[10px] font-bold ${log.status === '200' ? 'bg-emerald-100 dark:bg-emerald-500/20 text-emerald-600 dark:text-emerald-400' : 'bg-rose-100 dark:bg-rose-500/20 text-rose-600 dark:text-rose-400'}`}>
                                  {log.status}
                                </span>
                              </td>
                              <td className="py-2.5 text-slate-500">{log.time}</td>
                              <td className="py-2.5 text-slate-900 dark:text-slate-300 font-medium">{log.model}</td>
                              <td className="py-2.5 text-slate-500">{log.provider}</td>
                              <td className="py-2.5 text-slate-900 dark:text-slate-300">{log.latency}</td>
                              <td className="py-2.5 text-slate-500">{log.tokens}</td>
                              <td className="py-2.5 text-slate-900 dark:text-slate-300">{log.cost}</td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </div>
                </div>

                {/* Sidebar Trace & Models */}
                <div className="col-span-1 space-y-4">
                  {/* Models breakdown */}
                  <div className="p-5 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
                    <h3 className="text-sm font-bold text-slate-900 dark:text-white mb-4">تفکیک مدل‌ها</h3>
                    <div className="space-y-4 font-sans text-xs">
                      {[
                        { name: 'gpt-4o', percent: 45, color: 'bg-indigo-500' },
                        { name: 'claude-3.5-sonnet', percent: 30, color: 'bg-purple-500' },
                        { name: 'llama-3', percent: 15, color: 'bg-emerald-500' },
                        { name: 'other', percent: 10, color: 'bg-slate-400' },
                      ].map((item, i) => (
                        <div key={i} className="space-y-1.5">
                          <div className="flex justify-between text-slate-600 dark:text-slate-400">
                            <span>{item.name}</span>
                            <span className="font-bold">{item.percent}٪</span>
                          </div>
                          <div className="w-full bg-slate-100 dark:bg-slate-800 rounded-full h-1.5 overflow-hidden">
                            <div className={`${item.color} h-full rounded-full`} style={{ width: `${item.percent}%` }} />
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>

                  {/* Trace Timeline Preview */}
                  <div className="p-5 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
                    <h3 className="text-sm font-bold text-slate-900 dark:text-white mb-4">Trace Timeline</h3>
                    <div className="relative pl-2 space-y-4 before:absolute before:right-3.5 before:top-2 before:bottom-2 before:w-[2px] before:bg-slate-100 dark:before:bg-slate-800">
                      {[
                        { label: 'User Prompt', time: '0ms', icon: Users },
                        { label: 'Model Call', time: '12ms', icon: Cpu },
                        { label: 'Tool Call (Search)', time: '450ms', icon: Search },
                        { label: 'Final Response', time: '1.2s', icon: CheckCircle, color: 'text-emerald-500' },
                      ].map((step, i) => (
                        <div key={i} className="relative flex items-start gap-3">
                          <div className={`w-5 h-5 rounded-full bg-white dark:bg-slate-900 border-2 ${step.color ? 'border-emerald-500' : 'border-indigo-500'} flex items-center justify-center z-10 shrink-0`}>
                            <step.icon className={`w-2.5 h-2.5 ${step.color ? step.color : 'text-indigo-500'}`} />
                          </div>
                          <div>
                            <div className="text-xs font-bold text-slate-900 dark:text-white font-sans">{step.label}</div>
                            <div className="text-[10px] text-slate-500 font-sans">{step.time}</div>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>

              </div>
            </div>
          </motion.div>
        </div>
      </section>

      {/* 2. Social Proof / Trust Bar */}
      <section className="py-12 border-y border-slate-100 dark:border-slate-800/50 bg-slate-50/50 dark:bg-slate-900/20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-8">
            <h2 className="text-xl font-bold text-slate-900 dark:text-white mb-2" style={{ fontFamily: 'var(--font-heading)' }}>
              ساخته‌شده برای تیم‌هایی که با LLM محصول واقعی می‌سازند
            </h2>
            <p className="text-sm text-slate-600 dark:text-slate-400 max-w-2xl mx-auto">گدارAI به تیم‌های محصول، فنی و AI کمک می‌کند مصرف، کیفیت و عملکرد مدل‌های زبانی را در یک محیط واحد مدیریت کنند؛ با تجربه‌ای مناسب بازار ایران و بدون نیاز به پرداخت ارزی مستقیم.</p>
          </div>
          
          <div className="flex flex-wrap justify-center gap-3">
            {[
              'استارتاپ‌های AI', 'تیم‌های نرم‌افزاری', 'سازمان‌های داده‌محور', 
              'محصولات SaaS', 'تیم‌های Customer Support', 'تیم‌های R&D و AI Lab'
            ].map((chip, idx) => (
              <div key={idx} className="px-4 py-2 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-full text-sm font-medium text-slate-700 dark:text-slate-300 shadow-sm">
                {chip}
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* 3. Section — 360° Visibility */}
      <section className="py-24 relative">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16 max-w-3xl mx-auto">
            <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white mb-4" style={{ fontFamily: 'var(--font-heading)' }}>
              دید ۳۶۰ درجه روی هر تعامل AI
            </h2>
            <p className="text-lg text-slate-600 dark:text-slate-400">
              با ماژول LLM Observability گدارAI، هزینه، عملکرد، کیفیت پاسخ‌ها و رفتار مدل‌ها را در سطح هر درخواست بررسی کنید.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
            {/* Feature Card 1 */}
            <MagicCard className="bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-3xl p-8 flex flex-col" gradientColor="rgba(99,102,241,0.1)">
              <div className="w-12 h-12 bg-indigo-100 dark:bg-indigo-500/20 text-indigo-600 dark:text-indigo-400 rounded-xl flex items-center justify-center mb-6">
                <Table className="w-6 h-6" />
              </div>
              <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-3" style={{ fontFamily: 'var(--font-heading)' }}>
                ثبت لاگ کامل برای هر درخواست و پاسخ
              </h3>
              <p className="text-slate-600 dark:text-slate-400 mb-6 leading-relaxed flex-1">
                گدارAI تمام جزئیات مهم هر درخواست و پاسخ را ثبت می‌کند؛ از مدل انتخاب‌شده و تعداد توکن‌ها تا تأخیر، هزینه، وضعیت پاسخ، خطاها و Metadata. این یعنی تیم شما همیشه می‌داند در پشت هر تعامل AI چه اتفاقی افتاده است.
              </p>
              <div className="bg-indigo-50/50 dark:bg-indigo-500/5 border border-indigo-100 dark:border-indigo-500/10 rounded-xl p-4 text-xs text-indigo-800 dark:text-indigo-300 font-medium">
                💡 مناسب برای Debug، گزارش‌گیری، Audit و تحلیل رفتار کاربران.
              </div>
            </MagicCard>

            {/* Feature Card 2 */}
            <MagicCard className="bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-3xl p-8 flex flex-col" gradientColor="rgba(99,102,241,0.1)">
              <div className="w-12 h-12 bg-purple-100 dark:bg-purple-500/20 text-purple-600 dark:text-purple-400 rounded-xl flex items-center justify-center mb-6">
                <Workflow className="w-6 h-6" />
              </div>
              <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-3" style={{ fontFamily: 'var(--font-heading)' }}>
                Debug سریع‌تر با Tracing
              </h3>
              <p className="text-slate-600 dark:text-slate-400 mb-6 leading-relaxed flex-1">
                مسیر کامل اجرای یک درخواست LLM را در یک نمای زمانی و قابل‌فهم ببینید. از Prompt اولیه تا Tool Callها، Retryها، Fallbackها و پاسخ نهایی، همه‌چیز در یک Trace واحد قابل بررسی است.
              </p>
              <div className="bg-purple-50/50 dark:bg-purple-500/5 border border-purple-100 dark:border-purple-500/10 rounded-xl p-4 text-xs text-purple-800 dark:text-purple-300 font-medium">
                💡 برای Agentها، Workflowهای چندمرحله‌ای و Chainهای پیچیده ایده‌آل است.
              </div>
            </MagicCard>

            {/* Feature Card 3 */}
            <MagicCard className="bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-3xl p-8 flex flex-col" gradientColor="rgba(99,102,241,0.1)">
              <div className="w-12 h-12 bg-emerald-100 dark:bg-emerald-500/20 text-emerald-600 dark:text-emerald-400 rounded-xl flex items-center justify-center mb-6">
                <CheckCircle className="w-6 h-6" />
              </div>
              <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-3" style={{ fontFamily: 'var(--font-heading)' }}>
                ارزیابی و بهبود کیفیت پاسخ‌ها
              </h3>
              <p className="text-slate-600 dark:text-slate-400 mb-6 leading-relaxed flex-1">
                بازخورد کاربران یا تیم داخلی را در سطح درخواست یا مکالمه جمع‌آوری کنید و بفهمید کدام مدل، Prompt یا مسیر پاسخ‌دهی کیفیت بهتری دارد.
              </p>
              <div className="bg-emerald-50/50 dark:bg-emerald-500/5 border border-emerald-100 dark:border-emerald-500/10 rounded-xl p-4 text-xs text-emerald-800 dark:text-emerald-300 font-medium">
                💡 کیفیت خروجی LLM را فقط حدس نزنید؛ اندازه‌گیری کنید.
              </div>
            </MagicCard>

            {/* Feature Card 4 */}
            <MagicCard className="bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-3xl p-8 flex flex-col" gradientColor="rgba(99,102,241,0.1)">
              <div className="w-12 h-12 bg-amber-100 dark:bg-amber-500/20 text-amber-600 dark:text-amber-400 rounded-xl flex items-center justify-center mb-6">
                <DollarSign className="w-6 h-6" />
              </div>
              <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-3" style={{ fontFamily: 'var(--font-heading)' }}>
                کنترل هزینه‌ها با نگاه FinOps
              </h3>
              <p className="text-slate-600 dark:text-slate-400 mb-6 leading-relaxed flex-1">
                مصرف AI را در لحظه تحلیل کنید، هزینه‌ها را بر اساس مدل، تأمین‌کننده، کاربر، Workspace یا فیچر محصول ببینید و قبل از اینکه هزینه‌ها از کنترل خارج شوند، نقاط پرمصرف را شناسایی کنید.
              </p>
              <div className="bg-amber-50/50 dark:bg-amber-500/5 border border-amber-100 dark:border-amber-500/10 rounded-xl p-4 text-xs text-amber-800 dark:text-amber-300 font-medium">
                💡 مخصوص تیم‌هایی که می‌خواهند AI را مقیاس‌پذیر و اقتصادی در محصول خود استفاده کنند.
              </div>
            </MagicCard>
          </div>
        </div>
      </section>

      {/* 4. Section — Instant Visibility */}
      <section className="py-24 bg-slate-900 dark:bg-[#060913] text-white border-y border-slate-800">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 items-center">
            <div>
              <h2 className="text-3xl md:text-4xl font-bold mb-4" style={{ fontFamily: 'var(--font-heading)' }}>
                Observability فوری، بدون سربار فنی
              </h2>
              <p className="text-lg text-slate-400 mb-10">
                گدارAI به شما کمک می‌کند عملکرد AI app خود را در سطح Frameworkها، Promptها، Tool Callها، Agentها و تأمین‌کننده‌ها مانیتور کنید؛ بدون اینکه تیم فنی درگیر پیاده‌سازی پیچیده شود.
              </p>
              
              <div className="space-y-8">
                {[
                  {
                    title: 'اندازه‌گیری Performance-to-Cost ROI در لحظه',
                    desc: 'با داشبورد Real-time گدارAI، بیش از ۴۰ شاخص مهم را برای سنجش عملکرد، هزینه، نرخ خطا، مصرف توکن و کیفیت پاسخ‌ها مانیتور کنید.'
                  },
                  {
                    title: 'مشاهده کامل چرخه اجرای Agent',
                    desc: 'اگر از AI Agentها استفاده می‌کنید، گدارAI مسیر کامل اجرای هر Agent Run را نمایش می‌دهد؛ از تصمیم‌گیری مدل تا Tool Callها، Retryها، Fallbackها و خروجی نهایی.'
                  },
                  {
                    title: 'شناسایی سریع گلوگاه‌ها با Filterهای دقیق',
                    desc: 'با فیلترهای پیشرفته، داده‌ها را بر اساس مدل، تأمین‌کننده، Status Code، کاربر، تیم، Workspace، زمان، هزینه، تأخیر یا Metadata فیلتر کنید و سریع‌تر به علت مشکل برسید.'
                  }
                ].map((item, idx) => (
                  <div key={idx} className="flex gap-4">
                    <div className="w-8 h-8 rounded-full bg-indigo-500/20 text-indigo-400 flex items-center justify-center shrink-0 mt-1">
                      <Check className="w-4 h-4" />
                    </div>
                    <div>
                      <h4 className="text-lg font-bold mb-2">{item.title}</h4>
                      <p className="text-slate-400 text-sm leading-relaxed">{item.desc}</p>
                    </div>
                  </div>
                ))}
              </div>
            </div>
            
            <div className="relative">
              <div className="absolute -inset-4 bg-gradient-to-br from-indigo-500/20 to-purple-500/20 rounded-3xl blur-2xl z-0" />
              <div className="relative z-10 bg-[#0B0F19] border border-slate-800 rounded-2xl p-6 shadow-2xl">
                <div className="flex items-center justify-between mb-6 pb-4 border-b border-slate-800">
                  <h4 className="font-bold text-sm">Real-time Metrics</h4>
                  <div className="flex gap-2">
                    <span className="flex items-center gap-1 text-xs bg-slate-800 px-2 py-1 rounded text-slate-300">
                      <Clock className="w-3 h-3" /> Last 1h
                    </span>
                  </div>
                </div>
                
                <div className="grid grid-cols-2 gap-4 mb-6">
                  <div className="p-4 bg-slate-800/50 rounded-xl border border-slate-700/50">
                    <div className="text-slate-400 text-xs mb-1">Average Latency</div>
                    <div className="text-2xl font-bold font-sans text-amber-400">0.8s</div>
                  </div>
                  <div className="p-4 bg-slate-800/50 rounded-xl border border-slate-700/50">
                    <div className="text-slate-400 text-xs mb-1">Token Usage</div>
                    <div className="text-2xl font-bold font-sans text-emerald-400">12M</div>
                  </div>
                  <div className="p-4 bg-slate-800/50 rounded-xl border border-slate-700/50">
                    <div className="text-slate-400 text-xs mb-1">Total Cost</div>
                    <div className="text-2xl font-bold font-sans text-indigo-400">۴۵۲,۰۰۰ ت</div>
                  </div>
                  <div className="p-4 bg-slate-800/50 rounded-xl border border-slate-700/50">
                    <div className="text-slate-400 text-xs mb-1">Cache Hit Rate</div>
                    <div className="text-2xl font-bold font-sans text-purple-400">32.4%</div>
                  </div>
                </div>
                
                <div className="p-4 bg-slate-800/50 rounded-xl border border-slate-700/50">
                  <div className="flex items-center justify-between mb-4">
                    <div className="text-slate-400 text-xs">Cost by Provider</div>
                  </div>
                  <div className="space-y-3 font-sans text-xs">
                    <div className="flex items-center gap-3">
                      <div className="w-16 text-slate-300">OpenAI</div>
                      <div className="flex-1 h-2 bg-slate-800 rounded-full overflow-hidden">
                        <div className="w-[60%] h-full bg-indigo-500 rounded-full" />
                      </div>
                      <div className="w-12 text-left text-slate-400">60%</div>
                    </div>
                    <div className="flex items-center gap-3">
                      <div className="w-16 text-slate-300">Anthropic</div>
                      <div className="flex-1 h-2 bg-slate-800 rounded-full overflow-hidden">
                        <div className="w-[25%] h-full bg-purple-500 rounded-full" />
                      </div>
                      <div className="w-12 text-left text-slate-400">25%</div>
                    </div>
                    <div className="flex items-center gap-3">
                      <div className="w-16 text-slate-300">Meta</div>
                      <div className="flex-1 h-2 bg-slate-800 rounded-full overflow-hidden">
                        <div className="w-[15%] h-full bg-emerald-500 rounded-full" />
                      </div>
                      <div className="w-12 text-left text-slate-400">15%</div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* 5. Section — Everything You Need */}
      <section className="py-24 bg-slate-50 dark:bg-[#0B0F19]">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16 max-w-3xl mx-auto">
            <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white mb-4" style={{ fontFamily: 'var(--font-heading)' }}>
              هر چیزی که برای سنجش عملکرد نیاز دارید
            </h2>
            <p className="text-lg text-slate-600 dark:text-slate-400">
              یک ماژول Observability سازگار با نیاز تیم‌های LLM که دید end-to-end روی عملکرد، هزینه و کیفیت تعاملات AI فراهم می‌کند.
            </p>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-6">
            {[
              {
                icon: ListFilter,
                title: 'گزارش رخدادها و Traces کامل',
                desc: 'هر درخواست را ثبت کنید و مسیر کامل آن را از ابتدا تا انتها دنبال کنید. داده‌ها را Export کنید.'
              },
              {
                icon: Fingerprint,
                title: 'Metadata قابل‌تحلیل',
                desc: 'هزینه، عملکرد، Accuracy، کاربر، Workspace، تأمین‌کننده و اطلاعات سفارشی خودتان را روی هر درخواست ذخیره کنید.'
              },
              {
                icon: Code,
                title: 'Auto-instrumentation',
                desc: 'Tracing، ثبت لاگ و شاخص‌ها را برای Frameworkها و Agentهای مختلف با کمترین تغییر در کد فعال کنید.'
              },
              {
                icon: BarChart3,
                title: 'تحلیل لحظه‌ای',
                desc: 'Usage، خطاها، Caching، بازخورد، Metadata و هزینه‌ها را در یک داشبورد Real-time مشاهده کنید.'
              },
              {
                icon: Search,
                title: 'فیلترهای پیشرفته',
                desc: 'با بیش از ۱۵ فیلتر کاربردی، Viewهای دقیق و قابل‌اقدام برای تیم فنی، محصول و مدیریت بسازید.'
              },
              {
                icon: HelpCircle,
                title: 'بازخورد ساختاریافته',
                desc: 'بازخورد وزنی و قابل‌تحلیل جمع‌آوری کنید تا کیفیت پاسخ‌ها، Promptها و مدل‌ها را به‌مرور بهبود دهید.'
              }
            ].map((feature, idx) => (
              <div key={idx} className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 hover:shadow-lg transition-all group hover:border-indigo-300 dark:hover:border-indigo-500/50">
                <div className="w-10 h-10 rounded-lg bg-indigo-50 dark:bg-indigo-500/10 text-indigo-600 dark:text-indigo-400 flex items-center justify-center mb-4 group-hover:scale-110 transition-transform">
                  <feature.icon className="w-5 h-5" />
                </div>
                <h4 className="text-lg font-bold text-slate-900 dark:text-white mb-2">{feature.title}</h4>
                <p className="text-sm text-slate-600 dark:text-slate-400 leading-relaxed">{feature.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* 6. Shahrah Differentiation Section */}
      <section className="py-24 border-t border-slate-200 dark:border-slate-800 bg-white dark:bg-[#0B0F19]">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16 max-w-3xl mx-auto">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-rose-50 dark:bg-rose-500/10 text-rose-600 dark:text-rose-400 text-xs font-bold mb-4">
              چرا گدارAI؟
            </div>
            <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white mb-4" style={{ fontFamily: 'var(--font-heading)' }}>
              چرا Observability در گدارAI برای تیم‌های ایرانی مهم‌تر است؟
            </h2>
            <p className="text-lg text-slate-600 dark:text-slate-400">
              وقتی محصول شما به LLM وابسته است، فقط اتصال به مدل کافی نیست. باید بدانید هر درخواست چقدر هزینه دارد، چقدر طول می‌کشد، چه خطایی می‌دهد و کدام تأمین‌کننده بهترین عملکرد را دارد.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {[
              {
                icon: ShieldAlert,
                title: 'کاهش ریسک محدودیت‌های اتصال مستقیم',
                desc: 'با گدارAI، تیم شما برای دسترسی به مدل‌های LLM خارجی درگیر محدودیت‌های تحریم، مسدود شدن حساب یا ریسک قطع دسترسی مستقیم نمی‌شود.'
              },
              {
                icon: CreditCard,
                title: 'بدون پرداخت ارزی مستقیم',
                desc: 'هزینه‌های AI را بدون نیاز به کارت اعتباری خارجی یا پرداخت دلاری مدیریت کنید. گدارAI تجربه‌ای مناسب کسب‌وکارهای ایرانی فراهم می‌کند.'
              },
              {
                icon: Crosshair,
                title: 'Observability متناسب با نیاز بازار ایران',
                desc: 'داشبوردها، گزارش‌ها و جریان‌های پایش گدارAI برای تیم‌هایی طراحی شده‌اند که می‌خواهند AI را در محصول فارسی، بازار ایران و مقیاس واقعی استفاده کنند.'
              },
              {
                icon: Globe,
                title: 'یک نقطه کنترل برای چند مدل و تأمین‌کننده',
                desc: 'به‌جای اتصال جداگانه به چند تأمین‌کننده، همه تعاملات LLM را از یک مسیر واحد مدیریت، مانیتور و تحلیل کنید.'
              }
            ].map((diff, idx) => (
              <div key={idx} className="flex gap-4 p-6 bg-slate-50 dark:bg-slate-900/50 rounded-2xl border border-slate-200 dark:border-slate-800">
                <div className="w-12 h-12 rounded-xl bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 flex items-center justify-center shrink-0 shadow-sm">
                  <diff.icon className="w-6 h-6 text-indigo-600 dark:text-indigo-400" />
                </div>
                <div>
                  <h4 className="text-lg font-bold text-slate-900 dark:text-white mb-2">{diff.title}</h4>
                  <p className="text-sm text-slate-600 dark:text-slate-400 leading-relaxed">{diff.desc}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* 7. Use Cases Section */}
      <section className="py-24 bg-slate-50 dark:bg-slate-900/20 border-t border-slate-200 dark:border-slate-800">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white mb-4" style={{ fontFamily: 'var(--font-heading)' }}>
              برای چه تیم‌هایی ساخته شده است؟
            </h2>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
            {[
              {
                icon: TerminalSquare,
                title: 'تیم‌های محصول AI',
                desc: 'بفهمید کاربران چطور با قابلیت‌های AI تعامل می‌کنند، کدام پاسخ‌ها کیفیت پایین‌تری دارند و کجا باید تجربه محصول را بهبود دهید.'
              },
              {
                icon: Code,
                title: 'تیم‌های Backend و Platform',
                desc: 'تأخیر، خطا، Retry، Timeout و Fallback را در سطح زیرساخت بررسی کنید و پایداری AI app را بالا ببرید.'
              },
              {
                icon: Building,
                title: 'تیم‌های مالی و مدیریت',
                desc: 'مصرف AI را بر اساس تیم، فیچر، کاربر یا Workspace ببینید و هزینه‌ها را قابل پیش‌بینی‌تر کنید.'
              },
              {
                icon: Users,
                title: 'تیم‌های Customer Support',
                desc: 'کیفیت پاسخ‌های AI را در مکالمات واقعی بررسی کنید و موارد خطادار یا کم‌کیفیت را سریع‌تر پیدا کنید.'
              }
            ].map((useCase, idx) => (
              <MagicCard key={idx} className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 text-center flex flex-col items-center" gradientColor="rgba(99,102,241,0.08)">
                <div className="w-14 h-14 rounded-full bg-slate-100 dark:bg-slate-800 flex items-center justify-center mb-4 text-slate-700 dark:text-slate-300">
                  <useCase.icon className="w-6 h-6" />
                </div>
                <h4 className="text-lg font-bold text-slate-900 dark:text-white mb-3">{useCase.title}</h4>
                <p className="text-sm text-slate-600 dark:text-slate-400 leading-relaxed">
                  {useCase.desc}
                </p>
              </MagicCard>
            ))}
          </div>
        </div>
      </section>

      {/* 9. Resources Section */}
      <section className="py-24 bg-white dark:bg-[#0B0F19] border-t border-slate-200 dark:border-slate-800">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-end justify-between mb-12">
            <div>
              <h2 className="text-3xl font-bold text-slate-900 dark:text-white mb-2" style={{ fontFamily: 'var(--font-heading)' }}>
                راهنماها و منابع پیشنهادی
              </h2>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {[
              {
                title: 'چطور هزینه LLM را در سطح کاربر مانیتور کنیم؟',
                desc: 'مصرف توکن و هزینه هر کاربر را در مدل‌ها، محصولات و Workspaceهای مختلف تحلیل کنید.',
                tag: 'آموزش'
              },
              {
                title: 'بهینه‌سازی هزینه‌های GenAI با نگاه FinOps',
                desc: 'با اصول FinOps، هزینه‌های AI را قابل‌کنترل، قابل‌پیش‌بینی و قابل‌گزارش کنید.',
                tag: 'FinOps'
              },
              {
                title: 'چرا LLM Observability برای AI appها ضروری است؟',
                desc: 'بدون Observability، Debug، کنترل هزینه و بهبود کیفیت پاسخ‌ها در مقیاس واقعی دشوار می‌شود.',
                tag: 'مفاهیم پایه'
              }
            ].map((resource, idx) => (
              <a key={idx} href="#" className="group bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl overflow-hidden hover:border-indigo-400 dark:hover:border-indigo-500 transition-colors flex flex-col h-full">
                <div className="h-40 bg-slate-100 dark:bg-slate-800 relative overflow-hidden">
                  <div className="absolute inset-0 bg-gradient-to-br from-indigo-500/10 to-purple-500/10 group-hover:scale-105 transition-transform duration-500" />
                  <div className="absolute top-4 right-4 bg-white/90 dark:bg-slate-900/90 backdrop-blur text-xs font-bold px-2 py-1 rounded text-slate-700 dark:text-slate-300">
                    {resource.tag}
                  </div>
                </div>
                <div className="p-6 flex-1 flex flex-col">
                  <h3 className="text-lg font-bold text-slate-900 dark:text-white mb-3 group-hover:text-indigo-600 dark:group-hover:text-indigo-400 transition-colors">
                    {resource.title}
                  </h3>
                  <p className="text-sm text-slate-600 dark:text-slate-400 mb-4 flex-1">
                    {resource.desc}
                  </p>
                  <div className="text-indigo-600 dark:text-indigo-400 text-sm font-bold flex items-center gap-1">
                    مطالعه مقاله <ArrowLeft className="w-4 h-4" />
                  </div>
                </div>
              </a>
            ))}
          </div>
        </div>
      </section>

      {/* 10. FAQ Section */}
      <section className="py-24 bg-slate-50 dark:bg-slate-900/20 border-t border-slate-200 dark:border-slate-800">
        <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 className="text-3xl font-bold text-slate-900 dark:text-white" style={{ fontFamily: 'var(--font-heading)' }}>
              سوالات متداول
            </h2>
          </div>

          <FAQAccordion items={[
              {
                q: "آیا می‌توانم هر مرحله یا Tool Call داخل یک Agent Flow را ببینم؟",
                a: "بله. گدارAI مسیر اجرای Agent را به‌صورت Trace نمایش می‌دهد تا بتوانید Promptها، Model Callها، Tool Callها، Retryها و پاسخ نهایی را بررسی کنید."
              },
              {
                q: "آیا تأخیر و مصرف توکن را برای مدل‌ها و تیم‌های مختلف نشان می‌دهد؟",
                a: "بله. می‌توانید شاخص‌هایی مثل تأخیر، مصرف توکن، هزینه، نرخ خطا و تعداد درخواست‌ها را بر اساس مدل، تأمین‌کننده، Workspace، کاربر یا تیم مشاهده کنید."
              },
              {
                q: "آیا می‌توانم بفهمم کدام تأمین‌کننده یا مدل بیشترین هزینه را ایجاد می‌کند؟",
                a: "بله. گدارAI مصرف و هزینه را در سطح مدل، تأمین‌کننده، کاربر، Workspace و فیچر محصول نمایش می‌دهد تا بتوانید تصمیم‌های بهینه‌تری بگیرید."
              },
              {
                q: "آیا Retry، Fallback و Timeoutها قابل مانیتور هستند؟",
                a: "بله. می‌توانید رفتار درخواست‌ها را در شرایط خطا، Timeout، Retry یا Fallback بررسی کنید و مسیرهای ناپایدار را سریع‌تر شناسایی کنید."
              },
              {
                q: "آیا برای استفاده از گدارAI نیاز به پرداخت ارزی دارم؟",
                a: "خیر. گدارAI برای تیم‌های ایرانی طراحی شده و هدف آن ساده‌کردن دسترسی، مدیریت و پایش LLMها بدون نیاز به پرداخت ارزی مستقیم است."
              },
              {
                q: "آیا گدارAI فقط برای تیم‌های فنی است؟",
                a: "خیر. تیم‌های فنی، محصول، مالی، پشتیبانی و مدیریت می‌توانند از داده‌های Observability گدارAI برای تصمیم‌گیری بهتر استفاده کنند."
              }
            ].map((faq) => ({
              question: faq.q,
              answer: faq.a
            }))} />
        </div>
      </section>

      {/* 11. Final CTA Section */}
      <section className="py-24 relative overflow-hidden">
        <div className="absolute inset-0 bg-indigo-600 dark:bg-indigo-900/40" />
        
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10 text-center">
          <h2 className="text-3xl md:text-5xl font-bold text-white mb-6" style={{ fontFamily: 'var(--font-heading)' }}>از امروز AI Stack خود را مانیتور کنید</h2>
          <p className="text-lg md:text-xl text-indigo-100 mb-10 leading-relaxed max-w-3xl mx-auto">
            با گدارAI، دید کامل روی مصرف، هزینه، عملکرد و کیفیت تعاملات LLM داشته باشید؛ با کاهش ریسک محدودیت‌های اتصال مستقیم، بدون پرداخت ارزی مستقیم و با تجربه‌ای متناسب با تیم‌های ایرانی.
          </p>
          
          <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
            <Link href="/signup" className="w-full sm:w-auto px-8 py-4 bg-white text-indigo-600 hover:bg-slate-50 rounded-xl font-bold text-lg shadow-xl transition-all active:scale-95 flex items-center justify-center gap-2 group">
              شروع کنید
              <ArrowLeft className="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
            </Link>
            <Link href="/contact" className="w-full sm:w-auto px-8 py-4 bg-transparent border-2 border-white/30 hover:border-white text-white rounded-xl font-bold text-lg transition-all active:scale-95 flex items-center justify-center">
              درخواست دمو
            </Link>
          </div>
        </div>
      </section>
    </main>
  );
}

