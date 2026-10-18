import { MediaSlot } from '../MediaSlot';
import { Reveal } from '../Reveal';
import { Carousel } from './Carousel';
import { CircleLink, sectionPad, TwoTone } from './parts';
import type { HomeMediaId } from '@/app/lib/home-media';

const stories: { media: HomeMediaId; title: string }[] = [
  { media: 'storyOne', title: 'عنوان تجربه‌ی مشتری ۱' },
  { media: 'storyTwo', title: 'عنوان تجربه‌ی مشتری ۲' },
  { media: 'storyThree', title: 'عنوان تجربه‌ی مشتری ۳' },
];

export function WorkplaceGuide() {
  return (
    <section className="overflow-hidden">
      <div className={`${sectionPad} pb-20 md:pb-24 lg:pb-40`}>
        <div className="mx-auto w-full max-w-7xl">
          <Reveal className="mb-8 md:mb-12 lg:mb-16">
            <TwoTone lead="قدم‌به‌قدم کنارتان هستیم." muted="ببینید گدارAI چطور در سازمان شما کار می‌کند." />
          </Reveal>

          <div className="grid grid-cols-1 gap-8 md:grid-cols-[1fr_2fr] lg:gap-10">
            <Reveal className="flex">
              <article className="relative flex min-h-100 w-full flex-col overflow-hidden rounded-xl bg-zinc-200 p-8 text-zinc-900 dark:bg-white/10 dark:text-white sm:p-10 md:h-144">
                <h3 className="mb-2 text-lg font-medium">راهنمای شروع</h3>
                <p className="mb-6 text-base leading-8 opacity-75">از ساختن فضای کاری تا اولین دستیار، قدم‌به‌قدم.</p>
                <div className="mb-6">
                  <CircleLink href="/docs">مشاهده‌ی مستندات</CircleLink>
                </div>
                <div className="min-h-0 flex-1">
                  <MediaSlot id="guideHandbook" className="h-full" />
                </div>
              </article>
            </Reveal>

            <Reveal className="flex">
              <article className="relative flex min-h-100 w-full flex-col overflow-hidden rounded-xl bg-stone-800 p-8 text-white sm:p-10 md:h-144">
                <h3 className="mb-2 text-lg font-medium">گدارAI در عمل</h3>
                <p className="mb-6 max-w-md text-base leading-8 opacity-75">
                  یک جلسه‌ی خصوصی با تیم ما رزرو کنید تا نمونه‌ی خودتان را با هم بسازیم.
                </p>
                <div className="mb-6">
                  <CircleLink href="/demo" tone="light">
                    رزرو وقت دمو
                  </CircleLink>
                </div>
                <div className="min-h-0 flex-1">
                  <MediaSlot id="guideTour" className="h-full" />
                </div>
              </article>
            </Reveal>
          </div>
        </div>
      </div>

      <div className={`${sectionPad} pb-20 md:pb-24 lg:pb-40`}>
        <div className="mx-auto w-full max-w-7xl">
          <Reveal className="mb-8 md:mb-12 lg:mb-16">
            <TwoTone lead="تجربه‌ی مشتری‌ها." muted="ببینید چطور هوش مصنوعی را به کارشان آورده‌اند." />
          </Reveal>

          <Carousel label="تجربه‌ی مشتری‌ها">
            {stories.map((story) => (
              <article
                key={story.media}
                className="relative flex h-128 w-80 shrink-0 snap-start flex-col overflow-hidden rounded-xl bg-white p-8 text-zinc-900 dark:bg-white/5 dark:text-white sm:w-96"
              >
                <h3 className="mb-6 text-lg font-medium">{story.title}</h3>
                <div className="min-h-0 flex-1">
                  <MediaSlot id={story.media} className="h-full" />
                </div>
              </article>
            ))}
          </Carousel>
        </div>
      </div>
    </section>
  );
}
