import { ArrowLeft } from 'lucide-react';
import Link from 'next/link';
import { pillars as allPillars, toneSurface } from '@/app/lib/pillars';
import type { Pillar } from '@/app/lib/pillars';

const columnClass = {
  2: 'md:grid-cols-2',
  3: 'md:grid-cols-3',
  4: 'md:grid-cols-2 lg:grid-cols-4',
} as const;

type PillarGridProps = {
  items?: Pillar[];
  columns?: keyof typeof columnClass;
  numbered?: boolean;
  /** `compact` for smaller cards inside another section. */
  size?: 'default' | 'compact';
};

// One card per differentiator, each on its own tone. Cards are links; they lift on hover and focus.
export function PillarGrid({ items = allPillars, columns = 4, numbered = false, size = 'default' }: PillarGridProps) {
  const compact = size === 'compact';

  return (
    <ul className={`m-0 grid list-none grid-cols-1 gap-4 p-0 ${columnClass[columns]}`}>
      {items.map((pillar, index) => (
        <li key={pillar.id} className="flex">
          <Link
            href={pillar.href}
            className={`group flex w-full flex-col rounded-3xl text-site-ink transition-shadow duration-300 hover:shadow-card focus-visible:shadow-card focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-site-accent ${
              toneSurface[pillar.id]
            } ${compact ? 'gap-3 p-6' : 'gap-6 p-8'}`}
          >
            <span className="flex items-center justify-between">
              <span className="flex size-10 items-center justify-center rounded-2xl bg-site-ground/70 text-site-accent">
                <pillar.icon className="size-5" />
              </span>
              {numbered ? (
                <span className="text-lg font-medium text-site-ink-soft">{(index + 1).toLocaleString('fa-IR', { minimumIntegerDigits: 2 })}</span>
              ) : null}
            </span>
            <span className="flex flex-1 flex-col gap-2">
              <span className={`font-semibold leading-snug ${compact ? 'text-lg' : 'text-xl'}`}>{pillar.title}</span>
              <span className="text-base leading-8 text-site-ink-soft">{pillar.summary}</span>
            </span>
            <ArrowLeft aria-hidden="true" className="size-4 text-site-ink-soft transition-transform duration-300 group-hover:-translate-x-0.5" />
          </Link>
        </li>
      ))}
    </ul>
  );
}
