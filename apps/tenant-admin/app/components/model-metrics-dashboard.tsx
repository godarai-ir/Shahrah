"use client";

import { useEffect, useMemo, useState } from "react";
import { useTranslations } from "next-intl";
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
  const t = useTranslations("metrics");
  const common = useTranslations("common");

  const [viewBy, setViewBy] = useState<ViewBy>("modelName");
  const [pct, setPct] = useState<Pct>("p99");
  const [modelFilter, setModelFilter] = useState(props.initialModel ?? "");
  const [providerFilter, setProviderFilter] = useState(props.initialProvider ?? "");
  const [requestIDFilter, setRequestIDFilter] = useState(props.initialRequestID ?? "");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [from, setFrom] = useState(() => normalizeInitialIso(props.initialFrom, Date.now() - 3 * 60 * 60 * 1000));
  const [to, setTo] = useState(() => normalizeInitialIso(props.initialTo, Date.now()));

  const [distribution, setDistribution] = useState<Row[]>([]);
  const [timeseries, setTimeseries] = useState<Row[]>([]);
  const [errorBreakdown, setErrorBreakdown] = useState<Row[]>([]);

  const summary = useMemo<Summary>(() => {
    const totalRequests = distribution.reduce((acc, row) => acc + Number(row.countModelName ?? row.countAll ?? 0), 0);
    const inputTokens = distribution.reduce((acc, row) => acc + Number(row.sumInputTokens ?? 0), 0);
    const outputTokens = distribution.reduce((acc, row) => acc + Number(row.sumOutputTokens ?? 0), 0);
    const cost = distribution.reduce((acc, row) => acc + Number(row.sumCostInUSD ?? 0), 0);
    return { totalRequests, inputTokens, outputTokens, cost };
  }, [distribution]);

  const chartData = useMemo(
    () => buildChartData(timeseries, viewBy, pct, t("values.unknown")),
    [timeseries, viewBy, pct, t],
  );

  const errorChartData = useMemo(
    () =>
      errorBreakdown.map((row) => ({
        errorCode: String(row.errorCode ?? t("values.unknown")),
        count: Number(row.countModelName ?? row.countAll ?? 0),
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

  async function runQuery() {
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
          startTs: from,
          endTs: to,
          datasource: "modelMetrics",
          type: "distribution",
          aggregations: [
            { type: "count", column: "modelName" },
            { type: "sum", column: "inputTokens" },
            { type: "sum", column: "outputTokens" },
            { type: "sum", column: "costInUSD" },
            { type: pct, column: "latencyMs" },
          ],
          groupBy: [viewBy],
          filters,
        }, common("requestError")),
        postMetricsQuery({
          startTs: from,
          endTs: to,
          datasource: "modelMetrics",
          type: "timeseries",
          aggregations: [
            { type: "count", column: "modelName" },
            { type: pct, column: "latencyMs" },
            { type: "sum", column: "inputTokens" },
            { type: "sum", column: "outputTokens" },
            { type: "sum", column: "costInUSD" },
          ],
          groupBy: [viewBy],
          filters,
          intervalInSeconds: INTERVAL_SECONDS,
        }, common("requestError")),
        postMetricsQuery({
          startTs: from,
          endTs: to,
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
      setLoading(false);
    }
  }

  useEffect(() => {
    const id = window.setTimeout(() => {
      void runQuery();
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
          <Field id="metrics-percentile" label={t("filters.percentile")}>
            <select
              id="metrics-percentile"
              className="h-10 rounded-md border border-input bg-background px-3 text-sm text-foreground"
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

      <section aria-label={t("sections.summary")} className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <MetricCard title={t("cards.totalRequests")} value={formatNumber(summary.totalRequests)} description={t("cards.totalRequestsHint")} />
        <MetricCard title={t("cards.inputTokens")} value={formatNumber(summary.inputTokens)} description={t("cards.inputTokensHint")} />
        <MetricCard title={t("cards.outputTokens")} value={formatNumber(summary.outputTokens)} description={t("cards.outputTokensHint")} />
        <MetricCard title={t("cards.totalCost")} value={formatCurrency(summary.cost)} description={t("cards.totalCostHint")} />
      </section>

      <section aria-label={t("sections.charts")} className="grid gap-4 xl:grid-cols-2">
        <ChartCard title={t("charts.requests.title")} description={t("charts.requests.description")}>
          <ChartContainer config={seriesConfig} className="h-[320px] w-full">
            <AreaChart data={chartData.points} margin={{ top: 12, right: 12, left: 0, bottom: 0 }}>
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
                  fillOpacity={0.16}
                  stroke={`var(--color-${series.key})`}
                  strokeWidth={2}
                />
              ))}
            </AreaChart>
          </ChartContainer>
        </ChartCard>

        <ChartCard title={t("charts.latency.title", { percentile: pct.toUpperCase() })} description={t("charts.latency.description")}>
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

function ChartCard({ title, description, children }: { title: string; description: string; children: React.ReactNode }) {
  return (
    <Card className="rounded-lg">
      <CardHeader>
        <CardTitle className="text-base">{title}</CardTitle>
        <CardDescription>{description}</CardDescription>
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

function buildChartData(rows: Row[], viewBy: ViewBy, pct: Pct, unknownLabel: string) {
  const timeSet = new Set<string>();
  const grouped = new Map<string, Row[]>();

  for (const row of rows) {
    const key = String(row[viewBy] ?? unknownLabel);
    const ts = String(row.timestamp ?? "");
    if (!grouped.has(key)) grouped.set(key, []);
    grouped.get(key)!.push(row);
    if (ts) timeSet.add(ts);
  }

  const x = Array.from(timeSet).sort((a, b) => new Date(a).getTime() - new Date(b).getTime());
  const series = Array.from(grouped.entries())
    .slice(0, MAX_SERIES)
    .map(([label], index) => ({ label, key: `series${index}` }));

  const points: SeriesPoint[] = x.map((timestamp) => {
    const point: SeriesPoint = {
      timestamp,
      label: formatLocalTime(timestamp),
      tooltipLabel: formatLocalDateTime(timestamp),
    };

    series.forEach((seriesItem) => {
      const item = grouped.get(seriesItem.label)?.find((row) => String(row.timestamp ?? "") === timestamp);
      point[seriesItem.key] = Number(item?.countModelName ?? item?.countAll ?? 0) / INTERVAL_SECONDS;
    });

    return point;
  });

  const latencyPoints: SeriesPoint[] = x.map((timestamp) => {
    const point: SeriesPoint = {
      timestamp,
      label: formatLocalTime(timestamp),
      tooltipLabel: formatLocalDateTime(timestamp),
    };
    const latencyKey = `${pct}LatencyMs`;

    series.forEach((seriesItem) => {
      const item = grouped.get(seriesItem.label)?.find((row) => String(row.timestamp ?? "") === timestamp);
      point[seriesItem.key] = Number(item?.[latencyKey] ?? 0);
    });

    return point;
  });

  const tokenCostPoints = x.map((timestamp) => {
    let inputTokens = 0;
    let outputTokens = 0;
    let costUsd = 0;

    for (const [, items] of grouped) {
      const item = items.find((row) => String(row.timestamp ?? "") === timestamp);
      inputTokens += Number(item?.sumInputTokens ?? 0);
      outputTokens += Number(item?.sumOutputTokens ?? 0);
      costUsd += Number(item?.sumCostInUSD ?? 0);
    }

    return {
      timestamp,
      label: formatLocalTime(timestamp),
      tooltipLabel: formatLocalDateTime(timestamp),
      inputTokens,
      outputTokens,
      costUsd: Number(costUsd.toFixed(6)),
    };
  });

  return { points, latencyPoints, tokenCostPoints, series };
}

function formatTooltipLabel(value: unknown, payload: readonly unknown[]) {
  const first = payload[0] as { payload?: { tooltipLabel?: string } } | undefined;
  return first?.payload?.tooltipLabel ?? String(value ?? "");
}

function normalizeInitialIso(value: string | undefined, fallbackTime: number) {
  const trimmed = value?.trim();
  if (trimmed && !Number.isNaN(Date.parse(trimmed))) return new Date(trimmed).toISOString();
  return new Date(fallbackTime).toISOString();
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

function formatLocalTime(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return new Intl.DateTimeFormat(undefined, {
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

function formatLocalDateTime(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

function formatNumber(value: number) {
  return new Intl.NumberFormat().format(value);
}

function formatCurrency(value: number) {
  return new Intl.NumberFormat(undefined, {
    style: "currency",
    currency: "USD",
    maximumFractionDigits: 6,
  }).format(value);
}
