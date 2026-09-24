"use client";

import * as React from "react";
import { Check } from "lucide-react";

import { useNumberFormatter } from "@/hooks/use-locale-format";
import { cn } from "@/lib/utils";

export type WizardStepState = "complete" | "active" | "idle";

export type WizardStepperStep = {
  id: string;
  title: React.ReactNode;
  /** A short status line under the title, such as the chosen value or what the step covers. */
  description?: React.ReactNode;
  state: WizardStepState;
  disabled?: boolean;
  onSelect?: () => void;
};

export type WizardStepperProps = {
  steps: readonly WizardStepperStep[];
  "aria-label"?: string;
  columns?: 2 | 3 | 4 | 5;
  className?: string;
};

const columnClasses = {
  2: "grid-cols-2",
  3: "grid-cols-2 md:grid-cols-3",
  4: "grid-cols-2 md:grid-cols-4",
  5: "grid-cols-2 md:grid-cols-5",
} as const;

/**
 * The step list of a multi-step flow. Each step is a button with a localized number (or a check
 * once complete), a title, and an optional status line; the active step carries
 * `aria-current="step"` and an underline. Callers own reachability through `disabled`, and wrap the
 * list in a labelled `nav` when it is the flow's main navigation.
 */
export function WizardStepper({ steps, "aria-label": ariaLabel, columns = 4, className }: WizardStepperProps) {
  const numberFormatter = useNumberFormatter();

  return (
    <ol aria-label={ariaLabel} className={cn("grid gap-2", columnClasses[columns], className)}>
      {steps.map((step, index) => {
        const active = step.state === "active";
        const complete = step.state === "complete";
        const selectable = !active && !step.disabled && Boolean(step.onSelect);

        return (
          <li key={step.id} className="min-w-0">
            <button
              type="button"
              aria-current={active ? "step" : undefined}
              disabled={step.disabled && !active}
              onClick={selectable ? step.onSelect : undefined}
              className={cn(
                "group relative flex w-full min-w-0 items-center gap-3 px-1 pt-2 pb-3 text-start focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring",
                selectable ? "cursor-pointer" : "cursor-default",
              )}
            >
              <span
                aria-hidden="true"
                className={cn(
                  "absolute inset-x-0 bottom-0 h-0.5 rounded-full transition-colors",
                  active ? "bg-primary" : complete ? "bg-primary-subtle-border" : "bg-transparent",
                  selectable && "group-hover:bg-primary-subtle-border",
                )}
              />
              <span
                aria-hidden="true"
                className={cn(
                  "flex size-7 shrink-0 items-center justify-center rounded-full text-xs font-semibold tabular-nums ring-1 ring-inset",
                  complete ? "bg-primary text-primary-foreground ring-primary" : active ? "bg-primary-soft text-primary ring-primary" : "bg-card text-muted-foreground ring-input-border",
                )}
              >
                {complete ? <Check className="size-3.5" strokeWidth={3} /> : numberFormatter.format(index + 1)}
              </span>
              <span className="flex min-w-0 flex-col">
                <span className={cn("truncate text-sm", active ? "font-semibold text-foreground" : complete ? "font-medium text-foreground" : "font-medium text-muted-foreground")}>
                  {step.title}
                </span>
                {step.description ? <span className="truncate text-xs text-muted-foreground">{step.description}</span> : null}
              </span>
            </button>
          </li>
        );
      })}
    </ol>
  );
}
