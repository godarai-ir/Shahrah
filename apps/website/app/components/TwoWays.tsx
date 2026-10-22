import { ArrowLeft, Code2, LayoutGrid } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import Link from 'next/link';
import { Section } from './patterns';
import { sectionCopy, sectionHeading } from './sectionType';

type Way = { icon: LucideIcon; title: string; audience: string; text: string; href: string; cta: string };

const ways: Way[] = [
  {
    icon: LayoutGrid,
    title: 'فضای کاری هوش مصنوعی',
    audience: 'برای همه‌ی کارکنان',
    text: 'گفتگو با هر مدل، دستیار برای هر تیم و اتصال به ابزارها.',
    href: '/workplace',
    cta: 'فضای کاری را ببینید',
  },
  {
    icon: Code2,
    title: 'درگاه مدل‌های زبانی',
    audience: 'برای تیم فنی',
    text: 'یک نشانی سازگار با OpenAI برای ۱۲۰۰+ مدل از ۲۱ ارائه‌دهنده.',
    href: '/features/ai-gateway',
    cta: 'درگاه را ببینید',
  },
];

// The two products of equal weight, side by side, each leading to its own hub.
export function TwoWays() {
  return (
    <Section id="products" layout="flex flex-col gap-12">
      <div className="flex flex-col gap-6">
        <div>
          <h2 className={`max-w-3xl ${sectionHeading}`}>یک گدارAI، دو راه استفاده</h2>
        </div>
        <div>
          <p className={`max-w-2xl ${sectionCopy}`}>کارکنان در فضای کاری با دستیارها کار می‌کنند؛ تیم فنی همان مدل‌ها را از درگاه در محصول خودش به کار می‌برد.</p>
        </div>
      </div>

      <ul className="m-0 grid list-none grid-cols-1 gap-4 p-0 md:grid-cols-2">
        {ways.map((way) => (
          <li key={way.href} className="flex">
            <Link
              href={way.href}
              className="group flex w-full flex-col gap-10 rounded-3xl bg-site-panel p-8 transition-shadow duration-300 hover:shadow-card focus-visible:shadow-card focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-site-accent lg:p-12"
            >
              <span className="flex size-12 items-center justify-center rounded-2xl bg-site-panel-deep text-site-accent">
                <way.icon className="size-6" />
              </span>
              <span className="flex flex-1 flex-col gap-3">
                <span className="text-sm text-site-ink-soft">{way.audience}</span>
                <span className="text-2xl font-semibold leading-snug text-site-ink lg:text-3xl">{way.title}</span>
                <span className="max-w-md text-base leading-8 text-site-ink-soft">{way.text}</span>
              </span>
              <span className="inline-flex items-center gap-2 text-sm font-medium text-site-ink group-hover:text-site-accent">
                {way.cta}
                <ArrowLeft aria-hidden="true" className="size-4 transition-transform duration-300 group-hover:-translate-x-0.5" />
              </span>
            </Link>
          </li>
        ))}
      </ul>
    </Section>
  );
}
