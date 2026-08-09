import type { ReactNode } from "react";

import { Check } from "lucide-react";

import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

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
  variant?: "default" | "compact";
}) {
  const compact = props.variant === "compact";

  return (
    <button
      type="button"
      onClick={props.onClick}
      disabled={props.disabled}
      aria-current={props.state === "active" ? "step" : undefined}
      aria-label={`${props.number}. ${props.title}`}
      className={cn(
        "group flex flex-1 flex-row items-center text-start transition focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:cursor-not-allowed disabled:opacity-60",
        compact
          ? "min-h-16 min-w-0 gap-2 rounded-lg border border-card-border bg-card px-2 py-2 shadow-xs hover:border-primary-subtle-border hover:bg-primary-soft sm:gap-3 sm:px-3"
          : "min-w-14 justify-center gap-2 rounded-lg px-2 py-2 hover:bg-muted/50 disabled:hover:bg-transparent sm:min-w-24 sm:justify-start sm:py-1 lg:w-full lg:min-w-0 lg:flex-none lg:px-1 lg:py-2",
        compact && props.state === "active" && "border-primary-subtle-border bg-primary-soft",
        compact && props.state === "idle" && "bg-muted/20",
      )}
    >
      <div
        className={cn(
          "flex shrink-0 items-center justify-center rounded-full text-sm",
          compact ? "size-8" : "size-11 sm:size-9 sm:text-xs lg:size-11 lg:text-sm",
          props.state === "idle"
            ? "border border-border bg-background text-muted-foreground"
            : props.state === "complete" && compact
              ? "border border-primary-subtle-border bg-primary-soft font-semibold text-primary-soft-foreground"
              : "bg-primary font-semibold text-primary-foreground",
          !compact && props.state !== "idle" && "ring-8 ring-primary/10 sm:ring-4 lg:ring-8",
        )}
      >
        {props.state === "complete" ? <Check aria-hidden="true" className="size-4" /> : props.number}
      </div>
      <span
        className={cn(
          "min-w-0 flex-1 text-start text-xs leading-5 sm:text-sm",
          !compact && "hidden sm:block",
          props.state === "idle"
            ? "font-medium text-muted-foreground"
            : props.state === "active" && compact
              ? "font-semibold text-primary-soft-foreground"
              : compact
                ? "font-medium text-foreground"
                : "font-semibold text-primary",
        )}
      >
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
  const content = scrollable ? (
    <ScrollArea className="min-h-0 flex-1 [&_[data-slot=scroll-area-viewport]]:overscroll-contain">
      <div className={contentClassName ?? "p-4 pb-6 sm:p-5 sm:pb-7"}>{children}</div>
    </ScrollArea>
  ) : (
    <div className={contentClassName ?? "p-4 pb-6 sm:p-5 sm:pb-7"}>{children}</div>
  );

  return (
    <section
      id={id}
      ref={setSectionNode}
      className={fillHeight ? "flex min-h-0 flex-1 flex-col gap-3" : "flex flex-col gap-3"}
    >
      <header className="flex flex-col gap-3 px-1 text-start sm:flex-row sm:items-end sm:justify-between">
        <div className="min-w-0">
          <h2 className="text-lg font-semibold">{title}</h2>
          {description ? <p className="mt-1 text-sm text-muted-foreground">{description}</p> : null}
        </div>
        {actions ? <div className="shrink-0">{actions}</div> : null}
      </header>
      <div className={fillHeight ? "flex min-h-0 flex-1 flex-col overflow-hidden rounded-lg border border-border bg-card shadow-sm" : "flex flex-col rounded-lg border border-border bg-card shadow-sm"}>
        {content}
        {footer ? (
          <footer className="sticky bottom-0 z-10 flex shrink-0 flex-wrap justify-between gap-2 border-t border-border bg-card/95 px-5 py-3 shadow-lg">
            {footer}
          </footer>
        ) : null}
      </div>
    </section>
  );
}
