"use client";

import { Fragment, useId, useMemo, useState } from "react";
import Link from "next/link";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import {
  Brain,
  Check,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
  CircleAlert,
  Clipboard,
  Eye,
  Filter,
  Info,
  ListFilter,
  Search,
  SlidersHorizontal,
  Sparkles,
  Wrench,
} from "lucide-react";
import { CompareToggleButton } from "@/components/compare/compare-toggle-button";
import { ProviderMark } from "@/components/catalog/provider-mark";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { Slider } from "@/components/ui/slider";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { apiGet } from "@/lib/api";
import { cn } from "@/lib/utils";
import { useDisplayFormatters } from "@/lib/use-display-formatters";

type Pricing = {
  currency?: string;
  inputPerMillionTokens?: number | null;
  outputPerMillionTokens?: number | null;
  cachedInputPerMillionTokens?: number | null;
  imageInputPrice?: number | null;
  imageOutputPrice?: number | null;
  audioInputPrice?: number | null;
  audioOutputPrice?: number | null;
  requestPrice?: number | null;
};

type ModelRow = {
  id?: string;
  canonicalId: string;
  slug: string;
  name: string;
  provider: { slug: string; name: string; logoUrl?: string | null };
  description?: string;
  status: string;
  mode?: string[];
  contextWindow?: number | null;
  maxOutputTokens?: number | null;
  inputModalities: string[];
  outputModalities: string[];
  capabilities: string[];
  supportedParameters?: string[];
  pricing?: Pricing | null;
  free: boolean;
  updatedAt?: string;
  catalogPreview?: boolean;
};

type FacetOption = { slug: string; name: string; count?: number };
type FacetsResponse = {
  providers?: FacetOption[];
  capabilities?: FacetOption[];
  inputModalities?: FacetOption[];
  outputModalities?: FacetOption[];
  data?: {
    providers?: FacetOption[];
    capabilities?: FacetOption[];
    inputModalities?: FacetOption[];
    outputModalities?: FacetOption[];
  };
};

type ModelsResponse = {
  data: ModelRow[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
};

type SortValue =
  | "most_popular"
  | "newest"
  | "name"
  | "cheapest_input"
  | "cheapest_output"
  | "largest_context";

type FeatureKey = "vision" | "tools" | "reasoning";

const ALL_VALUE = "__all__";
const PAGE_SIZE = 24;
const CONTEXT_LIMIT = 2_000_000;
const PRICE_LIMIT = 100;

const FEATURE_CAPABILITIES: Record<FeatureKey, string> = {
  vision: "vision",
  tools: "function_calling",
  reasoning: "reasoning",
};

const PREVIEW_MODELS: ModelRow[] = [
  previewModel("openai", "OpenAI", "gpt-4.1-mini", "GPT-4.1 Mini", 0.4, 1.6, 0.1, ["vision", "function_calling"], 1_000_000),
  previewModel("anthropic", "Anthropic", "claude-sonnet-4", "Claude Sonnet 4", 3, 15, 0.3, ["vision", "reasoning", "function_calling"], 200_000),
  previewModel("google-gemini", "Google Gemini", "gemini-2.5-flash", "Gemini 2.5 Flash", 0.3, 2.5, 0.03, ["vision", "function_calling"], 1_000_000),
  previewModel("mistral-ai", "Mistral AI", "mistral-small-latest", "Mistral Small", 0.2, 0.6, null, ["function_calling"], 128_000),
  previewModel("cohere", "Cohere", "command-r-plus", "Command R+", 2.5, 10, null, ["function_calling"], 128_000),
  previewModel("groq", "Groq", "llama-3.3-70b-versatile", "Llama 3.3 70B", 0.59, 0.79, null, [], 128_000),
  previewModel("openrouter", "OpenRouter", "deepseek-r1", "DeepSeek R1", 0.55, 2.19, null, ["reasoning"], 163_840),
  previewModel("together-ai", "Together AI", "qwen-2.5-72b", "Qwen 2.5 72B", 1.2, 1.2, null, ["function_calling"], 128_000),
];

function previewModel(
  providerSlug: string,
  providerName: string,
  slug: string,
  name: string,
  input: number,
  output: number,
  cached: number | null,
  capabilities: string[],
  contextWindow: number,
): ModelRow {
  return {
    canonicalId: `${providerSlug}/${slug}`,
    slug,
    name,
    provider: { slug: providerSlug, name: providerName },
    status: "preview",
    mode: ["chat"],
    contextWindow,
    maxOutputTokens: 32_768,
    inputModalities: capabilities.includes("vision") ? ["text", "image"] : ["text"],
    outputModalities: ["text"],
    capabilities,
    supportedParameters: capabilities.includes("function_calling") ? ["tools"] : [],
    pricing: {
      currency: "USD",
      inputPerMillionTokens: input,
      outputPerMillionTokens: output,
      cachedInputPerMillionTokens: cached,
    },
    free: false,
    updatedAt: "2026-07-25T00:00:00.000Z",
    catalogPreview: true,
  };
}

export function ModelsExplorer() {
  const t = useTranslations("Catalog");
  const tCommon = useTranslations("Common");
  const { formatPrice } = useDisplayFormatters();
  const numberFormatter = useMemo(() => new Intl.NumberFormat("fa-IR"), []);
  const compactNumberFormatter = useMemo(
    () => new Intl.NumberFormat("fa-IR", { notation: "compact", maximumFractionDigits: 1 }),
    [],
  );
  const priceRangeFormatter = useMemo(
    () => new Intl.NumberFormat("fa-IR", { maximumFractionDigits: 1 }),
    [],
  );

  const [q, setQ] = useState("");
  const [provider, setProvider] = useState("");
  const [features, setFeatures] = useState<FeatureKey[]>([]);
  const [inputModality, setInputModality] = useState("");
  const [outputModality, setOutputModality] = useState("");
  const [contextRange, setContextRange] = useState<[number, number]>([0, CONTEXT_LIMIT]);
  const [priceRange, setPriceRange] = useState<[number, number]>([0, PRICE_LIMIT]);
  const [sort, setSort] = useState<SortValue>("most_popular");
  const [page, setPage] = useState(1);
  const [advancedFiltersOpen, setAdvancedFiltersOpen] = useState(false);
  const [expandedModel, setExpandedModel] = useState<string | null>(null);
  const [copiedModel, setCopiedModel] = useState<string | null>(null);

  const contextRangeActive = contextRange[0] !== 0 || contextRange[1] !== CONTEXT_LIMIT;
  const priceRangeActive = priceRange[0] !== 0 || priceRange[1] !== PRICE_LIMIT;
  const activeFilterCount =
    [q, provider, inputModality, outputModality].filter(Boolean).length +
    features.length +
    Number(contextRangeActive) +
    Number(priceRangeActive);

  const params = useMemo(() => {
    const searchParams = new URLSearchParams();
    if (q.trim()) searchParams.set("q", q.trim());
    if (provider) searchParams.set("provider", provider);
    if (features.length) {
      searchParams.set(
        "capabilities",
        features.map((feature) => FEATURE_CAPABILITIES[feature]).join(","),
      );
    }
    if (inputModality) searchParams.set("input_modalities", inputModality);
    if (outputModality) searchParams.set("output_modalities", outputModality);
    if (contextRange[0] > 0) searchParams.set("min_context_window", String(contextRange[0]));
    if (contextRange[1] < CONTEXT_LIMIT) searchParams.set("max_context_window", String(contextRange[1]));
    if (priceRange[0] > 0) searchParams.set("min_input_price", String(priceRange[0]));
    if (priceRange[1] < PRICE_LIMIT) searchParams.set("max_input_price", String(priceRange[1]));
    searchParams.set("sort", sort);
    searchParams.set("page", String(page));
    searchParams.set("limit", String(PAGE_SIZE));
    return searchParams.toString();
  }, [
    q,
    provider,
    features,
    inputModality,
    outputModality,
    contextRange,
    priceRange,
    sort,
    page,
  ]);

  const modelsQuery = useQuery({
    queryKey: ["models", params],
    queryFn: () => apiGet<ModelsResponse>(`/models?${params}`),
    placeholderData: keepPreviousData,
  });
  const facetsQuery = useQuery({
    queryKey: ["facets"],
    queryFn: () => apiGet<FacetsResponse>("/facets"),
  });

  const providers = facetsQuery.data?.providers ?? facetsQuery.data?.data?.providers ?? [];
  const inputModalities =
    facetsQuery.data?.inputModalities ?? facetsQuery.data?.data?.inputModalities ?? [];
  const outputModalities =
    facetsQuery.data?.outputModalities ?? facetsQuery.data?.data?.outputModalities ?? [];
  const liveModels = modelsQuery.data?.data ?? [];
  const showingPreview =
    !modelsQuery.isLoading &&
    !modelsQuery.isError &&
    liveModels.length === 0 &&
    activeFilterCount === 0 &&
    page === 1;
  const displayedModels = showingPreview ? sortPreviewModels(PREVIEW_MODELS, sort) : liveModels;
  const totalModels = showingPreview
    ? PREVIEW_MODELS.length
    : (modelsQuery.data?.pagination.total ?? 0);
  const totalPages = showingPreview
    ? 1
    : (modelsQuery.data?.pagination.totalPages ?? 1);

  const clearFilters = () => {
    setQ("");
    setProvider("");
    setFeatures([]);
    setInputModality("");
    setOutputModality("");
    setContextRange([0, CONTEXT_LIMIT]);
    setPriceRange([0, PRICE_LIMIT]);
    setPage(1);
  };

  const toggleFeature = (feature: FeatureKey) => {
    setFeatures((current) =>
      current.includes(feature)
        ? current.filter((item) => item !== feature)
        : [...current, feature],
    );
    setPage(1);
  };

  const copyModelId = async (canonicalId: string) => {
    await navigator.clipboard.writeText(canonicalId);
    setCopiedModel(canonicalId);
    window.setTimeout(() => setCopiedModel((current) => (current === canonicalId ? null : current)), 1600);
  };

  return (
    <TooltipProvider>
      <section aria-label={t("directoryLabel")} className="flex min-w-0 flex-col gap-4">
        <div className="flex flex-col gap-3 lg:flex-row lg:items-center">
          <div className="relative min-w-0 flex-1">
            <Search
              className="pointer-events-none absolute start-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground"
              aria-hidden="true"
            />
            <Input
              aria-label={t("searchLabel")}
              className="h-10 ps-10"
              placeholder={t("searchShortPlaceholder")}
              value={q}
              onChange={(event) => {
                setQ(event.target.value);
                setPage(1);
              }}
            />
          </div>

          <div className="scrollbar-none flex min-w-0 items-center gap-2 overflow-x-auto pb-1 lg:pb-0">
            <Select value={provider || ALL_VALUE} onValueChange={(value) => {
              setProvider(value === ALL_VALUE ? "" : value);
              setPage(1);
            }}>
              <SelectTrigger className="h-10 w-auto min-w-36">
                <Filter data-icon="inline-start" aria-hidden="true" />
                <SelectValue placeholder={t("providerLabel")} />
              </SelectTrigger>
              <SelectContent position="popper">
                <SelectGroup>
                  <SelectItem value={ALL_VALUE}>{t("allProviders")}</SelectItem>
                  {providers.map((item) => (
                    <SelectItem key={item.slug} value={item.slug}>
                      <span>{item.name}</span>
                      {item.count !== undefined ? (
                        <span className="ms-auto text-xs text-muted-foreground">
                          {numberFormatter.format(item.count)}
                        </span>
                      ) : null}
                    </SelectItem>
                  ))}
                </SelectGroup>
              </SelectContent>
            </Select>

            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" size="lg" className="h-10">
                  <Sparkles data-icon="inline-start" aria-hidden="true" />
                  {t("featuresFilter")}
                  {features.length ? <Badge>{numberFormatter.format(features.length)}</Badge> : null}
                  <ChevronDown data-icon="inline-end" aria-hidden="true" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="start" className="w-52">
                <DropdownMenuLabel>{t("featuresFilter")}</DropdownMenuLabel>
                <DropdownMenuSeparator />
                {(["vision", "tools", "reasoning"] as FeatureKey[]).map((feature) => (
                  <DropdownMenuCheckboxItem
                    key={feature}
                    checked={features.includes(feature)}
                    onCheckedChange={() => toggleFeature(feature)}
                  >
                    <FeatureIcon feature={feature} />
                    {t(`features.${feature}`)}
                  </DropdownMenuCheckboxItem>
                ))}
              </DropdownMenuContent>
            </DropdownMenu>

            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" size="lg" className="h-10">
                  <ListFilter data-icon="inline-start" aria-hidden="true" />
                  {t(`sort.${sort}`)}
                  <ChevronDown data-icon="inline-end" aria-hidden="true" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="start" className="w-64">
                <DropdownMenuLabel>{t("sort.title")}</DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuRadioGroup value={sort} onValueChange={(value) => {
                  setSort(value as SortValue);
                  setPage(1);
                }}>
                  <DropdownMenuRadioItem value="most_popular">{t("sort.most_popular")}</DropdownMenuRadioItem>
                  <DropdownMenuRadioItem value="newest">{t("sort.newest")}</DropdownMenuRadioItem>
                  <DropdownMenuRadioItem value="name">{t("sort.name")}</DropdownMenuRadioItem>
                  <DropdownMenuSeparator />
                  <DropdownMenuLabel>{t("sort.priceGroup")}</DropdownMenuLabel>
                  <DropdownMenuRadioItem value="cheapest_input">{t("sort.cheapest_input")}</DropdownMenuRadioItem>
                  <DropdownMenuRadioItem value="cheapest_output">{t("sort.cheapest_output")}</DropdownMenuRadioItem>
                  <DropdownMenuRadioItem value="largest_context">{t("sort.largest_context")}</DropdownMenuRadioItem>
                </DropdownMenuRadioGroup>
              </DropdownMenuContent>
            </DropdownMenu>

            <Button
              type="button"
              variant={advancedFiltersOpen ? "secondary" : "outline"}
              size="icon-lg"
              aria-label={t("showAdvancedFilters")}
              aria-expanded={advancedFiltersOpen}
              aria-controls="catalog-advanced-filters"
              onClick={() => setAdvancedFiltersOpen((open) => !open)}
            >
              <SlidersHorizontal aria-hidden="true" />
            </Button>
          </div>

          <span className="shrink-0 text-sm tabular-nums text-muted-foreground" aria-live="polite">
            {t("modelsCountCompact", { count: totalModels })}
          </span>
        </div>

        <div
          id="catalog-advanced-filters"
          className={cn(
            "grid gap-3 rounded-xl border border-border bg-card p-4 lg:grid-cols-2",
            !advancedFiltersOpen && "hidden",
          )}
        >
          <div className="grid gap-3 sm:grid-cols-2 lg:col-span-2">
            <FilterSelect
              label={t("inputModalityLabel")}
              allLabel={t("allInputModalities")}
              value={inputModality}
              options={inputModalities}
              onValueChange={(value) => {
                setInputModality(value);
                setPage(1);
              }}
            />
            <FilterSelect
              label={t("outputModalityLabel")}
              allLabel={t("allOutputModalities")}
              value={outputModality}
              options={outputModalities}
              onValueChange={(value) => {
                setOutputModality(value);
                setPage(1);
              }}
            />
          </div>
          <RangeFilter
            label={t("contextRange")}
            value={contextRange}
            min={0}
            max={CONTEXT_LIMIT}
            step={1_000}
            formatValue={(value) => compactNumberFormatter.format(value)}
            onValueChange={(value) => {
              setContextRange(value);
              setPage(1);
            }}
          />
          <RangeFilter
            label={t("inputPriceRange")}
            value={priceRange}
            min={0}
            max={PRICE_LIMIT}
            step={0.1}
            formatValue={(value) => `$${priceRangeFormatter.format(value)}`}
            onValueChange={(value) => {
              setPriceRange(value);
              setPage(1);
            }}
          />
          {activeFilterCount ? (
            <div className="flex justify-end lg:col-span-2">
              <Button type="button" variant="ghost" onClick={clearFilters}>
                {t("clearFilters", { count: activeFilterCount })}
              </Button>
            </div>
          ) : null}
        </div>

        {facetsQuery.isError ? (
          <Alert>
            <CircleAlert aria-hidden="true" />
            <AlertTitle>{t("filtersError")}</AlertTitle>
            <AlertDescription>{t("filtersErrorDescription")}</AlertDescription>
          </Alert>
        ) : null}

        {showingPreview ? (
          <Alert>
            <Info aria-hidden="true" />
            <AlertTitle>{t("previewDataTitle")}</AlertTitle>
            <AlertDescription>{t("previewDataDescription")}</AlertDescription>
          </Alert>
        ) : null}

        {modelsQuery.isLoading ? <ModelsTableSkeleton label={t("loading")} /> : null}

        {modelsQuery.isError ? (
          <Alert variant="destructive">
            <CircleAlert aria-hidden="true" />
            <AlertTitle>{t("loadError")}</AlertTitle>
            <AlertDescription>{t("loadErrorDescription")}</AlertDescription>
          </Alert>
        ) : null}

        {!modelsQuery.isLoading && !modelsQuery.isError && displayedModels.length === 0 ? (
          <Empty className="border border-card-border bg-card">
            <EmptyHeader>
              <EmptyMedia variant="icon"><Search aria-hidden="true" /></EmptyMedia>
              <EmptyTitle>{t("noMatchesTitle")}</EmptyTitle>
              <EmptyDescription>{t("noMatches")}</EmptyDescription>
            </EmptyHeader>
            <EmptyContent>
              <Button type="button" onClick={clearFilters}>{t("clearFilters", { count: activeFilterCount })}</Button>
            </EmptyContent>
          </Empty>
        ) : null}

        {!modelsQuery.isLoading && !modelsQuery.isError && displayedModels.length > 0 ? (
          <>
            <Card className="overflow-hidden p-0 shadow-sm">
              <CardContent className="p-0 sm:p-0">
                <Table>
                  <TableHeader className="bg-muted/45">
                    <TableRow>
                      <TableHead className="min-w-44">{tCommon("provider")}</TableHead>
                      <TableHead className="min-w-64">{t("table.modelId")}</TableHead>
                      <TableHead>{t("table.inputPrice")}</TableHead>
                      <TableHead>{t("table.outputPrice")}</TableHead>
                      <TableHead>{t("table.cacheRead")}</TableHead>
                      <TableHead>{t("table.cacheWrite")}</TableHead>
                      <TableHead>{t("table.features")}</TableHead>
                      <TableHead>{t("table.updated")}</TableHead>
                      <TableHead className="text-end">{t("table.actions")}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {displayedModels.map((model) => {
                      const expanded = expandedModel === model.canonicalId;
                      const detailCount = getDetailCount(model);

                      return (
                        <Fragment key={model.canonicalId}>
                          <TableRow className="group">
                            <TableCell>
                              <div className="flex items-center gap-2">
                                {detailCount ? (
                                  <Button
                                    type="button"
                                    variant="ghost"
                                    size="icon-xs"
                                    aria-label={expanded ? t("hideRowDetails") : t("showRowDetails")}
                                    aria-expanded={expanded}
                                    onClick={() => setExpandedModel(expanded ? null : model.canonicalId)}
                                  >
                                    <ChevronLeft className={cn("transition-transform", expanded && "-rotate-90")} aria-hidden="true" />
                                  </Button>
                                ) : null}
                                <ProviderMark name={model.provider.name} logoUrl={model.provider.logoUrl} size="sm" />
                                <Link
                                  href={`/providers/${model.provider.slug}`}
                                  className="truncate text-sm font-medium hover:text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                                >
                                  {model.provider.name}
                                </Link>
                              </div>
                            </TableCell>
                            <TableCell>
                              <div className="flex min-w-0 items-center gap-2">
                                {model.catalogPreview ? (
                                  <span className="truncate font-mono text-sm font-semibold" dir="ltr">{model.slug}</span>
                                ) : (
                                  <Link
                                    href={`/models/${model.provider.slug}/${model.slug}`}
                                    className="truncate font-mono text-sm font-semibold hover:text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                                    dir="ltr"
                                    title={model.canonicalId}
                                  >
                                    {model.slug}
                                  </Link>
                                )}
                                {detailCount ? (
                                  <button
                                    type="button"
                                    className="inline-flex shrink-0 items-center gap-1 rounded-md border border-warning-soft bg-warning-soft px-1.5 py-0.5 text-xs font-semibold text-warning-soft-foreground"
                                    onClick={() => setExpandedModel(expanded ? null : model.canonicalId)}
                                  >
                                    <Sparkles className="size-3" aria-hidden="true" />
                                    +{numberFormatter.format(detailCount)}
                                  </button>
                                ) : null}
                              </div>
                            </TableCell>
                            <PriceCell value={model.pricing?.inputPerMillionTokens} formatPrice={formatPrice} />
                            <PriceCell value={model.pricing?.outputPerMillionTokens} formatPrice={formatPrice} />
                            <PriceCell value={model.pricing?.cachedInputPerMillionTokens} formatPrice={formatPrice} />
                            <TableCell className="text-muted-foreground">—</TableCell>
                            <TableCell>
                              <FeatureIcons model={model} labels={{
                                vision: t("features.vision"),
                                tools: t("features.tools"),
                                reasoning: t("features.reasoning"),
                              }} />
                            </TableCell>
                            <TableCell className="whitespace-nowrap text-xs text-muted-foreground">
                              {model.updatedAt ? formatRelativeDate(model.updatedAt) : "—"}
                            </TableCell>
                            <TableCell>
                              <div className="flex items-center justify-end gap-1">
                                {!model.catalogPreview ? (
                                  <CompareToggleButton canonicalId={model.canonicalId} compact iconOnly />
                                ) : null}
                                <Tooltip>
                                  <TooltipTrigger asChild>
                                    <Button
                                      type="button"
                                      variant="ghost"
                                      size="icon"
                                      aria-label={t("copyModelId")}
                                      onClick={() => void copyModelId(model.canonicalId)}
                                    >
                                      {copiedModel === model.canonicalId ? <Check aria-hidden="true" /> : <Clipboard aria-hidden="true" />}
                                    </Button>
                                  </TooltipTrigger>
                                  <TooltipContent>{copiedModel === model.canonicalId ? tCommon("copied") : t("copyModelId")}</TooltipContent>
                                </Tooltip>
                              </div>
                            </TableCell>
                          </TableRow>
                          {expanded ? (
                            <TableRow className="bg-muted/20 hover:bg-muted/20">
                              <TableCell colSpan={9} className="p-0">
                                <ModelRowDetails model={model} formatPrice={formatPrice} formatNumber={numberFormatter.format} />
                              </TableCell>
                            </TableRow>
                          ) : null}
                        </Fragment>
                      );
                    })}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>

            <Pagination
              page={page}
              totalPages={totalPages}
              total={totalModels}
              pageSize={PAGE_SIZE}
              onPageChange={setPage}
            />
          </>
        ) : null}
      </section>
    </TooltipProvider>
  );
}

function FilterSelect({
  label,
  allLabel,
  value,
  options,
  onValueChange,
}: {
  label: string;
  allLabel: string;
  value: string;
  options: FacetOption[];
  onValueChange: (value: string) => void;
}) {
  const id = useId();

  return (
    <div className="flex min-w-0 flex-col gap-1.5">
      <Label htmlFor={id}>{label}</Label>
      <Select value={value || ALL_VALUE} onValueChange={(next) => onValueChange(next === ALL_VALUE ? "" : next)}>
        <SelectTrigger id={id}>
          <SelectValue />
        </SelectTrigger>
        <SelectContent position="popper">
          <SelectGroup>
            <SelectItem value={ALL_VALUE}>{allLabel}</SelectItem>
            {options.map((item) => <SelectItem key={item.slug} value={item.slug}>{item.name}</SelectItem>)}
          </SelectGroup>
        </SelectContent>
      </Select>
    </div>
  );
}

function RangeFilter({
  label,
  value,
  min,
  max,
  step,
  formatValue,
  onValueChange,
}: {
  label: string;
  value: [number, number];
  min: number;
  max: number;
  step: number;
  formatValue: (value: number) => string;
  onValueChange: (value: [number, number]) => void;
}) {
  const id = useId();
  return (
    <div className="flex min-w-0 flex-col gap-3 rounded-lg border border-border bg-muted/30 p-3">
      <div className="flex items-center justify-between gap-3">
        <Label id={id}>{label}</Label>
        <span className="text-xs tabular-nums text-muted-foreground" dir="ltr">
          {formatValue(value[0])} – {formatValue(value[1])}
        </span>
      </div>
      <Slider
        dir="ltr"
        aria-labelledby={id}
        min={min}
        max={max}
        step={step}
        value={value}
        onValueChange={(next) => onValueChange(next as [number, number])}
      />
      <div className="flex items-center justify-between text-xs tabular-nums text-muted-foreground" dir="ltr" aria-hidden="true">
        <span>{formatValue(min)}</span>
        <span>{formatValue(max)}</span>
      </div>
    </div>
  );
}

function PriceCell({
  value,
  formatPrice,
}: {
  value?: number | null;
  formatPrice: (value?: number | null) => string;
}) {
  return (
    <TableCell className="whitespace-nowrap font-mono text-sm tabular-nums" dir="ltr">
      {value === undefined || value === null ? "—" : formatPrice(value)}
    </TableCell>
  );
}

function FeatureIcons({
  model,
  labels,
}: {
  model: ModelRow;
  labels: Record<FeatureKey, string>;
}) {
  const visibleFeatures: FeatureKey[] = [];
  if (hasFeature(model, "vision")) visibleFeatures.push("vision");
  if (hasFeature(model, "tools")) visibleFeatures.push("tools");
  if (hasFeature(model, "reasoning")) visibleFeatures.push("reasoning");

  if (!visibleFeatures.length) return <span className="text-muted-foreground">—</span>;

  return (
    <div className="flex items-center gap-1">
      {visibleFeatures.map((feature) => (
        <Tooltip key={feature}>
          <TooltipTrigger asChild>
            <span className="inline-flex size-7 items-center justify-center rounded-md border border-border bg-muted/40 text-muted-foreground">
              <FeatureIcon feature={feature} />
              <span className="sr-only">{labels[feature]}</span>
            </span>
          </TooltipTrigger>
          <TooltipContent>{labels[feature]}</TooltipContent>
        </Tooltip>
      ))}
    </div>
  );
}

function FeatureIcon({ feature }: { feature: FeatureKey }) {
  if (feature === "vision") return <Eye aria-hidden="true" />;
  if (feature === "tools") return <Wrench aria-hidden="true" />;
  return <Brain aria-hidden="true" />;
}

function hasFeature(model: ModelRow, feature: FeatureKey) {
  const capabilities = model.capabilities.map((item) => item.toLowerCase());
  if (feature === "vision") {
    return capabilities.includes("vision") || model.inputModalities.includes("image");
  }
  if (feature === "tools") {
    return capabilities.some((item) => ["function_calling", "tool_calling", "tool_choice", "tools"].includes(item));
  }
  return capabilities.some((item) => item.includes("reason"));
}

function ModelRowDetails({
  model,
  formatPrice,
  formatNumber,
}: {
  model: ModelRow;
  formatPrice: (value?: number | null) => string;
  formatNumber: (value: number) => string;
}) {
  const t = useTranslations("Catalog");
  const pricingDetails = [
    [t("detail.imageInput"), model.pricing?.imageInputPrice],
    [t("detail.imageOutput"), model.pricing?.imageOutputPrice],
    [t("detail.audioInput"), model.pricing?.audioInputPrice],
    [t("detail.audioOutput"), model.pricing?.audioOutputPrice],
    [t("detail.request"), model.pricing?.requestPrice],
  ].filter((entry): entry is [string, number] => typeof entry[1] === "number");

  return (
    <div className="grid gap-4 border-y border-border px-6 py-5 lg:grid-cols-[1fr_2fr]">
      <div>
        <p className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">{t("detail.title")}</p>
        <dl className="mt-3 grid grid-cols-2 gap-3">
          <DetailValue label={t("contextWindow")} value={model.contextWindow ? formatNumber(model.contextWindow) : "—"} />
          <DetailValue label={t("maxOutputTokens")} value={model.maxOutputTokens ? formatNumber(model.maxOutputTokens) : "—"} />
          <DetailValue label={t("detail.modes")} value={model.mode?.join("، ") || "—"} technical />
          <DetailValue label={t("table.modalities")} value={[...model.inputModalities, ...model.outputModalities].join("، ") || "—"} technical />
        </dl>
      </div>
      <div>
        <p className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">{t("detail.additionalPricing")}</p>
        {pricingDetails.length ? (
          <div className="mt-3 flex flex-wrap gap-2">
            {pricingDetails.map(([label, value]) => (
              <span key={label} className="inline-flex items-center gap-2 rounded-lg border border-border bg-background px-3 py-2 text-xs">
                <span className="text-muted-foreground">{label}</span>
                <strong className="font-mono" dir="ltr">{formatPrice(value)}</strong>
              </span>
            ))}
          </div>
        ) : (
          <p className="mt-3 text-sm text-muted-foreground">{t("detail.noAdditionalPricing")}</p>
        )}
      </div>
    </div>
  );
}

function DetailValue({
  label,
  value,
  technical = false,
}: {
  label: string;
  value: string;
  technical?: boolean;
}) {
  return (
    <div>
      <dt className="text-xs text-muted-foreground">{label}</dt>
      <dd className="mt-1 truncate text-sm font-medium" dir={technical ? "ltr" : undefined}>{value}</dd>
    </div>
  );
}

function Pagination({
  page,
  totalPages,
  total,
  pageSize,
  onPageChange,
}: {
  page: number;
  totalPages: number;
  total: number;
  pageSize: number;
  onPageChange: (page: number) => void;
}) {
  const t = useTranslations("Catalog");
  const formatter = useMemo(() => new Intl.NumberFormat("fa-IR"), []);
  const start = total === 0 ? 0 : (page - 1) * pageSize + 1;
  const end = Math.min(page * pageSize, total);

  return (
    <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <p className="text-sm text-muted-foreground">
        {t("paginationSummary", {
          start: formatter.format(start),
          end: formatter.format(end),
          total: formatter.format(total),
        })}
      </p>
      <div className="flex items-center gap-1" dir="ltr">
        <Button type="button" variant="outline" size="icon" disabled={page <= 1} aria-label={t("firstPage")} onClick={() => onPageChange(1)}>
          <ChevronsLeft aria-hidden="true" />
        </Button>
        <Button type="button" variant="outline" size="icon" disabled={page <= 1} aria-label={t("previousPage")} onClick={() => onPageChange(page - 1)}>
          <ChevronLeft aria-hidden="true" />
        </Button>
        <span className="min-w-20 text-center text-sm tabular-nums">
          {formatter.format(page)} / {formatter.format(totalPages)}
        </span>
        <Button type="button" variant="outline" size="icon" disabled={page >= totalPages} aria-label={t("nextPage")} onClick={() => onPageChange(page + 1)}>
          <ChevronRight aria-hidden="true" />
        </Button>
        <Button type="button" variant="outline" size="icon" disabled={page >= totalPages} aria-label={t("lastPage")} onClick={() => onPageChange(totalPages)}>
          <ChevronsRight aria-hidden="true" />
        </Button>
      </div>
    </div>
  );
}

function ModelsTableSkeleton({ label }: { label: string }) {
  return (
    <Card className="overflow-hidden p-0" role="status" aria-label={label}>
      <CardContent className="p-0 sm:p-0">
        <Table>
          <TableHeader className="bg-muted/45">
            <TableRow>
              {Array.from({ length: 9 }, (_, index) => (
                <TableHead key={index}><Skeleton className="h-4 w-20" /></TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody>
            {Array.from({ length: 8 }, (_, rowIndex) => (
              <TableRow key={rowIndex}>
                {Array.from({ length: 9 }, (_, cellIndex) => (
                  <TableCell key={cellIndex}><Skeleton className="h-8 w-full min-w-20" /></TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}

function getDetailCount(model: ModelRow) {
  const extraPrices = [
    model.pricing?.imageInputPrice,
    model.pricing?.imageOutputPrice,
    model.pricing?.audioInputPrice,
    model.pricing?.audioOutputPrice,
    model.pricing?.requestPrice,
  ].filter((value) => typeof value === "number").length;
  return extraPrices || (model.contextWindow || model.maxOutputTokens || model.mode?.length ? 1 : 0);
}

function sortPreviewModels(models: ModelRow[], sort: SortValue) {
  const sorted = [...models];
  if (sort === "name") return sorted.sort((a, b) => a.slug.localeCompare(b.slug));
  if (sort === "cheapest_input") return sorted.sort((a, b) => (a.pricing?.inputPerMillionTokens ?? Infinity) - (b.pricing?.inputPerMillionTokens ?? Infinity));
  if (sort === "cheapest_output") return sorted.sort((a, b) => (a.pricing?.outputPerMillionTokens ?? Infinity) - (b.pricing?.outputPerMillionTokens ?? Infinity));
  if (sort === "largest_context") return sorted.sort((a, b) => (b.contextWindow ?? 0) - (a.contextWindow ?? 0));
  return sorted;
}

function formatRelativeDate(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "—";
  const days = Math.round((date.getTime() - Date.now()) / 86_400_000);
  const formatter = new Intl.RelativeTimeFormat("fa-IR", { numeric: "auto" });
  if (Math.abs(days) < 30) return formatter.format(days, "day");
  const months = Math.round(days / 30);
  if (Math.abs(months) < 12) return formatter.format(months, "month");
  return formatter.format(Math.round(months / 12), "year");
}
