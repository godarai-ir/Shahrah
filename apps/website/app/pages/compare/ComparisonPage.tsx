'use client';

import type { ComponentType } from 'react';
import { motion } from 'motion/react';
import Link from 'next/link';
import { FAQAccordion } from '../../components/ui/FAQAccordion';
import { AlertTriangle, ArrowLeft, BarChart3, CheckCircle, Code2, Globe, Scale, Zap } from 'lucide-react';

type Icon = ComponentType<{ className?: string }>;

export type ComparisonContent = {
  competitor: string;
  hero: {
    titleLead: string;
    titleHighlight: string;
    titleTail: string;
    intro: string;
    tagline: string;
    chips: string[];
  };
  differentiators: { subtitle: string; rows: Array<{ icon: Icon; category: string; godar: string; competitor: string }> };
  evaluation: { subtitle: string; items: Array<{ question: string; godar: string; competitor: string }> };
  sideBySide: { title: string; body: string; ctaLabel: string };
  painPoints: { subtitle: string; rows: Array<{ pain: string; benefit: string; impact: string }> };
  pitfalls: { subtitle: string; items: string[] };
  scenarios: { title: string; subtitle: string; items: Array<{ icon: Icon; title: string; desc: string }> };
  faq: Array<{ question: string; answer: string }>;
  finalCta: { title: string; body: string; primaryLabel: string; badges: string[] };
};

const fadeIn = {
  initial: { opacity: 0, y: 20 },
  whileInView: { opacity: 1, y: 0 },
  viewport: { once: true },
  transition: { duration: 0.4 },
};

// ─── Section heading ─────────────────────────────────────────────────────────
function SectionHeading({ title, subtitle }: { title: string; subtitle?: string }) {
  return (
    <div className="text-center mb-14 max-w-3xl mx-auto">
      <h2 className="text-3xl md:text-4xl font-bold mb-4 text-slate-900 dark:text-white" style={{ fontFamily: 'var(--font-heading)' }}>
        {title}
      </h2>
      {subtitle && <p className="text-lg leading-relaxed text-slate-600 dark:text-slate-400">{subtitle}</p>}
    </div>
  );
}

// ─── Main Component ──────────────────────────────────────────────────────────
export function ComparisonPage({ content }: { content: ComparisonContent }) {
  const { competitor, hero, differentiators, evaluation, sideBySide, painPoints, pitfalls, scenarios, faq, finalCta } = content;

  return (
    <main className="flex min-h-screen flex-col bg-white dark:bg-[#0B0F19] text-slate-900 dark:text-white" dir="rtl">

      {/* ═══ 1. Hero ══════════════════════════════════════════════════════════ */}
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
              <Scale className="w-4 h-4" />
              مقایسه گدارAI و {competitor}
            </motion.div>

            <motion.h1
              className="text-4xl md:text-5xl lg:text-6xl font-bold tracking-tight mb-8 leading-tight text-slate-900 dark:text-white"
              style={{ fontFamily: 'var(--font-heading)' }}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5 }}
            >
              {hero.titleLead}{' '}
              <span className="text-transparent bg-clip-text bg-gradient-to-l from-indigo-500 to-purple-500">{hero.titleHighlight}</span>{' '}
              {hero.titleTail}
            </motion.h1>

            <motion.p
              className="text-lg md:text-xl text-slate-600 dark:text-slate-400 mb-6 leading-relaxed max-w-3xl mx-auto"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.1 }}
            >
              {hero.intro}
            </motion.p>

            <motion.p
              className="text-xl md:text-2xl font-bold text-indigo-600 dark:text-indigo-400 mb-10"
              style={{ fontFamily: 'var(--font-heading)' }}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.15 }}
            >
              {hero.tagline}
            </motion.p>

            <motion.div
              className="flex flex-col sm:flex-row items-center justify-center gap-4"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.2 }}
            >
              <Link
                href="/onboarding"
                className="w-full sm:w-auto px-8 py-4 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl font-bold text-lg shadow-[0_0_20px_rgba(79,70,229,0.3)] hover:shadow-[0_0_30px_rgba(79,70,229,0.5)] transition-all active:scale-95 flex items-center justify-center gap-2 group"
              >
                ۱۵ روز رایگان شروع کنید
                <ArrowLeft className="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
              </Link>
              <Link
                href="/demo"
                className="w-full sm:w-auto px-8 py-4 bg-white dark:bg-slate-900 text-slate-900 dark:text-white border-2 border-slate-200 dark:border-slate-800 hover:border-indigo-600 dark:hover:border-indigo-500 rounded-xl font-bold text-lg transition-all active:scale-95 flex items-center justify-center"
              >
                گفت‌وگو با کارشناس
              </Link>
            </motion.div>

            <div className="mt-10 flex flex-wrap justify-center gap-3">
              {hero.chips.map((chip) => (
                <span key={chip} className="flex items-center gap-1.5 px-4 py-2 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-full text-sm font-medium text-slate-700 dark:text-slate-300 shadow-sm">
                  <CheckCircle className="w-4 h-4 text-emerald-500" />
                  {chip}
                </span>
              ))}
            </div>
          </div>
        </div>
      </section>

      {/* ═══ 2. Key differentiators ══════════════════════════════════════════ */}
      <section className="py-24 bg-slate-50 dark:bg-slate-900/30 border-y border-slate-100 dark:border-slate-800/50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <SectionHeading title="تفاوت‌های کلیدی" subtitle={differentiators.subtitle} />

          <div className="hidden md:grid grid-cols-[1fr_200px_1fr] gap-4 px-6 mb-3 text-sm font-bold" style={{ fontFamily: 'var(--font-heading)' }}>
            <div className="text-indigo-600 dark:text-indigo-400">گدارAI</div>
            <div className="text-center text-slate-500">حوزه</div>
            <div className="text-slate-500">{competitor}</div>
          </div>

          <div className="space-y-3">
            {differentiators.rows.map((row, i) => (
              <motion.div
                key={row.category}
                {...fadeIn}
                transition={{ duration: 0.4, delay: (i % 4) * 0.05 }}
                className="grid grid-cols-1 md:grid-cols-[1fr_200px_1fr] gap-4 bg-white dark:bg-slate-900/60 border border-slate-200 dark:border-slate-800 rounded-2xl p-5 md:p-6 md:items-center"
              >
                <div className="md:order-2 flex md:flex-col items-center gap-3 md:gap-2 md:text-center">
                  <div className="w-10 h-10 bg-indigo-50 dark:bg-indigo-500/10 text-indigo-600 dark:text-indigo-400 rounded-xl flex items-center justify-center shrink-0">
                    <row.icon className="w-5 h-5" />
                  </div>
                  <h3 className="font-bold text-slate-900 dark:text-white" style={{ fontFamily: 'var(--font-heading)' }}>{row.category}</h3>
                </div>
                <div className="md:order-1 rounded-xl bg-indigo-50/60 dark:bg-indigo-500/5 border border-indigo-100 dark:border-indigo-500/20 p-4">
                  <div className="md:hidden text-xs font-bold text-indigo-600 dark:text-indigo-400 mb-1">گدارAI</div>
                  <p className="text-sm text-slate-700 dark:text-slate-300 leading-relaxed">{row.godar}</p>
                </div>
                <div className="md:order-3 rounded-xl bg-slate-50 dark:bg-slate-800/40 border border-slate-100 dark:border-slate-700/50 p-4">
                  <div className="md:hidden text-xs font-bold text-slate-500 mb-1">{competitor}</div>
                  <p className="text-sm text-slate-600 dark:text-slate-400 leading-relaxed">{row.competitor}</p>
                </div>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      {/* ═══ 3. Evaluation questions ═════════════════════════════════════════ */}
      <section className="py-24">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
          <SectionHeading title="پرسش‌های کلیدی برای ارزیابی" subtitle={evaluation.subtitle} />

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {evaluation.items.map((item, i) => (
              <motion.div
                key={item.question}
                {...fadeIn}
                transition={{ duration: 0.4, delay: (i % 2) * 0.1 }}
                className="bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-3xl p-6 md:p-8"
              >
                <div className="flex items-start gap-3 mb-5">
                  <span className="w-8 h-8 rounded-lg bg-indigo-600 text-white flex items-center justify-center text-sm font-bold shrink-0">
                    {(i + 1).toLocaleString('fa-IR')}
                  </span>
                  <h3 className="text-lg font-bold text-slate-900 dark:text-white leading-relaxed" style={{ fontFamily: 'var(--font-heading)' }}>
                    {item.question}
                  </h3>
                </div>
                <div className="space-y-3">
                  <div className="rounded-xl bg-white dark:bg-slate-900 border border-indigo-100 dark:border-indigo-500/20 p-4">
                    <div className="text-xs font-bold text-indigo-600 dark:text-indigo-400 mb-1">گدارAI</div>
                    <p className="text-sm text-slate-700 dark:text-slate-300 leading-relaxed">{item.godar}</p>
                  </div>
                  <div className="rounded-xl bg-white/60 dark:bg-slate-900/40 border border-slate-200 dark:border-slate-800 p-4">
                    <div className="text-xs font-bold text-slate-500 mb-1">{competitor}</div>
                    <p className="text-sm text-slate-600 dark:text-slate-400 leading-relaxed">{item.competitor}</p>
                  </div>
                </div>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      {/* ═══ 4. Try before you migrate ═══════════════════════════════════════ */}
      <section className="py-24 bg-slate-900 dark:bg-[#060913] text-white border-y border-slate-800">
        <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <span className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-white/10 border border-white/15 text-indigo-300 text-sm font-bold mb-6">
            <Code2 className="w-4 h-4" />
            پیش از مهاجرت آزمایش کنید
          </span>
          <h2 className="text-3xl md:text-4xl font-bold mb-5" style={{ fontFamily: 'var(--font-heading)' }}>
            {sideBySide.title}
          </h2>
          <p className="text-lg text-slate-400 leading-relaxed max-w-3xl mx-auto mb-10">{sideBySide.body}</p>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-10 text-right">
            {[
              { icon: Zap, title: '۱۵ روز آزمایش رایگان', desc: 'برای پلن‌های غیرسازمانی' },
              { icon: Globe, title: 'بیش از ۱۲۰۰ مدل', desc: 'از یک نشانی پایه و یک توکن' },
              { icon: BarChart3, title: 'مقایسه با داده واقعی', desc: 'لاگ، رد درخواست و گزارش هزینه' },
            ].map((item) => (
              <div key={item.title} className="flex items-center gap-3 bg-slate-800/50 border border-slate-700/50 rounded-2xl p-5">
                <div className="w-10 h-10 bg-indigo-500/20 text-indigo-300 rounded-xl flex items-center justify-center shrink-0">
                  <item.icon className="w-5 h-5" />
                </div>
                <div>
                  <div className="font-bold text-white">{item.title}</div>
                  <div className="text-sm text-slate-400">{item.desc}</div>
                </div>
              </div>
            ))}
          </div>

          <Link
            href="/onboarding"
            className="inline-flex w-full sm:w-auto px-8 py-4 bg-indigo-600 hover:bg-indigo-500 text-white rounded-xl font-bold text-lg transition-all active:scale-95 items-center justify-center gap-2 group"
          >
            {sideBySide.ctaLabel}
            <ArrowLeft className="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
          </Link>
        </div>
      </section>

      {/* ═══ 5. Pain points ══════════════════════════════════════════════════ */}
      <section className="py-24">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <SectionHeading title="گدارAI کدام مشکل‌ها را حل می‌کند؟" subtitle={painPoints.subtitle} />

          <div className="hidden lg:grid grid-cols-3 gap-4 px-6 mb-3 text-sm font-bold text-slate-500" style={{ fontFamily: 'var(--font-heading)' }}>
            <div>چالش</div>
            <div>راهکار گدارAI</div>
            <div>اثر برای تیم</div>
          </div>

          <div className="space-y-3">
            {painPoints.rows.map((row, i) => (
              <motion.div
                key={row.pain}
                {...fadeIn}
                transition={{ duration: 0.4, delay: (i % 3) * 0.05 }}
                className="grid grid-cols-1 lg:grid-cols-3 gap-3 lg:gap-4 bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-2xl p-5 lg:p-6 lg:items-center"
              >
                <div className="flex items-center gap-2">
                  <AlertTriangle className="w-5 h-5 text-amber-500 shrink-0" />
                  <h3 className="font-bold text-slate-900 dark:text-white" style={{ fontFamily: 'var(--font-heading)' }}>{row.pain}</h3>
                </div>
                <div className="flex items-start gap-2">
                  <CheckCircle className="w-4 h-4 text-emerald-500 shrink-0 mt-1" />
                  <p className="text-sm text-slate-700 dark:text-slate-300 leading-relaxed">{row.benefit}</p>
                </div>
                <p className="text-sm text-slate-600 dark:text-slate-400 leading-relaxed lg:border-r lg:border-slate-200 lg:dark:border-slate-700 lg:pr-4">
                  {row.impact}
                </p>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      {/* ═══ 6. Pitfalls ═════════════════════════════════════════════════════ */}
      <section className="py-24 bg-slate-50 dark:bg-slate-900/30 border-y border-slate-100 dark:border-slate-800/50">
        <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8">
          <SectionHeading title="اشتباه‌های رایج هنگام انتخاب درگاه" subtitle={pitfalls.subtitle} />
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {pitfalls.items.map((item, i) => (
              <motion.div
                key={item}
                {...fadeIn}
                transition={{ duration: 0.4, delay: (i % 2) * 0.08 }}
                className="flex items-start gap-3 bg-white dark:bg-slate-900/60 border border-slate-200 dark:border-slate-800 rounded-2xl p-5"
              >
                <div className="w-8 h-8 rounded-lg bg-rose-50 dark:bg-rose-500/10 text-rose-600 dark:text-rose-400 flex items-center justify-center shrink-0">
                  <AlertTriangle className="w-4 h-4" />
                </div>
                <p className="text-slate-700 dark:text-slate-300 leading-relaxed">{item}</p>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      {/* ═══ 7. Migration scenarios ══════════════════════════════════════════ */}
      <section className="py-24">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <SectionHeading title={scenarios.title} subtitle={scenarios.subtitle} />
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5">
            {scenarios.items.map((item, i) => (
              <motion.div
                key={item.title}
                {...fadeIn}
                transition={{ duration: 0.4, delay: i * 0.08 }}
                className="bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 hover:border-indigo-200 dark:hover:border-indigo-800 transition-colors group"
              >
                <div className="w-10 h-10 bg-indigo-50 dark:bg-indigo-500/10 text-indigo-600 dark:text-indigo-400 rounded-xl flex items-center justify-center mb-4 group-hover:scale-110 transition-transform">
                  <item.icon className="w-5 h-5" />
                </div>
                <h3 className="font-bold text-slate-900 dark:text-white mb-2" style={{ fontFamily: 'var(--font-heading)' }}>{item.title}</h3>
                <p className="text-sm text-slate-600 dark:text-slate-400 leading-relaxed">{item.desc}</p>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      {/* ═══ 8. FAQ ══════════════════════════════════════════════════════════ */}
      <section className="py-24 bg-slate-50 dark:bg-slate-900/30 border-y border-slate-100 dark:border-slate-800/50">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
          <SectionHeading title="پرسش‌ها و تردیدهای رایج" />
          <FAQAccordion items={faq} />
          <p className="mt-8 text-center text-xs text-slate-500 dark:text-slate-500 leading-relaxed">
            اطلاعات {competitor} بر اساس مستندات و اطلاعیه‌های عمومی آن در زمان انتشار این صفحه است و ممکن است تغییر کرده باشد. پیش از تصمیم نهایی، نسخه فعلی هر دو محصول را در محیط خود ارزیابی کنید.
          </p>
        </div>
      </section>

      {/* ═══ 9. Final CTA ════════════════════════════════════════════════════ */}
      <section className="py-24 relative overflow-hidden">
        <div className="absolute inset-0 bg-indigo-600 dark:bg-indigo-900/40" />
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10 text-center">
          <h2 className="text-3xl md:text-4xl font-bold text-white mb-4" style={{ fontFamily: 'var(--font-heading)' }}>
            {finalCta.title}
          </h2>
          <p className="text-lg text-indigo-200 mb-10 leading-relaxed">{finalCta.body}</p>
          <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
            <Link
              href="/onboarding"
              className="w-full sm:w-auto px-8 py-4 bg-white text-indigo-700 hover:bg-indigo-50 rounded-xl font-bold text-lg shadow-[0_0_30px_rgba(0,0,0,0.2)] transition-all active:scale-95 flex items-center justify-center gap-2 group"
            >
              {finalCta.primaryLabel}
              <ArrowLeft className="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
            </Link>
            <Link
              href="/demo"
              className="w-full sm:w-auto px-8 py-4 bg-white/10 hover:bg-white/20 text-white border border-white/20 hover:border-white/40 rounded-xl font-bold text-lg transition-all active:scale-95 flex items-center justify-center"
            >
              گفت‌وگو با کارشناس
            </Link>
          </div>
          <div className="mt-8 flex flex-wrap justify-center gap-x-6 gap-y-2 text-sm text-indigo-100">
            {finalCta.badges.map((item) => (
              <span key={item} className="flex items-center gap-1.5">
                <CheckCircle className="w-4 h-4" />
                {item}
              </span>
            ))}
          </div>
        </div>
      </section>
    </main>
  );
}
