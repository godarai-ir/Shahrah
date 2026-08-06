'use client';

import { motion } from 'motion/react';
import Link from 'next/link';
import { MagicCard } from '../../components/ui/MagicCard';
import { FAQAccordion } from '../../components/ui/FAQAccordion';
import { 
  ArrowLeft, Activity, BarChart3, Clock, AlertTriangle, 
  Table, Network, CheckCircle, DollarSign, 
  ListFilter, Users, Building, HelpCircle, Code, Workflow, 
  Database, TerminalSquare, ShieldAlert, Cpu, Fingerprint, 
  Search, Check, Globe, CreditCard, Crosshair
} from 'lucide-react';

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
              مشاهده‌پذیری درخواست‌های مدل زبانی
            </motion.div>
            
            <motion.h1 
              className="text-4xl md:text-5xl lg:text-6xl font-bold tracking-tight mb-8 leading-tight text-slate-900 dark:text-white"
              style={{ fontFamily: 'var(--font-heading)' }}
              {...fadeIn}
            >
              از روند کلی مصرف تا <span className="text-transparent bg-clip-text bg-gradient-to-l from-indigo-500 to-purple-500">جزئیات هر درخواست</span>
            </motion.h1>
            
            <motion.p 
              className="text-lg md:text-xl text-slate-600 dark:text-slate-400 mb-10 leading-relaxed max-w-3xl mx-auto"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.1 }}
            >لاگ، رد درخواست و سنجه‌های تجمیعی گدارAI نشان می‌دهند کدام مدل پاسخ داده، چه میزان توکن و هزینه ثبت شده، چه خطایی رخ داده و هر مرحله چقدر زمان برده است.</motion.p>
            
            <motion.div 
              className="flex flex-col sm:flex-row items-center justify-center gap-4"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.2 }}
            >
              <Link href="/pricing" className="w-full sm:w-auto px-8 py-4 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl font-bold text-lg shadow-[0_0_20px_rgba(79,70,229,0.3)] hover:shadow-[0_0_30px_rgba(79,70,229,0.5)] transition-all active:scale-95 flex items-center justify-center gap-2 group">
                شروع کنید
                <ArrowLeft className="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
              </Link>
              <Link href="/demo" className="w-full sm:w-auto px-8 py-4 bg-white dark:bg-slate-900 text-slate-900 dark:text-white border-2 border-slate-200 dark:border-slate-800 hover:border-indigo-600 dark:hover:border-indigo-500 rounded-xl font-bold text-lg transition-all active:scale-95 flex items-center justify-center">
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
                    { label: 'میانگین تأخیر', value: '240ms', change: '-۱۵ms', icon: Clock, color: 'text-amber-500' },
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
                    <h3 className="text-sm font-bold text-slate-900 dark:text-white mb-4">نمونه لاگ درخواست‌ها</h3>
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
              برای تیم‌هایی که مدل زبانی را در محصول واقعی به کار می‌گیرند
            </h2>
            <p className="text-sm text-slate-600 dark:text-slate-400 max-w-2xl mx-auto">تیم‌های فنی، محصول و مالی می‌توانند داده‌های درخواست، مصرف و هزینه را از یک منبع مشترک بررسی کنند و سریع‌تر به علت تغییرها برسند.</p>
          </div>
          
          <div className="flex flex-wrap justify-center gap-3">
            {[
              'استارتاپ‌های هوش مصنوعی', 'تیم‌های نرم‌افزاری', 'سازمان‌های داده‌محور',
              'محصولات SaaS', 'تیم‌های پشتیبانی مشتری', 'تیم‌های تحقیق و توسعه'
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
              سه نمای مکمل برای بررسی درخواست‌ها
            </h2>
            <p className="text-lg text-slate-600 dark:text-slate-400">
              لاگ برای جزئیات یک درخواست، رد درخواست برای مسیر مرحله‌به‌مرحله و سنجه برای تحلیل روندهای کلی به کار می‌رود.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
            {/* Feature Card 1 */}
            <MagicCard className="bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-3xl p-8 flex flex-col" gradientColor="rgba(99,102,241,0.1)">
              <div className="w-12 h-12 bg-indigo-100 dark:bg-indigo-500/20 text-indigo-600 dark:text-indigo-400 rounded-xl flex items-center justify-center mb-6">
                <Table className="w-6 h-6" />
              </div>
              <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-3" style={{ fontFamily: 'var(--font-heading)' }}>
                بررسی لاگ هر درخواست
              </h3>
              <p className="text-slate-600 dark:text-slate-400 mb-6 leading-relaxed flex-1">
                مدل انتخاب‌شده، توکن، تأخیر، هزینه، وضعیت پاسخ، خطا و فراداده هر درخواست را بررسی کنید. ثبت محتوای درخواست و پاسخ نیز تابع تنظیمات حریم خصوصی فضای کاری است.
              </p>
              <div className="bg-indigo-50/50 dark:bg-indigo-500/5 border border-indigo-100 dark:border-indigo-500/10 rounded-xl p-4 text-xs text-indigo-800 dark:text-indigo-300 font-medium">
                مناسب برای عیب‌یابی، گزارش‌گیری و بررسی یک درخواست مشخص.
              </div>
            </MagicCard>

            {/* Feature Card 2 */}
            <MagicCard className="bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-3xl p-8 flex flex-col" gradientColor="rgba(99,102,241,0.1)">
              <div className="w-12 h-12 bg-purple-100 dark:bg-purple-500/20 text-purple-600 dark:text-purple-400 rounded-xl flex items-center justify-center mb-6">
                <Workflow className="w-6 h-6" />
              </div>
              <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-3" style={{ fontFamily: 'var(--font-heading)' }}>
                عیب‌یابی مرحله‌ای با رد درخواست
              </h3>
              <p className="text-slate-600 dark:text-slate-400 mb-6 leading-relaxed flex-1">
                تلاش‌ها، خطاها، مسیر جایگزین و زمان هر مرحله را در یک نمای زمانی ببینید و مشخص کنید کدام بخش باعث کندی یا شکست شده است.
              </p>
              <div className="bg-purple-50/50 dark:bg-purple-500/5 border border-purple-100 dark:border-purple-500/10 rounded-xl p-4 text-xs text-purple-800 dark:text-purple-300 font-medium">
                برای درخواست‌هایی که چند تلاش یا مقصد دارند، رد درخواست از لاگ جزئیات بیشتری می‌دهد.
              </div>
            </MagicCard>

            {/* Feature Card 3 */}
            <MagicCard className="bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-3xl p-8 flex flex-col" gradientColor="rgba(99,102,241,0.1)">
              <div className="w-12 h-12 bg-emerald-100 dark:bg-emerald-500/20 text-emerald-600 dark:text-emerald-400 rounded-xl flex items-center justify-center mb-6">
                <CheckCircle className="w-6 h-6" />
              </div>
              <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-3" style={{ fontFamily: 'var(--font-heading)' }}>
                تحلیل روند مصرف و عملکرد
              </h3>
              <p className="text-slate-600 dark:text-slate-400 mb-6 leading-relaxed flex-1">
                تعداد درخواست، خطا، تأخیر، مصرف توکن، هزینه و اثربخشی حافظه نهان را در بازه زمانی و بر اساس مدل یا ارائه‌دهنده مقایسه کنید.
              </p>
              <div className="bg-emerald-50/50 dark:bg-emerald-500/5 border border-emerald-100 dark:border-emerald-500/10 rounded-xl p-4 text-xs text-emerald-800 dark:text-emerald-300 font-medium">
                برای بررسی کیفیت محتوایی، این داده‌ها را با ارزیابی و بازخورد محصول خود ترکیب کنید.
              </div>
            </MagicCard>

            {/* Feature Card 4 */}
            <MagicCard className="bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-3xl p-8 flex flex-col" gradientColor="rgba(99,102,241,0.1)">
              <div className="w-12 h-12 bg-amber-100 dark:bg-amber-500/20 text-amber-600 dark:text-amber-400 rounded-xl flex items-center justify-center mb-6">
                <DollarSign className="w-6 h-6" />
              </div>
              <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-3" style={{ fontFamily: 'var(--font-heading)' }}>
                پیگیری هزینه و مصرف
              </h3>
              <p className="text-slate-600 dark:text-slate-400 mb-6 leading-relaxed flex-1">
                هزینه را بر اساس مدل، ارائه‌دهنده، تیم یا حساب مجازی بررسی کنید و افزایش مصرف را به درخواست‌های مربوط برسانید.
              </p>
              <div className="bg-amber-50/50 dark:bg-amber-500/5 border border-amber-100 dark:border-amber-500/10 rounded-xl p-4 text-xs text-amber-800 dark:text-amber-300 font-medium">
                این داده‌ها مبنای تنظیم سقف هزینه، محدودیت نرخ و انتخاب مدل مناسب‌تر هستند.
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
                یک مسیر مشترک برای داده‌های پایش
              </h2>
              <p className="text-lg text-slate-400 mb-10">
                داده‌های درخواست در همان مسیری ثبت می‌شوند که اپلیکیشن از طریق درگاه به مدل وصل است. برای همبستگی با سرویس‌های دیگر نیز می‌توانید از OpenTelemetry استفاده کنید.
              </p>
              
              <div className="space-y-8">
                {[
                  {
                    title: 'مقایسه عملکرد و هزینه',
                    desc: 'تأخیر، هزینه، نرخ خطا و مصرف توکن را در بازه زمانی و بر اساس مدل یا ارائه‌دهنده مقایسه کنید.'
                  },
                  {
                    title: 'بررسی تلاش‌ها و مسیر جایگزین',
                    desc: 'در رد درخواست ببینید هر تلاش به کدام مقصد رفته، چه خطایی رخ داده و مسیر جایگزین چگونه انتخاب شده است.'
                  },
                  {
                    title: 'یافتن سریع درخواست‌های مسئله‌دار',
                    desc: 'داده‌ها را بر اساس مدل، ارائه‌دهنده، کد وضعیت، تیم، زمان، هزینه، تأخیر یا فراداده فیلتر کنید و نمونه‌های مرتبط را باز کنید.'
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
              لاگ، رد درخواست، سنجه و خروجی داده، هرکدام برای یک سطح از بررسی طراحی شده‌اند.
            </p>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-6">
            {[
              {
                icon: ListFilter,
                title: 'لاگ و رد درخواست',
                desc: 'جزئیات یک درخواست را در لاگ ببینید و در صورت نیاز، مسیر اجرای آن را مرحله‌به‌مرحله دنبال کنید.'
              },
              {
                icon: Fingerprint,
                title: 'فراداده قابل تحلیل',
                desc: 'هزینه، تأخیر، کاربر، فضای کاری، ارائه‌دهنده و فراداده سفارشی را برای فیلتر و تحلیل درخواست‌ها به کار ببرید.'
              },
              {
                icon: Code,
                title: 'Auto-instrumentation',
                desc: 'داده گدارAI را با OpenTelemetry در کنار ردیابی و سنجه‌های سرویس‌های خود همبسته کنید.'
              },
              {
                icon: BarChart3,
                title: 'تحلیل روندها',
                desc: 'مصرف، خطا، حافظه نهان، فراداده و هزینه را در بازه زمانی دلخواه مشاهده کنید.'
              },
              {
                icon: Search,
                title: 'فیلتر و نمای ذخیره‌شده',
                desc: 'فیلترهای پرتکرار را برای بررسی‌های عملیاتی نگه دارید و درخواست‌های مرتبط را سریع‌تر پیدا کنید.'
              },
              {
                icon: HelpCircle,
                title: 'خروجی برای تحلیل بیرونی',
                desc: 'لاگ‌ها را با قالب CSV یا JSONL برای تحلیل، گزارش‌گیری یا نگهداری در سامانه دیگر دریافت کنید.'
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
              چرا پایش متمرکز برای تیم‌های ایرانی مهم است؟
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
                desc: 'عبور درخواست‌ها از یک درگاه، وابستگی کد محصول به اتصال مستقیم هر ارائه‌دهنده را کمتر می‌کند و تغییر مسیر را ساده‌تر می‌سازد.'
              },
              {
                icon: CreditCard,
                title: 'بدون پرداخت ارزی مستقیم',
                desc: 'هزینه مدل‌ها را بدون نیاز به کارت اعتباری خارجی یا پرداخت دلاری مستقیم مدیریت و با نمایش ریالی بررسی کنید.'
              },
              {
                icon: Crosshair,
                title: 'هزینه قابل مشاهده به ریال',
                desc: 'نمایش هزینه به ریال کمک می‌کند تیم فنی و مالی درباره مصرف مدل‌ها با واحدی مشترک تصمیم بگیرند.'
              },
              {
                icon: Globe,
                title: 'یک نقطه کنترل برای چند مدل و تأمین‌کننده',
                desc: 'به‌جای بررسی جداگانه چند ارائه‌دهنده، درخواست‌ها، هزینه و خطاها را از یک مسیر مشترک تحلیل کنید.'
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
                title: 'تیم‌های محصول هوش مصنوعی',
                desc: 'الگوی استفاده، خطا و زمان پاسخ قابلیت‌های مبتنی بر مدل را بررسی کنید و نقاط اصطکاک محصول را پیدا کنید.'
              },
              {
                icon: Code,
                title: 'تیم‌های Backend و Platform',
                desc: 'تأخیر، خطا، تلاش دوباره، مهلت پاسخ و مسیر جایگزین را بررسی کنید و علت ناپایداری را پیدا کنید.'
              },
              {
                icon: Building,
                title: 'تیم‌های مالی و مدیریت',
                desc: 'مصرف مدل را بر اساس تیم، قابلیت محصول، کاربر یا فضای کاری ببینید و هزینه‌ها را قابل پیش‌بینی‌تر کنید.'
              },
              {
                icon: Users,
                title: 'تیم‌های پشتیبانی مشتری',
                desc: 'درخواست‌های خطادار یا کند در مسیر پاسخ‌گویی خودکار را با شناسه درخواست پیدا و برای تیم فنی ارسال کنید.'
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

      {/* 10. FAQ Section */}
      <section className="py-24 bg-slate-50 dark:bg-slate-900/20 border-t border-slate-200 dark:border-slate-800">
        <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 className="text-3xl font-bold text-slate-900 dark:text-white" style={{ fontFamily: 'var(--font-heading)' }}>
              پرسش‌های پرتکرار
            </h2>
          </div>

          <FAQAccordion items={[
              {
                q: "تفاوت لاگ و رد درخواست چیست؟",
                a: "لاگ، جزئیات یک درخواست مانند مدل، وضعیت، توکن و هزینه را نشان می‌دهد. رد درخواست، تلاش‌ها و مراحل اجرای همان درخواست را در یک خط زمانی نمایش می‌دهد."
              },
              {
                q: "آیا تأخیر و مصرف توکن برای مدل‌ها و تیم‌های مختلف قابل مشاهده است؟",
                a: "بله. می‌توانید تأخیر، مصرف توکن، هزینه، نرخ خطا و تعداد درخواست‌ها را بر اساس مدل، ارائه‌دهنده، فضای کاری، کاربر یا تیم بررسی کنید."
              },
              {
                q: "چطور مدل یا ارائه‌دهنده پرهزینه را پیدا کنم؟",
                a: "در سنجه‌های مدل، هزینه را بر اساس مدل یا ارائه‌دهنده مقایسه کنید. سپس از لاگ درخواست‌ها نمونه‌های پرهزینه را باز کنید و جزئیات مصرف آن‌ها را ببینید."
              },
              {
                q: "آیا تلاش دوباره و مسیر جایگزین قابل پیگیری است؟",
                a: "بله. رد درخواست نشان می‌دهد هر تلاش به کدام مقصد رفته، چه خطایی رخ داده و مسیر جایگزین چگونه ادامه پیدا کرده است."
              },
              {
                q: "آیا برای استفاده از گدارAI نیاز به پرداخت ارزی دارم؟",
                a: "خیر. گدارAI برای تیم‌های ایرانی طراحی شده و هدف آن ساده‌کردن دسترسی، مدیریت و پایش LLMها بدون نیاز به پرداخت ارزی مستقیم است."
              },
              {
                q: "آیا محتوای کامل درخواست و پاسخ همیشه ذخیره می‌شود؟",
                a: "خیر. شیوه ثبت محتوا به تنظیمات لاگ و حریم خصوصی ردیابی بستگی دارد. برای شروع امن‌تر می‌توانید فقط فراداده را نگه دارید."
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
          <h2 className="text-3xl md:text-5xl font-bold text-white mb-6" style={{ fontFamily: 'var(--font-heading)' }}>رفتار درخواست‌های مدل را با داده بررسی کنید</h2>
          <p className="text-lg md:text-xl text-indigo-100 mb-10 leading-relaxed max-w-3xl mx-auto">
            از سنجه‌ها برای یافتن روند، از لاگ برای بررسی نمونه واقعی و از رد درخواست برای پیدا کردن علت خطا یا تأخیر استفاده کنید.
          </p>
          
          <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
            <Link href="/pricing" className="w-full sm:w-auto px-8 py-4 bg-white text-indigo-600 hover:bg-slate-50 rounded-xl font-bold text-lg shadow-xl transition-all active:scale-95 flex items-center justify-center gap-2 group">
              شروع کنید
              <ArrowLeft className="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
            </Link>
            <Link href="/demo" className="w-full sm:w-auto px-8 py-4 bg-transparent border-2 border-white/30 hover:border-white text-white rounded-xl font-bold text-lg transition-all active:scale-95 flex items-center justify-center">
              درخواست دمو
            </Link>
          </div>
        </div>
      </section>
    </main>
  );
}

