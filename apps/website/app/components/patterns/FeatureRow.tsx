import { ArrowLeft } from 'lucide-react';
import Link from 'next/link';
import type { ReactNode } from 'react';
import { sectionCopy, sectionHeading, textLink } from '../sectionType';
import { Section } from './Section';
import type { SectionTone } from './Section';
import type { Action } from './PageHero';

type FeatureRowProps = {
  title: string;
  copy: ReactNode;
  visual: ReactNode;
  id?: string;
  bullets?: string[];
  link?: Action;
  /** Puts the visual on the reading-start side (right in RTL). */
  reverse?: boolean;
  tone?: SectionTone;
};

// One idea per section: copy on one half, the product on the other. Rows alternate sides with `reverse`.
export function FeatureRow({ title, copy, visual, id, bullets, link, reverse = false, tone }: FeatureRowProps) {
  return (
    <Section
      id={id}
      tone={tone}
      layout={`flex flex-col items-stretch gap-8 lg:items-center lg:gap-16 ${reverse ? 'lg:flex-row-reverse' : 'lg:flex-row'}`}
    >
      <div className="flex w-full flex-col gap-4 lg:w-1/2">
        <div>
          <h2 className={sectionHeading}>{title}</h2>
        </div>
        <div>
          <p className={`max-w-md ${sectionCopy}`}>{copy}</p>
        </div>
        {bullets ? (
          <ul className="m-0 flex max-w-md list-none flex-col gap-3 p-0">
            {bullets.map((bullet) => (
              <li key={bullet} className="flex items-start gap-3">
                <span className="mt-3 block size-1.5 shrink-0 rounded-full bg-site-ink-soft" />
                <span className="text-body text-site-ink">{bullet}</span>
              </li>
            ))}
          </ul>
        ) : null}
        {link ? (
          <div>
            <Link
              href={link.href}
              className={textLink}
            >
              {link.label}
              <ArrowLeft aria-hidden="true" className="size-4 transition-transform duration-150 group-hover/link:-translate-x-0.5" />
            </Link>
          </div>
        ) : null}
      </div>

      <div className="flex w-full justify-center lg:w-1/2">{visual}</div>
    </Section>
  );
}
