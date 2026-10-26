import { ArrowLeft } from 'lucide-react';
import Link from 'next/link';
import type { ReactNode } from 'react';
import { buttonOnNavy, buttonPrimary, buttonSecondary } from '../sectionType';

export type Action = { href: string; label: string };

// Docs, catalog and blog are separate apps behind the same domain: they need a full page load, not client navigation.
const otherApp = /^\/(docs|catalog|blog)(\/|$)/;

export function ActionLink({ href, className, children }: { href: string; className: string; children: ReactNode }) {
  return otherApp.test(href) ? (
    <a href={href} className={className}>
      {children}
    </a>
  ) : (
    <Link href={href} className={className}>
      {children}
    </Link>
  );
}

type PageHeroProps = {
  title: ReactNode;
  lead: ReactNode;
  primary: Action;
  secondary?: Action;
  /** Usually a ProductStage holding a ProductWindow. Placed after the copy on mobile. */
  visual?: ReactNode;
  /** `night` sets the hero on the brand's navy field; use it on the page's lead story. */
  tone?: 'ground' | 'night';
};

// First viewport of a page: Display heading, one or two sentences, a primary and a secondary action, the product beside them.
export function PageHero({ title, lead, primary, secondary, visual, tone = 'ground' }: PageHeroProps) {
  const night = tone === 'night';

  return (
    <section className={night ? 'bg-site-night text-site-night-ink' : 'bg-site-ground text-site-ink'}>
      <div className="mx-auto flex w-full max-w-content flex-col items-stretch gap-10 px-4 py-10 sm:px-6 lg:flex-row lg:items-center lg:gap-16 lg:py-20">
        <div className={`z-10 flex w-full flex-col items-start gap-6 ${visual ? 'lg:w-5/12' : 'lg:max-w-3xl'}`}>
          <div>
            <h1 className="text-balance text-h1 font-extrabold lg:text-display">{title}</h1>
          </div>
          <div>
            <p className={`max-w-xl text-pretty text-lead font-light ${night ? 'text-site-night-soft' : 'text-site-ink-soft'}`}>{lead}</p>
          </div>
          <div className="flex flex-row flex-wrap items-center gap-2 pt-2">
            <ActionLink href={primary.href} className={buttonPrimary}>
              {primary.label}
            </ActionLink>
            {secondary ? (
              <ActionLink href={secondary.href} className={night ? buttonOnNavy : buttonSecondary}>
                {secondary.label}
                <ArrowLeft aria-hidden="true" className="size-4" />
              </ActionLink>
            ) : null}
          </div>
        </div>

        {visual ? <div className="relative w-full lg:h-136 lg:w-7/12">{visual}</div> : null}
      </div>
    </section>
  );
}
