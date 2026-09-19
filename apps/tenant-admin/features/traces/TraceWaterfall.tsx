"use client";

import { useLocale, useTranslations } from "next-intl";

import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

import type { TraceSpan } from "./detail-contracts";
import {
  buildWaterfallRows,
  formatSpanDuration,
  sortSpans,
  spanKindBadgeVariant,
  textStatusBadgeVariant,
} from "./trace-waterfall";

type TraceWaterfallProps = {
  spans: TraceSpan[];
  /** Stacks the label above the bar at every width, for narrow panels. */
  compact?: boolean;
};

/** Span waterfall for one trace. Used by the trace detail page and the /traces detail panel. */
export function TraceWaterfall({ spans, compact = false }: TraceWaterfallProps) {
  const t = useTranslations("traces");
  const locale = useLocale();
  const rows = buildWaterfallRows(sortSpans(spans));
  const msLabel = t("detail.units.ms");

  if (rows.length === 0) {
    return <p className="rounded-md bg-muted/40 px-3 py-2 text-sm text-muted-foreground">{t("detail.timeline.empty")}</p>;
  }

  const totalMS = rows[0].totalMS;
  const hasNonZeroSpanOffset = rows.some((row) => row.offsetMS > 0);
  const gridClassName = compact ? "grid gap-1 px-2" : "grid gap-2 px-2 md:grid-cols-[minmax(14rem,18rem)_1fr]";

  return (
    <div className="flex flex-col gap-1">
      <div className="rounded-md bg-muted/40 px-3 py-1.5 text-xs text-muted-foreground">
        {hasNonZeroSpanOffset ? t("detail.timeline.waterfallHint") : t("detail.timeline.durationHint")}
      </div>
      <div className={gridClassName}>
        <div dir="ltr" className={cn("flex items-center justify-between text-xs text-muted-foreground", !compact && "md:hidden")}>
          <span>{formatSpanDuration(0, msLabel, locale)}</span>
          <span>{formatSpanDuration(totalMS, msLabel, locale)}</span>
        </div>
        {compact ? null : (
          <>
            <div className="hidden md:block" />
            <TimelineAxis totalMS={totalMS} msLabel={msLabel} locale={locale} />
          </>
        )}
      </div>
      <ol className="flex flex-col gap-1">
        {rows.map((row, idx) => (
          <li key={row.span.span_id} className={cn(gridClassName, "py-1", !compact && "md:items-center")}>
            <div className="min-w-0">
              <div className="flex min-w-0 flex-wrap items-center gap-2">
                <TimelineStepMarker index={idx} kind={row.span.kind} total={rows.length} locale={locale} compact={compact} />
                <Badge variant={textStatusBadgeVariant(row.span.status)}>{row.span.status}</Badge>
                <p className="truncate text-sm font-medium">{row.span.name}</p>
              </div>
              <div className="mt-0.5 flex flex-wrap items-center gap-x-3 gap-y-1 pe-4 text-xs text-muted-foreground md:pe-9">
                <span>{formatSpanDuration(row.span.duration_ms, msLabel, locale)}</span>
                {row.offsetMS > 0 ? <span dir="ltr">+{formatSpanDuration(row.offsetMS, msLabel, locale)}</span> : null}
              </div>
            </div>
            <div dir="ltr">
              <div className={cn("relative rounded-md bg-muted", compact ? "h-5" : "h-8")}>
                <TimelineGridLines />
                <div
                  className={cn("absolute rounded-md bg-primary/70", compact ? "top-0.5 h-4" : "top-1 h-6")}
                  style={{ left: `${row.leftPct}%`, width: `${row.widthPct}%` }}
                />
                {!compact && row.widthPct >= 12 ? (
                  <span
                    className="absolute top-1/2 -translate-y-1/2 truncate px-2 text-xs font-medium text-primary-foreground"
                    style={{ left: `${row.leftPct}%`, width: `${row.widthPct}%` }}
                  >
                    {formatSpanDuration(row.span.duration_ms, msLabel, locale)}
                  </span>
                ) : null}
              </div>
            </div>
          </li>
        ))}
      </ol>
    </div>
  );
}

function TimelineStepMarker({ index, kind, total, locale, compact }: { index: number; kind: string; total: number; locale: string; compact: boolean }) {
  return (
    <span className="relative flex shrink-0 items-center gap-1.5">
      {!compact && index < total - 1 ? (
        <span className="absolute start-3 top-6 h-7 border-s border-border" aria-hidden="true" />
      ) : null}
      <span className="flex size-6 items-center justify-center rounded-full border bg-background text-xs font-semibold">
        {new Intl.NumberFormat(locale).format(index + 1)}
      </span>
      <Badge variant={spanKindBadgeVariant(kind)}>{kind}</Badge>
    </span>
  );
}

function TimelineAxis({ totalMS, msLabel, locale }: { totalMS: number; msLabel: string; locale: string }) {
  const ticks = [0, 25, 50, 75, 100];
  return (
    <div dir="ltr" className="hidden md:block">
      <div className="relative h-6">
        {ticks.map((tick) => (
          <span
            key={tick}
            className={cn(
              "absolute top-0 text-xs text-muted-foreground",
              tick === 0 ? "translate-x-0" : tick === 100 ? "-translate-x-full" : "-translate-x-1/2",
            )}
            style={{ insetInlineStart: `${tick}%` }}
          >
            {formatSpanDuration((totalMS * tick) / 100, msLabel, locale)}
          </span>
        ))}
      </div>
    </div>
  );
}

function TimelineGridLines() {
  return (
    <>
      {[25, 50, 75].map((tick) => (
        <span
          key={tick}
          className="absolute top-0 h-full border-s border-background/80"
          style={{ insetInlineStart: `${tick}%` }}
          aria-hidden="true"
        />
      ))}
    </>
  );
}
