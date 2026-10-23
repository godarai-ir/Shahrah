import type { LucideIcon } from 'lucide-react';
import { ArrowLeft, Check } from 'lucide-react';
import Link from 'next/link';
import {
  ClosingBand,
  CodeWindow,
  ComparisonTable,
  FAQ,
  FeatureRow,
  MockRows,
  PageHero,
  ProductStage,
  ProductWindow,
  Section,
  TileGrid,
} from '@/app/components/patterns';
import { sectionCopy, sectionHeading, textLink } from '@/app/components/sectionType';

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
  differentiators: { subtitle: string; rows: Array<{ icon: LucideIcon; category: string; godar: string; competitor: string }> };
  evaluation: { subtitle: string; items: Array<{ question: string; godar: string; competitor: string }> };
  sideBySide: { title: string; body: string; ctaLabel: string };
  painPoints: { subtitle: string; rows: Array<{ pain: string; benefit: string; impact: string }> };
  pitfalls: { subtitle: string; items: string[] };
  scenarios: { title: string; subtitle: string; items: Array<{ icon: LucideIcon; title: string; desc: string }> };
  faq: Array<{ question: string; answer: string }>;
  finalCta: { title: string; body: string; primaryLabel: string };
};

const migrationCode = `from openai import OpenAI

# فقط نشانی پایه، توکن و شناسه مدل عوض می‌شود
client = OpenAI(
    api_key="YOUR_GODARAI_TOKEN",
    base_url="https://YOUR_WORKSPACE.godarai.ir/v1",
)`;

function SectionHead({ title, copy }: { title: string; copy?: string }) {
  return (
    <div className="flex flex-col gap-4">
      <h2 className={`max-w-3xl ${sectionHeading}`}>{title}</h2>
      {copy ? <p className={`m-0 max-w-2xl ${sectionCopy}`}>{copy}</p> : null}
    </div>
  );
}

// Shared template for the gateway comparisons (/vs/litellm, /vs/portkey). The copy is fair to the competitor:
// each row says where it is the better fit, too.
export function ComparisonPage({ content }: { content: ComparisonContent }) {
  const { competitor, hero, differentiators, evaluation, sideBySide, painPoints, pitfalls, scenarios, faq, finalCta } = content;

  return (
    <main className="bg-site-ground">
      <PageHero
        title={`${hero.titleLead} ${hero.titleHighlight} ${hero.titleTail}`}
        lead={hero.intro}
        primary={{ href: '/onboarding', label: 'شروع ۱۵ روز رایگان' }}
        secondary={{ href: '/demo', label: 'رزرو دمو' }}
        visual={
          <ProductStage layout="size-full p-6 sm:p-10">
            <ProductWindow title={`گدارAI و ${competitor}`} label={`خلاصه‌ی تفاوت گدارAI با ${competitor}.`}>
              <MockRows rows={hero.chips.map((chip) => ({ icon: Check, title: chip }))} />
            </ProductWindow>
          </ProductStage>
        }
      />

      <Section id="differences" layout="flex flex-col gap-8">
        <SectionHead title="تفاوت‌ها، ردیف به ردیف." copy={differentiators.subtitle} />
        <ComparisonTable
          caption={`مقایسه‌ی گدارAI و ${competitor}`}
          columns={['گدارAI', competitor]}
          highlight={0}
          rows={differentiators.rows.map((row) => ({ label: row.category, values: [row.godar, row.competitor] }))}
        />
        <p className="m-0 text-lead font-medium text-site-ink">{hero.tagline}</p>
      </Section>

      <Section id="questions" tone="panel" layout="flex flex-col gap-8">
        <SectionHead title="پیش از انتخاب، این‌ها را بپرسید." copy={evaluation.subtitle} />
        <ul className="m-0 grid list-none grid-cols-1 gap-3 p-0 sm:gap-6 lg:grid-cols-2">
          {evaluation.items.map((item) => (
            <li key={item.question} className="flex flex-col gap-4 rounded-card border border-site-line bg-site-surface p-6 lg:p-8">
              <h3 className="m-0 text-lead font-semibold text-site-ink">{item.question}</h3>
              <dl className="m-0 flex flex-col gap-3">
                <div className="flex flex-col gap-1">
                  <dt className="text-caption font-semibold text-site-ink">گدارAI</dt>
                  <dd className="m-0 text-body text-site-ink">{item.godar}</dd>
                </div>
                <div className="flex flex-col gap-1">
                  <dt className="text-caption text-site-ink-soft">{competitor}</dt>
                  <dd className="m-0 text-body text-site-ink-soft">{item.competitor}</dd>
                </div>
              </dl>
            </li>
          ))}
        </ul>
      </Section>

      <FeatureRow
        id="side-by-side"
        reverse
        title={sideBySide.title}
        copy={sideBySide.body}
        link={{ href: '/features/ai-gateway', label: 'درگاه مدل‌های زبانی' }}
        visual={
          <ProductStage layout="flex w-full max-w-lg justify-center p-8 sm:p-12">
            <CodeWindow title="مهاجرت" meta="Python" code={migrationCode} />
          </ProductStage>
        }
      />

      <Section id="pains" tone="panel" layout="flex flex-col gap-8">
        <SectionHead title="کجا درد می‌کند؟ و چه چیزی عوض می‌شود." copy={painPoints.subtitle} />
        <ul className="m-0 grid list-none grid-cols-1 gap-3 p-0 sm:gap-6 md:grid-cols-2 lg:grid-cols-3">
          {painPoints.rows.map((row) => (
            <li key={row.pain} className="flex flex-col gap-3 rounded-card border border-site-line bg-site-surface p-6">
              <h3 className="m-0 text-lead font-semibold text-site-ink">{row.pain}</h3>
              <p className="m-0 text-body font-medium text-site-ink">{row.benefit}</p>
              <p className="m-0 text-body text-site-ink-soft">{row.impact}</p>
            </li>
          ))}
        </ul>
      </Section>

      <Section id="pitfalls" layout="grid gap-8 lg:grid-cols-2 lg:gap-16">
        <SectionHead title="اشتباه‌های رایج. هر درگاهی که انتخاب کنید." copy={pitfalls.subtitle} />
        <ol className="m-0 flex list-none flex-col divide-y divide-site-line border-y border-site-line p-0">
          {pitfalls.items.map((item, index) => (
            <li key={item} className="flex items-start gap-4 py-4">
              <span className="flex size-7 shrink-0 items-center justify-center rounded-full bg-site-sand text-caption font-semibold text-site-ink">
                {new Intl.NumberFormat('fa-IR').format(index + 1)}
              </span>
              <p className="m-0 text-body text-site-ink">{item}</p>
            </li>
          ))}
        </ol>
      </Section>

      <Section id="scenarios" tone="panel" layout="flex flex-col gap-8">
        <SectionHead title={scenarios.title} copy={scenarios.subtitle} />
        <TileGrid
          columns={4}
          size="compact"
          items={scenarios.items.map((item) => ({ key: item.title, icon: item.icon, title: item.title, text: item.desc }))}
        />
        <div className="flex flex-wrap gap-x-8 gap-y-3">
          <Link href="/features/cost-control" className={textLink}>
            کنترل هزینه
            <ArrowLeft aria-hidden="true" className="size-4" />
          </Link>
          <Link href="/features/ai-gateway" className={textLink}>
            درگاه مدل‌های زبانی
            <ArrowLeft aria-hidden="true" className="size-4" />
          </Link>
        </div>
      </Section>

      <FAQ title="تردیدهای رایج." items={faq} />

      <ClosingBand
        title={finalCta.title}
        copy={finalCta.body}
        primary={{ href: '/onboarding', label: finalCta.primaryLabel }}
        secondary={{ href: '/docs', label: 'مستندات' }}
      />
    </main>
  );
}
