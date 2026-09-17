import * as React from "react";
import Link from "next/link";
import { ArrowUpRight, type LucideIcon } from "lucide-react";

import { cn } from "@/lib/utils";

export type DashboardAlertTone = "error" | "warning" | "info";

export type DashboardAlertItem = {
  id: string;
  tone: DashboardAlertTone;
  icon: LucideIcon;
  /** Bold lead of the sentence. */
  title: React.ReactNode;
  /** Rest of the sentence, rendered with lower emphasis. */
  detail?: React.ReactNode;
  action?: { label: string; href: string };
};

const toneClasses: Record<DashboardAlertTone, string> = {
  error: "border-destructive-soft-border bg-destructive-soft text-destructive-soft-foreground",
  warning: "border-warning-soft-border bg-warning-soft text-warning-soft-foreground",
  info: "border-info-soft-border bg-info-soft text-info-soft-foreground",
};

const toneOrder: Record<DashboardAlertTone, number> = { error: 0, warning: 1, info: 2 };

export const DASHBOARD_ALERT_LIMIT = 3;

export function DashboardAlert({ tone, icon: Icon, title, detail, action }: Omit<DashboardAlertItem, "id">) {
  return (
    <div
      role={tone === "error" ? "alert" : "status"}
      data-tone={tone}
      className={cn("flex flex-wrap items-center gap-x-3 gap-y-2 rounded-lg border px-4 py-3 text-sm", toneClasses[tone])}
    >
      <Icon className="size-4 shrink-0" aria-hidden="true" />
      <p className="min-w-0 flex-1 leading-6">
        <span className="font-semibold">{title}</span>
        {detail ? <span className="opacity-80"> {detail}</span> : null}
      </p>
      {action ? (
        <Link href={action.href} className="inline-flex shrink-0 items-center gap-1 font-medium underline-offset-4 hover:underline">
          {action.label}
          <ArrowUpRight className="size-3.5" aria-hidden="true" />
        </Link>
      ) : null}
    </div>
  );
}

export function sortDashboardAlerts(items: DashboardAlertItem[]): DashboardAlertItem[] {
  return [...items].sort((a, b) => toneOrder[a.tone] - toneOrder[b.tone]);
}

export function DashboardAlertList({
  items,
  moreLabel,
  moreHref,
  className,
}: {
  items: DashboardAlertItem[];
  moreLabel: (count: number) => React.ReactNode;
  /** Defaults to the first hidden alert's action. */
  moreHref?: string;
  className?: string;
}) {
  if (items.length === 0) return null;
  const sorted = sortDashboardAlerts(items);
  const visible = sorted.slice(0, DASHBOARD_ALERT_LIMIT);
  const hiddenCount = sorted.length - visible.length;
  const overflowHref = moreHref ?? sorted[DASHBOARD_ALERT_LIMIT]?.action?.href;

  return (
    <section className={cn("flex flex-col gap-2", className)} data-slot="dashboard-alerts">
      {visible.map(({ id, ...item }) => (
        <DashboardAlert key={id} {...item} />
      ))}
      {hiddenCount > 0 ? (
        overflowHref ? (
          <Link href={overflowHref} className="self-end text-xs font-medium text-muted-foreground underline-offset-4 hover:underline">
            {moreLabel(hiddenCount)}
          </Link>
        ) : (
          <p className="self-end text-xs text-muted-foreground">{moreLabel(hiddenCount)}</p>
        )
      ) : null}
    </section>
  );
}
