import { ArrowLeft } from 'lucide-react';
import Link from 'next/link';
import type { ReactNode } from 'react';

export const pillPrimary =
  'inline-flex min-h-11 items-center justify-center gap-2 rounded-full bg-zinc-900 px-5 py-2.5 text-center text-sm font-medium text-white transition-colors hover:bg-zinc-700 dark:bg-white dark:text-zinc-900 dark:hover:bg-zinc-200';

export const pillOutline =
  'inline-flex min-h-11 items-center justify-center gap-2 rounded-full border border-zinc-900/20 px-5 py-2.5 text-center text-sm font-medium text-zinc-900 transition-colors hover:bg-zinc-900/5 dark:border-white/25 dark:text-white dark:hover:bg-white/10';

export const sectionPad = 'px-5 md:px-10';

/** Two-tone headline: the lead is full ink, the rest is muted. */
export function TwoTone({ lead, muted, as: Tag = 'h2', className = '' }: { lead: string; muted?: string; as?: 'h1' | 'h2'; className?: string }) {
  return (
    <Tag className={`text-balance text-2xl font-medium leading-snug text-zinc-900 dark:text-white md:text-4xl lg:text-5xl lg:leading-tight ${className}`}>
      {lead}
      {muted ? (
        <>
          {' '}
          <span className="text-zinc-900/55 dark:text-white/55">{muted}</span>
        </>
      ) : null}
    </Tag>
  );
}

/** Small text link with the round arrow button used across the page. */
export function CircleLink({ href, children, tone = 'default' }: { href: string; children: ReactNode; tone?: 'default' | 'light' | 'ink' }) {
  return (
    <Link
      href={href}
      className={`group/button-icon inline-flex items-center gap-3 text-sm transition-colors ${
        tone === 'light'
          ? 'text-white/85 hover:text-white'
          : tone === 'ink'
            ? 'text-zinc-900/75 hover:text-zinc-900'
            : 'text-zinc-900/75 hover:text-zinc-900 dark:text-white/75 dark:hover:text-white'
      }`}
    >
      <span
        className={`flex size-8 flex-none items-center justify-center rounded-full transition-transform ${
          tone === 'light' ? 'bg-white/20' : tone === 'ink' ? 'bg-zinc-900/10' : 'bg-zinc-900/10 dark:bg-white/15'
        }`}
      >
        <ArrowLeft className="size-4 transition-transform group-hover/button-icon:-rotate-45" />
      </span>
      {children}
    </Link>
  );
}
