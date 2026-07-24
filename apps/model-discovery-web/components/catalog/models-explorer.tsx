"use client";

import { useId, useMemo, useState } from "react";
import Link from "next/link";
import { useTranslations } from "next-intl";
import { useQuery } from "@tanstack/react-query";
import {
  ArrowLeft,
  Building2,
  CircleAlert,
  Search,
  SlidersHorizontal,
} from "lucide-react";
import { CompareToggleButton } from "@/components/compare/compare-toggle-button";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { apiGet, resolveAssetUrl } from "@/lib/api";
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

type StatusKey = "active" | "inactive" | "disabled" | "maintenance" | "deprecated" | "preview" | "available" | "unavailable";

const ALL_VALUE = "__all__";

function getStatusKey(status: string): StatusKey | null {
  const normalized = status.toLowerCase().replaceAll("_", "-");
  const keys: StatusKey[] = ["active", "inactive", "disabled", "maintenance", "deprecated", "preview", "available", "unavailable"];
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
      <Select value={value || ALL_VALUE} onValueChange={(nextValue) => onValueChange(nextValue === ALL_VALUE ? "" : nextValue)}>
        <SelectTrigger id={id}>
          <SelectValue />
        </SelectTrigger>
        <SelectContent position="popper" sideOffset={6}>
          <SelectGroup>
            <SelectItem value={ALL_VALUE}>{allLabel}</SelectItem>
            {options.map((item) => (
              <SelectItem key={item.slug} value={item.slug}>{item.name}</SelectItem>
            ))}
          </SelectGroup>
        </SelectContent>
      </Select>
    </div>
  );
}

function NumericFilter({
  id,
  label,
  value,
  step,
  onChange,
}: {
  id: string;
  label: string;
  value: string;
  step?: string;
  onChange: (value: string) => void;
}) {
  return (
    <div className="flex min-w-0 flex-col gap-1.5">
      <Label htmlFor={id}>{label}</Label>
      <Input
        id={id}
        dir="ltr"
        type="number"
        min="0"
        step={step}
        inputMode="decimal"
        value={value}
        onChange={(event) => onChange(event.target.value)}
      />
    </div>
  );
}

export function ModelsExplorer() {
  const t = useTranslations("Catalog");
  const tCommon = useTranslations("Common");
  const tStatus = useTranslations("Statuses");
  const { formatPrice } = useDisplayFormatters();
  const numberFormatter = useMemo(() => new Intl.NumberFormat("fa-IR"), []);
  const [q, setQ] = useState("");
  const [provider, setProvider] = useState("");
  const [inputModality, setInputModality] = useState("");
  const [outputModality, setOutputModality] = useState("");
  const [minContextWindow, setMinContextWindow] = useState("");
  const [maxContextWindow, setMaxContextWindow] = useState("");
  const [minPromptPrice, setMinPromptPrice] = useState("");
  const [maxPromptPrice, setMaxPromptPrice] = useState("");
  const [advancedFiltersOpen, setAdvancedFiltersOpen] = useState(false);

  const params = useMemo(() => {
    const searchParams = new URLSearchParams();
    if (q) searchParams.set("q", q);
    if (provider) searchParams.set("provider", provider);
    if (inputModality) searchParams.set("input_modalities", inputModality);
    if (outputModality) searchParams.set("output_modalities", outputModality);
    if (minContextWindow) searchParams.set("min_context_window", minContextWindow);
    if (maxContextWindow) searchParams.set("max_context_window", maxContextWindow);
    if (minPromptPrice) searchParams.set("min_input_price", minPromptPrice);
    if (maxPromptPrice) searchParams.set("max_input_price", maxPromptPrice);
    searchParams.set("limit", "24");
    return searchParams.toString();
  }, [q, provider, inputModality, outputModality, minContextWindow, maxContextWindow, minPromptPrice, maxPromptPrice]);

  const modelsQuery = useQuery({
    queryKey: ["models", params],
    queryFn: () => apiGet<{ data: ModelRow[]; pagination: { total: number } }>(`/models?${params}`),
  });
  const facetsQuery = useQuery({ queryKey: ["facets"], queryFn: () => apiGet<FacetsResponse>("/facets") });

  const providers = facetsQuery.data?.providers ?? facetsQuery.data?.data?.providers ?? [];
  const inputModalities = facetsQuery.data?.inputModalities ?? facetsQuery.data?.input_modalities ?? facetsQuery.data?.data?.inputModalities ?? facetsQuery.data?.data?.input_modalities ?? [];
  const outputModalities = facetsQuery.data?.outputModalities ?? facetsQuery.data?.output_modalities ?? facetsQuery.data?.data?.outputModalities ?? facetsQuery.data?.data?.output_modalities ?? [];
  const activeFilterCount = [q, provider, inputModality, outputModality, minContextWindow, maxContextWindow, minPromptPrice, maxPromptPrice].filter(Boolean).length;

  const clearFilters = () => {
    setQ("");
    setProvider("");
    setInputModality("");
    setOutputModality("");
    setMinContextWindow("");
    setMaxContextWindow("");
    setMinPromptPrice("");
    setMaxPromptPrice("");
  };

  const formatStatus = (status: string) => {
    const key = getStatusKey(status);
    return key ? tStatus(key) : status;
  };

  return (
    <div className="flex min-w-0 flex-col gap-6">
      <Card className="overflow-hidden p-0 shadow-sm lg:sticky lg:top-20 lg:z-10">
        <CardHeader className="border-b border-border bg-muted/35 sm:flex-row sm:items-start sm:justify-between sm:gap-4">
          <div className="flex min-w-0 flex-col gap-1">
            <div className="flex flex-wrap items-center gap-2">
              <h2 className="text-base font-semibold sm:text-lg">{t("filtersTitle")}</h2>
              {activeFilterCount > 0 ? (
                <Badge className="border-primary-soft bg-primary-soft text-primary-soft-foreground">
                  {t("activeFilters", { count: activeFilterCount })}
                </Badge>
              ) : null}
            </div>
            <CardDescription>{t("filtersDescription")}</CardDescription>
          </div>
          {activeFilterCount > 0 ? (
            <Button
              type="button"
              className="h-9 shrink-0 border border-border bg-background px-3 text-foreground hover:bg-accent"
              onClick={clearFilters}
            >
              {t("clearFilters", { count: activeFilterCount })}
            </Button>
          ) : null}
        </CardHeader>

        <CardContent className="flex flex-col gap-4 pt-4 sm:pt-5">
          <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
            <div className="flex min-w-0 flex-col gap-1.5 sm:col-span-2 xl:col-span-1">
              <Label htmlFor="catalog-model-search">{t("searchLabel")}</Label>
              <div className="relative">
                <Search className="pointer-events-none absolute start-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" aria-hidden="true" />
                <Input
                  id="catalog-model-search"
                  className="ps-10"
                  placeholder={t("searchPlaceholder")}
                  value={q}
                  onChange={(event) => setQ(event.target.value)}
                />
              </div>
            </div>
            <FilterSelect label={t("providerLabel")} allLabel={t("allProviders")} value={provider} options={providers} onValueChange={setProvider} />
            <FilterSelect label={t("inputModalityLabel")} allLabel={t("allInputModalities")} value={inputModality} options={inputModalities} onValueChange={setInputModality} />
            <FilterSelect label={t("outputModalityLabel")} allLabel={t("allOutputModalities")} value={outputModality} options={outputModalities} onValueChange={setOutputModality} />
          </div>

          <div id="catalog-advanced-filters" className={cn("grid gap-3 sm:grid-cols-2 xl:grid-cols-4", !advancedFiltersOpen && "hidden lg:grid")}>
            <NumericFilter id="catalog-min-context" label={t("minimumContext")} value={minContextWindow} onChange={setMinContextWindow} />
            <NumericFilter id="catalog-max-context" label={t("maximumContext")} value={maxContextWindow} onChange={setMaxContextWindow} />
            <NumericFilter id="catalog-min-price" label={t("minimumInputPrice")} value={minPromptPrice} step="0.0001" onChange={setMinPromptPrice} />
            <NumericFilter id="catalog-max-price" label={t("maximumInputPrice")} value={maxPromptPrice} step="0.0001" onChange={setMaxPromptPrice} />
          </div>

          <Button
            type="button"
            aria-expanded={advancedFiltersOpen}
            aria-controls="catalog-advanced-filters"
            className="h-9 self-start border border-border bg-background px-3 text-foreground hover:bg-accent lg:hidden"
            onClick={() => setAdvancedFiltersOpen((open) => !open)}
          >
            <SlidersHorizontal data-icon="inline-start" aria-hidden="true" />
            {advancedFiltersOpen ? t("hideAdvancedFilters") : t("showAdvancedFilters")}
          </Button>
        </CardContent>
      </Card>

      <section aria-labelledby="catalog-results-title" className="flex min-w-0 flex-col gap-4">
        <div className="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
          <div className="flex flex-col gap-1">
            <h2 id="catalog-results-title" className="text-base font-semibold sm:text-lg">{t("resultsTitle")}</h2>
            <p className="text-sm leading-6 text-muted-foreground">{t("resultsDescription")}</p>
          </div>
          {!modelsQuery.isLoading && !modelsQuery.isError ? (
            <Badge aria-live="polite" className="self-start px-3 py-1 sm:self-auto">
              {t("modelsFound", { count: modelsQuery.data?.pagination.total ?? 0 })}
            </Badge>
          ) : null}
        </div>

        {facetsQuery.isError ? (
          <ResultMessage tone="warning" title={t("filtersError")} />
        ) : null}

        {modelsQuery.isLoading ? (
          <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3" role="status" aria-label={t("loading")}>
            {Array.from({ length: 6 }, (_, index) => <ModelCardSkeleton key={index} />)}
          </div>
        ) : null}

        {modelsQuery.isError ? (
          <ResultMessage tone="error" title={t("loadError")} />
        ) : null}

        {!modelsQuery.isLoading && !modelsQuery.isError && modelsQuery.data?.data.length === 0 ? (
          <ResultMessage tone="neutral" title={t("noMatches")} action={activeFilterCount > 0 ? (
            <Button type="button" onClick={clearFilters}>{t("clearFilters", { count: activeFilterCount })}</Button>
          ) : (
            <Link
              href="/providers"
              className="inline-flex items-center justify-center rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            >
              {t("browseProviders")}
            </Link>
          )} />
        ) : null}

        {!modelsQuery.isLoading && !modelsQuery.isError && modelsQuery.data?.data.length ? (
          <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            {modelsQuery.data.data.map((model) => {
              const providerLogo = resolveAssetUrl(model.provider.logoUrl);
              const seriesLogo = resolveAssetUrl(model.modelSeries?.logoUrl);
              const statusKey = getStatusKey(model.status);
              const statusClassName = cn(
                (statusKey === "active" || statusKey === "available") && "border-success-soft bg-success-soft text-success-soft-foreground",
              );

              return (
                <Card key={model.canonicalId} className="group relative flex min-w-0 flex-col overflow-hidden p-0 transition hover:-translate-y-0.5 hover:border-primary/60 hover:shadow-lg">
                  <span className="absolute inset-x-0 top-0 h-1 bg-primary opacity-0 transition-opacity group-hover:opacity-100" aria-hidden="true" />
                  <div className="absolute end-3 top-3 z-10">
                    <CompareToggleButton canonicalId={model.canonicalId} compact iconOnly />
                  </div>
                  <Link
                    href={`/models/${model.provider.slug}/${model.slug}`}
                    className="flex h-full min-w-0 flex-col rounded-xl focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                  >
                    <CardHeader className="pe-16">
                      <div className="flex min-w-0 items-start gap-3">
                        <div className="flex shrink-0 items-center gap-1">
                          {providerLogo ? (
                            // eslint-disable-next-line @next/next/no-img-element
                            <img src={providerLogo} alt={model.provider.name} className="size-9 rounded-lg border border-border bg-background object-contain p-1" />
                          ) : (
                            <span className="flex size-9 items-center justify-center rounded-lg border border-border bg-muted text-muted-foreground">
                              <Building2 className="size-4" aria-hidden="true" />
                            </span>
                          )}
                          {seriesLogo ? (
                            // eslint-disable-next-line @next/next/no-img-element
                            <img src={seriesLogo} alt={t("modelSeriesLogoAlt", { name: model.modelSeries?.name ?? "" })} className="size-9 rounded-lg border border-border bg-background object-contain p-1" />
                          ) : null}
                        </div>
                        <div className="min-w-0 flex-1">
                          <div className="flex flex-wrap items-center gap-1.5">
                            <Badge className={statusClassName}>{formatStatus(model.status)}</Badge>
                            {model.free ? <Badge className="border-success-soft bg-success-soft text-success-soft-foreground">{t("free")}</Badge> : null}
                          </div>
                          <CardTitle className="mt-2 text-lg leading-7 transition-colors group-hover:text-primary">
                            <bdi dir="auto">{model.name}</bdi>
                          </CardTitle>
                          <CardDescription className="truncate text-xs" dir="ltr" data-technical>{model.canonicalId}</CardDescription>
                        </div>
                      </div>
                    </CardHeader>

                    <CardContent className="flex flex-1 flex-col gap-4">
                      <p dir="auto" className="line-clamp-2 min-h-12 text-sm leading-6 text-muted-foreground">
                        {model.description || t("noDescription")}
                      </p>

                      <dl className="grid grid-cols-2 gap-2">
                        <ModelMetric label={tCommon("provider")} value={model.provider.name} />
                        <ModelMetric label={t("contextWindow")} value={model.contextWindow === undefined ? "-" : numberFormatter.format(model.contextWindow)} />
                        <ModelMetric label={t("maxOutputTokens")} value={model.maxOutputTokens === undefined ? "-" : numberFormatter.format(model.maxOutputTokens)} />
                        <ModelMetric label={t("inputPrice")} value={formatPrice(model.pricing?.inputPerMillionTokens)} />
                      </dl>

                      {model.capabilities.length > 0 ? (
                        <div className="flex flex-wrap gap-1.5">
                          {model.capabilities.slice(0, 4).map((capability) => (
                            <Badge key={capability}><bdi dir="auto">{capability}</bdi></Badge>
                          ))}
                          {model.capabilities.length > 4 ? <Badge>+{numberFormatter.format(model.capabilities.length - 4)}</Badge> : null}
                        </div>
                      ) : null}
                    </CardContent>

                    <CardFooter className="justify-between gap-3 border-t border-border bg-muted/30">
                      <span className="text-sm font-medium text-primary">{t("viewDetails")}</span>
                      <ArrowLeft className="size-4 text-primary transition-transform group-hover:-translate-x-1" aria-hidden="true" />
                    </CardFooter>
                  </Link>
                </Card>
              );
            })}
          </div>
        ) : null}
      </section>
    </div>
  );
}

function ModelMetric({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="min-w-0 rounded-lg bg-muted/50 p-2.5">
      <dt className="text-xs text-muted-foreground">{label}</dt>
      <dd className="mt-1 truncate text-sm font-semibold" title={typeof value === "string" ? value : undefined}>{value}</dd>
    </div>
  );
}

function ModelCardSkeleton() {
  return (
    <Card className="flex flex-col gap-4 p-5" aria-hidden="true">
      <div className="flex items-start gap-3">
        <Skeleton className="size-9 shrink-0" />
        <div className="flex flex-1 flex-col gap-2">
          <Skeleton className="h-5 w-24" />
          <Skeleton className="h-6 w-2/3" />
          <Skeleton className="h-3 w-full" />
        </div>
      </div>
      <Skeleton className="h-12 w-full" />
      <div className="grid grid-cols-2 gap-2">
        <Skeleton className="h-14" />
        <Skeleton className="h-14" />
        <Skeleton className="h-14" />
        <Skeleton className="h-14" />
      </div>
    </Card>
  );
}

function ResultMessage({
  title,
  tone,
  action,
}: {
  title: string;
  tone: "error" | "warning" | "neutral";
  action?: React.ReactNode;
}) {
  return (
    <Card role={tone === "error" ? "alert" : "status"} className="flex flex-col items-start gap-4 p-5 sm:flex-row sm:items-center sm:justify-between">
      <div className="flex items-center gap-3">
        <span className={cn(
          "flex size-10 shrink-0 items-center justify-center rounded-xl",
          tone === "error" && "bg-destructive-soft text-destructive-soft-foreground",
          tone === "warning" && "bg-warning-soft text-warning-soft-foreground",
          tone === "neutral" && "bg-muted text-muted-foreground",
        )}>
          <CircleAlert className="size-5" aria-hidden="true" />
        </span>
        <p className="text-sm font-medium leading-6">{title}</p>
      </div>
      {action}
    </Card>
  );
}
