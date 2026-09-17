"use client";

import { useEffect, useMemo, useState } from "react";
import dynamic from "next/dynamic";
import { useLocale, useTranslations } from "next-intl";
import { Activity, ChartNoAxesCombined } from "lucide-react";

import { AppCard } from "@/components/app-ui/AppCard";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { Badge } from "@/components/ui/badge";
import type { ChartConfig } from "@/components/ui/chart";
import { Skeleton } from "@/components/ui/skeleton";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import { formatCompactCurrency } from "@/lib/compact-number";
import { cn } from "@/lib/utils";
import { getDashboardAnalytics, getDashboardTimeseries, type AnalyticsPoint } from "@/features/dashboard/api";
import {
  ANALYTICS_RANGE_KEYS,
  isAnalyticsRangeKey,
  orderPoints,
  resolveRangeWindow,
  type AnalyticsRangeKey,
} from "@/features/dashboard/dashboard-metrics";

type DashboardAnalyticsProps = {
  range: AnalyticsRangeKey;
  onRangeChange: (range: AnalyticsRangeKey) => void;
  refreshKey: number;
};

type ReliabilityChartPoint = {
  label: string;
  successful: number;
  errors: number;
};

type ProviderCostPoint = {
  provider: string;
  cost: number;
  share: number;
  fill: string;
};

const PROVIDER_COLORS = [
  "var(--chart-1)",
  "var(--chart-2)",
  "var(--chart-3)",
  "var(--chart-4)",
  "var(--chart-5)",
] as const;

const ProviderCostChart = dynamic(
  () => import("@/features/dashboard/dashboard-charts").then((module) => module.ProviderCostChart),
  { loading: () => <div className="h-64 w-full animate-pulse rounded-md bg-muted" aria-hidden="true" /> },
);
const ReliabilityChart = dynamic(
  () => import("@/features/dashboard/dashboard-charts").then((module) => module.ReliabilityChart),
  { loading: () => <div className="h-64 w-full animate-pulse rounded-md bg-muted" aria-hidden="true" /> },
);

export function DashboardAnalytics({ range, onRangeChange, refreshKey }: DashboardAnalyticsProps) {
  const t = useTranslations("dashboard.analytics");
  const tRange = useTranslations("dashboard.range");
  const tCharts = useTranslations("dashboard.charts");
  const locale = useLocale();
  const [timeseries, setTimeseries] = useState<AnalyticsPoint[]>([]);
  const [providerBreakdown, setProviderBreakdown] = useState<AnalyticsPoint[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isUnavailable, setIsUnavailable] = useState(false);

  useEffect(() => {
    const controller = new AbortController();
    const window = resolveRangeWindow(range);
    const providerQuery = new URLSearchParams({
      from: window.from,
      to: window.to,
      type: "breakdown",
      group_by: "provider",
    });

    void Promise.all([
      getDashboardTimeseries(window, controller.signal),
      getDashboardAnalytics(providerQuery, controller.signal),
    ])
      .then(([nextTimeseries, nextProviderBreakdown]) => {
        setTimeseries(nextTimeseries);
        setProviderBreakdown(nextProviderBreakdown);
        setIsUnavailable(false);
      })
      .catch((error: unknown) => {
        if (error instanceof Error && error.name === "AbortError") return;
        setTimeseries([]);
        setProviderBreakdown([]);
        setIsUnavailable(true);
      })
      .finally(() => {
        if (!controller.signal.aborted) setIsLoading(false);
      });

    return () => controller.abort();
  }, [range, refreshKey]);

  const bucketFormatter = useMemo(
    () =>
      new Intl.DateTimeFormat(
        locale,
        range === "24h"
          ? { hour: "2-digit", minute: "2-digit" }
          : range === "7d"
            ? { day: "numeric", month: "short", hour: "2-digit" }
            : { day: "numeric", month: "short" },
      ),
    [locale, range],
  );
  const reliabilityData = useMemo(
    () => buildReliabilityData(timeseries, bucketFormatter),
    [bucketFormatter, timeseries],
  );
  const providerCostData = useMemo(
    () => buildProviderCostData(providerBreakdown, t("otherProviders")),
    [providerBreakdown, t],
  );

  const totalRequests = reliabilityData.reduce(
    (total, point) => total + point.successful + point.errors,
    0,
  );
  const totalErrors = reliabilityData.reduce((total, point) => total + point.errors, 0);
  const successRate = totalRequests > 0 ? ((totalRequests - totalErrors) / totalRequests) * 100 : null;
  const totalProviderCost = providerCostData.reduce((total, point) => total + point.cost, 0);
  const percentFormatter = useMemo(() => new Intl.NumberFormat(locale, { maximumFractionDigits: 1 }), [locale]);

  const chartConfig = useMemo(
    () =>
      ({
        successful: { label: t("successful"), color: "var(--chart-2)" },
        errors: { label: t("failed"), color: "var(--destructive)" },
        cost: { label: t("cost"), color: "var(--chart-1)" },
      }) satisfies ChartConfig,
    [t],
  );

  if (isLoading) {
    return <DashboardAnalyticsSkeleton />;
  }

  if (isUnavailable) {
    return (
      <AppCard className="overflow-hidden">
        <EmptyState
          icon={ChartNoAxesCombined}
          title={t("unavailableTitle")}
          description={t("unavailableDescription")}
          variant="inline"
        />
      </AppCard>
    );
  }

  return (
    <section
      className="grid items-stretch gap-4 xl:grid-cols-[minmax(0,2fr)_minmax(0,1fr)]"
      aria-label={t("sectionLabel")}
    >
      <AppCard
        title={tCharts("trafficTitle")}
        description={t("reliabilityDescription")}
        actions={
          <div className="flex flex-wrap items-center gap-2">
            {successRate !== null ? (
              <Badge variant={successRate >= 99 ? "success" : successRate >= 95 ? "secondary" : "warning"}>
                {t("successRateBadge", { value: percentFormatter.format(successRate) })}
              </Badge>
            ) : null}
            <ToggleGroup
              type="single"
              value={range}
              onValueChange={(value) => {
                if (isAnalyticsRangeKey(value)) onRangeChange(value);
              }}
              aria-label={tRange("label")}
              className="rounded-lg border bg-muted/40 p-0.5"
            >
              {ANALYTICS_RANGE_KEYS.map((key) => (
                <ToggleGroupItem
                  key={key}
                  value={key}
                  className={cn(
                    "h-7 px-2.5 text-xs text-muted-foreground",
                    "data-[state=on]:bg-background data-[state=on]:text-foreground data-[state=on]:shadow-xs",
                  )}
                >
                  {tRange(key)}
                </ToggleGroupItem>
              ))}
            </ToggleGroup>
          </div>
        }
        className="h-full min-w-0 overflow-hidden"
        contentClassName="min-w-0"
      >
        {reliabilityData.length === 0 ? (
          <ChartEmptyState title={t("emptyTitle")} description={t("emptyDescription")} />
        ) : (
          <ReliabilityChart
            data={reliabilityData}
            config={chartConfig}
            ariaLabel={t("reliabilityAriaLabel")}
          />
        )}
      </AppCard>

      <AppCard
        title={t("providerCostTitle")}
        description={t("providerCostDescription")}
        className="h-full min-w-0 overflow-hidden"
        contentClassName="min-w-0"
      >
        {providerCostData.length === 0 ? (
          <ChartEmptyState title={t("emptyCostTitle")} description={t("emptyCostDescription")} />
        ) : (
          <div className="flex min-w-0 flex-col gap-4">
            <div className="relative min-w-0">
              <ProviderCostChart
                data={providerCostData}
                config={chartConfig}
                ariaLabel={t("providerCostAriaLabel")}
              />
              <div className="pointer-events-none absolute inset-0 flex flex-col items-center justify-center">
                <span className="text-xs text-muted-foreground">{t("totalCost")}</span>
                <strong className="mt-1 text-lg font-bold tabular-nums text-foreground" dir="ltr">
                  {compactCurrency(totalProviderCost)}
                </strong>
              </div>
            </div>

            <ul className="flex min-w-0 flex-col gap-3">
              {providerCostData.map((item) => (
                <li key={item.provider} className="flex min-w-0 flex-col gap-1.5 text-sm">
                  <div className="flex min-w-0 items-center justify-between gap-3">
                    <span className="truncate text-muted-foreground" dir="auto" title={item.provider}>
                      {item.provider}
                    </span>
                    <span className="shrink-0 font-medium tabular-nums text-foreground" dir="ltr">
                      {compactCurrency(item.cost)} · {percentFormatter.format(item.share * 100)}%
                    </span>
                  </div>
                  <div className="h-1.5 overflow-hidden rounded-full bg-muted">
                    <div
                      className="h-full rounded-full"
                      style={{ width: `${Math.max(2, item.share * 100)}%`, backgroundColor: item.fill }}
                    />
                  </div>
                </li>
              ))}
            </ul>

            <div className="flex items-center justify-between gap-3 border-t pt-3 text-sm">
              <span className="text-muted-foreground">{t("rangeTotalCost")}</span>
              <strong className="font-semibold tabular-nums text-foreground" dir="ltr">
                {compactCurrency(totalProviderCost)}
              </strong>
            </div>
          </div>
        )}
      </AppCard>
    </section>
  );
}

function DashboardAnalyticsSkeleton() {
  return (
    <section className="grid gap-4 xl:grid-cols-[minmax(0,2fr)_minmax(0,1fr)]" aria-hidden="true">
      {[0, 1].map((item) => (
        <AppCard key={item} className="overflow-hidden">
          <div className="flex flex-col gap-4">
            <Skeleton className="h-5 w-40" />
            <Skeleton className="h-64 w-full" />
          </div>
        </AppCard>
      ))}
    </section>
  );
}

function ChartEmptyState({ title, description }: { title: string; description: string }) {
  return (
    <EmptyState
      icon={Activity}
      title={title}
      description={description}
      variant="inline"
      density="compact"
      className="h-64"
    />
  );
}

function buildReliabilityData(
  points: AnalyticsPoint[],
  formatter: Intl.DateTimeFormat,
): ReliabilityChartPoint[] {
  return orderPoints(points).map(({ point, date }) => {
    const requests = Math.max(0, Number(point.request_count ?? 0));
    const errors = Math.max(0, Math.min(requests, Number(point.error_count ?? 0)));
    return {
      label: formatter.format(date),
      successful: requests - errors,
      errors,
    };
  });
}

function buildProviderCostData(points: AnalyticsPoint[], otherProvidersLabel: string): ProviderCostPoint[] {
  const ordered = points
    .map((point) => ({
      provider: point.group?.trim() || otherProvidersLabel,
      cost: Math.max(0, Number(point.cost_usd ?? 0)),
    }))
    .filter((point) => point.cost > 0)
    .sort((a, b) => b.cost - a.cost);

  const visible = ordered.slice(0, 4);
  const otherCost = ordered.slice(4).reduce((total, point) => total + point.cost, 0);
  if (otherCost > 0) visible.push({ provider: otherProvidersLabel, cost: otherCost });
  const total = visible.reduce((sum, point) => sum + point.cost, 0);

  return visible.map((point, index) => ({
    ...point,
    share: total > 0 ? point.cost / total : 0,
    fill: PROVIDER_COLORS[index % PROVIDER_COLORS.length],
  }));
}

function compactCurrency(value: number) {
  return formatCompactCurrency(value, {
    locale: "en-US",
    currency: "USD",
    compactThreshold: 1_000,
    maximumFractionDigits: 2,
  }).value;
}
