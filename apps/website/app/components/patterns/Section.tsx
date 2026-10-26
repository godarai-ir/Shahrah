import type { ReactNode } from 'react';
import { sectionSpace } from '../sectionType';

// Sections alternate ground and panel; surface and sand are for tiles and cards inside them.
export type SectionTone = 'ground' | 'panel' | 'surface' | 'sand';

export const surfaceClass: Record<SectionTone, string> = {
  ground: 'bg-site-ground',
  panel: 'bg-site-panel',
  surface: 'bg-site-surface',
  sand: 'bg-site-sand',
};

const widthClass = {
  narrow: 'max-w-3xl',
  default: 'max-w-content',
  wide: 'max-w-content',
} as const;

type SectionProps = {
  children: ReactNode;
  id?: string;
  tone?: SectionTone;
  width?: keyof typeof widthClass;
  /** Layout utilities for the inner container (flex, gap, alignment). Colour comes only from `tone`. */
  layout?: string;
};

// Standard section rhythm: 40px vertical space on mobile, 80px on desktop, inside a container of at most 1200px.
export function Section({ children, id, tone = 'ground', width = 'default', layout = '' }: SectionProps) {
  return (
    <section id={id} className={`w-full ${surfaceClass[tone]} ${sectionSpace}`}>
      <div className={`mx-auto w-full ${widthClass[width]} px-4 sm:px-6 ${layout}`}>{children}</div>
    </section>
  );
}
