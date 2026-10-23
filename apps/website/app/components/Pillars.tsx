import { PillarGrid, Section } from './patterns';
import { sectionCopy, sectionHeading } from './sectionType';

// The four differentiators as a row of flat tiles, each linked to its page once it exists.
export function Pillars() {
  return (
    <Section id="control" layout="flex flex-col gap-8">
      <div className="flex flex-col items-center gap-4 text-center">
        <div>
          <h2 className={`max-w-3xl ${sectionHeading}`}>کنترل را جدی بگیرید. نه دست‌وپاگیر.</h2>
        </div>
        <div>
          <p className={`max-w-2xl ${sectionCopy}`}>چهار سؤالی که مدیر فناوری پیش از تأیید می‌پرسد. هر کدام جواب روشن دارد.</p>
        </div>
      </div>
      <PillarGrid numbered />
    </Section>
  );
}
