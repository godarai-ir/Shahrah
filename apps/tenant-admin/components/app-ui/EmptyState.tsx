import * as React from "react";
import type { LucideIcon } from "lucide-react";

import { cn } from "@/lib/utils";

type EmptyStateProps = {
  icon?: LucideIcon;
  title: React.ReactNode;
  description?: React.ReactNode;
  primaryAction?: React.ReactNode;
  secondaryAction?: React.ReactNode;
  className?: string;
};

export function EmptyState({ icon: Icon, title, description, primaryAction, secondaryAction, className }: EmptyStateProps) {
  return (
    <div className={cn("flex min-h-48 flex-col items-center justify-center gap-4 rounded-xl border border-dashed border-border bg-card p-6 text-center", className)}>
      {Icon ? (
        <div className="flex size-11 items-center justify-center rounded-full bg-muted text-muted-foreground">
          <Icon aria-hidden="true" />
        </div>
      ) : null}
      <div className="flex max-w-md flex-col gap-1">
        <h3 className="text-base font-semibold text-foreground">{title}</h3>
        {description ? <p className="text-sm leading-6 text-muted-foreground">{description}</p> : null}
      </div>
      {primaryAction || secondaryAction ? <div className="flex flex-wrap items-center justify-center gap-2">{primaryAction}{secondaryAction}</div> : null}
    </div>
  );
}
