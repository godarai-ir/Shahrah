'use client';
import { CalendarDays } from 'lucide-react';
import Link from 'next/link';
import { SOCIAL_LINKS } from '@/app/lib/social-links';
import { GodarAILogo } from './GodarAILogo';
import { solutionHref, solutions } from '@/app/lib/solutions';

export function Footer() {
  return (
    <footer className="bg-slate-50 dark:bg-[#0B0F19] border-t border-gray-200 dark:border-gray-800 py-8 transition-colors">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="mb-8 flex flex-col items-start justify-between gap-4 rounded-2xl border border-indigo-200 bg-indigo-50 p-5 dark:border-indigo-500/20 dark:bg-indigo-500/10 sm:flex-row sm:items-center">
          <div>
            <h2 className="font-bold text-slate-950 dark:text-white">گدارAI را متناسب با نیاز تیم خود ببینید</h2>
            <p className="mt-1 text-sm leading-6 text-slate-600 dark:text-slate-300">یک جلسه خصوصی ۶۰ دقیقه‌ای با تیم ما رزرو کنید.</p>
          </div>
          <Link href="/demo" className="flex min-h-11 w-full shrink-0 items-center justify-center gap-2 rounded-lg bg-indigo-600 px-5 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-indigo-700 sm:w-auto">
            رزرو وقت دمو
            <CalendarDays className="w-5 h-5" />
          </Link>
        </div>
        <div className="mb-6 grid grid-cols-2 gap-x-4 gap-y-6 sm:grid-cols-4 md:grid-cols-5 md:gap-6">
          <div className="col-span-2 sm:col-span-4 md:col-span-1">
            <div className="mb-3 flex items-center gap-2">
              <GodarAILogo
                decorative
                className="h-9 w-auto max-w-12 shrink-0 object-contain"
                sizes="3rem"
              />
              <span className="font-bold text-xl text-slate-900 dark:text-white">گدارAI</span>
            </div>
            <p className="mb-3 max-w-2xl text-xs leading-6 text-slate-500 dark:text-gray-400 md:text-sm">
              درگاه مدل‌های زبانی برای اتصال، مسیریابی، کنترل هزینه و پایش مصرف مدل‌ها در تیم‌های ایرانی.
            </p>
            <div className="flex items-center gap-3">
              <a href={SOCIAL_LINKS.linkedin} target="_blank" rel="noopener noreferrer" aria-label="گدارAI در لینکدین" className="text-slate-400 hover:text-slate-600 dark:hover:text-white transition-colors">
                <svg className="h-4 w-4 md:h-5 md:w-5" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                  <path d="M20.45 20.45h-3.56v-5.57c0-1.33-.03-3.04-1.85-3.04-1.85 0-2.14 1.45-2.14 2.94v5.67H9.34V8.98h3.42v1.57h.05c.48-.9 1.64-1.85 3.37-1.85 3.61 0 4.27 2.37 4.27 5.46v6.29ZM5.32 7.41a2.06 2.06 0 1 1 0-4.12 2.06 2.06 0 0 1 0 4.12Zm1.78 13.04H3.54V8.98H7.1v11.47Z" />
                </svg>
              </a>
              <a href={SOCIAL_LINKS.telegram} target="_blank" rel="noopener noreferrer" aria-label="گدارAI در تلگرام" className="text-slate-400 hover:text-slate-600 dark:hover:text-white transition-colors">
                <svg className="h-4 w-4 md:h-5 md:w-5" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                  <path d="M21.94 4.66c.26-1.21-.44-1.69-1.21-1.4L2.4 10.33c-1.25.49-1.23 1.19-.21 1.5l4.71 1.47 10.91-6.88c.52-.31.99-.14.6.21l-8.84 7.98-.33 4.67c.48 0 .69-.22.96-.48l2.31-2.25 4.8 3.55c.89.49 1.52.24 1.75-.82l2.88-14.62Z" />
                </svg>
              </a>
              <a href={SOCIAL_LINKS.github} target="_blank" rel="noopener noreferrer" aria-label="گدارAI در گیت‌هاب" className="text-slate-400 hover:text-slate-600 dark:hover:text-white transition-colors">
                <svg className="h-4 w-4 md:h-5 md:w-5" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                  <path fillRule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" clipRule="evenodd" />
                </svg>
              </a>
            </div>
          </div>
          
          <div className="min-w-0">
            <h4 className="mb-2 text-sm font-bold text-slate-900 dark:text-white md:mb-3 md:text-base">محصول</h4>
            <ul className="space-y-1.5 text-xs md:space-y-2 md:text-sm">
              <li><Link href="/" className="text-slate-600 dark:text-gray-400 hover:text-indigo-500 dark:hover:text-indigo-400">گدارAI در یک نگاه</Link></li>
              <li><Link href="/features/observability" className="text-slate-600 dark:text-gray-400 hover:text-indigo-500 dark:hover:text-indigo-400">پایش و مشاهده‌پذیری</Link></li>
              <li><Link href="/features/ai-gateway" className="text-slate-600 dark:text-gray-400 hover:text-indigo-500 dark:hover:text-indigo-400">درگاه مدل‌های AI</Link></li>
              <li><Link href="/features/guardrails" className="text-slate-600 dark:text-gray-400 hover:text-indigo-500 dark:hover:text-indigo-400">گاردریل‌ها</Link></li>
              <li><Link href="/features/prompt-engineering" className="text-slate-600 dark:text-gray-400 hover:text-indigo-500 dark:hover:text-indigo-400">رجیستری پرامپت</Link></li>
              <li><Link href="/features/security" className="text-slate-600 dark:text-gray-400 hover:text-indigo-500 dark:hover:text-indigo-400">امنیت و تطابق</Link></li>
              <li><Link href="/pricing" className="text-slate-600 dark:text-gray-400 hover:text-indigo-500 dark:hover:text-indigo-400">پلن‌های اشتراک</Link></li>
            </ul>
          </div>

          <div className="min-w-0">
            <h4 className="mb-2 text-sm font-bold text-slate-900 dark:text-white md:mb-3 md:text-base">راهکارها</h4>
            <ul className="space-y-1.5 text-xs md:space-y-2 md:text-sm">
              {solutions.map((solution) => (
                <li key={solution.slug}><Link href={solutionHref(solution.slug)} className="text-slate-600 dark:text-gray-400 hover:text-indigo-500 dark:hover:text-indigo-400">{solution.navName}</Link></li>
              ))}
            </ul>
          </div>

          <div className="min-w-0">
            <h4 className="mb-2 text-sm font-bold text-slate-900 dark:text-white md:mb-3 md:text-base">منابع</h4>
            <ul className="space-y-1.5 text-xs md:space-y-2 md:text-sm">
              <li><Link href="/docs" className="text-slate-600 dark:text-gray-400 hover:text-indigo-500 dark:hover:text-indigo-400">مستندات</Link></li>
              <li><Link href="/blog" className="text-slate-600 dark:text-gray-400 hover:text-indigo-500 dark:hover:text-indigo-400">وبلاگ</Link></li>
              <li><Link href="/catalog" className="text-slate-600 dark:text-gray-400 hover:text-indigo-500 dark:hover:text-indigo-400">کاتالوگ مدل‌ها</Link></li>
              <li><Link href="/vs/litellm" className="text-slate-600 dark:text-gray-400 hover:text-indigo-500 dark:hover:text-indigo-400">گدارAI در برابر LiteLLM</Link></li>
              <li><Link href="/vs/portkey" className="text-slate-600 dark:text-gray-400 hover:text-indigo-500 dark:hover:text-indigo-400">گدارAI در برابر Portkey</Link></li>
              <li><a href={SOCIAL_LINKS.github} target="_blank" rel="noopener noreferrer" className="text-slate-600 dark:text-gray-400 hover:text-indigo-500 dark:hover:text-indigo-400">گیت‌هاب</a></li>
            </ul>
          </div>

          <div className="min-w-0">
            <h4 className="mb-2 text-sm font-bold text-slate-900 dark:text-white md:mb-3 md:text-base">شرکت</h4>
            <ul className="space-y-1.5 text-xs md:space-y-2 md:text-sm">
              <li><Link href="/about" className="text-slate-600 dark:text-gray-400 hover:text-indigo-500 dark:hover:text-indigo-400">درباره ما</Link></li>
              <li><Link href="/contact" className="text-slate-600 dark:text-gray-400 hover:text-indigo-500 dark:hover:text-indigo-400">تماس با ما</Link></li>
              <li><Link href="/demo" className="text-slate-600 dark:text-gray-400 hover:text-indigo-500 dark:hover:text-indigo-400">رزرو وقت دمو</Link></li>
              <li><Link href="/on-premise" className="text-slate-600 dark:text-gray-400 hover:text-indigo-500 dark:hover:text-indigo-400">استقرار درون‌سازمانی</Link></li>
              <li><Link href="/terms" className="text-slate-600 dark:text-gray-400 hover:text-indigo-500 dark:hover:text-indigo-400">قوانین و مقررات</Link></li>
              <li><Link href="/privacy" className="text-slate-600 dark:text-gray-400 hover:text-indigo-500 dark:hover:text-indigo-400">حریم خصوصی</Link></li>
            </ul>
          </div>
        </div>
        
        <div className="grid grid-cols-[1fr_auto] items-center gap-3 border-t border-gray-200 pt-4 text-right dark:border-gray-800 md:pt-6">
          <p className="min-w-0 text-xs leading-5 text-slate-500 dark:text-gray-500 md:text-sm">
            © ۱۴۰۵ گدارAI. تمامی حقوق محفوظ است.
          </p>
          <a
            referrerPolicy="origin"
            target="_blank"
            rel="noopener noreferrer"
            href="https://trustseal.enamad.ir/?id=737580&Code=Dc476GaakpMUDp6MG3lSrtPK8uL9943N"
            className="inline-flex h-11 w-14 shrink-0 items-center justify-center overflow-hidden rounded-lg border border-gray-200 bg-white p-1 text-[10px] leading-4 text-slate-500 dark:border-gray-800 dark:bg-[#111827] dark:text-gray-400 md:h-14 md:w-16"
          >
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              referrerPolicy="origin"
              src="https://trustseal.enamad.ir/logo.aspx?id=737580&Code=Dc476GaakpMUDp6MG3lSrtPK8uL9943N"
              alt="نماد اعتماد"
              style={{ cursor: "pointer" }}
              width={56}
              height={48}
              className="max-h-full w-auto object-contain"
              data-code="Dc476GaakpMUDp6MG3lSrtPK8uL9943N"
            />
          </a>
        </div>
      </div>
    </footer>
  );
}


