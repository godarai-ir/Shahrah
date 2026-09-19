"use client";

import { useLocale, useTranslations } from "next-intl";

import { CompactValue } from "@/components/app-ui/CompactValue";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { formatCompactNumber } from "@/lib/compact-number";
import { cn } from "@/lib/utils";

import { MAX_BREAKDOWN_ITEMS, type BreakdownBucket, type RequestTrace } from "./trace-types";
import { compactID, durationColor, formatDuration, formatLocalDateTime, statusBadgeVariant } from "./trace-utils";

export function BreakdownCard({
  title,
  buckets,
  activeValue,
  noDataLabel,
  selectLabel,
  onSelect,
  mono = false,
  compact = false,
}: {
  title: string;
  buckets: BreakdownBucket[];
  activeValue?: string;
  noDataLabel: string;
  selectLabel: (value: string, count: number) => string;
  onSelect: (value: string) => void;
  mono?: boolean;
  compact?: boolean;
}) {
  const locale = useLocale();
  const visibleBuckets = buckets.slice(0, MAX_BREAKDOWN_ITEMS);
  const maxCount = Math.max(0, ...visibleBuckets.map((item) => item.count));

  return (
    <Card className="rounded-lg">
      <CardHeader className="gap-1 p-3 pb-2 md:p-3 md:pb-2">
        <CardTitle className="truncate text-sm">{title}</CardTitle>
      </CardHeader>
      <CardContent className={cn("flex flex-col gap-1.5 px-3 pb-3 md:px-3 md:pb-3", compact && "max-h-40 overflow-y-auto pe-2")}>
        {visibleBuckets.length === 0 ? (
          <p className="py-4 text-center text-xs text-muted-foreground">{noDataLabel}</p>
        ) : (
          visibleBuckets.map((item) => {
            const active = activeValue === item.value;
            const width = maxCount > 0 ? Math.max(6, (item.count / maxCount) * 100) : 0;
            return (
              <button
                key={`${title}:${item.value}`}
                type="button"
                className={cn(
                  "grid gap-1 rounded-md border p-2 text-start transition-colors hover:bg-muted/50",
                  active && "border-primary bg-muted/50",
                )}
                title={selectLabel(item.label, item.count)}
                aria-label={selectLabel(item.label, item.count)}
                onClick={() => onSelect(item.value)}
              >
                <span className="flex min-w-0 items-center justify-between gap-2">
                  <span className={cn("min-w-0 truncate text-xs", mono && "font-mono")} dir="auto">
                    {item.label}
                  </span>
                  <CompactValue
                    display={formatCompactNumber(item.count, { locale })}
                    className="font-mono text-xs text-muted-foreground"
                    dir="ltr"
                  />
                </span>
                <span className="h-1.5 overflow-hidden rounded-full bg-muted">
                  <span
                    className={cn("block h-full rounded-full", active ? "bg-primary" : "bg-muted-foreground")}
                    style={{ width: `${width}%` }}
                  />
                </span>
              </button>
            );
          })
        )}
      </CardContent>
    </Card>
  );
}

/**
 * One row of the /traces master list. Span counts are not part of the list payload, so the row shows
 * attempts and model instead; the span count lives in the detail panel header.
 */
export function TraceListItem({
  item,
  selected,
  maxDuration,
  onSelect,
}: {
  item: RequestTrace;
  selected: boolean;
  maxDuration: number;
  onSelect: (traceID: string) => void;
}) {
  const t = useTranslations("traces");
  const common = useTranslations("common");
  const locale = useLocale();
  const attempts = item.attempts ?? 0;

  return (
    <button
      type="button"
      aria-pressed={selected}
      aria-label={t("list.actions.select", { traceId: item.trace_id })}
      onClick={() => onSelect(item.trace_id)}
      className={cn(
        "flex w-full min-w-0 flex-col gap-1.5 rounded-md border px-3 py-2 text-start transition-colors hover:bg-muted/50 focus-visible:outline-2 focus-visible:outline-ring",
        selected ? "border-primary bg-muted/50" : "border-border",
      )}
    >
      <span className="flex min-w-0 items-center gap-2">
        <Badge variant={statusBadgeVariant(item.status_code)}>{item.status_code}</Badge>
        <span className="min-w-0 truncate font-mono text-xs" dir="ltr" title={item.trace_id}>{compactID(item.trace_id)}</span>
        <span className="ms-auto shrink-0 text-xs text-muted-foreground">{formatLocalDateTime(item.started_at, locale)}</span>
      </span>
      <span className="flex min-w-0 items-center gap-2 text-xs">
        <span className="min-w-0 truncate" dir="auto" title={item.model || common("notSet")}>{item.model || common("notSet")}</span>
        {attempts > 1 ? <Badge variant="outline" className="shrink-0">{t("list.meta.attempts", { count: attempts })}</Badge> : null}
        {item.error_code ? (
          <span className="min-w-0 truncate font-mono text-destructive" dir="auto" title={item.error_code}>{item.error_code}</span>
        ) : null}
      </span>
      <DurationMeter value={item.duration_ms} max={maxDuration} msLabel={t("list.units.ms")} sLabel={t("list.units.s")} locale={locale} />
    </button>
  );
}

export function DurationMeter({
  value,
  max,
  msLabel,
  sLabel,
  locale,
  compact = false,
}: {
  value: number;
  max: number;
  msLabel: string;
  sLabel: string;
  locale: string;
  compact?: boolean;
}) {
  const width = max > 0 ? Math.max(6, (value / max) * 100) : 0;

  if (compact) {
    return (
      <span className="font-mono text-xs tabular-nums">
        {formatDuration(value, msLabel, sLabel, locale)}
      </span>
    );
  }

  return (
    <div className="flex min-w-0 items-center gap-2">
      <span className="w-16 shrink-0 text-end font-mono tabular-nums">
        {formatDuration(value, msLabel, sLabel, locale)}
      </span>
      <span className="h-1.5 min-w-16 flex-1 overflow-hidden rounded-full bg-muted">
        <span
          className="block h-full rounded-full"
          style={{ width: `${width}%`, backgroundColor: durationColor(value) }}
        />
      </span>
    </div>
  );
}
