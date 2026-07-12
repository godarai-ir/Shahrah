import * as React from "react";
import type { LucideIcon } from "lucide-react";

import { cn } from "@/lib/utils";

type SectionHeaderProps = {
  title: React.ReactNode;
  description?: React.ReactNode;
  action?: React.ReactNode;
  icon?: LucideIcon;
  density?: "default" | "compact";
  className?: string;
};

export function SectionHeader({ title, description, action, icon: Icon, density = "default", className }: SectionHeaderProps) {
  const compact = density === "compact";
  return (
    <div className={cn("flex gap-3 sm:items-start sm:justify-between", compact ? "items-start" : "flex-col sm:flex-row", className)}>
      <div className="flex min-w-0 items-start gap-2">
        {Icon ? <Icon className="mt-0.5 size-4 shrink-0 text-muted-foreground" aria-hidden="true" /> : null}
        <div className="min-w-0">
          <h2 className={cn("font-semibold text-foreground", compact ? "text-sm md:text-base" : "text-base md:text-lg")}>{title}</h2>
          {description ? <p className={cn("mt-1 text-muted-foreground", compact ? "text-xs leading-normal" : "text-sm leading-6")}>{description}</p> : null}
        </div>
      </div>
      {action ? <div className="flex shrink-0 items-center gap-2">{action}</div> : null}
    </div>
  );
}
