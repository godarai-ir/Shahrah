"use client";

import {
  Area,
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  ComposedChart,
  Line,
  Pie,
  PieChart,
  XAxis,
  YAxis,
} from "recharts";

import {
  type ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";
import { formatCompactCurrency, formatCompactNumber } from "@/lib/compact-number";

type UsageChartProps = {
  data: Array<{ label: string; cumulative: number | null; projected: number | null }>;
  config: ChartConfig;
  ariaLabel: string;
};

export function UsageChart({ data, config, ariaLabel }: UsageChartProps) {
  return (
    <ChartContainer
      config={config}
      className="h-72 min-w-0 w-full overflow-hidden"
      role="img"
      aria-label={ariaLabel}
    >
      <ComposedChart data={data} margin={{ top: 8, right: 8, bottom: 0, left: 0 }}>
        <defs>
          <linearGradient id="dashboard-usage-fill" x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor="var(--color-cumulative)" stopOpacity={0.32} />
            <stop offset="95%" stopColor="var(--color-cumulative)" stopOpacity={0.03} />
          </linearGradient>
        </defs>
        <CartesianGrid vertical={false} />
        <XAxis dataKey="label" tickLine={false} axisLine={false} minTickGap={28} />
        <YAxis tickLine={false} axisLine={false} width={48} tickFormatter={compactNumber} />
        <ChartTooltip
          content={<ChartTooltipContent formatter={(value) => compactNumber(Number(value))} />}
        />
        <Area
          type="monotone"
          dataKey="cumulative"
          stroke="var(--color-cumulative)"
          fill="url(#dashboard-usage-fill)"
          strokeWidth={2.5}
        />
        <Line
          type="monotone"
          dataKey="projected"
          stroke="var(--color-projected)"
          strokeWidth={2}
          strokeDasharray="6 5"
          dot={false}
          connectNulls
        />
      </ComposedChart>
    </ChartContainer>
  );
}

type ProviderCostChartProps = {
  data: Array<{ provider: string; cost: number; fill: string }>;
  config: ChartConfig;
  ariaLabel: string;
};

export function ProviderCostChart({ data, config, ariaLabel }: ProviderCostChartProps) {
  return (
    <ChartContainer
      config={config}
      className="h-64 min-w-0 w-full"
      role="img"
      aria-label={ariaLabel}
    >
      <PieChart>
        <ChartTooltip
          content={
            <ChartTooltipContent
              nameKey="provider"
              formatter={(value) => compactCurrency(Number(value))}
            />
          }
        />
        <Pie
          data={data}
          dataKey="cost"
          nameKey="provider"
          innerRadius={64}
          outerRadius={92}
          paddingAngle={3}
          strokeWidth={0}
        >
          {data.map((item) => (
            <Cell key={item.provider} fill={item.fill} />
          ))}
        </Pie>
      </PieChart>
    </ChartContainer>
  );
}

type ReliabilityChartProps = {
  data: Array<{ label: string; successful: number; errors: number }>;
  config: ChartConfig;
  ariaLabel: string;
};

export function ReliabilityChart({ data, config, ariaLabel }: ReliabilityChartProps) {
  return (
    <ChartContainer
      config={config}
      className="h-64 min-w-0 w-full overflow-hidden"
      role="img"
      aria-label={ariaLabel}
    >
      <BarChart data={data} margin={{ top: 8, right: 8, bottom: 0, left: 0 }}>
        <CartesianGrid vertical={false} />
        <XAxis dataKey="label" tickLine={false} axisLine={false} minTickGap={24} />
        <YAxis tickLine={false} axisLine={false} width={44} tickFormatter={compactNumber} />
        <ChartTooltip
          content={<ChartTooltipContent formatter={(value) => compactNumber(Number(value))} />}
        />
        <Bar
          dataKey="successful"
          stackId="requests"
          fill="var(--color-successful)"
          radius={[0, 0, 4, 4]}
        />
        <Bar
          dataKey="errors"
          stackId="requests"
          fill="var(--color-errors)"
          radius={[4, 4, 0, 0]}
        />
      </BarChart>
    </ChartContainer>
  );
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
