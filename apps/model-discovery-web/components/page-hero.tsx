import type { LucideIcon } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

export function PageHero({
  eyebrow,
  title,
  description,
  icon: Icon,
  afterTitle,
  actions,
  compact = false,
}: {
  eyebrow: string;
  title: string;
  description: string;
  icon: LucideIcon;
  afterTitle?: React.ReactNode;
  actions?: React.ReactNode;
  compact?: boolean;
}) {
  return (
    <header
      className={cn(
        "page-hero relative isolate overflow-hidden rounded-2xl border border-card-border bg-card",
        compact ? "p-5 sm:p-6" : "p-5 sm:p-7 lg:p-8",
      )}
    >
      <div className="pointer-events-none absolute inset-y-0 end-0 hidden w-2/5 sm:block" aria-hidden="true">
        <div className="hero-orbit absolute end-8 top-1/2 size-44 -translate-y-1/2 rounded-full border border-primary/20" />
        <div className="hero-orbit absolute end-20 top-1/2 size-28 -translate-y-1/2 rounded-full border border-primary/25" />
        <div className="absolute end-28 top-1/2 flex size-12 -translate-y-1/2 items-center justify-center rounded-2xl bg-primary text-primary-foreground shadow-lg">
          <Icon className="size-6" aria-hidden="true" />
        </div>
      </div>

      <div className="relative flex min-w-0 flex-col items-start gap-4 sm:max-w-3xl">
        <Badge className="border-primary-soft bg-primary-soft text-primary-soft-foreground">
          {eyebrow}
        </Badge>
        <div className="flex min-w-0 flex-col gap-2">
          <div className="flex flex-wrap items-center gap-2">
            <h1 className="text-2xl font-bold tracking-tight md:text-3xl lg:text-4xl">{title}</h1>
            {afterTitle}
          </div>
          <p className="max-w-2xl text-sm leading-7 text-muted-foreground sm:text-base">
            {description}
          </p>
        </div>
        {actions ? <div className="flex flex-wrap gap-2">{actions}</div> : null}
      </div>
    </header>
  );
}
