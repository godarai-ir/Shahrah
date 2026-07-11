import * as React from "react";

import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

type DataListPanelProps = {
  children: React.ReactNode;
  toolbar?: React.ReactNode;
  toolbarMeta?: React.ReactNode;
  loading?: boolean;
  loadingLabel?: string;
  loadingRows?: number;
  loadingState?: React.ReactNode;
  empty?: boolean;
  emptyState?: React.ReactNode;
  pagination?: React.ReactNode;
  className?: string;
  toolbarClassName?: string;
  contentClassName?: string;
  paginationClassName?: string;
  "aria-label"?: string;
};

export function DataListPanel({
  children,
  toolbar,
  toolbarMeta,
  loading = false,
  loadingLabel,
  loadingRows = 5,
  loadingState,
  empty = false,
  emptyState,
  pagination,
  className,
  toolbarClassName,
  contentClassName,
  paginationClassName,
  "aria-label": ariaLabel,
}: DataListPanelProps) {
  return (
    <section
      data-slot="data-list-panel"
      aria-label={ariaLabel}
      aria-busy={loading || undefined}
      className={cn("flex min-w-0 flex-col gap-4", className)}
    >
      {toolbar || toolbarMeta ? (
        <div
          data-slot="data-list-panel-toolbar"
          className={cn(
            "flex flex-col gap-3 rounded-xl border border-border bg-card p-4 sm:flex-row sm:items-center sm:justify-between",
            toolbarClassName,
          )}
        >
          {toolbar ? <div className="min-w-0 flex-1">{toolbar}</div> : null}
          {toolbarMeta ? <div className="shrink-0 text-sm text-muted-foreground">{toolbarMeta}</div> : null}
        </div>
      ) : null}

      {loading ? (
        loadingState ?? (
          <div
            data-slot="data-list-panel-loading"
            role="status"
            aria-label={loadingLabel}
            className="flex flex-col gap-3 rounded-xl border border-border bg-card p-4"
          >
            {Array.from({ length: loadingRows }, (_, index) => (
              <Skeleton key={index} className="h-10 w-full" />
            ))}
          </div>
        )
      ) : empty ? (
        emptyState ?? null
      ) : (
        <div
          data-slot="data-list-panel-content"
          className={cn("overflow-x-auto rounded-xl border border-border bg-card", contentClassName)}
        >
          {children}
        </div>
      )}

      {pagination ? (
        <div data-slot="data-list-panel-pagination" className={paginationClassName}>
          {pagination}
        </div>
      ) : null}
    </section>
  );
}
