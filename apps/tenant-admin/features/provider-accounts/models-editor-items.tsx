"use client";

import type React from "react";
import { memo, useMemo } from "react";
import { Code2, SlidersHorizontal } from "lucide-react";
import { useTranslations } from "next-intl";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Popover, PopoverTrigger } from "@/components/ui/popover";
import { Skeleton } from "@/components/ui/skeleton";
import { formatTokenCount, ModelCard, type ModelCardData } from "@/components/app-ui/ModelCard";
import { cn } from "@/lib/utils";

import { buildModelDiscoveryDetailsHref } from "./model-editor-utils";
import type { ModelDraft, ProviderCatalogModel, TenantProviderAccount } from "./types";
import {
  formatDisplayPrice,
  getModelCapabilities,
  getModelModes,
} from "./utils";

export const COMPACT_MODEL_PAGE_SIZE = 25;

export function ModelCardsSkeleton() {
  return (
    <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
      {Array.from({ length: 4 }).map((_, index) => (
        <Card key={`model-card-skeleton-${index}`} className="overflow-hidden">
          <CardHeader className="pb-3">
            <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
              <div className="flex min-w-0 flex-1 flex-col gap-2">
                <div className="flex items-center gap-2">
                  <Skeleton className="size-9 rounded-md" />
                  <div className="flex min-w-0 flex-1 flex-col gap-2">
                    <Skeleton className="h-5 w-48 max-w-full" />
                    <Skeleton className="h-4 w-64 max-w-full" />
                  </div>
                </div>
                <div className="flex flex-wrap gap-2">
                  <Skeleton className="h-6 w-20 rounded-full" />
                  <Skeleton className="h-6 w-24 rounded-full" />
                  <Skeleton className="h-6 w-28 rounded-full" />
                </div>
              </div>
              <div className="flex shrink-0 gap-2">
                <Skeleton className="h-9 w-24" />
                <Skeleton className="h-9 w-28" />
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <div className="grid gap-3 md:grid-cols-3">
              <Skeleton className="h-16" />
              <Skeleton className="h-16" />
              <Skeleton className="h-16" />
            </div>
            <div className="mt-4 flex flex-wrap items-center justify-between gap-3">
              <div className="flex gap-2">
                <Skeleton className="h-5 w-24" />
                <Skeleton className="h-5 w-24" />
              </div>
              <Skeleton className="h-9 w-32" />
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}

export type ModelCardLabels = NonNullable<React.ComponentProps<typeof ModelCard>["labels"]>;

export type PricingEditorState = {
  open: boolean;
  modelKey: string;
  pricingMode: "public" | "private";
  inputPrice: string;
  outputPrice: string;
  ratePrices: Record<string, string>;
  error: string;
};

const COMPACT_ROW_STYLE = {
  contentVisibility: "auto",
  containIntrinsicSize: "220px",
} as const;

const DEFAULT_ROW_STYLE = {
  contentVisibility: "auto",
  containIntrinsicSize: "360px",
} as const;

export function buildFallbackDraft(model: ProviderCatalogModel, selectableOptions: string[]): ModelDraft {
  return {
    enabled: false,
    selectedCapabilities: selectableOptions,
    pricingMode: "public",
    inputPrice: model.public_pricing?.input_usd_per_1m?.toString() ?? "",
    outputPrice: model.public_pricing?.output_usd_per_1m?.toString() ?? "",
  };
}

type ModelListItemProps = {
  account: TenantProviderAccount | null;
  canUseCustomPricing: boolean;
  canManage: boolean;
  compact: boolean;
  mobileDensity: "compact" | "detailed";
  displayCurrency: "USD" | "TOMAN";
  displayScale: "PER_MILLION_TOKENS" | "PER_TOKEN";
  draft?: ModelDraft;
  labels: ModelCardLabels;
  model: ProviderCatalogModel;
  openCurlGuide: (model: ProviderCatalogModel) => void;
  openPricingEditor: (model: ProviderCatalogModel) => void;
  closePricingEditor: () => void;
  pricingEditor: PricingEditorState;
  renderPricingEditorContent: (model: ProviderCatalogModel) => React.ReactNode;
  setModelEnabled: (model: ProviderCatalogModel, draft: ModelDraft, selectableOptions: string[], enabled: boolean) => void;
  toModelCardData: (model: ProviderCatalogModel, draft: ModelDraft) => ModelCardData;
  ux: ReturnType<typeof useTranslations>;
  usdToIrr: number;
  getSelectableOptions: (model: ProviderCatalogModel) => string[];
};

export const ALL_FILTER_OPTIONS = "__all__";

export function formatCapabilityLabel(value: string, ux: ReturnType<typeof useTranslations>): string {
  const normalized = value.trim().toLowerCase().replaceAll("-", "_");
  if (normalized === "function_calling" || normalized === "tool_calling" || normalized === "tools") {
    return ux("models.capabilityLabels.toolCalling");
  }
  if (normalized === "reasoning") return ux("models.capabilityLabels.reasoning");
  if (normalized === "vision" || normalized === "image") return ux("models.capabilityLabels.vision");
  if (normalized === "audio") return ux("models.capabilityLabels.audio");
  if (normalized === "streaming") return ux("models.capabilityLabels.streaming");
  return value.replaceAll("_", " ");
}

type CompactTableRowProps = {
  canManage: boolean;
  draft?: ModelDraft;
  model: ProviderCatalogModel;
  selectableOptions: string[];
  setModelEnabled: (model: ProviderCatalogModel, draft: ModelDraft, selectableOptions: string[], enabled: boolean) => void;
  displayCurrency: "USD" | "TOMAN";
  displayScale: "PER_MILLION_TOKENS" | "PER_TOKEN";
  usdToIrr: number;
  ux: ReturnType<typeof useTranslations>;
};

export const CompactTableRow = memo(function CompactTableRow(props: CompactTableRowProps) {
  const {
    canManage,
    draft: draftProp,
    model,
    selectableOptions,
    setModelEnabled,
    displayCurrency,
    displayScale,
    usdToIrr,
    ux,
  } = props;
  const draft = useMemo(
    () => draftProp ?? buildFallbackDraft(model, selectableOptions),
    [draftProp, model, selectableOptions],
  );
  const capabilities = useMemo(() => getModelCapabilities(model).slice(0, 2), [model]);
  const modes = useMemo(() => getModelModes(model).slice(0, 2), [model]);
  const priceLabel = useMemo(
    () =>
      formatDisplayPrice(model.public_pricing?.input_usd_per_1m, {
        currency: displayCurrency,
        scale: displayScale,
        usdToIrr,
        naLabel: ux("models.na"),
      }),
    [displayCurrency, displayScale, model.public_pricing?.input_usd_per_1m, usdToIrr, ux],
  );
  const outputPriceLabel = useMemo(
    () =>
      formatDisplayPrice(model.public_pricing?.output_usd_per_1m, {
        currency: displayCurrency,
        scale: displayScale,
        usdToIrr,
        naLabel: ux("models.na"),
      }),
    [displayCurrency, displayScale, model.public_pricing?.output_usd_per_1m, usdToIrr, ux],
  );
  const handleRowClick = (event: React.MouseEvent<HTMLTableRowElement>) => {
    if (!canManage || (event.target as HTMLElement).closest("button, a, input, [role='checkbox']")) return;

    setModelEnabled(model, draft, selectableOptions, !draft.enabled);
  };

  return (
    <tr
      onClick={handleRowClick}
      className={cn(
        "border-b border-border/60 align-top transition-colors",
        canManage && "cursor-pointer hover:bg-muted/30",
        draft.enabled && "bg-[var(--success-soft)]/45",
      )}
    >
      <td className="px-2 py-3 text-center">
        <Checkbox
          checked={draft.enabled}
          onCheckedChange={(checked) => setModelEnabled(model, draft, selectableOptions, checked === true)}
          aria-label={`${ux("common.enabled")} ${model.display_name || model.model}`}
          disabled={!canManage}
          className="size-4"
        />
      </td>
      <td className="w-px whitespace-nowrap px-3 py-3">
        <div className="min-w-0">
          <p className="whitespace-nowrap text-sm font-semibold text-foreground" dir="ltr" title={model.display_name || model.model}>
            {model.display_name || model.model}
          </p>
        </div>
      </td>
      <td className="px-3 py-3 text-center">
        <span className="inline-flex rounded-full bg-muted/35 px-2.5 py-1 text-xs font-medium" dir="ltr">
          {formatTokenCount(model.context_window)}
        </span>
      </td>
      <td className="px-3 py-3">
        <div className="truncate whitespace-nowrap text-xs font-medium" dir="ltr">
          {priceLabel}
        </div>
      </td>
      <td className="px-3 py-3">
        <div className="truncate whitespace-nowrap text-xs font-medium" dir="ltr">
          {outputPriceLabel}
        </div>
      </td>
      <td className="hidden px-3 py-3 md:table-cell">
        <div className="flex min-w-0 flex-wrap gap-1">
          {capabilities.map((capability) => (
            <Badge key={`${model.model}-cap-${capability}`} variant="outline" className="px-2 py-0 text-xs" title={capability}>
              {formatCapabilityLabel(capability, ux)}
            </Badge>
          ))}
          {modes.map((mode) => (
            <Badge key={`${model.model}-mode-${mode}`} variant="secondary" className="px-2 py-0 text-xs">
              {mode}
            </Badge>
          ))}
          {capabilities.length === 0 && modes.length === 0 ? (
            <span className="text-xs text-muted-foreground">{ux("models.na")}</span>
          ) : null}
        </div>
      </td>
    </tr>
  );
}, (prev, next) => {
  return (
    prev.canManage === next.canManage &&
    prev.draft === next.draft &&
    prev.model === next.model &&
    prev.selectableOptions === next.selectableOptions &&
    prev.displayCurrency === next.displayCurrency &&
    prev.displayScale === next.displayScale &&
    prev.usdToIrr === next.usdToIrr &&
    prev.ux === next.ux
  );
});

export const ModelListItem = memo(function ModelListItem(props: ModelListItemProps) {
  const {
    account,
    canManage,
    canUseCustomPricing,
    compact,
    mobileDensity,
    displayCurrency,
    displayScale,
    draft: draftProp,
    getSelectableOptions,
    labels,
    model,
    openCurlGuide,
    openPricingEditor,
    closePricingEditor,
    pricingEditor,
    renderPricingEditorContent,
    setModelEnabled,
    toModelCardData,
    usdToIrr,
    ux,
  } = props;
  const selectableOptions = useMemo(
    () => getSelectableOptions(model),
    [getSelectableOptions, model],
  );
  const draft = useMemo(
    () => draftProp ?? buildFallbackDraft(model, selectableOptions),
    [draftProp, model, selectableOptions],
  );
  const displayDraft = useMemo(
    () => (canUseCustomPricing ? draft : { ...draft, pricingMode: "public" as const }),
    [canUseCustomPricing, draft],
  );
  const cardData = useMemo(() => toModelCardData(model, displayDraft), [displayDraft, model, toModelCardData]);
  const pricingDisplay = useMemo(
    () => ({
      currency: displayCurrency,
      scale: displayScale,
      usdToIrr,
    }),
    [displayCurrency, displayScale, usdToIrr],
  );
  const isPricingEditorOpen = pricingEditor.open && pricingEditor.modelKey === model.model;

  return (
    <Popover
      open={isPricingEditorOpen}
      onOpenChange={(open) => {
        if (open) {
          if (!canUseCustomPricing) return;
          openPricingEditor(model);
          return;
        }
        closePricingEditor();
      }}
    >
      <div className="min-w-0 space-y-2" style={compact ? COMPACT_ROW_STYLE : DEFAULT_ROW_STYLE}>
        <ModelCard
          model={cardData}
          mobileDensity={mobileDensity}
          detailsHref={buildModelDiscoveryDetailsHref(model)}
          pricingDisplay={pricingDisplay}
          formatCapabilityLabel={(value) => formatCapabilityLabel(value, ux)}
          isSelected={displayDraft.enabled}
          onSelectedChange={(checked) => setModelEnabled(model, displayDraft, selectableOptions, checked)}
          selectedCheckboxLabel={`${ux("common.enabled")} ${model.display_name || model.model}`}
          selectedControlLabel={null}
          pricingModeBadge={displayDraft.pricingMode === "private" ? ux("models.pricing.private") : undefined}
          hasCustomPricing={displayDraft.pricingMode === "private"}
          showCopyAction={Boolean(account)}
          className={
            compact
              ? displayDraft.enabled
                ? "self-start border-success-soft-border"
                : "self-start"
              : displayDraft.enabled
                ? "border-success-soft-border"
                : undefined
          }
          footerActions={account ? (
            <Button type="button" variant="outline" size="sm" onClick={() => openCurlGuide(model)}>
              <Code2 data-icon="inline-start" />
              {ux("models.card.code")}
            </Button>
          ) : undefined}
          pricingAction={canUseCustomPricing ? (
            <PopoverTrigger asChild>
              <Button
                type="button"
                variant="outline"
                size="icon-sm"
                disabled={!canManage}
                aria-label={`${ux("models.actions.editPricing")} ${model.display_name || model.model}`}
                onClick={() => openPricingEditor(model)}
              >
                <SlidersHorizontal aria-hidden="true" />
              </Button>
            </PopoverTrigger>
          ) : undefined}
          labels={labels}
        />
      </div>
      {renderPricingEditorContent(model)}
    </Popover>
  );
}, (prev, next) => {
  if (
    prev.account !== next.account ||
    prev.canUseCustomPricing !== next.canUseCustomPricing ||
    prev.canManage !== next.canManage ||
    prev.compact !== next.compact ||
    prev.mobileDensity !== next.mobileDensity ||
    prev.displayCurrency !== next.displayCurrency ||
    prev.displayScale !== next.displayScale ||
    prev.draft !== next.draft ||
    prev.labels !== next.labels ||
    prev.model !== next.model ||
    prev.ux !== next.ux ||
    prev.usdToIrr !== next.usdToIrr
  ) {
    return false;
  }

  const prevOpen = prev.pricingEditor.open && prev.pricingEditor.modelKey === prev.model.model;
  const nextOpen = next.pricingEditor.open && next.pricingEditor.modelKey === next.model.model;
  if (prevOpen !== nextOpen) return false;

  if (
    nextOpen &&
    (
      prev.pricingEditor.pricingMode !== next.pricingEditor.pricingMode ||
      prev.pricingEditor.inputPrice !== next.pricingEditor.inputPrice ||
      prev.pricingEditor.outputPrice !== next.pricingEditor.outputPrice ||
      prev.pricingEditor.ratePrices !== next.pricingEditor.ratePrices ||
      prev.pricingEditor.error !== next.pricingEditor.error
    )
  ) {
    return false;
  }

  return true;
});
