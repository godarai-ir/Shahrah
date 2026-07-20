import type { ProviderCatalogModel } from "./types";
import {
  getModelCapabilities,
  getModelDiscoveryCatalogBaseURL,
  normalizeModelDiscoveryCatalogBaseURL,
  toNumberOrUndefined,
} from "./utils";

export type ModelFilterState = {
  search: string;
  minPrice: string;
  maxPrice: string;
  minContext: string;
  inputModality: string;
  outputModality: string;
  featureFilter: string;
  familyFilter: string;
  sortBy: "price_asc" | "price_desc" | "context_asc" | "context_desc";
};

export type ModelFilterLabels = {
  minPrice: string;
  maxPrice: string;
  minContext: string;
  allInputModalities: string;
  allOutputModalities: string;
  allFeatures: string;
  allFamilies: string;
};

export type DisplayPriceOptions = {
  currency: "USD" | "TOMAN";
  scale: "PER_MILLION_TOKENS" | "PER_TOKEN";
  usdToIrr: number;
  naLabel: string;
};

export function createInitialModelFilters(search = ""): ModelFilterState {
  return {
    search,
    minPrice: "",
    maxPrice: "",
    minContext: "",
    inputModality: "",
    outputModality: "",
    featureFilter: "",
    familyFilter: "",
    sortBy: "price_asc",
  };
}

export function filterAndSortProviderModels({
  models,
  filters,
  isSelectedOnly,
  isModelEnabled,
}: {
  models: ProviderCatalogModel[];
  filters: ModelFilterState;
  isSelectedOnly: boolean;
  isModelEnabled: (modelKey: string) => boolean;
}): ProviderCatalogModel[] {
  const minPriceValue = toNumberOrUndefined(filters.minPrice);
  const maxPriceValue = toNumberOrUndefined(filters.maxPrice);
  const minContextValue = toNumberOrUndefined(filters.minContext);
  const normalizedSearch = filters.search.trim().toLowerCase();
  const filtered = models.filter((model) => {
    const inputPrice = model.public_pricing?.input_usd_per_1m;
    const context = model.context_window;
    const searchableValues = [
      model.display_name,
      model.model,
      model.provider,
      ...(model.types ?? []),
      ...(model.mode ?? []),
      ...(model.modes ?? []),
      ...(model.input_modalities ?? []),
      ...(model.output_modalities ?? []),
      ...(model.capabilities ?? []),
      ...(model.features ?? []),
    ];
    if (isSelectedOnly && !isModelEnabled(model.model)) return false;
    if (
      normalizedSearch &&
      !searchableValues.some((value) => value?.toLowerCase().includes(normalizedSearch))
    ) {
      return false;
    }
    if (typeof minPriceValue === "number" && (typeof inputPrice !== "number" || inputPrice < minPriceValue)) return false;
    if (typeof maxPriceValue === "number" && (typeof inputPrice !== "number" || inputPrice > maxPriceValue)) return false;
    if (typeof minContextValue === "number" && (typeof context !== "number" || context < minContextValue)) return false;
    if (filters.inputModality && !(model.input_modalities ?? []).includes(filters.inputModality)) return false;
    if (filters.outputModality && !(model.output_modalities ?? []).includes(filters.outputModality)) return false;
    if (filters.featureFilter && !getModelCapabilities(model).includes(filters.featureFilter)) return false;
    if (filters.familyFilter && getModelFamily(model) !== filters.familyFilter) return false;
    return true;
  });

  const sorted = [...filtered];
  sorted.sort((a, b) => {
    const aPrice = a.public_pricing?.input_usd_per_1m ?? Number.POSITIVE_INFINITY;
    const bPrice = b.public_pricing?.input_usd_per_1m ?? Number.POSITIVE_INFINITY;
    const aContext = a.context_window ?? 0;
    const bContext = b.context_window ?? 0;
    switch (filters.sortBy) {
      case "price_desc":
        return bPrice - aPrice;
      case "context_asc":
        return aContext - bContext;
      case "context_desc":
        return bContext - aContext;
      case "price_asc":
      default:
        return aPrice - bPrice;
    }
  });
  return sorted;
}

export function getActiveModelFilterChips(filters: ModelFilterState, labels: ModelFilterLabels) {
  const chips: Array<{ key: keyof ModelFilterState; label: string }> = [];
  if (filters.minPrice.trim()) chips.push({ key: "minPrice", label: `${labels.minPrice}: ${filters.minPrice}` });
  if (filters.maxPrice.trim()) chips.push({ key: "maxPrice", label: `${labels.maxPrice}: ${filters.maxPrice}` });
  if (filters.minContext.trim()) chips.push({ key: "minContext", label: `${labels.minContext}: ${filters.minContext}` });
  if (filters.inputModality.trim()) chips.push({ key: "inputModality", label: `${labels.allInputModalities}: ${filters.inputModality}` });
  if (filters.outputModality.trim()) chips.push({ key: "outputModality", label: `${labels.allOutputModalities}: ${filters.outputModality}` });
  if (filters.featureFilter.trim()) chips.push({ key: "featureFilter", label: `${labels.allFeatures}: ${filters.featureFilter}` });
  if (filters.familyFilter.trim()) chips.push({ key: "familyFilter", label: `${labels.allFamilies}: ${filters.familyFilter}` });
  return chips;
}

export function getModelFamily(model: ProviderCatalogModel): string {
  const raw = (model.display_name || model.model.split("/").at(-1) || model.model).replaceAll(/[_-]+/g, " ");
  const words = raw.split(/\s+/).filter(Boolean);
  const familyWords = words.filter((word) => !/^v?\d+(?:\.\d+)*$/.test(word)).slice(0, 2);
  return familyWords.join(" ") || words[0] || model.provider;
}

export function getRecommendedModel(models: ProviderCatalogModel[]): ProviderCatalogModel | null {
  if (models.length === 0) return null;
  const chatModels = models.filter((model) => {
    const values = [...(model.types ?? []), ...(model.mode ?? []), ...(model.modes ?? [])];
    return values.some((value) => value.toLowerCase().includes("chat"));
  });
  const candidates = chatModels.length > 0 ? chatModels : models;
  return [...candidates].sort((a, b) => {
    const datedA = /(?:^|[-_/])20\d{2}[-_]?\d{2}[-_]?\d{2}(?:$|[-_/])/.test(a.model) ? 1 : 0;
    const datedB = /(?:^|[-_/])20\d{2}[-_]?\d{2}[-_]?\d{2}(?:$|[-_/])/.test(b.model) ? 1 : 0;
    if (datedA !== datedB) return datedA - datedB;
    const priceA = a.public_pricing?.input_usd_per_1m ?? Number.POSITIVE_INFINITY;
    const priceB = b.public_pricing?.input_usd_per_1m ?? Number.POSITIVE_INFINITY;
    if (priceA !== priceB) return priceA - priceB;
    return (b.context_window ?? 0) - (a.context_window ?? 0);
  })[0] ?? null;
}

export function hasInvalidModelPriceRange(filters: Pick<ModelFilterState, "minPrice" | "maxPrice">): boolean {
  const min = toNumberOrUndefined(filters.minPrice);
  const max = toNumberOrUndefined(filters.maxPrice);
  if (typeof min !== "number" || typeof max !== "number") return false;
  return min > max;
}

export function normalizeProviderScopedModelName(provider: string, modelName: string) {
  const normalizedProvider = provider.trim().toLowerCase();
  const normalizedModelName = modelName.trim();
  if (!normalizedProvider || !normalizedModelName) return normalizedModelName;
  const prefix = `${normalizedProvider}/`;
  if (normalizedModelName.toLowerCase().startsWith(prefix)) {
    return normalizedModelName.slice(prefix.length);
  }
  return normalizedModelName;
}

export function buildProviderAccountModelFqn(input: {
  account?: { provider: string; name: string } | null;
  pendingProvider?: string;
  pendingAccountName?: string;
  model: Pick<ProviderCatalogModel, "provider" | "model">;
}): string {
  const provider = (
    input.account?.provider
    || input.pendingProvider
    || input.model.provider
  ).trim().toLowerCase();
  const accountName = (
    input.account?.name
    || input.pendingAccountName
    || ""
  ).trim();

  if (!provider || !accountName) return input.model.model;

  const officialModel = normalizeProviderScopedModelName(provider, input.model.model);
  return `${provider}:${accountName}:${officialModel}`;
}

export function buildModelDiscoveryDetailsHref(model: ProviderCatalogModel): string {
  const base = process.env.NEXT_PUBLIC_MODEL_DISCOVERY_APP_URL
    ? normalizeModelDiscoveryCatalogBaseURL(process.env.NEXT_PUBLIC_MODEL_DISCOVERY_APP_URL)
    : getModelDiscoveryCatalogBaseURL();
  const provider = encodeURIComponent(model.provider);
  const canonical = model.model.startsWith(`${model.provider}/`)
    ? model.model.slice(model.provider.length + 1)
    : model.model;
  const path = canonical
    .split("/")
    .filter(Boolean)
    .map((part) => encodeURIComponent(part))
    .join("/");
  return `${base}/models/${provider}/${path || encodeURIComponent(model.model)}`;
}
