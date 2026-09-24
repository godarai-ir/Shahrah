"use client";

import * as React from "react";
import type { LucideIcon } from "lucide-react";

import { RadioGroup, RadioGroupCard } from "@/components/ui/radio-group";
import { cn } from "@/lib/utils";

export type OptionCardItem<T extends string = string> = {
  value: T;
  title: React.ReactNode;
  description?: React.ReactNode;
  /** A short emphasized note under the description, such as when to prefer this option. */
  tip?: React.ReactNode;
  icon?: LucideIcon;
  disabled?: boolean;
};

export type OptionCardGroupProps<T extends string = string> = {
  options: readonly OptionCardItem<T>[];
  value: T;
  onValueChange: (value: T) => void;
  "aria-label"?: string;
  "aria-labelledby"?: string;
  columns?: 1 | 2 | 3 | 4;
  density?: "default" | "compact";
  dir?: "rtl" | "ltr";
  disabled?: boolean;
  className?: string;
};

const columnClasses = {
  1: "grid-cols-1",
  2: "grid-cols-1 sm:grid-cols-2",
  3: "grid-cols-1 sm:grid-cols-2 xl:grid-cols-3",
  4: "grid-cols-1 sm:grid-cols-2 xl:grid-cols-4",
} as const;

/**
 * A single-choice group of descriptive cards. Each card is a radio inside a radiogroup, so arrow
 * keys move the selection and the title names the option while the description describes it.
 */
export function OptionCardGroup<T extends string>({
  options,
  value,
  onValueChange,
  "aria-label": ariaLabel,
  "aria-labelledby": ariaLabelledBy,
  columns = 2,
  density = "default",
  dir,
  disabled,
  className,
}: OptionCardGroupProps<T>) {
  const baseID = React.useId();

  return (
    <RadioGroup
      value={value}
      onValueChange={(next) => onValueChange(next as T)}
      aria-label={ariaLabel}
      aria-labelledby={ariaLabelledBy}
      dir={dir}
      disabled={disabled}
      className={cn("grid items-stretch gap-3", columnClasses[columns], className)}
    >
      {options.map((option) => {
        const Icon = option.icon;
        const titleID = `${baseID}-${option.value}-title`;
        const descriptionID = option.description || option.tip ? `${baseID}-${option.value}-description` : undefined;

        return (
          <RadioGroupCard
            key={option.value}
            value={option.value}
            disabled={option.disabled}
            aria-labelledby={titleID}
            aria-describedby={descriptionID}
            className={cn(
              "group flex min-w-0 items-start gap-3 rounded-xl border border-border bg-card text-start transition-colors",
              "hover:border-primary-subtle-border data-[state=checked]:border-primary data-[state=checked]:bg-primary-soft",
              "data-[disabled]:hover:border-border",
              density === "compact" ? "p-3" : "p-4",
            )}
          >
            {Icon ? (
              <span
                aria-hidden="true"
                className={cn(
                  "flex shrink-0 items-center justify-center rounded-lg bg-muted text-muted-foreground group-data-[state=checked]:text-primary",
                  density === "compact" ? "size-8 [&_svg]:size-4" : "size-9 [&_svg]:size-4",
                )}
              >
                <Icon />
              </span>
            ) : null}
            <span className="flex min-w-0 flex-col gap-1">
              <span id={titleID} className="text-sm font-semibold text-foreground">{option.title}</span>
              {descriptionID ? (
                <span id={descriptionID} className="flex flex-col gap-1">
                  {option.description ? <span className="text-xs leading-5 text-muted-foreground">{option.description}</span> : null}
                  {option.tip ? <span className="text-xs leading-5 text-primary">{option.tip}</span> : null}
                </span>
              ) : null}
            </span>
          </RadioGroupCard>
        );
      })}
    </RadioGroup>
  );
}
