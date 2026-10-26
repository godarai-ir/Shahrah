import { ArrowLeft, Check, CircleAlert, CircleCheck, CircleSlash } from 'lucide-react';
import { ActionLink } from '@/app/components/patterns/PageHero';
import {
  ClosingBand,
  CodeWindow,
  FAQ,
  PageHero,
  ProductStage,
  ProductWindow,
  Section,
  TileGrid,
} from '@/app/components/patterns';
import { sectionCopy, sectionHeading, textLink } from '@/app/components/sectionType';
import {
  getSolution,
  solutionHref,
  solutions,
  type AuditEvent,
  type RoutingTarget,
  type SolutionVisual,
  type UsageRow,
} from '@/app/lib/solutions';

const featureLinks: Record<string, string> = {
  '/features/ai-gateway': 'درگاه مدل‌های زبانی',
  '/features/observability': 'مشاهده‌پذیری',
  '/features/model-discovery': 'کشف و مقایسه مدل‌ها',
  '/features/guardrails': 'گاردریل‌های معنایی',
  '/features/security': 'امنیت و تطابق',
  '/features/prompt-engineering': 'رجیستری پرامپت',
};

const faNumber = new Intl.NumberFormat('fa-IR');

const statusTone = {
  success: 'border-site-success text-site-success',
  danger: 'border-site-danger text-site-danger',
  neutral: 'border-site-line-strong text-site-ink-soft',
} as const;

function Status({ label, tone }: { label: string; tone: keyof typeof statusTone }) {
  return <span className={`shrink-0 rounded-full border px-2 py-0.5 text-caption font-medium ${statusTone[tone]}`}>{label}</span>;
}

function UsageVisual({ entityLabel, unit, rows }: { entityLabel: string; unit: string; rows: UsageRow[] }) {
  return (
    <ul className="m-0 flex list-none flex-col divide-y divide-site-line p-0" aria-label={`${entityLabel}، مصرف نسبت به سقف (${unit})`}>
      {rows.map((row) => {
        const ratio = Math.min(row.spent / row.cap, 1);
        const blocked = ratio >= 1;
        const warn = !blocked && ratio >= 0.9;

        return (
          <li key={row.name} className="flex flex-col gap-2 py-3">
            <div className="flex items-center justify-between gap-3">
              <div className="min-w-0">
                <div dir="ltr" className="truncate text-end text-label font-semibold text-site-ink">{row.name}</div>
                <div className="truncate text-caption text-site-ink-soft">{row.detail}</div>
              </div>
              <Status label={blocked ? 'متوقف' : warn ? 'هشدار' : 'عادی'} tone={blocked ? 'danger' : warn ? 'neutral' : 'success'} />
            </div>
            <div className="flex items-center gap-3">
              <div className="h-2 flex-1 overflow-hidden rounded-full bg-site-panel">
                <div className={`h-full rounded-full ${blocked ? 'bg-site-danger' : 'bg-site-ink'}`} style={{ width: `${Math.round(ratio * 100)}%` }} />
              </div>
              <span className="shrink-0 text-caption text-site-ink-soft">
                {faNumber.format(row.spent)} / {faNumber.format(row.cap)} {unit}
              </span>
            </div>
          </li>
        );
      })}
    </ul>
  );
}

const targetState: Record<RoutingTarget['state'], { label: string; icon: typeof CircleCheck; tone: keyof typeof statusTone }> = {
  down: { label: 'قطع', icon: CircleSlash, tone: 'danger' },
  active: { label: 'پاسخ‌گو', icon: CircleCheck, tone: 'success' },
  standby: { label: 'آماده', icon: CircleAlert, tone: 'neutral' },
};

function RoutingVisual({ virtualModel, targets }: { virtualModel: string; targets: RoutingTarget[] }) {
  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between gap-3 rounded-field bg-site-panel px-4 py-3">
        <span className="text-caption text-site-ink-soft">مدل مجازی</span>
        <span dir="ltr" className="text-label font-semibold text-site-ink">{virtualModel}</span>
      </div>
      <ol className="m-0 flex list-none flex-col divide-y divide-site-line p-0">
        {targets.map((target, index) => {
          const state = targetState[target.state];

          return (
            <li key={target.model} className="flex items-center justify-between gap-3 py-3">
              <div className="flex min-w-0 items-center gap-3">
                <span className="flex size-6 shrink-0 items-center justify-center rounded-full bg-site-sand text-caption font-semibold text-site-ink">
                  {faNumber.format(index + 1)}
                </span>
                <div dir="ltr" className="min-w-0 text-end">
                  <div className="truncate text-label font-semibold text-site-ink">{target.model}</div>
                  <div className="text-caption text-site-ink-soft">{target.provider}</div>
                </div>
              </div>
              <Status label={state.label} tone={state.tone} />
            </li>
          );
        })}
      </ol>
    </div>
  );
}

const auditTone: Record<AuditEvent['tone'], keyof typeof statusTone> = { info: 'neutral', warn: 'neutral', block: 'danger' };

function AuditVisual({ events }: { events: AuditEvent[] }) {
  return (
    <ul className="m-0 flex list-none flex-col divide-y divide-site-line p-0">
      {events.map((event) => (
        <li key={event.time} className="flex flex-col gap-1 py-3">
          <div className="flex items-center justify-between gap-3">
            <span dir="ltr" className="truncate text-caption font-medium text-site-ink">{event.actor}</span>
            <span className="shrink-0 text-caption text-site-ink-soft">{event.time}</span>
          </div>
          <div className="flex items-center justify-between gap-3">
            <span className="text-label text-site-ink">{event.action}</span>
            {event.tone === 'block' ? <Status label="رد شد" tone={auditTone[event.tone]} /> : null}
          </div>
        </li>
      ))}
    </ul>
  );
}

function SolutionVisualBlock({ visual }: { visual: SolutionVisual }) {
  if (visual.kind === 'usage') {
    return <UsageVisual entityLabel={visual.entityLabel} unit={visual.unit} rows={visual.rows} />;
  }

  if (visual.kind === 'routing') {
    return <RoutingVisual virtualModel={visual.virtualModel} targets={visual.targets} />;
  }

  return <AuditVisual events={visual.events} />;
}

// One template for the six /solutions/[slug] pages; all copy comes from lib/solutions.ts.
export function SolutionPage({ slug }: { slug: string }) {
  const solution = getSolution(slug);

  if (!solution) {
    return null;
  }

  const others = solutions.filter((item) => item.slug !== slug);

  return (
    <main className="bg-site-ground">
      <PageHero
        title={solution.title}
        lead={solution.highlight}
        primary={{ href: '/onboarding', label: 'شروع با اعتبار هدیه' }}
        secondary={{ href: '/demo', label: 'رزرو دمو' }}
        visual={
          <ProductStage layout="flex size-full items-center justify-center p-6 sm:p-10">
            <CodeWindow title={solution.badge} meta={solution.code.filename} code={solution.code.lines.join('\n')} />
          </ProductStage>
        }
      />

      <Section id="fit" tone="panel" layout="grid gap-8 lg:grid-cols-2 lg:gap-16">
        <div className="flex flex-col gap-4">
          <h2 className={sectionHeading}>گدارAI به کار شما می‌آید، اگر…</h2>
          <p className={`m-0 ${sectionCopy}`}>{solution.intro}</p>
        </div>
        <ul className="m-0 flex list-none flex-col divide-y divide-site-line border-y border-site-line p-0">
          {solution.qualifiers.map((item) => (
            <li key={item} className="flex items-start gap-3 py-4">
              <Check aria-hidden="true" className="mt-1 size-5 shrink-0 text-site-ink" />
              <p className="m-0 text-body text-site-ink">{item}</p>
            </li>
          ))}
        </ul>
      </Section>

      <Section id="how" layout="flex flex-col gap-8">
        <h2 className={`max-w-3xl ${sectionHeading}`}>چطور کار می‌کند.</h2>
        <TileGrid
          columns={2}
          items={solution.features.map((feature) => ({ key: feature.title, icon: feature.icon, title: feature.title, text: feature.description }))}
        />
      </Section>

      <Section id="practice" tone="panel" layout="flex flex-col gap-8">
        <h2 className={`max-w-3xl ${sectionHeading}`}>{solution.visual.title}</h2>
        <ProductStage layout="flex w-full justify-center p-6 sm:p-12">
          <ProductWindow title="پنل گدارAI" meta="نمونه" size="w-full max-w-2xl" label={`نمونه: ${solution.visual.title}. عددها ساختگی‌اند.`}>
            <SolutionVisualBlock visual={solution.visual} />
          </ProductWindow>
        </ProductStage>
      </Section>

      <Section id="pitch" width="narrow" layout="flex flex-col gap-4">
        <p className="m-0 text-label font-medium text-site-accent-text">{solution.pitch.kicker}</p>
        <p className="m-0 text-balance text-h3 font-semibold text-site-ink">{solution.pitch.body}</p>
      </Section>

      <FAQ title={solution.faqTitle} items={solution.faqs} />

      <Section id="related" layout="flex flex-col gap-4">
        <h2 className="m-0 text-h3 font-semibold text-site-ink">جزئیات بیشتر.</h2>
        <ul className="m-0 flex list-none flex-wrap gap-x-8 gap-y-3 p-0">
          {[...solution.related, '/catalog'].map((href) => (
            <li key={href}>
              <ActionLink href={href} className={textLink}>
                {featureLinks[href] ?? 'کاتالوگ مدل‌ها'}
                <ArrowLeft aria-hidden="true" className="size-4" />
              </ActionLink>
            </li>
          ))}
        </ul>
      </Section>

      <ClosingBand
        title={solution.finalCta}
        copy="با اعتبار هدیه وارد پنل شوید؛ تاریخ انقضا ندارد."
        primary={{ href: '/onboarding', label: 'شروع با اعتبار هدیه' }}
        secondary={{ href: '/demo', label: 'رزرو دمو' }}
      />

      <Section id="others" tone="panel" layout="flex flex-col gap-8">
        <h2 className={`max-w-3xl ${sectionHeading}`}>راهکارهای دیگر گدارAI.</h2>
        <TileGrid
          columns={3}
          size="compact"
          items={others.map((item) => ({ key: item.slug, icon: item.icon, title: item.navName, text: item.navDescription, href: solutionHref(item.slug) }))}
        />
      </Section>
    </main>
  );
}
