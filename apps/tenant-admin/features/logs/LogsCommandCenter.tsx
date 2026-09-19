"use client";

import dynamic from "next/dynamic";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { memo, useCallback, useEffect, useMemo, useRef, useState, type ComponentProps, type Dispatch, type SetStateAction } from "react";
import type { ColumnDef } from "@tanstack/react-table";
import {
  ChevronDown,
  Download,
  Eye,
  FolderOpen,
  RefreshCw,
  RotateCcw,
  Save,
  Settings2,
  Trash2,
} from "lucide-react";
import { useLocale, useTranslations } from "next-intl";

import { ActiveFilterChips } from "@/components/app-ui/ActiveFilterChips";
import { AppCard } from "@/components/app-ui/AppCard";
import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { AppTextField } from "@/components/app-ui/AppTextField";
import { CompactValue } from "@/components/app-ui/CompactValue";
import { DateTimeRangeField } from "@/components/app-ui/DateTimeRangeField";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { FilterCombobox } from "@/components/app-ui/FilterCombobox";
import { DebouncedTextFilter, FilterField, RangeFilterField } from "@/components/app-ui/FilterField";
import { FormDialog } from "@/components/app-ui/FormDialog";
import { PageHeader } from "@/components/app-ui/PageHeader";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { Skeleton } from "@/components/ui/skeleton";
import { Checkbox } from "@/components/ui/checkbox";

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { PrincipalPicker, type PrincipalPickerItem } from "@/components/app-ui/PrincipalPicker";
import { type PrincipalSearchKind, type PrincipalSearchOption } from "@/components/app-ui/PrincipalEntitySelect";
import { buildPrincipalSelectorLabels } from "@/components/app-ui/buildPrincipalSelectorLabels";
import RequestLoggingSettingsPanel from "@/features/logs/RequestLoggingSettingsPanel";
import { usePermissions } from "@/components/permissions-context";
import { canAnyPermission } from "@/lib/authz";
import { cn } from "@/lib/utils";
import {
  createLogExport,
  createSavedView,
  deleteSavedView as deleteSavedViewRequest,
  getLogAnalytics,
  getLogDetail,
  getLogFacets,
  getLogRangeStats,
  getLogs,
  getLogSummary,
  getOrganizationUnitLookups,
  getVirtualAccountLookups,
  parseInitialAnalytics,
  parseInitialFacets,
  parseInitialLogs,
  parseInitialRangeStats,
  parseInitialSavedViews,
  parseInitialSummary,
  searchAccessPrincipals,
  traceExists,
} from "@/features/logs/api";
import { LogsTable } from "@/features/logs/logs-table";
import { RangeSummary } from "@/features/logs/RangeSummary";
import {
  DEFAULT_COLUMNS,
  DEFAULT_FILTERS,
  rangeMs,
  type AnalyticsPoint,
  type ExpandedChartKey,
  type Filters,
  type LogDetail,
  type LogFacets,
  type LogRange,
  type LogRangeStats,
  type LogsCommandCenterProps,
  type LogsFormatter,
  type LogsResponse,
  type RequestLog,
  type SavedView,
  type Summary,
} from "@/features/logs/logs-types";
import {
  buildQuery,
  buildRangeStatsQuery,
  createLogFormatters,
  filtersFromSearchParams,
  formatCacheTimeseries,
  formatFilterNumber,
  formatTimeseries,
  isAbortError,
  limitChartPoints,
  normalizeRange,
  normalizeRangeValue,
  normalizeVisibleColumns,
  payloadStatus,
  rangeStep,
  serializeFilters,
} from "@/features/logs/logs-utils";

const LogDetailSheet = dynamic(
  () => import("@/features/logs/log-detail-sheet").then((module) => module.LogDetailSheet),
  {
    ssr: false,
    loading: () => null,
  },
);


const LogsCharts = dynamic(
  () => import("@/features/logs/logs-charts").then((module) => module.LogsCharts),
  { ssr: false },
);

const ExpandedChartDialog = dynamic(
  () => import("@/features/logs/logs-charts").then((module) => module.ExpandedChartDialog),
  { ssr: false },
);

export default function LogsCommandCenter({
  initialLogs,
  initialSummary,
  initialFacets,
  initialSavedViews,
  initialTimeseries,
  initialBreakdown,
  initialRangeStats,
  initialError,
}: LogsCommandCenterProps) {
  const t = useTranslations("logs");
  const common = useTranslations("common");
  const requestLoggingT = useTranslations("settings.requestLogging");
  const locale = useLocale();
  const pathname = usePathname();
  const router = useRouter();
  const searchParams = useSearchParams();
  const permissions = usePermissions();
  const [filters, setFilters] = useState<Filters>(() => filtersFromSearchParams(searchParams));
  const [limit, setLimit] = useState(25);
  const [offset, setOffset] = useState(0);
  const [logs, setLogs] = useState<LogsResponse>(() => parseInitialLogs(initialLogs));

  useEffect(() => {
    const nextParams = new URLSearchParams(searchParams.toString());
    for (const key of ["provider", "model"] as const) {
      const value = filters[key].trim();
      if (value) nextParams.set(key, value);
      else nextParams.delete(key);
    }
    const nextQuery = nextParams.toString();
    if (nextQuery === searchParams.toString()) return;
    router.replace(nextQuery ? `${pathname}?${nextQuery}` : pathname, { scroll: false });
  }, [filters, pathname, router, searchParams]);
  const [summary, setSummary] = useState<Summary>(() => parseInitialSummary(initialSummary));
  const [facets, setFacets] = useState<LogFacets>(() => parseInitialFacets(initialFacets));
  const [rangeStats, setRangeStats] = useState<LogRangeStats>(() => parseInitialRangeStats(initialRangeStats));
  const [savedViews, setSavedViews] = useState<SavedView[]>(() => parseInitialSavedViews(initialSavedViews));
  const [timeseries, setTimeseries] = useState<AnalyticsPoint[]>(() => limitChartPoints(parseInitialAnalytics(initialTimeseries)));
  const [breakdown, setBreakdown] = useState<AnalyticsPoint[]>(() => parseInitialAnalytics(initialBreakdown));
  const [selectedLog, setSelectedLog] = useState<RequestLog | null>(null);
  const [detail, setDetail] = useState<LogDetail | null>(null);
  const [traceAvailable, setTraceAvailable] = useState(false);
  const [viewName, setViewName] = useState("");
  const [saveViewOpen, setSaveViewOpen] = useState(false);
  const [savedViewsOpen, setSavedViewsOpen] = useState(false);
  const [expandedChart, setExpandedChart] = useState<ExpandedChartKey | null>(null);
  const [requestLoggingOpen, setRequestLoggingOpen] = useState(false);
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
  const rangeStatsAbortRef = useRef<AbortController | null>(null);
  const rangeStatsRequestRef = useRef(0);
  const detailAbortRef = useRef<AbortController | null>(null);
  const detailRequestRef = useRef(0);
  const hasCompleteInitialDataRef = useRef(
    initialLogs !== undefined
      && initialSummary !== undefined
      && initialFacets !== undefined
      && initialTimeseries !== undefined
      && initialBreakdown !== undefined
      && !initialError,
  );
  const hasInitialRangeStatsRef = useRef(initialRangeStats !== undefined);

  const formatters = useMemo(() => createLogFormatters(locale), [locale]);
  const timeseriesChartData = useMemo(() => formatTimeseries(timeseries, formatters), [timeseries, formatters]);
  const cacheTimeseriesChartData = useMemo(() => formatCacheTimeseries(timeseries, formatters), [timeseries, formatters]);
  const providerBreakdownData = useMemo(
    () => breakdown.map((item) => ({ name: item.group || t("values.unknown"), requests: item.request_count })),
    [breakdown, t],
  );
  const canManageRequestLogging = useMemo(() => canAnyPermission(permissions ?? [], ["settings:list"]), [permissions]);
  const principalT = useTranslations("providerAccounts");
  const principalUx = useTranslations("providerAccounts.ux");
  const principalSelectorLabels = useMemo(() => buildPrincipalSelectorLabels(principalT, principalUx), [principalT, principalUx]);

  const loadData = useCallback(async () => {
    loadAbortRef.current?.abort();
    const controller = new AbortController();
    loadAbortRef.current = controller;
    const requestID = ++loadRequestRef.current;
    const isCurrentRequest = () => loadRequestRef.current === requestID && loadAbortRef.current === controller && !controller.signal.aborted;
    const nextQueryString = buildQuery(filters, limit, offset);
    const nextBaseQueryString = buildQuery(filters, undefined, undefined);

    setLoading(true);
    setError("");
    setSummaryUnavailable(false);
    setTimeseriesUnavailable(false);
    setBreakdownUnavailable(false);
    try {
      const [logsResult, summaryResult, facetsResult, timeseriesResult, breakdownResult] = await Promise.allSettled([
        getLogs(nextQueryString, controller.signal),
        getLogSummary(nextBaseQueryString, controller.signal),
        getLogFacets(nextBaseQueryString, controller.signal),
        getLogAnalytics(`type=timeseries&group_by=&interval_seconds=3600&${nextBaseQueryString}`, controller.signal),
        getLogAnalytics(`type=breakdown&group_by=provider&${nextBaseQueryString}`, controller.signal),
      ]);

      if (!isCurrentRequest()) return;
      if (logsResult.status === "rejected") throw logsResult.reason;
      setLogs(logsResult.value);
      if (summaryResult.status === "fulfilled") setSummary(summaryResult.value.data ?? {});
      else setSummaryUnavailable(true);
      if (facetsResult.status === "fulfilled") setFacets(facetsResult.value.data ?? {});
      if (timeseriesResult.status === "fulfilled") setTimeseries(limitChartPoints(timeseriesResult.value.data ?? []));
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
  }, [filters, limit, offset, t]);

  const loadRangeStats = useCallback(async () => {
    rangeStatsAbortRef.current?.abort();
    const controller = new AbortController();
    rangeStatsAbortRef.current = controller;
    const requestID = ++rangeStatsRequestRef.current;
    const isCurrentRequest = () => rangeStatsRequestRef.current === requestID && rangeStatsAbortRef.current === controller && !controller.signal.aborted;
    const nextRangeStatsQueryString = buildRangeStatsQuery(filters);
    try {
      const payload = await getLogRangeStats(nextRangeStatsQueryString, controller.signal);
      if (!isCurrentRequest()) return;
      setRangeStats(payload.data ?? {});
    } catch (err) {
      if (!isAbortError(err) && isCurrentRequest()) setRangeStats({});
    } finally {
      if (rangeStatsAbortRef.current === controller) rangeStatsAbortRef.current = null;
    }
  }, [filters]);

  const refreshReport = useCallback(async () => {
    await Promise.allSettled([loadData(), loadRangeStats()]);
  }, [loadData, loadRangeStats]);

  useEffect(() => {
    if (hasCompleteInitialDataRef.current) {
      hasCompleteInitialDataRef.current = false;
      return;
    }

    let disposed = false;
    queueMicrotask(() => {
      if (!disposed) void loadData();
    });
    return () => {
      disposed = true;
      loadAbortRef.current?.abort();
      rangeStatsAbortRef.current?.abort();
    };
  }, [loadData]);

  useEffect(() => {
    if (hasInitialRangeStatsRef.current) {
      hasInitialRangeStatsRef.current = false;
      return;
    }

    let disposed = false;
    queueMicrotask(() => {
      if (!disposed) void loadRangeStats();
    });
    return () => {
      disposed = true;
    };
  }, [loadRangeStats]);

  useEffect(() => {
    if (!canManageRequestLogging) return;
    if (searchParams.get("settings") !== "request-logging") return;

    queueMicrotask(() => {
      setRequestLoggingOpen(true);
    });

    const nextParams = new URLSearchParams(searchParams.toString());
    nextParams.delete("settings");
    const nextHref = nextParams.size > 0 ? `${pathname}?${nextParams.toString()}` : pathname;
    router.replace(nextHref, { scroll: false });
  }, [canManageRequestLogging, pathname, router, searchParams]);

  const openDetail = useCallback(async (item: RequestLog) => {
    detailAbortRef.current?.abort();
    const controller = new AbortController();
    detailAbortRef.current = controller;
    const requestID = ++detailRequestRef.current;
    const isCurrentRequest = () => detailRequestRef.current === requestID && detailAbortRef.current === controller && !controller.signal.aborted;

    setSelectedLog(item);
    setDetail(null);
    setTraceAvailable(false);
    try {
      const payload = await getLogDetail(item.log_id, controller.signal);
      if (!isCurrentRequest()) return;
      setDetail(payload);

      const traceID = payload.log.trace_id?.trim();
      if (traceID) {
        try {
          await traceExists(traceID, controller.signal);
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
  }, []);

  const rows = logs.data ?? [];
  const total = Math.max(logs.total ?? 0, summary.request_count ?? 0, offset + rows.length);
  const pageStart = total === 0 ? 0 : offset + 1;
  const pageEnd = Math.min(offset + rows.length, total);
  const hasNextPage = typeof logs.next_offset === "number" ? logs.next_offset > offset : offset + rows.length < total;

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
        cell: ({ row }) => <CompactValue display={formatters.numberDisplay(row.original.total_tokens ?? 0)} className="whitespace-nowrap" />,
      },
      {
        id: "cost",
        header: t("table.cost"),
        cell: ({ row }) => (
          <div className="whitespace-nowrap">
            <div><CompactValue display={formatters.usdDisplay(row.original.cost_usd ?? 0)} /></div>
            <div className="text-muted-foreground"><CompactValue display={formatters.irrDisplay(row.original.cost_irr ?? 0)} /></div>
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
        cell: ({ row }) => <StatusBadge status={payloadStatus(row.original.payload_mode)} className="whitespace-nowrap">{row.original.payload_mode || "-"}</StatusBadge>,
      },
      {
        id: "actions",
        header: "",
        cell: ({ row }) => (
          <div className="flex items-center gap-1">
            <Button variant="ghost" size="icon" onClick={() => void openDetail(row.original)} aria-label={t("actions.inspect")} title={t("actions.inspect")}>
              <Eye aria-hidden="true" />
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
      const payload = await searchAccessPrincipals(kind, normalized);
      return (payload.data ?? []).map((item) => ({
        principal_id: item.principal_id,
        display_name: item.display_name,
        secondary: item.secondary,
        type: kind,
      }));
    }
    if (kind === "org_unit") {
      const payload = await getOrganizationUnitLookups();
      return filterLookupPrincipals(payload.data ?? [], normalized, "org_unit");
    }
    if (kind === "virtual_account") {
      const payload = await getVirtualAccountLookups();
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
    setOffset((current) => {
      if (typeof logs.next_offset === "number" && logs.next_offset > current) {
        return logs.next_offset;
      }
      return current + limit;
    });
  }, [limit, logs.next_offset]);

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
      const created = await createSavedView({
        name: viewName.trim(),
        filters: serializeFilters(filters),
        visible_columns: visibleColumns,
        shared_scope: "tenant",
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
      await deleteSavedViewRequest(viewID);
      setSavedViews((items) => items.filter((item) => item.saved_view_id !== viewID));
    } catch (err) {
      setError(err instanceof Error ? err.message : t("savedViews.deleteError"));
    } finally {
      setDeletingViewID("");
    }
  }, [t]);

  const createExport = useCallback(async (format: "csv" | "jsonl") => {
    const job = await createLogExport({
      filters: {
        ...serializeFilters(filters),
        limit: String(limit),
        offset: String(offset),
        sort_by: filters.sort_by,
        sort_order: filters.sort_order,
      },
      format,
      include_content: includeContentExport,
    });
    window.location.href = `/api/logs/exports/${encodeURIComponent(job.export_id)}/download`;
  }, [filters, includeContentExport, limit, offset]);

  return (
    <TooltipProvider delayDuration={120}>
    <div className="flex flex-col gap-4">
      <PageHeader
        eyebrow={t("eyebrow")}
        title={t("title")}
        description={t("subtitle")}
        actions={
          <div className="flex flex-wrap items-center gap-2">
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="outline"
                  size="icon-sm"
                  onClick={() => void refreshReport()}
                  disabled={loading}
                  aria-busy={loading}
                  aria-label={t("actions.refresh")}
                  title={t("actions.refresh")}
                >
                  {loading ? <Spinner /> : <RefreshCw />}
                  <span className="sr-only">{t("actions.refresh")}</span>
                </Button>
              </TooltipTrigger>
              <TooltipContent>{t("actions.refresh")}</TooltipContent>
            </Tooltip>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  variant="outline"
                  size="icon-sm"
                  aria-label={t("actions.export")}
                  title={t("actions.export")}
                >
                  <Download className="h-4 w-4" />
                  <span className="sr-only">{t("actions.export")}</span>
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
            {canManageRequestLogging ? (
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant="outline"
                    size="icon-sm"
                    onClick={() => setRequestLoggingOpen(true)}
                    aria-label={t("actions.settings")}
                    title={t("actions.settings")}
                  >
                    <Settings2 className="h-4 w-4" />
                    <span className="sr-only">{t("actions.settings")}</span>
                  </Button>
                </TooltipTrigger>
                <TooltipContent>{t("actions.settings")}</TooltipContent>
              </Tooltip>
            ) : null}
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
      <RangeSummary summary={summary} summaryUnavailable={summaryUnavailable} formatters={formatters} t={t}>
        <DeferredLogsCharts
          timeseriesUnavailable={timeseriesUnavailable}
          breakdownUnavailable={breakdownUnavailable}
          timeseriesChartData={timeseriesChartData}
          cacheTimeseriesChartData={cacheTimeseriesChartData}
          providerBreakdownData={providerBreakdownData}
          onExpandChart={setExpandedChart}
          t={t}
        />
      </RangeSummary>
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
        hasNextPage={hasNextPage}
        onPageLimitChange={setPageLimit}
        t={t}
      />
      <ExpandedChartDialog
        chart={expandedChart}
        timeseriesUnavailable={timeseriesUnavailable}
        breakdownUnavailable={breakdownUnavailable}
        timeseriesChartData={timeseriesChartData}
        cacheTimeseriesChartData={cacheTimeseriesChartData}
        providerBreakdownData={providerBreakdownData}
        onOpenChange={(open) => {
          if (!open) setExpandedChart(null);
        }}
        formatters={formatters}
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
      {canManageRequestLogging ? (
        <Sheet open={requestLoggingOpen} onOpenChange={setRequestLoggingOpen}>
          <SheetContent side="left" closeLabel={common("cancel")} className="w-full overflow-y-auto sm:max-w-xl">
            <SheetHeader>
              <SheetTitle>{requestLoggingT("title")}</SheetTitle>
              <SheetDescription>{requestLoggingT("description")}</SheetDescription>
            </SheetHeader>
            <div className="pt-4">
              <RequestLoggingSettingsPanel
                initialSettings={null}
                active={requestLoggingOpen}
                showHeader={false}
                surface="plain"
              />
            </div>
          </SheetContent>
        </Sheet>
      ) : null}
      <LogDetailSheet
        selectedLog={selectedLog}
        detail={detail}
        traceAvailable={traceAvailable}
        onClose={closeDetail}
        formatters={formatters}
        closeLabel={common("cancel")}
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
  const showAdvancedLabel = t("filters.showAdvanced");
  const hideAdvancedLabel = t("filters.hideAdvanced");
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
    <AppCard>
      <Collapsible open={advancedOpen} onOpenChange={setAdvancedOpen} className="flex flex-col gap-3">
        <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
          <FilterField label={t("filters.range")}>
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
          </FilterField>
          <FacetSelect label={t("filters.provider")} value={filters.provider} onChange={(value) => onSetFilter("provider", value)} items={facets.providers} empty={t("filters.noOptions")} />
          <FacetSelect label={t("filters.model")} value={filters.model} onChange={(value) => onSetFilter("model", value)} items={facets.models} empty={t("filters.noOptions")} />
          {filters.range === "custom" ? (
            <DateTimeRangeField
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
      <ActiveFilterChips items={activeChips} onRemove={(key) => onSetFilter(key as keyof Filters, "")} getRemoveLabel={(item) => `${t("actions.removeFilter")}: ${item.label}`} className="pt-4" />
    </AppCard>
  );
});

function DeferredLogsCharts(props: ComponentProps<typeof LogsCharts>) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const [shouldLoad, setShouldLoad] = useState(false);

  useEffect(() => {
    if (shouldLoad) return;
    const container = containerRef.current;
    if (!container) return;

    if (!("IntersectionObserver" in window)) {
      queueMicrotask(() => setShouldLoad(true));
      return;
    }

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (!entry?.isIntersecting) return;
        setShouldLoad(true);
        observer.disconnect();
      },
      { rootMargin: "500px 0px" },
    );
    observer.observe(container);
    return () => observer.disconnect();
  }, [shouldLoad]);

  return (
    <div ref={containerRef} className="flex flex-col gap-4">
      {shouldLoad ? (
        <LogsCharts {...props} />
      ) : (
        <div className="grid gap-4 xl:grid-cols-2" aria-hidden="true">
          {Array.from({ length: 4 }).map((_, index) => (
            <Skeleton key={`logs-chart-${index}`} className="h-80 rounded-xl" />
          ))}
        </div>
      )}
    </div>
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
      <FormDialog open={saveViewOpen} onOpenChange={setSaveViewOpen} title={t("savedViews.saveTitle")} description={t("savedViews.saveDescription")} footer={(
          <>
            <Button type="button" variant="outline" onClick={() => setSaveViewOpen(false)}>
              {t("actions.cancel")}
            </Button>
            <Button type="button" onClick={() => void onSaveCurrentView()} disabled={!viewName.trim()}>
              <Save data-icon="inline-start" />
              {t("actions.saveView")}
            </Button>
          </>
      )}>
        <AppTextField id="logs-save-view-name" label={t("savedViews.nameLabel")} value={viewName} onChange={(event) => setViewName(event.target.value)} placeholder={t("savedViews.namePlaceholder")} />
      </FormDialog>

      <FormDialog open={savedViewsOpen} onOpenChange={setSavedViewsOpen} title={t("savedViews.manageTitle")} description={t("savedViews.manageDescription")} size="md">
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
      </FormDialog>
    </>
  );
});


const TextFilter = memo(function TextFilter({ label, value, onChange, mono }: { label: string; value: string; onChange: (value: string) => void; mono?: boolean }) {
  return <DebouncedTextFilter label={label} value={value} onValueChange={onChange} mono={mono} />;
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
      <RangeFilterField label={label} value={draft} min={bounds.min} max={sliderMax} step={sliderStep} disabled={disabled} minLabel={minLabel} maxLabel={maxLabel} formatValue={(next) => hasRangeData ? formatValue(next) : "-"}
          onValueChange={(next) => setDraftState({ key: valueKey, value: normalizeRangeValue(String(next[0]), String(next[1]), bounds) })}
          onValueCommit={(next) => {
            const committed = normalizeRangeValue(String(next[0]), String(next[1]), bounds);
            setDraftState({ key: valueKey, value: committed });
            onChange(
              committed[0] <= bounds.min ? "" : formatFilterNumber(committed[0]),
              committed[1] >= bounds.max ? "" : formatFilterNumber(committed[1]),
            );
          }}
        />
  );
});

const FacetSelect = memo(function FacetSelect({ label, value, onChange, items, empty }: { label: string; value: string; onChange: (value: string) => void; items?: Array<{ value: string; count: number }>; empty: string }) {
  const options = [{ value: "", label: label }, ...(items ?? []).filter((item) => item.value).map((item) => ({ value: item.value, label: `${item.value} (${item.count})` }))];
  return (
    <FilterField label={label}>
      <FilterCombobox id={`logs-${label}`} value={value} onValueChange={onChange} options={options} placeholder={label} emptyText={empty} />
    </FilterField>
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
        <Badge variant="outline" className="max-w-28 cursor-help truncate text-xs">
          <span dir="auto">{muted ? "-" : value}</span>
        </Badge>
      </TooltipTrigger>
      <TooltipContent side="top" className="max-w-64 text-start leading-5">
        {description}
      </TooltipContent>
    </Tooltip>
  );
}


function MonoText({ value }: { value?: string }) {
  return <span className="block max-w-40 truncate font-mono text-xs" dir="auto">{value || "-"}</span>;
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
