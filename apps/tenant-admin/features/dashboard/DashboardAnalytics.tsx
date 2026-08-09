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
import { formatCompactCurrency, formatCompactNumber } from "@/lib/compact-number";
import { getDashboardAnalytics } from "@/features/dashboard/api";

type AnalyticsPoint = {
  timestamp?: string;
  group?: string;
  request_count?: number;
  error_count?: number;
  cost_usd?: number;
};

type DashboardAnalyticsProps = {
  currentPeriodStart?: string;
  currentPeriodEnd?: string;
  currentUsage: number | null;
  requestLimit: number | null;
  refreshKey: number;
};

type UsageChartPoint = {
  label: string;
  cumulative: number | null;
  projected: number | null;
};

type ReliabilityChartPoint = {
  label: string;
  successful: number;
  errors: number;
};

type ProviderCostPoint = {
  provider: string;
  cost: number;
  fill: string;
};

const DAY_MS = 24 * 60 * 60 * 1_000;
const PROVIDER_COLORS = [
  "var(--chart-1)",
  "var(--chart-2)",
  "var(--chart-3)",
  "var(--chart-4)",
  "var(--chart-5)",
] as const;

const UsageChart = dynamic(
  () => import("@/features/dashboard/dashboard-charts").then((module) => module.UsageChart),
  { loading: () => <div className="h-72 w-full animate-pulse rounded-md bg-muted" aria-hidden="true" /> },
);
const ProviderCostChart = dynamic(
  () => import("@/features/dashboard/dashboard-charts").then((module) => module.ProviderCostChart),
  { loading: () => <div className="h-64 w-full animate-pulse rounded-md bg-muted" aria-hidden="true" /> },
);
const ReliabilityChart = dynamic(
  () => import("@/features/dashboard/dashboard-charts").then((module) => module.ReliabilityChart),
  { loading: () => <div className="h-64 w-full animate-pulse rounded-md bg-muted" aria-hidden="true" /> },
);

export function DashboardAnalytics({
  currentPeriodStart,
  currentPeriodEnd,
  currentUsage,
  requestLimit,
  refreshKey,
}: DashboardAnalyticsProps) {
  const t = useTranslations("dashboard.analytics");
  const locale = useLocale();
  const [timeseries, setTimeseries] = useState<AnalyticsPoint[]>([]);
  const [providerBreakdown, setProviderBreakdown] = useState<AnalyticsPoint[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isUnavailable, setIsUnavailable] = useState(false);

  const range = useMemo(
    () => resolveAnalyticsRange(currentPeriodStart, currentPeriodEnd),
    [currentPeriodEnd, currentPeriodStart],
  );

  useEffect(() => {
    const controller = new AbortController();
    const query = new URLSearchParams({
      from: range.from,
      to: range.to,
    });
    const timeseriesQuery = new URLSearchParams(query);
    timeseriesQuery.set("type", "timeseries");
    timeseriesQuery.set("group_by", "");
    timeseriesQuery.set("interval_seconds", String(DAY_MS / 1_000));
    const providerQuery = new URLSearchParams(query);
    providerQuery.set("type", "breakdown");
    providerQuery.set("group_by", "provider");

    void Promise.all([
      getDashboardAnalytics(timeseriesQuery, controller.signal),
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
  }, [range.from, range.to, refreshKey]);

  const dateFormatter = useMemo(
    () => new Intl.DateTimeFormat(locale, { day: "numeric", month: "short" }),
    [locale],
  );
  const usageData = useMemo(
    () => buildUsageData(timeseries, dateFormatter, range, currentUsage),
    [currentUsage, dateFormatter, range, timeseries],
  );
  const reliabilityData = useMemo(
    () => buildReliabilityData(timeseries, dateFormatter),
    [dateFormatter, timeseries],
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
  const projectedUsage = usageData.at(-1)?.projected ?? null;
  const capacityUsage =
    requestLimit && currentUsage !== null
      ? Math.max(0, (currentUsage / requestLimit) * 100)
      : null;

  const chartConfig = useMemo(
    () =>
      ({
        cumulative: { label: t("used"), color: "var(--chart-1)" },
        projected: { label: t("forecast"), color: "var(--chart-3)" },
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
    <section className="flex flex-col gap-4" aria-label={t("sectionLabel")}>
      <AppCard
        title={t("usageTitle")}
        description={t("usageDescription")}
        actions={
          <div className="flex flex-wrap gap-2">
            <Badge variant="outline">{t("currentPeriod")}</Badge>
            {capacityUsage !== null ? (
              <Badge variant={capacityUsage >= 85 ? "warning" : "secondary"}>
                {t("capacityBadge", {
                  value: new Intl.NumberFormat(locale, { maximumFractionDigits: 1 }).format(capacityUsage),
                })}
              </Badge>
            ) : null}
            {projectedUsage !== null ? (
              <Badge variant={requestLimit !== null && projectedUsage > requestLimit ? "warning" : "secondary"}>
                {t("forecastBadge", {
                  value: compactNumber(projectedUsage),
                })}
              </Badge>
            ) : (
              <Badge variant="outline">{t("forecastUnavailableBadge")}</Badge>
            )}
          </div>
        }
        className="min-w-0 overflow-hidden"
        contentClassName="min-w-0"
      >
        {usageData.length === 0 ? (
          <ChartEmptyState title={t("emptyTitle")} description={t("emptyDescription")} />
        ) : (
          <UsageChart
            data={usageData}
            config={chartConfig}
            ariaLabel={t("usageAriaLabel")}
          />
        )}
      </AppCard>

      <div className="grid items-stretch gap-4 xl:grid-cols-2">
        <AppCard
          title={t("providerCostTitle")}
          description={t("providerCostDescription")}
          className="h-full min-w-0 overflow-hidden"
          contentClassName="min-w-0"
        >
          {providerCostData.length === 0 ? (
            <ChartEmptyState title={t("emptyCostTitle")} description={t("emptyCostDescription")} />
          ) : (
            <div className="grid items-center gap-4 sm:grid-cols-[minmax(0,1fr)_minmax(10rem,0.8fr)]">
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

              <div className="flex min-w-0 flex-col gap-3">
                {providerCostData.map((item) => (
                  <div key={item.provider} className="flex min-w-0 items-center justify-between gap-3 text-sm">
                    <span className="flex min-w-0 items-center gap-2">
                      <span className="size-2.5 shrink-0 rounded-full" style={{ backgroundColor: item.fill }} />
                      <span className="truncate text-muted-foreground" dir="auto" title={item.provider}>
                        {item.provider}
                      </span>
                    </span>
                    <span className="shrink-0 font-medium tabular-nums text-foreground" dir="ltr">
                      {compactCurrency(item.cost)}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </AppCard>

        <AppCard
          title={t("reliabilityTitle")}
          description={t("reliabilityDescription")}
          actions={
            successRate !== null ? (
              <Badge variant={successRate >= 99 ? "success" : successRate >= 95 ? "secondary" : "warning"}>
                {t("successRateBadge", {
                  value: new Intl.NumberFormat(locale, { maximumFractionDigits: 1 }).format(successRate),
                })}
              </Badge>
            ) : null
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
      </div>
    </section>
  );
}

function DashboardAnalyticsSkeleton() {
  return (
    <section className="flex flex-col gap-4" aria-hidden="true">
      <AppCard className="overflow-hidden">
        <div className="flex flex-col gap-4">
          <Skeleton className="h-5 w-48" />
          <Skeleton className="h-72 w-full" />
        </div>
      </AppCard>
      <div className="grid gap-4 xl:grid-cols-2">
        {[0, 1].map((item) => (
          <AppCard key={item} className="overflow-hidden">
            <div className="flex flex-col gap-4">
              <Skeleton className="h-5 w-40" />
              <Skeleton className="h-64 w-full" />
            </div>
          </AppCard>
        ))}
      </div>
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

function resolveAnalyticsRange(currentPeriodStart?: string, currentPeriodEnd?: string) {
  const now = new Date();
  const fallbackStart = new Date(now.getTime() - 30 * DAY_MS);
  const start = parseValidDate(currentPeriodStart) ?? fallbackStart;
  const configuredEnd = parseValidDate(currentPeriodEnd);
  const periodEnd = configuredEnd && configuredEnd > now ? configuredEnd : now;

  return {
    from: start.toISOString(),
    to: now.toISOString(),
    start,
    periodEnd,
  };
}

function parseValidDate(value?: string) {
  if (!value) return null;
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? null : date;
}

function buildUsageData(
  points: AnalyticsPoint[],
  dateFormatter: Intl.DateTimeFormat,
  range: ReturnType<typeof resolveAnalyticsRange>,
  currentUsage: number | null,
): UsageChartPoint[] {
  const ordered = points
    .map((point) => ({ point, date: parseValidDate(point.timestamp) }))
    .filter((item): item is { point: AnalyticsPoint; date: Date } => item.date !== null)
    .sort((a, b) => a.date.getTime() - b.date.getTime());

  const observedUsage = ordered.reduce(
    (total, item) => total + Math.max(0, Number(item.point.request_count ?? 0)),
    0,
  );
  let cumulative = currentUsage === null ? 0 : Math.max(0, currentUsage - observedUsage);
  const actual: UsageChartPoint[] = ordered.map(({ point, date }) => {
    cumulative += Math.max(0, Number(point.request_count ?? 0));
    return {
      label: dateFormatter.format(date),
      cumulative,
      projected: null,
    };
  });

  if (actual.length === 0) return [];

  const baseline = currentUsage ?? cumulative;
  const now = new Date(range.to);
  const last = actual.at(-1);

  if (last) last.cumulative = baseline;

  const completeDays = ordered
    .filter((item) => item.date.getTime() + DAY_MS <= now.getTime())
    .slice(-14);
  if (completeDays.length < 7 || range.periodEnd.getTime() <= now.getTime() + DAY_MS) return actual;

  const dailyRate = calculateTheilSenDailyRate(completeDays);
  if (dailyRate === null) return actual;

  const remainingDays = Math.max(0, (range.periodEnd.getTime() - now.getTime()) / DAY_MS);
  const projectedTotal = Math.max(baseline, Math.round(baseline + dailyRate * remainingDays));

  if (last) last.projected = baseline;
  actual.push({
    label: dateFormatter.format(range.periodEnd),
    cumulative: null,
    projected: projectedTotal,
  });

  return actual;
}

function calculateTheilSenDailyRate(
  samples: Array<{ point: AnalyticsPoint; date: Date }>,
): number | null {
  let cumulative = 0;
  const cumulativeSamples = samples.map(({ point, date }) => {
    cumulative += Math.max(0, Number(point.request_count ?? 0));
    return { timestamp: date.getTime(), cumulative };
  });
  const slopes: number[] = [];

  for (let startIndex = 0; startIndex < cumulativeSamples.length - 1; startIndex += 1) {
    for (let endIndex = startIndex + 1; endIndex < cumulativeSamples.length; endIndex += 1) {
      const start = cumulativeSamples[startIndex];
      const end = cumulativeSamples[endIndex];
      const elapsedDays = (end.timestamp - start.timestamp) / DAY_MS;
      if (elapsedDays <= 0) continue;
      slopes.push((end.cumulative - start.cumulative) / elapsedDays);
    }
  }

  if (slopes.length === 0) return null;
  slopes.sort((a, b) => a - b);
  const middle = Math.floor(slopes.length / 2);
  const median =
    slopes.length % 2 === 0
      ? (slopes[middle - 1] + slopes[middle]) / 2
      : slopes[middle];
  return Number.isFinite(median) ? Math.max(0, median) : null;
}

function buildReliabilityData(
  points: AnalyticsPoint[],
  dateFormatter: Intl.DateTimeFormat,
): ReliabilityChartPoint[] {
  return [...points]
    .filter((point) => parseValidDate(point.timestamp))
    .sort((a, b) => new Date(a.timestamp ?? 0).getTime() - new Date(b.timestamp ?? 0).getTime())
    .map((point) => {
      const requests = Math.max(0, Number(point.request_count ?? 0));
      const errors = Math.max(0, Math.min(requests, Number(point.error_count ?? 0)));
      return {
        label: dateFormatter.format(new Date(point.timestamp ?? 0)),
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

  return visible.map((point, index) => ({
    ...point,
    fill: PROVIDER_COLORS[index % PROVIDER_COLORS.length],
  }));
}

function compactNumber(value: number) {
  return formatCompactNumber(value, {
    locale: "en-US",
    compactThreshold: 1_000,
    maximumFractionDigits: 1,
  }).value;
}

function compactCurrency(value: number) {
  return formatCompactCurrency(value, {
    locale: "en-US",
    currency: "USD",
    compactThreshold: 1_000,
    maximumFractionDigits: 2,
  }).value;
}
