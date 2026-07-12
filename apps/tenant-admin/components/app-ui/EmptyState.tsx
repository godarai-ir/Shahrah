import * as React from "react";
import type { LucideIcon } from "lucide-react";

import { cn } from "@/lib/utils";

export type EmptyStateVariant = "surface" | "inline" | "table";
export type EmptyStateDensity = "default" | "compact";

export type EmptyStateProps = {
  icon?: LucideIcon;
  title: React.ReactNode;
  description?: React.ReactNode;
  primaryAction?: React.ReactNode;
  secondaryAction?: React.ReactNode;
  variant?: EmptyStateVariant;
  density?: EmptyStateDensity;
  className?: string;
};

const variantClasses: Record<EmptyStateVariant, string> = {
  surface: "min-h-48 rounded-xl border border-dashed border-border bg-card p-6",
  inline: "min-h-32 p-4",
  table: "min-h-40 w-full p-4",
};

const densityClasses: Record<EmptyStateDensity, string> = {
  default: "gap-4",
  compact: "gap-3",
};

export function EmptyState({
  icon: Icon,
  title,
  description,
  primaryAction,
  secondaryAction,
  variant = "surface",
  density = "default",
  className,
}: EmptyStateProps) {
  return (
    <div
      data-variant={variant}
      data-density={density}
      className={cn("flex flex-col items-center justify-center text-center", variantClasses[variant], densityClasses[density], className)}
    >
      {Icon ? (
        <div className={cn("flex items-center justify-center rounded-full bg-muted text-muted-foreground", density === "compact" ? "size-9 [&_svg]:size-4" : "size-11 [&_svg]:size-5")}>
          <Icon aria-hidden="true" />
        </div>
      ) : null}
      <div className="flex max-w-md flex-col gap-1">
        <h3 className={cn("font-semibold text-foreground", density === "compact" ? "text-sm" : "text-base")}>{title}</h3>
        {description ? <p className="text-sm leading-6 text-muted-foreground">{description}</p> : null}
      </div>
      {primaryAction || secondaryAction ? <div className="flex flex-wrap items-center justify-center gap-2">{primaryAction}{secondaryAction}</div> : null}
    </div>
  );
}
