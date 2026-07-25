"use client";

import { useId, useMemo, useState } from "react";
import Link from "next/link";
import { useTranslations } from "next-intl";
import { useQuery } from "@tanstack/react-query";
import {
  ArrowLeft,
  Boxes,
  CircleAlert,
  Info,
  Search,
  SlidersHorizontal,
} from "lucide-react";
import { CompareToggleButton } from "@/components/compare/compare-toggle-button";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
} from "@/components/ui/card";
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
import { apiGet } from "@/lib/api";
import { cn } from "@/lib/utils";
import { useDisplayFormatters } from "@/lib/use-display-formatters";

type ModelRow = {
  canonicalId: string;
  slug: string;
  name: string;
  provider: { slug: string; name: string; logoUrl?: string | null };
  modelSeries?: { slug: string; name: string; logoUrl?: string | null } | null;
  description?: string;
  status: string;
  contextWindow?: number;
  maxOutputTokens?: number;
  inputModalities: string[];
  outputModalities: string[];
  capabilities: string[];
  pricing?: { inputPerMillionTokens?: number; outputPerMillionTokens?: number } | null;
  free: boolean;
  catalogPreview?: boolean;
};

type FacetOption = { slug: string; name: string };
type FacetsResponse = {
  providers?: FacetOption[];
  inputModalities?: FacetOption[];
  input_modalities?: FacetOption[];
  outputModalities?: FacetOption[];
  output_modalities?: FacetOption[];
  data?: {
    providers?: FacetOption[];
    inputModalities?: FacetOption[];
    input_modalities?: FacetOption[];
    outputModalities?: FacetOption[];
    output_modalities?: FacetOption[];
  };
};

type StatusKey =
  | "active"
  | "inactive"
  | "disabled"
  | "maintenance"
  | "deprecated"
  | "preview"
  | "available"
  | "unavailable";

const ALL_VALUE = "__all__";
const CONTEXT_LIMIT = 2_000_000;
const PRICE_LIMIT = 100;

const PREVIEW_MODELS: ModelRow[] = [
  {
    canonicalId: "openai/gpt-4.1-mini",
    slug: "gpt-4.1-mini",
    name: "GPT-4.1 Mini",
    provider: { slug: "openai", name: "OpenAI" },
    status: "preview",
    contextWindow: 1_000_000,
    maxOutputTokens: 32_768,
    inputModalities: ["text", "image"],
    outputModalities: ["text"],
    capabilities: ["vision", "tool_calling", "json_mode", "streaming"],
    pricing: { inputPerMillionTokens: 0.4 },
    free: false,
    catalogPreview: true,
  },
  {
    canonicalId: "anthropic/claude-sonnet-4",
    slug: "claude-sonnet-4",
    name: "Claude Sonnet 4",
    provider: { slug: "anthropic", name: "Anthropic" },
    status: "preview",
    contextWindow: 200_000,
    inputModalities: ["text", "image"],
    outputModalities: ["text"],
    capabilities: [],
    free: false,
    catalogPreview: true,
  },
  {
    canonicalId: "google-gemini/gemini-2.5-flash",
    slug: "gemini-2.5-flash",
    name: "Gemini 2.5 Flash",
    provider: { slug: "google-gemini", name: "Google Gemini" },
    status: "preview",
    contextWindow: 1_000_000,
    inputModalities: ["text", "image", "audio"],
    outputModalities: ["text"],
    capabilities: [],
    free: false,
    catalogPreview: true,
  },
  {
    canonicalId: "mistral-ai/mistral-small-latest",
    slug: "mistral-small-latest",
    name: "Mistral Small",
    provider: { slug: "mistral-ai", name: "Mistral AI" },
    status: "preview",
    contextWindow: 128_000,
    inputModalities: ["text"],
    outputModalities: ["text"],
    capabilities: [],
    free: false,
    catalogPreview: true,
  },
  {
    canonicalId: "cohere/command-r-plus",
    slug: "command-r-plus",
    name: "Command R+",
    provider: { slug: "cohere", name: "Cohere" },
    status: "preview",
    contextWindow: 128_000,
    inputModalities: ["text"],
    outputModalities: ["text"],
    capabilities: [],
    free: false,
    catalogPreview: true,
  },
  {
    canonicalId: "groq/llama-3.3-70b-versatile",
    slug: "llama-3.3-70b-versatile",
    name: "Llama 3.3 70B",
    provider: { slug: "groq", name: "Groq" },
    status: "preview",
    contextWindow: 128_000,
    inputModalities: ["text"],
    outputModalities: ["text"],
    capabilities: [],
    free: false,
    catalogPreview: true,
  },
];

function getStatusKey(status: string): StatusKey | null {
  const normalized = status.toLowerCase().replaceAll("_", "-");
  const keys: StatusKey[] = [
    "active",
    "inactive",
    "disabled",
    "maintenance",
    "deprecated",
    "preview",
    "available",
    "unavailable",
  ];
  return keys.includes(normalized as StatusKey) ? (normalized as StatusKey) : null;
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
      <Select
        value={value || ALL_VALUE}
        onValueChange={(nextValue) =>
          onValueChange(nextValue === ALL_VALUE ? "" : nextValue)
        }
      >
        <SelectTrigger id={id}>
          <SelectValue />
        </SelectTrigger>
        <SelectContent position="popper" sideOffset={6}>
          <SelectGroup>
            <SelectItem value={ALL_VALUE}>{allLabel}</SelectItem>
            {options.map((item) => (
              <SelectItem key={item.slug} value={item.slug}>
                {item.name}
              </SelectItem>
            ))}
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
        onValueChange={(nextValue) =>
          onValueChange(nextValue as [number, number])
        }
      />
      <div
        className="flex items-center justify-between text-xs tabular-nums text-muted-foreground"
        dir="ltr"
        aria-hidden="true"
      >
        <span>{formatValue(min)}</span>
        <span>{formatValue(max)}</span>
      </div>
    </div>
  );
}

export function ModelsExplorer() {
  const t = useTranslations("Catalog");
  const tCommon = useTranslations("Common");
  const tStatus = useTranslations("Statuses");
  const { formatPrice } = useDisplayFormatters();
  const numberFormatter = useMemo(() => new Intl.NumberFormat("fa-IR"), []);
  const compactNumberFormatter = useMemo(
    () =>
      new Intl.NumberFormat("fa-IR", {
        notation: "compact",
        maximumFractionDigits: 1,
      }),
    [],
  );
  const priceRangeFormatter = useMemo(
    () =>
      new Intl.NumberFormat("fa-IR", {
        maximumFractionDigits: 1,
      }),
    [],
  );
  const [q, setQ] = useState("");
  const [provider, setProvider] = useState("");
  const [inputModality, setInputModality] = useState("");
  const [outputModality, setOutputModality] = useState("");
  const [contextRange, setContextRange] = useState<[number, number]>([
    0,
    CONTEXT_LIMIT,
  ]);
  const [priceRange, setPriceRange] = useState<[number, number]>([
    0,
    PRICE_LIMIT,
  ]);
  const [advancedFiltersOpen, setAdvancedFiltersOpen] = useState(false);

  const contextRangeActive =
    contextRange[0] !== 0 || contextRange[1] !== CONTEXT_LIMIT;
  const priceRangeActive =
    priceRange[0] !== 0 || priceRange[1] !== PRICE_LIMIT;

  const params = useMemo(() => {
    const searchParams = new URLSearchParams();
    if (q) searchParams.set("q", q);
    if (provider) searchParams.set("provider", provider);
    if (inputModality) searchParams.set("input_modalities", inputModality);
    if (outputModality) searchParams.set("output_modalities", outputModality);
    if (contextRange[0] > 0) {
      searchParams.set("min_context_window", String(contextRange[0]));
    }
    if (contextRange[1] < CONTEXT_LIMIT) {
      searchParams.set("max_context_window", String(contextRange[1]));
    }
    if (priceRange[0] > 0) {
      searchParams.set("min_input_price", String(priceRange[0]));
    }
    if (priceRange[1] < PRICE_LIMIT) {
      searchParams.set("max_input_price", String(priceRange[1]));
    }
    searchParams.set("limit", "24");
    return searchParams.toString();
  }, [
    q,
    provider,
    inputModality,
    outputModality,
    contextRange,
    priceRange,
  ]);

  const modelsQuery = useQuery({
    queryKey: ["models", params],
    queryFn: () =>
      apiGet<{ data: ModelRow[]; pagination: { total: number } }>(
        `/models?${params}`,
      ),
  });
  const facetsQuery = useQuery({
    queryKey: ["facets"],
    queryFn: () => apiGet<FacetsResponse>("/facets"),
  });

  const providers =
    facetsQuery.data?.providers ??
    facetsQuery.data?.data?.providers ??
    [];
  const inputModalities =
    facetsQuery.data?.inputModalities ??
    facetsQuery.data?.input_modalities ??
    facetsQuery.data?.data?.inputModalities ??
    facetsQuery.data?.data?.input_modalities ??
    [];
  const outputModalities =
    facetsQuery.data?.outputModalities ??
    facetsQuery.data?.output_modalities ??
    facetsQuery.data?.data?.outputModalities ??
    facetsQuery.data?.data?.output_modalities ??
    [];
  const activeFilterCount =
    [q, provider, inputModality, outputModality].filter(Boolean).length +
    Number(contextRangeActive) +
    Number(priceRangeActive);
  const liveModels = modelsQuery.data?.data ?? [];
  const showingPreview =
    !modelsQuery.isLoading &&
    !modelsQuery.isError &&
    liveModels.length === 0 &&
    activeFilterCount === 0;
  const displayedModels = showingPreview ? PREVIEW_MODELS : liveModels;

  const clearFilters = () => {
    setQ("");
    setProvider("");
    setInputModality("");
    setOutputModality("");
    setContextRange([0, CONTEXT_LIMIT]);
    setPriceRange([0, PRICE_LIMIT]);
  };

  const formatStatus = (model: ModelRow) => {
    if (model.catalogPreview) return t("previewRow");
    const key = getStatusKey(model.status);
    return key ? tStatus(key) : model.status;
  };

  return (
    <div className="flex min-w-0 flex-col gap-6">
      <Card className="overflow-hidden p-0 shadow-sm">
        <CardHeader className="border-b border-border bg-muted/35 sm:flex-row sm:items-start sm:justify-between sm:gap-4">
          <div className="flex min-w-0 flex-col gap-1">
            <div className="flex flex-wrap items-center gap-2">
              <h2 className="text-base font-semibold sm:text-lg">
                {t("filtersTitle")}
              </h2>
              {activeFilterCount > 0 ? (
                <Badge className="border-primary-soft bg-primary-soft text-primary-soft-foreground">
                  {t("activeFilters", { count: activeFilterCount })}
                </Badge>
              ) : null}
            </div>
            <CardDescription>{t("filtersDescription")}</CardDescription>
          </div>
        </CardHeader>

        <CardContent className="flex flex-col gap-4 pt-4 sm:pt-5">
          <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
            <div className="flex min-w-0 flex-col gap-1.5 sm:col-span-2 xl:col-span-1">
              <Label htmlFor="catalog-model-search">{t("searchLabel")}</Label>
              <div className="relative">
                <Search
                  className="pointer-events-none absolute start-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground"
                  aria-hidden="true"
                />
                <Input
                  id="catalog-model-search"
                  className="ps-10"
                  placeholder={t("searchPlaceholder")}
                  value={q}
                  onChange={(event) => setQ(event.target.value)}
                />
              </div>
            </div>
            <FilterSelect
              label={t("providerLabel")}
              allLabel={t("allProviders")}
              value={provider}
              options={providers}
              onValueChange={setProvider}
            />
            <FilterSelect
              label={t("inputModalityLabel")}
              allLabel={t("allInputModalities")}
              value={inputModality}
              options={inputModalities}
              onValueChange={setInputModality}
            />
            <FilterSelect
              label={t("outputModalityLabel")}
              allLabel={t("allOutputModalities")}
              value={outputModality}
              options={outputModalities}
              onValueChange={setOutputModality}
            />
          </div>

          <div
            id="catalog-advanced-filters"
            className={cn(
              "grid gap-3 lg:grid-cols-2",
              !advancedFiltersOpen && "hidden lg:grid",
            )}
          >
            <RangeFilter
              label={t("contextRange")}
              value={contextRange}
              min={0}
              max={CONTEXT_LIMIT}
              step={1_000}
              formatValue={(value) => compactNumberFormatter.format(value)}
              onValueChange={setContextRange}
            />
            <RangeFilter
              label={t("inputPriceRange")}
              value={priceRange}
              min={0}
              max={PRICE_LIMIT}
              step={0.1}
              formatValue={(value) => `$${priceRangeFormatter.format(value)}`}
              onValueChange={setPriceRange}
            />
          </div>
        </CardContent>

        <CardFooter className="justify-between gap-3 border-t border-border bg-muted/20">
          <Button
            type="button"
            variant="outline"
            aria-expanded={advancedFiltersOpen}
            aria-controls="catalog-advanced-filters"
            className="lg:hidden"
            onClick={() => setAdvancedFiltersOpen((open) => !open)}
          >
            <SlidersHorizontal data-icon="inline-start" aria-hidden="true" />
            {advancedFiltersOpen
              ? t("hideAdvancedFilters")
              : t("showAdvancedFilters")}
          </Button>
          {activeFilterCount > 0 ? (
            <Button
              type="button"
              variant="ghost"
              className="ms-auto"
              onClick={clearFilters}
            >
              {t("clearFilters", { count: activeFilterCount })}
            </Button>
          ) : (
            <span className="text-xs text-muted-foreground">
              {t("rangeFilterHint")}
            </span>
          )}
        </CardFooter>
      </Card>

      <section
        aria-labelledby="catalog-results-title"
        className="flex min-w-0 flex-col gap-4"
      >
        <div className="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
          <div className="flex flex-col gap-1">
            <h2
              id="catalog-results-title"
              className="text-base font-semibold sm:text-lg"
            >
              {t("resultsTitle")}
            </h2>
            <p className="text-sm leading-6 text-muted-foreground">
              {t("resultsDescription")}
            </p>
          </div>
          {!modelsQuery.isLoading && !modelsQuery.isError ? (
            <Badge aria-live="polite" className="self-start px-3 py-1 sm:self-auto">
              {showingPreview
                ? t("previewModelsCount", { count: PREVIEW_MODELS.length })
                : t("modelsFound", {
                    count: modelsQuery.data?.pagination.total ?? 0,
                  })}
            </Badge>
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

        {modelsQuery.isLoading ? (
          <ModelsTableSkeleton label={t("loading")} />
        ) : null}

        {modelsQuery.isError ? (
          <Alert variant="destructive">
            <CircleAlert aria-hidden="true" />
            <AlertTitle>{t("loadError")}</AlertTitle>
            <AlertDescription>{t("loadErrorDescription")}</AlertDescription>
          </Alert>
        ) : null}

        {!modelsQuery.isLoading &&
        !modelsQuery.isError &&
        displayedModels.length === 0 ? (
          <Empty className="border border-card-border bg-card">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <Boxes aria-hidden="true" />
              </EmptyMedia>
              <EmptyTitle>{t("noMatchesTitle")}</EmptyTitle>
              <EmptyDescription>{t("noMatches")}</EmptyDescription>
            </EmptyHeader>
            <EmptyContent>
              {activeFilterCount > 0 ? (
                <Button type="button" onClick={clearFilters}>
                  {t("clearFilters", { count: activeFilterCount })}
                </Button>
              ) : (
                <Button asChild>
                  <Link href="/providers">{t("browseProviders")}</Link>
                </Button>
              )}
            </EmptyContent>
          </Empty>
        ) : null}

        {!modelsQuery.isLoading &&
        !modelsQuery.isError &&
        displayedModels.length > 0 ? (
          <ModelsTable
            models={displayedModels}
            formatStatus={formatStatus}
            formatPrice={formatPrice}
            formatNumber={(value) => numberFormatter.format(value)}
            labels={{
              model: t("table.model"),
              provider: tCommon("provider"),
              modalities: t("table.modalities"),
              context: t("contextWindow"),
              inputPrice: t("inputPrice"),
              status: tCommon("status"),
              actions: t("table.actions"),
              input: t("table.input"),
              output: t("table.output"),
              viewDetails: t("viewDetails"),
              preview: t("previewRow"),
              unavailable: tCommon("unknown"),
            }}
          />
        ) : null}
      </section>
    </div>
  );
}

function ModelsTable({
  models,
  formatStatus,
  formatPrice,
  formatNumber,
  labels,
}: {
  models: ModelRow[];
  formatStatus: (model: ModelRow) => string;
  formatPrice: (value?: number) => string;
  formatNumber: (value: number) => string;
  labels: {
    model: string;
    provider: string;
    modalities: string;
    context: string;
    inputPrice: string;
    status: string;
    actions: string;
    input: string;
    output: string;
    viewDetails: string;
    preview: string;
    unavailable: string;
  };
}) {
  return (
    <Card className="overflow-hidden p-0 shadow-sm">
      <CardContent className="p-0 sm:p-0">
        <Table>
          <TableHeader className="bg-muted/45">
            <TableRow>
              <TableHead className="min-w-64">{labels.model}</TableHead>
              <TableHead>{labels.provider}</TableHead>
              <TableHead>{labels.modalities}</TableHead>
              <TableHead>{labels.context}</TableHead>
              <TableHead>{labels.inputPrice}</TableHead>
              <TableHead>{labels.status}</TableHead>
              <TableHead className="text-end">{labels.actions}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {models.map((model) => {
              const statusKey = getStatusKey(model.status);
              const statusClassName = cn(
                !model.catalogPreview &&
                  (statusKey === "active" || statusKey === "available") &&
                  "border-success-soft bg-success-soft text-success-soft-foreground",
                model.catalogPreview &&
                  "border-primary-soft bg-primary-soft text-primary-soft-foreground",
              );

              return (
                <TableRow key={model.canonicalId}>
                  <TableCell>
                    <div className="flex min-w-0 flex-col gap-1">
                      {model.catalogPreview ? (
                        <span className="font-semibold" dir="auto">
                          {model.name}
                        </span>
                      ) : (
                        <Link
                          href={`/models/${model.provider.slug}/${model.slug}`}
                          className="font-semibold text-foreground transition-colors hover:text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                          dir="auto"
                        >
                          {model.name}
                        </Link>
                      )}
                      <span
                        className="max-w-64 truncate text-xs text-muted-foreground"
                        dir="ltr"
                        data-technical
                        title={model.canonicalId}
                      >
                        {model.canonicalId}
                      </span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <span className="font-medium">{model.provider.name}</span>
                  </TableCell>
                  <TableCell>
                    <div className="flex min-w-48 flex-col gap-1.5">
                      <ModalityList
                        label={labels.input}
                        values={model.inputModalities}
                        unavailable={labels.unavailable}
                      />
                      <ModalityList
                        label={labels.output}
                        values={model.outputModalities}
                        unavailable={labels.unavailable}
                      />
                    </div>
                  </TableCell>
                  <TableCell className="tabular-nums" dir="ltr">
                    {model.contextWindow === undefined
                      ? "—"
                      : formatNumber(model.contextWindow)}
                  </TableCell>
                  <TableCell className="tabular-nums" dir="ltr">
                    {model.pricing?.inputPerMillionTokens === undefined
                      ? "—"
                      : formatPrice(model.pricing.inputPerMillionTokens)}
                  </TableCell>
                  <TableCell>
                    <Badge className={statusClassName}>
                      {formatStatus(model)}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center justify-end gap-1">
                      {model.catalogPreview ? (
                        <Button type="button" variant="outline" size="sm" disabled>
                          {labels.preview}
                        </Button>
                      ) : (
                        <>
                          <CompareToggleButton
                            canonicalId={model.canonicalId}
                            compact
                            iconOnly
                          />
                          <Button asChild variant="ghost" size="icon">
                            <Link
                              href={`/models/${model.provider.slug}/${model.slug}`}
                              aria-label={`${labels.viewDetails}: ${model.name}`}
                              title={labels.viewDetails}
                            >
                              <ArrowLeft aria-hidden="true" />
                            </Link>
                          </Button>
                        </>
                      )}
                    </div>
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}

function ModalityList({
  label,
  values,
  unavailable,
}: {
  label: string;
  values: string[];
  unavailable: string;
}) {
  return (
    <div className="flex min-w-0 items-center gap-1.5">
      <span className="shrink-0 text-xs text-muted-foreground">{label}</span>
      <span className="truncate text-xs" dir="auto">
        {values.length > 0 ? values.slice(0, 3).join("، ") : unavailable}
      </span>
    </div>
  );
}

function ModelsTableSkeleton({ label }: { label: string }) {
  return (
    <Card
      className="overflow-hidden p-0"
      role="status"
      aria-label={label}
    >
      <CardContent className="p-0 sm:p-0">
        <Table>
          <TableHeader className="bg-muted/45">
            <TableRow>
              {Array.from({ length: 7 }, (_, index) => (
                <TableHead key={index}>
                  <Skeleton className="h-4 w-20" />
                </TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody>
            {Array.from({ length: 6 }, (_, rowIndex) => (
              <TableRow key={rowIndex}>
                {Array.from({ length: 7 }, (_, cellIndex) => (
                  <TableCell key={cellIndex}>
                    <Skeleton className="h-8 w-full min-w-20" />
                  </TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}
