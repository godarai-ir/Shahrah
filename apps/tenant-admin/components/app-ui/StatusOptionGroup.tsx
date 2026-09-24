"use client";

import * as React from "react";
import type { LucideIcon } from "lucide-react";
import { CheckCircle2 } from "lucide-react";

import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import { cn } from "@/lib/utils";

export type StatusOptionTone = "neutral" | "info" | "warning" | "success";

export type StatusOption<T extends string = string> = {
  value: T;
  badge: React.ReactNode;
  description: React.ReactNode;
  icon: LucideIcon;
  title: React.ReactNode;
  tone: StatusOptionTone;
  disabled?: boolean;
};

export type StatusOptionGroupProps<T extends string = string> = {
  label: string;
  options: readonly StatusOption<T>[];
  value: T;
  onValueChange: (value: T) => void;
  columns?: 1 | 2 | 3 | 4;
  orientation?: "horizontal" | "vertical";
  disabled?: boolean;
  dir?: "rtl" | "ltr";
  className?: string;
};

const toneClasses: Record<StatusOptionTone, string> = {
  neutral:
    "data-[state=on]:border-neutral-soft-border data-[state=on]:bg-neutral-soft data-[state=on]:text-neutral-soft-foreground",
  info: "data-[state=on]:border-info-soft-border data-[state=on]:bg-info-soft data-[state=on]:text-info-soft-foreground",
  warning:
    "data-[state=on]:border-warning-soft-border data-[state=on]:bg-warning-soft data-[state=on]:text-warning-soft-foreground",
  success:
    "data-[state=on]:border-success-soft-border data-[state=on]:bg-success-soft data-[state=on]:text-success-soft-foreground",
};

const columnClasses = {
  1: "grid-cols-1",
  2: "grid-cols-1 md:grid-cols-2",
  3: "grid-cols-1 md:grid-cols-2 xl:grid-cols-3",
  4: "grid-cols-1 md:grid-cols-2 xl:grid-cols-4",
} as const;

export function StatusOptionGroup<T extends string>({
  label,
  options,
  value,
  onValueChange,
  columns = 2,
  orientation = "horizontal",
  disabled,
  dir = "rtl",
  className,
}: StatusOptionGroupProps<T>) {
  return (
    <ToggleGroup
      type="single"
      value={value}
      onValueChange={(nextValue) => {
        if (nextValue) onValueChange(nextValue as T);
      }}
      aria-label={label}
      orientation={orientation}
      disabled={disabled}
      dir={dir}
      className={cn("grid w-full items-stretch gap-3", columnClasses[columns], className)}
    >
      {options.map((option) => {
        const Icon = option.icon;
        const selected = value === option.value;

        return (
          <ToggleGroupItem
            key={option.value}
            value={option.value}
            aria-label={typeof option.title === "string" ? option.title : undefined}
            disabled={option.disabled}
            className={cn(
              "group h-auto min-w-0 flex-col items-stretch whitespace-normal rounded-xl border border-border bg-background p-4 text-start",
              "hover:shadow-sm data-[state=off]:hover:border-primary/30 data-[state=off]:hover:bg-muted/30",
              toneClasses[option.tone],
            )}
          >
            <span className="flex items-start justify-between gap-3">
              <span
                className={cn(
                  "flex size-10 shrink-0 items-center justify-center rounded-lg border border-border bg-muted/40 text-muted-foreground transition-colors [&_svg]:size-5",
                  selected && "border-current bg-transparent text-current",
                )}
                aria-hidden="true"
              >
                <Icon />
              </span>
              <span className="flex min-w-0 items-center gap-2">
                {selected ? (
                  <span className="text-primary [&_svg]:size-4" aria-hidden="true">
                    <CheckCircle2 />
                  </span>
                ) : null}
                <StatusBadge status={selected ? "active" : "inactive"} className="px-2 py-0 text-xs">
                  {option.badge}
                </StatusBadge>
              </span>
            </span>
            <span className="mt-3 flex flex-col gap-1">
              <span className="text-sm font-semibold text-foreground">{option.title}</span>
              <span className="text-xs leading-6 text-muted-foreground">{option.description}</span>
            </span>
          </ToggleGroupItem>
        );
      })}
    </ToggleGroup>
  );
}
