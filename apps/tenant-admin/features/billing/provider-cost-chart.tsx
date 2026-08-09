"use client";

import { Pie, PieChart, ResponsiveContainer, Tooltip } from "recharts";

import { formatIRR } from "@/features/billing/domain";

type ProviderCostChartProps = {
  data: Array<{ name: string; value: number }>;
};

export default function ProviderCostChart({ data }: ProviderCostChartProps) {
  return (
    <ResponsiveContainer width="100%" height="100%">
      <PieChart>
        <Pie data={data} dataKey="value" nameKey="name" outerRadius={86} fill="var(--primary)" />
        <Tooltip formatter={(value) => formatIRR(Number(value))} />
      </PieChart>
    </ResponsiveContainer>
  );
}
