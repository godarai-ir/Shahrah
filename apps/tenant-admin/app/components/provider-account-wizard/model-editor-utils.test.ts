import { describe, expect, it } from "vitest";

import {
  createInitialModelFilters,
  filterAndSortProviderModels,
  getActiveModelFilterChips,
  getBestCompareMarkers,
  getCompareKpis,
  hasInvalidModelPriceRange,
  normalizeProviderScopedModelName,
  sortComparedModels,
} from "./model-editor-utils";
import type { ProviderCatalogModel } from "./types";

const models: ProviderCatalogModel[] = [
  {
    provider: "openai",
    model: "openai/gpt-fast",
    display_name: "GPT Fast",
    types: ["chat"],
    context_window: 128_000,
    input_modalities: ["text"],
    output_modalities: ["text"],
    capabilities: ["tools"],
    public_pricing: { input_usd_per_1m: 1, output_usd_per_1m: 4 },
  },
  {
    provider: "openai",
    model: "gpt-vision",
    display_name: "GPT Vision",
    types: ["chat"],
    context_window: 32_000,
    input_modalities: ["text", "image"],
    output_modalities: ["text"],
    capabilities: ["vision"],
    public_pricing: { input_usd_per_1m: 5, output_usd_per_1m: 15 },
  },
  {
    provider: "openai",
    model: "gpt-long",
    display_name: "GPT Long",
    types: ["chat"],
    context_window: 1_000_000,
    input_modalities: ["text"],
    output_modalities: ["text"],
    capabilities: ["tools", "reasoning"],
    public_pricing: { input_usd_per_1m: 2, output_usd_per_1m: 8 },
  },
];

describe("model-editor-utils", () => {
  it("filters selected models and sorts by context", () => {
    const result = filterAndSortProviderModels({
      models,
      filters: {
        ...createInitialModelFilters(),
        minPrice: "1",
        maxPrice: "3",
        inputModality: "text",
        sortBy: "context_desc",
      },
      isSelectedOnly: true,
      isModelEnabled: (modelKey) => modelKey === "openai/gpt-fast" || modelKey === "gpt-long",
    });

    expect(result.map((model) => model.model)).toEqual(["gpt-long", "openai/gpt-fast"]);
  });

  it("builds active filter chips from non-empty filters", () => {
    const chips = getActiveModelFilterChips(
      { ...createInitialModelFilters(), minPrice: "1", featureFilter: "vision" },
      {
        minPrice: "Min",
        maxPrice: "Max",
        minContext: "Context",
        allInputModalities: "Input",
        allOutputModalities: "Output",
        allFeatures: "Feature",
      },
    );

    expect(chips).toEqual([
      { key: "minPrice", label: "Min: 1" },
      { key: "featureFilter", label: "Feature: vision" },
    ]);
  });

  it("detects invalid price ranges", () => {
    expect(hasInvalidModelPriceRange({ minPrice: "10", maxPrice: "2" })).toBe(true);
    expect(hasInvalidModelPriceRange({ minPrice: "2", maxPrice: "10" })).toBe(false);
  });

  it("sorts compared models and computes compare summaries", () => {
    const compared = sortComparedModels(models, "context_desc");
    expect(compared.map((model) => model.model)).toEqual(["gpt-long", "openai/gpt-fast", "gpt-vision"]);

    expect(getCompareKpis(models).cheapestInput?.model).toBe("openai/gpt-fast");
    expect(getCompareKpis(models).largestContext?.model).toBe("gpt-long");
    expect(getBestCompareMarkers(models)).toEqual({ input: 1, output: 4, context: 1_000_000 });
  });

  it("normalizes provider-scoped model names", () => {
    expect(normalizeProviderScopedModelName("openai", "openai/gpt-4.1")).toBe("gpt-4.1");
    expect(normalizeProviderScopedModelName("openai", "anthropic/claude")).toBe("anthropic/claude");
  });
});
