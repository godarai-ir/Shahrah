import { MediaSlot } from '../MediaSlot';
import { Reveal } from '../Reveal';
import { CircleLink, sectionPad, TwoTone } from './parts';

export function WorkplaceProof() {
  return (
    <section className={`${sectionPad} py-20 md:py-28 lg:py-40`}>
      <div className="mx-auto w-full max-w-7xl">
        <div className="grid grid-cols-1 gap-12 md:grid-cols-2 md:gap-x-16 md:gap-y-20">
          <Reveal className="order-first md:order-none md:col-span-2">
            <div className="mx-auto max-w-4xl">
              <MediaSlot id="logos" />
            </div>
          </Reveal>

          <Reveal className="flex max-w-xl flex-col items-start gap-6 md:order-last">
            <TwoTone lead="سازمان‌هایی که به گدارAI اعتماد کرده‌اند." muted="ببینید چطور هوش مصنوعی را به کارشان آورده‌اند." />
            <CircleLink href="/blog">تجربه‌ی مشتری‌ها</CircleLink>
          </Reveal>

          <Reveal className="flex w-full flex-col gap-4 rounded-xl bg-white p-8 dark:bg-white/5 md:bg-transparent md:p-0 md:dark:bg-transparent">
            <p className="border-2 border-dashed border-zinc-900/20 p-4 text-lg leading-9 text-zinc-900/60 dark:border-white/20 dark:text-white/60">
              جای نقل‌قول مشتری: دو یا سه جمله از زبان او درباره‌ی اینکه هوش مصنوعی چطور وارد کار روزمره‌ی تیمشان شد.
            </p>
            <p className="text-sm font-light leading-snug text-zinc-900/55 dark:text-white/55">نام و نام خانوادگی، سمت در شرکت</p>
          </Reveal>
        </div>
      </div>
    </section>
  );
}
