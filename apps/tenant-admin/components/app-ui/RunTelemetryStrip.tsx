import * as React from "react";

import { CopyAction } from "@/components/app-ui/CopyAction";
import { cn } from "@/lib/utils";

export type RunTelemetryItem = {
  key: string;
  label: React.ReactNode;
  value: React.ReactNode;
  rawValue?: string;
  technical?: boolean;
};

export type RunTelemetryStripProps = {
  items: RunTelemetryItem[];
  unavailableLabel: string;
  copyLabel: string;
  copiedLabel: string;
  className?: string;
};

export function RunTelemetryStrip({ items, unavailableLabel, copyLabel, copiedLabel, className }: RunTelemetryStripProps) {
  return (
    <div className={cn("@container", className)}>
      <dl className="grid grid-cols-2 gap-px overflow-hidden rounded-lg border border-border bg-border @lg:grid-cols-3">
        {items.map((item) => (
          <div key={item.key} className="min-w-0 bg-card px-3 py-2">
            <dt className="truncate text-xs text-muted-foreground">{item.label}</dt>
            <dd className={cn("mt-1 flex min-w-0 items-center gap-1 text-sm font-medium tabular-nums", item.technical && "font-mono text-xs")}>
              <span className="min-w-0 truncate" dir={item.technical ? "ltr" : undefined}>{item.value ?? unavailableLabel}</span>
              {item.rawValue ? (
                <CopyAction value={item.rawValue} copyLabel={copyLabel} copiedLabel={copiedLabel} size="icon-xs" />
              ) : null}
            </dd>
          </div>
        ))}
      </dl>
    </div>
  );
}
