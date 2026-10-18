import Link from 'next/link';
import { Reveal } from './Reveal';
import { MediaSlot } from './MediaSlot';

export function Hero() {
  return (
    <section className="bg-white pb-12 dark:bg-slate-950">
      <div className="mx-auto flex w-full max-w-screen-2xl flex-col items-stretch gap-10 px-6 pt-12 lg:flex-row lg:items-center lg:gap-0 lg:px-10 lg:pt-20">
        <div className="z-10 flex w-full flex-col items-start gap-6 lg:w-5/12 lg:pe-8">
          <Reveal>
            <h1 className="text-balance text-4xl font-semibold leading-tight text-slate-950 dark:text-white sm:text-5xl lg:text-6xl xl:text-7xl xl:leading-tight">
              دستیاری که شرکتتان را می‌شناسد
            </h1>
          </Reveal>
          <Reveal delay={0.1}>
            <p className="max-w-xl text-pretty text-lg leading-9 text-slate-600 dark:text-gray-400 lg:text-xl lg:leading-10">
              گدارAI هوش مصنوعی را به ابزارهای روزمره‌ی تیم‌تان وصل می‌کند، از CRM تا بله و ایتا. کارهای تکراری با دستیار؛ تصمیم‌های مهم با خودتان.
            </p>
          </Reveal>
          <Reveal delay={0.2}>
            <div className="flex flex-row flex-wrap items-center justify-start gap-3">
              <Link
                href="/demo"
                className="inline-flex h-12 items-center justify-center rounded-2xl bg-blue-500 px-6 text-base font-medium text-white transition-colors hover:bg-blue-600"
              >
                با ما حرف بزنید
              </Link>
              <Link
                href="/onboarding"
                className="inline-flex h-12 items-center justify-center rounded-2xl px-5 text-base font-medium text-slate-600 transition-colors hover:text-slate-950 dark:text-gray-300 dark:hover:text-white"
              >
                ۱۵ روز رایگان امتحان کنید ←
              </Link>
            </div>
          </Reveal>
        </div>

        <Reveal className="relative w-full lg:h-144 lg:w-7/12">
          <MediaSlot id="heroDemo" priority className="rounded-3xl" />
        </Reveal>
      </div>
    </section>
  );
}
