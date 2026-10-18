'use client';

import { ArrowLeft, ArrowRight } from 'lucide-react';
import { useRef } from 'react';
import type { ReactNode } from 'react';

/**
 * Horizontal scroll-snap row with round previous/next buttons. In an RTL page
 * the next item sits to the left, so "next" scrolls by a negative offset.
 */
export function Carousel({ label, children }: { label: string; children: ReactNode }) {
  const trackRef = useRef<HTMLDivElement>(null);

  const scroll = (direction: 'next' | 'previous') => {
    const track = trackRef.current;

    if (!track) {
      return;
    }

    const step = track.clientWidth * 0.8;
    const rtl = getComputedStyle(track).direction === 'rtl';
    const sign = (direction === 'next' ? 1 : -1) * (rtl ? -1 : 1);

    track.scrollBy({ left: sign * step, behavior: 'smooth' });
  };

  return (
    <div>
      <div
        ref={trackRef}
        role="group"
        aria-label={label}
        className="flex snap-x snap-mandatory gap-6 overflow-x-auto pb-2"
        style={{ scrollbarWidth: 'none' }}
      >
        {children}
      </div>
      <div className="mt-8 hidden items-center justify-end gap-3 md:flex">
        <button
          type="button"
          aria-label="قبلی"
          onClick={() => scroll('previous')}
          className="flex size-10 items-center justify-center rounded-full bg-zinc-900/5 text-zinc-900 transition-colors hover:bg-zinc-900/10 dark:bg-white/10 dark:text-white dark:hover:bg-white/20"
        >
          <ArrowRight className="size-4" />
        </button>
        <button
          type="button"
          aria-label="بعدی"
          onClick={() => scroll('next')}
          className="flex size-10 items-center justify-center rounded-full bg-zinc-900/5 text-zinc-900 transition-colors hover:bg-zinc-900/10 dark:bg-white/10 dark:text-white dark:hover:bg-white/20"
        >
          <ArrowLeft className="size-4" />
        </button>
      </div>
    </div>
  );
}
