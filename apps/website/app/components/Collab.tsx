import { Eyebrow, sectionCopy, sectionHeading } from './Eyebrow';
import { MediaSlot } from './MediaSlot';
import { Reveal } from './Reveal';

export function Collab() {
  return (
    <section id="collab" className="w-full bg-white py-14 dark:bg-slate-950 lg:py-24">
      <div className="mx-auto flex w-full max-w-6xl flex-col px-6">
        <div className="flex flex-col items-stretch gap-12 lg:flex-row lg:items-center lg:gap-20">
          <div className="flex w-full flex-col gap-6 lg:w-1/2">
            <Reveal>
              <Eyebrow>همکاری تیمی</Eyebrow>
            </Reveal>
            <Reveal>
              <h2 className={sectionHeading}>کار میان همکارها و دستیار دست‌به‌دست می‌شود</h2>
            </Reveal>
            <Reveal>
              <p className={`max-w-md ${sectionCopy}`}>
                دستیار پیش‌نویس را آماده می‌کند، همکارتان نگاهی می‌اندازد و کار از همان‌جا که تیم‌تان هست پیش می‌رود: بله، ایتا، واتساپ، تلگرام، Jira یا Notion.
              </p>
            </Reveal>
            <Reveal className="mt-4 w-full">
              <figure className="m-0 flex w-full max-w-lg flex-col gap-4 border-s-2 border-slate-950/15 ps-5 dark:border-white/20">
                <blockquote className="m-0 border-2 border-dashed border-slate-300 p-4 text-sm leading-7 text-slate-600 dark:border-slate-700 dark:text-gray-400">
                  جای نقل‌قول کوتاه مشتری.
                </blockquote>
                <figcaption className="flex flex-col text-sm">
                  <span className="font-medium text-slate-950 dark:text-white">نام و نام خانوادگی</span>
                  <span className="text-xs text-slate-500 dark:text-gray-400">سمت، نام شرکت</span>
                </figcaption>
              </figure>
            </Reveal>
          </div>

          <Reveal className="flex w-full justify-center self-stretch lg:w-1/2">
            <div className="relative w-full max-w-lg self-stretch overflow-hidden rounded-3xl">
              <MediaSlot id="collabPhoto" className="h-full rounded-3xl" />
            </div>
          </Reveal>
        </div>
      </div>
    </section>
  );
}
