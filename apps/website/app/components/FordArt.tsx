// Authored water and stone forms for the home page. Drawn paths, not masks: each
// line drifts a little differently and each pebble has its own weight.

export function WaterLines({ className = '' }: { className?: string }) {
  return (
    <svg
      aria-hidden="true"
      viewBox="0 0 800 600"
      preserveAspectRatio="none"
      className={`pointer-events-none absolute inset-0 size-full ${className}`}
      fill="none"
    >
      <g stroke="currentColor" strokeWidth="1.5" strokeLinecap="round">
        <path d="M0 96 C 120 76, 230 118, 350 98 S 570 74, 690 102 S 780 112, 800 106" />
        <path d="M0 176 C 90 198, 240 154, 370 180 S 590 206, 700 174 S 790 168, 800 172" opacity="0.85" />
        <path d="M0 262 C 130 240, 210 286, 340 264 S 560 238, 690 268 S 780 280, 800 274" opacity="0.7" />
        <path d="M0 352 C 100 372, 250 330, 380 356 S 600 384, 710 350 S 790 346, 800 350" opacity="0.55" />
        <path d="M0 440 C 140 420, 220 462, 350 442 S 570 420, 700 448 S 780 456, 800 452" opacity="0.4" />
      </g>
    </svg>
  );
}

/** A thin, uneven water line used as a quiet break between sections. */
export function WaterEdge({ className = '' }: { className?: string }) {
  return (
    <svg aria-hidden="true" viewBox="0 0 1200 28" preserveAspectRatio="none" className={`h-7 w-full ${className}`} fill="none">
      <path
        d="M0 14 C 90 4, 180 24, 290 13 S 470 3, 590 15 S 800 26, 910 12 S 1110 6, 1200 15"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
      />
    </svg>
  );
}

/**
 * A small set of stones. Each has an irregular contour, a lit top and a soft
 * grounded shadow, so they read as weight resting in shallow water.
 */
export function Pebbles({ className = '' }: { className?: string }) {
  const stones = [
    { body: 'M14 62 C 8 40, 40 22, 78 30 C 112 37, 124 62, 100 76 C 70 90, 20 84, 14 62 Z', lit: 'M30 48 C 38 38, 62 34, 80 40 C 92 44, 96 50, 86 54 C 68 52, 44 54, 30 48 Z', shadow: [68, 86, 54, 7] },
    { body: 'M148 74 C 146 58, 168 50, 190 55 C 210 60, 214 78, 198 86 C 178 94, 150 90, 148 74 Z', lit: 'M160 66 C 166 60, 182 58, 194 62 C 198 65, 194 68, 186 68 C 176 68, 166 70, 160 66 Z', shadow: [180, 90, 34, 5] },
    { body: 'M236 80 C 236 70, 250 65, 262 68 C 274 71, 276 82, 266 87 C 254 92, 236 90, 236 80 Z', lit: 'M244 76 C 248 72, 258 71, 264 73 C 266 75, 262 77, 256 77 C 250 77, 246 78, 244 76 Z', shadow: [256, 91, 20, 3] },
  ];

  return (
    <svg aria-hidden="true" viewBox="0 0 290 100" className={`pointer-events-none ${className}`} fill="none">
      <defs>
        <filter id="water-soft" x="-30%" y="-80%" width="160%" height="260%">
          <feGaussianBlur stdDeviation="3" />
        </filter>
      </defs>
      {stones.map((stone, index) => (
        <g key={index}>
          <ellipse cx={stone.shadow[0]} cy={stone.shadow[1]} rx={stone.shadow[2]} ry={stone.shadow[3]} className="fill-site-ink" opacity="0.14" filter="url(#water-soft)" />
          <path d={stone.body} className="fill-site-panel-deep" />
          <path d={stone.lit} className="fill-site-ground" opacity="0.75" />
        </g>
      ))}
    </svg>
  );
}

export function Ripples({ className = '' }: { className?: string }) {
  return (
    <div
      aria-hidden="true"
      className={`pointer-events-none absolute inset-0 flex items-center justify-center ${className}`}
      style={{ containerType: 'size' }}
    >
      {[0, 1, 2].map((index) => (
        <span
          key={index}
          className="water-ripple absolute aspect-square rounded-full border border-site-accent"
          style={{ width: `${58 + index * 14}cqmin`, animationDelay: `${2.6 + index * 0.9}s` }}
        />
      ))}
    </div>
  );
}
