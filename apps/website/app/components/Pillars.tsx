'use client';

import { useRef, useState } from 'react';
import type { MouseEvent } from 'react';
import { MediaSlot } from './MediaSlot';
import { Section } from './patterns';
import { sectionCopy, sectionHeading } from './sectionType';
import type { HomeMediaId } from '@/app/lib/home-media';

type Pillar = {
  number: string;
  title: string;
  body: string;
  media: HomeMediaId;
  tools?: string[];
};

const pillars: Pillar[] = [
  {
    number: '۰۱',
    title: 'دستیار با ابزارهای خودتان کار می‌کند',
    body: 'CRM، پیام‌رسان، تسک، سند و کد؛ هر چیزی که تیم‌تان هر روز باز می‌کند.',
    media: 'pillarContext',
    tools: ['بله', 'ایتا', 'واتساپ', 'تلگرام', 'CRM', 'Jira', 'Notion', 'Git', 'نشان', 'MCP'],
  },
  {
    number: '۰۲',
    title: 'آخرین حرف با شماست',
    body: 'کارهای حساس را دستیار آماده می‌کند و منتظر می‌ماند. فرستادن یا نفرستادنش تصمیم همکار شماست.',
    media: 'pillarHuman',
  },
  {
    number: '۰۳',
    title: 'مدل را خودتان انتخاب می‌کنید',
    body: 'بیش از ۱۲۰۰ مدل از ۲۱ ارائه‌دهنده. برای هر کار مدل مناسبش را بگذارید و اگر خوب جواب نداد، عوضش کنید.',
    media: 'pillarModels',
  },
  {
    number: '۰۴',
    title: 'چه کسی به چه چیزی برسد، شما می‌گویید',
    body: 'هر دستیار فقط به چیزهایی دست می‌زند که به او سپرده‌اید. هر تیم سقف مصرف دارد و هر درخواست ثبت می‌شود.',
    media: 'pillarControl',
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
    <Section id="platform" layout="flex flex-col gap-16">
      <div className="flex flex-col items-center gap-6 text-center">
        <div>
          <h2 className={`max-w-3xl ${sectionHeading}`}>هوش مصنوعی شرکت را یک‌جا اداره کنید</h2>
        </div>
        <div>
          <p className={`max-w-2xl ${sectionCopy}`}>
            ابزار و مدل و تیم، همه سر یک میز. هر تیم دستیار خودش را می‌سازد و کنترل همچنان دست شماست.
          </p>
        </div>
      </div>

      <div
        ref={gridRef}
        onMouseMove={track}
        onMouseLeave={() => setActive(null)}
        className="relative grid grid-cols-1 gap-px overflow-hidden rounded-3xl bg-site-line md:grid-cols-2"
      >
        {pillars.map((pillar, index) => (
          <article
            key={pillar.number}
            onMouseEnter={() => setActive(index)}
            className="group relative z-0 flex flex-col gap-6 overflow-hidden bg-site-ground p-8 transition-colors duration-500 hover:bg-site-panel md:p-12"
          >
            <div className="flex items-center gap-4">
              <span className="text-xl font-medium text-site-accent">{pillar.number}</span>
              <span className="h-px flex-1 bg-site-line" />
            </div>
            <div className="flex flex-col gap-3">
              <h3 className="text-xl font-semibold leading-snug text-site-ink md:text-2xl">{pillar.title}</h3>
              <p className="max-w-md text-base leading-8 text-site-ink-soft">{pillar.body}</p>
              {pillar.tools ? (
                <ul className="mt-2 flex flex-wrap gap-2">
                  {pillar.tools.map((tool) => (
                    <li key={tool} className="rounded-full bg-site-panel px-3 py-1 text-xs text-site-ink-soft group-hover:bg-site-panel-deep">
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
          className={`pointer-events-none absolute left-0 top-0 z-10 hidden w-72 transition-opacity duration-500 md:block ${
            active === null ? 'opacity-0' : 'opacity-100'
          }`}
          style={{ transform: `translate(${point.x - 144}px, ${point.y + 24}px)` }}
        >
          <div className="overflow-hidden rounded-2xl bg-site-ground shadow-window">
            <MediaSlot id={pillars[active ?? 0].media} />
          </div>
        </div>
      </div>
    </Section>
  );
}
