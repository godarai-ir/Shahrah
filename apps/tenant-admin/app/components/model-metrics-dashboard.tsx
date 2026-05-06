"use client";

import dynamic from "next/dynamic";
import { useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import type { EChartsOption } from "echarts";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

const ReactECharts = dynamic(() => import("echarts-for-react"), { ssr: false });

const VIEW_BY = ["modelName", "virtualModelName", "userEmail", "virtualAccount", "team"] as const;
const PCTS = ["p50", "p75", "p90", "p99"] as const;

type ViewBy = (typeof VIEW_BY)[number];
type Pct = (typeof PCTS)[number];
type Row = Record<string, unknown>;

export default function ModelMetricsDashboard() {
  const t = useTranslations("metrics");
  const common = useTranslations("common");

  const [viewBy, setViewBy] = useState<ViewBy>("modelName");
  const [pct, setPct] = useState<Pct>("p99");
  const [modelFilter, setModelFilter] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [from, setFrom] = useState(() => new Date(Date.now() - 3 * 60 * 60 * 1000).toISOString());
  const [to, setTo] = useState(() => new Date().toISOString());

  const [distribution, setDistribution] = useState<Row[]>([]);
  const [timeseries, setTimeseries] = useState<Row[]>([]);
  const [errorBreakdown, setErrorBreakdown] = useState<Row[]>([]);

  const summary = useMemo(() => {
    const totalRequests = distribution.reduce((acc, row) => acc + Number(row.countModelName ?? row.countAll ?? 0), 0);
    const inputTokens = distribution.reduce((acc, row) => acc + Number(row.sumInputTokens ?? 0), 0);
    const outputTokens = distribution.reduce((acc, row) => acc + Number(row.sumOutputTokens ?? 0), 0);
    const cost = distribution.reduce((acc, row) => acc + Number(row.sumCostInUSD ?? 0), 0);
    return { totalRequests, inputTokens, outputTokens, cost };
  }, [distribution]);

  async function runQuery() {
    setLoading(true);
    setError(null);
    try {
      const filters: Array<Record<string, unknown>> = [];
      if (modelFilter.trim()) {
        filters.push({ fieldName: "modelName", operator: "STRING_CONTAINS", value: modelFilter.trim() });
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
        }),
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
          intervalInSeconds: 300,
        }),
        postMetricsQuery({
          startTs: from,
          endTs: to,
          datasource: "modelMetrics",
          type: "distribution",
          aggregations: [{ type: "count", column: "modelName" }],
          groupBy: ["errorCode"],
          filters: [...filters, { fieldName: "statusCode", operator: "GREATER_THAN_EQUAL", value: 400 }],
        }),
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

  const chartData = useMemo(() => buildSeries(timeseries, viewBy, pct), [timeseries, viewBy, pct]);

  const rpsOption = useMemo<EChartsOption>(() => ({
    backgroundColor: "transparent",
    tooltip: { trigger: "axis" },
    legend: { textStyle: { color: "#94a3b8" } },
    grid: { left: 30, right: 20, top: 40, bottom: 28 },
    xAxis: { type: "category", data: chartData.x, boundaryGap: false, axisLine: { lineStyle: { color: "#334155" } } },
    yAxis: { type: "value", splitLine: { lineStyle: { color: "rgba(148,163,184,0.15)" } } },
    series: chartData.series.map((s, i) => ({
      name: s.name,
      type: "line",
      smooth: true,
      showSymbol: false,
      lineStyle: { width: 2 },
      areaStyle: { opacity: 0.18 },
      emphasis: { focus: "series" },
      data: s.rps,
      color: palette(i),
    })),
  }), [chartData]);

  const latencyOption = useMemo<EChartsOption>(() => ({
    tooltip: { trigger: "axis" },
    legend: { textStyle: { color: "#94a3b8" } },
    grid: { left: 30, right: 20, top: 40, bottom: 28 },
    xAxis: { type: "category", data: chartData.x, boundaryGap: false },
    yAxis: { type: "value" },
    series: chartData.series.map((s, i) => ({
      name: s.name,
      type: "line",
      smooth: 0.25,
      showSymbol: false,
      data: s.latency,
      color: palette(i),
      lineStyle: { width: 2.2, shadowBlur: 8, shadowColor: "rgba(0,0,0,0.2)" },
    })),
  }), [chartData]);

  const tokenCostOption = useMemo<EChartsOption>(() => ({
    tooltip: { trigger: "axis" },
    legend: { textStyle: { color: "#94a3b8" } },
    grid: { left: 30, right: 20, top: 40, bottom: 28 },
    xAxis: { type: "category", data: chartData.x },
    yAxis: [{ type: "value", name: "tokens" }, { type: "value", name: "USD" }],
    series: [
      {
        name: "input tokens",
        type: "bar",
        stack: "tokens",
        data: chartData.totalInput,
        itemStyle: { color: "#22d3ee" },
      },
      {
        name: "output tokens",
        type: "bar",
        stack: "tokens",
        data: chartData.totalOutput,
        itemStyle: { color: "#818cf8" },
      },
      {
        name: "cost usd",
        type: "line",
        yAxisIndex: 1,
        smooth: true,
        data: chartData.totalCost,
        itemStyle: { color: "#f59e0b" },
        lineStyle: { width: 2.5 },
      },
    ],
  }), [chartData]);

  const errorBreakdownOption = useMemo<EChartsOption>(() => ({
    tooltip: { trigger: "axis" },
    xAxis: { type: "category", data: errorBreakdown.map((r) => String(r.errorCode ?? "unknown")) },
    yAxis: { type: "value" },
    series: [{
      type: "bar",
      data: errorBreakdown.map((r) => Number(r.countModelName ?? r.countAll ?? 0)),
      itemStyle: { color: "#fb7185", borderRadius: [6, 6, 0, 0] },
    }],
  }), [errorBreakdown]);

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>{t("title")}</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-3 md:grid-cols-3">
          <div className="grid gap-1.5">
            <Label>{t("filters.from")}</Label>
            <Input value={from} onChange={(e) => setFrom(e.target.value)} />
          </div>
          <div className="grid gap-1.5">
            <Label>{t("filters.to")}</Label>
            <Input value={to} onChange={(e) => setTo(e.target.value)} />
          </div>
          <div className="grid gap-1.5">
            <Label>{t("filters.model")}</Label>
            <Input value={modelFilter} onChange={(e) => setModelFilter(e.target.value)} />
          </div>
          <div className="grid gap-1.5">
            <Label>{t("filters.viewBy")}</Label>
            <select className="h-10 rounded-md border bg-background px-3 text-sm" value={viewBy} onChange={(e) => setViewBy(e.target.value as ViewBy)}>
              {VIEW_BY.map((v) => <option key={v} value={v}>{v}</option>)}
            </select>
          </div>
          <div className="grid gap-1.5">
            <Label>Percentile</Label>
            <select className="h-10 rounded-md border bg-background px-3 text-sm" value={pct} onChange={(e) => setPct(e.target.value as Pct)}>
              {PCTS.map((p) => <option key={p} value={p}>{p.toUpperCase()}</option>)}
            </select>
          </div>
          <div className="flex items-end">
            <Button onClick={runQuery} disabled={loading}>{loading ? common("loading") : common("refresh")}</Button>
          </div>
        </CardContent>
      </Card>

      {error ? (
        <Alert variant="destructive">
          <AlertTitle>{t("alerts.errorTitle")}</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}

      <div className="grid gap-3 md:grid-cols-4">
        <MetricCard title={t("cards.totalRequests")} value={String(summary.totalRequests)} />
        <MetricCard title={t("cards.inputTokens")} value={String(summary.inputTokens)} />
        <MetricCard title={t("cards.outputTokens")} value={String(summary.outputTokens)} />
        <MetricCard title={t("cards.totalCost")} value={summary.cost.toFixed(6)} />
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <ChartCard title="Requests Per Second" option={rpsOption} />
        <ChartCard title={`Request Latency (${pct.toUpperCase()})`} option={latencyOption} />
        <ChartCard title="Tokens & Cost Over Time" option={tokenCostOption} />
        <ChartCard title="Request Failures Breakdown" option={errorBreakdownOption} />
      </div>
    </div>
  );
}

function MetricCard({ title, value }: { title: string; value: string }) {
  return (
    <Card className="bg-gradient-to-br from-slate-950/60 via-slate-900/40 to-slate-950/70 border-slate-800">
      <CardHeader>
        <CardTitle className="text-sm text-slate-300">{title}</CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-2xl font-semibold tracking-tight text-slate-100">{value}</p>
      </CardContent>
    </Card>
  );
}

function ChartCard({ title, option }: { title: string; option: EChartsOption }) {
  return (
    <Card className="border-slate-800 bg-gradient-to-b from-slate-950/60 to-slate-900/50">
      <CardHeader>
        <CardTitle className="text-sm text-slate-200">{title}</CardTitle>
      </CardHeader>
      <CardContent>
        <ReactECharts option={option} style={{ height: 330 }} notMerge lazyUpdate />
      </CardContent>
    </Card>
  );
}

async function postMetricsQuery(body: Record<string, unknown>): Promise<{ data?: Row[] }> {
  const res = await fetch("/api/metrics/query", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
    cache: "no-store",
  });
  const payload = await res.json();
  if (!res.ok) throw new Error(payload?.error?.message ?? `Request failed (${res.status})`);
  return payload as { data?: Row[] };
}

function buildSeries(rows: Row[], viewBy: ViewBy, pct: Pct) {
  const timeSet = new Set<string>();
  const grouped = new Map<string, Row[]>();
  for (const row of rows) {
    const key = String(row[viewBy] ?? "unknown");
    const ts = String(row.timestamp ?? "");
    if (!grouped.has(key)) grouped.set(key, []);
    grouped.get(key)!.push(row);
    if (ts) timeSet.add(ts);
  }
  const x = Array.from(timeSet).sort((a, b) => new Date(a).getTime() - new Date(b).getTime());
  const series = Array.from(grouped.entries()).map(([name, items]) => {
    const index = new Map<string, Row>();
    for (const i of items) index.set(String(i.timestamp ?? ""), i);
    const rps = x.map((ts) => Number(index.get(ts)?.countModelName ?? index.get(ts)?.countAll ?? 0) / 300);
    const latencyKey = `${pct}LatencyMs`;
    const latency = x.map((ts) => Number(index.get(ts)?.[latencyKey] ?? 0));
    return { name, rps, latency };
  });

  const totalInput = x.map((ts) => {
    let total = 0;
    for (const [, items] of grouped) total += Number(items.find((r) => String(r.timestamp) === ts)?.sumInputTokens ?? 0);
    return total;
  });
  const totalOutput = x.map((ts) => {
    let total = 0;
    for (const [, items] of grouped) total += Number(items.find((r) => String(r.timestamp) === ts)?.sumOutputTokens ?? 0);
    return total;
  });
  const totalCost = x.map((ts) => {
    let total = 0;
    for (const [, items] of grouped) total += Number(items.find((r) => String(r.timestamp) === ts)?.sumCostInUSD ?? 0);
    return Number(total.toFixed(6));
  });

  return { x, series, totalInput, totalOutput, totalCost };
}

function palette(i: number) {
  const colors = ["#22d3ee", "#818cf8", "#34d399", "#f59e0b", "#fb7185", "#a78bfa", "#38bdf8"];
  return colors[i % colors.length];
}
