"use client";

import type React from "react";
import { useCallback, useEffect, useMemo, useState } from "react";
import { ArrowUpDown, CheckSquare, Code2, Copy, GitCompare, RotateCcw, Search, SlidersHorizontal, Trash2, X } from "lucide-react";
import { useTranslations } from "next-intl";

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
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { Switch } from "@/components/ui/switch";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ModelCard, type ModelCardData } from "@/app/components/model-card";
import {
  clearVisibleModelSelection,
  toggleVisibleModelSelection,
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

export function ModelsEditor(props: {
  t: ReturnType<typeof useTranslations>;
  account: TenantProviderAccount | null;
  setBanner: React.Dispatch<React.SetStateAction<Banner | null>>;
  filteredProviderModels: ProviderCatalogModel[];
  modelDrafts: Record<string, ModelDraft>;
  setModelDrafts: React.Dispatch<React.SetStateAction<Record<string, ModelDraft>>>;
  modelSearchQuery: string;
  setModelSearchQuery: (value: string) => void;
  isSaving: boolean;
  isLoadingModels: boolean;
  canManage: boolean;
  onSave: () => void;
  variant?: "default" | "compact";
}) {
  const ux = useTranslations("providerAccounts.ux");
  const initialFilters = useMemo(() => createInitialModelFilters(props.modelSearchQuery), [props.modelSearchQuery]);
  const [filters, setFilters] = useState<ModelFilterState>(() => createInitialModelFilters(props.modelSearchQuery));
  const [compareOpen, setCompareOpen] = useState(false);
  const [compareSortBy, setCompareSortBy] = useState<CompareSortMode>("input_asc");
  const [compareDiffOnly, setCompareDiffOnly] = useState(false);
  const [compareNormalizedPricing, setCompareNormalizedPricing] = useState(true);
  const [displayCurrency, setDisplayCurrency] = useState<"USD" | "TOMAN">("USD");
  const [displayScale, setDisplayScale] = useState<"PER_MILLION_TOKENS" | "PER_TOKEN">("PER_MILLION_TOKENS");
  const [usdToIrr, setUsdToIrr] = useState(0);
  const [curlGuide, setCurlGuide] = useState<CurlGuide | null>(null);
  const [curlGuideShell, setCurlGuideShell] = useState<CurlGuideShell>("powershell");
  const [autoEnableNewModels, setAutoEnableNewModels] = useState(true);
  const [showSelectedOnly, setShowSelectedOnly] = useState(false);
  const [pricingEditor, setPricingEditor] = useState<{
    open: boolean;
    modelKey: string;
    pricingMode: "public" | "private";
    inputPrice: string;
    outputPrice: string;
    error: string;
  }>({
    open: false,
    modelKey: "",
    pricingMode: "public",
    inputPrice: "",
    outputPrice: "",
    error: "",
  });
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
    () => props.filteredProviderModels.filter((model) => props.modelDrafts[model.model]?.compared),
    [props.filteredProviderModels, props.modelDrafts],
  );
  const enabledFilteredModelsCount = useMemo(
    () => props.filteredProviderModels.filter((model) => props.modelDrafts[model.model]?.enabled).length,
    [props.filteredProviderModels, props.modelDrafts],
  );
  const availableFilteredModelsCount = Math.max(props.filteredProviderModels.length - enabledFilteredModelsCount, 0);
  const selectedOnlyActive = showSelectedOnly && enabledFilteredModelsCount > 0;

  const allInputModalities = useMemo(
    () => Array.from(new Set(props.filteredProviderModels.flatMap((model) => model.input_modalities ?? []))),
    [props.filteredProviderModels],
  );
  const allOutputModalities = useMemo(
    () => Array.from(new Set(props.filteredProviderModels.flatMap((model) => model.output_modalities ?? []))),
    [props.filteredProviderModels],
  );
  const allFeatures = useMemo(
    () => Array.from(new Set(props.filteredProviderModels.flatMap((model) => getModelCapabilities(model)))),
    [props.filteredProviderModels],
  );

  const visibleModels = useMemo(
    () =>
      filterAndSortProviderModels({
        models: props.filteredProviderModels,
        filters,
        isSelectedOnly: selectedOnlyActive,
        isModelEnabled: (modelKey) => Boolean(props.modelDrafts[modelKey]?.enabled),
      }),
    [filters, props.filteredProviderModels, props.modelDrafts, selectedOnlyActive],
  );
  const visibleSelectedModelsCount = useMemo(
    () => visibleModels.filter((model) => props.modelDrafts[model.model]?.enabled).length,
    [props.modelDrafts, visibleModels],
  );
  const allVisibleModelsSelected = visibleModels.length > 0 && visibleSelectedModelsCount === visibleModels.length;

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

  const removeComparedModel = (modelKey: string) => {
    props.setModelDrafts((prev) => {
      const draft = prev[modelKey];
      if (!draft) return prev;
      return { ...prev, [modelKey]: { ...draft, compared: false } };
    });
  };

  const renderCompareValue = (rowKey: string, value: unknown, render: (value: unknown) => string): string => {
    if (!compareNormalizedPricing && (rowKey === "input" || rowKey === "output")) {
      if (typeof value !== "number") return render(value);
      return render(value / 1_000_000);
    }
    return render(value);
  };

  useEffect(() => {
    const timer = setTimeout(() => {
      props.setModelSearchQuery(filters.search);
    }, 250);
    return () => clearTimeout(timer);
  }, [filters.search, props]);

  const openPricingEditor = (model: ProviderCatalogModel) => {
    const draft = props.modelDrafts[model.model] ?? {
      enabled: false,
      compared: false,
      selectedCapabilities: [],
      pricingMode: "public" as const,
      inputPrice: model.public_pricing?.input_usd_per_1m?.toString() ?? "",
      outputPrice: model.public_pricing?.output_usd_per_1m?.toString() ?? "",
    };
    setPricingEditor({
      open: true,
      modelKey: model.model,
      pricingMode: draft.pricingMode,
      inputPrice: draft.inputPrice,
      outputPrice: draft.outputPrice,
      error: "",
    });
  };

  const applyPricingEditor = () => {
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

    props.setModelDrafts((prev) => {
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
  };

  const renderPricingEditorContent = (model: ProviderCatalogModel) => (
    <PopoverContent align="start" className="w-72 p-3">
      <div className="flex flex-col gap-3">
        <div className="min-w-0">
          <p className="text-sm font-medium">{ux("models.actions.editPricing")}</p>
          <p className="truncate text-xs text-muted-foreground" dir="ltr" title={model.display_name || model.model}>
            {model.display_name || model.model}
          </p>
        </div>
        <div className="flex flex-col gap-1.5">
          <Label>{ux("models.columns.pricingMode")}</Label>
          <Select
            value={pricingEditor.pricingMode}
            onValueChange={(value) =>
              setPricingEditor((prev) => ({
                ...prev,
                pricingMode: value as "public" | "private",
                error: "",
              }))
            }
          >
            <SelectTrigger className="h-8">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                <SelectItem value="public">{ux("models.pricing.public")}</SelectItem>
                <SelectItem value="private">{ux("models.pricing.private")}</SelectItem>
              </SelectGroup>
            </SelectContent>
          </Select>
        </div>
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
            onClick={() => setPricingEditor((prev) => ({ ...prev, open: false, error: "" }))}
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

  const toModelCardData = (model: ProviderCatalogModel, draft?: ModelDraft): ModelCardData => {
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
  };

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
      <section className="flex flex-col gap-5">
        <div className="min-w-0 overflow-hidden rounded-lg border border-border bg-background p-4">
          <div className="flex flex-col gap-4">
            <div dir="rtl" className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
              <div className="min-w-0 text-right">
                <h3 dir="rtl" className="text-base font-semibold">
                  {ux("integrations.wizard.modelsTitle", { count: props.filteredProviderModels.length })}
                </h3>
                <p className="mt-1 text-sm text-muted-foreground">
                  {ux("integrations.wizard.selectedCount", { count: enabledFilteredModelsCount })}
                </p>
              </div>
              <div className="relative w-full lg:max-w-md">
                <Search className="pointer-events-none absolute end-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  dir="rtl"
                  value={filters.search}
                  onChange={(event) => setFilters((prev) => ({ ...prev, search: event.target.value }))}
                  placeholder={props.t("step3.searchPlaceholder")}
                  className="h-10 pe-9 text-right placeholder:text-right"
                />
              </div>
            </div>
            <div
              dir="rtl"
              className="flex flex-col gap-3 rounded-md border border-border/70 bg-muted/20 p-3 text-right sm:flex-row sm:items-center sm:justify-between"
            >
              <div className="flex flex-wrap items-center gap-2">
                <Badge variant="secondary">
                  {ux("models.summary.showing", { visible: visibleModels.length, total: props.filteredProviderModels.length })}
                </Badge>
                <Badge variant={enabledFilteredModelsCount > 0 ? "default" : "outline"}>
                  {ux("integrations.wizard.selectedCount", { count: enabledFilteredModelsCount })}
                </Badge>
              </div>
              <div className="flex flex-wrap items-center gap-4 text-sm">
                <label
                  htmlFor="compact-select-all-models"
                  className="inline-flex cursor-pointer flex-row items-center gap-2 text-foreground has-[:disabled]:pointer-events-none has-[:disabled]:opacity-50"
                >
                  <Checkbox
                    id="compact-select-all-models"
                    checked={allVisibleModelsSelected}
                    disabled={!props.canManage || visibleModels.length === 0}
                    onCheckedChange={(checked) =>
                      props.setModelDrafts((prev) =>
                        checked === true
                          ? toggleVisibleModelSelection(visibleModels, prev)
                          : clearVisibleModelSelection(visibleModels, prev),
                      )
                    }
                  />
                  <span>{ux("integrations.wizard.selectAll")}</span>
                </label>
                <label
                  htmlFor="compact-selected-only-models"
                  className="inline-flex cursor-pointer flex-row items-center gap-2 text-foreground has-[:disabled]:pointer-events-none has-[:disabled]:opacity-50"
                >
                  <Checkbox
                    id="compact-selected-only-models"
                    checked={selectedOnlyActive}
                    disabled={enabledFilteredModelsCount === 0}
                    onCheckedChange={(checked) => setShowSelectedOnly(checked === true)}
                  />
                  <span>{ux("integrations.wizard.selectedOnly")}</span>
                </label>
              </div>
            </div>
          </div>

          {props.isLoadingModels ? (
            <div className="mt-4 grid min-w-0 gap-3">
              {Array.from({ length: 8 }).map((_, index) => (
                <Skeleton key={`compact-model-skeleton-${index}`} className="h-24 rounded-xl" />
              ))}
            </div>
          ) : visibleModels.length === 0 ? (
            <p className="mt-4 text-sm text-muted-foreground">{props.t("step3.empty")}</p>
          ) : (
            <div className="mt-4 grid min-w-0 gap-3">
              {visibleModels.map((model) => {
                const availableCapabilities = getModelSelectionTypes(model);
                const draft = props.modelDrafts[model.model] ?? {
                  enabled: false,
                  compared: false,
                  selectedCapabilities: availableCapabilities.length === 1 ? availableCapabilities : [],
                  pricingMode: "public" as const,
                  inputPrice: model.public_pricing?.input_usd_per_1m?.toString() ?? "",
                  outputPrice: model.public_pricing?.output_usd_per_1m?.toString() ?? "",
                };

                return (
                  <Popover
                    key={model.model}
                    open={pricingEditor.open && pricingEditor.modelKey === model.model}
                    onOpenChange={(open) => {
                      if (!open) {
                        setPricingEditor((prev) => ({ ...prev, open: false, error: "" }));
                      } else {
                        openPricingEditor(model);
                      }
                    }}
                  >
                    <div className="min-w-0">
                      <ModelCard
                        model={toModelCardData(model, draft)}
                        detailsHref={buildModelDiscoveryDetailsHref(model)}
                        pricingDisplay={{
                          currency: displayCurrency,
                          scale: displayScale,
                          usdToIrr,
                        }}
                        isSelected={draft.enabled}
                        onSelectedChange={(checked) =>
                          props.setModelDrafts((prev) => ({
                            ...prev,
                            [model.model]: {
                              ...draft,
                              enabled: checked,
                              selectedCapabilities:
                                checked && draft.selectedCapabilities.length === 0 && availableCapabilities.length === 1
                                  ? availableCapabilities
                                  : draft.selectedCapabilities,
                            },
                          }))
                        }
                        selectedCheckboxLabel={ux("common.enabled")}
                        stateBadge={
                          draft.enabled
                            ? { label: ux("models.state.enabledOnAccount"), tone: "active" as const }
                            : { label: ux("models.state.availableToAdd"), tone: "inactive" as const }
                        }
                        pricingModeBadge={draft.pricingMode === "private" ? ux("models.pricing.private") : ux("models.pricing.public")}
                        hasCustomPricing={draft.pricingMode === "private"}
                        className={draft.enabled ? "border-emerald-500/70" : undefined}
                        pricingAction={
                          <PopoverTrigger asChild>
                            <Button
                              type="button"
                              variant="outline"
                              size="icon"
                              className="size-7"
                              disabled={!props.canManage}
                              aria-label={`${ux("models.actions.editPricing")} ${model.display_name || model.model}`}
                              onClick={() => openPricingEditor(model)}
                            >
                              <SlidersHorizontal className="h-3.5 w-3.5" />
                            </Button>
                          </PopoverTrigger>
                        }
                        footerActions={
                          <Button type="button" variant="outline" size="sm" className="gap-1" onClick={() => openCurlGuide(model)}>
                            <Code2 className="h-3.5 w-3.5" />
                            {ux("models.card.curlGuide")}
                          </Button>
                        }
                        labels={modelCardLabels}
                      />
                    </div>
                    {renderPricingEditorContent(model)}
                  </Popover>
                );
              })}
            </div>
          )}
        </div>

        <div className="flex items-start gap-3 rounded-xl border border-border bg-background p-4">
          <Switch checked={autoEnableNewModels} onCheckedChange={setAutoEnableNewModels} aria-label={ux("integrations.wizard.autoEnableTitle")} />
          <div className="space-y-1">
            <p className="text-sm font-medium">{ux("integrations.wizard.autoEnableTitle")}</p>
            <p className="text-sm text-muted-foreground">{ux("integrations.wizard.autoEnableDescription")}</p>
          </div>
        </div>
      </section>
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

      <div className="rounded-lg border border-border/70 bg-gradient-to-l from-background to-emerald-500/[0.05] p-3">
        <div className="flex flex-wrap items-center gap-2">
          <Badge variant="outline">{ux("models.summary.catalog", { count: props.filteredProviderModels.length })}</Badge>
          <Badge className="border-emerald-500/30 bg-emerald-500/15 text-emerald-700 hover:bg-emerald-500/20 dark:text-emerald-300" variant="outline">
            {ux("models.summary.enabled", { count: enabledFilteredModelsCount })}
          </Badge>
          <Badge variant="secondary">{ux("models.summary.available", { count: availableFilteredModelsCount })}</Badge>
          <Badge variant="secondary">{ux("models.summary.showing", { visible: visibleModels.length, total: props.filteredProviderModels.length })}</Badge>
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
          <select className="h-10 rounded-md border border-input bg-background px-3 text-sm" value={filters.inputModality} onChange={(event) => setFilters((prev) => ({ ...prev, inputModality: event.target.value }))}>
          <option value="">{ux("models.filters.allInputModalities")}</option>
          {allInputModalities.map((item) => <option key={item} value={item}>{item}</option>)}
          </select>
          <select className="h-10 rounded-md border border-input bg-background px-3 text-sm" value={filters.outputModality} onChange={(event) => setFilters((prev) => ({ ...prev, outputModality: event.target.value }))}>
          <option value="">{ux("models.filters.allOutputModalities")}</option>
          {allOutputModalities.map((item) => <option key={item} value={item}>{item}</option>)}
          </select>
          <select className="h-10 rounded-md border border-input bg-background px-3 text-sm" value={filters.featureFilter} onChange={(event) => setFilters((prev) => ({ ...prev, featureFilter: event.target.value }))}>
          <option value="">{ux("models.filters.allFeatures")}</option>
          {allFeatures.map((item) => <option key={item} value={item}>{item}</option>)}
          </select>
          <select className="h-10 rounded-md border border-input bg-background px-3 text-sm" value={filters.sortBy} onChange={(event) => setFilters((prev) => ({ ...prev, sortBy: event.target.value as ModelFilterState["sortBy"] }))}>
          <option value="price_asc">{ux("models.sort.priceAsc")}</option>
          <option value="price_desc">{ux("models.sort.priceDesc")}</option>
          <option value="context_asc">{ux("models.sort.contextAsc")}</option>
          <option value="context_desc">{ux("models.sort.contextDesc")}</option>
          </select>
          <select className="h-10 rounded-md border border-input bg-background px-3 text-sm" value={displayCurrency} onChange={(event) => setDisplayCurrency(event.target.value as "USD" | "TOMAN")}>
          <option value="USD">{ux("models.displayUnit.currencyUsd")}</option>
          <option value="TOMAN">{ux("models.displayUnit.currencyToman")}</option>
          </select>
          <select className="h-10 rounded-md border border-input bg-background px-3 text-sm" value={displayScale} onChange={(event) => setDisplayScale(event.target.value as "PER_MILLION_TOKENS" | "PER_TOKEN")}>
          <option value="PER_MILLION_TOKENS">{ux("models.displayUnit.perMillionTokens")}</option>
          <option value="PER_TOKEN">{ux("models.displayUnit.perToken")}</option>
          </select>
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
          onClick={() => {
            props.setModelDrafts((prev) => toggleVisibleModelSelection(visibleModels, prev));
          }}
          disabled={!props.canManage || visibleModels.length === 0}
        >
          <CheckSquare className="h-4 w-4" />
          {props.t("step3.toggleSelectAll")}
        </Button>
        <Button
          type="button"
          variant="outline"
          onClick={() => {
            props.setModelDrafts((prev) => clearVisibleModelSelection(visibleModels, prev));
          }}
          disabled={!props.canManage || visibleModels.length === 0}
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
          <span>{visibleModels.length}</span>
        </div>
      </div>

      {props.isLoadingModels ? (
        <ModelCardsSkeleton />
      ) : visibleModels.length === 0 ? (
        <p className="text-sm text-muted-foreground">{props.t("step3.empty")}</p>
      ) : (
        <div className="space-y-3">
          {visibleModels.map((model) => {
            const availableCapabilities = getModelSelectionTypes(model);
            const draft = props.modelDrafts[model.model] ?? {
              enabled: false,
              compared: false,
              selectedCapabilities: availableCapabilities.length === 1 ? availableCapabilities : [],
              pricingMode: "public" as const,
              inputPrice: model.public_pricing?.input_usd_per_1m?.toString() ?? "",
              outputPrice: model.public_pricing?.output_usd_per_1m?.toString() ?? "",
            };
            return (
              <Popover
                key={model.model}
                  open={pricingEditor.open && pricingEditor.modelKey === model.model}
                  onOpenChange={(open) => {
                    if (!open) {
                      setPricingEditor((prev) => ({ ...prev, open: false, error: "" }));
                    } else {
                      openPricingEditor(model);
                    }
                  }}
                >
                  <div className="min-w-0 space-y-2">
                    <ModelCard
                      model={toModelCardData(model, draft)}
                      detailsHref={buildModelDiscoveryDetailsHref(model)}
                      pricingDisplay={{
                        currency: displayCurrency,
                        scale: displayScale,
                        usdToIrr,
                      }}
                      isSelected={draft.enabled}
                      stateBadge={
                        draft.enabled
                          ? { label: ux("models.state.enabledOnAccount"), tone: "active" as const }
                          : { label: ux("models.state.availableToAdd"), tone: "inactive" as const }
                      }
                      onSelectedChange={(checked) =>
                        props.setModelDrafts((prev) => ({
                          ...prev,
                          [model.model]: {
                            ...draft,
                            enabled: checked,
                            selectedCapabilities:
                              checked && draft.selectedCapabilities.length === 0 && availableCapabilities.length === 1
                                ? availableCapabilities
                                : draft.selectedCapabilities,
                          },
                        }))
                      }
                      selectedCheckboxLabel={ux("common.enabled")}
                      isCompared={draft.compared}
                      onComparedChange={(checked) =>
                        props.setModelDrafts((prev) => ({
                          ...prev,
                          [model.model]: { ...draft, compared: checked },
                        }))
                      }
                      compareCheckboxLabel={ux("models.actions.compareToggle")}
                      pricingModeBadge={draft.pricingMode === "private" ? ux("models.pricing.private") : ux("models.pricing.public")}
                      hasCustomPricing={draft.pricingMode === "private"}
                      className={draft.enabled ? "border-emerald-500/70" : undefined}
                      footerActions={
                        <Button type="button" variant="outline" size="sm" className="gap-1" onClick={() => openCurlGuide(model)}>
                          <Code2 className="h-3.5 w-3.5" />
                          {ux("models.card.curlGuide")}
                        </Button>
                      }
                      pricingAction={
                        <PopoverTrigger asChild>
                          <Button
                            type="button"
                            variant="outline"
                            size="icon"
                            className="size-7"
                            disabled={!props.canManage}
                            aria-label={`${ux("models.actions.editPricing")} ${model.display_name || model.model}`}
                            onClick={() => openPricingEditor(model)}
                          >
                            <SlidersHorizontal className="h-3.5 w-3.5" />
                          </Button>
                        </PopoverTrigger>
                      }
                      labels={modelCardLabels}
                    />
                    {draft.enabled && availableCapabilities.length > 0 ? (
                      <div className="rounded-xl border border-border/70 bg-muted/20 p-3">
                        <div className="flex items-center justify-between gap-2">
                          <p className="text-sm font-medium">{ux("models.columns.capabilities")}</p>
                          <Badge variant="secondary">{draft.selectedCapabilities.length}/{availableCapabilities.length}</Badge>
                        </div>
                        <div className="mt-2 flex flex-wrap gap-2">
                          {availableCapabilities.map((capability) => {
                            const checked = draft.selectedCapabilities.includes(capability);
                            return (
                              <label
                                key={`${model.model}-${capability}`}
                                className={cn(
                                  "inline-flex cursor-pointer items-center gap-2 rounded-full border px-3 py-1.5 text-sm",
                                  checked ? "border-emerald-500 bg-emerald-500/10 text-emerald-700" : "border-border bg-background",
                                )}
                              >
                                <input
                                  type="checkbox"
                                  className="h-4 w-4"
                                  checked={checked}
                                  onChange={(event) =>
                                    props.setModelDrafts((prev) => {
                                      const current = prev[model.model] ?? draft;
                                      const nextCapabilities = event.target.checked
                                        ? [...current.selectedCapabilities, capability]
                                        : current.selectedCapabilities.filter((item) => item !== capability);
                                      return {
                                        ...prev,
                                        [model.model]: {
                                          ...current,
                                          selectedCapabilities: normalizeCapabilityList(nextCapabilities),
                                        },
                                      };
                                    })
                                  }
                                />
                                <span>{capability}</span>
                              </label>
                            );
                          })}
                        </div>
                        {draft.selectedCapabilities.length === 0 ? (
                          <p className="mt-2 text-xs text-destructive">{ux("messages.modelTypeRequired")}</p>
                        ) : null}
                      </div>
                    ) : null}
                  </div>
                  {renderPricingEditorContent(model)}
                </Popover>
            );
          })}
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
              <Card className="border-emerald-500/30 bg-emerald-500/5">
                <CardHeader className="pb-2">
                  <CardDescription>{ux("models.compare.kpis.bestInput")}</CardDescription>
                </CardHeader>
                <CardContent className="text-sm font-medium">{compareKpis.cheapestInput ? compareKpis.cheapestInput.display_name || compareKpis.cheapestInput.model : ux("models.na")}</CardContent>
              </Card>
              <Card className="border-emerald-500/30 bg-emerald-500/5">
                <CardHeader className="pb-2">
                  <CardDescription>{ux("models.compare.kpis.bestOutput")}</CardDescription>
                </CardHeader>
                <CardContent className="text-sm font-medium">{compareKpis.cheapestOutput ? compareKpis.cheapestOutput.display_name || compareKpis.cheapestOutput.model : ux("models.na")}</CardContent>
              </Card>
              <Card className="border-sky-500/30 bg-sky-500/5">
                <CardHeader className="pb-2">
                  <CardDescription>{ux("models.compare.kpis.largestContext")}</CardDescription>
                </CardHeader>
                <CardContent className="text-sm font-medium">{compareKpis.largestContext ? compareKpis.largestContext.display_name || compareKpis.largestContext.model : ux("models.na")}</CardContent>
              </Card>
              <Card className="border-violet-500/30 bg-violet-500/5">
                <CardHeader className="pb-2">
                  <CardDescription>{ux("models.compare.kpis.richestFeatures")}</CardDescription>
                </CardHeader>
                <CardContent className="text-sm font-medium">{compareKpis.mostFeatures ? compareKpis.mostFeatures.display_name || compareKpis.mostFeatures.model : ux("models.na")}</CardContent>
              </Card>
            </div>

            <div className="flex flex-wrap items-center gap-3 rounded-md border p-3">
              <Label htmlFor="compare-sort">{ux("models.compare.sort.label")}</Label>
              <select
                id="compare-sort"
                className="h-9 rounded-md border border-input bg-background px-2 text-sm"
                value={compareSortBy}
                onChange={(event) => setCompareSortBy(event.target.value as typeof compareSortBy)}
              >
                <option value="input_asc">{ux("models.compare.sort.inputAsc")}</option>
                <option value="output_asc">{ux("models.compare.sort.outputAsc")}</option>
                <option value="context_desc">{ux("models.compare.sort.contextDesc")}</option>
                <option value="name_asc">{ux("models.compare.sort.nameAsc")}</option>
              </select>
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
                              className={`border-s p-3 text-sm ${row.isDifferent ? "bg-amber-50/40 dark:bg-amber-950/10" : ""} ${(isBestInput || isBestOutput || isBestContext) ? "bg-emerald-50 dark:bg-emerald-950/20" : ""}`}
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
                <div dir="ltr" className="mt-2 break-all text-left font-mono text-sm text-foreground">{curlGuide?.modelFqn}</div>
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
                    void navigator.clipboard.writeText(command);
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
}


