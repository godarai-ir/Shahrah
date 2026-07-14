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
  sortBy: "price_asc" | "price_desc" | "context_asc" | "context_desc";
};

export type CompareSortMode = "input_asc" | "output_asc" | "context_desc" | "name_asc";

export type ModelFilterLabels = {
  minPrice: string;
  maxPrice: string;
  minContext: string;
  allInputModalities: string;
  allOutputModalities: string;
  allFeatures: string;
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
  return chips;
}

export function hasInvalidModelPriceRange(filters: Pick<ModelFilterState, "minPrice" | "maxPrice">): boolean {
  const min = toNumberOrUndefined(filters.minPrice);
  const max = toNumberOrUndefined(filters.maxPrice);
  if (typeof min !== "number" || typeof max !== "number") return false;
  return min > max;
}

export function sortComparedModels(models: ProviderCatalogModel[], sortBy: CompareSortMode): ProviderCatalogModel[] {
  const items = [...models];
  items.sort((a, b) => {
    const aInput = a.public_pricing?.input_usd_per_1m ?? Number.POSITIVE_INFINITY;
    const bInput = b.public_pricing?.input_usd_per_1m ?? Number.POSITIVE_INFINITY;
    const aOutput = a.public_pricing?.output_usd_per_1m ?? Number.POSITIVE_INFINITY;
    const bOutput = b.public_pricing?.output_usd_per_1m ?? Number.POSITIVE_INFINITY;
    const aContext = a.context_window ?? 0;
    const bContext = b.context_window ?? 0;
    const aName = (a.display_name || a.model).toLowerCase();
    const bName = (b.display_name || b.model).toLowerCase();
    switch (sortBy) {
      case "output_asc":
        return aOutput - bOutput;
      case "context_desc":
        return bContext - aContext;
      case "name_asc":
        return aName.localeCompare(bName);
      case "input_asc":
      default:
        return aInput - bInput;
    }
  });
  return items;
}

export function getCompareKpis(compareModels: ProviderCatalogModel[]) {
  const validInput = compareModels.filter((item) => typeof item.public_pricing?.input_usd_per_1m === "number");
  const validOutput = compareModels.filter((item) => typeof item.public_pricing?.output_usd_per_1m === "number");
  const validContext = compareModels.filter((item) => typeof item.context_window === "number");
  const cheapestInput = validInput.sort((a, b) => (a.public_pricing?.input_usd_per_1m ?? 0) - (b.public_pricing?.input_usd_per_1m ?? 0))[0];
  const cheapestOutput = validOutput.sort((a, b) => (a.public_pricing?.output_usd_per_1m ?? 0) - (b.public_pricing?.output_usd_per_1m ?? 0))[0];
  const largestContext = validContext.sort((a, b) => (b.context_window ?? 0) - (a.context_window ?? 0))[0];
  const maxFeatures = compareModels.reduce(
    (best, item) => {
      const count = getModelCapabilities(item).length;
      return count > best.count ? { model: item, count } : best;
    },
    { model: null as ProviderCatalogModel | null, count: -1 },
  );
  return { cheapestInput, cheapestOutput, largestContext, mostFeatures: maxFeatures.model };
}

export function getBestCompareMarkers(compareModels: ProviderCatalogModel[]) {
  const minInput = Math.min(...compareModels.map((item) => item.public_pricing?.input_usd_per_1m ?? Number.POSITIVE_INFINITY));
  const minOutput = Math.min(...compareModels.map((item) => item.public_pricing?.output_usd_per_1m ?? Number.POSITIVE_INFINITY));
  const maxContext = Math.max(...compareModels.map((item) => item.context_window ?? 0));
  return {
    input: Number.isFinite(minInput) ? minInput : null,
    output: Number.isFinite(minOutput) ? minOutput : null,
    context: maxContext > 0 ? maxContext : null,
  };
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
