import { PillarGrid, Section } from './patterns';
import { sectionCopy, sectionHeading } from './sectionType';

// The four differentiators, each on its own tone and linked to its page.
export function Pillars() {
  return (
    <Section id="control" layout="flex flex-col gap-12">
      <div className="flex flex-col items-center gap-6 text-center">
        <div>
          <h2 className={`max-w-3xl ${sectionHeading}`}>کنترل، بخشی از خود محصول است</h2>
        </div>
        <div>
          <p className={`max-w-2xl ${sectionCopy}`}>چهار چیزی که مدیر فناوری پیش از تأیید می‌پرسد، هر کدام جواب مشخص دارد.</p>
        </div>
      </div>
      <PillarGrid numbered />
    </Section>
  );
}
