'use client';

import { useEffect, useRef, useState } from 'react';
import type { ReactNode } from 'react';

// Holds the `step-*` animations in theme.css paused until the block scrolls into view, then lets them run once.
export function PlayOnView({ children }: { children: ReactNode }) {
  const ref = useRef<HTMLDivElement>(null);
  const [inView, setInView] = useState(false);

  useEffect(() => {
    const node = ref.current;

    if (!node || inView) {
      return;
    }

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setInView(true);
          observer.disconnect();
        }
      },
      { threshold: 0.4 },
    );

    observer.observe(node);

    return () => observer.disconnect();
  }, [inView]);

  return (
    <div ref={ref} data-in-view={inView ? '' : undefined}>
      {/* Without script nothing would ever mark the block in view, so show the finished state instead. */}
      <noscript>
        <style>{'.step-row,.step-complete{animation:none!important}.step-pending{animation:none!important;opacity:0}'}</style>
      </noscript>
      {children}
    </div>
  );
}
