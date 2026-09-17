import { describe, expect, it } from "vitest";

import { filterCatalogModels, matchesModelQuery, summarizeModelReadiness } from "./readiness-summary";
import type { CapabilityReadiness, CatalogModel, ModelAvailability } from "./types";

function capability(state: CapabilityReadiness["state"]): CapabilityReadiness {
  return { endpoint: "/v1/chat/completions", state, live_validated: false };
}

function availability(overrides: Partial<ModelAvailability> = {}): ModelAvailability {
  return {
    account_id: "acc-1",
    account_name: "openai:main",
    fqn: "openai:main/gpt-4.1",
    credential_source: "byok",
    effective_capabilities: ["chat"],
    pricing_mode: "public",
    pricing_resolved: true,
    readiness: "ready",
    capability_readiness: [capability("contract_ready")],
    ...overrides,
  };
}

function model(overrides: Partial<CatalogModel> = {}): CatalogModel {
  return {
    provider: "openai",
    model: "gpt-4.1",
    display_name: "GPT 4.1",
    types: ["chat"],
    availability: [availability()],
    ...overrides,
  };
}

describe("summarizeModelReadiness", () => {
  it("is ready when every capability has a ready contract", () => {
    const item = model();
    expect(summarizeModelReadiness(item, item.availability[0])).toBe("ready");
  });

  it("is ready when no capability rows are reported", () => {
    const item = model({ availability: [availability({ capability_readiness: undefined })] });
    expect(summarizeModelReadiness(item, item.availability[0])).toBe("ready");
  });

  it("flags unresolved pricing from either the availability or a capability", () => {
    const unresolved = model({ availability: [availability({ pricing_resolved: false })] });
    expect(summarizeModelReadiness(unresolved, unresolved.availability[0])).toBe("pricing");

    const incomplete = model({
      availability: [availability({ capability_readiness: [capability("contract_ready"), capability("pricing_incomplete")] })],
    });
    expect(summarizeModelReadiness(incomplete, incomplete.availability[0])).toBe("pricing");
  });

  it("marks unsupported transports and lifecycle blocks as limited", () => {
    const transport = model({ availability: [availability({ capability_readiness: [capability("unsupported_transport")] })] });
    expect(summarizeModelReadiness(transport, transport.availability[0])).toBe("limited");

    const lifecycle = model({ availability: [availability({ capability_readiness: [capability("blocked_lifecycle")] })] });
    expect(summarizeModelReadiness(lifecycle, lifecycle.availability[0])).toBe("limited");
  });

  it("applies deprecated > pricing > limited priority", () => {
    const rows = [capability("unsupported_transport"), capability("pricing_incomplete")];
    const pricing = model({ availability: [availability({ capability_readiness: rows })] });
    expect(summarizeModelReadiness(pricing, pricing.availability[0])).toBe("pricing");

    const deprecated = model({ is_deprecated: true, availability: [availability({ capability_readiness: rows })] });
    expect(summarizeModelReadiness(deprecated, deprecated.availability[0])).toBe("deprecated");
  });

  it("falls back to catalog readiness when there is no availability", () => {
    const item = model({ availability: [], readiness: [capability("unsupported_transport")] });
    expect(summarizeModelReadiness(item, null)).toBe("limited");
  });
});

describe("matchesModelQuery", () => {
  it("matches name, display name and capability types case-insensitively", () => {
    const item = model({ types: ["chat", "embedding"], capabilities: ["vision"] });
    expect(matchesModelQuery(item, "GPT-4")).toBe(true);
    expect(matchesModelQuery(item, "gpt 4.1")).toBe(true);
    expect(matchesModelQuery(item, "embed")).toBe(true);
    expect(matchesModelQuery(item, "vision")).toBe(true);
    expect(matchesModelQuery(item, "whisper")).toBe(false);
    expect(matchesModelQuery(item, "  ")).toBe(true);
  });
});

describe("filterCatalogModels", () => {
  const models = [
    model({ model: "gpt-4.1" }),
    model({ model: "gpt-4o", availability: [availability({ pricing_resolved: false })] }),
    model({ model: "gpt-3.5", is_deprecated: true, replacement_model: "gpt-4.1" }),
    model({ model: "text-embedding-3", display_name: "Text Embedding 3", types: ["embedding"] }),
  ];

  it("counts readiness over the search results, independent of the active chip", () => {
    const result = filterCatalogModels(models, "acc-1", "gpt", "ready");
    expect(result.rows.map((row) => row.model.model)).toEqual(["gpt-4.1"]);
    expect(result.counts).toEqual({ deprecated: 1, pricing: 1, limited: 0, ready: 1 });
    expect(result.queryMatches).toBe(3);
  });

  it("returns every model when no filter is set", () => {
    expect(filterCatalogModels(models, "acc-1", "", "").rows).toHaveLength(4);
  });
});
