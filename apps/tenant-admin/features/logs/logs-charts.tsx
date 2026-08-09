"use client";

import { memo, type Dispatch, type ReactNode, type SetStateAction } from "react";
import { Maximize2 } from "lucide-react";
import { useTranslations } from "next-intl";
import { Area, AreaChart, Bar, BarChart, CartesianGrid, ComposedChart, Line, LineChart, XAxis, YAxis } from "recharts";

import { AppCard } from "@/components/app-ui/AppCard";
import { CompactValue } from "@/components/app-ui/CompactValue";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ChartContainer, ChartTooltip, ChartTooltipContent, type ChartConfig } from "@/components/ui/chart";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";

import type { ChartDataPoint, ExpandedChartKey, LogsFormatter } from "./logs-types";
import { formatCacheTimeseries, formatTimeseries } from "./logs-utils";

const chartConfig = {
  requests: { label: "Requests", color: "var(--primary)" },
  errors: { label: "Errors", color: "var(--destructive)" },
  latency: { label: "P95", color: "var(--chart-2)" },
  cost: { label: "Cost", color: "var(--chart-3)" },
  cacheHit: { label: "Cache hit", color: "var(--chart-2)" },
  cacheMiss: { label: "Cache miss", color: "var(--chart-5)" },
} satisfies ChartConfig;

export const LogsCharts = memo(function LogsCharts({
  timeseriesUnavailable,
  breakdownUnavailable,
  timeseriesChartData,
  cacheTimeseriesChartData,
  providerBreakdownData,
  onExpandChart,
  t,
}: {
  timeseriesUnavailable: boolean;
  breakdownUnavailable: boolean;
  timeseriesChartData: ReturnType<typeof formatTimeseries>;
  cacheTimeseriesChartData: ReturnType<typeof formatCacheTimeseries>;
  providerBreakdownData: Array<{ name: string; requests: number }>;
  onExpandChart: Dispatch<SetStateAction<ExpandedChartKey | null>>;
  t: ReturnType<typeof useTranslations>;
}) {
  return (
    <>
      <div className="grid items-stretch gap-4 xl:grid-cols-[1.4fr_1fr]">
        <AppCard
          title={t("charts.traffic")}
          description={t("charts.trafficDescription")}
          actions={<ChartExpandButton label={t("charts.traffic")} onClick={() => onExpandChart("traffic")} t={t} />}
          className="h-full min-w-0 overflow-hidden"
          contentClassName="min-w-0"
        >
          <UnavailableFrame unavailable={timeseriesUnavailable} message={t("analyticsUnavailable")}>
            <ChartContainer config={chartConfig} className="h-72 min-w-0 w-full overflow-hidden">
              <AreaChart data={timeseriesChartData}>
                <CartesianGrid vertical={false} />
                <XAxis dataKey="label" tickLine={false} axisLine={false} minTickGap={24} />
                <YAxis tickLine={false} axisLine={false} />
                <ChartTooltip content={<ChartTooltipContent />} />
                <Area type="monotone" dataKey="requests" stroke="var(--color-requests)" fill="var(--color-requests)" fillOpacity={0.16} />
                <Area type="monotone" dataKey="errors" stroke="var(--color-errors)" fill="var(--color-errors)" fillOpacity={0.12} />
              </AreaChart>
            </ChartContainer>
          </UnavailableFrame>
        </AppCard>

        <AppCard
          title={t("charts.latencyCost")}
          description={t("charts.latencyCostDescription")}
          actions={<ChartExpandButton label={t("charts.latencyCost")} onClick={() => onExpandChart("latencyCost")} t={t} />}
          className="h-full min-w-0 overflow-hidden"
          contentClassName="min-w-0"
        >
          <UnavailableFrame unavailable={timeseriesUnavailable} message={t("analyticsUnavailable")}>
            <ChartContainer config={chartConfig} className="h-72 min-w-0 w-full overflow-hidden">
              <LineChart data={timeseriesChartData}>
                <CartesianGrid vertical={false} />
                <XAxis dataKey="label" tickLine={false} axisLine={false} minTickGap={24} />
                <YAxis tickLine={false} axisLine={false} />
                <ChartTooltip content={<ChartTooltipContent />} />
                <Line type="monotone" dataKey="latency" stroke="var(--color-latency)" dot={false} />
                <Line type="monotone" dataKey="cost" stroke="var(--color-cost)" dot={false} />
              </LineChart>
            </ChartContainer>
          </UnavailableFrame>
        </AppCard>
      </div>

      <div className="grid items-stretch gap-4 xl:grid-cols-[1.4fr_1fr]">
        <AppCard
          title={t("charts.providerBreakdown")}
          description={t("charts.providerBreakdownDescription")}
          actions={<ChartExpandButton label={t("charts.providerBreakdown")} onClick={() => onExpandChart("providerBreakdown")} t={t} />}
          className="h-full min-w-0 overflow-hidden"
          contentClassName="min-w-0"
        >
          <UnavailableFrame unavailable={breakdownUnavailable} message={t("analyticsUnavailable")}>
            <ChartContainer config={chartConfig} className="h-72 min-w-0 w-full overflow-hidden">
              <BarChart data={providerBreakdownData}>
                <CartesianGrid vertical={false} />
                <XAxis dataKey="name" tickLine={false} axisLine={false} />
                <YAxis tickLine={false} axisLine={false} />
                <ChartTooltip content={<ChartTooltipContent />} />
                <Bar dataKey="requests" fill="var(--color-requests)" radius={4} />
              </BarChart>
            </ChartContainer>
          </UnavailableFrame>
        </AppCard>

        <AppCard
          title={t("charts.cacheHitMiss")}
          description={t("charts.cacheHitMissDescription")}
          actions={<ChartExpandButton label={t("charts.cacheHitMiss")} onClick={() => onExpandChart("cacheHitMiss")} t={t} />}
          className="h-full min-w-0 overflow-hidden"
          contentClassName="min-w-0"
        >
          <UnavailableFrame unavailable={timeseriesUnavailable} message={t("analyticsUnavailable")}>
            <ChartContainer config={chartConfig} className="h-72 min-w-0 w-full overflow-hidden">
              <BarChart data={cacheTimeseriesChartData}>
                <CartesianGrid vertical={false} />
                <XAxis dataKey="label" tickLine={false} axisLine={false} minTickGap={24} />
                <YAxis tickLine={false} axisLine={false} />
                <ChartTooltip content={<ChartTooltipContent />} />
                <Bar dataKey="cacheHit" stackId="cache" fill="var(--color-cacheHit)" radius={[4, 4, 0, 0]} />
                <Bar dataKey="cacheMiss" stackId="cache" fill="var(--color-cacheMiss)" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ChartContainer>
          </UnavailableFrame>
        </AppCard>
      </div>
    </>
  );
});

function ChartExpandButton({ label, onClick, t }: { label: string; onClick: () => void; t: ReturnType<typeof useTranslations> }) {
  return (
    <Button variant="outline" size="icon-sm" onClick={onClick} aria-label={t("actions.expandChart", { chart: label })} title={t("actions.expandChart", { chart: label })}>
      <Maximize2 className="h-4 w-4" />
      <span className="sr-only">{t("actions.expandChart", { chart: label })}</span>
    </Button>
  );
}

export function ExpandedChartDialog({
  chart,
  timeseriesUnavailable,
  breakdownUnavailable,
  timeseriesChartData,
  cacheTimeseriesChartData,
  providerBreakdownData,
  onOpenChange,
  formatters,
  t,
}: {
  chart: ExpandedChartKey | null;
  timeseriesUnavailable: boolean;
  breakdownUnavailable: boolean;
  timeseriesChartData: ReturnType<typeof formatTimeseries>;
  cacheTimeseriesChartData: ReturnType<typeof formatCacheTimeseries>;
  providerBreakdownData: Array<{ name: string; requests: number }>;
  onOpenChange: (open: boolean) => void;
  formatters: LogsFormatter;
  t: ReturnType<typeof useTranslations>;
}) {
  if (!chart) return null;

  const chartMeta = getExpandedChartMeta(chart, t);
  const unavailable = chart === "providerBreakdown" ? breakdownUnavailable : timeseriesUnavailable;
  const detailItems = buildExpandedChartDetails(chart, timeseriesChartData, cacheTimeseriesChartData, providerBreakdownData, formatters, t);
  const legendItems = buildExpandedChartLegend(chart, t);

  return (
    <Dialog open={Boolean(chart)} onOpenChange={onOpenChange}>
      <DialogContent className="flex h-[calc(100dvh-1rem)] max-h-[calc(100dvh-1rem)] w-[calc(100vw-1rem)] max-w-[calc(100vw-1rem)] flex-col overflow-hidden p-0 sm:h-[calc(100dvh-2rem)] sm:max-h-[calc(100dvh-2rem)] sm:w-[calc(100vw-2rem)] sm:max-w-[min(1400px,calc(100vw-2rem))]">
        <DialogHeader className="border-b border-border px-4 py-3 sm:px-5 sm:py-3">
          <DialogTitle>{chartMeta.title}</DialogTitle>
          <DialogDescription>{chartMeta.description}</DialogDescription>
        </DialogHeader>
        <div className="flex min-h-0 flex-1 overflow-hidden px-4 py-4 sm:px-5 sm:py-4">
          <div className="flex min-h-0 w-full min-w-0 flex-col gap-3 rounded-xl border border-border bg-card p-3 sm:p-4">
            <div className="flex flex-wrap items-center gap-2">
              {legendItems.map((item) => (
                <Badge key={item.label} variant="outline" className="gap-2 rounded-full px-3 py-1 text-xs">
                  <span className="size-2 rounded-full" style={{ backgroundColor: item.color }} />
                  {item.label}
                </Badge>
              ))}
            </div>
            <div className="min-h-0 flex-1">
              <UnavailableFrame unavailable={unavailable} message={t("analyticsUnavailable")}>
                <ChartContainer config={chartConfig} className="h-full min-h-0 w-full">
                  {renderExpandedChart(chart, timeseriesChartData, cacheTimeseriesChartData, providerBreakdownData, formatters)}
                </ChartContainer>
              </UnavailableFrame>
            </div>
            <ExpandedChartSummaryRow items={detailItems} />
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}

function getExpandedChartMeta(chart: ExpandedChartKey, t: ReturnType<typeof useTranslations>) {
  switch (chart) {
    case "traffic":
      return { title: t("charts.traffic"), description: t("charts.trafficDescription") };
    case "latencyCost":
      return { title: t("charts.latencyCost"), description: t("charts.latencyCostDescription") };
    case "providerBreakdown":
      return { title: t("charts.providerBreakdown"), description: t("charts.providerBreakdownDescription") };
    case "cacheHitMiss":
      return { title: t("charts.cacheHitMiss"), description: t("charts.cacheHitMissDescription") };
  }
}

function buildExpandedChartLegend(chart: ExpandedChartKey, t: ReturnType<typeof useTranslations>) {
  switch (chart) {
    case "traffic":
      return [
        { label: t("chartLegend.requests"), color: "var(--color-requests)" },
        { label: t("chartLegend.errors"), color: "var(--color-errors)" },
      ];
    case "latencyCost":
      return [
        { label: t("chartLegend.latency"), color: "var(--color-latency)" },
        { label: t("chartLegend.cost"), color: "var(--color-cost)" },
      ];
    case "providerBreakdown":
      return [{ label: t("chartLegend.requestShare"), color: "var(--color-requests)" }];
    case "cacheHitMiss":
      return [
        { label: t("chartLegend.cacheHit"), color: "var(--color-cacheHit)" },
        { label: t("chartLegend.cacheMiss"), color: "var(--color-cacheMiss)" },
      ];
  }
}

function renderExpandedChart(
  chart: ExpandedChartKey,
  timeseriesChartData: ReturnType<typeof formatTimeseries>,
  cacheTimeseriesChartData: ReturnType<typeof formatCacheTimeseries>,
  providerBreakdownData: Array<{ name: string; requests: number }>,
  formatters: LogsFormatter,
) {
  switch (chart) {
    case "traffic":
      return (
        <ComposedChart data={timeseriesChartData} margin={{ top: 8, right: 16, bottom: 8, left: 8 }}>
          <CartesianGrid vertical={false} />
          <XAxis dataKey="label" tickLine={false} axisLine={false} minTickGap={24} />
          <YAxis yAxisId="requests" tickLine={false} axisLine={false} width={64} tickFormatter={(value: number) => formatters.compactNumber(value)} />
          <YAxis yAxisId="errors" orientation="right" tickLine={false} axisLine={false} width={56} tickFormatter={(value: number) => formatters.number(value)} />
          <ChartTooltip content={<ChartTooltipContent />} />
          <Area yAxisId="requests" type="monotone" dataKey="requests" stroke="var(--color-requests)" strokeWidth={2} fill="var(--color-requests)" fillOpacity={0.16} />
          <Line yAxisId="errors" type="monotone" dataKey="errors" stroke="var(--color-errors)" dot={false} strokeWidth={2.5} />
        </ComposedChart>
      );
    case "latencyCost":
      return (
        <LineChart data={timeseriesChartData} margin={{ top: 8, right: 12, bottom: 8, left: 8 }}>
          <CartesianGrid vertical={false} />
          <XAxis dataKey="label" tickLine={false} axisLine={false} minTickGap={24} />
          <YAxis yAxisId="latency" tickLine={false} axisLine={false} width={64} tickFormatter={(value: number) => formatters.compactNumber(value)} />
          <YAxis
            yAxisId="cost"
            orientation="right"
            tickLine={false}
            axisLine={false}
            width={68}
            tickFormatter={(value: number) => value >= 1 ? `$${value.toFixed(1)}` : `$${value.toFixed(3)}`}
          />
          <ChartTooltip content={<ChartTooltipContent />} />
          <Line yAxisId="latency" type="monotone" dataKey="latency" stroke="var(--color-latency)" dot={false} strokeWidth={2.5} />
          <Line yAxisId="cost" type="monotone" dataKey="cost" stroke="var(--color-cost)" dot={false} strokeWidth={2.5} />
        </LineChart>
      );
    case "providerBreakdown":
      return (
        <BarChart data={[...providerBreakdownData].sort((a, b) => b.requests - a.requests)} layout="vertical" margin={{ top: 8, right: 24, bottom: 8, left: 24 }}>
          <CartesianGrid horizontal={false} />
          <XAxis type="number" tickLine={false} axisLine={false} tickFormatter={(value: number) => formatters.compactNumber(value)} />
          <YAxis type="category" dataKey="name" tickLine={false} axisLine={false} width={120} />
          <ChartTooltip content={<ChartTooltipContent />} />
          <Bar dataKey="requests" fill="var(--color-requests)" radius={6} maxBarSize={40} />
        </BarChart>
      );
    case "cacheHitMiss":
      return (
        <BarChart data={cacheTimeseriesChartData} margin={{ top: 8, right: 16, bottom: 8, left: 8 }} barGap={10}>
          <CartesianGrid vertical={false} />
          <XAxis dataKey="label" tickLine={false} axisLine={false} minTickGap={24} />
          <YAxis tickLine={false} axisLine={false} width={64} tickFormatter={(value: number) => formatters.compactNumber(value)} />
          <ChartTooltip content={<ChartTooltipContent />} />
          <Bar dataKey="cacheHit" fill="var(--color-cacheHit)" radius={[6, 6, 0, 0]} maxBarSize={28} />
          <Bar dataKey="cacheMiss" fill="var(--color-cacheMiss)" radius={[6, 6, 0, 0]} maxBarSize={28} />
        </BarChart>
      );
  }
}

function buildExpandedChartDetails(
  chart: ExpandedChartKey,
  timeseriesChartData: ReturnType<typeof formatTimeseries>,
  cacheTimeseriesChartData: ReturnType<typeof formatCacheTimeseries>,
  providerBreakdownData: Array<{ name: string; requests: number }>,
  formatters: LogsFormatter,
  t: ReturnType<typeof useTranslations>,
) {
  if (chart === "providerBreakdown") {
    const totalRequests = providerBreakdownData.reduce((sum, item) => sum + item.requests, 0);
    const topProvider = providerBreakdownData.reduce<{ name: string; requests: number } | null>((best, item) => {
      if (!best || item.requests > best.requests) return item;
      return best;
    }, null);

    return [
      { label: t("chartDetails.totalRequests"), value: <CompactValue display={formatters.numberDisplay(totalRequests)} /> },
      {
        label: t("chartDetails.topProvider"),
        value: topProvider?.name ?? t("values.unknown"),
        description: topProvider ? t("chartDetails.topProviderShare", { share: formatters.percent(totalRequests > 0 ? topProvider.requests / totalRequests : 0) }) : undefined,
      },
      { label: t("chartDetails.providerCount"), value: <CompactValue display={formatters.numberDisplay(providerBreakdownData.length)} /> },
      { label: t("chartDetails.longTail"), value: <CompactValue display={formatters.numberDisplay(Math.max(0, totalRequests - (topProvider?.requests ?? 0)))} /> },
    ];
  }

  if (chart === "cacheHitMiss") {
    const totalHits = cacheTimeseriesChartData.reduce((sum, item) => sum + (item.cacheHit ?? 0), 0);
    const totalMisses = cacheTimeseriesChartData.reduce((sum, item) => sum + (item.cacheMiss ?? 0), 0);
    const busiestMiss = cacheTimeseriesChartData.reduce<ChartDataPoint | null>((best, item) => {
      if (!best || (item.cacheMiss ?? 0) > (best.cacheMiss ?? 0)) return item;
      return best;
    }, null);

    return [
      { label: t("chartDetails.totalHits"), value: <CompactValue display={formatters.numberDisplay(totalHits)} /> },
      { label: t("chartDetails.totalMisses"), value: <CompactValue display={formatters.numberDisplay(totalMisses)} /> },
      { label: t("chartDetails.hitRate"), value: formatters.percent(totalHits + totalMisses > 0 ? totalHits / (totalHits + totalMisses) : 0) },
      { label: t("chartDetails.peakMissWindow"), value: busiestMiss?.label ?? "-", description: busiestMiss ? <CompactValue display={formatters.numberDisplay(busiestMiss.cacheMiss ?? 0)} /> : undefined },
    ];
  }

  const totalRequests = timeseriesChartData.reduce((sum, item) => sum + (item.requests ?? 0), 0);
  const totalErrors = timeseriesChartData.reduce((sum, item) => sum + (item.errors ?? 0), 0);
  const peakRequests = timeseriesChartData.reduce<ChartDataPoint | null>((best, item) => {
    if (!best || (item.requests ?? 0) > (best.requests ?? 0)) return item;
    return best;
  }, null);
  const peakLatency = timeseriesChartData.reduce<ChartDataPoint | null>((best, item) => {
    if (!best || (item.latency ?? 0) > (best.latency ?? 0)) return item;
    return best;
  }, null);
  const peakCost = timeseriesChartData.reduce<ChartDataPoint | null>((best, item) => {
    if (!best || (item.cost ?? 0) > (best.cost ?? 0)) return item;
    return best;
  }, null);

  if (chart === "traffic") {
    return [
      { label: t("chartDetails.totalRequests"), value: <CompactValue display={formatters.numberDisplay(totalRequests)} /> },
      { label: t("chartDetails.totalErrors"), value: <CompactValue display={formatters.numberDisplay(totalErrors)} /> },
      { label: t("chartDetails.errorRate"), value: formatters.percent(totalRequests > 0 ? totalErrors / totalRequests : 0) },
      { label: t("chartDetails.peakRequestWindow"), value: peakRequests?.label ?? "-", description: peakRequests ? <CompactValue display={formatters.numberDisplay(peakRequests.requests ?? 0)} /> : undefined },
    ];
  }

  return [
    { label: t("chartDetails.maxLatency"), value: peakLatency ? formatters.ms(peakLatency.latency ?? 0) : "-", description: peakLatency?.label },
    { label: t("chartDetails.maxCost"), value: peakCost ? <CompactValue display={formatters.usdDisplay(peakCost.cost ?? 0)} /> : "-", description: peakCost?.label },
    { label: t("chartDetails.avgLatency"), value: timeseriesChartData.length ? formatters.ms(timeseriesChartData.reduce((sum, item) => sum + (item.latency ?? 0), 0) / timeseriesChartData.length) : "-" },
    { label: t("chartDetails.totalCost"), value: <CompactValue display={formatters.usdDisplay(timeseriesChartData.reduce((sum, item) => sum + (item.cost ?? 0), 0))} /> },
  ];
}

function ExpandedChartSummaryRow({ items }: { items: Array<{ label: string; value: ReactNode; description?: ReactNode }> }) {
  return (
    <div className="grid gap-2 md:grid-cols-2 xl:grid-cols-4">
      {items.map((item) => (
        <div key={item.label} className="rounded-lg border border-border/70 bg-background px-3 py-2.5">
          <div className="truncate text-xs font-medium leading-4 text-muted-foreground">{item.label}</div>
          <div className="mt-1 truncate text-base font-semibold tracking-tight text-foreground sm:text-lg">{item.value}</div>
          {item.description ? <div className="mt-0.5 truncate text-xs leading-4 text-muted-foreground">{item.description}</div> : null}
        </div>
      ))}
    </div>
  );
}

function UnavailableFrame({ unavailable, message, children }: { unavailable: boolean; message: string; children: ReactNode }) {
  if (!unavailable) {
    return children;
  }

  return (
    <div className="flex h-72 items-center justify-center rounded-md border border-dashed border-border bg-muted/40 p-6 text-center text-sm text-muted-foreground">
      {message}
    </div>
  );
}
