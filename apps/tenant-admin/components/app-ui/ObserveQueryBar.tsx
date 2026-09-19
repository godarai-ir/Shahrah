"use client";

import * as React from "react";
import { RefreshCw, RotateCcw, Search, X } from "lucide-react";
import { useLocale, useTranslations } from "next-intl";

import { AppCard } from "@/components/app-ui/AppCard";
import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { DateTimeRangeField } from "@/components/app-ui/DateTimeRangeField";
import { DebouncedTextFilter, FilterField } from "@/components/app-ui/FilterField";
import { FilterCombobox } from "@/components/app-ui/FilterCombobox";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import {
  DEFAULT_OBSERVE_RANGE,
  EMPTY_OBSERVE_FILTERS,
  OBSERVE_FILTER_KEYS,
  OBSERVE_RANGE_PRESETS,
  isObserveFilterSupported,
  resolveObserveRange,
  type ObserveFilterKey,
  type ObserveFilters,
  type ObserveRangePreset,
  type ObserveTab,
} from "@/lib/observe-filters";
import { cn } from "@/lib/utils";

const ALL_VALUE = "__all__";

export type ObserveExtraChip = { key: string; label: string };

export type ObserveQueryBarProps = {
  /** The view rendering the bar; decides which shared filters are marked as not applied. */
  tab: ObserveTab;
  filters: ObserveFilters;
  onFiltersChange: (next: ObserveFilters) => void;
  modelOptions?: readonly string[];
  providerOptions?: readonly string[];
  onRefresh?: () => void;
  loading?: boolean;
  /** View-specific controls placed at the end of the query row (advanced filters, export, settings). */
  actions?: React.ReactNode;
  /** Active view-specific filters, shown after the shared chips. */
  extraChips?: readonly ObserveExtraChip[];
  onRemoveExtraChip?: (key: string) => void;
  /** Clears view-specific filters too; the bar resets the shared ones itself. */
  onClearAll?: () => void;
  /** View-specific panel under the query row, e.g. an advanced filter drawer. */
  children?: React.ReactNode;
  className?: string;
};

/**
 * Query bar shared by the observe views. Each view renders it first under the section tabs so the
 * three tabs look like one area with one set of filters.
 */
export function ObserveQueryBar({
  tab,
  filters,
  onFiltersChange,
  modelOptions = [],
  providerOptions = [],
  onRefresh,
  loading = false,
  actions,
  extraChips = [],
  onRemoveExtraChip,
  onClearAll,
  children,
  className,
}: ObserveQueryBarProps) {
  const t = useTranslations("common.observe");
  const locale = useLocale();
  const id = React.useId();

  const set = React.useCallback(
    (patch: Partial<ObserveFilters>) => onFiltersChange({ ...filters, ...patch }),
    [filters, onFiltersChange],
  );

  const setRange = (value: string) => {
    const range = value as ObserveRangePreset;
    if (range !== "custom") {
      set({ range, from: "", to: "" });
      return;
    }
    // Start a custom range from the window the user was looking at.
    const resolved = resolveObserveRange(filters);
    set({ range, from: filters.from || resolved.from, to: filters.to || resolved.to });
  };

  const rangeOptions = React.useMemo(
    () => OBSERVE_RANGE_PRESETS.map((value) => ({ value, label: t(`ranges.${value}`) })),
    [t],
  );
  const modelItems = React.useMemo(() => withCurrent(modelOptions, filters.model, t("allModels")), [filters.model, modelOptions, t]);
  const providerItems = React.useMemo(() => withCurrent(providerOptions, filters.provider, t("allProviders")), [filters.provider, providerOptions, t]);

  const sharedChips = OBSERVE_FILTER_KEYS.filter((key) => filters[key] !== "");
  const customRangeChip = filters.range === "custom" && filters.from && filters.to;
  const hasChips = sharedChips.length > 0 || extraChips.length > 0 || customRangeChip;
  const canClear = hasChips || filters.range !== DEFAULT_OBSERVE_RANGE;

  const clearAll = () => {
    onClearAll?.();
    onFiltersChange(EMPTY_OBSERVE_FILTERS);
  };

  return (
    <TooltipProvider delayDuration={120}>
      <AppCard className={cn("shadow-none", className)} contentClassName="flex flex-col gap-3 p-3">
        <div className="flex flex-wrap items-center gap-2" role="toolbar" aria-label={t("toolbar")}>
          <div className="relative min-w-48 basis-64 grow">
            <Search className="pointer-events-none absolute start-2 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" aria-hidden="true" />
            <DebouncedTextFilter
              id={`${id}-q`}
              label={t("search")}
              hideLabel
              value={filters.q}
              onValueChange={(q) => set({ q: q.trim() })}
              placeholder={t("searchPlaceholder")}
              mono
              className="ps-7"
            />
          </div>
          <FilterField label={t("range")} hideLabel className="w-40">
            <AppSelectField
              value={filters.range}
              onValueChange={setRange}
              options={rangeOptions}
              ariaLabel={t("range")}
              triggerClassName="h-8 text-xs"
            />
          </FilterField>
          <FilterField label={t("model")} hideLabel className="min-w-40 flex-1">
            <FilterCombobox
              id={`${id}-model`}
              options={modelItems}
              value={filters.model || ALL_VALUE}
              onValueChange={(value) => set({ model: value === ALL_VALUE ? "" : value })}
              placeholder={t("model")}
              emptyText={t("noOptions")}
            />
          </FilterField>
          <FilterField label={t("provider")} hideLabel className="min-w-40 flex-1">
            <FilterCombobox
              id={`${id}-provider`}
              options={providerItems}
              value={filters.provider || ALL_VALUE}
              onValueChange={(value) => set({ provider: value === ALL_VALUE ? "" : value })}
              placeholder={t("provider")}
              emptyText={t("noOptions")}
            />
          </FilterField>
          {actions}
          {onRefresh ? (
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  type="button"
                  variant="outline"
                  size="icon-sm"
                  onClick={onRefresh}
                  disabled={loading}
                  aria-busy={loading}
                  aria-label={t("refresh")}
                >
                  {loading ? <Spinner /> : <RefreshCw />}
                </Button>
              </TooltipTrigger>
              <TooltipContent>{t("refresh")}</TooltipContent>
            </Tooltip>
          ) : null}
        </div>

        {filters.range === "custom" ? (
          <DateTimeRangeField
            fromID={`${id}-from`}
            toID={`${id}-to`}
            from={filters.from}
            to={filters.to}
            fromLabel={t("from")}
            toLabel={t("to")}
            onFromChange={(from) => set({ from })}
            onToChange={(to) => set({ to })}
          />
        ) : null}

        {children}

        <div className="flex flex-wrap items-center gap-2">
          {customRangeChip ? (
            <Chip
              label={`${t("chips.range")}: ${formatRange(filters.from, filters.to, locale)}`}
              removeLabel={t("removeFilter", { label: t("chips.range") })}
              onRemove={() => set({ range: DEFAULT_OBSERVE_RANGE, from: "", to: "" })}
            />
          ) : null}
          {sharedChips.map((key) => (
            <SharedChip
              key={key}
              filterKey={key}
              value={filters[key]}
              applied={isObserveFilterSupported(tab, key)}
              onRemove={() => set({ [key]: "" })}
            />
          ))}
          {extraChips.map((chip) => (
            <Chip
              key={chip.key}
              label={chip.label}
              removeLabel={t("removeFilter", { label: chip.label })}
              onRemove={() => onRemoveExtraChip?.(chip.key)}
            />
          ))}
          {canClear ? (
            <Button type="button" variant="ghost" size="sm" className="h-7 px-2 text-xs" onClick={clearAll}>
              <RotateCcw data-icon="inline-start" />
              {t("clearAll")}
            </Button>
          ) : null}
          <p className="ms-auto text-xs text-muted-foreground">{t("sharedNote")}</p>
        </div>
      </AppCard>
    </TooltipProvider>
  );
}

function SharedChip({ filterKey, value, applied, onRemove }: { filterKey: ObserveFilterKey; value: string; applied: boolean; onRemove: () => void }) {
  const t = useTranslations("common.observe");
  const label = t(`chips.${filterKey}`);
  const chip = (
    <Chip
      label={`${label}: ${value}`}
      removeLabel={t("removeFilter", { label })}
      onRemove={onRemove}
      muted={!applied}
      note={applied ? undefined : t("notApplied")}
    />
  );
  if (applied) return chip;
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <span className="inline-flex">{chip}</span>
      </TooltipTrigger>
      <TooltipContent side="top" className="max-w-64 text-start leading-5">{t("notAppliedHint", { label })}</TooltipContent>
    </Tooltip>
  );
}

function Chip({ label, removeLabel, onRemove, muted = false, note }: { label: string; removeLabel: string; onRemove: () => void; muted?: boolean; note?: string }) {
  return (
    <span
      data-applied={muted ? "false" : "true"}
      className={cn(
        "inline-flex h-7 min-w-0 max-w-full items-center gap-1 rounded-full border px-2 text-xs",
        muted ? "border-dashed border-border bg-muted text-muted-foreground" : "border-border bg-background text-foreground",
      )}
    >
      <span className="min-w-0 truncate" dir="auto">{label}</span>
      {note ? <span className="shrink-0 text-muted-foreground">· {note}</span> : null}
      <button
        type="button"
        onClick={onRemove}
        aria-label={removeLabel}
        title={removeLabel}
        className="inline-flex size-4 shrink-0 items-center justify-center rounded-full hover:bg-accent focus-visible:outline-2 focus-visible:outline-ring"
      >
        <X className="size-3" aria-hidden="true" />
      </button>
    </span>
  );
}

/** Keeps a value that came from the URL selectable even when the option list does not contain it. */
function withCurrent(options: readonly string[], current: string, allLabel: string) {
  const values = Array.from(new Set([...options, ...(current ? [current] : [])].map((value) => value.trim()).filter(Boolean)));
  values.sort((a, b) => a.localeCompare(b));
  return [{ value: ALL_VALUE, label: allLabel }, ...values.map((value) => ({ value, label: value }))];
}

function formatRange(from: string, to: string, locale: string) {
  const format = new Intl.DateTimeFormat(locale, { month: "short", day: "numeric", hour: "2-digit", minute: "2-digit" });
  const start = new Date(from);
  const end = new Date(to);
  if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime())) return `${from} – ${to}`;
  return `${format.format(start)} – ${format.format(end)}`;
}
