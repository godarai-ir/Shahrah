import Link from 'next/link';
import { surfaceClass } from './Section';
import type { SectionTone } from './Section';

export type ControlItem = { label: string; href?: string };

export type ControlGroup = {
  title: string;
  items: ControlItem[];
  /** The column's fill; columns usually step between surface and sand. */
  surface?: SectionTone;
};

const groupSurface = (surface: ControlGroup['surface'] = 'surface') => surfaceClass[surface];

const columnClass = ['', 'md:grid-cols-1', 'md:grid-cols-2', 'md:grid-cols-3', 'md:grid-cols-2 lg:grid-cols-4'];

// Enterprise controls as joined columns. Only confirmed controls belong here (PRODUCT.md, Capabilities and Constraints).
export function ControlChecklist({ groups }: { groups: ControlGroup[] }) {
  return (
    <div className={`grid grid-cols-1 overflow-hidden rounded-card border border-site-line ${columnClass[Math.min(groups.length, 4)]}`}>
      {groups.map((group) => (
        <div key={group.title} className={`flex flex-col ${groupSurface(group.surface)}`}>
          <header className="flex flex-col px-6 pb-4 pt-6">
            <h3 className="m-0 text-h3 font-semibold text-site-ink">{group.title}</h3>
          </header>
          <ul className="m-0 flex list-none flex-col p-0 pb-6">
            {group.items.map((item) => (
              <li key={item.label} className="flex items-start gap-3 px-6 py-2">
                <span className="mt-2.5 block size-1.5 shrink-0 rounded-full bg-site-ink-soft" />
                {item.href ? (
                  <Link
                    href={item.href}
                    className="rounded text-body text-site-ink underline underline-offset-4 decoration-site-line-strong hover:text-site-accent-text focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-site-focus"
                  >
                    {item.label}
                  </Link>
                ) : (
                  <span className="text-body text-site-ink">{item.label}</span>
                )}
              </li>
            ))}
          </ul>
        </div>
      ))}
    </div>
  );
}
