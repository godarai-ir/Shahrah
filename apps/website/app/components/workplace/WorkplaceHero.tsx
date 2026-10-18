import { ArrowLeft, Play } from 'lucide-react';
import Link from 'next/link';
import { MediaSlot } from '../MediaSlot';
import { Reveal } from '../Reveal';
import { pillOutline, pillPrimary, sectionPad, TwoTone } from './parts';

export function AnnounceBar() {
  return (
    <Link
      href="/onboarding"
      className="flex min-h-10 items-center justify-center gap-3 bg-zinc-900/5 px-4 py-2 text-sm text-zinc-900 transition-colors hover:bg-zinc-900/10 dark:bg-white/10 dark:text-white dark:hover:bg-white/15"
    >
      <span className="opacity-75">آزمایش ۱۵ روزه‌ی گدارAI رایگان است.</span>
      <span className="flex items-center gap-1 opacity-55">
        شروع کنید
        <ArrowLeft className="size-4" />
      </span>
    </Link>
  );
}

export function WorkplaceHero() {
  return (
    <>
      <section className={`${sectionPad} pb-12 pt-16 md:pt-20`}>
        <div className="mx-auto w-full max-w-7xl text-start md:text-center">
          <Reveal className="mb-3">
            <TwoTone as="h1" lead="هوش مصنوعی را وارد کار روزانه‌ی سازمانتان کنید" className="mx-auto max-w-3xl" />
          </Reveal>
          <Reveal className="mb-8" delay={0.08}>
            <p className="mx-auto max-w-xl text-lg font-light max-md:mx-0 leading-9 text-zinc-900/55 dark:text-white/55">
              گدارAI هوش مصنوعی را به دست همه‌ی تیم‌ها می‌رساند؛ امن، قابل کنترل و بدون نیاز به برنامه‌نویسی. هر دستیار به
              ابزارهای کار روزانه‌تان وصل است.
            </p>
          </Reveal>
          <Reveal className="flex flex-wrap items-center justify-start gap-3 md:justify-center" delay={0.16}>
            <Link href="/onboarding" className={pillPrimary}>
              ۱۵ روز رایگان امتحان کنید
            </Link>
            <Link href="/demo" className={pillOutline}>
              با ما حرف بزنید
            </Link>
          </Reveal>
        </div>
      </section>

      <section className="overflow-hidden">
        <div className="px-0 md:px-10">
          <Reveal className="mx-auto w-full max-w-7xl">
            <Link
              href="/demo"
              className="group/video relative block w-full overflow-hidden text-white md:rounded-xl"
              aria-label="تور کوتاه محصول را ببینید"
            >
              <MediaSlot id="tourVideo" className="rounded-none md:rounded-xl" />
              <span className="absolute bottom-5 end-5 inline-flex items-center gap-4 rounded-full bg-zinc-900/70 py-3 pe-5 ps-4 text-sm backdrop-blur-sm transition-colors group-hover/video:bg-zinc-900/85">
                <Play className="size-4 fill-white" />
                تور کوتاه محصول را ببینید
              </span>
            </Link>
          </Reveal>
        </div>
      </section>
    </>
  );
}
