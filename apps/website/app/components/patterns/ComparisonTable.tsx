import type { ReactNode } from 'react';

export type ComparisonRow = { label: string; values: ReactNode[]; note?: string };

type ComparisonTableProps = {
  /** Column headings after the row-label column, one per compared option. */
  columns: string[];
  rows: ComparisonRow[];
  caption: string;
  /** Index of the column to highlight (usually گدارAI). */
  highlight?: number;
};

// A fair side-by-side comparison. A table from `md` up; stacked cards below so nothing scrolls sideways.
export function ComparisonTable({ columns, rows, caption, highlight }: ComparisonTableProps) {
  return (
    <>
      <div className="hidden overflow-hidden rounded-card border border-site-line bg-site-surface md:block">
        <table className="w-full border-collapse text-start">
          <caption className="sr-only">{caption}</caption>
          <thead>
            <tr className="bg-site-panel">
              <th scope="col" className="w-1/4 px-6 py-4 text-start text-sm font-medium text-site-ink-soft">
                <span className="sr-only">ویژگی</span>
              </th>
              {columns.map((column, index) => (
                <th
                  key={column}
                  scope="col"
                  className={`px-6 py-4 text-start text-base font-semibold text-site-ink ${index === highlight ? 'bg-site-sand' : ''}`}
                >
                  {column}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.label} className="border-t border-site-line">
                <th scope="row" className="px-6 py-4 text-start align-top text-sm font-medium text-site-ink">
                  {row.label}
                  {row.note ? <span className="mt-1 block text-xs font-normal leading-6 text-site-ink-soft">{row.note}</span> : null}
                </th>
                {row.values.map((value, index) => (
                  <td
                    key={columns[index]}
                    className={`px-6 py-4 align-top text-sm leading-7 text-site-ink-soft ${index === highlight ? 'bg-site-panel text-site-ink' : ''}`}
                  >
                    {value}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <ul className="m-0 flex list-none flex-col gap-4 p-0 md:hidden" aria-label={caption}>
        {rows.map((row) => (
          <li key={row.label} className="rounded-card border border-site-line bg-site-surface p-6">
            <p className="mb-4 text-base font-semibold text-site-ink">{row.label}</p>
            <dl className="m-0 flex flex-col gap-3">
              {columns.map((column, index) => (
                <div key={column} className="flex flex-col gap-1">
                  <dt className={`text-xs ${index === highlight ? 'font-semibold text-site-ink' : 'text-site-ink-soft'}`}>{column}</dt>
                  <dd className="m-0 text-sm leading-7 text-site-ink">{row.values[index]}</dd>
                </div>
              ))}
            </dl>
          </li>
        ))}
      </ul>
    </>
  );
}
