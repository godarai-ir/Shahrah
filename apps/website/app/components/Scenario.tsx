import { Eyebrow, sectionCopy, sectionHeading } from './Eyebrow';
import { Reveal } from './Reveal';
import { WorkflowCard } from './WorkflowCard';

export function Scenario() {
  return (
    <section id="scenario" className="w-full bg-white py-14 dark:bg-slate-950 lg:py-24">
      <div className="mx-auto flex w-full max-w-6xl flex-col items-center gap-12 px-6 lg:flex-row-reverse lg:items-center lg:gap-20">
        <div className="flex w-full flex-col gap-6 lg:w-1/2">
          <Reveal>
            <Eyebrow>یک مثال ساده</Eyebrow>
          </Reveal>
          <Reveal>
            <h2 className={sectionHeading}>از سابقه‌ی مشتری تا پیش‌نویس جواب</h2>
          </Reveal>
          <Reveal>
            <p className={`max-w-md ${sectionCopy}`}>
              مشتری می‌نویسد فاکتورش اشتباه است. دستیار می‌فهمد درباره‌ی کدام سفارش حرف می‌زند، فاکتور و سابقه‌اش را از سیستم‌های شما بیرون می‌کشد و جواب را می‌نویسد. همکار پشتیبانی نگاهی می‌اندازد و می‌فرستد. همین.
            </p>
          </Reveal>
        </div>

        <Reveal className="flex w-full justify-center self-stretch lg:w-1/2">
          <div className="flex w-full max-w-lg items-center justify-center self-stretch rounded-3xl bg-blue-50 p-8 dark:bg-blue-500/10">
            <WorkflowCard />
          </div>
        </Reveal>
      </div>
    </section>
  );
}
