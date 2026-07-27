import { describe, expect, it } from "vitest";

import {
  buildDraftsForAccount,
  buildSelectedModelPayload,
  clearVisibleModelSelection,
  toggleVisibleModelSelection,
  type WizardAccountModel,
  type WizardCatalogModel,
  type WizardModelDraft,
} from "./provider-account-wizard.models";
import type { ModelPrice } from "./pricing";

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

const visibleModels: WizardCatalogModel[] = [
  {
    model: "gpt-4.1",
    types: ["chat"],
    public_pricing: tokenPricing(1, 2),
  },
  {
    model: "gpt-4.1-mini",
    types: ["chat", "embedding"],
    public_pricing: tokenPricing(0.5, 1),
  },
];

function makeDraft(partial: Partial<WizardModelDraft>): WizardModelDraft {
  return {
    enabled: false,
    selectedCapabilities: [],
    pricingMode: "public",
    ...partial,
  };
}

describe("provider-account-wizard model helpers", () => {
  it("toggleVisibleModelSelection changes enabled state", () => {
    const drafts = {
      "gpt-4.1": makeDraft({ enabled: false }),
      "gpt-4.1-mini": makeDraft({ enabled: false }),
    };

    const next = toggleVisibleModelSelection(visibleModels, drafts);

    expect(next["gpt-4.1"].enabled).toBe(true);
    expect(next["gpt-4.1-mini"].enabled).toBe(true);
    expect(next["gpt-4.1"].selectedCapabilities).toEqual(["chat"]);
    expect(next["gpt-4.1-mini"].selectedCapabilities).toEqual([]);
  });

  it("clearVisibleModelSelection clears only enabled state", () => {
    const drafts = {
      "gpt-4.1": makeDraft({ enabled: true }),
      "gpt-4.1-mini": makeDraft({ enabled: true }),
    };

    const next = clearVisibleModelSelection(visibleModels, drafts);

    expect(next["gpt-4.1"].enabled).toBe(false);
    expect(next["gpt-4.1-mini"].enabled).toBe(false);
  });

  it("buildSelectedModelPayload saves enabled models only", () => {
    const drafts = {
      "gpt-4.1": makeDraft({
        enabled: true,
        selectedCapabilities: ["chat", "embedding"],
        pricingMode: "private",
        privatePricing: tokenPricing(1.25, 2),
      }),
      "gpt-4.1-mini": makeDraft({
        enabled: false,
        pricingMode: "public",
      }),
    };

    const result = buildSelectedModelPayload(drafts);

    expect(result.validationErrors).toEqual([]);
    expect(result.models).toEqual([
      {
        model: "gpt-4.1",
        effective_capabilities: ["chat", "embedding"],
        pricing_mode: "private",
        private_pricing: {
          status: "priced",
          profiles: [{
            region: "global",
            currency: "USD",
            rates: [
              { key: "input-token", usage: "input", unit: "token", amount: "1.25", quantity: 1_000_000, service_tier: "standard" },
              { key: "output-token", usage: "output", unit: "token", amount: "2", quantity: 1_000_000, service_tier: "standard" },
            ],
          }],
        },
      },
    ]);
  });

  it("buildSelectedModelPayload reports private-pricing validation errors for enabled models", () => {
    const drafts = {
      "gpt-4.1": makeDraft({
        enabled: true,
        selectedCapabilities: ["chat"],
        pricingMode: "private",
      }),
    };

    const result = buildSelectedModelPayload(drafts);

    expect(result.models).toHaveLength(1);
    expect(result.validationErrors).toEqual([{ model: "gpt-4.1", reason: "private_pricing_required" }]);
  });

  it("buildSelectedModelPayload strips private pricing when policy disallows custom pricing", () => {
    const drafts = {
      "gpt-4.1": makeDraft({
        enabled: true,
        selectedCapabilities: ["chat"],
        pricingMode: "private",
        privatePricing: tokenPricing(1.25, 2.5),
      }),
    };

    const result = buildSelectedModelPayload(drafts, { canUseCustomPricing: false });

    expect(result.validationErrors).toEqual([]);
    expect(result.models).toEqual([
      {
        model: "gpt-4.1",
        effective_capabilities: ["chat"],
        pricing_mode: "public",
        private_pricing: undefined,
      },
    ]);
  });

  it("buildSelectedModelPayload reports capability validation errors for enabled models", () => {
    const drafts = {
      "gpt-4.1": makeDraft({
        enabled: true,
        selectedCapabilities: [],
        pricingMode: "public",
      }),
    };

    const result = buildSelectedModelPayload(drafts);

    expect(result.models).toEqual([
      {
        model: "gpt-4.1",
        effective_capabilities: [],
        pricing_mode: "public",
        private_pricing: undefined,
      },
    ]);
    expect(result.validationErrors).toEqual([{ model: "gpt-4.1", reason: "effective_capabilities_required" }]);
  });

  it("buildDraftsForAccount preserves saved capabilities", () => {
    const accountModels: WizardAccountModel[] = [
      {
        model: "gpt-4.1-mini",
        enabled: true,
        types: ["Embedding", "chat", "embedding"],
        pricing_mode: "public",
      },
    ];

    const drafts = buildDraftsForAccount(accountModels, visibleModels);

    expect(drafts["gpt-4.1-mini"].enabled).toBe(true);
    expect(drafts["gpt-4.1-mini"].selectedCapabilities).toEqual(["embedding", "chat"]);
  });
});
