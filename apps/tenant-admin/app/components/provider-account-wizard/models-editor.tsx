"use client";

import type React from "react";
import { copyTextToClipboard } from "@/components/app-ui/CopyableValue";
import { memo, useCallback, useDeferredValue, useEffect, useMemo, useState } from "react";
import { ArrowUpDown, CheckSquare, Code2, Copy, GitCompare, RotateCcw, Search, SlidersHorizontal, Trash2, X } from "lucide-react";
import { useTranslations } from "next-intl";

import { AppSelectField, EmptyState } from "@/components/app-ui";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import { formatTokenCount, ModelCard, type ModelCardData } from "@/app/components/model-card";
import {
  clearVisibleModelSelection,
} from "@/app/components/provider-account-wizard.models";
import { cn } from "@/lib/utils";
import { CodeExampleBlock } from "./code-example-block";
import {
  buildModelDiscoveryDetailsHref,
  createInitialModelFilters,
  filterAndSortProviderModels,
  getActiveModelFilterChips,
  getBestCompareMarkers,
  getCompareKpis,
  getModelFamily,
  getRecommendedModel,
  hasInvalidModelPriceRange,
  normalizeProviderScopedModelName,
  sortComparedModels,
  type CompareSortMode,
  type ModelFilterState,
} from "./model-editor-utils";
import type {
  Banner,
  CurlGuide,
  CurlGuideShell,
  DisplaySettingsResponse,
  ModelDraft,
  ProviderCatalogModel,
  TenantProviderAccount,
} from "./types";
import { CURL_GUIDE_LANGUAGE } from "./types";
import {
  buildBashCurlGuideCommand,
  buildCurlGuidePayload,
  buildCurlGuideSampleResponse,
  buildPowerShellCurlGuideCommand,
  formatDisplayPrice,
  getCurlGuideEndpoint,
  getModelCapabilities,
  getModelModes,
  getModelSelectionTypes,
  getTenantApiBaseURL,
  normalizeCapabilityList,
  requestJSON,
  resolveCurlGuideMode,
  toNumberOrUndefined,
  unwrapData,
} from "./utils";

function ModelCardsSkeleton() {
  return (
    <div className="flex flex-col gap-3">
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

type ModelCardLabels = NonNullable<React.ComponentProps<typeof ModelCard>["labels"]>;

type PricingEditorState = {
  open: boolean;
  modelKey: string;
  pricingMode: "public" | "private";
  inputPrice: string;
  outputPrice: string;
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

function buildFallbackDraft(model: ProviderCatalogModel, selectableOptions: string[]): ModelDraft {
  return {
    enabled: false,
    compared: false,
    selectedCapabilities: selectableOptions,
    pricingMode: "public",
    inputPrice: model.public_pricing?.input_usd_per_1m?.toString() ?? "",
    outputPrice: model.public_pricing?.output_usd_per_1m?.toString() ?? "",
  };
}

type ModelListItemProps = {
  account: TenantProviderAccount | null;
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
  setModelCompared: (model: ProviderCatalogModel, draft: ModelDraft, compared: boolean) => void;
  setModelEnabled: (model: ProviderCatalogModel, draft: ModelDraft, selectableOptions: string[], enabled: boolean) => void;
  toModelCardData: (model: ProviderCatalogModel, draft: ModelDraft) => ModelCardData;
  ux: ReturnType<typeof useTranslations>;
  usdToIrr: number;
  getSelectableOptions: (model: ProviderCatalogModel) => string[];
};

const ALL_FILTER_OPTIONS = "__all__";

function formatCapabilityLabel(value: string, ux: ReturnType<typeof useTranslations>): string {
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
  recommended: boolean;
};

const CompactTableRow = memo(function CompactTableRow(props: CompactTableRowProps) {
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
    recommended,
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

  return (
    <tr className={cn("border-b border-border/60 align-top", draft.enabled && "bg-[var(--success-soft)]/45")}>
      <td className="px-2 py-3 text-center">
        <Checkbox
          checked={draft.enabled}
          onCheckedChange={(checked) => setModelEnabled(model, draft, selectableOptions, checked === true)}
          aria-label={`${ux("common.enabled")} ${model.display_name || model.model}`}
          disabled={!canManage}
          className="size-4"
        />
      </td>
      <td className="px-3 py-3">
        <div className="min-w-0 space-y-1">
          <p className="line-clamp-2 text-sm font-semibold text-foreground" dir="ltr" title={model.display_name || model.model}>
            {model.display_name || model.model}
          </p>
          {recommended ? <Badge className="mt-1" variant="secondary">{ux("models.card.recommended")}</Badge> : null}
          <p className="truncate text-xs text-muted-foreground" dir="ltr" title={model.model}>
            {normalizeProviderScopedModelName(model.provider, model.model)}
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
    prev.ux === next.ux &&
    prev.recommended === next.recommended
  );
});

const ModelListItem = memo(function ModelListItem(props: ModelListItemProps) {
  const {
    canManage,
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
    setModelCompared,
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
  const cardData = useMemo(() => toModelCardData(model, draft), [draft, model, toModelCardData]);
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
          isSelected={draft.enabled}
          onSelectedChange={(checked) => setModelEnabled(model, draft, selectableOptions, checked)}
          selectedCheckboxLabel={ux("common.enabled")}
          isCompared={compact ? undefined : draft.compared}
          onComparedChange={
            compact
              ? undefined
              : (checked) => setModelCompared(model, draft, checked)
          }
          compareCheckboxLabel={compact ? undefined : ux("models.actions.compareToggle")}
          pricingModeBadge={draft.pricingMode === "private" ? ux("models.pricing.private") : undefined}
          hasCustomPricing={draft.pricingMode === "private"}
          showCopyAction
          className={
            compact
              ? draft.enabled
                ? "border-success/40 self-start"
                : "self-start"
              : draft.enabled
                ? "border-success/40"
                : undefined
          }
          footerActions={
            <Button type="button" variant="outline" size="sm" onClick={() => openCurlGuide(model)}>
                <Code2 data-icon="inline-start" />
                {ux("models.card.curlGuide")}
            </Button>
          }
          pricingAction={
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
          }
          labels={labels}
        />
      </div>
      {renderPricingEditorContent(model)}
    </Popover>
  );
}, (prev, next) => {
  if (
    prev.account !== next.account ||
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
      prev.pricingEditor.error !== next.pricingEditor.error
    )
  ) {
    return false;
  }

  return true;
});

export const ModelsEditor = memo(function ModelsEditor(props: {
  t: ReturnType<typeof useTranslations>;
  account: TenantProviderAccount | null;
  setBanner: React.Dispatch<React.SetStateAction<Banner | null>>;
  providerModels: ProviderCatalogModel[];
  modelDrafts: Record<string, ModelDraft>;
  setModelDrafts: React.Dispatch<React.SetStateAction<Record<string, ModelDraft>>>;
  isSaving: boolean;
  isLoadingModels: boolean;
  canManage: boolean;
  onSave: () => void;
  variant?: "default" | "compact";
}) {
  const ux = useTranslations("providerAccounts.ux");
  const {
    modelDrafts,
    providerModels,
    setModelDrafts,
  } = props;
  const initialFilters = useMemo(() => createInitialModelFilters(""), []);
  const [filters, setFilters] = useState<ModelFilterState>(() => createInitialModelFilters(""));
  const [compareOpen, setCompareOpen] = useState(false);
  const [compareSortBy, setCompareSortBy] = useState<CompareSortMode>("input_asc");
  const [compareDiffOnly, setCompareDiffOnly] = useState(false);
  const [compareNormalizedPricing, setCompareNormalizedPricing] = useState(true);
  const [displayCurrency, setDisplayCurrency] = useState<"USD" | "TOMAN">("USD");
  const [displayScale, setDisplayScale] = useState<"PER_MILLION_TOKENS" | "PER_TOKEN">("PER_MILLION_TOKENS");
  const [usdToIrr, setUsdToIrr] = useState(0);
  const [curlGuide, setCurlGuide] = useState<CurlGuide | null>(null);
  const [curlGuideShell, setCurlGuideShell] = useState<CurlGuideShell>("powershell");
  const [showSelectedOnly, setShowSelectedOnly] = useState(false);
  const [mobileViewMode, setMobileViewMode] = useState<"table" | "card">("card");
  const [pricingEditor, setPricingEditor] = useState<PricingEditorState>({
    open: false,
    modelKey: "",
    pricingMode: "public",
    inputPrice: "",
    outputPrice: "",
    error: "",
  });
  const deferredFilters = useDeferredValue(filters);
  const tenantBaseUrl = useMemo(() => {
    return getTenantApiBaseURL();
  }, []);
  const modelCardLabels = useMemo(
    () => ({
      context: ux("models.card.context"),
      maxOutput: ux("models.card.maxOutput"),
      input: ux("models.labels.input"),
      output: ux("models.labels.output"),
      latency: ux("models.card.latency"),
      quality: ux("models.card.quality"),
      pricing: ux("models.card.pricing"),
      priceInputOutput: ux("models.card.priceInputOutput"),
      pricingUnavailable: ux("models.card.pricingUnavailable"),
      technicalMetadata: ux("models.card.technicalMetadata"),
      inputFormats: ux("models.card.inputFormats"),
      outputFormats: ux("models.card.outputFormats"),
      capabilities: ux("models.card.capabilities"),
      invocationModes: ux("models.card.invocationModes"),
      fqn: ux("models.card.fqn"),
      customPricing: ux("models.card.customPricing"),
      pricingPerMillion: ux("models.card.pricingPerMillion"),
      modelId: ux("models.card.modelId"),
      providerId: ux("models.card.providerId"),
      apiMode: ux("models.card.apiMode"),
      releaseDate: ux("models.card.releaseDate"),
      knowledgeCutoff: ux("models.card.knowledgeCutoff"),
      regions: ux("models.card.regions"),
      rateLimitNotes: ux("models.card.rateLimitNotes"),
      license: ux("models.card.license"),
      viewDetails: ux("models.card.viewDetails"),
      copyId: ux("models.card.copyFqn"),
      copied: ux("models.card.copied"),
      lowerCost: ux("models.card.lowerCost"),
      recommended: ux("models.card.recommended"),
      popular: ux("models.card.popular"),
      tradeOffs: ux("models.card.tradeOffs"),
      cachedInput: ux("models.card.cachedInput"),
      image: ux("models.card.image"),
      audio: ux("models.card.audio"),
      statusStable: ux("models.card.statusStable"),
      statusPreview: ux("models.card.statusPreview"),
      statusDeprecated: ux("models.card.statusDeprecated"),
      statusNew: ux("models.card.statusNew"),
      categoryChat: ux("models.card.categoryChat"),
      categoryReasoning: ux("models.card.categoryReasoning"),
      categoryEmbedding: ux("models.card.categoryEmbedding"),
      categoryVision: ux("models.card.categoryVision"),
      categoryAudio: ux("models.card.categoryAudio"),
      categoryCode: ux("models.card.categoryCode"),
      categoryMultimodal: ux("models.card.categoryMultimodal"),
    }),
    [ux],
  );

  const buildModelFqn = useCallback((modelValue: ProviderCatalogModel) => {
    if (!props.account) return modelValue.model;
    const officialModel = normalizeProviderScopedModelName(props.account.provider, modelValue.model);
    return `${props.account.provider}:${props.account.name}:${officialModel}`;
  }, [props.account]);

  const openCurlGuide = useCallback((modelValue: ProviderCatalogModel) => {
    const modelFqn = buildModelFqn(modelValue);
    const officialModel = normalizeProviderScopedModelName(modelValue.provider, modelValue.model);
    const guideMode = resolveCurlGuideMode(modelValue);
    const endpoint = getCurlGuideEndpoint(guideMode);
    const bashCommand = buildBashCurlGuideCommand({
      baseUrl: tenantBaseUrl,
      endpoint,
      payload: buildCurlGuidePayload(guideMode, modelFqn),
    });
    const powershellCommand = buildPowerShellCurlGuideCommand({
      baseUrl: tenantBaseUrl,
      endpoint,
      mode: guideMode,
      modelFqn,
    });
    setCurlGuideShell("powershell");
    setCurlGuide({
      modelName: modelValue.display_name || modelValue.model,
      modelFqn,
      mode: guideMode,
      commands: {
        powershell: powershellCommand,
        bash: bashCommand,
      },
      sampleResponse: buildCurlGuideSampleResponse(guideMode, officialModel),
    });
  }, [buildModelFqn, tenantBaseUrl]);

  const comparedModels = useMemo(
    () => providerModels.filter((model) => modelDrafts[model.model]?.compared),
    [modelDrafts, providerModels],
  );
  const enabledModelCount = useMemo(
    () => Object.values(modelDrafts).filter((draft) => draft.enabled).length,
    [modelDrafts],
  );
  const selectedOnlyActive = showSelectedOnly && enabledModelCount > 0;

  const allInputModalities = useMemo(
    () => Array.from(new Set(providerModels.flatMap((model) => model.input_modalities ?? []))),
    [providerModels],
  );
  const allOutputModalities = useMemo(
    () => Array.from(new Set(providerModels.flatMap((model) => model.output_modalities ?? []))),
    [providerModels],
  );
  const allFeatures = useMemo(
    () => Array.from(new Set(providerModels.flatMap((model) => getModelCapabilities(model)))),
    [providerModels],
  );
  const allFamilies = useMemo(
    () => Array.from(new Set(providerModels.map(getModelFamily))).sort((a, b) => a.localeCompare(b)),
    [providerModels],
  );
  const recommendedModel = useMemo(() => getRecommendedModel(providerModels), [providerModels]);

  const visibleModels = useMemo(
    () =>
      filterAndSortProviderModels({
        models: providerModels,
        filters: deferredFilters,
        isSelectedOnly: selectedOnlyActive,
        isModelEnabled: (modelKey) => Boolean(modelDrafts[modelKey]?.enabled),
      }),
    [deferredFilters, modelDrafts, providerModels, selectedOnlyActive],
  );
  const renderedVisibleModels = useDeferredValue(visibleModels);
  const visibleSelectedModelsCount = useMemo(
    () => renderedVisibleModels.filter((model) => modelDrafts[model.model]?.enabled).length,
    [modelDrafts, renderedVisibleModels],
  );
  const allVisibleModelsSelected =
    renderedVisibleModels.length > 0 && visibleSelectedModelsCount === renderedVisibleModels.length;
  const availableVisibleModelsCount = Math.max(renderedVisibleModels.length - visibleSelectedModelsCount, 0);

  const resetFilters = () => {
    setFilters((prev) => ({
      ...initialFilters,
      search: prev.search,
    }));
  };

  const activeFilters = useMemo(
    () =>
      getActiveModelFilterChips(filters, {
        minPrice: ux("models.filters.minPrice"),
        maxPrice: ux("models.filters.maxPrice"),
        minContext: ux("models.filters.minContext"),
        allInputModalities: ux("models.filters.allInputModalities"),
        allOutputModalities: ux("models.filters.allOutputModalities"),
        allFeatures: ux("models.filters.allFeatures"),
        allFamilies: ux("models.filters.allFamilies"),
      }),
    [filters, ux],
  );

  const hasInvalidPriceRange = useMemo(() => hasInvalidModelPriceRange(filters), [filters]);

  const compareModels = useMemo(() => sortComparedModels(comparedModels, compareSortBy), [compareSortBy, comparedModels]);

  const compareKpis = useMemo(() => getCompareKpis(compareModels), [compareModels]);

  const bestMarkers = useMemo(() => getBestCompareMarkers(compareModels), [compareModels]);

  const compareRows = useMemo(() => {
    const rows = [
      {
        section: "pricing",
        key: "input",
        label: ux("models.compare.rows.inputPrice"),
        values: compareModels.map((item) => item.public_pricing?.input_usd_per_1m),
        render: (value: unknown) => formatDisplayPrice(typeof value === "number" ? value : undefined, { currency: displayCurrency, scale: displayScale, usdToIrr, naLabel: ux("models.na") }),
      },
      {
        section: "pricing",
        key: "output",
        label: ux("models.compare.rows.outputPrice"),
        values: compareModels.map((item) => item.public_pricing?.output_usd_per_1m),
        render: (value: unknown) => formatDisplayPrice(typeof value === "number" ? value : undefined, { currency: displayCurrency, scale: displayScale, usdToIrr, naLabel: ux("models.na") }),
      },
      {
        section: "capacity",
        key: "context",
        label: ux("models.compare.rows.context"),
        values: compareModels.map((item) => item.context_window),
        render: (value: unknown) => (typeof value === "number" ? value.toLocaleString() : ux("models.na")),
      },
      {
        section: "modalities",
        key: "inputModality",
        label: ux("models.compare.rows.inputModality"),
        values: compareModels.map((item) => (item.input_modalities ?? []).slice().sort().join(", ")),
        render: (value: unknown) => (typeof value === "string" && value ? value : ux("models.na")),
      },
      {
        section: "modalities",
        key: "outputModality",
        label: ux("models.compare.rows.outputModality"),
        values: compareModels.map((item) => (item.output_modalities ?? []).slice().sort().join(", ")),
        render: (value: unknown) => (typeof value === "string" && value ? value : ux("models.na")),
      },
      {
        section: "capabilities",
        key: "features",
        label: ux("models.compare.rows.features"),
        values: compareModels.map((item) => getModelCapabilities(item).slice().sort().join(", ")),
        render: (value: unknown) => (typeof value === "string" && value ? value : ux("models.na")),
      },
    ] as const;

    return rows.map((row) => {
      const uniqueCount = new Set(row.values.map((value) => `${value ?? ""}`)).size;
      return { ...row, isDifferent: uniqueCount > 1 };
    });
  }, [compareModels, displayCurrency, displayScale, usdToIrr, ux]);

  const visibleCompareRows = useMemo(
    () => (compareDiffOnly ? compareRows.filter((row) => row.isDifferent) : compareRows),
    [compareDiffOnly, compareRows],
  );

  const removeComparedModel = useCallback((modelKey: string) => {
    setModelDrafts((prev) => {
      const draft = prev[modelKey];
      if (!draft) return prev;
      return { ...prev, [modelKey]: { ...draft, compared: false } };
    });
  }, [setModelDrafts]);

  const renderCompareValue = useCallback((rowKey: string, value: unknown, render: (value: unknown) => string): string => {
    if (!compareNormalizedPricing && (rowKey === "input" || rowKey === "output")) {
      if (typeof value !== "number") return render(value);
      return render(value / 1_000_000);
    }
    return render(value);
  }, [compareNormalizedPricing]);

  const openPricingEditor = useCallback((model: ProviderCatalogModel) => {
    const draft = modelDrafts[model.model] ?? buildFallbackDraft(model, []);
    setPricingEditor({
      open: true,
      modelKey: model.model,
      pricingMode: draft.pricingMode,
      inputPrice: draft.inputPrice,
      outputPrice: draft.outputPrice,
      error: "",
    });
  }, [modelDrafts]);

  const closePricingEditor = useCallback(() => {
    setPricingEditor((prev) => ({ ...prev, open: false, error: "" }));
  }, []);

  const applyPricingEditor = useCallback(() => {
    const inputNumber = toNumberOrUndefined(pricingEditor.inputPrice);
    const outputNumber = toNumberOrUndefined(pricingEditor.outputPrice);
    if (pricingEditor.pricingMode === "private") {
      if (pricingEditor.inputPrice.trim() && (typeof inputNumber !== "number" || inputNumber < 0)) {
        setPricingEditor((prev) => ({ ...prev, error: ux("models.validation.invalidInputPrice") }));
        return;
      }
      if (pricingEditor.outputPrice.trim() && (typeof outputNumber !== "number" || outputNumber < 0)) {
        setPricingEditor((prev) => ({ ...prev, error: ux("models.validation.invalidOutputPrice") }));
        return;
      }
    }

    setModelDrafts((prev) => {
      const current = prev[pricingEditor.modelKey];
      if (!current) return prev;
      return {
        ...prev,
        [pricingEditor.modelKey]: {
          ...current,
          enabled: true,
          pricingMode: pricingEditor.pricingMode,
          inputPrice: pricingEditor.inputPrice.trim(),
          outputPrice: pricingEditor.outputPrice.trim(),
        },
      };
    });
    setPricingEditor((prev) => ({ ...prev, open: false, error: "" }));
  }, [pricingEditor, setModelDrafts, ux]);

  const renderPricingEditorContent = (model: ProviderCatalogModel) => (
    <PopoverContent align="start" className="w-72 p-3">
      <div className="flex flex-col gap-3">
        <div className="min-w-0">
          <p className="text-sm font-medium">{ux("models.actions.editPricing")}</p>
          <p className="truncate text-xs text-muted-foreground" dir="ltr" title={model.display_name || model.model}>
            {model.display_name || model.model}
          </p>
        </div>
        <AppSelectField
          label={ux("models.columns.pricingMode")}
          size="sm"
          triggerClassName="min-w-0"
          value={pricingEditor.pricingMode}
          onValueChange={(value) =>
            setPricingEditor((prev) => ({
              ...prev,
              pricingMode: value as "public" | "private",
              error: "",
            }))
          }
          options={[
            { value: "public", label: ux("models.pricing.public") },
            { value: "private", label: ux("models.pricing.private") },
          ]}
        />
        {pricingEditor.pricingMode === "private" ? (
          <div className="grid gap-2 sm:grid-cols-2">
            <div className="flex flex-col gap-1.5">
              <Label>{ux("models.labels.input")}</Label>
              <Input
                className="h-8"
                value={pricingEditor.inputPrice}
                onChange={(event) => setPricingEditor((prev) => ({ ...prev, inputPrice: event.target.value, error: "" }))}
                placeholder={ux("models.placeholders.input")}
                dir="ltr"
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>{ux("models.labels.output")}</Label>
              <Input
                className="h-8"
                value={pricingEditor.outputPrice}
                onChange={(event) => setPricingEditor((prev) => ({ ...prev, outputPrice: event.target.value, error: "" }))}
                placeholder={ux("models.placeholders.output")}
                dir="ltr"
              />
            </div>
          </div>
        ) : (
          <div className="rounded-md border border-border/70 bg-muted/30 p-2 text-xs text-muted-foreground">
            <div className="font-medium text-foreground">{ux("models.catalogPricing")}</div>
            <div className="mt-1 flex items-center justify-between gap-2">
              <span>{ux("models.labels.input")}</span>
              <span dir="ltr">
                {formatDisplayPrice(model.public_pricing?.input_usd_per_1m, { currency: displayCurrency, scale: displayScale, usdToIrr, naLabel: ux("models.na") })}
              </span>
            </div>
            <div className="mt-1 flex items-center justify-between gap-2">
              <span>{ux("models.labels.output")}</span>
              <span dir="ltr">
                {formatDisplayPrice(model.public_pricing?.output_usd_per_1m, { currency: displayCurrency, scale: displayScale, usdToIrr, naLabel: ux("models.na") })}
              </span>
            </div>
          </div>
        )}
        {pricingEditor.error ? (
          <p className="text-xs text-destructive">{pricingEditor.error}</p>
        ) : null}
        <div className="flex justify-end gap-2">
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={closePricingEditor}
            >
              {ux("models.actions.cancel")}
            </Button>
          <Button type="button" size="sm" onClick={applyPricingEditor}>
            {ux("models.actions.apply")}
          </Button>
        </div>
      </div>
    </PopoverContent>
  );

  const toModelCardData = useCallback((model: ProviderCatalogModel, draft?: ModelDraft): ModelCardData => {
    const modes = getModelModes(model);
    const caps = getModelCapabilities(model);
    const selectionTypes = getModelSelectionTypes(model);
    const inputModalities = normalizeCapabilityList(model.input_modalities);
    const outputModalities = normalizeCapabilityList(model.output_modalities);
    const privateInputPrice = toNumberOrUndefined(draft?.inputPrice ?? "");
    const privateOutputPrice = toNumberOrUndefined(draft?.outputPrice ?? "");
    const inputPrice = draft?.pricingMode === "private" && typeof privateInputPrice === "number"
      ? privateInputPrice
      : model.public_pricing?.input_usd_per_1m;
    const outputPrice = draft?.pricingMode === "private" && typeof privateOutputPrice === "number"
      ? privateOutputPrice
      : model.public_pricing?.output_usd_per_1m;
    const categories: ModelCardData["category"] = [];
    if (modes.includes("embedding") || selectionTypes.includes("embedding")) categories.push("embedding");
    if (caps.some((c) => c.toLowerCase().includes("vision") || c.toLowerCase().includes("image")) || inputModalities.includes("image") || outputModalities.includes("image")) categories.push("vision");
    if (caps.some((c) => c.toLowerCase().includes("reason"))) categories.push("reasoning");
    if (caps.some((c) => c.toLowerCase().includes("code")) || inputModalities.includes("code") || outputModalities.includes("code")) categories.push("code");
    if (caps.some((c) => c.toLowerCase().includes("audio")) || inputModalities.includes("audio") || outputModalities.includes("audio")) categories.push("audio");
    if (inputModalities.length + outputModalities.length > 2) categories.push("multimodal");
    if (categories.length === 0) categories.push("chat");
    const modelFqn = buildModelFqn(model);

    return {
      id: model.model,
      name: model.display_name || model.model,
      provider: {
        id: model.provider,
        name: model.provider,
      },
      description: `${ux("models.columns.modalities")}: ${inputModalities.join(", ") || ux("models.na")} / ${outputModalities.join(", ") || ux("models.na")}`,
      category: categories,
      status: "stable",
      recommended: recommendedModel?.model === model.model,
      contextWindow: model.context_window,
      maxOutputTokens: undefined,
      pricing: {
        inputPerMillion: inputPrice,
        outputPerMillion: outputPrice,
      },
      latencyTier: "medium",
      qualityTier: "good",
      capabilities: caps,
      inputModalities,
      outputModalities,
      invocationModes: modes.length ? modes : selectionTypes,
      tradeoffs: [],
      metadata: {
        modelId: modelFqn,
        fqn: modelFqn,
        mode: modes.join(", ") || undefined,
      },
    };
  }, [buildModelFqn, recommendedModel?.model, ux]);

  const getSelectableOptions = useCallback((model: ProviderCatalogModel) => {
    const modes = getModelModes(model);
    if (modes.length > 1) return modes;
    const selectionTypes = getModelSelectionTypes(model);
    if (selectionTypes.length > 0) return selectionTypes;
    return modes;
  }, []);

  const setModelEnabled = useCallback(
    (model: ProviderCatalogModel, draft: ModelDraft, selectableOptions: string[], enabled: boolean) => {
      setModelDrafts((prev) => ({
        ...prev,
        [model.model]: {
          ...draft,
          enabled,
          selectedCapabilities: enabled ? selectableOptions : draft.selectedCapabilities,
        },
      }));
    },
    [setModelDrafts],
  );

  const setModelCompared = useCallback(
    (model: ProviderCatalogModel, draft: ModelDraft, compared: boolean) => {
      setModelDrafts((prev) => ({
        ...prev,
        [model.model]: {
          ...draft,
          compared,
        },
      }));
    },
    [setModelDrafts],
  );

  const handleVisibleSelectionChange = useCallback((models: ProviderCatalogModel[], checked: boolean) => {
    setModelDrafts((prev) => {
      if (!checked) return clearVisibleModelSelection(models, prev);
      const next = { ...prev };
      for (const model of models) {
        const selectableOptions = getSelectableOptions(model);
        const draft = next[model.model] ?? buildFallbackDraft(model, selectableOptions);
        next[model.model] = {
          ...draft,
          enabled: true,
          selectedCapabilities: selectableOptions,
        };
      }
      return next;
    });
  }, [getSelectableOptions, setModelDrafts]);

  useEffect(() => {
    let canceled = false;
    void (async () => {
      try {
        const payload = await requestJSON<DisplaySettingsResponse>("/api/provider-catalog/display-settings");
        if (canceled) return;
        const data = unwrapData<DisplaySettingsResponse["data"]>(payload) ?? payload.data;
        if (!data) return;
        if (data.currencyUnit === "TOMAN") setDisplayCurrency("TOMAN");
        else setDisplayCurrency("USD");
        if (data.priceScale === "PER_TOKEN" || data.priceScale === "PER_MILLION_TOKENS") {
          setDisplayScale(data.priceScale);
        }
        setUsdToIrr(Number(data.exchangeRate?.usdToIrr ?? 0));
      } catch {
        // Keep defaults when display settings are unavailable.
      }
    })();
    return () => {
      canceled = true;
    };
  }, []);

  if (props.variant === "compact") {
    return (
      <>
        <div className="sticky top-0 z-10 shrink-0 bg-card px-4 py-3 shadow-sm sm:px-5">
          <div className="flex flex-col gap-2.5">
            <div className="flex justify-end">
              <div className="relative w-full">
                <Search className="pointer-events-none absolute end-3 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
                <Input
                  value={filters.search}
                  onChange={(event) => setFilters((prev) => ({ ...prev, search: event.target.value }))}
                  placeholder={props.t("step3.searchPlaceholder")}
                  className="h-9 pe-9 text-start text-sm placeholder:text-start"
                />
              </div>
            </div>
            {recommendedModel ? (
              <div className="rounded-lg border border-primary/25 bg-primary/5 px-3 py-2 text-sm">
                <span className="font-medium">{ux("models.recommendation.title")}: </span>
                <bdi dir="ltr">{recommendedModel.display_name || recommendedModel.model}</bdi>
                <span className="text-muted-foreground"> — {ux("models.recommendation.reason")}</span>
              </div>
            ) : null}
            <div className="grid gap-2 sm:grid-cols-3">
              <AppSelectField
                ariaLabel={ux("models.filters.allFamilies")}
                size="sm"
                value={filters.familyFilter || ALL_FILTER_OPTIONS}
                onValueChange={(value) => setFilters((prev) => ({ ...prev, familyFilter: value === ALL_FILTER_OPTIONS ? "" : value }))}
                options={[
                  { value: ALL_FILTER_OPTIONS, label: ux("models.filters.allFamilies") },
                  ...allFamilies.map((item) => ({ value: item, label: item })),
                ]}
              />
              <AppSelectField
                ariaLabel={ux("models.filters.allFeatures")}
                size="sm"
                value={filters.featureFilter || ALL_FILTER_OPTIONS}
                onValueChange={(value) => setFilters((prev) => ({ ...prev, featureFilter: value === ALL_FILTER_OPTIONS ? "" : value }))}
                options={[
                  { value: ALL_FILTER_OPTIONS, label: ux("models.filters.allFeatures") },
                  ...allFeatures.map((item) => ({ value: item, label: formatCapabilityLabel(item, ux) })),
                ]}
              />
              <AppSelectField
                ariaLabel={ux("models.sort.priceAsc")}
                size="sm"
                value={filters.sortBy}
                onValueChange={(value) => setFilters((prev) => ({ ...prev, sortBy: value as ModelFilterState["sortBy"] }))}
                options={[
                  { value: "price_asc", label: ux("models.sort.priceAsc") },
                  { value: "price_desc", label: ux("models.sort.priceDesc") },
                  { value: "context_desc", label: ux("models.sort.contextDesc") },
                  { value: "context_asc", label: ux("models.sort.contextAsc") },
                ]}
              />
            </div>
            <div
              className="flex flex-col gap-2 px-1 py-1 text-start text-sm sm:flex-row sm:items-center sm:justify-between"
            >
              <div className="flex flex-wrap items-center gap-2">
                <Badge variant={enabledModelCount > 0 ? "default" : "outline"} className="px-2 py-0 text-xs">
                  {ux("integrations.wizard.selectedCountDetailed", {
                    count: enabledModelCount,
                    total: props.providerModels.length,
                  })}
                </Badge>
                <ToggleGroup
                  type="single"
                  value={mobileViewMode}
                  onValueChange={(value) => {
                    if (value === "table" || value === "card") setMobileViewMode(value);
                  }}
                  aria-label={ux("integrations.wizard.viewMode")}
                  className="inline-flex rounded-full border border-border/70 bg-muted/30 p-1"
                >
                  <ToggleGroupItem value="table" className="h-8 rounded-full px-3 text-xs data-[state=on]:bg-secondary">
                    {ux("integrations.wizard.tableView")}
                  </ToggleGroupItem>
                  <ToggleGroupItem value="card" className="h-8 rounded-full px-3 text-xs data-[state=on]:bg-secondary">
                    {ux("integrations.wizard.cardView")}
                  </ToggleGroupItem>
                </ToggleGroup>
              </div>
              <div className="flex flex-wrap items-center gap-3 text-sm">
                <label
                  htmlFor="compact-select-all-models"
                  className="inline-flex cursor-pointer flex-row items-center gap-1.5 text-foreground has-[:disabled]:pointer-events-none has-[:disabled]:opacity-50"
                >
                  <Checkbox
                    id="compact-select-all-models"
                    checked={allVisibleModelsSelected}
                    disabled={!props.canManage || renderedVisibleModels.length === 0}
                    onCheckedChange={(checked) => handleVisibleSelectionChange(renderedVisibleModels, checked === true)}
                  />
                  <span>{ux("integrations.wizard.selectAll")}</span>
                </label>
                <label
                  htmlFor="compact-selected-only-models"
                  className="inline-flex cursor-pointer flex-row items-center gap-1.5 text-foreground has-[:disabled]:pointer-events-none has-[:disabled]:opacity-50"
                >
                  <Checkbox
                    id="compact-selected-only-models"
                    checked={selectedOnlyActive}
                    disabled={enabledModelCount === 0}
                    onCheckedChange={(checked) => setShowSelectedOnly(checked === true)}
                  />
                  <span>{ux("integrations.wizard.selectedOnly")}</span>
                </label>
              </div>
            </div>
          </div>
        </div>

        <section className="flex h-full min-h-0 flex-col overflow-hidden bg-card">
          <div className="min-h-0 flex-1 overflow-y-auto px-4 pb-3 pt-3 sm:px-5">
            <div className="flex min-h-full min-w-0 flex-col gap-2 pe-2">
              {props.isLoadingModels ? (
                <>
                  {Array.from({ length: 8 }).map((_, index) => (
                    <Skeleton key={`compact-model-skeleton-${index}`} className="h-24 rounded-xl" />
                  ))}
                </>
              ) : renderedVisibleModels.length === 0 ? (
                <EmptyState variant="inline" density="compact" title={props.t("step3.empty")} />
              ) : mobileViewMode === "table" ? (
                <div className="overflow-hidden rounded-2xl border border-border/70 bg-card shadow-sm">
                  <div className="overflow-x-auto">
                    <table className="w-full table-fixed border-collapse text-start">
                      <colgroup>
                        <col className="w-12" />
                        <col />
                        <col className="w-20" />
                        <col className="w-20" />
                        <col className="w-20" />
                        <col className="hidden w-36 md:table-column" />
                      </colgroup>
                      <thead className="bg-muted/35 text-xs text-muted-foreground">
                        <tr>
                          <th className="px-2 py-2 text-center font-medium">{ux("common.enabled")}</th>
                          <th className="px-3 py-2 text-start font-medium">{ux("models.card.modelId")}</th>
                          <th className="px-3 py-2 text-center font-medium">{ux("models.card.context")}</th>
                          <th className="px-3 py-2 text-start font-medium">{ux("models.labels.input")}</th>
                          <th className="px-3 py-2 text-start font-medium">{ux("models.labels.output")}</th>
                          <th className="hidden px-3 py-2 text-start font-medium md:table-cell">{ux("models.card.capabilities")}</th>
                        </tr>
                      </thead>
                      <tbody>
                        {renderedVisibleModels.map((model) => (
                          <CompactTableRow
                            key={model.model}
                            canManage={props.canManage}
                            draft={props.modelDrafts[model.model]}
                            model={model}
                            selectableOptions={getSelectableOptions(model)}
                            setModelEnabled={setModelEnabled}
                            displayCurrency={displayCurrency}
                            displayScale={displayScale}
                            usdToIrr={usdToIrr}
                            ux={ux}
                            recommended={recommendedModel?.model === model.model}
                          />
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              ) : (
                <>
                  {renderedVisibleModels.map((model) => (
                    <ModelListItem
                      key={model.model}
                      account={props.account}
                      canManage={props.canManage}
                      compact
                      mobileDensity="detailed"
                      displayCurrency={displayCurrency}
                      displayScale={displayScale}
                      draft={props.modelDrafts[model.model]}
                      labels={modelCardLabels}
                      model={model}
                      openCurlGuide={openCurlGuide}
                      openPricingEditor={openPricingEditor}
                      closePricingEditor={closePricingEditor}
                      pricingEditor={pricingEditor}
                      renderPricingEditorContent={renderPricingEditorContent}
                      setModelCompared={setModelCompared}
                      setModelEnabled={setModelEnabled}
                      toModelCardData={toModelCardData}
                      ux={ux}
                      usdToIrr={usdToIrr}
                      getSelectableOptions={getSelectableOptions}
                    />
                  ))}
                </>
              )}
            </div>
          </div>
        </section>
      </>
    );
  }

  return (
    <section className="space-y-3 rounded-lg border p-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div>
          <h3 className="text-base font-semibold">{ux("tabs.models")}</h3>
          <p className="text-sm text-muted-foreground">{ux("models.description")}</p>
        </div>
        <Button type="button" onClick={props.onSave} disabled={!props.canManage || props.isSaving}>
          {props.isSaving ? props.t("saving") : props.t("saveSelectedModels")}
        </Button>
      </div>

      <div className="surface-elevated rounded-lg border p-3">
        <div className="flex flex-wrap items-center gap-2">
          <Badge variant="outline">{ux("models.summary.catalog", { count: props.providerModels.length })}</Badge>
          <Badge className="border-[var(--success-soft-border)] bg-[var(--success-soft)] text-[var(--success-soft-foreground)] hover:bg-[color-mix(in_oklab,var(--success-soft)_88%,black)]" variant="outline">
            {ux("models.summary.enabled", { count: enabledModelCount })}
          </Badge>
          <Badge variant="secondary">{ux("models.summary.available", { count: availableVisibleModelsCount })}</Badge>
          <Badge variant="secondary">{ux("models.summary.showing", { visible: renderedVisibleModels.length, total: props.providerModels.length })}</Badge>
        </div>
        <p className="mt-2 text-sm text-muted-foreground">{ux("models.catalogScopeHint")}</p>
      </div>

      <div className="space-y-3 rounded-lg border border-border/70 bg-muted/20 p-3">
        <div className="relative">
          <Search className="pointer-events-none absolute start-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={filters.search}
            onChange={(event) => setFilters((prev) => ({ ...prev, search: event.target.value }))}
            placeholder={props.t("step3.searchPlaceholder")}
            className="ps-9 pe-9"
          />
          {filters.search.trim() ? (
            <Button
              type="button"
              variant="ghost"
              size="icon"
              className="absolute end-1 top-1/2 h-8 w-8 -translate-y-1/2"
              onClick={() => setFilters((prev) => ({ ...prev, search: "" }))}
              aria-label={ux("actions.clear")}
              title={ux("actions.clear")}
            >
              <X className="h-4 w-4" />
            </Button>
          ) : null}
        </div>

        <div className="grid gap-2 md:grid-cols-2 xl:grid-cols-4">
          <Input
            value={filters.minPrice}
            onChange={(event) => setFilters((prev) => ({ ...prev, minPrice: event.target.value }))}
            placeholder={ux("models.filters.minPrice")}
            inputMode="decimal"
          />
          <Input
            value={filters.maxPrice}
            onChange={(event) => setFilters((prev) => ({ ...prev, maxPrice: event.target.value }))}
            placeholder={ux("models.filters.maxPrice")}
            inputMode="decimal"
          />
          <Input
            value={filters.minContext}
            onChange={(event) => setFilters((prev) => ({ ...prev, minContext: event.target.value }))}
            placeholder={ux("models.filters.minContext")}
            inputMode="numeric"
          />
          <AppSelectField
            ariaLabel={ux("models.filters.allInputModalities")}
            value={filters.inputModality || ALL_FILTER_OPTIONS}
            onValueChange={(value) => setFilters((prev) => ({ ...prev, inputModality: value === ALL_FILTER_OPTIONS ? "" : value }))}
            optionTextDir="ltr"
            options={[
              { value: ALL_FILTER_OPTIONS, label: ux("models.filters.allInputModalities") },
              ...allInputModalities.map((item) => ({ value: item, label: item })),
            ]}
          />
          <AppSelectField
            ariaLabel={ux("models.filters.allOutputModalities")}
            value={filters.outputModality || ALL_FILTER_OPTIONS}
            onValueChange={(value) => setFilters((prev) => ({ ...prev, outputModality: value === ALL_FILTER_OPTIONS ? "" : value }))}
            optionTextDir="ltr"
            options={[
              { value: ALL_FILTER_OPTIONS, label: ux("models.filters.allOutputModalities") },
              ...allOutputModalities.map((item) => ({ value: item, label: item })),
            ]}
          />
          <AppSelectField
            ariaLabel={ux("models.filters.allFeatures")}
            value={filters.featureFilter || ALL_FILTER_OPTIONS}
            onValueChange={(value) => setFilters((prev) => ({ ...prev, featureFilter: value === ALL_FILTER_OPTIONS ? "" : value }))}
            optionTextDir="ltr"
            options={[
              { value: ALL_FILTER_OPTIONS, label: ux("models.filters.allFeatures") },
              ...allFeatures.map((item) => ({ value: item, label: item })),
            ]}
          />
          <AppSelectField
            ariaLabel={ux("models.sort.priceAsc")}
            value={filters.sortBy}
            onValueChange={(value) => setFilters((prev) => ({ ...prev, sortBy: value as ModelFilterState["sortBy"] }))}
            options={[
              { value: "price_asc", label: ux("models.sort.priceAsc") },
              { value: "price_desc", label: ux("models.sort.priceDesc") },
              { value: "context_asc", label: ux("models.sort.contextAsc") },
              { value: "context_desc", label: ux("models.sort.contextDesc") },
            ]}
          />
          <AppSelectField
            ariaLabel={ux("models.displayUnit.currencyUsd")}
            value={displayCurrency}
            onValueChange={(value) => setDisplayCurrency(value as "USD" | "TOMAN")}
            options={[
              { value: "USD", label: ux("models.displayUnit.currencyUsd") },
              { value: "TOMAN", label: ux("models.displayUnit.currencyToman") },
            ]}
          />
          <AppSelectField
            ariaLabel={ux("models.displayUnit.perMillionTokens")}
            value={displayScale}
            onValueChange={(value) => setDisplayScale(value as "PER_MILLION_TOKENS" | "PER_TOKEN")}
            options={[
              { value: "PER_MILLION_TOKENS", label: ux("models.displayUnit.perMillionTokens") },
              { value: "PER_TOKEN", label: ux("models.displayUnit.perToken") },
            ]}
          />
        </div>

        {hasInvalidPriceRange ? (
          <p className="text-xs text-destructive">{ux("models.validation.invalidInputPrice")}</p>
        ) : null}

        {activeFilters.length > 0 ? (
          <div className="flex flex-wrap items-center gap-2">
            {activeFilters.map((item) => (
              <Badge key={item.key} variant="secondary" className="flex items-center gap-1 pe-1">
                <span className="max-w-56 truncate">{item.label}</span>
                <button
                  type="button"
                  className="rounded p-0.5 hover:bg-muted"
                  onClick={() => setFilters((prev) => ({ ...prev, [item.key]: "" }))}
                  aria-label={ux("actions.clear")}
                  title={ux("actions.clear")}
                >
                  <X className="h-3 w-3" />
                </button>
              </Badge>
            ))}
          </div>
        ) : null}
      </div>

      <div className="flex flex-wrap items-center gap-2">
        <Button
          type="button"
          variant="outline"
          onClick={() => handleVisibleSelectionChange(renderedVisibleModels, true)}
          disabled={!props.canManage || renderedVisibleModels.length === 0}
        >
          <CheckSquare className="h-4 w-4" />
          {props.t("step3.toggleSelectAll")}
        </Button>
        <Button
          type="button"
          variant="outline"
          onClick={() => handleVisibleSelectionChange(renderedVisibleModels, false)}
          disabled={!props.canManage || renderedVisibleModels.length === 0}
        >
          <X className="h-4 w-4" />
          {ux("actions.clear")}
        </Button>
        <Button type="button" variant="outline" onClick={resetFilters}>
          <RotateCcw className="h-4 w-4" />
          {ux("models.actions.clearFilters")}
        </Button>
        <Button type="button" variant="outline" disabled={comparedModels.length < 2} onClick={() => setCompareOpen(true)}>
          <GitCompare className="h-4 w-4" />
          {ux("models.actions.compare", { count: comparedModels.length })}
        </Button>
        <div className="ms-auto flex items-center gap-2 text-xs text-muted-foreground">
          <ArrowUpDown className="h-4 w-4" />
          <span>{renderedVisibleModels.length}</span>
        </div>
      </div>

      {props.isLoadingModels ? (
        <ModelCardsSkeleton />
      ) : renderedVisibleModels.length === 0 ? (
        <EmptyState variant="inline" density="compact" title={props.t("step3.empty")} />
      ) : (
        <div className="space-y-3">
          {renderedVisibleModels.map((model) => (
            <ModelListItem
              key={model.model}
              account={props.account}
              canManage={props.canManage}
              compact={false}
              mobileDensity="detailed"
              displayCurrency={displayCurrency}
              displayScale={displayScale}
              draft={props.modelDrafts[model.model]}
              labels={modelCardLabels}
              model={model}
              openCurlGuide={openCurlGuide}
              openPricingEditor={openPricingEditor}
              closePricingEditor={closePricingEditor}
              pricingEditor={pricingEditor}
              renderPricingEditorContent={renderPricingEditorContent}
              setModelCompared={setModelCompared}
              setModelEnabled={setModelEnabled}
              toModelCardData={toModelCardData}
              ux={ux}
              usdToIrr={usdToIrr}
              getSelectableOptions={getSelectableOptions}
            />
          ))}
        </div>
      )}
      <Dialog open={compareOpen} onOpenChange={setCompareOpen}>
        <DialogContent className="max-h-[90vh] max-w-7xl overflow-hidden p-0">
          <DialogHeader>
            <div className="space-y-1 px-6 pt-6">
              <DialogTitle>{ux("models.compare.title")}</DialogTitle>
              <DialogDescription>{ux("models.compare.description")}</DialogDescription>
            </div>
          </DialogHeader>
          <div className="space-y-4 overflow-y-auto px-6 pb-6">
            <div className="grid gap-2 sm:grid-cols-2 xl:grid-cols-4">
              <Card className="border-[var(--success-soft-border)] bg-[var(--success-soft)]">
                <CardHeader className="pb-2">
                  <CardDescription>{ux("models.compare.kpis.bestInput")}</CardDescription>
                </CardHeader>
                <CardContent className="text-sm font-medium">{compareKpis.cheapestInput ? compareKpis.cheapestInput.display_name || compareKpis.cheapestInput.model : ux("models.na")}</CardContent>
              </Card>
              <Card className="border-[var(--success-soft-border)] bg-[var(--success-soft)]">
                <CardHeader className="pb-2">
                  <CardDescription>{ux("models.compare.kpis.bestOutput")}</CardDescription>
                </CardHeader>
                <CardContent className="text-sm font-medium">{compareKpis.cheapestOutput ? compareKpis.cheapestOutput.display_name || compareKpis.cheapestOutput.model : ux("models.na")}</CardContent>
              </Card>
              <Card className="border-[var(--info-soft-border)] bg-[var(--info-soft)]">
                <CardHeader className="pb-2">
                  <CardDescription>{ux("models.compare.kpis.largestContext")}</CardDescription>
                </CardHeader>
                <CardContent className="text-sm font-medium">{compareKpis.largestContext ? compareKpis.largestContext.display_name || compareKpis.largestContext.model : ux("models.na")}</CardContent>
              </Card>
              <Card className="surface-primary-soft border">
                <CardHeader className="pb-2">
                  <CardDescription>{ux("models.compare.kpis.richestFeatures")}</CardDescription>
                </CardHeader>
                <CardContent className="text-sm font-medium">{compareKpis.mostFeatures ? compareKpis.mostFeatures.display_name || compareKpis.mostFeatures.model : ux("models.na")}</CardContent>
              </Card>
            </div>

            <div className="flex flex-wrap items-center gap-3 rounded-md border p-3">
              <AppSelectField
                id="compare-sort"
                label={ux("models.compare.sort.label")}
                value={compareSortBy}
                onValueChange={(value) => setCompareSortBy(value as typeof compareSortBy)}
                size="sm"
                className="w-auto"
                options={[
                  { value: "input_asc", label: ux("models.compare.sort.inputAsc") },
                  { value: "output_asc", label: ux("models.compare.sort.outputAsc") },
                  { value: "context_desc", label: ux("models.compare.sort.contextDesc") },
                  { value: "name_asc", label: ux("models.compare.sort.nameAsc") },
                ]}
              />
              <label className="inline-flex items-center gap-2 text-sm">
                <input type="checkbox" checked={compareDiffOnly} onChange={(event) => setCompareDiffOnly(event.target.checked)} />
                {ux("models.compare.toggles.differencesOnly")}
              </label>
              <label className="inline-flex items-center gap-2 text-sm">
                <input type="checkbox" checked={compareNormalizedPricing} onChange={(event) => setCompareNormalizedPricing(event.target.checked)} />
                {ux("models.compare.toggles.normalizedPricing")}
              </label>
            </div>

            <div className="hidden md:block overflow-x-auto rounded-md border">
              <div className="min-w-[980px]">
                <div
                  className="sticky top-0 z-10 grid border-b bg-background"
                  style={{ gridTemplateColumns: `240px repeat(${compareModels.length}, minmax(220px, 1fr))` }}
                >
                  <div className="sticky start-0 z-20 border-e bg-background p-3 text-sm font-semibold">{ux("models.compare.columns.attribute")}</div>
                  {compareModels.map((model) => (
                    <div key={`compare-head-${model.model}`} className="border-s px-3 py-3">
                      <div className="flex items-start justify-between gap-2">
                        <div className="min-w-0">
                          <p className="truncate text-sm font-semibold">{model.display_name || model.model}</p>
                          <p className="truncate text-xs text-muted-foreground">{model.provider}</p>
                        </div>
                        <Button type="button" size="icon" variant="ghost" onClick={() => removeComparedModel(model.model)} aria-label={ux("models.compare.actions.removeModel")}>
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </div>
                  ))}
                </div>

                {[
                  ["pricing", ux("models.compare.sections.pricing")],
                  ["capacity", ux("models.compare.sections.capacity")],
                  ["modalities", ux("models.compare.sections.modalities")],
                  ["capabilities", ux("models.compare.sections.capabilities")],
                ].map(([sectionKey, sectionLabel]) => (
                  <div key={sectionKey}>
                    <div
                      className="grid border-b bg-muted/40"
                      style={{ gridTemplateColumns: `240px repeat(${compareModels.length}, minmax(220px, 1fr))` }}
                    >
                      <div className="sticky start-0 z-[1] border-e bg-muted/50 p-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">{sectionLabel}</div>
                      {compareModels.map((model) => <div key={`${sectionKey}-${model.model}`} className="border-s p-2" />)}
                    </div>
                    {visibleCompareRows.filter((row) => row.section === sectionKey).map((row) => (
                      <div key={row.key} className="grid border-b" style={{ gridTemplateColumns: `240px repeat(${compareModels.length}, minmax(220px, 1fr))` }}>
                        <div className="sticky start-0 z-[1] border-e bg-background p-3 text-sm font-medium">{row.label}</div>
                        {row.values.map((value, index) => {
                          const model = compareModels[index];
                          const isBestInput = row.key === "input" && typeof value === "number" && value === bestMarkers.input;
                          const isBestOutput = row.key === "output" && typeof value === "number" && value === bestMarkers.output;
                          const isBestContext = row.key === "context" && typeof value === "number" && value === bestMarkers.context;
                          return (
                            <div
                              key={`${row.key}-${model.model}`}
                              className={cn(
                                "border-s p-3 text-sm",
                                row.isDifferent && "bg-[var(--warning-soft)]",
                                (isBestInput || isBestOutput || isBestContext) && "bg-[var(--success-soft)]",
                              )}
                            >
                              {row.key === "features" || row.key === "inputModality" || row.key === "outputModality" ? (
                                <div className="flex flex-wrap gap-1">
                                  {`${value ?? ""}`.split(",").map((item) => item.trim()).filter(Boolean).length > 0 ? (
                                    `${value ?? ""}`.split(",").map((item) => item.trim()).filter(Boolean).map((item) => (
                                      <Badge key={`${row.key}-${model.model}-${item}`} variant="secondary">{item}</Badge>
                                    ))
                                  ) : (
                                    <span className="text-muted-foreground">{ux("models.na")}</span>
                                  )}
                                </div>
                              ) : (
                                <span>
                                  {renderCompareValue(row.key, value, row.render)}
                                </span>
                              )}
                            </div>
                          );
                        })}
                      </div>
                    ))}
                  </div>
                ))}
              </div>
            </div>

            <div className="space-y-3 md:hidden">
              {compareModels.map((model) => (
                <Card key={`compare-mobile-${model.model}`}>
                  <CardHeader className="pb-3">
                    <div className="flex items-start justify-between gap-2">
                      <div className="min-w-0">
                        <CardTitle className="truncate text-sm">{model.display_name || model.model}</CardTitle>
                        <CardDescription className="truncate">{model.provider}</CardDescription>
                      </div>
                      <Button type="button" size="icon" variant="ghost" onClick={() => removeComparedModel(model.model)} aria-label={ux("models.compare.actions.removeModel")}>
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </CardHeader>
                  <CardContent className="space-y-2 text-sm">
                    {visibleCompareRows.map((row) => {
                      const rowIndex = compareModels.findIndex((item) => item.model === model.model);
                      const value = row.values[rowIndex];
                      return (
                        <div key={`mobile-${model.model}-${row.key}`} className="rounded border p-2">
                          <div className="text-xs text-muted-foreground">{row.label}</div>
                          <div className="mt-1">
                            {row.key === "features" || row.key === "inputModality" || row.key === "outputModality"
                              ? (`${value ?? ""}` || ux("models.na"))
                              : row.key === "input" || row.key === "output"
                                ? renderCompareValue(row.key, value, row.render)
                                : row.render(value)}
                          </div>
                        </div>
                      );
                    })}
                  </CardContent>
                </Card>
              ))}
            </div>
          </div>
        </DialogContent>
      </Dialog>
      <Dialog open={curlGuide !== null} onOpenChange={(open) => { if (!open) setCurlGuide(null); }}>
        <DialogContent className="flex max-h-[calc(100dvh-1rem)] w-[calc(100vw-1rem)] max-w-4xl flex-col gap-0 overflow-hidden p-0 sm:max-h-[calc(100dvh-2rem)] sm:w-[calc(100vw-2rem)]">
          <div className="shrink-0 border-b border-border/70 bg-[linear-gradient(135deg,color-mix(in_oklab,var(--primary)_10%,transparent),transparent_58%),linear-gradient(180deg,color-mix(in_oklab,var(--muted)_65%,transparent),transparent)] px-4 py-4 sm:px-6 sm:py-6">
            <DialogHeader className="space-y-3">
              <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div className="min-w-0 space-y-2">
                  <DialogTitle>{ux("models.curlGuide.title")}</DialogTitle>
                  <DialogDescription>{ux("models.curlGuide.description", { model: curlGuide?.modelName ?? "" })}</DialogDescription>
                </div>
                <Badge className="w-fit shrink-0 border border-primary/20 bg-primary/10 text-primary">
                  <Code2 className="me-1 h-3.5 w-3.5" />
                  {ux("models.curlGuide.readyBadge")}
                </Badge>
              </div>
            </DialogHeader>
            <div className="mt-4 grid min-w-0 gap-3 md:mt-5 md:grid-cols-[minmax(0,1.2fr)_minmax(0,0.8fr)]">
              <div className="min-w-0 rounded-lg border border-border/70 bg-background/80 p-3 shadow-sm backdrop-blur sm:p-4">
                <div className="text-xs font-medium text-muted-foreground">{ux("models.curlGuide.modelFqnLabel")}</div>
                <div dir="ltr" className="mt-2 break-all text-start font-mono text-sm text-foreground">{curlGuide?.modelFqn}</div>
              </div>
              <div className="min-w-0 rounded-lg border border-border/70 bg-background/80 p-3 shadow-sm backdrop-blur sm:p-4">
                <div className="text-xs font-medium text-muted-foreground">{ux("models.curlGuide.instructionsLabel")}</div>
                <p className="mt-2 text-sm leading-6 text-muted-foreground">{ux("models.curlGuide.instructions")}</p>
              </div>
            </div>
          </div>
          <div className="custom-scrollbar flex min-h-0 flex-1 flex-col gap-5 overflow-y-auto px-4 py-4 sm:px-6 sm:py-6">
            <Tabs value={curlGuideShell} onValueChange={(value) => setCurlGuideShell(value as CurlGuideShell)} className="min-w-0 space-y-4">
              <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
                <TabsList className="h-auto w-full justify-start overflow-x-auto rounded-lg bg-muted/70 p-1 md:w-auto">
                  <TabsTrigger value="powershell" className="min-w-28 sm:min-w-32">{ux("models.curlGuide.shells.powershell")}</TabsTrigger>
                  <TabsTrigger value="bash" className="min-w-20 sm:min-w-24">{ux("models.curlGuide.shells.bash")}</TabsTrigger>
                </TabsList>
                <Button
                  type="button"
                  variant="outline"
                  className="w-full md:w-auto"
                  onClick={() => {
                    const command = curlGuide?.commands[curlGuideShell];
                    if (!command) return;
                    void copyTextToClipboard(command);
                    props.setBanner({ kind: "success", message: ux("messages.curlCopied") });
                  }}
                >
                  <Copy className="me-2 h-4 w-4" />
                  {ux("models.curlGuide.copyCommand")}
                </Button>
              </div>
              <TabsContent value="powershell">
                <CodeExampleBlock
                  code={curlGuide?.commands.powershell ?? ""}
                  language={CURL_GUIDE_LANGUAGE.powershell}
                  shellLabel={ux("models.curlGuide.shells.powershell")}
                />
              </TabsContent>
              <TabsContent value="bash">
                <CodeExampleBlock
                  code={curlGuide?.commands.bash ?? ""}
                  language={CURL_GUIDE_LANGUAGE.bash}
                  shellLabel={ux("models.curlGuide.shells.bash")}
                />
              </TabsContent>
            </Tabs>
            {curlGuide?.sampleResponse ? (
              <div className="flex flex-col gap-3">
                <div className="text-start text-sm font-medium text-foreground">{ux("models.curlGuide.sampleResponse")}</div>
                <CodeExampleBlock
                  code={curlGuide.sampleResponse}
                  language="json"
                  shellLabel="JSON"
                />
              </div>
            ) : null}
          </div>
          <DialogFooter className="shrink-0 border-t border-border/70 bg-muted/20 px-4 py-3 sm:px-6 sm:py-4">
            <Button type="button" onClick={() => setCurlGuide(null)}>
              {ux("models.curlGuide.close")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </section>
  );
}, (prev, next) => (
  prev.account === next.account &&
  prev.canManage === next.canManage &&
  prev.isLoadingModels === next.isLoadingModels &&
  prev.isSaving === next.isSaving &&
  prev.modelDrafts === next.modelDrafts &&
  prev.onSave === next.onSave &&
  prev.providerModels === next.providerModels &&
  prev.setBanner === next.setBanner &&
  prev.setModelDrafts === next.setModelDrafts &&
  prev.t === next.t &&
  prev.variant === next.variant
));


