import { ArrowLeft } from 'lucide-react';
import Link from 'next/link';
import type { ReactNode } from 'react';
import { buttonPrimary, buttonQuiet } from '../sectionType';

export type Action = { href: string; label: string };

type PageHeroProps = {
  title: ReactNode;
  lead: ReactNode;
  primary: Action;
  secondary?: Action;
  /** Usually a ProductStage holding a ProductWindow. Placed after the copy on mobile. */
  visual?: ReactNode;
};

// First viewport of a page: Display heading, one or two sentences, two actions, and the product beside them.
export function PageHero({ title, lead, primary, secondary, visual }: PageHeroProps) {
  return (
    <section className="bg-site-ground pb-16 lg:pb-24">
      <div className="mx-auto flex w-full max-w-7xl flex-col items-stretch gap-14 px-6 pt-16 lg:flex-row lg:items-center lg:gap-16 lg:px-10 lg:pt-28">
        <div className={`z-10 flex w-full flex-col items-start gap-8 ${visual ? 'lg:w-5/12' : 'lg:max-w-3xl'}`}>
          <div>
            <h1 className="text-balance text-4xl font-semibold leading-snug text-site-ink sm:text-5xl lg:text-6xl lg:leading-snug">
              {title}
            </h1>
          </div>
          <div>
            <p className="max-w-xl text-pretty text-lg leading-9 text-site-ink-soft lg:text-xl lg:leading-10">{lead}</p>
          </div>
          <div>
            <div className="flex flex-row flex-wrap items-center gap-3">
              <Link href={primary.href} className={buttonPrimary}>
                {primary.label}
              </Link>
              {secondary ? (
                <Link href={secondary.href} className={`${buttonQuiet} gap-2`}>
                  {secondary.label}
                  <ArrowLeft className="size-4" />
                </Link>
              ) : null}
            </div>
          </div>
        </div>

        {visual ? <div className="relative w-full lg:h-144 lg:w-7/12">{visual}</div> : null}
      </div>
    </section>
  );
}
