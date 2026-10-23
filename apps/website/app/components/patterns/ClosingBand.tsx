import { ArrowLeft } from 'lucide-react';
import type { ReactNode } from 'react';
import { buttonOnNavy, buttonPrimary } from '../sectionType';
import { ActionLink } from './PageHero';
import type { Action } from './PageHero';

type ClosingBandProps = {
  title: string;
  /** Second clause of the heading, the turn, set lighter. */
  titleMuted?: string;
  copy: string;
  primary: Action;
  secondary?: Action;
  /** Decoration drawn behind the content. */
  backdrop?: ReactNode;
};

// The page's closing call to action: one inverted navy card on the ground, kept apart from the navy footer below it.
export function ClosingBand({ title, titleMuted, copy, primary, secondary, backdrop }: ClosingBandProps) {
  return (
    <section className="w-full bg-site-ground px-4 py-10 sm:px-6 lg:py-20">
      <div className="relative mx-auto flex w-full max-w-content flex-col items-center gap-6 overflow-hidden rounded-card bg-site-night px-6 py-16 text-center text-site-night-ink lg:py-24">
        {backdrop}
        <h2 className="relative m-0 max-w-3xl text-balance text-h2 font-extrabold lg:text-h1">
          {title}
          {titleMuted ? (
            <>
              {' '}
              <span className="font-light text-site-night-soft">{titleMuted}</span>
            </>
          ) : null}
        </h2>
        <p className="relative m-0 max-w-2xl text-lead font-light text-site-night-soft">{copy}</p>
        <div className="relative flex flex-wrap items-center justify-center gap-2 pt-2">
          <ActionLink href={primary.href} className={buttonPrimary}>
            {primary.label}
          </ActionLink>
          {secondary ? (
            <ActionLink href={secondary.href} className={buttonOnNavy}>
              {secondary.label}
              <ArrowLeft aria-hidden="true" className="size-4" />
            </ActionLink>
          ) : null}
        </div>
      </div>
    </section>
  );
}
