"use client";

import Link from "next/link";
import { memo, useCallback, useEffect, useMemo, useRef, useState, type Dispatch, type ReactNode, type SetStateAction } from "react";
import { flexRender, getCoreRowModel, useReactTable, type ColumnDef } from "@tanstack/react-table";
import {
  BarChart3,
  ChevronDown,
  Columns3,
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
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { Slider } from "@/components/ui/slider";
import { PrincipalPicker, type PrincipalPickerItem } from "@/app/components/principal-picker";
import { type PrincipalSearchKind, type PrincipalSearchOption } from "@/app/components/principal-search-input";
import { buildPrincipalSelectorLabels } from "@/app/components/provider-account-wizard/principal-labels";
import { cn } from "@/lib/utils";

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

type LogRange = {
  min?: number;
  max?: number;
};

type LogRangeStats = {
  duration_ms?: LogRange;
  total_tokens?: LogRange;
  cost_usd?: LogRange;
  cost_irr?: LogRange;
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
  safe_headers?: Record<string, string>;
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
  actor_type: string;
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

type LogsFormatter = {
  number: (value: number) => string;
  percent: (value: number) => string;
  usd: (value: number) => string;
  irr: (value: number) => string;
  ms: (value: number) => string;
  msValue: (value: number) => string;
  timeLabel: (value?: string) => string | undefined;
  dateTime: (value: string) => string;
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
  actor_type: "",
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
  "payload",
  "actions",
];

const OPTIONAL_COLUMNS = [
  "cache_status",
  "retry_count",
  "fallback_status",
  "load_balance_status",
];

const COLUMN_CHOICES = [
  "status",
  "started_at",
  "provider",
  "model",
  "route",
  "latency",
  "tokens",
  "cost",
  "payload",
  "cache_status",
  "retry_count",
  "fallback_status",
  "load_balance_status",
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
  const [rangeStats, setRangeStats] = useState<LogRangeStats>({});
  const [savedViews, setSavedViews] = useState<SavedView[]>(() => normalizeDataList<SavedView>(initialSavedViews));
  const [timeseries, setTimeseries] = useState<AnalyticsPoint[]>([]);
  const [breakdown, setBreakdown] = useState<AnalyticsPoint[]>([]);
  const [selectedLog, setSelectedLog] = useState<RequestLog | null>(null);
  const [detail, setDetail] = useState<LogDetail | null>(null);
  const [traceAvailable, setTraceAvailable] = useState(false);
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
  const loadAbortRef = useRef<AbortController | null>(null);
  const loadRequestRef = useRef(0);
  const detailAbortRef = useRef<AbortController | null>(null);
  const detailRequestRef = useRef(0);

  const formatters = useMemo(() => createLogFormatters(locale), [locale]);
  const queryString = useMemo(() => buildQuery(filters, limit, offset), [filters, limit, offset]);
  const baseQueryString = useMemo(() => buildQuery(filters, undefined, undefined), [filters]);
  const rangeStatsQueryString = useMemo(() => buildRangeStatsQuery(filters), [filters]);
  const timeseriesChartData = useMemo(() => formatTimeseries(timeseries, formatters), [timeseries, formatters]);
  const cacheTimeseriesChartData = useMemo(() => formatCacheTimeseries(timeseries, formatters), [timeseries, formatters]);
  const providerBreakdownData = useMemo(
    () => breakdown.map((item) => ({ name: item.group || t("values.unknown"), requests: item.request_count })),
    [breakdown, t],
  );
  const principalT = useTranslations("providerAccounts");
  const principalUx = useTranslations("providerAccounts.ux");
  const principalSelectorLabels = useMemo(() => buildPrincipalSelectorLabels(principalT, principalUx), [principalT, principalUx]);

  const loadData = useCallback(async () => {
    loadAbortRef.current?.abort();
    const controller = new AbortController();
    loadAbortRef.current = controller;
    const requestID = ++loadRequestRef.current;
    const isCurrentRequest = () => loadRequestRef.current === requestID && loadAbortRef.current === controller && !controller.signal.aborted;

    setLoading(true);
    setError("");
    setSummaryUnavailable(false);
    setTimeseriesUnavailable(false);
    setBreakdownUnavailable(false);
    try {
      const logsPayload = await fetchJSON<LogsResponse>(`/api/logs?${queryString}`, { signal: controller.signal });
      if (!isCurrentRequest()) return;
      setLogs(logsPayload);

      const [summaryResult, facetsResult, timeseriesResult, breakdownResult] = await Promise.allSettled([
        fetchJSON<{ data: Summary }>(`/api/logs/summary?${baseQueryString}`, { signal: controller.signal }),
        fetchJSON<{ data: LogFacets }>(`/api/logs/facets?${baseQueryString}`, { signal: controller.signal }),
        fetchJSON<{ data: AnalyticsPoint[] }>(`/api/logs/analytics?type=timeseries&group_by=&interval_seconds=3600&${baseQueryString}`, { signal: controller.signal }),
        fetchJSON<{ data: AnalyticsPoint[] }>(`/api/logs/analytics?type=breakdown&group_by=provider&${baseQueryString}`, { signal: controller.signal }),
      ]);

      if (!isCurrentRequest()) return;
      if (summaryResult.status === "fulfilled") setSummary(summaryResult.value.data ?? {});
      else setSummaryUnavailable(true);
      if (facetsResult.status === "fulfilled") setFacets(facetsResult.value.data ?? {});
      if (timeseriesResult.status === "fulfilled") setTimeseries(timeseriesResult.value.data ?? []);
      else setTimeseriesUnavailable(true);
      if (breakdownResult.status === "fulfilled") setBreakdown(breakdownResult.value.data ?? []);
      else setBreakdownUnavailable(true);
    } catch (err) {
      if (isAbortError(err)) return;
      if (!isCurrentRequest()) return;
      setError(err instanceof Error ? err.message : t("loadError"));
    } finally {
      if (isCurrentRequest()) setLoading(false);
      if (loadAbortRef.current === controller) loadAbortRef.current = null;
    }
  }, [baseQueryString, queryString, t]);

  useEffect(() => {
    let disposed = false;
    queueMicrotask(() => {
      if (!disposed) void loadData();
    });
    return () => {
      disposed = true;
      loadAbortRef.current?.abort();
    };
  }, [loadData]);

  useEffect(() => {
    const controller = new AbortController();
    const suffix = rangeStatsQueryString ? `?${rangeStatsQueryString}` : "";
    void fetchCachedJSON<{ data?: LogRangeStats }>(`/api/logs/range-stats${suffix}`, { signal: controller.signal })
      .then((payload) => {
        setRangeStats(payload.data ?? {});
      })
      .catch((err) => {
        if (!isAbortError(err)) setRangeStats({});
      });
    return () => controller.abort();
  }, [rangeStatsQueryString]);

  const openDetail = useCallback(async (item: RequestLog) => {
    detailAbortRef.current?.abort();
    const controller = new AbortController();
    detailAbortRef.current = controller;
    const requestID = ++detailRequestRef.current;
    const isCurrentRequest = () => detailRequestRef.current === requestID && detailAbortRef.current === controller && !controller.signal.aborted;

    setSelectedLog(item);
    setDetail(null);
    setTraceAvailable(false);
    setReplayDraft(null);
    try {
      const payload = await fetchJSON<LogDetail>(`/api/logs/${encodeURIComponent(item.log_id)}`, { signal: controller.signal });
      if (!isCurrentRequest()) return;
      setDetail(payload);

      const traceID = payload.log.trace_id?.trim();
      if (traceID) {
        try {
          await fetchJSON<unknown>(`/api/traces/${encodeURIComponent(traceID)}`, { signal: controller.signal });
          if (isCurrentRequest()) setTraceAvailable(true);
        } catch (traceErr) {
          if (!isAbortError(traceErr) && isCurrentRequest()) setTraceAvailable(false);
        }
      }
    } catch (err) {
      if (isAbortError(err) || !isCurrentRequest()) return;
      setDetail(null);
      setTraceAvailable(false);
    } finally {
      if (detailAbortRef.current === controller) detailAbortRef.current = null;
    }
  }, []);

  const closeDetail = useCallback(() => {
    detailAbortRef.current?.abort();
    detailAbortRef.current = null;
    detailRequestRef.current += 1;
    setSelectedLog(null);
    setDetail(null);
    setTraceAvailable(false);
    setReplayDraft(null);
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
        cell: ({ row }) => <span className="whitespace-nowrap">{formatters.dateTime(row.original.started_at)}</span>,
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
        cell: ({ row }) => <span className="whitespace-nowrap">{formatters.msValue(row.original.duration_ms)}</span>,
      },
      {
        id: "tokens",
        header: t("table.tokens"),
        cell: ({ row }) => formatters.number(row.original.total_tokens ?? 0),
      },
      {
        id: "cost",
        header: t("table.cost"),
        cell: ({ row }) => (
          <div className="whitespace-nowrap">
            <div>{formatters.usd(row.original.cost_usd ?? 0)}</div>
            <div className="text-muted-foreground">{formatters.irr(row.original.cost_irr ?? 0)}</div>
          </div>
        ),
      },
      {
        id: "cache_status",
        header: t("table.cacheStatus"),
        cell: ({ row }) => <RoutingValue value={row.original.cache_status || t("values.noCache")} description={t("routing.cacheHelp")} />,
      },
      {
        id: "retry_count",
        header: t("table.retryCount"),
        cell: ({ row }) => <RoutingValue value={formatters.number(row.original.retry_count ?? 0)} description={t("routing.retryHelp")} />,
      },
      {
        id: "fallback_status",
        header: t("table.fallbackStatus"),
        cell: ({ row }) => <RoutingValue value={row.original.fallback_status || ""} description={t("routing.fallbackHelp")} />,
      },
      {
        id: "load_balance_status",
        header: t("table.loadBalanceStatus"),
        cell: ({ row }) => <RoutingValue value={row.original.load_balance_status || ""} description={t("routing.loadBalanceHelp")} />,
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
          </div>
        ),
      },
    ];
    return allColumns.filter((column) => visibleColumns.includes(column.id));
  }, [formatters, openDetail, t, visibleColumns]);

  const setFilter = useCallback(<K extends keyof Filters>(key: K, value: Filters[K]) => {
    setOffset(0);
    setFilters((current) => ({ ...current, [key]: value }));
  }, []);

  const setStringFilter = useCallback((key: keyof Filters, value: string) => {
    setFilter(key, value as never);
  }, [setFilter]);

  const searchLogPrincipals = useCallback(async (kind: PrincipalSearchKind, query: string): Promise<PrincipalSearchOption[]> => {
    const normalized = query.trim();
    if (kind === "user" || kind === "team") {
      const payload = await fetchJSON<{ data?: Array<{ principal_id: string; display_name: string; secondary?: string; type?: string }> }>(
        `/api/access-principals?kind=${kind}&q=${encodeURIComponent(normalized)}&limit=20`,
      );
      return (payload.data ?? []).map((item) => ({
        principal_id: item.principal_id,
        display_name: item.display_name,
        secondary: item.secondary,
        type: kind,
      }));
    }
    if (kind === "org_unit") {
      const payload = await fetchJSON<{ data?: Array<{ id?: string; org_unit_id?: string; alias?: string; display_name?: string; name?: string }> }>("/api/gateway/lookups/organization-units");
      return filterLookupPrincipals(payload.data ?? [], normalized, "org_unit");
    }
    if (kind === "virtual_account") {
      const payload = await fetchJSON<{ data?: Array<{ id?: string; virtual_user_id?: string; alias?: string; display_name?: string; name?: string }> }>("/api/gateway/lookups/virtual-accounts");
      return filterLookupPrincipals(payload.data ?? [], normalized, "virtual_account");
    }
    return [];
  }, []);

  const resetFilters = useCallback(() => {
    setFilters(DEFAULT_FILTERS);
    setOffset(0);
  }, []);

  const setPageLimit = useCallback((value: number) => {
    setLimit(value);
    setOffset(0);
  }, []);

  const previousPage = useCallback(() => {
    setOffset((current) => Math.max(0, current - limit));
  }, [limit]);

  const nextPage = useCallback(() => {
    setOffset((current) => current + limit);
  }, [limit]);

  const openSaveViewDialog = useCallback(() => setSaveViewOpen(true), []);
  const openSavedViewsDialog = useCallback(() => setSavedViewsOpen(true), []);

  const activeChips = useMemo(() => activeFilterChips(filters, t), [filters, t]);

  const applySavedView = useCallback((view: SavedView) => {
    setFilters({ ...DEFAULT_FILTERS, ...view.filters } as Filters);
    setVisibleColumns(normalizeVisibleColumns(view.visible_columns));
    setOffset(0);
    setSavedViewsOpen(false);
  }, []);

  const saveCurrentView = useCallback(async () => {
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
  }, [filters, t, viewName, visibleColumns]);

  const deleteSavedView = useCallback(async (viewID: string) => {
    setDeletingViewID(viewID);
    try {
      await fetchJSON<Record<string, never>>(`/api/logs/saved-views/${encodeURIComponent(viewID)}`, { method: "DELETE" });
      setSavedViews((items) => items.filter((item) => item.saved_view_id !== viewID));
    } catch (err) {
      setError(err instanceof Error ? err.message : t("savedViews.deleteError"));
    } finally {
      setDeletingViewID("");
    }
  }, [t]);

  const createExport = useCallback(async (format: "csv" | "jsonl") => {
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
  }, [filters, includeContentExport, limit, offset]);

  const createReplayDraft = useCallback(async () => {
    if (!selectedLog) return;
    const draft = await fetchJSON<ReplayDraft>(`/api/logs/${encodeURIComponent(selectedLog.log_id)}/replay-draft`, {
      method: "POST",
    });
    setReplayDraft(draft);
  }, [selectedLog]);

  return (
    <TooltipProvider delayDuration={120}>
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
      <LogsFilters
        filters={filters}
        facets={facets}
        rangeStats={rangeStats}
        activeChips={activeChips}
        onSetFilter={setStringFilter}
        onSearchPrincipal={searchLogPrincipals}
        principalSelectorLabels={principalSelectorLabels}
        onReset={resetFilters}
        formatters={formatters}
        t={t}
      />
      <LogsMetrics summary={summary} summaryUnavailable={summaryUnavailable} formatters={formatters} t={t} />
      <LogsCharts
        timeseriesUnavailable={timeseriesUnavailable}
        breakdownUnavailable={breakdownUnavailable}
        timeseriesChartData={timeseriesChartData}
        cacheTimeseriesChartData={cacheTimeseriesChartData}
        providerBreakdownData={providerBreakdownData}
        t={t}
      />
      <LogsTable
        rows={rows}
        columns={columns}
        pageStart={pageStart}
        pageEnd={pageEnd}
        total={total}
        offset={offset}
        limit={limit}
        loading={loading}
        visibleColumns={visibleColumns}
        setVisibleColumns={setVisibleColumns}
        onOpenSaveView={openSaveViewDialog}
        onOpenSavedViews={openSavedViewsDialog}
        onPreviousPage={previousPage}
        onNextPage={nextPage}
        onPageLimitChange={setPageLimit}
        t={t}
      />
      <SavedViewsDialogs
        saveViewOpen={saveViewOpen}
        savedViewsOpen={savedViewsOpen}
        viewName={viewName}
        savedViews={savedViews}
        deletingViewID={deletingViewID}
        setSaveViewOpen={setSaveViewOpen}
        setSavedViewsOpen={setSavedViewsOpen}
        setViewName={setViewName}
        onSaveCurrentView={saveCurrentView}
        onApplySavedView={applySavedView}
        onDeleteSavedView={deleteSavedView}
        t={t}
      />
      <LogDetailSheet
        selectedLog={selectedLog}
        detail={detail}
        traceAvailable={traceAvailable}
        replayDraft={replayDraft}
        onClose={closeDetail}
        onCreateReplayDraft={createReplayDraft}
        formatters={formatters}
        t={t}
      />
    </div>
    </TooltipProvider>
  );
}

const LogsFilters = memo(function LogsFilters({
  filters,
  facets,
  rangeStats,
  activeChips,
  onSetFilter,
  onSearchPrincipal,
  principalSelectorLabels,
  onReset,
  formatters,
  t,
}: {
  filters: Filters;
  facets: LogFacets;
  rangeStats: LogRangeStats;
  activeChips: Array<{ key: string; label: string }>;
  onSetFilter: (key: keyof Filters, value: string) => void;
  onSearchPrincipal: (kind: PrincipalSearchKind, query: string) => Promise<PrincipalSearchOption[]>;
  principalSelectorLabels: ReturnType<typeof buildPrincipalSelectorLabels>;
  onReset: () => void;
  formatters: LogsFormatter;
  t: ReturnType<typeof useTranslations>;
}) {
  const [advancedOpen, setAdvancedOpen] = useState(false);
  const [principalNamesByKey, setPrincipalNamesByKey] = useState<Record<string, string>>({});
  const selectedPrincipals = useMemo(
    () => principalsFromFilters(filters, principalNamesByKey),
    [filters, principalNamesByKey],
  );
  const isPersian = t("filters.range") !== "Range";
  const showAdvancedLabel = isPersian ? "فیلترهای پیشرفته‌تر" : "Advanced filters";
  const hideAdvancedLabel = isPersian ? "پنهان کردن فیلترهای پیشرفته" : "Hide advanced filters";
  const setRange = useCallback((value: string) => {
    onSetFilter("range", value);
    if (value === "custom") {
      const now = new Date();
      if (!filters.from) onSetFilter("from", new Date(now.getTime() - rangeMs["24h"]).toISOString());
      if (!filters.to) onSetFilter("to", now.toISOString());
      return;
    }
    onSetFilter("from", "");
    onSetFilter("to", "");
  }, [filters.from, filters.to, onSetFilter]);
  const setSelectedPrincipals = useCallback((items: PrincipalPickerItem[]) => {
    setPrincipalNamesByKey((current) => {
      const next = { ...current };
      for (const item of items) {
        const key = principalKey(item);
        if (key) next[key] = item.display_name || item.principal_id || item.type;
      }
      return next;
    });
    const users = items.filter((item) => item.type === "user" || item.type === "virtual_account");
    onSetFilter("actor_type", joinUnique(users.map((item) => item.type === "virtual_account" ? "virtual_user" : "user")));
    onSetFilter("actor_id", joinUnique(users.map((item) => item.principal_id ?? "")));
    onSetFilter("team_id", joinUnique(items.filter((item) => item.type === "team").map((item) => item.principal_id ?? "")));
    onSetFilter("org_unit_id", joinUnique(items.filter((item) => item.type === "org_unit").map((item) => item.principal_id ?? "")));
  }, [onSetFilter]);

  return (
    <AppCard title={t("filters.title")} description={t("filters.description")}>
      <Collapsible open={advancedOpen} onOpenChange={setAdvancedOpen} className="flex flex-col gap-3">
        <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
          <LabeledControl label={t("filters.range")}>
            <AppSelectField
              value={filters.range}
              onValueChange={setRange}
              options={[
                { value: "1h", label: t("ranges.1h") },
                { value: "6h", label: t("ranges.6h") },
                { value: "24h", label: t("ranges.24h") },
                { value: "7d", label: t("ranges.7d") },
                { value: "custom", label: t("ranges.custom") },
              ]}
              ariaLabel={t("filters.range")}
              triggerClassName="h-8 text-xs"
            />
          </LabeledControl>
          <FacetSelect label={t("filters.provider")} value={filters.provider} onChange={(value) => onSetFilter("provider", value)} items={facets.providers} empty={t("filters.noOptions")} />
          <FacetSelect label={t("filters.model")} value={filters.model} onChange={(value) => onSetFilter("model", value)} items={facets.models} empty={t("filters.noOptions")} />
          {filters.range === "custom" ? (
            <CustomDateRange
              from={filters.from}
              to={filters.to}
              fromLabel={t("filters.from")}
              toLabel={t("filters.to")}
              onFromChange={(value) => onSetFilter("from", value)}
              onToChange={(value) => onSetFilter("to", value)}
              className="md:col-span-2 xl:col-span-3"
            />
          ) : null}
        </div>
        <PrincipalPicker
          id="logs-principal-filter"
          title={t("filters.principals")}
          selected={selectedPrincipals}
          setSelected={setSelectedPrincipals}
          onSearchPrincipal={onSearchPrincipal}
          principalSelectorLabels={principalSelectorLabels}
          allowEveryone={false}
          addLabel={t("filters.addPrincipal")}
          removeLabel={t("filters.removePrincipal")}
          kinds={["user", "virtual_account", "team", "org_unit"]}
          variant="filter"
        />
        <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
          <RangeFilter
            label={t("filters.duration")}
            minLabel={t("filters.duration_min")}
            maxLabel={t("filters.duration_max")}
            range={rangeStats.duration_ms}
            minValue={filters.duration_min}
            maxValue={filters.duration_max}
            onChange={(min, max) => {
              onSetFilter("duration_min", min);
              onSetFilter("duration_max", max);
            }}
            formatValue={formatters.ms}
          />
          <RangeFilter
            label={t("filters.tokens")}
            minLabel={t("filters.tokens_min")}
            maxLabel={t("filters.tokens_max")}
            range={rangeStats.total_tokens}
            minValue={filters.tokens_min}
            maxValue={filters.tokens_max}
            onChange={(min, max) => {
              onSetFilter("tokens_min", min);
              onSetFilter("tokens_max", max);
            }}
            formatValue={formatters.number}
          />
          <RangeFilter
            label={t("filters.cost_usd")}
            minLabel={t("filters.cost_usd_min")}
            maxLabel={t("filters.cost_usd_max")}
            range={rangeStats.cost_usd}
            minValue={filters.cost_usd_min}
            maxValue={filters.cost_usd_max}
            onChange={(min, max) => {
              onSetFilter("cost_usd_min", min);
              onSetFilter("cost_usd_max", max);
            }}
            formatValue={formatters.usd}
          />
          <RangeFilter
            label={t("filters.cost_irr")}
            minLabel={t("filters.cost_irr_min")}
            maxLabel={t("filters.cost_irr_max")}
            range={rangeStats.cost_irr}
            minValue={filters.cost_irr_min}
            maxValue={filters.cost_irr_max}
            onChange={(min, max) => {
              onSetFilter("cost_irr_min", min);
              onSetFilter("cost_irr_max", max);
            }}
            formatValue={formatters.irr}
          />
        </div>
        <div className="flex flex-wrap items-center justify-between gap-2 pt-1">
          <CollapsibleTrigger asChild>
            <Button type="button" variant="outline" size="sm" aria-expanded={advancedOpen}>
              <ChevronDown className={cn("h-4 w-4 transition-transform", advancedOpen && "rotate-180")} />
              {advancedOpen ? hideAdvancedLabel : showAdvancedLabel}
            </Button>
          </CollapsibleTrigger>
          <Button type="button" variant="outline" size="sm" onClick={onReset}>
            <RotateCcw className="h-4 w-4" />
            {t("actions.reset")}
          </Button>
        </div>
        <CollapsibleContent>
          <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
            <FacetSelect label={t("filters.route")} value={filters.route} onChange={(value) => onSetFilter("route", value)} items={facets.routes} empty={t("filters.noOptions")} />
            <TextFilter label={t("filters.log_id")} value={filters.log_id} onChange={(value) => onSetFilter("log_id", value)} mono />
            <TextFilter label={t("filters.request_id")} value={filters.request_id} onChange={(value) => onSetFilter("request_id", value)} mono />
            <TextFilter label={t("filters.trace_id")} value={filters.trace_id} onChange={(value) => onSetFilter("trace_id", value)} mono />
            <FacetSelect
              label={t("filters.status_code")}
              value={filters.status_code}
              onChange={(value) => onSetFilter("status_code", value)}
              items={facets.status_codes?.map((item) => ({ value: String(item.status_code), count: item.count }))}
              empty={t("filters.noOptions")}
            />
            <FacetSelect label={t("filters.error_type")} value={filters.error_type} onChange={(value) => onSetFilter("error_type", value)} items={facets.error_types} empty={t("filters.noOptions")} />
            <FacetSelect label={t("filters.cache_status")} value={filters.cache_status} onChange={(value) => onSetFilter("cache_status", value)} items={facets.cache_statuses} empty={t("filters.noOptions")} />
            <FacetSelect label={t("filters.fallback_status")} value={filters.fallback_status} onChange={(value) => onSetFilter("fallback_status", value)} items={facets.fallback_statuses} empty={t("filters.noOptions")} />
            <FacetSelect label={t("filters.load_balance_status")} value={filters.load_balance_status} onChange={(value) => onSetFilter("load_balance_status", value)} items={facets.load_balance_statuses} empty={t("filters.noOptions")} />
            <FacetSelect label={t("filters.payload_mode")} value={filters.payload_mode} onChange={(value) => onSetFilter("payload_mode", value)} items={facets.payload_modes} empty={t("filters.noOptions")} />
          </div>
        </CollapsibleContent>
      </Collapsible>
      {activeChips.length ? (
      <div className="flex flex-wrap items-center gap-2 pt-4">
        {activeChips.map((chip) => (
          <Button key={chip.key} type="button" variant="secondary" size="sm" onClick={() => onSetFilter(chip.key as keyof Filters, "")}>
            <X className="h-3 w-3" />
            {chip.label}
          </Button>
        ))}
      </div>
      ) : null}
    </AppCard>
  );
});

const LogsMetrics = memo(function LogsMetrics({
  summary,
  summaryUnavailable,
  formatters,
  t,
}: {
  summary: Summary;
  summaryUnavailable: boolean;
  formatters: LogsFormatter;
  t: ReturnType<typeof useTranslations>;
}) {
  return (
    <div className="grid grid-cols-2 gap-2 sm:gap-3 md:grid-cols-3 xl:grid-cols-5">
      <AnalyticsMetric unavailable={summaryUnavailable} label={t("cards.requests")} value={formatters.number(summary.request_count ?? 0)} icon={BarChart3} t={t} />
      <AnalyticsMetric unavailable={summaryUnavailable} label={t("cards.errorRate")} value={formatters.percent(summary.error_rate ?? 0)} description={formatters.number(summary.error_count ?? 0)} icon={ShieldCheck} t={t} />
      <AnalyticsMetric unavailable={summaryUnavailable} label={t("cards.p95")} value={formatters.ms(summary.latency_p95_ms ?? 0)} description={`${t("cards.p99")} ${formatters.ms(summary.latency_p99_ms ?? 0)}`} t={t} />
      <AnalyticsMetric unavailable={summaryUnavailable} label={t("cards.cost")} value={formatters.usd(summary.cost_usd ?? 0)} description={formatters.irr(summary.cost_irr ?? 0)} t={t} />
      <AnalyticsMetric unavailable={summaryUnavailable} label={t("cards.cache")} value={formatters.percent(summary.cache_hit_rate ?? 0)} description={`${t("cards.retries")} ${formatters.number(summary.retry_count ?? 0)}`} t={t} />
    </div>
  );
});

const LogsCharts = memo(function LogsCharts({
  timeseriesUnavailable,
  breakdownUnavailable,
  timeseriesChartData,
  cacheTimeseriesChartData,
  providerBreakdownData,
  t,
}: {
  timeseriesUnavailable: boolean;
  breakdownUnavailable: boolean;
  timeseriesChartData: ReturnType<typeof formatTimeseries>;
  cacheTimeseriesChartData: ReturnType<typeof formatCacheTimeseries>;
  providerBreakdownData: Array<{ name: string; requests: number }>;
  t: ReturnType<typeof useTranslations>;
}) {
  return (
    <>
      <div className="grid items-stretch gap-4 xl:grid-cols-[1.4fr_1fr]">
        <AppCard title={t("charts.traffic")} description={t("charts.trafficDescription")} className="h-full min-w-0 overflow-hidden" contentClassName="min-w-0">
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

        <AppCard title={t("charts.latencyCost")} description={t("charts.latencyCostDescription")} className="h-full min-w-0 overflow-hidden" contentClassName="min-w-0">
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
        <AppCard title={t("charts.providerBreakdown")} description={t("charts.providerBreakdownDescription")} className="h-full min-w-0 overflow-hidden" contentClassName="min-w-0">
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

        <AppCard title={t("charts.cacheHitMiss")} description={t("charts.cacheHitMissDescription")} className="h-full min-w-0 overflow-hidden" contentClassName="min-w-0">
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

const LogsTable = memo(function LogsTable({
  rows,
  columns,
  pageStart,
  pageEnd,
  total,
  offset,
  limit,
  loading,
  visibleColumns,
  setVisibleColumns,
  onOpenSaveView,
  onOpenSavedViews,
  onPreviousPage,
  onNextPage,
  onPageLimitChange,
  t,
}: {
  rows: RequestLog[];
  columns: ColumnDef<RequestLog>[];
  pageStart: number;
  pageEnd: number;
  total: number;
  offset: number;
  limit: number;
  loading: boolean;
  visibleColumns: string[];
  setVisibleColumns: Dispatch<SetStateAction<string[]>>;
  onOpenSaveView: () => void;
  onOpenSavedViews: () => void;
  onPreviousPage: () => void;
  onNextPage: () => void;
  onPageLimitChange: (value: number) => void;
  t: ReturnType<typeof useTranslations>;
}) {
  return (
    <AppCard
      title={t("table.title")}
      actions={
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" className="h-8 px-2 sm:px-3" onClick={onOpenSaveView} aria-label={t("actions.saveView")} title={t("actions.saveView")}>
            <Save className="h-4 w-4" />
            <span className="hidden sm:inline">{t("actions.saveView")}</span>
          </Button>
          <Button variant="outline" size="sm" className="h-8 px-2 sm:px-3" onClick={onOpenSavedViews} aria-label={t("actions.savedViews")} title={t("actions.savedViews")}>
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
      {rows.length ? (
        <LogsDataTable
          columns={columns}
          data={rows}
          emptyLabel={t("empty")}
        />
      ) : (
        <EmptyState title={t("empty")} description={t("emptyDescription")} icon={Filter} />
      )}
      <div className="flex flex-wrap items-center justify-between gap-2 pt-4">
        <div className="text-xs text-muted-foreground">{t("table.description", { start: pageStart, end: pageEnd, total })}</div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" disabled={offset === 0 || loading} onClick={onPreviousPage}>{t("actions.previous")}</Button>
          <AppSelectField
            value={String(limit)}
            onValueChange={(value) => onPageLimitChange(Number(value))}
            options={[25, 50, 100, 200].map((value) => ({ value: String(value), label: value }))}
            ariaLabel={t("table.pageSize")}
            triggerClassName="h-8 min-w-20 text-xs"
          />
          <Button variant="outline" size="sm" disabled={offset + rows.length >= total || loading} onClick={onNextPage}>{t("actions.next")}</Button>
        </div>
      </div>
    </AppCard>
  );
});

function LogsDataTable({
  columns,
  data,
  emptyLabel,
}: {
  columns: ColumnDef<RequestLog>[];
  data: RequestLog[];
  emptyLabel: string;
}) {
  // TanStack Table intentionally returns table helpers that React Compiler cannot memoize safely.
  // eslint-disable-next-line react-hooks/incompatible-library
  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
  });

  return (
    <Table className="text-xs">
      <TableHeader>
        {table.getHeaderGroups().map((headerGroup) => (
          <TableRow key={headerGroup.id}>
            {headerGroup.headers.map((header) => (
              <TableHead key={header.id} style={{ width: header.getSize() }}>
                {header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}
              </TableHead>
            ))}
          </TableRow>
        ))}
      </TableHeader>
      <TableBody>
        {table.getRowModel().rows.length ? (
          table.getRowModel().rows.map((row) => (
            <TableRow key={row.id} data-state={row.getIsSelected() && "selected"}>
              {row.getVisibleCells().map((cell) => (
                <TableCell key={cell.id}>
                  {flexRender(cell.column.columnDef.cell, cell.getContext())}
                </TableCell>
              ))}
            </TableRow>
          ))
        ) : (
          <TableRow>
            <TableCell colSpan={columns.length} className="py-8 text-center text-muted-foreground">
              {emptyLabel}
            </TableCell>
          </TableRow>
        )}
      </TableBody>
    </Table>
  );
}

const SavedViewsDialogs = memo(function SavedViewsDialogs({
  saveViewOpen,
  savedViewsOpen,
  viewName,
  savedViews,
  deletingViewID,
  setSaveViewOpen,
  setSavedViewsOpen,
  setViewName,
  onSaveCurrentView,
  onApplySavedView,
  onDeleteSavedView,
  t,
}: {
  saveViewOpen: boolean;
  savedViewsOpen: boolean;
  viewName: string;
  savedViews: SavedView[];
  deletingViewID: string;
  setSaveViewOpen: Dispatch<SetStateAction<boolean>>;
  setSavedViewsOpen: Dispatch<SetStateAction<boolean>>;
  setViewName: Dispatch<SetStateAction<string>>;
  onSaveCurrentView: () => Promise<void>;
  onApplySavedView: (view: SavedView) => void;
  onDeleteSavedView: (viewID: string) => Promise<void>;
  t: ReturnType<typeof useTranslations>;
}) {
  return (
    <>
      <Dialog open={saveViewOpen} onOpenChange={setSaveViewOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("savedViews.saveTitle")}</DialogTitle>
            <DialogDescription>{t("savedViews.saveDescription")}</DialogDescription>
          </DialogHeader>
          <div className="grid gap-2">
            <Label htmlFor="logs-save-view-name">{t("savedViews.nameLabel")}</Label>
            <Input id="logs-save-view-name" value={viewName} onChange={(event) => setViewName(event.target.value)} placeholder={t("savedViews.namePlaceholder")} />
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setSaveViewOpen(false)}>
              {t("actions.cancel")}
            </Button>
            <Button type="button" onClick={() => void onSaveCurrentView()} disabled={!viewName.trim()}>
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
                    <Button type="button" size="sm" variant="outline" onClick={() => onApplySavedView(view)}>
                      <FolderOpen className="h-4 w-4" />
                      {t("actions.load")}
                    </Button>
                    <Button type="button" size="sm" variant="destructive" onClick={() => void onDeleteSavedView(view.saved_view_id)} disabled={deletingViewID === view.saved_view_id}>
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
    </>
  );
});

const LogDetailSheet = memo(function LogDetailSheet({
  selectedLog,
  detail,
  traceAvailable,
  replayDraft,
  onClose,
  onCreateReplayDraft,
  formatters,
  t,
}: {
  selectedLog: RequestLog | null;
  detail: LogDetail | null;
  traceAvailable: boolean;
  replayDraft: ReplayDraft | null;
  onClose: () => void;
  onCreateReplayDraft: () => Promise<void>;
  formatters: LogsFormatter;
  t: ReturnType<typeof useTranslations>;
}) {
  return (
    <Sheet open={selectedLog !== null} onOpenChange={(open) => { if (!open) onClose(); }}>
      <SheetContent side="left" className="w-full overflow-y-auto sm:max-w-2xl">
        <SheetHeader>
          <SheetTitle>{t("detailTitle")}</SheetTitle>
          <SheetDescription>{selectedLog?.log_id}</SheetDescription>
        </SheetHeader>
        {detail ? (
          <div className="flex flex-col gap-4 pt-4">
            <div className="grid grid-cols-3 gap-2">
              <MetricCard label={t("table.status")} value={detail.log.status_code} />
              <MetricCard label={t("table.latency")} value={formatters.ms(detail.log.duration_ms)} />
              <MetricCard label={t("table.cost")} value={formatters.usd(detail.log.cost_usd ?? 0)} />
            </div>
            <div className="flex flex-wrap justify-end gap-2">
              {traceAvailable && detail.log.trace_id ? <Button asChild size="sm" variant="outline"><Link href={`/traces/${detail.log.trace_id}`}>{t("actions.openTrace")}</Link></Button> : null}
              <Button size="sm" variant="outline" onClick={() => void onCreateReplayDraft()}>
                <FileJson className="h-4 w-4" />
                {t("actions.replayDraft")}
              </Button>
            </div>
            {replayDraft ? <ReplayDraftPanel draft={replayDraft} t={t} /> : null}
            <Tabs defaultValue="summary">
              <TabsList className="ml-auto">
                <TabsTrigger value="metadata">{t("tabs.metadata")}</TabsTrigger>
                <TabsTrigger value="response">{t("content.response")}</TabsTrigger>
                <TabsTrigger value="request">{t("content.request")}</TabsTrigger>
                <TabsTrigger value="summary">{t("tabs.summary")}</TabsTrigger>
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
  );
});

function LabeledControl({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="flex min-w-0 flex-col gap-1">
      <Label className="text-xs text-muted-foreground">{label}</Label>
      {children}
    </div>
  );
}

const TextFilter = memo(function TextFilter({ label, value, onChange, mono }: { label: string; value: string; onChange: (value: string) => void; mono?: boolean }) {
  const [draft, setDraft] = useState(value);
  const onChangeRef = useRef(onChange);

  useEffect(() => {
    onChangeRef.current = onChange;
  }, [onChange]);

  useEffect(() => {
    let disposed = false;
    queueMicrotask(() => {
      if (!disposed) setDraft(value);
    });
    return () => {
      disposed = true;
    };
  }, [value]);

  useEffect(() => {
    if (draft === value) return;
    const timer = window.setTimeout(() => {
      onChangeRef.current(draft);
    }, 350);
    return () => window.clearTimeout(timer);
  }, [draft, value]);

  return (
    <LabeledControl label={label}>
      <Input
        value={draft}
        onChange={(event) => setDraft(event.target.value)}
        dir={mono ? "ltr" : undefined}
        className={cn("h-8 text-xs", mono && "text-left font-mono")}
      />
    </LabeledControl>
  );
});

const CustomDateRange = memo(function CustomDateRange({
  from,
  to,
  fromLabel,
  toLabel,
  onFromChange,
  onToChange,
  className,
}: {
  from: string;
  to: string;
  fromLabel: string;
  toLabel: string;
  onFromChange: (value: string) => void;
  onToChange: (value: string) => void;
  className?: string;
}) {
  return (
    <div className={cn("grid min-w-0 gap-3 md:grid-cols-2", className)}>
      <LabeledControl label={fromLabel}>
        <Input
          type="datetime-local"
          value={toDateTimeLocalValue(from)}
          onChange={(event) => onFromChange(fromDateTimeLocalValue(event.target.value))}
          className="h-8 text-left font-mono text-xs"
          dir="ltr"
        />
      </LabeledControl>
      <LabeledControl label={toLabel}>
        <Input
          type="datetime-local"
          value={toDateTimeLocalValue(to)}
          onChange={(event) => onToChange(fromDateTimeLocalValue(event.target.value))}
          className="h-8 text-left font-mono text-xs"
          dir="ltr"
        />
      </LabeledControl>
    </div>
  );
});

const RangeFilter = memo(function RangeFilter({
  label,
  minLabel,
  maxLabel,
  range,
  minValue,
  maxValue,
  onChange,
  formatValue,
}: {
  label: string;
  minLabel: string;
  maxLabel: string;
  range?: LogRange;
  minValue: string;
  maxValue: string;
  onChange: (min: string, max: string) => void;
  formatValue: (value: number) => string;
}) {
  const bounds = useMemo(() => normalizeRange(range), [range]);
  const disabled = bounds.max <= bounds.min;
  const value = useMemo(
    () => normalizeRangeValue(minValue, maxValue, bounds),
    [bounds, maxValue, minValue],
  );
  const valueKey = `${bounds.min}:${bounds.max}:${minValue}:${maxValue}`;
  const [draftState, setDraftState] = useState<{ key: string; value: [number, number] }>(() => ({ key: valueKey, value }));
  const draft = draftState.key === valueKey ? draftState.value : value;
  const sliderMax = disabled ? bounds.min + 1 : bounds.max;
  const sliderStep = disabled ? 1 : rangeStep(bounds);
  const hasRangeData = bounds.min !== 0 || bounds.max !== 0;

  return (
    <LabeledControl label={label}>
      <div className="flex min-h-20 flex-col justify-center gap-3 rounded-md border border-border bg-transparent px-3 py-2">
        <div className="flex items-center justify-between gap-2 text-xs text-muted-foreground" dir="ltr">
          <span title={minLabel}>{hasRangeData ? formatValue(draft[0]) : "-"}</span>
          <span title={maxLabel}>{hasRangeData ? formatValue(draft[1]) : "-"}</span>
        </div>
        <Slider
          value={draft}
          min={bounds.min}
          max={sliderMax}
          step={sliderStep}
          disabled={disabled}
          onValueChange={(next) => setDraftState({ key: valueKey, value: normalizeRangeValue(String(next[0] ?? bounds.min), String(next[1] ?? bounds.max), bounds) })}
          onValueCommit={(next) => {
            const committed = normalizeRangeValue(String(next[0] ?? bounds.min), String(next[1] ?? bounds.max), bounds);
            setDraftState({ key: valueKey, value: committed });
            onChange(
              committed[0] <= bounds.min ? "" : formatFilterNumber(committed[0]),
              committed[1] >= bounds.max ? "" : formatFilterNumber(committed[1]),
            );
          }}
        />
        <div className="flex items-center justify-between gap-2 text-[11px] text-muted-foreground" dir="ltr">
          <span>{hasRangeData ? formatValue(bounds.min) : "-"}</span>
          <span>{hasRangeData ? formatValue(bounds.max) : "-"}</span>
        </div>
      </div>
    </LabeledControl>
  );
});

const FacetSelect = memo(function FacetSelect({ label, value, onChange, items, empty }: { label: string; value: string; onChange: (value: string) => void; items?: Array<{ value: string; count: number }>; empty: string }) {
  const options = [{ value: "", label: label }, ...(items ?? []).filter((item) => item.value).map((item) => ({ value: item.value, label: `${item.value} (${item.count})` }))];
  return (
    <LabeledControl label={label}>
      <FilterCombobox id={`logs-${label}`} value={value} onValueChange={onChange} options={options} placeholder={label} emptyText={empty} />
    </LabeledControl>
  );
});

function StatusPill({ code }: { code: number }) {
  if (code >= 500) return <StatusBadge status="error">{code}</StatusBadge>;
  if (code >= 400) return <StatusBadge status="warning">{code}</StatusBadge>;
  return <StatusBadge status="success">{code || "-"}</StatusBadge>;
}

function RoutingValue({ value, description }: { value: string; description: string }) {
  const normalized = value.trim().toLowerCase();
  const muted = !normalized || normalized === "inactive";
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Badge variant="outline" className="max-w-28 cursor-help truncate text-[11px]">
          <span dir="auto">{muted ? "-" : value}</span>
        </Badge>
      </TooltipTrigger>
      <TooltipContent side="top" className="max-w-64 text-right leading-5" dir="rtl">
        {description}
      </TooltipContent>
    </Tooltip>
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
          {COLUMN_CHOICES.map((column) => (
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
    const Icon = icon;
    return (
      <AppCard contentClassName="p-3 sm:p-4 xl:pt-6">
        <div className="flex min-w-0 items-start justify-between gap-2 sm:gap-3">
          <div className="min-w-0">
            <p className="truncate text-xs font-medium text-muted-foreground sm:text-sm">{label}</p>
            <p className="mt-1 truncate text-lg font-bold tracking-tight text-foreground sm:mt-2 sm:text-2xl xl:text-3xl">{value}</p>
            {description ? <p className="mt-0.5 truncate text-[11px] leading-5 text-muted-foreground sm:mt-1 sm:text-xs">{description}</p> : null}
          </div>
          {Icon ? (
            <div className="flex size-8 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary sm:size-10 sm:rounded-xl">
              <Icon aria-hidden="true" className="h-4 w-4 sm:h-5 sm:w-5" />
            </div>
          ) : null}
        </div>
      </AppCard>
    );
  }

  return (
    <div className="rounded-lg border border-dashed border-border bg-muted/40 p-3 opacity-70 sm:p-4">
      <div className="text-xs font-medium text-muted-foreground">{label}</div>
      <div className="pt-1 text-base font-semibold text-muted-foreground sm:pt-2 sm:text-lg">{t("analyticsUnavailableShort")}</div>
      <div className="pt-1 text-[11px] text-muted-foreground sm:text-xs">{t("analyticsUnavailable")}</div>
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

function ReplayDraftPanel({ draft, t }: { draft: ReplayDraft; t: ReturnType<typeof useTranslations> }) {
  const payload = {
    log_id: draft.log_id,
    route: draft.route,
    request_type: draft.request_type,
    model: draft.model,
    provider: draft.provider,
    safe_headers: draft.safe_headers ?? {},
    request_payload: draft.request_payload ?? {},
  };

  return (
    <div className="flex flex-col gap-3">
      <Alert variant={draft.supported ? "default" : "destructive"}>
        <AlertTitle>{draft.supported ? t("replay.ready") : t("replay.unsupported")}</AlertTitle>
        <AlertDescription>{replayDraftMessage(draft, t)}</AlertDescription>
      </Alert>
      {draft.supported ? (
        <div className="flex flex-col gap-2">
          <div className="text-sm font-medium">{t("replay.payloadTitle")}</div>
          <JSONBlock value={payload} />
        </div>
      ) : null}
    </div>
  );
}

function JSONBlock({ value }: { value: unknown }) {
  return (
    <pre className="max-h-[520px] overflow-auto rounded-md border border-border bg-muted p-3 text-left font-mono text-xs leading-relaxed" dir="ltr">
      <code>{highlightJSON(value ?? {})}</code>
    </pre>
  );
}

function highlightJSON(value: unknown) {
  const json = JSON.stringify(value, null, 2);
  const tokenPattern = /("(?:\\u[\da-fA-F]{4}|\\[^u]|[^\\"])*"(\s*:)?|\btrue\b|\bfalse\b|\bnull\b|-?\d+(?:\.\d*)?(?:[eE][+-]?\d+)?)/g;

  const nodes: ReactNode[] = [];
  let lastIndex = 0;

  for (const match of json.matchAll(tokenPattern)) {
    const token = match[0];
    const index = match.index ?? 0;
    if (index > lastIndex) nodes.push(json.slice(lastIndex, index));

    const className = jsonTokenClassName(token);
    nodes.push(
      <span key={`${index}-${token}`} className={className}>
        {token}
      </span>,
    );
    lastIndex = index + token.length;
  }

  if (lastIndex < json.length) nodes.push(json.slice(lastIndex));
  return nodes;
}

function jsonTokenClassName(token: string) {
  if (token.endsWith(":")) return "text-primary";
  if (token.startsWith("\"")) return "text-foreground";
  if (token === "true" || token === "false") return "text-primary";
  if (token === "null") return "text-muted-foreground";
  return "text-destructive";
}

function replayDraftMessage(draft: ReplayDraft, t: ReturnType<typeof useTranslations>) {
  if (!draft.supported) {
    if (draft.reason === "unsupported_route") return t("replay.reasons.unsupportedRoute");
    if (draft.reason === "payload_hidden_or_unavailable") return t("replay.reasons.payloadUnavailable");
    return draft.reason || t("replay.reasons.unknown");
  }

  return t("replay.readyDescription");
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

async function fetchCachedJSON<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, init);
  const payload = await response.json().catch(() => ({}));
  if (!response.ok) {
    const message = payload?.error?.message ?? payload?.message ?? `Request failed (${response.status})`;
    throw new Error(message);
  }
  return payload as T;
}

function isAbortError(err: unknown) {
  return err instanceof Error && err.name === "AbortError";
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

function buildRangeStatsQuery(filters: Filters) {
  const params = new URLSearchParams();
  const serialized = serializeFilters(filters);
  for (const [key, value] of Object.entries(serialized)) {
    if (
      key === "duration_min" ||
      key === "duration_max" ||
      key === "tokens_min" ||
      key === "tokens_max" ||
      key === "cost_usd_min" ||
      key === "cost_usd_max" ||
      key === "cost_irr_min" ||
      key === "cost_irr_max"
    ) {
      continue;
    }
    if (value !== "") params.set(key, String(value));
  }
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

function normalizeVisibleColumns(columns?: string[]) {
  if (!columns?.length) return DEFAULT_COLUMNS;
  const expanded = columns.flatMap((column) => (column === "routing" ? OPTIONAL_COLUMNS : [column]));
  return Array.from(new Set(expanded)).filter((column) => COLUMN_CHOICES.includes(column));
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
    actor_type: t("filters.actor_type"),
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

function normalizeRange(range?: LogRange): { min: number; max: number } {
  const min = Number(range?.min ?? 0);
  const max = Number(range?.max ?? 0);
  if (!Number.isFinite(min) || !Number.isFinite(max)) return { min: 0, max: 0 };
  return { min: Math.min(min, max), max: Math.max(min, max) };
}

function normalizeRangeValue(minValue: string, maxValue: string, bounds: { min: number; max: number }): [number, number] {
  const min = parseFiniteNumber(minValue) ?? bounds.min;
  const max = parseFiniteNumber(maxValue) ?? bounds.max;
  const low = clampNumber(Math.min(min, max), bounds.min, bounds.max);
  const high = clampNumber(Math.max(min, max), bounds.min, bounds.max);
  return [low, high];
}

function toDateTimeLocalValue(value: string) {
  if (!value) return "";
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

function parseFiniteNumber(value: string): number | undefined {
  if (!value.trim()) return undefined;
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : undefined;
}

function clampNumber(value: number, min: number, max: number) {
  return Math.min(Math.max(value, min), max);
}

function rangeStep(bounds: { min: number; max: number }) {
  const span = bounds.max - bounds.min;
  if (span <= 1) return Math.max(span / 100, 0.0001);
  return Math.max(1, Math.round(span / 100));
}

function formatFilterNumber(value: number) {
  if (Number.isInteger(value)) return String(value);
  return value.toFixed(6).replace(/0+$/, "").replace(/\.$/, "");
}

function principalsFromFilters(filters: Filters, namesByKey: Record<string, string>): PrincipalPickerItem[] {
  const items: PrincipalPickerItem[] = [];
  const actorIDs = splitFilterList(filters.actor_id);
  const actorTypes = splitFilterList(filters.actor_type);
  const actorType = actorTypes.length === 1 ? actorTypes[0] : "user";
  for (const id of actorIDs) {
    const type = namesByKey[`virtual_account:${id}`] || actorType === "virtual_user" ? "virtual_account" : "user";
    const key = `${type}:${id}`;
    items.push({ type, principal_id: id, display_name: namesByKey[key] ?? id });
  }
  for (const id of splitFilterList(filters.team_id)) {
    const key = `team:${id}`;
    items.push({ type: "team", principal_id: id, display_name: namesByKey[key] ?? id });
  }
  for (const id of splitFilterList(filters.org_unit_id)) {
    const key = `org_unit:${id}`;
    items.push({ type: "org_unit", principal_id: id, display_name: namesByKey[key] ?? id });
  }
  return items;
}

function principalKey(item: PrincipalPickerItem) {
  if (!item.principal_id) return "";
  return `${item.type}:${item.principal_id}`;
}

function splitFilterList(value: string) {
  return value.split(",").map((item) => item.trim()).filter(Boolean);
}

function joinUnique(values: string[]) {
  return Array.from(new Set(values.map((item) => item.trim()).filter(Boolean))).join(",");
}

function filterLookupPrincipals(
  items: Array<{ id?: string; org_unit_id?: string; virtual_user_id?: string; alias?: string; display_name?: string; name?: string }>,
  query: string,
  type: PrincipalSearchKind,
): PrincipalSearchOption[] {
  const normalized = query.toLowerCase();
  return items
    .filter((item) => `${item.display_name ?? ""} ${item.name ?? ""} ${item.alias ?? ""}`.toLowerCase().includes(normalized))
    .slice(0, 20)
    .map((item) => {
      const principalID = item.id ?? item.org_unit_id ?? item.virtual_user_id ?? item.alias ?? "";
      const displayName = item.display_name || item.name || item.alias || principalID;
      return {
        principal_id: principalID,
        display_name: displayName,
        secondary: item.alias,
        type,
      };
    })
    .filter((item) => item.principal_id);
}

function createLogFormatters(locale: string): LogsFormatter {
  const number = new Intl.NumberFormat(locale);
  const percent = new Intl.NumberFormat(locale, { style: "percent", maximumFractionDigits: 1 });
  const usd = new Intl.NumberFormat(locale, { style: "currency", currency: "USD", maximumFractionDigits: 4 });
  const irr = new Intl.NumberFormat(locale, { maximumFractionDigits: 0 });
  const shortTime = new Intl.DateTimeFormat(locale, { hour: "2-digit", minute: "2-digit" });
  const weekday = new Intl.DateTimeFormat(locale, { weekday: "short" });
  const fullTime = new Intl.DateTimeFormat(locale, { hour: "2-digit", minute: "2-digit", second: "2-digit" });

  return {
    number: (value) => number.format(value),
    percent: (value) => percent.format(value),
    usd: (value) => usd.format(value),
    irr: (value) => `${irr.format(value)} IRR`,
    ms: (value) => `${number.format(Math.round(value))} ms`,
    msValue: (value) => number.format(Math.round(value)),
    timeLabel: (value) => (value ? shortTime.format(new Date(value)) : undefined),
    dateTime: (value) => {
      if (!value) return "-";
      const date = new Date(value);
      return `${weekday.format(date)}${locale.startsWith("fa") ? "،" : ","} ${fullTime.format(date)}`;
    },
  };
}

function formatTimeseries(points: AnalyticsPoint[], formatters: LogsFormatter) {
  return points.map((point) => ({
    label: formatters.timeLabel(point.timestamp) ?? point.group,
    requests: point.request_count,
    errors: point.error_count,
    latency: Math.round(point.latency_p95_ms ?? 0),
    cost: Number((point.cost_usd ?? 0).toFixed(4)),
  }));
}

function formatCacheTimeseries(points: AnalyticsPoint[], formatters: LogsFormatter) {
  return points.map((point) => {
    const cacheHit = point.cache_hit_count ?? 0;
    return {
      label: formatters.timeLabel(point.timestamp) ?? point.group,
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
