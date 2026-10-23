import { pillars } from '@/app/lib/pillars';
import type { PillarId } from '@/app/lib/pillars';
import { sectionHeading } from '../sectionType';
import { PillarGrid } from './PillarGrid';
import { Section } from './Section';

type PillarCrossLinksProps = {
  current: PillarId;
  title?: string;
};

// Closes a differentiator page by pointing to the other three.
export function PillarCrossLinks({ current, title = 'کنترل‌های دیگر' }: PillarCrossLinksProps) {
  return (
    <Section layout="flex flex-col gap-10">
      <div>
        <h2 className={sectionHeading}>{title}</h2>
      </div>
      <PillarGrid items={pillars.filter((pillar) => pillar.id !== current)} columns={3} size="compact" />
    </Section>
  );
}
