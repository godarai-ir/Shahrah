'use client';

import { useRef, useState } from 'react';
import type { MouseEvent } from 'react';
import { Eyebrow, sectionCopy, sectionHeading } from './Eyebrow';
import { MediaSlot } from './MediaSlot';
import { Reveal } from './Reveal';
import type { HomeMediaId } from '@/app/lib/home-media';

type Pillar = {
  number: string;
  title: string;
  body: string;
  media: HomeMediaId;
  numberColor: string;
  tools?: string[];
};

const pillars: Pillar[] = [
  {
    number: '۰۱',
    title: 'دستیار با ابزارهای خودتان کار می‌کند',
    body: 'CRM، پیام‌رسان، تسک، سند و کد؛ هر چیزی که تیم‌تان هر روز باز می‌کند.',
    media: 'pillarContext',
    numberColor: 'text-blue-500',
    tools: ['بله', 'ایتا', 'واتساپ', 'تلگرام', 'CRM', 'Jira', 'Notion', 'Git', 'نشان', 'MCP'],
  },
  {
    number: '۰۲',
    title: 'آخرین حرف با شماست',
    body: 'کارهای حساس را دستیار آماده می‌کند و منتظر می‌ماند. فرستادن یا نفرستادنش تصمیم همکار شماست.',
    media: 'pillarHuman',
    numberColor: 'text-rose-500',
  },
  {
    number: '۰۳',
    title: 'مدل را خودتان انتخاب می‌کنید',
    body: 'بیش از ۱۲۰۰ مدل از ۲۱ ارائه‌دهنده. برای هر کار مدل مناسبش را بگذارید و اگر خوب جواب نداد، عوضش کنید.',
    media: 'pillarModels',
    numberColor: 'text-amber-500',
  },
  {
    number: '۰۴',
    title: 'چه کسی به چه چیزی برسد، شما می‌گویید',
    body: 'هر دستیار فقط به چیزهایی دست می‌زند که به او سپرده‌اید. هر تیم سقف مصرف دارد و هر درخواست ثبت می‌شود.',
    media: 'pillarControl',
    numberColor: 'text-green-600 dark:text-green-400',
  },
];

export function Pillars() {
  const gridRef = useRef<HTMLDivElement>(null);
  const [active, setActive] = useState<number | null>(null);
  const [point, setPoint] = useState({ x: 0, y: 0 });

  const track = (event: MouseEvent<HTMLDivElement>) => {
    const box = gridRef.current?.getBoundingClientRect();

    if (box) {
      setPoint({ x: event.clientX - box.left, y: event.clientY - box.top });
    }
  };

  return (
    <section id="platform" className="w-full bg-white py-14 dark:bg-slate-950 lg:py-24">
      <div className="mx-auto flex w-full max-w-6xl flex-col gap-16 px-6">
        <div className="flex flex-col items-center gap-6 text-center">
          <Reveal>
            <Eyebrow>چهار چیز که مهم است</Eyebrow>
          </Reveal>
          <Reveal>
            <h2 className={`max-w-3xl ${sectionHeading}`}>هوش مصنوعی شرکت را یک‌جا اداره کنید</h2>
          </Reveal>
          <Reveal>
            <p className={`max-w-2xl ${sectionCopy}`}>
              ابزار و مدل و تیم، همه سر یک میز. هر تیم دستیار خودش را می‌سازد و کنترل همچنان دست شماست.
            </p>
          </Reveal>
        </div>

        <div
          ref={gridRef}
          onMouseMove={track}
          onMouseLeave={() => setActive(null)}
          className="relative grid grid-cols-1 gap-px overflow-hidden rounded-3xl bg-gray-200 dark:bg-gray-800 md:grid-cols-2"
        >
          {pillars.map((pillar, index) => (
            <article
              key={pillar.number}
              onMouseEnter={() => setActive(index)}
              className="group relative z-0 flex flex-col gap-6 overflow-hidden bg-white p-8 transition-colors duration-200 hover:bg-blue-50/60 dark:bg-slate-950 dark:hover:bg-blue-500/10 md:p-10"
            >
              <div className="flex items-center gap-4">
                <span className={`inline-block text-2xl font-medium transition-transform duration-200 group-hover:scale-110 ${pillar.numberColor}`}>
                  {pillar.number}
                </span>
                <span className="h-px flex-1 bg-gray-200 dark:bg-gray-800" />
              </div>
              <div className="flex flex-col gap-2">
                <h3 className="text-xl font-semibold leading-snug text-slate-950 dark:text-white md:text-2xl">{pillar.title}</h3>
                <p className="max-w-md text-sm leading-7 text-slate-600 dark:text-gray-400">{pillar.body}</p>
                {pillar.tools ? (
                  <ul className="mt-2 flex flex-wrap gap-2">
                    {pillar.tools.map((tool) => (
                      <li
                        key={tool}
                        className="rounded-lg bg-slate-100 px-2.5 py-1 text-xs text-slate-700 dark:bg-slate-800 dark:text-gray-300"
                      >
                        {tool}
                      </li>
                    ))}
                  </ul>
                ) : null}
              </div>
              <div className="md:hidden">
                <MediaSlot id={pillar.media} />
              </div>
            </article>
          ))}

          <div
            aria-hidden="true"
            className={`pointer-events-none absolute left-0 top-0 z-10 hidden w-72 transition-opacity duration-200 md:block ${
              active === null ? 'opacity-0' : 'opacity-100'
            }`}
            style={{ transform: `translate(${point.x - 144}px, ${point.y + 24}px)` }}
          >
            <div className="overflow-hidden rounded-2xl bg-white shadow-xl dark:bg-slate-900">
              <MediaSlot id={pillars[active ?? 0].media} />
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
