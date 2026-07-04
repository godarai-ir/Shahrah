import * as React from "react";
import type { LucideIcon } from "lucide-react";

import { AppCard } from "@/components/app-ui/AppCard";
import { cn } from "@/lib/utils";

type MetricCardProps = {
  label: React.ReactNode;
  value: React.ReactNode;
  description?: React.ReactNode;
  trend?: React.ReactNode;
  icon?: LucideIcon;
  className?: string;
  iconClassName?: string;
};

export function MetricCard({ label, value, description, trend, icon: Icon, className, iconClassName }: MetricCardProps) {
  return (
    <AppCard className={className} contentClassName="pt-4 md:pt-6">
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0">
          <p className="text-sm font-medium text-muted-foreground">{label}</p>
          <p className="mt-2 text-2xl font-bold tracking-tight text-foreground md:text-3xl">{value}</p>
          {description ? <p className="mt-1 text-xs leading-6 text-muted-foreground">{description}</p> : null}
          {trend ? <div className="mt-3 text-xs font-medium">{trend}</div> : null}
        </div>
        {Icon ? (
          <div className={cn("flex size-10 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary", iconClassName)}>
            <Icon aria-hidden="true" />
          </div>
        ) : null}
      </div>
    </AppCard>
  );
}
