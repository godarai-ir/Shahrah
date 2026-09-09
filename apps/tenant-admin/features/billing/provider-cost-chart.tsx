"use client";

import { Pie, PieChart, ResponsiveContainer, Tooltip } from "recharts";

import { useBillingFormat } from "@/features/billing/use-billing-format";

type ProviderCostChartProps = {
  data: Array<{ name: string; value: number }>;
};

export default function ProviderCostChart({ data }: ProviderCostChartProps) {
  const format = useBillingFormat();

  return (
    <ResponsiveContainer width="100%" height="100%">
      <PieChart>
        <Pie data={data} dataKey="value" nameKey="name" outerRadius={86} fill="var(--primary)" />
        <Tooltip formatter={(value) => format.irr(Number(value))} />
      </PieChart>
    </ResponsiveContainer>
  );
}
