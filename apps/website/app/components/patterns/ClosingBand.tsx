import { ArrowLeft } from 'lucide-react';
import Link from 'next/link';
import type { ReactNode } from 'react';
import type { Action } from './PageHero';

type ClosingBandProps = {
  title: string;
  /** Second half of the heading, set lighter in Dusk Grey. */
  titleMuted?: string;
  copy: string;
  primary: Action;
  secondary?: Action;
  /** Decoration drawn behind the content. */
  backdrop?: ReactNode;
};

// The one dark band per page, holding the final call to action.
export function ClosingBand({ title, titleMuted, copy, primary, secondary, backdrop }: ClosingBandProps) {
  return (
    <section className="relative w-full overflow-hidden bg-site-night py-28 text-site-night-ink lg:py-40">
      {backdrop}
      <div className="relative mx-auto flex w-full max-w-4xl flex-col items-center gap-10 px-6 text-center">
        <div>
          <h2 className="m-0 max-w-3xl text-balance text-center text-4xl font-semibold leading-snug md:text-5xl xl:text-6xl xl:leading-snug">
            {title}
            {titleMuted ? (
              <>
                {' '}
                <span className="font-normal text-site-night-soft">{titleMuted}</span>
              </>
            ) : null}
          </h2>
        </div>
        <div>
          <p className="m-0 max-w-2xl text-lg leading-9 text-site-night-soft">{copy}</p>
        </div>
        <div className="flex flex-col items-center gap-5">
          <Link
            href={primary.href}
            className="inline-flex h-12 items-center justify-center rounded-full bg-site-night-ink px-7 text-base font-medium text-site-night transition-colors hover:bg-site-night-strong focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-site-night-ink"
          >
            {primary.label}
          </Link>
          {secondary ? (
            <Link
              href={secondary.href}
              className="inline-flex items-center gap-2 rounded-full text-sm text-site-night-soft transition-colors hover:text-site-night-ink focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-site-night-ink"
            >
              {secondary.label}
              <ArrowLeft className="size-4" />
            </Link>
          ) : null}
        </div>
      </div>
    </section>
  );
}
