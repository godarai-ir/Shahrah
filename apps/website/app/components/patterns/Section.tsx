import type { ReactNode } from 'react';
import { toneSurface } from '@/app/lib/pillars';
import type { PillarId } from '@/app/lib/pillars';
import { sectionSpace } from '../sectionType';

export type SectionTone = 'ground' | 'panel' | PillarId;

export const surfaceClass: Record<SectionTone, string> = {
  ground: 'bg-site-ground',
  panel: 'bg-site-panel',
  ...toneSurface,
};

const widthClass = {
  narrow: 'max-w-3xl',
  default: 'max-w-6xl',
  wide: 'max-w-7xl',
} as const;

type SectionProps = {
  children: ReactNode;
  id?: string;
  tone?: SectionTone;
  width?: keyof typeof widthClass;
  /** Layout utilities for the inner container (flex, gap, alignment). Colour comes only from `tone`. */
  layout?: string;
};

// Standard section rhythm: 80px vertical space rising to 128px, inside a centred container with a 24px gutter.
export function Section({ children, id, tone = 'ground', width = 'default', layout = '' }: SectionProps) {
  return (
    <section id={id} className={`w-full ${surfaceClass[tone]} ${sectionSpace}`}>
      <div className={`mx-auto w-full ${widthClass[width]} px-6 ${layout}`}>{children}</div>
    </section>
  );
}
