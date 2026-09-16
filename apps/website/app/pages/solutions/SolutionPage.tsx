'use client';

import type { ReactNode } from 'react';
import Link from 'next/link';
import { motion } from 'motion/react';
import {
  ArrowLeft,
  CalendarDays,
  Check,
  ChevronDown,
  CircleAlert,
  CircleCheck,
  CircleSlash,
  Network,
  Sparkles,
} from 'lucide-react';
import {
  getSolution,
  solutionHref,
  solutions,
  type AuditEvent,
  type RoutingTarget,
  type SolutionVisual,
  type UsageRow,
} from '@/app/lib/solutions';

const reveal = {
  initial: { opacity: 0, y: 20 },
  whileInView: { opacity: 1, y: 0 },
  viewport: { once: true, margin: '-60px' },
  transition: { duration: 0.5 },
};

const featureLinks: Record<string, string> = {
  '/features/ai-gateway': 'درگاه مدل‌های زبانی',
  '/features/observability': 'پایش و مشاهده‌پذیری',
  '/features/model-discovery': 'کشف و مقایسه مدل‌ها',
  '/features/guardrails': 'گاردریل‌ها',
  '/features/security': 'امنیت و تطابق',
  '/features/prompt-engineering': 'رجیستری پرامپت',
};

const faNumber = new Intl.NumberFormat('fa-IR');

function SectionHead({ eyebrow, title }: { eyebrow?: string; title: string }) {
  return (
    <div className="mx-auto mb-10 max-w-2xl text-center">
      {eyebrow ? (
        <div className="mb-3 text-sm font-bold text-indigo-600 dark:text-indigo-400">{eyebrow}</div>
      ) : null}
      <h2 className="text-2xl font-bold leading-relaxed text-slate-900 dark:text-white sm:text-3xl">{title}</h2>
    </div>
  );
}

function CtaButtons() {
  return (
    <div className="flex flex-col items-stretch justify-center gap-3 sm:flex-row sm:items-center">
      <Link
        href="/demo"
        className="flex min-h-12 items-center justify-center gap-2 rounded-xl bg-indigo-500 px-7 py-3.5 text-base font-medium text-white shadow-[0_0_20px_rgba(99,102,241,0.35)] transition-all hover:-translate-y-0.5 hover:bg-indigo-600"
      >
        رزرو وقت دمو
        <CalendarDays className="h-5 w-5" />
      </Link>
      <Link
        href="/onboarding"
        className="flex min-h-12 items-center justify-center gap-2 rounded-xl border border-slate-200 bg-white px-7 py-3.5 text-base font-medium text-slate-800 transition-colors hover:border-indigo-300 hover:text-indigo-600 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100 dark:hover:border-indigo-500/50 dark:hover:text-indigo-300"
      >
        ۱۵ روز رایگان شروع کنید
        <ArrowLeft className="h-5 w-5" />
      </Link>
    </div>
  );
}

// Monospace fonts have no Persian glyphs, so comments use the site font.
const commentFont = { fontFamily: "'Vazirmatn', sans-serif" };

// Lines starting with "#" are comments; quoted literals and trailing
// "# ..." comments are highlighted inline.
function CodeLine({ line }: { line: string }) {
  if (line.trim().startsWith('#')) {
    return <span className="text-slate-500" style={commentFont}>{line}</span>;
  }
  const commentAt = line.indexOf('  # ');
  const code = commentAt >= 0 ? line.slice(0, commentAt) : line;
  const comment = commentAt >= 0 ? line.slice(commentAt) : '';
  return (
    <>
      {code.split(/("[^"]*")/g).map((part, idx) =>
        part.startsWith('"') ? (
          <span key={idx} className="text-emerald-300">{part}</span>
        ) : (
          <span key={idx}>{part}</span>
        ),
      )}
      {comment ? <span className="text-slate-500" style={commentFont}>{comment}</span> : null}
    </>
  );
}

function CodeWindow({ filename, lines }: { filename: string; lines: string[] }) {
  return (
    <div className="mx-auto mt-12 max-w-2xl overflow-hidden rounded-2xl border border-slate-800 bg-[#0B0F19] text-left shadow-2xl dark:bg-black" dir="ltr">
      <div className="flex items-center gap-2 border-b border-white/5 px-4 py-3">
        <span className="h-3 w-3 rounded-full bg-rose-500/80" />
        <span className="h-3 w-3 rounded-full bg-amber-500/80" />
        <span className="h-3 w-3 rounded-full bg-emerald-500/80" />
        <span className="ml-auto font-mono text-xs text-slate-400">{filename}</span>
      </div>
      <pre className="overflow-x-auto p-5 font-mono text-[13px] leading-7 text-slate-200">
        {lines.map((line, idx) => (
          <div key={idx} className="min-h-7 whitespace-pre" style={{ unicodeBidi: 'plaintext', textAlign: 'left' }}>
            <CodeLine line={line} />
          </div>
        ))}
      </pre>
    </div>
  );
}

function SampleBadge() {
  return (
    <span className="rounded-full bg-slate-100 px-2.5 py-1 text-[11px] font-medium text-slate-500 dark:bg-slate-800 dark:text-slate-400">
      نمای نمونه
    </span>
  );
}

function VisualFrame({ children }: { children: ReactNode }) {
  return (
    <div className="overflow-hidden rounded-2xl border border-slate-200 bg-slate-50 shadow-xl dark:border-slate-800 dark:bg-[#121827]">
      <div className="flex items-center justify-between gap-3 border-b border-slate-200 bg-white/60 px-5 py-3 dark:border-slate-800 dark:bg-slate-900/60">
        <div className="flex min-w-0 items-center gap-2 text-sm font-bold text-slate-800 dark:text-slate-100">
          <Network className="h-4 w-4 shrink-0 text-indigo-500" />
          <span className="truncate">پنل گدارAI</span>
        </div>
        <SampleBadge />
      </div>
      <div className="p-4 sm:p-6">{children}</div>
    </div>
  );
}

function UsageVisual({ entityLabel, unit, rows }: { entityLabel: string; unit: string; rows: UsageRow[] }) {
  return (
    <div className="space-y-3">
      <div className="hidden grid-cols-[1.4fr_2fr_auto] gap-4 px-2 text-xs text-slate-500 dark:text-slate-400 sm:grid">
        <span>{entityLabel}</span>
        <span>مصرف نسبت به سقف بودجه</span>
        <span>مصرف / سقف ({unit})</span>
      </div>
      {rows.map((row) => {
        const ratio = Math.min(row.spent / row.cap, 1);
        const blocked = ratio >= 1;
        const warn = !blocked && ratio >= 0.9;
        const bar = blocked ? 'bg-rose-500' : warn ? 'bg-amber-500' : 'bg-indigo-500';
        return (
          <div
            key={row.name}
            className="grid grid-cols-1 items-center gap-3 rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-800 dark:bg-slate-900 sm:grid-cols-[1.4fr_2fr_auto] sm:gap-4"
          >
            <div className="min-w-0">
              <div className="truncate font-mono text-sm font-semibold text-slate-900 dark:text-white" dir="ltr" style={{ textAlign: 'right' }}>
                {row.name}
              </div>
              <div className="truncate text-xs text-slate-500 dark:text-slate-400">{row.detail}</div>
            </div>
            <div className="flex items-center gap-3">
              <div className="h-2.5 flex-1 overflow-hidden rounded-full bg-slate-100 dark:bg-slate-800">
                <div className={`h-full rounded-full ${bar}`} style={{ width: `${Math.round(ratio * 100)}%` }} />
              </div>
              <span
                className={`shrink-0 rounded-full px-2 py-0.5 text-[11px] font-bold ${
                  blocked
                    ? 'bg-rose-50 text-rose-600 dark:bg-rose-500/10 dark:text-rose-400'
                    : warn
                      ? 'bg-amber-50 text-amber-600 dark:bg-amber-500/10 dark:text-amber-400'
                      : 'bg-emerald-50 text-emerald-600 dark:bg-emerald-500/10 dark:text-emerald-400'
                }`}
              >
                {blocked ? 'متوقف' : warn ? 'هشدار' : 'عادی'}
              </span>
            </div>
            <div className="text-sm font-semibold text-slate-700 dark:text-slate-200">
              {faNumber.format(row.spent)} / {faNumber.format(row.cap)}
              <span className="mr-1 text-xs font-normal text-slate-400 sm:hidden">{unit}</span>
            </div>
          </div>
        );
      })}
    </div>
  );
}

const targetStyles: Record<RoutingTarget['state'], { label: string; icon: typeof CircleCheck; className: string }> = {
  down: { label: 'قطع', icon: CircleSlash, className: 'border-rose-200 bg-rose-50 text-rose-600 dark:border-rose-500/30 dark:bg-rose-500/10 dark:text-rose-400' },
  active: { label: 'پاسخ‌گو', icon: CircleCheck, className: 'border-emerald-200 bg-emerald-50 text-emerald-600 dark:border-emerald-500/30 dark:bg-emerald-500/10 dark:text-emerald-400' },
  standby: { label: 'آماده', icon: CircleAlert, className: 'border-slate-200 bg-white text-slate-500 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-400' },
};

function RoutingVisual({ virtualModel, targets }: { virtualModel: string; targets: RoutingTarget[] }) {
  return (
    <div className="grid items-center gap-4 md:grid-cols-[1fr_auto_1.4fr]">
      <div className="rounded-xl border border-slate-200 bg-white p-4 text-center dark:border-slate-800 dark:bg-slate-900">
        <div className="text-xs text-slate-500 dark:text-slate-400">درخواست محصول به مدل مجازی</div>
        <div className="mt-2 font-mono text-base font-bold text-indigo-600 dark:text-indigo-400" dir="ltr">
          {virtualModel}
        </div>
      </div>
      <ArrowLeft className="mx-auto h-6 w-6 rotate-[-90deg] text-indigo-400 md:rotate-0" />
      <ol className="space-y-2">
        {targets.map((target, idx) => {
          const style = targetStyles[target.state];
          return (
            <li key={target.model} className={`flex items-center justify-between gap-3 rounded-xl border px-4 py-3 ${style.className}`}>
              <div className="flex min-w-0 items-center gap-3">
                <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-white/70 text-xs font-bold dark:bg-black/20">
                  {faNumber.format(idx + 1)}
                </span>
                <div className="min-w-0" dir="ltr" style={{ textAlign: 'right' }}>
                  <div className="truncate font-mono text-sm font-semibold">{target.model}</div>
                  <div className="text-xs opacity-80">{target.provider}</div>
                </div>
              </div>
              <span className="flex shrink-0 items-center gap-1 text-xs font-bold">
                <style.icon className="h-4 w-4" />
                {style.label}
              </span>
            </li>
          );
        })}
      </ol>
    </div>
  );
}

const auditTone: Record<AuditEvent['tone'], string> = {
  info: 'bg-indigo-500',
  warn: 'bg-amber-500',
  block: 'bg-rose-500',
};

function AuditVisual({ events }: { events: AuditEvent[] }) {
  return (
    <ul className="divide-y divide-slate-200 overflow-hidden rounded-xl border border-slate-200 bg-white dark:divide-slate-800 dark:border-slate-800 dark:bg-slate-900">
      {events.map((event) => (
        <li key={event.time} className="flex flex-col gap-1 px-4 py-3 sm:flex-row sm:items-center sm:gap-4">
          <span className="flex items-center gap-2 text-xs text-slate-500 dark:text-slate-400 sm:w-24 sm:shrink-0">
            <span className={`h-2 w-2 rounded-full ${auditTone[event.tone]}`} />
            {event.time}
          </span>
          <span className="font-mono text-xs text-indigo-600 dark:text-indigo-400 sm:w-52 sm:shrink-0">{event.actor}</span>
          <span className="text-sm text-slate-800 dark:text-slate-200">{event.action}</span>
        </li>
      ))}
    </ul>
  );
}

function SolutionVisualBlock({ visual }: { visual: SolutionVisual }) {
  return (
    <VisualFrame>
      {visual.kind === 'usage' ? (
        <UsageVisual entityLabel={visual.entityLabel} unit={visual.unit} rows={visual.rows} />
      ) : visual.kind === 'routing' ? (
        <RoutingVisual virtualModel={visual.virtualModel} targets={visual.targets} />
      ) : (
        <AuditVisual events={visual.events} />
      )}
    </VisualFrame>
  );
}

export function SolutionPage({ slug }: { slug: string }) {
  const solution = getSolution(slug);
  if (!solution) return null;
  const others = solutions.filter((item) => item.slug !== slug);

  return (
    <main className="overflow-hidden">
      {/* Hero */}
      <section className="relative px-4 pb-16 pt-14 sm:px-6 sm:pt-20 lg:px-8">
        <div className="pointer-events-none absolute inset-x-0 top-0 -z-0 h-[420px] bg-gradient-to-b from-indigo-500/10 to-transparent" />
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
          className="relative mx-auto max-w-4xl text-center"
        >
          <span className="mb-6 inline-flex items-center gap-2 rounded-full border border-indigo-200 bg-indigo-50 px-4 py-1.5 text-sm font-semibold text-indigo-700 dark:border-indigo-500/30 dark:bg-indigo-500/10 dark:text-indigo-300">
            <solution.icon className="h-4 w-4 shrink-0" />
            {solution.badge}
          </span>
          <h1 className="mx-auto max-w-3xl text-3xl font-extrabold leading-snug text-slate-900 dark:text-white sm:text-5xl sm:leading-tight">
            {solution.title}
          </h1>
          <p className="mt-4 bg-gradient-to-l from-indigo-500 to-purple-500 bg-clip-text text-xl font-bold leading-relaxed text-transparent sm:text-2xl">
            {solution.highlight}
          </p>
          <p className="mx-auto mt-6 max-w-2xl text-base leading-8 text-slate-600 dark:text-slate-300 sm:text-lg">
            {solution.intro}
          </p>
          <div className="mt-9">
            <CtaButtons />
          </div>
          <CodeWindow filename={solution.code.filename} lines={solution.code.lines} />
        </motion.div>
      </section>

      {/* Self-qualify */}
      <section className="bg-slate-50 px-4 py-16 dark:bg-slate-900/40 sm:px-6 sm:py-20 lg:px-8">
        <motion.div {...reveal} className="mx-auto max-w-3xl">
          <SectionHead eyebrow="مخاطب این راهکار" title="گدارAI مناسب شماست، اگر:" />
          <ul className="divide-y divide-slate-200 rounded-2xl border border-slate-200 bg-white px-5 dark:divide-slate-800 dark:border-slate-800 dark:bg-[#0B0F19] sm:px-7">
            {solution.qualifiers.map((item) => (
              <li key={item} className="flex items-start gap-3 py-4">
                <span className="mt-1 flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-emerald-50 text-emerald-600 dark:bg-emerald-500/10 dark:text-emerald-400">
                  <Check className="h-4 w-4" />
                </span>
                <p className="text-base leading-8 text-slate-800 dark:text-slate-200">{item}</p>
              </li>
            ))}
          </ul>
        </motion.div>
      </section>

      {/* Mechanism */}
      <section className="px-4 py-16 sm:px-6 sm:py-20 lg:px-8">
        <div className="mx-auto max-w-5xl">
          <SectionHead eyebrow="مکانیزم" title="چطور کار می‌کند" />
          <div className="grid gap-5 sm:grid-cols-2">
            {solution.features.map((feature, idx) => (
              <motion.div
                key={feature.title}
                {...reveal}
                transition={{ duration: 0.5, delay: idx * 0.08 }}
                className="rounded-2xl border border-slate-200 bg-white p-6 transition-colors hover:border-indigo-300 dark:border-slate-800 dark:bg-slate-900/60 dark:hover:border-indigo-500/40"
              >
                <div className="mb-4 flex h-11 w-11 items-center justify-center rounded-xl bg-indigo-50 text-indigo-600 dark:bg-indigo-500/10 dark:text-indigo-400">
                  <feature.icon className="h-5 w-5" />
                </div>
                <h3 className="mb-2 text-lg font-bold text-slate-900 dark:text-white">{feature.title}</h3>
                <p className="text-sm leading-7 text-slate-600 dark:text-slate-400">{feature.description}</p>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      {/* In practice */}
      <section className="bg-slate-50 px-4 py-16 dark:bg-slate-900/40 sm:px-6 sm:py-20 lg:px-8">
        <motion.div {...reveal} className="mx-auto max-w-4xl">
          <SectionHead eyebrow="در عمل" title={solution.visual.title} />
          <SolutionVisualBlock visual={solution.visual} />
        </motion.div>
      </section>

      {/* Pitch */}
      <section className="px-4 py-16 sm:px-6 sm:py-20 lg:px-8">
        <motion.div
          {...reveal}
          className="relative mx-auto max-w-4xl overflow-hidden rounded-3xl bg-gradient-to-br from-indigo-600 to-purple-600 p-8 text-white shadow-[0_20px_60px_-20px_rgba(99,102,241,0.6)] sm:p-12"
        >
          <Sparkles className="absolute -left-4 -top-4 h-28 w-28 text-white/10" />
          <div className="relative mb-4 text-sm font-bold text-white/80">{solution.pitch.kicker}</div>
          <p className="relative text-lg font-medium leading-9 sm:text-xl sm:leading-10">{solution.pitch.body}</p>
        </motion.div>
      </section>

      {/* FAQ */}
      <section className="bg-slate-50 px-4 py-16 dark:bg-slate-900/40 sm:px-6 sm:py-20 lg:px-8">
        <div className="mx-auto max-w-3xl">
          <SectionHead eyebrow="سؤالات رایج" title={solution.faqTitle} />
          <div className="space-y-3">
            {solution.faqs.map((faq) => (
              <details
                key={faq.question}
                className="group rounded-xl border border-slate-200 bg-white px-5 py-4 dark:border-slate-800 dark:bg-slate-900 [&_summary::-webkit-details-marker]:hidden"
              >
                <summary className="flex cursor-pointer list-none items-center justify-between gap-4 font-bold text-slate-900 dark:text-white">
                  {faq.question}
                  <ChevronDown className="h-5 w-5 shrink-0 text-slate-500 transition-transform group-open:rotate-180" />
                </summary>
                <p className="mt-3 text-sm leading-7 text-slate-600 dark:text-slate-400 sm:text-base sm:leading-8">
                  {faq.answer}
                  {faq.link ? (
                    <>
                      {' '}
                      <Link href={faq.link.href} className="font-semibold text-indigo-600 hover:underline dark:text-indigo-400">
                        {faq.link.label}
                      </Link>
                    </>
                  ) : null}
                </p>
              </details>
            ))}
          </div>
        </div>
      </section>

      {/* Related features + catalog */}
      <section className="px-4 py-14 sm:px-6 lg:px-8">
        <div className="mx-auto flex max-w-4xl flex-col items-center gap-5 text-center">
          <p className="text-sm text-slate-500 dark:text-slate-400">
            قابلیت‌های مرتبط با این راهکار را دقیق‌تر ببینید، یا فهرست به‌روز مدل‌ها و ارائه‌دهندگان را در{' '}
            <Link href="/catalog" className="font-semibold text-indigo-600 hover:underline dark:text-indigo-400">
              کاتالوگ مدل‌ها
            </Link>{' '}
            بررسی کنید.
          </p>
          <div className="flex flex-wrap justify-center gap-2">
            {solution.related.map((href) => (
              <Link
                key={href}
                href={href}
                className="inline-flex min-h-10 items-center gap-1.5 rounded-full border border-slate-200 px-4 py-2 text-sm font-medium text-slate-700 transition-colors hover:border-indigo-300 hover:text-indigo-600 dark:border-slate-700 dark:text-slate-300 dark:hover:border-indigo-500/50 dark:hover:text-indigo-300"
              >
                {featureLinks[href] ?? href}
                <ArrowLeft className="h-4 w-4" />
              </Link>
            ))}
          </div>
        </div>
      </section>

      {/* Final CTA */}
      <section className="px-4 pb-16 sm:px-6 sm:pb-24 lg:px-8">
        <motion.div
          {...reveal}
          className="mx-auto max-w-4xl rounded-3xl border border-slate-200 bg-white px-6 py-12 text-center dark:border-slate-800 dark:bg-slate-900/60"
        >
          <h2 className="mb-8 text-2xl font-bold text-slate-900 dark:text-white sm:text-3xl">{solution.finalCta}</h2>
          <CtaButtons />
        </motion.div>
      </section>

      {/* Other solutions */}
      <section className="border-t border-slate-200 px-4 py-14 dark:border-slate-800 sm:px-6 lg:px-8">
        <div className="mx-auto max-w-6xl">
          <h2 className="mb-6 text-center text-lg font-bold text-slate-900 dark:text-white">راهکارهای دیگر گدارAI</h2>
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-5">
            {others.map((item) => (
              <Link
                key={item.slug}
                href={solutionHref(item.slug)}
                className="group flex items-center gap-3 rounded-xl border border-slate-200 p-4 transition-colors hover:border-indigo-300 hover:bg-indigo-50/40 dark:border-slate-800 dark:hover:border-indigo-500/40 dark:hover:bg-indigo-500/5"
              >
                <item.icon className="h-5 w-5 shrink-0 text-indigo-500" />
                <span className="text-sm font-semibold text-slate-800 group-hover:text-indigo-600 dark:text-slate-200 dark:group-hover:text-indigo-300">
                  {item.navName}
                </span>
              </Link>
            ))}
          </div>
        </div>
      </section>
    </main>
  );
}
