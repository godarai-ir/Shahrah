"use client";

import { useEffect, useMemo, useRef, useState } from "react";
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
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  type ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

const VIEW_BY = ["modelName", "virtualModelName", "userEmail", "virtualAccount", "team"] as const;
const PCTS = ["p50", "p75", "p90", "p99"] as const;
const INTERVAL_SECONDS = 300;
const MAX_SERIES = 5;
const DEFAULT_RANGE_MS = 24 * 60 * 60 * 1000;

type ViewBy = (typeof VIEW_BY)[number];
type Pct = (typeof PCTS)[number];
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
  const didRunInitialQuery = useRef(false);
  const [pct, setPct] = useState<Pct>("p99");
  const [modelFilter, setModelFilter] = useState(props.initialModel ?? "");
  const [providerFilter, setProviderFilter] = useState(props.initialProvider ?? "");
  const [requestIDFilter, setRequestIDFilter] = useState(props.initialRequestID ?? "");
  const [loading, setLoading] = useState(false);
  const [hasLoaded, setHasLoaded] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [from, setFrom] = useState(() => normalizeInitialIso(props.initialFrom));
  const [to, setTo] = useState(() => normalizeInitialIso(props.initialTo));

  const [distribution, setDistribution] = useState<Row[]>([]);
  const [timeseries, setTimeseries] = useState<Row[]>([]);
  const [errorBreakdown, setErrorBreakdown] = useState<Row[]>([]);

  const summary = useMemo<Summary>(() => {
    const totalRequests = distribution.reduce((acc, row) => acc + metricNumber(row, "countModelName", "countAll"), 0);
    const inputTokens = distribution.reduce((acc, row) => acc + metricNumber(row, "sumInputTokens"), 0);
    const outputTokens = distribution.reduce((acc, row) => acc + metricNumber(row, "sumOutputTokens"), 0);
    const cost = distribution.reduce((acc, row) => acc + metricNumber(row, "sumCostInUSD", "sumCostUsd"), 0);
    return { totalRequests, inputTokens, outputTokens, cost };
  }, [distribution]);

  const chartData = useMemo(
    () => buildChartData(timeseries, viewBy, pct, t("values.unknown"), locale, from, to),
    [timeseries, viewBy, pct, t, locale, from, to],
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

  async function runQuery(range?: { from: string; to: string }) {
    const queryFrom = range?.from ?? from;
    const queryTo = range?.to ?? to;
    if (!queryFrom || !queryTo) return;

    setLoading(true);
    setError(null);
    try {
      const filters: Array<Record<string, unknown>> = [];
      if (modelFilter.trim()) {
        filters.push({ fieldName: "modelName", operator: "STRING_CONTAINS", value: modelFilter.trim() });
      }
      if (providerFilter.trim()) {
        filters.push({ fieldName: "providerName", operator: "EQUAL", value: providerFilter.trim() });
      }
      if (requestIDFilter.trim()) {
        filters.push({ fieldName: "requestID", operator: "EQUAL", value: requestIDFilter.trim() });
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
          groupBy: [viewBy],
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
          groupBy: [viewBy],
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

  useEffect(() => {
    const id = window.setTimeout(() => {
      if (didRunInitialQuery.current) return;
      didRunInitialQuery.current = true;

      let queryTo = to;
      let queryFrom = from;
      if (!queryTo || !queryFrom) {
        const now = Date.now();
        queryTo = new Date(now).toISOString();
        queryFrom = new Date(now - DEFAULT_RANGE_MS).toISOString();
        setTo(queryTo);
        setFrom(queryFrom);
      }
      void runQuery({ from: queryFrom, to: queryTo });
    }, 0);
    return () => window.clearTimeout(id);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <div className="flex flex-col gap-6">
      <header className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <div className="mb-2 flex flex-wrap items-center gap-2">
            <Badge variant="secondary">{t("badges.datasource")}</Badge>
            <Badge variant="outline">{t(`filters.viewByOptions.${viewBy}`)}</Badge>
          </div>
          <h1 className="text-2xl font-semibold tracking-normal">{t("title")}</h1>
          <p className="mt-1 max-w-3xl text-sm text-muted-foreground">{t("description")}</p>
        </div>
        <div className="flex flex-wrap items-center gap-2 sm:justify-end">
          <Button type="button" variant="outline" onClick={() => void runQuery()} disabled={loading}>
            {loading ? common("loading") : common("refresh")}
          </Button>
        </div>
      </header>

      <Card className="rounded-lg">
        <CardHeader>
          <CardTitle className="text-base">{t("filters.title")}</CardTitle>
          <CardDescription>{t("filters.description")}</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <DateTimePicker id="metrics-from" label={t("filters.from")} value={from} onChange={setFrom} />
          <DateTimePicker id="metrics-to" label={t("filters.to")} value={to} onChange={setTo} />
          <Field id="metrics-model" label={t("filters.model")}>
            <Input
              id="metrics-model"
              value={modelFilter}
              onChange={(event) => setModelFilter(event.target.value)}
              placeholder={t("filters.placeholders.model")}
              dir="ltr"
              className="font-mono"
            />
          </Field>
          <Field id="metrics-provider" label={t("filters.provider")}>
            <Input
              id="metrics-provider"
              value={providerFilter}
              onChange={(event) => setProviderFilter(event.target.value)}
              placeholder={t("filters.placeholders.provider")}
              dir="ltr"
              className="font-mono"
            />
          </Field>
          <Field id="metrics-request-id" label={t("filters.requestId")}>
            <Input
              id="metrics-request-id"
              value={requestIDFilter}
              onChange={(event) => setRequestIDFilter(event.target.value)}
              placeholder={t("filters.placeholders.requestId")}
              dir="ltr"
              className="font-mono"
            />
          </Field>
          <Field id="metrics-view-by" label={t("filters.viewBy")}>
            <select
              id="metrics-view-by"
              className="h-10 rounded-md border border-input bg-background px-3 text-sm text-foreground"
              value={viewBy}
              onChange={(event) => setViewBy(event.target.value as ViewBy)}
            >
              {VIEW_BY.map((value) => (
                <option key={value} value={value}>
                  {t(`filters.viewByOptions.${value}`)}
                </option>
              ))}
            </select>
          </Field>
          <div className="flex items-end">
            <Button type="button" onClick={() => void runQuery()} disabled={loading}>
              {loading ? common("loading") : t("actions.applyFilters")}
            </Button>
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

      <section aria-label={t("sections.summary")} className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <MetricCard title={t("cards.totalRequests")} value={formatNumber(summary.totalRequests, locale)} description={t("cards.totalRequestsHint")} />
        <MetricCard title={t("cards.inputTokens")} value={formatNumber(summary.inputTokens, locale)} description={t("cards.inputTokensHint")} />
        <MetricCard title={t("cards.outputTokens")} value={formatNumber(summary.outputTokens, locale)} description={t("cards.outputTokensHint")} />
        <MetricCard title={t("cards.totalCost")} value={formatCurrency(summary.cost, locale)} description={t("cards.totalCostHint")} />
      </section>

      <section aria-label={t("sections.charts")} className="grid gap-4 xl:grid-cols-2">
        <ChartCard title={t("charts.requests.title")} description={t("charts.requests.description")}>
          <ChartContainer config={seriesConfig} className="h-[320px] w-full">
            <AreaChart data={chartData.points} margin={{ top: 12, right: 12, left: 0, bottom: 0 }}>
              <CartesianGrid vertical={false} />
              <XAxis dataKey="label" tickLine={false} axisLine={false} minTickGap={24} />
              <YAxis tickLine={false} axisLine={false} width={42} allowDecimals={false} />
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
          description={t("charts.latency.description")}
          action={
            <Field id="metrics-percentile" label={t("filters.percentile")}>
              <select
                id="metrics-percentile"
                className="h-10 min-w-28 rounded-md border border-input bg-background px-3 text-sm text-foreground"
                value={pct}
                onChange={(event) => setPct(event.target.value as Pct)}
              >
                {PCTS.map((value) => (
                  <option key={value} value={value}>
                    {t(`filters.percentiles.${value}`)}
                  </option>
                ))}
              </select>
            </Field>
          }
        >
          <ChartContainer config={seriesConfig} className="h-[320px] w-full">
            <AreaChart data={chartData.latencyPoints} margin={{ top: 12, right: 12, left: 0, bottom: 0 }}>
              <CartesianGrid vertical={false} />
              <XAxis dataKey="label" tickLine={false} axisLine={false} minTickGap={24} />
              <YAxis tickLine={false} axisLine={false} width={42} />
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

        <ChartCard title={t("charts.tokensCost.title")} description={t("charts.tokensCost.description")}>
          <ChartContainer config={tokenChartConfig} className="h-[320px] w-full">
            <ComposedChart data={chartData.tokenCostPoints} margin={{ top: 12, right: 12, left: 0, bottom: 0 }}>
              <CartesianGrid vertical={false} />
              <XAxis dataKey="label" tickLine={false} axisLine={false} minTickGap={24} />
              <YAxis yAxisId="tokens" tickLine={false} axisLine={false} width={42} />
              <YAxis yAxisId="cost" orientation="right" tickLine={false} axisLine={false} width={42} />
              <ChartTooltip content={<ChartTooltipContent labelFormatter={formatTooltipLabel} />} />
              <Bar yAxisId="tokens" dataKey="inputTokens" stackId="tokens" fill="var(--color-inputTokens)" radius={[4, 4, 0, 0]} />
              <Bar yAxisId="tokens" dataKey="outputTokens" stackId="tokens" fill="var(--color-outputTokens)" radius={[4, 4, 0, 0]} />
              <Line yAxisId="cost" dataKey="costUsd" type="monotone" stroke="var(--color-costUsd)" strokeWidth={2} dot={false} />
            </ComposedChart>
          </ChartContainer>
        </ChartCard>

        <ChartCard title={t("charts.failures.title")} description={t("charts.failures.description")}>
          <ChartContainer config={errorChartConfig} className="h-[320px] w-full">
            <BarChart data={errorChartData} margin={{ top: 12, right: 12, left: 0, bottom: 0 }}>
              <CartesianGrid vertical={false} />
              <XAxis dataKey="errorCode" tickLine={false} axisLine={false} minTickGap={24} />
              <YAxis tickLine={false} axisLine={false} width={42} />
              <ChartTooltip content={<ChartTooltipContent />} />
              <Bar dataKey="count" fill="var(--color-count)" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ChartContainer>
        </ChartCard>
      </section>
    </div>
  );
}

function Field({ id, label, children }: { id: string; label: string; children: React.ReactNode }) {
  return (
    <div className="grid gap-2">
      <Label htmlFor={id}>{label}</Label>
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
      <Input
        id={id}
        type="datetime-local"
        value={toDateTimeLocalValue(value)}
        onChange={(event) => {
          const next = fromDateTimeLocalValue(event.target.value);
          if (next) onChange(next);
        }}
      />
    </Field>
  );
}

function MetricCard({ title, value, description }: { title: string; value: string; description: string }) {
  return (
    <Card className="rounded-lg">
      <CardHeader className="gap-2 p-4">
        <CardDescription>{title}</CardDescription>
        <CardTitle className="text-2xl">{value}</CardTitle>
        <CardDescription>{description}</CardDescription>
      </CardHeader>
    </Card>
  );
}

function ChartCard({
  title,
  description,
  action,
  children,
}: {
  title: string;
  description: string;
  action?: React.ReactNode;
  children: React.ReactNode;
}) {
  return (
    <Card className="rounded-lg">
      <CardHeader className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <CardTitle className="text-base">{title}</CardTitle>
          <CardDescription>{description}</CardDescription>
        </div>
        {action ? <div className="shrink-0">{action}</div> : null}
      </CardHeader>
      <CardContent>{children}</CardContent>
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
