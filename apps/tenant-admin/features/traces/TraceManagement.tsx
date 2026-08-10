"use client";

import Link from "next/link";
import dynamic from "next/dynamic";
import { useCallback, useEffect, useMemo, useState } from "react";
import type { ColumnDef } from "@tanstack/react-table";
import { ChevronDown, Filter, RefreshCw, RotateCcw, Settings2 } from "lucide-react";
import { useLocale, useTranslations } from "next-intl";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { ActiveFilterChips } from "@/components/app-ui/ActiveFilterChips";
import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { CatalogFilterCombobox } from "@/components/app-ui/CatalogFilterCombobox";
import { CompactValue } from "@/components/app-ui/CompactValue";
import { DateTimeRangeField } from "@/components/app-ui/DateTimeRangeField";
import { FilterField } from "@/components/app-ui/FilterField";
import { FormSheetActions } from "@/components/app-ui/FormSheetActions";
import { MetricCard } from "@/components/app-ui/MetricCard";
import { PageContainer } from "@/components/app-ui/PageContainer";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import { DataTable } from "@/components/ui/data-table";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Sheet, SheetContent, SheetDescription, SheetFooter, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { useIsMobile } from "@/hooks/use-mobile";
import { formatCompactNumber } from "@/lib/compact-number";
import { cn } from "@/lib/utils";
import {
  BreakdownCard,
  CompactTraceCard,
  DurationMeter,
  SortButton,
  TraceColumnVisibilityMenu,
} from "@/features/traces/trace-components";
import {
  ALL_MODELS_VALUE,
  ALL_PROVIDERS_VALUE,
  DEFAULT_RANGE_PRESET,
  PAGE_SIZE,
  RANGE_PRESETS,
  TRACE_COLUMN_KEYS,
  type RangePreset,
  type RequestTrace,
  type TraceColumnKey,
  type TraceFacets,
  type TraceFilters,
  type TraceManagementProps,
} from "@/features/traces/trace-types";
import { loadTraces } from "@/features/traces/api";
import {
  aggregateStatusBuckets,
  compactID,
  createDefaultColumnVisibility,
  createDefaultFilters,
  createEmptyFilters,
  formatIsoDateTime,
  formatLocalDateTime,
  formatPercent,
  normalizeComboboxOptions,
  normalizeError,
  normalizeFacetBuckets,
  normalizeFilters,
  rangeForPreset,
  resolveSelectedRange,
  sortBuckets,
  statusBadgeVariant,
} from "@/features/traces/trace-utils";

const TracePrivacySettingsPanel = dynamic(
  () => import("@/features/traces/TracePrivacySettingsPanel"),
  { ssr: false },
);


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
      const { list, facets } = await loadTraces(listQueryString, filtersQueryString);
      setItems(list.data ?? []);
      setTotal(list.total ?? 0);
      setFacets(facets.data ?? {});
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
                  <span className="truncate font-mono text-xs text-muted-foreground" title={item.request_id}>
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
            <span className="block max-w-64 truncate font-mono text-xs" dir="ltr" title={row.original.route}>
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
                <span className="truncate font-mono text-xs" dir="auto" title={row.original.error_code}>
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
          <div className="grid gap-3 md:grid-cols-3">
            <AppSelectField
              label={t("filters.timeRange")}
              value={draftRangePreset}
              onValueChange={(value) => handleRangePresetChange(value as RangePreset)}
              options={rangeOptions}
              ariaLabel={t("filters.timeRange")}
              className="gap-1 [&_[data-slot=field-label]]:text-xs [&_[data-slot=field-label]]:text-muted-foreground"
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
            <DateTimeRangeField
              fromID="from-filter"
              toID="to-filter"
              from={draft.from}
              to={draft.to}
              fromLabel={t("filters.from")}
              toLabel={t("filters.to")}
              onFromChange={(from) => setDraft((prev) => ({ ...prev, from }))}
              onToChange={(to) => setDraft((prev) => ({ ...prev, to }))}
              className="[&_[data-slot=field]]:gap-1 [&_[data-slot=field-label]]:text-xs [&_[data-slot=field-label]]:text-muted-foreground"
            />
          ) : null}

          {isCompactLayout ? (
            <>
              <div className="grid gap-2 border-t border-card-border/70 pt-3 sm:grid-cols-2 md:grid-cols-3">
                <Button type="button" variant="outline" size="sm" className="h-9 text-sm sm:col-span-2 md:col-span-1" onClick={() => setAdvancedSheetOpen(true)}>
                  <ChevronDown data-icon="inline-start" aria-hidden="true" />
                  {advancedFilterLabel}
                </Button>
                <Button type="button" size="sm" className="h-9 text-sm" onClick={applyFilters} disabled={loading}>
                  <Filter data-icon="inline-start" aria-hidden="true" />
                  {t("actions.apply")}
                </Button>
                <Button type="button" variant="outline" size="sm" className="h-9 text-sm" onClick={clearFilters} disabled={loading}>
                  <RotateCcw data-icon="inline-start" aria-hidden="true" />
                  {t("actions.clear")}
                </Button>
              </div>

              <Sheet open={advancedSheetOpen} onOpenChange={setAdvancedSheetOpen}>
                <SheetContent side="bottom" closeLabel={common("cancel")} className="flex max-h-svh flex-col rounded-t-2xl px-4 py-5 sm:max-w-none">
                  <SheetHeader>
                    <SheetTitle>{advancedFilterLabel}</SheetTitle>
                    <SheetDescription>{t("description")}</SheetDescription>
                  </SheetHeader>
                  <ScrollArea className="mt-4 flex-1">
                    <div className="pb-4">
                      {advancedFiltersFields}
                    </div>
                  </ScrollArea>
                  <SheetFooter className="border-t border-border px-6 py-4">
                    <FormSheetActions
                      secondary={<Button type="button" variant="outline" onClick={clearFilters} disabled={loading}>
                        <RotateCcw data-icon="inline-start" aria-hidden="true" />
                        {t("actions.clear")}
                      </Button>}
                      primary={<Button type="button" onClick={() => { applyFilters(); setAdvancedSheetOpen(false); }} disabled={loading}>
                      <Filter data-icon="inline-start" aria-hidden="true" />
                      {t("actions.apply")}
                      </Button>}
                    />
                  </SheetFooter>
                </SheetContent>
              </Sheet>
            </>
          ) : (
            <Collapsible open={advancedOpen} onOpenChange={setAdvancedOpen} className="flex flex-col gap-3">
              <div className="flex flex-wrap items-center justify-between gap-2 border-t border-card-border/70 pt-3">
                <CollapsibleTrigger asChild>
                  <Button type="button" variant="outline" size="sm" className="h-9 text-sm" aria-expanded={advancedOpen}>
                    <ChevronDown data-icon="inline-start" className={cn("transition-transform", advancedOpen && "rotate-180")} aria-hidden="true" />
                    {advancedFilterLabel}
                  </Button>
                </CollapsibleTrigger>
                <div className="flex flex-wrap items-center gap-2">
                  <Button type="button" size="sm" className="h-9 text-sm" onClick={applyFilters} disabled={loading}>
                    <Filter data-icon="inline-start" aria-hidden="true" />
                    {t("actions.apply")}
                  </Button>
                  <Button type="button" variant="outline" size="sm" className="h-9 text-sm" onClick={clearFilters} disabled={loading}>
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
        <CardHeader className="flex flex-col gap-2 p-3 md:flex-row md:items-end md:justify-between md:p-3 lg:items-start">
          <div className="min-w-0">
            <CardTitle className="text-sm md:text-base">{t("list.title")}</CardTitle>
            <CardDescription className="mt-1 text-xs">
              {t("pagination.showing", { from: showingFrom, to: showingTo, total })}
            </CardDescription>
          </div>
          {isCompactLayout ? (
            <div className="grid w-full gap-2 sm:grid-cols-[minmax(0,1fr)_auto] md:w-auto md:min-w-80 lg:w-auto">
              <AppSelectField
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
                className="h-9 text-sm sm:self-end"
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
              <div className="grid gap-2 md:grid-cols-2">
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
        <SheetContent side="left" closeLabel={common("cancel")} className="w-full overflow-y-auto sm:max-w-xl">
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
