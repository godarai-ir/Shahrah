import type { ReactNode } from "react";

import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Skeleton } from "@/components/ui/skeleton";

export type CreateWizardSectionID = "details" | "models" | "access" | "review";

export function ProviderAccountSidebarSkeleton() {
  return (
    <>
      <Card>
        <CardHeader className="pb-3">
          <Skeleton className="h-5 w-40" />
          <Skeleton className="h-4 w-56" />
        </CardHeader>
        <CardContent>
          <div className="flex flex-col gap-2">
            {Array.from({ length: 4 }).map((_, index) => (
              <div key={`provider-skeleton-${index}`} className="flex items-center justify-between gap-2 rounded-lg border px-3 py-2">
                <div className="flex items-center gap-2">
                  <Skeleton className="size-8 rounded-full" />
                  <Skeleton className="h-4 w-28" />
                </div>
                <Skeleton className="h-5 w-8 rounded-full" />
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="pb-3">
          <Skeleton className="h-5 w-36" />
          <Skeleton className="h-4 w-44" />
        </CardHeader>
        <CardContent>
          <div className="flex flex-col gap-2">
            {Array.from({ length: 3 }).map((_, index) => (
              <div key={`account-skeleton-${index}`} className="rounded-lg border px-2 py-2">
                <div className="flex items-center justify-between gap-2">
                  <div className="flex min-w-0 flex-1 flex-col gap-2">
                    <Skeleton className="h-4 w-32" />
                    <Skeleton className="h-3 w-20" />
                  </div>
                  <Skeleton className="h-6 w-24 rounded-full" />
                </div>
                <div className="mt-2 flex justify-end">
                  <Skeleton className="h-8 w-20" />
                </div>
              </div>
            ))}
            <div className="flex gap-2 pt-2">
              <Skeleton className="h-8 flex-1" />
              <Skeleton className="h-8 flex-1" />
            </div>
          </div>
        </CardContent>
      </Card>
    </>
  );
}

export function CreateFlowStepIndicator(props: {
  number: number;
  title: string;
  state: "active" | "complete" | "idle";
  onClick: () => void;
  disabled?: boolean;
}) {
  return (
    <button
      type="button"
      onClick={props.onClick}
      disabled={props.disabled}
      dir="rtl"
      className="group flex w-full flex-row items-center justify-start gap-2 rounded-lg px-1 py-2 text-right transition hover:bg-muted/50 focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-60 disabled:hover:bg-transparent"
    >
      <div
        className={
          props.state === "idle"
            ? "flex size-11 shrink-0 items-center justify-center rounded-full border bg-background text-sm text-muted-foreground"
            : "flex size-11 shrink-0 items-center justify-center rounded-full bg-primary text-sm font-semibold text-primary-foreground ring-8 ring-primary/10"
        }
      >
        {props.number}
      </div>
      <span className={props.state === "idle" ? "min-w-0 flex-1 text-right text-sm font-medium text-muted-foreground" : "min-w-0 flex-1 text-right text-sm font-semibold text-primary"}>
        {props.title}
      </span>
    </button>
  );
}

export function WizardPanel({
  title,
  description,
  children,
  footer,
  actions,
  id,
  setSectionNode,
}: {
  title: string;
  description: string;
  children: ReactNode;
  footer?: ReactNode;
  actions?: ReactNode;
  id?: string;
  setSectionNode?: (node: HTMLElement | null) => void;
}) {
  return (
    <section
      id={id}
      ref={setSectionNode}
      className="flex h-[min(640px,calc(100dvh-13.5rem))] min-h-[420px] flex-col gap-3"
    >
      <header dir="rtl" className="flex flex-col gap-3 px-1 text-right sm:flex-row sm:items-end sm:justify-between">
        <div className="min-w-0">
          <h2 className="text-lg font-semibold">{title}</h2>
          <p className="mt-1 text-sm text-muted-foreground">{description}</p>
        </div>
        {actions ? <div className="shrink-0">{actions}</div> : null}
      </header>
      <div className="flex min-h-0 flex-1 flex-col overflow-hidden rounded-lg border border-border bg-card shadow-sm">
        <ScrollArea className="min-h-0 flex-1 [&_[data-slot=scroll-area-viewport]]:overscroll-contain">
          <div className="p-4 pb-6 sm:p-5 sm:pb-7">{children}</div>
        </ScrollArea>
        {footer ? (
          <footer className="flex shrink-0 flex-wrap justify-between gap-2 border-t border-border bg-card/95 px-5 py-3 shadow-[0_-8px_18px_rgba(15,23,42,0.04)]">
            {footer}
          </footer>
        ) : null}
      </div>
    </section>
  );
}
