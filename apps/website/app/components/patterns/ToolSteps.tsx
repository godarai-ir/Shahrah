import { Check, Loader2 } from 'lucide-react';
import type { CSSProperties } from 'react';
import type { LucideIcon } from 'lucide-react';
import { iconWell } from '../sectionType';
import { PlayOnView } from './PlayOnView';

export type ToolStep = { icon: LucideIcon; name: string; text: string };

type ToolStepsProps = {
  steps: ToolStep[];
  result: { title: string; text: string };
};

// Tool rows inside a Product Window arrive one by one, then the result resolves once from a spinner to a success check
// (blue success, so the section's single orange stays its primary button).
// The `step-*` classes in theme.css wait for PlayOnView and show the final state under reduced motion.
export function ToolSteps({ steps, result }: ToolStepsProps) {
  const timing = { '--done': `${steps.length * 0.35 + 0.8}s` } as CSSProperties;

  return (
    <PlayOnView>
      <div style={timing}>
        <ul className="m-0 flex list-none flex-col gap-1 p-3">
          {steps.map((step, index) => (
            <li
              key={step.name}
              className="step-row flex items-center gap-3 rounded-card px-3 py-2.5"
              style={{ '--i': index } as CSSProperties}
            >
              <span className={`size-9 ${iconWell}`}>
                <step.icon className="size-4" />
              </span>
              <span className="flex flex-col">
                <span className="text-label font-medium text-site-ink">{step.name}</span>
                <span className="text-caption text-site-ink-soft">{step.text}</span>
              </span>
            </li>
          ))}
        </ul>

        <div className="m-3 mt-1 flex items-center gap-3 rounded-card bg-site-panel px-4 py-3.5">
          <span className="relative flex size-7 shrink-0 items-center justify-center">
            <Loader2 className="step-pending absolute size-5 animate-spin text-site-ink-soft" />
            <span className="step-complete absolute flex size-7 items-center justify-center rounded-full bg-site-success text-site-surface">
              <Check className="size-4" />
            </span>
          </span>
          <span className="flex flex-col">
            <span className="text-label font-medium text-site-ink">{result.title}</span>
            <span className="text-caption text-site-ink-soft">{result.text}</span>
          </span>
        </div>
      </div>
    </PlayOnView>
  );
}
