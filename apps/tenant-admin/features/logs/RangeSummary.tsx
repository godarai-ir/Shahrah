"use client";

import { Fragment, memo, useState, type ReactNode } from "react";
import { BarChart3, ChevronDown, ShieldCheck, type LucideIcon } from "lucide-react";
import { useTranslations } from "next-intl";

import { AppCard } from "@/components/app-ui/AppCard";
import { CompactValue } from "@/components/app-ui/CompactValue";
import { cn } from "@/lib/utils";

import type { LogsFormatter, Summary } from "./logs-types";
import { useCollapsibleState } from "./use-collapsible-state";

export const RANGE_SUMMARY_STORAGE_KEY = "logs.rangeSummaryOpen";

type RangeSummaryProps = {
  summary: Summary;
  summaryUnavailable: boolean;
  formatters: LogsFormatter;
  t: ReturnType<typeof useTranslations>;
  /** Extra content shown under the metric cards once the section is open, such as the charts. */
  children?: ReactNode;
};

/**
 * Collapsible range overview. Closed, it keeps every headline number on one line so collapsing
 * only frees space; open, it shows the metric cards and whatever the caller passes as children.
 * Children mount on first open, so deferred charts do not load until someone asks for them.
 */
export const RangeSummary = memo(function RangeSummary({ summary, summaryUnavailable, formatters, t, children }: RangeSummaryProps) {
  const [open, setOpen] = useCollapsibleState(RANGE_SUMMARY_STORAGE_KEY);
  const [hasOpened, setHasOpened] = useState(open);
  if (open && !hasOpened) setHasOpened(true);

  return (
    <details
      open={open}
      onToggle={(event) => {
        const next = event.currentTarget.open;
        if (next !== open) setOpen(next);
      }}
      className="group rounded-lg border border-card-border bg-card text-card-foreground shadow-xs"
      data-testid="logs-range-summary"
    >
      <summary className="flex min-h-11 cursor-pointer list-none items-center gap-3 rounded-lg px-4 py-2 text-xs outline-none focus-visible:ring-[3px] focus-visible:ring-ring/50 [&::-webkit-details-marker]:hidden">
        <ChevronDown
          aria-hidden="true"
          className="size-4 shrink-0 text-muted-foreground transition-transform ltr:-rotate-90 rtl:rotate-90 group-open:ltr:rotate-0 group-open:rtl:rotate-0"
        />
        <span className="shrink-0 text-sm font-semibold text-foreground">{t("rangeSummary.title")}</span>
        <RangeSummaryLine summary={summary} summaryUnavailable={summaryUnavailable} formatters={formatters} t={t} />
      </summary>
      {hasOpened ? (
        <div className="flex flex-col gap-4 border-t border-border p-4">
          <LogsMetrics summary={summary} summaryUnavailable={summaryUnavailable} formatters={formatters} t={t} />
          {children}
        </div>
      ) : null}
    </details>
  );
});

function RangeSummaryLine({ summary, summaryUnavailable, formatters, t }: Omit<RangeSummaryProps, "children">) {
  if (summaryUnavailable) {
    return <span className="min-w-0 truncate text-muted-foreground">{t("analyticsUnavailable")}</span>;
  }

  const errorRate = summary.error_rate ?? 0;
  const items: Array<{ key: string; label: string; value: ReactNode; tone?: "destructive" }> = [
    { key: "requests", label: t("rangeSummary.requests"), value: <CompactValue display={formatters.numberDisplay(summary.request_count ?? 0)} /> },
    { key: "errorRate", label: t("rangeSummary.errorRate"), value: formatters.percent(errorRate), tone: errorRate > 0 ? "destructive" : undefined },
    { key: "p95", label: t("rangeSummary.p95"), value: formatters.ms(summary.latency_p95_ms ?? 0) },
    { key: "cost", label: t("rangeSummary.cost"), value: <CompactValue display={formatters.usdDisplay(summary.cost_usd ?? 0)} /> },
    { key: "cache", label: t("rangeSummary.cache"), value: formatters.percent(summary.cache_hit_rate ?? 0) },
  ];

  return (
    <dl className="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1">
      {items.map((item, index) => (
        <Fragment key={item.key}>
          {index > 0 ? <span aria-hidden="true" className="text-muted-foreground">·</span> : null}
          <div className="flex items-baseline gap-1 whitespace-nowrap" data-metric={item.key}>
            <dt className="text-muted-foreground">{item.label}</dt>
            <dd className={cn("font-semibold tabular-nums", item.tone === "destructive" ? "text-destructive" : "text-foreground")}>{item.value}</dd>
          </div>
        </Fragment>
      ))}
    </dl>
  );
}

export const LogsMetrics = memo(function LogsMetrics({
  summary,
  summaryUnavailable,
  formatters,
  t,
}: {
  summary: Summary;
  summaryUnavailable: boolean;
  formatters: LogsFormatter;
  t: ReturnType<typeof useTranslations>;
}) {
  return (
    <div className="grid grid-cols-2 gap-2 sm:gap-3 md:grid-cols-3 xl:grid-cols-5">
      <AnalyticsMetric unavailable={summaryUnavailable} label={t("cards.requests")} value={<CompactValue display={formatters.numberDisplay(summary.request_count ?? 0)} />} icon={BarChart3} t={t} />
      <AnalyticsMetric unavailable={summaryUnavailable} label={t("cards.errorRate")} value={formatters.percent(summary.error_rate ?? 0)} description={<CompactValue display={formatters.numberDisplay(summary.error_count ?? 0)} />} icon={ShieldCheck} t={t} />
      <AnalyticsMetric unavailable={summaryUnavailable} label={t("cards.p95")} value={formatters.ms(summary.latency_p95_ms ?? 0)} description={`${t("cards.p99")} ${formatters.ms(summary.latency_p99_ms ?? 0)}`} t={t} />
      <AnalyticsMetric unavailable={summaryUnavailable} label={t("cards.cost")} value={<CompactValue display={formatters.usdDisplay(summary.cost_usd ?? 0)} />} description={<CompactValue display={formatters.irrDisplay(summary.cost_irr ?? 0)} />} t={t} />
      <AnalyticsMetric unavailable={summaryUnavailable} label={t("cards.cache")} value={formatters.percent(summary.cache_hit_rate ?? 0)} description={<span>{t("cards.retries")} <CompactValue display={formatters.numberDisplay(summary.retry_count ?? 0)} /></span>} t={t} />
    </div>
  );
});

function AnalyticsMetric({
  unavailable,
  label,
  value,
  description,
  icon,
  t,
}: {
  unavailable: boolean;
  label: string;
  value: ReactNode;
  description?: ReactNode;
  icon?: LucideIcon;
  t: ReturnType<typeof useTranslations>;
}) {
  if (!unavailable) {
    const Icon = icon;
    return (
      <AppCard contentClassName="p-3 sm:p-4 xl:pt-6">
        <div className="flex min-w-0 items-start justify-between gap-2 sm:gap-3">
          <div className="min-w-0">
            <p className="truncate text-xs font-medium text-muted-foreground sm:text-sm">{label}</p>
            <p className="mt-1 truncate text-lg font-bold tracking-tight text-foreground sm:mt-2 sm:text-2xl xl:text-3xl">{value}</p>
            {description ? <p className="mt-0.5 truncate text-xs leading-5 text-muted-foreground sm:mt-1">{description}</p> : null}
          </div>
          {Icon ? (
            <div className="flex size-8 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary sm:size-10 sm:rounded-xl">
              <Icon aria-hidden="true" className="h-4 w-4 sm:h-5 sm:w-5" />
            </div>
          ) : null}
        </div>
      </AppCard>
    );
  }

  return (
    <div className="rounded-lg border border-dashed border-border bg-muted/40 p-3 opacity-70 sm:p-4">
      <div className="text-xs font-medium text-muted-foreground">{label}</div>
      <div className="pt-1 text-base font-semibold text-muted-foreground sm:pt-2 sm:text-lg">{t("analyticsUnavailableShort")}</div>
      <div className="pt-1 text-xs text-muted-foreground">{t("analyticsUnavailable")}</div>
    </div>
  );
}
