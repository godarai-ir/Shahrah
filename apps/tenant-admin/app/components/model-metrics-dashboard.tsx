"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { Info, RotateCcw } from "lucide-react";
import { useLocale, useTranslations } from "next-intl";
import {
  Area,
  AreaChart,
  Bar,
  BarChart,
  CartesianGrid,
  ComposedChart,
  Line,
  XAxis,
  YAxis,
} from "recharts";

import { AppCard, CatalogFilterCombobox, CompactValue, DateTimeRangeField, EmptyState, FilterCombobox, MetricCard, PageHeader } from "@/components/app-ui";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  type ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";
import { Input } from "@/components/ui/input";
import { Field, FieldLabel } from "@/components/ui/field";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { formatCompactCurrency, formatCompactNumber } from "@/lib/compact-number";
import {
  buildCacheBreakdown,
  buildChartData,
  buildSavingsChartData,
  metricNumber,
  metricString,
  normalizeStringList,
  type MetricsRow as Row,
} from "./model-metrics-utils";

const VIEW_BY = ["modelName", "virtualModelName", "userEmail", "virtualAccount", "team"] as const;
const PCTS = ["p50", "p75", "p90", "p99"] as const;
const RANGE_PRESETS = ["last15m", "last1h", "last24h", "yesterday", "last3d", "last7d", "last14d", "last30d", "custom"] as const;
const INTERVAL_SECONDS = 300;
const DEFAULT_RANGE_MS = 24 * 60 * 60 * 1000;
const ALL_MODELS_VALUE = "__all_models__";
const ALL_PROVIDERS_VALUE = "__all_providers__";

type ViewBy = (typeof VIEW_BY)[number];
type Pct = (typeof PCTS)[number];
type RangePreset = (typeof RANGE_PRESETS)[number];

type ModelMetricsDashboardProps = {
  initialFrom?: string;
  initialTo?: string;
  initialModel?: string;
  initialProvider?: string;
  initialRequestID?: string;
};

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

type ModelMetricsFacetsResponse = {
  models?: unknown;
  providers?: unknown;
};

export default function ModelMetricsDashboard(props: ModelMetricsDashboardProps) {
  const locale = useLocale();
  const t = useTranslations("metrics");
  const common = useTranslations("common");

  const [viewBy, setViewBy] = useState<ViewBy>("modelName");
  const [displayViewBy, setDisplayViewBy] = useState<ViewBy>("modelName");
  const didRunInitialQuery = useRef(false);
  const [pct, setPct] = useState<Pct>("p99");
  const [modelFilter, setModelFilter] = useState(props.initialModel ?? "");
  const [providerFilter, setProviderFilter] = useState(props.initialProvider ?? "");
  const [requestIDFilter, setRequestIDFilter] = useState(props.initialRequestID ?? "");
  const [rangePreset, setRangePreset] = useState<RangePreset>(() =>
    props.initialFrom || props.initialTo ? "custom" : "last24h",
  );
  const [historicalModels, setHistoricalModels] = useState<string[]>([]);
  const [historicalProviders, setHistoricalProviders] = useState<string[]>([]);
  const [loadingOptions, setLoadingOptions] = useState(false);
  const [loading, setLoading] = useState(false);
  const [hasLoaded, setHasLoaded] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [from, setFrom] = useState(() => normalizeInitialIso(props.initialFrom));
  const [to, setTo] = useState(() => normalizeInitialIso(props.initialTo));

  const [distribution, setDistribution] = useState<Row[]>([]);
  const [timeseries, setTimeseries] = useState<Row[]>([]);
  const [errorBreakdown, setErrorBreakdown] = useState<Row[]>([]);
  const [cacheBreakdown, setCacheBreakdown] = useState<Row[]>([]);
  const [cacheTimeseries, setCacheTimeseries] = useState<Row[]>([]);

  const rangeOptions = useMemo(() => RANGE_PRESETS.map((value) => ({ value, label: t(`filters.rangePresets.${value}`) })), [t]);
  const modelOptions = useMemo(() => historicalModels.map((value) => ({ value, label: value })), [historicalModels]);
  const providerOptions = useMemo(() => historicalProviders.map((value) => ({ value, label: value })), [historicalProviders]);
  const viewByOptions = useMemo(() => VIEW_BY.map((value) => ({ value, label: t(`filters.viewByOptions.${value}`) })), [t]);
  const pctOptions = useMemo(() => PCTS.map((value) => ({ value, label: t(`filters.percentiles.${value}`) })), [t]);
  const modelComboboxOptions = useMemo(() => [{ value: ALL_MODELS_VALUE, label: t("filters.placeholders.model") }, ...modelOptions], [modelOptions, t]);
  const providerComboboxOptions = useMemo(() => [{ value: ALL_PROVIDERS_VALUE, label: t("filters.placeholders.provider") }, ...providerOptions], [providerOptions, t]);

  const summary = useMemo<Summary>(() => {
    const totalRequests = distribution.reduce((acc, row) => acc + metricNumber(row, "countModelName", "countAll"), 0);
    const inputTokens = distribution.reduce((acc, row) => acc + metricNumber(row, "sumInputTokens"), 0);
    const outputTokens = distribution.reduce((acc, row) => acc + metricNumber(row, "sumOutputTokens"), 0);
    const cost = distribution.reduce((acc, row) => acc + metricNumber(row, "sumCostInUSD", "sumCostUsd"), 0);
    const avoidedCost = distribution.reduce((acc, row) => acc + metricNumber(row, "sumAvoidedCostUSD"), 0);
    const avoidedTokens = distribution.reduce(
      (acc, row) => acc + metricNumber(row, "sumAvoidedInputTokens") + metricNumber(row, "sumAvoidedOutputTokens"),
      0,
    );
    const costWithoutCache = cost + avoidedCost;
    const cacheHits = cacheBreakdown.reduce(
      (acc, row) => acc + (metricString(row, "cacheStatus") === "hit" ? metricNumber(row, "countModelName", "countAll") : 0),
      0,
    );
    const cacheMisses = cacheBreakdown.reduce(
      (acc, row) => acc + (metricString(row, "cacheStatus") === "miss" ? metricNumber(row, "countModelName", "countAll") : 0),
      0,
    );
    return {
      totalRequests,
      inputTokens,
      outputTokens,
      cost,
      avoidedCost,
      costWithoutCache,
      savingsRate: costWithoutCache > 0 ? avoidedCost / costWithoutCache : 0,
      cacheHits,
      cacheHitRate: cacheHits + cacheMisses > 0 ? cacheHits / (cacheHits + cacheMisses) : 0,
      avoidedTokens,
    };
  }, [distribution, cacheBreakdown]);

  const cacheBreakdownItems = useMemo(() => buildCacheBreakdown(cacheBreakdown), [cacheBreakdown]);
  const savingsChartData = useMemo(() => buildSavingsChartData(cacheTimeseries, locale, from, to), [cacheTimeseries, locale, from, to]);

  const chartData = useMemo(
    () => buildChartData(timeseries, displayViewBy, pct, t("values.unknown"), locale, from, to),
    [timeseries, displayViewBy, pct, t, locale, from, to],
  );

  const errorChartData = useMemo(
    () =>
      errorBreakdown.map((row) => ({
        errorCode: metricString(row, "errorCode", "statusCode") || t("values.unknown"),
        count: metricNumber(row, "countModelName", "countAll"),
      })),
    [errorBreakdown, t],
  );

  const seriesConfig = useMemo<ChartConfig>(() => {
    return chartData.series.reduce<ChartConfig>((config, series, index) => {
      config[series.key] = {
        label: series.label,
        color: `var(--chart-${(index % 5) + 1})`,
      };
      return config;
    }, {});
  }, [chartData.series]);

  const errorChartConfig = useMemo<ChartConfig>(
    () => ({
      count: {
        label: t("charts.failures.count"),
        color: "var(--destructive)",
      },
    }),
    [t],
  );

  const tokenChartConfig = useMemo<ChartConfig>(
    () => ({
      inputTokens: {
        label: t("charts.tokensCost.inputTokens"),
        color: "var(--chart-1)",
      },
      outputTokens: {
        label: t("charts.tokensCost.outputTokens"),
        color: "var(--chart-2)",
      },
      costUsd: {
        label: t("charts.tokensCost.costUsd"),
        color: "var(--chart-4)",
      },
    }),
    [t],
  );
  const savingsChartConfig = useMemo<ChartConfig>(
    () => ({
      actualCost: { label: t("charts.savings.actualCost"), color: "var(--chart-1)" },
      simpleSavings: { label: t("charts.savings.simpleSavings"), color: "var(--chart-2)" },
      smartSavings: { label: t("charts.savings.smartSavings"), color: "var(--chart-3)" },
    }),
    [t],
  );
  const requestChartTitle = t("charts.requests.titleByView", {
    dimension: t(`filters.viewByOptions.${displayViewBy}`),
  });

  async function runQuery(
    range?: { from: string; to: string },
    filterOverride?: { model: string; provider: string; requestID: string },
    viewByOverride?: ViewBy,
  ) {
    const resolvedRange = range ?? resolveSelectedRange(rangePreset, from, to);
    const queryFrom = resolvedRange.from;
    const queryTo = resolvedRange.to;
    const queryModelFilter = filterOverride?.model ?? modelFilter;
    const queryProviderFilter = filterOverride?.provider ?? providerFilter;
    const queryRequestIDFilter = filterOverride?.requestID ?? requestIDFilter;
    const queryViewBy = viewByOverride ?? viewBy;
    if (!queryFrom || !queryTo) return;
    if (queryFrom !== from) setFrom(queryFrom);
    if (queryTo !== to) setTo(queryTo);

    setLoading(true);
    setError(null);
    try {
      const filters: Array<Record<string, unknown>> = [];
      if (queryModelFilter.trim()) {
        filters.push({ fieldName: "modelName", operator: "STRING_CONTAINS", value: queryModelFilter.trim() });
      }
      if (queryProviderFilter.trim()) {
        filters.push({ fieldName: "providerName", operator: "EQUAL", value: queryProviderFilter.trim() });
      }
      if (queryRequestIDFilter.trim()) {
        filters.push({ fieldName: "requestID", operator: "EQUAL", value: queryRequestIDFilter.trim() });
      }

      const [dist, ts, errb, cacheDist, cacheTs] = await Promise.all([
        postMetricsQuery({
          startTs: queryFrom,
          endTs: queryTo,
          datasource: "modelMetrics",
          type: "distribution",
          aggregations: [
            { type: "count", column: "modelName" },
            { type: "sum", column: "inputTokens" },
            { type: "sum", column: "outputTokens" },
            { type: "sum", column: "costInUSD" },
            { type: "sum", column: "avoidedCostUSD" },
            { type: "sum", column: "avoidedInputTokens" },
            { type: "sum", column: "avoidedOutputTokens" },
          ],
          groupBy: [queryViewBy],
          filters,
        }, common("requestError")),
        postMetricsQuery({
          startTs: queryFrom,
          endTs: queryTo,
          datasource: "modelMetrics",
          type: "timeseries",
          aggregations: [
            { type: "count", column: "modelName" },
            ...PCTS.map((percentile) => ({ type: percentile, column: "latencyMs" })),
            { type: "sum", column: "inputTokens" },
            { type: "sum", column: "outputTokens" },
            { type: "sum", column: "costInUSD" },
            { type: "sum", column: "avoidedCostUSD" },
          ],
          groupBy: [queryViewBy],
          filters,
          intervalInSeconds: INTERVAL_SECONDS,
        }, common("requestError")),
        postMetricsQuery({
          startTs: queryFrom,
          endTs: queryTo,
          datasource: "modelMetrics",
          type: "distribution",
          aggregations: [{ type: "count", column: "modelName" }],
          groupBy: ["errorCode"],
          filters: [...filters, { fieldName: "statusCode", operator: "GREATER_THAN_EQUAL", value: 400 }],
        }, common("requestError")),
        postMetricsQuery({
          startTs: queryFrom,
          endTs: queryTo,
          datasource: "modelMetrics",
          type: "distribution",
          aggregations: [
            { type: "count", column: "modelName" },
            { type: "sum", column: "avoidedCostUSD" },
          ],
          groupBy: ["cacheMode", "cacheStatus", "cacheMatch"],
          filters,
        }, common("requestError")),
        postMetricsQuery({
          startTs: queryFrom,
          endTs: queryTo,
          datasource: "modelMetrics",
          type: "timeseries",
          aggregations: [
            { type: "sum", column: "costInUSD" },
            { type: "sum", column: "avoidedCostUSD" },
          ],
          groupBy: ["cacheMode"],
          filters,
          intervalInSeconds: INTERVAL_SECONDS,
        }, common("requestError")),
      ]);

      setDisplayViewBy(queryViewBy);
      setDistribution(dist.data ?? []);
      setTimeseries(ts.data ?? []);
      setErrorBreakdown(errb.data ?? []);
      setCacheBreakdown(cacheDist.data ?? []);
      setCacheTimeseries(cacheTs.data ?? []);
    } catch (e) {
      setError(e instanceof Error ? e.message : common("requestError"));
      setDistribution([]);
      setTimeseries([]);
      setErrorBreakdown([]);
      setCacheBreakdown([]);
      setCacheTimeseries([]);
    } finally {
      setHasLoaded(true);
      setLoading(false);
    }
  }

  async function loadMetricFacets() {
    setLoadingOptions(true);
    try {
      const payload = await requestJSON<ModelMetricsFacetsResponse>("/api/metrics/facets", common("requestError"));
      setHistoricalModels(normalizeStringList(payload.models));
      setHistoricalProviders(normalizeStringList(payload.providers));
    } catch {
      setHistoricalModels([]);
      setHistoricalProviders([]);
    } finally {
      setLoadingOptions(false);
    }
  }

  function handleRangePresetChange(nextPreset: RangePreset) {
    setRangePreset(nextPreset);
    if (nextPreset === "custom") return;
    const nextRange = rangeForPreset(nextPreset);
    setFrom(nextRange.from);
    setTo(nextRange.to);
  }

  function resetFilters() {
    const nextRange = rangeForPreset("last24h");
    setRangePreset("last24h");
    setFrom(nextRange.from);
    setTo(nextRange.to);
    setModelFilter("");
    setProviderFilter("");
    setRequestIDFilter("");
    setViewBy("modelName");
    setPct("p99");
    void runQuery(nextRange, { model: "", provider: "", requestID: "" });
  }

  function handleViewByChange(nextViewBy: ViewBy) {
    setViewBy(nextViewBy);
    void runQuery(undefined, undefined, nextViewBy);
  }

  useEffect(() => {
    const id = window.setTimeout(() => {
      if (didRunInitialQuery.current) return;
      didRunInitialQuery.current = true;

      let queryTo = to;
      let queryFrom = from;
      if (!queryTo || !queryFrom) {
        const defaultRange = rangeForPreset("last24h");
        queryTo = defaultRange.to;
        queryFrom = defaultRange.from;
        setTo(queryTo);
        setFrom(queryFrom);
      }
      void runQuery({ from: queryFrom, to: queryTo });
    }, 0);
    return () => window.clearTimeout(id);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    const id = window.setTimeout(() => {
      void loadMetricFacets();
    }, 0);
    return () => window.clearTimeout(id);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <div className="flex flex-col gap-4">
      <PageHeader title={t("title")} description={t("description")} />

      <AppCard
        contentClassName="grid gap-2 sm:grid-cols-2 lg:grid-cols-4 2xl:grid-cols-6 pt-4 md:pt-4"
        className="shadow-none"
      >
          <Field className="min-w-0 gap-1">
            <FieldLabel htmlFor="metrics-range-preset" className="text-xs text-muted-foreground">{t("filters.timeRange")}</FieldLabel>
            <FilterCombobox
              id="metrics-range-preset"
              options={rangeOptions}
              value={rangePreset}
              onValueChange={(value) => handleRangePresetChange(value as RangePreset)}
              placeholder={t("filters.timeRange")}
              emptyText={t("filters.emptyOptions")}
              searchable={false}
            />
          </Field>
          {rangePreset === "custom" ? (
            <DateTimeRangeField fromID="metrics-from" toID="metrics-to" from={from} to={to} fromLabel={t("filters.from")} toLabel={t("filters.to")} onFromChange={setFrom} onToChange={setTo} className="contents" />
          ) : null}
          <Field className="min-w-0 gap-1">
            <FieldLabel htmlFor="metrics-model" className="text-xs text-muted-foreground">{t("filters.model")}</FieldLabel>
            <CatalogFilterCombobox
              id="metrics-model"
              options={modelComboboxOptions}
              value={modelFilter || ALL_MODELS_VALUE}
              onValueChange={(value) => setModelFilter(value === ALL_MODELS_VALUE ? "" : value)}
              placeholder={loadingOptions ? common("loading") : t("filters.search")}
              emptyText={t("filters.emptyOptions")}
            />
          </Field>
          <Field className="min-w-0 gap-1">
            <FieldLabel htmlFor="metrics-provider" className="text-xs text-muted-foreground">{t("filters.provider")}</FieldLabel>
            <CatalogFilterCombobox
              id="metrics-provider"
              options={providerComboboxOptions}
              value={providerFilter || ALL_PROVIDERS_VALUE}
              onValueChange={(value) => setProviderFilter(value === ALL_PROVIDERS_VALUE ? "" : value)}
              placeholder={loadingOptions ? common("loading") : t("filters.search")}
              emptyText={t("filters.emptyOptions")}
            />
          </Field>
          <Field className="min-w-0 gap-1">
            <FieldLabel htmlFor="metrics-request-id" className="text-xs text-muted-foreground">{t("filters.requestId")}</FieldLabel>
            <Input
              id="metrics-request-id"
              value={requestIDFilter}
              onChange={(event) => setRequestIDFilter(event.target.value)}
              placeholder={t("filters.placeholders.requestId")}
              dir="ltr"
              className="h-8 px-2 text-xs font-mono text-start"
            />
          </Field>
          <Field className="min-w-0 gap-1">
            <FieldLabel htmlFor="metrics-view-by" className="text-xs text-muted-foreground">{t("filters.viewBy")}</FieldLabel>
            <FilterCombobox
              id="metrics-view-by"
              options={viewByOptions}
              value={viewBy}
              onValueChange={(value) => handleViewByChange(value as ViewBy)}
              placeholder={t("filters.viewBy")}
              emptyText={t("filters.emptyOptions")}
              searchable={false}
            />
          </Field>
          <div className="flex items-end 2xl:col-span-2">
            <div className="grid w-full grid-cols-2 gap-2 sm:flex sm:justify-end">
              <Button type="button" size="sm" onClick={() => void runQuery()} disabled={loading}>
                {loading ? common("loading") : t("actions.applyFilters")}
              </Button>
              <Button type="button" variant="outline" size="sm" onClick={resetFilters} disabled={loading}>
                <RotateCcw data-icon="inline-start" aria-hidden="true" />
                {t("actions.reset")}
              </Button>
            </div>
          </div>
      </AppCard>

      {error ? (
        <Alert variant="destructive">
          <AlertTitle>{t("alerts.errorTitle")}</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}

      {hasLoaded && !loading && !error && distribution.length === 0 ? (
        <EmptyState title={t("table.empty")} className="min-h-32" />
      ) : null}

      <section aria-label={t("sections.summary")} className="grid gap-2 sm:grid-cols-2 xl:grid-cols-5">
        <MetricCard label={t("cards.totalRequests")} value={<CompactValue display={formatCompactNumber(summary.totalRequests, { locale })} />} density="compact" className="shadow-none" />
        <MetricCard label={t("cards.inputTokens")} value={<CompactValue display={formatCompactNumber(summary.inputTokens, { locale })} />} density="compact" className="shadow-none" />
        <MetricCard label={t("cards.outputTokens")} value={<CompactValue display={formatCompactNumber(summary.outputTokens, { locale })} />} density="compact" className="shadow-none" />
        <MetricCard label={t("cards.totalCost")} value={<CompactValue display={formatCompactCurrency(summary.cost, { locale, currency: "USD", maximumFractionDigits: 6 })} />} density="compact" className="shadow-none" />
        <MetricCard label={<InfoMetricLabel label={t("cards.avoidedCost")} tooltip={t("cards.avoidedCostHint")} />} value={<CompactValue display={formatCompactCurrency(summary.avoidedCost, { locale, currency: "USD", maximumFractionDigits: 6 })} />} density="compact" className="shadow-none" />
        <MetricCard label={t("cards.costWithoutCache")} value={<CompactValue display={formatCompactCurrency(summary.costWithoutCache, { locale, currency: "USD", maximumFractionDigits: 6 })} />} density="compact" className="shadow-none" />
        <MetricCard label={t("cards.savingsRate")} value={formatPercent(summary.savingsRate, locale)} density="compact" className="shadow-none" />
        <MetricCard label={t("cards.cacheHits")} value={<CompactValue display={formatCompactNumber(summary.cacheHits, { locale })} />} density="compact" className="shadow-none" />
        <MetricCard label={t("cards.cacheHitRate")} value={formatPercent(summary.cacheHitRate, locale)} density="compact" className="shadow-none" />
        <MetricCard label={<InfoMetricLabel label={t("cards.cachedTokens")} tooltip={t("cards.cachedTokensHint")} />} value={<CompactValue display={formatCompactNumber(summary.avoidedTokens, { locale })} />} density="compact" className="shadow-none" />
      </section>

      <section aria-label={t("sections.charts")} className="grid gap-3 xl:grid-cols-2">
        <ChartCard title={requestChartTitle}>
          <ChartContainer config={seriesConfig} className="h-[240px] w-full md:h-[260px]">
            <AreaChart data={chartData.points} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
              <CartesianGrid vertical={false} />
              <XAxis dataKey="label" tickLine={false} axisLine={false} minTickGap={24} />
              <YAxis tickLine={false} axisLine={false} width={44} allowDecimals={false} tickFormatter={(value: number) => formatCompactNumber(value, { locale, compactThreshold: 1_000 }).value} />
              <ChartTooltip content={<ChartTooltipContent labelFormatter={formatTooltipLabel} />} />
              {chartData.series.map((series) => (
                <Area
                  key={series.key}
                  dataKey={series.key}
                  type="monotone"
                  fill={`var(--color-${series.key})`}
                  fillOpacity={0.16}
                  stroke={`var(--color-${series.key})`}
                  strokeWidth={2}
                />
              ))}
            </AreaChart>
          </ChartContainer>
        </ChartCard>

        <ChartCard
          title={t("charts.latency.title", { percentile: pct.toUpperCase() })}
          action={
            <Field className="w-32 gap-1">
              <FieldLabel htmlFor="metrics-percentile" className="text-xs text-muted-foreground">{t("filters.percentile")}</FieldLabel>
              <FilterCombobox
                id="metrics-percentile"
                options={pctOptions}
                value={pct}
                onValueChange={(value) => setPct(value as Pct)}
                placeholder={t("filters.percentile")}
                emptyText={t("filters.emptyOptions")}
                searchable={false}
              />
            </Field>
          }
        >
          <ChartContainer config={seriesConfig} className="h-[240px] w-full md:h-[260px]">
            <AreaChart data={chartData.latencyPoints} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
              <CartesianGrid vertical={false} />
              <XAxis dataKey="label" tickLine={false} axisLine={false} minTickGap={24} />
              <YAxis tickLine={false} axisLine={false} width={44} tickFormatter={(value: number) => formatCompactNumber(value, { locale, compactThreshold: 1_000 }).value} />
              <ChartTooltip content={<ChartTooltipContent labelFormatter={formatTooltipLabel} />} />
              {chartData.series.map((series) => (
                <Area
                  key={series.key}
                  dataKey={series.key}
                  type="monotone"
                  fill={`var(--color-${series.key})`}
                  fillOpacity={0.1}
                  stroke={`var(--color-${series.key})`}
                  strokeWidth={2}
                />
              ))}
            </AreaChart>
          </ChartContainer>
        </ChartCard>

        <ChartCard title={t("charts.tokensCost.title")}>
          <ChartContainer config={tokenChartConfig} className="h-[240px] w-full md:h-[260px]">
            <ComposedChart data={chartData.tokenCostPoints} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
              <CartesianGrid vertical={false} />
              <XAxis dataKey="label" tickLine={false} axisLine={false} minTickGap={24} />
              <YAxis yAxisId="tokens" tickLine={false} axisLine={false} width={44} tickFormatter={(value: number) => formatCompactNumber(value, { locale, compactThreshold: 1_000 }).value} />
              <YAxis yAxisId="cost" orientation="right" tickLine={false} axisLine={false} width={52} tickFormatter={(value: number) => formatCompactCurrency(value, { locale, currency: "USD", compactThreshold: 1_000, maximumFractionDigits: 6 }).value} />
              <ChartTooltip content={<ChartTooltipContent labelFormatter={formatTooltipLabel} />} />
              <Bar yAxisId="tokens" dataKey="inputTokens" stackId="tokens" fill="var(--color-inputTokens)" radius={[4, 4, 0, 0]} />
              <Bar yAxisId="tokens" dataKey="outputTokens" stackId="tokens" fill="var(--color-outputTokens)" radius={[4, 4, 0, 0]} />
              <Line yAxisId="cost" dataKey="costUsd" type="monotone" stroke="var(--color-costUsd)" strokeWidth={2} dot={false} />
            </ComposedChart>
          </ChartContainer>
        </ChartCard>

        <ChartCard title={t("charts.failures.title")}>
          <ChartContainer config={errorChartConfig} className="h-[240px] w-full md:h-[260px]">
            <BarChart data={errorChartData} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
              <CartesianGrid vertical={false} />
              <XAxis dataKey="errorCode" tickLine={false} axisLine={false} minTickGap={24} />
              <YAxis tickLine={false} axisLine={false} width={44} tickFormatter={(value: number) => formatCompactNumber(value, { locale, compactThreshold: 1_000 }).value} />
              <ChartTooltip content={<ChartTooltipContent />} />
              <Bar dataKey="count" fill="var(--color-count)" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ChartContainer>
        </ChartCard>

        <ChartCard title={t("charts.savings.title")}>
          <ChartContainer config={savingsChartConfig} className="h-[240px] w-full md:h-[260px]">
            <BarChart data={savingsChartData} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
              <CartesianGrid vertical={false} />
              <XAxis dataKey="label" tickLine={false} axisLine={false} minTickGap={24} />
              <YAxis tickLine={false} axisLine={false} width={52} tickFormatter={(value: number) => formatCompactCurrency(value, { locale, currency: "USD", compactThreshold: 1_000, maximumFractionDigits: 6 }).value} />
              <ChartTooltip content={<ChartTooltipContent labelFormatter={formatTooltipLabel} />} />
              <Bar dataKey="actualCost" stackId="cost" fill="var(--color-actualCost)" radius={[4, 4, 0, 0]} />
              <Bar dataKey="simpleSavings" stackId="cost" fill="var(--color-simpleSavings)" />
              <Bar dataKey="smartSavings" stackId="cost" fill="var(--color-smartSavings)" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ChartContainer>
        </ChartCard>
      </section>

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
              {cacheBreakdownItems.map((item) => {
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
    </div>
  );
}

function ChartCard({
  title,
  action,
  children,
}: {
  title: string;
  action?: React.ReactNode;
  children: React.ReactNode;
}) {
  return (
    <AppCard
      title={<span className="truncate text-sm">{title}</span>}
      actions={action}
      contentClassName="pt-0"
      className="shadow-none"
    >
      {children}
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

async function postMetricsQuery(body: Record<string, unknown>, fallbackError: string): Promise<{ data?: Row[] }> {
  const res = await fetch("/api/metrics/query", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
    cache: "no-store",
  });
  const payload = await res.json();
  if (!res.ok) throw new Error(payload?.error?.message ?? `${fallbackError} (${res.status})`);
  return payload as { data?: Row[] };
}

async function requestJSON<T>(path: string, fallbackError: string): Promise<T> {
  const res = await fetch(path, { cache: "no-store" });
  const text = await res.text();
  const payload = text.trim() ? JSON.parse(text) : {};
  if (!res.ok) throw new Error(extractMessageFromPayload(payload) ?? `${fallbackError} (${res.status})`);
  return payload as T;
}

function extractMessageFromPayload(payload: unknown): string | null {
  if (typeof payload === "string") return payload;
  if (!payload || typeof payload !== "object") return null;
  const candidate = payload as { error?: { message?: string } | string; message?: string };
  if (typeof candidate.error === "string") return candidate.error;
  return candidate.error?.message ?? candidate.message ?? null;
}

function formatTooltipLabel(value: unknown, payload: readonly unknown[]) {
  const first = payload[0] as { payload?: { tooltipLabel?: string } } | undefined;
  return first?.payload?.tooltipLabel ?? String(value ?? "");
}

function normalizeInitialIso(value: string | undefined) {
  const trimmed = value?.trim();
  if (trimmed && !Number.isNaN(Date.parse(trimmed))) return new Date(trimmed).toISOString();
  return "";
}

function resolveSelectedRange(preset: RangePreset, from: string, to: string) {
  if (preset === "custom") return { from, to };
  return rangeForPreset(preset);
}

function rangeForPreset(preset: Exclude<RangePreset, "custom">) {
  const now = new Date();
  if (preset === "yesterday") {
    const start = new Date(now);
    start.setHours(0, 0, 0, 0);
    start.setDate(start.getDate() - 1);
    const end = new Date(start);
    end.setDate(end.getDate() + 1);
    return { from: start.toISOString(), to: end.toISOString() };
  }

  const durationMs = {
    last15m: 15 * 60 * 1000,
    last1h: 60 * 60 * 1000,
    last24h: DEFAULT_RANGE_MS,
    last3d: 3 * DEFAULT_RANGE_MS,
    last7d: 7 * DEFAULT_RANGE_MS,
    last14d: 14 * DEFAULT_RANGE_MS,
    last30d: 30 * DEFAULT_RANGE_MS,
  }[preset];

  return {
    from: new Date(now.getTime() - durationMs).toISOString(),
    to: now.toISOString(),
  };
}

function formatPercent(value: number, locale: string) {
  return new Intl.NumberFormat(locale, { style: "percent", maximumFractionDigits: 1 }).format(value);
}
