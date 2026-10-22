import type { ReactNode } from 'react';
import { MediaSlot } from '../MediaSlot';
import type { HomeMediaId } from '@/app/lib/home-media';
import { surfaceClass } from './Section';
import type { SectionTone } from './Section';

type ProductWindowProps = {
  title: string;
  /** Right of the title bar; a plain string renders as quiet metadata, such as «نمونه». */
  meta?: ReactNode;
  /** A screenshot slot from `lib/home-media.ts`; used instead of `children` when given. */
  media?: HomeMediaId;
  children?: ReactNode;
  /** Description for screen readers when the window is a drawn illustration rather than a screenshot. */
  label?: string;
  /** Width utilities, for example `max-w-sm`. */
  size?: string;
};

// The system's signature: real product UI framed as a window resting on a panel (DESIGN.md, Product Window).
export function ProductWindow({ title, meta, media, children, label, size = 'max-w-sm' }: ProductWindowProps) {
  return (
    <div
      role={label ? 'img' : undefined}
      aria-label={label}
      className={`w-full ${size} overflow-hidden rounded-3xl bg-site-ground shadow-window`}
    >
      <div className="flex items-center justify-between border-b border-site-line px-5 py-4">
        <span className="text-sm font-semibold text-site-ink">{title}</span>
        {typeof meta === 'string' ? <span className="text-xs text-site-ink-soft">{meta}</span> : meta}
      </div>
      {media ? <MediaSlot id={media} className="rounded-none" /> : children}
    </div>
  );
}

type ProductStageProps = {
  children: ReactNode;
  tone?: SectionTone;
  /** Decoration drawn behind the window. */
  backdrop?: ReactNode;
  /** Size and padding utilities. */
  layout?: string;
  /** Lets the window cross the panel's edge, so it reads as an object resting on the bench. */
  overhang?: boolean;
};

// The bench a Product Window rests on: a 24px-cornered panel, tinted when the section belongs to a differentiator.
export function ProductStage({ children, tone = 'panel', backdrop, layout = 'w-full p-6 sm:p-10', overhang = false }: ProductStageProps) {
  return (
    <div className={`relative flex items-center justify-center rounded-3xl ${overhang ? '' : 'overflow-hidden'} ${surfaceClass[tone]} ${layout}`}>
      {backdrop}
      {children}
    </div>
  );
}
