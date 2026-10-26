'use client';

import { useId, useState } from 'react';
import type { KeyboardEvent } from 'react';
import { ArrowLeft, Check } from 'lucide-react';
import Link from 'next/link';
import { isReady } from '@/app/lib/navigation';
import { Section } from './patterns';
import { sectionCopy, sectionHeading, textLink } from './sectionType';
import { MediaSlot } from './MediaSlot';
import { teams } from '@/app/lib/teams';

export function TeamTabs() {
  const baseId = useId();
  const [activeId, setActiveId] = useState(teams[0].id);
  const active = teams.find((team) => team.id === activeId) ?? teams[0];

  // Automatic activation; in RTL the left arrow moves forward in reading order.
  const moveFocus = (event: KeyboardEvent<HTMLDivElement>) => {
    const index = teams.findIndex((team) => team.id === activeId);
    const steps: Record<string, number> = { ArrowLeft: 1, ArrowRight: -1 };
    let next: number | null = null;

    if (event.key in steps) {
      next = (index + steps[event.key] + teams.length) % teams.length;
    } else if (event.key === 'Home') {
      next = 0;
    } else if (event.key === 'End') {
      next = teams.length - 1;
    }

    if (next === null) {
      return;
    }

    event.preventDefault();
    setActiveId(teams[next].id);
    document.getElementById(`${baseId}-tab-${teams[next].id}`)?.focus();
  };

  return (
    <Section id="teams" tone="panel" layout="flex flex-col gap-8">
      <div className="flex flex-col items-center gap-4 text-center">
        <div>
          <h2 className={`max-w-3xl ${sectionHeading}`}>هر تیم، دستیار خودش. نه یک چت‌بات برای همه.</h2>
        </div>
        <div>
          <p className={`m-0 max-w-xl ${sectionCopy}`}>
            چند کاری که معمولاً اول به دستیار سپرده می‌شود.
          </p>
        </div>
      </div>

      <div>
        <div
          role="tablist"
          aria-label="تیم‌ها"
          onKeyDown={moveFocus}
          className="relative flex w-full overflow-x-auto overflow-y-hidden border-b border-site-line"
        >
          {teams.map((team) => {
            const selected = team.id === activeId;

            return (
              <button
                key={team.id}
                type="button"
                role="tab"
                id={`${baseId}-tab-${team.id}`}
                aria-selected={selected}
                tabIndex={selected ? 0 : -1}
                aria-controls={`${baseId}-panel`}
                onClick={() => setActiveId(team.id)}
                className={`-mb-px flex-1 whitespace-nowrap border-b-2 px-4 py-4 text-body transition-colors duration-150 focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-site-focus md:text-lead ${
                  selected
                    ? 'border-site-ink font-semibold text-site-ink'
                    : 'border-transparent font-medium text-site-ink-soft hover:text-site-ink'
                }`}
              >
                {team.label}
              </button>
            );
          })}
        </div>
      </div>

      <div
        role="tabpanel"
        id={`${baseId}-panel`}
        aria-labelledby={`${baseId}-tab-${active.id}`}
        className="grid grid-cols-1 gap-8 lg:grid-cols-2 lg:gap-16"
      >
        <div className="relative">
          <MediaSlot key={active.media} id={active.media} className="rounded-card lg:aspect-auto lg:h-full lg:min-h-96" />
        </div>
        <div className="flex flex-col gap-6 lg:gap-8 lg:py-4">
          <h3 className="m-0 whitespace-pre-line text-h2 font-bold text-site-ink lg:text-h1">
            {`تیم ${active.label}\nبا دستیار گدارAI`}
          </h3>
          <ul className="m-0 flex list-none flex-col p-0">
            {active.tasks.map((task, index) => (
              <li
                key={task.title}
                className={`flex items-start gap-4 py-4 ${index === 0 ? '' : 'border-t border-site-line'}`}
              >
                <span className="flex h-7 items-center pt-px">
                  <Check aria-hidden="true" className="size-5 text-site-ink" />
                </span>
                <div className="flex flex-col gap-1.5">
                  <span className="text-body font-semibold text-site-ink">{task.title}</span>
                  <span className="text-label text-site-ink-soft">{task.text}</span>
                </div>
              </li>
            ))}
          </ul>
          {isReady(`/teams/${active.id}`) ? (
            <Link
              href={`/teams/${active.id}`}
              className={textLink}
            >
              {`دیدن همه‌ی کارهای تیم ${active.label}`}
              <ArrowLeft aria-hidden="true" className="size-4 transition-transform duration-150 group-hover/link:-translate-x-0.5" />
            </Link>
          ) : null}
        </div>
      </div>
    </Section>
  );
}
