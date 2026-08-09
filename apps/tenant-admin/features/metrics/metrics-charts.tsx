"use client";

import type { ReactNode } from "react";
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

import { AppCard } from "@/components/app-ui/AppCard";
import {
  type ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";
import { formatCompactCurrency, formatCompactNumber } from "@/lib/compact-number";
import type { buildChartData, buildSavingsChartData } from "@/features/metrics/utils";

type MetricsChartsProps = {
  sectionLabel: string;
  requestTitle: string;
  latencyTitle: string;
  tokensCostTitle: string;
  failuresTitle: string;
  savingsTitle: string;
  latencyAction: ReactNode;
  locale: string;
  chartData: ReturnType<typeof buildChartData>;
  errorChartData: Array<{ errorCode: string; count: number }>;
  savingsChartData: ReturnType<typeof buildSavingsChartData>;
  seriesConfig: ChartConfig;
  tokenChartConfig: ChartConfig;
  errorChartConfig: ChartConfig;
  savingsChartConfig: ChartConfig;
};

export function MetricsCharts({
  sectionLabel,
  requestTitle,
  latencyTitle,
  tokensCostTitle,
  failuresTitle,
  savingsTitle,
  latencyAction,
  locale,
  chartData,
  errorChartData,
  savingsChartData,
  seriesConfig,
  tokenChartConfig,
  errorChartConfig,
  savingsChartConfig,
}: MetricsChartsProps) {
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
    <section aria-label={sectionLabel} className="grid min-w-0 gap-3 xl:grid-cols-2">
      <ChartCard title={requestTitle}>
        <ChartContainer config={seriesConfig} className="h-[240px] w-full md:h-[260px]">
          <AreaChart data={chartData.points} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
            <CartesianGrid vertical={false} />
            <XAxis dataKey="label" tickLine={false} axisLine={false} minTickGap={24} />
            <YAxis tickLine={false} axisLine={false} width={44} allowDecimals={false} tickFormatter={compactNumber} />
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

      <ChartCard title={latencyTitle} action={latencyAction}>
        <ChartContainer config={seriesConfig} className="h-[240px] w-full md:h-[260px]">
          <AreaChart data={chartData.latencyPoints} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
            <CartesianGrid vertical={false} />
            <XAxis dataKey="label" tickLine={false} axisLine={false} minTickGap={24} />
            <YAxis tickLine={false} axisLine={false} width={44} tickFormatter={compactNumber} />
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

      <ChartCard title={tokensCostTitle}>
        <ChartContainer config={tokenChartConfig} className="h-[240px] w-full md:h-[260px]">
          <ComposedChart data={chartData.tokenCostPoints} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
            <CartesianGrid vertical={false} />
            <XAxis dataKey="label" tickLine={false} axisLine={false} minTickGap={24} />
            <YAxis yAxisId="tokens" tickLine={false} axisLine={false} width={44} tickFormatter={compactNumber} />
            <YAxis yAxisId="cost" orientation="right" tickLine={false} axisLine={false} width={52} tickFormatter={compactCurrency} />
            <ChartTooltip content={<ChartTooltipContent labelFormatter={formatTooltipLabel} />} />
            <Bar yAxisId="tokens" dataKey="inputTokens" stackId="tokens" fill="var(--color-inputTokens)" radius={[4, 4, 0, 0]} />
            <Bar yAxisId="tokens" dataKey="outputTokens" stackId="tokens" fill="var(--color-outputTokens)" radius={[4, 4, 0, 0]} />
            <Line yAxisId="cost" dataKey="costUsd" type="monotone" stroke="var(--color-costUsd)" strokeWidth={2} dot={false} />
          </ComposedChart>
        </ChartContainer>
      </ChartCard>

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
    </section>
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
