"use client";

import type React from "react";
import { CopyAction } from "@/components/app-ui/CopyAction";
import { TablePagination } from "@/components/app-ui/TablePagination";
import { memo, useCallback, useDeferredValue, useEffect, useMemo, useState } from "react";
import { Code2, ExternalLink, LayoutGrid, RotateCcw, Search, Settings, SlidersHorizontal, Table2, X } from "lucide-react";
import { useLocale, useTranslations } from "next-intl";

import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { showFeedbackToast } from "@/components/app-ui/feedback-toast";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Field, FieldLabel } from "@/components/ui/field";
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from "@/components/ui/input-group";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { PopoverContent } from "@/components/ui/popover";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import type { ModelCardData } from "@/components/app-ui/ModelCard";
import {
  canUseCustomModelPricing,
  clearVisibleModelSelection,
} from "@/features/provider-accounts/models";
import { CodeExampleBlock } from "./code-example-block";
import {
  ALL_FILTER_OPTIONS,
  COMPACT_MODEL_PAGE_SIZE,
  CompactTableRow,
  ModelCardsSkeleton,
  ModelListItem,
  buildFallbackDraft,
  formatCapabilityLabel,
  type PricingEditorState,
} from "./models-editor-items";
import {
  buildProviderAccountModelFqn,
  createInitialModelFilters,
  filterAndSortProviderModels,
  getActiveModelFilterChips,
  getModelFamily,
  hasInvalidModelPriceRange,
  normalizeProviderScopedModelName,
  type ModelFilterState,
} from "./model-editor-utils";
import type {
  Banner,
  CurlGuide,
  CurlGuideShell,
  ModelDraft,
  ProviderCatalogModel,
  TenantProviderAccount,
} from "./types";
import { CURL_GUIDE_LANGUAGE } from "./types";
import { getProviderDisplaySettings } from "./api";
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
  resolveCurlGuideMode,
  toNumberOrUndefined,
} from "./utils";


export const ModelsEditor = memo(function ModelsEditor(props: {
  t: ReturnType<typeof useTranslations>;
  account: TenantProviderAccount | null;
  pendingAccountName?: string;
  pendingProvider?: string;
  setBanner: React.Dispatch<React.SetStateAction<Banner | null>>;
  providerModels: ProviderCatalogModel[];
  modelDrafts: Record<string, ModelDraft>;
  setModelDrafts: React.Dispatch<React.SetStateAction<Record<string, ModelDraft>>>;
  isSaving: boolean;
  isLoadingModels: boolean;
  canManage: boolean;
  onSave: (draftOverride?: Record<string, ModelDraft>) => void;
  variant?: "default" | "compact";
}) {
  const ux = useTranslations("providerAccounts.ux");
  const locale = useLocale();
  const {
    account,
    modelDrafts,
    onSave,
    providerModels,
    setModelDrafts,
  } = props;
  const canUseCustomPricing = canUseCustomModelPricing(account);
  const initialFilters = useMemo(() => createInitialModelFilters(""), []);
  const [compactPage, setCompactPage] = useState(0);
  const [filters, setFiltersState] = useState<ModelFilterState>(() => createInitialModelFilters(""));
  const setFilters = useCallback<React.Dispatch<React.SetStateAction<ModelFilterState>>>((update) => {
    setCompactPage(0);
    setFiltersState(update);
  }, []);
  const [displayCurrency, setDisplayCurrency] = useState<"USD" | "TOMAN">("USD");
  const [displayScale, setDisplayScale] = useState<"PER_MILLION_TOKENS" | "PER_TOKEN">("PER_MILLION_TOKENS");
  const [usdToIrr, setUsdToIrr] = useState(0);
  const [curlGuide, setCurlGuide] = useState<CurlGuide | null>(null);
  const [curlGuideShell, setCurlGuideShell] = useState<CurlGuideShell>("powershell");
  const [moreFiltersOpen, setMoreFiltersOpen] = useState(false);
  const [priceDisplayOpen, setPriceDisplayOpen] = useState(false);
  const [mobileViewMode, setMobileViewMode] = useState<"table" | "card">("table");
  const [pricingEditor, setPricingEditor] = useState<PricingEditorState>({
    open: false,
    modelKey: "",
    pricingMode: "public",
    inputPrice: "",
    outputPrice: "",
    ratePrices: {},
    error: "",
  });
  const pricingEditorView = useMemo<PricingEditorState>(() => {
    if (canUseCustomPricing) return pricingEditor;
    return { ...pricingEditor, open: false, error: "" };
  }, [canUseCustomPricing, pricingEditor]);
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
      pricingPerToken: ux("models.displayUnit.perToken"),
      modelId: ux("models.card.modelId"),
      providerId: ux("models.card.providerId"),
      apiMode: ux("models.card.apiMode"),
      releaseDate: ux("models.card.releaseDate"),
      knowledgeCutoff: ux("models.card.knowledgeCutoff"),
      regions: ux("models.card.regions"),
      rateLimitNotes: ux("models.card.rateLimitNotes"),
      license: ux("models.card.license"),
      viewDetails: ux("models.card.viewDetails"),
      copyId: "FQN",
      copied: ux("models.card.copied"),
      lowerCost: ux("models.card.lowerCost"),
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
    return buildProviderAccountModelFqn({
      account: props.account,
      pendingAccountName: props.pendingAccountName,
      pendingProvider: props.pendingProvider,
      model: modelValue,
    });
  }, [props.account, props.pendingAccountName, props.pendingProvider]);

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
  const visibleModels = useMemo(
    () =>
      filterAndSortProviderModels({
        models: providerModels,
        filters: deferredFilters,
        isSelectedOnly: false,
        isModelEnabled: (modelKey) => Boolean(modelDrafts[modelKey]?.enabled),
      }),
    [deferredFilters, modelDrafts, providerModels],
  );
  const renderedVisibleModels = useDeferredValue(visibleModels);
  const lastCompactPage = Math.max(0, Math.ceil(renderedVisibleModels.length / COMPACT_MODEL_PAGE_SIZE) - 1);
  const effectiveCompactPage = Math.min(compactPage, lastCompactPage);
  const compactPageStart = effectiveCompactPage * COMPACT_MODEL_PAGE_SIZE;
  const compactPageModels = useMemo(
    () => renderedVisibleModels.slice(compactPageStart, compactPageStart + COMPACT_MODEL_PAGE_SIZE),
    [compactPageStart, renderedVisibleModels],
  );
  const compactPageEnd = Math.min(compactPageStart + COMPACT_MODEL_PAGE_SIZE, renderedVisibleModels.length);
  const visibleSelectedModelsCount = useMemo(
    () => renderedVisibleModels.filter((model) => modelDrafts[model.model]?.enabled).length,
    [modelDrafts, renderedVisibleModels],
  );
  const allVisibleModelsSelected =
    renderedVisibleModels.length > 0 && visibleSelectedModelsCount === renderedVisibleModels.length;

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

  const openPricingEditor = useCallback((model: ProviderCatalogModel) => {
    if (!canUseCustomPricing) return;
    const draft = modelDrafts[model.model] ?? buildFallbackDraft(model, []);
    setPricingEditor({
      open: true,
      modelKey: model.model,
      pricingMode: draft.pricingMode,
      inputPrice: draft.inputPrice,
      outputPrice: draft.outputPrice,
      ratePrices: Object.fromEntries((draft.rateOverrides ?? []).map((rate) => [rate.key, rate.unit_price_usd])),
      error: "",
    });
  }, [canUseCustomPricing, modelDrafts]);

  const closePricingEditor = useCallback(() => {
    setPricingEditor((prev) => ({ ...prev, open: false, error: "" }));
  }, []);

  const applyPricingEditor = useCallback(() => {
    if (!canUseCustomPricing) {
      setPricingEditor((prev) => ({ ...prev, open: false, error: "" }));
      return;
    }
    const inputNumber = toNumberOrUndefined(pricingEditor.inputPrice);
    const outputNumber = toNumberOrUndefined(pricingEditor.outputPrice);
    const model = providerModels.find((item) => item.model === pricingEditor.modelKey);
    const publicRates = model?.public_pricing?.rates ?? [];
    if (pricingEditor.pricingMode === "private") {
      if (pricingEditor.inputPrice.trim() && (typeof inputNumber !== "number" || inputNumber < 0)) {
        setPricingEditor((prev) => ({ ...prev, error: ux("models.validation.invalidInputPrice") }));
        return;
      }
      if (pricingEditor.outputPrice.trim() && (typeof outputNumber !== "number" || outputNumber < 0)) {
        setPricingEditor((prev) => ({ ...prev, error: ux("models.validation.invalidOutputPrice") }));
        return;
      }
      for (const rate of publicRates) {
        const value = pricingEditor.ratePrices[rate.key]?.trim() ?? "";
        if (value && (typeof toNumberOrUndefined(value) !== "number" || Number(value) < 0)) {
          setPricingEditor((prev) => ({ ...prev, error: ux("models.validation.invalidInputPrice") }));
          return;
        }
      }
    }

    const current = modelDrafts[pricingEditor.modelKey];
    if (!current) return;
    const nextDrafts = {
      ...modelDrafts,
      [pricingEditor.modelKey]: {
        ...current,
        enabled: true,
        pricingMode: pricingEditor.pricingMode,
        inputPrice: pricingEditor.inputPrice.trim(),
        outputPrice: pricingEditor.outputPrice.trim(),
        rateOverrides: publicRates.flatMap((rate) => {
          const value = pricingEditor.ratePrices[rate.key]?.trim() ?? "";
          return value ? [{ ...rate, unit_price_usd: value }] : [];
        }),
      },
    };
    setModelDrafts(nextDrafts);
    if (account) onSave(nextDrafts);
    setPricingEditor((prev) => ({ ...prev, open: false, error: "" }));
  }, [account, canUseCustomPricing, modelDrafts, onSave, pricingEditor, providerModels, setModelDrafts, ux]);

  const renderPricingEditorContent = (model: ProviderCatalogModel) => (
    canUseCustomPricing ? <PopoverContent align="start" className="w-72 p-3">
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
          model.public_pricing?.rates?.length ? (
            <div className="max-h-72 space-y-2 overflow-y-auto pe-1">
              {model.public_pricing.rates.map((rate) => (
                <div key={rate.key} className="rounded-md border border-border p-2">
                  <Label className="font-mono text-xs" dir="ltr">{rate.dimension}</Label>
                  <Input
                    className="mt-1 h-8"
                    value={pricingEditor.ratePrices[rate.key] ?? ""}
                    onChange={(event) =>
                      setPricingEditor((prev) => ({
                        ...prev,
                        ratePrices: { ...prev.ratePrices, [rate.key]: event.target.value },
                        error: "",
                      }))
                    }
                    placeholder={`${rate.unit_price_usd} USD / ${rate.unit}`}
                    dir="ltr"
                  />
                  {rate.condition && Object.keys(rate.condition).length ? (
                    <p className="mt-1 truncate font-mono text-xs text-muted-foreground" dir="ltr">
                      {JSON.stringify(rate.condition)}
                    </p>
                  ) : null}
                </div>
              ))}
            </div>
          ) : (
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
          )
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
    </PopoverContent> : null
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
  }, [buildModelFqn, ux]);

  const getSelectableOptions = useCallback((model: ProviderCatalogModel) => {
    const modes = getModelModes(model);
    if (modes.length > 1) return modes;
    const selectionTypes = getModelSelectionTypes(model);
    if (selectionTypes.length > 0) return selectionTypes;
    return modes;
  }, []);

  const setModelEnabled = useCallback(
    (model: ProviderCatalogModel, draft: ModelDraft, selectableOptions: string[], enabled: boolean) => {
      const nextDrafts = {
        ...modelDrafts,
        [model.model]: {
          ...draft,
          enabled,
          selectedCapabilities: enabled ? selectableOptions : draft.selectedCapabilities,
        },
      };
      setModelDrafts(nextDrafts);
      if (account) onSave(nextDrafts);
    },
    [account, modelDrafts, onSave, setModelDrafts],
  );

  const handleVisibleSelectionChange = useCallback((models: ProviderCatalogModel[], checked: boolean) => {
    let nextDrafts: Record<string, ModelDraft>;
    if (!checked) {
      nextDrafts = clearVisibleModelSelection(models, modelDrafts);
    } else {
      nextDrafts = { ...modelDrafts };
      for (const model of models) {
        const selectableOptions = getSelectableOptions(model);
        const draft = nextDrafts[model.model] ?? buildFallbackDraft(model, selectableOptions);
        nextDrafts[model.model] = {
          ...draft,
          enabled: true,
          selectedCapabilities: selectableOptions,
        };
      }
    }
    setModelDrafts(nextDrafts);
    if (account) onSave(nextDrafts);
  }, [account, getSelectableOptions, modelDrafts, onSave, setModelDrafts]);

  useEffect(() => {
    let canceled = false;
    void (async () => {
      try {
        const data = await getProviderDisplaySettings();
        if (canceled) return;
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

  const curlGuideDialog = (
    <Dialog open={curlGuide !== null} onOpenChange={(open) => { if (!open) setCurlGuide(null); }}>
      <DialogContent
        showCloseButton={false}
        className="flex max-h-[calc(100dvh-1rem)] w-[calc(100vw-1rem)] max-w-4xl flex-col gap-0 overflow-hidden rounded-xl p-0 sm:max-h-[calc(100dvh-2rem)] sm:w-[calc(100vw-2rem)] sm:max-w-4xl"
      >
        <div className="flex shrink-0 items-start gap-4 px-4 py-4 sm:px-6 sm:py-5">
          <DialogHeader className="min-w-0 flex-1 items-start gap-1.5 text-start">
            <div className="flex min-w-0 flex-wrap items-center gap-2.5">
              <span className="inline-flex size-10 shrink-0 items-center justify-center rounded-lg bg-primary text-primary-foreground shadow-xs">
                  <Code2 aria-hidden="true" />
              </span>
              <DialogTitle className="min-w-0 text-xl leading-7 sm:text-2xl sm:leading-8">
                {ux("models.curlGuide.title")}
              </DialogTitle>
            </div>
            <DialogDescription className="max-w-3xl text-start text-sm leading-6">
              {ux("models.curlGuide.descriptionPrefix")}{" "}
              <bdi dir="ltr" className="font-medium text-foreground">{curlGuide?.modelName ?? ""}</bdi>{" "}
              {ux("models.curlGuide.descriptionSuffix")}
            </DialogDescription>
          </DialogHeader>
          <DialogClose asChild>
            <Button type="button" variant="ghost" size="icon" className="shrink-0" aria-label={ux("models.curlGuide.close")}>
              <X aria-hidden="true" />
            </Button>
          </DialogClose>
        </div>

        <Separator />

        <div className="custom-scrollbar min-h-0 flex-1 overflow-y-auto">
          <div className="flex flex-col gap-4 px-4 py-4 sm:px-6 sm:py-5">
            <Field>
              <FieldLabel htmlFor="curl-guide-model-fqn">{ux("models.curlGuide.modelFqnLabel")}</FieldLabel>
              <InputGroup className="h-10 bg-muted/20">
                <InputGroupInput
                  id="curl-guide-model-fqn"
                  readOnly
                  value={curlGuide?.modelFqn ?? ""}
                  dir="ltr"
                  className="truncate font-mono text-xs font-semibold text-primary"
                />
                <InputGroupAddon align="inline-end" className="pe-1">
                  <CopyAction
                    value={curlGuide?.modelFqn ?? ""}
                    copyLabel={ux("models.card.copyFqn")}
                    copiedLabel={ux("models.card.copied")}
                    showLabel
                    variant="ghost"
                    size="sm"
                    className="h-8 shrink-0"
                    onCopied={() => showFeedbackToast({ kind: "success", title: ux("messages.fqnCopied") })}
                  />
                </InputGroupAddon>
              </InputGroup>
            </Field>

            <Alert variant="info">
              <AlertTitle>{ux("models.curlGuide.instructionsLabel")}</AlertTitle>
              <AlertDescription>{ux("models.curlGuide.instructions")}</AlertDescription>
            </Alert>

            <Tabs value={curlGuideShell} onValueChange={(value) => setCurlGuideShell(value as CurlGuideShell)} className="min-w-0 gap-3">
              <TabsList className="w-full sm:w-fit">
                <TabsTrigger value="powershell" className="min-w-0 sm:min-w-28">{ux("models.curlGuide.shells.powershell")}</TabsTrigger>
                <TabsTrigger value="bash" className="min-w-0 sm:min-w-20">{ux("models.curlGuide.shells.bash")}</TabsTrigger>
              </TabsList>
              <TabsContent value="powershell" className="mt-0">
                <CodeExampleBlock
                  title={ux("models.curlGuide.requestTitle")}
                  code={curlGuide?.commands.powershell ?? ""}
                  language={CURL_GUIDE_LANGUAGE.powershell}
                  shellLabel={ux("models.curlGuide.shells.powershell")}
                  copyLabel={ux("models.curlGuide.copySample")}
                  copiedLabel={ux("models.card.copied")}
                  onCopied={() => showFeedbackToast({ kind: "success", title: ux("messages.curlCopied") })}
                  footer={renderCurlGuideDocsFooter()}
                />
              </TabsContent>
              <TabsContent value="bash" className="mt-0">
                <CodeExampleBlock
                  title={ux("models.curlGuide.requestTitle")}
                  code={curlGuide?.commands.bash ?? ""}
                  language={CURL_GUIDE_LANGUAGE.bash}
                  shellLabel={ux("models.curlGuide.shells.bash")}
                  copyLabel={ux("models.curlGuide.copySample")}
                  copiedLabel={ux("models.card.copied")}
                  onCopied={() => showFeedbackToast({ kind: "success", title: ux("messages.curlCopied") })}
                  footer={renderCurlGuideDocsFooter()}
                />
              </TabsContent>
            </Tabs>
            {curlGuide?.sampleResponse ? (
              <div className="mt-1">
                <CodeExampleBlock
                  title={ux("models.curlGuide.sampleResponse")}
                  code={curlGuide.sampleResponse}
                  language="json"
                  shellLabel="JSON"
                />
              </div>
            ) : null}
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );

  if (props.variant === "compact") {
    return (
      <>
        <div className="shrink-0 bg-card px-4 py-3 shadow-sm sm:px-5">
          <div className="flex flex-col gap-2.5">
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
                <TooltipProvider>
                  <ToggleGroup
                    type="single"
                    value={mobileViewMode}
                    onValueChange={(value) => {
                      if (value === "table" || value === "card") setMobileViewMode(value);
                    }}
                    aria-label={ux("integrations.wizard.viewMode")}
                    className="inline-flex rounded-full border border-border/70 bg-muted/30 p-1"
                  >
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <ToggleGroupItem
                          value="table"
                          aria-label={ux("integrations.wizard.tableView")}
                          className="size-8 rounded-full p-0 data-[state=on]:bg-secondary [&>svg]:size-4"
                        >
                          <Table2 aria-hidden="true" />
                        </ToggleGroupItem>
                      </TooltipTrigger>
                      <TooltipContent>{ux("integrations.wizard.tableView")}</TooltipContent>
                    </Tooltip>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <ToggleGroupItem
                          value="card"
                          aria-label={ux("integrations.wizard.cardView")}
                          className="size-8 rounded-full p-0 data-[state=on]:bg-secondary [&>svg]:size-4"
                        >
                          <LayoutGrid aria-hidden="true" />
                        </ToggleGroupItem>
                      </TooltipTrigger>
                      <TooltipContent>{ux("integrations.wizard.cardView")}</TooltipContent>
                    </Tooltip>
                  </ToggleGroup>
                </TooltipProvider>
              </div>
              <div className="relative w-full min-w-0 sm:w-auto sm:flex-1">
                <Search className="pointer-events-none absolute end-3 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
                <Input
                  value={filters.search}
                  onChange={(event) => setFilters((prev) => ({ ...prev, search: event.target.value }))}
                  placeholder={props.t("step3.searchPlaceholder")}
                  aria-label={props.t("step3.searchPlaceholder")}
                  className="h-9 pe-9 text-start text-sm placeholder:text-start"
                />
              </div>
              <div className="flex flex-wrap items-center gap-3 text-sm">
                <label
                  htmlFor="compact-select-all-models"
                  className="inline-flex cursor-pointer flex-row items-center gap-1.5 text-foreground has-[:disabled]:pointer-events-none has-[:disabled]:opacity-50"
                >
                  <Checkbox
                    id="compact-select-all-models"
                    checked={allVisibleModelsSelected}
                    disabled={!props.canManage || props.isSaving || renderedVisibleModels.length === 0}
                    onCheckedChange={(checked) => handleVisibleSelectionChange(renderedVisibleModels, checked === true)}
                  />
                  <span>{ux("integrations.wizard.selectAll")}</span>
                </label>
              </div>
            </div>
          </div>
        </div>

        <section className="flex h-full min-h-0 flex-col overflow-hidden bg-card">
          <div className="min-h-0 flex-1 overflow-y-auto pb-3 pt-3">
            <div className="flex min-h-full min-w-0 flex-col gap-2">
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
                    <table className="w-full table-auto border-collapse text-start">
                      <colgroup>
                        <col className="w-12" />
                        <col className="w-px" />
                        <col className="w-20" />
                        <col className="w-20" />
                        <col className="w-20" />
                        <col className="hidden md:table-column" />
                      </colgroup>
                      <thead className="bg-muted/35 text-xs text-muted-foreground">
                        <tr>
                          <th className="px-2 py-2 text-center font-medium">
                            <span className="sr-only">{ux("common.enabled")}</span>
                          </th>
                          <th className="w-px whitespace-nowrap px-3 py-2 text-start font-medium">{ux("models.card.modelId")}</th>
                          <th className="px-3 py-2 text-center font-medium">{ux("models.card.context")}</th>
                          <th className="px-3 py-2 text-start font-medium">{ux("models.labels.input")}</th>
                          <th className="px-3 py-2 text-start font-medium">{ux("models.labels.output")}</th>
                          <th className="hidden px-3 py-2 text-start font-medium md:table-cell">{ux("models.card.capabilities")}</th>
                        </tr>
                      </thead>
                      <tbody>
                        {compactPageModels.map((model) => (
                          <CompactTableRow
                            key={model.model}
                            canManage={props.canManage && !props.isSaving}
                            draft={props.modelDrafts[model.model]}
                            model={model}
                            selectableOptions={getSelectableOptions(model)}
                            setModelEnabled={setModelEnabled}
                            displayCurrency={displayCurrency}
                            displayScale={displayScale}
                            usdToIrr={usdToIrr}
                            ux={ux}
                          />
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              ) : (
                <>
                  {compactPageModels.map((model) => (
                    <ModelListItem
                      key={model.model}
                      account={props.account}
                      canUseCustomPricing={canUseCustomPricing}
                      canManage={props.canManage && !props.isSaving}
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
                      pricingEditor={pricingEditorView}
                      renderPricingEditorContent={renderPricingEditorContent}
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
        {renderedVisibleModels.length > COMPACT_MODEL_PAGE_SIZE ? (
          <div className="shrink-0 border-t border-border bg-card px-4 py-3 sm:px-5">
            <TablePagination
              label={ux("models.pagination.label")}
              summary={ux("models.pagination.summary", {
                start: compactPageStart + 1,
                end: compactPageEnd,
                total: renderedVisibleModels.length,
              })}
              previousLabel={ux("models.pagination.previous")}
              nextLabel={ux("models.pagination.next")}
              onPrevious={() => setCompactPage(Math.max(0, effectiveCompactPage - 1))}
              onNext={() => setCompactPage(effectiveCompactPage + 1)}
              canPrevious={effectiveCompactPage > 0}
              canNext={compactPageEnd < renderedVisibleModels.length}
              disabled={props.isSaving || props.isLoadingModels}
              page={new Intl.NumberFormat(locale).format(effectiveCompactPage + 1)}
              pageSize={new Intl.NumberFormat(locale).format(COMPACT_MODEL_PAGE_SIZE)}
              pageSizeLabel={ux("models.pagination.pageSize")}
            />
          </div>
        ) : null}
        {curlGuideDialog}
      </>
    );
  }

  function renderCurlGuideDocsFooter() {
    return (
      <p className="text-end text-xs leading-5 text-muted-foreground">
        {ux("models.curlGuide.docsDescriptionPrefix")}{" "}
        <a href="/quick-start" className="inline-flex items-center gap-1 font-medium text-primary hover:underline">
          {ux("models.curlGuide.docsLabel")}
          <ExternalLink data-icon="inline-end" />
        </a>{" "}
        {ux("models.curlGuide.docsDescriptionSuffix")}
      </p>
    );
  }

  return (
    <section className="flex flex-col gap-3">
      <div className="flex flex-col gap-2">
        <Collapsible open={moreFiltersOpen} onOpenChange={setMoreFiltersOpen} className="flex flex-col gap-2">
          <div className="flex items-stretch gap-2">
            <InputGroup className="min-w-0 flex-1 has-[[data-slot=input-group-control]:focus-visible]:border-ring has-[[data-slot=input-group-control]:focus-visible]:ring-0">
              <InputGroupAddon align="inline-start" className="ps-3 pe-1">
                <Search aria-hidden="true" />
              </InputGroupAddon>
              <InputGroupInput
                value={filters.search}
                onChange={(event) => setFilters((prev) => ({ ...prev, search: event.target.value }))}
                placeholder={props.t("step3.searchPlaceholder")}
                className="!border-0 text-xs shadow-none focus-visible:!border-0 focus-visible:!ring-0 focus-visible:!shadow-none sm:text-sm"
              />
              {filters.search.trim() ? (
                <InputGroupAddon align="inline-end" className="ps-1 pe-1">
                  <InputGroupButton
                    size="icon-xs"
                    className="text-muted-foreground hover:text-foreground"
                    onClick={() => setFilters((prev) => ({ ...prev, search: "" }))}
                    aria-label={ux("actions.clear")}
                    title={ux("actions.clear")}
                  >
                    <X aria-hidden="true" />
                  </InputGroupButton>
                </InputGroupAddon>
              ) : null}
            </InputGroup>

            <CollapsibleTrigger asChild>
              <Button type="button" variant="outline" size="sm" className="h-9 shrink-0 whitespace-nowrap">
                <SlidersHorizontal data-icon="inline-start" />
                {ux("models.filters.more")}
              </Button>
            </CollapsibleTrigger>
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    type="button"
                    variant={priceDisplayOpen ? "secondary" : "outline"}
                    size="icon-sm"
                    className="h-9 shrink-0"
                    onClick={() => setPriceDisplayOpen((open) => !open)}
                    aria-label={ux("models.actions.priceDisplaySettings")}
                  >
                    <Settings aria-hidden="true" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>{ux("models.actions.priceDisplaySettings")}</TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>
          <CollapsibleContent>
            <div className="flex items-start gap-2 rounded-lg border border-border/70 bg-muted/20 p-2">
              <div className="grid min-w-0 flex-1 grid-cols-2 gap-2 md:grid-cols-3 xl:grid-cols-4">
                <Input
                  value={filters.minPrice}
                  onChange={(event) => setFilters((prev) => ({ ...prev, minPrice: event.target.value }))}
                  placeholder={ux("models.filters.minPrice")}
                  inputMode="decimal"
                  className="h-8 text-xs sm:h-10 sm:text-sm"
                />
                <Input
                  value={filters.maxPrice}
                  onChange={(event) => setFilters((prev) => ({ ...prev, maxPrice: event.target.value }))}
                  placeholder={ux("models.filters.maxPrice")}
                  inputMode="decimal"
                  className="h-8 text-xs sm:h-10 sm:text-sm"
                />
                <Input
                  value={filters.minContext}
                  onChange={(event) => setFilters((prev) => ({ ...prev, minContext: event.target.value }))}
                  placeholder={ux("models.filters.minContext")}
                  inputMode="numeric"
                  className="h-8 text-xs sm:h-10 sm:text-sm"
                />
                <AppSelectField
                  ariaLabel={ux("models.filters.allInputModalities")}
                  value={filters.inputModality || ALL_FILTER_OPTIONS}
                  onValueChange={(value) => setFilters((prev) => ({ ...prev, inputModality: value === ALL_FILTER_OPTIONS ? "" : value }))}
                  optionTextDir="ltr"
                  size="sm"
                  triggerClassName="min-w-0"
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
                  size="sm"
                  triggerClassName="min-w-0"
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
                  size="sm"
                  triggerClassName="min-w-0"
                  options={[
                    { value: ALL_FILTER_OPTIONS, label: ux("models.filters.allFeatures") },
                    ...allFeatures.map((item) => ({ value: item, label: item })),
                  ]}
                />
                <AppSelectField
                  ariaLabel={ux("models.sort.priceAsc")}
                  value={filters.sortBy}
                  onValueChange={(value) => setFilters((prev) => ({ ...prev, sortBy: value as ModelFilterState["sortBy"] }))}
                  size="sm"
                  triggerClassName="min-w-0"
                  options={[
                    { value: "price_asc", label: ux("models.sort.priceAsc") },
                    { value: "price_desc", label: ux("models.sort.priceDesc") },
                    { value: "context_asc", label: ux("models.sort.contextAsc") },
                    { value: "context_desc", label: ux("models.sort.contextDesc") },
                  ]}
                />
              </div>
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      type="button"
                      variant="outline"
                      size="icon-sm"
                      onClick={resetFilters}
                      aria-label={ux("models.actions.clearFilters")}
                    >
                      <RotateCcw aria-hidden="true" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>{ux("models.actions.clearFilters")}</TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </div>
          </CollapsibleContent>
        </Collapsible>

        {priceDisplayOpen ? (
          <div className="flex items-start gap-2 rounded-lg border border-border/70 bg-muted/20 p-2">
            <div className="grid min-w-0 flex-1 grid-cols-1 gap-2 sm:grid-cols-2">
              <AppSelectField
                ariaLabel={ux("models.displayUnit.currencyUsd")}
                value={displayCurrency}
                onValueChange={(value) => setDisplayCurrency(value as "USD" | "TOMAN")}
                size="sm"
                triggerClassName="min-w-0"
                options={[
                  { value: "USD", label: ux("models.displayUnit.currencyUsd") },
                  { value: "TOMAN", label: ux("models.displayUnit.currencyToman") },
                ]}
              />
              <AppSelectField
                ariaLabel={ux("models.displayUnit.perMillionTokens")}
                value={displayScale}
                onValueChange={(value) => setDisplayScale(value as "PER_MILLION_TOKENS" | "PER_TOKEN")}
                size="sm"
                triggerClassName="min-w-0"
                options={[
                  { value: "PER_MILLION_TOKENS", label: ux("models.displayUnit.perMillionTokens") },
                  { value: "PER_TOKEN", label: ux("models.displayUnit.perToken") },
                ]}
              />
            </div>
            <Button
              type="button"
              variant="outline"
              size="icon-sm"
              onClick={() => setPriceDisplayOpen(false)}
              aria-label={ux("models.curlGuide.close")}
              title={ux("models.curlGuide.close")}
            >
              <X aria-hidden="true" />
            </Button>
          </div>
        ) : null}

        {hasInvalidPriceRange ? (
          <p className="text-xs text-destructive">{ux("models.validation.invalidInputPrice")}</p>
        ) : null}

        {activeFilters.length > 0 ? (
          <div className="flex flex-wrap items-center gap-1.5">
            {activeFilters.map((item) => (
              <Badge key={item.key} variant="secondary" className="flex items-center gap-1 pe-1 text-xs">
                <span className="max-w-44 truncate sm:max-w-56">{item.label}</span>
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

      {props.isLoadingModels ? (
        <ModelCardsSkeleton />
      ) : renderedVisibleModels.length === 0 ? (
        <EmptyState variant="inline" density="compact" title={props.t("step3.empty")} />
      ) : (
        <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
          {renderedVisibleModels.map((model) => (
            <ModelListItem
              key={model.model}
              account={props.account}
              canUseCustomPricing={canUseCustomPricing}
              canManage={props.canManage && !props.isSaving}
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
              pricingEditor={pricingEditorView}
              renderPricingEditorContent={renderPricingEditorContent}
              setModelEnabled={setModelEnabled}
              toModelCardData={toModelCardData}
              ux={ux}
              usdToIrr={usdToIrr}
              getSelectableOptions={getSelectableOptions}
            />
          ))}
        </div>
      )}
      {curlGuideDialog}
    </section>
  );
}, (prev, next) => (
  prev.account === next.account &&
  prev.pendingAccountName === next.pendingAccountName &&
  prev.pendingProvider === next.pendingProvider &&
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


