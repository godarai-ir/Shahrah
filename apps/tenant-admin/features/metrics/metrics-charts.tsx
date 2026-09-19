"use client";

import type { ReactNode } from "react";
import {
  Area,
  AreaChart,
  Bar,
  BarChart,
  CartesianGrid,
  XAxis,
  YAxis,
} from "recharts";

import { AppCard } from "@/components/app-ui/AppCard";
import {
  type ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";
import { formatCompactCurrency, formatCompactNumber } from "@/lib/compact-number";
import type { buildChartData, buildSavingsChartData, MetricKind } from "@/features/metrics/utils";

type FocusMetricChartProps = {
  title: string;
  action?: ReactNode;
  kind: MetricKind;
  locale: string;
  chartData: ReturnType<typeof buildChartData>;
  seriesConfig: ChartConfig;
};

/** The one large chart on /metrics: the selected metric over time, one series per group. */
export function FocusMetricChart({ title, action, kind, locale, chartData, seriesConfig }: FocusMetricChartProps) {
  const data = {
    requests: chartData.points,
    latency: chartData.latencyPoints,
    tokens: chartData.tokenPoints,
    cost: chartData.costPoints,
  }[kind];
  const formatValue = kind === "cost"
    ? (value: number) => formatCompactCurrency(value, { locale, currency: "USD", compactThreshold: 1_000, maximumFractionDigits: 6 }).value
    : (value: number) => formatCompactNumber(value, { locale, compactThreshold: 1_000 }).value;

  return (
    <ChartCard title={title} action={action}>
      <ChartContainer config={seriesConfig} className="h-72 w-full md:h-80">
        <AreaChart data={data} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
          <CartesianGrid vertical={false} />
          <XAxis dataKey="label" tickLine={false} axisLine={false} minTickGap={24} />
          <YAxis tickLine={false} axisLine={false} width={52} allowDecimals={kind !== "requests"} tickFormatter={formatValue} />
          <ChartTooltip content={<ChartTooltipContent labelFormatter={formatTooltipLabel} />} />
          {chartData.series.map((series) => (
            <Area
              key={series.key}
              dataKey={series.key}
              type="monotone"
              fill={`var(--color-${series.key})`}
              fillOpacity={kind === "latency" ? 0.1 : 0.16}
              stroke={`var(--color-${series.key})`}
              strokeWidth={2}
            />
          ))}
        </AreaChart>
      </ChartContainer>
    </ChartCard>
  );
}

type MetricsDetailChartsProps = {
  failuresTitle: string;
  savingsTitle: string;
  locale: string;
  errorChartData: Array<{ errorCode: string; count: number }>;
  savingsChartData: ReturnType<typeof buildSavingsChartData>;
  errorChartConfig: ChartConfig;
  savingsChartConfig: ChartConfig;
};

/** Secondary charts shown inside the collapsible range summary. */
export function MetricsDetailCharts({
  failuresTitle,
  savingsTitle,
  locale,
  errorChartData,
  savingsChartData,
  errorChartConfig,
  savingsChartConfig,
}: MetricsDetailChartsProps) {
  const compactNumber = (value: number) =>
    formatCompactNumber(value, { locale, compactThreshold: 1_000 }).value;
  const compactCurrency = (value: number) =>
    formatCompactCurrency(value, {
      locale,
      currency: "USD",
      compactThreshold: 1_000,
      maximumFractionDigits: 6,
    }).value;

  return (
    <div className="grid min-w-0 gap-3 xl:grid-cols-2">
      <ChartCard title={failuresTitle}>
        <ChartContainer config={errorChartConfig} className="h-[240px] w-full md:h-[260px]">
          <BarChart data={errorChartData} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
            <CartesianGrid vertical={false} />
            <XAxis dataKey="errorCode" tickLine={false} axisLine={false} minTickGap={24} />
            <YAxis tickLine={false} axisLine={false} width={44} tickFormatter={compactNumber} />
            <ChartTooltip content={<ChartTooltipContent />} />
            <Bar dataKey="count" fill="var(--color-count)" radius={[4, 4, 0, 0]} />
          </BarChart>
        </ChartContainer>
      </ChartCard>

      <ChartCard title={savingsTitle}>
        <ChartContainer config={savingsChartConfig} className="h-[240px] w-full md:h-[260px]">
          <BarChart data={savingsChartData} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
            <CartesianGrid vertical={false} />
            <XAxis dataKey="label" tickLine={false} axisLine={false} minTickGap={24} />
            <YAxis tickLine={false} axisLine={false} width={52} tickFormatter={compactCurrency} />
            <ChartTooltip content={<ChartTooltipContent labelFormatter={formatTooltipLabel} />} />
            <Bar dataKey="actualCost" stackId="cost" fill="var(--color-actualCost)" radius={[4, 4, 0, 0]} />
            <Bar dataKey="simpleSavings" stackId="cost" fill="var(--color-simpleSavings)" />
            <Bar dataKey="smartSavings" stackId="cost" fill="var(--color-smartSavings)" radius={[4, 4, 0, 0]} />
          </BarChart>
        </ChartContainer>
      </ChartCard>
    </div>
  );
}

function ChartCard({
  title,
  action,
  children,
}: {
  title: string;
  action?: ReactNode;
  children: ReactNode;
}) {
  return (
    <AppCard
      title={<span className="truncate text-sm">{title}</span>}
      actions={action}
      contentClassName="min-w-0 overflow-hidden pt-0"
      className="min-w-0 overflow-hidden shadow-none"
    >
      {children}
    </AppCard>
  );
}

function formatTooltipLabel(value: unknown, payload: readonly unknown[]) {
  const first = payload[0] as { payload?: { tooltipLabel?: string } } | undefined;
  return first?.payload?.tooltipLabel ?? String(value ?? "");
}
