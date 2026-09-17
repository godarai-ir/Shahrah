"use client";

import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
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

export { Sparkline } from "./Sparkline";

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
