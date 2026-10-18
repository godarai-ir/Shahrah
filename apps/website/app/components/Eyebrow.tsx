import type { ReactNode } from 'react';

export function Eyebrow({ children }: { children: ReactNode }) {
  return (
    <span className="inline-flex h-7 w-fit items-center gap-2 rounded-full bg-blue-100 px-3 text-xs font-medium text-blue-700 dark:bg-blue-500/15 dark:text-blue-300">
      <span className="h-1.5 w-1.5 rounded-full bg-blue-500" />
      {children}
    </span>
  );
}

export const sectionHeading =
  'py-2 text-balance text-3xl font-semibold leading-tight text-slate-950 dark:text-white md:text-4xl lg:text-5xl';

export const sectionCopy = 'text-base leading-8 text-slate-600 dark:text-gray-400 lg:text-lg lg:leading-9';
