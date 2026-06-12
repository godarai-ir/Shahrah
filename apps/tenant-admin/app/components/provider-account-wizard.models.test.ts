import { describe, expect, it } from "vitest";

import {
  buildSelectedModelPayload,
  clearVisibleModelSelection,
  toggleVisibleModelSelection,
  type WizardCatalogModel,
  type WizardModelDraft,
} from "./provider-account-wizard.models";

const visibleModels: WizardCatalogModel[] = [
  {
    model: "gpt-4.1",
    public_pricing: { input_usd_per_1m: 1, output_usd_per_1m: 2 },
  },
  {
    model: "gpt-4.1-mini",
    public_pricing: { input_usd_per_1m: 0.5, output_usd_per_1m: 1 },
  },
];

function makeDraft(partial: Partial<WizardModelDraft>): WizardModelDraft {
  return {
    enabled: false,
    compared: false,
    pricingMode: "public",
    inputPrice: "",
    outputPrice: "",
    ...partial,
  };
}

describe("provider-account-wizard model helpers", () => {
  it("toggleVisibleModelSelection changes enabled state without touching compared state", () => {
    const drafts = {
      "gpt-4.1": makeDraft({ enabled: false, compared: true }),
      "gpt-4.1-mini": makeDraft({ enabled: false, compared: false }),
    };

    const next = toggleVisibleModelSelection(visibleModels, drafts);

    expect(next["gpt-4.1"].enabled).toBe(true);
    expect(next["gpt-4.1-mini"].enabled).toBe(true);
    expect(next["gpt-4.1"].compared).toBe(true);
    expect(next["gpt-4.1-mini"].compared).toBe(false);
  });

  it("clearVisibleModelSelection clears only enabled state", () => {
    const drafts = {
      "gpt-4.1": makeDraft({ enabled: true, compared: true }),
      "gpt-4.1-mini": makeDraft({ enabled: true, compared: false }),
    };

    const next = clearVisibleModelSelection(visibleModels, drafts);

    expect(next["gpt-4.1"].enabled).toBe(false);
    expect(next["gpt-4.1-mini"].enabled).toBe(false);
    expect(next["gpt-4.1"].compared).toBe(true);
    expect(next["gpt-4.1-mini"].compared).toBe(false);
  });

  it("buildSelectedModelPayload saves enabled models only and ignores compare-only state", () => {
    const drafts = {
      "gpt-4.1": makeDraft({
        enabled: true,
        compared: false,
        pricingMode: "private",
        inputPrice: "1.25",
        outputPrice: "",
      }),
      "gpt-4.1-mini": makeDraft({
        enabled: false,
        compared: true,
        pricingMode: "public",
      }),
    };

    const result = buildSelectedModelPayload(drafts);

    expect(result.validationErrors).toEqual([]);
    expect(result.models).toEqual([
      {
        model: "gpt-4.1",
        pricing_mode: "private",
        private_pricing: {
          input_usd_per_1m: 1.25,
          output_usd_per_1m: undefined,
        },
      },
    ]);
  });

  it("buildSelectedModelPayload reports private-pricing validation errors for enabled models", () => {
    const drafts = {
      "gpt-4.1": makeDraft({
        enabled: true,
        pricingMode: "private",
        inputPrice: "",
        outputPrice: "",
      }),
    };

    const result = buildSelectedModelPayload(drafts);

    expect(result.models).toHaveLength(1);
    expect(result.validationErrors).toEqual(["gpt-4.1"]);
  });
});
