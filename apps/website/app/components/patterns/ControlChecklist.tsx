import Link from 'next/link';
import { surfaceClass } from './Section';
import type { SectionTone } from './Section';

export type ControlItem = { label: string; href?: string };

export type ControlGroup = {
  title: string;
  items: ControlItem[];
  /** `panel`, `panel-deep` or a differentiator tone. */
  surface?: SectionTone | 'panel-deep';
};

const groupSurface = (surface: ControlGroup['surface'] = 'panel') =>
  surface === 'panel-deep' ? 'bg-site-panel-deep' : surfaceClass[surface];

const columnClass = ['', 'md:grid-cols-1', 'md:grid-cols-2', 'md:grid-cols-3', 'md:grid-cols-2 lg:grid-cols-4'];

// Enterprise controls as joined columns. Only confirmed controls belong here (PRODUCT.md, Capabilities and Constraints).
export function ControlChecklist({ groups }: { groups: ControlGroup[] }) {
  return (
    <div className={`grid grid-cols-1 overflow-hidden rounded-3xl ${columnClass[Math.min(groups.length, 4)]}`}>
      {groups.map((group) => (
        <div key={group.title} className={`flex flex-col ${groupSurface(group.surface)}`}>
          <header className="flex flex-col px-8 pb-6 pt-8">
            <span className="text-lg font-semibold leading-snug text-site-ink">{group.title}</span>
          </header>
          <ul className="m-0 flex list-none flex-col p-0 pb-5">
            {group.items.map((item) => (
              <li key={item.label} className="flex items-start gap-3 px-8 py-3">
                <span className="mt-3 block size-1.5 shrink-0 rounded-full bg-site-accent" />
                {item.href ? (
                  <Link
                    href={item.href}
                    className="rounded text-sm leading-7 text-site-ink underline-offset-4 hover:text-site-accent hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-site-accent"
                  >
                    {item.label}
                  </Link>
                ) : (
                  <span className="text-sm leading-7 text-site-ink">{item.label}</span>
                )}
              </li>
            ))}
          </ul>
        </div>
      ))}
    </div>
  );
}
