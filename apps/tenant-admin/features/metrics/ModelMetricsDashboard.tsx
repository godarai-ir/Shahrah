"use client";

import { Fragment, useEffect, useMemo, useRef, useState, type ReactNode } from "react";
import { ChevronDown, Info } from "lucide-react";
import dynamic from "next/dynamic";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { useLocale, useTranslations } from "next-intl";

import { AppCard } from "@/components/app-ui/AppCard";
import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { CompactValue } from "@/components/app-ui/CompactValue";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { MetricCard } from "@/components/app-ui/MetricCard";
import { ObserveQueryBar } from "@/components/app-ui/ObserveQueryBar";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import type { ChartConfig } from "@/components/ui/chart";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { formatCompactCurrency, formatCompactNumber } from "@/lib/compact-number";
import {
  mergeObserveSearchParams,
  observeFiltersFromSearchParams,
  resolveObserveRange,
  type ObserveFilters,
} from "@/lib/observe-filters";
import {
  METRIC_KINDS,
  METRICS_PERCENTILES,
  METRICS_VIEW_BY,
  buildBreakdownRows,
  buildCacheBreakdown,
  buildMetricsFilters,
  buildChartData,
  buildSavingsChartData,
  latencyValue,
  metricNumber,
  metricString,
  normalizeStringList,
  type BreakdownRow,
  type MetricKind,
  type MetricsPercentile as Pct,
  type MetricsRow as Row,
  type MetricsViewBy as ViewBy,
} from "./utils";
import { getMetricsFacets, queryMetrics } from "./api";

const MIN_INTERVAL_SECONDS = 300;
const MAX_CHART_POINTS = 360;
const MAX_BREAKDOWN_ROWS = 10;
const latencyAggregations = METRICS_PERCENTILES.map((percentile) => ({ type: percentile, column: "latencyMs" }));

const FocusMetricChart = dynamic(
  () => import("@/features/metrics/metrics-charts").then((module) => module.FocusMetricChart),
  { loading: () => <ChartSkeleton /> },
);
const MetricsDetailCharts = dynamic(
  () => import("@/features/metrics/metrics-charts").then((module) => module.MetricsDetailCharts),
  { loading: () => <ChartSkeleton /> },
);

type Summary = {
  totalRequests: number;
  inputTokens: number;
  outputTokens: number;
  cost: number;
  avoidedCost: number;
  costWithoutCache: number;
  savingsRate: number;
  cacheHits: number;
  cacheHitRate: number;
  avoidedTokens: number;
};

type MetricsData = {
  distribution: Row[];
  overall: Row | undefined;
  timeseries: Row[];
  errorBreakdown: Row[];
  cacheBreakdown: Row[];
  cacheTimeseries: Row[];
};

const EMPTY_DATA: MetricsData = { distribution: [], overall: undefined, timeseries: [], errorBreakdown: [], cacheBreakdown: [], cacheTimeseries: [] };

export default function ModelMetricsDashboard() {
  const locale = useLocale();
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const t = useTranslations("metrics");
  const common = useTranslations("common");

  const [observe, setObserve] = useState<ObserveFilters>(() => observeFiltersFromSearchParams(searchParams));
  const [viewBy, setViewBy] = useState<ViewBy>("modelName");
  const [metric, setMetric] = useState<MetricKind>("requests");
  const [pct, setPct] = useState<Pct>("p99");
  // Relative ranges are re-anchored to "now" on every refresh.
  const [rangeAnchor, setRangeAnchor] = useState(() => Date.now());
  const [historicalModels, setHistoricalModels] = useState<string[]>([]);
  const [historicalProviders, setHistoricalProviders] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [hasLoaded, setHasLoaded] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [data, setData] = useState<MetricsData>(EMPTY_DATA);
  const [displayViewBy, setDisplayViewBy] = useState<ViewBy>("modelName");
  const requestRef = useRef(0);

  const range = useMemo(() => resolveObserveRange(observe, new Date(rangeAnchor)), [observe, rangeAnchor]);

  useEffect(() => {
    const nextQuery = mergeObserveSearchParams(searchParams, observe).toString();
    if (nextQuery === searchParams.toString()) return;
    router.replace(nextQuery ? `${pathname}?${nextQuery}` : pathname, { scroll: false });
  }, [observe, pathname, router, searchParams]);

  useEffect(() => {
    if (!range.from || !range.to) return;
    const requestID = ++requestRef.current;
    const isCurrent = () => requestRef.current === requestID;
    const filters = buildMetricsFilters(observe);
    const base = { startTs: range.from, endTs: range.to, datasource: "modelMetrics", filters };
    const intervalInSeconds = resolveMetricsIntervalSeconds(range.from, range.to);

    queueMicrotask(() => {
      if (!isCurrent()) return;
      setLoading(true);
      setError(null);
      Promise.all([
        queryMetrics({
          ...base,
          type: "distribution",
          aggregations: [
            { type: "count", column: "modelName" },
            { type: "sum", column: "inputTokens" },
            { type: "sum", column: "outputTokens" },
            { type: "sum", column: "costInUSD" },
            { type: "sum", column: "avoidedCostUSD" },
            { type: "sum", column: "avoidedInputTokens" },
            { type: "sum", column: "avoidedOutputTokens" },
            ...latencyAggregations,
          ],
          groupBy: [viewBy],
        }),
        // Percentiles do not add up across groups, so the overall latency needs its own ungrouped query.
        // Optional: without it the latency breakdown shows no ratio, but the page still loads.
        queryMetrics({ ...base, type: "distribution", aggregations: latencyAggregations, groupBy: [] }).catch(() => ({ data: [] as Row[] })),
        queryMetrics({
          ...base,
          type: "timeseries",
          aggregations: [
            { type: "count", column: "modelName" },
            ...latencyAggregations,
            { type: "sum", column: "inputTokens" },
            { type: "sum", column: "outputTokens" },
            { type: "sum", column: "costInUSD" },
            { type: "sum", column: "avoidedCostUSD" },
          ],
          groupBy: [viewBy],
          intervalInSeconds,
        }),
        queryMetrics({
          ...base,
          type: "distribution",
          aggregations: [{ type: "count", column: "modelName" }],
          groupBy: ["errorCode"],
          filters: [...filters, { fieldName: "statusCode", operator: "GREATER_THAN_EQUAL", value: 400 }],
        }),
        queryMetrics({
          ...base,
          type: "distribution",
          aggregations: [{ type: "count", column: "modelName" }, { type: "sum", column: "avoidedCostUSD" }],
          groupBy: ["cacheMode", "cacheStatus", "cacheMatch"],
        }),
        queryMetrics({
          ...base,
          type: "timeseries",
          aggregations: [{ type: "sum", column: "costInUSD" }, { type: "sum", column: "avoidedCostUSD" }],
          groupBy: ["cacheMode"],
          intervalInSeconds,
        }),
      ])
        .then(([dist, overall, ts, errb, cacheDist, cacheTs]) => {
          if (!isCurrent()) return;
          setDisplayViewBy(viewBy);
          setData({
            distribution: dist.data ?? [],
            overall: overall.data?.[0],
            timeseries: ts.data ?? [],
            errorBreakdown: errb.data ?? [],
            cacheBreakdown: cacheDist.data ?? [],
            cacheTimeseries: cacheTs.data ?? [],
          });
        })
        .catch((e: unknown) => {
          if (!isCurrent()) return;
          setError(e instanceof Error ? e.message : common("requestError"));
          setData(EMPTY_DATA);
        })
        .finally(() => {
          if (!isCurrent()) return;
          setHasLoaded(true);
          setLoading(false);
        });
    });
  }, [common, observe, range, viewBy]);

  useEffect(() => {
    let disposed = false;
    getMetricsFacets()
      .then((payload) => {
        if (disposed) return;
        setHistoricalModels(normalizeStringList(payload.models));
        setHistoricalProviders(normalizeStringList(payload.providers));
      })
      .catch(() => undefined);
    return () => {
      disposed = true;
    };
  }, []);

  const summary = useMemo(() => buildSummary(data.distribution, data.cacheBreakdown), [data.distribution, data.cacheBreakdown]);
  const cacheBreakdownItems = useMemo(() => buildCacheBreakdown(data.cacheBreakdown), [data.cacheBreakdown]);
  const savingsChartData = useMemo(() => buildSavingsChartData(data.cacheTimeseries, locale, range.from, range.to), [data.cacheTimeseries, locale, range]);
  const chartData = useMemo(
    () => buildChartData(data.timeseries, displayViewBy, pct, t("values.unknown"), locale, range.from, range.to),
    [data.timeseries, displayViewBy, pct, t, locale, range],
  );
  const breakdownRows = useMemo(
    () => buildBreakdownRows(data.distribution, displayViewBy, metric, pct, latencyValue(data.overall, pct), t("values.unknown")).slice(0, MAX_BREAKDOWN_ROWS),
    [data.distribution, data.overall, displayViewBy, metric, pct, t],
  );
  const errorChartData = useMemo(
    () => data.errorBreakdown.map((row) => ({
      errorCode: metricString(row, "errorCode", "statusCode") || t("values.unknown"),
      count: metricNumber(row, "countModelName", "countAll"),
    })),
    [data.errorBreakdown, t],
  );
  const seriesConfig = useMemo<ChartConfig>(
    () => chartData.series.reduce<ChartConfig>((config, series, index) => {
      config[series.key] = { label: series.label, color: `var(--chart-${(index % 5) + 1})` };
      return config;
    }, {}),
    [chartData.series],
  );

  const dimension = t(`filters.viewByOptions.${displayViewBy}`);
  const focusTitle = metric === "latency"
    ? t("focus.latencyTitle", { percentile: t(`filters.percentiles.${pct}`), dimension })
    : t("focus.title", { metric: t(`controls.metricKinds.${metric}`), dimension });
  const formatValue = (value: number) => formatMetricValue(value, metric, locale, t("units.ms"));

  return (
    <div className="flex flex-col gap-4">
      <ObserveQueryBar
        tab="metrics"
        filters={observe}
        onFiltersChange={setObserve}
        modelOptions={historicalModels}
        providerOptions={historicalProviders}
        onRefresh={() => setRangeAnchor(Date.now())}
        loading={loading}
      />

      {error ? (
        <Alert variant="destructive">
          <AlertTitle>{t("alerts.errorTitle")}</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}

      {hasLoaded && !loading && !error && data.distribution.length === 0 ? (
        <EmptyState title={t("table.empty")} className="min-h-32" />
      ) : null}

      <MetricsRangeSummary summary={summary} locale={locale}>
        <section aria-label={t("sections.summary")} className="grid gap-2 sm:grid-cols-2 xl:grid-cols-5">
          <SummaryCards summary={summary} locale={locale} />
        </section>
        <MetricsDetailCharts
          failuresTitle={t("charts.failures.title")}
          savingsTitle={t("charts.savings.title")}
          locale={locale}
          errorChartData={errorChartData}
          savingsChartData={savingsChartData}
          errorChartConfig={{ count: { label: t("charts.failures.count"), color: "var(--destructive)" } }}
          savingsChartConfig={{
            actualCost: { label: t("charts.savings.actualCost"), color: "var(--chart-1)" },
            simpleSavings: { label: t("charts.savings.simpleSavings"), color: "var(--chart-2)" },
            smartSavings: { label: t("charts.savings.smartSavings"), color: "var(--chart-3)" },
          }}
        />
        <CacheTable items={cacheBreakdownItems} locale={locale} />
      </MetricsRangeSummary>

      <div className="flex flex-wrap items-center gap-x-4 gap-y-2">
        <Segmented
          label={t("controls.metric")}
          value={metric}
          options={METRIC_KINDS.map((value) => ({ value, label: t(`controls.metricKinds.${value}`) }))}
          onChange={setMetric}
        />
        <Segmented
          label={t("controls.groupBy")}
          value={viewBy}
          options={METRICS_VIEW_BY.map((value) => ({ value, label: t(`filters.viewByOptions.${value}`) }))}
          onChange={setViewBy}
        />
      </div>

      <FocusMetricChart
        title={focusTitle}
        kind={metric}
        locale={locale}
        chartData={chartData}
        seriesConfig={seriesConfig}
        action={metric === "latency" ? (
          <AppSelectField
            value={pct}
            onValueChange={(value) => setPct(value as Pct)}
            options={METRICS_PERCENTILES.map((value) => ({ value, label: t(`filters.percentiles.${value}`) }))}
            ariaLabel={t("filters.percentile")}
            triggerClassName="h-8 w-24 text-xs"
          />
        ) : null}
      />

      <BreakdownTable
        rows={breakdownRows}
        metric={metric}
        title={t("breakdown.title", { dimension })}
        note={metric === "latency" ? t("breakdown.latencyNote", { percentile: t(`filters.percentiles.${pct}`) }) : undefined}
        formatValue={formatValue}
        locale={locale}
      />
    </div>
  );
}

function Segmented<T extends string>({ label, value, options, onChange }: { label: string; value: T; options: Array<{ value: T; label: string }>; onChange: (value: T) => void }) {
  return (
    <div className="flex min-w-0 flex-wrap items-center gap-2">
      <span className="text-xs text-muted-foreground">{label}</span>
      <ToggleGroup
        type="single"
        value={value}
        onValueChange={(next) => {
          const option = options.find((item) => item.value === next);
          if (option) onChange(option.value);
        }}
        aria-label={label}
        className="flex-wrap rounded-lg border bg-muted/40 p-0.5"
      >
        {options.map((option) => (
          <ToggleGroupItem
            key={option.value}
            value={option.value}
            className="h-7 px-2.5 text-xs text-muted-foreground data-[state=on]:bg-background data-[state=on]:text-foreground data-[state=on]:shadow-xs"
          >
            {option.label}
          </ToggleGroupItem>
        ))}
      </ToggleGroup>
    </div>
  );
}

function BreakdownTable({ rows, metric, title, note, formatValue, locale }: { rows: BreakdownRow[]; metric: MetricKind; title: string; note?: string; formatValue: (value: number) => ReactNode; locale: string }) {
  const t = useTranslations("metrics");
  const percent = new Intl.NumberFormat(locale, { style: "percent", maximumFractionDigits: 1 });
  const ratio = new Intl.NumberFormat(locale, { maximumFractionDigits: 2 });
  const isLatency = metric === "latency";

  return (
    <AppCard title={title} description={note} contentClassName="overflow-x-auto" className="shadow-none">
      {rows.length === 0 ? (
        <p className="py-6 text-center text-sm text-muted-foreground">{t("breakdown.empty")}</p>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t("breakdown.group")}</TableHead>
              <TableHead>{t("breakdown.value")}</TableHead>
              <TableHead className="w-2/5">{isLatency ? t("breakdown.ratio") : t("breakdown.share")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {rows.map((row) => (
              <TableRow key={row.label}>
                <TableCell className="max-w-64 truncate" dir="auto">{row.label}</TableCell>
                <TableCell className="tabular-nums">{formatValue(row.value)}</TableCell>
                <TableCell>
                  {isLatency ? (
                    <span className="tabular-nums">{row.ratio === undefined ? "-" : t("breakdown.ratioValue", { value: ratio.format(row.ratio) })}</span>
                  ) : (
                    <div className="flex items-center gap-2">
                      <span className="h-1.5 min-w-16 flex-1 overflow-hidden rounded-full bg-muted" aria-hidden="true">
                        <span className="block h-full rounded-full bg-primary" style={{ width: `${Math.round((row.share ?? 0) * 100)}%` }} />
                      </span>
                      <span className="w-14 shrink-0 text-end tabular-nums">{percent.format(row.share ?? 0)}</span>
                    </div>
                  )}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
    </AppCard>
  );
}

/** Collapsed by default: the headline numbers on one line, the cards and secondary charts on open. */
function MetricsRangeSummary({ summary, locale, children }: { summary: Summary; locale: string; children: ReactNode }) {
  const t = useTranslations("metrics");
  const [hasOpened, setHasOpened] = useState(false);
  const percent = (value: number) => formatPercent(value, locale);
  const items = [
    { key: "requests", label: t("rangeSummary.requests"), value: <CompactValue display={formatCompactNumber(summary.totalRequests, { locale })} /> },
    { key: "cost", label: t("rangeSummary.cost"), value: <CompactValue display={formatCompactCurrency(summary.cost, { locale, currency: "USD", maximumFractionDigits: 6 })} /> },
    { key: "cacheHitRate", label: t("rangeSummary.cacheHitRate"), value: percent(summary.cacheHitRate) },
    { key: "avoidedCost", label: t("rangeSummary.avoidedCost"), value: <CompactValue display={formatCompactCurrency(summary.avoidedCost, { locale, currency: "USD", maximumFractionDigits: 6 })} /> },
  ];

  return (
    <details
      onToggle={(event) => {
        if (event.currentTarget.open) setHasOpened(true);
      }}
      className="group rounded-lg border border-card-border bg-card text-card-foreground shadow-xs"
      data-testid="metrics-range-summary"
    >
      <summary className="flex min-h-11 cursor-pointer list-none items-center gap-3 rounded-lg px-4 py-2 text-xs outline-none focus-visible:ring-2 focus-visible:ring-ring [&::-webkit-details-marker]:hidden">
        <ChevronDown aria-hidden="true" className="size-4 shrink-0 text-muted-foreground transition-transform ltr:-rotate-90 rtl:rotate-90 group-open:ltr:rotate-0 group-open:rtl:rotate-0" />
        <span className="shrink-0 text-sm font-semibold text-foreground">{t("rangeSummary.title")}</span>
        <dl className="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1">
          {items.map((item, index) => (
            <Fragment key={item.key}>
              {index > 0 ? <span aria-hidden="true" className="text-muted-foreground">·</span> : null}
              <div className="flex items-center gap-1">
                <dt className="text-muted-foreground">{item.label}</dt>
                <dd className="font-medium tabular-nums text-foreground">{item.value}</dd>
              </div>
            </Fragment>
          ))}
        </dl>
      </summary>
      {hasOpened ? <div className="flex flex-col gap-4 border-t border-border p-4">{children}</div> : null}
    </details>
  );
}

function SummaryCards({ summary, locale }: { summary: Summary; locale: string }) {
  const t = useTranslations("metrics");
  const usd = (value: number) => <CompactValue display={formatCompactCurrency(value, { locale, currency: "USD", maximumFractionDigits: 6 })} />;
  const count = (value: number) => <CompactValue display={formatCompactNumber(value, { locale })} />;
  const cards: Array<{ key: string; label: ReactNode; value: ReactNode }> = [
    { key: "totalRequests", label: t("cards.totalRequests"), value: count(summary.totalRequests) },
    { key: "inputTokens", label: t("cards.inputTokens"), value: count(summary.inputTokens) },
    { key: "outputTokens", label: t("cards.outputTokens"), value: count(summary.outputTokens) },
    { key: "totalCost", label: t("cards.totalCost"), value: usd(summary.cost) },
    { key: "avoidedCost", label: <InfoMetricLabel label={t("cards.avoidedCost")} tooltip={t("cards.avoidedCostHint")} />, value: usd(summary.avoidedCost) },
    { key: "costWithoutCache", label: t("cards.costWithoutCache"), value: usd(summary.costWithoutCache) },
    { key: "savingsRate", label: t("cards.savingsRate"), value: formatPercent(summary.savingsRate, locale) },
    { key: "cacheHits", label: t("cards.cacheHits"), value: count(summary.cacheHits) },
    { key: "cacheHitRate", label: t("cards.cacheHitRate"), value: formatPercent(summary.cacheHitRate, locale) },
    { key: "cachedTokens", label: <InfoMetricLabel label={t("cards.cachedTokens")} tooltip={t("cards.cachedTokensHint")} />, value: count(summary.avoidedTokens) },
  ];
  return <>{cards.map((card) => <MetricCard key={card.key} label={card.label} value={card.value} density="compact" className="shadow-none" />)}</>;
}

function CacheTable({ items, locale }: { items: ReturnType<typeof buildCacheBreakdown>; locale: string }) {
  const t = useTranslations("metrics");
  return (
    <AppCard title={t("cacheTable.title")} description={t("cacheTable.description")} contentClassName="overflow-x-auto" className="shadow-none">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{t("cacheTable.mode")}</TableHead>
            <TableHead>{t("cacheTable.exactHits")}</TableHead>
            <TableHead>{t("cacheTable.semanticHits")}</TableHead>
            <TableHead>{t("cacheTable.hitRate")}</TableHead>
            <TableHead>{t("cacheTable.avoidedCost")}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {items.map((item) => {
            const hits = item.exactHits + item.semanticHits;
            const hitRate = hits + item.misses > 0 ? hits / (hits + item.misses) : 0;
            return (
              <TableRow key={item.mode}>
                <TableCell>{t(`cacheModes.${item.mode}`)}</TableCell>
                <TableCell><CompactValue display={formatCompactNumber(item.exactHits, { locale })} /></TableCell>
                <TableCell><CompactValue display={formatCompactNumber(item.semanticHits, { locale })} /></TableCell>
                <TableCell>{formatPercent(hitRate, locale)}</TableCell>
                <TableCell><CompactValue display={formatCompactCurrency(item.avoidedCost, { locale, currency: "USD", maximumFractionDigits: 6 })} /></TableCell>
              </TableRow>
            );
          })}
        </TableBody>
      </Table>
    </AppCard>
  );
}

function InfoMetricLabel({ label, tooltip }: { label: string; tooltip: string }) {
  return (
    <span className="inline-flex items-center gap-1.5 align-middle">
      <span>{label}</span>
      <TooltipProvider delayDuration={120}>
        <Tooltip>
          <TooltipTrigger asChild>
            <button
              type="button"
              className="inline-flex size-5 shrink-0 items-center justify-center rounded-full text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background"
              aria-label={tooltip}
            >
              <Info className="size-3.5" aria-hidden="true" />
            </button>
          </TooltipTrigger>
          <TooltipContent side="top" className="max-w-72 text-start leading-5">
            {tooltip}
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    </span>
  );
}

function ChartSkeleton() {
  return <div className="h-72 w-full animate-pulse rounded-lg bg-muted md:h-80" aria-hidden="true" />;
}

function buildSummary(distribution: Row[], cacheBreakdown: Row[]): Summary {
  const sum = (...keys: string[]) => distribution.reduce((acc, row) => acc + metricNumber(row, ...keys), 0);
  const cost = sum("sumCostInUSD", "sumCostUsd");
  const avoidedCost = sum("sumAvoidedCostUSD");
  const costWithoutCache = cost + avoidedCost;
  const countWithStatus = (status: string) => cacheBreakdown.reduce(
    (acc, row) => acc + (metricString(row, "cacheStatus") === status ? metricNumber(row, "countModelName", "countAll") : 0),
    0,
  );
  const cacheHits = countWithStatus("hit");
  const cacheMisses = countWithStatus("miss");
  return {
    totalRequests: sum("countModelName", "countAll"),
    inputTokens: sum("sumInputTokens"),
    outputTokens: sum("sumOutputTokens"),
    cost,
    avoidedCost,
    costWithoutCache,
    savingsRate: costWithoutCache > 0 ? avoidedCost / costWithoutCache : 0,
    cacheHits,
    cacheHitRate: cacheHits + cacheMisses > 0 ? cacheHits / (cacheHits + cacheMisses) : 0,
    avoidedTokens: sum("sumAvoidedInputTokens") + sum("sumAvoidedOutputTokens"),
  };
}

function formatMetricValue(value: number, metric: MetricKind, locale: string, msLabel: string): ReactNode {
  if (metric === "cost") return <CompactValue display={formatCompactCurrency(value, { locale, currency: "USD", maximumFractionDigits: 6 })} />;
  if (metric === "latency") return `${new Intl.NumberFormat(locale, { maximumFractionDigits: 0 }).format(value)} ${msLabel}`;
  return <CompactValue display={formatCompactNumber(value, { locale })} />;
}

export function resolveMetricsIntervalSeconds(from: string, to: string) {
  const durationSeconds = Math.max(0, (new Date(to).getTime() - new Date(from).getTime()) / 1_000);
  if (!Number.isFinite(durationSeconds)) return MIN_INTERVAL_SECONDS;
  const targetInterval = Math.ceil(durationSeconds / MAX_CHART_POINTS);
  return Math.max(MIN_INTERVAL_SECONDS, Math.ceil(targetInterval / 60) * 60);
}

function formatPercent(value: number, locale: string) {
  return new Intl.NumberFormat(locale, { style: "percent", maximumFractionDigits: 1 }).format(value);
}
