"use client";

import Link from "next/link";
import { useCallback, useEffect, useMemo, useState, type ReactNode } from "react";
import type { ColumnDef } from "@tanstack/react-table";
import { ChevronDown, Columns3, Filter, RefreshCw, RotateCcw, Settings2 } from "lucide-react";
import { useLocale, useTranslations } from "next-intl";

import TracePrivacySettingsPanel, { type TracePrivacySettings } from "@/app/components/trace-privacy-settings";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { ActiveFilterChips, AppSelectField, CatalogFilterCombobox, CompactValue, DateTimeRangeField, FilterField, MetricCard, PageContainer } from "@/components/app-ui";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import { DataTable } from "@/components/ui/data-table";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Sheet, SheetContent, SheetDescription, SheetFooter, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { useIsMobile } from "@/hooks/use-mobile";
import { formatCompactNumber } from "@/lib/compact-number";
import { cn } from "@/lib/utils";

type RequestTrace = {
  trace_id: string;
  tenant_id: string;
  request_id?: string;
  model?: string;
  provider?: string;
  route: string;
  status_code: number;
  started_at: string;
  duration_ms: number;
  error_code?: string;
  sampled: boolean;
  attempts?: number;
};

type TraceListResponse = {
  data?: RequestTrace[];
  total?: number;
  limit?: number;
  offset?: number;
  next_offset?: number;
};

type FacetBucket = {
  value: string;
  count: number;
};

type StatusFacetBucket = {
  status_code: number;
  count: number;
};

type TraceFacets = {
  providers?: FacetBucket[];
  models?: FacetBucket[];
  routes?: FacetBucket[];
  error_codes?: FacetBucket[];
  status_codes?: StatusFacetBucket[];
};

type FacetsResponse = {
  data?: TraceFacets;
};

type TraceFilters = {
  traceID: string;
  requestID: string;
  model: string;
  provider: string;
  route: string;
  statusCode: string;
  errorCode: string;
  from: string;
  to: string;
  sortBy: string;
  sortOrder: "asc" | "desc";
};

type RangePreset = (typeof RANGE_PRESETS)[number];
type TraceColumnKey = (typeof TRACE_COLUMN_KEYS)[number];

type TraceManagementProps = {
  initialData: RequestTrace[];
  initialTotal: number;
  initialError: string | null;
  initialTracePrivacySettings: TracePrivacySettings | null;
  initialTracePrivacyError: string | null;
  initialTracePrivacyOpen?: boolean;
  currentUserID: string;
  currentUserLabel: string;
};

type BreakdownBucket = {
  value: string;
  label: string;
  count: number;
};

const PAGE_SIZE = 25;
const MAX_BREAKDOWN_ITEMS = 5;
const RANGE_PRESETS = ["last15m", "last1h", "last24h", "yesterday", "last3d", "last7d", "last14d", "last30d", "custom"] as const;
const DEFAULT_RANGE_PRESET = "last24h";
const DEFAULT_RANGE_MS = 24 * 60 * 60 * 1000;
const ALL_MODELS_VALUE = "__all_models__";
const ALL_PROVIDERS_VALUE = "__all_providers__";
const TRACE_COLUMN_KEYS = ["startedAt", "status", "duration", "trace", "context", "endpoint", "error"] as const;

const BASE_FILTERS: Omit<TraceFilters, "from" | "to"> = {
  traceID: "",
  requestID: "",
  model: "",
  provider: "",
  route: "",
  statusCode: "",
  errorCode: "",
  sortBy: "started_at",
  sortOrder: "desc",
};

export default function TraceManagement(props: TraceManagementProps) {
  const locale = useLocale();
  const t = useTranslations("traces");
  const tracePrivacyT = useTranslations("settings.tracePrivacy");
  const common = useTranslations("common");
  const isCompactLayout = useIsMobile();

  const [items, setItems] = useState<RequestTrace[]>(props.initialData);
  const [total, setTotal] = useState(props.initialTotal);
  const [offset, setOffset] = useState(0);
  const [error, setError] = useState<string | null>(props.initialError);
  const [loading, setLoading] = useState(false);
  const [facets, setFacets] = useState<TraceFacets>({});

  const [draft, setDraft] = useState<TraceFilters>(() => createDefaultFilters());
  const [applied, setApplied] = useState<TraceFilters>(() => createDefaultFilters());
  const [draftRangePreset, setDraftRangePreset] = useState<RangePreset>(DEFAULT_RANGE_PRESET);
  const [appliedRangePreset, setAppliedRangePreset] = useState<RangePreset>(DEFAULT_RANGE_PRESET);
  const [advancedOpen, setAdvancedOpen] = useState(false);
  const [advancedSheetOpen, setAdvancedSheetOpen] = useState(false);
  const [tracePrivacyOpen, setTracePrivacyOpen] = useState(props.initialTracePrivacyOpen ?? false);
  const [visibleColumns, setVisibleColumns] = useState<Record<TraceColumnKey, boolean>>(() => createDefaultColumnVisibility());

  const canPrev = offset > 0;
  const canNext = offset + items.length < total;
  const showingFrom = total === 0 ? 0 : offset + 1;
  const showingTo = Math.min(offset + items.length, total);

  const statusBuckets = useMemo(() => {
    const source = facets.status_codes?.length
      ? facets.status_codes
      : aggregateStatusBuckets(items);
    return source
      .map((item) => ({
        value: String(item.status_code),
        label: String(item.status_code),
        count: item.count,
      }))
      .sort(sortBuckets);
  }, [facets.status_codes, items]);

  const breakdowns = useMemo(
    () => ({
      providers: normalizeFacetBuckets(facets.providers, t("facets.empty")),
      models: normalizeFacetBuckets(facets.models, t("facets.empty")),
      routes: normalizeFacetBuckets(facets.routes, t("facets.empty")),
      errorCodes: normalizeFacetBuckets(facets.error_codes, t("facets.empty")),
      statusCodes: statusBuckets,
    }),
    [facets.error_codes, facets.models, facets.providers, facets.routes, statusBuckets, t],
  );

  const rangeOptions = useMemo(() => RANGE_PRESETS.map((value) => ({ value, label: t(`filters.rangePresets.${value}`) })), [t]);
  const modelComboboxOptions = useMemo(
    () => [
      { value: ALL_MODELS_VALUE, label: t("filters.placeholders.allModels") },
      ...normalizeComboboxOptions(facets.models),
    ],
    [facets.models, t],
  );
  const providerComboboxOptions = useMemo(
    () => [
      { value: ALL_PROVIDERS_VALUE, label: t("filters.placeholders.allProviders") },
      ...normalizeComboboxOptions(facets.providers),
    ],
    [facets.providers, t],
  );

  const statusTotal = statusBuckets.reduce((sum, item) => sum + item.count, 0);
  const errorCount = statusBuckets.reduce((sum, item) => {
    const code = Number(item.value);
    return code >= 400 ? sum + item.count : sum;
  }, 0);
  const errorRate = statusTotal > 0 ? errorCount / statusTotal : 0;
  const maxDuration = Math.max(0, ...items.map((item) => item.duration_ms));

  const filtersQueryString = useMemo(() => {
    const params = new URLSearchParams();
    if (applied.traceID) params.set("trace_id", applied.traceID);
    if (applied.requestID) params.set("request_id", applied.requestID);
    if (applied.model) params.set("model", applied.model);
    if (applied.provider) params.set("provider", applied.provider);
    if (applied.route) params.set("route", applied.route);
    if (applied.statusCode) params.set("status_code", applied.statusCode);
    if (applied.errorCode) params.set("error_code", applied.errorCode);
    if (applied.from) params.set("from", applied.from);
    if (applied.to) params.set("to", applied.to);
    if (applied.sortBy) params.set("sort_by", applied.sortBy);
    if (applied.sortOrder) params.set("sort_order", applied.sortOrder);
    return params.toString();
  }, [applied]);

  const listQueryString = useMemo(() => {
    const params = new URLSearchParams(filtersQueryString);
    params.set("limit", String(PAGE_SIZE));
    params.set("offset", String(offset));
    return params.toString();
  }, [filtersQueryString, offset]);

  const activeFilters = useMemo(() => {
    const out: Array<{ key: keyof TraceFilters | "timeRange"; label: string; value: string }> = [];
    if (applied.traceID) out.push({ key: "traceID", label: t("activeFilters.traceId"), value: compactID(applied.traceID) });
    if (applied.requestID) out.push({ key: "requestID", label: t("activeFilters.requestId"), value: compactID(applied.requestID) });
    if (applied.model) out.push({ key: "model", label: t("activeFilters.model"), value: applied.model });
    if (applied.provider) out.push({ key: "provider", label: t("activeFilters.provider"), value: applied.provider });
    if (applied.route) out.push({ key: "route", label: t("activeFilters.route"), value: applied.route });
    if (applied.statusCode) out.push({ key: "statusCode", label: t("activeFilters.status"), value: applied.statusCode });
    if (applied.errorCode) out.push({ key: "errorCode", label: t("activeFilters.errorCode"), value: applied.errorCode });
    if (appliedRangePreset === "custom") {
      if (applied.from) out.push({ key: "from", label: t("activeFilters.from"), value: formatLocalDateTime(applied.from, locale) });
      if (applied.to) out.push({ key: "to", label: t("activeFilters.to"), value: formatLocalDateTime(applied.to, locale) });
    } else {
      out.push({ key: "timeRange", label: t("activeFilters.timeRange"), value: t(`filters.rangePresets.${appliedRangePreset}`) });
    }
    return out;
  }, [applied, appliedRangePreset, locale, t]);

  const advancedFilterLabel = advancedOpen ? t("filters.advanced.hide") : t("filters.advanced.show");

  useEffect(() => {
    void load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [listQueryString]);

  async function load() {
    setLoading(true);
    try {
      const [listPayload, facetsPayload] = await Promise.all([
        requestJSON<TraceListResponse>(`/api/traces?${listQueryString}`, common("requestError")),
        requestJSON<FacetsResponse>(`/api/traces/facets${filtersQueryString ? `?${filtersQueryString}` : ""}`, common("requestError")),
      ]);
      setItems(listPayload.data ?? []);
      setTotal(listPayload.total ?? 0);
      setFacets(facetsPayload.data ?? {});
      setError(null);
    } catch (err) {
      setError(normalizeError(err, common("requestError")));
    } finally {
      setLoading(false);
    }
  }

  function applyFilters() {
    const resolvedRange = resolveSelectedRange(draftRangePreset, draft.from, draft.to);
    const next = normalizeFilters({
      ...draft,
      from: resolvedRange.from,
      to: resolvedRange.to,
    });
    setDraft(next);
    setApplied(next);
    setAppliedRangePreset(draftRangePreset);
    setOffset(0);
  }

  function clearFilters() {
    const next = createDefaultFilters();
    setDraft(next);
    setApplied(next);
    setDraftRangePreset(DEFAULT_RANGE_PRESET);
    setAppliedRangePreset(DEFAULT_RANGE_PRESET);
    setOffset(0);
  }

  function removeFilter(key: keyof TraceFilters | "timeRange") {
    const emptyFilters = createEmptyFilters();
    if (key === "timeRange") {
      const nextDraft = { ...draft, from: "", to: "" };
      const nextApplied = { ...applied, from: "", to: "" };
      setDraft(nextDraft);
      setApplied(nextApplied);
      setDraftRangePreset("custom");
      setAppliedRangePreset("custom");
      setOffset(0);
      return;
    }
    const nextDraft = { ...draft, [key]: emptyFilters[key] };
    const nextApplied = { ...applied, [key]: emptyFilters[key] };
    if (key === "from" || key === "to") {
      setDraftRangePreset("custom");
      setAppliedRangePreset("custom");
    }
    setDraft(nextDraft);
    setApplied(nextApplied);
    setOffset(0);
  }

  function selectFacet(update: Partial<TraceFilters>) {
    const nextDraft = normalizeFilters({ ...draft, ...update });
    const nextApplied = normalizeFilters({ ...applied, ...update });
    setDraft(nextDraft);
    setApplied(nextApplied);
    setOffset(0);
  }

  function handleRangePresetChange(nextPreset: RangePreset) {
    setDraftRangePreset(nextPreset);
    if (nextPreset === "custom") return;
    const nextRange = rangeForPreset(nextPreset);
    setDraft((prev) => ({ ...prev, from: nextRange.from, to: nextRange.to }));
  }

  const setSort = useCallback((sortBy: string, forcedOrder?: "asc" | "desc") => {
    const nextOrder: "asc" | "desc" =
      forcedOrder ?? (applied.sortBy === sortBy && applied.sortOrder === "desc" ? "asc" : "desc");
    const next = normalizeFilters({ ...applied, sortBy, sortOrder: nextOrder });
    setDraft(next);
    setApplied(next);
    setOffset(0);
  }, [applied]);

  const traceColumnLabels = useMemo<Record<TraceColumnKey, string>>(
    () => ({
      startedAt: t("list.columns.startedAt"),
      status: t("list.columns.status"),
      duration: t("list.columns.duration"),
      trace: t("list.columns.trace"),
      context: t("list.columns.context"),
      endpoint: t("list.columns.route"),
      error: t("list.columns.error"),
    }),
    [t],
  );
  const visibleColumnCount = TRACE_COLUMN_KEYS.filter((key) => visibleColumns[key]).length;
  const compactSortOptions = useMemo(
    () => [
      { value: "started_at", label: t("list.columns.startedAt") },
      { value: "status_code", label: t("list.columns.status") },
      { value: "duration_ms", label: t("list.columns.duration") },
    ],
    [t],
  );
  const breakdownCards = [
    {
      key: "statusCodes",
      title: t("breakdown.statusCodes"),
      buckets: breakdowns.statusCodes,
      activeValue: applied.statusCode,
      onSelect: (value: string) => selectFacet({ statusCode: value }),
      mono: true,
    },
    {
      key: "providers",
      title: t("breakdown.providers"),
      buckets: breakdowns.providers,
      activeValue: applied.provider,
      onSelect: (value: string) => selectFacet({ provider: value }),
      mono: false,
    },
    {
      key: "models",
      title: t("breakdown.models"),
      buckets: breakdowns.models,
      activeValue: applied.model,
      onSelect: (value: string) => selectFacet({ model: value }),
      mono: true,
    },
    {
      key: "routes",
      title: t("breakdown.routes"),
      buckets: breakdowns.routes,
      activeValue: applied.route,
      onSelect: (value: string) => selectFacet({ route: value }),
      mono: true,
    },
    {
      key: "errorCodes",
      title: t("breakdown.errorCodes"),
      buckets: breakdowns.errorCodes,
      activeValue: applied.errorCode,
      onSelect: (value: string) => selectFacet({ errorCode: value }),
      mono: true,
    },
  ];
  const advancedFiltersFields = (
    <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-5">
      <FilterField htmlFor="status-filter" label={t("filters.statusCode")}>
        <Input
          id="status-filter"
          value={draft.statusCode}
          onChange={(event) => setDraft((prev) => ({ ...prev, statusCode: event.target.value }))}
          placeholder={t("filters.placeholders.statusCode")}
          dir="ltr"
          className="h-9 font-mono"
        />
      </FilterField>
      <FilterField htmlFor="error-filter" label={t("filters.errorCode")}>
        <Input
          id="error-filter"
          value={draft.errorCode}
          onChange={(event) => setDraft((prev) => ({ ...prev, errorCode: event.target.value }))}
          placeholder={t("filters.placeholders.errorCode")}
          dir="ltr"
          className="h-9 font-mono"
        />
      </FilterField>
      <FilterField htmlFor="route-filter" label={t("filters.route")}>
        <Input
          id="route-filter"
          value={draft.route}
          onChange={(event) => setDraft((prev) => ({ ...prev, route: event.target.value }))}
          placeholder={t("filters.placeholders.route")}
          dir="ltr"
          className="h-9 font-mono"
        />
      </FilterField>
      <FilterField htmlFor="trace-id-filter" label={t("filters.traceId")}>
        <Input
          id="trace-id-filter"
          value={draft.traceID}
          onChange={(event) => setDraft((prev) => ({ ...prev, traceID: event.target.value }))}
          placeholder={t("filters.placeholders.traceId")}
          dir="ltr"
          className="h-9 font-mono"
        />
      </FilterField>
      <FilterField htmlFor="request-id-filter" label={t("filters.requestId")}>
        <Input
          id="request-id-filter"
          value={draft.requestID}
          onChange={(event) => setDraft((prev) => ({ ...prev, requestID: event.target.value }))}
          placeholder={t("filters.placeholders.requestId")}
          dir="ltr"
          className="h-9 font-mono"
        />
      </FilterField>
    </div>
  );

  function toggleColumnVisibility(key: TraceColumnKey, checked: boolean) {
    setVisibleColumns((prev) => {
      if (!checked && Object.values(prev).filter(Boolean).length <= 1) return prev;
      return { ...prev, [key]: checked };
    });
  }

  const traceColumnConfigs = useMemo<Array<{ key: TraceColumnKey; column: ColumnDef<RequestTrace> }>>(
    () => [
      {
        key: "startedAt",
        column: {
          accessorKey: "started_at",
          size: 160,
          header: () => (
            <SortButton
              label={t("list.columns.startedAt")}
              isActive={applied.sortBy === "started_at"}
              sortOrder={applied.sortOrder}
              onClick={() => setSort("started_at")}
              sortLabel={t("list.sort.label", { column: t("list.columns.startedAt") })}
            />
          ),
          cell: ({ row }) => (
            <span className="whitespace-nowrap" title={formatIsoDateTime(row.original.started_at)}>
              {formatLocalDateTime(row.original.started_at, locale)}
            </span>
          ),
        },
      },
      {
        key: "status",
        column: {
          accessorKey: "status_code",
          size: 96,
          header: () => (
            <SortButton
              label={t("list.columns.status")}
              isActive={applied.sortBy === "status_code"}
              sortOrder={applied.sortOrder}
              onClick={() => setSort("status_code")}
              sortLabel={t("list.sort.label", { column: t("list.columns.status") })}
            />
          ),
          cell: ({ row }) => {
            const item = row.original;
            return (
              <div className="flex flex-wrap items-center gap-1">
                <Badge variant={statusBadgeVariant(item.status_code)}>{item.status_code}</Badge>
                {item.attempts && item.attempts > 1 ? (
                  <Badge variant="outline" title={t("list.meta.attempts", { count: item.attempts })}>
                    {t("list.meta.attemptsShort", { count: item.attempts })}
                  </Badge>
                ) : null}
              </div>
            );
          },
        },
      },
      {
        key: "duration",
        column: {
          accessorKey: "duration_ms",
          size: 160,
          header: () => (
            <SortButton
              label={t("list.columns.duration")}
              isActive={applied.sortBy === "duration_ms"}
              sortOrder={applied.sortOrder}
              onClick={() => setSort("duration_ms")}
              sortLabel={t("list.sort.label", { column: t("list.columns.duration") })}
            />
          ),
          cell: ({ row }) => (
            <div className="min-w-40">
              <DurationMeter
                value={row.original.duration_ms}
                max={maxDuration}
                msLabel={t("list.units.ms")}
                sLabel={t("list.units.s")}
                locale={locale}
              />
            </div>
          ),
        },
      },
      {
        key: "trace",
        column: {
          id: "trace",
          size: 224,
          header: () => <span className="block min-w-56">{t("list.columns.trace")}</span>,
          cell: ({ row }) => {
            const item = row.original;
            return (
              <div className="flex min-w-0 max-w-72 flex-col gap-0.5">
                <Link
                  className="truncate font-mono text-xs text-primary underline-offset-2 hover:underline"
                  href={`/traces/${encodeURIComponent(item.trace_id)}`}
                  title={item.trace_id}
                  aria-label={t("list.actions.openTrace", { traceId: item.trace_id })}
                >
                  {compactID(item.trace_id)}
                </Link>
                {item.request_id ? (
                  <span className="truncate font-mono text-[11px] text-muted-foreground" title={item.request_id}>
                    {t("list.meta.requestId", { id: compactID(item.request_id) })}
                  </span>
                ) : null}
              </div>
            );
          },
        },
      },
      {
        key: "context",
        column: {
          id: "context",
          size: 224,
          header: () => <span className="block min-w-56">{t("list.columns.context")}</span>,
          cell: ({ row }) => {
            const item = row.original;
            return (
              <div className="flex min-w-0 max-w-72 flex-col gap-0.5">
                <span className="truncate" dir="auto" title={item.provider || common("notSet")}>
                  {item.provider || common("notSet")}
                </span>
                <span className="truncate text-muted-foreground" dir="auto" title={item.model || common("notSet")}>
                  {item.model || common("notSet")}
                </span>
              </div>
            );
          },
        },
      },
      {
        key: "endpoint",
        column: {
          accessorKey: "route",
          size: 192,
          header: () => <span className="block min-w-48">{t("list.columns.route")}</span>,
          cell: ({ row }) => (
            <span className="block max-w-64 truncate font-mono text-[11px]" dir="ltr" title={row.original.route}>
              {row.original.route}
            </span>
          ),
        },
      },
      {
        key: "error",
        column: {
          accessorKey: "error_code",
          size: 144,
          header: () => <span className="block min-w-36">{t("list.columns.error")}</span>,
          cell: ({ row }) =>
            row.original.error_code ? (
              <Badge variant="outline" className="max-w-48">
                <span className="truncate font-mono text-[11px]" dir="auto" title={row.original.error_code}>
                  {row.original.error_code}
                </span>
              </Badge>
            ) : (
              <span className="text-muted-foreground">{common("none")}</span>
            ),
        },
      },
    ],
    [applied.sortBy, applied.sortOrder, common, locale, maxDuration, setSort, t],
  );
  const traceColumns = useMemo(
    () => traceColumnConfigs.filter((item) => visibleColumns[item.key]).map((item) => item.column),
    [traceColumnConfigs, visibleColumns],
  );

  return (
    <PageContainer spacing="compact" padding="none" className="md:gap-6">
      <div className="flex justify-end gap-2">
            <Tooltip>
              <TooltipTrigger asChild>
                <Button type="button" variant="outline" size="icon" onClick={() => void load()} disabled={loading} aria-label={loading ? common("loading") : common("refresh")}>
                  <RefreshCw aria-hidden="true" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>{loading ? common("loading") : common("refresh")}</TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button type="button" variant="outline" size="icon" onClick={() => setTracePrivacyOpen(true)} aria-label={t("privacySettings")}>
                  <Settings2 aria-hidden="true" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>{t("privacySettings")}</TooltipContent>
            </Tooltip>
      </div>

      <Card className="rounded-lg">
        <CardContent className="flex flex-col gap-3 px-3 pb-3 pt-2 md:px-4 md:pb-4 md:pt-3">
          <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
            <AppSelectField
              label={t("filters.timeRange")}
              value={draftRangePreset}
              onValueChange={(value) => handleRangePresetChange(value as RangePreset)}
              options={rangeOptions}
              ariaLabel={t("filters.timeRange")}
              triggerClassName="h-9"
            />
            <FilterField htmlFor="provider-filter" label={t("filters.provider")}>
              <CatalogFilterCombobox
                id="provider-filter"
                options={providerComboboxOptions}
                value={draft.provider || ALL_PROVIDERS_VALUE}
                onValueChange={(value) => setDraft((prev) => ({ ...prev, provider: value === ALL_PROVIDERS_VALUE ? "" : value }))}
                placeholder={t("filters.search")}
                emptyText={t("filters.emptyOptions")}
              />
            </FilterField>
            <FilterField htmlFor="model-filter" label={t("filters.model")}>
              <CatalogFilterCombobox
                id="model-filter"
                options={modelComboboxOptions}
                value={draft.model || ALL_MODELS_VALUE}
                onValueChange={(value) => setDraft((prev) => ({ ...prev, model: value === ALL_MODELS_VALUE ? "" : value }))}
                placeholder={t("filters.search")}
                emptyText={t("filters.emptyOptions")}
              />
            </FilterField>
          </div>

          {draftRangePreset === "custom" ? (
            <DateTimeRangeField fromID="from-filter" toID="to-filter" from={draft.from} to={draft.to} fromLabel={t("filters.from")} toLabel={t("filters.to")} onFromChange={(from) => setDraft((prev) => ({ ...prev, from }))} onToChange={(to) => setDraft((prev) => ({ ...prev, to }))} />
          ) : null}

          {isCompactLayout ? (
            <>
              <div className="grid gap-2 border-t border-card-border/70 pt-3 sm:grid-cols-2">
                <Button type="button" variant="outline" size="sm" className="sm:col-span-2" onClick={() => setAdvancedSheetOpen(true)}>
                  <ChevronDown data-icon="inline-start" aria-hidden="true" />
                  {advancedFilterLabel}
                </Button>
                <Button type="button" size="sm" onClick={applyFilters} disabled={loading}>
                  <Filter data-icon="inline-start" aria-hidden="true" />
                  {t("actions.apply")}
                </Button>
                <Button type="button" variant="outline" size="sm" onClick={clearFilters} disabled={loading}>
                  <RotateCcw data-icon="inline-start" aria-hidden="true" />
                  {t("actions.clear")}
                </Button>
              </div>

              <Sheet open={advancedSheetOpen} onOpenChange={setAdvancedSheetOpen}>
                <SheetContent side="bottom" className="flex max-h-svh flex-col rounded-t-2xl px-4 py-5 sm:max-w-none">
                  <SheetHeader>
                    <SheetTitle>{advancedFilterLabel}</SheetTitle>
                    <SheetDescription>{t("description")}</SheetDescription>
                  </SheetHeader>
                  <ScrollArea className="mt-4 flex-1">
                    <div className="pb-4">
                      {advancedFiltersFields}
                    </div>
                  </ScrollArea>
                  <SheetFooter>
                    <Button type="button" onClick={() => { applyFilters(); setAdvancedSheetOpen(false); }} disabled={loading}>
                      <Filter data-icon="inline-start" aria-hidden="true" />
                      {t("actions.apply")}
                    </Button>
                    <Button type="button" variant="outline" onClick={clearFilters} disabled={loading}>
                      <RotateCcw data-icon="inline-start" aria-hidden="true" />
                      {t("actions.clear")}
                    </Button>
                  </SheetFooter>
                </SheetContent>
              </Sheet>
            </>
          ) : (
            <Collapsible open={advancedOpen} onOpenChange={setAdvancedOpen} className="flex flex-col gap-3">
              <div className="flex flex-wrap items-center justify-between gap-2 border-t border-card-border/70 pt-3">
                <CollapsibleTrigger asChild>
                  <Button type="button" variant="outline" size="sm" aria-expanded={advancedOpen}>
                    <ChevronDown data-icon="inline-start" className={cn("transition-transform", advancedOpen && "rotate-180")} aria-hidden="true" />
                    {advancedFilterLabel}
                  </Button>
                </CollapsibleTrigger>
                <div className="flex flex-wrap items-center gap-2">
                  <Button type="button" size="sm" onClick={applyFilters} disabled={loading}>
                    <Filter data-icon="inline-start" aria-hidden="true" />
                    {t("actions.apply")}
                  </Button>
                  <Button type="button" variant="outline" size="sm" onClick={clearFilters} disabled={loading}>
                    <RotateCcw data-icon="inline-start" aria-hidden="true" />
                    {t("actions.clear")}
                  </Button>
                </div>
              </div>
              <CollapsibleContent>
                {advancedFiltersFields}
              </CollapsibleContent>
            </Collapsible>
          )}

          <ActiveFilterChips items={activeFilters} onRemove={(key) => removeFilter(key as keyof TraceFilters | "timeRange")} getRemoveLabel={(item) => t("activeFilters.removeLabel", { label: item.label, value: item.value ?? "" })} className="pt-1" />
        </CardContent>
      </Card>

      {error ? (
        <Alert variant="destructive">
          <AlertTitle>{t("alerts.errorTitle")}</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}

      <section aria-label={t("summary.title")} className="grid grid-cols-2 gap-2">
        <MetricCard
          label={t("summary.totalRequests")}
          value={<CompactValue display={formatCompactNumber(total, { locale })} />}
          description={t("summary.currentFilters")}
          density="compact"
        />
        <MetricCard
          label={t("summary.errorRate")}
          value={formatPercent(errorRate, locale)}
          description={t("summary.failureCount", { count: formatCompactNumber(errorCount, { locale }).value })}
          density="compact"
        />
      </section>

      <section aria-label={t("breakdown.title")} className="flex flex-col gap-2">
        {isCompactLayout ? (
          <ScrollArea className="w-full">
            <div className="flex items-start gap-2 pb-3">
              {breakdownCards.map((item) => (
                <div key={item.key} className="min-w-72 max-w-72 shrink-0">
                  <BreakdownCard
                    title={item.title}
                    buckets={item.buckets}
                    activeValue={item.activeValue}
                    noDataLabel={t("facets.noData")}
                    selectLabel={(value, count) => t("breakdown.select", { label: item.title, value, count })}
                    onSelect={item.onSelect}
                    mono={item.mono}
                    compact
                  />
                </div>
              ))}
            </div>
          </ScrollArea>
        ) : (
          <div className="grid gap-2 lg:grid-cols-2 xl:grid-cols-5">
            {breakdownCards.map((item) => (
              <BreakdownCard
                key={item.key}
                title={item.title}
                buckets={item.buckets}
                activeValue={item.activeValue}
                noDataLabel={t("facets.noData")}
                selectLabel={(value, count) => t("breakdown.select", { label: item.title, value, count })}
                onSelect={item.onSelect}
                mono={item.mono}
              />
            ))}
          </div>
        )}
      </section>

      <Card className="rounded-lg">
        <CardHeader className="flex flex-col gap-3 p-3 md:p-4 lg:flex-row lg:items-start lg:justify-between">
          <div className="min-w-0">
            <CardTitle className="text-sm md:text-base">{t("list.title")}</CardTitle>
            <CardDescription className="mt-1 text-xs">
              {t("pagination.showing", { from: showingFrom, to: showingTo, total })}
            </CardDescription>
          </div>
          {isCompactLayout ? (
            <div className="grid w-full gap-2 sm:grid-cols-[minmax(0,1fr)_auto] lg:w-auto">
              <AppSelectField
                label={t("list.sort.field")}
                value={applied.sortBy}
                onValueChange={(value) => setSort(value, value === applied.sortBy ? applied.sortOrder : "desc")}
                options={compactSortOptions}
                ariaLabel={t("list.sort.field")}
                triggerClassName="h-9 min-w-0 w-full"
                className="min-w-0"
              />
              <Button
                type="button"
                variant="outline"
                size="sm"
                className="sm:self-end"
                aria-label={t("list.sort.directionLabel", {
                  direction: applied.sortOrder === "asc" ? t("list.sort.directionAsc") : t("list.sort.directionDesc"),
                })}
                onClick={() => setSort(applied.sortBy, applied.sortOrder === "desc" ? "asc" : "desc")}
              >
                <span aria-hidden="true">{applied.sortOrder === "asc" ? "↑" : "↓"}</span>
                {applied.sortOrder === "asc" ? t("list.sort.directionAsc") : t("list.sort.directionDesc")}
              </Button>
            </div>
          ) : (
            <TraceColumnVisibilityMenu
              labels={traceColumnLabels}
              visibleColumns={visibleColumns}
              visibleColumnCount={visibleColumnCount}
              columnsLabel={t("list.actions.columns")}
              menuLabel={t("list.actions.columnsMenu")}
              onToggle={toggleColumnVisibility}
            />
          )}
        </CardHeader>
        <CardContent className="px-3 pb-3 md:px-3 md:pb-3">
          {isCompactLayout ? (
            items.length > 0 ? (
              <div className="grid gap-3 md:grid-cols-2">
                {items.map((item) => (
                  <CompactTraceCard key={item.trace_id} item={item} locale={locale} maxDuration={maxDuration} t={t} common={common} />
                ))}
              </div>
            ) : (
              <div className="rounded-lg border border-dashed border-card-border px-4 py-10 text-center text-sm text-muted-foreground">
                {t("list.empty")}
              </div>
            )
          ) : (
            <DataTable columns={traceColumns} data={items} emptyLabel={t("list.empty")} />
          )}

          <div className="mt-3 flex flex-col gap-2 text-xs text-muted-foreground sm:flex-row sm:items-center sm:justify-end">
            <div className="flex gap-2">
              <Button size="sm" variant="outline" disabled={!canPrev || loading} onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}>
                {t("pagination.prev")}
              </Button>
              <Button size="sm" variant="outline" disabled={!canNext || loading} onClick={() => setOffset(offset + PAGE_SIZE)}>
                {t("pagination.next")}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      <Sheet open={tracePrivacyOpen} onOpenChange={setTracePrivacyOpen}>
        <SheetContent side="left" className="w-full overflow-y-auto sm:max-w-xl">
          <SheetHeader>
            <SheetTitle>{tracePrivacyT("title")}</SheetTitle>
          </SheetHeader>
          <div className="pt-4">
            <TracePrivacySettingsPanel
              initialSettings={props.initialTracePrivacySettings}
              initialError={props.initialTracePrivacyError}
              showHeader={false}
              surface="plain"
              currentUserID={props.currentUserID}
              currentUserLabel={props.currentUserLabel}
            />
          </div>
        </SheetContent>
      </Sheet>
    </PageContainer>
  );
}

function createEmptyFilters(): TraceFilters {
  return {
    ...BASE_FILTERS,
    from: "",
    to: "",
  };
}

function createDefaultFilters(): TraceFilters {
  const range = rangeForPreset(DEFAULT_RANGE_PRESET);
  return {
    ...BASE_FILTERS,
    from: range.from,
    to: range.to,
  };
}

function createDefaultColumnVisibility(): Record<TraceColumnKey, boolean> {
  return TRACE_COLUMN_KEYS.reduce(
    (acc, key) => {
      acc[key] = true;
      return acc;
    },
    {} as Record<TraceColumnKey, boolean>,
  );
}

function TraceColumnVisibilityMenu({
  labels,
  visibleColumns,
  visibleColumnCount,
  columnsLabel,
  menuLabel,
  onToggle,
}: {
  labels: Record<TraceColumnKey, string>;
  visibleColumns: Record<TraceColumnKey, boolean>;
  visibleColumnCount: number;
  columnsLabel: string;
  menuLabel: string;
  onToggle: (key: TraceColumnKey, checked: boolean) => void;
}) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button type="button" variant="outline" size="sm">
          <Columns3 data-icon="inline-start" aria-hidden="true" />
          {columnsLabel}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-48">
        <DropdownMenuLabel>{menuLabel}</DropdownMenuLabel>
        <DropdownMenuSeparator />
        {TRACE_COLUMN_KEYS.map((key) => {
          const checked = visibleColumns[key];
          return (
            <DropdownMenuCheckboxItem
              key={key}
              checked={checked}
              disabled={checked && visibleColumnCount <= 1}
              onCheckedChange={(nextChecked) => onToggle(key, Boolean(nextChecked))}
            >
              {labels[key]}
            </DropdownMenuCheckboxItem>
          );
        })}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

function BreakdownCard({
  title,
  buckets,
  activeValue,
  noDataLabel,
  selectLabel,
  onSelect,
  mono = false,
  compact = false,
}: {
  title: string;
  buckets: BreakdownBucket[];
  activeValue?: string;
  noDataLabel: string;
  selectLabel: (value: string, count: number) => string;
  onSelect: (value: string) => void;
  mono?: boolean;
  compact?: boolean;
}) {
  const locale = useLocale();
  const visibleBuckets = buckets.slice(0, MAX_BREAKDOWN_ITEMS);
  const maxCount = Math.max(0, ...visibleBuckets.map((item) => item.count));

  return (
    <Card className="rounded-lg">
      <CardHeader className="gap-1 p-3 pb-2 md:p-3 md:pb-2">
        <CardTitle className="truncate text-sm">{title}</CardTitle>
      </CardHeader>
      <CardContent className={cn("flex flex-col gap-1.5 px-3 pb-3 md:px-3 md:pb-3", compact && "max-h-40 overflow-y-auto pe-2")}>
        {visibleBuckets.length === 0 ? (
          <p className="py-4 text-center text-xs text-muted-foreground">{noDataLabel}</p>
        ) : (
          visibleBuckets.map((item) => {
            const active = activeValue === item.value;
            const width = maxCount > 0 ? Math.max(6, (item.count / maxCount) * 100) : 0;
            return (
              <button
                key={`${title}:${item.value}`}
                type="button"
                className={cn(
                  "grid gap-1 rounded-md border p-2 text-start transition-colors hover:bg-muted/50",
                  active && "border-primary bg-muted/50",
                )}
                title={selectLabel(item.label, item.count)}
                aria-label={selectLabel(item.label, item.count)}
                onClick={() => onSelect(item.value)}
              >
                <span className="flex min-w-0 items-center justify-between gap-2">
                  <span className={cn("min-w-0 truncate text-xs", mono && "font-mono")} dir="auto">
                    {item.label}
                  </span>
                  <CompactValue
                    display={formatCompactNumber(item.count, { locale })}
                    className="font-mono text-[11px] text-muted-foreground"
                    dir="ltr"
                  />
                </span>
                <span className="h-1.5 overflow-hidden rounded-full bg-muted">
                  <span
                    className={cn("block h-full rounded-full", active ? "bg-primary" : "bg-muted-foreground")}
                    style={{ width: `${width}%` }}
                  />
                </span>
              </button>
            );
          })
        )}
      </CardContent>
    </Card>
  );
}

function CompactTraceCard({
  item,
  locale,
  maxDuration,
  t,
  common,
}: {
  item: RequestTrace;
  locale: string;
  maxDuration: number;
  t: (key: string, values?: Record<string, string | number>) => string;
  common: (key: string, values?: Record<string, string | number>) => string;
}) {
  return (
    <Card className="rounded-lg">
      <CardHeader className="gap-1.5 p-3 pb-2">
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0">
            <CardDescription className="text-xs">{formatLocalDateTime(item.started_at, locale)}</CardDescription>
            <CardTitle className="mt-1 min-w-0 text-sm">
              <Link
                className="block truncate font-mono text-primary underline-offset-2 hover:underline"
                href={`/traces/${encodeURIComponent(item.trace_id)}`}
                title={item.trace_id}
                aria-label={t("list.actions.openTrace", { traceId: item.trace_id })}
              >
                {compactID(item.trace_id)}
              </Link>
            </CardTitle>
            {item.request_id ? (
              <p className="mt-1 truncate font-mono text-[11px] text-muted-foreground" title={item.request_id}>
                {t("list.meta.requestId", { id: compactID(item.request_id) })}
              </p>
            ) : null}
          </div>
          <div className="flex shrink-0 flex-wrap items-center justify-end gap-1">
            <Badge variant={statusBadgeVariant(item.status_code)}>{item.status_code}</Badge>
            {item.attempts && item.attempts > 1 ? (
              <Badge variant="outline">{t("list.meta.attemptsShort", { count: item.attempts })}</Badge>
            ) : null}
          </div>
        </div>
      </CardHeader>
      <CardContent className="flex flex-col gap-2 px-3 pb-3 pt-0">
        <div className="grid grid-cols-3 gap-x-3 gap-y-2">
          <CompactListField
            label={t("list.columns.model")}
            value={
              <span className="truncate" dir="auto" title={item.model || common("notSet")}>
                {item.model || common("notSet")}
              </span>
            }
          />
          <CompactListField
            label={t("list.columns.provider")}
            value={
              <span className="truncate" dir="auto" title={item.provider || common("notSet")}>
                {item.provider || common("notSet")}
              </span>
            }
          />
          <CompactListField
            label={t("list.columns.duration")}
            value={<DurationMeter value={item.duration_ms} max={Math.max(maxDuration, 1)} msLabel={t("list.units.ms")} sLabel={t("list.units.s")} locale={locale} compact />}
            inline
          />
        </div>
        <div className="grid grid-cols-2 gap-x-4 gap-y-2">
          <CompactListField
            label={t("list.columns.route")}
            inline
            value={
              <span className="truncate font-mono text-xs" dir="ltr" title={item.route}>
                {item.route}
              </span>
            }
          />
          <CompactListField
            label={t("list.columns.error")}
            inline
            value={item.error_code ? (
              <Badge variant="outline" className="max-w-full">
                <span className="truncate font-mono text-[11px]" dir="auto" title={item.error_code}>
                  {item.error_code}
                </span>
              </Badge>
            ) : (
              <span className="text-muted-foreground">{common("none")}</span>
            )}
          />
        </div>
      </CardContent>
    </Card>
  );
}

function CompactListField({
  label,
  value,
  secondary,
  secondaryLabel,
  className,
  inline = false,
}: {
  label: string;
  value: ReactNode;
  secondary?: ReactNode;
  secondaryLabel?: string;
  className?: string;
  inline?: boolean;
}) {
  return (
    <div className={cn("flex min-w-0 flex-col gap-0.5", inline && "justify-between gap-1", className)}>
      {inline ? (
        <div className="flex min-w-0 items-start justify-between gap-3">
          <span className="shrink-0 text-xs text-muted-foreground">{label}</span>
          <div className="min-w-0 text-sm font-medium text-end">{value}</div>
        </div>
      ) : (
        <>
          <span className="text-xs text-muted-foreground">{label}</span>
          <div className="min-w-0 text-sm font-medium">{value}</div>
        </>
      )}
      {secondary ? (
        <div className="flex min-w-0 flex-col gap-0.5">
          {secondaryLabel ? <span className="text-xs text-muted-foreground">{secondaryLabel}</span> : null}
          <div className="min-w-0 text-xs">{secondary}</div>
        </div>
      ) : null}
    </div>
  );
}

function SortButton({
  label,
  isActive,
  sortOrder,
  sortLabel,
  onClick,
}: {
  label: string;
  isActive: boolean;
  sortOrder: "asc" | "desc";
  sortLabel: string;
  onClick: () => void;
}) {
  return (
    <button type="button" className="inline-flex items-center gap-1 text-start hover:underline" aria-label={sortLabel} onClick={onClick}>
      {label}
      <span className="font-mono text-[10px] text-muted-foreground" aria-hidden="true">
        {isActive ? (sortOrder === "asc" ? "^" : "v") : ""}
      </span>
    </button>
  );
}

function DurationMeter({
  value,
  max,
  msLabel,
  sLabel,
  locale,
  compact = false,
}: {
  value: number;
  max: number;
  msLabel: string;
  sLabel: string;
  locale: string;
  compact?: boolean;
}) {
  const width = max > 0 ? Math.max(6, (value / max) * 100) : 0;

  if (compact) {
    return (
      <span className="font-mono text-xs tabular-nums">
        {formatDuration(value, msLabel, sLabel, locale)}
      </span>
    );
  }

  return (
    <div className="flex min-w-0 items-center gap-2">
      <span className="w-16 shrink-0 text-end font-mono tabular-nums">
        {formatDuration(value, msLabel, sLabel, locale)}
      </span>
      <span className="h-1.5 min-w-16 flex-1 overflow-hidden rounded-full bg-muted">
        <span
          className="block h-full rounded-full"
          style={{ width: `${width}%`, backgroundColor: durationColor(value) }}
        />
      </span>
    </div>
  );
}

async function requestJSON<T = unknown>(path: string, fallbackError: string, options?: RequestInit): Promise<T> {
  const headers = new Headers(options?.headers ?? {});
  if (options?.body && !headers.has("Content-Type")) headers.set("Content-Type", "application/json");
  const response = await fetch(path, { ...options, headers, cache: "no-store" });
  const text = await response.text();
  const payload = parsePayload(text);
  if (!response.ok) throw new Error(extractMessage(payload) ?? `${fallbackError} (${response.status})`);
  return payload as T;
}

function parsePayload(text: string): unknown {
  if (!text.trim()) return {};
  try {
    return JSON.parse(text) as unknown;
  } catch {
    return { message: text };
  }
}

function extractMessage(payload: unknown): string | null {
  if (typeof payload === "string") return payload;
  if (!payload || typeof payload !== "object") return null;
  const candidate = payload as { error?: { message?: string }; message?: string };
  return candidate.error?.message ?? candidate.message ?? null;
}

function normalizeError(error: unknown, fallback: string) {
  if (error instanceof Error) return error.message;
  return fallback;
}

function normalizeFilters(filters: TraceFilters): TraceFilters {
  return {
    ...filters,
    traceID: filters.traceID.trim(),
    requestID: filters.requestID.trim(),
    model: filters.model.trim(),
    provider: filters.provider.trim(),
    route: filters.route.trim(),
    statusCode: filters.statusCode.trim(),
    errorCode: filters.errorCode.trim(),
    from: filters.from.trim(),
    to: filters.to.trim(),
    sortBy: (filters.sortBy || "started_at").trim(),
    sortOrder: filters.sortOrder === "asc" ? "asc" : "desc",
  };
}

function normalizeFacetBuckets(buckets: FacetBucket[] | undefined, emptyLabel: string): BreakdownBucket[] {
  return (buckets ?? [])
    .map((item) => ({
      value: item.value || "",
      label: item.value || emptyLabel,
      count: item.count,
    }))
    .sort(sortBuckets);
}

function normalizeComboboxOptions(buckets: FacetBucket[] | undefined) {
  return Array.from(
    new Set(
      (buckets ?? [])
        .map((item) => item.value.trim())
        .filter(Boolean),
    ),
  )
    .sort((a, b) => a.localeCompare(b))
    .map((value) => ({ value, label: value }));
}

function aggregateStatusBuckets(items: RequestTrace[]): StatusFacetBucket[] {
  const map = new Map<number, number>();
  for (const item of items) {
    map.set(item.status_code, (map.get(item.status_code) ?? 0) + 1);
  }
  return Array.from(map.entries()).map(([status_code, count]) => ({ status_code, count }));
}

function sortBuckets(a: { count: number; value: string }, b: { count: number; value: string }) {
  if (b.count !== a.count) return b.count - a.count;
  return a.value.localeCompare(b.value);
}

function statusBadgeVariant(statusCode: number): "success" | "info" | "warning" | "destructive" | "outline" {
  if (statusCode >= 500) return "destructive";
  if (statusCode >= 400) return "warning";
  if (statusCode >= 300) return "info";
  if (statusCode >= 200) return "success";
  return "outline";
}

function compactID(value: string) {
  if (value.length <= 22) return value;
  return `${value.slice(0, 12)}...${value.slice(-6)}`;
}

function formatPercent(value: number, locale: string) {
  return new Intl.NumberFormat(locale, {
    style: "percent",
    maximumFractionDigits: 1,
  }).format(value);
}

function formatDuration(value: number, msLabel: string, sLabel: string, locale: string) {
  if (!Number.isFinite(value)) return `0 ${msLabel}`;
  if (value >= 1000) {
    return `${new Intl.NumberFormat(locale, { maximumFractionDigits: 1 }).format(value / 1000)} ${sLabel}`;
  }
  return `${new Intl.NumberFormat(locale, { maximumFractionDigits: 0 }).format(value)} ${msLabel}`;
}

function durationColor(value: number) {
  if (!Number.isFinite(value)) return "hsl(212 90% 45%)";
  const clamped = Math.max(0, Math.min(value, 5000));
  const progress = clamped / 5000;
  const hue = 212 - progress * 212;
  const saturation = 90 - progress * 8;
  const lightness = 45 + progress * 8;
  return `hsl(${hue.toFixed(0)} ${saturation.toFixed(0)}% ${lightness.toFixed(0)}%)`;
}

function formatLocalDateTime(value: string, locale: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "-";
  return new Intl.DateTimeFormat(locale, {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  }).format(date);
}

function formatIsoDateTime(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "-";
  return date.toISOString();
}

function resolveSelectedRange(preset: RangePreset, from: string, to: string) {
  if (preset === "custom") return { from, to };
  return rangeForPreset(preset);
}

function rangeForPreset(preset: Exclude<RangePreset, "custom">) {
  const now = new Date();
  if (preset === "yesterday") {
    const start = new Date(now);
    start.setHours(0, 0, 0, 0);
    start.setDate(start.getDate() - 1);
    const end = new Date(start);
    end.setDate(end.getDate() + 1);
    return { from: start.toISOString(), to: end.toISOString() };
  }

  const durationMs = {
    last15m: 15 * 60 * 1000,
    last1h: 60 * 60 * 1000,
    last24h: DEFAULT_RANGE_MS,
    last3d: 3 * DEFAULT_RANGE_MS,
    last7d: 7 * DEFAULT_RANGE_MS,
    last14d: 14 * DEFAULT_RANGE_MS,
    last30d: 30 * DEFAULT_RANGE_MS,
  }[preset];

  return {
    from: new Date(now.getTime() - durationMs).toISOString(),
    to: now.toISOString(),
  };
}
