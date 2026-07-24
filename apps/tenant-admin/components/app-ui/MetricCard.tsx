import * as React from "react";
import type { LucideIcon } from "lucide-react";

import { AppCard } from "@/components/app-ui/AppCard";
import { cn } from "@/lib/utils";

export type MetricCardProps = {
  label: React.ReactNode;
  value: React.ReactNode;
  description?: React.ReactNode;
  trend?: React.ReactNode;
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

export function MetricCard({ label, value, description, trend, icon: Icon, className, iconClassName, valueClassName, density = "default", tone = "default" }: MetricCardProps) {
  const compact = density === "compact";
  return (
    <AppCard
      className={cn("border-s-2", toneBorderClasses[tone], className)}
      contentClassName={compact ? "pt-3 md:pt-3" : "pt-4 md:pt-5"}
    >
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0 flex-1">
          <p className="text-sm font-medium text-muted-foreground">{label}</p>
          <div className={cn("mt-2 font-semibold tracking-tight text-foreground", compact ? "text-lg" : "text-2xl", valueClassName)}>{value}</div>
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
