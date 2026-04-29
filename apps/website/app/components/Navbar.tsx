'use client';

import { motion, AnimatePresence } from 'motion/react';
import { Moon, Sun, Menu, X, ArrowLeft, ChevronDown, Activity, Network, ShieldCheck, TerminalSquare, Bot, Lock, Server } from 'lucide-react';
import { useTheme } from './ThemeProvider';
import { useState } from 'react';
import Link from 'next/link';

const features = [
  { name: 'پایش و مشاهده‌پذیری', description: 'ثبت لاگ، ردیابی و پایش متریک‌های کلیدی', icon: Activity, href: '/features/observability' },
  { name: 'درگاه هوش مصنوعی', description: 'مسیریابی به بیش از ۱۶۰۰ مدل زبانی بزرگ', icon: Network, href: '/features/ai-gateway' },
  { name: 'حفاظت و کنترل', description: 'اعمال کنترل و محدودیت‌های بلادرنگ روی رفتار مدل‌ها', icon: ShieldCheck, href: '/features/guardrails' },
  { name: 'مهندسی پرامپت', description: 'ساخت، نسخه‌بندی و استقرار پرامپت‌ها', icon: TerminalSquare, href: '/features/prompt-engineering' },
  { name: 'ایجنت‌ها', description: 'استقرار ایجنت‌های هوش مصنوعی در محیط عملیاتی', icon: Bot, href: '/features/agents' },
  { name: 'امنیت و تطابق', description: 'امنیت در سطح سازمانی و کنترل دسترسی‌ها', icon: Lock, href: '/features/security' },
  { name: 'درگاه MCP', description: 'استقرار و مدیریت سرورهای MCP در محیط عملیاتی', icon: Server, href: '/features/mcp-gateway' },
];

export function Navbar() {
  const { theme, setTheme } = useTheme();
  const [isOpen, setIsOpen] = useState(false);
  const [isMobileFeaturesOpen, setIsMobileFeaturesOpen] = useState(false);

  return (
    <nav className="sticky top-0 z-50 w-full backdrop-blur-xl bg-white/70 dark:bg-[#0B0F19]/70 border-b border-gray-200 dark:border-gray-800 transition-colors duration-300">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between h-16 items-center">
          {/* Logo & Menu */}
          <div className="flex items-center gap-8">
            <Link href="/" className="flex items-center gap-2">
              <div className="w-8 h-8 rounded-lg bg-indigo-500 flex items-center justify-center text-white font-bold text-xl">ش</div>
              <span className="font-bold text-xl text-slate-900 dark:text-white">شاهراه</span>
            </Link>
            <div className="hidden md:flex items-center gap-6 h-16">
              <div className="relative group/features h-full flex items-center">
                <button className="text-sm font-medium text-slate-600 dark:text-gray-300 group-hover/features:text-indigo-500 dark:group-hover/features:text-indigo-400 transition-colors flex items-center gap-1">
                  امکانات
                  <ChevronDown className="w-4 h-4 transition-transform duration-300 group-hover/features:rotate-180" />
                </button>
                <div className="absolute top-full right-0 w-[650px] opacity-0 invisible group-hover/features:opacity-100 group-hover/features:visible transition-all duration-300 -translate-y-2 group-hover/features:translate-y-0">
                  <div className="bg-white dark:bg-[#0B0F19] rounded-2xl shadow-[0_10px_40px_-10px_rgba(0,0,0,0.1)] dark:shadow-[0_10px_40px_-10px_rgba(0,0,0,0.5)] border border-gray-100 dark:border-gray-800/60 p-5 grid grid-cols-2 gap-2 relative before:absolute before:inset-0 before:bg-gradient-to-b before:from-indigo-500/5 before:to-transparent before:rounded-2xl before:pointer-events-none">
                    {features.map((feature, idx) => (
                      <Link key={idx} href={feature.href} className="flex items-start gap-4 p-3 rounded-xl hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors group/item relative z-10">
                        <div className="p-2.5 bg-indigo-50 dark:bg-indigo-500/10 text-indigo-600 dark:text-indigo-400 rounded-xl group-hover/item:bg-indigo-500 group-hover/item:text-white transition-colors shadow-sm">
                          <feature.icon className="w-5 h-5" />
                        </div>
                        <div>
                          <div className="text-sm font-bold text-slate-900 dark:text-white mb-1.5 transition-colors group-hover/item:text-indigo-600 dark:group-hover/item:text-indigo-400" style={{ fontFamily: 'var(--font-heading)' }}>{feature.name}</div>
                          <div className="text-xs text-slate-500 dark:text-slate-400 leading-relaxed">{feature.description}</div>
                        </div>
                      </Link>
                    ))}
                  </div>
                </div>
              </div>
              <Link href="/pricing" className="text-sm font-medium text-slate-600 dark:text-gray-300 hover:text-indigo-500 dark:hover:text-indigo-400 transition-colors">قیمت‌گذاری</Link>
              <Link href="/about" className="text-sm font-medium text-slate-600 dark:text-gray-300 hover:text-indigo-500 dark:hover:text-indigo-400 transition-colors">درباره ما</Link>
              <Link href="/blog" className="text-sm font-medium text-slate-600 dark:text-gray-300 hover:text-indigo-500 dark:hover:text-indigo-400 transition-colors">وبلاگ</Link>
              <Link href="/contact" className="text-sm font-medium text-slate-600 dark:text-gray-300 hover:text-indigo-500 dark:hover:text-indigo-400 transition-colors">تماس</Link>
            </div>
          </div>

          {/* Actions */}
          <div className="hidden md:flex items-center gap-4">
            <button
              onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
              className="p-2 rounded-full text-slate-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
            >
              {theme === 'dark' ? <Sun className="w-5 h-5" /> : <Moon className="w-5 h-5" />}
            </button>
            <Link href="/signup" className="flex items-center gap-2 px-4 py-2 rounded-lg bg-indigo-500 hover:bg-indigo-600 text-white text-sm font-medium transition-all shadow-[0_0_15px_rgba(99,102,241,0.3)] hover:shadow-[0_0_25px_rgba(99,102,241,0.5)]">
              شروع کنید
              <ArrowLeft className="w-4 h-4" />
            </Link>
          </div>

          {/* Mobile menu button */}
          <div className="md:hidden flex items-center gap-4">
            <button
              onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
              className="p-2 rounded-full text-slate-600 dark:text-gray-300"
            >
              {theme === 'dark' ? <Sun className="w-5 h-5" /> : <Moon className="w-5 h-5" />}
            </button>
            <button onClick={() => setIsOpen(!isOpen)} className="text-slate-600 dark:text-gray-300">
              {isOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
            </button>
          </div>
        </div>
      </div>

      {/* Mobile Menu */}
      {isOpen && (
        <div className="md:hidden border-t border-gray-200 dark:border-gray-800 bg-white dark:bg-[#0B0F19] overflow-y-auto max-h-[calc(100vh-64px)]">
          <div className="px-4 pt-2 pb-6 flex flex-col gap-4">
            <div>
              <button 
                onClick={() => setIsMobileFeaturesOpen(!isMobileFeaturesOpen)}
                className="flex items-center justify-between w-full text-sm font-medium text-slate-600 dark:text-gray-300 py-2"
              >
                امکانات
                <ChevronDown className={`w-4 h-4 transition-transform ${isMobileFeaturesOpen ? 'rotate-180' : ''}`} />
              </button>
              {isMobileFeaturesOpen && (
                <div className="pl-4 pr-2 py-2 flex flex-col gap-3 border-r-2 border-indigo-100 dark:border-gray-800 mr-2 mt-2">
                  {features.map((feature, idx) => (
                    <Link 
                      key={idx} 
                      href={feature.href} 
                      className="flex items-start gap-3"
                      onClick={() => {
                        setIsOpen(false);
                        setIsMobileFeaturesOpen(false);
                      }}
                    >
                      <feature.icon className="w-5 h-5 text-indigo-500 mt-0.5" />
                      <div>
                        <div className="text-sm font-bold text-slate-900 dark:text-white mb-0.5">{feature.name}</div>
                        <div className="text-xs text-slate-500 dark:text-slate-400">{feature.description}</div>
                      </div>
                    </Link>
                  ))}
                </div>
              )}
            </div>
            <Link href="/pricing" className="text-sm font-medium text-slate-600 dark:text-gray-300 block py-2" onClick={() => setIsOpen(false)}>قیمت‌گذاری</Link>
            <Link href="/about" className="text-sm font-medium text-slate-600 dark:text-gray-300 block py-2" onClick={() => setIsOpen(false)}>درباره ما</Link>
            <Link href="/blog" className="text-sm font-medium text-slate-600 dark:text-gray-300 block py-2" onClick={() => setIsOpen(false)}>وبلاگ</Link>
            <Link href="/contact" className="text-sm font-medium text-slate-600 dark:text-gray-300 block py-2" onClick={() => setIsOpen(false)}>تماس</Link>
            <div className="border-t border-gray-200 dark:border-gray-800 pt-4 flex flex-col gap-4">
              <Link href="/signup" className="text-center block px-4 py-2 rounded-lg bg-indigo-500 text-white text-sm font-medium" onClick={() => setIsOpen(false)}>شروع کنید</Link>
            </div>
          </div>
        </div>
      )}
    </nav>
  );
}


