import { ArrowLeft, Code2, LayoutGrid } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import Link from 'next/link';
import { Section } from './patterns';
import { iconWell, sectionCopy, sectionHeading } from './sectionType';

type Way = { icon: LucideIcon; title: string; audience: string; text: string; href: string; cta: string };

const ways: Way[] = [
  {
    icon: LayoutGrid,
    title: 'فضای کاری هوش مصنوعی',
    audience: 'برای همه‌ی کارکنان',
    text: 'گفتگو با هر مدل، دستیار برای هر تیم و اتصال به ابزارها.',
    href: '/workplace',
    cta: 'دیدن فضای کاری',
  },
  {
    icon: Code2,
    title: 'درگاه مدل‌های زبانی',
    audience: 'برای تیم فنی',
    text: 'یک نشانی سازگار با OpenAI برای ۱۲۰۰+ مدل از ۲۱ ارائه‌دهنده.',
    href: '/features/ai-gateway',
    cta: 'دیدن درگاه',
  },
];

// The two products of equal weight, side by side as flat tiles, each leading to its own hub.
export function TwoWays() {
  return (
    <Section id="products" tone="panel" layout="flex flex-col gap-8">
      <div className="flex flex-col gap-4">
        <div>
          <h2 className={`max-w-3xl ${sectionHeading}`}>یک گدارAI. دو راه ورود.</h2>
        </div>
        <div>
          <p className={`max-w-2xl ${sectionCopy}`}>کارکنان در فضای کاری با دستیارها کار می‌کنند؛ تیم فنی همان مدل‌ها را از درگاه در محصول خودش به کار می‌برد.</p>
        </div>
      </div>

      <ul className="m-0 grid list-none grid-cols-1 gap-3 p-0 sm:gap-6 md:grid-cols-2">
        {ways.map((way) => (
          <li key={way.href} className="flex">
            <Link
              href={way.href}
              className="group flex w-full flex-col gap-8 rounded-card border border-site-line bg-site-surface p-6 transition-colors duration-150 hover:bg-site-ground focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-site-focus lg:p-10"
            >
              <span className={`size-12 ${iconWell}`}>
                <way.icon aria-hidden="true" className="size-6" />
              </span>
              <span className="flex flex-1 flex-col gap-2">
                <span className="text-label text-site-ink-soft">{way.audience}</span>
                <span className="text-h3 font-bold text-site-ink lg:text-h2">{way.title}</span>
                <span className="max-w-md text-body text-site-ink-soft">{way.text}</span>
              </span>
              <span className="inline-flex items-center gap-2 text-label font-medium text-site-accent-text">
                {way.cta}
                <ArrowLeft aria-hidden="true" className="size-4 transition-transform duration-150 group-hover:-translate-x-0.5" />
              </span>
            </Link>
          </li>
        ))}
      </ul>
    </Section>
  );
}
