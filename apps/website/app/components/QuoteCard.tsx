import { Reveal } from './Reveal';
import { MediaSlot } from './MediaSlot';

export function QuoteCard() {
  return (
    <section className="w-full bg-white py-14 dark:bg-slate-950 lg:py-24">
      <div className="mx-auto w-full max-w-6xl px-6">
        <Reveal className="flex flex-col items-stretch gap-4 lg:flex-row">
          <div className="flex flex-1 flex-col gap-8 rounded-2xl bg-blue-50 p-10 dark:bg-blue-500/10 md:p-14 lg:gap-12">
            <div className="flex flex-1 items-center justify-center rounded-xl border-2 border-dashed border-blue-200 p-6 text-center dark:border-blue-500/30">
              <p className="text-sm leading-7 text-slate-600 dark:text-gray-300">
                جای نقل‌قول مشتری: دو یا سه جمله از زبان او درباره‌ی اینکه گدارAI چه کاری را برایشان عوض کرد.
              </p>
            </div>
            <div>
              <p className="text-sm font-medium text-slate-950 dark:text-white">نام و نام خانوادگی</p>
              <p className="text-xs text-slate-500 dark:text-gray-400">سمت، نام شرکت</p>
            </div>
          </div>
          <div className="w-full lg:w-96 lg:shrink-0 lg:self-stretch">
            <MediaSlot id="quotePortrait" className="lg:aspect-auto lg:h-full" />
          </div>
        </Reveal>
      </div>
    </section>
  );
}
