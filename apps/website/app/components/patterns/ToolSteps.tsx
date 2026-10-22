import { Check, Loader2 } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';

export type ToolStep = { icon: LucideIcon; name: string; text: string };

type ToolStepsProps = {
  steps: ToolStep[];
  result: { title: string; text: string };
};

// Tool rows inside a Product Window, then a result that resolves once from a spinner to a check.
// The `step-*` classes in theme.css play it once and show the final state under reduced motion.
export function ToolSteps({ steps, result }: ToolStepsProps) {
  return (
    <>
      <ul className="flex flex-col gap-1 p-3">
        {steps.map((step) => (
          <li key={step.name} className="flex items-center gap-3 rounded-2xl px-3 py-2.5">
            <span className="flex size-9 shrink-0 items-center justify-center rounded-xl bg-site-panel-deep text-site-accent">
              <step.icon className="size-4" />
            </span>
            <span className="flex flex-col">
              <span className="text-sm font-medium text-site-ink">{step.name}</span>
              <span className="text-xs text-site-ink-soft">{step.text}</span>
            </span>
          </li>
        ))}
      </ul>

      <div className="m-3 mt-1 flex items-center gap-3 rounded-2xl bg-site-panel px-4 py-3.5">
        <span className="relative flex size-7 shrink-0 items-center justify-center">
          <Loader2 className="step-pending absolute size-5 animate-spin text-site-ink-soft" />
          <span className="step-complete absolute flex size-7 items-center justify-center rounded-full bg-site-accent text-site-on-accent">
            <Check className="size-4" />
          </span>
        </span>
        <span className="flex flex-col">
          <span className="text-sm font-medium text-site-ink">{result.title}</span>
          <span className="text-xs text-site-ink-soft">{result.text}</span>
        </span>
      </div>
    </>
  );
}
