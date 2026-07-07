"use client";

import Link from "next/link";
import { useCallback, useEffect, useMemo, useState, type Dispatch, type ReactNode, type SetStateAction } from "react";
import type { ColumnDef } from "@tanstack/react-table";
import {
  BarChart3,
  Columns3,
  Copy,
  Download,
  FileJson,
  Filter,
  FolderOpen,
  RefreshCw,
  RotateCcw,
  Save,
  Search,
  ShieldCheck,
  Trash2,
  X,
  type LucideIcon,
} from "lucide-react";
import { useLocale, useTranslations } from "next-intl";
import { Area, AreaChart, Bar, BarChart, CartesianGrid, Line, LineChart, XAxis, YAxis } from "recharts";

import { AppCard, AppSelectField, EmptyState, FilterCombobox, MetricCard, PageHeader, StatusBadge } from "@/components/app-ui";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ChartContainer, ChartTooltip, ChartTooltipContent, type ChartConfig } from "@/components/ui/chart";
import { Checkbox } from "@/components/ui/checkbox";
import { DataTable } from "@/components/ui/data-table";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";

type RequestLog = {
  log_id: string;
  request_id: string;
  trace_id?: string;
  provider?: string;
  provider_model?: string;
  model?: string;
  virtual_model?: string;
  route?: string;
  request_type?: string;
  status?: string;
  status_code: number;
  error_type?: string;
  started_at: string;
  duration_ms: number;
  input_tokens?: number;
  output_tokens?: number;
  total_tokens?: number;
  cost_usd?: number;
  cost_irr?: number;
  cache_status?: string;
  retry_count?: number;
  fallback_count?: number;
  fallback_status?: string;
  load_balance_status?: string;
  payload_mode?: string;
};

type LogsResponse = {
  data?: RequestLog[];
  total?: number;
  limit?: number;
  offset?: number;
  next_offset?: number;
};

type Bucket = {
  value: string;
  count: number;
};

type StatusBucket = {
  status_code: number;
  count: number;
};

type LogFacets = {
  providers?: Bucket[];
  models?: Bucket[];
  virtual_models?: Bucket[];
  routes?: Bucket[];
  request_types?: Bucket[];
  error_types?: Bucket[];
  cache_statuses?: Bucket[];
  fallback_statuses?: Bucket[];
  load_balance_statuses?: Bucket[];
  payload_modes?: Bucket[];
  status_codes?: StatusBucket[];
};

type Summary = {
  request_count?: number;
  error_count?: number;
  error_rate?: number;
  latency_avg_ms?: number;
  latency_p50_ms?: number;
  latency_p95_ms?: number;
  latency_p99_ms?: number;
  total_tokens?: number;
  cost_usd?: number;
  cost_irr?: number;
  cache_hit_count?: number;
  cache_hit_rate?: number;
  retry_count?: number;
  fallback_count?: number;
  load_balance_count?: number;
  payload_modes?: Bucket[];
  status_codes?: StatusBucket[];
  providers?: Bucket[];
  models?: Bucket[];
  routes?: Bucket[];
};

type AnalyticsPoint = {
  timestamp?: string;
  group?: string;
  request_count: number;
  error_count: number;
  error_rate: number;
  latency_p95_ms: number;
  total_tokens: number;
  cost_usd: number;
  cache_hit_count: number;
  retry_count: number;
  fallback_count: number;
  load_balance_count: number;
};

type SavedView = {
  saved_view_id: string;
  name: string;
  description?: string;
  filters: Record<string, unknown>;
  visible_columns: string[];
  shared_scope: string;
};

type LogDetail = {
  log: RequestLog;
  content_visibility: {
    can_view_content: boolean;
    request_reason?: string;
    response_reason?: string;
    request_available: boolean;
    response_available: boolean;
  };
  request_payload?: Record<string, unknown>;
  response_payload?: Record<string, unknown>;
  request_payload_meta?: Record<string, unknown>;
  response_payload_meta?: Record<string, unknown>;
  metadata?: Record<string, unknown>;
  actor?: Record<string, string>;
  scope?: Record<string, string>;
};

type ReplayDraft = {
  log_id: string;
  supported: boolean;
  reason?: string;
  route?: string;
  request_type?: string;
  model?: string;
  provider?: string;
  request_payload?: Record<string, unknown>;
  warnings?: string[];
};

type Filters = {
  range: "1h" | "6h" | "24h" | "7d" | "custom";
  from: string;
  to: string;
  log_id: string;
  request_id: string;
  trace_id: string;
  provider: string;
  model: string;
  virtual_model: string;
  route: string;
  request_type: string;
  status_code: string;
  error_type: string;
  cache_status: string;
  fallback_status: string;
  load_balance_status: string;
  payload_mode: string;
  actor_id: string;
  team_id: string;
  org_unit_id: string;
  duration_min: string;
  duration_max: string;
  tokens_min: string;
  tokens_max: string;
  cost_usd_min: string;
  cost_usd_max: string;
  cost_irr_min: string;
  cost_irr_max: string;
  sort_by: string;
  sort_order: "asc" | "desc";
};

type LogsCommandCenterProps = {
  initialLogs?: unknown;
  initialSummary?: unknown;
  initialFacets?: unknown;
  initialSavedViews?: unknown;
  initialError?: string;
};

const DEFAULT_FILTERS: Filters = {
  range: "24h",
  from: "",
  to: "",
  log_id: "",
  request_id: "",
  trace_id: "",
  provider: "",
  model: "",
  virtual_model: "",
  route: "",
  request_type: "",
  status_code: "",
  error_type: "",
  cache_status: "",
  fallback_status: "",
  load_balance_status: "",
  payload_mode: "",
  actor_id: "",
  team_id: "",
  org_unit_id: "",
  duration_min: "",
  duration_max: "",
  tokens_min: "",
  tokens_max: "",
  cost_usd_min: "",
  cost_usd_max: "",
  cost_irr_min: "",
  cost_irr_max: "",
  sort_by: "started_at",
  sort_order: "desc",
};

const DEFAULT_COLUMNS = [
  "status",
  "started_at",
  "provider",
  "model",
  "route",
  "latency",
  "tokens",
  "cost",
  "routing",
  "payload",
  "actions",
];

const rangeMs: Record<Filters["range"], number> = {
  "1h": 60 * 60 * 1000,
  "6h": 6 * 60 * 60 * 1000,
  "24h": 24 * 60 * 60 * 1000,
  "7d": 7 * 24 * 60 * 60 * 1000,
  custom: 0,
};

const chartConfig = {
  requests: { label: "Requests", color: "var(--primary)" },
  errors: { label: "Errors", color: "var(--destructive)" },
  latency: { label: "P95", color: "var(--chart-2)" },
  cost: { label: "Cost", color: "var(--chart-3)" },
  cacheHit: { label: "Cache hit", color: "var(--chart-2)" },
  cacheMiss: { label: "Cache miss", color: "var(--chart-5)" },
} satisfies ChartConfig;

export default function LogsCommandCenter({
  initialLogs,
  initialSummary,
  initialFacets,
  initialSavedViews,
  initialError,
}: LogsCommandCenterProps) {
  const t = useTranslations("logs");
  const locale = useLocale();
  const [filters, setFilters] = useState<Filters>(DEFAULT_FILTERS);
  const [limit, setLimit] = useState(25);
  const [offset, setOffset] = useState(0);
  const [logs, setLogs] = useState<LogsResponse>(() => normalizeLogs(initialLogs));
  const [summary, setSummary] = useState<Summary>(() => normalizeData<Summary>(initialSummary));
  const [facets, setFacets] = useState<LogFacets>(() => normalizeData<LogFacets>(initialFacets));
  const [savedViews, setSavedViews] = useState<SavedView[]>(() => normalizeDataList<SavedView>(initialSavedViews));
  const [timeseries, setTimeseries] = useState<AnalyticsPoint[]>([]);
  const [breakdown, setBreakdown] = useState<AnalyticsPoint[]>([]);
  const [selectedLog, setSelectedLog] = useState<RequestLog | null>(null);
  const [detail, setDetail] = useState<LogDetail | null>(null);
  const [replayDraft, setReplayDraft] = useState<ReplayDraft | null>(null);
  const [viewName, setViewName] = useState("");
  const [saveViewOpen, setSaveViewOpen] = useState(false);
  const [savedViewsOpen, setSavedViewsOpen] = useState(false);
  const [deletingViewID, setDeletingViewID] = useState("");
  const [includeContentExport, setIncludeContentExport] = useState(false);
  const [visibleColumns, setVisibleColumns] = useState<string[]>(DEFAULT_COLUMNS);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(initialError ?? "");
  const [summaryUnavailable, setSummaryUnavailable] = useState(false);
  const [timeseriesUnavailable, setTimeseriesUnavailable] = useState(false);
  const [breakdownUnavailable, setBreakdownUnavailable] = useState(false);

  const queryString = useMemo(() => buildQuery(filters, limit, offset), [filters, limit, offset]);
  const baseQueryString = useMemo(() => buildQuery(filters, undefined, undefined), [filters]);

  const loadData = useCallback(async () => {
    setLoading(true);
    setError("");
    setSummaryUnavailable(false);
    setTimeseriesUnavailable(false);
    setBreakdownUnavailable(false);
    try {
      const logsPayload = await fetchJSON<LogsResponse>(`/api/logs?${queryString}`);
      setLogs(logsPayload);

      const [summaryResult, facetsResult, timeseriesResult, breakdownResult] = await Promise.allSettled([
        fetchJSON<{ data: Summary }>(`/api/logs/summary?${baseQueryString}`),
        fetchJSON<{ data: LogFacets }>(`/api/logs/facets?${baseQueryString}`),
        fetchJSON<{ data: AnalyticsPoint[] }>(`/api/logs/analytics?type=timeseries&group_by=&interval_seconds=3600&${baseQueryString}`),
        fetchJSON<{ data: AnalyticsPoint[] }>(`/api/logs/analytics?type=breakdown&group_by=provider&${baseQueryString}`),
      ]);

      if (summaryResult.status === "fulfilled") setSummary(summaryResult.value.data ?? {});
      else setSummaryUnavailable(true);
      if (facetsResult.status === "fulfilled") setFacets(facetsResult.value.data ?? {});
      if (timeseriesResult.status === "fulfilled") setTimeseries(timeseriesResult.value.data ?? []);
      else setTimeseriesUnavailable(true);
      if (breakdownResult.status === "fulfilled") setBreakdown(breakdownResult.value.data ?? []);
      else setBreakdownUnavailable(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("loadError"));
    } finally {
      setLoading(false);
    }
  }, [baseQueryString, queryString, t]);

  useEffect(() => {
    queueMicrotask(() => {
      void loadData();
    });
  }, [loadData]);

  const openDetail = useCallback(async (item: RequestLog) => {
    setSelectedLog(item);
    setDetail(null);
    setReplayDraft(null);
    try {
      setDetail(await fetchJSON<LogDetail>(`/api/logs/${encodeURIComponent(item.log_id)}`));
    } catch {
      setDetail(null);
    }
  }, []);

  const rows = logs.data ?? [];
  const total = logs.total ?? rows.length;
  const pageStart = total === 0 ? 0 : offset + 1;
  const pageEnd = Math.min(offset + rows.length, total);

  const columns = useMemo<ColumnDef<RequestLog>[]>(() => {
    const allColumns: Array<ColumnDef<RequestLog> & { id: string }> = [
      {
        id: "status",
        header: t("table.status"),
        cell: ({ row }) => <StatusPill code={row.original.status_code} />,
      },
      {
        id: "started_at",
        header: t("table.time"),
        cell: ({ row }) => <span className="whitespace-nowrap">{formatDateTime(row.original.started_at, locale)}</span>,
      },
      {
        id: "provider",
        header: t("table.provider"),
        cell: ({ row }) => <MonoText value={row.original.provider} />,
      },
      {
        id: "model",
        header: t("table.model"),
        cell: ({ row }) => (
          <div className="max-w-44">
            <div className="truncate font-medium" dir="auto">{row.original.model || "-"}</div>
            <div className="truncate text-muted-foreground" dir="auto">{row.original.virtual_model || row.original.provider_model || "-"}</div>
          </div>
        ),
      },
      {
        id: "route",
        header: t("table.route"),
        cell: ({ row }) => <MonoText value={row.original.route} />,
      },
      {
        id: "latency",
        header: t("table.latency"),
        cell: ({ row }) => <span className="whitespace-nowrap">{formatMsValue(row.original.duration_ms, locale)}</span>,
      },
      {
        id: "tokens",
        header: t("table.tokens"),
        cell: ({ row }) => formatNumber(row.original.total_tokens ?? 0, locale),
      },
      {
        id: "cost",
        header: t("table.cost"),
        cell: ({ row }) => (
          <div className="whitespace-nowrap">
            <div>{formatUSD(row.original.cost_usd ?? 0, locale)}</div>
            <div className="text-muted-foreground">{formatIRR(row.original.cost_irr ?? 0, locale)}</div>
          </div>
        ),
      },
      {
        id: "routing",
        header: t("table.routing"),
        cell: ({ row }) => (
          <div className="flex flex-wrap gap-1">
            <RoutingBadge label={t("routing.cache")} value={row.original.cache_status || t("values.noCache")} description={t("routing.cacheHelp")} />
            <RoutingBadge label={t("routing.retry")} value={formatNumber(row.original.retry_count ?? 0, locale)} description={t("routing.retryHelp")} />
            <RoutingBadge label={t("routing.fallback")} value={row.original.fallback_status || t("values.noFallback")} description={t("routing.fallbackHelp")} />
            <RoutingBadge label={t("routing.loadBalance")} value={row.original.load_balance_status || t("values.noLoadBalance")} description={t("routing.loadBalanceHelp")} />
          </div>
        ),
      },
      {
        id: "payload",
        header: t("table.payload"),
        cell: ({ row }) => <StatusBadge status={payloadStatus(row.original.payload_mode)}>{row.original.payload_mode || "-"}</StatusBadge>,
      },
      {
        id: "actions",
        header: "",
        cell: ({ row }) => (
          <div className="flex items-center gap-1">
            <Button variant="ghost" size="icon" onClick={() => void openDetail(row.original)} title={t("actions.inspect")}>
              <Search className="h-4 w-4" />
            </Button>
            {row.original.trace_id ? (
              <Button asChild variant="ghost" size="icon" title={t("actions.openTrace")}>
                <Link href={`/traces/${row.original.trace_id}`}>
                  <BarChart3 className="h-4 w-4" />
                </Link>
              </Button>
            ) : null}
          </div>
        ),
      },
    ];
    return allColumns.filter((column) => visibleColumns.includes(column.id));
  }, [locale, openDetail, t, visibleColumns]);

  const setFilter = <K extends keyof Filters>(key: K, value: Filters[K]) => {
    setOffset(0);
    setFilters((current) => ({ ...current, [key]: value }));
  };

  const activeChips = useMemo(() => activeFilterChips(filters, t), [filters, t]);

  function applySavedView(view: SavedView) {
    setFilters({ ...DEFAULT_FILTERS, ...view.filters } as Filters);
    setVisibleColumns(view.visible_columns?.length ? view.visible_columns : DEFAULT_COLUMNS);
    setOffset(0);
    setSavedViewsOpen(false);
  }

  async function saveCurrentView() {
    if (!viewName.trim()) return;
    try {
      const created = await fetchJSON<SavedView>("/api/logs/saved-views", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: viewName.trim(),
          filters: serializeFilters(filters),
          visible_columns: visibleColumns,
          shared_scope: "tenant",
        }),
      });
      setSavedViews((items) => [created, ...items.filter((item) => item.saved_view_id !== created.saved_view_id)]);
      setViewName("");
      setSaveViewOpen(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("savedViews.saveError"));
    }
  }

  async function deleteSavedView(viewID: string) {
    setDeletingViewID(viewID);
    try {
      await fetchJSON<Record<string, never>>(`/api/logs/saved-views/${encodeURIComponent(viewID)}`, { method: "DELETE" });
      setSavedViews((items) => items.filter((item) => item.saved_view_id !== viewID));
    } catch (err) {
      setError(err instanceof Error ? err.message : t("savedViews.deleteError"));
    } finally {
      setDeletingViewID("");
    }
  }

  async function createExport(format: "csv" | "jsonl") {
    const job = await fetchJSON<{ export_id: string }>("/api/logs/exports", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        filters: {
          ...serializeFilters(filters),
          limit: String(limit),
          offset: String(offset),
          sort_by: filters.sort_by,
          sort_order: filters.sort_order,
        },
        format,
        include_content: includeContentExport,
      }),
    });
    window.location.href = `/api/logs/exports/${encodeURIComponent(job.export_id)}/download`;
  }

  async function createReplayDraft() {
    if (!selectedLog) return;
    const draft = await fetchJSON<ReplayDraft>(`/api/logs/${encodeURIComponent(selectedLog.log_id)}/replay-draft`, {
      method: "POST",
    });
    setReplayDraft(draft);
  }

  return (
    <div className="flex flex-col gap-4">
      <PageHeader
        eyebrow={t("eyebrow")}
        title={t("title")}
        description={t("subtitle")}
        actions={
          <div className="flex flex-wrap items-center gap-2">
            <Button variant="outline" size="sm" onClick={() => void loadData()} disabled={loading}>
              <RefreshCw className="h-4 w-4" />
              {t("actions.refresh")}
            </Button>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" size="sm">
                  <Download className="h-4 w-4" />
                  {t("actions.export")}
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-56">
                <div className="flex items-center gap-2 px-2 py-2 text-xs">
                  <Checkbox checked={includeContentExport} onCheckedChange={(value) => setIncludeContentExport(value === true)} />
                  <span>{t("export.includeContent")}</span>
                </div>
                <Button variant="ghost" className="w-full justify-start" onClick={() => void createExport("csv")}>
                  {t("export.csv")}
                </Button>
                <Button variant="ghost" className="w-full justify-start" onClick={() => void createExport("jsonl")}>
                  {t("export.jsonl")}
                </Button>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        }
      />

      {error ? (
        <Alert variant="destructive">
          <AlertTitle>{t("loadError")}</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}
      <AppCard
        title={t("filters.title")}
        description={t("filters.description")}
      >
        <div className="grid gap-3 lg:grid-cols-6">
          <LabeledControl label={t("filters.range")}>
            <AppSelectField
              value={filters.range}
              onValueChange={(value) => setFilter("range", value as Filters["range"])}
              options={(["1h", "6h", "24h", "7d", "custom"] as const).map((value) => ({ value, label: t(`ranges.${value}`) }))}
              ariaLabel={t("filters.range")}
              triggerClassName="h-8 min-w-0 text-xs"
            />
          </LabeledControl>
          <FacetSelect label={t("filters.provider")} value={filters.provider} onChange={(value) => setFilter("provider", value)} items={facets.providers} empty={t("filters.noOptions")} />
          <FacetSelect label={t("filters.model")} value={filters.model} onChange={(value) => setFilter("model", value)} items={facets.models} empty={t("filters.noOptions")} />
          <FacetSelect label={t("filters.route")} value={filters.route} onChange={(value) => setFilter("route", value)} items={facets.routes} empty={t("filters.noOptions")} />
          <FacetSelect label={t("filters.status")} value={filters.status_code} onChange={(value) => setFilter("status_code", value)} items={(facets.status_codes ?? []).map((item) => ({ value: String(item.status_code), count: item.count }))} empty={t("filters.noOptions")} />
          <FacetSelect label={t("filters.payload")} value={filters.payload_mode} onChange={(value) => setFilter("payload_mode", value)} items={facets.payload_modes} empty={t("filters.noOptions")} />
        </div>

        <div className="grid gap-3 pt-3 lg:grid-cols-6">
          <TextFilter label={t("filters.logId")} value={filters.log_id} onChange={(value) => setFilter("log_id", value)} mono />
          <TextFilter label={t("filters.requestId")} value={filters.request_id} onChange={(value) => setFilter("request_id", value)} mono />
          <TextFilter label={t("filters.traceId")} value={filters.trace_id} onChange={(value) => setFilter("trace_id", value)} mono />
          <FacetSelect label={t("filters.cache")} value={filters.cache_status} onChange={(value) => setFilter("cache_status", value)} items={facets.cache_statuses} empty={t("filters.noOptions")} />
          <FacetSelect label={t("filters.fallback")} value={filters.fallback_status} onChange={(value) => setFilter("fallback_status", value)} items={facets.fallback_statuses} empty={t("filters.noOptions")} />
          <FacetSelect label={t("filters.loadBalance")} value={filters.load_balance_status} onChange={(value) => setFilter("load_balance_status", value)} items={facets.load_balance_statuses} empty={t("filters.noOptions")} />
        </div>

        <div className="grid gap-3 pt-3 lg:grid-cols-8">
          <TextFilter label={t("filters.actor")} value={filters.actor_id} onChange={(value) => setFilter("actor_id", value)} mono />
          <TextFilter label={t("filters.team")} value={filters.team_id} onChange={(value) => setFilter("team_id", value)} mono />
          <TextFilter label={t("filters.orgUnit")} value={filters.org_unit_id} onChange={(value) => setFilter("org_unit_id", value)} mono />
          <TextFilter label={t("filters.durationMin")} value={filters.duration_min} onChange={(value) => setFilter("duration_min", value)} />
          <TextFilter label={t("filters.durationMax")} value={filters.duration_max} onChange={(value) => setFilter("duration_max", value)} />
          <TextFilter label={t("filters.tokensMin")} value={filters.tokens_min} onChange={(value) => setFilter("tokens_min", value)} />
          <TextFilter label={t("filters.costMin")} value={filters.cost_usd_min} onChange={(value) => setFilter("cost_usd_min", value)} />
          <Button variant="outline" size="sm" className="self-end" onClick={() => { setFilters(DEFAULT_FILTERS); setOffset(0); }}>
            <RotateCcw className="h-4 w-4" />
            {t("actions.reset")}
          </Button>
        </div>

        {activeChips.length ? (
          <div className="flex flex-wrap gap-2 pt-3">
            {activeChips.map((chip) => (
              <Badge key={chip.key} variant="secondary" className="gap-1">
                {chip.label}
                <button type="button" onClick={() => setFilter(chip.key as keyof Filters, "" as never)} title={t("actions.removeFilter")}>
                  <X className="h-3 w-3" />
                </button>
              </Badge>
            ))}
          </div>
        ) : null}
      </AppCard>

      <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-5">
        <AnalyticsMetric unavailable={summaryUnavailable} label={t("cards.requests")} value={formatNumber(summary.request_count ?? 0, locale)} icon={BarChart3} t={t} />
        <AnalyticsMetric unavailable={summaryUnavailable} label={t("cards.errorRate")} value={formatPercent(summary.error_rate ?? 0, locale)} description={formatNumber(summary.error_count ?? 0, locale)} icon={ShieldCheck} t={t} />
        <AnalyticsMetric unavailable={summaryUnavailable} label={t("cards.p95")} value={formatMs(summary.latency_p95_ms ?? 0, locale)} description={`${t("cards.p99")} ${formatMs(summary.latency_p99_ms ?? 0, locale)}`} t={t} />
        <AnalyticsMetric unavailable={summaryUnavailable} label={t("cards.cost")} value={formatUSD(summary.cost_usd ?? 0, locale)} description={formatIRR(summary.cost_irr ?? 0, locale)} t={t} />
        <AnalyticsMetric unavailable={summaryUnavailable} label={t("cards.cache")} value={formatPercent(summary.cache_hit_rate ?? 0, locale)} description={`${t("cards.retries")} ${formatNumber(summary.retry_count ?? 0, locale)}`} t={t} />
      </div>

      <div className="grid gap-4 xl:grid-cols-[1.4fr_1fr]">
        <AppCard title={t("charts.traffic")} description={t("charts.trafficDescription")}>
          <UnavailableFrame unavailable={timeseriesUnavailable} message={t("analyticsUnavailable")}>
            <ChartContainer config={chartConfig} className="h-72 w-full">
              <AreaChart data={formatTimeseries(timeseries, locale)}>
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

        <AppCard title={t("charts.latencyCost")} description={t("charts.latencyCostDescription")}>
          <UnavailableFrame unavailable={timeseriesUnavailable} message={t("analyticsUnavailable")}>
            <ChartContainer config={chartConfig} className="h-72 w-full">
              <LineChart data={formatTimeseries(timeseries, locale)}>
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

      <div className="grid gap-4 xl:grid-cols-2">
        <AppCard title={t("charts.providerBreakdown")} description={t("charts.providerBreakdownDescription")}>
          <UnavailableFrame unavailable={breakdownUnavailable} message={t("analyticsUnavailable")}>
            <ChartContainer config={chartConfig} className="h-72 w-full">
              <BarChart data={breakdown.map((item) => ({ name: item.group || t("values.unknown"), requests: item.request_count }))}>
                <CartesianGrid vertical={false} />
                <XAxis dataKey="name" tickLine={false} axisLine={false} />
                <YAxis tickLine={false} axisLine={false} />
                <ChartTooltip content={<ChartTooltipContent />} />
                <Bar dataKey="requests" fill="var(--color-requests)" radius={4} />
              </BarChart>
            </ChartContainer>
          </UnavailableFrame>
        </AppCard>

        <AppCard title={t("charts.cacheHitMiss")} description={t("charts.cacheHitMissDescription")}>
          <UnavailableFrame unavailable={timeseriesUnavailable} message={t("analyticsUnavailable")}>
            <ChartContainer config={chartConfig} className="h-72 w-full">
              <BarChart data={formatCacheTimeseries(timeseries, locale)}>
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

      <AppCard
        title={t("table.title")}
        actions={
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              className="h-8 px-2 sm:px-3"
              onClick={() => setSaveViewOpen(true)}
              aria-label={t("actions.saveView")}
              title={t("actions.saveView")}
            >
              <Save className="h-4 w-4" />
              <span className="hidden sm:inline">{t("actions.saveView")}</span>
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="h-8 px-2 sm:px-3"
              onClick={() => setSavedViewsOpen(true)}
              aria-label={t("actions.savedViews")}
              title={t("actions.savedViews")}
            >
              <FolderOpen className="h-4 w-4" />
              <span className="hidden sm:inline">{t("actions.savedViews")}</span>
            </Button>
            <ColumnMenu visibleColumns={visibleColumns} setVisibleColumns={setVisibleColumns} t={t} />
          </div>
        }
        headerClassName="flex-row items-center gap-2"
        titleContainerClassName="flex-1"
        actionsClassName="flex-nowrap"
      >
        {rows.length ? <DataTable columns={columns} data={rows} emptyLabel={t("empty")} /> : <EmptyState title={t("empty")} description={t("emptyDescription")} icon={Filter} />}
        <div className="flex flex-wrap items-center justify-between gap-2 pt-4">
          <div className="text-xs text-muted-foreground">{t("table.description", { start: pageStart, end: pageEnd, total })}</div>
          <div className="flex items-center gap-2">
            <Button variant="outline" size="sm" disabled={offset === 0 || loading} onClick={() => setOffset(Math.max(0, offset - limit))}>{t("actions.previous")}</Button>
            <AppSelectField
              value={String(limit)}
              onValueChange={(value) => {
                setLimit(Number(value));
                setOffset(0);
              }}
              options={[25, 50, 100, 200].map((value) => ({ value: String(value), label: value }))}
              ariaLabel={t("table.pageSize")}
              triggerClassName="h-8 min-w-20 text-xs"
            />
            <Button variant="outline" size="sm" disabled={offset + rows.length >= total || loading} onClick={() => setOffset(offset + limit)}>{t("actions.next")}</Button>
          </div>
        </div>
      </AppCard>

      <Dialog open={saveViewOpen} onOpenChange={setSaveViewOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("savedViews.saveTitle")}</DialogTitle>
            <DialogDescription>{t("savedViews.saveDescription")}</DialogDescription>
          </DialogHeader>
          <div className="grid gap-2">
            <Label htmlFor="logs-save-view-name">{t("savedViews.nameLabel")}</Label>
            <Input
              id="logs-save-view-name"
              value={viewName}
              onChange={(event) => setViewName(event.target.value)}
              placeholder={t("savedViews.namePlaceholder")}
            />
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setSaveViewOpen(false)}>
              {t("actions.cancel")}
            </Button>
            <Button type="button" onClick={() => void saveCurrentView()} disabled={!viewName.trim()}>
              <Save className="h-4 w-4" />
              {t("actions.saveView")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={savedViewsOpen} onOpenChange={setSavedViewsOpen}>
        <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>{t("savedViews.manageTitle")}</DialogTitle>
            <DialogDescription>{t("savedViews.manageDescription")}</DialogDescription>
          </DialogHeader>
          {savedViews.length ? (
            <div className="grid gap-2">
              {savedViews.map((view) => (
                <div key={view.saved_view_id} className="flex flex-col gap-3 rounded-md border border-border bg-background p-3 sm:flex-row sm:items-center sm:justify-between">
                  <div className="min-w-0">
                    <div className="truncate text-sm font-medium">{view.name}</div>
                    <div className="truncate text-xs text-muted-foreground">{view.shared_scope}</div>
                  </div>
                  <div className="flex shrink-0 items-center gap-2">
                    <Button type="button" size="sm" variant="outline" onClick={() => applySavedView(view)}>
                      <FolderOpen className="h-4 w-4" />
                      {t("actions.load")}
                    </Button>
                    <Button
                      type="button"
                      size="sm"
                      variant="destructive"
                      onClick={() => void deleteSavedView(view.saved_view_id)}
                      disabled={deletingViewID === view.saved_view_id}
                    >
                      <Trash2 className="h-4 w-4" />
                      {deletingViewID === view.saved_view_id ? t("savedViews.deleting") : t("actions.delete")}
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <EmptyState title={t("savedViews.emptyTitle")} description={t("savedViews.emptyDescription")} />
          )}
        </DialogContent>
      </Dialog>

      <Sheet open={selectedLog !== null} onOpenChange={(open) => { if (!open) setSelectedLog(null); }}>
        <SheetContent side="left" className="w-full overflow-y-auto sm:max-w-2xl">
          <SheetHeader>
            <SheetTitle>{t("detailTitle")}</SheetTitle>
            <SheetDescription>{selectedLog?.log_id}</SheetDescription>
          </SheetHeader>
          {detail ? (
            <div className="flex flex-col gap-4 pt-4">
              <div className="grid gap-2 md:grid-cols-3">
                <MetricCard label={t("table.status")} value={detail.log.status_code} />
                <MetricCard label={t("table.latency")} value={formatMs(detail.log.duration_ms, locale)} />
                <MetricCard label={t("table.cost")} value={formatUSD(detail.log.cost_usd ?? 0, locale)} />
              </div>
              <div className="flex flex-wrap gap-2">
                {detail.log.trace_id ? <Button asChild size="sm" variant="outline"><Link href={`/traces/${detail.log.trace_id}`}>{t("actions.openTrace")}</Link></Button> : null}
                <Button size="sm" variant="outline" onClick={() => void navigator.clipboard.writeText(window.location.href)}>
                  <Copy className="h-4 w-4" />
                  {t("actions.copyLink")}
                </Button>
                <Button size="sm" onClick={() => void createReplayDraft()}>
                  <FileJson className="h-4 w-4" />
                  {t("actions.replayDraft")}
                </Button>
              </div>
              {replayDraft ? (
                <Alert variant={replayDraft.supported ? "default" : "destructive"}>
                  <AlertTitle>{replayDraft.supported ? t("replay.ready") : t("replay.unsupported")}</AlertTitle>
                  <AlertDescription>{replayDraft.reason || replayDraft.warnings?.join(" / ")}</AlertDescription>
                </Alert>
              ) : null}
              <Tabs defaultValue="summary">
                <TabsList>
                  <TabsTrigger value="summary">{t("tabs.summary")}</TabsTrigger>
                  <TabsTrigger value="request">{t("content.request")}</TabsTrigger>
                  <TabsTrigger value="response">{t("content.response")}</TabsTrigger>
                  <TabsTrigger value="metadata">{t("tabs.metadata")}</TabsTrigger>
                </TabsList>
                <TabsContent value="summary"><JSONBlock value={{ log: detail.log, actor: detail.actor, scope: detail.scope, content_visibility: detail.content_visibility }} /></TabsContent>
                <TabsContent value="request"><PayloadBlock available={detail.content_visibility.request_available} value={detail.request_payload} reason={detail.content_visibility.request_reason} hidden={t("content.hidden")} /></TabsContent>
                <TabsContent value="response"><PayloadBlock available={detail.content_visibility.response_available} value={detail.response_payload} reason={detail.content_visibility.response_reason} hidden={t("content.hidden")} /></TabsContent>
                <TabsContent value="metadata"><JSONBlock value={{ metadata: detail.metadata, request_payload_meta: detail.request_payload_meta, response_payload_meta: detail.response_payload_meta }} /></TabsContent>
              </Tabs>
            </div>
          ) : (
            <EmptyState title={t("detailLoading")} description={t("detailLoadingDescription")} />
          )}
        </SheetContent>
      </Sheet>
    </div>
  );
}

function LabeledControl({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="flex min-w-0 flex-col gap-1">
      <Label className="text-xs text-muted-foreground">{label}</Label>
      {children}
    </div>
  );
}

function TextFilter({ label, value, onChange, mono }: { label: string; value: string; onChange: (value: string) => void; mono?: boolean }) {
  return (
    <LabeledControl label={label}>
      <Input value={value} onChange={(event) => onChange(event.target.value)} className={mono ? "h-8 font-mono text-xs" : "h-8 text-xs"} />
    </LabeledControl>
  );
}

function FacetSelect({ label, value, onChange, items, empty }: { label: string; value: string; onChange: (value: string) => void; items?: Array<{ value: string; count: number }>; empty: string }) {
  const options = [{ value: "", label: label }, ...(items ?? []).filter((item) => item.value).map((item) => ({ value: item.value, label: `${item.value} (${item.count})` }))];
  return (
    <LabeledControl label={label}>
      <FilterCombobox id={`logs-${label}`} value={value} onValueChange={onChange} options={options} placeholder={label} emptyText={empty} />
    </LabeledControl>
  );
}

function StatusPill({ code }: { code: number }) {
  if (code >= 500) return <StatusBadge status="error">{code}</StatusBadge>;
  if (code >= 400) return <StatusBadge status="warning">{code}</StatusBadge>;
  return <StatusBadge status="success">{code || "-"}</StatusBadge>;
}

function RoutingBadge({ label, value, description }: { label: string; value: string; description: string }) {
  return (
    <TooltipProvider delayDuration={120}>
      <Tooltip>
        <TooltipTrigger asChild>
          <Badge variant="outline" className="max-w-36 cursor-help gap-1 truncate text-[11px]">
            <span className="text-muted-foreground">{label}</span>
            <span dir="auto">{value}</span>
          </Badge>
        </TooltipTrigger>
        <TooltipContent side="top" className="max-w-64 text-right leading-5" dir="rtl">
          {description}
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

function ColumnMenu({
  visibleColumns,
  setVisibleColumns,
  t,
}: {
  visibleColumns: string[];
  setVisibleColumns: Dispatch<SetStateAction<string[]>>;
  t: ReturnType<typeof useTranslations>;
}) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" size="sm" className="h-8 px-2 sm:px-3" aria-label={t("actions.columns")} title={t("actions.columns")}>
          <Columns3 className="h-4 w-4" />
          <span className="hidden sm:inline">{t("actions.columns")}</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-48 text-right">
        <div dir="rtl">
          {DEFAULT_COLUMNS.map((column) => (
            <DropdownMenuCheckboxItem
              key={column}
              checked={visibleColumns.includes(column)}
              className="text-right"
              onCheckedChange={(checked) => {
                setVisibleColumns((current) => checked ? [...current, column] : current.filter((item) => item !== column));
              }}
            >
              <span className="w-full text-right">{t(`columns.${column}`)}</span>
            </DropdownMenuCheckboxItem>
          ))}
        </div>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

function AnalyticsMetric({
  unavailable,
  label,
  value,
  description,
  icon,
  t,
}: {
  unavailable: boolean;
  label: string;
  value: ReactNode;
  description?: ReactNode;
  icon?: LucideIcon;
  t: ReturnType<typeof useTranslations>;
}) {
  if (!unavailable) {
    return <MetricCard label={label} value={value} description={description} icon={icon} />;
  }

  return (
    <div className="rounded-lg border border-dashed border-border bg-muted/40 p-4 opacity-70">
      <div className="text-xs font-medium text-muted-foreground">{label}</div>
      <div className="pt-2 text-lg font-semibold text-muted-foreground">{t("analyticsUnavailableShort")}</div>
      <div className="pt-1 text-xs text-muted-foreground">{t("analyticsUnavailable")}</div>
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

function MonoText({ value }: { value?: string }) {
  return <span className="block max-w-40 truncate font-mono text-xs" dir="auto">{value || "-"}</span>;
}

function JSONBlock({ value }: { value: unknown }) {
  return <pre className="max-h-[520px] overflow-auto rounded-md border border-border bg-muted p-3 text-xs leading-relaxed" dir="ltr">{JSON.stringify(value ?? {}, null, 2)}</pre>;
}

function PayloadBlock({ available, value, reason, hidden }: { available: boolean; value?: unknown; reason?: string; hidden: string }) {
  if (!available || !value) {
    return <Alert><AlertTitle>{hidden}</AlertTitle><AlertDescription>{reason}</AlertDescription></Alert>;
  }
  return <JSONBlock value={value} />;
}

function normalizeLogs(payload: unknown): LogsResponse {
  if (!payload || typeof payload !== "object") return { data: [], total: 0, limit: 25, offset: 0 };
  return payload as LogsResponse;
}

function normalizeData<T>(payload: unknown): T {
  if (payload && typeof payload === "object" && "data" in payload) {
    return (payload as { data?: T }).data ?? ({} as T);
  }
  return {} as T;
}

function normalizeDataList<T>(payload: unknown): T[] {
  if (payload && typeof payload === "object" && "data" in payload) {
    return (payload as { data?: T[] }).data ?? [];
  }
  return [];
}

async function fetchJSON<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, { ...init, cache: "no-store" });
  const payload = await response.json().catch(() => ({}));
  if (!response.ok) {
    const message = payload?.error?.message ?? payload?.message ?? `Request failed (${response.status})`;
    throw new Error(message);
  }
  return payload as T;
}

function buildQuery(filters: Filters, limit?: number, offset?: number) {
  const params = new URLSearchParams();
  const serialized = serializeFilters(filters);
  for (const [key, value] of Object.entries(serialized)) {
    if (value !== "") params.set(key, String(value));
  }
  if (limit !== undefined) params.set("limit", String(limit));
  if (offset !== undefined) params.set("offset", String(offset));
  return params.toString();
}

function serializeFilters(filters: Filters): Record<string, string> {
  const out: Record<string, string> = {};
  for (const [key, value] of Object.entries(filters)) {
    if (key === "range") continue;
    if (value) out[key] = String(value);
  }
  if (filters.range !== "custom") {
    out.from = new Date(Date.now() - rangeMs[filters.range]).toISOString();
    out.to = new Date().toISOString();
  }
  return out;
}

function activeFilterChips(filters: Filters, t: ReturnType<typeof useTranslations>) {
  const labels: Partial<Record<keyof Filters, string>> = {
    from: t("filters.from"),
    to: t("filters.to"),
    log_id: t("filters.log_id"),
    request_id: t("filters.request_id"),
    trace_id: t("filters.trace_id"),
    provider: t("filters.provider"),
    model: t("filters.model"),
    virtual_model: t("filters.virtual_model"),
    route: t("filters.route"),
    request_type: t("filters.request_type"),
    status_code: t("filters.status_code"),
    error_type: t("filters.error_type"),
    cache_status: t("filters.cache_status"),
    fallback_status: t("filters.fallback_status"),
    load_balance_status: t("filters.load_balance_status"),
    payload_mode: t("filters.payload_mode"),
    actor_id: t("filters.actor_id"),
    team_id: t("filters.team_id"),
    org_unit_id: t("filters.org_unit_id"),
    duration_min: t("filters.duration_min"),
    duration_max: t("filters.duration_max"),
    tokens_min: t("filters.tokens_min"),
    tokens_max: t("filters.tokens_max"),
    cost_usd_min: t("filters.cost_usd_min"),
    cost_usd_max: t("filters.cost_usd_max"),
    cost_irr_min: t("filters.cost_irr_min"),
    cost_irr_max: t("filters.cost_irr_max"),
  };
  return Object.entries(filters)
    .filter(([key, value]) => key !== "sort_by" && key !== "sort_order" && key !== "range" && Boolean(value))
    .map(([key, value]) => ({ key, label: `${labels[key as keyof Filters] ?? key}: ${value}` }));
}

function formatTimeseries(points: AnalyticsPoint[], locale: string) {
  return points.map((point) => ({
    label: point.timestamp ? new Intl.DateTimeFormat(locale, { hour: "2-digit", minute: "2-digit" }).format(new Date(point.timestamp)) : point.group,
    requests: point.request_count,
    errors: point.error_count,
    latency: Math.round(point.latency_p95_ms ?? 0),
    cost: Number((point.cost_usd ?? 0).toFixed(4)),
  }));
}

function formatCacheTimeseries(points: AnalyticsPoint[], locale: string) {
  return points.map((point) => {
    const cacheHit = point.cache_hit_count ?? 0;
    return {
      label: point.timestamp ? new Intl.DateTimeFormat(locale, { hour: "2-digit", minute: "2-digit" }).format(new Date(point.timestamp)) : point.group,
      cacheHit,
      cacheMiss: Math.max(0, (point.request_count ?? 0) - cacheHit),
    };
  });
}

function payloadStatus(mode?: string) {
  if (!mode || mode === "none") return "inactive";
  if (mode.includes("redacted") || mode.includes("metadata")) return "warning";
  return "success";
}

function formatNumber(value: number, locale: string) {
  return new Intl.NumberFormat(locale).format(value);
}

function formatPercent(value: number, locale: string) {
  return new Intl.NumberFormat(locale, { style: "percent", maximumFractionDigits: 1 }).format(value);
}

function formatUSD(value: number, locale: string) {
  return new Intl.NumberFormat(locale, { style: "currency", currency: "USD", maximumFractionDigits: 4 }).format(value);
}

function formatIRR(value: number, locale: string) {
  return new Intl.NumberFormat(locale, { maximumFractionDigits: 0 }).format(value) + " IRR";
}

function formatMs(value: number, locale: string) {
  return `${formatNumber(Math.round(value), locale)} ms`;
}

function formatMsValue(value: number, locale: string) {
  return formatNumber(Math.round(value), locale);
}

function formatDateTime(value: string, locale: string) {
  if (!value) return "-";
  const date = new Date(value);
  const day = new Intl.DateTimeFormat(locale, { weekday: "short" }).format(date);
  const time = new Intl.DateTimeFormat(locale, { hour: "2-digit", minute: "2-digit", second: "2-digit" }).format(date);
  return `${day}${locale.startsWith("fa") ? "،" : ","} ${time}`;
}
