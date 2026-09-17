import * as React from "react";
import { ArrowDownRight, ArrowUpRight, Minus, type LucideIcon } from "lucide-react";

import { AppCard } from "@/components/app-ui/AppCard";
import { cn } from "@/lib/utils";

export type MetricCardDelta = {
  /** Omit to render the label without an arrow (e.g. "12 days left"). */
  direction?: "up" | "down" | "flat";
  label: React.ReactNode;
  tone: "positive" | "negative" | "muted";
};

export type MetricCardProps = {
  label: React.ReactNode;
  value: React.ReactNode;
  description?: React.ReactNode;
  trend?: React.ReactNode;
  delta?: MetricCardDelta;
  sparkline?: React.ReactNode;
  icon?: LucideIcon;
  className?: string;
  iconClassName?: string;
  valueClassName?: string;
  density?: "default" | "compact";
  tone?: "default" | "success" | "warning" | "muted";
};

const toneClasses = {
  default: "border-primary-subtle-border bg-primary-soft text-primary-soft-foreground",
  success: "border-success-soft-border bg-success-soft text-success-soft-foreground",
  warning: "border-warning-soft-border bg-warning-soft text-warning-soft-foreground",
  muted: "border-neutral-soft-border bg-neutral-soft text-neutral-soft-foreground",
} as const;

const toneBorderClasses = {
  default: "border-s-primary",
  success: "border-s-success",
  warning: "border-s-warning",
  muted: "border-s-border",
} as const;

const deltaToneClasses = {
  positive: "border-success-soft-border bg-success-soft text-success-soft-foreground",
  negative: "border-destructive-soft-border bg-destructive-soft text-destructive-soft-foreground",
  muted: "border-neutral-soft-border bg-neutral-soft text-neutral-soft-foreground",
} as const;

const deltaIcons = { up: ArrowUpRight, down: ArrowDownRight, flat: Minus } as const;

export function MetricCard({ label, value, description, trend, delta, sparkline, icon: Icon, className, iconClassName, valueClassName, density = "default", tone = "default" }: MetricCardProps) {
  const compact = density === "compact";
  return (
    <AppCard
      className={cn("border-s-2", toneBorderClasses[tone], className)}
      contentClassName={compact ? "pt-3 md:pt-3" : "pt-4 md:pt-5"}
    >
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0 flex-1">
          <div className="flex items-start justify-between gap-2">
            <p className="text-sm font-medium text-muted-foreground">{label}</p>
            {delta ? <MetricDelta {...delta} /> : null}
          </div>
          <div className={cn("mt-2 font-semibold tracking-tight text-foreground", compact ? "text-lg" : "text-2xl", valueClassName)}>{value}</div>
          {sparkline ? <div className="mt-2">{sparkline}</div> : null}
          {description ? <p className="mt-1 text-xs leading-6 text-muted-foreground">{description}</p> : null}
          {trend ? <div className="mt-3 text-xs font-medium">{trend}</div> : null}
        </div>
        {Icon ? (
          <div className={cn("flex size-10 shrink-0 items-center justify-center rounded-lg border", toneClasses[tone], iconClassName)}>
            <Icon aria-hidden="true" />
          </div>
        ) : null}
      </div>
    </AppCard>
  );
}

function MetricDelta({ direction, label, tone }: MetricCardDelta) {
  const Icon = direction ? deltaIcons[direction] : null;
  return (
    <span
      data-slot="metric-delta"
      data-tone={tone}
      className={cn("inline-flex shrink-0 items-center gap-1 rounded-full border px-2 py-0.5 text-xs font-medium tabular-nums", deltaToneClasses[tone])}
    >
      {Icon ? <Icon className="size-3" aria-hidden="true" /> : null}
      {label}
    </span>
  );
}
