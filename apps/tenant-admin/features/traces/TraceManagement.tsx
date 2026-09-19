"use client";

import dynamic from "next/dynamic";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { ChevronDown, Settings2, SlidersHorizontal } from "lucide-react";
import { useLocale, useTranslations } from "next-intl";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { CompactValue } from "@/components/app-ui/CompactValue";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { DebouncedTextFilter, FilterField } from "@/components/app-ui/FilterField";
import { FilterCombobox } from "@/components/app-ui/FilterCombobox";
import { MetricCard } from "@/components/app-ui/MetricCard";
import { ObserveQueryBar, type ObserveExtraChip } from "@/components/app-ui/ObserveQueryBar";
import { PageContainer } from "@/components/app-ui/PageContainer";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Sheet, SheetContent, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { useIsMobile } from "@/hooks/use-mobile";
import { formatCompactNumber } from "@/lib/compact-number";
import {
  mergeObserveSearchParams,
  observeFiltersFromSearchParams,
  resolveObserveRange,
  type ObserveFilters,
} from "@/lib/observe-filters";
import { cn } from "@/lib/utils";
import { BreakdownCard, TraceListItem } from "@/features/traces/trace-components";
import { TraceDetailPanel } from "@/features/traces/TraceDetailPanel";
import {
  PAGE_SIZE,
  type RequestTrace,
  type TraceFacets,
  type TraceLocalFilters,
  type TraceManagementProps,
} from "@/features/traces/trace-types";
import { loadTraces } from "@/features/traces/api";
import {
  aggregateStatusBuckets,
  buildTraceFiltersQuery,
  compactID,
  createDefaultLocalFilters,
  formatPercent,
  normalizeComboboxOptions,
  normalizeError,
  normalizeFacetBuckets,
  sortBuckets,
  traceFiltersFromObserve,
} from "@/features/traces/trace-utils";

const TracePrivacySettingsPanel = dynamic(
  () => import("@/features/traces/TracePrivacySettingsPanel"),
  { ssr: false },
);

/** URL parameter holding the trace open in the detail panel. */
const SELECTED_TRACE_PARAM = "trace";
const ANY_VALUE = "__any__";

export default function TraceManagement(props: TraceManagementProps) {
  const locale = useLocale();
  const t = useTranslations("traces");
  const tracePrivacyT = useTranslations("settings.tracePrivacy");
  const common = useTranslations("common");
  const isCompactLayout = useIsMobile();
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  const [items, setItems] = useState<RequestTrace[]>(props.initialData);
  const [total, setTotal] = useState(props.initialTotal);
  const [offset, setOffset] = useState(0);
  const [error, setError] = useState<string | null>(props.initialError);
  const [loading, setLoading] = useState(false);
  const [facets, setFacets] = useState<TraceFacets>({});

  const [observe, setObserve] = useState<ObserveFilters>(() => observeFiltersFromSearchParams(searchParams));
  const [local, setLocal] = useState<TraceLocalFilters>(() => createDefaultLocalFilters());
  const [selectedID, setSelectedID] = useState(() => searchParams.get(SELECTED_TRACE_PARAM)?.trim() ?? "");
  // Relative ranges are re-anchored to "now" on every refresh.
  const [rangeAnchor, setRangeAnchor] = useState(() => Date.now());
  const [advancedOpen, setAdvancedOpen] = useState(false);
  const [tracePrivacyOpen, setTracePrivacyOpen] = useState(props.initialTracePrivacyOpen ?? false);
  const loadRequestRef = useRef(0);

  const applied = useMemo(
    () => traceFiltersFromObserve(local, observe, resolveObserveRange(observe, new Date(rangeAnchor))),
    [local, observe, rangeAnchor],
  );
  const filtersQueryString = useMemo(() => buildTraceFiltersQuery(applied), [applied]);
  const listQueryString = useMemo(() => {
    const params = new URLSearchParams(filtersQueryString);
    params.set("limit", String(PAGE_SIZE));
    params.set("offset", String(offset));
    return params.toString();
  }, [filtersQueryString, offset]);

  useEffect(() => {
    const params = mergeObserveSearchParams(searchParams, observe);
    if (selectedID) params.set(SELECTED_TRACE_PARAM, selectedID);
    else params.delete(SELECTED_TRACE_PARAM);
    const nextQuery = params.toString();
    if (nextQuery === searchParams.toString()) return;
    router.replace(nextQuery ? `${pathname}?${nextQuery}` : pathname, { scroll: false });
  }, [observe, pathname, router, searchParams, selectedID]);

  useEffect(() => {
    const requestID = ++loadRequestRef.current;
    const isCurrent = () => loadRequestRef.current === requestID;
    queueMicrotask(() => {
      if (!isCurrent()) return;
      setLoading(true);
      loadTraces(listQueryString, filtersQueryString)
        .then(({ list, facets: nextFacets }) => {
          if (!isCurrent()) return;
          setItems(list.data ?? []);
          setTotal(list.total ?? 0);
          setFacets(nextFacets.data ?? {});
          setError(null);
        })
        .catch((err: unknown) => {
          if (isCurrent()) setError(normalizeError(err, common("requestError")));
        })
        .finally(() => {
          if (isCurrent()) setLoading(false);
        });
    });
  }, [common, filtersQueryString, listQueryString]);

  // A filter change re-anchors relative ranges to now; paging keeps the anchor so pages stay consistent.
  const updateObserve = useCallback((next: ObserveFilters) => {
    setObserve(next);
    setOffset(0);
    setRangeAnchor(Date.now());
  }, []);
  const updateLocal = useCallback((patch: Partial<TraceLocalFilters>) => {
    setLocal((current) => ({ ...current, ...patch }));
    setOffset(0);
    setRangeAnchor(Date.now());
  }, []);
  const refresh = useCallback(() => setRangeAnchor(Date.now()), []);

  const statusBuckets = useMemo(() => {
    const source = facets.status_codes?.length ? facets.status_codes : aggregateStatusBuckets(items);
    return source
      .map((item) => ({ value: String(item.status_code), label: String(item.status_code), count: item.count }))
      .sort(sortBuckets);
  }, [facets.status_codes, items]);

  const statusTotal = statusBuckets.reduce((sum, item) => sum + item.count, 0);
  const errorCount = statusBuckets.reduce((sum, item) => (Number(item.value) >= 400 ? sum + item.count : sum), 0);
  const errorRate = statusTotal > 0 ? errorCount / statusTotal : 0;
  const maxDuration = Math.max(0, ...items.map((item) => item.duration_ms));
  const showingFrom = total === 0 ? 0 : offset + 1;
  const showingTo = Math.min(offset + items.length, total);
  const selectedItem = items.find((item) => item.trace_id === selectedID) ?? null;
  const capturePolicy = props.initialTracePrivacySettings?.capture_policy ?? null;

  const extraChips = useMemo<ObserveExtraChip[]>(() => {
    const out: ObserveExtraChip[] = [];
    if (local.traceID) out.push({ key: "traceID", label: `${t("activeFilters.traceId")}: ${compactID(local.traceID)}` });
    if (local.route) out.push({ key: "route", label: `${t("activeFilters.route")}: ${local.route}` });
    return out;
  }, [local.route, local.traceID, t]);

  const breakdownCards = [
    { key: "statusCodes", title: t("breakdown.statusCodes"), buckets: statusBuckets, activeValue: observe.status, onSelect: (value: string) => updateObserve({ ...observe, status: value }), mono: true },
    { key: "providers", title: t("breakdown.providers"), buckets: normalizeFacetBuckets(facets.providers, t("facets.empty")), activeValue: observe.provider, onSelect: (value: string) => updateObserve({ ...observe, provider: value }), mono: false },
    { key: "models", title: t("breakdown.models"), buckets: normalizeFacetBuckets(facets.models, t("facets.empty")), activeValue: observe.model, onSelect: (value: string) => updateObserve({ ...observe, model: value }), mono: true },
    { key: "routes", title: t("breakdown.routes"), buckets: normalizeFacetBuckets(facets.routes, t("facets.empty")), activeValue: local.route, onSelect: (value: string) => updateLocal({ route: value }), mono: true },
    { key: "errorCodes", title: t("breakdown.errorCodes"), buckets: normalizeFacetBuckets(facets.error_codes, t("facets.empty")), activeValue: observe.errorCode, onSelect: (value: string) => updateObserve({ ...observe, errorCode: value }), mono: true },
  ];

  const renderBreakdown = (item: (typeof breakdownCards)[number], compact: boolean) => (
    <BreakdownCard
      title={item.title}
      buckets={item.buckets}
      activeValue={item.activeValue}
      noDataLabel={t("facets.noData")}
      selectLabel={(value, count) => t("breakdown.select", { label: item.title, value, count })}
      onSelect={item.onSelect}
      mono={item.mono}
      compact={compact}
    />
  );

  const sortOptions = [
    { value: "started_at", label: t("list.columns.startedAt") },
    { value: "status_code", label: t("list.columns.status") },
    { value: "duration_ms", label: t("list.columns.duration") },
  ];

  const detailPanel = selectedID ? (
    <TraceDetailPanel
      key={selectedID}
      traceID={selectedID}
      item={selectedItem}
      capturePolicy={capturePolicy}
      onClose={() => setSelectedID("")}
    />
  ) : (
    <EmptyState title={t("panel.placeholderTitle")} description={t("panel.placeholderDescription")} density="compact" />
  );

  return (
    <PageContainer spacing="compact" padding="none" className="md:gap-6">
      <Collapsible open={advancedOpen} onOpenChange={setAdvancedOpen}>
        <ObserveQueryBar
          tab="traces"
          filters={observe}
          onFiltersChange={updateObserve}
          modelOptions={normalizeComboboxOptions(facets.models).map((option) => option.value)}
          providerOptions={normalizeComboboxOptions(facets.providers).map((option) => option.value)}
          onRefresh={refresh}
          loading={loading}
          extraChips={extraChips}
          onRemoveExtraChip={(key) => updateLocal({ [key]: "" })}
          onClearAll={() => updateLocal(createDefaultLocalFilters())}
          actions={
            <>
              <CollapsibleTrigger asChild>
                <Button type="button" variant="outline" size="sm" aria-expanded={advancedOpen}>
                  <SlidersHorizontal data-icon="inline-start" aria-hidden="true" />
                  {advancedOpen ? t("filters.advanced.hide") : t("filters.advanced.show")}
                  <ChevronDown data-icon="inline-end" className={cn("transition-transform", advancedOpen && "rotate-180")} aria-hidden="true" />
                </Button>
              </CollapsibleTrigger>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button type="button" variant="outline" size="icon-sm" onClick={() => setTracePrivacyOpen(true)} aria-label={t("privacySettings")}>
                    <Settings2 aria-hidden="true" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>{t("privacySettings")}</TooltipContent>
              </Tooltip>
            </>
          }
        >
          <CollapsibleContent className="grid gap-3 border-t border-border pt-3 md:grid-cols-2 xl:grid-cols-4">
            <FacetFilter
              id="trace-status-filter"
              label={t("filters.statusCode")}
              value={observe.status}
              options={statusBuckets.map((bucket) => bucket.value)}
              onChange={(status) => updateObserve({ ...observe, status })}
              emptyText={t("filters.emptyOptions")}
            />
            <FacetFilter
              id="trace-error-filter"
              label={t("filters.errorCode")}
              value={observe.errorCode}
              options={normalizeComboboxOptions(facets.error_codes).map((option) => option.value)}
              onChange={(errorCode) => updateObserve({ ...observe, errorCode })}
              emptyText={t("filters.emptyOptions")}
            />
            <FacetFilter
              id="trace-route-filter"
              label={t("filters.route")}
              value={local.route}
              options={normalizeComboboxOptions(facets.routes).map((option) => option.value)}
              onChange={(route) => updateLocal({ route })}
              emptyText={t("filters.emptyOptions")}
            />
            <DebouncedTextFilter
              label={t("filters.traceId")}
              value={local.traceID}
              onValueChange={(traceID) => updateLocal({ traceID: traceID.trim() })}
              placeholder={t("filters.placeholders.traceId")}
              mono
            />
          </CollapsibleContent>
        </ObserveQueryBar>
      </Collapsible>

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

      <section aria-label={t("breakdown.title")}>
        {isCompactLayout ? (
          <ScrollArea className="w-full">
            <div className="flex items-start gap-2 pb-3">
              {breakdownCards.map((item) => (
                <div key={item.key} className="min-w-72 max-w-72 shrink-0">
                  {renderBreakdown(item, true)}
                </div>
              ))}
            </div>
          </ScrollArea>
        ) : (
          <div className="grid gap-2 lg:grid-cols-2 xl:grid-cols-5">
            {breakdownCards.map((item) => <div key={item.key}>{renderBreakdown(item, false)}</div>)}
          </div>
        )}
      </section>

      <div className="grid items-start gap-3 lg:grid-cols-[20rem_minmax(0,1fr)]">
        <Card className="rounded-lg">
          <CardHeader className="flex flex-col gap-2 p-3 md:p-3">
            <div className="min-w-0">
              <CardTitle className="text-sm">{t("list.title")}</CardTitle>
              <CardDescription className="mt-1 text-xs">{t("pagination.showing", { from: showingFrom, to: showingTo, total })}</CardDescription>
            </div>
            <div className="flex items-end gap-2">
              <AppSelectField
                value={local.sortBy}
                onValueChange={(sortBy) => updateLocal({ sortBy, sortOrder: sortBy === local.sortBy ? local.sortOrder : "desc" })}
                options={sortOptions}
                ariaLabel={t("list.sort.field")}
                triggerClassName="h-8 w-full min-w-0 text-xs"
                className="min-w-0 flex-1"
              />
              <Button
                type="button"
                variant="outline"
                size="sm"
                aria-label={t("list.sort.directionLabel", { direction: local.sortOrder === "asc" ? t("list.sort.directionAsc") : t("list.sort.directionDesc") })}
                onClick={() => updateLocal({ sortOrder: local.sortOrder === "desc" ? "asc" : "desc" })}
              >
                <span aria-hidden="true">{local.sortOrder === "asc" ? "↑" : "↓"}</span>
                {local.sortOrder === "asc" ? t("list.sort.directionAsc") : t("list.sort.directionDesc")}
              </Button>
            </div>
          </CardHeader>
          <CardContent className="flex flex-col gap-2 px-3 pb-3 md:px-3 md:pb-3">
            {items.length > 0 ? (
              <ul className="flex flex-col gap-1.5">
                {items.map((item) => (
                  <li key={item.trace_id}>
                    <TraceListItem item={item} selected={item.trace_id === selectedID} maxDuration={maxDuration} onSelect={setSelectedID} />
                  </li>
                ))}
              </ul>
            ) : (
              <p className="rounded-lg border border-dashed border-card-border px-4 py-10 text-center text-sm text-muted-foreground">{t("list.empty")}</p>
            )}
            <div className="flex justify-end gap-2">
              <Button size="sm" variant="outline" disabled={offset === 0 || loading} onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}>
                {t("pagination.prev")}
              </Button>
              <Button size="sm" variant="outline" disabled={offset + items.length >= total || loading} onClick={() => setOffset(offset + PAGE_SIZE)}>
                {t("pagination.next")}
              </Button>
            </div>
          </CardContent>
        </Card>

        {isCompactLayout ? null : (
          <Card className="rounded-lg lg:sticky lg:top-20">
            <CardContent className="p-4 md:p-4">{detailPanel}</CardContent>
          </Card>
        )}
      </div>

      {isCompactLayout ? (
        <Sheet open={Boolean(selectedID)} onOpenChange={(open) => { if (!open) setSelectedID(""); }}>
          <SheetContent side="bottom" closeLabel={common("cancel")} className="max-h-svh overflow-y-auto rounded-t-2xl px-4 py-5 sm:max-w-none">
            <SheetHeader className="sr-only">
              <SheetTitle>{t("panel.title")}</SheetTitle>
            </SheetHeader>
            {selectedID ? detailPanel : null}
          </SheetContent>
        </Sheet>
      ) : null}

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

function FacetFilter({ id, label, value, options, onChange, emptyText }: { id: string; label: string; value: string; options: string[]; onChange: (value: string) => void; emptyText: string }) {
  const items = [{ value: ANY_VALUE, label }, ...Array.from(new Set([...options, ...(value ? [value] : [])])).map((option) => ({ value: option, label: option }))];
  return (
    <FilterField label={label} htmlFor={id}>
      <FilterCombobox id={id} options={items} value={value || ANY_VALUE} onValueChange={(next) => onChange(next === ANY_VALUE ? "" : next)} placeholder={label} emptyText={emptyText} mono />
    </FilterField>
  );
}
