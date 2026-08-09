import type { ReactNode } from "react";

import { FormSection } from "@/components/app-ui/FormSection";
import { SectionHeader } from "@/components/app-ui/SectionHeader";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

export function CatalogCreateSection({
  title,
  description,
  children,
  footer,
  actions,
  id,
  setSectionNode,
  fillHeight = false,
  scrollable = false,
  contentClassName,
}: {
  title: string;
  description?: string;
  children: ReactNode;
  footer?: ReactNode;
  actions?: ReactNode;
  id?: string;
  setSectionNode?: (node: HTMLElement | null) => void;
  fillHeight?: boolean;
  scrollable?: boolean;
  contentClassName?: string;
}) {
  return (
    <section id={id} ref={setSectionNode} className={cn("min-w-0", fillHeight && "flex min-h-0 flex-1 flex-col")}>
      <FormSection footer={footer} className={cn(fillHeight && "min-h-0 flex-1", scrollable && "overflow-y-auto")}>
        {title || description || actions ? (
          <SectionHeader title={title} description={description} action={actions} />
        ) : null}
        <div className={cn("min-w-0", contentClassName)}>{children}</div>
      </FormSection>
    </section>
  );
}

export function ProviderDetailsSkeleton({ label }: { label: string }) {
  return (
    <div className="flex flex-col gap-5" role="status" aria-label={label}>
      <div className="flex flex-col gap-3 rounded-xl border border-border bg-card p-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="flex min-w-0 items-center gap-3">
          <Skeleton className="size-10 shrink-0 rounded-lg" />
          <div className="flex flex-col gap-2">
            <Skeleton className="h-5 w-48 max-w-full" />
            <Skeleton className="h-4 w-64 max-w-full" />
          </div>
        </div>
        <Skeleton className="h-9 w-32" />
      </div>
      <Skeleton className="h-12 w-full rounded-lg" />
      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        {Array.from({ length: 4 }).map((_, index) => (
          <Skeleton key={`provider-detail-metric-${index}`} className="h-24 rounded-xl" />
        ))}
      </div>
      <Skeleton className="h-64 w-full rounded-xl" />
    </div>
  );
}
