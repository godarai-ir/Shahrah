import type { LucideIcon } from 'lucide-react';
import { iconWell } from '../sectionType';

export type MockBadge = { label: string; tone: 'success' | 'danger' | 'neutral' };

export type MockRow = { title: string; text?: string; icon?: LucideIcon; badge?: MockBadge };

const badgeClass: Record<MockBadge['tone'], string> = {
  success: 'border-site-success text-site-success',
  danger: 'border-site-danger text-site-danger',
  neutral: 'border-site-line-strong text-site-ink-soft',
};

// Rows inside a drawn Product Window: a log, a list of rules, an access table. Status is always a word, never colour alone.
export function MockRows({ rows }: { rows: MockRow[] }) {
  return (
    <ul className="m-0 flex list-none flex-col divide-y divide-site-line p-0">
      {rows.map((row) => (
        <li key={`${row.title}-${row.text ?? ''}`} className="flex items-center gap-3 px-5 py-3">
          {row.icon ? (
            <span className={`size-8 ${iconWell}`}>
              <row.icon aria-hidden="true" className="size-4" />
            </span>
          ) : null}
          <span className="flex min-w-0 flex-1 flex-col">
            <span className="truncate text-label font-medium text-site-ink">{row.title}</span>
            {row.text ? <span className="truncate text-caption text-site-ink-soft">{row.text}</span> : null}
          </span>
          {row.badge ? (
            <span className={`shrink-0 rounded-full border px-2 py-0.5 text-caption font-medium ${badgeClass[row.badge.tone]}`}>{row.badge.label}</span>
          ) : null}
        </li>
      ))}
    </ul>
  );
}

export type MockBar = { label: string; percent: number; caption: string; over?: boolean };

// Usage or budget bars inside a drawn Product Window; the caption always states the value in words.
export function MockBars({ bars }: { bars: MockBar[] }) {
  return (
    <ul className="m-0 flex list-none flex-col gap-4 p-5">
      {bars.map((bar) => (
        <li key={bar.label} className="flex flex-col gap-1.5">
          <span className="flex items-baseline justify-between gap-3">
            <span className="text-label font-medium text-site-ink">{bar.label}</span>
            <span className={`text-caption ${bar.over ? 'text-site-danger' : 'text-site-ink-soft'}`}>{bar.caption}</span>
          </span>
          <span className="block h-2 overflow-hidden rounded-full bg-site-panel" aria-hidden="true">
            <span
              className={`block h-full rounded-full ${bar.over ? 'bg-site-danger' : 'bg-site-ink'}`}
              style={{ width: `${Math.min(bar.percent, 100)}%` }}
            />
          </span>
        </li>
      ))}
    </ul>
  );
}
