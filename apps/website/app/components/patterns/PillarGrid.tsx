import { ArrowLeft } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import Link from 'next/link';
import type { ReactNode } from 'react';
import { isReady } from '@/app/lib/navigation';
import { pillars as allPillars } from '@/app/lib/pillars';
import type { Pillar } from '@/app/lib/pillars';
import { iconWell } from '../sectionType';

const columnClass = {
  2: 'md:grid-cols-2',
  3: 'md:grid-cols-3',
  4: 'md:grid-cols-2 lg:grid-cols-4',
} as const;

export type Tile = { key: string; icon: LucideIcon; title: string; text: string; href?: string };

function CardFrame({ href, className, children }: { href?: string; className: string; children: ReactNode }) {
  if (!href) {
    return <div className={className}>{children}</div>;
  }

  return (
    <Link
      href={href}
      className={`${className} transition-colors duration-150 hover:bg-site-ground focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-site-focus`}
    >
      {children}
    </Link>
  );
}

type TileGridProps = {
  items: Tile[];
  columns?: keyof typeof columnClass;
  numbered?: boolean;
  /** `compact` for smaller tiles inside another section. */
  size?: 'default' | 'compact';
};

// A row of flat tiles: an icon well, a bold H3 label and one line of body (DESIGN.md, Tiles).
// A tile links only when its page exists; otherwise it is a plain tile, so nothing links to a 404.
export function TileGrid({ items, columns = 4, numbered = false, size = 'default' }: TileGridProps) {
  const compact = size === 'compact';

  return (
    <ul className={`m-0 grid list-none grid-cols-1 gap-3 p-0 sm:gap-6 ${columnClass[columns]}`}>
      {items.map((item, index) => {
        const href = item.href && isReady(item.href) ? item.href : undefined;

        return (
          <li key={item.key} className="flex">
            <CardFrame
              href={href}
              className={`group flex w-full flex-col rounded-card border border-site-line bg-site-surface text-site-ink ${compact ? 'gap-3 p-6' : 'gap-6 p-6 lg:p-8'}`}
            >
              <span className="flex items-center justify-between">
                <span className={`size-10 ${iconWell}`}>
                  <item.icon aria-hidden="true" className="size-5" />
                </span>
                {numbered ? (
                  <span className="text-label font-medium text-site-ink-soft">{(index + 1).toLocaleString('fa-IR', { minimumIntegerDigits: 2 })}</span>
                ) : null}
              </span>
              <span className="flex flex-1 flex-col gap-2">
                <span className={`font-semibold ${compact ? 'text-lead' : 'text-h3'}`}>{item.title}</span>
                <span className="text-body text-site-ink-soft">{item.text}</span>
              </span>
              {href ? (
                <ArrowLeft aria-hidden="true" className="size-4 text-site-ink-soft transition-transform duration-150 group-hover:-translate-x-0.5" />
              ) : null}
            </CardFrame>
          </li>
        );
      })}
    </ul>
  );
}

const toTile = (pillar: Pillar): Tile => ({ key: pillar.id, icon: pillar.icon, title: pillar.title, text: pillar.summary, href: pillar.href });

type PillarGridProps = Omit<TileGridProps, 'items'> & { items?: Pillar[] };

// The four differentiators as tiles, each linked to its page once it exists.
export function PillarGrid({ items = allPillars, ...rest }: PillarGridProps) {
  return <TileGrid items={items.map(toTile)} {...rest} />;
}
