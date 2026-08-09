"use client";

import Link from "next/link";
import { type ReactNode } from "react";
import { Columns3 } from "lucide-react";
import { useLocale } from "next-intl";

import { CompactValue } from "@/components/app-ui/CompactValue";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { formatCompactNumber } from "@/lib/compact-number";
import { cn } from "@/lib/utils";

import { MAX_BREAKDOWN_ITEMS, TRACE_COLUMN_KEYS, type BreakdownBucket, type RequestTrace, type TraceColumnKey } from "./trace-types";
import { compactID, durationColor, formatDuration, formatLocalDateTime, statusBadgeVariant } from "./trace-utils";

export function TraceColumnVisibilityMenu({
  labels,
  visibleColumns,
  visibleColumnCount,
  columnsLabel,
  menuLabel,
  onToggle,
}: {
  labels: Record<TraceColumnKey, string>;
  visibleColumns: Record<TraceColumnKey, boolean>;
  visibleColumnCount: number;
  columnsLabel: string;
  menuLabel: string;
  onToggle: (key: TraceColumnKey, checked: boolean) => void;
}) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button type="button" variant="outline" size="sm">
          <Columns3 data-icon="inline-start" aria-hidden="true" />
          {columnsLabel}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-48">
        <DropdownMenuLabel>{menuLabel}</DropdownMenuLabel>
        <DropdownMenuSeparator />
        {TRACE_COLUMN_KEYS.map((key) => {
          const checked = visibleColumns[key];
          return (
            <DropdownMenuCheckboxItem
              key={key}
              checked={checked}
              disabled={checked && visibleColumnCount <= 1}
              onCheckedChange={(nextChecked) => onToggle(key, Boolean(nextChecked))}
            >
              {labels[key]}
            </DropdownMenuCheckboxItem>
          );
        })}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

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

export function CompactTraceCard({
  item,
  locale,
  maxDuration,
  t,
  common,
}: {
  item: RequestTrace;
  locale: string;
  maxDuration: number;
  t: (key: string, values?: Record<string, string | number>) => string;
  common: (key: string, values?: Record<string, string | number>) => string;
}) {
  return (
    <Card className="rounded-lg">
      <CardHeader className="gap-1 p-3 pb-2 md:p-3 md:pb-2">
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0">
            <CardDescription className="text-xs leading-5">{formatLocalDateTime(item.started_at, locale)}</CardDescription>
            <CardTitle className="mt-1 min-w-0 text-sm md:text-sm">
              <Link
                className="block truncate font-mono text-primary underline-offset-2 hover:underline"
                href={`/traces/${encodeURIComponent(item.trace_id)}`}
                title={item.trace_id}
                aria-label={t("list.actions.openTrace", { traceId: item.trace_id })}
              >
                {compactID(item.trace_id)}
              </Link>
            </CardTitle>
            {item.request_id ? (
              <p className="mt-0.5 truncate font-mono text-xs text-muted-foreground" title={item.request_id}>
                {t("list.meta.requestId", { id: compactID(item.request_id) })}
              </p>
            ) : null}
          </div>
          <div className="flex shrink-0 flex-wrap items-center justify-end gap-1">
            <Badge variant={statusBadgeVariant(item.status_code)}>{item.status_code}</Badge>
            {item.attempts && item.attempts > 1 ? (
              <Badge variant="outline">{t("list.meta.attemptsShort", { count: item.attempts })}</Badge>
            ) : null}
          </div>
        </div>
      </CardHeader>
      <CardContent className="flex flex-col gap-1.5 px-3 pb-3 pt-0 md:px-3 md:pb-3 md:pt-0">
        <div className="grid grid-cols-3 gap-x-2 gap-y-1.5">
          <CompactListField
            label={t("list.columns.model")}
            value={
              <span className="truncate" dir="auto" title={item.model || common("notSet")}>
                {item.model || common("notSet")}
              </span>
            }
          />
          <CompactListField
            label={t("list.columns.provider")}
            value={
              <span className="truncate" dir="auto" title={item.provider || common("notSet")}>
                {item.provider || common("notSet")}
              </span>
            }
          />
          <CompactListField
            label={t("list.columns.duration")}
            value={<DurationMeter value={item.duration_ms} max={Math.max(maxDuration, 1)} msLabel={t("list.units.ms")} sLabel={t("list.units.s")} locale={locale} compact />}
            inline
          />
        </div>
        <div className="grid grid-cols-2 gap-x-3 gap-y-1.5">
          <CompactListField
            label={t("list.columns.route")}
            inline
            value={
              <span className="truncate font-mono text-xs" dir="ltr" title={item.route}>
                {item.route}
              </span>
            }
          />
          <CompactListField
            label={t("list.columns.error")}
            inline
            value={item.error_code ? (
              <Badge variant="outline" className="max-w-full">
                <span className="truncate font-mono text-xs" dir="auto" title={item.error_code}>
                  {item.error_code}
                </span>
              </Badge>
            ) : (
              <span className="text-muted-foreground">{common("none")}</span>
            )}
          />
        </div>
      </CardContent>
    </Card>
  );
}

function CompactListField({
  label,
  value,
  secondary,
  secondaryLabel,
  className,
  inline = false,
}: {
  label: string;
  value: ReactNode;
  secondary?: ReactNode;
  secondaryLabel?: string;
  className?: string;
  inline?: boolean;
}) {
  return (
    <div className={cn("flex min-w-0 flex-col gap-0.5", inline && "justify-between gap-1", className)}>
      {inline ? (
        <div className="flex min-w-0 items-start justify-between gap-3">
          <span className="shrink-0 text-xs text-muted-foreground">{label}</span>
          <div className="min-w-0 text-sm font-medium text-end">{value}</div>
        </div>
      ) : (
        <>
          <span className="text-xs text-muted-foreground">{label}</span>
          <div className="min-w-0 text-sm font-medium">{value}</div>
        </>
      )}
      {secondary ? (
        <div className="flex min-w-0 flex-col gap-0.5">
          {secondaryLabel ? <span className="text-xs text-muted-foreground">{secondaryLabel}</span> : null}
          <div className="min-w-0 text-xs">{secondary}</div>
        </div>
      ) : null}
    </div>
  );
}

export function SortButton({
  label,
  isActive,
  sortOrder,
  sortLabel,
  onClick,
}: {
  label: string;
  isActive: boolean;
  sortOrder: "asc" | "desc";
  sortLabel: string;
  onClick: () => void;
}) {
  return (
    <button type="button" className="inline-flex items-center gap-1 text-start hover:underline" aria-label={sortLabel} onClick={onClick}>
      {label}
      <span className="font-mono text-xs text-muted-foreground" aria-hidden="true">
        {isActive ? (sortOrder === "asc" ? "^" : "v") : ""}
      </span>
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
