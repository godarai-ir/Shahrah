'use client';

import { Moon, Sun, Menu, X, ArrowLeft, ChevronDown, Activity, Network, ShieldCheck, TerminalSquare, Lock, Database, CalendarDays, Scale } from 'lucide-react';
import { useTheme } from './ThemeProvider';
import { useState } from 'react';
import Link from 'next/link';
import { GodarAILogo } from './GodarAILogo';
import { solutionGroups, solutionHref, solutions } from '@/app/lib/solutions';

const features = [
  { name: 'پایش و مشاهده‌پذیری', description: 'گزارش رخداد، رد درخواست و سنجه‌های مصرف', icon: Activity, href: '/features/observability' },
  { name: 'درگاه مدل‌های زبانی', description: 'مسیریابی، مدل مجازی و کنترل هزینه از یک مسیر', icon: Network, href: '/features/ai-gateway' },
  { name: 'کشف و مقایسه مدل‌ها', description: 'جست‌وجو، فیلتر و مقایسه مدل‌ها از ۲۱ پروایدر', icon: Database, href: '/features/model-discovery' },
  { name: 'گاردریل‌ها', description: 'کنترل ورودی، خروجی و داده حساس در مسیر مدل', icon: ShieldCheck, href: '/features/guardrails' },
  { name: 'رجیستری پرامپت', description: 'ساخت، نسخه‌بندی و اجرای پرامپت‌های قابل تکرار', icon: TerminalSquare, href: '/features/prompt-engineering' },
  { name: 'امنیت و تطابق', description: 'امنیت در سطح سازمانی و کنترل دسترسی‌ها', icon: Lock, href: '/features/security' },
];

const mobileFeatures = features;

const comparisons = [
  { name: 'گدارAI در برابر LiteLLM', description: 'تفاوت‌ها و زمان مناسب مهاجرت', icon: Scale, href: '/vs/litellm' },
  { name: 'گدارAI در برابر Portkey', description: 'درگاه مستقل با استقرار کامل داخلی', icon: Scale, href: '/vs/portkey' },
];

const primaryLinks = [
  { label: 'مستندات', href: '/docs' },
  { label: 'کاتالوگ مدل‌ها', href: '/catalog' },
  { label: 'قیمت‌گذاری', href: '/pricing' },
];

const actionLinks = [
  { label: 'ورود', href: '/login', variant: 'secondary' },
  { label: '۱۵ روز رایگان شروع کنید', href: '/onboarding', variant: 'primary' },
] as const;

export function Navbar() {
  const { resolvedTheme, setTheme } = useTheme();
  const [isOpen, setIsOpen] = useState(false);

  return (
    <nav className="sticky top-0 z-50 w-full backdrop-blur-xl bg-white/70 dark:bg-[#0B0F19]/70 border-b border-gray-200 dark:border-gray-800 transition-colors duration-300">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex min-h-16 justify-between items-center gap-3">
          {/* Logo & Menu */}
          <div className="flex min-w-0 items-center gap-4 sm:gap-8">
            <Link href="/" className="flex min-w-0 items-center gap-2">
              <GodarAILogo
                decorative
                priority
                className="h-9 w-auto max-w-12 shrink-0 object-contain"
                sizes="3rem"
              />
              <span className="truncate font-bold text-lg sm:text-xl text-slate-900 dark:text-white">گدارAI</span>
            </Link>
            <div className="hidden lg:flex items-center gap-6 h-16">
              <div className="relative group/features h-full flex items-center">
                <button className="text-sm font-medium text-slate-600 dark:text-gray-300 group-hover/features:text-indigo-500 dark:group-hover/features:text-indigo-400 transition-colors flex items-center gap-1">
                  امکانات
                  <ChevronDown className="w-4 h-4 transition-transform duration-300 group-hover/features:rotate-180" />
                </button>
                <div className="absolute top-full right-0 w-[720px] opacity-0 invisible group-hover/features:opacity-100 group-hover/features:visible transition-all duration-300 -translate-y-2 group-hover/features:translate-y-0">
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
              <div className="relative group/solutions h-full flex items-center">
                <button className="text-sm font-medium text-slate-600 dark:text-gray-300 group-hover/solutions:text-indigo-500 dark:group-hover/solutions:text-indigo-400 transition-colors flex items-center gap-1">
                  راهکارها
                  <ChevronDown className="w-4 h-4 transition-transform duration-300 group-hover/solutions:rotate-180" />
                </button>
                <div className="absolute top-full right-0 w-[760px] opacity-0 invisible group-hover/solutions:opacity-100 group-hover/solutions:visible transition-all duration-300 -translate-y-2 group-hover/solutions:translate-y-0">
                  <div className="bg-white dark:bg-[#0B0F19] rounded-2xl shadow-[0_10px_40px_-10px_rgba(0,0,0,0.1)] dark:shadow-[0_10px_40px_-10px_rgba(0,0,0,0.5)] border border-gray-100 dark:border-gray-800/60 p-5 grid grid-cols-2 gap-x-4 relative before:absolute before:inset-0 before:bg-gradient-to-b before:from-indigo-500/5 before:to-transparent before:rounded-2xl before:pointer-events-none">
                    {solutionGroups.map((group) => (
                      <div key={group.id} className="relative z-10">
                        <div className="px-3 pb-2 text-xs font-bold text-slate-400 dark:text-slate-500">{group.label}</div>
                        <div className="flex flex-col gap-1">
                          {solutions.filter((solution) => solution.group === group.id).map((solution) => (
                            <Link key={solution.slug} href={solutionHref(solution.slug)} className="flex items-start gap-4 p-3 rounded-xl hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors group/item">
                              <div className="p-2.5 bg-indigo-50 dark:bg-indigo-500/10 text-indigo-600 dark:text-indigo-400 rounded-xl group-hover/item:bg-indigo-500 group-hover/item:text-white transition-colors shadow-sm">
                                <solution.icon className="w-5 h-5" />
                              </div>
                              <div>
                                <div className="text-sm font-bold text-slate-900 dark:text-white mb-1.5 transition-colors group-hover/item:text-indigo-600 dark:group-hover/item:text-indigo-400" style={{ fontFamily: 'var(--font-heading)' }}>{solution.navName}</div>
                                <div className="text-xs text-slate-500 dark:text-slate-400 leading-relaxed">{solution.navDescription}</div>
                              </div>
                            </Link>
                          ))}
                        </div>
                      </div>
                    ))}
                    <div className="relative z-10 col-span-2 mt-3 border-t border-gray-100 pt-3 dark:border-gray-800/60">
                      <div className="px-3 pb-2 text-xs font-bold text-slate-400 dark:text-slate-500">مقایسه‌ها</div>
                      <div className="grid grid-cols-2 gap-x-4">
                        {comparisons.map((item) => (
                          <Link key={item.href} href={item.href} className="flex items-start gap-4 p-3 rounded-xl hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors group/item">
                            <div className="p-2.5 bg-indigo-50 dark:bg-indigo-500/10 text-indigo-600 dark:text-indigo-400 rounded-xl group-hover/item:bg-indigo-500 group-hover/item:text-white transition-colors shadow-sm">
                              <item.icon className="w-5 h-5" />
                            </div>
                            <div>
                              <div className="text-sm font-bold text-slate-900 dark:text-white mb-1.5 transition-colors group-hover/item:text-indigo-600 dark:group-hover/item:text-indigo-400" style={{ fontFamily: 'var(--font-heading)' }}>{item.name}</div>
                              <div className="text-xs text-slate-500 dark:text-slate-400 leading-relaxed">{item.description}</div>
                            </div>
                          </Link>
                        ))}
                      </div>
                    </div>
                  </div>
                </div>
              </div>
              {primaryLinks.map((item) => (
                <Link
                  key={item.href}
                  href={item.href}
                  className="text-sm font-medium text-slate-600 transition-colors hover:text-indigo-500 dark:text-gray-300 dark:hover:text-indigo-400"
                >
                  {item.label}
                </Link>
              ))}
            </div>
          </div>

          {/* Actions */}
          <div className="hidden lg:flex items-center gap-4">
            <button
              onClick={() => setTheme(resolvedTheme === 'dark' ? 'light' : 'dark')}
              className="flex items-center justify-center p-2 rounded-full text-slate-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
            >
              {resolvedTheme === 'dark' ? <Sun className="w-5 h-5" /> : <Moon className="w-5 h-5" />}
            </button>
            <Link
              href="/demo"
              className="flex items-center gap-2 rounded-lg border border-indigo-200 px-4 py-2 text-sm font-medium text-indigo-700 transition-colors hover:border-indigo-300 hover:bg-indigo-50 dark:border-indigo-500/30 dark:text-indigo-300 dark:hover:bg-indigo-500/10"
            >
              رزرو وقت دمو
              <CalendarDays className="h-4 w-4" />
            </Link>
            {actionLinks.map((item) => (
              <Link
                key={item.href}
                href={item.href}
                className={
                  item.variant === 'primary'
                    ? 'flex items-center gap-2 rounded-lg bg-indigo-500 px-4 py-2 text-sm font-medium text-white transition-all shadow-[0_0_15px_rgba(99,102,241,0.3)] hover:bg-indigo-600 hover:shadow-[0_0_25px_rgba(99,102,241,0.5)]'
                    : 'text-sm font-medium text-slate-600 transition-colors hover:text-indigo-500 dark:text-gray-300 dark:hover:text-indigo-400'
                }
              >
                {item.label}
                {item.variant === 'primary' ? <ArrowLeft className="h-4 w-4" /> : null}
              </Link>
            ))}
          </div>

          {/* Mobile menu button */}
          <div className="flex shrink-0 items-center gap-2 lg:hidden">
            <button
              aria-label={isOpen ? "Close menu" : "Open menu"}
              aria-expanded={isOpen}
              aria-controls="mobile-menu"
              onClick={() => setIsOpen(!isOpen)}
              className="flex min-h-11 min-w-11 items-center justify-center text-slate-600 dark:text-gray-300"
            >
              {isOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
            </button>
            <button
              aria-label="Toggle theme"
              onClick={() => setTheme(resolvedTheme === 'dark' ? 'light' : 'dark')}
              className="flex min-h-11 min-w-11 items-center justify-center p-2 rounded-full text-slate-600 dark:text-gray-300"
            >
              {resolvedTheme === 'dark' ? <Sun className="w-5 h-5" /> : <Moon className="w-5 h-5" />}
            </button>
          </div>
        </div>
      </div>

      {/* Mobile Menu */}
      {isOpen && (
        <div id="mobile-menu" className="lg:hidden border-t border-gray-200 dark:border-gray-800 bg-white dark:bg-[#0B0F19] overflow-y-auto max-h-[calc(100vh-64px)]">
          <div className="px-4 pt-2 pb-6 flex flex-col gap-3">
            {mobileFeatures.map((feature) => (
              <Link
                key={feature.href}
                href={feature.href}
                className="flex min-h-11 items-start gap-3 py-2"
                onClick={() => setIsOpen(false)}
              >
                <feature.icon className="mt-0.5 h-5 w-5 shrink-0 text-indigo-500" />
                <span className="min-w-0">
                  <span className="block text-sm font-bold text-slate-900 dark:text-white">{feature.name}</span>
                  <span className="block truncate text-xs text-slate-500 dark:text-slate-400">{feature.description}</span>
                </span>
              </Link>
            ))}
            <div className="border-t border-gray-200 dark:border-gray-800 pt-4 flex flex-col gap-3">
              <span className="text-xs font-bold text-slate-400 dark:text-slate-500">راهکارها</span>
              {solutions.map((solution) => (
                <Link
                  key={solution.slug}
                  href={solutionHref(solution.slug)}
                  className="flex min-h-11 items-start gap-3 py-2"
                  onClick={() => setIsOpen(false)}
                >
                  <solution.icon className="mt-0.5 h-5 w-5 shrink-0 text-indigo-500" />
                  <span className="min-w-0">
                    <span className="block text-sm font-bold text-slate-900 dark:text-white">{solution.navName}</span>
                    <span className="block truncate text-xs text-slate-500 dark:text-slate-400">{solution.navDescription}</span>
                  </span>
                </Link>
              ))}
            </div>
            <div className="border-t border-gray-200 dark:border-gray-800 pt-4 flex flex-col gap-3">
              <span className="text-xs font-bold text-slate-400 dark:text-slate-500">مقایسه‌ها</span>
              {comparisons.map((item) => (
                <Link
                  key={item.href}
                  href={item.href}
                  className="flex min-h-11 items-start gap-3 py-2"
                  onClick={() => setIsOpen(false)}
                >
                  <item.icon className="mt-0.5 h-5 w-5 shrink-0 text-indigo-500" />
                  <span className="min-w-0">
                    <span className="block text-sm font-bold text-slate-900 dark:text-white">{item.name}</span>
                    <span className="block truncate text-xs text-slate-500 dark:text-slate-400">{item.description}</span>
                  </span>
                </Link>
              ))}
            </div>
            <div className="border-t border-gray-200 dark:border-gray-800 pt-4 flex flex-col gap-3">
              {[
                { label: 'کاتالوگ مدل‌ها', href: '/catalog' },
                { label: 'قیمت‌گذاری', href: '/pricing' },
              ].map((item) => (
                <Link
                  key={item.href}
                  href={item.href}
                  className="block min-h-11 py-2 text-sm font-bold text-slate-900 dark:text-white"
                  onClick={() => setIsOpen(false)}
                >
                  {item.label}
                </Link>
              ))}
            </div>
            <div className="border-t border-gray-200 dark:border-gray-800 pt-4 flex flex-col gap-4">
              <Link href="/login" className="block min-h-12 rounded-lg border border-gray-200 px-4 py-3 text-center text-sm font-medium text-slate-700 dark:border-gray-800 dark:text-gray-200" onClick={() => setIsOpen(false)}>ورود</Link>
              <Link href="/demo" className="flex min-h-12 items-center justify-center gap-2 rounded-lg border border-indigo-200 px-4 py-3 text-center text-sm font-medium text-indigo-700 dark:border-indigo-500/30 dark:text-indigo-300" onClick={() => setIsOpen(false)}>
                رزرو وقت دمو
                <CalendarDays className="h-4 w-4" />
              </Link>
              <Link href="/onboarding" className="block min-h-12 rounded-lg bg-indigo-500 px-4 py-3 text-center text-sm font-medium text-white" onClick={() => setIsOpen(false)}>۱۵ روز رایگان شروع کنید</Link>
            </div>
          </div>
        </div>
      )}
    </nav>
  );
}


