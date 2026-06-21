"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { Calendar, RotateCcw } from "lucide-react";
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

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  type ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";
import {
  Combobox,
  ComboboxContent,
  ComboboxEmpty,
  ComboboxInput,
  ComboboxItem,
  ComboboxList,
} from "@/components/ui/combobox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";

const VIEW_BY = ["modelName", "virtualModelName", "userEmail", "virtualAccount", "team"] as const;
const PCTS = ["p50", "p75", "p90", "p99"] as const;
const RANGE_PRESETS = ["last15m", "last1h", "last24h", "yesterday", "last3d", "last7d", "last14d", "last30d", "custom"] as const;
const INTERVAL_SECONDS = 300;
const MAX_SERIES = 5;
const DEFAULT_RANGE_MS = 24 * 60 * 60 * 1000;
const ALL_MODELS_VALUE = "__all_models__";
const ALL_PROVIDERS_VALUE = "__all_providers__";

type ViewBy = (typeof VIEW_BY)[number];
type Pct = (typeof PCTS)[number];
type RangePreset = (typeof RANGE_PRESETS)[number];
type Row = Record<string, unknown>;

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
};

type MetricComboboxOption = {
  value: string;
  label: string;
};

type ModelMetricsFacetsResponse = {
  models?: unknown;
  providers?: unknown;
};

type SeriesPoint = {
  timestamp: string;
  label: string;
  tooltipLabel: string;
} & Record<string, number | string>;

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
    return { totalRequests, inputTokens, outputTokens, cost };
  }, [distribution]);

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

      const [dist, ts, errb] = await Promise.all([
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
      ]);

      setDisplayViewBy(queryViewBy);
      setDistribution(dist.data ?? []);
      setTimeseries(ts.data ?? []);
      setErrorBreakdown(errb.data ?? []);
    } catch (e) {
      setError(e instanceof Error ? e.message : common("requestError"));
      setDistribution([]);
      setTimeseries([]);
      setErrorBreakdown([]);
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
    <div className="flex flex-col gap-3">
      <header className="flex items-center justify-between gap-3">
        <div className="min-w-0">
          <h1 className="text-xl font-semibold tracking-normal">{t("title")}</h1>
        </div>
      </header>

      <Card className="rounded-lg">
        <CardHeader className="sr-only">
          <CardTitle>{t("filters.title")}</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-2 p-3 pt-3 sm:grid-cols-2 lg:grid-cols-4 2xl:grid-cols-6 md:p-3 md:pt-3">
          <Field id="metrics-range-preset" label={t("filters.timeRange")}>
            <MetricCombobox
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
            <>
              <DateTimePicker id="metrics-from" label={t("filters.from")} value={from} onChange={setFrom} />
              <DateTimePicker id="metrics-to" label={t("filters.to")} value={to} onChange={setTo} />
            </>
          ) : null}
          <Field id="metrics-model" label={t("filters.model")}>
            <MetricCombobox
              id="metrics-model"
              options={modelComboboxOptions}
              value={modelFilter || ALL_MODELS_VALUE}
              onValueChange={(value) => setModelFilter(value === ALL_MODELS_VALUE ? "" : value)}
              placeholder={loadingOptions ? common("loading") : t("filters.search")}
              emptyText={t("filters.emptyOptions")}
              mono
            />
          </Field>
          <Field id="metrics-provider" label={t("filters.provider")}>
            <MetricCombobox
              id="metrics-provider"
              options={providerComboboxOptions}
              value={providerFilter || ALL_PROVIDERS_VALUE}
              onValueChange={(value) => setProviderFilter(value === ALL_PROVIDERS_VALUE ? "" : value)}
              placeholder={loadingOptions ? common("loading") : t("filters.search")}
              emptyText={t("filters.emptyOptions")}
              mono
            />
          </Field>
          <Field id="metrics-request-id" label={t("filters.requestId")}>
            <Input
              id="metrics-request-id"
              value={requestIDFilter}
              onChange={(event) => setRequestIDFilter(event.target.value)}
              placeholder={t("filters.placeholders.requestId")}
              dir="ltr"
              className="h-8 px-2 text-xs font-mono"
            />
          </Field>
          <Field id="metrics-view-by" label={t("filters.viewBy")}>
            <MetricCombobox
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
        </CardContent>
      </Card>

      {error ? (
        <Alert variant="destructive">
          <AlertTitle>{t("alerts.errorTitle")}</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}

      {hasLoaded && !loading && !error && distribution.length === 0 ? (
        <Alert>
          <AlertDescription>{t("table.empty")}</AlertDescription>
        </Alert>
      ) : null}

      <section aria-label={t("sections.summary")} className="grid gap-2 sm:grid-cols-2 xl:grid-cols-4">
        <MetricCard title={t("cards.totalRequests")} value={formatNumber(summary.totalRequests, locale)} />
        <MetricCard title={t("cards.inputTokens")} value={formatNumber(summary.inputTokens, locale)} />
        <MetricCard title={t("cards.outputTokens")} value={formatNumber(summary.outputTokens, locale)} />
        <MetricCard title={t("cards.totalCost")} value={formatCurrency(summary.cost, locale)} />
      </section>

      <section aria-label={t("sections.charts")} className="grid gap-3 xl:grid-cols-2">
        <ChartCard title={requestChartTitle}>
          <ChartContainer config={seriesConfig} className="h-[240px] w-full md:h-[260px]">
            <AreaChart data={chartData.points} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
              <CartesianGrid vertical={false} />
              <XAxis dataKey="label" tickLine={false} axisLine={false} minTickGap={24} />
              <YAxis tickLine={false} axisLine={false} width={36} allowDecimals={false} />
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
            <Field id="metrics-percentile" label={t("filters.percentile")} className="w-32">
              <MetricCombobox
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
              <YAxis tickLine={false} axisLine={false} width={36} />
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
              <YAxis yAxisId="tokens" tickLine={false} axisLine={false} width={36} />
              <YAxis yAxisId="cost" orientation="right" tickLine={false} axisLine={false} width={36} />
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
              <YAxis tickLine={false} axisLine={false} width={36} />
              <ChartTooltip content={<ChartTooltipContent />} />
              <Bar dataKey="count" fill="var(--color-count)" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ChartContainer>
        </ChartCard>
      </section>
    </div>
  );
}

function Field({
  id,
  label,
  children,
  className,
}: {
  id: string;
  label: string;
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <div className={cn("grid min-w-0 gap-1", className)}>
      <Label htmlFor={id} className="text-xs text-muted-foreground">
        {label}
      </Label>
      {children}
    </div>
  );
}

function DateTimePicker({
  id,
  label,
  value,
  onChange,
}: {
  id: string;
  label: string;
  value: string;
  onChange: (value: string) => void;
}) {
  return (
    <Field id={id} label={label}>
      <div className="relative">
        <Calendar className="pointer-events-none absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" aria-hidden="true" />
        <Input
          id={id}
          type="datetime-local"
          value={toDateTimeLocalValue(value)}
          onChange={(event) => {
            const next = fromDateTimeLocalValue(event.target.value);
            if (next) onChange(next);
          }}
          className="h-8 px-2 pl-8 text-xs"
        />
      </div>
    </Field>
  );
}

function MetricCombobox({
  id,
  options,
  value,
  onValueChange,
  placeholder,
  emptyText,
  searchable = true,
  className,
  mono = false,
}: {
  id: string;
  options: MetricComboboxOption[];
  value: string;
  onValueChange: (value: string) => void;
  placeholder: string;
  emptyText: string;
  searchable?: boolean;
  className?: string;
  mono?: boolean;
}) {
  const selectedOption = options.find((option) => option.value === value) ?? null;

  return (
    <Combobox
      items={options}
      value={selectedOption}
      onValueChange={(option) => {
        if (option) onValueChange(option.value);
      }}
      itemToStringValue={(option) => option.label}
    >
      <ComboboxInput
        id={id}
        dir="rtl"
        placeholder={placeholder}
        readOnly={!searchable}
        className={cn(
          "h-8 w-full bg-background [&_[data-slot=input-group-control]]:min-w-0 [&_[data-slot=input-group-control]]:text-right [&_[data-slot=input-group-control]]:text-xs",
          mono && "[&_[data-slot=input-group-control]]:font-mono",
          className,
        )}
      />
      <ComboboxContent align="start" className="text-right" dir="rtl">
        <ComboboxEmpty>{emptyText}</ComboboxEmpty>
        <ComboboxList>
          {(option) => (
            <ComboboxItem key={option.value} value={option} className={cn("text-right", mono && "font-mono")} dir="rtl">
              <span className="min-w-0 truncate" dir="auto">
                {option.label}
              </span>
            </ComboboxItem>
          )}
        </ComboboxList>
      </ComboboxContent>
    </Combobox>
  );
}

function MetricCard({ title, value }: { title: string; value: string }) {
  return (
    <Card className="rounded-lg">
      <CardHeader className="gap-1 p-3 md:p-3">
        <CardDescription className="text-xs leading-none">{title}</CardDescription>
        <CardTitle className="truncate text-xl leading-tight tabular-nums" dir="auto">
          {value}
        </CardTitle>
      </CardHeader>
    </Card>
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
    <Card className="rounded-lg">
      <CardHeader className="flex flex-col gap-2 p-3 sm:flex-row sm:items-center sm:justify-between md:p-3">
        <div className="min-w-0">
          <CardTitle className="truncate text-sm">{title}</CardTitle>
        </div>
        {action ? <div className="shrink-0">{action}</div> : null}
      </CardHeader>
      <CardContent className="px-3 pb-3 md:px-3 md:pb-3">{children}</CardContent>
    </Card>
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

export function buildChartData(
  rows: Row[],
  viewBy: ViewBy,
  pct: Pct,
  unknownLabel: string,
  locale?: string,
  rangeFrom?: string,
  rangeTo?: string,
) {
  const timeSet = new Set<string>();
  const grouped = new Map<string, Row[]>();

  for (const row of rows) {
    const key = metricString(row, viewBy) || unknownLabel;
    const ts = metricString(row, "timestamp");
    if (!grouped.has(key)) grouped.set(key, []);
    grouped.get(key)!.push(row);
    if (ts) timeSet.add(ts);
  }

  const x = Array.from(timeSet).sort((a, b) => new Date(a).getTime() - new Date(b).getTime());
  const labelMode = getAxisLabelMode(x, rangeFrom, rangeTo);
  const series = Array.from(grouped.entries())
    .slice(0, MAX_SERIES)
    .map(([label], index) => ({ label, key: `series${index}` }));

  const points: SeriesPoint[] = x.map((timestamp) => {
    const point: SeriesPoint = {
      timestamp,
      label: formatAxisLabel(timestamp, labelMode, locale),
      tooltipLabel: formatLocalDateTime(timestamp, locale),
    };

    series.forEach((seriesItem) => {
      const item = grouped.get(seriesItem.label)?.find((row) => metricString(row, "timestamp") === timestamp);
      point[seriesItem.key] = metricNumber(item, "countModelName", "countAll");
    });

    return point;
  });

  const latencyPoints: SeriesPoint[] = x.map((timestamp) => {
    const point: SeriesPoint = {
      timestamp,
      label: formatAxisLabel(timestamp, labelMode, locale),
      tooltipLabel: formatLocalDateTime(timestamp, locale),
    };
    const latencyKey = `${pct}LatencyMs`;
    const latencyKeyAlt = `${pct}LatencyMS`;

    series.forEach((seriesItem) => {
      const item = grouped.get(seriesItem.label)?.find((row) => metricString(row, "timestamp") === timestamp);
      point[seriesItem.key] = metricNumber(item, latencyKey, latencyKeyAlt);
    });

    return point;
  });

  const tokenCostPoints = x.map((timestamp) => {
    let inputTokens = 0;
    let outputTokens = 0;
    let costUsd = 0;

    for (const [, items] of grouped) {
      const item = items.find((row) => metricString(row, "timestamp") === timestamp);
      inputTokens += metricNumber(item, "sumInputTokens");
      outputTokens += metricNumber(item, "sumOutputTokens");
      costUsd += metricNumber(item, "sumCostInUSD", "sumCostUsd");
    }

    return {
      timestamp,
      label: formatAxisLabel(timestamp, labelMode, locale),
      tooltipLabel: formatLocalDateTime(timestamp, locale),
      inputTokens,
      outputTokens,
      costUsd: Number(costUsd.toFixed(6)),
    };
  });

  return { points, latencyPoints, tokenCostPoints, series };
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

export function metricNumber(row: Row | undefined, ...keys: string[]) {
  const value = metricValue(row, keys);
  if (typeof value === "number") return Number.isFinite(value) ? value : 0;
  if (typeof value === "bigint") return Number(value);
  if (typeof value === "string") {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : 0;
  }
  return 0;
}

function metricString(row: Row | undefined, ...keys: string[]) {
  const value = metricValue(row, keys);
  if (value == null) return "";
  if (value instanceof Date) return value.toISOString();
  return String(value);
}

function metricValue(row: Row | undefined, keys: string[]) {
  if (!row) return undefined;
  const candidates = new Set<string>();
  for (const key of keys) {
    candidates.add(key);
    candidates.add(key.toLowerCase());
    candidates.add(toSnakeCase(key));
    for (const alias of metricAliases[key] ?? []) {
      candidates.add(alias);
      candidates.add(alias.toLowerCase());
    }
  }
  for (const key of candidates) {
    if (Object.prototype.hasOwnProperty.call(row, key)) {
      return row[key];
    }
  }
  return undefined;
}

function toSnakeCase(value: string) {
  return value.replace(/([a-z0-9])([A-Z])/g, "$1_$2").toLowerCase();
}

const metricAliases: Record<string, string[]> = {
  modelName: ["model_name", "view_model_name"],
  virtualModelName: ["virtual_model_name", "view_virtual_model_name"],
  userEmail: ["user_email", "view_user_email"],
  virtualAccount: ["virtual_account", "virtual_account_name", "view_virtual_account_name"],
  team: ["team_name", "view_team_name"],
  errorCode: ["error_code", "status_code"],
  timestamp: ["bucket_ts"],
  countModelName: ["count_model_name", "countmodelname", "request_count"],
  countAll: ["count_all", "countall"],
  sumInputTokens: ["sum_input_tokens", "suminputtokens", "input_tokens_sum"],
  sumOutputTokens: ["sum_output_tokens", "sumoutputtokens", "output_tokens_sum"],
  sumCostInUSD: ["sum_cost_in_usd", "sumcostinusd", "sum_cost_usd", "cost_usd_sum"],
  sumCostUsd: ["sum_cost_usd", "sumcostusd", "cost_usd_sum"],
  p50LatencyMs: ["p50_latency_ms", "p50latencyms", "latency_p50"],
  p75LatencyMs: ["p75_latency_ms", "p75latencyms", "latency_p75"],
  p90LatencyMs: ["p90_latency_ms", "p90latencyms", "latency_p90"],
  p99LatencyMs: ["p99_latency_ms", "p99latencyms", "latency_p99"],
};

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

export function normalizeStringList(value: unknown) {
  if (!Array.isArray(value)) return [];
  return Array.from(
    new Set(
      value
        .map((item) => (typeof item === "string" ? item.trim() : ""))
        .filter(Boolean),
    ),
  ).sort((a, b) => a.localeCompare(b));
}

function toDateTimeLocalValue(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  const pad = (part: number) => String(part).padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
}

function fromDateTimeLocalValue(value: string) {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return date.toISOString();
}

type AxisLabelMode = "time" | "dateTime";

function getAxisLabelMode(timestamps: string[], rangeFrom?: string, rangeTo?: string): AxisLabelMode {
  const rangeStart = parseTime(rangeFrom);
  const rangeEnd = parseTime(rangeTo);
  if (rangeStart != null && rangeEnd != null && rangeEnd - rangeStart >= 24 * 60 * 60 * 1000) {
    return "dateTime";
  }

  const first = parseTime(timestamps[0]);
  const last = parseTime(timestamps[timestamps.length - 1]);
  if (first != null && last != null && last - first >= 24 * 60 * 60 * 1000) {
    return "dateTime";
  }

  return "time";
}

function parseTime(value: string | undefined) {
  if (!value) return undefined;
  const time = new Date(value).getTime();
  return Number.isNaN(time) ? undefined : time;
}

function formatAxisLabel(value: string, mode: AxisLabelMode, locale?: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  if (mode === "dateTime") {
    return new Intl.DateTimeFormat(locale, {
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    }).format(date);
  }
  return new Intl.DateTimeFormat(locale, {
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

function formatLocalDateTime(value: string, locale?: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return new Intl.DateTimeFormat(locale, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

function formatNumber(value: number, locale: string) {
  return new Intl.NumberFormat(locale).format(value);
}

function formatCurrency(value: number, locale: string) {
  return new Intl.NumberFormat(locale, {
    style: "currency",
    currency: "USD",
    maximumFractionDigits: 6,
  }).format(value);
}
