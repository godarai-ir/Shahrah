import { describe, expect, it } from "vitest";

import {
  buildProviderAccountModelFqn,
  createInitialModelFilters,
  filterAndSortProviderModels,
  getActiveModelFilterChips,
  getModelFamily,
  getRecommendedModel,
  hasInvalidModelPriceRange,
  normalizeProviderScopedModelName,
} from "./model-editor-utils";
import type { ProviderCatalogModel } from "./types";
import type { ModelPrice } from "@/app/components/pricing";

function tokenPricing(input: number, output: number): ModelPrice {
  return {
    status: "priced",
    profiles: [{
      region: "global",
      currency: "USD",
      rates: [
        { key: "input-token", usage: "input", unit: "token", amount: String(input), quantity: 1_000_000, service_tier: "standard" },
        { key: "output-token", usage: "output", unit: "token", amount: String(output), quantity: 1_000_000, service_tier: "standard" },
      ],
    }],
  };
}

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
    public_pricing: tokenPricing(1, 4),
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
    public_pricing: tokenPricing(5, 15),
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
    public_pricing: tokenPricing(2, 8),
  },
];

describe("model-editor-utils", () => {
  it("builds a three-part model FQN from the pending account identity", () => {
    expect(buildProviderAccountModelFqn({
      account: null,
      pendingProvider: "AI21",
      pendingAccountName: "ai21-brisk-orbit",
      model: {
        provider: "ai21",
        model: "ai21/jamba-large-1.7",
      },
    })).toBe("ai21:ai21-brisk-orbit:jamba-large-1.7");
  });

  it("prefers the persisted account identity when one is available", () => {
    expect(buildProviderAccountModelFqn({
      account: { provider: "openai", name: "production" },
      pendingProvider: "ai21",
      pendingAccountName: "draft",
      model: {
        provider: "openai",
        model: "openai/gpt-4o-mini",
      },
    })).toBe("openai:production:gpt-4o-mini");
  });

  it("filters models by free-text search across model metadata", () => {
    const result = filterAndSortProviderModels({
      models,
      filters: {
        ...createInitialModelFilters(),
        search: "vision",
      },
      isSelectedOnly: false,
      isModelEnabled: () => false,
    });

    expect(result.map((model) => model.model)).toEqual(["gpt-vision"]);
  });

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
        allFamilies: "Family",
      },
    );

    expect(chips).toEqual([
      { key: "minPrice", label: "Min: 1" },
      { key: "featureFilter", label: "Feature: vision" },
    ]);
  });

  it("prefers a stable chat alias over a dated model version", () => {
    const recommended = getRecommendedModel([
      { ...models[0], model: "claude-sonnet-4-20250514", display_name: "Claude Sonnet 4 (2025-05-14)" },
      { ...models[0], model: "claude-sonnet-4", display_name: "Claude Sonnet 4" },
    ]);

    expect(recommended?.model).toBe("claude-sonnet-4");
    expect(getModelFamily(recommended!)).toBe("Claude Sonnet");
  });

  it("detects invalid price ranges", () => {
    expect(hasInvalidModelPriceRange({ minPrice: "10", maxPrice: "2" })).toBe(true);
    expect(hasInvalidModelPriceRange({ minPrice: "2", maxPrice: "10" })).toBe(false);
  });

  it("normalizes provider-scoped model names", () => {
    expect(normalizeProviderScopedModelName("openai", "openai/gpt-4.1")).toBe("gpt-4.1");
    expect(normalizeProviderScopedModelName("openai", "anthropic/claude")).toBe("anthropic/claude");
  });
});
