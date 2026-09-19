import { describe, expect, it } from "vitest";

import { EMPTY_OBSERVE_FILTERS } from "@/lib/observe-filters";

import { buildBreakdownRows, buildCacheBreakdown, buildChartData, buildMetricsFilters, buildSavingsChartData, normalizeStringList } from "./utils";

describe("buildChartData", () => {
  it("uses raw request counts for the request chart", () => {
    const data = buildChartData(
      [
        {
          timestamp: "2026-06-21T08:30:00.000Z",
          modelName: "openai/gpt-4.1-mini",
          count_model_name: 7,
          p99_latency_ms: 1200,
        },
      ],
      "modelName",
      "p99",
      "Unknown",
    );

    expect(data.points).toHaveLength(1);
    expect(data.points[0].series0).toBe(7);
  });

  it("includes dates in axis labels when the selected range spans multiple days", () => {
    const data = buildChartData(
      [
        {
          timestamp: "2026-06-21T08:30:00.000Z",
          modelName: "openai/gpt-4.1-mini",
          count_model_name: 3,
        },
      ],
      "modelName",
      "p99",
      "Unknown",
      "en-US",
      "2026-01-20T08:32:00.000Z",
      "2026-06-21T08:32:00.000Z",
    );

    expect(data.points[0].label).toContain("Jun 21");
  });
});

describe("cache savings helpers", () => {
  it("separates simple exact and smart semantic savings", () => {
    const breakdown = buildCacheBreakdown([
      { cache_mode: "simple", cache_status: "hit", cache_match: "exact", count_model_name: 3, sum_avoided_cost_usd: 1.25 },
      { cache_mode: "smart", cache_status: "hit", cache_match: "semantic", count_model_name: 2, sum_avoided_cost_usd: 2.5 },
      { cache_mode: "smart", cache_status: "miss", cache_match: "none", count_model_name: 4, sum_avoided_cost_usd: 0 },
    ]);

    expect(breakdown[0]).toMatchObject({ mode: "simple", exactHits: 3, semanticHits: 0, avoidedCost: 1.25 });
    expect(breakdown[1]).toMatchObject({ mode: "smart", exactHits: 0, semanticHits: 2, misses: 4, avoidedCost: 2.5 });
  });

  it("builds actual, simple, and smart cost series", () => {
    const points = buildSavingsChartData([
      { timestamp: "2026-06-21T08:30:00.000Z", cache_mode: "none", sum_cost_usd: 2, sum_avoided_cost_usd: 0 },
      { timestamp: "2026-06-21T08:30:00.000Z", cache_mode: "simple", sum_cost_usd: 0, sum_avoided_cost_usd: 1 },
      { timestamp: "2026-06-21T08:30:00.000Z", cache_mode: "smart", sum_cost_usd: 0, sum_avoided_cost_usd: 3 },
    ]);

    expect(points[0]).toMatchObject({ actualCost: 2, simpleSavings: 1, smartSavings: 3 });
  });
});

describe("metrics filter option helpers", () => {
  it("dedupes and sorts historical facet values", () => {
    expect(normalizeStringList(["openai:gpt-4o", " anthropic/claude-3-5 ", "openai:gpt-4o", 42, ""])).toEqual([
      "anthropic/claude-3-5",
      "openai:gpt-4o",
    ]);
  });
});

describe("metrics breakdown table", () => {
  const rows = [
    { modelName: "fast", count_model_name: 30, sum_cost_usd: 1, p90_latency_ms: 200 },
    { modelName: "slow", count_model_name: 10, sum_cost_usd: 3, p90_latency_ms: 900 },
  ];

  it("gives additive metrics a share of the total", () => {
    const breakdown = buildBreakdownRows(rows, "modelName", "requests", "p90", 0, "Unknown");
    expect(breakdown).toEqual([
      { label: "fast", value: 30, share: 0.75 },
      { label: "slow", value: 10, share: 0.25 },
    ]);
    expect(buildBreakdownRows(rows, "modelName", "cost", "p90", 0, "Unknown")[0]).toMatchObject({ label: "slow", share: 0.75 });
  });

  it("compares latency with the overall percentile instead of producing a share", () => {
    const breakdown = buildBreakdownRows(rows, "modelName", "latency", "p90", 450, "Unknown");
    expect(breakdown.map((row) => row.share)).toEqual([undefined, undefined]);
    expect(breakdown[0]).toMatchObject({ label: "slow", value: 900, ratio: 2 });
    expect(breakdown[1].ratio).toBeCloseTo(200 / 450);
  });

  it("leaves the latency ratio empty when the overall percentile is unknown", () => {
    expect(buildBreakdownRows(rows, "modelName", "latency", "p90", 0, "Unknown").map((row) => row.ratio)).toEqual([undefined, undefined]);
  });
});

describe("shared observe filters on metrics", () => {
  it("maps request ID and status onto metrics fields and leaves the trace error code out", () => {
    expect(buildMetricsFilters({ ...EMPTY_OBSERVE_FILTERS, q: "req_1", status: "503", provider: "openai", errorCode: "timeout" })).toEqual([
      { fieldName: "providerName", operator: "EQUAL", value: "openai" },
      { fieldName: "requestID", operator: "EQUAL", value: "req_1" },
      { fieldName: "statusCode", operator: "EQUAL", value: 503 },
    ]);
  });
});
